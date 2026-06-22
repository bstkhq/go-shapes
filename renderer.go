package shapes

import (
	"image"
	"image/color"
	"slices"

	"github.com/hajimehoshi/ebiten/v2"
)

// Renderer is the heart of the go-shapes package and provides access to most of its
// operations. It stores offscreens, vertices and other data reused across rendering
// methods.
//
// Valid renderers must be created through [NewRenderer](). Once created, most
// users should consider setting a logger for warnings, e.g.:
//
//	renderer.Warnings.SetHandler(NewWarningLogOnceHandler()).
//
// Unless stated otherwise, renderer state management should respect the following
// conventions:
//
//   - Renderer color should not be assumed; set it explicitly before operation.
//   - If [Renderer.Options]() and [Renderer.Tint]() are modified during operation,
//     they should be restored afterwards. Tint is assumed to be zero, and the only
//     common modification to the renderer options is setting the blend, which is
//     expected to be [ebiten.BlendSourceOver].
type Renderer struct {
	vertices []ebiten.Vertex
	indices  []uint32
	opts     ebiten.DrawTrianglesShaderOptions

	lastBlend           ebiten.Blend
	lastBlendSafeToCrop bool

	singleClr bool
	tint      float32 // mix rate for renderer colors in supported operations

	temps     []offscreen
	vogelMemo *vogelMemory // helper used by vogel blur

	// Warnings registers events like invalid parameters being sent to
	// rendering operations and makes them easy to detect, log and fix.
	Warnings Warnings
}

// NewRenderer initializes and returns a new [Renderer].
func NewRenderer() *Renderer {
	var renderer Renderer
	renderer.vertices = make([]ebiten.Vertex, 4)
	renderer.SetColor(color.RGBA{255, 255, 255, 255})
	renderer.indices = []uint32{0, 1, 2, 0, 2, 3}
	renderer.lastBlendSafeToCrop = true
	renderer.lastBlend = ebiten.BlendSourceOver
	renderer.opts.Blend = ebiten.BlendSourceOver
	renderer.opts.Uniforms = make(map[string]any, 8)
	return &renderer
}

func (r *Renderer) restoreIndices() {
	r.indices = r.indices[:6]
	r.indices[0] = 0
	r.indices[1] = 1
	r.indices[2] = 2
	r.indices[3] = 0
	r.indices[4] = 2
	r.indices[5] = 3
}

// GetColorF32 returns the current color of the requested vertex
// (0 = top-left, 1 = top-right, 2 = bottom-right, 3 = bottom-left).
func (r *Renderer) GetColorF32(vertexIndex int) [4]float32 {
	return [4]float32{r.vertices[vertexIndex].ColorR, r.vertices[vertexIndex].ColorG, r.vertices[vertexIndex].ColorB, r.vertices[vertexIndex].ColorA}
}

// SetColor sets the color of the renderer vertices. If vertexIndices are not provided,
// all 4 vertex colors are set; otherwise, only the specified vertices will have the
// color applied.
//
// Vertex indices start at top-left and follow in clockwise order: 0 = top-left, 1 = top-right,
// 2 = bottom-right, 3 = bottom-left.
//
// Colors are internally converted to float32. See also [Renderer.SetColorF32]().
func (r *Renderer) SetColor(clr color.Color, vertexIndices ...int) {
	clrF32 := RGBAF32(clr)
	r.SetColorF32(clrF32[0], clrF32[1], clrF32[2], clrF32[3], vertexIndices...)
}

// SetColorF32 is the float32 version of [Renderer.SetColor](). Passing colors directly
// as float32 values avoids conversions.
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

// SetColorF32A is the [4]float32 version of [Renderer.SetColorF32]().
func (r *Renderer) SetColorF32A(color [4]float32, vertexIndices ...int) {
	r.SetColorF32(color[0], color[1], color[2], color[3], vertexIndices...)
}

// ScaleAlphaBy adjusts the vertex colors by multiplying them by alphaFactor.
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

// Options returns the underlying shader options.
func (r *Renderer) Options() *ebiten.DrawTrianglesShaderOptions {
	return &r.opts
}

func (r *Renderer) memorizeColors() [16]float32 {
	var memo [16]float32
	for i := range 4 {
		memo[i<<2+0] = r.vertices[i].ColorR
		memo[i<<2+1] = r.vertices[i].ColorG
		memo[i<<2+2] = r.vertices[i].ColorB
		memo[i<<2+3] = r.vertices[i].ColorA
	}
	return memo
}

func (r *Renderer) vertexColors(i int) [4]float32 {
	return [4]float32{r.vertices[i].ColorR, r.vertices[i].ColorG, r.vertices[i].ColorB, r.vertices[i].ColorA}
}

// notice: internal use only, doesn't touch singleClr flag
func (r *Renderer) restoreColors(values [16]float32) {
	for i := range 4 {
		r.vertices[i].ColorR = values[i<<2+0]
		r.vertices[i].ColorG = values[i<<2+1]
		r.vertices[i].ColorB = values[i<<2+2]
		r.vertices[i].ColorA = values[i<<2+3]
	}
}

// set all the vertex colors based on interpolation over a quad region and the
// original quad colors
func (r *Renderer) applyTriQuadColors(minX, minY, maxX, maxY float32, baseColors [16]float32) {
	origin := PtF32(minX, minY)
	size := PtF32(maxX-minX, maxY-minY)

	tl := [4]float32(baseColors[0:4])
	tr := [4]float32(baseColors[4:8])
	br := [4]float32(baseColors[8:12])
	bl := [4]float32(baseColors[12:16])
	for i := range r.vertices {
		interpCoords := PtF32(r.vertices[i].DstX, r.vertices[i].DstY)
		clr := interpTriQuadColor(tl, tr, br, bl, origin, size, interpCoords)
		setVertexColor(&r.vertices[i], clr[0], clr[1], clr[2], clr[3])
	}
}

func (r *Renderer) applySingleColor(cr, cg, cb, ca float32) {
	for i := range r.vertices {
		setVertexColor(&r.vertices[i], cr, cg, cb, ca)
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

// Scale draws source into target with the specified parameters. opts can be nil,
// but in that case there's no advantage over Ebitengine's default functions.
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

	r.setDstRectCoords(ox, oy, ox+srcWidthF32*scale, oy+srcHeightF32*scale)
	minX, minY := float32(srcBounds.Min.X), float32(srcBounds.Min.Y)
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
	target.DrawTrianglesShader32(r.vertices[:], r.indices[:], shader, &r.opts)
	r.opts.Images[0] = nil
}

// UnsafeTemp allows requesting offscreens to the renderer. These offscreens might have
// already been created while the renderer was doing complex operations, so reusing them
// can prevent the creation of additional offscreens.
//
// The 'clear' argument allows specifying whether the image should be cleared or not
// (this includes an extra 1px transparent margin if the returned offscreen is part of
// a larger image).
//
// Returned offscreens always have origin (0, 0).
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
//
// If padding > 0 is given, the padding is always cleared and output images have width
// and height 2*padding larger than source.
//
// Returned offscreens always have origin (0, 0).
func (r *Renderer) UnsafeTempCopy(offscreenIndex int, source *ebiten.Image, padding int, clear bool) *ebiten.Image {
	padding = warnZeroNegativeValue(r, padding)
	ow, oh := rectSize(source.Bounds())
	ow, oh = ow+padding*2, oh+padding*2
	temp, clear := r.getTemp(offscreenIndex, ow, oh, clear)
	if padding > 0 && !clear {
		Clear(temp, image.Rect(0, 0, ow, padding))
		Clear(temp, image.Rect(0, oh-padding, ow, oh))
		Clear(temp, image.Rect(0, padding, padding, oh-padding))
		Clear(temp, image.Rect(ow-padding, padding, ow, oh-padding))
	}

	var opts ebiten.DrawImageOptions
	opts.Blend = ebiten.BlendCopy
	opts.GeoM.Translate(float64(padding), float64(padding))
	temp.DrawImage(source, &opts)
	return temp
}

// UnsafeTempDual calls [Renderer.UnsafeTemp](), copies the contents of source into the
// specified offscreen, and returns both this copy and an extra image of the same size
// on the same offscreen. This function is highly specific and meant to prepare images
// for shaders that use two source images: an original source and variant or mask for it.
//
// If padding > 0 is given, the padding is always cleared and output images have width
// and height 2*padding larger than source.
//
// See safety warnings and docs for UnsafeTemp.
func (r *Renderer) UnsafeTempDual(offscreenIndex int, source *ebiten.Image, padding int, clear bool) (sourceTemp, variantTemp *ebiten.Image) {
	padding = warnZeroNegativeValue(r, padding)
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
		Clear(temp, image.Rect(0, 0, pw, padding))
		Clear(temp, image.Rect(0, ph-padding, pw, ph))
		Clear(temp, image.Rect(0, padding, padding, ph-padding))
		Clear(temp, image.Rect(pw-padding, padding, pw, ph-padding))
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

func (r *Renderer) blendSafeToCrop() bool {
	if r.lastBlend != r.opts.Blend {
		r.lastBlendSafeToCrop = blendOnlyAffectsNonZeroSourceArea(r.opts.Blend)
		r.lastBlend = r.opts.Blend
	}
	return r.lastBlendSafeToCrop
}

func blendOnlyAffectsNonZeroSourceArea(blend ebiten.Blend) bool {
	return blend == ebiten.BlendSourceOver ||
		blend == ebiten.BlendDestinationOver ||
		blend == ebiten.BlendLighter ||
		blend == ebiten.BlendSourceAtop ||
		blend == ebiten.BlendDestinationOut ||
		blend == ebiten.BlendXor ||
		blend == BlendSubtract
}

// read Bilinear and Dithered flags
func (r *Renderer) readRenderFlags(flags ...Flag) (bilinear, dither bool) {
	for _, flag := range flags {
		switch flag {
		case Bilinear:
			if bilinear {
				r.Warnings.report(WarnRepeatedFlag, flag)
			}
			bilinear = true
		case Dithered:
			if dither {
				r.Warnings.report(WarnRepeatedFlag, flag)
			}
			dither = true
		default:
			r.Warnings.report(WarnInvalidFlag, flag)
		}
	}

	return bilinear, dither
}

// read optional bounding and colorMode flags. if no optional values are
// provided, the return values will be noFlag and a suitable fallback should
// be executed
func (r *Renderer) readBoundingAndColorModeFlags(optInBounding, optInColorMode Flag, flags ...Flag) (bounding Flag, colorMode Flag) {
	bounding, colorMode = noFlag, noFlag
	for _, flag := range flags {
		if flag == optInBounding {
			if bounding != noFlag {
				r.Warnings.report(WarnRepeatedFlag, flag)
			}
			bounding = optInBounding
		} else if flag == optInColorMode {
			if colorMode != noFlag {
				r.Warnings.report(WarnRepeatedFlag, flag)
			}
			colorMode = optInColorMode
		} else {
			r.Warnings.report(WarnInvalidFlag, flag)
		}
	}
	return bounding, colorMode
}

func (r *Renderer) readOptInFlag(optInFlag Flag, flags ...Flag) Flag {
	flag := noFlag
	for _, flag := range flags {
		if flag == optInFlag {
			if flag != noFlag {
				r.Warnings.report(WarnRepeatedFlag, flag)
			}
			flag = optInFlag
		} else {
			r.Warnings.report(WarnInvalidFlag, flag)
		}
	}
	return flag
}
