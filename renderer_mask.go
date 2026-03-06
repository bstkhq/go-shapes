package shapes

import (
	"github.com/hajimehoshi/ebiten/v2"
)

// Mask draws 'source' over 'target' using 'mask' as an alpha mask. If the source
// and mask sizes are different, the mask will be adjusted to fit the source.
// For manual mask placement, see [Renderer.MaskAt]() instead.
//
// Supported flags: [Bilinear] (in destination space for mask), [Dithered].
func (r *Renderer) Mask(target, source, mask *ebiten.Image, ox, oy float32, flags ...Flag) {
	bilinear, dither := r.readFlags(flags...)
	r.opts.Uniforms["Bilinear"] = mapBool(bilinear, 0, 1)
	r.opts.Uniforms["Dither"] = mapBool(dither, 0, 1)
	if dither {
		r.loadBlueNoiseAt(2)
	}

	srcOX, srcOY, srcWidthF32, srcHeightF32 := rectOriginSizeF32(source.Bounds())
	r.setDstRectCoords(ox, oy, ox+srcWidthF32, oy+srcHeightF32)
	r.setSrcRectCoords(srcOX, srcOY, srcOX+srcWidthF32, srcOY+srcHeightF32)

	maskWidthF32, maskHeightF32 := rectSizeF32(mask.Bounds())
	r.setFlatCustomVAs01(maskWidthF32/srcWidthF32, maskHeightF32/srcHeightF32)
	r.opts.Images[0] = source
	r.opts.Images[1] = mask
	target.DrawTrianglesShader(r.vertices[:], r.indices[:], shaderMask.Load(), &r.opts)
	r.opts.Images[0] = nil
	r.opts.Images[1] = nil
	r.opts.Images[2] = nil
	clear(r.opts.Uniforms)
}

// MaskAt draws 'source' over 'target' using 'mask' as an alpha mask at the given position.
// If you want the mask to be fit to the source instead, see [Renderer.Mask](). For especialized
// masking methods with predefined shapes, see [Renderer.MaskHorz]() and [Renderer.MaskCircle]().
//
// Supported flags: [Bilinear], [Dithered].
func (r *Renderer) MaskAt(target, source, mask *ebiten.Image, ox, oy, oxMask, oyMask float32, flags ...Flag) {
	bilinear, dither := r.readFlags(flags...)
	r.opts.Uniforms["Bilinear"] = mapBool(bilinear, 0, 1)
	r.opts.Uniforms["Dither"] = mapBool(dither, 0, 1)
	if dither {
		r.loadBlueNoiseAt(2)
	}

	srcOX, srcOY, srcWidthF32, srcHeightF32 := rectOriginSizeF32(source.Bounds())
	r.setDstRectCoords(ox, oy, ox+srcWidthF32, oy+srcHeightF32)
	r.setSrcRectCoords(srcOX, srcOY, srcOX+srcWidthF32, srcOY+srcHeightF32)

	r.setFlatCustomVAs01(ox-oxMask, oy-oyMask)
	r.opts.Images[0] = source
	r.opts.Images[1] = mask
	target.DrawTrianglesShader(r.vertices[:], r.indices[:], shaderMaskAt.Load(), &r.opts)
	r.opts.Images[0] = nil
	r.opts.Images[1] = nil
	r.opts.Images[2] = nil
	clear(r.opts.Uniforms)
}

// MaskThreshold draws source into target at the given position, using mask to hide
// the pixels where reveal < mask.alpha.
//
// For example, if a mask goes from transparent to opaque, left to right, the source will
// start appearing from left to right as the reveal threshold increases from 0 to 1.
//
// If source and mask sizes differ, the mask is adjusted like in [Renderer.Mask]().
//
// Supported flags: [Bilinear].
func (r *Renderer) MaskThreshold(target, source, mask *ebiten.Image, reveal, ox, oy float32, flags ...Flag) {
	bilinear, dither := r.readFlags(flags...)
	r.opts.Uniforms["Bilinear"] = mapBool(bilinear, 0, 1)
	if dither {
		r.Warnings.report(WarnInvalidFlag, Dithered)
	}

	srcOX, srcOY, srcWidthF32, srcHeightF32 := rectOriginSizeF32(source.Bounds())
	r.setDstRectCoords(ox, oy, ox+srcWidthF32, oy+srcHeightF32)
	r.setSrcRectCoords(srcOX, srcOY, srcOX+srcWidthF32, srcOY+srcHeightF32)

	maskWidthF32, maskHeightF32 := rectSizeF32(mask.Bounds())
	r.setFlatCustomVAs(maskWidthF32/srcWidthF32, maskHeightF32/srcHeightF32, reveal, 0.0)
	r.opts.Images[0] = source
	r.opts.Images[1] = mask
	target.DrawTrianglesShader(r.vertices[:], r.indices[:], shaderMaskThreshold.Load(), &r.opts)
	r.opts.Images[0] = nil
	r.opts.Images[1] = nil
	clear(r.opts.Uniforms)
}

// MaskHorz is a specialized form of [Renderer.MaskAt]() that draws source over target
// with an horizontal alpha fade between inX, outX.
//
// Supported flags: [Bilinear], [Dithered].
func (r *Renderer) MaskHorz(target, source *ebiten.Image, ox, oy, inX, outX float32, flags ...Flag) {
	bilinear, dither := r.readFlags(flags...)
	r.opts.Uniforms["Bilinear"] = mapBool(bilinear, 0, 1)
	r.opts.Uniforms["Dither"] = mapBool(dither, 0, 1)
	if dither {
		r.loadBlueNoiseAt(1)
	}

	tox, _ := rectOriginF32(target.Bounds())
	lo, hi, ltr := inX, outX, float32(1.0)
	if hi < lo {
		lo, hi = hi, lo
		ltr = 0.0
	}
	r.setFlatCustomVAs(lo-tox, hi-tox, ltr, 0.0)
	r.DrawImgShader(target, source, ox, oy, NoMargins, shaderMaskHorz.Load())

	r.opts.Images[1] = nil
	clear(r.opts.Uniforms)
}

// MaskCircle is a specialized form of [Renderer.MaskAt]() that draws source over target
// with a circular alpha fade between hardRadius and hardRadius + softEdge.
//
// The radial fade is centered at the given circOrigin relative to the source, typically
// [CTR].
//
// Supported flags: [Bilinear], [Dithered].
//
// TODO: make arguments ox, oy, circOrigin Origin?
func (r *Renderer) MaskCircle(target, source *ebiten.Image, cx, cy, srcOffsetX, srcOffsetY, hardRadius, softEdge float32, flags ...Flag) {
	if softEdge < 0.0 {
		hardRadius += softEdge
		softEdge = min(-softEdge, hardRadius-softEdge)
		if softEdge < 0 {
			softEdge = 0
		}
	}
	if hardRadius < 0.0 {
		r.Warnings.report(WarnNegativeValueZeroed, hardRadius)
		hardRadius = 0.0
	}
	if hardRadius == 0 && softEdge == 0 && r.hasSkippableBlend() {
		return
	}

	bilinear, dither := r.readFlags(flags...)
	r.opts.Uniforms["Bilinear"] = mapBool(bilinear, 0, 1)
	r.opts.Uniforms["Dither"] = mapBool(dither, 0, 1)
	if dither {
		r.loadBlueNoiseAt(1)
	}

	srcOX, srcOY, srcWidthF32, srcHeightF32 := rectOriginSizeF32(source.Bounds())
	ox, oy := cx-srcWidthF32/2.0+srcOffsetX, cy-srcHeightF32/2.0+srcOffsetY
	r.setDstRectCoords(ox, oy, ox+srcWidthF32, oy+srcHeightF32)
	r.setSrcRectCoords(srcOX, srcOY, srcOX+srcWidthF32, srcOY+srcHeightF32)

	r.opts.Images[0] = source
	tox, toy := rectOriginF32(target.Bounds())
	r.setFlatCustomVAs(cx-tox, cy-toy, hardRadius, softEdge)
	target.DrawTrianglesShader(r.vertices[:], r.indices[:], shaderMaskCircle.Load(), &r.opts)
	r.opts.Images[0] = nil
}

// Pattern type for [Renderer.DrawAlphaMaskCirc]().
type AlphaMaskPattern int

const (
	MaskPatternDefault AlphaMaskPattern = iota // particles

	MaskPatternFlare       // lines, elliptical, flare
	MaskPatternEllipseCuts // modern elliptical cuts
	MaskPatternCircMesh    // circular mesh
	MaskPatternPhiGrid     // artistic phi-based grid geometry
)

// DrawAlphaMaskCirc draws a circular mask going from RGBA(0, 0, 0, 0) at cx, cy
// to the renderer's color at >= dist. This is primarily a utility method to create
// masks for [Renderer.Mask]() or [Renderer.MaskThreshold]() operations.
func (r *Renderer) DrawAlphaMaskCirc(target *ebiten.Image, cx, cy, dist, distRand float32, pattern AlphaMaskPattern) {
	r.opts.Uniforms["RngPattern"] = int(pattern)
	tox, toy, tw, th := rectOriginSizeF32(target.Bounds())
	r.setFlatCustomVAs(cx-tox, cy-toy, dist, distRand)
	r.DrawRectShader(target, 0, 0, tw, th, NoMargins, shaderAlphaMaskCirc.Load())
	clear(r.opts.Uniforms)
}
