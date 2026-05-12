package shapes

import (
	"github.com/hajimehoshi/ebiten/v2"
)

// HalftoneTri renders source into target with a halftone effect based on equilateral triangles.
//   - triBaseSize defines the side of the triangles conforming the main grid.
//   - loSize and hiSize defined the size of the triangle fill at minimum and maximum lightness,
//     respectively. In most cases, it's more intuitive to use triBaseSize*0.2, triBaseSize*1.0
//     or similar to control the "fill ratio".
//
// This operation is affected by [Renderer.Tint].
//
// Notice that only the triangles are drawn and there's no explicit color for the background.
func (r *Renderer) HalftoneTri(target, source *ebiten.Image, ox, oy, triBaseSize, loSize, hiSize, xOffset, yOffset float32) {
	hasOffsets := (xOffset != 0 || yOffset != 0)
	if hasOffsets {
		r.opts.Uniforms["Offsets"] = [2]float32{xOffset, yOffset}
	}
	r.setFlatCustomVAs(triBaseSize, loSize, hiSize, r.tint)
	r.DrawImgShader(target, source, ox, oy, NoMargins, shaderHalftoneTri.Load())
	if hasOffsets {
		clear(r.opts.Uniforms)
	}
}

// func (r *Renderer) HalftoneDots(target, source *ebiten.Image, ox, oy, cellSize, loSize, hiSize, xOffset, yOffset float32) {}
