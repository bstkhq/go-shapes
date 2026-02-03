package shapes

import (
	"image"
	"image/color"
	"math"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// go test -run ^TestFlatPaint$ . -count 1
func TestFlatPaint(t *testing.T) {
	app := NewTestApp(func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Fill(color.Black)

		lx, ly := ctx.LeftClickF32()
		ctx.Renderer.SetColorF32(1.0, 0.0, 0.0, 1.0, 0, 1)
		ctx.Renderer.SetColorF32(1.0, 0.0, 1.0, 1.0, 2, 3)
		ctx.Renderer.FlatPaint(canvas, ctx.Images[0], lx, ly)

		rx, ry := ctx.RightClickF32()
		ctx.Renderer.SetColorF32(0.0, 1.0, 0.0, 1.0, 0, 3)
		ctx.Renderer.SetColorF32(0.0, 1.0, 1.0, 1.0, 1, 2)
		ctx.Renderer.FlatPaint(canvas, ctx.Images[1], rx, ry)
	})

	rect := app.Renderer.NewRect(120, 80)
	circ := app.Renderer.NewCircle(64.0)
	app.Renderer.Options().Blend = ebiten.BlendDestinationOut
	app.Renderer.SetColorF32(0.8, 0.8, 0.8, 0.8, 0, 1)
	app.Renderer.SetColorF32(0.3, 0.3, 0.3, 0.3, 2, 3)
	app.Renderer.DrawCircle(circ, 64.0, 64.0, 42.0)
	app.Renderer.Options().Blend = ebiten.BlendSourceOver
	app.Images = append(app.Images, rect, circ)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestGradient$ . -count 1
func TestGradient(t *testing.T) {
	app := NewTestApp(func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Fill(color.Black)

		lx, ly := ctx.LeftClickF32()
		ctx.DrawAtF32(canvas, ctx.Images[0], lx, ly)

		rx, ry := ctx.RightClickF32()
		ctx.DrawAtF32(canvas, ctx.Images[1], rx, ry)

		ox, oy := 50, 400
		sub := canvas.SubImage(image.Rect(ox, oy, ox+80, oy+60)).(*ebiten.Image)
		ctx.Renderer.SimpleGradient(sub, color.RGBA{0, 255, 0, 255}, color.RGBA{0, 0, 255, 255}, DirRadsBRTL)
	})

	rectA := app.Renderer.NewRect(120, 80)
	circ := app.Renderer.NewCircle(64.0)
	rectB := ebiten.NewImageWithOptions(circ.Bounds(), nil)
	app.Renderer.Gradient(rectA, nil, 0, 0, color.RGBA{0, 0, 255, 255}, color.RGBA{0, 255, 0, 255}, 4, math.Pi/7, 1.0)
	app.Renderer.Gradient(rectB, circ, 0, 0, color.RGBA{0, 0, 255, 255}, color.RGBA{0, 255, 0, 255}, -1, math.Pi, 0.2)
	app.Images = append(app.Images, rectA, rectB)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestGradientDither$ . -count 1
func TestGradientDither(t *testing.T) {
	dirs := []float32{DirRadsRTL, DirRadsBLTR, DirRadsTTB, DirRadsTLBR}
	dirIndex := 0
	app := NewTestApp(func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Fill(color.Black)
		if ctx.NewInput && inpututil.IsKeyJustPressed(ebiten.KeySpace) {
			dirIndex = (dirIndex + 1) % len(dirs)
		}

		cw, ch := rectSizeF32(canvas.Bounds())
		from, to := color.RGBA{25, 25, 52, 255}, color.RGBA{50, 50, 50, 255}
		dir := dirs[dirIndex]
		ctx.Renderer.GradientDither(canvas, 0, 0, cw, ch/2, from, to, dir, 1.0)
		sub := canvas.SubImage(image.Rect(0, int(ch/2), int(cw), int(ch))).(*ebiten.Image)
		ctx.Renderer.Gradient(sub, nil, 0, 0, from, to, -1, dir, 1.0)
	})

	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestGradientRadial$ . -count 1
func TestGradientRadial(t *testing.T) {
	app := NewTestApp(func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Fill(color.Black)

		lx, ly := ctx.LeftClickF32()
		rx, ry := ctx.RightClickF32()
		ctx.DrawAtF32(canvas, ctx.Images[0], lx-64, ly-64)

		curveShift := float32(ctx.DistAnim(2.0, 2.0))
		ctx.Renderer.GradientRadial(canvas, lx, ly, color.RGBA{0, 255, 0, 255}, color.RGBA{0, 0, 255, 255}, 16.0, 64.0, 64.0, -1, 1.0+curveShift)
		steps := int(math.Floor(ctx.DistAnim(8.0, 1.0)))
		inner := float32(ctx.DistAnim(32.0, 1.0))
		ctx.Renderer.GradientRadial(canvas, rx, ry, color.RGBA{0, 255, 255, 255}, color.RGBA{255, 0, 255, 255}, inner, 96.0, 128.0, steps, 1.0)

		ctx.DrawAtF32(canvas, ctx.Images[1], 120, 320)
	})

	circ := app.Renderer.NewCircle(64.0)
	circ2 := app.Renderer.NewCircle(48.0)

	app.Renderer.SetBlend(ebiten.BlendSourceIn)
	app.Renderer.GradientRadial(circ2, 48, 48, color.RGBA{255, 255, 0, 255}, color.RGBA{255, 0, 255, 255}, 0.0, 48.0, Float32Inf(), -1, 2.5)
	app.Renderer.SetBlend(ebiten.BlendSourceOver)

	app.Images = append(app.Images, circ, circ2)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestGradientRadialDither$ . -count 1
func TestGradientRadialDither(t *testing.T) {
	app := NewTestApp(func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Fill(color.Black)

		w, h := rectSizeF32(canvas.Bounds())
		curveShift := float32(ctx.DistAnim(2.0, 2.0))
		from, to := color.RGBA{25, 25, 50, 96}, color.RGBA{50, 50, 50, 96}
		if ebiten.IsKeyPressed(ebiten.KeySpace) {
			ctx.Renderer.GradientRadial(canvas, w/2, h/2, from, to, 0.0, 320.0, 320.0, -1, 1.0+curveShift)
		} else {
			ctx.Renderer.GradientRadialDither(canvas, w/2, h/2, from, to, 0.0, 320.0, 320.0, 1.0+curveShift)
		}
	})

	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestColorizeByLightness$ . -count 1
func TestColorizeByLightness(t *testing.T) {
	app := NewTestApp(func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Fill(color.Black)

		cw, ch := rectSizeF32(canvas.Bounds())
		iw, ih := rectSizeF32(ctx.Images[0].Bounds())
		x, y := (cw-iw)/2.0, (ch-ih)/2.0
		from, to := color.RGBA{0, 196, 196, 255}, color.RGBA{255, 0, 255, 255}
		curveFactor := float32(0.75 + ctx.DistAnim(1.75, 1.0))
		if ebiten.IsKeyPressed(ebiten.KeySpace) {
			ctx.DrawAtF32(canvas, ctx.Images[0], x, y)
		} else {
			var thresA, thresB float32 = 0.0, 1.0
			ctx.Renderer.ColorizeByLightness(canvas, ctx.Images[0], x, y, from, to, thresA, thresB, 7, curveFactor)
		}
	})

	const Radius = 48
	base := app.Renderer.NewSimpleGradient(Radius*8, Radius*8, color.RGBA{0, 0, 0, 255}, color.RGBA{255, 255, 255, 255}, DirRadsTLBR)
	app.Renderer.Noise(base, 0.1, 26.26, 0.0)
	app.Renderer.SetColorF32(1.0, 0.0, 1.0, 1.0, 0, 3)
	app.Renderer.SetColorF32(0.5, 0.0, 0.0, 1.0, 1, 2)
	app.Renderer.DrawCircle(base, Radius*2, Radius*4, Radius)
	app.Renderer.SetColorF32(0.0, 1.0, 1.0, 1.0, 0, 3)
	app.Renderer.SetColorF32(0.5, 0.0, 0.0, 1.0, 1, 2)
	app.Renderer.DrawCircle(base, Radius*6, Radius*4, Radius)
	app.Renderer.SetColorF32(1.0, 1.0, 0.0, 1.0, 0, 3)
	app.Renderer.SetColorF32(0.5, 0.0, 0.0, 1.0, 1, 2)
	app.Renderer.DrawCircle(base, Radius*4, Radius*2, Radius)
	app.Renderer.SetColorF32(0.0, 1.0, 0.0, 1.0, 0, 3)
	app.Renderer.SetColorF32(0.5, 0.0, 0.0, 1.0, 1, 2)
	app.Renderer.DrawCircle(base, Radius*4, Radius*6, Radius)

	app.Renderer.SetColorF32(1.0, 1.0, 1.0, 1.0)
	app.Images = append(app.Images, base)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestOklabShift$ . -count 1
func TestOklabShift(t *testing.T) {
	app := NewTestApp(func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Fill(color.Black)

		lx, ly := ctx.LeftClickF32()
		chromaShift := float32(0.3 - ctx.DistAnim(0.6, 0.4))
		lightnessShift := float32(0.5 - ctx.DistAnim(1.0, 1.0))
		hueShift := float32(ctx.ModAnim(2*math.Pi, 0.5))
		ctx.Renderer.OklabShift(canvas, ctx.Images[0], lx, ly, lightnessShift, chromaShift, hueShift)

		rx, ry := ctx.RightClickF32()
		ctx.DrawAtF32(canvas, ctx.Images[0], rx, ry)
	})

	app.Renderer.SetColorF32(0.8, 0.5, 0.0, 1.0)
	circ := app.Renderer.NewCircle(64.0)
	app.Images = append(app.Images, circ)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestDitherMat4$ . -count 1
func TestDitherMat4(t *testing.T) {
	app := NewTestApp(func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Fill(color.RGBA{128, 0, 128, 255})

		var mat [16]float32
		switch int(ctx.ModAnim(5.0, 0.25)) {
		case 0:
			mat = DitherBayes
		case 1:
			mat = DitherDots
		case 2:
			mat = DitherGlitch
		case 3:
			mat = DitherSerp
		case 4:
			mat = DitherCrumbs
		}
		mat = combineDitherMat4(mat, DitherBayes)

		lx, ly := ctx.LeftClickF32()
		ctx.Renderer.SetColorF32(1.0, 0.0, 0.0, 1.0, 0, 1)
		ctx.Renderer.SetColorF32(1.0, 0.0, 1.0, 1.0, 2, 3)
		anim := float32(ctx.DistAnim(1.0, 1.0))
		yOffset := int(ctx.ModAnim(4.0, 1.0))
		xOffset := 8 - int(ctx.DistAnim(16.0, 1.0))
		ctx.Renderer.DitherMat4(canvas, ctx.Images[0], lx, ly, xOffset, yOffset, DitherBW4, mat, anim, 0.0)

		rx, ry := ctx.RightClickF32()
		ctx.Renderer.DitherMat4(canvas, ctx.Images[1], rx, ry, 0, 0, DitherAlpha8, mat, 0, anim)
	})

	from, to := color.RGBA{0, 0, 0, 255}, color.RGBA{255, 255, 255, 255}
	gradient := app.Renderer.NewSimpleGradient(160, 160, from, to, DirRadsLTR)
	app.Images = append(app.Images, gradient)
	from, to = color.RGBA{0, 0, 0, 255}, color.RGBA{255, 255, 255, 255}
	gradient = app.Renderer.NewSimpleGradient(160, 160, from, to, DirRadsTLBR)
	app.Images = append(app.Images, gradient)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

func combineDitherMat4(a, b [16]float32) [16]float32 {
	var out [16]float32
	for i := range 16 {
		out[i] = (a[i] + b[i]) / 2.0
	}
	return out
}

// go test -run ^TestColorMix$ . -count 1
func TestColorMix(t *testing.T) {
	app := NewTestApp(func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Fill(color.Black)
		lvl := float32(ctx.DistAnim(1.0, 1.0))
		ctx.Renderer.ColorMix(canvas, ctx.Images[0], ctx.Images[1], ctx.LeftClick.X, ctx.LeftClick.Y, 0.5, lvl)
	})

	circ := app.Renderer.NewCircle(64.0)
	app.Renderer.SetColorF32(1.0, 0, 1.0, 1.0)
	circ2 := app.Renderer.NewCircle(64.0)
	app.Images = append(app.Images, circ, circ2)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}
