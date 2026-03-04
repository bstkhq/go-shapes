package shapes

import (
	"github.com/hajimehoshi/ebiten/v2"
)

func (r *Renderer) TileRectsGrid(target *ebiten.Image, outWidth, outHeight, inWidth, inHeight, xOffset, yOffset float32) {
	useOffsets := (xOffset != 0 || yOffset != 0)
	if useOffsets {
		r.opts.Uniforms["Offsets"] = [2]float32{xOffset, yOffset}
	}
	r.setFlatCustomVAs(outWidth, outHeight, inWidth, inHeight)
	tw, th := rectSizeF32(target.Bounds())
	r.DrawRectShader(target, 0, 0, tw, th, NoMargins, shaderTileRectsGrid.Load())
	if useOffsets {
		clear(r.opts.Uniforms)
	}
}

// TileDotsHex draws dots of the given radius distributed in a hexagonal
// lattice. HorzSpacing should always be at least twice the radius.
func (r *Renderer) TileDotsHex(target *ebiten.Image, radius, horzSpacing, xOffset, yOffset float32) {
	r.setFlatCustomVAs(radius, horzSpacing, xOffset, yOffset)
	tw, th := rectSizeF32(target.Bounds())
	r.DrawRectShader(target, 0, 0, tw, th, NoMargins, shaderTileDotsHex.Load())
}

func (r *Renderer) TileDotsGrid(target *ebiten.Image, radius, spacing, xOffset, yOffset float32) {
	r.setFlatCustomVAs(radius, spacing, xOffset, yOffset)
	tw, th := rectSizeF32(target.Bounds())
	r.DrawRectShader(target, 0, 0, tw, th, NoMargins, shaderTileDotsGrid.Load())
}

func (r *Renderer) TileTriUpGrid(target *ebiten.Image, outTriBase, inTriBase, xOffset, yOffset float32) {
	r.setFlatCustomVAs(xOffset, yOffset, outTriBase, inTriBase)
	tw, th := rectSizeF32(target.Bounds())
	r.DrawRectShader(target, 0, 0, tw, th, NoMargins, shaderTileTriUpGrid.Load())
}

func (r *Renderer) TileTriHex(target *ebiten.Image, outTriBase, inTriBase, xOffset, yOffset float32) {
	r.setFlatCustomVAs(xOffset, yOffset, outTriBase, inTriBase)
	tw, th := rectSizeF32(target.Bounds())
	r.DrawRectShader(target, 0, 0, tw, th, NoMargins, shaderTileTriHex.Load())
}
