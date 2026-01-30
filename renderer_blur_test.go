package shapes

import (
	"fmt"
	"image/color"
	"math/rand/v2"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// go test -run ^TestApplyBlurVogel$ . -count 1
func TestApplyBlurVogel(t *testing.T) {
	radius := float32(64.0)
	sampling := 8
	downscale := DownscaleNone
	app := NewTestApp(func(canvas *ebiten.Image, ctx TestAppCtx) {
		const FxRadius = 16
		canvas.Fill(color.RGBA{0, 0, 255, 255})

		if ctx.NewInput {
			if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
				sampling *= 2
				if sampling > 64 {
					sampling = 8
				}
			}
			if inpututil.IsKeyJustPressed(ebiten.KeyD) {
				downscale += 1
				if downscale > DownscaleX16 {
					downscale = DownscaleNone
				}
			}
		}
		seed := float32(1.0)
		if ebiten.IsKeyPressed(ebiten.KeyN) {
			seed = rand.Float32()
		}
		ebiten.SetWindowTitle(ctx.Title() + fmt.Sprintf(" [sampling = %d, downscaling x%d]", sampling, downscale.Factor()))

		cw, ch := rectSizeF32(canvas.Bounds())
		ctx.Renderer.ApplyBlur(canvas, ctx.Images[0], 16, ch-16-radius*2, FxRadius, 1.0)
		ctx.Renderer.ApplyBlur(canvas, ctx.Images[0], 32+radius*2, ch-16-radius*2, FxRadius, 1.0)
		ctx.Renderer.SetColor(color.RGBA{0, 0, 255, 255})
		ctx.Renderer.ApplyBlurVogel(canvas, ctx.Images[0], 16, ch-16-radius*2, FxRadius, 0.0, sampling, downscale, seed)
		ctx.Renderer.ApplyBlur(canvas, ctx.Images[0], 32+radius*2, ch-16-radius*2, FxRadius, 0.0)

		ctx.Renderer.ApplyBlurVogel(canvas, ctx.Images[1], cw-float32(ctx.Images[1].Bounds().Dx())-16, 16, FxRadius, 1.0, sampling, downscale, seed)

		ctx.Renderer.SetColor(color.RGBA{255, 0, 0, 255})
		clrMix := float32(ctx.DistAnim(1.0, 1.0))

		lx, ly := ctx.LeftClickF32()
		ctx.Renderer.ApplyBlur(canvas, ctx.Images[0], lx-radius, ly-radius, FxRadius, clrMix)

		rx, ry := ctx.RightClickF32()
		ctx.Renderer.SetColor(color.RGBA{255, 0, 0, 255})
		ctx.Renderer.ApplyBlurVogel(canvas, ctx.Images[0], rx-radius, ry-radius, FxRadius, clrMix, sampling, downscale, seed)

	})

	circle := app.Renderer.NewCircle(float64(radius))
	rect := ebiten.NewImage(120, 80)
	app.Renderer.StrokeIntRect(rect, rect.Bounds(), 0, 16)
	app.Images = append(app.Images, circle, rect)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestApplyBlurVogelFull$ . -count 1
func TestApplyBlurVogelFull(t *testing.T) {
	const Sampling = 16
	const Downscale = DownscaleNone

	var full *ebiten.Image = ebiten.NewImage(1920, 1080)

	app := NewTestApp(func(canvas *ebiten.Image, ctx TestAppCtx) {
		seed := float32(1.0)
		if ebiten.IsKeyPressed(ebiten.KeySpace) {
			seed = rand.Float32()
		}

		radius := ctx.DistAnim(64, 0.5)
		ctx.Renderer.ApplyBlurVogel(canvas, full, 0, 0, float32(radius), 1.0, Sampling, Downscale, seed)
	})

	app.Renderer.GradientRadialDither(full, 1920/2, 960, color.RGBA{0, 64, 196, 255}, color.RGBA{0, 128, 255, 255}, 320, 960, Float32Inf(), 1.0)
	app.Renderer.GradientDither(full, 0, 0, 1920, 1080, color.RGBA{64, 0, 64, 64}, color.RGBA{32, 0, 0, 64}, DirRadsBLTR, 1.0)
	for range 256 {
		rand.Float64()
		app.Renderer.SetColorF32(rand.Float32(), rand.Float32(), rand.Float32(), 1.0)
		app.Renderer.ScaleAlphaBy(0.5)
		app.Renderer.DrawCircle(full, rand.Float32()*1920, rand.Float32()*1080, 16+rand.Float32()*64)
	}
	app.Renderer.SetColorF32(1, 1, 1, 1)

	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}
