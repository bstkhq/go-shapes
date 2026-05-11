package shapes

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

func TestGaussSolver8x8(t *testing.T) {
	const tolerance float32 = 1e-6

	var tests = []struct {
		system    [8][8]float32
		weights   [8]float32
		solutions [8]float32
	}{
		{
			system: [8][8]float32{
				{1, 0, 0, 0, 0, 0, 0, 0},
				{0, 1, 0, 0, 0, 0, 0, 0},
				{0, 0, 1, 0, 0, 0, 0, 0},
				{0, 0, 0, 1, 0, 0, 0, 0},
				{0, 0, 0, 0, 1, 0, 0, 0},
				{0, 0, 0, 0, 0, 1, 0, 0},
				{0, 0, 0, 0, 0, 0, 1, 0},
				{0, 0, 0, 0, 0, 0, 0, 1},
			},
			weights:   [8]float32{1, 2, 3, 4, 5, 6, 7, 8},
			solutions: [8]float32{1, 2, 3, 4, 5, 6, 7, 8},
		},

		{
			system: [8][8]float32{
				{1, 1, 0, 0, 0, 0, 0, 0},
				{1, -1, 0, 0, 0, 0, 0, 0},
				{0, 0, 1, 0, 0, 0, 0, 0},
				{0, 0, 0, 1, 0, 0, 0, 0},
				{0, 0, 0, 0, 1, 0, 0, 0},
				{0, 0, 0, 0, 0, 1, 0, 0},
				{0, 0, 0, 0, 0, 0, 1, 0},
				{0, 0, 0, 0, 0, 0, 0, 1},
			},
			weights:   [8]float32{3, 1, 5, 6, 7, 8, 9, 10},
			solutions: [8]float32{2, 1, 5, 6, 7, 8, 9, 10},
		},
	}

	for i, test := range tests {
		solutions := gaussSolver8x8(test.system, test.weights)
		if !similarSliceF32(solutions[:], test.solutions[:], tolerance) {
			t.Fatalf("test #%d, expected %v, got %v", i, test.solutions, solutions)
		}
	}
}

func similarSliceF32(a, b []float32, tolerance float32) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if abs(a[i]-b[i]) < tolerance {
			continue // positive logic for NaNs
		}
		return false
	}
	return true
}

func TestComputeHomography(t *testing.T) {
	const tolerance float32 = 0.001

	uvQuad := [4]PointF32{{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 1, Y: 1}, {X: 0, Y: 1}}
	var tests = []struct {
		fromQuad  [4]PointF32
		outMatrix [9]float32
	}{
		{
			fromQuad: [4]PointF32{{0, 0}, {1, 0}, {1, 1}, {0, 1}},
			outMatrix: [9]float32{
				1, 0, 0,
				0, 1, 0,
				0, 0, 1,
			},
		},
		{ // scale
			fromQuad: [4]PointF32{{0, 0}, {2, 0}, {2, 2}, {0, 2}},
			outMatrix: [9]float32{
				0.5, 0, 0,
				0, 0.5, 0,
				0, 0, 1,
			},
		},
		{ // translate
			fromQuad: [4]PointF32{{1, 1}, {2, 1}, {2, 2}, {1, 2}},
			outMatrix: [9]float32{
				1, 0, -1,
				0, 1, -1,
				0, 0, 1,
			},
		},
		{ // scale and translate
			fromQuad: [4]PointF32{{1, 1}, {3, 1}, {3, 3}, {1, 3}},
			outMatrix: [9]float32{
				0.5, 0, -0.5,
				0, 0.5, -0.5,
				0, 0, 1,
			},
		},
		{ // 90 degs rotation
			fromQuad: [4]PointF32{{1, 0}, {1, 1}, {0, 1}, {0, 0}},
			outMatrix: [9]float32{
				0, 1, 0,
				-1, 0, 1,
				0, 0, 1,
			},
		},
		{ // perspective transform, wider bottom trapezoid
			fromQuad: [4]PointF32{{0, 0}, {1, 0}, {1.5, 1}, {-0.5, 1}},
			outMatrix: [9]float32{
				1, 0.5, 0,
				0, 2, 0,
				0, 1, 1,
			},
		},
		{ // arbitrary perspective transform
			fromQuad: [4]PointF32{{0, 0}, {1.5, 0.1}, {1.1, 1.2}, {-0.2, 0.9}},
			outMatrix: [9]float32{ // ~approx
				0.827844, 0.183965, 0,
				-0.06694, 1.004189, 0,
				0.176956, -0.052721, 1,
			},
		},
	}

	for i, test := range tests {
		matrix := computeHomography(test.fromQuad, uvQuad)
		if !similarSliceF32(test.outMatrix[:], matrix[:], tolerance) {
			t.Fatalf("test #%d, expected %v, got %v", i, test.outMatrix, matrix)
		}
	}
}

// go test -run ^TestCircIntersect$ . -count 1
func TestCircIntersect(t *testing.T) {
	updater := func(TestAppCtx) {}
	drawer := func(canvas *ebiten.Image, ctx TestAppCtx) {
		w, h := rectSizeF32(canvas.Bounds())
		cx, cy := w/2.0, h/2.0
		lx, ly := ctx.LeftClickF32()

		rA, rB := min(w, h)*0.4, float32(32.0)
		if ctx.SpacePressed {
			rA, rB = rB, rA
		}
		xy1, xy2, numSolutions := circIntersect(float64(cx), float64(cy), float64(rA), float64(lx), float64(ly), float64(rB))

		ctx.Renderer.SetColorF32(0.8, 0.8, 0.8, 0.8)
		ctx.Renderer.StrokeCircle(canvas, cx, cy, rA, 3.0)
		ctx.Renderer.StrokeCircle(canvas, lx, ly, rB, 3.0)

		if numSolutions > 0 {
			ctx.Renderer.SetColorF32(0.8, 0.0, 0.8, 0.8)
			ctx.Renderer.FillCircle(canvas, float32(xy1[0]), float32(xy1[1]), 3.0)
		}
		if numSolutions > 1 {
			ctx.Renderer.SetColorF32(0.0, 0.8, 0.8, 0.8)
			ctx.Renderer.FillCircle(canvas, float32(xy2[0]), float32(xy2[1]), 3.0)
		}
	}

	app := NewTestApp(updater, drawer)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}
