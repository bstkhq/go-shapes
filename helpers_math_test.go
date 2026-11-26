package shapes

import "testing"

func TestFloorCeilF32(t *testing.T) {
	tests := []struct {
		value float32
		floor float32
		ceil  float32
	}{
		{-1.75, -2, -1},
		{-1, -1, -1},
		{-0.25, -1, 0},
		{0, 0, 0},
		{0.25, 0, 1},
		{1, 1, 1},
		{1.75, 1, 2},
	}

	for _, test := range tests {
		if got := floorF32(test.value); got != test.floor {
			t.Errorf("floorF32(%v): got %v, want %v", test.value, got, test.floor)
		}
		if got := ceilF32(test.value); got != test.ceil {
			t.Errorf("ceilF32(%v): got %v, want %v", test.value, got, test.ceil)
		}
	}
}

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
