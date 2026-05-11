package shapes

import (
	"image"
	"image/color"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

// go test -run ^TestHalftoneTri$ . -count 1
func TestHalftoneTri(t *testing.T) {
	updater := func(TestAppCtx) {}
	drawer := func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Fill(color.Black)
		if !ctx.SpacePressed {
			ctx.Renderer.SetTint(float32(ctx.DistAnim(1.0, 1.0)))
			ctx.Renderer.SetColorF32(1.0, 0.5, 0, 1.0)
			size := float32(16.0)
			xOffset := float32(ctx.ModAnim(float64(size*2.0), 1.0))
			ctx.Renderer.HalftoneTri(canvas, ctx.Images[0], 0, 0, size, size*0.2, size*1.0, xOffset, 0)
			ctx.Renderer.SetColorF32(1.0, 1.0, 1.0, 1.0)
			ctx.Renderer.SetTint(0)
		} else {
			ctx.Renderer.DrawAt(canvas, ctx.Images[0], 0, 0, 1.0)
		}
	}

	app := NewTestApp(updater, drawer)
	img := ebiten.NewImage(640, 480)
	app.Renderer.SetColorF32(0.2, 0.2, 0.2, 0.2, 0, 1)
	app.Renderer.SetColorF32(1.0, 1.0, 1.0, 1.0, 2, 3)
	app.Renderer.FillIntRect(img, image.Rect(0, 0, 640, 480), 0)
	app.Renderer.SetColorF32(0.666, 0.666, 0.666, 0.666)
	app.Renderer.FillCircle(img, 640/2, 480/3, 64)
	app.Renderer.SetColorF32(1.0, 1.0, 1.0, 1.0)

	app.Images = append(app.Images, img)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}
