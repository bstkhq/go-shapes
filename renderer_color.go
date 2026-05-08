package shapes

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
)

// OklabShift draws the source image to the target, at the given coordinates,
// with the given LCh shifts applied on Oklab color space. Shifts can be positive
// or negative, with the typical ranges being common:
//   - chroma: [0, 0.37]. Most clearly distinguishable colors don't even exceed 0.15.
//   - lightness: [0, 1]. High chromas happen around 0.9 lightness for yellow, around
//     0.8 for orange, cyan, pink and lime green, 0.7 for light blues, magentas, greens
//     and light reds, 0.5 for deeper reds, green, purple and blue, and down to 0.3 for
//     deep blue. Lightness below or above these values in the given hue ranges leads
//     to very dark or whitish colors with significantly lower chromas.
//   - hue: in radians, wrapping is done automatically
//
// Notice: absolute shifts in perceptually uniform color spaces aren't particularly
// helpful. Hue itself is not perceptually uniform, so that's one of the most useful
// values to tweak. Other more effective tools might be exposed in the future.
func (r *Renderer) OklabShift(target, source *ebiten.Image, x, y, lightnessShift, chromaShift, hueShift float32) {
	r.setFlatCustomVAs(lightnessShift, chromaShift, hueShift, 0.0)
	r.DrawImgShader(target, source, x, y, NoMargins, shaderOklabShift.Load())
}

// ColorizeByLightness draws source into target at the given (x, y), taking the
// lightness of each source pixel and remapping it to a color between 'from' and 'to'.
//
// Key parameters:
//   - fromLightness: pixels before this threshold take 'from' color.
//     Expected range: [0.0, 1.0]
//   - toLightness: pixels after this threshold take 'to' color.
//     Expected range: [0.0, 1.0].
func (r *Renderer) ColorizeByLightness(target, source *ebiten.Image, opts GradientOptions, x, y, fromLightness, toLightness float32) {
	if fromLightness > 1.0 || fromLightness < 0.0 {
		r.Warnings.report(WarnInvalidRateClamped, fromLightness)
		fromLightness = clamp(fromLightness, 0.0, 1.0)
	}
	if toLightness > 1.0 || toLightness < 0.0 {
		r.Warnings.report(WarnInvalidRateClamped, toLightness)
		toLightness = clamp(fromLightness, 0.0, 1.0)
	}
	if opts.Bias < -1.0 || opts.Bias > 1.0 {
		r.Warnings.report(WarnInvalidBiasClamped, opts.Bias)
		opts.Bias = clamp(opts.Bias, -1.0, 1.0)
	}
	if opts.Dither {
		r.opts.Uniforms["Dither"] = 1
		r.loadBlueNoise64RGBAt(1)
	}

	fromL, fromA, fromB := toOklab(opts.From[0], opts.From[1], opts.From[2])
	toL, toA, toB := toOklab(opts.To[0], opts.To[1], opts.To[2])

	from := [4]float32{float32(fromL), float32(fromA), float32(fromB), float32(opts.From[3])}
	to := [4]float32{float32(toL), float32(toA), float32(toB), float32(opts.To[3])}
	if fromLightness > toLightness {
		fromLightness, toLightness = toLightness, fromLightness
		from, to = to, from
	}
	r.opts.Uniforms["From"] = from
	r.opts.Uniforms["To"] = to
	r.setFlatCustomVAs(fromLightness, toLightness, float32(max(opts.Steps, 0)), (opts.Bias+1.0)/2.0)
	r.DrawImgShader(target, source, x, y, NoMargins, shaderColorizeByLightness.Load())
	clear(r.opts.Uniforms)
	if opts.Dither {
		r.opts.Images[1] = nil
	}
}

// ColorMix draws 'base' and 'over' to 'target' using the mix() function for color
// mixing instead of standard composition operations.
//
// This is the cleanest way to interpolate a transition between two images (morphing)
// while there's also an alpha transition, or the two images have different alphas at
// different pixel positions.
//
// The sizes of 'base' and 'over' must match.
func (r *Renderer) ColorMix(target, base, over *ebiten.Image, x, y float32, alpha, mixLevel float32, flags ...Flag) {
	baseBounds, overBounds := base.Bounds(), over.Bounds()
	if baseBounds.Dx() != overBounds.Dx() || baseBounds.Dy() != overBounds.Dy() {
		panic(fmt.Sprintf(
			"'base' and 'over' sizes must match (found %dx%d vs %dx%d)",
			baseBounds.Dx(), baseBounds.Dy(), overBounds.Dx(), overBounds.Dy(),
		))
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

	r.opts.Images[1] = over
	if dither {
		r.opts.Uniforms["Dither"] = 1
		r.loadBlueNoise64RGBAt(2)
	}

	if bilinear {
		r.setFlatCustomVAs01(alpha, mixLevel)
		r.DrawImgShader(target, base, x, y, NoMargins, shaderColorMixBilinear.Load())
	} else {
		r.setFlatCustomVAs01(alpha, mixLevel)
		r.DrawImgShader(target, base, x, y, NoMargins, shaderColorMix.Load())
	}

	r.opts.Images[1] = nil
	if dither {
		clear(r.opts.Uniforms)
		r.opts.Images[2] = nil
	}
}

// Predefined color palettes for use with [Renderer.DitherMat4].
var (
	PaletteBW []float32 = []float32{
		0.0, 0.0, 0.0, 1.0,
		1.0, 1.0, 1.0, 1.0,
	}
	PaletteBW4 []float32 = []float32{
		0.0, 0.0, 0.0, 1.0,
		0.333, 0.333, 0.333, 1.0,
		0.666, 0.666, 0.666, 1.0,
		1.0, 1.0, 1.0, 1.0,
	}
	PaletteAlpha8 []float32 = []float32{
		0.0, 0.0, 0.0, 0.0,
		1.0 / 7.0, 1.0 / 7.0, 1.0 / 7.0, 1.0 / 7.0,
		2.0 / 7.0, 2.0 / 7.0, 2.0 / 7.0, 2.0 / 7.0,
		3.0 / 7.0, 3.0 / 7.0, 3.0 / 7.0, 3.0 / 7.0,
		4.0 / 7.0, 4.0 / 7.0, 4.0 / 7.0, 4.0 / 7.0,
		5.0 / 7.0, 5.0 / 7.0, 5.0 / 7.0, 5.0 / 7.0,
		6.0 / 7.0, 6.0 / 7.0, 6.0 / 7.0, 6.0 / 7.0,
		1.0, 1.0, 1.0, 1.0,
	}
	PaletteBRG []float32 = []float32{
		0.0, 0.0, 1.0, 1.0,
		1.0, 0.0, 0.0, 1.0,
		0.0, 1.0, 0.0, 1.0,
	}
)

// Predefined 4x4 dither matrices for use with [Renderer.DitherMat4].
// Tip: average multiple matrices for more interesting variations.
var (
	DitherBayes [16]float32 = [16]float32{
		0.0 / 16.0, 12.0 / 16.0, 3.0 / 16.0, 15.0 / 16.0,
		8.0 / 16.0, 4.0 / 16.0, 11.0 / 16.0, 7.0 / 16.0,
		2.0 / 16.0, 14.0 / 16.0, 1.0 / 16.0, 13.0 / 16.0,
		10.0 / 16.0, 6.0 / 16.0, 9.0 / 16.0, 5.0 / 16.0,
	}
	DitherDots [16]float32 = [16]float32{
		12.0 / 16.0, 4.0 / 16.0, 11.0 / 16.0, 15.0 / 16.0,
		5.0 / 16.0, 0.0 / 16.0, 3.0 / 16.0, 10.0 / 16.0,
		6.0 / 16.0, 1.0 / 16.0, 2.0 / 16.0, 9.0 / 16.0,
		13.0 / 16.0, 7.0 / 16.0, 8.0 / 16.0, 14.0 / 16.0,
	}
	DitherSerp [16]float32 = [16]float32{
		0.0 / 16.0, 12.0 / 16.0, 13.0 / 16.0, 1.0 / 16.0,
		3.0 / 16.0, 7.0 / 16.0, 6.0 / 16.0, 2.0 / 16.0,
		4.0 / 16.0, 8.0 / 16.0, 9.0 / 16.0, 5.0 / 16.0,
		11.0 / 16.0, 15.0 / 16.0, 14.0 / 16.0, 10.0 / 16.0,
	}
	DitherGlitch [16]float32 = [16]float32{
		0.0 / 16.0, 1.0 / 16.0, 2.0 / 16.0, 3.0 / 16.0,
		4.0 / 16.0, 5.0 / 16.0, 6.0 / 16.0, 7.0 / 16.0,
		8.0 / 16.0, 9.0 / 16.0, 10.0 / 16.0, 11.0 / 16.0,
		12.0 / 16.0, 13.0 / 16.0, 14.0 / 16.0, 15.0 / 16.0,
	}
	DitherCrumbs [16]float32 = [16]float32{
		0.0 / 16.0, 4.0 / 16.0, 8.0 / 16.0, 1.0 / 16.0,
		11.0 / 16.0, 14.0 / 16.0, 12.0 / 16.0, 5.0 / 16.0,
		7.0 / 16.0, 13.0 / 16.0, 15.0 / 16.0, 9.0 / 16.0,
		3.0 / 16.0, 10.0 / 16.0, 6.0 / 16.0, 2.0 / 16.0,
	}
)

// DitherMat4 draws the given mask to the target applying a static 4x4 dithering pattern to select colors from
// rgbaColors.
//   - The rgbaColors argument can contain up to 8 colors, flattened as RGBA quadruplets in [0...1] range.
//     See [PaletteBW4] and others for predefined palettes.
//   - The ditherMatrix argument is a 4x4 dithering matrix in column major order (like GLSL), where the
//     values indicate the thresholds of the pattern in 0...1 range. See [DitherBayes] and others for
//     predefined matrices.
func (r *Renderer) DitherMat4(target, mask *ebiten.Image, ox, oy float32, xOffset, yOffset int, rgbaColors []float32, ditherMatrix [16]float32, rendererClrMix, maskColorMix float32) {
	if len(rgbaColors)%4 != 0 {
		panic("rgbaColors must have length multiple of 4")
	}
	if len(rgbaColors) > 8*4 {
		r.Warnings.report(WarnTooManyColorsClamped, len(rgbaColors)/4)
		rgbaColors = rgbaColors[:8*4]
	}
	numColors := len(rgbaColors) / 4
	if numColors <= 1 {
		panic("DitherMat4 requires at least 2 colors (as 8 float32 values)")
	}
	// TODO: check alpha premult?
	var palette [4 * 8]float32
	copy(palette[:], rgbaColors)

	r.setFlatCustomVAs(float32(xOffset), float32(yOffset), rendererClrMix, maskColorMix)
	r.opts.Uniforms["Matrix"] = ditherMatrix
	r.opts.Uniforms["NumColors"] = numColors
	r.opts.Uniforms["Colors"] = palette
	r.DrawImgShader(target, mask, ox, oy, NoMargins, shaderDitherMat4.Load())
	clear(r.opts.Uniforms)
}
