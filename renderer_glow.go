package shapes

import (
	"github.com/hajimehoshi/ebiten/v2"
)

// helper for glow effects.
// this operation is affected by [Renderer.Tint].
func (r *Renderer) glowVert(shader *ebiten.Shader, target *ebiten.Image, mask *ebiten.Image, ox, oy, radius, threshStart, threshEnd float32, blend *ebiten.Blend) {
	sox, soy, sw, sh := rectOriginSizeF32(mask.Bounds())
	ceilRadius := ceilF32(radius)
	r.setDstRectCoords(ox, oy-ceilRadius, ox+sw, oy+sh+ceilRadius)
	r.setSrcRectCoords(sox, soy-ceilRadius, sox+sw, soy+sh+ceilRadius)
	r.setFlatCustomVAs(radius, threshStart, threshEnd, r.tint)

	if blend != nil {
		r.opts.Blend, blend = *blend, &r.opts.Blend
	}
	r.opts.Images[0] = mask
	target.DrawTrianglesShader32(r.vertices[:], r.indices[:], shader, &r.opts)
	r.opts.Images[0] = nil
	if blend != nil {
		r.opts.Blend = *blend
	}
}

// helper for glow effects.
// this operation is affected by [Renderer.Tint].
func (r *Renderer) glowHorz(shader *ebiten.Shader, target *ebiten.Image, mask *ebiten.Image, ox, oy, radius, threshStart, threshEnd float32, blend *ebiten.Blend) {
	sox, soy, sw, sh := rectOriginSizeF32(mask.Bounds())
	ceilRadius := ceilF32(radius)
	r.setDstRectCoords(ox-ceilRadius, oy, ox+sw+ceilRadius, oy+sh)
	r.setSrcRectCoords(sox-ceilRadius, soy, sox+sw+ceilRadius, soy+sh)
	r.setFlatCustomVAs(radius, threshStart, threshEnd, r.tint)

	if blend != nil {
		r.opts.Blend, blend = *blend, &r.opts.Blend
	}
	r.opts.Images[0] = mask
	target.DrawTrianglesShader32(r.vertices[:], r.indices[:], shader, &r.opts)
	r.opts.Images[0] = nil
	if blend != nil {
		r.opts.Blend = *blend
	}
}

// Glow2 draws a separable two-pass glow effect. Radiuses can't exceed 32.
//
// threshStart and threshEnd indicate the start luminosity threshold at which the glow
// effect kicks in and the point at which it's fully active. threshStart must be <=
// threshEnd, and the values must be in [0, 1] range. For reference, 0.4 to 0.7 is
// typically a good starting range.
//
// Unlike blurs, glows require you to draw the original image first, and then apply the
// glow effect on top. Glow is always applied with source over blend.
//
// This operation is affected by [Renderer.Tint], but the renderer's current alpha is
// always applied as an effect intensity factor even if the tint is 0.
//
// This function uses an internal offscreen (#0), and target and mask can be on the same
// internal atlas.
//
// See also [Renderer.GlowK]() if fixed radiuses are acceptable.
func (r *Renderer) Glow2(target *ebiten.Image, mask *ebiten.Image, ox, oy, horzRadius, vertRadius, threshStart, threshEnd float32) {
	if mask == nil {
		r.Warnings.report(WarnMissingSourceOpSkipped, mask)
		return
	}
	if threshStart > threshEnd {
		r.Warnings.report(WarnInconsistentRangeOpSkipped, [2]float32{threshStart, threshEnd})
		return
	}

	horzRadius = r.warnClampNonNegArgF32(horzRadius, 32, WarnRadiusClamped)
	vertRadius = r.warnClampNonNegArgF32(vertRadius, 32, WarnRadiusClamped)
	if horzRadius == 0 {
		r.glowVert(shaderGlowVert.Load(), target, mask, ox, oy, vertRadius, threshStart, threshEnd, &ebiten.BlendLighter)
		return
	} else if vertRadius == 0 {
		r.glowHorz(shaderGlowHorz.Load(), target, mask, ox, oy, horzRadius, threshStart, threshEnd, &ebiten.BlendLighter)
		return
	}

	ceilVertRadius := ceilF32(vertRadius)
	w32, h32 := rectSizeF32(mask.Bounds())
	h32 += 2.0 * ceilVertRadius
	tmp, _ := r.getTemp(0, int(w32), int(h32), false)
	preBlend := r.opts.Blend
	r.opts.Blend = ebiten.BlendCopy
	r.glowVert(shaderGlowVert.Load(), tmp, mask, 0, ceilVertRadius, vertRadius, threshStart, threshEnd, nil)

	memo := r.tint
	r.tint = 0.0
	r.opts.Blend = ebiten.BlendLighter
	r.blurHorz(target, tmp, ox, oy-ceilVertRadius, horzRadius)
	r.opts.Blend = preBlend
	r.tint = memo
}

// GlowK is the fixed kernel version of [Renderer.Glow2](). See [KernelOptions] for more
// details.
//
// This operation is affected by [Renderer.Tint], but the renderer's current alpha is
// always applied as an effect intensity factor even if the tint is 0.
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

// GlowDark2 is the "negative" version of [Renderer.Glow2](). Instead of using additive
// blending around high luminosity areas, it uses multiplicative blending around dark
// areas. Unlike regular glow effects, dark glows expects threshStart >= threshEnd.
// Radiuses can't exceed 32.
//
// This operation is affected by [Renderer.Tint], but the renderer's current alpha is
// always applied as an effect intensity factor even if the tint is 0. The effect is
// always applied with 'source over' blend.
//
// This function uses an internal offscreen (#0), and target and mask can be on the same
// internal atlas.
//
// See also [Renderer.GlowDarkK]() if fixed radiuses are acceptable.
func (r *Renderer) GlowDark2(target *ebiten.Image, mask *ebiten.Image, ox, oy float32, horzRadius, vertRadius, threshStart, threshEnd float32) {
	if mask == nil {
		r.Warnings.report(WarnMissingSourceOpSkipped, mask)
		return
	}
	if threshStart < threshEnd {
		r.Warnings.report(WarnInconsistentRangeOpSkipped, [2]float32{threshStart, threshEnd})
		return
	}
	horzRadius = r.warnClampNonNegArgF32(horzRadius, 32, WarnRadiusClamped)
	vertRadius = r.warnClampNonNegArgF32(vertRadius, 32, WarnRadiusClamped)
	if horzRadius == 0 {
		r.glowVert(shaderGlowDarkVert.Load(), target, mask, ox, oy, vertRadius, threshStart, threshEnd, &blendDarkGlow)
		return
	} else if vertRadius == 0 {
		r.glowHorz(shaderGlowDarkHorz.Load(), target, mask, ox, oy, horzRadius, threshStart, threshEnd, &blendDarkGlow)
		return
	}

	ceilVertRadius := ceilF32(vertRadius)
	w32, h32 := rectSizeF32(mask.Bounds())
	h32 += 2.0 * ceilVertRadius
	tmp, _ := r.getTemp(0, int(w32), int(h32), true)
	r.glowVert(shaderGlowDarkVert.Load(), tmp, mask, 0, ceilVertRadius, vertRadius, threshStart, threshEnd, &blendDarkGlow)

	memo := r.tint
	r.tint = 0.0
	preBlend := r.opts.Blend
	r.opts.Blend = ebiten.BlendSourceOver
	r.blurHorz(target, tmp, ox, oy-ceilVertRadius, horzRadius)
	r.opts.Blend = preBlend
	r.tint = memo
}

// GlowDarkK is the fixed kernel version of [Renderer.GlowDark2](). See [KernelOptions]
// for more details.
//
// This operation is affected by [Renderer.Tint], but the renderer's current alpha is
// always applied as an effect intensity factor even if the tint is 0.
//
// This function uses the internal offscreen (#0), and if downscaling also (#1).
// Target and mask can be on the same internal atlas.
//
// TODO: untested
func (r *Renderer) GlowDarkK(target *ebiten.Image, mask *ebiten.Image, ox, oy, threshStart, threshEnd float32, opts KernelOptions) {
	r.applyKernel(target, mask, ox, oy, opts, func(downHorzTarget *ebiten.Image) {
		preBlend := r.opts.Blend
		r.opts.Blend = blendDarkGlow
		r.setFlatCustomVAs(threshStart, threshEnd, r.tint, 0)
		downHorzTarget.DrawTrianglesShader32(r.vertices[:], r.indices[:], shaderGlowDarkHorzKern.Load(), &r.opts)
		r.opts.Blend = preBlend
	}, false)
}

// GlowColor2 is the color-similarity version of [Renderer.Glow2](). Instead of applying
// glow to high luminosity areas, it applies the glow wherever the colors are close to
// the given one.
//
// This operation is affected by [Renderer.Tint], but the renderer's current alpha is
// always applied as an effect intensity factor even if the tint is 0. The effect is
// always applied with 'source over' blend.
//
// This function uses an internal offscreen (#0), and target and mask can be on the same
// internal atlas.
//
// See also [Renderer.GlowColorK]() if fixed radiuses are acceptable.
func (r *Renderer) GlowColor2(target *ebiten.Image, mask *ebiten.Image, ox, oy float32, horzRadius, vertRadius float32, rgb [3]float32, threshStart, threshEnd float32) {
	if mask == nil {
		r.Warnings.report(WarnMissingSourceOpSkipped, mask)
		return
	}
	if threshStart > threshEnd {
		r.Warnings.report(WarnInconsistentRangeOpSkipped, [2]float32{threshStart, threshEnd})
		return
	}
	horzRadius = r.warnClampNonNegArgF32(horzRadius, 32, WarnRadiusClamped)
	vertRadius = r.warnClampNonNegArgF32(vertRadius, 32, WarnRadiusClamped)

	r.opts.Uniforms["RGB"] = rgb
	if horzRadius == 0 {
		r.glowVert(shaderGlowColorHorz.Load(), target, mask, ox, oy, vertRadius, threshStart, threshEnd, &ebiten.BlendLighter)
		clear(r.opts.Uniforms)
		return
	} else if vertRadius == 0 {
		r.glowHorz(shaderGlowColorVert.Load(), target, mask, ox, oy, horzRadius, threshStart, threshEnd, &ebiten.BlendLighter)
		clear(r.opts.Uniforms)
		return
	}

	ceilVertRadius := ceilF32(vertRadius)
	w32, h32 := rectSizeF32(mask.Bounds())
	h32 += 2.0 * ceilVertRadius
	tmp, _ := r.getTemp(0, int(w32), int(h32), false)
	preBlend := r.opts.Blend
	r.opts.Blend = ebiten.BlendCopy
	r.glowVert(shaderGlowColorVert.Load(), tmp, mask, 0, ceilVertRadius, vertRadius, threshStart, threshEnd, nil)
	clear(r.opts.Uniforms)

	memo := r.tint
	r.tint = 0.0
	r.opts.Blend = ebiten.BlendLighter
	r.blurHorz(target, tmp, ox, oy-ceilVertRadius, horzRadius)
	r.opts.Blend = preBlend
	r.tint = memo
}

// GlowColorK is the fixed kernel version of [Renderer.GlowColor2](). See [KernelOptions]
// for more details.
//
// This operation is affected by [Renderer.Tint], but the renderer's current alpha is
// always applied as an effect intensity factor even if the tint is 0.
//
// This function uses the internal offscreen (#0), and if downscaling also (#1).
// Target and mask can be on the same internal atlas.
//
// TODO: untested
func (r *Renderer) GlowColorK(target *ebiten.Image, mask *ebiten.Image, ox, oy float32, rgb [3]float32, threshStart, threshEnd float32, opts KernelOptions) {
	if threshStart > threshEnd {
		r.Warnings.report(WarnInconsistentRangeOpSkipped, [2]float32{threshStart, threshEnd})
		return
	}

	r.applyKernel(target, mask, ox, oy, opts, func(downHorzTarget *ebiten.Image) {
		r.opts.Uniforms["RGB"] = rgb
		r.setFlatCustomVAs(threshStart, threshEnd, r.tint, 0)
		downHorzTarget.DrawTrianglesShader32(r.vertices[:], r.indices[:], shaderGlowColorHorzKern.Load(), &r.opts)
		delete(r.opts.Uniforms, "RGB")
	}, true)
}
