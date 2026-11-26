package shapes

import (
	"fmt"
	"image"
	"image/color"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/bstkhq/go-shapes/internal/pixcheck"
)

// A stroke covers the rect a fill of its outer edge would, except for a hole
// inset by the total thickness, which disappears once the stroke is thick
// enough to cover the whole rect. With a translucent color, triangles that
// overlap would blend some pixels twice, which happened when thick strokes
// inverted the inner rect.
func TestStrokeIntRectNoOverlap(t *testing.T) {
	sizes := []image.Point{{60, 24}, {61, 25}, {80, 1}, {80, 0}, {0, 0}}
	var thicknesses []struct{ in, out int }
	for in := range 15 { // inner thickness only, up to past half of every size
		thicknesses = append(thicknesses, struct{ in, out int }{in, 0})
	}
	thicknesses = append(thicknesses, []struct{ in, out int }{{40, 0}, {0, 3}, {4, 4}, {10, 10}, {3, 12}}...)

	report := pixcheck.Report{Rule: pixcheck.Rule{MaxDelta: 1}}
	pixcheck.Run(t, func() {
		r := NewRenderer()
		r.SetColor(color.RGBA{128, 128, 128, 128})
		for _, size := range sizes {
			for _, thick := range thicknesses {
				if thick.in+thick.out == 0 {
					continue // nothing to draw
				}
				rect := RectWithSize(8, 8, size.X, size.Y)
				outer := rect.Inset(-thick.out)
				stroke := ebiten.NewImage(100, 60)
				r.StrokeIntRect(stroke, rect, thick.in, thick.out)
				fill := ebiten.NewImage(100, 60)
				r.FillIntRect(fill, outer, 0)

				// Inset returns an empty rect near the center when the
				// thickness is more than half of a side
				want := pixcheck.Read(fill)
				hole := outer.Inset(thick.in + thick.out)
				for y := hole.Min.Y; y < hole.Max.Y; y++ {
					for x := hole.Min.X; x < hole.Max.X; x++ {
						want.SetRGBA(x, y, color.RGBA{})
					}
				}
				label := fmt.Sprintf("size %v, in %d, out %d", size, thick.in, thick.out)
				report.CompareRGBA(label, pixcheck.Read(stroke), want)
			}
		}
	})
	report.Check(t)
}
