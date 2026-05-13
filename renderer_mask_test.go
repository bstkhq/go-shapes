package shapes

import (
	"fmt"
	"image"
	"image/color"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

func setMaskFlagsAndTitle(ctx TestAppCtx, flags flagList) {
	ebiten.SetWindowTitle(fmt.Sprintf("%s [[B]ilinear: %t, [D]ither: %t]", ctx.Title(), flags.Has(Bilinear), flags.Has(Dithered)))
	flags.UpdateFlag(Bilinear, ebiten.KeyB)
	flags.UpdateFlag(Dithered, ebiten.KeyD)
}

// go test -run ^TestMask$ . -count 1
func TestMask(t *testing.T) {
	flags := newFlagList()
	updater := func(ctx TestAppCtx) { setMaskFlagsAndTitle(ctx, flags) }
	drawer := func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Fill(color.Black)

		// dither check strip
		_, ch := rectSizeF32(canvas.Bounds())
		ctx.Renderer.Mask(canvas, ctx.Images[4], ctx.Images[3], 16, ch-16-float32(ctx.Images[3].Bounds().Dy()), flags...)

		// circle with slightly movement for bilinear testing
		lx, ly := ctx.LeftClickF32()
		ly += float32(-4 + ctx.DistAnim(8, 1.0))
		ctx.Renderer.Mask(canvas, ctx.Images[0], ctx.Images[1], lx, ly, flags...)

		// small rect mask scaling and dst space bilinear test
		rx, ry := ctx.RightClickF32()
		ctx.Renderer.Mask(canvas, ctx.Images[0], ctx.Images[2], rx, ry, flags...)
	}

	app := NewTestApp(updater, drawer)
	app.Renderer.SetColorF32(0.5, 0.5, 0.5, 1.0)
	circ := app.Renderer.NewFilledCircle(72.0)
	app.Renderer.SetColorF32(1, 1, 1, 1)
	app.Renderer.SetColorF32(0, 0, 0, 0, 1, 2) // right side to zero
	bigRect := app.Renderer.NewFilledRect(256, 128)
	smallRect := app.Renderer.NewFilledRect(16, 8) // being small creates an step effect automatically

	longRect := app.Renderer.NewFilledRect(640, 128)
	app.Renderer.SetColorF32(0.0, 0.2, 0.2, 1)
	longImg := app.Renderer.NewFilledRect(640, 128)

	app.Renderer.SetColorF32(1, 1, 1, 1)
	app.Images = append(app.Images, circ, bigRect, smallRect, longRect, longImg)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestMaskAt$ . -count 1
func TestMaskAt(t *testing.T) {
	flags := newFlagList()
	updater := func(ctx TestAppCtx) { setMaskFlagsAndTitle(ctx, flags) }
	drawer := func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Fill(color.Black)
		lx, ly := ctx.LeftClickF32()
		dist := float32(ctx.DistAnim(256.0, 1.0))
		ctx.Renderer.MaskAt(canvas, ctx.Images[0], ctx.Images[1], lx+dist, ly, lx, ly, flags...)

		_, ch := rectSizeF32(canvas.Bounds())
		x, y := float32(16.0), ch-16-float32(ctx.Images[3].Bounds().Dy())
		maskSlide := float32(ctx.DistAnim(32.0, 1.0))
		ctx.Renderer.MaskAt(canvas, ctx.Images[2], ctx.Images[3], x+32, y, x+maskSlide, y, flags...)
	}

	app := NewTestApp(updater, drawer)
	circ := app.Renderer.NewFilledCircle(32.0)
	app.Renderer.SetColorF32(0, 0, 0, 0, 1, 2)
	trans := app.Renderer.NewFilledRect(256, 64)
	longRect := app.Renderer.NewFilledRect(640, 128)
	app.Renderer.SetColorF32(0.0, 0.2, 0.2, 1)
	longImg := app.Renderer.NewFilledRect(640-64, 128)
	app.Renderer.SetColorF32(1, 1, 1, 1)
	app.Images = append(app.Images, circ, trans, longImg, longRect)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestMaskHorz$ . -count 1
func TestMaskHorz(t *testing.T) {
	flags := newFlagList()
	updater := func(ctx TestAppCtx) { setMaskFlagsAndTitle(ctx, flags) }
	drawer := func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Fill(color.Black)
		lx, ly := ctx.LeftClickF32()
		x, _ := ebiten.CursorPosition()
		yShift := float32(-4.0 + ctx.DistAnim(8.0, 1.0))
		ctx.Renderer.MaskHorz(canvas, ctx.Images[0], lx, ly+yShift, lx+256/2.0, float32(x), flags...)

		cw, ch := rectSizeF32(canvas.Bounds())
		xm, ym := CTR.Adjust(ctx.Images[1], cw/2, ch*2/3)
		i1w := float32(ctx.Images[1].Bounds().Dx())
		ctx.Renderer.MaskHorz(canvas, ctx.Images[1], xm, ym, xm, xm+i1w, flags...)
	}

	app := NewTestApp(updater, drawer)
	rect := app.Renderer.NewFilledRect(256, 64)
	app.Renderer.SetColorF32(0.2, 0.2, 0.2, 1.0)
	rectDither := app.Renderer.NewFilledRect(640, 96)
	app.Renderer.SetColorF32(1, 1, 1, 1)
	app.Images = append(app.Images, rect, rectDither)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestMaskCirc$ . -count 1
func TestMaskCirc(t *testing.T) {
	flags := newFlagList()
	updater := func(ctx TestAppCtx) { setMaskFlagsAndTitle(ctx, flags) }
	drawer := func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Fill(color.Black)
		w, h := rectSizeF32(canvas.Bounds())
		hardRadius := 48.0 + float32(ctx.DistAnim(16.0, 0.25))
		softEdge := float32(ctx.DistAnim(16.0, 1.0))
		ox, oy := CTR.Adjust(ctx.Images[0], w/2, h/2)
		ctx.Renderer.MaskCirc(canvas, ctx.Images[0], ox, oy, w/2, h/2, hardRadius, softEdge, flags...)

		lx, ly := ctx.LeftClickF32()
		ox, oy = CTR.Adjust(ctx.Images[0], lx, ly)
		circleYShift := float32(-4.0 + ctx.DistAnim(8.0, 1.0))
		ctx.Renderer.MaskCirc(canvas, ctx.Images[0], ox, oy, lx, ly+circleYShift, 32, 16, flags...)
	}

	app := NewTestApp(updater, drawer)
	rect := app.Renderer.NewFilledRect(360, 360)
	app.Images = append(app.Images, rect)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestMaskThreshold$ . -count 1
func TestMaskThreshold(t *testing.T) {
	const Size = 256

	flags := newFlagList()
	updater := func(ctx TestAppCtx) { setMaskFlagsAndTitle(ctx, flags) }
	drawer := func(canvas *ebiten.Image, ctx TestAppCtx) {
		reveal := -0.1 + float32(ctx.ModAnim(1.2, 0.5))
		// TODO: Text(canvas, fmt.Sprintf("Reveal: %.02f", reveal))

		canvas.Fill(color.Black)
		w, h := rectSizeF32(canvas.Bounds())
		ox, oy := w/2-Size/2, h/2-Size/2
		oxi, oyi := int(ox), int(oy)
		Paint(canvas, image.Rect(oxi, oyi, oxi+Size, oyi+Size), [4]float32{0.5, 0, 0, 0.5}, ebiten.BlendSourceOver)
		oy += float32(-4.0 + ctx.DistAnim(8.0, 1.0))
		ctx.Renderer.MaskThreshold(canvas, ctx.Images[0], ctx.Images[1], reveal, ox, oy)
	}

	app := NewTestApp(updater, drawer)
	maskTarget := ebiten.NewImage(Size, Size)
	gradientOpts := StepGradientOpts(color.RGBA{0, 0, 0, 0}, color.RGBA{255, 255, 255, 255}, 16)
	app.Renderer.Gradient(maskTarget, gradientOpts, DirRadsLTR)
	whiteRect := app.Renderer.NewFilledRect(Size, Size)
	app.Images = append(app.Images, whiteRect, maskTarget)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestBakeAlphaMaskRadial$ . -count 1
func TestBakeAlphaMaskRadial(t *testing.T) {
	const Size = 256
	randomness := float32(0.3)
	var justClicked bool

	flags := newFlagList()
	updater := func(ctx TestAppCtx) {
		justClicked = false
		setMaskFlagsAndTitle(ctx, flags)
		switch {
		case inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft):
			justClicked = true
		case inpututil.IsKeyJustPressed(ebiten.KeyArrowUp):
			randomness = min(randomness+0.1, 1.0)
		case inpututil.IsKeyJustPressed(ebiten.KeyArrowDown):
			randomness = max(randomness-0.1, 0.0)
		}
		ebiten.SetWindowTitle(ctx.Title() + fmt.Sprintf(" [randomness %.02f]", randomness))
	}

	drawer := func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Fill(color.Black)

		w, h := rectSizeF32(canvas.Bounds())
		ox, oy := w/2-Size/2, h/2-Size/2
		if justClicked {
			lx, ly := ctx.LeftClickF32()
			ctx.Renderer.Options().Blend = ebiten.BlendCopy
			ctx.Renderer.BakeAlphaMaskRadial(ctx.Images[1], lx-ox, ly-oy, Size*1.44, randomness, MaskPatternEllipseCuts)
			ctx.Renderer.Options().Blend = ebiten.BlendSourceOver
		}

		reveal := -0.1 + float32(ctx.ModAnim(2.0, 0.2))
		ctx.Renderer.MaskThreshold(canvas, ctx.Images[0], ctx.Images[1], reveal, ox, oy, flags...)
	}

	app := NewTestApp(updater, drawer)
	maskTarget := ebiten.NewImage(Size, Size)
	whiteRect := app.Renderer.NewFilledRect(Size, Size)
	app.Renderer.BakeAlphaMaskRadial(maskTarget, Size/2, Size/2, Size, randomness, MaskPatternDefault)
	app.Images = append(app.Images, whiteRect, maskTarget)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}
