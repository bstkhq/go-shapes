package shapes

import (
	"cmp"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

var roFloat32Inf = float32(math.Inf(1))

func Float32Inf() float32 {
	return roFloat32Inf
}

type GoldenRatioGen struct {
	n float64
}

func (gen *GoldenRatioGen) Reset() {
	gen.n = 0
}

func (gen *GoldenRatioGen) Float64() float64 {
	const phi = 1.618033988749895 // golden ratio
	gen.n += 1.0
	if gen.n == 514230.0 {
		gen.n = 1.0
	}
	v := math.Mod(gen.n/phi, 1.0)
	return v
}

// RadsSpan returns the start and end angles (in radians) centered
// around centerDir. fillRate must be in [0...1].
//
// This is a helper function often used with [Renderer.DrawCircSector]()
// and similar functions.
//
// See [RadsRight] constants for angle conventions and docs.
func RadsSpan[Float ~float64 | ~float32](centerDir Float, fillRate Float) (start, end Float) {
	if fillRate <= 0 {
		return centerDir, centerDir
	}
	if fillRate >= 1.0 {
		start := uradsAddCW(centerDir, math.Pi)
		return start, start + 2*Float(math.Pi)
	}

	centerDir = normURads(centerDir)
	ratePi := fillRate * math.Pi
	return uradsAddCW(centerDir, -ratePi), uradsAddCW(centerDir, ratePi)
}

func ceilF32(x float32) float32 {
	return float32(math.Ceil(float64(x)))
}

func clamp[T cmp.Ordered](v, minValue, maxValue T) T {
	return min(max(v, minValue), maxValue)
}

func abs[Float float32 | float64](a Float) Float {
	if a < 0 {
		return -a
	}
	return a
}

func lerp[Float float32 | float64](a, b, t Float) Float {
	return a + t*(b-a)
}

// umod returns the non-negative remainder of x mod m, similar
// to rust's [rem_euclid]. This is often used in the package for
// normalizing angles.
//
// [rem_euclid]: https://doc.rust-lang.org/std/primitive.f64.html#method.rem_euclid
func umod(x, m float64) float64 {
	r := math.Mod(x, m)
	if r < 0 {
		r += m
	}
	return r
}

// normURads calls [umod](r, 2*math.Pi) to normalize r to [0, 2*pi) range.
func normURads[Float ~float64 | ~float32](r Float) Float {
	if r >= 0 && r <= 2*math.Pi {
		return r
	}
	return Float(umod(float64(r), 2*math.Pi))
}

// Notice: geometry code is derived from etxt@v0.0.8 emask/helper_funcs.go

// Given two points of a line, it returns its A, B and C
// coefficients from the form "Ax + By + C = 0".
func toLinearFormABC[Float ~float32 | ~float64](ox, oy, fx, fy Float) (Float, Float, Float) {
	a, b, c := fy-oy, -(fx - ox), (fx-ox)*oy-(fy-oy)*ox
	return a, b, c
}

// If we had two line equations like this:
// >> a1*x + b1*y = c1
// >> a2*x + b2*y = c2
// We would apply cramer's rule to solve the system:
// >> x = (b2*c1 - b1*c2)/(b2*a1 - b1*a2)
// This function solves this system, but assuming c1 and c2 have
// a negative sign (ax + by + c = 0).
func shortCramer[Float ~float32 | ~float64](a1, b1, c1, a2, b2, c2 Float) (Float, Float) {
	xdiv := b2*a1 - b1*a2
	if xdiv == 0 {
		panic("parallel lines")
	}

	// actual application of cramer's rule
	x := (b2*-c1 - b1*-c2) / xdiv
	if b1 != 0 {
		return x, (-c1 - a1*x) / b1
	}
	return x, (-c2 - a2*x) / b2
}

// given a line equation in the form Ax + By + C = 0, it returns
// C1 and C2 such that two new line equations can be created that
// are parallel to the original line, but at distance 'dist' from it.
// It also returns hypot(a, b), which is the length of the line and
// can be useful in some contexts.
func parallelsAtDist[Float ~float32 | ~float64](a, b, c, dist Float) (Float, Float, Float) {
	norm := Float(math.Hypot(float64(a), float64(b)))
	if norm == 0 {
		return c, c, norm // degenerate case
	}
	shift := dist * norm
	return c - shift, c + shift, norm
}

func distToLineF32(coords, start, end PointF32) float32 {
	if start.X == end.X {
		return abs(coords.X - end.X)
	}
	if start.Y == end.Y {
		return abs(coords.Y - end.Y)
	}

	// given ax + by + c = 0, and point (x1, y1):
	// >> d = |ax1 + by1 + c|/sqrt(a² + b²)
	a, b, c := toLinearFormABC(start.X, start.Y, end.X, end.Y)
	return abs(a*coords.X+b*coords.Y+c) / float32(math.Sqrt(float64(a*a+b*b)))
}

func distToArcF32(coords, start, end PointF32, radius float32) float32 {
	panic("unimplemented")
}

func distToQuadF32(coords, start, end PointF32, cx, cy float32) float32 {
	panic("unimplemented")
}

func distToCubeF32(coords, start, end PointF32, cax, cay, cbx, cby float32) float32 {
	panic("unimplemented")
}

func snapEdges[Float ~float32 | ~float64](value, min, max, tolerance Float) Float {
	switch {
	case value+tolerance > max:
		return max
	case value-tolerance < min:
		return min
	default:
		return value
	}
}

// gaussian elimination 8x8 homogeneous linear system solver
func gaussSolver8x8(sys [8][8]float32, weights [8]float32) [8]float32 {
	var x [8]float32
	for i := range 8 {
		// find pivot
		maxRow := i
		for k := i + 1; k < 8; k++ {
			if abs(sys[k][i]) > abs(sys[maxRow][i]) {
				maxRow = k
			}
		}

		// swap rows
		sys[i], sys[maxRow] = sys[maxRow], sys[i]
		weights[i], weights[maxRow] = weights[maxRow], weights[i]

		// eliminate
		for k := i + 1; k < 8; k++ {
			f := sys[k][i] / sys[i][i]
			for j := i; j < 8; j++ {
				sys[k][j] -= f * sys[i][j]
			}
			weights[k] -= f * weights[i]
		}
	}

	// substitution
	for i := 7; i >= 0; i-- {
		sum := float32(0)
		for j := i + 1; j < 8; j++ {
			sum += sys[i][j] * x[j]
		}
		x[i] = (weights[i] - sum) / sys[i][i]
	}

	return x
}

// points are given in clockwise order, from top-left.
// returned matrix is row-major order
func computeHomography(fromQuad, toQuad [4]PointF32) [9]float32 {
	var system [8][8]float32
	var weights [8]float32

	var i int
	for j, pt := range fromQuad {
		u, v := toQuad[j].X, toQuad[j].Y
		system[i+0] = [8]float32{pt.X, pt.Y, 1, 0, 0, 0, -u * pt.X, -u * pt.Y}
		system[i+1] = [8]float32{0, 0, 0, pt.X, pt.Y, 1, -v * pt.X, -v * pt.Y}
		weights[i+0] = u
		weights[i+1] = v
		i += 2
	}

	solutions := gaussSolver8x8(system, weights)
	var homography [9]float32
	_ = copy(homography[:], solutions[:])
	homography[8] = 1.0
	return homography
}

// quad points must be given in clockwise order, +y axis goes down
func expandQuad(quad [4]PointF32, thickness float32) [4]PointF32 {
	if thickness == 0 {
		return quad
	}

	// edges
	e0 := quad[1].Sub(quad[0])
	e1 := quad[2].Sub(quad[1])
	e2 := quad[3].Sub(quad[2])
	e3 := quad[0].Sub(quad[3])

	// normals
	n0 := PointF32{X: e0.Y, Y: -e0.X}.Normalize().Scale(thickness)
	n1 := PointF32{X: e1.Y, Y: -e1.X}.Normalize().Scale(thickness)
	n2 := PointF32{X: e2.Y, Y: -e2.X}.Normalize().Scale(thickness)
	n3 := PointF32{X: e3.Y, Y: -e3.X}.Normalize().Scale(thickness)

	// offset points
	p0a := quad[0].Add(n0)
	p1a := quad[1].Add(n1)
	p2a := quad[2].Add(n2)
	p3a := quad[3].Add(n3)

	// final intersection
	out := [4]PointF32{
		lineIntersect(p3a, e3, p0a, e0),
		lineIntersect(p0a, e0, p1a, e1),
		lineIntersect(p1a, e1, p2a, e2),
		lineIntersect(p2a, e2, p3a, e3),
	}
	return out
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

// precondition: angles must be normalized by normURads, outRadius >= inRadius
func circSectorBounds(cx, cy float32, inRadius, outRadius float32, startRads, endRads float64) (minX, minY, maxX, maxY float32) {
	ss, sc := math.Sincos(startRads)
	es, ec := math.Sincos(endRads)
	ss32, sc32, es32, ec32 := float32(ss), float32(sc), float32(es), float32(ec)
	pi1x, pi1y := cx+inRadius*sc32, cy+inRadius*ss32
	po1x, po1y := cx+outRadius*sc32, cy+outRadius*ss32
	pi2x, pi2y := cx+inRadius*ec32, cy+inRadius*es32
	po2x, po2y := cx+outRadius*ec32, cy+outRadius*es32
	minX, minY = min(pi1x, po1x, pi2x, po2x), min(pi1y, po1y, pi2y, po2y)
	maxX, maxY = max(pi1x, po1x, pi2x, po2x), max(pi1y, po1y, pi2y, po2y)

	if uradsWithinCW(RadsRight, startRads, endRads) {
		maxX = cx + outRadius
	}
	if uradsWithinCW(RadsBottom, startRads, endRads) {
		maxY = cy + outRadius
	}
	if uradsWithinCW(RadsLeft, startRads, endRads) {
		minX = cx - outRadius
	}
	if uradsWithinCW(RadsTop, startRads, endRads) {
		minY = cy - outRadius
	}
	return minX, minY, maxX, maxY
}

// uradsWithinCW returns whether 'rads' is within the clockwise segment [start, end],
// assumming that all angles are normalized in the [0, 2*pi) range (e.g. normURads)
func uradsWithinCW[Float ~float64 | ~float32](rads, start, end Float) bool {
	if start < end {
		return rads >= start && rads <= end
	}
	return rads >= start || rads <= end
}

func uradsDeltaCW[Float ~float64 | ~float32](start, end Float) Float {
	if end >= start {
		return end - start
	}
	return 2*math.Pi - start + end
}

// precondition: start is in [0, 2*pi) range, delta is in (-2*pi, 2*pi) range
func uradsAddCW[Float ~float64 | ~float32](start, delta Float) Float {
	total := start + delta
	if total > 2*math.Pi {
		total -= 2 * math.Pi
	} else if total < 0 {
		total += 2 * math.Pi
	}
	return total
}

// Vertices are appended in (in, out) pairs.
// precondition: tolerance >= 0.1
func appendArcVertices(vertices []ebiten.Vertex, radius, thickness, startRads, rads float64, tolerance float64) []ebiten.Vertex {
	if tolerance < 0.095 {
		panic("tolerance < 0.1")
	}

	var sidePad float64
	var fullCircle bool
	if rads >= 2*math.Pi-0.000001 {
		rads = 2 * math.Pi
		startRads = 0.0
		fullCircle = true
	} else if tolerance >= 1.0 {
		sidePad = 1.0
	}

	radiusIn := max(radius-thickness*0.5, 0)
	radiusOut := radius + thickness*0.5 + tolerance

	// - compute maximum chord length -
	// maximum chord length is determined by maxOvershoot, which must be <= sagitta
	// (s, the distance between the midpoints of an arc and its chord of length c).
	// this is given by the formula c = 2*sqrt(2*R*s - s²), which can be derived from
	// the triangle with hypotenuse = R, side = R - s, base = c/2. apply pythagoras
	// and you get (R - s)² + (c/2)² = R²
	maxChordLen := 2 * math.Sqrt(2*radiusOut*tolerance-tolerance*tolerance)

	// compute max angle from max chord length and turn into number of segments
	maxAngle := 2 * math.Asin(maxChordLen/(2*radiusOut)) // chord/angle formula c = 2*R*sin(θ/2)
	if maxAngle <= 0 {
		panic(maxAngle)
	}
	numSegments := math.Ceil(rads / maxAngle)
	// fmt.Printf("numSegments: %d\n", int(numSegments))
	step := rads / float64(numSegments)

	// tighten radiusOut using the final angle
	actualS := radiusOut * (1 - math.Cos(step*0.5))
	radiusOut = radius + thickness*0.5 + actualS

	// sidePad = 0.0
	iters := int(numSegments)
	if !fullCircle {
		iters += 1
	}
	for i := range iters {
		sin, cos := math.Sincos(uradsAddCW(startRads, step*float64(i)))
		vertices = append(vertices,
			ebiten.Vertex{
				DstX: float32(radiusIn * cos),
				DstY: float32(radiusIn * sin),
			},
			ebiten.Vertex{
				DstX: float32(radiusOut * cos),
				DstY: float32(radiusOut * sin),
			},
		)
		if sidePad > 0 {
			if i == 0 {
				shiftX, shiftY := float32(-sin), float32(cos)
				vertices[0].DstX += shiftX
				vertices[0].DstY += shiftY
				vertices[1].DstX += shiftX
				vertices[1].DstY += shiftY
			} else if i == iters-1 {
				shiftX, shiftY := float32(sin), float32(-cos)
				vertices[i].DstX += shiftX
				vertices[i].DstY += shiftY
				vertices[i-1].DstX += shiftX
				vertices[i-1].DstY += shiftY
			}
		}
	}
	if fullCircle { // perfect wrap
		vertices = append(vertices, vertices[0], vertices[1])
	}

	return vertices
}

// (cx, cy) is a point on circle, (ox, oy) is an outside point.
// Circle center is assumed to be (0,0). this function will return
// (cx, cy) unless the line from c to o crosses another point
// in the circle first.
func lineCircIntersect(radius, cx, cy, ox, oy float64) (float64, float64) {
	dx, dy := ox-cx, oy-cy
	dd := dx*dx + dy*dy
	if dd == 0 {
		return cx, cy
	}

	// direct second root (t=0 is not relevant)
	t := -2.0 * (cx*dx + cy*dy) / dd

	// forward check
	if t > 0 {
		return cx + t*dx, cy + t*dy
	}

	return cx, cy
}
