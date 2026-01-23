package shapes

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

// WarpBarrel draws the given image with a simple, CRT-like barrel warp.
// Intensity should be in ~[0.2, 1.5], with 0.5 being a good starting value
// to play with.
//
// The size of the output image will always be equal or smaller than the
// input source, as the corner vertices are warped towards the interior.
//
// If both warp values are <= 0, a quadratic curve-based pincushion effect
// will be used instead. Notice that very large warp factors in both axes
// will make the image start to "shrink".
//
// For visually pleasing effects, you usually want to normalize the warps
// by the image aspect ratio.
//
// Notice that the effects are designed to be adjustable per axis and fast.
// More mathematically accurate atan/sin based warps are possible, but for
// soft warps the quadratic versions are pleasant enough.
//
// Warps of different signs will panic.
func (r *Renderer) WarpBarrel(target, source *ebiten.Image, ox, oy float32, horzWarp, vertWarp float32) {
	if horzWarp == 0 && vertWarp == 0 {
		return
	}
	if (horzWarp < 0 || vertWarp < 0) && (horzWarp > 0 || vertWarp > 0) {
		panic("horzWarp and vertWarp must have the same sign")
	}

	switch {
	case horzWarp <= 0 && vertWarp <= 0:
		r.warpPincushionQuad(target, source, ox, oy, -horzWarp, -vertWarp)
	default:
		r.setFlatCustomVAs01(horzWarp, vertWarp)
		ensureShaderWarpBarrelLoaded()
		r.DrawShaderAt(target, source, ox, oy, 0, 0, shaderWarpBarrel)
	}
}

func (r *Renderer) warpPincushionQuad(target, source *ebiten.Image, ox, oy float32, horzWarp, vertWarp float32) {
	if horzWarp < 0 || vertWarp < 0 {
		panic("broken internal code")
	}

	// perceptual adjustment to better match WarpBarrel strength
	horzWarp *= 0.2
	vertWarp *= 0.2

	r.setFlatCustomVAs01(horzWarp, vertWarp)
	ensureShaderWarpPincushionQuadLoaded()
	r.DrawShaderAt(target, source, ox, oy, 0, 0, shaderWarpPincushionQuad)
}

// WarpArc projects the given source image onto a curved arc on target.
// The arc is characterized by (cx, cy) and outRadius. The content is
// horizontally centered at 'rads'. If the source's width > circumference,
// the content is automatically clamped.
//
// See [RadsRight] constants for angle conventions and docs.
func (r *Renderer) WarpArc(target, source *ebiten.Image, cx, cy, outRadius float32, rads float64) {
	if outRadius < 0 {
		return // nothing to draw
	}

	srcBounds := source.Bounds()
	sw, sh := rectSizeF32(srcBounds)
	inRadius := outRadius - sh
	circumference := 2 * math.Pi * outRadius
	radsHalfDelta := min(float64(sw/circumference)*math.Pi, math.Pi)
	startRads := normURads(rads - radsHalfDelta)
	var minX, minY, maxX, maxY float32
	if radsHalfDelta >= math.Pi {
		minX, minY = cx-outRadius, cy-outRadius
		maxX, maxY = cx+outRadius, cy+outRadius
	} else {
		minX, minY, maxX, maxY = ringSectorBounds(cx, cy, inRadius, outRadius, startRads, normURads(rads+radsHalfDelta))
	}

	minX, minY, maxX, maxY = minX-1.0, minY-1.0, maxX+1.0, maxY+1.0
	dstOX, dstOY := rectOriginF32(target.Bounds())
	r.setDstRectCoords(dstOX+minX, dstOY+minY, dstOX+maxX, dstOY+maxY)

	sox, soy, sfx, sfy := rectPointsF32(srcBounds)
	r.setSrcRectCoords(sox, soy, sfx, sfy)

	r.setFlatCustomVAs(outRadius, sw, float32(startRads), float32(radsHalfDelta*2.0))
	ensureShaderWarpArcLoaded()
	r.opts.Images[0] = source
	r.opts.Uniforms["Center"] = [2]float32{cx, cy}
	target.DrawTrianglesShader(r.vertices[:], r.indices[:], shaderWarpArc, &r.opts)
	r.opts.Images[0] = nil
	clear(r.opts.Uniforms)
}
