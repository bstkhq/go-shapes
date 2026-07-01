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

// go test -run ^TestFlatPaint$ . -count 1
func TestFlatPaint(t *testing.T) {
	// NOTICE: FlatPaint was removed in favor of DrawAt
	updater := func(TestAppCtx) {}
	drawer := func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Fill(color.Black)

		lc := ctx.LeftClickF32()
		ctx.Renderer.SetTint(1)
		ctx.Renderer.SetColorF32(1.0, 0.0, 0.0, 1.0, 0, 1)
		ctx.Renderer.SetColorF32(1.0, 0.0, 1.0, 1.0, 2, 3)
		lc = CTR.Adjust(ctx.Images[0], lc)
		ctx.Renderer.DrawAt(canvas, ctx.Images[0], lc.X, lc.Y, 1.0)

		rc := ctx.RightClickF32()
		ctx.Renderer.SetColorF32(0.0, 1.0, 0.0, 1.0, 0, 3)
		ctx.Renderer.SetColorF32(0.0, 1.0, 1.0, 1.0, 1, 2)
		rc = CTR.Adjust(ctx.Images[1], rc)
		ctx.Renderer.DrawAt(canvas, ctx.Images[1], rc.X, rc.Y, 1.0)
		ctx.Renderer.SetTint(0)
	}

	app := NewTestApp(updater, drawer)
	rect := app.Renderer.NewFilledRect(120, 80)
	circ := app.Renderer.NewFilledCircle(64.0)
	app.Renderer.Options().Blend = ebiten.BlendDestinationOut
	app.Renderer.SetColorF32(0.8, 0.8, 0.8, 0.8, 0, 1)
	app.Renderer.SetColorF32(0.3, 0.3, 0.3, 0.3, 2, 3)
	app.Renderer.FillCircle(circ, 64.0, 64.0, 42.0)
	app.Renderer.Options().Blend = ebiten.BlendSourceOver
	app.Images = append(app.Images, rect, circ)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestColorizeByLightness$ . -count 1
func TestColorizeByLightness(t *testing.T) {
	bias := float32(0.0)
	steps := 7
	dither := false
	var fromThresh, toThresh float32 = 0.0, 1.0

	updater := func(ctx TestAppCtx) {
		ebiten.SetWindowTitle(fmt.Sprintf(
			"%s [[S]teps: %d, [B]ias: %.02f, [D]ither: %t, [F]romL: %.02f, [T]oL: %.02f]",
			ctx.Title(), steps, bias, dither, fromThresh, toThresh,
		))
		steps = updateParam(ctx, ebiten.KeyS, steps, 0, 12, 1)
		fromThresh = updateParam(ctx, ebiten.KeyF, fromThresh, 0.0, 1.0, 0.1)
		toThresh = updateParam(ctx, ebiten.KeyT, toThresh, 0.0, 1.0, 0.1)
		bias = updateParam(ctx, ebiten.KeyB, bias, -1.0, 1.0, 0.1)
		if inpututil.IsKeyJustPressed(ebiten.KeyD) {
			dither = !dither
		}
	}
	drawer := func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Fill(color.Black)

		cw, ch := rectSizeF32(canvas.Bounds())
		iw, ih := rectSizeF32(ctx.Images[0].Bounds())
		x, y := (cw-iw)/2.0, (ch-ih)/2.0

		opts := GradientOpts(color.RGBA{0, 196, 196, 255}, color.RGBA{255, 0, 255, 255}, dither)
		opts.Steps = steps
		opts.Bias = bias

		if ctx.SpacePressed {
			ctx.DrawAtF32(canvas, ctx.Images[0], x, y)
		} else {
			ctx.Renderer.ColorizeByLightness(canvas, ctx.Images[0], opts, x, y, fromThresh, toThresh)
		}
	}
	app := NewTestApp(updater, drawer)

	const Radius = 48
	gradientOpts := GradientOpts(color.RGBA{0, 0, 0, 255}, color.RGBA{255, 255, 255, 255}, false)
	base := ebiten.NewImage(Radius*8, Radius*8)
	app.Renderer.Gradient(base, gradientOpts, DirRadsTLBR)
	app.Renderer.Noise(base, 0.1, 0.26, 0.0)
	app.Renderer.SetColorF32(1.0, 0.0, 1.0, 1.0, 0, 3)
	app.Renderer.SetColorF32(0.5, 0.0, 0.0, 1.0, 1, 2)
	app.Renderer.FillCircle(base, Radius*2, Radius*4, Radius)
	app.Renderer.SetColorF32(0.0, 1.0, 1.0, 1.0, 0, 3)
	app.Renderer.SetColorF32(0.5, 0.0, 0.0, 1.0, 1, 2)
	app.Renderer.FillCircle(base, Radius*6, Radius*4, Radius)
	app.Renderer.SetColorF32(1.0, 1.0, 0.0, 1.0, 0, 3)
	app.Renderer.SetColorF32(0.5, 0.0, 0.0, 1.0, 1, 2)
	app.Renderer.FillCircle(base, Radius*4, Radius*2, Radius)
	app.Renderer.SetColorF32(0.0, 1.0, 0.0, 1.0, 0, 3)
	app.Renderer.SetColorF32(0.5, 0.0, 0.0, 1.0, 1, 2)
	app.Renderer.FillCircle(base, Radius*4, Radius*6, Radius)

	app.Renderer.SetColorF32(1.0, 1.0, 1.0, 1.0)
	app.Images = append(app.Images, base)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestOklabShift$ . -count 1
func TestOklabShift(t *testing.T) {
	updater := func(TestAppCtx) {}
	drawer := func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Fill(color.Black)

		lc := ctx.LeftClickF32()
		chromaShift := float32(0.3 - ctx.DistAnim(0.6, 0.4))
		lightnessShift := float32(0.5 - ctx.DistAnim(1.0, 1.0))
		hueShift := float32(ctx.ModAnim(2*math.Pi, 0.5))

		lc = CTR.Adjust(ctx.Images[0], lc)
		ctx.Renderer.OklabShift(canvas, ctx.Images[0], lc.X, lc.Y, lightnessShift, chromaShift, hueShift)

		rc := ctx.RightClickF32()
		rc = CTR.Adjust(ctx.Images[0], rc)
		ctx.DrawAtF32(canvas, ctx.Images[0], rc.X, rc.Y)
	}

	app := NewTestApp(updater, drawer)
	app.Renderer.SetColorF32(0.8, 0.5, 0.0, 1.0)
	circ := app.Renderer.NewFilledCircle(64.0)
	app.Images = append(app.Images, circ)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestDitherMat4$ . -count 1
func TestDitherMat4(t *testing.T) {
	updater := func(TestAppCtx) {}
	drawer := func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Fill(color.RGBA{128, 0, 128, 255})

		mats := [][16]float32{DitherBayes, DitherDots, DitherGlitch, DitherSerp, DitherCrumbs}
		mat := mats[int(ctx.ModAnim(float64(len(mats)), 0.25))]

		if ctx.SpacePressed {
			mat = combineDitherMat4(mat, DitherBayes)
		}

		lc := ctx.LeftClickF32()
		ctx.Renderer.SetColorF32(1.0, 0.0, 0.0, 1.0, 0, 1)
		ctx.Renderer.SetColorF32(1.0, 0.0, 1.0, 1.0, 2, 3)
		anim := float32(ctx.DistAnim(1.0, 1.0))
		yOffset := int(ctx.ModAnim(4.0, 1.0))
		xOffset := 8 - int(ctx.DistAnim(16.0, 1.0))
		lc = CTR.Adjust(ctx.Images[0], lc)
		ctx.Renderer.DitherMat4(canvas, ctx.Images[0], lc.X, lc.Y, xOffset, yOffset, PaletteBW4, mat, anim, 0.0)

		rc := ctx.RightClickF32()
		rc = CTR.Adjust(ctx.Images[0], rc)
		ctx.Renderer.DitherMat4(canvas, ctx.Images[1], rc.X, rc.Y, 0, 0, PaletteAlpha8, mat, 0.0, anim)
	}

	app := NewTestApp(updater, drawer)
	gradient := ebiten.NewImage(160, 160)
	gradientOpts := GradientOpts(color.RGBA{0, 0, 0, 255}, color.RGBA{255, 255, 255, 255}, false)
	app.Renderer.Gradient(gradient, gradientOpts, DirRadsLTR)
	app.Images = append(app.Images, gradient)

	gradient = ebiten.NewImage(160, 160)
	app.Renderer.Gradient(gradient, gradientOpts, DirRadsTLBR)
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
	var flags flagList
	updater := func(ctx TestAppCtx) {
		setMaskFlagsAndTitle(ctx, flags)
	}
	drawer := func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Fill(color.Black)
		lvl := float32(ctx.DistAnim(1.0, 1.0))
		lc := ctx.LeftClickF32()
		lc = CTR.Adjust(ctx.Images[0], lc)
		ctx.Renderer.ColorMix(canvas, ctx.Images[0], ctx.Images[1], lc.X, lc.Y, 0.5, lvl, flags...)

		rc := ctx.RightClickF32()
		rc = CTR.Adjust(ctx.Images[0], rc)
		alpha := float32(ctx.DistAnim(1.0, 0.333))
		offX := float32(-8.0 + ctx.DistAnim(16.0, 1.0))
		ctx.Renderer.ColorMix(canvas, ctx.Images[1], ctx.Images[0], rc.X+offX, rc.Y, alpha, lvl, flags...)
	}

	app := NewTestApp(updater, drawer)
	circ := app.Renderer.NewFilledCircle(64.0)
	app.Renderer.SetColorF32(1.0, 0, 1.0, 1.0)
	circ2 := ebiten.NewImage(128+16, 128+16)
	app.Renderer.FillCircle(circ2, 64+16, 64+16, 64)
	circ2 = circ2.SubImage(image.Rect(16, 16, 16+128, 16+128)).(*ebiten.Image)
	app.Images = append(app.Images, circ, circ2)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}
