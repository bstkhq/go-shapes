package shapes

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// go test -run ^TestGradient$ . -count 1
func TestGradient(t *testing.T) {
	// to appreciate dithering, resize the window or press F11 to go fullscreen
	var steps float64
	var dither bool
	var bias float64 = 0.0
	app := NewTestApp(func(canvas *ebiten.Image, ctx TestAppCtx) {
		ebiten.SetWindowTitle(ctx.Title() + fmt.Sprintf(" [[S]teps: %d, [D]ither: %t, [B]ias: %+.02f]", int(steps), dither, bias))
		canvas.Fill(color.Black)

		bias = updateParam(ctx, ebiten.KeyB, bias, -1.0, 1.0, 0.05)
		steps = updateParam(ctx, ebiten.KeyS, steps, 0, 8.0, 1)
		if ctx.NewInput && inpututil.IsKeyJustPressed(ebiten.KeyD) {
			dither = !dither
		}

		_, _, cw, ch := rectOriginSize(canvas.Bounds())
		cwF32, chF32 := float32(cw), float32(ch)
		ctx.DrawAtF32(canvas, ctx.Images[0], 16, 16)
		ctx.DrawAtF32(canvas, ctx.Images[1], cwF32-16-float32(ctx.Images[1].Bounds().Dx()), 16)
		ctx.DrawAtF32(canvas, ctx.Images[2], 16, chF32-16-float32(ctx.Images[2].Bounds().Dy()))

		sub := canvas.SubImage(image.Rect(cw-16-80, ch-16-60, cw-16, ch-16)).(*ebiten.Image)
		opts := GradientOpts(color.RGBA{0, 255, 0, 255}, color.RGBA{0, 0, 255, 255}, false)
		ctx.Renderer.Gradient(sub, opts, DirRadsBRTL)

		halfsize := min(cw, ch) * 2 / 6
		sub = canvas.SubImage(image.Rect(cw/2-halfsize, ch/2-halfsize, cw/2+halfsize, ch/2+halfsize)).(*ebiten.Image)
		opts = GradientOpts(color.RGBA{0, 25, 50, 96}, color.RGBA{75, 50, 75, 96}, dither)
		opts.Steps = int(steps)
		opts.Bias = float32(bias)
		ctx.Renderer.Gradient(sub, opts, DirRadsTLBR)
	})

	rect := app.Renderer.NewRect(120, 80)
	square := app.Renderer.NewRect(64, 64)
	circ := app.Renderer.NewCircle(64.0)
	app.Renderer.Options().Blend = ebiten.BlendSourceIn

	opts := StepGradientOpts(color.RGBA{0, 0, 255, 255}, color.RGBA{0, 255, 0, 255}, 4)
	app.Renderer.Gradient(rect, opts, math.Pi/7)

	opts = GradientOpts(color.RGBA{0, 0, 255, 255}, color.RGBA{0, 255, 0, 255}, false)
	opts.Bias = -0.5
	app.Renderer.Gradient(circ, opts, DirRadsRTL)

	opts = StepGradientOpts(color.RGBA{0, 255, 255, 255}, color.RGBA{255, 255, 0, 255}, 2)
	app.Renderer.Gradient(square, opts, DirRadsBLTR)

	app.Renderer.Options().Blend = ebiten.BlendSourceOver
	app.Images = append(app.Images, rect, circ, square)

	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestGradientRadial$ . -count 1
func TestGradientRadial(t *testing.T) {
	// to appreciate dithering, resize the window or press F11 to go fullscreen
	var steps float64
	var dither bool
	var bias float64 = 0.0
	app := NewTestApp(func(canvas *ebiten.Image, ctx TestAppCtx) {
		ebiten.SetWindowTitle(ctx.Title() + fmt.Sprintf(" [[S]teps: %d, [D]ither: %t, [B]ias: %+.02f]", int(steps), dither, bias))
		canvas.Fill(color.Black)

		bias = updateParam(ctx, ebiten.KeyB, bias, -1.0, 1.0, 0.05)
		steps = updateParam(ctx, ebiten.KeyS, steps, 0, 8.0, 1)
		if ctx.NewInput && inpututil.IsKeyJustPressed(ebiten.KeyD) {
			dither = !dither
		}

		lx, ly := ctx.LeftClickF32()
		ctx.DrawAtF32(canvas, ctx.Images[0], lx-64, ly-64)

		_, _, cw, ch := rectOriginSize(canvas.Bounds())
		halfsize := min(cw, ch) * 2 / 6
		sub := canvas.SubImage(image.Rect(cw/2-halfsize, ch/2-halfsize, cw/2+halfsize, ch/2+halfsize)).(*ebiten.Image)
		opts := GradientOpts(color.RGBA{0, 25, 50, 96}, color.RGBA{75, 50, 75, 96}, dither)
		opts.Steps = int(steps)
		opts.Bias = float32(bias)
		hsF32 := float32(halfsize)
		ctx.Renderer.GradientRadial(sub, opts, hsF32, hsF32, 24, hsF32-48, hsF32)
	})

	circ := app.Renderer.NewCircle(48.0)
	opts := GradientOpts(color.RGBA{255, 255, 0, 255}, color.RGBA{255, 0, 255, 255}, true)
	opts.Bias = +0.5
	app.Renderer.Options().Blend = ebiten.BlendSourceIn
	app.Renderer.GradientRadial(circ, opts, 48, 48, 0.0, 48.0, Float32Inf())
	app.Renderer.Options().Blend = ebiten.BlendSourceOver

	app.Images = append(app.Images, circ)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}
