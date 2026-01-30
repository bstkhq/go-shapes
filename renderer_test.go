package shapes

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

// go test -run ^TestScale$ . -count 1
func TestScale(t *testing.T) {
	app := NewTestApp(func(canvas *ebiten.Image, ctx TestAppCtx) {
		_, _, w, h := rectOriginSize(canvas.Bounds())
		scale := float64(min(w, h)) / 6.0

		var opts ebiten.DrawImageOptions
		opts.GeoM.Scale(scale, scale)
		canvas.DrawImage(ctx.Images[0], &opts)

		opts.Filter = ebiten.FilterLinear
		opts.GeoM.Reset()
		opts.GeoM.Translate(-3, 0)
		opts.GeoM.Scale(scale, scale)
		opts.GeoM.Translate(float64(w), 0)
		canvas.DrawImage(ctx.Images[0], &opts)

		size := float32(3 * scale)
		var scOpts ScaleOptions
		scOpts.Clamp = ebiten.IsKeyPressed(ebiten.KeyC)
		scOpts.DstSampling = ebiten.IsKeyPressed(ebiten.KeyD)

		ctx.Renderer.Scale(canvas, ctx.Images[0], 0, float32(h)-size, float32(scale), &scOpts)
		scOpts.Bicubic = true
		ctx.Renderer.Scale(canvas, ctx.Images[0], float32(w)-size, float32(h)-size, float32(scale), &scOpts)
	})

	sq9 := ebiten.NewImage(3, 3)
	sq9.WritePixels([]byte{
		255, 0, 0, 255 /* */, 0, 128, 255, 255 /* */, 255, 128, 0, 255,
		0, 0, 255, 255 /* */, 255, 0, 255, 255 /* */, 0, 255, 0, 255,
		128, 255, 0, 255 /* */, 255, 255, 0, 255 /* */, 0, 255, 255, 255,
	})
	app.Images = append(app.Images, sq9)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}
