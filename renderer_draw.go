package shapes

import (
	"math"
	"slices"

	"github.com/hajimehoshi/ebiten/v2"
)

// DrawAt draws a source image with the given parameters. Supported flags: [Bilinear], [Dithered].
//
// This operation is affected by [Renderer.Tint].
//
// At first glance DrawAt might seem redundant with [ebiten.DrawImageOptions], but there are two
// differential features:
//   - Applying dithering during composition, which can be especially critical at low alphas
//     and visibility transitions when image and background colors are similar.
//   - Painting or interpolating an image with the renderer colors, which can be set per vertex.
//     This is color mixing, as opposed to [ebiten.ColorScale]'s multiplication.
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
	if alpha == 0 && r.blendSafeToCrop() {
		return
	}

	bilinear, dither := r.readFlags(flags...)
	if dither {
		r.opts.Uniforms["Dither"] = 1
		r.loadBlueNoise64RGBAt(1)
	}

	var shader *ebiten.Shader
	if bilinear {
		shader = shaderDrawTintBilinear.Load()
	} else {
		shader = shaderDrawTintNearest.Load()
	}

	r.setFlatCustomVAs01(r.tint, alpha)
	r.DrawImgShader(target, source, x, y, NoMargins, shader)

	if dither {
		clear(r.opts.Uniforms)
		r.opts.Images[1] = nil
	}
}

// DrawImgShader calls DrawTrianglesShader using the renderer's option and
// passing the given source image as imageSrc0.
//
// Notice that the target origin matters; to align the shader to the top left
// corner, (ox, oy) must match target.Bounds().Min.
//
// This is a low level method mostly used by other higher level renderer calls.
func (r *Renderer) DrawImgShader(target, source *ebiten.Image, ox, oy float32, margins Margins, shader *ebiten.Shader) {
	srcOX, srcOY, srcWidthF32, srcHeightF32 := rectOriginSizeF32(source.Bounds())
	r.setDstRectCoords(ox-margins.Left, oy-margins.Top, ox+srcWidthF32+margins.Right, oy+srcHeightF32+margins.Bottom)
	r.setSrcRectCoords(srcOX-margins.Left, srcOY-margins.Top, srcOX+srcWidthF32+margins.Right, srcOY+srcHeightF32+margins.Bottom)

	r.opts.Images[0] = source
	target.DrawTrianglesShader32(r.vertices[:], r.indices[:], shader, &r.opts)
	r.opts.Images[0] = nil
}

// DrawRectShader calls target.DrawTrianglesShader using the renderer's options.
//
// Notice that the target origin matters; to align the shader to the top left
// corner, (ox, oy) must match target.Bounds().Min.
//
// This is a low level method mostly used by other higher level renderer calls.
func (r *Renderer) DrawRectShader(target *ebiten.Image, ox, oy, w, h float32, margins Margins, shader *ebiten.Shader) {
	r.setDstRectCoords(ox-margins.Left, oy-margins.Top, ox+w+margins.Right, oy+h+margins.Bottom)
	r.setSrcRectCoords(-margins.Left, -margins.Top, w+margins.Right, h+margins.Bottom)
	target.DrawTrianglesShader32(r.vertices[:], r.indices[:], shader, &r.opts)
}

// CircShaderOptions are used for [Renderer.DrawCircShader]().
type CircShaderOptions struct {
	// Radius defines the radius of the circular region.
	Radius float32

	// Thickness defines the width of the circular outline:
	//  - If thickness > 0, outline expands [-thickness/2, thickness/2]
	//    around the radius.
	//  - If thickness < 0, the outline goes from [-thickness, 0].
	Thickness float32

	// See [RadsRight] constants for angle conventions and docs.
	StartAngle float32

	// See [RadsRight] constants for angle conventions and docs.
	EndAngle float32

	// Tolerance sets the maximum allowed tessellation overshoot
	// in pixels.
	//
	// If zero, a default coarse tolerance of 7.5 is used.
	// The minimum non-zero value is 0.1.
	Tolerance float32
}

// CircShaderOpts creates a full circle [CircShaderOptions] with the given radius
// and thickness.
func CircShaderOpts(radius, thickness float32) CircShaderOptions {
	return CircShaderOptions{
		Radius:     radius,
		Thickness:  thickness,
		StartAngle: 0.0,
		EndAngle:   2 * math.Pi,
	}
}

// DrawCircShader draws a shader within a circular triangle strip. This is particularly
// useful when the bounding rectangle of a circle is much bigger than the actual area
// that we need to render.
//
// Due to the dynamically adjusted number of segments and rendering area, notice that:
//   - You can't use per-vertex colors, only the first vertex color will be used.
//   - You shouldn't rely on blends like BlendSourceIn or BlendClear that aren't "safe
//     to crop".
func (r *Renderer) DrawCircShader(target *ebiten.Image, cx, cy float32, opts CircShaderOptions, shader *ebiten.Shader) {
	if opts.Thickness == 0 {
		return // ignore blends, this is already a crop operation
	}
	if opts.Tolerance == 0.0 {
		opts.Tolerance = 7.5 // default tolerance, update opts.Tolerance docs if changed
	}
	if opts.Radius < 0.0 {
		r.Warnings.report(WarnRadiusClamped, opts.Radius)
		opts.Radius = 0.0
	}
	if opts.Tolerance < 0.1 {
		r.Warnings.report(WarnLowToleranceRaised, opts.Tolerance)
		opts.Tolerance = 0.1
	}
	if opts.StartAngle == opts.EndAngle {
		return
	}

	// prepare vertices and indices
	memo := r.memorizeColors()
	r.vertices = r.vertices[:0]
	rads := uradsDeltaCW(opts.StartAngle, opts.EndAngle)
	r.vertices = appendArcVertices(r.vertices, float64(opts.Radius), float64(opts.Thickness), float64(opts.StartAngle), float64(rads), float64(opts.Tolerance))

	for i := range r.vertices {
		r.vertices[i].DstX += cx
		r.vertices[i].DstY += cy
		r.vertices[i].ColorR = memo[0]
		r.vertices[i].ColorG = memo[1]
		r.vertices[i].ColorB = memo[2]
		r.vertices[i].ColorA = memo[3]
	}

	r.indices = r.indices[:0]
	numQuads := len(r.vertices)/2 - 1
	numIndices := numQuads * 6
	r.indices = slices.Grow(r.indices, numIndices)[:numIndices]
	i := 0
	for q := range uint32(numQuads) {
		vertIndex := q << 1
		r.indices[i+0] = vertIndex + 0
		r.indices[i+1] = vertIndex + 1
		r.indices[i+2] = vertIndex + 2
		r.indices[i+3] = vertIndex + 2
		r.indices[i+4] = vertIndex + 1
		r.indices[i+5] = vertIndex + 3
		i += 6
	}

	// draw and restore state
	target.DrawTrianglesShader32(r.vertices[:], r.indices[:], shader, &r.opts)

	r.vertices = r.vertices[:4]
	r.restoreIndices()
	r.setColors(memo)
}
