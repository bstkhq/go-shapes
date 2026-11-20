package shapes

import (
	"fmt"
	"math"
	"testing"
)

// When drawing a stroked rect with hull bounding (the default), we skip a
// rectangle at its center (the "hole"), since the stroke leaves that area
// empty anyway (with AABB bounding, the whole rect is drawn instead). For a
// wide range of sizes, thicknesses and roundings, this test checks that the
// hole:
//   - never overlaps the stroke, which would make pixels go missing,
//   - is close to the biggest hole possible, so we don't process extra pixels,
//   - doesn't jump when the rect goes from wide to tall, which would be
//     visible on animations with gradient colors.
func TestStrokeRectHoleHalfSize(t *testing.T) {
	// distance from a point (relative to the rect center) to the outer edge
	// of the stroke, negative inside. same as distanceToRoundedRect in
	// stroke_rect.kage. points deeper than the stroke thickness are empty
	distance := func(x, y, w, h, rounding float64) float64 {
		dx, dy := math.Abs(x)-(w/2-rounding), math.Abs(y)-(h/2-rounding)
		return math.Hypot(max(dx, 0), max(dy, 0)) + min(max(dx, dy), 0) - rounding
	}

	var worstExtra float64
	var worstCase string
	for w := float32(4); w <= 160; w += 6 {
		for h := float32(4); h <= 160; h += 7 {
			for thick := float32(1); thick <= 40; thick += 5 {
				for rounding := float32(0); rounding <= 2*min(w, h); rounding += 5 {
					desc := fmt.Sprintf("w=%v h=%v thickness=%v rounding=%v", w, h, thick, rounding)
					halfW, halfH := strokeRectHoleHalfSize(w, h, thick, rounding)
					if halfW < 0 || halfH < 0 {
						t.Fatalf("negative hole size %v x %v for %s", halfW, halfH, desc)
					}

					// the empty area is convex, so if the hole corners are inside
					// it, the whole hole is
					W, H, T, R := float64(w), float64(h), float64(thick), float64(rounding)
					if halfW > 0 && halfH > 0 && distance(float64(halfW), float64(halfH), W, H, R) > -T+1e-3 {
						t.Fatalf("hole overlaps the stroke for %s", desc)
					}

					// find the biggest possible hole by trying many points along the
					// corner curve, and measure how many extra pixels we process
					// compared to it
					radius := max(R-T, 0)
					curveX, curveY := W/2-T-radius, H/2-T-radius
					var bestArea float64
					for i := range 301 {
						sin, cos := math.Sincos(math.Pi / 2 * float64(i) / 300)
						bestArea = max(bestArea, 4*max(curveX+radius*cos, 0)*max(curveY+radius*sin, 0))
					}
					extra := (bestArea - 4*float64(halfW)*float64(halfH)) / (W*H - bestArea)
					if extra > worstExtra {
						worstExtra, worstCase = extra, desc
					}
				}
			}
		}
	}
	t.Logf("worst case processes %.3f%% more pixels than the best hole (%s)", worstExtra*100, worstCase)
	if worstExtra > 0.03 { // ~3% is expected for very thin rects with very big rounding
		t.Fatalf("hole too small: %.3f%% more pixels than the best hole (%s)", worstExtra*100, worstCase)
	}

	// an earlier implementation chose the hole differently for wide and tall
	// rects, so the hole jumped when width and height crossed
	halfW0, halfH0 := strokeRectHoleHalfSize(100, 100-1e-3, 8, 25)
	halfW1, halfH1 := strokeRectHoleHalfSize(100, 100+1e-3, 8, 25)
	if abs(halfW1-halfW0) > 0.01 || abs(halfH1-halfH0) > 0.01 {
		t.Fatalf("hole jumps when w == h: %v x %v vs %v x %v", halfW0, halfH0, halfW1, halfH1)
	}
}
