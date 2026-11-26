//go:build interactive

package shapes

import (
	"fmt"
	"image"
	"image/color"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

func setMaskFlagsAndTitle(ctx TestAppCtx, flags *flagList) {
	ebiten.SetWindowTitle(fmt.Sprintf("%s [[B]ilinear: %t, [D]ither: %t]", ctx.Title(), flags.Has(Bilinear), flags.Has(Dithered)))
	flags.UpdateFlag(Bilinear, ebiten.KeyB)
	flags.UpdateFlag(Dithered, ebiten.KeyD)
}

// go test -tags interactive -run ^TestMask$ . -count 1
func TestMask(t *testing.T) {
	var flags flagList
	smoothMovement := false
	updater := func(ctx TestAppCtx) {
		flags.UpdateFlag(Bilinear, ebiten.KeyB)
		flags.UpdateFlag(Dithered, ebiten.KeyD)
		smoothMovement = updateToggle(ctx, ebiten.KeyM, smoothMovement)
	}
	drawer := func(canvas *ebiten.Image, ctx TestAppCtx) {
		circleSource := ctx.Images[0]
		hiResGradientMask, loResGradientMask := ctx.Images[1], ctx.Images[2]
		hiResLongGradientMask, stripSource := ctx.Images[3], ctx.Images[4]

		canvas.Fill(backTestColor)
		if ctx.SpacePressed {
			ctx.Renderer.Options().Blend = ebiten.BlendCopy
		}
		xShift, yShift := float32(0), float32(0)
		if smoothMovement {
			xShift = float32(-4.0 + ctx.DistAnim(8.0, 1.3))
			yShift = float32(-4.0 + ctx.DistAnim(8.0, 1.0))
		}

		// dither check strip
		_, ch := rectSizeF32(canvas.Bounds())
		ctx.Renderer.Mask(canvas, stripSource, hiResLongGradientMask, 16, ch-16-float32(hiResLongGradientMask.Bounds().Dy()), flags...)

		lc := ctx.LeftClickF32()
		lc = lc.AddXY(xShift, yShift)
		ctx.Renderer.Mask(canvas, circleSource, hiResGradientMask, lc.X, lc.Y, flags...)

		rc := ctx.RightClickF32()
		rc = rc.AddXY(xShift, yShift)
		ctx.Renderer.Mask(canvas, circleSource, loResGradientMask, rc.X, rc.Y, flags...)
		ctx.Renderer.Options().Blend = ebiten.BlendSourceOver

		ctx.Renderer.SetColor(color.White)
		info := fmt.Sprintf(
			"Bilinear: %t [B]\nDither: %t [D]\nSmooth movement: %t [M]",
			flags.Has(Bilinear), flags.Has(Dithered), smoothMovement,
		)
		ctx.Renderer.Text(canvas, info, 8, 8, TextOpts(1.0, TopLeft.Snap(CapLine)))
	}

	app := NewTestApp(updater, drawer)
	app.Renderer.SetColorF32(0.5, 0.5, 0.5, 1.0)
	circleSource := app.Renderer.NewFilledCircle(72.0)
	setColorsLR(app.Renderer, tcWhite, tcClear)
	hiResGradientMask := app.Renderer.NewFilledRect(256, 128)
	loResGradientMask := app.Renderer.NewFilledRect(16, 8)

	hiResLongGradientMask := app.Renderer.NewFilledRect(640, 128)
	app.Renderer.SetColorF32(0.0, 0.2, 0.2, 1)
	stripSource := app.Renderer.NewFilledRect(640, 128)

	app.Renderer.SetColorF32A(tcWhite)
	app.Images = append(app.Images, circleSource, hiResGradientMask, loResGradientMask, hiResLongGradientMask, stripSource)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -tags interactive -run ^TestMaskAt$ . -count 1
func TestMaskAt(t *testing.T) {
	var flags flagList
	verticalMotion := false
	updater := func(ctx TestAppCtx) {
		flags.UpdateFlag(Bilinear, ebiten.KeyB)
		flags.UpdateFlag(Dithered, ebiten.KeyD)
		verticalMotion = updateToggle(ctx, ebiten.KeyM, verticalMotion)
	}
	drawer := func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Fill(backTestColor)
		if ctx.SpacePressed {
			ctx.Renderer.Options().Blend = ebiten.BlendCopy
		}

		lc := ctx.LeftClickF32()
		xShift := float32(ctx.DistAnim(256.0, 1.0))
		yShift := float32(0)
		if verticalMotion {
			yShift = float32(-8.0 + ctx.DistAnim(16.0, 1.3))
		}
		ctx.Renderer.MaskAt(canvas, ctx.Images[0], ctx.Images[1], lc.X+xShift, lc.Y+yShift, lc.X, lc.Y-16, flags...)

		_, ch := rectSizeF32(canvas.Bounds())
		x, y := float32(16.0), ch-16-float32(ctx.Images[3].Bounds().Dy())
		maskSlide := float32(ctx.DistAnim(32.0, 1.0))
		ctx.Renderer.MaskAt(canvas, ctx.Images[2], ctx.Images[3], x+32, y, x+maskSlide, y, flags...)
		ctx.Renderer.Options().Blend = ebiten.BlendSourceOver

		ctx.Renderer.SetColor(color.White)
		info := fmt.Sprintf(
			"Bilinear: %t [B]\nDither: %t [D]\nSoft vert motion: %t [M]",
			flags.Has(Bilinear), flags.Has(Dithered), verticalMotion,
		)
		ctx.Renderer.Text(canvas, info, 8, 8, TextOpts(1.0, TopLeft.Snap(CapLine)))
	}

	app := NewTestApp(updater, drawer)
	circ := app.Renderer.NewFilledCircle(32.0)
	setColorsLR(app.Renderer, tcWhite, tcClear)
	trans := app.Renderer.NewFilledRect(256, 96)
	longRect := app.Renderer.NewFilledRect(640, 128)
	app.Renderer.SetColorF32(0.0, 0.2, 0.2, 1)
	longImg := app.Renderer.NewFilledRect(640-64, 128)
	app.Renderer.SetColorF32A(tcWhite)
	app.Images = append(app.Images, circ, trans, longImg, longRect)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -tags interactive -run ^TestMaskHorz$ . -count 1
func TestMaskHorz(t *testing.T) {
	var flags flagList
	smoothMovement := false
	updater := func(ctx TestAppCtx) {
		flags.UpdateFlag(Bilinear, ebiten.KeyB)
		flags.UpdateFlag(Dithered, ebiten.KeyD)
		smoothMovement = updateToggle(ctx, ebiten.KeyM, smoothMovement)
	}
	drawer := func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Fill(backTestColor)
		if ctx.SpacePressed {
			ctx.Renderer.Options().Blend = ebiten.BlendCopy
		}

		lc := ctx.LeftClickF32()
		x, _ := ebiten.CursorPosition()
		xShift, yShift := float32(0), float32(0)
		if smoothMovement {
			xShift = float32(-4.0 + ctx.DistAnim(8.0, 1.3))
			yShift = float32(-4.0 + ctx.DistAnim(8.0, 1.0))
		}
		ctx.Renderer.MaskHorz(canvas, ctx.Images[0], lc.X+xShift, lc.Y+yShift, lc.X+256/2.0, float32(x), flags...)

		cw, ch := rectSizeF32(canvas.Bounds())
		om := CTR.AdjustXY(ctx.Images[1], cw/2, ch*2/3)
		i1w := float32(ctx.Images[1].Bounds().Dx())
		ctx.Renderer.MaskHorz(canvas, ctx.Images[1], om.X, om.Y, om.X, om.X+i1w, flags...)
		ctx.Renderer.Options().Blend = ebiten.BlendSourceOver

		ctx.Renderer.SetColor(color.White)
		info := fmt.Sprintf(
			"Bilinear: %t [B]\nDither: %t [D]\nSmooth movement: %t [M]",
			flags.Has(Bilinear), flags.Has(Dithered), smoothMovement,
		)
		ctx.Renderer.Text(canvas, info, 8, 8, TextOpts(1.0, TopLeft.Snap(CapLine)))
	}

	app := NewTestApp(updater, drawer)
	rect := app.Renderer.NewFilledRect(256, 64)
	app.Renderer.SetColorF32(0.2, 0.2, 0.2, 1.0)
	rectDither := app.Renderer.NewFilledRect(640, 96)
	app.Renderer.SetColorF32A(tcWhite)
	app.Images = append(app.Images, rect, rectDither)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -tags interactive -run ^TestMaskCirc$ . -count 1
func TestMaskCirc(t *testing.T) {
	var flags flagList
	smoothMovement := false
	updater := func(ctx TestAppCtx) {
		flags.UpdateFlag(Bilinear, ebiten.KeyB)
		flags.UpdateFlag(Dithered, ebiten.KeyD)
		smoothMovement = updateToggle(ctx, ebiten.KeyM, smoothMovement)
	}
	drawer := func(canvas *ebiten.Image, ctx TestAppCtx) {
		largeSquareSource, smallSquareSource := ctx.Images[0], ctx.Images[1]

		canvas.Fill(backTestColor)
		if ctx.SpacePressed {
			ctx.Renderer.Options().Blend = ebiten.BlendCopy
		}
		xShift, yShift := float32(0), float32(0)
		if smoothMovement {
			xShift = float32(-4.0 + ctx.DistAnim(8.0, 1.3))
			yShift = float32(-4.0 + ctx.DistAnim(8.0, 1.0))
		}
		lc := ctx.LeftClickF32().AddXY(xShift, yShift)
		rc := ctx.RightClickF32().AddXY(xShift, yShift)
		hardRadius := 48.0 + float32(ctx.DistAnim(16.0, 0.25))
		softEdge := float32(ctx.DistAnim(16.0, 1.0))
		o := CTR.AdjustXY(largeSquareSource, lc.X, lc.Y)
		ctx.Renderer.MaskCirc(canvas, largeSquareSource, o.X, o.Y, lc.X, lc.Y, hardRadius, softEdge, flags...)

		// the circular fade only cuts the square's corners, leaving straight
		// source edges visible to check bilinear filtering during movement.
		o = CTR.AdjustXY(smallSquareSource, rc.X, rc.Y)
		ctx.Renderer.MaskCirc(canvas, smallSquareSource, o.X, o.Y, rc.X, rc.Y, 32, 16, flags...)
		ctx.Renderer.Options().Blend = ebiten.BlendSourceOver

		ctx.Renderer.SetColor(color.White)
		labelOpts := TextOpts(1.0, TopCenter.Snap(CapLine))
		ctx.Renderer.Text(canvas, "Circular fade", lc.X, lc.Y+hardRadius+softEdge+8, labelOpts)
		ctx.Renderer.Text(canvas, "Source edges", rc.X, rc.Y+32+8, labelOpts)
		info := fmt.Sprintf(
			"Bilinear: %t [B]\nDither: %t [D]\nSmooth movement: %t [M]",
			flags.Has(Bilinear), flags.Has(Dithered), smoothMovement,
		)
		ctx.Renderer.Text(canvas, info, 8, 8, TextOpts(1.0, TopLeft.Snap(CapLine)))
	}

	app := NewTestApp(updater, drawer)
	largeSquareSource := app.Renderer.NewFilledRect(360, 360)
	smallSquareSource := app.Renderer.NewFilledRect(64, 64)
	app.Images = append(app.Images, largeSquareSource, smallSquareSource)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -tags interactive -run ^TestMaskThreshold$ . -count 1
func TestMaskThreshold(t *testing.T) {
	const Size = 256

	var flags flagList
	smoothMovement := false
	updater := func(ctx TestAppCtx) {
		flags.UpdateFlag(Bilinear, ebiten.KeyB)
		smoothMovement = updateToggle(ctx, ebiten.KeyM, smoothMovement)
	}
	drawer := func(canvas *ebiten.Image, ctx TestAppCtx) {
		reveal := -0.1 + float32(ctx.ModAnim(1.2, 0.5))
		source, mask, reference := ctx.Images[0], ctx.Images[1], ctx.Images[2]

		canvas.Fill(backTestColor)
		w, h := rectSizeF32(canvas.Bounds())
		ox, oy := w/2-Size/2, h/2-Size/2
		if smoothMovement {
			ox += float32(-4.0 + ctx.DistAnim(8.0, 1.3))
			oy += float32(-4.0 + ctx.DistAnim(8.0, 1.0))
		}
		ctx.Renderer.Options().Blend = ebiten.BlendSourceOver
		ctx.Renderer.DrawAt(canvas, reference, ox, oy, 1.0, flags...)
		if ctx.SpacePressed {
			ctx.Renderer.Options().Blend = ebiten.BlendCopy
		}
		ctx.Renderer.MaskThreshold(canvas, source, mask, reveal, ox, oy, flags...)
		ctx.Renderer.Options().Blend = ebiten.BlendSourceOver

		ctx.Renderer.SetColor(color.White)
		info := fmt.Sprintf(
			"Reveal: %.02f\nBilinear: %t [B]\nSmooth movement: %t [M]",
			reveal, flags.Has(Bilinear), smoothMovement,
		)
		ctx.Renderer.Text(canvas, info, 8, 8, TextOpts(1.0, TopLeft.Snap(CapLine)))
	}

	app := NewTestApp(updater, drawer)
	maskTarget := ebiten.NewImage(Size, Size)
	gradientOpts := StepGradientOpts(color.RGBA{0, 0, 0, 0}, color.RGBA{255, 255, 255, 255}, 16)
	app.Renderer.Gradient(maskTarget, gradientOpts, DirRadsLTR)
	whiteRect := app.Renderer.NewFilledRect(Size, Size)
	reference := ebiten.NewImage(Size, Size)
	reference.Fill(color.RGBA{128, 0, 0, 128})
	app.Images = append(app.Images, whiteRect, maskTarget, reference)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -tags interactive -run ^TestBakeAlphaMaskRadial$ . -count 1
func TestBakeAlphaMaskRadial(t *testing.T) {
	const (
		Size             = 256
		MaskOriginOffset = 64
	)
	randomness := float32(0.3)
	pattern := MaskPatternDefault
	bakeCX, bakeCY := float32(Size/2), float32(Size/2)
	var justClicked bool
	var rebake bool
	var revealStartTick uint64
	var smoothMovement = true
	var useNonZeroMaskOrigin bool

	var flags flagList
	updater := func(ctx TestAppCtx) {
		justClicked = false
		rebake = false
		flags.UpdateFlag(Bilinear, ebiten.KeyB)
		smoothMovement = updateToggle(ctx, ebiten.KeyM, smoothMovement)
		nextPattern := updateParam(ctx, ebiten.KeyP, pattern, MaskPatternDefault, maskPatternEndSentinel-1, 1)
		if nextPattern != pattern {
			pattern = nextPattern
			rebake = true
			revealStartTick = ctx.Ticks
		}
		switch {
		case inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft):
			justClicked = true
			rebake = true
			revealStartTick = ctx.Ticks
		case inpututil.IsKeyJustPressed(ebiten.KeyArrowUp):
			randomness = min(randomness+0.1, 1.0)
		case inpututil.IsKeyJustPressed(ebiten.KeyArrowDown):
			randomness = max(randomness-0.1, 0.0)
		case inpututil.IsKeyJustPressed(ebiten.KeyO):
			useNonZeroMaskOrigin = !useNonZeroMaskOrigin
		}
	}

	drawer := func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Fill(color.Black)
		info := fmt.Sprintf(
			"Pattern: %d/%d [P]\nRandomness: %.02f [Up/Down]\nNon-zero mask origin: %t [O]\nBilinear: %t [B]\nSmooth movement: %t [M]\nClick to rebake and restart",
			pattern+1, maskPatternEndSentinel, randomness, useNonZeroMaskOrigin, flags.Has(Bilinear), smoothMovement,
		)
		tox, toy := rectOriginF32(canvas.Bounds())
		ctx.Renderer.SetColorF32A(tcWhite)
		ctx.Renderer.Text(canvas, info, tox+8, toy+8, TextOpts(1.0, TopLeft.Snap(CapLine)))

		w, h := rectSizeF32(canvas.Bounds())
		ox, oy := w/2-Size/2, h/2-Size/2
		if smoothMovement {
			ox += float32(-4.0 + ctx.DistAnim(8.0, 0.666))
			oy += float32(-4.0 + ctx.DistAnim(8.0, 0.5))
		}
		if justClicked {
			lc := ctx.LeftClickF32()
			bakeCX, bakeCY = lc.X-ox, lc.Y-oy
		}
		if rebake {
			ctx.Renderer.Options().Blend = ebiten.BlendCopy
			for _, maskOffscreen := range ctx.Images[1:] {
				tox, toy := rectOriginF32(maskOffscreen.Bounds())
				ctx.Renderer.BakeAlphaMaskRadial(maskOffscreen, tox+bakeCX, toy+bakeCY, Size*1.44, randomness, pattern)
			}
			ctx.Renderer.Options().Blend = ebiten.BlendSourceOver
		}

		maskIndex := 1
		if useNonZeroMaskOrigin {
			maskIndex = 2
		}
		mask := ctx.Renderer.UnsafeTempCopy(0, ctx.Images[maskIndex], 0, true)
		revealCtx := ctx
		revealCtx.Ticks -= revealStartTick
		reveal := -0.1 + float32(revealCtx.ModAnim(2.0, 0.2))
		ctx.Renderer.MaskThreshold(canvas, ctx.Images[0], mask, reveal, ox, oy, flags...)
	}

	app := NewTestApp(updater, drawer)
	zeroOriginMask := ebiten.NewImage(Size, Size)
	nonZeroOriginMask := ebiten.NewImageWithOptions(image.Rect(MaskOriginOffset, MaskOriginOffset, MaskOriginOffset+Size, MaskOriginOffset+Size), nil)
	whiteRect := app.Renderer.NewFilledRect(Size, Size)
	app.Renderer.BakeAlphaMaskRadial(zeroOriginMask, bakeCX, bakeCY, Size, randomness, pattern)
	app.Renderer.BakeAlphaMaskRadial(nonZeroOriginMask, MaskOriginOffset+bakeCX, MaskOriginOffset+bakeCY, Size, randomness, pattern)
	app.Images = append(app.Images, whiteRect, zeroOriginMask, nonZeroOriginMask)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}
