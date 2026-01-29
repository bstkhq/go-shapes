package shapes

import (
	"fmt"
	"image/color"
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
					downscale = DownscaleX2
				}
			}
		}
		ebiten.SetWindowTitle(ctx.Title() + fmt.Sprintf(" [sampling = %d, downscaling x%d]", sampling, downscale.Factor()))

		cw, ch := rectSizeF32(canvas.Bounds())
		ctx.Renderer.ApplyBlur(canvas, ctx.Images[0], 16, ch-16-radius*2, FxRadius, 1.0)
		ctx.Renderer.ApplyBlur(canvas, ctx.Images[0], 32+radius*2, ch-16-radius*2, FxRadius, 1.0)
		ctx.Renderer.SetColor(color.RGBA{0, 0, 255, 255})
		ctx.Renderer.ApplyBlurVogel(canvas, ctx.Images[0], 16, ch-16-radius*2, FxRadius, 0.0, sampling, downscale)
		ctx.Renderer.ApplyBlur(canvas, ctx.Images[0], 32+radius*2, ch-16-radius*2, FxRadius, 0.0)

		ctx.Renderer.ApplyBlurVogel(canvas, ctx.Images[1], cw-float32(ctx.Images[1].Bounds().Dx())-16, 16, FxRadius, 1.0, sampling, downscale)

		ctx.Renderer.SetColor(color.RGBA{255, 0, 0, 255})
		clrMix := float32(ctx.DistAnim(1.0, 1.0))

		lx, ly := ctx.LeftClickF32()
		ctx.Renderer.ApplyBlur(canvas, ctx.Images[0], lx-radius, ly-radius, FxRadius, clrMix)

		rx, ry := ctx.RightClickF32()
		ctx.Renderer.SetColor(color.RGBA{255, 0, 0, 255})
		ctx.Renderer.ApplyBlurVogel(canvas, ctx.Images[0], rx-radius, ry-radius, FxRadius, clrMix, sampling, downscale)

	})

	circle := app.Renderer.NewCircle(float64(radius))
	rect := ebiten.NewImage(120, 80)
	app.Renderer.StrokeIntRect(rect, rect.Bounds(), 0, 16)
	app.Images = append(app.Images, circle, rect)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}
