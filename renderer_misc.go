package shapes

import (
	"math"

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

// TODO: rename to Fx* or Misc*?

// ScanlinesSharp is a miscellaneous effect that draws a simple scanline effect.
//
// Offset can be progressively increased to animate the scanlines, but notice that the
// shader uses nearest sampling, not smooth interpolation.
func (r *Renderer) ScanlinesSharp(target *ebiten.Image, darkThick, clearThick int, intensity, offset float32) {
	r.setFlatCustomVAs(float32(darkThick), float32(clearThick), intensity, offset)
	tw, th := rectSizeF32(target.Bounds())
	r.DrawRectShader(target, 0, 0, tw, th, NoMargins, shaderScanlinesSharp.Load())
}

// WaveLines is a miscellaneous effect that draws a pattern of lines with oscillating widths.
//
// dir defines the direction in which the lines advance if offset is used. The actual
// line direction is perpendicular to dir. For common values, see [DirRadsLTR] and
// related constants.
func (r *Renderer) WaveLines(target *ebiten.Image, lineThick, minFillRate, maxFillRate, linesPerOsc, offset float32, dir float64) {
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
	drs, drc := math.Sincos(dir)
	hypot := math.Hypot(drs, drc)
	drs, drc = drs/hypot, drc/hypot
	r.opts.Uniforms["DirRadsSin"] = float32(drs)
	r.opts.Uniforms["DirRadsCos"] = float32(drc)
	r.setFlatCustomVAs(lineThick, minFillThick, maxFillThick, waveLen)
	tw, th := rectSizeF32(target.Bounds())
	r.DrawRectShader(target, 0, 0, tw, th, NoMargins, shaderWaveLines.Load())
	clear(r.opts.Uniforms)
}
