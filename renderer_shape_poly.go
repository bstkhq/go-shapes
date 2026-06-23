package shapes

import (
	"image"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

// this file contains Fill* and Stroke* functions for basic polygons:
// rectangles, triangles, quads, hexagons, ... for circular shapes,
// see renderer_shape_circ.go instead

// NewFilledRect returns a new image filled with the renderer's current color.
func (r *Renderer) NewFilledRect(width, height int) *ebiten.Image {
	img := ebiten.NewImage(width, height)
	r.internalFillIntRect(img, 0, 0, width, height)
	return img
}

// FillIntRect is the [image.Rectangle]-compatible equivalent of [Renderer.FillRect]().
//
// For rectangle creation, consider [image.Rect]() and [RectWithSize]().
func (r *Renderer) FillIntRect(target *ebiten.Image, rect image.Rectangle, rounding float32) {
	if rounding == 0 {
		r.internalFillIntRect(target, rect.Min.X, rect.Min.Y, rect.Dx(), rect.Dy())
	} else {
		r.FillRect(target, float32(rect.Min.X), float32(rect.Min.Y), float32(rect.Dx()), float32(rect.Dy()), rounding)
	}
}

func (r *Renderer) internalFillIntRect(target *ebiten.Image, ox, oy, w, h int) {
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
	r.setDstRectCoords(float32(ox), float32(oy), float32(ox+w), float32(oy+h))
	target.DrawTrianglesShader32(r.vertices[:], r.indices[:], shaderDefault.Load(), &r.opts)
}

// FillRect draws a filled rectangle with the given properties. Rounding can be zero,
// positive for outwards rounding, or negative for inwards rounding.
//
// If you need shadows for rects or capsules drawn with this method, consider
// [Renderer.FillRectSoft]().
func (r *Renderer) FillRect(target *ebiten.Image, ox, oy, w, h, rounding float32) {
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

	tox, toy := rectOriginF32(target.Bounds())
	r.setFlatCustomVAs(ox-tox, oy-toy, w, h)
	r.opts.Uniforms["Rounding"] = rounding
	margins := NewMargins(hmargin, vmargin)
	r.DrawRectShader(target, ox, oy, w, h, margins, shaderRect.Load())
	clear(r.opts.Uniforms)
}

// FillRectSoft draws a rect like [Renderer.FillRect]() but with a soft edge.
// This is ideal for rect shadows in UIs and avoiding the more expensive
// raster-based blurs.
//
// Rounding can be zero, positive for outwards rounding, or negative for
// inwards rounding. When positive, the soft radius will extend [-softEdge,
// +softEdge] around the boundary, closely approximating a gaussian blur. When
// negative, the softening will extend inwards [-softEdge, 0].
func (r *Renderer) FillRectSoft(target *ebiten.Image, ox, oy, w, h, rounding, softEdge float32) {
	if w < 0 {
		w = -w
		ox -= w
	}
	if h < 0 {
		h = -h
		oy -= h
	}

	// always make rounding negative (inwards)
	if rounding > 0 {
		r2 := rounding + rounding
		w, h = w+r2, h+r2
		ox -= rounding
		oy -= rounding
		rounding = -rounding
	}

	// collapse case
	if rounding+max(softEdge, 0) < -max(w, h)*2 {
		return // ignore
	}

	margin := max(softEdge, 0)
	var shader *ebiten.Shader
	if softEdge > 0 {
		rounding -= softEdge / 1.65 // empirical adjustment
		shader = shaderRectSoftBlur.Load()
	} else {
		softEdge = -softEdge
		shader = shaderRectSoftIn.Load()
	}

	tox, toy := rectOriginF32(target.Bounds())
	r.setFlatCustomVAs(ox-tox, oy-toy, w, h)
	r.opts.Uniforms["InRounding"] = -rounding
	r.opts.Uniforms["BlurRadius"] = softEdge
	r.DrawRectShader(target, ox, oy, w, h, NewMargins(margin, margin), shader)
	clear(r.opts.Uniforms)
}

// StrokeLine draws a smooth line between the given two points, with rounded ends.
//
// Supported flags: [AABB], [ColorAABB].
func (r *Renderer) StrokeLine(target *ebiten.Image, origin, end PointF32, thickness float32, flags ...Flag) {
	if thickness == 0 {
		return
	}
	if thickness < 0 {
		r.Warnings.report(WarnNegativeValueOpSkipped, thickness)
		return
	}

	bounding, colorMode := r.readBoundingAndColorModeFlags(AABB, ColorAABB, flags...)
	if bounding == AABB {
		r.strokeAABBLine(target, origin, end, thickness)
	} else {
		r.strokeHullLine(target, origin, end, thickness, colorMode)
	}
}

func (r *Renderer) strokeAABBLine(target *ebiten.Image, origin, end PointF32, thickness float32) {
	halfThick := thickness / 2.0
	minX, maxX := floorF32(min(origin.X, end.X)-halfThick), ceilF32(max(origin.X, end.X)+halfThick)
	minY, maxY := floorF32(min(origin.Y, end.Y)-halfThick), ceilF32(max(origin.Y, end.Y)+halfThick)
	r.setDstRectCoords(minX, minY, maxX, maxY)
	tox, toy := rectOriginF32(target.Bounds())
	r.setFlatCustomVAs(origin.X-tox, origin.Y-toy, end.X-tox, end.Y-toy)
	r.opts.Uniforms["Thickness"] = float32(thickness)
	target.DrawTrianglesShader32(r.vertices[:], r.indices[:], shaderLine.Load(), &r.opts)
	clear(r.opts.Uniforms)
}

func (r *Renderer) strokeHullLine(target *ebiten.Image, origin, end PointF32, thickness float32, colorMode Flag) {
	const padOffset = 0.333 // to prevent diagonal clipping (affects color interpolation)
	quad := lineToQuad(origin, end, thickness, padOffset)
	r.vertices[0].DstX, r.vertices[0].DstY = quad[0].X, quad[0].Y
	r.vertices[1].DstX, r.vertices[1].DstY = quad[1].X, quad[1].Y
	r.vertices[2].DstX, r.vertices[2].DstY = quad[2].X, quad[2].Y
	r.vertices[3].DstX, r.vertices[3].DstY = quad[3].X, quad[3].Y

	// apply ColorAABB if requested
	var memo [16]float32
	hasMemo := false
	if colorMode == ColorAABB && !r.singleClr {
		memo = r.memorizeColors()
		hasMemo = true
		halfThick := thickness / 2.0
		minX, maxX := min(origin.X, end.X)-halfThick, max(origin.X, end.X)+halfThick
		minY, maxY := min(origin.Y, end.Y)-halfThick, max(origin.Y, end.Y)+halfThick
		r.applyTriQuadColors(minX, minY, maxX, maxY, memo)
	}

	// render
	tox, toy := rectOriginF32(target.Bounds())
	r.setFlatCustomVAs(origin.X-tox, origin.Y-toy, end.X-tox, end.Y-toy)
	r.opts.Uniforms["Thickness"] = float32(thickness)
	target.DrawTrianglesShader32(r.vertices[:], r.indices[:], shaderLine.Load(), &r.opts)

	clear(r.opts.Uniforms)
	if hasMemo {
		r.restoreColors(memo)
	}
}

var rectStrokeIndices = []uint32{
	0, 1, 4,
	4, 1, 5,
	5, 1, 2,
	5, 2, 6,
	6, 2, 3,
	6, 3, 7,
	7, 3, 0,
	0, 4, 7,
}

// StrokeRect is the [image.Rectangle]-compatible equivalent of [Renderer.StrokeRect]().
//
// For rectangle creation, consider [image.Rect]() and [RectWithSize]().
//
// Supported flags: [ColorIntrinsic] (colors "twist" around the border along
// the direction of the underlying triangles).
func (r *Renderer) StrokeIntRect(target *ebiten.Image, rect image.Rectangle, inThickness, outThickness int, flags ...Flag) {
	ox, oy := rect.Min.X, rect.Min.Y
	w, h := rect.Dx(), rect.Dy()
	if w < 0 {
		w = -w
		ox -= w
	}
	if h < 0 {
		h = -h
		oy -= h
	}

	colorMode := r.readOptInFlag(ColorIntrinsic, flags...)
	if colorMode == noFlag {
		colorMode = ColorAABB
	}
	outThickness = warnZeroNegativeValue(r, outThickness)
	inThickness = warnZeroNegativeValue(r, inThickness)
	if outThickness+inThickness == 0 {
		return
	}

	if outThickness > 0 {
		ox -= outThickness
		oy -= outThickness
		w += outThickness * 2
		h += outThickness * 2
		inThickness += outThickness
	}
	r.strokeIntInnerRect(target, ox, oy, w, h, inThickness, colorMode)
}

func (r *Renderer) strokeIntInnerRect(target *ebiten.Image, ox, oy, w, h, thickness int, colorMode Flag) {
	oox, ooy := float32(ox), float32(oy)
	ofx, ofy := float32(ox+w), float32(oy+h)
	r.setDstRectCoords(oox, ooy, ofx, ofy)

	// add inner points
	thickF32 := float32(thickness)
	iox, ioy := oox+thickF32, ooy+thickF32
	ifx, ify := ofx-thickF32, ofy-thickF32
	r.vertices = append(r.vertices,
		ebiten.Vertex{DstX: iox, DstY: ioy},
		ebiten.Vertex{DstX: ifx, DstY: ioy},
		ebiten.Vertex{DstX: ifx, DstY: ify},
		ebiten.Vertex{DstX: iox, DstY: ify},
	)
	if r.singleClr || r.opts.Blend == ebiten.BlendClear {
		for i := range 4 {
			r.vertices[4+i].ColorR = r.vertices[i].ColorR
			r.vertices[4+i].ColorG = r.vertices[i].ColorG
			r.vertices[4+i].ColorB = r.vertices[i].ColorB
			r.vertices[4+i].ColorA = r.vertices[i].ColorA
		}
	} else {
		tl, tr, br, bl := r.vertexColors(0), r.vertexColors(1), r.vertexColors(2), r.vertexColors(3)
		origin, size := PtF32(oox, ooy), PtF32(float32(w), float32(h))
		indexOffset := 0
		if colorMode == ColorIntrinsic {
			indexOffset = 3 // "twist" colors following triangle directions. this is just a fancy effect
		}
		for i := range 4 {
			clr := interpTriQuadColor(tl, tr, br, bl, origin, size, PtF32(r.vertices[4+i].DstX, r.vertices[4+i].DstY))
			ci := 4 + (i+indexOffset)%4
			setVertexColor(&r.vertices[ci], clr[0], clr[1], clr[2], clr[3])
		}
	}

	target.DrawTrianglesShader32(r.vertices[:], rectStrokeIndices[:], shaderDefault.Load(), &r.opts)
	r.vertices = r.vertices[:4]
}

// StrokeRect draws an outline on the given area's boundary, with explicit controls for in/out
// border thickness. Rounding controls the outer edge rounding radius (inner radius if negative).
//
// If you have an [image.Rectangle] for the rect, consider [Renderer.StrokeIntRect]() instead,
// or if you need rounding convert with [RectPointsF32]().
//
// Supported flags: [AABB].
func (r *Renderer) StrokeRect(target *ebiten.Image, ox, oy, w, h, inThickness, outThickness, rounding float32, flags ...Flag) {
	if w < 0 {
		w = -w
		ox -= w
	}
	if h < 0 {
		h = -h
		oy -= h
	}

	boundingMode := r.readOptInFlag(AABB, flags...)
	if boundingMode == noFlag {
		boundingMode = Hull
	}
	outThickness = warnZeroNegativeValue(r, outThickness)
	inThickness = warnZeroNegativeValue(r, inThickness)
	if outThickness+inThickness == 0 {
		return
	}

	if outThickness > 0 {
		ox -= outThickness
		oy -= outThickness
		w += outThickness * 2
		h += outThickness * 2
		inThickness += outThickness
	}
	r.strokeInnerRect(target, ox, oy, w, h, inThickness, rounding, boundingMode)
}

func (r *Renderer) strokeInnerRect(target *ebiten.Image, ox, oy, w, h, inThickness, rounding float32, boundingMode Flag) {
	if rounding < 0 {
		// adjust for inner boundary
		rounding = -(rounding - inThickness)
	}

	const SafeMargin = 2.0
	if rounding >= min(w, h)*2.0+SafeMargin {
		return // ignore
	}

	tox, toy := rectOriginF32(target.Bounds())
	r.opts.Uniforms["InnerThickness"] = inThickness
	r.opts.Uniforms["Rounding"] = rounding
	inRounding := max(rounding-inThickness, 0)
	if boundingMode == Hull && (w >= 2*inRounding || h >= 2*inRounding) {
		r.setDstRectCoords(floorF32(ox), floorF32(oy), ceilF32(ox+w), ceilF32(oy+h))
		iox, ioy := ox+inThickness, oy+inThickness
		ifx, ify := ox+w-inThickness, oy+h-inThickness

		// note: the point at the center of the inner curve is often more
		// optimal for square-ish rects, but this is simple and reasonable
		if w >= h {
			iox += inRounding
			ifx -= inRounding
		} else {
			ioy += inRounding
			ify -= inRounding
		}
		iox, ioy = ceilF32(iox), ceilF32(ioy)
		ifx, ify = floorF32(ifx), floorF32(ify)
		r.vertices = append(r.vertices,
			ebiten.Vertex{DstX: iox, DstY: ioy},
			ebiten.Vertex{DstX: ifx, DstY: ioy},
			ebiten.Vertex{DstX: ifx, DstY: ify},
			ebiten.Vertex{DstX: iox, DstY: ify},
		)

		tl, tr, br, bl := r.vertexColors(0), r.vertexColors(1), r.vertexColors(2), r.vertexColors(3)
		origin, size := PtF32(ox, oy), PtF32(w, h)
		for i := range 4 {
			clr := interpTriQuadColor(tl, tr, br, bl, origin, size, PtF32(r.vertices[4+i].DstX, r.vertices[4+i].DstY))
			setVertexColor(&r.vertices[4+i], clr[0], clr[1], clr[2], clr[3])
		}

		r.setFlatCustomVAs(ox-tox, oy-toy, w, h)
		target.DrawTrianglesShader32(r.vertices[:], rectStrokeIndices[:], shaderStrokeRect.Load(), &r.opts)
		r.vertices = r.vertices[:4]
	} else { // assume AABB
		r.setFlatCustomVAs(ox-tox, oy-toy, w, h)
		r.DrawRectShader(target, ox, oy, w, h, NoMargins, shaderStrokeRect.Load())
	}
	clear(r.opts.Uniforms)
}

// FillTriangle draws a smooth filled triangle using the given vertices and an optional rounding factor.
//
// Rounding can be positive for outwards rounding, or negative for inwards rounding. Notice that
// inwards rounding requires non-trivial CPU precalculations.
func (r *Renderer) FillTriangle(target *ebiten.Image, points [3]PointF32, rounding float32) {
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

// TODO: support Hull and ColorAABB flags
func (r *Renderer) drawTriangle(target *ebiten.Image, points [3]PointF32, thickness, rounding float32) {
	points, shape, rounding := preprocessTriangle(points, rounding)
	if shape == shapePoint {
		points[1], points[2] = points[0], points[0]
		rounding, thickness = cleanStrokeRadiusThickness(rounding, thickness)
		if rounding+thickness <= 0 {
			return
		}
	}

	minX, maxX := min(points[0].X, points[1].X, points[2].X), max(points[0].X, points[1].X, points[2].X)
	minY, maxY := min(points[0].Y, points[1].Y, points[2].Y), max(points[0].Y, points[1].Y, points[2].Y)
	margin := max(thickness/2.0, 0) + max(rounding, 0)
	r.setDstRectCoords(floorF32(minX-margin), floorF32(minY-margin), ceilF32(maxX+margin), ceilF32(maxY+margin))

	// draw shader
	tox, toy := rectOriginF32(target.Bounds())
	r.opts.Uniforms["P0"] = [2]float32{points[0].X - tox, points[0].Y - toy}
	r.opts.Uniforms["P1"] = [2]float32{points[1].X - tox, points[1].Y - toy}
	r.opts.Uniforms["P2"] = [2]float32{points[2].X - tox, points[2].Y - toy}

	var shader *ebiten.Shader
	if thickness == 0 {
		shader = shaderTriangle.Load()
		r.setFlatCustomVA0(abs(rounding))
	} else {
		shader = shaderTriangleStroke.Load()
		var thickOffset float32
		if thickness < 0 {
			thickOffset = thickness
			thickness = -thickness
		} else {
			thickOffset = -thickness / 2.0
		}
		r.setFlatCustomVAs(abs(rounding), thickness, thickOffset, 0.0)
	}

	target.DrawTrianglesShader32(r.vertices[:], r.indices[:], shader, &r.opts)
	clear(r.opts.Uniforms)
}

// preprocessTriangle handles inner rounding, shrinking the geometry and
// converting to outer rounding while also handling collapse cases. notice
// that outer rounding can still be negative, as thickness might have to
// be applied on top.
//
// when normalizing as CW, the point 0 is always preserved, and the other
// 2 are swapped
func preprocessTriangle(points [3]PointF32, rounding float32) ([3]PointF32, shapeType, float32) {
	area := triangleSignedArea(points[0], points[1], points[2])
	if area < 0 { // normalize as CW
		points[1], points[2] = points[2], points[1]
		area = -area
	}
	if area < 1e-6 {
		midpoint := points[0].Add(points[1]).Add(points[2]).Scale(1.0 / 3.0)
		return [3]PointF32{midpoint}, shapePoint, rounding // notice: rounding result can be negative
	}

	if rounding > 0 {
		return points, shapeTriangle, rounding
	}

	var shape shapeType
	var offsetReached float32
	points[0], points[1], points[2], shape, offsetReached = shrinkTriangle(points[0], points[1], points[2], area, -rounding)
	if shape == shapePoint {
		return points, shapePoint, offsetReached*2 + rounding // notice: rounding result can be negative
	}
	return points, shapeTriangle, -rounding
}

// Sqrt3Div2 is commonly used to derive a hexagon's apothem from its radius, or
// viceversa (apothem = radius*Sqrt3Div2).
const Sqrt3Div2 = 0.86602540378443864676372317075293618347140262690519031402790348 // https://oeis.org/A010527

// FillHexagon renders an hexagon that can be fully contained within the given radius.
//
// Roundness must be non-negative. When > 0, the sides of the hexagon will expand while
// the radius of the shape is maintained, effectively rounding the vertices. Roundness
// >= radius will turn the hexagon into a perfect circle and start increasing the
// effective radius. For inwards/outwards rounding, see [Renderer.FillHexagonApothem]()
// instead.
//
// Rads can be used to rotate the hexagon. See [RadsRight] constants for angle conventions and docs.
func (r *Renderer) FillHexagon(target *ebiten.Image, cx, cy, radius, roundness, rads float32) {
	roundness = warnZeroNegativeValue(r, roundness)
	if radius == 0 {
		return
	}
	if radius < 0 {
		r.Warnings.report(WarnNegativeValueOpSkipped, radius)
		return
	}

	if roundness >= radius {
		r.FillCircle(target, cx, cy, roundness)
		return
	}

	r.setDstRectCoords(cx-radius, cy-radius, cx+radius, cy+radius)
	apothem := (radius - roundness) * Sqrt3Div2
	tox, toy := rectOriginF32(target.Bounds())
	r.setFlatCustomVAs(cx-tox, cy-toy, apothem, rads)
	r.opts.Uniforms["Rounding"] = roundness
	target.DrawTrianglesShader32(r.vertices[:], r.indices[:], shaderHexagon.Load(), &r.opts)
	clear(r.opts.Uniforms)
}

// FillHexagonApothem is an alternative form to [Renderer.FillHexagon]() that defines the apothem
// of the hexagon instead of its radius.
//
// Rounding values above 0 will increase the effective apothem by that amount while rounding corners
// outwards. Values between 0 and -apothem will preserve the apothem while rounding corners inwards.
// Values below -apothem collapse the shape into a circle.
func (r *Renderer) FillHexagonApothem(target *ebiten.Image, ox, oy, apothem, rounding, rads float32) {
	if apothem == 0 {
		return
	}
	if apothem < 0 {
		r.Warnings.report(WarnNegativeValueOpSkipped, apothem)
		return
	}
	inset := -(apothem + rounding)
	if inset >= 0 {
		r.FillCircle(target, ox, oy, apothem-inset)
		return
	}
	boundingApothem := apothem + max(rounding, 0)
	radius := boundingApothem / Sqrt3Div2
	r.setDstRectCoords(ox-radius, oy-radius, ox+radius, oy+radius)
	tox, toy := rectOriginF32(target.Bounds())
	r.setFlatCustomVAs(ox-tox, oy-toy, apothem, rads)
	r.opts.Uniforms["Rounding"] = rounding
	target.DrawTrianglesShader32(r.vertices[:], r.indices[:], shaderHexagon.Load(), &r.opts)
	clear(r.opts.Uniforms)
}

// FillQuad renders a quadrilateral with the current renderer colors.
//
// Rounding can be zero, positive for outwards rounding, or negative for
// inwards rounding. Notice that non-zero rounding or self-intersecting quads
// trigger additional precomputations, which can be particularly complex for
// inner rounding.
func (r *Renderer) FillQuad(target *ebiten.Image, quad [4]PointF32, rounding float32) {
	var simple bool
	quad, simple = canonicalizeQuadCW(quad)
	if !simple {
		r.fillSelfIntersectingQuad(target, quad, rounding)
		return
	}

	if rounding < 0 {
		var nonEmpty bool
		quad, rounding, nonEmpty = innerRoundQuad(quad, rounding)
		if !nonEmpty {
			return
		}
	}

	// if opts.include(Hull) {
	//     r.hullBuff = quadHull(r.hullBuff, quad, rounding)
	// }
	r.internalFillQuad(target, quad, shaderQuad.Load(), rounding, 0.0)
}

// the returned bool indicates whether the result is non-empty
func innerRoundQuad(quad [4]PointF32, rounding float32) ([4]PointF32, float32, bool) {
	if rounding < 0 {
		var shape shapeType
		var offsetReached float32
		quad, shape, offsetReached = offsetQuad(quad, rounding)
		switch shape {
		case shapePoint:
			radius := rounding - offsetReached*2
			quad[1], quad[2], quad[3] = quad[0], quad[0], quad[0]
			rounding = radius
			if radius <= 0 {
				return quad, 0.0, false
			}
		case shapeLine:
			radius := rounding - offsetReached*2
			quad[2], quad[3] = quad[1], quad[0]
			rounding = radius
			if radius <= 0 {
				return quad, 0.0, false
			}
		case shapeTriangle:
			quad[3] = quad[2] // repeat one of the points
			rounding = -rounding
		case shapeQuad:
			rounding = -rounding
		default:
			panic(shape) // broken code
		}
	}
	return quad, rounding, true
}

// precondition: rounding >= 0
func (r *Renderer) internalFillQuad(target *ebiten.Image, quad [4]PointF32, shader *ebiten.Shader, rounding, softEdge float32) {
	margin := rounding + max(softEdge, 0)
	minX, maxX := floorF32(min(quad[0].X, quad[1].X, quad[2].X, quad[3].X)-margin), ceilF32(max(quad[0].X, quad[1].X, quad[2].X, quad[3].X)+margin)
	minY, maxY := floorF32(min(quad[0].Y, quad[1].Y, quad[2].Y, quad[3].Y)-margin), ceilF32(max(quad[0].Y, quad[1].Y, quad[2].Y, quad[3].Y)+margin)
	r.setDstRectCoords(minX, minY, maxX, maxY)

	tox, toy := rectOriginF32(target.Bounds())
	r.setFlatCustomVAs01(rounding, abs(softEdge))
	r.opts.Uniforms["Quad"] = [8]float32{
		quad[0].X - tox, quad[0].Y - toy, quad[1].X - tox, quad[1].Y - toy,
		quad[2].X - tox, quad[2].Y - toy, quad[3].X - tox, quad[3].Y - toy,
	}
	target.DrawTrianglesShader32(r.vertices[:], r.indices[:], shader, &r.opts)
	clear(r.opts.Uniforms)
}

// precondition: quad must be self-intersecting
func (r *Renderer) fillSelfIntersectingQuad(target *ebiten.Image, quad [4]PointF32, rounding float32) {
	inter, ok := findSelfIntersection(quad)
	if !ok { // collinear or almost collinear
		if rounding > 0 {
			r.internalFillQuad(target, quad, shaderQuad.Load(), rounding, 0)
		}
		return
	}

	triA := [3]PointF32{inter.A[0], inter.A[1], inter.Intersection}
	triB := [3]PointF32{inter.B[0], inter.B[1], inter.Intersection}
	var roundingA, roundingB float32

	if rounding < 0 {
		// treat as two triangles, but avoid FillTriangle directly to stay on
		// the same shader and improve multi-vertex color interpolation
		var shape shapeType
		triA, shape, roundingA = preprocessTriangle([3]PointF32{inter.A[0], inter.A[1], inter.Intersection}, rounding)
		if shape == shapePoint {
			triA[1], triA[2] = triA[0], triA[0]
		}
		triB, _, roundingB = preprocessTriangle([3]PointF32{inter.B[0], inter.B[1], inter.Intersection}, rounding)
		if shape == shapePoint {
			triB[1], triB[2] = triB[0], triB[0]
		}

		if roundingA <= 0 {
			if roundingB <= 0 {
				return // nothing visible
			}
		}
		roundingA = max(roundingA, 0)
		roundingB = max(roundingB, 0)
	} else {
		// ensure A and B form clockwise triangles with CW ordering
		if triangleSignedArea(triA[0], triA[1], triA[2]) < 0 {
			triA[0], triA[1] = triA[1], triA[0]
		}
		if triangleSignedArea(triB[0], triB[1], triB[2]) < 0 {
			triB[0], triB[1] = triB[1], triB[0]
		}
		roundingA, roundingB = rounding, rounding
	}

	minAX, minBX := min(triA[0].X, triA[1].X, triA[2].X)-roundingA, min(triB[0].X, triB[1].X, triB[2].X)-roundingB
	minAY, minBY := min(triA[0].Y, triA[1].Y, triA[2].Y)-roundingA, min(triB[0].Y, triB[1].Y, triB[2].Y)-roundingB
	maxAX, maxBX := max(triA[0].X, triA[1].X, triA[2].X)+roundingA, max(triB[0].X, triB[1].X, triB[2].X)+roundingB
	maxAY, maxBY := max(triA[0].Y, triA[1].Y, triA[2].Y)+roundingA, max(triB[0].Y, triB[1].Y, triB[2].Y)+roundingB
	minX, maxX := floorF32(min(minAX, minBX)), ceilF32(max(maxAX, maxBX))
	minY, maxY := floorF32(min(minAY, minBY)), ceilF32(max(maxAY, maxBY))
	r.setDstRectCoords(minX, minY, maxX, maxY)
	r.setFlatCustomVAs01(roundingA, roundingB)
	tox, toy := rectOriginF32(target.Bounds())
	r.opts.Uniforms["TriangleA"] = [6]float32{triA[0].X - tox, triA[0].Y - toy, triA[1].X - tox, triA[1].Y - toy, triA[2].X - tox, triA[2].Y - toy}
	r.opts.Uniforms["TriangleB"] = [6]float32{triB[0].X - tox, triB[0].Y - toy, triB[1].X - tox, triB[1].Y - toy, triB[2].X - tox, triB[2].Y - toy}
	target.DrawTrianglesShader32(r.vertices[:], r.indices[:], shaderQuadSelfIntersect.Load(), &r.opts)
	clear(r.opts.Uniforms)
}

// FillQuadSoft draws a quad like [Renderer.FillQuad]() but with a soft edge.
//
// Rounding can be zero, positive for outwards rounding, or negative for
// inwards rounding. When positive, the soft radius will extend [-softEdge,
// +softEdge] around the boundary, approximating a gaussian blur. When
// negative, the softening will extend inwards [-softEdge, 0].
//
// Limitations: self-intersecting quads are not supported, and geometric
// fidelity is lower than the FillRectSoft approximation, as general quads
// have far more complex geometric shapes.
func (r *Renderer) FillQuadSoft(target *ebiten.Image, quad [4]PointF32, rounding, softEdge float32) {
	var simple bool
	quad, simple = canonicalizeQuadCW(quad)
	if !simple {
		r.Warnings.report(WarnSelfIntersectingQuad, quad)
	}

	var shader *ebiten.Shader
	if softEdge > 0 {
		// note: the empirical rounding correction applied to soft rects can't
		// be applied here (rounding -= softEdge / 1.65), as concave shapes
		// would be eroded way too fast (this is why the docs mention "lower
		// geometric fidelity"; the blur rounding is not approximated here)
		shader = shaderQuadSoftBlur.Load()
	} else {
		shader = shaderQuadSoftIn.Load()
	}

	if rounding < 0 {
		var nonEmpty bool
		quad, rounding, nonEmpty = innerRoundQuad(quad, rounding)
		if !nonEmpty {
			return
		}
	}

	r.internalFillQuad(target, quad, shader, rounding, softEdge)
}
