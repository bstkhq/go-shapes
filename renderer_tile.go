package shapes

import (
	"github.com/hajimehoshi/ebiten/v2"
)

// TileRectsGrid draws rectangles of the given size distributed in a grid
// pattern. outWidth and outHeight control the grid spacing, while inWidth
// and inHeight define the size of the rectangles to draw.
//
// All widths and heights must be positive. in sizes can be >= out sizes,
// but rectangles will start overlapping.
//
// The color of the rectangles is determined by the renderer's current color.
func (r *Renderer) TileRectsGrid(target *ebiten.Image, inWidth, inHeight, outWidth, outHeight, xOffset, yOffset float32) {
	if outWidth <= 0 {
		r.Warnings.report(WarnNonPositiveValueOpSkipped, outWidth)
		return
	}
	if outHeight <= 0 {
		r.Warnings.report(WarnNonPositiveValueOpSkipped, inWidth)
		return
	}
	inWidth = warnZeroNegativeValue(r, inWidth)
	inHeight = warnZeroNegativeValue(r, inHeight)
	if (inWidth == 0 || inHeight == 0) && r.blendSafeToCrop() {
		return
	}

	useOffsets := (xOffset != 0 || yOffset != 0)
	if useOffsets {
		r.opts.Uniforms["Offsets"] = [2]float32{xOffset, yOffset}
	}
	r.setFlatCustomVAs(inWidth, inHeight, outWidth, outHeight)
	tw, th := rectSizeF32(target.Bounds())
	r.DrawRectShader(target, 0, 0, tw, th, NoMargins, shaderTileRectsGrid.Load())
	if useOffsets {
		clear(r.opts.Uniforms)
	}
}

// TileDotsHex draws dots of the given radius distributed in a hexagonal
// lattice. horzSpacing defines the horizontal spacing between the
// dot centers, and the row height is horzSpacing*[Sqrt3Div2].
//
// radius and horzSpacing must be positive. horzSpacing can be <= radius, but
// dots will start overlapping.
//
// The color of the dots is determined by the renderer's current color.
func (r *Renderer) TileDotsHex(target *ebiten.Image, radius, horzSpacing, xOffset, yOffset float32) {
	if horzSpacing <= 0 {
		r.Warnings.report(WarnNonPositiveValueOpSkipped, horzSpacing)
		return
	}
	radius = warnZeroNegativeValue(r, radius)
	if radius == 0 && r.blendSafeToCrop() {
		return
	}

	r.setFlatCustomVAs(radius, horzSpacing, xOffset, yOffset)
	tw, th := rectSizeF32(target.Bounds())
	r.DrawRectShader(target, 0, 0, tw, th, NoMargins, shaderTileDotsHex.Load())
}

// TileDotsGrid draws dots of the given radius distributed in a grid
// pattern. spacing defines the grid cell size.
//
// radius and spacing must be positive. horzSpacing can be <= radius,
// but dots will start overlapping.
//
// The color of the dots is determined by the renderer's current color.
func (r *Renderer) TileDotsGrid(target *ebiten.Image, radius, spacing, xOffset, yOffset float32) {
	if spacing <= 0 {
		r.Warnings.report(WarnNonPositiveValueOpSkipped, spacing)
		return
	}
	radius = warnZeroNegativeValue(r, radius)
	if radius == 0 && r.blendSafeToCrop() {
		return
	}

	r.setFlatCustomVAs(radius, spacing, xOffset, yOffset)
	tw, th := rectSizeF32(target.Bounds())
	r.DrawRectShader(target, 0, 0, tw, th, NoMargins, shaderTileDotsGrid.Load())
}

// TileTriUpGrid draws upwards equilateral triangles in a grid. The triangles have base inTriBase
// and height inTriBase*[Sqrt3Div2]. The grid is divided in rectangles of width outTriBase and
// height outTriBase*[Sqrt3Div2].
//
// outTriBase and inTribase must be positive. outTriBase can be <= inTriBase, but triangles will
// start overlapping.
//
// The color of the triangles is determined by the renderer's current color.
func (r *Renderer) TileTriUpGrid(target *ebiten.Image, inTriBase, outTriBase, xOffset, yOffset float32) {
	if outTriBase <= 0 {
		r.Warnings.report(WarnNonPositiveValueOpSkipped, outTriBase)
		return
	}
	inTriBase = warnZeroNegativeValue(r, inTriBase)
	if inTriBase == 0 && r.blendSafeToCrop() {
		return
	}

	r.setFlatCustomVAs(xOffset, yOffset, inTriBase, outTriBase)
	tw, th := rectSizeF32(target.Bounds())
	r.DrawRectShader(target, 0, 0, tw, th, NoMargins, shaderTileTriUpGrid.Load())
}

// TileTriHex draws equilateral triangles alternating up and down in a hexagonal latice.
// outTriBase defines the width of the lattice's triangular cells, and the row height is
// outTriBase*[Sqrt3Div2].
//
// inTriBase and outTriBase must be positive. outTriBase can be <= inTriBase, but triangles
// will start overlapping.
//
// The color of the triangles is determined by the renderer's current color.
func (r *Renderer) TileTriHex(target *ebiten.Image, inTriBase, outTriBase, xOffset, yOffset float32) {
	if outTriBase <= 0 {
		r.Warnings.report(WarnNonPositiveValueOpSkipped, outTriBase)
		return
	}
	inTriBase = warnZeroNegativeValue(r, inTriBase)
	if inTriBase == 0 && r.blendSafeToCrop() {
		return
	}

	r.setFlatCustomVAs(xOffset, yOffset, inTriBase, outTriBase)
	tw, th := rectSizeF32(target.Bounds())
	r.DrawRectShader(target, 0, 0, tw, th, NoMargins, shaderTileTriHex.Load())
}
