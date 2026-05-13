package shapes

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

// Blur applies a naive, quadratic gaussian blur to the given mask and draws it onto
// the given target.
//
// Radius can't exceed 16. Internally, the gaussian's std deviation is σ = radius/3.
//
// This operation is affected by [Renderer.Tint].
//
// Notice that this method is designed mostly as a comparison baseline due to its high
// cost (a radius of 8 will sample (8*2)^2 = 256 pixels!). Most use-cases should rely
// on [Renderer.Blur2]() instead.
func (r *Renderer) Blur(target *ebiten.Image, mask *ebiten.Image, ox, oy, radius float32) {
	if mask == nil {
		r.Warnings.report(WarnMissingSourceOpSkipped, mask)
		return
	}
	if radius > 16 {
		r.Warnings.report(WarnRadiusClamped, radius)
		radius = 16
	} else if radius < 0 {
		r.Warnings.report(WarnNegativeValueZeroed, radius)
		radius = 0
	}

	srcBounds := mask.Bounds()
	srcWidth, srcHeight := float32(srcBounds.Dx()), float32(srcBounds.Dy())
	minX, minY := ox-radius, oy-radius
	maxX, maxY := ox+srcWidth+radius, oy+srcHeight+radius
	r.setDstRectCoords(minX-1, minY-1, maxX+1, maxY+1)

	srcMinX, srcMinY := float32(srcBounds.Min.X), float32(srcBounds.Min.Y)
	srcMaxX, srcMaxY := float32(srcBounds.Max.X), float32(srcBounds.Max.Y)
	r.setSrcRectCoords(srcMinX-radius-1, srcMinY-radius-1, srcMaxX+radius+1.0, srcMaxY+radius+1.0)
	r.setFlatCustomVAs01(radius, r.tint)

	// draw shader
	r.opts.Images[0] = mask
	target.DrawTrianglesShader32(r.vertices[:], r.indices[:], shaderBlurNaive.Load(), &r.opts)
	r.opts.Images[0] = nil
}

// Blur2 is similar to [Renderer.Blur](), but uses two 1D passes instead of a single 2D
// pass. This greatly reduces the amount of sampled pixels for the shader, and despite
// breaking batching tends to be much more efficient than [Renderer.Blur](). Radius can't
// exceed 32.
//
// This operation is affected by [Renderer.Tint].
//
// This function uses one internal offscreen (#0), and target and mask can be on the same
// internal atlas.
//
// See also [Renderer.BlurK]() if fixed radiuses are acceptable.
func (r *Renderer) Blur2(target *ebiten.Image, mask *ebiten.Image, ox, oy, radius float32) {
	if mask == nil {
		r.Warnings.report(WarnMissingSourceOpSkipped, mask)
		return
	}
	if radius > 32 {
		r.Warnings.report(WarnRadiusClamped, radius)
		radius = 32
	} else if radius < 0 {
		r.Warnings.report(WarnNegativeValueZeroed, radius)
		radius = 0
	}

	ceilRadius := ceilF32(radius)
	w32, h32 := rectSizeF32(mask.Bounds())
	h32 += 2.0 * ceilRadius
	tmp, _ := r.getTemp(0, int(w32), int(h32), false)
	preBlend := r.opts.Blend
	r.opts.Blend = ebiten.BlendCopy
	memo := r.tint
	r.tint = 0.0
	r.blurVert(tmp, mask, 0, ceilRadius, radius)
	r.opts.Blend = preBlend
	r.tint = memo
	r.blurHorz(target, tmp, ox, oy-ceilRadius, radius)
}

// blurVert applies a 1D vertical blur pass of the given radius, which can't exceed 32.
//
// This operation is affected by [Renderer.Tint].
func (r *Renderer) blurVert(target *ebiten.Image, mask *ebiten.Image, ox, oy, radius float32) {
	if mask == nil {
		r.Warnings.report(WarnMissingSourceOpSkipped, mask)
		return
	}
	if radius > 32 {
		r.Warnings.report(WarnRadiusClamped, radius)
		radius = 32
	} else if radius < 0 {
		r.Warnings.report(WarnNegativeValueZeroed, radius)
		radius = 0
	}

	sox, soy, sw, sh := rectOriginSizeF32(mask.Bounds())
	ceilRadius := ceilF32(radius)
	r.setDstRectCoords(ox, oy-ceilRadius, ox+sw, oy+sh+ceilRadius)
	r.setSrcRectCoords(sox, soy-ceilRadius, sox+sw, soy+sh+ceilRadius)
	r.setFlatCustomVAs01(radius, r.tint)

	// draw shader
	r.opts.Images[0] = mask
	target.DrawTrianglesShader32(r.vertices[:], r.indices[:], shaderBlurVert.Load(), &r.opts)
	r.opts.Images[0] = nil
}

// blurHorz applies a 1D horizontal blur pass of the given radius, which can't exceed 32.
//
// This operation is affected by [Renderer.Tint].
func (r *Renderer) blurHorz(target *ebiten.Image, mask *ebiten.Image, ox, oy, radius float32) {
	if mask == nil {
		r.Warnings.report(WarnMissingSourceOpSkipped, mask)
		return
	}
	if radius > 32 {
		r.Warnings.report(WarnRadiusClamped, radius)
		radius = 32
	} else if radius < 0 {
		r.Warnings.report(WarnNegativeValueZeroed, radius)
		radius = 0
	}

	srcBounds := mask.Bounds()
	srcWidth, srcHeight := float32(srcBounds.Dx()), float32(srcBounds.Dy())
	minX, minY := ox-radius, oy
	maxX, maxY := ox+srcWidth+radius, oy+srcHeight
	r.setDstRectCoords(minX, minY, maxX, maxY)

	srcMinX, srcMinY := float32(srcBounds.Min.X), float32(srcBounds.Min.Y)
	srcMaxX, srcMaxY := float32(srcBounds.Max.X), float32(srcBounds.Max.Y)
	r.setSrcRectCoords(srcMinX-radius, srcMinY, srcMaxX+radius, srcMaxY)
	r.setFlatCustomVAs01(radius, r.tint)

	// draw shader
	r.opts.Images[0] = mask
	target.DrawTrianglesShader32(r.vertices[:], r.indices[:], shaderBlurHorz.Load(), &r.opts)
	r.opts.Images[0] = nil
}

// BlurK is a separable multipass blur with fixed radius and optional downscaling.
// See [KernelOptions] for more details.
//
// This operation is affected by [Renderer.Tint].
//
// This function uses the internal offscreen (#0), and if downscaling also (#1).
// Target and mask can be on the same internal atlas.
func (r *Renderer) BlurK(target *ebiten.Image, mask *ebiten.Image, ox, oy float32, opts KernelOptions) {
	invokeShader := func(downHorzTarget *ebiten.Image) {
		r.setFlatCustomVA0(r.tint)
		downHorzTarget.DrawTrianglesShader32(r.vertices[:], r.indices[:], shaderBlurHorzKern.Load(), &r.opts)
	}
	r.applyKernel(target, mask, ox, oy, opts, invokeShader, false)
}

// BlurVogel applies a gaussian blur using numSamples distributed with a vogel disk.
//
// In comparison to pure gaussian blurs, vogel blurs have a more grainy, frosted glass look.
// This can look anywhere from artistic to noisy. It can be a reasonable way to implement
// bokeh or depth of field effects. It works well on organic images, much less so for isolated
// shapes with very few colors.
//
// Common numSamples values:
//   - 16: low quality and noisy, but fast and practical in certain scenarios.
//   - 32: medium quality and useful for a wide variety of blur effects.
//   - 64: maximum allowed value.
//
// This operation is affected by [Renderer.Tint].
//
// If downscaling is used:
//   - The function will use internal offscreens (#0, #1). In this case, target and mask
//     can be on the same internal atlas.
//   - The radius will be applied 'as is' to a downscaled version of mask before upscaling
//     back to draw on target. For example: if radius 16 and [DownscaleX4] are used, the
//     resulting radius will be closer to 16*4 = 64.
func (r *Renderer) BlurVogel(target, mask *ebiten.Image, ox, oy, radius float32, numSamples int, downscaling Downscaling, seed float32) {
	if mask == nil {
		r.Warnings.report(WarnMissingSourceOpSkipped, mask)
		return
	}
	if !downscaling.valid() {
		panic("invalid downscaling value")
	}
	if numSamples < 1 {
		r.Warnings.report(WarnNotEnoughSamplesOpSkipped, numSamples)
		return
	}
	if numSamples > 64 {
		r.Warnings.report(WarnNumSamplesClamped, numSamples)
		numSamples = 64
	}
	if radius < 0 {
		r.Warnings.report(WarnNegativeValueZeroed, radius)
		radius = 0
	}
	if seed < 0 || seed > 1 {
		r.Warnings.report(WarnInvalidNoiseSeedClamped, seed)
		seed = clamp(seed, 0, 1)
	}

	if r.vogelMemo == nil {
		r.vogelMemo = &vogelMemory{}
	}
	r.vogelMemo.Refresh(numSamples, 1.0)
	df := downscaling.Factor()
	if df == 1 {
		r.blurVogelDirect(target, mask, ox, oy, radius, numSamples, seed, 0.0)
	} else {
		r.blurVogelDownscaled(target, mask, ox, oy, radius, numSamples, df, seed)
	}
}

// helper function for BlurVogel. precondition: all input parameters have already been validated
//
// This operation is affected by [Renderer.Tint].
func (r *Renderer) blurVogelDirect(target, mask *ebiten.Image, ox, oy, radius float32, numSamples int, seed, padOffset float32) {
	sox, soy, sw, sh := rectOriginSizeF32(mask.Bounds())
	radiusF32 := float32(radius)

	minX, minY := ox-radiusF32, oy-radiusF32
	maxX, maxY := ox+sw+radiusF32, oy+sh+radiusF32
	r.setDstRectCoords(minX-1, minY-1, maxX+1, maxY+1)
	r.setSrcRectCoords(sox-radiusF32-1, soy-radiusF32-1, sox+sw+radiusF32+1.0, soy+sh+radiusF32+1.0)
	r.setFlatCustomVAs(radiusF32, r.tint, seed, padOffset)

	// draw shader
	r.opts.Images[0] = mask
	r.opts.Uniforms["NumSamples"] = numSamples
	r.opts.Uniforms["Disk"] = r.vogelMemo.Points
	target.DrawTrianglesShader32(r.vertices[:], r.indices[:], shaderBlurVogel.Load(), &r.opts)
	r.opts.Images[0] = nil
	clear(r.opts.Uniforms)
}

// helper function for BlurVogel. precondition: all input parameters have already been validated,
// and downscale is an strictly positive even number
//
// This operation is affected by [Renderer.Tint].
func (r *Renderer) blurVogelDownscaled(target, mask *ebiten.Image, ox, oy, radius float32, numSamples int, downscale int, seed float32) {
	_, _, sw, sh := rectOriginSize(mask.Bounds())
	ds := 1.0 / float64(downscale)
	tw := float64(sw) * ds
	th := float64(sh) * ds

	tmp, clear := r.getTemp(0, int(math.Ceil(tw)), int(math.Ceil(th)), false)
	if !clear {
		b := tmp.Bounds()
		preBlend := r.opts.Blend
		r.opts.Blend = ebiten.BlendClear
		r.FillIntRect(tmp, clockwiseRightBorder(b, 1), 0)
		r.FillIntRect(tmp, bottomBorder(b, 1), 0)
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
	r.blurVogelDirect(mid, tmp, shift, shift, radius, numSamples, seed, shift)
	scaledShift := shift * float32(downscale)
	ox -= scaledShift
	oy -= scaledShift

	// upscale
	var scaleOpts ScaleOptions
	if downscale < 4 {
		scaleOpts.Bicubic = true
	}
	r.Scale(target, mid, ox, oy, float32(downscale), &scaleOpts)
}

// vogelMemory is a helper for vogel disks data
type vogelMemory struct {
	Points        [128]float32
	SinCos        [][2]float64
	LastRadius    float64
	StickyPtCount int // highest point count for LastRadius
}

func (m *vogelMemory) Refresh(numPoints int, radius float64) {
	const GoldenAngle = 2.39996322972865332223155550663361385312499901105811504293511275 // https://oeis.org/A131988

	if numPoints > 128 {
		panic("vogel disk numPoints can't exceed 128")
	}

	if numPoints <= m.StickyPtCount && radius == m.LastRadius {
		return // already cached
	}

	for i := float64(len(m.SinCos)); i < float64(numPoints); i += 1.0 {
		theta := i * GoldenAngle
		sin, cos := math.Sincos(theta)
		m.SinCos = append(m.SinCos, [2]float64{sin, cos})
	}

	for i := range numPoints {
		dist := radius * math.Sqrt(float64(i)/float64(numPoints))
		m.Points[i<<1+0] = float32(dist * m.SinCos[i][1]) // X
		m.Points[i<<1+1] = float32(dist * m.SinCos[i][0]) // Y
	}

	if m.LastRadius != radius || numPoints > m.StickyPtCount {
		m.StickyPtCount = numPoints
	}
	m.LastRadius = radius
}
