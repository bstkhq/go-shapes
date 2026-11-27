package shapes

import (
	"github.com/hajimehoshi/ebiten/v2"
)

const (
	// Precision cushion for morph contributions that reach RegionExact boundaries.
	morphMarginEpsilon = 0.001
	// Maximum output reach beyond the morphological radius. The shader's
	// one-pixel sample transition starts at source centers, half a pixel inwards.
	morphOutlineExtraReach = 0.5
)

// MorphExpansion performs morphological expansion of the given mask and draws
// it onto the given target. Notice that this is a quadratic algorithm. For
// large expansion operations, consider [Renderer.MorphExpansionRect]() and
// [Renderer.JFMExpand]().
//
// thickness must be in [0, 16].
//
// softEdge can soften edges inwards (towards the original edges). It can't be
// negative, and values above thickness are clamped to thickness.
//
// Smooth position animation requires sufficiently large thickness and softEdge.
// Animating either value alongside position can still cause stuttering.
//
// Only the mask's alpha is used; colors are taken from the renderer.
func (r *Renderer) MorphExpansion(target *ebiten.Image, mask *ebiten.Image, ox, oy, thickness, softEdge float32) {
	if mask == nil {
		r.Warnings.report(WarnMissingSourceOpSkipped, mask)
		return
	}
	thickness = r.warnClampNonNegArgF32(thickness, 16, WarnThicknessClamped)
	softEdge = min(warnZeroNegativeValue(r, softEdge), thickness)

	r.setFlatCustomVAs01(thickness, softEdge)
	margin := float32(0)
	if thickness > 0 {
		// The shader returns zero at pixel centers beyond thickness.
		// The epsilon helps include centers exactly on the boundary.
		margin = thickness + morphMarginEpsilon
	}
	margins := NewMargins(margin, margin)
	r.DrawImgShader(target, mask, ox, oy, margins, RegionExact, shaderMorphExpansion.Load())
}

// MorphExpansionRect performs double pass expansion with a square kernel. This
// is more efficient than [Renderer.MorphExpansion](), but less general.
//
// thickness must be in [0, 32].
//
// Only the mask's alpha is used; colors are taken from the renderer.
//
// This function uses one internal offscreen (#0), and target and mask can be on
// the same internal atlas.
func (r *Renderer) MorphExpansionRect(target *ebiten.Image, mask *ebiten.Image, ox, oy, thickness float32) {
	if mask == nil {
		r.Warnings.report(WarnMissingSourceOpSkipped, mask)
		return
	}
	thickness = r.warnClampNonNegArgF32(thickness, 32, WarnThicknessClamped)

	// the first pass is shifted by oyFract so it lines up when drawn from
	// the temp to the final target on the second pass
	oyWhole := floorF32(oy)
	oyFract := oy - oyWhole

	hmargin := thickness + 0.5        // reaches >= the center of the last affected pixel
	vmargin := ceilF32(thickness) + 1 // upper bound, exact would be ±ceil(thickness ± oyFract)

	sx, sy, sw, sh := rectOriginSize(mask.Bounds())
	sx32, sy32, sw32, sh32 := float32(sx), float32(sy), float32(sw), float32(sh)
	tempH := sh32 + vmargin*2
	temp, _ := r.getTemp(0, sw, int(tempH), false)

	// first pass (vert)
	memoBlend := r.opts.Blend
	r.opts.Blend = ebiten.BlendCopy
	r.setSrcRectCoords(sx32, sy32-vmargin-oyFract, sx32+sw32, sy32-vmargin-oyFract+tempH)
	r.setDstRectCoords(0, 0, sw32, tempH)
	r.setFlatCustomVA0(thickness)
	r.opts.Images[0] = mask
	temp.DrawTrianglesShader32(r.vertices[:], r.indices[:], shaderMorphExpansionRectVert.Load(), &r.opts)
	r.opts.Images[0] = nil

	// second pass (horz)
	r.opts.Blend = memoBlend
	r.setSrcRectCoords(-hmargin, 0, sw32+hmargin, tempH)
	r.setDstRectCoords(ox-hmargin, oyWhole-vmargin, ox+sw32+hmargin, oyWhole-vmargin+tempH)
	r.opts.Images[0] = temp
	target.DrawTrianglesShader32(r.vertices[:], r.indices[:], shaderMorphExpansionRectHorz.Load(), &r.opts)
	r.opts.Images[0] = nil
}

// MorphErosion performs morphological erosion of the given source and draws it
// onto the given target. Notice that this is a quadratic algorithm. For large
// erosion operations, consider [Renderer.JFMErode]().
//
// thickness must be in [0, 16].
//
// softEdge can soften edges outwards (towards the original edges). It can't be
// negative, and values above thickness are clamped to thickness.
//
// Smooth position animation requires sufficiently large thickness and softEdge.
// Animating either value alongside position can still cause stuttering.
//
// Supported flags: [Clamp], which prevents erosion from starting at the
// source borders.
//
// This operation is affected by [Renderer.Tint].
func (r *Renderer) MorphErosion(target *ebiten.Image, source *ebiten.Image, ox, oy, thickness, softEdge float32, flags ...Flag) {
	if source == nil {
		r.Warnings.report(WarnMissingSourceOpSkipped, source)
		return
	}
	clampBorders := r.readOptInFlag(Clamp, flags...)
	thickness = r.warnClampNonNegArgF32(thickness, 16, WarnThicknessClamped)
	softEdge = min(warnZeroNegativeValue(r, softEdge), thickness)
	r.setFlatCustomVAs(thickness, softEdge, mapBool[float32](clampBorders == Clamp, 0, 1), r.tint)
	r.DrawImgShader(target, source, ox, oy, NoMargins, RegionExact, shaderMorphErosion.Load())
}

// MorphOutline draws an outline of the mask into the given target. This
// operation is implemented as the difference between morphological expansion
// and erosion. Notice that this is a quadratic algorithm. For large outlines,
// consider [Renderer.JFMOutline]().
//
// If thickness > 0, the outline is centered on the mask boundary. If
// thickness < 0, the outline extends inwards by -thickness. thickness must be
// in [-8, 16].
//
// softEdge softens the edge(s) of the outline inwards. This value must be in
// [0, thickness/2] for balanced outlines, and [0, -thickness] for inner
// outlines. A zero value gives the sharpest result; neighboring pixels still
// enter the outline gradually to reduce flicker during subpixel motion.
//
// Supported flags: [Clamp], which prevents outlines along the mask borders.
// With Clamp, nothing is drawn outside the mask area.
//
// Only the mask's alpha is used; colors are taken from the renderer.
func (r *Renderer) MorphOutline(target *ebiten.Image, mask *ebiten.Image, ox, oy, thickness, softEdge float32, flags ...Flag) {
	if mask == nil {
		r.Warnings.report(WarnMissingSourceOpSkipped, mask)
		return
	}
	clampBorders := r.readOptInFlag(Clamp, flags...)
	if thickness < -8 || thickness > 16 {
		r.Warnings.report(WarnThicknessClamped, thickness)
		thickness = clamp(thickness, -8, 16)
	}
	morphDistance := thickness / 2.0
	if thickness < 0 {
		morphDistance = -thickness
	}
	softEdge = min(warnZeroNegativeValue(r, softEdge), morphDistance)
	if thickness == 0 {
		return
	}

	r.setFlatCustomVAs(thickness, softEdge, mapBool[float32](clampBorders == Clamp, 0, 1), 0)
	margin := float32(0)
	if thickness > 0 && clampBorders != Clamp { // see Clamp docs above
		margin = morphDistance + morphOutlineExtraReach + morphMarginEpsilon
	}
	r.DrawImgShader(target, mask, ox, oy, NewMargins(margin, margin), RegionExact, shaderMorphOutline.Load())
}
