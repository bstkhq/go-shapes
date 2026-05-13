package shapes

import (
	"image"
	"image/color"
	"math"
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
			xOffset := float32(ctx.ModAnim(float64(size*2.0), 5.0))
			yOffset := float32(ctx.ModAnim(float64(size*Sqrt3Div2)*2.0, 3.0))
			ctx.Renderer.HalftoneTri(canvas, ctx.Images[0], 0, 0, size, size*0.2, size*1.0, xOffset, yOffset)
			ctx.Renderer.SetColorF32(1.0, 1.0, 1.0, 1.0)
			ctx.Renderer.SetTint(0)

			ctx.Renderer.SetColorF32(0.0, 0.3, 1.0, 1.0)
			ctx.Renderer.FillCircle(canvas, 8+xOffset, 8+yOffset, 4.0)
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

// go test -run ^TestScanlinesSharp . -count 1
func TestScanlinesSharp(t *testing.T) {
	updater := func(TestAppCtx) {}
	drawer := func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Fill(color.White)
		const darkThick, clearThick = 3, 1
		offset := float32(ctx.ModAnim(darkThick+clearThick, 1.0))
		ctx.Renderer.ScanlinesSharp(canvas, darkThick, clearThick, 0.05, offset)
	}

	app := NewTestApp(updater, drawer)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestWaveLines$ . -count 1
func TestWaveLines(t *testing.T) {
	const LineThick = 6.0

	offset := float32(0.0)
	updater := func(TestAppCtx) {}
	drawer := func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Fill(color.White)
		offset += 0.666
		minFillRate := float32(0.2)
		maxFillRate := float32(0.8)
		radsOffset := ctx.ModAnim(2*math.Pi, 0.2)
		if ctx.SpacePressed {
			radsOffset = 0.0
		}

		dir := DirRadsLTR + radsOffset
		ctx.Renderer.SetColorF32(0, 0, 0, 0.2, 0)
		ctx.Renderer.SetColorF32(0, 0.1, 0, 0.2, 1)
		ctx.Renderer.SetColorF32(0, 0.1, 0.1, 0.2, 3)
		ctx.Renderer.WaveLines(canvas, LineThick, minFillRate, maxFillRate, 16.0, offset, dir)

		sin, cos := math.Sincos(normURads(dir))
		ctx.Renderer.SetColorF32(0, 0.0, 0.0, 0.666)
		w, h := rectSizeF64(canvas.Bounds())
		ctx.Renderer.StrokeLine(canvas, w/2, h/2, w/2+(60.0*cos), h/2+(60.0*sin), 4.0)
	}

	app := NewTestApp(updater, drawer)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}
