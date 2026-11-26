package shapes

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

// go test -run ^TestCircIntersect$ . -count 1
func TestCircIntersect(t *testing.T) {
	updater := func(TestAppCtx) {}
	drawer := func(canvas *ebiten.Image, ctx TestAppCtx) {
		w, h := rectSizeF32(canvas.Bounds())
		cx, cy := w/2.0, h/2.0
		lc := ctx.LeftClickF32()

		rA, rB := min(w, h)*0.4, float32(32.0)
		if ctx.SpacePressed {
			rA, rB = rB, rA
		}
		xy1, xy2, numSolutions := circIntersect(float64(cx), float64(cy), float64(rA), float64(lc.X), float64(lc.Y), float64(rB))

		ctx.Renderer.SetColorF32(0.8, 0.8, 0.8, 0.8)
		ctx.Renderer.StrokeCircle(canvas, cx, cy, rA, 3.0)
		ctx.Renderer.StrokeCircle(canvas, lc.X, lc.Y, rB, 3.0)

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
