package shapes

import (
	"fmt"
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

// FillRectSoft draws a rect like [Renderer.FillRect]() but with a soft edge. This is
// ideal for rect shadows in UIs and avoiding the more expensive raster-based blurs.
//
// Rounding can be zero, positive for outwards rounding, or negative for inwards
// rounding. When positive, the soft radius will extend [-softEdge, +softEdge] around the
// boundary, closely approximating a gaussian blur. When negative, the softening will
// extend inwards [-softEdge, 0].
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

	var shader *ebiten.Shader
	if softEdge > 0 {
		rounding -= softEdge / 1.65 // empirical adjustment
		shader = shaderRectSoftBlur.Load()
	} else {
		softEdge = -softEdge
		shader = shaderRectSoftIn.Load()
		if ebiten.IsKeyPressed(ebiten.KeyQ) {
			shader = shaderRectSoftInSmoothstep.Load()
		}
	}

	tox, toy := rectOriginF32(target.Bounds())
	r.setFlatCustomVAs(ox-tox, oy-toy, w, h)
	r.opts.Uniforms["InRounding"] = -rounding
	r.opts.Uniforms["BlurRadius"] = softEdge
	margin := max(softEdge, 0)
	r.DrawRectShader(target, ox, oy, w, h, NewMargins(margin, margin), shader)
	clear(r.opts.Uniforms)
}

// StrokeLine draws a smooth line between the given two points, with rounded ends.
func (r *Renderer) StrokeLine(target *ebiten.Image, ox, oy, fx, fy float64, thickness float64) {
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

	tox, toy := rectOriginF32(target.Bounds())
	r.setFlatCustomVAs(float32(ox)-tox, float32(oy)-toy, float32(fx)-tox, float32(fy)-toy)
	r.opts.Uniforms["Thickness"] = float32(thickness)

	// draw shader
	target.DrawTrianglesShader32(r.vertices[:], r.indices[:], shaderLine.Load(), &r.opts)
}

func (r *Renderer) internalStrokeIntRect(target *ebiten.Image, ox, oy, w, h, outThickness, inThickness int) {
	if w < 0 {
		w = -w
		ox -= w
	}
	if h < 0 {
		h = -h
		oy -= h
	}

	outThickness = warnZeroNegativeValue(r, outThickness)
	inThickness = warnZeroNegativeValue(r, inThickness)
	if outThickness+inThickness == 0 {
		return
	}

	if outThickness == 0 {
		if inThickness != 0 {
			r.strokeIntInnerRect(target, ox, oy, w, h, inThickness)
		}
	} else {
		r.strokeIntInnerRect(target, ox-outThickness, oy-outThickness, w+outThickness*2, h+outThickness*2, outThickness+inThickness)
	}
}

var strokeIndices = []uint32{
	0, 1, 4,
	4, 1, 5,
	5, 1, 2,
	5, 2, 6,
	6, 2, 3,
	6, 3, 7,
	7, 3, 0,
	0, 4, 7,
}

func (r *Renderer) strokeIntInnerRect(target *ebiten.Image, ox, oy, w, h, thickness int) {
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
		// we need to interpolate colors. this code takes advantage of
		// the heavy symmetries in the geometry to reduce the number of
		// operations, but as a downside, it's a bit tricky to understand

		// compute uv coords for inner points
		iou := min(max((iox-oox)/(ofx-oox), 0), 1)
		iov := min(max((ioy-ooy)/(ofy-ooy), 0), 1)

		// compute top and bottom left colors
		tR, tG, tB, tA := interpVertexColor(r.vertices[0], r.vertices[1], iou)
		bR, bG, bB, bA := interpVertexColor(r.vertices[3], r.vertices[2], iou)

		// compute left side colors
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

	target.DrawTrianglesShader32(r.vertices[:], strokeIndices[:], shaderDefault.Load(), &r.opts)
	r.vertices = r.vertices[:4]
}

// StrokeRect is the [image.Rectangle]-compatible equivalent of [Renderer.StrokeRect]().
//
// For rectangle creation, consider [image.Rect]() and [RectWithSize]().
func (r *Renderer) StrokeIntRect(target *ebiten.Image, rect image.Rectangle, inThickness, outThickness int) {
	r.internalStrokeIntRect(target, rect.Min.X, rect.Min.Y, rect.Dx(), rect.Dy(), outThickness, inThickness)
}

// StrokeRect draws an outline on the given area's boundary, with explicit controls for in/out
// border thickness. Rounding controls the outer edge rounding radius (inner radius if negative).
//
// If you have an [image.Rectangle] for the rect, consider [Renderer.StrokeIntRect]() instead,
// or if you need rounding convert with [RectPointsF32]().
func (r *Renderer) StrokeRect(target *ebiten.Image, ox, oy, w, h, inThickness, outThickness, rounding float32) {
	// NOTE: should we consider optional segmentation more similar to what internalStrokeIntRect does?
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
			r.strokeInnerRect(target, ox, oy, w, h, inThickness, rounding)
		}
	} else {
		r.strokeInnerRect(target, ox-outThickness, oy-outThickness, w+outThickness*2, h+outThickness*2, outThickness+inThickness, rounding)
	}
}

func (r *Renderer) strokeInnerRect(target *ebiten.Image, ox, oy, w, h, inThickness, rounding float32) {
	if rounding < 0 {
		// adjust for inner boundary
		rounding = -(rounding - inThickness)
	}

	const SafeMargin = 2.0
	if rounding >= min(w, h)*2.0+SafeMargin {
		return // ignore
	}

	tox, toy := rectOriginF32(target.Bounds())
	r.setFlatCustomVAs(ox-tox, oy-toy, w, h)
	r.opts.Uniforms["InnerThickness"] = inThickness
	r.opts.Uniforms["Rounding"] = rounding
	r.DrawRectShader(target, ox, oy, w, h, NoMargins, shaderStrokeRect.Load())
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

// TODO: support Hull flag
func (r *Renderer) drawTriangle(target *ebiten.Image, points [3]PointF32, thickness, rounding float32) {
	points, shape, rounding := preprocessTriangle(points, rounding)
	if shape == shapePoint {
		if thickness == 0 {
			r.FillCircle(target, points[0].X, points[0].Y, max(rounding, 0))
		} else {
			r.StrokeCircle(target, points[0].X, points[0].Y, rounding, thickness)
		}
		return
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
	r.setFlatCustomVAs01(abs(rounding), thickness)
	target.DrawTrianglesShader32(r.vertices[:], r.indices[:], shaderTriangle.Load(), &r.opts)
	clear(r.opts.Uniforms)
}

// preprocessTriangle handles inner rounding, shrinking the geometry and
// converting to outer rounding while also handling collapse cases. notice
// that outer rounding can still be negative, as thickness might have to
// be applied on top
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
// triggers additional precomputations, which are particularly complex for
// inner rounding.
//
// Limitations: self-intersecting quads with inner rounding are drawn as two
// triangles, so multi-vertex color is only approximated.
func (r *Renderer) FillQuad(target *ebiten.Image, quad [4]PointF32, rounding float32) {
	var simple bool
	quad, simple = canonicalizeQuadCW(quad)
	if !simple {
		r.fillSelfIntersectingQuad(target, quad, rounding)
		return
	}

	if rounding < 0 {
		dbgX, dbgY := min(quad[0].X, quad[1].X, quad[2].X, quad[3].X), min(quad[0].Y, quad[1].Y, quad[2].Y, quad[3].Y)
		quad, shape, offsetReached := offsetQuad(quad, rounding)
		r.Text(target, fmt.Sprintf("shape: %s, offsetReached: %.02f", shape.String(), offsetReached), dbgX, dbgY, TextOpts(1.0, BottomLeft))

		switch shape {
		case shapePoint:
			radius := rounding - offsetReached*2
			if radius <= 0 {
				return // empty
			}
			r.FillCircle(target, quad[0].X, quad[0].Y, radius)
		case shapeLine:
			radius := rounding - offsetReached*2
			if radius <= 0 {
				return // empty
			}
			r.StrokeLine(target, float64(quad[0].X), float64(quad[0].Y), float64(quad[1].X), float64(quad[1].Y), float64(radius*2.0))
		case shapeTriangle:
			var tri [3]PointF32
			tri[0], tri[1], tri[2] = quad[0], quad[1], quad[2]
			r.FillTriangle(target, tri, -rounding)
		case shapeQuad:
			r.FillQuad(target, quad, -rounding)
		default:
			panic(shape) // broken code
		}
		return
	}

	// if opts.include(Hull) {
	//     r.hullBuff = quadHull(r.hullBuff, quad, rounding)
	// }
	r.internalFillQuad(target, quad, rounding)
}

// precondition: rounding >= 0
func (r *Renderer) internalFillQuad(target *ebiten.Image, quad [4]PointF32, rounding float32) {
	minX, maxX := floorF32(min(quad[0].X, quad[1].X, quad[2].X, quad[3].X)-rounding), ceilF32(max(quad[0].X, quad[1].X, quad[2].X, quad[3].X)+rounding)
	minY, maxY := floorF32(min(quad[0].Y, quad[1].Y, quad[2].Y, quad[3].Y)-rounding), ceilF32(max(quad[0].Y, quad[1].Y, quad[2].Y, quad[3].Y)+rounding)
	r.setDstRectCoords(minX, minY, maxX, maxY)

	tox, toy := rectOriginF32(target.Bounds())
	r.setFlatCustomVA0(rounding)
	r.opts.Uniforms["Quad"] = [8]float32{
		quad[0].X - tox, quad[0].Y - toy, quad[1].X - tox, quad[1].Y - toy,
		quad[2].X - tox, quad[2].Y - toy, quad[3].X - tox, quad[3].Y - toy,
	}
	target.DrawTrianglesShader32(r.vertices[:], r.indices[:], shaderQuad.Load(), &r.opts)
	clear(r.opts.Uniforms)
}

// precondition: quad must be self-intersecting
func (r *Renderer) fillSelfIntersectingQuad(target *ebiten.Image, quad [4]PointF32, rounding float32) {
	// TODO
}

// FillQuadSoft draws a quad like [Renderer.FillQuad]() but with an extra softEdge, which
// creates a shadow-like soft edge.
//
// TODO: thickening -> rounding, soft edge both positive and negative (inwards)
func (r *Renderer) FillQuadSoft(target *ebiten.Image, quad [4]PointF32, offset, softEdge float32) {
	quad, shape, offsetReached := offsetQuad(quad, offset)
	_ = offsetReached
	switch shape {
	case shapeLine:
		panic("unimplemented soft shapeLine")
	case shapePoint:
		panic("unimplemented soft shapePoint")
	case shapeTriangle:
		panic("unimplemented soft shapeTriangle")
	case shapeQuad:
		for i, pt := range quad {
			r.vertices[i].DstX = pt.X
			r.vertices[i].DstY = pt.Y
		}

		r.setFlatCustomVAs01(offset, softEdge)
		tox, toy := rectOriginF32(target.Bounds())
		r.opts.Uniforms["Quad"] = [8]float32{
			quad[0].X - tox, quad[0].Y - toy, quad[1].X - tox, quad[1].Y - toy,
			quad[2].X - tox, quad[2].Y - toy, quad[3].X - tox, quad[3].Y - toy,
		}
		target.DrawTrianglesShader32(r.vertices[:], r.indices[:], shaderQuad.Load(), &r.opts)
		clear(r.opts.Uniforms)
	}
}
