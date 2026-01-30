package shapes

import (
	"fmt"
	"image/color"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// go test -run ^TestNoise$ . -count 1
func TestNoise(t *testing.T) {
	move := true
	app := NewTestApp(func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Fill(color.Black)

		if ctx.NewInput && inpututil.IsKeyJustPressed(ebiten.KeyControl) {
			move = !move
		}

		sub := canvas
		if move {
			const PosShift = 96
			bounds := canvas.Bounds()
			xShift, yShift := ctx.DistAnim(PosShift, 1.0)-PosShift/2.0, ctx.DistAnim(PosShift, 0.75)-PosShift/2.0
			ixShift, iyShift := int(xShift), int(yShift)
			bounds.Min.X = bounds.Min.X + PosShift + ixShift
			bounds.Min.Y = bounds.Min.Y + PosShift + iyShift
			bounds.Max.X = bounds.Max.X - PosShift + ixShift
			bounds.Max.Y = bounds.Max.Y - PosShift + iyShift
			sub = canvas.SubImage(bounds).(*ebiten.Image)
		}

		anim := float32(ctx.ModAnim(1.0, 0.5))
		ctx.Renderer.Noise(sub, 0.8, 0.26, anim)
	})

	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestNoiseAspectRatio$ . -count 1
func TestNoiseAspectRatio(t *testing.T) {
	const SubWidth, SubHeight = 96 * 3, 96
	sub := ebiten.NewImage(SubWidth, SubHeight)
	app := NewTestApp(func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Fill(color.Black)

		anim := float32(ctx.ModAnim(1.0, 0.5))
		sub.Clear()
		ctx.Renderer.Noise(sub, 0.8, 0.26, anim)

		var opts ebiten.DrawImageOptions
		_, _, w, h := rectOriginSize(canvas.Bounds())
		wF64, hF64 := float64(w), float64(h)
		opts.GeoM.Translate(-SubWidth/2, -SubHeight/2)
		scale := min(wF64/SubWidth, hF64/SubHeight)
		opts.GeoM.Scale(scale, scale)
		opts.GeoM.Translate(wF64/2, hF64/2)
		canvas.DrawImage(sub, &opts)
	})

	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestNoiseGolden$ . -count 1
func TestNoiseGolden(t *testing.T) {
	scale := float32(1.0)
	move := true

	app := NewTestApp(func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Fill(color.Black)

		shift := ebiten.IsKeyPressed(ebiten.KeyShift)
		up := ctx.NewInput && inpututil.IsKeyJustPressed(ebiten.KeyArrowUp)
		down := ctx.NewInput && inpututil.IsKeyJustPressed(ebiten.KeyArrowDown)
		switch {
		case inpututil.IsKeyJustPressed(ebiten.KeyControl):
			move = !move
		case up && shift:
			scale += 1.0
		case down && shift:
			scale -= 1.0
		case up:
			scale *= 2.0
		case down:
			scale /= 2.0
		}
		ebiten.SetWindowTitle(ctx.Title() + fmt.Sprintf(" - scale: %.02f", scale))

		sub := canvas
		if move {
			const PosShift = 96
			bounds := canvas.Bounds()
			xShift, yShift := ctx.DistAnim(PosShift, 1.0)-PosShift/2.0, ctx.DistAnim(PosShift, 0.75)-PosShift/2.0
			ixShift, iyShift := int(xShift), int(yShift)
			bounds.Min.X = bounds.Min.X + PosShift + ixShift
			bounds.Min.Y = bounds.Min.Y + PosShift + iyShift
			bounds.Max.X = bounds.Max.X - PosShift + ixShift
			bounds.Max.Y = bounds.Max.Y - PosShift + iyShift
			sub = canvas.SubImage(bounds).(*ebiten.Image)
		}

		anim := float64(ctx.Ticks) / 60.0
		ctx.Renderer.NoiseGolden(sub, scale, 1.0, float32(anim))
	})

	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}
