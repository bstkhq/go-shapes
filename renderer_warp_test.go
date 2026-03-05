package shapes

import (
	"image"
	"image/color"
	"math"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

// go test -run ^TestWarpBarrel$ . -count 1
func TestWarpBarrel(t *testing.T) {
	app := NewTestApp(func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Fill(color.Black)
		d := 4.0 - float32(ctx.DistAnim(8.0, 0.5))
		ctx.Renderer.WarpBarrel(canvas, ctx.Images[0], 64, 32, d/2.0, d)
	})
	w, h := 640/2, 480/2
	img := ebiten.NewImage(w, h)
	app.Renderer.SetColor(color.RGBA{255, 0, 0, 255}, 0)
	app.Renderer.SetColor(color.RGBA{255, 255, 0, 255}, 1)
	app.Renderer.SetColor(color.RGBA{0, 255, 0, 255}, 2)
	app.Renderer.SetColor(color.RGBA{0, 255, 255, 255}, 3)
	app.Renderer.DrawIntArea(img, 0, 0, w, h)

	app.Images = append(app.Images, img)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestWarpArc$ . -count 1
func TestWarpArc(t *testing.T) {
	app := NewTestApp(func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Fill(color.Black)
		ctx.DrawAtF32(canvas, ctx.Images[0], 0, 0)
		outRadius := 64 + float32(ctx.DistAnim(172, 0.5))
		rads := ctx.ModAnim(2*math.Pi, 0.5)
		//rads = RadsBottomRight
		cw, ch := rectSizeF32(canvas.Bounds())
		ctx.Renderer.WarpArc(canvas, ctx.Images[0], cw/2.0, ch/2.0, outRadius, rads)
	})

	const W, H = 512, 64
	img := ebiten.NewImage(W, H)
	from, to := color.RGBA{0, 0, 0, 255}, color.RGBA{255, 255, 255, 255}
	mask := app.Renderer.NewSimpleGradient(W, H, from, to, DirRadsRTL)
	app.Renderer.DitherMat4(img, mask, 0, 0, 0, 0, PaletteBW, DitherGlitch, 0.0, 0.0)
	app.Renderer.SetColorF32(0, 0.5, 0, 0.5)
	app.Renderer.DrawRect(img, img.Bounds(), 0)
	app.Renderer.SetColorF32(0.5, 0.0, 0.5, 0.5)
	app.Renderer.DrawRect(img, image.Rect(0, 0, W/8, H/16), 0)
	app.Renderer.SetColorF32(1, 1, 1, 1)
	app.Images = append(app.Images, img)

	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}
