package shapes

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

// this file contains Fill* and Stroke* functions for circular shapes:
// circles, circular sectors and arcs, ellipses, etc. for rectangles and
// polygons, see renderer_shape_poly.go instead

// NewFilledCircle returns a new image with circle of the given radius, drawn
// with the renderer's current color.
func (r *Renderer) NewFilledCircle(radius float64) *ebiten.Image {
	side := float32(math.Ceil(radius * 2))
	img := ebiten.NewImage(int(side), int(side))
	r.FillCircle(img, side/2, side/2, float32(radius))
	return img
}

// FillCircle draws a filled circle. Radius can't be negative.
// See also [Renderer.StrokeCircle](), [Renderer.FillCircSector]().
func (r *Renderer) FillCircle(target *ebiten.Image, cx, cy, radius float32) {
	if radius == 0 {
		return
	}
	if radius < 0 {
		r.Warnings.report(WarnNegativeValueOpSkipped, radius)
		return
	}
	r.setDstRectCoords(cx-radius, cy-radius, cx+radius, cy+radius)
	tox, toy := rectOriginF32(target.Bounds())
	r.setFlatCustomVAs(cx-tox, cy-toy, radius, 0.0)
	target.DrawTrianglesShader32(r.vertices[:], r.indices[:], shaderCircle.Load(), &r.opts)
}

// StrokeCircle draws a circle outline. If thickness > 0, the outline expands [-thickness/2, thickness/2]
// around the radius. If thickness < 0, the outline goes from [-thickness, 0].
//
// For arcs, see [Renderer.StrokeCircArc]().
func (r *Renderer) StrokeCircle(target *ebiten.Image, cx, cy, radius, thickness float32) {
	if thickness == 0 {
		return // nothing to draw
	}
	if thickness < 0 {
		thickness = -thickness
		radius -= thickness / 2.0
	}
	if thickness/2.0 >= radius {
		radius := radius + thickness/2.0
		if radius < 0 {
			return
		}
		r.FillCircle(target, cx, cy, radius)
		return
	}

	hthickCeil := ceilF32(thickness / 2.0)
	r.setDstRectCoords(cx-radius-hthickCeil, cy-radius-hthickCeil, cx+radius+hthickCeil, cy+radius+hthickCeil)
	tox, toy := rectOriginF32(target.Bounds())
	r.setFlatCustomVAs(cx-tox, cy-toy, radius, thickness)
	target.DrawTrianglesShader32(r.vertices[:], r.indices[:], shaderStrokeCircle.Load(), &r.opts)
}

// StrokeCircArc draws an arc of the given radius. For stroking full circles,
// consider [Renderer.StrokeCircle]() instead.
func (r *Renderer) StrokeCircArc(target *ebiten.Image, cx, cy, radius, startRads, endRads, thickness float64) {
	if thickness <= 0 {
		if thickness < 0 {
			r.Warnings.report(WarnNegativeValueOpSkipped, thickness)
		}
		return
	}

	if radius < 0 {
		r.Warnings.report(WarnNegativeValueZeroed, radius)
		thickness += radius
		radius = 0
		if thickness <= 0 {
			return
		}
	}

	// normalize radians
	delta := uradsDeltaCW(startRads, endRads)
	if delta >= 2.0*math.Pi {
		startRads, endRads = 0.0, 2.0*math.Pi
		delta = 2.0 * math.Pi
	} else {
		startRads, endRads = normURads(startRads), normURads(endRads)
	}

	// bounds
	cx32, cy32, radius32, thick32 := float32(cx), float32(cy), float32(radius), float32(thickness)
	hthick32 := thick32 / 2.0
	minX, minY, maxX, maxY := circSectorBounds(cx32, cy32, radius32, radius32, startRads, endRads)
	minX, minY = minX-hthick32-1.0, minY-hthick32-1.0
	maxX, maxY = maxX+hthick32+1.0, maxY+hthick32+1.0

	// shader
	r.setDstRectCoords(minX, minY, maxX, maxY)
	s, c := math.Sincos(delta / 2.0)
	r.setFlatCustomVAs(float32(c), float32(s), radius32, thick32)
	tox, toy := rectOriginF32(target.Bounds())
	r.opts.Uniforms["Center"] = [2]float32{cx32 - tox, cy32 - toy}
	s, c = math.Sincos(uradsAddCW(startRads, delta*0.5))
	r.opts.Uniforms["Rotation"] = [2]float32{float32(c), float32(s)}
	target.DrawTrianglesShader32(r.vertices[:], r.indices[:], shaderCircLine.Load(), &r.opts)
	clear(r.opts.Uniforms)
}

// fillCircWedge draws a circular sector with the contact points at the inner and
// outer radius being controlled by inAperture and outAperture, which must be in
// [0..2*Pi]. inner rounding is not implemented because of analytical geometry skill
// issue with the many edge cases.
func (r *Renderer) fillCircWedge(target *ebiten.Image, cx, cy, inRadius, outRadius, centerDir, inAperture, outAperture, rounding float64) {
	if inRadius < 0 {
		r.Warnings.report(WarnRadiusClamped, inRadius)
		inRadius = 0.0
	}
	if outRadius < 0 {
		r.Warnings.report(WarnRadiusClamped, outRadius)
		outRadius = 0.0
	}
	if inRadius == outRadius && rounding <= 0 {
		return
	}
	if inRadius > outRadius {
		r.Warnings.report(WarnInconsistentRangeOpSkipped, [2]float64{inRadius, outRadius})
		return
	}
	if inAperture > 2*math.Pi || inAperture < 0.0 {
		r.Warnings.report(WarnInvalidApertureClamped, inAperture)
		inAperture = clamp(inAperture, 0.0, 2.0*math.Pi)
	}
	if outAperture > 2*math.Pi || outAperture < 0.0 {
		r.Warnings.report(WarnInvalidRateClamped, outAperture)
		outAperture = clamp(outAperture, 0.0, 2.0*math.Pi)
	}

	// compute in and out radius points relative to centerDir = 0 (right)
	inSin, inCos := math.Sincos(inAperture / 2.0)
	outSin, outCos := math.Sincos(outAperture / 2.0)
	inX, inY := inCos*inRadius, inSin*inRadius
	outX, outY := outCos*outRadius, outSin*outRadius

	// perform line-circle intersection to clamp the in point to the
	// nearest side of the outer point (within the in circle)
	inX, inY = lineCircIntersect(inRadius, inX, inY, outX, outY)

	centerDirSin, centerDirCos := math.Sincos(centerDir)
	r.innerFillCircWedge(target, cx, cy, inRadius, outRadius, centerDirSin, centerDirCos, inX, inY, outX, outY, rounding)
}

// precondition: inX, inY, outX and outY given relative to centerDir = 0 (right), distances match inRadius and outRadius
func (r *Renderer) innerFillCircWedge(target *ebiten.Image, cx, cy, inRadius, outRadius, centerDirSin, centerDirCos, inX, inY, outX, outY, rounding float64) {
	// TODO: bounding can use min(inX, outX), min(-inY, -outY), max(inX, outX), max(inY, outY) probably?
	// minX, maxX := min(inX, outX), max(inX, outX)
	// minY, maxY := min(-inY, -outY), max(inY, outY)
	minX, maxX := -outRadius-rounding, outRadius+rounding // TODO: unbounded
	minY, maxY := -outRadius-rounding, outRadius+rounding
	r.setDstRectCoords(float32(cx+minX), float32(cy+minY), float32(cx+maxX), float32(cy+maxY))
	r.opts.Uniforms["Radiuses"] = [2]float32{float32(inRadius), float32(outRadius)}
	r.opts.Uniforms["InPoint"] = [2]float32{float32(inX), float32(inY)}
	r.opts.Uniforms["OutPoint"] = [2]float32{float32(outX), float32(outY)}
	r.opts.Uniforms["Rotation"] = [2]float32{float32(centerDirCos), float32(centerDirSin)}

	tox, toy := rectOriginF32(target.Bounds())
	r.setFlatCustomVAs(float32(cx)-tox, float32(cy)-toy, float32(rounding), 0)
	target.DrawTrianglesShader32(r.vertices[:], r.indices[:], shaderCircSectorSegment.Load(), &r.opts)
	clear(r.opts.Uniforms)
}

// FillCircSector draws a smooth circular sector. Use inRadius = 0 for pie shapes, inRadius > 0 for
// rings. Rounding can be positive for outwards rounding, or negative for inwards rounding. Notice
// that inwards rounding requires non-trivial CPU precalculations and a different shader from outwards
// or no rounding.
//
// Consider [RadsSpan]() if you need to derive (startRads, endRads) from a central direction.
//
// See [RadsRight] constants for angle conventions and docs.
func (r *Renderer) FillCircSector(target *ebiten.Image, cx, cy, inRadius, outRadius float32, startRads, endRads float64, rounding float32) {
	if inRadius >= outRadius || outRadius < 0 || startRads == endRads {
		return // skip empty draws
	}
	if inRadius < 0 {
		r.Warnings.report(WarnRadiusClamped, inRadius)
		inRadius = 0.0
	}
	if outRadius < 0 {
		r.Warnings.report(WarnRadiusClamped, outRadius)
		outRadius = 0.0
	}
	if inRadius == outRadius {
		return
	}
	if inRadius > outRadius {
		r.Warnings.report(WarnInconsistentRangeOpSkipped, [2]float32{inRadius, outRadius})
		return
	}

	if endRads >= startRads+2*math.Pi {
		thickness := outRadius - inRadius + max(rounding, 0)
		r.StrokeCircle(target, cx, cy, outRadius-thickness/2, thickness)
		return
	}

	startRads, endRads = normURads(startRads), normURads(endRads)
	r.internalFillCircSector(target, cx, cy, inRadius, outRadius, startRads, endRads, rounding)
}

// precondition: angles are normalized to [0, 2*pi)
func (r *Renderer) internalFillCircSector(target *ebiten.Image, cx, cy, inRadius, outRadius float32, startRads, endRads float64, rounding float32) {
	delta := uradsDeltaCW(startRads, endRads)
	halfDelta := delta * 0.5
	centerDir := uradsAddCW(startRads, halfDelta)
	ws, wc := math.Sincos(halfDelta)

	if rounding < 0 {
		r.innerRoundingFillCircSector(target, float64(cx), float64(cy), centerDir, halfDelta, ws, wc, float64(inRadius), float64(outRadius), startRads, endRads, float64(-rounding))
		return
	}

	pieMinX, pieMinY, pieMaxX, pieMaxY := circSectorBounds(cx, cy, inRadius, outRadius, startRads, endRads)
	if rounding > 0 {
		pieMinX -= rounding
		pieMinY -= rounding
		pieMaxX += rounding
		pieMaxY += rounding
	}

	r.setDstRectCoords(float32(pieMinX), float32(pieMinY), float32(pieMaxX), float32(pieMaxY))
	r.opts.Uniforms["WedgeNormal"] = [2]float32{float32(ws), float32(wc)}
	r.opts.Uniforms["InRadius"] = inRadius
	r.opts.Uniforms["Rounding"] = rounding
	tox, toy := rectOriginF32(target.Bounds())
	r.setFlatCustomVAs(cx-tox, cy-toy, float32(centerDir), outRadius)
	target.DrawTrianglesShader32(r.vertices[:], r.indices[:], shaderCircSector.Load(), &r.opts)
	clear(r.opts.Uniforms)
}

// precondition: rounding must be > 0. ws, wc are sin and cos of halfDelta
func (r *Renderer) innerRoundingFillCircSector(target *ebiten.Image, cx, cy, centerDir, halfDelta, ws, wc, inRadius, outRadius, startRads, endRads, rounding float64) {
	if rounding <= 0 {
		panic("rounding <= 0")
	}

	// shrink radiuses
	arcCollapse := -min((outRadius-inRadius)-rounding*2, 0)
	if arcCollapse == 0 {
		inRadius += rounding
		outRadius -= rounding
	} else if arcCollapse >= rounding {
		return // nothing to draw, shape fully collapsed
	} else { // partial collapse, in/out radius become the same and expansion starts to shrink
		outRadius = inRadius + (outRadius-inRadius)/2.0
		inRadius = outRadius
		// effective rounding will become rounding - arcCollapse for the arc,
		// but full rounding still has to be applied to the cut itself
	}

	if rounding > 2*inRadius || inRadius <= 0 {
		panic("unreachable")
	}

	// compute cut sector origin, which is displaced alongside centerDir
	dirOffset := rounding / ws
	cds, cdc := math.Sincos(centerDir)

	// handle shape collapses
	if dirOffset >= inRadius {
		panic("TODO: validate later")
		if dirOffset >= outRadius {
			return // treat as fully collapsed
		}

		// cone collapse
		cutCX, cutCY := cx+cdc*dirOffset, cy+cds*dirOffset
		r.internalFillCircSector(target, float32(cutCX), float32(cutCY), 0, float32(outRadius-dirOffset), startRads, endRads, float32(rounding-arcCollapse))
		return
	}

	var relInCut, relOutCut [2]float64
	if inRadius > 0 {
		relInCut, _, _ = circIntersect(0, 0, inRadius, wc*inRadius, ws*inRadius, rounding)
	} else { // inRadius <= 0, treat separately as a collapsed pie
		relInCut[0], relInCut[1] = dirOffset, 0
	}

	relOutCut, _, _ = circIntersect(0, 0, outRadius, wc*outRadius, ws*outRadius, rounding)
	r.innerFillCircWedge(target, cx, cy, inRadius, outRadius, cds, cdc, relInCut[0], relInCut[1], relOutCut[0], relOutCut[1], rounding-arcCollapse)
}

// StrokeCircSector draws the outline of a circular sector. Thickness must be >= 0.
// Rounding can be positive for outwards rounding, or negative for inwards rounding.
// Notice that inwards rounding requires non-trivial CPU precalculations.
//
// See [RadsRight] constants for angle conventions and docs.
func (r *Renderer) StrokeCircSector(target *ebiten.Image, cx, cy, inRadius, outRadius, thickness float32, startRads, endRads float64, rounding float32) {
	if inRadius >= outRadius || outRadius < 0 || startRads == endRads || thickness <= 0 {
		return // skip empty draws
	}
	if inRadius < 0 {
		r.Warnings.report(WarnNegativeValueZeroed, inRadius)
		inRadius = 0
	}
	if endRads >= startRads+2*math.Pi {
		r.StrokeCircle(target, cx, cy, inRadius, thickness)
		r.StrokeCircle(target, cx, cy, outRadius, thickness)
		return
	}

	startRads, endRads = normURads(startRads), normURads(endRads)
	r.internalStrokeCircSector(target, cx, cy, inRadius, outRadius, thickness, startRads, endRads, rounding)
}

// precondition: angles are normalized to [0, 2*pi)
func (r *Renderer) internalStrokeCircSector(target *ebiten.Image, cx, cy, inRadius, outRadius, thickness float32, startRads, endRads float64, rounding float32) {
	pieMinX, pieMinY, pieMaxX, pieMaxY := circSectorBounds(cx, cy, inRadius, outRadius, startRads, endRads)
	if rounding != 0 {
		r := abs(rounding)
		pieMinX -= r
		pieMinY -= r
		pieMaxX += r
		pieMaxY += r
	}
	pieMinX -= thickness / 2.0
	pieMinY -= thickness / 2.0
	pieMaxX += thickness / 2.0
	pieMaxY += thickness / 2.0

	r.setDstRectCoords(pieMinX, pieMinY, pieMaxX, pieMaxY)
	delta := uradsDeltaCW(startRads, endRads)
	centerDir := uradsAddCW(startRads, delta/2.0)
	ws, wc := math.Sincos(delta / 2.0)
	r.opts.Uniforms["WedgeNormal"] = [2]float32{float32(ws), float32(wc)}
	r.opts.Uniforms["InRadius"] = inRadius
	r.opts.Uniforms["Rounding"] = rounding
	r.opts.Uniforms["Thickness"] = thickness
	tox, toy := rectOriginF32(target.Bounds())
	r.setFlatCustomVAs(cx-tox, cy-toy, float32(centerDir), outRadius)
	target.DrawTrianglesShader32(r.vertices[:], r.indices[:], shaderStrokeCircSector.Load(), &r.opts)
	clear(r.opts.Uniforms)
}

// FillEllipse draws a smooth filled ellipse.
//
// Notice: ellipses don't have a perfect SDF, so approximations can be very slightly
// bigger or smaller than the requested radiuses.
func (r *Renderer) FillEllipse(target *ebiten.Image, cx, cy, horzRadius, vertRadius float32, rads float64) {
	if horzRadius < 0 {
		r.Warnings.report(WarnNegativeValueOpSkipped, horzRadius)
		horzRadius = 0
	}
	if vertRadius < 0 {
		r.Warnings.report(WarnNegativeValueOpSkipped, vertRadius)
		vertRadius = 0
	}
	if horzRadius == 0 || vertRadius == 0 {
		return
	}

	if rads == 0 {
		r.setDstRectCoords(cx-horzRadius, cy-vertRadius, cx+horzRadius, cy+vertRadius)
		r.opts.Uniforms["Radians"] = 0
	} else {
		hRadiusF64, vRadiusF64 := float64(horzRadius), float64(vertRadius)
		rs, rc := math.Sincos(rads)
		halfWidth := float32(math.Hypot(hRadiusF64*rc, vRadiusF64*rs))
		halfHeight := float32(math.Hypot(hRadiusF64*rs, vRadiusF64*rc))
		r.setDstRectCoords(cx-halfWidth, cy-halfHeight, cx+halfWidth, cy+halfHeight)
		r.opts.Uniforms["Radians"] = rads
	}
	tox, toy := rectOriginF32(target.Bounds())
	r.setFlatCustomVAs(cx-tox, cy-toy, horzRadius, vertRadius)
	target.DrawTrianglesShader32(r.vertices[:], r.indices[:], shaderEllipse.Load(), &r.opts)
	clear(r.opts.Uniforms)
}
