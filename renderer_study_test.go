package shapes

import (
	"fmt"
	"image/color"
	"math"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

// go test -run ^TestStudyWaveFuncs$ . -count 1
func TestStudyWaveFuncs(t *testing.T) {
	updater := func(TestAppCtx) {}
	drawer := func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Fill(color.RGBA{128, 0, 0, 255})
		ctx.Renderer.studyWaveFuncs(canvas, 1.0, 8.0)
	}

	app := NewTestApp(updater, drawer)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestStudyRadians$ . -count 1
func TestStudyRadians(t *testing.T) {
	const PointRadius = 3.5
	const Dist = 96.0

	rads := -math.Pi
	updater := func(TestAppCtx) { rads += 0.01 }
	drawer := func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Fill(color.Black)
		w, h := rectSizeF32(canvas.Bounds())
		cx, cy := w/2.0, h/2.0
		ctx.Renderer.FillCircle(canvas, cx, cy, PointRadius)

		sin, cos := math.Sincos(normURads(rads))
		ctx.Renderer.FillCircle(canvas, cx+float32(Dist*cos), cy+float32(Dist*sin), PointRadius)
		ebiten.SetWindowTitle(fmt.Sprintf("%s [rads: %02f]", ctx.Title(), rads))
	}

	app := NewTestApp(updater, drawer)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestStudyGaussian$ . -count 1
func TestStudyGaussian(t *testing.T) {
	const Radius = 64.0
	const H = 192.0
	const Sigma = Radius / 3.0
	const Sigma2 = 2.0 * Sigma * Sigma
	const MarkerRadius = 2.0

	updater := func(TestAppCtx) {}
	drawer := func(canvas *ebiten.Image, ctx TestAppCtx) {
		w, h := rectSizeF64(canvas.Bounds())
		canvas.Fill(color.Black)

		gaussian := func(x float64) float32 {
			return float32(H * math.Exp(-(x*x)/Sigma2))
		}

		cx, by := float32(w*0.5), float32(h*0.666)
		ctx.Renderer.FillCircle(canvas, cx, by-gaussian(0.0), MarkerRadius)
		for i := range 30 {
			x := float64(i+1) * (Radius * 0.05)
			y := gaussian(x)
			ctx.Renderer.FillCircle(canvas, cx+float32(x), by-y, MarkerRadius)
			ctx.Renderer.FillCircle(canvas, cx-float32(x), by-y, MarkerRadius)
		}
	}

	app := NewTestApp(updater, drawer)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}
