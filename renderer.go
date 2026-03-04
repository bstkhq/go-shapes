package shapes

import (
	"embed"
	"image"
	"image/color"
	"slices"

	"github.com/hajimehoshi/ebiten/v2"
)

//go:embed assets
var assets embed.FS

// Renderer is a helper type for basic shape rendering which
// reuses vertices and options for slightly reduced memory usage.
type Renderer struct {
	vertices []ebiten.Vertex
	indices  []uint16
	opts     ebiten.DrawTrianglesShaderOptions

	tint          float32
	singleClr     bool
	strokeIndices []uint16

	temps           []offscreen
	blueNoise64RGB  *ebiten.Image
	vogelPoints     [128]float32
	vogelSinCos     [][2]float64
	vogelLastRadius float64
	vogelStickyN    int

	// Warnings registers events like invalid parameters being sent to
	// rendering operations and makes them easy to detect, log and fix.
	Warnings Warnings
}

func NewRenderer() *Renderer {
	var renderer Renderer
	renderer.vertices = make([]ebiten.Vertex, 4)
	renderer.SetColor(color.RGBA{255, 255, 255, 255})
	renderer.indices = []uint16{0, 1, 2, 0, 2, 3}
	renderer.opts.Uniforms = make(map[string]any, 8)
	renderer.strokeIndices = []uint16{
		0, 1, 4,
		4, 1, 5,
		5, 1, 2,
		5, 2, 6,
		6, 2, 3,
		6, 3, 7,
		7, 3, 0,
		0, 4, 7,
	}
	return &renderer
}

func (r *Renderer) GetColorF32() [4]float32 {
	return [4]float32{r.vertices[0].ColorR, r.vertices[0].ColorG, r.vertices[0].ColorB, r.vertices[0].ColorA}
}

// SetColor sets the color of all vertices, unless vertexIndices are specifically provided, in
// which case only the given indices will be set. In general, most shaders use vertex 0 as top-left,
// vertex 1 as top-right, vertex 2 as bottom-right, vertex 3 as bottom-left, but this is shader
// dependent (or even variable in some cases).
func (r *Renderer) SetColor(clr color.Color, vertexIndices ...int) {
	clrF32 := RGBAF32(clr)
	r.SetColorF32(clrF32[0], clrF32[1], clrF32[2], clrF32[3], vertexIndices...)
}

func (r *Renderer) SetColorF32(red, green, blue, alpha float32, vertexIndices ...int) {
	if len(vertexIndices) == 0 {
		r.singleClr = true
		vertexIndices = []int{0, 1, 2, 3}
	} else {
		r.singleClr = false
	}
	for _, i := range vertexIndices {
		r.vertices[i].ColorR = red
		r.vertices[i].ColorG = green
		r.vertices[i].ColorB = blue
		r.vertices[i].ColorA = alpha
	}
}

func (r *Renderer) ScaleAlphaBy(alphaFactor float32) {
	for i := range r.vertices {
		r.vertices[i].ColorR *= alphaFactor
		r.vertices[i].ColorG *= alphaFactor
		r.vertices[i].ColorB *= alphaFactor
		r.vertices[i].ColorA *= alphaFactor
	}
}

// SetTint sets the renderer's tint value, which controls the weight
// of the renderer vertex colors on documented operations:
//   - A value of 0 means only source image colors are used. This is the default.
//   - A value of 1 means only renderer colors are used.
//   - Other values between 0..1 can be used to control the mix.
//
// Operations that can use tint are documented with "This operation is
// affected by [Renderer.Tint]".
//
// Most workflows should only set tint locally during a sequence of operations and
// restore it to zero afterwards.
func (r *Renderer) SetTint(rendererColorWeight float32) {
	if rendererColorWeight != r.tint {
		if rendererColorWeight < 0 || rendererColorWeight > 1 {
			r.Warnings.report(WarnInvalidTintClamped, rendererColorWeight)
			rendererColorWeight = clamp(rendererColorWeight, 0, 1)
		}
		r.tint = rendererColorWeight
	}
}

// Tint returns the renderer's current tint value, which controls
// the weight of the renderer vertex colors on documented operations.
//
// See [Renderer.SetTint]() for more details.
func (r *Renderer) Tint() float32 {
	return r.tint
}

func (r *Renderer) Options() *ebiten.DrawTrianglesShaderOptions {
	return &r.opts
}

func (r *Renderer) DrawShaderAt(target, source *ebiten.Image, ox, oy, horzMargin, vertMargin float32, shader *ebiten.Shader) {
	srcOX, srcOY, srcWidthF32, srcHeightF32 := rectOriginSizeF32(source.Bounds())
	r.setDstRectCoords(ox-horzMargin, oy-vertMargin, ox+srcWidthF32+horzMargin, oy+srcHeightF32+vertMargin)
	r.setSrcRectCoords(srcOX-horzMargin, srcOY-vertMargin, srcOX+srcWidthF32+horzMargin, srcOY+srcHeightF32+vertMargin)

	r.opts.Images[0] = source
	target.DrawTrianglesShader(r.vertices[:], r.indices[:], shader, &r.opts)
	r.opts.Images[0] = nil
}

func (r *Renderer) DrawRectShader(target *ebiten.Image, ox, oy, w, h, horzMargin, vertMargin float32, shader *ebiten.Shader) {
	r.setDstRectCoords(ox-horzMargin, oy-vertMargin, ox+w+horzMargin, oy+h+vertMargin)
	r.setSrcRectCoords(-horzMargin, -vertMargin, w+horzMargin, h+vertMargin)
	target.DrawTrianglesShader(r.vertices[:], r.indices[:], shader, &r.opts)
}

func (r *Renderer) DrawShader(target *ebiten.Image, horzMargin, vertMargin float32, shader *ebiten.Shader) {
	bounds := target.Bounds()
	r.DrawRectShader(target, 0, 0, float32(bounds.Dx()), float32(bounds.Dy()), horzMargin, vertMargin, shader)
}

// DrawAt draws a source image with the given parameters. Supported flags: [Bilinear], [Dithered].
//
// This operation is affected by [Renderer.Tint].
//
// At first glance DrawAt and ebiten.Image.DrawImage overlap heavily, but DrawAt exposes
// two features that DrawImageOptions doesn't have:
//   - Applying dithering during composition, which can be especially critical at low alphas
//     and visibility transitions when image and background colors are similar.
//   - Painting or interpolating an image with the renderer colors, which can be set per vertex.
//     This is color mixing, as opposed to ebiten.ColorScale's multiplication.
//
// Usage example:
//
//	r.SetTint(1.0) // use only renderer colors
//	r.SetColorF32(1.0, 0.0, 0.0, 1.0, 0, 1) // make top vertices red
//	r.SetColorF32(1.0, 0.0, 1.0, 1.0, 2, 3) // make bottom vertices magenta
//	x, y = shapes.CTR.Adjust(src, x, y) // adjust coordinates to CTR (center)
//	r.DrawAt(target, source, x, y, alpha, shapes.Dithered) // draw with dithering
func (r *Renderer) DrawAt(target *ebiten.Image, source *ebiten.Image, x, y float32, alpha float32, flags ...Flag) {
	if alpha < 0.0 || alpha > 1.0 {
		r.Warnings.report(WarnInvalidAlphaClamped, alpha)
		alpha = clamp(alpha, 0, 1)
	}
	if alpha == 0 && r.hasSkippableBlend() {
		return
	}

	var dither, bilinear bool
	for _, flag := range flags {
		switch flag {
		case noFlag:
			// ignore
		case Dithered:
			dither = true
		case Bilinear:
			bilinear = true
		default:
			r.Warnings.report(WarnInvalidFlag, flag)
		}
	}

	if dither {
		r.ensureBlueNoiseLoaded()
		r.opts.Uniforms["Dither"] = 1
		r.opts.Images[1] = r.blueNoise64RGB
	}

	if bilinear {
		r.setFlatCustomVAs01(r.tint, alpha)
		r.DrawShaderAt(target, source, x, y, 0, 0, shaderDrawTintBilinear.Load())
	} else {
		r.setFlatCustomVAs01(r.tint, alpha)
		r.DrawShaderAt(target, source, x, y, 0, 0, shaderDrawTintNearest.Load())
	}

	if dither {
		clear(r.opts.Uniforms)
		r.opts.Images[1] = nil
	}
}

// ScaleOptions are used in [Renderer.Scale]().
type ScaleOptions struct {
	// When true, samples outside bounds will be clamped to the image limits.
	// When false, samples outside bounds will be considered transparent (0, 0, 0, 0).
	Clamp bool

	// When true, sampling is done in destination space (1.0/scale px offsets).
	// When false, sampling is done in source space (1.0px offsets).
	//
	// DstSampling = true is equivalent to Ebitengine's v2.9.0 FilterPixelated.
	DstSampling bool

	// When true, bicubic algorithm will be used instead of bilinear.
	Bicubic bool
}

// Scale draws the source into the given target with the given parameters. opts can be nil,
// but notice that in that case Ebitengine default functions can already do the job fine.
func (r *Renderer) Scale(target, source *ebiten.Image, ox, oy, scale float32, opts *ScaleOptions) {
	if scale <= 0 {
		if scale < 0 {
			r.Warnings.report(WarnNegativeValueOpSkipped, scale)
		}
		return
	}

	srcBounds := source.Bounds()
	srcWidth, srcHeight := srcBounds.Dx(), srcBounds.Dy()
	srcWidthF32, srcHeightF32 := float32(srcWidth), float32(srcHeight)

	dstBounds := target.Bounds()
	minX := float32(dstBounds.Min.X) + ox
	minY := float32(dstBounds.Min.Y) + oy
	r.setDstRectCoords(minX, minY, minX+srcWidthF32*scale, minY+srcHeightF32*scale)

	minX, minY = float32(srcBounds.Min.X), float32(srcBounds.Min.Y)
	r.setSrcRectCoords(minX, minY, minX+srcWidthF32, minY+srcHeightF32)
	r.opts.Images[0] = source

	var clamp, sampling float32 = 0.0, 1.0
	var shader *ebiten.Shader
	if opts != nil {
		if opts.Clamp {
			clamp = 1.0
		}
		if opts.DstSampling {
			sampling /= scale
		}
		if opts.Bicubic {
			shader = shaderBicubic.Load()
		}
	}
	if shader == nil {
		shader = shaderBilinear.Load()
	}

	r.setFlatCustomVAs(sampling, sampling, clamp, 0)
	target.DrawTrianglesShader(r.vertices[:], r.indices[:], shader, &r.opts)
	r.opts.Images[0] = nil
}

// UnsafeTemp allows requesting offscreens to the renderer. These offscreens might have
// already been created while the renderer was doing complex operations, so reusing them
// can prevent the creation of additional offscreens.
//
// The offscreens returned by this function should only be used for local operations, and
// the offscreen must not be stored. Any renderer function documented to use an internal
// offscreen can panic or fail in any other way if an offscreen returned by this function
// is passed as an input parameter.
func (r *Renderer) UnsafeTemp(offscreenIndex int, w, h int, clear bool) *ebiten.Image {
	temp, _ := r.getTemp(offscreenIndex, w, h, clear)
	return temp
}

// UnsafeTempCopy calls [Renderer.UnsafeTemp]() and copies the contents of source into
// the returned offscreen. See safety warnings and docs for UnsafeTemp. The 'clear'
// argument allows specifying whether a 1 pixel clear margin is required or not.
func (r *Renderer) UnsafeTempCopy(offscreenIndex int, source *ebiten.Image, clear bool) *ebiten.Image {
	bounds := source.Bounds()
	temp, _ := r.getTemp(offscreenIndex, bounds.Dx(), bounds.Dy(), clear)
	var opts ebiten.DrawImageOptions
	opts.Blend = ebiten.BlendCopy
	temp.DrawImage(source, &opts)
	return temp
}

// UnsafeTempDual calls [Renderer.UnsafeTemp](), copies the contents of source into the
// specified offscreen, and returns both this copy and an extra image of the same size
// on the same offscreen. This function is highly specific and meant to prepare images
// for shaders that use two source images: an original source and variant or mask for it.
//
// If any padding is given, the padding is always cleared.
//
// See safety warnings and docs for UnsafeTemp.
func (r *Renderer) UnsafeTempDual(offscreenIndex int, source *ebiten.Image, padding int, clear bool) (sourceTemp, variantTemp *ebiten.Image) {
	if padding < 0 {
		r.Warnings.report(WarnNegativeValueZeroed, padding)
	}

	_, _, w, h := rectOriginSize(source.Bounds())
	ox, oy := 0, 0
	pw, ph := w+padding*2, h+padding*2
	if h <= w {
		oy = ph
	} else {
		ox = pw
	}
	temp, clear := r.getTemp(offscreenIndex, ox+pw, oy+ph, clear)
	if padding > 0 && !clear {
		memoBlend := r.opts.Blend
		r.opts.Blend = ebiten.BlendClear
		r.StrokeIntArea(temp, 0, 0, pw, ph, 0, padding)
		r.opts.Blend = memoBlend
	}
	var opts ebiten.DrawImageOptions
	opts.Blend = ebiten.BlendCopy
	if padding > 0 {
		opts.GeoM.Translate(float64(padding), float64(padding))
	}
	temp.DrawImage(source, &opts)

	sourceTemp = temp.SubImage(image.Rect(0, 0, pw, ph)).(*ebiten.Image)
	variantTemp = temp.SubImage(image.Rect(ox, oy, ox+pw, oy+ph)).(*ebiten.Image)
	return sourceTemp, variantTemp
}

func (r *Renderer) setDstRectCoords(minX, minY, maxX, maxY float32) {
	setVertDstCoords(r.vertices, minX, minY, maxX, maxY)
}

func (r *Renderer) setSrcRectCoords(minX, minY, maxX, maxY float32) {
	setVertSrcCoords(r.vertices, minX, minY, maxX, maxY)
}

func (r *Renderer) setFlatCustomVAs(cva0, cva1, cva2, cva3 float32) {
	for i := range len(r.vertices) {
		r.vertices[i].Custom0 = cva0
		r.vertices[i].Custom1 = cva1
		r.vertices[i].Custom2 = cva2
		r.vertices[i].Custom3 = cva3
	}
}

func (r *Renderer) setFlatCustomVA0(cva0 float32) {
	for i := range len(r.vertices) {
		r.vertices[i].Custom0 = cva0
	}
}

func (r *Renderer) setFlatCustomVAs01(cva0, cva1 float32) {
	for i := range len(r.vertices) {
		r.vertices[i].Custom0 = cva0
		r.vertices[i].Custom1 = cva1
	}
}

// SetCustomVAs configures up to 4 custom vertex attributes.
func (r *Renderer) SetCustomVAs(vas ...float32) {
	switch len(vas) {
	case 0:
		// nothing
	case 1:
		r.setFlatCustomVA0(vas[0])
	case 2:
		r.setFlatCustomVAs01(vas[0], vas[1])
	case 3:
		r.setFlatCustomVAs(vas[0], vas[1], vas[2], 0.0)
	default:
		if len(vas) > 4 {
			r.Warnings.report(WarnTooManyVertexAttribs, len(vas))
		}
		r.setFlatCustomVAs(vas[0], vas[1], vas[2], vas[3])
	}
}

// the returned bool indicates whether the returned offscreen is clear
// (this will always be true if clear was requested, but can be true in
// some cases even if clear is not requested)
func (r *Renderer) getTemp(offscreenIndex int, w, h int, clear bool) (*ebiten.Image, bool) {
	if offscreenIndex >= len(r.temps) {
		growth := offscreenIndex + 1 - len(r.temps)
		r.temps = slices.Grow(r.temps, growth)
		r.temps = r.temps[:offscreenIndex+1]
		r.temps[offscreenIndex] = newOffscreen(0, 0, 64)
	}
	return r.temps[offscreenIndex].WithSize(w, h, clear)
}

func (r *Renderer) hasSkippableBlend() bool {
	// ebiten.BlendLighter should also be skippable, but it's so rare
	// it's probably not be worth it. maybe also shapes.BlendSubtract
	return r.opts.Blend == ebiten.BlendSourceOver
}
