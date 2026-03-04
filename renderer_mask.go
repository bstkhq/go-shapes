package shapes

import (
	"github.com/hajimehoshi/ebiten/v2"
)

// Mask draws 'source' over 'target' using 'mask' as an alpha mask. If the source
// and mask sizes are different, the mask will be adjusted to fit the source
// (sampling is always nearest, not bilinear). For manual mask placement, see
// [Renderer.MaskAt]() instead.
func (r *Renderer) Mask(target, source, mask *ebiten.Image, ox, oy float32) {
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
}

// MaskAt draws 'source' over 'target' using 'mask' as an alpha mask at the given position.
// If you want the mask to be fit to the source instead, see [Renderer.Mask]().
func (r *Renderer) MaskAt(target, source, mask *ebiten.Image, ox, oy, oxMask, oyMask float32) {
	srcOX, srcOY, srcWidthF32, srcHeightF32 := rectOriginSizeF32(source.Bounds())
	r.setDstRectCoords(ox, oy, ox+srcWidthF32, oy+srcHeightF32)
	r.setSrcRectCoords(srcOX, srcOY, srcOX+srcWidthF32, srcOY+srcHeightF32)

	r.setFlatCustomVAs01(ox-oxMask, oy-oyMask)
	r.opts.Images[0] = source
	r.opts.Images[1] = mask
	target.DrawTrianglesShader(r.vertices[:], r.indices[:], shaderMaskAt.Load(), &r.opts)
	r.opts.Images[0] = nil
	r.opts.Images[1] = nil
}

// MaskThreshold draws source into target, at the given position, using 'mask' to hide
// the pixels where 'reveal' < mask.alpha.
//
// For example, if a mask goes from transparent to opaque, left to right, the source will
// start appearing from left to right as the 'reveal' threshold increases from 0 to 1.
//
// If source and mask sizes differ, the mask is adjusted like in [Renderer.Mask]().
func (r *Renderer) MaskThreshold(target, source, mask *ebiten.Image, reveal, ox, oy float32) {
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
}

// MaskHorz draws 'source' over 'target' but with an horizontal alpha fade between
// the given points.
func (r *Renderer) MaskHorz(target, source *ebiten.Image, x, y, inX, outX float32) {
	tox, toy := rectOriginF32(target.Bounds())
	r.setFlatCustomVAs01(inX-tox, outX-toy)
	r.DrawShaderAt(target, source, x, y, 0, 0, shaderMaskHorz.Load())
}

// MaskCircle draws 'source' into 'target', centered at (cx + srcOffsetX, cy + srcOffset),
// but filtering out pixels beyond a distance of hardRadius + softEdge from (cx, cy).
func (r *Renderer) MaskCircle(target, source *ebiten.Image, cx, cy, srcOffsetX, srcOffsetY, hardRadius, softEdge float32) {
	if softEdge < 0.0 {
		hardRadius += softEdge
		softEdge = -softEdge
	}
	if hardRadius < 0.0 {
		hardRadius = 0.0
	}
	if hardRadius == 0.0 && softEdge == 0.0 {
		return // omit draw, nothing to draw
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

// Related to DrawAlphaMaskCirc
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
func (r *Renderer) DrawAlphaMaskCirc(target *ebiten.Image, ox, oy, dist, distRand float32, pattern AlphaMaskPattern) {
	r.opts.Uniforms["RngPattern"] = int(pattern)
	tox, toy := rectOriginF32(target.Bounds())
	r.setFlatCustomVAs(ox-tox, oy-toy, dist, distRand)
	r.DrawShader(target, 0, 0, shaderAlphaMaskCirc.Load())
	clear(r.opts.Uniforms)
}
