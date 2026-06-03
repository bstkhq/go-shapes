package shapes

import (
	"cmp"
	"math"
	"slices"
	"strconv"
)

// this file contains functions for quad and triangle expansion and shrinking.
// shrinking requires careful handling of collapse cases (and mitering for
// hulls).

type shape int

const (
	shapePoint shape = iota + 1
	shapeLine
	shapeTriangle
	shapeQuad
)

func (s shape) NumPoints() int {
	return int(s)
}

func (s shape) String() string {
	switch s {
	case shapePoint:
		return "Point"
	case shapeLine:
		return "Line"
	case shapeTriangle:
		return "Triangle"
	case shapeQuad:
		return "Quad"
	default:
		return "shape#" + strconv.Itoa(int(s))
	}
}

// offsetQuad applies a signed offset to the given quad. the returned shapes
// can be shapeQuad, shapeTriangle, shapeLine and shapePoint. the float is the
// offset reached (signed, used to detect the maximum collapse into a point)
//
// quad points must be given in clockwise order, +y axis goes down
func offsetQuad(quad [4]PointF32, offset float32) ([4]PointF32, shape, float32) {
	if offset > 0 {
		return offsetQuadNaive(quad, offset), shapeQuad, offset
	}
	out, shape, offsetReached := shrinkQuad(quad, -offset)
	return out, shape, -offsetReached
}

// offsetQuadNaive applies a naive offsetting of the edges of the quad. in the
// case of outsetting or outwards expansion, this is always safe to run
// directly. in the case of insetting, collapse into triangles, lines or points
// may happen, so an straight skeleton approach must be used first to detect
// such cases, and offset must only be applied when found to be safe. see also
// [shrinkQuad].
//
// quad points must be given in clockwise order, +y axis goes down
func offsetQuadNaive(quad [4]PointF32, offset float32) [4]PointF32 {
	if offset == 0 {
		return quad
	}

	return offsetQuadNaiveWithEdges(quad, quadNormalizedEdges(quad), offset)
}

// precondition: edges = quadNormalizedEdges(quad)
func offsetQuadNaiveWithEdges(quad [4]PointF32, edges [4]PointF32, offset float32) [4]PointF32 {
	if offset == 0 {
		return quad
	}

	return [4]PointF32{
		bisectorSlide(quad[0], edges[3], edges[0], offset),
		bisectorSlide(quad[1], edges[0], edges[1], offset),
		bisectorSlide(quad[2], edges[1], edges[2], offset),
		bisectorSlide(quad[3], edges[2], edges[3], offset),
	}
}

// quadHull computes a hull for the given quad, applying mitering if/when
// required to prevent tight angles from leading to large area fragments.
func quadHull(buff []PointF32, quad [4]PointF32) []PointF32 {
	const Offset = 1.0
	const MiterRadius = 5.0

	// ... TODO

	return buff
}

func quadNormalizedEdges(quad [4]PointF32) [4]PointF32 {
	e01 := quad[1].Sub(quad[0]).Normalize()
	e12 := quad[2].Sub(quad[1]).Normalize()
	e23 := quad[3].Sub(quad[2]).Normalize()
	e30 := quad[0].Sub(quad[3]).Normalize()
	return [4]PointF32{e01, e12, e23, e30}
}

// returns the intersection of p1 + t·d1 and p2 + u·d2 (two lines in
// parametric form: point + direction)
func lineIntersect(p1, d1, p2, d2 PointF32) PointF32 {
	// p1 + t*d1 = p2 + s*d2 => solve for t
	det := d1.X*d2.Y - d1.Y*d2.X
	if math.Abs(float64(det)) < 1e-6 {
		return p1.Add(p2).Scale(0.5) // ~parallel lines, return midpoint
	}

	t := ((p2.X-p1.X)*d2.Y - (p2.Y-p1.Y)*d2.X) / det
	return PointF32{
		X: p1.X + d1.X*t,
		Y: p1.Y + d1.Y*t,
	}
}

// shrinkQuad returns the quad with the edges offset inwards by offset, the
// resulting shape and the offset reached (magnitude)
//
// offset is a magnitude, so it must be >= 0. quad points must be given
// in clockwise order, +y axis goes down
func shrinkQuad(quad [4]PointF32, offset float32) ([4]PointF32, shape, float32) {
	if offset < 0 {
		panic("offset must be non-negative")
	}
	if offset == 0 {
		return quad, shapeQuad, offset
	}

	// compute first straight skeleton intersection
	edges := quadNormalizedEdges(quad)
	skeleton := firstQuadSkeletonOffset(quad, edges, offset)

	// simple case (no collapse)
	if offset < skeleton.Offset {
		return offsetQuadNaiveWithEdges(quad, edges, -offset), shapeQuad, offset
	}

	// hard cases (collapse)
	switch skeleton.Shape {
	case shapeLine:
		p1, p2, shape, offsetReached := shrinkLine(skeleton.Points[0], skeleton.Points[1], offset-skeleton.Offset)
		return [4]PointF32{p1, p2}, shape, offsetReached + skeleton.Offset
	case shapeTriangle:
		pi1, pi2, pi3 := skeleton.Points[0], skeleton.Points[1], skeleton.Points[2]
		po1, po2, po3, shape, offsetReached := shrinkTriangle(pi1, pi2, pi3, offset-skeleton.Offset)
		return [4]PointF32{po1, po2, po3}, shape, offsetReached + skeleton.Offset
	default:
		panic(skeleton.Shape)
	}
}

// offset is a magnitude (>= 0). the returned float32 is the offset reached
// (magnitude)
func shrinkLine(p1, p2 PointF32, offset float32) (PointF32, PointF32, shape, float32) {
	if offset < 0 {
		panic("offset must be >= 0")
	}

	segmentVector := p2.Sub(p1)
	segmentLength := segmentVector.Length()
	maxOffset := segmentLength * 0.5
	if offset > maxOffset-1e-6 {
		return p1.Add(segmentVector.Scale(0.5)), PointF32{}, shapePoint, maxOffset
	}

	dir := segmentVector.Scale(1.0 / segmentLength)
	shift := dir.Scale(offset)
	return p1.Add(shift), p2.Sub(shift), shapeLine, offset
}

// offset is a magnitude (>= 0). the returned float32 is the offset reached
// (magnitude)
//
// triangle points must be given in clockwise order, +y axis goes down
func shrinkTriangle(p1, p2, p3 PointF32, offset float32) (PointF32, PointF32, PointF32, shape, float32) {
	a12, b12, c12 := toLinearFormABC(p1.X, p1.Y, p2.X, p2.Y)
	a23, b23, c23 := toLinearFormABC(p2.X, p2.Y, p3.X, p3.Y)
	a31, b31, c31 := toLinearFormABC(p3.X, p3.Y, p1.X, p1.Y)
	c1_12, c2_12, l12 := parallelsAtDist(a12, b12, c12, offset)
	c1_23, c2_23, l23 := parallelsAtDist(a23, b23, c23, offset)
	c1_31, c2_31, l31 := parallelsAtDist(a31, b31, c31, offset)

	// handle collapse into point
	perimeter := l12 + l23 + l31
	if perimeter < 1e-6 {
		return p1, PointF32{}, PointF32{}, shapePoint, offset
	}
	area := 0.5 * abs((p1.X*(p2.Y-p3.Y) + p2.X*(p3.Y-p1.Y) + p3.X*(p1.Y-p2.Y)))
	maxOffset := area / (0.5 * perimeter)
	if offset >= maxOffset {
		cx := (l23*p1.X + l31*p2.X + l12*p3.X) / perimeter
		cy := (l23*p1.Y + l31*p2.Y + l12*p3.Y) / perimeter
		return PtF32(cx, cy), PointF32{}, PointF32{}, shapePoint, maxOffset
	}

	// no collapse, finish calculations
	if a12*p3.X+b12*p3.Y+c12 > 0 { // fancy winding order test
		c12, c23, c31 = c1_12, c1_23, c1_31
	} else {
		c12, c23, c31 = c2_12, c2_23, c2_31
	}
	p1.X, p1.Y = shortCramer(a31, b31, c31, a12, b12, c12)
	p2.X, p2.Y = shortCramer(a12, b12, c12, a23, b23, c23)
	p3.X, p3.Y = shortCramer(a23, b23, c23, a31, b31, c31)
	return p1, p2, p3, shapeTriangle, offset
}

type skeletonOffset struct {
	Offset float32     // perpendicular offset until first intersection
	Shape  shape       // line or triangle
	Points [3]PointF32 // 2 points for line case, 3 for triangle
}

// firstQuadSkeletonOffset computes the distance for the first straight
// skeleton edge collapse. check wikipedia or whatever for more details on
// straight skeleton algorithms
//
// precondition: edges = quadNormalizedEdges(quad)
func firstQuadSkeletonOffset(quad [4]PointF32, edges [4]PointF32, maxOffset float32) skeletonOffset {
	// bisectors for each vertex (unnormalized)
	var bisectors [4]PointF32
	bisectors[0] = edges[0].Sub(edges[3])
	bisectors[1] = edges[1].Sub(edges[0])
	bisectors[2] = edges[2].Sub(edges[1])
	bisectors[3] = edges[3].Sub(edges[2])

	// collapse vertices. notice that collapseKJ is the distance *along
	// the bisector*, while offsetJK is the distance *along the normals*
	vertsCollapse := func(vertexA, bisectorA, vertexB, bisectorB, edgeAB PointF32) (float32, PointF32) {
		collapseAB, okAB := intersectRays(vertexA, bisectorA, vertexB, bisectorB)
		if okAB && collapseAB > 0 {
			intersection := vertexA.Add(bisectorA.Scale(collapseAB))
			return distanceToNormalizedLine(intersection, vertexA, edgeAB), intersection
		}
		return float32(math.MaxFloat32), PointF32{}
	}

	offset01, inter01 := vertsCollapse(quad[0], bisectors[0], quad[1], bisectors[1], edges[0])
	offset12, inter12 := vertsCollapse(quad[1], bisectors[1], quad[2], bisectors[2], edges[1])
	offset23, inter23 := vertsCollapse(quad[2], bisectors[2], quad[3], bisectors[3], edges[2])
	offset30, inter30 := vertsCollapse(quad[3], bisectors[3], quad[0], bisectors[0], edges[3])

	// handle degenerate point/line cases
	out := skeletonOffset{Shape: shapeLine}
	if offset01 == math.MaxFloat32 && offset12 == math.MaxFloat32 && offset23 == math.MaxFloat32 && offset30 == math.MaxFloat32 {
		dir := cmp.Or(edges[0], edges[1], edges[2], edges[3])
		if dir == (PointF32{}) {
			out.Shape = shapePoint
			out.Points[0] = quad[0]
			return out
		}

		ext1 := (quad[1].X-quad[0].X)*dir.X + (quad[1].Y-quad[0].Y)*dir.Y
		ext2 := (quad[2].X-quad[0].X)*dir.X + (quad[2].Y-quad[0].Y)*dir.Y
		ext3 := (quad[3].X-quad[0].X)*dir.X + (quad[3].Y-quad[0].Y)*dir.Y
		minExt, maxExt := min(0, ext1, ext2, ext3), max(0, ext1, ext2, ext3)
		out.Points[0] = PointF32{X: quad[0].X + dir.X*minExt, Y: quad[0].Y + dir.Y*minExt}
		out.Points[1] = PointF32{X: quad[0].X + dir.X*maxExt, Y: quad[0].Y + dir.Y*maxExt}
		return out
	}

	// line collapse (technically it might also be a point, but line covers it as well)
	if offset01 < math.MaxFloat32 && abs(offset01-offset23) < 1e-4 {
		out.Offset = offset01
		out.Points[0], out.Points[1] = inter01, inter23
		return out
	}
	if offset12 < math.MaxFloat32 && abs(offset12-offset30) < 1e-4 {
		out.Offset = offset12
		out.Points[0], out.Points[1] = inter30, inter12
		return out
	}

	// triangle collapse
	out.Shape = shapeTriangle
	out.Offset = min(offset01, offset12, offset23, offset30)
	if out.Offset > maxOffset {
		return out // performance cutoff
	}
	switch out.Offset {
	case offset01:
		out.Points[0] = inter01
		out.Points[1] = bisectorSlide(quad[2], edges[1], edges[2], -out.Offset)
		out.Points[2] = bisectorSlide(quad[3], edges[2], edges[3], -out.Offset)
	case offset12:
		out.Points[0] = bisectorSlide(quad[0], edges[3], edges[0], -out.Offset)
		out.Points[1] = inter12
		out.Points[2] = bisectorSlide(quad[3], edges[2], edges[3], -out.Offset)
	case offset23:
		out.Points[0] = bisectorSlide(quad[0], edges[3], edges[0], -out.Offset)
		out.Points[1] = bisectorSlide(quad[1], edges[0], edges[1], -out.Offset)
		out.Points[2] = inter23
	case offset30:
		out.Points[0] = inter30
		out.Points[1] = bisectorSlide(quad[1], edges[0], edges[1], -out.Offset)
		out.Points[2] = bisectorSlide(quad[2], edges[1], edges[2], -out.Offset)
	default:
		panic("NaN?") // this shouldn't be possible by construction
	}

	return out
}

// returns the distance t1 along ray1 where the rays intersect. parallel rays
// return false
func intersectRays(p1, d1, p2, d2 PointF32) (float32, bool) {
	denom := d1.X*d2.Y - d1.Y*d2.X
	if abs(denom) < 1e-6 { // Parallel lines
		return 0, false
	}
	num := (p2.X-p1.X)*d2.Y - (p2.Y-p1.Y)*d2.X
	return num / denom, true
}

func distanceToNormalizedLine(p, lineOrigin, lineDir PointF32) float32 {
	v := p.Sub(lineOrigin)
	return abs(v.X*lineDir.Y - v.Y*lineDir.X) // cross product
}

// bisectorSlide slides a corner vertex inward by a perpendicular offset
// distance from the edges. the edge vectors must be normalized.
func bisectorSlide(v, edgeIn, edgeOut PointF32, offset float32) PointF32 {
	bisector := edgeOut.Sub(edgeIn)

	// cross product magnitude (perpendicular projection factor)
	denom := bisector.X*edgeOut.Y - bisector.Y*edgeOut.X
	if abs(denom) < 1e-6 {
		return v // parallel
	}

	// scale factor along the bisector
	return v.Add(bisector.Scale(offset / denom))
}

// see shoelace formula. the function returns a positive value for CW quads,
// negative for CCW and 0 for degenerate cases (self-intersecting, all points
// on a line)
func shoelaceArea(quad [4]PointF32) float32 {
	return (quad[1].X-quad[0].X)*(quad[1].Y+quad[0].Y) +
		(quad[2].X-quad[1].X)*(quad[2].Y+quad[1].Y) +
		(quad[3].X-quad[2].X)*(quad[3].Y+quad[2].Y) +
		(quad[0].X-quad[3].X)*(quad[0].Y+quad[3].Y)
}

// normalizeTriangleCW checks the winding of the triangle and makes it CW
func normalizeTriangleCW(points [3]PointF32) [3]PointF32 {
	cross := (points[1].X-points[0].X)*(points[2].Y-points[0].Y) - (points[1].Y-points[0].Y)*(points[2].X-points[0].X)
	if cross < 0 {
		points[1], points[2] = points[2], points[1]
	}
	return points
}

// normalizeQuadCW checks the winding of the quad and makes it CW
func normalizeQuadCW(points [4]PointF32) [4]PointF32 {
	if shoelaceArea(points) < 0 {
		points[1], points[3] = points[3], points[1]
	}
	return points
}

// canonicalizeQuadCW converts the quad to its canonical form: CW orientation,
// if concave then quad[2] contains the reflex angle. The method returns false
// if the quad can't be canonicalized (self-intersecting)
func canonicalizeQuadCW(points [4]PointF32) ([4]PointF32, bool) {
	turn := func(a, b, c PointF32) float32 {
		ab, bc := a.Sub(b), b.Sub(c)
		return ab.X*bc.Y - ab.Y*bc.X
	}
	turns := [4]float32{
		turn(points[3], points[0], points[1]),
		turn(points[0], points[1], points[2]),
		turn(points[1], points[2], points[3]),
		turn(points[2], points[3], points[0]),
	}

	var cw, ccw int // assume x+ right, y+ down
	for _, t := range turns {
		if t > 0 {
			cw += 1
		} else if t < 0 {
			ccw += 1
		}
	}

	switch {
	case cw > 1 && ccw > 1:
		return points, false // self-intersecting
	case ccw == 1 && cw > 1: // CW concave
		reflex := slices.IndexFunc(turns[:], func(t float32) bool { return t < 0 })
		points[0], points[1], points[2], points[3] = points[(reflex+2)&3], points[(reflex+3)&3], points[reflex], points[(reflex+1)&3]
	case cw == 1 && ccw > 1: // CCW concave
		reflex := slices.IndexFunc(turns[:], func(t float32) bool { return t > 0 })
		points[0], points[1], points[2], points[3] = points[(reflex+2)&3], points[(reflex+1)&3], points[reflex], points[(reflex+3)&3]
	default: // convex or line or point
		if ccw > cw {
			points[1], points[3] = points[3], points[1]
		}
	}
	return points, true
}
