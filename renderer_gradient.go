package shapes

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

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

// notice: internal use only, doesn't touch singleClr flag
func (r *Renderer) setColors(values [16]float32) {
	for i := range 4 {
		r.vertices[i].ColorR = values[i<<2+0]
		r.vertices[i].ColorG = values[i<<2+1]
		r.vertices[i].ColorB = values[i<<2+2]
		r.vertices[i].ColorA = values[i<<2+3]
	}
}

// Gradient paints a gradient over the given target, interpolating in Oklab space.
// For common dirRadians values, consider [DirRadsLTR] and related constants.
//
// If you need to apply the gradient over a mask, use ebiten.BlendSourceIn:
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
	r.setColors(memo)
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
	r.setColors(memo)
}
