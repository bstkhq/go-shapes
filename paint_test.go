package shapes

import (
	"image"
	"image/color"
	"sync/atomic"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

// go test -run ^TestPaint$ . -count 1
func TestPaint(t *testing.T) {
	const RectWidth, RectHeight = 128, 64
	app := NewTestApp(func(canvas *ebiten.Image, ctx TestAppCtx) {
		const DrawRepeats = 2000 // for benchmarking

		canvas.Fill(color.RGBA{255, 0, 0, 255})

		restore := false
		if ebiten.IsKeyPressed(ebiten.KeySpace) {
			restore = atomic.CompareAndSwapUint32(&paintInUse, 0, 1)
		}

		rect := image.Rect(32, 32, 32+RectWidth, 32+RectHeight)
		if ebiten.IsKeyPressed(ebiten.KeyC) {
			canvas.SubImage(rect).(*ebiten.Image).Clear()
		} else if ebiten.IsKeyPressed(ebiten.KeyP) {
			a := float32(ctx.DistAnim(1, 1))
			for range DrawRepeats {
				Paint(canvas, rect, [4]float32{0, a, 0, a}, ebiten.BlendSourceOver)
			}
		} else if ebiten.IsKeyPressed(ebiten.KeyZ) {
			a := float32(ctx.DistAnim(1, 1))
			clr := color.RGBA{0, uint8(255 * a), 0, uint8(255 * a)}
			for range DrawRepeats {
				canvas.SubImage(rect).(*ebiten.Image).Fill(clr)
			}
		} else if ebiten.IsKeyPressed(ebiten.KeyR) {
			var opts ebiten.DrawImageOptions
			a := float32(ctx.DistAnim(1, 1))
			clr := color.RGBA{0, uint8(255 * a), 0, uint8(255 * a)}
			opts.ColorScale.ScaleWithColor(clr)
			opts.GeoM.Translate(32, 32)
			for range DrawRepeats {
				canvas.DrawImage(ctx.Images[0], &opts)
			}
		} else {
			Paint(canvas, rect, White, ebiten.BlendClear)
		}

		if restore {
			atomic.StoreUint32(&paintInUse, 0)
		}
	})

	img := ebiten.NewImage(RectWidth, RectHeight)
	img.Fill(color.White)
	app.Images = append(app.Images, img)

	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}
