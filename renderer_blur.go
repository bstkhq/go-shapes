package shapes

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

// TODO: implement new blurs on this file and progressively refactor

// Downscaling can be used with some operations
type Downscaling uint8

const (
	DownscaleNone Downscaling = iota
	DownscaleX2
	DownscaleX4
	DownscaleX8
	DownscaleX16
)

func (d Downscaling) valid() bool {
	return d >= DownscaleNone && d <= DownscaleX16
}

func (d Downscaling) Factor() int {
	return int(1 << d)
}

// ApplyBlurVogel applies a gaussian blur using numSamples distributed with a vogel disk.
//
// In comparison to pure gaussian blurs, vogel blurs have a more grainy, frosted glass look.
// This can look anywhere from artistic to noisy. It can be an efficient way to implement
// bokeh or depth of field effects. It works well on full images, much less so for specific,
// isolated shapes.
//
// Common numSamples values:
//   - 16: low quality, but fast and practical in many scenarios.
//   - 32: medium quality and useful for a wide variety of blur effects.
//   - 64: maximum allowed value.
//
// Use colorMix = 0 to use the renderer colors, 1 for mask colors, or anything in between
// to perform interpolation. seed controls the jitter noise, which can be static or animated
// between [0, 1].
//
// If downscaling is != DownscaleNone, notice that:
//   - The function will use internal offscreens (#0, #1), and target and mask can be on the
//     same internal atlas.
//   - radius will be applied 'as is' to a downscaled version of mask before upscaling back to
//     draw on target. This means that if radius 16 and DownscaleX4 are used, the actual radius
//     effect will be closer to 16*4 = 64.
func (r *Renderer) ApplyBlurVogel(target, mask *ebiten.Image, ox, oy, radius, colorMix float32, numSamples int, downscaling Downscaling, seed float32) {
	if mask == nil {
		r.Warnings.report(WarnMissingSourceOpSkipped, mask)
		return
	}
	if !downscaling.valid() {
		panic("invalid downscaling value")
	}
	if radius < 0 {
		r.Warnings.report(WarnNegativeValueOpSkipped, radius)
		return
	}
	if numSamples < 1 {
		r.Warnings.report(WarnNotEnoughSamplesOpSkipped, numSamples)
		return
	}
	if numSamples > 64 {
		r.Warnings.report(WarnNumSamplesClamped, numSamples)
		numSamples = 64
	}
	if colorMix < 0 || colorMix > 1 {
		r.Warnings.report(WarnInvalidColorMixClamped, colorMix)
		colorMix = clamp(colorMix, 0, 1)
	}
	if seed < 0 || seed > 1 {
		r.Warnings.report(WarnInvalidNoiseSeedClamped, seed)
		seed = clamp(seed, 0, 1)
	}

	r.refreshVogelPoints(numSamples, 1.0)
	df := downscaling.Factor()
	if df == 1 {
		r.applyBlurVogelDirect(target, mask, ox, oy, radius, colorMix, numSamples, seed, 0.0)
	} else {
		r.applyBlurVogelDownscaled(target, mask, ox, oy, radius, colorMix, numSamples, df, seed)
	}
}

// helper function for ApplyBlurVogel. precondition: all input parameters have already been validated
func (r *Renderer) applyBlurVogelDirect(target, mask *ebiten.Image, ox, oy, radius, colorMix float32, numSamples int, seed, padOffset float32) {
	dox, doy := rectOriginF32(target.Bounds())
	sox, soy, sw, sh := rectOriginSizeF32(mask.Bounds())
	radiusF32 := float32(radius)

	minX, minY := dox+ox-radiusF32, doy+oy-radiusF32
	maxX, maxY := dox+ox+sw+radiusF32, doy+oy+sh+radiusF32
	r.setDstRectCoords(minX-1, minY-1, maxX+1, maxY+1)
	r.setSrcRectCoords(sox-radiusF32-1, soy-radiusF32-1, sox+sw+radiusF32+1.0, soy+sh+radiusF32+1.0)
	r.setFlatCustomVAs(radiusF32, colorMix, seed, padOffset)

	// draw shader
	r.opts.Images[0] = mask
	r.opts.Uniforms["NumSamples"] = numSamples
	r.opts.Uniforms["Disk"] = r.vogelPoints
	target.DrawTrianglesShader(r.vertices[:], r.indices[:], shaderBlurVogel.Load(), &r.opts)
	r.opts.Images[0] = nil
	clear(r.opts.Uniforms)
}

// helper function for ApplyBlurVogel. precondition: all input parameters have already been validated,
// and downscale is an strictly positive even number
func (r *Renderer) applyBlurVogelDownscaled(target, mask *ebiten.Image, ox, oy, radius, colorMix float32, numSamples int, downscale int, seed float32) {
	_, _, sw, sh := rectOriginSize(mask.Bounds())
	ds := 1.0 / float64(downscale)
	tw := float64(sw) * ds
	th := float64(sh) * ds

	tmp, clear := r.getTemp(0, int(math.Ceil(tw)), int(math.Ceil(th)), false)
	if !clear {
		b := tmp.Bounds()
		preBlend := r.opts.Blend
		r.opts.Blend = ebiten.BlendClear
		r.DrawIntRect(tmp, clockwiseRightBorder(b, 1))
		r.DrawIntRect(tmp, bottomBorder(b, 1))
		r.opts.Blend = preBlend
	}

	// downscale
	var opts ebiten.DrawImageOptions
	opts.Filter = ebiten.FilterLinear
	opts.GeoM.Scale(ds, ds) // TODO: see if this is correct with mask non-zero origins
	opts.Blend = ebiten.BlendCopy
	tmp.DrawImage(mask, &opts)

	// apply blur
	radiusInt := int(math.Ceil(float64(radius)))
	effectBounds := tmp.Bounds().Inset(-radiusInt - 1)
	mid, _ := r.getTemp(1, effectBounds.Dx(), effectBounds.Dy(), true) // TODO: maybe I should optimize clears here, consider r.brClear(n) and r.borderClear(n)
	shift := float32(radiusInt + 1)
	r.applyBlurVogelDirect(mid, tmp, shift, shift, radius, colorMix, numSamples, seed, shift)
	scaledShift := shift * float32(downscale)
	ox -= scaledShift
	oy -= scaledShift

	// upscale
	if downscale < 4 {
		r.Scale(target, mid, ox, oy, float32(downscale), false)
	} else {
		r.ScaleBicubic(target, mid, ox, oy, float32(downscale), false)
	}
}

func (r *Renderer) refreshVogelPoints(n int, radius float64) {
	const GoldenAngle = 2.39996322972865332223155550663361385312499901105811504293511275 // https://oeis.org/A131988

	if n <= r.vogelStickyN && radius == r.vogelLastRadius {
		return // already cached
	}

	for i := float64(len(r.vogelSinCos)); i < float64(n); i += 1.0 {
		theta := i * GoldenAngle
		sin, cos := math.Sincos(theta)
		r.vogelSinCos = append(r.vogelSinCos, [2]float64{sin, cos})
	}

	for i := range n {
		dist := radius * math.Sqrt(float64(i)/float64(n))
		r.vogelPoints[i<<1+0] = float32(dist * r.vogelSinCos[i][1]) // X
		r.vogelPoints[i<<1+1] = float32(dist * r.vogelSinCos[i][0]) // Y
	}

	if r.vogelLastRadius != radius {
		r.vogelStickyN = n
	} else {
		r.vogelStickyN = max(r.vogelStickyN)
	}
	r.vogelLastRadius = radius
}
