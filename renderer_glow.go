package shapes

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

// Glow2 draws a separable two-pass glow effect for the given mask into the target,
// at the given coordinates.
//
// threshStart and threshEnd indicate the start luminosity threshold at which the glow
// effect kicks in and the point at which it's fully active. threshStart must be <=
// threshEnd, and the values must be in [0, 1] range.
//
// For reference thresholds, 0.4 to 0.7 is a good general default range.
//
// This operation is affected by [Renderer.Tint].
//
// Notice that this effect uses an internal offscreen (#0). Target and mask can be
// on the same internal atlas. Neither horzRadius nor vertRadius can exceed 32.
func (r *Renderer) Glow2(target *ebiten.Image, mask *ebiten.Image, ox, oy, horzRadius, vertRadius, threshStart, threshEnd float32) {
	if threshStart > threshEnd {
		r.Warnings.report(WarnInconsistentRangeOpSkipped, [2]float32{threshStart, threshEnd})
		return
	}
	horzRadius = r.warnClampNonNegArgF32(horzRadius, 32, WarnRadiusClamped)
	vertRadius = r.warnClampNonNegArgF32(vertRadius, 32, WarnRadiusClamped)
	if vertRadius == 0 {
		// TODO: optimize only horz / only vert cases
	}

	srcBounds := mask.Bounds()
	srcWidth, srcHeight := float32(srcBounds.Dx()), float32(srcBounds.Dy())
	w32, h32 := float32(srcWidth), float32(srcHeight)+vertRadius*2.0
	w, h := int(w32), int(math.Ceil(float64(h32)))
	tmp, _ := r.getTemp(0, w, h, false)

	r.setDstRectCoords(0, 0, w32, h32+2)

	srcMinX, srcMinY := float32(srcBounds.Min.X), float32(srcBounds.Min.Y)
	srcMaxX, srcMaxY := float32(srcBounds.Max.X), float32(srcBounds.Max.Y)
	r.setSrcRectCoords(srcMinX, srcMinY-vertRadius-1, srcMaxX, srcMaxY+vertRadius+1.0)
	r.setFlatCustomVAs(vertRadius, threshStart, threshEnd, r.tint)

	// first pass (threshold + vertical blur)
	r.opts.Images[0] = mask
	preBlend := r.opts.Blend
	r.opts.Blend = ebiten.BlendCopy
	tmp.DrawTrianglesShader32(r.vertices[:], r.indices[:], shaderGlowVert.Load(), &r.opts)
	r.opts.Images[0] = nil

	// second pass
	r.opts.Blend = ebiten.BlendLighter
	r.blurHorz(target, tmp, ox, oy-vertRadius-1.0, horzRadius)
	r.opts.Blend = preBlend
}

// glowHorz draws a horizontal glow effect for the given mask into the target, at the
// given coordinates. See [Renderer.Glow]() for additional documentation. Compared to
// Renderer.Glow, this effect only applies the glow horizontally and it's much cheaper,
// requiring no offscreen and a single pass.
//
// This operation is affected by [Renderer.Tint].
//
// horzRadius can't exceed 32.
func (r *Renderer) glowHorz(target *ebiten.Image, mask *ebiten.Image, ox, oy, horzRadius, threshStart, threshEnd float32) {
	if threshStart > threshEnd {
		r.Warnings.report(WarnInconsistentRangeOpSkipped, [2]float32{threshStart, threshEnd})
		return
	}
	horzRadius = r.warnClampNonNegArgF32(horzRadius, 32, WarnRadiusClamped)

	srcBounds := mask.Bounds()
	srcWidth, srcHeight := float32(srcBounds.Dx()), float32(srcBounds.Dy())

	r.setDstRectCoords(ox-horzRadius-1.0, oy, ox+float32(srcWidth)+horzRadius+1.0, oy+float32(srcHeight))

	srcMinX, srcMinY := float32(srcBounds.Min.X), float32(srcBounds.Min.Y)
	srcMaxX, srcMaxY := float32(srcBounds.Max.X), float32(srcBounds.Max.Y)
	r.setSrcRectCoords(srcMinX-horzRadius-1, srcMinY, srcMaxX+horzRadius+1, srcMaxY)
	r.setFlatCustomVAs(horzRadius, threshStart, threshEnd, r.tint)

	r.opts.Images[0] = mask
	preBlend := r.opts.Blend
	r.opts.Blend = ebiten.BlendLighter
	target.DrawTrianglesShader32(r.vertices[:], r.indices[:], shaderGlowHorz.Load(), &r.opts)
	r.opts.Blend = preBlend
	r.opts.Images[0] = nil
}

// glowDarkHorz is the "negative" version of glowHorz. Instead of using an additive
// blending effect around high luminosity areas, it uses multiplicative blending around
// dark areas.
//
// horzRadius can't exceed 32.
//
// This operation is affected by [Renderer.Tint].
//
// Notice that unlike regular glow effects, dark glows expects threshStart >= threshEnd.
func (r *Renderer) glowDarkHorz(target *ebiten.Image, mask *ebiten.Image, ox, oy, horzRadius, threshStart, threshEnd float32) {
	if threshStart < threshEnd {
		r.Warnings.report(WarnInconsistentRangeOpSkipped, [2]float32{threshStart, threshEnd})
		return
	}
	horzRadius = r.warnClampNonNegArgF32(horzRadius, 32, WarnRadiusClamped)

	srcBounds := mask.Bounds()
	srcWidth, srcHeight := float32(srcBounds.Dx()), float32(srcBounds.Dy())

	r.setDstRectCoords(ox-horzRadius-1.0, oy, ox+float32(srcWidth)+horzRadius+1.0, oy+float32(srcHeight))

	srcMinX, srcMinY := float32(srcBounds.Min.X), float32(srcBounds.Min.Y)
	srcMaxX, srcMaxY := float32(srcBounds.Max.X), float32(srcBounds.Max.Y)
	r.setSrcRectCoords(srcMinX-horzRadius-1, srcMinY, srcMaxX+horzRadius+1, srcMaxY)
	r.setFlatCustomVAs(horzRadius, threshStart, threshEnd, r.tint)

	r.opts.Images[0] = mask
	preBlend := r.opts.Blend
	r.opts.Blend = BlendMultiply
	target.DrawTrianglesShader32(r.vertices[:], r.indices[:], shaderGlowDarkHorz.Load(), &r.opts)
	r.opts.Blend = preBlend
	r.opts.Images[0] = nil
}

// func (r *Renderer) ApplyGlowDark2(target *ebiten.Image, mask *ebiten.Image, ox, oy float32) {}

// func (r *Renderer) ApplyGlowDarkK(target *ebiten.Image, mask *ebiten.Image, ox, oy, threshStart, threshEnd float32, opts KernelOptions) {
// 	r.applyKernel(target, mask, ox, oy, opts, func(downHorzTarget *ebiten.Image) {
// 		r.setFlatCustomVAs(float32(opts.HorzKernel.Radius()), threshStart, threshEnd, r.tint)
// 		preBlend := r.opts.Blend
// 		r.opts.Blend = BlendMultiply
// 		downHorzTarget.DrawTrianglesShader32(r.vertices[:], r.indices[:], shaderDarkHorzGlowKern.Load(), &r.opts)
// 		r.opts.Blend = preBlend

// 	}, true)
// }

// GlowK is a separable multipass glow with fixed radius and optional downscaling.
// See [KernelOptions] for more details.
//
// This operation is affected by [Renderer.Tint].
//
// This function uses the internal offscreen (#0), and if downscaling also (#1).
// Target and mask can be on the same internal atlas.
func (r *Renderer) GlowK(target *ebiten.Image, mask *ebiten.Image, ox, oy float32, threshStart, threshEnd float32, opts KernelOptions) {
	if threshStart > threshEnd {
		r.Warnings.report(WarnInconsistentRangeOpSkipped, [2]float32{threshStart, threshEnd})
		return
	}

	r.applyKernel(target, mask, ox, oy, opts, func(downHorzTarget *ebiten.Image) {
		r.setFlatCustomVAs(threshStart, threshEnd, r.tint, 0)
		downHorzTarget.DrawTrianglesShader32(r.vertices[:], r.indices[:], shaderGlowHorzKern.Load(), &r.opts)
	}, true)
}

// TODO: GlowColor2

// GlowColorK is a color-specific version of [Renderer.GlowK](), where glow
// intensity is determined by color similarity instead of lightness.
//
// This operation is affected by [Renderer.Tint].
//
// This function uses the internal offscreen (#0), and if downscaling also (#1).
// Target and mask can be on the same internal atlas.
func (r *Renderer) GlowColorK(target *ebiten.Image, mask *ebiten.Image, ox, oy float32, rgb [3]float32, threshStart, threshEnd float32, opts KernelOptions) {
	if threshStart > threshEnd {
		r.Warnings.report(WarnInconsistentRangeOpSkipped, [2]float32{threshStart, threshEnd})
		return
	}

	r.applyKernel(target, mask, ox, oy, opts, func(downHorzTarget *ebiten.Image) {
		r.opts.Uniforms["RGB"] = rgb
		r.setFlatCustomVAs(threshStart, threshEnd, r.tint, 0)
		downHorzTarget.DrawTrianglesShader32(r.vertices[:], r.indices[:], shaderGlowColorHorz.Load(), &r.opts)
		clear(r.opts.Uniforms)
	}, true)
}
