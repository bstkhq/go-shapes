package shapes

import "testing"

func TestQuadSkeletonShape(t *testing.T) {
	tests := []struct {
		quad     [4]PointF32
		outShape shapeType
	}{
		{ // normal quad A
			quad:     [4]PointF32{PtF32(0, 0), PtF32(100, 0), PtF32(100, 100), PtF32(20, 90)},
			outShape: shapeTriangle,
		},
		{ // normal quad B
			quad:     [4]PointF32{PtF32(10, 0), PtF32(8, 10), PtF32(0, 4), PtF32(0, 1)},
			outShape: shapeTriangle,
		},
		{ // parallelogram with flat top/bottom
			quad:     [4]PointF32{PtF32(0, 0), PtF32(100, 0), PtF32(120, 100), PtF32(20, 100)},
			outShape: shapeLine,
		},
		{ // symmetric quad A
			quad:     [4]PointF32{PtF32(0, 0), PtF32(100, 0), PtF32(100, 100), PtF32(10, 90)},
			outShape: shapeLine,
		},
		{ // point
			quad:     [4]PointF32{PtF32(50, 50), PtF32(50, 50), PtF32(50, 50), PtF32(50, 50)},
			outShape: shapePoint,
		},
		{ // line A
			quad:     [4]PointF32{PtF32(0, 50), PtF32(100, 50), PtF32(100, 50), PtF32(0, 50)},
			outShape: shapeLine,
		},
		{ // line B
			quad:     [4]PointF32{PtF32(0, 0), PtF32(0, 0), PtF32(10, 0), PtF32(0, 0)},
			outShape: shapeLine,
		},
		// { // symmetric bowtie A
		// 	quad:     [4]PointF32{PtF32(0, 0), PtF32(100, 100), PtF32(100, 0), PtF32(0, 100)},
		// 	outShape: shapeNonSimple,
		// },
		// { // symmetric bowtie B
		// 	quad:     [4]PointF32{PtF32(6, 8), PtF32(0, 0), PtF32(6, 0), PtF32(0, 8)},
		// 	outShape: shapeNonSimple,
		// },
		// { // asymmetric bowtie A
		// 	quad:     [4]PointF32{PtF32(0, 0), PtF32(10, 10), PtF32(0, 9), PtF32(0, 10)},
		// 	outShape: shapeNonSimple,
		// },
	}

	for i, test := range tests {
		edges := quadNormalizedEdges(test.quad)
		result := firstQuadSkeletonOffset(test.quad, edges, Float32Inf())
		if result.Shape != test.outShape {
			t.Errorf("test #%d: expected shape %v, got %v", i, test.outShape, result.Shape)
		}
	}
}

func TestShrinkLine(t *testing.T) {
	tests := []struct {
		in1           PointF32
		in2           PointF32
		offset        float32
		out1          PointF32
		out2          PointF32
		offsetReached float32
		shape         shapeType
	}{
		// zero offset
		{in1: PtF32(-1, 0), in2: PtF32(1, 0), offset: 0.0, out1: PtF32(-1, 0), out2: PtF32(1, 0), offsetReached: 0.0, shape: shapeLine},

		// basic cases
		{in1: PtF32(-1, 0), in2: PtF32(1, 0), offset: 0.5, out1: PtF32(-0.5, 0), out2: PtF32(0.5, 0), offsetReached: 0.5, shape: shapeLine},
		{in1: PtF32(-1, 0), in2: PtF32(1, 0), offset: 1.0, out1: PtF32(0, 0), out2: PtF32(0, 0), offsetReached: 1.0, shape: shapePoint},
		{in1: PtF32(-1, 0), in2: PtF32(1, 0), offset: 1.5, out1: PtF32(0, 0), out2: PtF32(0, 0), offsetReached: 1.0, shape: shapePoint},
		{in1: PtF32(-1, 0), in2: PtF32(1, 0), offset: 50.0, out1: PtF32(0, 0), out2: PtF32(0, 0), offsetReached: 1.0, shape: shapePoint},
		{in1: PtF32(0, -2), in2: PtF32(0, 2), offset: 0.5, out1: PtF32(0, -1.5), out2: PtF32(0, 1.5), offsetReached: 0.5, shape: shapeLine},
		{in1: PtF32(0, -2), in2: PtF32(0, 2), offset: 2.0, out1: PtF32(0, 0), out2: PtF32(0, 0), offsetReached: 2.0, shape: shapePoint},
		{in1: PtF32(0, -2), in2: PtF32(0, 2), offset: 2.5, out1: PtF32(0, 0), out2: PtF32(0, 0), offsetReached: 2.0, shape: shapePoint},

		// point inputs
		{in1: PtF32(0, 0), in2: PtF32(0, 0), offset: 0.0, out1: PtF32(0, 0), out2: PtF32(0, 0), offsetReached: 0.0, shape: shapePoint},
		{in1: PtF32(0, 0), in2: PtF32(0, 0), offset: 1.0, out1: PtF32(0, 0), out2: PtF32(0, 0), offsetReached: 0.0, shape: shapePoint},
		{in1: PtF32(10, 20), in2: PtF32(10, 20), offset: 5.0, out1: PtF32(10, 20), out2: PtF32(0, 0), offsetReached: 0.0, shape: shapePoint},
	}

	for i, test := range tests {
		out1, out2, shape, offsetReached := shrinkLine(test.in1, test.in2, test.offset)
		if out1 != test.out1 || out2 != test.out2 || shape != test.shape || offsetReached != test.offsetReached {
			t.Errorf("test #%d: expected %v, %v, %s, %v, got %v, %v, %s, %v", i, test.out1, test.out2, test.shape.String(), test.offsetReached, out1, out2, shape.String(), offsetReached)
		}
	}
}

func TestCanonicalizeQuadCW(t *testing.T) {
	tests := []struct {
		in               [4]PointF32
		out              [4]PointF32
		selfIntersecting bool
	}{
		{ // convex CW
			in:  [4]PointF32{PtF32(0, 0), PtF32(1, 0), PtF32(1, 1), PtF32(0, 1)},
			out: [4]PointF32{PtF32(0, 0), PtF32(1, 0), PtF32(1, 1), PtF32(0, 1)},
		},
		{ // convex CCW
			in:  [4]PointF32{PtF32(0, 0), PtF32(0, 1), PtF32(1, 1), PtF32(1, 0)},
			out: [4]PointF32{PtF32(0, 0), PtF32(1, 0), PtF32(1, 1), PtF32(0, 1)},
		},
		{ // concave CW
			in:  [4]PointF32{PtF32(0, 0), PtF32(3, 0), PtF32(1, 1), PtF32(0, 3)},
			out: [4]PointF32{PtF32(0, 0), PtF32(3, 0), PtF32(1, 1), PtF32(0, 3)},
		},
		{ // concave CW (shifted)
			in:  [4]PointF32{PtF32(3, 0), PtF32(1, 1), PtF32(0, 3), PtF32(0, 0)},
			out: [4]PointF32{PtF32(0, 0), PtF32(3, 0), PtF32(1, 1), PtF32(0, 3)},
		},
		{ // concave CCW
			in:  [4]PointF32{PtF32(0, 0), PtF32(0, 3), PtF32(1, 1), PtF32(3, 0)},
			out: [4]PointF32{PtF32(0, 0), PtF32(3, 0), PtF32(1, 1), PtF32(0, 3)},
		},
		{ // bow-tie
			in:               [4]PointF32{PtF32(0, 0), PtF32(1, 0), PtF32(0, 1), PtF32(1, 1)},
			selfIntersecting: true,
		},
		{ // degenerate triangle
			in:  [4]PointF32{PtF32(0, 0), PtF32(0, 1), PtF32(1, 0), PtF32(1, 0)},
			out: [4]PointF32{PtF32(0, 0), PtF32(1, 0), PtF32(1, 0), PtF32(0, 1)},
		},
		{ // collinear
			in:  [4]PointF32{PtF32(0, 0), PtF32(1, 0), PtF32(2, 0), PtF32(3, 0)},
			out: [4]PointF32{PtF32(0, 0), PtF32(1, 0), PtF32(2, 0), PtF32(3, 0)},
		},
	}

	for i, test := range tests {
		out, ok := canonicalizeQuadCW(test.in)
		selfIntersecting := !ok
		if selfIntersecting != test.selfIntersecting {
			t.Fatalf("test#%d: expected selfIntersecting=%t, got %t", i, test.selfIntersecting, selfIntersecting)
		}
		if test.selfIntersecting {
			continue
		}

		for pi, p := range test.out {
			if out[pi] != p {
				t.Fatalf("test#%d: expected %v, got %v", i, test.out, out)
				continue
			}
		}
	}
}
