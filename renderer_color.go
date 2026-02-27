package shapes

import (
	"fmt"
	"image"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

const MinGradientCurveFactor = 0.001

// NewSimpleGradient returns a new image filled with the given gradient.
//
// For common dirRadians values, consider [DirRadsLTR] and related constants.
func (r *Renderer) NewSimpleGradient(w, h int, from, to color.RGBA, dirRadians float32) *ebiten.Image {
	img := ebiten.NewImage(w, h)
	r.SimpleGradient(img, from, to, dirRadians)
	return img
}

// SimpleGradient paints a gradient over the given target, interpolating in Oklab space.
//
// For common dirRadians values, consider [DirRadsLTR] and related constants.
//
// See [Renderer.Gradient]() for additional controls.
func (r *Renderer) SimpleGradient(target *ebiten.Image, from, to color.RGBA, dirRadians float32) {
	r.Gradient(target, nil, 0, 0, from, to, -1, dirRadians, 1.0)
}

// Gradient paints a gradient over the given target, interpolating in Oklab space.
// If mask is nil, the target will have the gradient applied starting from (ox, oy)
// throughout the entire image.
//   - numSteps: use -1 for continuous gradient, > 0 for discrete color steps.
//   - curveFactor: use 1.0 for linear, or see [MinGradientCurveFactor] for more details.
//   - dirRadians: gradient direction. For common values, consider [DirRadsLTR] and related constants.
//
// See also [Renderer.SimpleGradient](), [Renderer.GradientDither]() and [Renderer.GradientRadial]().
func (r *Renderer) Gradient(target, mask *ebiten.Image, ox, oy float32, from, to color.RGBA, numSteps int, dirRadians, curveFactor float32) {
	if curveFactor < MinGradientCurveFactor {
		r.Warnings.report(WarnGradientCurveFactorLifted, curveFactor)
		curveFactor = MinGradientCurveFactor
	}

	var srcBounds image.Rectangle
	dstBounds := target.Bounds()
	if mask == nil {
		srcBounds = dstBounds
	} else {
		srcBounds = mask.Bounds()
	}

	srcWidth, srcHeight := float32(srcBounds.Dx()), float32(srcBounds.Dy())
	dstMinX, dstMinY := float32(dstBounds.Min.X), float32(dstBounds.Min.Y)
	minX, minY := dstMinX+ox, dstMinY+oy
	maxX, maxY := minX+srcWidth, minY+srcHeight
	r.setDstRectCoords(minX, minY, maxX, maxY)

	srcMinX, srcMinY := float32(srcBounds.Min.X), float32(srcBounds.Min.Y)
	srcMaxX, srcMaxY := float32(srcBounds.Max.X), float32(srcBounds.Max.Y)
	r.setSrcRectCoords(srcMinX, srcMinY, srcMaxX, srcMaxY)
	fromF64, toF64 := colorToF64(from), colorToF64(to)
	memo := r.GetColorF32()
	fromOklab := rgbToOklab([3]float64(fromF64[:3]))
	toOklab := rgbToOklab([3]float64(toF64[:3]))
	r.SetColorF32(float32(toOklab[0]), float32(toOklab[1]), float32(toOklab[2]), float32(toF64[3]))
	r.setFlatCustomVAs(float32(fromOklab[0]), float32(fromOklab[1]), float32(fromOklab[2]), float32(fromF64[3]))

	r.opts.Uniforms["Area"] = [4]float32{ox, oy, srcWidth, srcHeight}
	r.opts.Uniforms["DirRadians"] = dirRadians
	r.opts.Uniforms["NumSteps"] = numSteps
	r.opts.Uniforms["CurveFactor"] = curveFactor
	if mask != nil {
		r.opts.Uniforms["UseMask"] = 1
	} else {
		r.opts.Uniforms["UseMask"] = 0
	}

	// draw shader
	r.opts.Images[0] = mask
	target.DrawTrianglesShader(r.vertices[:], r.indices[:], shaderGradient.Load(), &r.opts)
	r.opts.Images[0] = nil
	clear(r.opts.Uniforms)
	r.SetColorF32(memo[0], memo[1], memo[2], memo[3])
}

// Gradient paints a gradient over the given target, interpolating in Oklab space and
// using blue noise dithering to avoid color banding on subtle gradients.
//
// Function arguments behave the same as [Renderer.Gradient]().
func (r *Renderer) GradientDither(target *ebiten.Image, ox, oy, w, h float32, from, to color.RGBA, dirRadians, curveFactor float32) {
	if curveFactor < MinGradientCurveFactor {
		r.Warnings.report(WarnGradientCurveFactorLifted, curveFactor)
		curveFactor = MinGradientCurveFactor
	}
	r.ensureBlueNoiseLoaded()

	r.setDstRectCoords(ox, oy, ox+w, oy+h)

	fromF64, toF64 := colorToF64(from), colorToF64(to)
	memo := r.GetColorF32()
	fromOklab := rgbToOklab([3]float64(fromF64[:3]))
	toOklab := rgbToOklab([3]float64(toF64[:3]))
	r.SetColorF32(float32(toOklab[0]), float32(toOklab[1]), float32(toOklab[2]), float32(toF64[3]))
	r.setFlatCustomVAs(float32(fromOklab[0]), float32(fromOklab[1]), float32(fromOklab[2]), float32(fromF64[3]))

	sin, cos := math.Sincos(float64(dirRadians))
	r.opts.Uniforms["Area"] = [4]float32{ox, oy, w, h}
	r.opts.Uniforms["DirCosSin"] = [2]float32{float32(cos), float32(sin)}
	r.opts.Uniforms["CurveFactor"] = curveFactor

	// draw shader
	r.opts.Images[0] = r.blueNoise64RGB
	target.DrawTrianglesShader(r.vertices[:], r.indices[:], shaderGradientDither.Load(), &r.opts)
	r.opts.Images[0] = nil
	clear(r.opts.Uniforms)
	r.SetColorF32(memo[0], memo[1], memo[2], memo[3])
}

// GradientRadial paints a radial gradient over the given target, interpolating
// in Oklab space.
//
// Three radiuses are necessary:
//   - fromRadius: distances below this threshold take 'from' color. Use 0.0 if
//     you don't need a solid central area.
//   - transRadius: distances below this threshold but above fromRadius interpolate
//     colors between 'from' and 'to'.
//   - toRadius: distances below this threshold but above transRadius take 'to' color.
//     Distances above this threshold are not painted. Use toRadius = transRadius for
//     a gradient that ends at the given radius, or Float32Inf() if you want 'to'
//     color to extend beyond the gradient radius.
//
// Other parameters:
//   - numSteps: use -1 for continuous gradient, > 0 for discrete color steps.
//   - curveFactor: use 1.0 for linear, or see [MinGradientCurveFactor] for more details.
//
// To mask the gradient over an existing image, consider [Renderer.SetBlend](ebiten.BlendSourceIn)
// and similar tricks.
func (r *Renderer) GradientRadial(target *ebiten.Image, cx, cy float32, from, to color.RGBA, fromRadius, transRadius, toRadius float32, numSteps int, curveFactor float32) {
	if curveFactor < MinGradientCurveFactor {
		r.Warnings.report(WarnGradientCurveFactorLifted, curveFactor)
		curveFactor = MinGradientCurveFactor
	}
	if transRadius < fromRadius || toRadius < transRadius {
		r.Warnings.report(WarnInconsistentRangeOpSkipped, [3]float32{fromRadius, transRadius, toRadius})
	}

	dstBounds := target.Bounds()
	dstMinX, dstMinY := float32(dstBounds.Min.X), float32(dstBounds.Min.Y)
	dstWidthF64, dstHeightF64 := float64(dstBounds.Dx()), float64(dstBounds.Dy())
	cxF64, cyF64, toRadiusF64 := float64(cx), float64(cy), float64(toRadius)
	ox, oy := float32(max(math.Floor(cxF64-toRadiusF64), 0)), float32(max(math.Floor(cyF64-toRadiusF64), 0))
	fx, fy := float32(min(math.Ceil(cxF64+toRadiusF64), dstWidthF64)), float32(min(math.Ceil(cyF64+toRadiusF64), dstHeightF64))
	minX, minY := dstMinX+ox, dstMinY+oy
	maxX, maxY := dstMinX+fx, dstMinY+fy
	r.setDstRectCoords(minX, minY, maxX, maxY)

	fromF64, toF64 := colorToF64(from), colorToF64(to)
	memo := r.GetColorF32()
	fromOklab := rgbToOklab([3]float64(fromF64[:3]))
	toOklab := rgbToOklab([3]float64(toF64[:3]))
	r.SetColorF32(float32(toOklab[0]), float32(toOklab[1]), float32(toOklab[2]), float32(toF64[3]))
	r.setFlatCustomVAs(float32(fromOklab[0]), float32(fromOklab[1]), float32(fromOklab[2]), float32(fromF64[3]))

	r.opts.Uniforms["Radius"] = [3]float32{fromRadius, transRadius, toRadius}
	r.opts.Uniforms["Origin"] = [2]float32{cx, cy}
	r.opts.Uniforms["NumSteps"] = numSteps
	r.opts.Uniforms["CurveFactor"] = curveFactor

	// draw shader
	target.DrawTrianglesShader(r.vertices[:], r.indices[:], shaderGradientRadial.Load(), &r.opts)
	clear(r.opts.Uniforms)
	r.SetColorF32(memo[0], memo[1], memo[2], memo[3])
}

// GradientRadialDither paints a high quality radial gradient over the given target.
// GradientRadialDither paints a high quality radial gradient over the given target,
// using dithering to avoid color banding on subtle gradients. Function arguments
// behave the same as [Renderer.GradientRadial]().
func (r *Renderer) GradientRadialDither(target *ebiten.Image, cx, cy float32, from, to color.RGBA, fromRadius, transRadius, toRadius float32, curveFactor float32) {
	if curveFactor < MinGradientCurveFactor {
		r.Warnings.report(WarnGradientCurveFactorLifted, curveFactor)
		curveFactor = MinGradientCurveFactor
	}
	if transRadius < fromRadius || toRadius < transRadius {
		r.Warnings.report(WarnInconsistentRangeOpSkipped, [3]float32{fromRadius, transRadius, toRadius})
	}

	r.ensureBlueNoiseLoaded()

	dstBounds := target.Bounds()
	dstMinX, dstMinY := float32(dstBounds.Min.X), float32(dstBounds.Min.Y)
	dstWidthF64, dstHeightF64 := float64(dstBounds.Dx()), float64(dstBounds.Dy())
	cxF64, cyF64, toRadiusF64 := float64(cx), float64(cy), float64(toRadius)
	ox, oy := float32(max(math.Floor(cxF64-toRadiusF64), 0)), float32(max(math.Floor(cyF64-toRadiusF64), 0))
	fx, fy := float32(min(math.Ceil(cxF64+toRadiusF64), dstWidthF64)), float32(min(math.Ceil(cyF64+toRadiusF64), dstHeightF64))
	minX, minY := dstMinX+ox, dstMinY+oy
	maxX, maxY := dstMinX+fx, dstMinY+fy
	r.setDstRectCoords(minX, minY, maxX, maxY)

	fromF64, toF64 := colorToF64(from), colorToF64(to)
	memo := r.GetColorF32()
	fromOklab := rgbToOklab([3]float64(fromF64[:3]))
	toOklab := rgbToOklab([3]float64(toF64[:3]))
	r.SetColorF32(float32(toOklab[0]), float32(toOklab[1]), float32(toOklab[2]), float32(toF64[3]))
	r.setFlatCustomVAs(float32(fromOklab[0]), float32(fromOklab[1]), float32(fromOklab[2]), float32(fromF64[3]))

	r.opts.Uniforms["Radius"] = [3]float32{fromRadius, transRadius, toRadius}
	r.opts.Uniforms["Origin"] = [2]float32{cx, cy}
	r.opts.Uniforms["CurveFactor"] = curveFactor

	// draw shader
	r.opts.Images[0] = r.blueNoise64RGB
	target.DrawTrianglesShader(r.vertices[:], r.indices[:], shaderGradientRadialDither.Load(), &r.opts)
	r.opts.Images[0] = nil
	clear(r.opts.Uniforms)
	r.SetColorF32(memo[0], memo[1], memo[2], memo[3])
}

func rgbToOklab(rgb [3]float64) [3]float64 {
	linR, linG, linB := linearize(rgb[0]), linearize(rgb[1]), linearize(rgb[2])
	x := math.Pow(0.4122214708*linR+0.5363325363*linG+0.0514459929*linB, 1.0/3.0)
	y := math.Pow(0.2119034982*linR+0.6806995451*linG+0.1073969566*linB, 1.0/3.0)
	z := math.Pow(0.0883024619*linR+0.2817188376*linG+0.6299787005*linB, 1.0/3.0)

	l := 0.2104542553*x + 0.7936177850*y - 0.0040720468*z
	a := 1.9779984951*x - 2.4285922050*y + 0.4505937099*z
	b := 0.0259040371*x + 0.7827717662*y - 0.8086757660*z
	return [3]float64{l, a, b}
}

func linearize(colorChan float64) float64 {
	if colorChan >= 0.04045 {
		return math.Pow((colorChan+0.055)/1.055, 2.4)
	} else {
		return colorChan / 12.92
	}
}

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
	r.DrawShaderAt(target, source, x, y, 0, 0, shaderOklabShift.Load())
}

// ColorizeByLightness draws source into target at the given (x, y), taking the
// lightness of each source pixel and remapping it to a color between 'from' and 'to'.
//
// Key parameters:
//   - fromLightness: pixels below this threshold take 'from' color.
//     Expected range: [0.0, 1.0]
//   - toLightness: pixels above this threshold take 'to' color.
//     Expected range: [0.0, 1.0].
//   - steps: number of color steps in the gradient. Use steps <= 0 for a continuous gradient.
//   - curveFactor: adjusts the gradient's interpolation curve; use 1.0 for linear,
//     <= 1.0 to bias towards 'from', > 1.0 to bias towards 'to'. Recommended
//     range: [0.2, 5.0].
func (r *Renderer) ColorizeByLightness(target, source *ebiten.Image, x, y float32, from, to color.RGBA, fromLightness, toLightness float32, steps int, curveFactor float32) {
	fromF64, toF64 := colorToF64(from), colorToF64(to)
	fromOklab, toOklab := rgbToOklab([3]float64(fromF64[:3])), rgbToOklab([3]float64(toF64[:3]))
	r.opts.Uniforms["From"] = [4]float32{float32(fromOklab[0]), float32(fromOklab[1]), float32(fromOklab[2]), float32(fromF64[3])}
	r.opts.Uniforms["To"] = [4]float32{float32(toOklab[0]), float32(toOklab[1]), float32(toOklab[2]), float32(toF64[3])}
	r.setFlatCustomVAs(fromLightness, toLightness, float32(steps), curveFactor)
	r.DrawShaderAt(target, source, x, y, 0, 0, shaderColorizeByLightness.Load())
	clear(r.opts.Uniforms)
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
		r.ensureBlueNoiseLoaded()
		r.opts.Uniforms["Dither"] = 1
		r.opts.Images[2] = r.blueNoise64RGB
	}

	if bilinear {
		r.setFlatCustomVAs01(alpha, mixLevel)
		r.DrawShaderAt(target, base, x, y, 0, 0, shaderColorMixBilinear.Load())
	} else {
		r.setFlatCustomVAs01(alpha, mixLevel)
		r.DrawShaderAt(target, base, x, y, 0, 0, shaderColorMix.Load())
	}

	r.opts.Images[1] = nil
	if dither {
		clear(r.opts.Uniforms)
		r.opts.Images[2] = nil
	}
}

var DitherBayes [16]float32 = [16]float32{
	0.0 / 16.0, 12.0 / 16.0, 3.0 / 16.0, 15.0 / 16.0,
	8.0 / 16.0, 4.0 / 16.0, 11.0 / 16.0, 7.0 / 16.0,
	2.0 / 16.0, 14.0 / 16.0, 1.0 / 16.0, 13.0 / 16.0,
	10.0 / 16.0, 6.0 / 16.0, 9.0 / 16.0, 5.0 / 16.0,
}
var DitherDots [16]float32 = [16]float32{
	12.0 / 16.0, 4.0 / 16.0, 11.0 / 16.0, 15.0 / 16.0,
	5.0 / 16.0, 0.0 / 16.0, 3.0 / 16.0, 10.0 / 16.0,
	6.0 / 16.0, 1.0 / 16.0, 2.0 / 16.0, 9.0 / 16.0,
	13.0 / 16.0, 7.0 / 16.0, 8.0 / 16.0, 14.0 / 16.0,
}
var DitherSerp [16]float32 = [16]float32{
	0.0 / 16.0, 12.0 / 16.0, 13.0 / 16.0, 1.0 / 16.0,
	3.0 / 16.0, 7.0 / 16.0, 6.0 / 16.0, 2.0 / 16.0,
	4.0 / 16.0, 8.0 / 16.0, 9.0 / 16.0, 5.0 / 16.0,
	11.0 / 16.0, 15.0 / 16.0, 14.0 / 16.0, 10.0 / 16.0,
}
var DitherGlitch [16]float32 = [16]float32{
	0.0 / 16.0, 1.0 / 16.0, 2.0 / 16.0, 3.0 / 16.0,
	4.0 / 16.0, 5.0 / 16.0, 6.0 / 16.0, 7.0 / 16.0,
	8.0 / 16.0, 9.0 / 16.0, 10.0 / 16.0, 11.0 / 16.0,
	12.0 / 16.0, 13.0 / 16.0, 14.0 / 16.0, 15.0 / 16.0,
}

var DitherCrumbs [16]float32 = [16]float32{
	0.0 / 16.0, 4.0 / 16.0, 8.0 / 16.0, 1.0 / 16.0,
	11.0 / 16.0, 14.0 / 16.0, 12.0 / 16.0, 5.0 / 16.0,
	7.0 / 16.0, 13.0 / 16.0, 15.0 / 16.0, 9.0 / 16.0,
	3.0 / 16.0, 10.0 / 16.0, 6.0 / 16.0, 2.0 / 16.0,
}

var DitherBW []float32 = []float32{
	0.0, 0.0, 0.0, 1.0,
	1.0, 1.0, 1.0, 1.0,
}
var DitherBW4 []float32 = []float32{
	0.0, 0.0, 0.0, 1.0,
	0.333, 0.333, 0.333, 1.0,
	0.666, 0.666, 0.666, 1.0,
	1.0, 1.0, 1.0, 1.0,
}

var DitherAlpha8 []float32 = []float32{
	0.0, 0.0, 0.0, 0.0,
	1.0 / 7.0, 1.0 / 7.0, 1.0 / 7.0, 1.0 / 7.0,
	2.0 / 7.0, 2.0 / 7.0, 2.0 / 7.0, 2.0 / 7.0,
	3.0 / 7.0, 3.0 / 7.0, 3.0 / 7.0, 3.0 / 7.0,
	4.0 / 7.0, 4.0 / 7.0, 4.0 / 7.0, 4.0 / 7.0,
	5.0 / 7.0, 5.0 / 7.0, 5.0 / 7.0, 5.0 / 7.0,
	6.0 / 7.0, 6.0 / 7.0, 6.0 / 7.0, 6.0 / 7.0,
	1.0, 1.0, 1.0, 1.0,
}
var DitherBRG []float32 = []float32{
	0.0, 0.0, 1.0, 1.0,
	1.0, 0.0, 0.0, 1.0,
	0.0, 1.0, 0.0, 1.0,
}

// DitherMat4 draws the given mask to the target applying a static 4x4 dithering pattern to select colors from
// rgbaColors. The rgbaColors argument can contain up to 8 colors, flattened as RGBA quadruplets in [0...1] range.
// You can test with DitherBW4. The ditherMatrix argument is a 4x4 dithering matrix in column major order (like
// GLSL), where the values indicate the thresholds of the pattern in 0...1 range. You can test with [DitherBayes].
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
	r.DrawShaderAt(target, mask, ox, oy, 0, 0, shaderDitherMat4.Load())
	clear(r.opts.Uniforms)
}
