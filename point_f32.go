package shapes

import "math"

// PointF32 is a helper type for operations with triangles and quads like
// [Renderer.DrawTriangle](), [Renderer.DrawQuad](), [Renderer.MapQuad4]()...
type PointF32 struct {
	X, Y float32
}

func PtF32(x, y float32) PointF32 {
	return PointF32{X: x, Y: y}
}

func (p PointF32) Sub(o PointF32) PointF32 {
	return PointF32{p.X - o.X, p.Y - o.Y}
}

func (p PointF32) Add(o PointF32) PointF32 {
	return PointF32{p.X + o.X, p.Y + o.Y}
}

func (p PointF32) AddXY(x, y float32) PointF32 {
	return PointF32{p.X + x, p.Y + y}
}

func (p PointF32) Scale(s float32) PointF32 {
	return PointF32{p.X * s, p.Y * s}
}

func (p PointF32) Dot(o PointF32) float32 {
	return p.X*o.X + p.Y*o.Y
}

func (p PointF32) Length() float32 {
	return float32(math.Hypot(float64(p.X), float64(p.Y)))
}

func (p PointF32) lengthSq() float32 {
	return p.X*p.X + p.Y*p.Y
}

func (p PointF32) Normalize() PointF32 {
	l := p.Length()
	if l == 0 {
		return PointF32{0, 0}
	}
	return PointF32{p.X / l, p.Y / l}
}
