package shapes

import (
	"fmt"
	"image/color"
	"math/bits"

	"github.com/hajimehoshi/ebiten/v2"
)

// JFMapComputeOptions are the options for [Renderer.JFMapCompute]().
type JFMapComputeOptions struct {
	// MaxDistance specifies the range of the jumping flood map. The higher the
	// distance, the more passes needed to compute the map (logarithmic
	// increase, up to 17). Pixels out of reach are treated as unknown /
	// infinite distance.
	//
	// The exact boundaries from which distances are calculated are specified
	// by MinAlpha/MaxAlpha; see those fields for more details.
	//
	// JFM operations can read slightly beyond their requested operation range,
	// so maps should be built with at least the desired distance + 2.
	// MaxDistance must be in [0, 32k].
	MaxDistance int

	// Inset shifts where clamping is applied. This is only relevant for
	// clamped outlines, where the source needs padding, but the padding itself
	// would bring back the unwanted boundaries. Inset can be set to match the
	// source padding to compensate for this. Inset can't be negative.
	Inset int

	// Normals are optional alpha normals of the source, computed with
	// [Renderer.AlphaNormalsCompute]() and the same flags as the map. They
	// make distances more accurate near corners and oblique edges, at the cost
	// of more work: computing them, ~20-30% more to compute maps, and extra
	// reads in the operations. The gains are often small, and thin lines can
	// be better without them. They must be the same size as the source.
	Normals *ebiten.Image

	// MinAlpha/MaxAlpha specify a region of the source. The boundary of this
	// region (or boundaries, as multiple areas can be part of it) is what the
	// jumping flood map measures distances to.
	//
	// If both values are zero, the region defaults to all the non-transparent
	// pixels [0.001, 1]. This is equivalent to a boundary delimiting the
	// shapes in source, which is suitable for most operations.
	//
	// Values must be 0 <= MinAlpha <= MaxAlpha <= 1.
	MinAlpha, MaxAlpha float32
}

// JFMap is a jumping flood map of a source image, used by operations like
// [Renderer.JFMExpand](). Precomputed maps are returned by
// [Renderer.JFMapCompute](), with Normals and MinAlpha/MaxAlpha set to the
// ones they were computed with, which the operations need unchanged.
type JFMap struct {
	// Offsets is the jumping flood map, which stores the offset from each
	// pixel to its nearest seed (see [Renderer.JFMapCompute]()).
	//
	// Operations must draw the same source it was computed from. For
	// expansions and outlines to have room, pad the source before computing
	// the map.
	//
	// If nil, operations compute it on each draw, which is very expensive.
	Offsets *ebiten.Image

	// Normals are optional alpha normals of the source, as described in
	// [JFMapComputeOptions].
	Normals *ebiten.Image

	// MinAlpha/MaxAlpha specify the region that operations expand, erode or
	// outline: the pixels with alpha between them (inclusive). If both are
	// zero, the region is all the non-transparent pixels.
	//
	// Values must be 0 <= MinAlpha <= MaxAlpha <= 1.
	MinAlpha, MaxAlpha float32
}

// region returns the region limits, with the default when both are zero.
func (jfmap JFMap) region() (minAlpha, maxAlpha float32) {
	return jfmRegion(jfmap.MinAlpha, jfmap.MaxAlpha)
}

// jfmRegion returns the region limits, with the default when both are zero.
func jfmRegion(minAlpha, maxAlpha float32) (float32, float32) {
	if minAlpha == 0 && maxAlpha == 0 {
		return 0.001, 1
	}
	return minAlpha, maxAlpha
}

// JFMapCompute computes a jumping flood map of the given source, as described
// by opts, and stores it in target, which must be the same size as source.
// Besides computing target, the method also returns a [JFMap] ready to use
// with operations like [Renderer.JFMExpand]().
//
// A jumping flood map encodes offsets from each pixel to the nearest seed,
// which here are the pixels that the region boundary crosses. This allows
// computing precise signed distances to the boundary, both inside and outside
// the region, and can make large radius morphological operations like
// outlining, expansion and erosion viable (see [Renderer.JFMOutline](),
// [Renderer.JFMExpand]() and [Renderer.JFMErode]()). Seeds are compared by the
// same distance that these operations measure, so a single map serves all of
// them. As a downside, maps are expensive to recompute on the fly.
//
// Jumping flood map internal encoding details are documented in
// shaders/morph/jfm_pass.kage.
//
// Supported flags: [Clamp]. Clamping in jumping flood maps is rather unique:
// it's only applied to prevent the creation of boundaries on the source
// borders, but it doesn't extend source content beyond them. As a result:
//   - For erosion, it prevents erosion from starting at the source borders.
//   - For outlines, it prevents outlines along the source borders.
//   - For expansion, it's not intended to be used, as there's no meaningful
//     geometric interpretation.
//
// This function uses one internal offscreen (#0). target is always overwritten
// and doesn't need to be cleared before operation. It can't be a subimage of
// the same image as source or opts.Normals, as they're read while the map is
// written.
func (r *Renderer) JFMapCompute(target, source *ebiten.Image, opts JFMapComputeOptions, flags ...Flag) JFMap {
	jfmap := JFMap{Offsets: target, Normals: opts.Normals, MinAlpha: opts.MinAlpha, MaxAlpha: opts.MaxAlpha}
	if source == nil {
		r.Warnings.report(WarnMissingSourceOpSkipped, source)
		return jfmap
	}
	clampBorders := r.readOptInFlag(Clamp, flags...)
	if target == nil {
		panic("target is nil")
	}
	maxDistance := opts.MaxDistance
	if maxDistance > 32000 { // up to 32766 should be technically distinguishable
		r.Warnings.report(WarnDistanceClamped, maxDistance)
		maxDistance = 32000
	} else if maxDistance < 0 {
		panic("maxDistance < 0")
	}
	sw, sh := rectSize(source.Bounds())
	if mw, mh := rectSize(target.Bounds()); sw != mw || sh != mh {
		panic(fmt.Sprintf("source size != target size (%dx%d != %dx%d)", sw, sh, mw, mh))
	}
	checkNormalsSize(source, opts.Normals)
	minAlpha, maxAlpha, valid := r.checkJFMRegion(opts.MinAlpha, opts.MaxAlpha)
	jfmap.MinAlpha, jfmap.MaxAlpha = minAlpha, maxAlpha
	inset := warnZeroNegativeValue(r, opts.Inset)
	if !valid || 2*inset >= min(sw, sh) {
		target.Fill(color.White) // no seeds
		return jfmap
	}

	// mark the seeds on offscreen #0, which then serves as the second buffer
	// of the passes
	memoBlend := r.opts.Blend
	r.opts.Blend = ebiten.BlendCopy
	r.opts.Uniforms["HasNormals"] = mapBool[float32](opts.Normals != nil, 0, 1)
	seeds, _ := r.getTemp(0, sw, sh, false)
	inner := source
	if inset > 0 {
		seeds.Fill(color.White)
		inner = source.SubImage(source.Bounds().Inset(inset)).(*ebiten.Image)
	}
	ix, iy, iw, ih := rectOriginSizeF32(inner.Bounds())
	fi := float32(inset)
	r.setFlatCustomVAs(mapBool[float32](clampBorders == Clamp, 0, 1), minAlpha, maxAlpha, 0)
	r.setDstRectCoords(fi, fi, fi+iw, fi+ih)
	r.setSrcRectCoords(ix, iy, ix+iw, iy+ih)
	r.opts.Images[0] = inner
	seeds.DrawTrianglesShader32(r.vertices[:], r.indices[:], shaderJFMInit.Load(), &r.opts)

	jfmOX, jfmOY := rectOriginF32(target.Bounds())
	w, h := float32(sw), float32(sh)
	if maxDistance == 0 {
		var drawOpts ebiten.DrawImageOptions
		drawOpts.Blend = ebiten.BlendCopy
		drawOpts.GeoM.Translate(float64(jfmOX), float64(jfmOY))
		target.DrawImage(seeds, &drawOpts)
	} else {
		// 1+JFA+1: a first pass with jump size 1 from the seeds into target,
		// then passes from the largest jump size down, alternating between
		// target and the seeds buffer, and a final pass with jump size 1. Seeds
		// are compared by boundary distances, which shift by up to half a
		// pixel with the seeds' alpha, so the nearest one can arrive at a
		// neighbor only in the last regular pass, too late to spread
		shader := shaderJFMPass.Load()
		r.setFlatCustomVAs(1.0, float32(maxDistance), minAlpha, maxAlpha)
		r.opts.Images[0] = seeds
		r.opts.Images[1] = opts.Normals
		r.opts.Images[2] = source
		r.setDstRectCoords(jfmOX, jfmOY, jfmOX+w, jfmOY+h)
		r.setSrcRectCoords(0, 0, w, h)
		target.DrawTrianglesShader32(r.vertices[:], r.indices[:], shader, &r.opts)

		bufferCoords := [2][4]float32{{jfmOX, jfmOY, jfmOX + w, jfmOY + h}, {0, 0, w, h}}
		maps := [2]*ebiten.Image{target, seeds}
		mapIndex := 1
		var jumpSizes []int
		for jumpSize := 1 << (15 - bits.LeadingZeros16(uint16(maxDistance))); jumpSize > 0; jumpSize /= 2 {
			jumpSizes = append(jumpSizes, jumpSize)
		}
		for _, jumpSize := range append(jumpSizes, 1) {
			r.setFlatCustomVA0(float32(jumpSize))
			newIndex := 1 - mapIndex
			dst, src := bufferCoords[mapIndex], bufferCoords[newIndex]
			r.setDstRectCoords(dst[0], dst[1], dst[2], dst[3])
			r.setSrcRectCoords(src[0], src[1], src[2], src[3])
			r.opts.Images[0] = maps[newIndex]
			maps[mapIndex].DrawTrianglesShader32(r.vertices[:], r.indices[:], shader, &r.opts)
			mapIndex = newIndex
		}
		if mapIndex == 0 { // the last pass was drawn on the seeds buffer
			var drawOpts ebiten.DrawImageOptions
			drawOpts.Blend = ebiten.BlendCopy
			drawOpts.GeoM.Translate(float64(jfmOX), float64(jfmOY))
			target.DrawImage(seeds, &drawOpts)
		}
	}

	r.opts.Images[0] = nil
	r.opts.Images[1] = nil
	r.opts.Images[2] = nil
	clear(r.opts.Uniforms)
	r.opts.Blend = memoBlend
	return jfmap
}

// JFMExpand performs morphological expansion of the given source and draws it
// onto the given target. Pixels inside the region keep their own coverage plus
// the distance, which also preserves fades.
//   - distance must be in [0, 32k].
//   - If jfmap.Offsets is nil, it is generated using [Renderer.JFMapCompute]().
//     This uses internal offscreens (#0, #1, #3, and #2 with normals).
//   - With precomputed jfmap.Offsets, the expansion stops at the source borders
//     (see [JFMap] for how to make room for it).
//
// Only the source's alpha is used; colors are taken from the renderer.
//
// For additional context on jumping flood maps, see [Renderer.JFMapCompute]()
// and [JFMap]. See also [Renderer.MorphExpansion]().
func (r *Renderer) JFMExpand(target, source *ebiten.Image, ox, oy, distance float32, jfmap JFMap) {
	distance = r.checkJFMDistance(distance)
	r.jfmRead(target, source, ox, oy, distance, distance, true, jfmap, shaderJFMRead.Load(), jfmOpExpand, 0)
}

// JFMErode performs morphological erosion of the given source and draws it onto
// the given target, keeping the source colors.
//   - distance must be in [0, 32k].
//   - If jfmap.Offsets is nil, it is generated using [Renderer.JFMapCompute]().
//     This uses internal offscreens (#0, #3).
//
// Supported flags: [Clamp], which only applies to internal jumping flood map
// computation when jfmap.Offsets is nil. Clamping prevents erosion from
// starting at the source borders.
//
// This operation is affected by [Renderer.Tint].
//
// For additional context on jumping flood maps, see [Renderer.JFMapCompute]()
// and [JFMap]. See also [Renderer.MorphErosion]().
func (r *Renderer) JFMErode(target, source *ebiten.Image, ox, oy, distance float32, jfmap JFMap, flags ...Flag) {
	distance = r.checkJFMDistance(distance)
	r.jfmRead(target, source, ox, oy, distance, distance, false, jfmap, shaderJFMRead.Load(), jfmOpErode, r.tint, flags...)
}

// JFMOutline draws an outline of the given source onto the given target,
// centered on the region boundary, with a total width of thickness.
//   - thickness must be in [0, 32k].
//   - If jfmap.Offsets is nil, it is generated using [Renderer.JFMapCompute]().
//     This uses internal offscreens (#0, #1, #3, and #2 with normals).
//   - With precomputed jfmap.Offsets, the outline stops at the source borders
//     (see [JFMap] for how to make room for it), and the map needs a
//     MaxDistance of at least thickness/2 + 2.
//
// Supported flags: [Clamp], which only applies to internal jumping flood map
// computation when jfmap.Offsets is nil. Clamping prevents outlines along the
// source borders.
//
// Only the source's alpha is used; colors are taken from the renderer.
//
// For additional context on jumping flood maps, see [Renderer.JFMapCompute]()
// and [JFMap]. See also [Renderer.MorphOutline]().
func (r *Renderer) JFMOutline(target, source *ebiten.Image, ox, oy, thickness float32, jfmap JFMap, flags ...Flag) {
	thickness = r.checkJFMDistance(thickness)
	r.jfmRead(target, source, ox, oy, thickness, thickness/2, true, jfmap, shaderJFMRead.Load(), jfmOpOutline, 0, flags...)
}

// JFMapHeat is a debug and utility method to draw a heatmap of the distances in
// a jumping flood map into the given target, using 0 and maxDistance as
// reference distances for "hot" and "cold". With inside, it shows the
// distances inside the region, and otherwise the ones outside. The other side
// is drawn as distance 0, and the pixels on the shown side without any seed
// within the map range as maxDistance. The colors are multiplied by the
// renderer's color.
//
// If jfmap.Offsets is nil, it is generated as in [Renderer.JFMOutline](), which
// also describes the supported flags.
//
// For additional context on jumping flood maps, see [Renderer.JFMapCompute]().
// See also [Renderer.JFMapViz](), which shows both sides.
func (r *Renderer) JFMapHeat(target, source *ebiten.Image, ox, oy, maxDistance float32, inside bool, jfmap JFMap, flags ...Flag) {
	maxDistance = r.checkJFMDistance(maxDistance)
	r.jfmRead(target, source, ox, oy, maxDistance, maxDistance, true, jfmap, shaderJFMHeat.Load(), jfmNoOp, mapBool[float32](inside, 0, 1), flags...)
}

// JFMapViz is a debug and utility method to draw the signed distances in a
// jumping flood map into the given target, in the style of Inigo Quilez's
// signed distance field shaders: orange outside the region and blue inside,
// fading towards maxDistance, with darker lines every 8 pixels of distance and
// a white line along the boundary. Pixels without any seed within the map
// range are left transparent. The colors are multiplied by the renderer's
// color.
//
// If jfmap.Offsets is nil, it is generated as in [Renderer.JFMOutline](), which
// also describes the supported flags.
//
// For additional context on jumping flood maps, see [Renderer.JFMapCompute]().
func (r *Renderer) JFMapViz(target, source *ebiten.Image, ox, oy, maxDistance float32, jfmap JFMap, flags ...Flag) {
	maxDistance = r.checkJFMDistance(maxDistance)
	r.jfmRead(target, source, ox, oy, maxDistance, maxDistance, true, jfmap, shaderJFMViz.Load(), jfmNoOp, 0, flags...)
}

// operations for jfm_read.kage. jfmNoOp is for the other shaders
const (
	jfmNoOp = iota - 1
	jfmOpExpand
	jfmOpErode
	jfmOpOutline
)

// helper for the operations that read jumping flood maps. reach is how far
// from the boundary the operation can draw, and outside tells whether that
// includes the area outside the source, which then needs padding when the map
// is generated. extra is the tint for erosion, and 1 to show the inside for
// heatmaps.
func (r *Renderer) jfmRead(target, source *ebiten.Image, ox, oy, amount, reach float32, outside bool, jfmap JFMap, shader *ebiten.Shader, op int, extra float32, flags ...Flag) {
	if source == nil {
		r.Warnings.report(WarnMissingSourceOpSkipped, source)
		return
	}
	clampBorders := r.readOptInFlag(Clamp, flags...)
	if clampBorders == Clamp && jfmap.Offsets != nil {
		r.Warnings.report(WarnIncompatibleFlag, clampBorders)
		clampBorders = noFlag
	}
	minAlpha, maxAlpha, valid := r.checkJFMRegion(jfmap.MinAlpha, jfmap.MaxAlpha)
	if !valid {
		return
	}
	checkNormalsSize(source, jfmap.Normals)

	if jfmap.Offsets == nil {
		// boundaries can be up to ~1.5 pixels away from the seed centers, so
		// the map and the padding need some extra range
		mapRange := int(ceilF32(reach)) + 2
		padding := mapBool(outside, 0, mapRange)
		if padding > 0 {
			source = r.UnsafeTempCopy(1, source, padding, false)
			if jfmap.Normals != nil {
				jfmap.Normals = r.UnsafeTempCopy(2, jfmap.Normals, padding, false)
			}
		}
		sw, sh := rectSize(source.Bounds())
		temp, _ := r.getTemp(3, sw, sh, false)
		// with Clamp, the inset skips the padding so that clamping happens at
		// the original borders
		opts := JFMapComputeOptions{
			MaxDistance: mapRange,
			Inset:       mapBool(clampBorders == Clamp, 0, padding),
			Normals:     jfmap.Normals,
			MinAlpha:    minAlpha,
			MaxAlpha:    maxAlpha,
		}
		jfmap = r.JFMapCompute(temp, source, opts, clampBorders)
		ox -= float32(padding)
		oy -= float32(padding)
	} else {
		sw, sh := rectSize(source.Bounds())
		if mw, mh := rectSize(jfmap.Offsets.Bounds()); sw != mw || sh != mh {
			panic(fmt.Sprintf("source size != jfmap size (%dx%d != %dx%d)", sw, sh, mw, mh))
		}
	}

	r.opts.Uniforms["HasNormals"] = mapBool[float32](jfmap.Normals != nil, 0, 1)
	if op != jfmNoOp {
		r.opts.Uniforms["Operation"] = float32(op)
	}
	r.opts.Images[1] = jfmap.Offsets
	r.opts.Images[2] = jfmap.Normals
	r.setFlatCustomVAs(amount, extra, minAlpha, maxAlpha)
	r.DrawImgShader(target, source, ox, oy, NoMargins, RegionExact, shader)
	r.opts.Images[1] = nil
	r.opts.Images[2] = nil
	clear(r.opts.Uniforms)
}

// checkJFMRegion sets the default region if needed, clamps the region limits
// to [0, 1] and reports whether the region is valid, with warnings for the
// problems found.
func (r *Renderer) checkJFMRegion(minAlpha, maxAlpha float32) (float32, float32, bool) {
	minAlpha, maxAlpha = jfmRegion(minAlpha, maxAlpha)
	if minAlpha < 0 {
		r.Warnings.report(WarnInvalidAlphaClamped, minAlpha)
		minAlpha = 0
	}
	if maxAlpha > 1 {
		r.Warnings.report(WarnInvalidAlphaClamped, maxAlpha)
		maxAlpha = 1
	}
	if minAlpha > maxAlpha {
		r.Warnings.report(WarnInconsistentRangeInvalidated, [2]float32{minAlpha, maxAlpha})
		return minAlpha, maxAlpha, false
	}
	return minAlpha, maxAlpha, true
}

// checkJFMDistance clamps distances to the range jfmaps can represent.
func (r *Renderer) checkJFMDistance(distance float32) float32 {
	if distance > 32000 { // up to 32766 should be technically distinguishable
		r.Warnings.report(WarnDistanceClamped, distance)
		return 32000
	} else if distance < 0 {
		panic("distance < 0")
	}
	return distance
}

// panics if normals is given and its size differs from source
func checkNormalsSize(source, normals *ebiten.Image) {
	if normals == nil {
		return
	}
	sw, sh := rectSize(source.Bounds())
	nw, nh := rectSize(normals.Bounds())
	if sw != nw || sh != nh {
		panic(fmt.Sprintf("source size != normals size (%dx%d != %dx%d)", sw, sh, nw, nh))
	}
}
