package shapes

import (
	"fmt"
	"image/color"
	"math"
	"math/bits"

	"github.com/hajimehoshi/ebiten/v2"
)

// TODO: mention that jfmaps only works with / assumes solid shapes with crisp anti-aliasing (at most one translucent pixel on the shape boundary transition)

// TODO: accept bilinear flags for expand and all others

// JFMapCompute computes a jumping flood map from the given seeds and stores it in jfmap.
//
// A jumping flood map encodes offsets to nearest seeds, which allows computing precise
// distances and can make large radius morphological operations like outlining, expansion
// and erosion viable. As a downside, they are expensive to recompute on the fly and they
// are based on binary seeds, which means additional techniques might have to be used to
// smooth results in many contexts.
//
// Jumping flood map internal encoding details are documented in shaders/jfm_pass.kage.
//
// Seed pixels in 'seeds' must be marked as trasparent vec4(0); all other pixels must be
// pure white. maxDistance acts as the cutoff distance for the algorithm, leaving pixels
// beyond it as pure white. maxDistance must be in [0, 32000] (inclusive). Higher maxDistance
// values require more iterations of the algorithm (logarithmically), up to a maximum of 16.
//
// This function uses one internal offscreen (#0). seeds can be on #0 if the image being
// overwritten is not a concern. jfmap is always overwritten and doesn't need to be cleared
// before operation.
//
// This is a low-level operation; most users should use [Renderer.JFMapFill]() or
// [Renderer.JFMapBoundary]() instead.
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
	shader := shaderJFMPass.Load()
	memoBlend := r.opts.Blend
	r.opts.Blend = ebiten.BlendCopy

	jfmOX, jfmOY := rectOriginF32(tbounds)
	seedOX, seedOY := rectOriginF32(sbounds)
	w, h := float32(sw), float32(sh)
	mapCoords := [2][4]float32{{jfmOX, jfmOY, jfmOX + w, jfmOY + h}, {seedOX, seedOY, seedOX + w, seedOY + h}}
	r.setFlatCustomVAs01(1.0, float32(maxDistance))
	r.opts.Images[0] = seeds
	r.setDstRectCoords(mapCoords[0][0], mapCoords[0][1], mapCoords[0][2], mapCoords[0][3])
	r.setSrcRectCoords(mapCoords[1][0], mapCoords[1][1], mapCoords[1][2], mapCoords[1][3])
	jfmap.DrawTrianglesShader(r.vertices[:], r.indices[:], shader, &r.opts)

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
		maps[mapIndex].DrawTrianglesShader(r.vertices[:], r.indices[:], shader, &r.opts)
		mapIndex = newIndex
		jumpSize /= 2
	}

	// cleanup
	r.opts.Images[0] = nil

	// copy to jfmap if last step was done on temp
	if mapIndex == 0 {
		var opts ebiten.DrawImageOptions
		opts.Blend = ebiten.BlendCopy
		jfmap.DrawImage(temp, &opts)
	}

	// cleanup
	r.opts.Blend = memoBlend
}

// JFMapFill computes a jumping flood map of the given source image and stores it
// in jfmap. minAlpha and maxAlpha delimit the seeds area (inclusive). For exclusive
// bounds, shift by +/-0.001.
//
// Preconditions:
//   - source and jfmap must have the same size
//   - 0 <= maxDistance <= 32k
//   - 0.0 <= minAlpha <= maxAlpha <= 1.0
//
// This function uses one internal offscreen (#0); neither source nor jfmap can use it,
// but they can share internal atlas. jfmap doesn't need to be cleared before operation.
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

// BoundaryMode is a parameter type for [Renderer.JFMapBoundary]().
type BoundaryMode struct {
	// If false, the boundary is marked at the last pixel inside the
	// boundary region (minAlpha, maxAlpha). If true, at the first pixel
	// outside it.
	Outer bool

	// If false, out-of-bound pixels are treated as zero. If true,
	// nearest edge pixel color is used instead.
	ExtendEdges bool
}

// JFMapBoundary computes a jumping flood map of the given source image
// and stores it in jfmap. minAlpha and maxAlpha delimit the area inside
// the boundary (inclusive). For exclusive bounds, shift by +/-0.001.
//
// Preconditions (panics if violated):
//   - source and jfmap must have the same size
//   - 0 <= maxDistance <= 32k
//   - 0.0 <= minAlpha <= maxAlpha <= 1.0
//
// This function uses one internal offscreen (#0); neither source nor jfmap can use it,
// but they can share internal atlas. jfmap doesn't need to be cleared before operation.
//
// For additional context on jumping flood maps, see [Renderer.JFMapCompute]().
func (r *Renderer) JFMapBoundary(jfmap, source *ebiten.Image, maxDistance int, minAlpha, maxAlpha float32, mode BoundaryMode) {
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

	outer := mapBool[float32](mode.Outer, 0, 1)
	extend := mapBool[float32](mode.ExtendEdges, 0, 1)
	r.setFlatCustomVAs(minAlpha, maxAlpha, outer, extend)
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
//
// For additional context on jumping flood maps, see [Renderer.JFMapCompute]().
func (r *Renderer) JFMHeat(target, jfmap *ebiten.Image, ox, oy float32, maxDistance float32) {
	r.setFlatCustomVA0(maxDistance)
	r.DrawImgShader(target, jfmap, ox, oy, NoMargins, shaderJFMHeat.Load())
}

// JFMExpand performs morphological expansion.
//   - distance must be in [0, 32k].
//   - source and jfmap must be the same size, and they should be in the same atlas to avoid
//     automatic atlasing issues.
//   - smooth can be set to true to sample a 3x3 area for higher morphological accuracy.
//   - jfmap can be nil, in which case it will be automatically generated for only this operation
//     using [Renderer.JFMapFill]() with [0.001, 1.0] alpha interval (all non-transparent pixels
//     are seeds). If jfmap is nil, this function uses the internal offscreens (#0, #1).
//
// For additional context on jumping flood maps, see [Renderer.JFMapCompute]().
func (r *Renderer) JFMExpand(target, source, jfmap *ebiten.Image, ox, oy, distance float32, smooth bool) {
	if distance > 32000 { // up to 32766 should be technically distinguishable
		r.Warnings.report(WarnDistanceClamped, distance)
		distance = 32000
	} else if distance < 0 {
		panic("distance < 0")
	}

	if jfmap == nil {
		jfmapMaxDist := ceilF32(distance)
		source, jfmap = r.UnsafeTempDual(1, source, int(jfmapMaxDist), false)
		r.JFMapFill(jfmap, source, int(jfmapMaxDist), 0.001, 1.0)
		ox -= jfmapMaxDist // compensate drawing position
		oy -= jfmapMaxDist
	} else {
		sw, sh := rectSize(source.Bounds())
		mw, mh := rectSize(jfmap.Bounds())
		if sw != mw || sh != mh {
			panic(fmt.Sprintf("source size != jfmap size (%dx%d != %dx%d)", sw, sh, mw, mh))
		}
	}

	if smooth {
		r.opts.Uniforms["Smooth"] = 1
	}

	r.opts.Images[1] = jfmap
	r.setFlatCustomVA0(distance)
	r.DrawImgShader(target, source, ox, oy, NoMargins, shaderJFMExpansion.Load())
	r.opts.Images[1] = nil

	if smooth {
		clear(r.opts.Uniforms)
	}
}

// NOTE: expand and erode can be done with rel dist product. We might support a Feather flag alongside Bilinear.
//func (r *Renderr) JFMFeather(target, source, jfmap *ebiten.Image, ox, oy, radius, curve float32) {}

// JFMErode performs morphological erosion.
//   - distance must be in [0, 32k].
//   - source and jfmap should be in the same atlas to avoid automatic atlasing issues.
//   - jfmap can be nil, in which case it will be automatically generated for only this operation
//     using [Renderer.JFMapBoundary]() with [0.0, 0.0] alpha interval + outer. If jfmap is nil,
//     this function uses the internal offscreens (#0, #1).
//
// This operation is affected by [Renderer.Tint]. (TODO)
//
// For additional context on jumping flood maps, see [Renderer.JFMapCompute]().
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
		r.JFMapBoundary(jfmap, source, jfmapMaxDist, 0.0, 0.0, BoundaryMode{Outer: true})
	}

	r.opts.Images[1] = jfmap
	r.setFlatCustomVA0(distance)
	r.DrawImgShader(target, source, ox, oy, NoMargins, shaderJFMErosion.Load())
	r.opts.Images[1] = nil
}

// TODO: unimplemented
//
// JFMOutline performs morphological outlining.
//   - thicknesses must be in [0, 32k].
//   - source and jfmap should be in the same atlas to avoid automatic atlasing issues.
//   - jfmap can be nil, in which case it will be automatically generated for only this
//     operation using [JFMBoundary] mode.
//
// This operation is affected by [Renderer.Tint].
// TODO: specify offscreens being used
//
// For additional context on jumping flood maps, see [Renderer.JFMapCompute]().
func (r *Renderer) JFMOutline(target, source, jfmap *ebiten.Image, ox, oy, inThickness, outThickness, inOpacity float32) {
	panic("unimplemented")
}

// TODO: unimplemented
//
// JFMInsetContour is a specific effect designed mainly for text animations. It creates an
// internal outline, which includes the image borders where the target clips the source, and
// also allows control over the inner fill opacity.
//   - inThickness must be in [0, 32k].
//   - source and jfmap should be in the same atlas to avoid automatic atlasing issues.
//   - jfmap can be nil, in which case it will be automatically generated for only this
//     operation using [JFMBoundary] mode. If jfmap is nil, this function uses the internal
//     offscreens (#0, #1).
//
// This operation is affected by [Renderer.Tint]. (TODO)
//
// For additional context on jumping flood maps, see [Renderer.JFMapCompute]().
func (r *Renderer) JFMInsetContour(target, source, jfmap *ebiten.Image, ox, oy, inThickness, inOpacity float32) {
	panic("unimplemented")
}
