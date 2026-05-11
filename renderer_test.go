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

// go test -run ^TestScale$ . -count 1
func TestScale(t *testing.T) {
	// TODO: test minification, unclear if we should disallow DstSampling in that case
	updater := func(TestAppCtx) {}
	drawer := func(canvas *ebiten.Image, ctx TestAppCtx) {
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
	}

	app := NewTestApp(updater, drawer)
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

// go test -run ^TestDrawAt$ . -count 1
func TestDrawAt(t *testing.T) {
	flags := newFlagList()
	updater := func(ctx TestAppCtx) {
		ebiten.SetWindowTitle(fmt.Sprintf("%s [bilinear: %t, dithered: %t]", ctx.Title(), flags[Bilinear] == Bilinear, flags[Dithered] == Dithered))
		if inpututil.IsKeyJustPressed(ebiten.KeyF) {
			flags.Flip(Bilinear)
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyD) {
			flags.Flip(Dithered)
		}
	}
	drawer := func(canvas *ebiten.Image, ctx TestAppCtx) {
		lx, ly := ctx.LeftClickF32()
		x, y := CTR.Adjust(ctx.Images[0], lx, ly)
		y += -8.0 + float32(ctx.DistAnim(16, 1.0))
		if !ebiten.IsKeyPressed(ebiten.KeySpace) {
			ctx.Renderer.DrawAt(canvas, ctx.Images[0], x, y, 1.0, flags...)
		} else {
			mark := image.Rectangle{Max: canvas.Bounds().Max}
			mark.Min = mark.Max.Sub(image.Pt(16, 8))
			Paint(canvas, mark, [4]float32{0.5, 0.5, 0, 1.0}, ebiten.BlendSourceOver)

			var opts ebiten.DrawImageOptions
			if flags.Has(Bilinear) {
				opts.Filter = ebiten.FilterLinear
			}
			opts.GeoM.Translate(float64(x), float64(y))
			canvas.DrawImage(ctx.Images[0], &opts)
		}

		rx, ry := ctx.RightClickF32()
		x, y = CTR.Adjust(ctx.Images[1], rx, ry)
		ctx.Renderer.SetTint(0.5 + float32(ctx.DistAnim(0.5, 1.0)))
		alpha := float32(ctx.DistAnim(1.0, 0.333))
		ctx.Renderer.DrawAt(canvas, ctx.Images[1], x, y, alpha, flags...)
		ctx.Renderer.SetTint(0)

		cw, ch := rectSizeF32(canvas.Bounds())
		alpha = float32(0.025 + ctx.DistAnim(0.025, 1.0))
		ctx.Renderer.SetTint(1)
		x, y = CTR.Adjust(ctx.Images[1], cw/2, ch/2)
		y += float32(-4 + ctx.DistAnim(8.0, 1.0))
		ctx.Renderer.DrawAt(canvas, ctx.Images[1], x, y, alpha, flags...)
		ctx.Renderer.SetTint(0)
	}

	const RW, RH = 96 * 2, 64 * 2
	const CR = 72
	app := NewTestApp(updater, drawer)
	rect := ebiten.NewImage(RW, RH)
	gradientOpts := GradientOpts(color.RGBA{255, 230, 200, 255}, color.RGBA{200, 230, 255, 255}, false)
	app.Renderer.Gradient(rect, gradientOpts, DirRadsBLTR)
	circ := app.Renderer.NewFilledCircle(CR)
	app.Renderer.Options().Blend = ebiten.BlendSourceIn
	gradientOpts = GradientOpts(color.RGBA{0, 230, 200, 255}, color.RGBA{255, 0, 236, 255}, false)
	app.Renderer.Gradient(circ, gradientOpts, DirRadsTTB)
	app.Renderer.Options().Blend = ebiten.BlendSourceOver
	app.Images = append(app.Images, rect, circ)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestDrawImgShader$ . -count 1
func TestDrawImgShader(t *testing.T) {
	updater := func(TestAppCtx) {}
	drawer := func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Fill(color.Black)

		for i := range 2 {
			ox, oy := float32(i)*200+8+float32(ctx.DistAnim(16, 1.0)), 8+float32(ctx.DistAnim(16, 0.5))
			if ebiten.IsKeyPressed(ebiten.KeySpace) {
				var opts ebiten.DrawImageOptions
				opts.GeoM.Translate(float64(ox), float64(oy))
				canvas.DrawImage(ctx.Images[i], &opts)
				opts.GeoM.Translate(16, 72)
				canvas.DrawImage(ctx.Images[i], &opts)
			} else {
				ctx.Renderer.setFlatCustomVAs01(1, 1)
				ctx.Renderer.DrawImgShader(canvas, ctx.Images[i], ox, oy, NoMargins, shaderBilinear.Load())
				ctx.Renderer.DrawAt(canvas, ctx.Images[i], ox+16, oy+72, 1.0)
			}
		}

		for i := range 2 {
			ox, oy := float32(i)*200+8+float32(ctx.DistAnim(16, 1.0)), 300+float32(ctx.DistAnim(16, 0.5))
			bounds := ctx.Images[i].Bounds()
			subox, suboy := int(ox)+16, int(oy)+72
			subRect := image.Rect(subox, suboy, subox+32, suboy+24)
			sub := canvas.SubImage(subRect).(*ebiten.Image)
			if ebiten.IsKeyPressed(ebiten.KeySpace) {
				var opts ebiten.DrawRectShaderOptions
				opts.GeoM.Translate(float64(ox), float64(oy))
				canvas.DrawRectShader(bounds.Dx(), bounds.Dy(), shaderDefault.Load(), &opts)
				sub.Fill(color.White)
			} else {
				ctx.Renderer.setFlatCustomVAs01(1, 1)
				ctx.Renderer.DrawRectShader(canvas, ox, oy, float32(bounds.Dx()), float32(bounds.Dy()), NoMargins, shaderDefault.Load())
				sox, soy, sw, sh := rectOriginSizeF32(subRect)
				ctx.Renderer.DrawRectShader(sub, sox, soy, sw, sh, NoMargins, shaderDefault.Load())
			}
		}
	}

	app := NewTestApp(updater, drawer)
	rect := app.Renderer.NewFilledRect(64, 48)
	rectO := ebiten.NewImageWithOptions(image.Rect(16, 16, 16+64, 16+48), nil)
	rectO.Fill(color.White)
	app.Images = append(app.Images, rect, rectO)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestDrawCircShader$ . -count 1
func TestDrawCircShader(t *testing.T) {
	var startDegs float32
	var degs float32 = 360
	var tolerance float32

	updater := func(ctx TestAppCtx) {
		ebiten.SetWindowTitle(fmt.Sprintf(
			"%s [[S]tartDegs: %.02f, [D]egrees: %.02f, [T]olerance: %.02f]",
			ctx.Title(), startDegs, degs, tolerance,
		))
		startDegs = updateParam(ctx, ebiten.KeyS, startDegs, 0, 360, 15)
		degs = updateParam(ctx, ebiten.KeyD, degs, 0, 360, 15)
		tolerance = updateParam(ctx, ebiten.KeyT, tolerance, 0, 10.0, 0.2)
	}
	drawer := func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Fill(color.Black)
		w, h := rectSizeF32(canvas.Bounds())
		r := min(w, h) * 0.333
		th := r * 0.1

		ctx.Renderer.SetColorF32(0.5, 0.5, 0.5, 0.5)
		ctx.Renderer.StrokeCircle(canvas, w/2, h/2, r, th)

		ctx.Renderer.SetColorF32(0.5, 0.0, 0.0, 0.5)
		opts := CircShaderOpts(r, th)
		opts.StartAngle = startDegs * math.Pi / 180
		if degs >= 359.999 {
			opts.EndAngle = opts.StartAngle + degs*math.Pi/180
		} else {
			opts.EndAngle = uradsAddCW(opts.StartAngle, degs*math.Pi/180)
		}
		opts.Tolerance = tolerance
		ctx.Renderer.DrawCircShader(canvas, w/2, h/2, opts, shaderDefault.Load())
	}

	app := NewTestApp(updater, drawer)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}
