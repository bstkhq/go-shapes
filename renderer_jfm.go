package shapes

import (
	"fmt"
	"image/color"
	"math"
	"math/bits"

	"github.com/hajimehoshi/ebiten/v2"
)

// JFMapCompute computes a jumping flood map from the given seeds and
// stores it in jfmap.
//
// A jumping flood map encodes offsets to nearest seeds, which allows
// computing precise distances and can make large radius morphological
// operations like outlining, expansion and erosion viable.
//
// Jumping flood map internal encoding details are documented in shaders/jfm_pass.kage.
//
// Seed pixels in 'seeds' must be marked as trasparent vec4(0); all other
// pixels must be pure white. maxDistance acts as the cutoff distance for the
// algorithm, leaving pixels beyond it as pure white. maxDistance must be in
// [0, 32000] (inclusive). Higher maxDistance values require more iterations
// of the algorithm, up to a maximum of 16.
//
// This function uses one internal offscreen (#0). seeds can be on #0 if
// the image being overwritten is not a concern. jfmap is always overwritten
// and doesn't need to be cleared before operation.
//
// This is a low-level operation; most users should use [Renderer.JFMapFill]()
// or [Renderer.JFMapBoundary]() instead.
func (r *Renderer) JFMapCompute(jfmap, seeds *ebiten.Image, maxDistance int) {
	if maxDistance > 32000 { // up to 32766 should be technically distinguishable
		r.Warnings.report(WarnDistanceClamped, maxDistance)
		maxDistance = 32000
	} else if maxDistance < 0 {
		panic("maxDistance < 0")
	}

	sbounds := seeds.Bounds()
	tbounds := jfmap.Bounds()
	sw, sh := sbounds.Dx(), sbounds.Dy()
	tw, th := tbounds.Dx(), tbounds.Dy()
	if sw != tw || sh != th {
		panic(fmt.Sprintf("seeds size != jfmap size (%dx%d != %dx%d)", sw, sh, tw, th))
	}

	// base case
	if maxDistance == 0 {
		var opts ebiten.DrawImageOptions
		opts.Blend = ebiten.BlendCopy
		bounds := jfmap.Bounds()
		opts.GeoM.Translate(float64(bounds.Min.X), float64(bounds.Min.Y))
		jfmap.DrawImage(seeds, &opts)
		return
	}

	// we use 1+JFA, so the first pass uses jump size = 1
	memoBlend := r.opts.Blend
	r.opts.Blend = ebiten.BlendCopy

	dstOX, dstOY := float32(tbounds.Min.X), float32(tbounds.Min.Y)
	w, h := float32(sw), float32(sh)
	mapCoords := [2][4]float32{{dstOX, dstOY, dstOX + w, dstOY + h}, {0, 0, w, h}}
	r.setFlatCustomVAs01(1.0, float32(maxDistance))
	r.opts.Images[0] = seeds
	r.setDstRectCoords(mapCoords[0][0], mapCoords[0][1], mapCoords[0][2], mapCoords[0][3])
	r.setSrcRectCoords(mapCoords[1][0], mapCoords[1][1], mapCoords[1][2], mapCoords[1][3])
	jfmap.DrawTrianglesShader(r.vertices[:], r.indices[:], shaderJFMPass.Load(), &r.opts)

	// - main JFA loop -
	// jump size starts at the base power of 2 of the current number
	temp, _ := r.getTemp(0, sw, sh, false)
	jumpSize := 1 << (15 - bits.LeadingZeros16(uint16(maxDistance)))
	maps := [2]*ebiten.Image{jfmap, temp}
	mapIndex := 1
	for jumpSize > 0 {
		r.setFlatCustomVA0(float32(jumpSize)) // set only jump size, maxDistance is already ok
		r.setDstRectCoords(mapCoords[mapIndex][0], mapCoords[mapIndex][1], mapCoords[mapIndex][2], mapCoords[mapIndex][3])
		newIndex := 1 - mapIndex
		r.setSrcRectCoords(mapCoords[newIndex][0], mapCoords[newIndex][1], mapCoords[newIndex][2], mapCoords[newIndex][3])
		r.opts.Images[0] = maps[newIndex]
		maps[mapIndex].DrawTrianglesShader(r.vertices[:], r.indices[:], shaderJFMPass.Load(), &r.opts)
		mapIndex = newIndex
		jumpSize /= 2
	}

	// copy to jfmap if last step was done on temp
	if mapIndex == 0 {
		var opts ebiten.DrawImageOptions
		opts.Blend = ebiten.BlendCopy
		jfmap.DrawImage(temp, &opts)
	}

	// cleanup
	r.opts.Blend = memoBlend
	r.opts.Images[0] = nil
}

// JFMapFill computes a jumping flood map of the given source image
// and stores it in jfmap.
//
// Preconditions:
//   - source and jfmap must have the same size
//   - 0 <= maxDistance <= 32k
//   - 0.0 <= minAlpha <= maxAlpha <= 1.0
//
// Parameters:
//   - minAlpha, maxAlpha: inclusive range defining the area inside the boundary.
//     For exclusive bounds, shift by +/-0.001.
//
// This function uses one internal offscreen (#0); neither source nor jfmap can use it,
// but they can share internal atlas otherwise. jfmap doesn't need to be cleared before
// operation.
//
// For additional context on jumping flood maps, see [Renderer.JFMapCompute]().
func (r *Renderer) JFMapFill(jfmap, source *ebiten.Image, maxDistance int, minAlpha, maxAlpha float32) {
	if minAlpha < 0 {
		r.Warnings.report(WarnInvalidAlphaClamped, minAlpha)
		minAlpha = 0
	}
	if maxAlpha > 1.0 {
		r.Warnings.report(WarnInvalidAlphaClamped, maxAlpha)
		maxAlpha = 1.0
	}
	if minAlpha > maxAlpha {
		r.Warnings.report(WarnInconsistentRangeInvalidated, [2]float32{minAlpha, maxAlpha})
		jfmap.Fill(color.RGBA{255, 255, 255, 255})
		return
	}

	r.setFlatCustomVAs01(minAlpha, maxAlpha)
	r.jfmInit(jfmap, source, maxDistance, shaderJFMInitFill.Load())
}

// JFMapBoundary computes a jumping flood map of the given source image
// and stores it in jfmap.
//
// Preconditions (panics if violated):
//   - source and jfmap must have the same size
//   - 0 <= maxDistance <= 32k
//   - 0.0 <= minAlpha <= maxAlpha <= 1.0
//
// Parameters:
//   - minAlpha, maxAlpha: inclusive range defining the area inside the boundary.
//     For exclusive bounds, shift by +/-0.001.
//   - outer: if false, the boundary is marked at the last pixel inside the
//     (minAlpha, maxAlpha) region; if true, at the first pixel outside it.
//   - extendEdges: if false, out-of-bounds pixels are treated as zero (vec4(0));
//     if true, the nearest edge pixel is repeated.
//
// This function uses one internal offscreen (#0); neither source nor jfmap can use it,
// but they can share internal atlas otherwise. jfmap doesn't need to be cleared before
// operation.
//
// For additional context on jumping flood maps, see [Renderer.JFMapCompute]().
func (r *Renderer) JFMapBoundary(jfmap, source *ebiten.Image, maxDistance int, minAlpha, maxAlpha float32, outer bool, extendEdges bool) {
	if minAlpha < 0 {
		r.Warnings.report(WarnInvalidAlphaClamped, minAlpha)
		minAlpha = 0
	}
	if maxAlpha > 1.0 {
		r.Warnings.report(WarnInvalidAlphaClamped, maxAlpha)
		maxAlpha = 1.0
	}
	if minAlpha > maxAlpha {
		r.Warnings.report(WarnInconsistentRangeInvalidated, [2]float32{minAlpha, maxAlpha})
		jfmap.Fill(color.RGBA{255, 255, 255, 255})
		return
	}

	boolToF32 := func(b bool) float32 {
		if b {
			return 1.0
		}
		return 0.0
	}
	r.setFlatCustomVAs(minAlpha, maxAlpha, boolToF32(outer), boolToF32(extendEdges))
	r.jfmInit(jfmap, source, maxDistance, shaderJFMInitBoundary.Load())
}

// helper for JFMapFill and JFMapBoundary. init shader vertex attributes must have already
// been set. uses unsafe offscreen #0.
func (r *Renderer) jfmInit(jfmap, source *ebiten.Image, maxDistance int, initShader *ebiten.Shader) {
	ox, oy, sw, sh := rectOriginSizeF32(source.Bounds())
	seeds, _ := r.getTemp(0, int(sw), int(sh), false)
	r.setDstRectCoords(0, 0, sw, sh)
	r.setSrcRectCoords(ox, oy, ox+sw, oy+sh)

	memoBlend := r.opts.Blend
	r.opts.Blend = ebiten.BlendCopy
	r.opts.Images[0] = source
	seeds.DrawTrianglesShader(r.vertices[:], r.indices[:], initShader, &r.opts)
	r.opts.Blend = memoBlend

	r.JFMapCompute(jfmap, seeds, maxDistance)
}

// JFMHeat is a debug and utility method to draw a heatmap for jfmap into the given target,
// using 0 and maxDistance as reference distances for "hot" and "cold". The seeds of a
// jfmap can be visualized by setting maxDistance to a positive value below 1 (e.g. 0.1).
func (r *Renderer) JFMHeat(target, jfmap *ebiten.Image, ox, oy float32, maxDistance float32) {
	r.setFlatCustomVA0(maxDistance)
	r.DrawShaderAt(target, jfmap, ox, oy, 0, 0, shaderJFMHeat.Load())
}

// JFMExpand performs morphological expansion. distance must be in [0, 32k].
// Notice that since jumping flood algorithms are based on distances to seeds,
// the only work well for shapes with hard edges. For soft edges, pure
// [Renderer.ApplyExpansion]() is the only real high quality option.
//
//   - jfmap can be nil, in which case it will be automatically generated for only this operation
//     using [JFMPixel] mode with [0.001, 1.0] alpha interval (all not fully transparent pixels
//     are seeds).
//   - source and jfmap should be in the same atlas to avoid automatic atlasing issues.
func (r *Renderer) JFMExpand(target, source, jfmap *ebiten.Image, ox, oy, distance float32) {
	if distance > 32000 { // up to 32766 should be technically distinguishable
		r.Warnings.report(WarnDistanceClamped, distance)
		distance = 32000
	} else if distance < 0 {
		panic("distance < 0")
	}

	var jfmapMaxDist float64
	if jfmap == nil {
		jfmapMaxDist = math.Ceil(float64(distance))
		source, jfmap = r.UnsafeTempDual(1, source, int(jfmapMaxDist), false)
		r.JFMapFill(jfmap, source, int(jfmapMaxDist), 0.001, 1.0)
	}

	r.opts.Images[1] = jfmap
	r.setFlatCustomVA0(distance)
	r.DrawShaderAt(target, source, ox-float32(jfmapMaxDist), oy-float32(jfmapMaxDist), 0, 0, shaderJFMExpansion.Load())
	r.opts.Images[1] = nil
}

// JFMErode performs morphological erosion. distance must be in [0, 32k].
//
//   - jfmap can be nil, in which case it will be automatically generated for only this operation
//     using [JFMPixel] mode with [0.0, 0.0] alpha interval (transparent pixels are seeds).
//   - source and jfmap should be in the same atlas to avoid automatic atlasing issues.
func (r *Renderer) JFMErode(target, source, jfmap *ebiten.Image, ox, oy, distance float32) {
	if distance > 32000 { // up to 32766 should be technically distinguishable
		r.Warnings.report(WarnDistanceClamped, distance)
		distance = 32000
	} else if distance < 0 {
		panic("distance < 0")
	}

	if jfmap == nil {
		jfmapMaxDist := int(math.Ceil(float64(distance)))
		source, jfmap = r.UnsafeTempDual(1, source, jfmapMaxDist, false)
		r.JFMapFill(jfmap, source, jfmapMaxDist, 0.0, 0.0)
	}

	r.opts.Images[1] = jfmap
	r.setFlatCustomVA0(distance)
	r.DrawShaderAt(target, source, ox, oy, 0, 0, shaderJFMErosion.Load())
	r.opts.Images[1] = nil
}

// TODO: unimplemented
//
// JFMOutline performs morphological outlining.
//
//   - colorMix controls the outline color (0 = use vertex colors, 1 = use source colors)
//   - jfmap can be nil, in which case it will be automatically generated for only this operation
//     using [JFMBoundary] mode.
//   - source and jfmap should be in the same atlas to avoid automatic atlasing issues.
func (r *Renderer) JFMOutline(target, source, jfmap *ebiten.Image, ox, oy, inThickness, outThickness, inOpacity, colorMix float32) {
	panic("unimplemented")
}

// TODO: unimplemented
//
// JFMInsetContour is a specific effect designed mainly for text animations. It creates an
// internal outline, which includes the image borders where the target clips the source, while
// also allowing to control the inner fill opacity.
//
//   - colorMix controls the outline color (0 = use vertex colors, 1 = use source colors)
//   - jfmap can be nil, in which case it will be automatically generated for only this operation
//     using [JFMBoundary] mode.
//   - source and jfmap should be in the same atlas to avoid automatic atlasing issues.
func (r *Renderer) JFMInsetContour(target, source, jfmap *ebiten.Image, ox, oy, inThickness, inOpacity, colorMix float32) {
	panic("unimplemented")
}

//func (r *Renderr) JFMFeather(target, source, jfmap *ebiten.Image, ox, oy, radius, curve float32) {}
