package shapes

import (
	"image"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

// NewRect returns a new image filled with the renderer's current color.
func (r *Renderer) NewRect(width, height int) *ebiten.Image {
	img := ebiten.NewImage(width, height)
	r.DrawIntArea(img, 0, 0, width, height)
	return img
}

// NewRect returns a new image with circle of the given radius, drawn
// with the renderer's current color.
func (r *Renderer) NewCircle(radius float64) *ebiten.Image {
	side := float32(math.Ceil(radius * 2))
	img := ebiten.NewImage(int(side), int(side))
	r.DrawCircle(img, side/2, side/2, float32(radius))
	return img
}

// NewRing returns a new image with a ring of the given dimensions, drawn
// with the renderer's current color.
func (r *Renderer) NewRing(inRadius, outRadius float64) *ebiten.Image {
	side := float32(math.Ceil(outRadius * 2))
	img := ebiten.NewImage(int(side), int(side))
	r.DrawRing(img, side/2, side/2, float32(inRadius), float32(outRadius))
	return img
}

// DrawRect is the image.Rectangle compatible equivalent of [Renderer.DrawArea]().
func (r *Renderer) DrawRect(target *ebiten.Image, rect image.Rectangle, rounding float32) {
	if rounding == 0 {
		r.DrawIntArea(target, rect.Min.X, rect.Min.Y, rect.Dx(), rect.Dy())
	} else {
		r.DrawArea(target, float32(rect.Min.X), float32(rect.Min.Y), float32(rect.Dx()), float32(rect.Dy()), rounding)
	}
}

// DrawArea draws a rectangle with the given properties. Rounding can be zero,
// positive for outwards rounding, or negative for inwards rounding.
//
// If you need shadows for rects or capsules drawn with this method, consider
// [Renderer.DrawAreaSoft]().
func (r *Renderer) DrawArea(target *ebiten.Image, ox, oy, w, h, rounding float32) {
	if w < 0 {
		w = -w
		ox -= w
	}
	if h < 0 {
		h = -h
		oy -= h
	}
	if rounding < -max(w, h)*2 {
		return // ignore
	}

	hmargin, vmargin := float32(0.0), float32(0.0)
	if rounding > 0 {
		hmargin = float32(math.Ceil(float64(rounding)))
		vmargin = hmargin
	} else if rounding < 0 {
		// NOTE: collapse values are conservative bounds found empirically,
		// the actual behavior is not linear and harder to narrowly match.
		// actual collapse is closer to [0.5, 1.707]
		const CollapseStart, CollapseEnd = 0.76, 1.86
		if hCS := h * CollapseStart; -rounding > hCS {
			ht := (-rounding - hCS) / (h * (CollapseEnd - CollapseStart))
			hmargin = -w / 2 * min(ht*(h/w), 1.0)
		}
		if wCS := w * CollapseStart; -rounding > wCS {
			wt := (-rounding - wCS) / (w * (CollapseEnd - CollapseStart))
			vmargin = -h / 2 * min(wt*(w/h), 1.0)
		}
	}

	r.setFlatCustomVAs(ox, oy, w, h)
	r.opts.Uniforms["Rounding"] = rounding
	r.DrawRectShader(target, ox, oy, w, h, hmargin, vmargin, shaderRect.Load())
	clear(r.opts.Uniforms)
}

// DrawLine draws a smooth line between the given two points, with rounded ends.
func (r *Renderer) DrawLine(target *ebiten.Image, ox, oy, fx, fy float64, thickness float64) {
	vdx, vdy := fx-ox, fy-oy // non-normalized vector
	vpx, vpy := -vdy, vdx    // perpendicular vector
	length := math.Hypot(vdx, vdy)
	if length == 0 {
		length = 1
	}
	// scale for vector normalization
	scale := (thickness / 2) / length

	// adjust bounding ends to include thickness rounding
	box, boy := ox-vdx*scale, oy-vdy*scale
	bfx, bfy := fx+vdx*scale, fy+vdy*scale

	// compute bounding vertices applying the perpendicular offset
	svpx, svpy := vpx*scale, vpy*scale
	r.vertices[0].DstX = float32(box + svpx)
	r.vertices[0].DstY = float32(boy + svpy)
	r.vertices[1].DstX = float32(bfx + svpx)
	r.vertices[1].DstY = float32(bfy + svpy)
	r.vertices[2].DstX = float32(bfx - svpx)
	r.vertices[2].DstY = float32(bfy - svpy)
	r.vertices[3].DstX = float32(box - svpx)
	r.vertices[3].DstY = float32(boy - svpy)

	r.setFlatCustomVAs(float32(ox), float32(oy), float32(fx), float32(fy))
	r.opts.Uniforms["Thickness"] = float32(thickness)

	// draw shader
	target.DrawTrianglesShader(r.vertices[:], r.indices[:], shaderLine.Load(), &r.opts)
}

// DrawCircle draws a filled circle. Radius can't be negative.
func (r *Renderer) DrawCircle(target *ebiten.Image, cx, cy, radius float32) {
	if radius == 0 {
		return
	}
	if radius < 0 {
		r.Warnings.report(WarnNegativeValueOpSkipped, radius)
		return
	}
	r.setDstRectCoords(cx-radius, cy-radius, cx+radius, cy+radius)
	r.setFlatCustomVAs(cx, cy, radius, 0.0)
	target.DrawTrianglesShader(r.vertices[:], r.indices[:], shaderCircle.Load(), &r.opts)
}

// StrokeCircle draws a circle outline. If thickness > 0, the outline expands [-thickness/2, thickness/2]
// around the radius. If thickness < 0, the outline goes from [-thickness, 0].
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
		r.DrawCircle(target, cx, cy, radius)
		return
	}

	hthickCeil := ceilF32(thickness / 2.0)
	r.setDstRectCoords(cx-radius-hthickCeil, cy-radius-hthickCeil, cx+radius+hthickCeil, cy+radius+hthickCeil)
	r.setFlatCustomVAs(cx, cy, radius, thickness)
	target.DrawTrianglesShader(r.vertices[:], r.indices[:], shaderStrokeCircle.Load(), &r.opts)
}

// DrawRing draws a smooth ring at the given position. For ring segments
// with start and end angles, see [Renderer.DrawRingSector]() instead.
func (r *Renderer) DrawRing(target *ebiten.Image, cx, cy, inRadius, outRadius float32) {
	if inRadius >= outRadius {
		return // skip empty draws
	}
	r.setDstRectCoords(cx-outRadius, cy-outRadius, cx+outRadius, cy+outRadius)
	r.setFlatCustomVAs(cx, cy, outRadius, inRadius)
	target.DrawTrianglesShader(r.vertices[:], r.indices[:], shaderRing.Load(), &r.opts)
}

// DrawRingSector draws a smooth ring segment. See [RadsRight] constants for
// angle conventions and docs.
//
// Only outer rounding is supported at the moment.
func (r *Renderer) DrawRingSector(target *ebiten.Image, cx, cy, inRadius, outRadius float32, startRads, endRads float64, rounding float32) {
	if inRadius >= outRadius || outRadius < 0 || startRads == endRads {
		return // skip empty draws
	}
	if inRadius <= 0 {
		r.DrawCircle(target, cx, cy, outRadius)
	}
	if endRads >= startRads+2*math.Pi {
		r.DrawRing(target, cx, cy, inRadius, outRadius)
	}

	startRads, endRads = normURads(startRads), normURads(endRads)
	r.internalDrawRingSector(target, cx, cy, inRadius, outRadius, startRads, endRads, rounding)
}

// precondition: angles are normalized to [0, 2*pi)
func (r *Renderer) internalDrawRingSector(target *ebiten.Image, cx, cy, inRadius, outRadius float32, startRads, endRads float64, rounding float32) {
	pieMinX, pieMinY, pieMaxX, pieMaxY := ringSectorBounds(cx, cy, inRadius, outRadius, startRads, endRads)
	if rounding != 0 {
		r := abs(rounding)
		pieMinX -= r
		pieMinY -= r
		pieMaxX += r
		pieMaxY += r
	}

	dstOX, dstOY := rectOriginF32(target.Bounds())
	r.vertices[0].DstX = dstOX + pieMinX
	r.vertices[0].DstY = dstOY + pieMinY
	r.vertices[1].DstX = dstOX + pieMaxX
	r.vertices[1].DstY = dstOY + pieMinY
	r.vertices[2].DstX = dstOX + pieMaxX
	r.vertices[2].DstY = dstOY + pieMaxY
	r.vertices[3].DstX = dstOX + pieMinX
	r.vertices[3].DstY = dstOY + pieMaxY

	delta := uradsDeltaCW(startRads, endRads)
	centerDir := uradsAddCW(startRads, delta/2.0)
	ws, wc := math.Sincos(delta / 2.0)
	r.opts.Uniforms["WedgeNormal"] = [2]float32{float32(ws), float32(wc)}
	r.opts.Uniforms["InRadius"] = inRadius
	r.opts.Uniforms["Rounding"] = rounding
	r.setFlatCustomVAs(cx, cy, float32(centerDir), outRadius)
	target.DrawTrianglesShader(r.vertices[:], r.indices[:], shaderRingSector.Load(), &r.opts)
	clear(r.opts.Uniforms)
}

// StrokeRingSector draws the outline of a smooth ring segment. See [RadsRight]
// constants for angle conventions and docs.
//
// Only outer rounding is supported at the moment.
func (r *Renderer) StrokeRingSector(target *ebiten.Image, cx, cy, inRadius, outRadius, thickness float32, startRads, endRads float64, rounding float32) {
	if inRadius >= outRadius || outRadius < 0 || startRads == endRads || thickness <= 0 {
		return // skip empty draws
	}
	if inRadius <= 0 {
		r.StrokeCircle(target, cx, cy, outRadius, thickness)
	}
	if endRads >= startRads+2*math.Pi {
		r.StrokeCircle(target, cx, cy, inRadius, thickness)
		r.StrokeCircle(target, cx, cy, outRadius, thickness)
	}

	startRads, endRads = normURads(startRads), normURads(endRads)
	r.internalStrokeRingSector(target, cx, cy, inRadius, outRadius, thickness, startRads, endRads, rounding)
}

// precondition: angles are normalized to [0, 2*pi)
func (r *Renderer) internalStrokeRingSector(target *ebiten.Image, cx, cy, inRadius, outRadius, thickness float32, startRads, endRads float64, rounding float32) {
	pieMinX, pieMinY, pieMaxX, pieMaxY := ringSectorBounds(cx, cy, inRadius, outRadius, startRads, endRads)
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

	dstOX, dstOY := rectOriginF32(target.Bounds())
	r.vertices[0].DstX = dstOX + pieMinX
	r.vertices[0].DstY = dstOY + pieMinY
	r.vertices[1].DstX = dstOX + pieMaxX
	r.vertices[1].DstY = dstOY + pieMinY
	r.vertices[2].DstX = dstOX + pieMaxX
	r.vertices[2].DstY = dstOY + pieMaxY
	r.vertices[3].DstX = dstOX + pieMinX
	r.vertices[3].DstY = dstOY + pieMaxY

	delta := uradsDeltaCW(startRads, endRads)
	centerDir := uradsAddCW(startRads, delta/2.0)
	ws, wc := math.Sincos(delta / 2.0)
	r.opts.Uniforms["WedgeNormal"] = [2]float32{float32(ws), float32(wc)}
	r.opts.Uniforms["InRadius"] = inRadius
	r.opts.Uniforms["Rounding"] = rounding
	r.opts.Uniforms["Thickness"] = thickness
	r.setFlatCustomVAs(cx, cy, float32(centerDir), outRadius)
	target.DrawTrianglesShader(r.vertices[:], r.indices[:], shaderStrokeRingSector.Load(), &r.opts)
	clear(r.opts.Uniforms)
}

// DrawPie draws circular sector defined by (startRads, endRads).
// See [RadsRight] constants for angle conventions and docs.
//
// Some examples:
//   - startRads = RadsRight, endRads = RadsBottom will draw the bottom-right quarter circle pie.
//   - startRads = RadsBottom, endRads = RadsRight will draw a pie missing the bottom-right quarter.
//
// Notice that the rounding parameter doesn't expand the shape beyond the radius, but it does expand
// the side flat faces of the pie.
func (r *Renderer) DrawPie(target *ebiten.Image, cx, cy, radius float32, startRads, endRads float64, rounding float32) {
	if startRads == endRads || radius < 0 {
		return // empty
	}
	if endRads >= startRads+2*math.Pi {
		r.DrawCircle(target, cx, cy, radius)
		return // full circle
	}
	startRads, endRads = normURads(startRads), normURads(endRads)
	delta := uradsDeltaCW(startRads, endRads)
	centerDir := uradsAddCW(startRads, delta/2.0)
	r.internalDrawPieRate(target, cx, cy, radius, centerDir, startRads, endRads, float64(delta)/(2*math.Pi), rounding)
}

// DrawPieRate is similar to DrawPie, but using a single direction for the center of the pie
// slice and a rate value between (0, 1), with 0 being empty pie and 1 being completely filled.
// See [RadsRight] constants for angle conventions and docs.
func (r *Renderer) DrawPieRate(target *ebiten.Image, cx, cy, radius float32, centerDir, rate float64, rounding float32) {
	if rate <= 0 || radius < 0 {
		return // empty
	}
	if rate > 1.0 {
		r.DrawCircle(target, cx, cy, radius)
		return // full circle
	}

	ratePi := rate * math.Pi
	startRads, endRads := uradsAddCW(centerDir, ratePi), uradsAddCW(centerDir, ratePi)
	r.internalDrawPieRate(target, cx, cy, radius, centerDir, startRads, endRads, rate, rounding)
}

// preconditions: 0 < rate < 1.0, centerDir, startRads, endRads and rate are consistent
func (r *Renderer) internalDrawPieRate(target *ebiten.Image, cx, cy, radius float32, centerDir, startRads, endRads, rate float64, rounding float32) {
	pieMinX, pieMinY, pieMaxX, pieMaxY := pieBounds(cx, cy, radius, startRads, endRads)
	pieMinX -= rounding
	pieMinY -= rounding
	pieMaxX += rounding
	pieMaxY += rounding

	dstOX, dstOY := rectOriginF32(target.Bounds())
	r.setDstRectCoords(dstOX+pieMinX, dstOY+pieMinY, dstOX+pieMaxX, dstOY+pieMaxY)

	ws, wc := math.Sincos(rate * math.Pi)
	r.opts.Uniforms["WedgeNormal"] = [2]float32{float32(ws), float32(wc)}
	r.opts.Uniforms["Rounding"] = rounding
	r.setFlatCustomVAs(cx, cy, float32(normURads(centerDir)), radius)
	target.DrawTrianglesShader(r.vertices[:], r.indices[:], shaderPie.Load(), &r.opts)
	clear(r.opts.Uniforms)
}

// StrokePie is the stroke version of [Renderer.DrawPie](). The shape is drawn with an ouline of the given thickness.
func (r *Renderer) StrokePie(target *ebiten.Image, cx, cy, radius, thickness float32, startRads, endRads float64, rounding float32) {
	if startRads == endRads || radius < 0 {
		return // empty
	}
	if endRads >= startRads+2*math.Pi {
		r.StrokeCircle(target, cx, cy, radius, thickness)
		return // full circle
	}
	startRads, endRads = normURads(startRads), normURads(endRads)
	delta := uradsDeltaCW(startRads, endRads)
	centerDir := uradsAddCW(startRads, delta/2.0)
	r.internalStrokePieRate(target, cx, cy, radius, thickness, centerDir, startRads, endRads, float64(delta)/(2*math.Pi), rounding)
}

// StrokePie is the stroke version of [Renderer.DrawPieRate](). The shape is drawn with an ouline of the given thickness.
func (r *Renderer) StrokePieRate(target *ebiten.Image, cx, cy, radius, thickness float32, centerDir, rate float64, rounding float32) {
	if rate <= 0 || radius < 0 {
		return // empty
	}
	if rate > 1.0 {
		r.StrokeCircle(target, cx, cy, radius, thickness)
		return // full circle
	}

	ratePi := rate * math.Pi
	startRads, endRads := uradsAddCW(centerDir, ratePi), uradsAddCW(centerDir, ratePi)
	r.internalStrokePieRate(target, cx, cy, radius, thickness, centerDir, startRads, endRads, rate, rounding)
}

// preconditions: 0 < rate < 1.0, centerDir, startRads, endRads and rate are consistent
func (r *Renderer) internalStrokePieRate(target *ebiten.Image, cx, cy, radius, thickness float32, centerDir, startRads, endRads, rate float64, rounding float32) {
	pieMinX, pieMinY, pieMaxX, pieMaxY := pieBounds(cx, cy, radius, startRads, endRads)
	pieMinX -= (rounding + thickness)
	pieMinY -= (rounding + thickness)
	pieMaxX += (rounding + thickness)
	pieMaxY += (rounding + thickness)

	dstOX, dstOY := rectOriginF32(target.Bounds())
	r.setDstRectCoords(dstOX+pieMinX, dstOY+pieMinY, dstOX+pieMaxX, dstOY+pieMaxY)

	ws, wc := math.Sincos(rate * math.Pi)
	r.opts.Uniforms["WedgeNormal"] = [2]float32{float32(ws), float32(wc)}
	r.opts.Uniforms["Rounding"] = rounding
	r.opts.Uniforms["Thickness"] = thickness
	r.setFlatCustomVAs(cx, cy, float32(normURads(centerDir)), radius)
	target.DrawTrianglesShader(r.vertices[:], r.indices[:], shaderStrokePie.Load(), &r.opts)
	clear(r.opts.Uniforms)
}

// Notice: ellipses don't have a perfect SDF, so approximations can be very slightly
// bigger or smaller than the requested radiuses.
func (r *Renderer) DrawEllipse(target *ebiten.Image, cx, cy, horzRadius, vertRadius float32, rads float64) {
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
	r.setFlatCustomVAs(cx, cy, horzRadius, vertRadius)
	target.DrawTrianglesShader(r.vertices[:], r.indices[:], shaderEllipse.Load(), &r.opts)
}

// DrawIntArea is a simpler version of [Renderer.DrawArea]() that uses integer coordinates.
//
// See also [Renderer.DrawRect]() when working with image.Rectangle.
func (r *Renderer) DrawIntArea(target *ebiten.Image, ox, oy, w, h int) {
	if w == 0 || h == 0 {
		return
	}
	if w < 0 {
		w = -w
		ox -= w
	}
	if h < 0 {
		h = -h
		oy -= h
	}
	bounds := target.Bounds()
	minX, minY := bounds.Min.X, bounds.Min.Y
	r.setDstRectCoords(float32(minX+ox), float32(minY+oy), float32(minX+ox+w), float32(minY+oy+h))
	target.DrawTrianglesShader(r.vertices[:], r.indices[:], shaderDefault.Load(), &r.opts)
}

// StrokeIntRect is the image.Rectangle compatible equivalent of [Renderer.StrokeIntArea]().
func (r *Renderer) StrokeIntRect(target *ebiten.Image, area image.Rectangle, outThickness, inThickness int) {
	r.StrokeIntArea(target, area.Min.X, area.Min.Y, area.Dx(), area.Dy(), outThickness, inThickness)
}

func (r *Renderer) StrokeIntArea(target *ebiten.Image, ox, oy, w, h, outThickness, inThickness int) {
	if w < 0 {
		w = -w
		ox -= w
	}
	if h < 0 {
		h = -h
		oy -= h
	}
	if outThickness < 0 {
		r.Warnings.report(WarnNegativeValueZeroed, outThickness)
		outThickness = 0
	}
	if inThickness < 0 {
		r.Warnings.report(WarnNegativeValueZeroed, inThickness)
		inThickness = 0
	}
	if outThickness+inThickness == 0 {
		return
	}

	if outThickness == 0 {
		if inThickness != 0 {
			r.strokeIntInnerArea(target, ox, oy, w, h, inThickness)
		}
	} else {
		r.strokeIntInnerArea(target, ox-outThickness, oy-outThickness, w+outThickness*2, h+outThickness*2, outThickness+inThickness)
	}
}

func (r *Renderer) strokeIntInnerArea(target *ebiten.Image, ox, oy, w, h, thickness int) {
	bounds := target.Bounds()
	minX, minY := bounds.Min.X, bounds.Min.Y
	oox, ooy := float32(minX+ox), float32(minY+oy)
	ofx, ofy := float32(minX+ox+w), float32(minY+oy+h)
	r.setDstRectCoords(oox, ooy, ofx, ofy)

	// add inner points
	thickF32 := float32(thickness)
	iox, ioy := oox+thickF32, ooy+thickF32
	ifx, ify := ofx-thickF32, ofy-thickF32
	if r.singleClr || r.opts.Blend == ebiten.BlendClear {
		r.vertices = append(r.vertices,
			ebiten.Vertex{DstX: iox, DstY: ioy},
			ebiten.Vertex{DstX: ifx, DstY: ioy},
			ebiten.Vertex{DstX: ifx, DstY: ify},
			ebiten.Vertex{DstX: iox, DstY: ify},
		)
		for i := range 4 {
			r.vertices[4+i].ColorR = r.vertices[i].ColorR
			r.vertices[4+i].ColorG = r.vertices[i].ColorG
			r.vertices[4+i].ColorB = r.vertices[i].ColorB
			r.vertices[4+i].ColorA = r.vertices[i].ColorA
		}
	} else {
		// we need to interpolate colors. this code takes advantage of
		// the heavy symmetries in the geometry to reduce the number of
		// operations, but as a downside, it's a bit tricky to understand

		// compute uv coords for inner points
		iou := min(max((iox-oox)/(ofx-oox), 0), 1)
		iov := min(max((ioy-ooy)/(ofy-ooy), 0), 1)

		// compute top and bottom left colors
		tR, tG, tB, tA := interpVertexColor(r.vertices[0], r.vertices[1], iou)
		bR, bG, bB, bA := interpVertexColor(r.vertices[3], r.vertices[2], iou)

		// append all vertices with left side colors set
		r.vertices = append(r.vertices,
			ebiten.Vertex{DstX: iox, DstY: ioy},
			ebiten.Vertex{DstX: ifx, DstY: ioy},
			ebiten.Vertex{DstX: ifx, DstY: ify},
			ebiten.Vertex{DstX: iox, DstY: ify},
		)

		tli, tri, bli, bri := 4, 5, 7, 6 // NOTE: use other orders for cool effects
		r.vertices[tli].ColorR = lerp(tR, bR, iov)
		r.vertices[tli].ColorG = lerp(tG, bG, iov)
		r.vertices[tli].ColorB = lerp(tB, bB, iov)
		r.vertices[tli].ColorA = lerp(tA, bA, iov)

		r.vertices[bli].ColorR = lerp(bR, tR, iov)
		r.vertices[bli].ColorG = lerp(bG, tG, iov)
		r.vertices[bli].ColorB = lerp(bB, tB, iov)
		r.vertices[bli].ColorA = lerp(bA, tA, iov)

		// compute right side colors by symmetry
		tR = r.vertices[1].ColorR - (tR - r.vertices[0].ColorR)
		tG = r.vertices[1].ColorG - (tG - r.vertices[0].ColorG)
		tB = r.vertices[1].ColorB - (tB - r.vertices[0].ColorB)
		tA = r.vertices[1].ColorA - (tA - r.vertices[0].ColorA)
		bR = r.vertices[2].ColorR - (bR - r.vertices[3].ColorR)
		bG = r.vertices[2].ColorG - (bG - r.vertices[3].ColorG)
		bB = r.vertices[2].ColorB - (bB - r.vertices[3].ColorB)
		bA = r.vertices[2].ColorA - (bA - r.vertices[3].ColorA)

		// set right vertex colors
		r.vertices[tri].ColorR = lerp(tR, bR, iov)
		r.vertices[tri].ColorG = lerp(tG, bG, iov)
		r.vertices[tri].ColorB = lerp(tB, bB, iov)
		r.vertices[tri].ColorA = lerp(tA, bA, iov)

		r.vertices[bri].ColorR = lerp(bR, tR, iov)
		r.vertices[bri].ColorG = lerp(bG, tG, iov)
		r.vertices[bri].ColorB = lerp(bB, tB, iov)
		r.vertices[bri].ColorA = lerp(bA, tA, iov)
	}

	target.DrawTrianglesShader(r.vertices[:], r.strokeIndices[:], shaderDefault.Load(), &r.opts)
	r.vertices = r.vertices[:4]
}

// StrokeRect is the image.Rectangle compatible equivalent of [Renderer.StrokeArea]().
// When no rounding is required, prefer [Renderer.StrokeIntArea]() instead.
func (r *Renderer) StrokeRect(target *ebiten.Image, rect image.Rectangle, inThickness, outThickness, rounding float32) {
	r.StrokeArea(target, float32(rect.Min.X), float32(rect.Min.Y), float32(rect.Dx()), float32(rect.Dy()), inThickness, outThickness, rounding)
}

// StrokeArea draws an outline on the given area's boundary, with explicit controls for in/out
// border thickness. Rounding controls the outer edge rounding radius (inner radius if negative).
//
// If you have an image.Rectangle for the rect, consider [Renderer.StrokeRect]() instead.
func (r *Renderer) StrokeArea(target *ebiten.Image, ox, oy, w, h, inThickness, outThickness, rounding float32) {
	if w < 0 {
		w = -w
		ox -= w
	}
	if h < 0 {
		h = -h
		oy -= h
	}

	if outThickness < 0 || inThickness < 0 {
		panic("outThickness < 0 || inThickness < 0")
	}

	if outThickness == 0 {
		if inThickness != 0 {
			r.strokeInnerArea(target, ox, oy, w, h, inThickness, rounding)
		}
	} else {
		r.strokeInnerArea(target, ox-outThickness, oy-outThickness, w+outThickness*2, h+outThickness*2, outThickness+inThickness, rounding)
	}
}

func (r *Renderer) strokeInnerArea(target *ebiten.Image, ox, oy, w, h, inThickness, rounding float32) {
	if rounding < 0 {
		// adjust for inner boundary
		rounding = -(rounding - inThickness)
	}

	const SafeMargin = 2.0
	if rounding >= min(w, h)*2.0+SafeMargin {
		return // ignore
	}

	r.setFlatCustomVAs(ox, oy, w, h)
	r.opts.Uniforms["InnerThickness"] = inThickness
	r.opts.Uniforms["Rounding"] = rounding
	r.DrawRectShader(target, ox, oy, w, h, 0, 0, shaderStrokeRect.Load())
	clear(r.opts.Uniforms)
}

// DrawTriangle draws a smooth triangle using the given vertices and an optional rounding factor.
//
// Rounding can be positive for outwards rounding, or negative for inwards rounding. Notice that
// inwards rounding is non-trivial (two dozen f64 products and 3 square roots for CPU-side
// precomputations).
func (r *Renderer) DrawTriangle(target *ebiten.Image, points [3]PointF32, rounding float32) {
	r.drawTriangle(target, points, 0.0, rounding)
}

// StrokeTriangle draws an unfilled triangle. The outline will expand [-thickness/2, +thickness/2] around
// the given points, unless the passed thickness is negative, in which case the outline will be interior
// only, going from [-thickness, 0].
//
// For more details on rounding, see [Renderer.DrawTriangle]().
func (r *Renderer) StrokeTriangle(target *ebiten.Image, points [3]PointF32, thickness, rounding float32) {
	if thickness == 0 {
		return
	}
	r.drawTriangle(target, points, thickness, rounding)
}

// TODO: skip if inner rounding collapses the shape (currently it degenerates), test all orientations
func (r *Renderer) drawTriangle(target *ebiten.Image, points [3]PointF32, thickness, rounding float32) {
	area := abs((points[0].X*(points[1].Y-points[2].Y) + points[1].X*(points[2].Y-points[0].Y) + points[2].X*(points[0].Y-points[1].Y)) / 2)
	if area < 1e-6 {
		return // empty triangle
	}

	var p0, p1, p2 PointF32 = points[0], points[1], points[2]
	if rounding < 0 {
		a12, b12, c12 := toLinearFormABC(points[0].X, points[0].Y, points[1].X, points[1].Y)
		a23, b23, c23 := toLinearFormABC(points[1].X, points[1].Y, points[2].X, points[2].Y)
		a31, b31, c31 := toLinearFormABC(points[2].X, points[2].Y, points[0].X, points[0].Y)
		c1_12, c2_12, l12 := parallelsAtDist(a12, b12, c12, -rounding)
		c1_23, c2_23, l23 := parallelsAtDist(a23, b23, c23, -rounding)
		c1_31, c2_31, l31 := parallelsAtDist(a31, b31, c31, -rounding)

		// handle collapse into circle
		area := 0.5 * abs((p0.X*(p1.Y-p2.Y) + p1.X*(p2.Y-p0.Y) + p2.X*(p0.Y-p1.Y)))
		perimeter := l12 + l23 + l31
		maxRounding := area / (0.5 * perimeter)
		if -rounding >= maxRounding {
			cx := (l23*p0.X + l31*p1.X + l12*p2.X) / perimeter
			cy := (l23*p0.Y + l31*p1.Y + l12*p2.Y) / perimeter
			radius := maxRounding + (maxRounding + rounding)
			if thickness == 0 {
				if radius <= 0 {
					return
				}
				r.DrawCircle(target, cx, cy, radius)
			} else {
				if thickness < 0 {
					thickness = -thickness
					radius -= thickness / 2.0
				}
				r.StrokeCircle(target, cx, cy, radius, thickness)
			}
			return
		}

		// no collapse, finish calculations
		if a12*points[2].X+b12*points[2].Y+c12 > 0 { // fancy winding order test
			c12, c23, c31 = c1_12, c1_23, c1_31
		} else {
			c12, c23, c31 = c2_12, c2_23, c2_31
		}
		p0.X, p0.Y = shortCramer(a31, b31, c31, a12, b12, c12)
		p1.X, p1.Y = shortCramer(a12, b12, c12, a23, b23, c23)
		p2.X, p2.Y = shortCramer(a23, b23, c23, a31, b31, c31)
	}

	minX, maxX := min(points[0].X, points[1].X, points[2].X), max(points[0].X, points[1].X, points[2].X)
	minY, maxY := min(points[0].Y, points[1].Y, points[2].Y), max(points[0].Y, points[1].Y, points[2].Y)
	hthick := max(thickness/2.0, 0)
	outRounding := max(rounding, 0)
	r.setDstRectCoords(minX-hthick-outRounding, minY-hthick-outRounding, maxX+hthick+outRounding, maxY+hthick+outRounding)

	// draw shader
	r.opts.Uniforms["P0"] = [2]float32{p0.X, p0.Y}
	r.opts.Uniforms["P1"] = [2]float32{p1.X, p1.Y}
	r.opts.Uniforms["P2"] = [2]float32{p2.X, p2.Y}
	r.setFlatCustomVAs01(abs(rounding), thickness)
	target.DrawTrianglesShader(r.vertices[:], r.indices[:], shaderTriangle.Load(), &r.opts)
	clear(r.opts.Uniforms)
}

// Sqrt3Div2 is commonly used to derive a hexagon's apothem from its radius, or
// viceversa (apothem = radius*Sqrt3Div2).
const Sqrt3Div2 = 0.86602540378443864676372317075293618347140262690519031402790348 // https://oeis.org/A010527

// DrawHexagon renders an hexagon that can be fully contained within the given radius.
// Roundness can be used to round the corners. Rads can be used to rotate the hexagon,
// in radians.
//
// Roundness must be non-negative. When > 0, the sides of the hexagon will expand while
// the radius of the shape is maintained, effectively rounding the vertices. Roundness
// >= 'radius' will turn the hexagon into a perfect circle and start increasing the
// effective radius. For inwards/outwards rounding, consider [Renderer.DrawHexagonApothem]()
// instead.
func (r *Renderer) DrawHexagon(target *ebiten.Image, cx, cy, radius, roundness, rads float32) {
	if roundness < 0 {
		r.Warnings.report(WarnNegativeValueZeroed, roundness)
		roundness = 0.0
	}
	if radius == 0 {
		return
	}
	if radius < 0 {
		r.Warnings.report(WarnNegativeValueOpSkipped, radius)
		return
	}

	if roundness >= radius {
		r.DrawCircle(target, cx, cy, roundness)
		return
	}

	r.setDstRectCoords(cx-radius, cy-radius, cx+radius, cy+radius)
	apothem := (radius - roundness) * Sqrt3Div2
	r.setFlatCustomVAs(cx, cy, apothem, rads)
	r.opts.Uniforms["Rounding"] = roundness
	target.DrawTrianglesShader(r.vertices[:], r.indices[:], shaderHexagon.Load(), &r.opts)
}

// DrawHexagonApothem is an alternative form to [Renderer.DrawHexagon]() that requires the apothem
// of the hexagon instead of its radius and supports signed rounding.
//
// Rounding values above 0 will increase the effective apothem by that amount while rounding corners
// outwards. Values between 0 and -apothem will preserve the apothem while rounding corners inwards.
// Values between -apothem and -2*apothem draw a perfect circle.
func (r *Renderer) DrawHexagonApothem(target *ebiten.Image, ox, oy, apothem, rounding, rads float32) {
	if apothem == 0 {
		return
	}
	if apothem < 0 {
		r.Warnings.report(WarnNegativeValueOpSkipped, apothem)
		return
	}
	inset := -(apothem + rounding)
	if inset >= 0 {
		r.DrawCircle(target, ox, oy, apothem-inset)
		return
	}
	boundingApothem := apothem + max(rounding, 0)
	radius := boundingApothem / Sqrt3Div2
	r.setDstRectCoords(ox-radius, oy-radius, ox+radius, oy+radius)
	r.setFlatCustomVAs(ox, oy, apothem, rads)
	r.opts.Uniforms["Rounding"] = rounding
	target.DrawTrianglesShader(r.vertices[:], r.indices[:], shaderHexagon.Load(), &r.opts)
}

// DrawQuad renders a convex quad with the current renderer colors.
// The thickening acts as a rounding parameter, but it extends the shape outwards
// instead of "cutting" the corners. Notice that non-zero thickening involves
// additional CPU-side precomputations.
//
// quad must be given in clockwise order starting from top-left.
func (r *Renderer) DrawQuad(target *ebiten.Image, quad [4]PointF32, thickening float32) {
	r.DrawQuadSoft(target, quad, thickening, 1.3333)
}

func (r *Renderer) DrawQuadSoft(target *ebiten.Image, quad [4]PointF32, thickening, softEdge float32) {
	minX, minY, _, _ := rectOriginSizeF32(target.Bounds())
	for i, pt := range expandQuad(quad, thickening) {
		r.vertices[i].DstX = minX + pt.X
		r.vertices[i].DstY = minY + pt.Y
	}

	r.setFlatCustomVAs01(thickening, softEdge)
	r.opts.Uniforms["Quad"] = [8]float32{
		quad[0].X, quad[0].Y, quad[1].X, quad[1].Y,
		quad[2].X, quad[2].Y, quad[3].X, quad[3].Y,
	}
	target.DrawTrianglesShader(r.vertices[:], r.indices[:], shaderQuad.Load(), &r.opts)
	clear(r.opts.Uniforms)
}

// DrawAreaSoft draws a rect like [Renderer.DrawArea]() but with an extra softRadius, which
// creates a shadow-like soft edge. This is ideal for rendering rect shadows in UIs avoiding
// the more expensive raster-based blurs.
//
// Rounding can be zero, positive for outwards rounding, or negative for inwards rounding.
// SoftRadius extends beyond the boundary.
func (r *Renderer) DrawAreaSoft(target *ebiten.Image, ox, oy, w, h, rounding, softRadius float32) {
	if w < 0 {
		w = -w
		ox -= w
	}
	if h < 0 {
		h = -h
		oy -= h
	}

	if rounding > 0 {
		r2 := rounding + rounding
		w, h = w+r2, h+r2
		ox -= rounding
		oy -= rounding
		rounding = -rounding
	}

	if rounding+max(softRadius, 0) < -max(w, h)*2 {
		return // ignore
	}

	r.setFlatCustomVAs(ox, oy, w, h)
	r.opts.Uniforms["Rounding"] = -rounding
	r.opts.Uniforms["SoftRadius"] = softRadius
	margin := max(softRadius, 0)
	r.DrawRectShader(target, ox, oy, w, h, margin, margin, shaderRectSoft.Load())
	clear(r.opts.Uniforms)
}

// DrawAreaBlur behaves very similarly to [Renderer.DrawAreaSoft](), but accepts only non-negative
// blur radiuses, and "blurs" around the boundary instead of before or after it.
func (r *Renderer) DrawAreaBlur(target *ebiten.Image, ox, oy, w, h, rounding, blurRadius float32) {
	if w < 0 {
		w = -w
		ox -= w
	}
	if h < 0 {
		h = -h
		oy -= h
	}

	if rounding > 0 {
		r2 := rounding + rounding
		w, h = w+r2, h+r2
		ox -= rounding
		oy -= rounding
		rounding = -rounding
	}

	if blurRadius < 0 {
		r.Warnings.report(WarnNegativeValueZeroed, blurRadius)
		blurRadius = 0.0
	}

	if rounding+blurRadius < -max(w, h)*2 {
		return // ignore
	}

	r.setFlatCustomVAs(ox, oy, w, h)
	r.opts.Uniforms["Rounding"] = -rounding
	r.opts.Uniforms["BlurRadius"] = blurRadius
	r.DrawRectShader(target, ox, oy, w, h, blurRadius, blurRadius, shaderRectBlur.Load())
	clear(r.opts.Uniforms)
}
