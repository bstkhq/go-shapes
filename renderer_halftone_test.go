package shapes

import (
	"image/color"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

// go test -run ^TestHalftoneTri$ . -count 1
func TestHalftoneTri(t *testing.T) {
	app := NewTestApp(func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Fill(color.Black)
		if ebiten.IsKeyPressed(ebiten.KeySpace) {
			ctx.Renderer.DrawAt(canvas, ctx.Images[0], 0, 0, 1.0)
		} else {
			ctx.Renderer.SetTint(float32(ctx.DistAnim(1.0, 1.0)))
			ctx.Renderer.SetColorF32(1.0, 0.5, 0, 1.0)
			ctx.Renderer.HalftoneTri(canvas, ctx.Images[0], 0, 0, 16.0, 6.0, 15.0, 0, 0)
			ctx.Renderer.SetColorF32(1.0, 1.0, 1.0, 1.0)
			ctx.Renderer.SetTint(0)
		}
	})
	img := ebiten.NewImage(640, 480)
	app.Renderer.SetColorF32(0.2, 0.2, 0.2, 0.2, 0, 1)
	app.Renderer.SetColorF32(1.0, 1.0, 1.0, 1.0, 2, 3)
	app.Renderer.DrawIntArea(img, 0, 0, 640, 480)
	app.Renderer.SetColorF32(0.666, 0.666, 0.666, 0.666)
	app.Renderer.DrawCircle(img, 640/2, 480/3, 64)
	app.Renderer.SetColorF32(1.0, 1.0, 1.0, 1.0)

	app.Images = append(app.Images, img)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}
