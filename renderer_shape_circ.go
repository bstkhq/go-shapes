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

// StrokeCircle draws a circle outline or ring. If thickness > 0, the outline expands
// [-thickness/2, thickness/2] around the radius. If thickness < 0, the outline goes from
// [-thickness, 0].
//
// For arcs, see [Renderer.StrokeArc]().
func (r *Renderer) StrokeCircle(target *ebiten.Image, cx, cy, radius, thickness float32) {
	if thickness == 0 {
		return // nothing to draw
	}
	radius, thickness = cleanStrokeRadiusThickness(radius, thickness)
	if radius+thickness <= 0 {
		return // nothing to draw
	}
	if thickness == 0 {
		radius, thickness = 0.0, radius*2.0
	}

	hthickCeil := ceilF32(thickness / 2.0)
	r.setDstRectCoords(cx-radius-hthickCeil, cy-radius-hthickCeil, cx+radius+hthickCeil, cy+radius+hthickCeil)
	tox, toy := rectOriginF32(target.Bounds())
	r.setFlatCustomVAs(cx-tox, cy-toy, radius, thickness)
	target.DrawTrianglesShader32(r.vertices[:], r.indices[:], shaderStrokeCircle.Load(), &r.opts)
}

// cleanStrokeRadiusThickness converts negative thicknesses (inner) to positive
// (outer), and collapses thicknesses to zero if the circle stroke can become a
// single filled circle
func cleanStrokeRadiusThickness(radius, thickness float32) (float32, float32) {
	if thickness < 0 {
		thickness = -thickness
		radius -= thickness / 2.0
	}
	if thickness/2.0 >= radius {
		radius = max(radius+thickness/2.0, 0)
		thickness = 0.0
	}
	return radius, thickness
}

// StrokeArc draws an arc of the given radius. For stroking full circles and rings,
// consider [Renderer.StrokeCircle]() instead.
func (r *Renderer) StrokeArc(target *ebiten.Image, cx, cy, radius, startRads, endRads, thickness float64) {
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
	minX, minY, maxX, maxY := radialSectorBounds(cx32, cy32, radius32, radius32, startRads, endRads)
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
	target.DrawTrianglesShader32(r.vertices[:], r.indices[:], shaderArc.Load(), &r.opts)
	clear(r.opts.Uniforms)
}

func (r *Renderer) fillCircularSector(target *ebiten.Image, cx, cy float32, radius, startRads, endRads, rounding float64) {
	if rounding >= 0 {
		panic("unimplemented")
	}
	centralAngle := uradsDeltaCW(startRads, endRads)
	halfCtrAngle := centralAngle * 0.5
	centerDir := uradsAddCW(startRads, halfCtrAngle)
	halfCtrAngleSin, halfCtrAngleCos := math.Sincos(halfCtrAngle)
	centerDirSin, centerDirCos := math.Sincos(centerDir)

	r.innerRoundingFillCircularSector(target, cx, cy, startRads, endRads, radius, centerDirSin, centerDirCos, halfCtrAngleSin, halfCtrAngleCos, -rounding)
}

func (r *Renderer) innerRoundingFillCircularSector(target *ebiten.Image, cx, cy float32, startRads, endRads, radius, centerDirSin, centerDirCos, halfCtrAngleSin, halfCtrAngleCos, rounding float64) {
	if rounding <= 0 {
		panic("rounding <= 0")
	}

	nearCircDist := rounding / halfCtrAngleSin
	if nearCircDist+rounding >= radius {
		pointRadius := radius - nearCircDist // rounding - (nearCircDist + rounding - radius)
		if pointRadius > 0 {
			pointDist := radius - (rounding+pointRadius)*0.5
			pcx, pcy := centerDirCos*pointDist, centerDirSin*pointDist
			r.FillCircle(target, cx+float32(pcx), cy+float32(pcy), float32(pointRadius))
		}
		return
	}

	distToFarCircTangent := math.Sqrt(radius*radius - 2*radius*rounding) // derived from d^2 + rounding^2 = (radius - rounding)^2
	farCenterX := distToFarCircTangent*halfCtrAngleCos + rounding*halfCtrAngleSin
	farCenterY := distToFarCircTangent*halfCtrAngleSin - rounding*halfCtrAngleCos

	nearDist := float32(math.Sqrt(nearCircDist*nearCircDist - rounding*rounding))
	r.opts.Uniforms["TangentDists"] = [2]float32{float32(nearDist), float32(distToFarCircTangent)}
	r.opts.Uniforms["NearCircDist"] = float32(nearCircDist)
	r.opts.Uniforms["OutRadius"] = float32(radius)
	r.opts.Uniforms["FarCircCenter"] = [2]float32{float32(farCenterX), float32(farCenterY)}
	r.opts.Uniforms["Rotation"] = [2]float32{float32(centerDirCos), float32(centerDirSin)}

	r.setFlatCustomVAs(cx, cy, float32(halfCtrAngleSin*(1.0/max(halfCtrAngleCos, 0.0001))), float32(rounding))
	minX, minY, maxX, maxY := circularSectorBounds(cx, cy, float32(radius), startRads, endRads)
	r.setDstRectCoords(minX, minY, maxX, maxY)

	target.DrawTrianglesShader32(r.vertices[:], r.indices[:], shaderCircularSectorInner.Load(), &r.opts)
	clear(r.opts.Uniforms)
}

// fillRadialWedge draws a circular sector/annulus with the contact points at the inner
// and outer radius being controlled by inAperture and outAperture, which must be in
// [0..2*Pi]. inner rounding is not implemented because of analytical geometry skill
// issue with the many edge cases.
func (r *Renderer) fillRadialWedge(target *ebiten.Image, cx, cy, inRadius, outRadius, centerDir, inAperture, outAperture, rounding float64) {
	inRadius = warnZeroNegativeValue(r, inRadius)
	outRadius = warnZeroNegativeValue(r, outRadius)
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
	r.innerFillRadialWedge(target, cx, cy, inRadius, outRadius, centerDir, centerDirSin, centerDirCos, inX, inY, outX, outY, rounding)
}

// precondition: inX, inY, outX and outY given relative to centerDir = 0 (right), distances match inRadius and outRadius
func (r *Renderer) innerFillRadialWedge(target *ebiten.Image, cx, cy, inRadius, outRadius, centerDir, centerDirSin, centerDirCos, inX, inY, outX, outY, rounding float64) {
	minX, minY, maxX, maxY := radialWedgeBounds(cx, cy, outRadius, inX, inY, outX, outY, centerDir, centerDirSin, centerDirCos)
	margin := float32(rounding)
	minX -= margin
	maxX += margin
	minY -= margin
	maxY += margin

	r.setDstRectCoords(float32(minX), float32(minY), float32(maxX), float32(maxY))
	r.opts.Uniforms["Radiuses"] = [2]float32{float32(inRadius), float32(outRadius)}
	r.opts.Uniforms["InPoint"] = [2]float32{float32(inX), float32(inY)}
	r.opts.Uniforms["OutPoint"] = [2]float32{float32(outX), float32(outY)}
	r.opts.Uniforms["Rotation"] = [2]float32{float32(centerDirCos), float32(centerDirSin)}

	tox, toy := rectOriginF32(target.Bounds())
	r.setFlatCustomVAs(float32(cx)-tox, float32(cy)-toy, float32(rounding), 0)
	target.DrawTrianglesShader32(r.vertices[:], r.indices[:], shaderRadialSectorSegment.Load(), &r.opts)
	clear(r.opts.Uniforms)
}

// FillRadialSector draws a smooth radial sector.
//   - Use inRadius = 0 for pie shapes (circular sectors).
//   - Use inRadius > 0 for rings (annular sectors).
//
// Rounding can be positive for outwards rounding, or negative for inwards rounding.
// Notice that inwards rounding requires non-trivial CPU precalculations and different
// shaders from outwards or no rounding.
//
// Consider [RadsSpan]() if you need to derive (startRads, endRads) from a central direction.
//
// See [RadsRight] constants for angle conventions and docs.
func (r *Renderer) FillRadialSector(target *ebiten.Image, cx, cy, inRadius, outRadius float32, startRads, endRads float64, rounding float32) {
	if inRadius >= outRadius || outRadius < 0 || startRads == endRads {
		return // skip empty draws
	}
	inRadius = warnZeroNegativeValue(r, inRadius)
	outRadius = warnZeroNegativeValue(r, outRadius)
	if inRadius == outRadius {
		return
	}
	if inRadius > outRadius {
		r.Warnings.report(WarnInconsistentRangeOpSkipped, [2]float32{inRadius, outRadius})
		return
	}

	if endRads >= startRads+2*math.Pi {
		var thickChange float32
		if rounding < 0 {
			deltaRadius := (outRadius - inRadius)
			thickChange = min(0, deltaRadius/2.0+rounding)
		} else {
			thickChange = rounding
		}

		thickness := thickChange*2.0 + (outRadius - inRadius)
		if thickness > 0 {
			r.StrokeCircle(target, cx, cy, (outRadius+inRadius)/2.0, thickness)
		}
		return
	}

	startRads, endRads = normURads(startRads), normURads(endRads)
	r.internalFillRadialSector(target, cx, cy, inRadius, outRadius, startRads, endRads, rounding)
}

// precondition: angles are normalized to [0, 2*pi)
func (r *Renderer) internalFillRadialSector(target *ebiten.Image, cx, cy, inRadius, outRadius float32, startRads, endRads float64, rounding float32) {
	if rounding < 0 {
		if inRadius == 0 {
			r.fillCircularSector(target, cx, cy, float64(outRadius), startRads, endRads, float64(rounding))
		} else {
			r.innerRoundingFillAnnularSector(target, float64(cx), float64(cy), float64(inRadius), float64(outRadius), startRads, endRads, float64(-rounding))
		}
		return
	}

	centralAngle := uradsDeltaCW(startRads, endRads)
	halfCtrAngle := centralAngle * 0.5
	centerDir := uradsAddCW(startRads, halfCtrAngle)
	halfCtrAngleSin, halfCtrAngleCos := math.Sincos(halfCtrAngle)

	pieMinX, pieMinY, pieMaxX, pieMaxY := radialSectorBounds(cx, cy, inRadius, outRadius, startRads, endRads)
	if rounding > 0 {
		pieMinX -= rounding
		pieMinY -= rounding
		pieMaxX += rounding
		pieMaxY += rounding
	}

	r.setDstRectCoords(float32(pieMinX), float32(pieMinY), float32(pieMaxX), float32(pieMaxY))
	r.opts.Uniforms["WedgeNormal"] = [2]float32{float32(halfCtrAngleSin), float32(halfCtrAngleCos)}
	r.opts.Uniforms["InRadius"] = inRadius
	r.opts.Uniforms["Rounding"] = rounding
	tox, toy := rectOriginF32(target.Bounds())
	r.setFlatCustomVAs(cx-tox, cy-toy, float32(centerDir), outRadius)
	target.DrawTrianglesShader32(r.vertices[:], r.indices[:], shaderRadialSector.Load(), &r.opts)
	clear(r.opts.Uniforms)
}

// precondition: rounding must be > 0. ws, wc are sin and cos of halfDelta
func (r *Renderer) innerRoundingFillAnnularSector(target *ebiten.Image, cx, cy, inRadius, outRadius, startRads, endRads, rounding float64) {
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
	centralAngle := uradsDeltaCW(startRads, endRads)
	halfCtrAngle := centralAngle * 0.5
	centerDir := uradsAddCW(startRads, halfCtrAngle)
	halfCtrAngleSin, halfCtrAngleCos := math.Sincos(halfCtrAngle)
	dirOffset := rounding / halfCtrAngleSin
	cds, cdc := math.Sincos(centerDir)

	// handle shape collapses
	if dirOffset >= inRadius && halfCtrAngle < math.Pi/2 {
		// note: when aperture is very large, point and circular sector collapse can't
		// happen. we check delta < math.Pi as a broad estimate, but I'm not sure this is
		// always correct / sufficient.

		// collapse into circular sector
		// TODO: I don't think arc collapse can ever be hit here?
		r.innerRoundingFillCircularSector(target, float32(cx), float32(cy), startRads, endRads, outRadius, cds, cdc, halfCtrAngleSin, halfCtrAngleCos, rounding)
		return
	}

	var relInCut, relOutCut [2]float64
	if inRadius > 0 {
		relInCut, _, _ = circIntersect(0, 0, inRadius, halfCtrAngleCos*inRadius, halfCtrAngleSin*inRadius, rounding)
	} else { // inRadius <= 0, treat separately as a collapsed pie
		relInCut[0], relInCut[1] = dirOffset, 0
	}

	relOutCut, _, _ = circIntersect(0, 0, outRadius, halfCtrAngleCos*outRadius, halfCtrAngleSin*outRadius, rounding)
	r.innerFillRadialWedge(target, cx, cy, inRadius, outRadius, centerDir, cds, cdc, relInCut[0], relInCut[1], relOutCut[0], relOutCut[1], rounding-arcCollapse)
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
