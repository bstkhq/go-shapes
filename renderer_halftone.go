package shapes

import (
	"github.com/hajimehoshi/ebiten/v2"
)

// HalftoneTri renders source into target with a halftone effect based on triangle shapes.
//
// This operation is affected by [Renderer.Tint].
//
// Notice that only the triangles are drawn and there's no explicit color for the background.
func (r *Renderer) HalftoneTri(target, source *ebiten.Image, ox, oy, outTriBaseSize, minInTriBaseSize, maxInTriBaseSize, xOffset, yOffset float32) {
	hasOffsets := (xOffset != 0 || yOffset != 0)
	if hasOffsets {
		r.opts.Uniforms["Offsets"] = [2]float32{xOffset, yOffset}
	}
	r.setFlatCustomVAs(outTriBaseSize, minInTriBaseSize, maxInTriBaseSize, r.tint)
	r.DrawImgShader(target, source, ox, oy, NoMargins, shaderHalftoneTri.Load())
	if hasOffsets {
		clear(r.opts.Uniforms)
	}
}
