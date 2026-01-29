package shapes

import (
	"github.com/hajimehoshi/ebiten/v2"
)

func (r *Renderer) HalftoneTri(target, source *ebiten.Image, ox, oy, outTriBaseSize, minInTriBaseSize, maxInTriBaseSize, xOffset, yOffset float32) {
	hasOffsets := (xOffset != 0 || yOffset != 0)
	if hasOffsets {
		r.opts.Uniforms["Offsets"] = [2]float32{xOffset, yOffset}
	}
	r.setFlatCustomVAs(outTriBaseSize, minInTriBaseSize, maxInTriBaseSize, 0)
	r.DrawShaderAt(target, source, ox, oy, 0, 0, shaderHalftoneTri.Load())
	if hasOffsets {
		clear(r.opts.Uniforms)
	}
}
