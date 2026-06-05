package shapes

import (
	"image"
	"math"
)

// PointF32 is a helper type for operations with triangles, quads and paths like
// [Renderer.FillTriangle](), [Renderer.FillQuad](), etc.
type PointF32 struct {
	X, Y float32
}

// PtF32 is shorthand for PointF32{X: x, Y: y}.
func PtF32(x, y float32) PointF32 {
	return PointF32{X: x, Y: y}
}

// RoundInt rounds a PointF32 into an image.Point.
func (p PointF32) RoundInt() image.Point {
	return image.Pt(int(p.X+0.5), int(p.Y+0.5))
}

// Sub returns p - o.
func (p PointF32) Sub(o PointF32) PointF32 {
	return PointF32{p.X - o.X, p.Y - o.Y}
}

// Add returns p + o.
func (p PointF32) Add(o PointF32) PointF32 {
	return PointF32{p.X + o.X, p.Y + o.Y}
}

// AddXY returns p + PtF32(x, y).
func (p PointF32) AddXY(x, y float32) PointF32 {
	return PointF32{p.X + x, p.Y + y}
}

// Mul returns the component-wise product of p*o.
func (p PointF32) Mul(o PointF32) PointF32 {
	return PointF32{X: p.X * o.X, Y: p.Y * o.Y}
}

// Scale returns p*s.
func (p PointF32) Scale(s float32) PointF32 {
	return PointF32{p.X * s, p.Y * s}
}

// Dot returns p.X*o.X + p.Y*o.Y.
func (p PointF32) Dot(o PointF32) float32 {
	return p.X*o.X + p.Y*o.Y
}

// signed area of the parallelogram created by directions p x o.
// or torque. or measure of perpendicularity. or...
func (p PointF32) cross(o PointF32) float32 {
	return p.X*o.Y - p.Y*o.X
}

// Length returns sqrt(p.X^2 + p.Y^2).
func (p PointF32) Length() float32 {
	return float32(math.Hypot(float64(p.X), float64(p.Y)))
}

func (p PointF32) lengthSq() float32 {
	return p.X*p.X + p.Y*p.Y
}

// Normalize returns the point scaled to have length = 1.
func (p PointF32) Normalize() PointF32 {
	l := p.Length()
	if l == 0 {
		return PointF32{0, 0}
	}
	return PointF32{p.X / l, p.Y / l}
}

// Rotate rotates the point by the given radians. For rotations with non-zero
// origins, use p.Sub(origin).Rotate(rads).Add(origin).
func (p PointF32) Rotate(rads float64) PointF32 {
	s, c := math.Sincos(rads)
	x, y := rotate(p.X, p.Y, float32(s), float32(c))
	return PointF32{X: x, Y: y}
}
