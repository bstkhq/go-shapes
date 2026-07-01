package shapes

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

// GradientOptions are used for multiple color operations like
// [Renderer.Gradient]() and [Renderer.GradientRadial]().
type GradientOptions struct {
	// Starting gradient color.
	From [4]float64

	// End gradient color.
	To [4]float64

	// Steps controls the number of colors in the gradient.
	//  - Steps <= 0 specifies a continuous gradient (no color limit).
	//  - Steps > 0 specifies a stepped gradient.
	Steps int

	// Dither determines whether the gradient will have dithering applied.
	//
	// Dithering is only necessary on subtle gradients or alpha transitions,
	// where very similar colors can cause banding or flickering.
	Dither bool

	// Bias is a value in [-1, 1] that controls the color interpolation:
	//  - Zero generates a linear gradient (both colors have the same presence).
	//  - Negative values give the start color more presence.
	//  - Positive values give the end color more presence.
	//
	// The interpolation is based on Schlick's bias function.
	Bias float32
}

// GradientOpts creates GradientOptions for a continuous gradient.
func GradientOpts(from, to color.Color, dither bool) GradientOptions {
	return GradientOptions{
		From:   colorToF64(from),
		To:     colorToF64(to),
		Dither: dither,
	}
}

// StepGradientOpts creates GradientOptions for a stepped gradient.
func StepGradientOpts(from, to color.Color, steps int) GradientOptions {
	return GradientOptions{
		From:  colorToF64(from),
		To:    colorToF64(to),
		Steps: steps,
	}
}

// Gradient paints a gradient over the given target, interpolating in Oklab space.
// For common dirRadians values, see [DirRadsLTR] and related constants.
//
// If you need to apply the gradient over a mask, use [ebiten.BlendSourceIn]:
//
//	tmp := r.UnsafeTempCopy(0, mask, 0, false) // optionally make a mask copy
//	r.Options().Blend = ebiten.BlendSourceIn
//	r.Gradient(tmp, opts, DirRadsBLTR) // gradient applied to mask copy
//	r.Options().Blend = ebiten.BlendSourceOver
//	r.DrawAt(target, tmp, x, y, alpha)
//
// See also [Renderer.GradientRadial]().
func (r *Renderer) Gradient(target *ebiten.Image, opts GradientOptions, dirRadians float64) {
	if opts.Bias < -1.0 || opts.Bias > 1.0 {
		r.Warnings.report(WarnInvalidBiasClamped, opts.Bias)
		opts.Bias = clamp(opts.Bias, -1.0, 1.0)
	}

	memo := r.memorizeColors()

	fromL, fromA, fromB := toOklab(opts.From[0], opts.From[1], opts.From[2])
	toL, toA, toB := toOklab(opts.To[0], opts.To[1], opts.To[2])
	r.SetColorF32(float32(toL), float32(toA), float32(toB), float32(opts.To[3]))
	r.setFlatCustomVAs(float32(fromL), float32(fromA), float32(fromB), float32(opts.From[3]))

	if opts.Dither {
		r.opts.Uniforms["Dither"] = 1
		r.loadBlueNoise64RGBAt(0)
	}
	dirSin, dirCos := math.Sincos(dirRadians)
	tox, toy, tw, th := rectOriginSizeF32(target.Bounds())

	r.opts.Uniforms["Dir"] = [2]float32{float32(dirCos), float32(dirSin)}
	r.opts.Uniforms["NumSteps"] = max(opts.Steps, 0)
	r.opts.Uniforms["Bias"] = (opts.Bias + 1.0) / 2.0
	r.DrawRectShader(target, tox, toy, tw, th, NoMargins, shaderGradient.Load())

	r.opts.Images[0] = nil
	clear(r.opts.Uniforms)
	r.restoreColors(memo)
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
// See also [Renderer.Gradient]().
func (r *Renderer) GradientRadial(target *ebiten.Image, opts GradientOptions, cx, cy float32, fromRadius, transRadius, toRadius float32) {
	if opts.Bias < -1.0 || opts.Bias > 1.0 {
		r.Warnings.report(WarnInvalidBiasClamped, opts.Bias)
		opts.Bias = clamp(opts.Bias, -1.0, 1.0)
	}
	if transRadius < fromRadius || toRadius < transRadius {
		r.Warnings.report(WarnInconsistentRangeOpSkipped, [3]float32{fromRadius, transRadius, toRadius})
		return // TODO: relax?
	}

	memo := r.memorizeColors()

	fromL, fromA, fromB := toOklab(opts.From[0], opts.From[1], opts.From[2])
	toL, toA, toB := toOklab(opts.To[0], opts.To[1], opts.To[2])
	r.SetColorF32(float32(toL), float32(toA), float32(toB), float32(opts.To[3]))
	r.setFlatCustomVAs(float32(fromL), float32(fromA), float32(fromB), float32(opts.From[3]))

	if opts.Dither {
		r.opts.Uniforms["Dither"] = 1
		r.loadBlueNoise64RGBAt(0)
	}

	tox, toy, tw, th := rectOriginSizeF32(target.Bounds())
	r.opts.Uniforms["Radius"] = [3]float32{fromRadius, transRadius, toRadius}
	r.opts.Uniforms["Origin"] = [2]float32{cx, cy}
	r.opts.Uniforms["NumSteps"] = opts.Steps
	r.opts.Uniforms["Bias"] = (opts.Bias + 1.0) / 2.0
	r.DrawRectShader(target, tox, toy, tw, th, NoMargins, shaderGradientRadial.Load())

	r.opts.Images[0] = nil
	clear(r.opts.Uniforms)
	r.restoreColors(memo)
}
