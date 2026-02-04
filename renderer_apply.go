package shapes

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

// Precondition: thickness can't exceed 16.
//
// WARNING: this is a quadratic algorithm on GPU. For large expansions,
// consider [Renderer.ApplyExpansionRect]() or [Renderer.JFMExpansion]()
// instead, but both of those are only useful in specific situations.
func (r *Renderer) ApplyExpansion(target *ebiten.Image, mask *ebiten.Image, ox, oy, thickness float32) {
	if mask == nil {
		r.Warnings.report(WarnMissingSourceOpSkipped, mask)
		return
	}
	if thickness > 16 {
		r.Warnings.report(WarnThicknessClamped, thickness)
		thickness = 16
	} else if thickness < 0 {
		r.Warnings.report(WarnNegativeValueZeroed, thickness)
		thickness = 0
	}

	srcBounds := mask.Bounds()
	srcWidth, srcHeight := float32(srcBounds.Dx()), float32(srcBounds.Dy())
	ht32 := thickness
	dstBounds := target.Bounds()
	dstMinX, dstMinY := float32(dstBounds.Min.X), float32(dstBounds.Min.Y)
	minX, minY := dstMinX+ox-ht32, dstMinY+oy-ht32
	maxX, maxY := dstMinX+ox+srcWidth+ht32, dstMinY+oy+srcHeight+ht32
	r.setDstRectCoords(minX-1, minY-1, maxX+1, maxY+1)

	srcMinX, srcMinY := float32(srcBounds.Min.X), float32(srcBounds.Min.Y)
	srcMaxX, srcMaxY := float32(srcBounds.Max.X), float32(srcBounds.Max.Y)
	r.setSrcRectCoords(srcMinX-ht32-1, srcMinY-ht32-1, srcMaxX+ht32+1, srcMaxY+ht32+1)
	r.setFlatCustomVA0(thickness)

	// draw shader
	r.opts.Images[0] = mask
	target.DrawTrianglesShader(r.vertices[:], r.indices[:], shaderExpansion.Load(), &r.opts)
	r.opts.Images[0] = nil
}

// ApplyExpansionRect performs double pass expansion with a square kernel.
// This is less general but more efficient than [Renderer.ApplyExpansion]().
//
// Precondition: thickness can't exceed 16.
//
// This function uses one internal offscreen (#0), and target and mask
// can be on the same internal atlas.
//
// Cost: two passes, 4*ceil(thickness) samples per pixel (2*ceil(thickness) on each pass).
func (r *Renderer) ApplyExpansionRect(target *ebiten.Image, mask *ebiten.Image, ox, oy, thickness float32) {
	if mask == nil {
		r.Warnings.report(WarnMissingSourceOpSkipped, mask)
		return
	}
	if thickness > 16 {
		r.Warnings.report(WarnThicknessClamped, thickness)
		thickness = 16
	} else if thickness < 0 {
		r.Warnings.report(WarnNegativeValueZeroed, thickness)
		thickness = 0
	}

	// first pass (vert)
	thickCeil := float32(math.Ceil(float64(thickness)))
	sx, sy, sw, sh := rectOriginSize(mask.Bounds())
	temp, _ := r.getTemp(0, sw, sh+int(thickCeil)*2.0, false)
	sx32, sy32, sw32, sh32 := float32(sx), float32(sy), float32(sw), float32(sh)
	memoBlend := r.opts.Blend
	r.opts.Blend = ebiten.BlendCopy
	r.setSrcRectCoords(sx32, sy32-thickCeil, sx32+sw32, sy32+sh32+thickCeil)
	r.setDstRectCoords(0, 0, sw32, sh32+thickCeil*2)
	r.setFlatCustomVA0(thickness)
	r.opts.Images[0] = mask
	temp.DrawTrianglesShader(r.vertices[:], r.indices[:], shaderExpansionVert.Load(), &r.opts)
	r.opts.Images[0] = nil

	// second pass (horz)
	r.opts.Blend = memoBlend
	r.setSrcRectCoords(-thickCeil, 0, sw32+thickCeil, sh32+thickCeil*2.0)
	dx, dy := rectOriginF32(target.Bounds())
	ox += dx
	oy += dy
	r.setDstRectCoords(ox-thickCeil, oy-thickCeil, ox+sw32+thickCeil, oy+sh32+thickCeil)
	r.opts.Images[0] = temp
	target.DrawTrianglesShader(r.vertices[:], r.indices[:], shaderExpansionHorz.Load(), &r.opts)
	r.opts.Images[0] = nil
}

// Precondition: thickness can't exceed 16.
//
// WARNING: this is a quadratic algorithm on GPU. For large erosions,
// consider [Renderer.JFMErosion]() instead.
func (r *Renderer) ApplyErosion(target *ebiten.Image, mask *ebiten.Image, ox, oy, thickness float32) {
	if mask == nil {
		r.Warnings.report(WarnMissingSourceOpSkipped, mask)
		return
	}
	if thickness > 16 {
		r.Warnings.report(WarnThicknessClamped, thickness)
		thickness = 16
	} else if thickness < 0 {
		r.Warnings.report(WarnNegativeValueZeroed, thickness)
		thickness = 0
		return
	}

	srcBounds := mask.Bounds()
	srcWidth, srcHeight := float32(srcBounds.Dx()), float32(srcBounds.Dy())
	dstBounds := target.Bounds()
	dstMinX, dstMinY := float32(dstBounds.Min.X), float32(dstBounds.Min.Y)
	minX, minY := dstMinX+ox, dstMinY+oy
	maxX, maxY := dstMinX+ox+srcWidth, dstMinY+oy+srcHeight
	r.setDstRectCoords(minX-1, minY-1, maxX+1.0, maxY+1.0)

	srcMinX, srcMinY := float32(srcBounds.Min.X), float32(srcBounds.Min.Y)
	srcMaxX, srcMaxY := float32(srcBounds.Max.X), float32(srcBounds.Max.Y)
	r.setSrcRectCoords(srcMinX-1, srcMinY-1, srcMaxX+1.0, srcMaxY+1.0)
	r.setFlatCustomVA0(thickness)

	// draw shader
	r.opts.Images[0] = mask
	target.DrawTrianglesShader(r.vertices[:], r.indices[:], shaderErosion.Load(), &r.opts)
	r.opts.Images[0] = nil
}

// Precondition: thickness can't exceed 32.
//
// WARNING: this is a quadratic algorithm on GPU. For large outlines,
// consider [Renderer.JFMOutline]() instead.
func (r *Renderer) ApplyOutline(target *ebiten.Image, mask *ebiten.Image, ox, oy, thickness float32) {
	if mask == nil {
		r.Warnings.report(WarnMissingSourceOpSkipped, mask)
		return
	}
	if thickness > 32 {
		r.Warnings.report(WarnThicknessClamped, thickness)
		thickness = 32
	} else if thickness <= 0 {
		if thickness < 0 {
			r.Warnings.report(WarnNegativeValueOpSkipped, thickness)
		}
		return
	}

	srcBounds := mask.Bounds()
	srcWidth, srcHeight := float32(srcBounds.Dx()), float32(srcBounds.Dy())
	ht32 := thickness / 2.0
	dstBounds := target.Bounds()
	dstMinX, dstMinY := float32(dstBounds.Min.X), float32(dstBounds.Min.Y)
	minX, minY := dstMinX+ox-ht32, dstMinY+oy-ht32
	maxX, maxY := dstMinX+ox+srcWidth+ht32, dstMinY+oy+srcHeight+ht32
	r.setDstRectCoords(minX-1, minY-1, maxX+1, maxY+1)

	srcMinX, srcMinY := float32(srcBounds.Min.X), float32(srcBounds.Min.Y)
	srcMaxX, srcMaxY := float32(srcBounds.Max.X), float32(srcBounds.Max.Y)
	r.setSrcRectCoords(srcMinX-ht32-1, srcMinY-ht32-1, srcMaxX+ht32+1.0, srcMaxY+ht32+1.0)
	r.setFlatCustomVA0(thickness)

	// draw shader
	r.opts.Images[0] = mask
	target.DrawTrianglesShader(r.vertices[:], r.indices[:], shaderOutline.Load(), &r.opts)
	r.opts.Images[0] = nil
}

func (r *Renderer) ApplyHardShadow(target *ebiten.Image, mask *ebiten.Image, ox, oy, xOffset, yOffset float32, clamping Clamping) {
	if mask == nil {
		r.Warnings.report(WarnMissingSourceOpSkipped, mask)
		return
	}

	dstBounds, srcBounds := target.Bounds(), mask.Bounds()
	srcWidth, srcHeight := float32(srcBounds.Dx()), float32(srcBounds.Dy())
	dstMinX, dstMinY := float32(dstBounds.Min.X), float32(dstBounds.Min.Y)
	leftOff, topOff := min(0, xOffset), min(0, yOffset)
	rightOff, bottomOff := max(0, xOffset), max(0, yOffset)
	minX, minY := dstMinX+ox+leftOff, dstMinY+oy+topOff
	maxX, maxY := dstMinX+srcWidth+ox+rightOff, dstMinY+srcHeight+oy+bottomOff
	r.setDstRectCoords(minX, minY, maxX, maxY)

	srcMinX, srcMinY := float32(srcBounds.Min.X), float32(srcBounds.Min.Y)
	srcMaxX, srcMaxY := float32(srcBounds.Max.X), float32(srcBounds.Max.Y)
	r.setSrcRectCoords(srcMinX+leftOff, srcMinY+topOff, srcMaxX+rightOff, srcMaxY+bottomOff)
	r.setFlatCustomVAs(xOffset, yOffset, float32(clamping), 0)

	// draw shader
	r.opts.Images[0] = mask
	target.DrawTrianglesShader(r.vertices[:], r.indices[:], shaderHardShadow.Load(), &r.opts)
	r.opts.Images[0] = nil
}

// ApplyShadow is a variant of [Renderer.ApplyBlur]() with offsets and clamping.
// The same performance considerations apply.
//
// The function panics if the radius exceeds 16.
func (r *Renderer) ApplyShadow(target *ebiten.Image, mask *ebiten.Image, ox, oy, xOffset, yOffset, radius float32, clamping Clamping) {
	if radius > 16 {
		r.Warnings.report(WarnRadiusClamped, radius)
		radius = 16
	} else if radius <= 0 {
		if radius < 0 {
			r.Warnings.report(WarnNegativeValueOpSkipped, radius)
		}
		return
	}

	dstBounds, srcBounds := target.Bounds(), mask.Bounds()
	srcWidth, srcHeight := float32(srcBounds.Dx()), float32(srcBounds.Dy())
	dstMinX, dstMinY := float32(dstBounds.Min.X), float32(dstBounds.Min.Y)
	leftOff, topOff := min(0, xOffset), min(0, yOffset)
	rightOff, bottomOff := max(0, xOffset), max(0, yOffset)
	topR32, bottomR32, leftR32, rightR32 := radius, radius, radius, radius
	if clamping&ClampBottom != 0 {
		bottomR32 = 0
	}
	if clamping&ClampTop != 0 {
		topR32 = 0
	}
	if clamping&ClampLeft != 0 {
		leftR32 = 0
	}
	if clamping&ClampRight != 0 {
		rightR32 = 0
	}

	minX, minY := dstMinX+ox+leftOff-leftR32, dstMinY+oy+topOff-topR32
	maxX, maxY := dstMinX+srcWidth+ox+rightOff+rightR32, dstMinY+srcHeight+oy+bottomOff+bottomR32
	r.setDstRectCoords(minX, minY, maxX, maxY)

	srcMinX, srcMinY := float32(srcBounds.Min.X), float32(srcBounds.Min.Y)
	srcMaxX, srcMaxY := float32(srcBounds.Max.X), float32(srcBounds.Max.Y)
	r.setSrcRectCoords(srcMinX+leftOff-leftR32, srcMinY+topOff-topR32, srcMaxX+rightOff+rightR32, srcMaxY+bottomOff+bottomR32)
	r.setFlatCustomVAs(xOffset, yOffset, radius, float32(clamping))

	// draw shader
	r.opts.Images[0] = mask
	target.DrawTrianglesShader(r.vertices[:], r.indices[:], shaderShadow.Load(), &r.opts)
	r.opts.Images[0] = nil
}

func (r *Renderer) ApplyZoomShadow(target *ebiten.Image, mask *ebiten.Image, ox, oy, xOffset, yOffset, zoom float32, clamping Clamping) {
	if zoom < 1.0 || zoom > 16.0 {
		r.Warnings.report(WarnInvalidZoomClamped, zoom)
		zoom = clamp(zoom, 1.0, 16.0)
	}

	dstBounds, srcBounds := target.Bounds(), mask.Bounds()
	srcWidth, srcHeight := float32(srcBounds.Dx()), float32(srcBounds.Dy())
	dstMinX, dstMinY := float32(dstBounds.Min.X), float32(dstBounds.Min.Y)
	leftOff, topOff := min(0, xOffset), min(0, yOffset)
	rightOff, bottomOff := max(0, xOffset), max(0, yOffset)
	var topCut, leftCut, bottomCut, rightCut float32
	topPad, leftPad := srcHeight*0.5*(zoom-1.0), srcWidth*0.5*(zoom-1.0)
	bottomPad, rightPad := topPad, leftPad
	if clamping&ClampBottom != 0 {
		bottomCut = bottomPad / zoom
		bottomPad = 0
	}
	if clamping&ClampTop != 0 {
		topCut = topPad / zoom
		topPad = 0
	}
	if clamping&ClampLeft != 0 {
		leftCut = leftPad / zoom
		leftPad = 0
	}
	if clamping&ClampRight != 0 {
		rightCut = rightPad / zoom
		rightPad = 0
	}

	minX, minY := dstMinX+ox+leftOff-leftPad, dstMinY+oy+topOff-topPad
	maxX, maxY := dstMinX+srcWidth+ox+rightOff+rightPad, dstMinY+srcHeight+oy+bottomOff+bottomPad
	r.setDstRectCoords(minX, minY, maxX, maxY)

	srcMinX, srcMinY := float32(srcBounds.Min.X), float32(srcBounds.Min.Y)
	srcMaxX, srcMaxY := float32(srcBounds.Max.X), float32(srcBounds.Max.Y)
	r.setSrcRectCoords(srcMinX+leftOff/zoom-leftCut, srcMinY+topOff/zoom-topCut, srcMaxX+rightOff/zoom-rightCut, srcMaxY+bottomOff/zoom-bottomCut)
	r.setFlatCustomVAs(xOffset/zoom, yOffset/zoom, float32(clamping), zoom)

	// draw shader
	r.opts.Images[0] = mask
	target.DrawTrianglesShader(r.vertices[:], r.indices[:], shaderZoomShadow.Load(), &r.opts)
	r.opts.Images[0] = nil
}

// ApplyGlow draws a horizontal glow effect for the given mask into the target, at the
// given coordinates. The effect mix intensity is determined by the renderer's color alphas.
//
// Regarding the advanced control parameters:
//   - threshStart and threshEnd indicate the start luminosity threshold at which the glow
//     effect kicks in and the point at which it's fully active. threshStart must be <=
//     threshEnd, and the values must be in [0, 1] range.
//   - colorMix controls the glow's color. If 0, the glow color will be determined fully
//     by the renderer's vertex colors. If 1, the glow color will be determined by the original
//     mask colors. Any values in between will lead to linear interpolation.
//
// For reference thresholds, 0.4 to 0.7 is a good general default range.
//
// Notice that this effect uses an internal offscreen (#0) and two passes. Target and mask
// can be on the same internal atlas. Neither horzRadius nor vertRadius can exceed 16.
func (r *Renderer) ApplyGlow2(target *ebiten.Image, mask *ebiten.Image, ox, oy, horzRadius, vertRadius, threshStart, threshEnd, colorMix float32) {
	if threshStart > threshEnd {
		r.Warnings.report(WarnInconsistentRangeOpSkipped, [2]float32{threshStart, threshEnd})
		return
	}
	if horzRadius > 16 {
		r.Warnings.report(WarnRadiusClamped, horzRadius)
		horzRadius = 16
	}
	if vertRadius > 16 {
		r.Warnings.report(WarnRadiusClamped, vertRadius)
		vertRadius = 16
	}
	if colorMix < 0 || colorMix > 1 {
		r.Warnings.report(WarnInvalidColorMixClamped, colorMix)
		colorMix = clamp(colorMix, 0, 1)
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
	r.setFlatCustomVAs(vertRadius, threshStart, threshEnd, 1.0)

	// first pass (threshold + vertical blur)
	r.opts.Images[0] = mask
	preBlend := r.opts.Blend
	r.opts.Blend = ebiten.BlendCopy
	tmp.DrawTrianglesShader(r.vertices[:], r.indices[:], shaderGlowFirstPass.Load(), &r.opts)
	r.opts.Images[0] = nil

	// second pass
	r.opts.Blend = ebiten.BlendLighter
	r.ApplyHorzBlur(target, tmp, ox, oy-vertRadius-1.0, horzRadius, colorMix)
	r.opts.Blend = preBlend
}

// ApplyHorzGlow draws a horizontal glow effect for the given mask into the target, at the
// given coordinates. See [Renderer.ApplyGlow]() for additional documentation. Comparedto
// Renderer.ApplyGlow, this effect only applies the glow horizontally and it's much cheaper,
// requiring no offscreen and a single pass.
//
// horzRadius can't exceed 16.
func (r *Renderer) ApplyHorzGlow(target *ebiten.Image, mask *ebiten.Image, ox, oy, horzRadius, threshStart, threshEnd, colorMix float32) {
	if threshStart > threshEnd {
		r.Warnings.report(WarnInconsistentRangeOpSkipped, [2]float32{threshStart, threshEnd})
		return
	}
	if horzRadius > 16 {
		r.Warnings.report(WarnRadiusClamped, horzRadius)
		horzRadius = 16
	}

	srcBounds := mask.Bounds()
	srcWidth, srcHeight := float32(srcBounds.Dx()), float32(srcBounds.Dy())

	r.setDstRectCoords(ox-horzRadius-1.0, oy, ox+float32(srcWidth)+horzRadius+1.0, oy+float32(srcHeight))

	srcMinX, srcMinY := float32(srcBounds.Min.X), float32(srcBounds.Min.Y)
	srcMaxX, srcMaxY := float32(srcBounds.Max.X), float32(srcBounds.Max.Y)
	r.setSrcRectCoords(srcMinX-horzRadius-1, srcMinY, srcMaxX+horzRadius+1, srcMaxY)
	r.setFlatCustomVAs(horzRadius, threshStart, threshEnd, colorMix)

	r.opts.Images[0] = mask
	preBlend := r.opts.Blend
	r.opts.Blend = ebiten.BlendLighter
	target.DrawTrianglesShader(r.vertices[:], r.indices[:], shaderHorzGlow.Load(), &r.opts)
	r.opts.Blend = preBlend
	r.opts.Images[0] = nil
}

// ApplyDarkHorzGlow is the "negative" version of [Renderer.ApplyHorzGlow](). Instead of
// using an additive blending effect around high luminosity areas, it uses multiplicative
// blending around dark areas.
//
// horzRadius can't exceed 16.
//
// Notice that unlike regular glow effects, dark glows expects threshStart >= threshEnd.
func (r *Renderer) ApplyDarkHorzGlow(target *ebiten.Image, mask *ebiten.Image, ox, oy, horzRadius, threshStart, threshEnd, colorMix float32) {
	if threshStart < threshEnd {
		r.Warnings.report(WarnInconsistentRangeOpSkipped, [2]float32{threshStart, threshEnd})
		return
	}
	if horzRadius > 16 {
		r.Warnings.report(WarnRadiusClamped, horzRadius)
		horzRadius = 16
	}

	srcBounds := mask.Bounds()
	srcWidth, srcHeight := float32(srcBounds.Dx()), float32(srcBounds.Dy())

	r.setDstRectCoords(ox-horzRadius-1.0, oy, ox+float32(srcWidth)+horzRadius+1.0, oy+float32(srcHeight))

	srcMinX, srcMinY := float32(srcBounds.Min.X), float32(srcBounds.Min.Y)
	srcMaxX, srcMaxY := float32(srcBounds.Max.X), float32(srcBounds.Max.Y)
	r.setSrcRectCoords(srcMinX-horzRadius-1, srcMinY, srcMaxX+horzRadius+1, srcMaxY)
	r.setFlatCustomVAs(horzRadius, threshStart, threshEnd, colorMix)

	r.opts.Images[0] = mask
	preBlend := r.opts.Blend
	r.opts.Blend = BlendMultiply
	//r.opts.Blend = BlendSubtract // also possible with a shader flag, but multiply feels more natural
	target.DrawTrianglesShader(r.vertices[:], r.indices[:], shaderDarkHorzGlow.Load(), &r.opts)
	r.opts.Blend = preBlend
	r.opts.Images[0] = nil
}

// ApplyGlowK is the multipass downscaling version of [Renderer.ApplyGlow]().
// See [Renderer.ApplyBlurKernel]() for further docs and context.
//
// This function uses the internal offscreen (#0), and if downscaling also (#1).
// Target and mask can be on the same internal atlas.
func (r *Renderer) ApplyGlowK(target *ebiten.Image, mask *ebiten.Image, ox, oy float32, threshStart, threshEnd float32, opts KernelOptions) {
	if threshStart > threshEnd {
		r.Warnings.report(WarnInconsistentRangeOpSkipped, [2]float32{threshStart, threshEnd})
		return
	}

	r.applyKernel(target, mask, ox, oy, opts, func(downHorzTarget *ebiten.Image) {
		r.setFlatCustomVAs(threshStart, threshEnd, opts.ColorMix, 0)
		downHorzTarget.DrawTrianglesShader(r.vertices[:], r.indices[:], shaderHorzGlowKern.Load(), &r.opts)
	}, true)
}

// ApplyColorGlowK is a color-specific version of [Renderer.ApplyGlowK](), where glow
// intensity is determined by color similarity instead of lightness.
//
// This function uses the internal offscreen (#0), and if downscaling also (#1).
// Target and mask can be on the same internal atlas.
func (r *Renderer) ApplyColorGlowK(target *ebiten.Image, mask *ebiten.Image, ox, oy float32, rgb [3]float32, threshStart, threshEnd float32, opts KernelOptions) {
	if threshStart > threshEnd {
		r.Warnings.report(WarnInconsistentRangeOpSkipped, [2]float32{threshStart, threshEnd})
		return
	}

	r.applyKernel(target, mask, ox, oy, opts, func(downHorzTarget *ebiten.Image) {
		r.opts.Uniforms["RGB"] = rgb
		r.setFlatCustomVAs(threshStart, threshEnd, opts.ColorMix, 0)
		downHorzTarget.DrawTrianglesShader(r.vertices[:], r.indices[:], shaderHorzColorGlow.Load(), &r.opts)
		clear(r.opts.Uniforms)
	}, true)
}

// Internal function used by ApplyBlurK, ApplyGlowK and ApplyColorGlowK. It downscales the
// mask, applies a custom horizontal kernel shader, then a standard vertical blur shader, and
// upscales the result back, optionally with a BlendLighter blend. At invokeShader, KernelLen
// and Kernel uniforms have been set, as well as the downscaled source image, but other uniforms
// and custom VAs have to be set during invocation.
//
// When downscaling is used, this function uses two internal offscreens (#0, #1), and target and
// mask can be on the same internal atlas.
func (r *Renderer) applyKernel(target *ebiten.Image, mask *ebiten.Image, ox, oy float32, opts KernelOptions, invokeShader func(downHorzTarget *ebiten.Image), lighterBlend bool) {
	if !opts.Downscaling.valid() {
		panic("invalid downscaling value")
	}
	if !opts.HorzKernel.valid() || !opts.VertKernel.valid() {
		panic("invalid GaussKernel value")
	}
	if mask == nil {
		r.Warnings.report(WarnMissingSourceOpSkipped, mask)
		return
	}

	if opts.Downscaling == DownscaleNone {
		r.applyKernelDirect(target, mask, ox, oy, opts, invokeShader, lighterBlend)
		return
	}

	// measures
	df := opts.Downscaling.Factor()
	maskW64, maskH64 := rectSizeF64(mask.Bounds())
	downW64, downH64 := maskW64/float64(df), maskH64/float64(df)
	halfHorzMargin, halfVertMargin := float64(opts.HorzKernel.Radius()), float64(opts.VertKernel.Radius())
	dkernW64, dkernH64 := downW64+halfHorzMargin+halfHorzMargin, downH64+halfVertMargin+halfVertMargin

	// get offscreens and smart clears
	downImgWidth, downImgHeight := math.Ceil(downW64)+2, math.Ceil(downH64)+2
	dkernImgWidth, dkernImgHeight := math.Ceil(dkernW64)+2, math.Ceil(dkernH64)+2
	dkern, _ := r.getTemp(0, int(dkernImgWidth), int(dkernImgHeight), false) // get first as the biggest offscreen
	down, _ := r.getTemp(0, int(downImgWidth), int(downImgHeight), false)    // shared with dkern
	dkernHorz, _ := r.getTemp(1, int(dkernImgWidth), int(downImgHeight), false)
	preBlend := r.opts.Blend
	r.opts.Blend = ebiten.BlendClear
	r.StrokeIntRect(down, down.Bounds(), 0, 2)
	r.DrawIntRect(dkern, clockwiseRightBorder(dkern.Bounds(), 1)) // *
	r.DrawIntRect(dkern, bottomBorder(dkern.Bounds(), 1))
	r.DrawIntRect(dkernHorz, clockwiseRightBorder(dkernHorz.Bounds(), 1))
	r.DrawIntRect(dkernHorz, bottomBorder(dkernHorz.Bounds(), 1))
	// * Notice that technically dkern content could be overwritten by operations
	//   on 'down' after the clear, but since kernels can't be zero and 'down' already
	//   has 1 pixel margins, this won't happen in practice. Otherwise the clear should
	//   be delayed until after the horz kernel application

	// downscaling
	r.opts.Blend = ebiten.BlendCopy
	df32 := float32(df)
	r.Scale(down, mask, 1, 1, 1.0/df32, opts.Scaling)

	// apply effect
	r.applyKernelOp(down, dkern, dkernHorz, dkernW64, dkernH64, downW64, downH64, opts, invokeShader)

	// upscale
	if lighterBlend {
		r.opts.Blend = ebiten.BlendLighter
	} else {
		r.opts.Blend = preBlend
	}
	fx, fy := ox+-df32-float32(halfHorzMargin)*df32, oy+-df32-float32(halfVertMargin)*df32
	r.Scale(target, dkern, fx, fy, df32, opts.Scaling)
	if lighterBlend {
		r.opts.Blend = preBlend
	}
}

func (r *Renderer) applyKernelDirect(target, mask *ebiten.Image, ox, oy float32, opts KernelOptions, invokeShader func(horzTarget *ebiten.Image), lighterBlend bool) {
	horzKernelLen := opts.HorzKernel.Size()
	ceilHRadius := float32(horzKernelLen)
	ox32, oy32, w32, h32 := rectOriginSizeF32(mask.Bounds())
	w32 += float32(horzKernelLen + horzKernelLen)
	tmp, _ := r.getTemp(0, int(w32), int(h32), true) // TODO: remove clear flag after debug
	preBlend := r.opts.Blend

	// apply horz kern shader
	r.setDstRectCoords(0, 0, w32, h32)
	//ox32, oy32 = 0.0, 0.0
	sx := ox32 - ceilHRadius
	r.setSrcRectCoords(sx, oy32, sx+w32, oy32+h32)
	r.opts.Blend = ebiten.BlendCopy
	r.opts.Images[0] = mask
	r.opts.Uniforms["KernelLen"] = opts.HorzKernel.Size()
	r.opts.Uniforms["Kernel"] = gaussKernels[opts.HorzKernel]
	invokeShader(tmp) // set VAs, more uniforms, invoke shader and clear(r.opts.Uniforms) if needed

	dox, doy := rectOriginF32(target.Bounds())
	ceilVRadius := float32(opts.VertKernel.Radius())
	dx := dox + ox - ceilHRadius
	r.setDstRectCoords(dx, doy+oy-ceilVRadius, dx+w32, doy+oy+h32+ceilVRadius)
	r.setSrcRectCoords(0, -ceilVRadius, w32, h32+ceilVRadius)

	r.opts.Blend = preBlend
	if lighterBlend {
		r.opts.Blend = ebiten.BlendLighter
	}
	r.opts.Uniforms["KernelLen"] = opts.VertKernel.Size()
	r.opts.Uniforms["Kernel"] = gaussKernels[opts.VertKernel]
	r.opts.Images[0] = tmp
	target.DrawTrianglesShader(r.vertices[:], r.indices[:], shaderVertBlurKern.Load(), &r.opts)
	r.opts.Images[0] = nil
	clear(r.opts.Uniforms)
	r.opts.Blend = preBlend
}

func (r *Renderer) applyKernelOp(down, dkern, dkernHorz *ebiten.Image, dkernW64, dkernH64, downW64, downH64 float64, opts KernelOptions, invokeShader func(downHorzTarget *ebiten.Image)) {
	halfHorzMargin, halfVertMargin := float64(opts.HorzKernel.Radius()), float64(opts.VertKernel.Radius())

	// apply horz kern shader
	r.setDstRectCoords(0, 0, float32(dkernW64)+2, float32(downH64)+2)
	r.setSrcRectCoords(float32(-halfHorzMargin), float32(0), float32(downW64+halfHorzMargin)+2, float32(downH64)+2)
	r.opts.Blend = ebiten.BlendCopy
	r.opts.Images[0] = down
	r.opts.Uniforms["KernelLen"] = opts.HorzKernel.Size()
	r.opts.Uniforms["Kernel"] = gaussKernels[opts.HorzKernel]
	invokeShader(dkernHorz) // set VAs, more uniforms, invoke shader and clear(r.opts.Uniforms) if needed

	// apply vert blur kern
	r.opts.Uniforms["KernelLen"] = opts.VertKernel.Size()
	r.opts.Uniforms["Kernel"] = gaussKernels[opts.VertKernel]
	r.setDstRectCoords(0, 0, float32(dkernW64)+2, float32(dkernH64)+2)
	r.setSrcRectCoords(0, float32(-halfVertMargin), float32(dkernW64)+2, float32(downH64+halfVertMargin)+2)
	r.opts.Images[0] = dkernHorz
	dkern.DrawTrianglesShader(r.vertices[:], r.indices[:], shaderVertBlurKern.Load(), &r.opts)
	r.opts.Images[0] = nil
	clear(r.opts.Uniforms)
}

func (r *Renderer) ApplyScanlinesSharp(target *ebiten.Image, darkThick, clearThick int, intensity, offset float32) {
	r.setFlatCustomVAs(float32(darkThick), float32(clearThick), intensity, offset)
	r.DrawShader(target, 0, 0, shaderScanlinesSharp.Load())
}

func (r *Renderer) ApplyWaveLines(target *ebiten.Image, lineThick, minFillRate, maxFillRate, linesPerOsc, offset float32, dirRadians float64) {
	if minFillRate > maxFillRate {
		r.Warnings.report(WarnInconsistentRangeOpSkipped, [2]float32{minFillRate, maxFillRate})
	}
	if minFillRate < 0 {
		r.Warnings.report(WarnInvalidRateClamped, minFillRate)
		minFillRate = 0
	}
	if maxFillRate > 1.0 {
		r.Warnings.report(WarnInvalidRateClamped, maxFillRate)
		maxFillRate = 1.0
	}
	if maxFillRate == 0 {
		return
	}

	minFillThick := minFillRate * lineThick
	maxFillThick := maxFillRate * lineThick
	waveLen := linesPerOsc * lineThick
	r.opts.Uniforms["Offset"] = float32(math.Mod(float64(offset), float64(waveLen)))
	drs, drc := math.Sincos(dirRadians)
	hypot := math.Hypot(drs, drc)
	drs, drc = drs/hypot, drc/hypot
	r.opts.Uniforms["DirRadsSin"] = float32(drs)
	r.opts.Uniforms["DirRadsCos"] = float32(drc)
	r.setFlatCustomVAs(lineThick, minFillThick, maxFillThick, waveLen)
	r.DrawShader(target, 0, 0, shaderWaveLines.Load())
	clear(r.opts.Uniforms)
}
