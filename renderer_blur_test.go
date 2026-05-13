package shapes

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"math/rand/v2"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// go test -run ^TestBlur$ . -count 1
func TestBlur(t *testing.T) {
	radius := float32(64.0)
	fxRadius := float32(16.0)
	updater := func(TestAppCtx) {}
	drawer := func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Fill(color.RGBA{0, 0, 255, 255})

		lx, ly := ctx.LeftClickF32()
		ctx.Renderer.SetColorF32(0, 0, 0, 1.0)
		ctx.Renderer.FillCircle(canvas, lx, ly, radius+fxRadius)
		ctx.Renderer.SetColor(color.RGBA{0, 0, 255, 255})
		modRadius := float32(ctx.DistAnim(float64(fxRadius), 1.0))
		ctx.Renderer.Blur(canvas, ctx.Images[0], lx-radius, ly-radius, modRadius)

		rx, ry := ctx.RightClickF32()
		ctx.Renderer.SetColor(color.RGBA{255, 0, 0, 255})
		ctx.Renderer.SetTint(float32(ctx.DistAnim(1.0, 1.0)))
		ctx.Renderer.Blur(canvas, ctx.Images[0], rx-radius, ry-radius, fxRadius)
		ctx.Renderer.SetTint(0)
	}

	app := NewTestApp(updater, drawer)
	app.Images = append(app.Images, app.Renderer.NewFilledCircle(float64(radius)))
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestBlur2$ . -count 1
func TestBlur2(t *testing.T) {
	radius := float32(64.0)
	fxRadius := float32(16.0)
	updater := func(TestAppCtx) {}
	drawer := func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Fill(color.Black)

		lx, ly := ctx.LeftClickF32()
		ctx.Renderer.SetColor(color.RGBA{0, 0, 255, 255})
		r := float32(ctx.DistAnim(float64(fxRadius), 1.0))
		ctx.Renderer.Blur2(canvas, ctx.Images[0], lx-radius, ly-radius, r)
		ctx.Renderer.Blur(canvas, ctx.Images[0], lx+radius, ly-radius, r)
		if ctx.SpacePressed {
			// NOTE: there are still differences between blur and blur2, as can be
			// seen here, though they are fairly small. I tested many things, but
			// I can't see where it comes from. Floating point precision loss is the
			// most likely candidate, alongside gamma/linearization, but even individual
			// horz/vert blurs have differences, which is the suspicious part. short on
			// both directions, slightly offset on vertical (see TestApplyDirBlur)
			ctx.Renderer.Options().Blend = ebiten.BlendXor
			ctx.Renderer.Blur2(canvas, ctx.Images[0], lx+radius, ly-radius, r)
			ctx.Renderer.Options().Blend = ebiten.BlendSourceOver
		}

		rx, ry := ctx.RightClickF32()
		ctx.Renderer.SetColor(color.RGBA{255, 0, 0, 255})
		ctx.Renderer.SetTint(float32(ctx.DistAnim(1.0, 1.0)))
		ctx.Renderer.Blur2(canvas, ctx.Images[0], rx-radius, ry-radius, fxRadius)
		ctx.Renderer.SetTint(0)
	}

	app := NewTestApp(updater, drawer)
	app.Images = append(app.Images, app.Renderer.NewFilledCircle(float64(radius)))
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestDirBlur$ . -count 1
func TestDirBlur(t *testing.T) {
	radius := float32(64.0)
	fxRadius := float32(24.0)
	updater := func(TestAppCtx) {}
	drawer := func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Fill(color.Black)

		lx, ly := ctx.LeftClickF32()
		ctx.Renderer.SetColor(color.RGBA{0, 0, 255, 255})
		r := float32(ctx.DistAnim(float64(fxRadius), 1.0))
		ctx.Renderer.blurVert(canvas, ctx.Images[0], lx-radius, ly-radius, r)

		rx, ry := ctx.RightClickF32()
		ctx.Renderer.SetColor(color.RGBA{255, 0, 0, 255})
		ctx.Renderer.SetTint(1)
		ctx.Renderer.blurHorz(canvas, ctx.Images[0], rx-radius, ry-radius, fxRadius)
		ctx.Renderer.SetTint(0)

		rect := image.Rect(480-8, 96-16, 480+80-8, 96+16)
		Paint(canvas, rect, [4]float32{0, 1, 0, 1}, ebiten.BlendSourceOver)
		ctx.Renderer.SetColor(color.RGBA{255, 255, 255, 255})
		if ctx.SpacePressed {
			// see notes on TestBlur2
			ctx.Renderer.blurVert(canvas, ctx.Images[1], 480, 96, 15.5)
			ctx.Renderer.Options().Blend = ebiten.BlendXor
			ctx.Renderer.Blur2(canvas, ctx.Images[2], 480, 96, 15.5)
			ctx.Renderer.Options().Blend = ebiten.BlendSourceOver
		} else {
			ctx.Renderer.blurVert(canvas, ctx.Images[1], 480, 96, 15.5)
		}
	}

	app := NewTestApp(updater, drawer)
	rect := ebiten.NewImage(80, 80)
	rect.Fill(color.RGBA{255, 0, 0, 255})
	rect2 := ebiten.NewImage(80, 80)
	app.Renderer.SetColorF32(1, 0, 0, 1)
	app.Renderer.FillIntRect(rect2, RectWithSize(0, 20, 80, 40), 0)
	app.Renderer.FillIntRect(rect2, RectWithSize(20, 0, 40, 80), 0)
	app.Renderer.SetColorF32(1, 1, 1, 1)
	app.Images = append(app.Images, app.Renderer.NewFilledCircle(float64(radius)), rect, rect2)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestBlurK$ . -count 1
func TestBlurK(t *testing.T) {
	radius := float32(64.0)
	dscale := DownscaleX4
	bicubic := false
	updater := func(ctx TestAppCtx) {
		ebiten.SetWindowTitle(fmt.Sprintf("%s [downscaling: x%d (D), bicubic: %t (B)]", ctx.Title(), dscale.Factor(), bicubic))
		dscale = updateParam(ctx, ebiten.KeyD, dscale, DownscaleNone, DownscaleX16, 1)
		if inpututil.IsKeyJustPressed(ebiten.KeyB) {
			bicubic = !bicubic
		}
	}
	drawer := func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Fill(color.Black)

		lx, ly := ctx.LeftClickF32()
		ctx.Renderer.SetColor(color.RGBA{255, 0, 255, 255})
		if ctx.SpacePressed {
			ctx.Renderer.Blur2(canvas, ctx.Images[0], lx-radius, ly-radius, 16.0)
		} else {
			kOpts := KernelOptions{Downscaling: dscale, HorzKernel: GaussK17, VertKernel: GaussK5}
			kOpts.Scaling = &ScaleOptions{Bicubic: bicubic}
			ctx.Renderer.BlurK(canvas, ctx.Images[0], lx-radius, ly-radius, kOpts)
		}
		ctx.Renderer.SetColor(color.RGBA{128, 0, 128, 128})
		ctx.Renderer.FillCircle(canvas, lx, ly, radius)

		rx, ry := ctx.RightClickF32()
		k := GaussK9
		if ctx.SpacePressed {
			ctx.Renderer.Blur2(canvas, ctx.Images[0], rx-radius, ry-radius, float32(k.Radius())*4.0)
		} else {
			kOpts := KernelOptions{Downscaling: dscale, HorzKernel: k, VertKernel: k}
			kOpts.Scaling = &ScaleOptions{Bicubic: bicubic}
			ctx.Renderer.BlurK(canvas, ctx.Images[0], rx-radius, ry-radius, kOpts)
		}

		_, ch := rectSizeF32(canvas.Bounds())
		k = GaussK29
		if ctx.SpacePressed {
			ctx.Renderer.Blur(canvas, ctx.Images[0], 16, ch-16-radius*2, float32(k.Radius()))
		} else {
			kOpts := KernelOptions{Downscaling: DownscaleNone, HorzKernel: k, VertKernel: k}
			ctx.Renderer.BlurK(canvas, ctx.Images[0], 16, ch-16-radius*2, kOpts)
		}
	}

	app := NewTestApp(updater, drawer)
	app.Images = append(app.Images, app.Renderer.NewFilledCircle(float64(radius)))
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestBlurKLoop$ . -count 1
func TestBlurKLoop(t *testing.T) {
	const Radius = 64.0

	updater := func(TestAppCtx) {}
	drawer := func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Fill(color.Black)

		lx, ly := ctx.LeftClickF32()
		ctx.Renderer.SetColor(color.RGBA{255, 0, 255, 255})
		const mink, maxk = GaussK3, GaussK17

		delta := uint64(maxk - mink)
		loop := (120 * delta)
		t := float64(ctx.Ticks%loop) / float64(loop)
		kern := GaussKernel(math.Round(lerp(float64(mink), float64(maxk), t)))
		ebiten.SetWindowTitle(fmt.Sprintf("kern size: %d, radius: %d", kern.Size(), kern.Radius()))

		kOpts := KernelOptions{Downscaling: DownscaleX4, HorzKernel: kern, VertKernel: kern}
		ctx.Renderer.BlurK(canvas, ctx.Images[0], lx-Radius, ly-Radius, kOpts)

		rx, ry := ctx.RightClickF32()
		ctx.Renderer.Blur2(canvas, ctx.Images[0], rx-Radius, ry-Radius, min(float32(kern.Radius())*4.0, 32.0))
	}

	app := NewTestApp(updater, drawer)
	app.Images = append(app.Images, app.Renderer.NewFilledCircle(Radius))
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestApplyBlurKBleed$ . -count 1
func TestApplyBlurKBleed(t *testing.T) {
	updater := func(TestAppCtx) {}
	drawer := func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Fill(color.Black)

		i1, i2, i3 := 0, 1, 2
		if ctx.SpacePressed {
			i1, i2, i3 = i2, i3, i1
		}
		kOpts := KernelOptions{Downscaling: DownscaleX4}
		withKernel := func(opts KernelOptions, kernel GaussKernel) KernelOptions {
			opts.HorzKernel = kernel
			opts.VertKernel = kernel
			return opts
		}
		ctx.Renderer.BlurK(canvas, ctx.Images[i1], 16, 16, withKernel(kOpts, GaussK9))
		ctx.Renderer.BlurK(canvas, ctx.Images[i2], 16+96*1, 16, withKernel(kOpts, GaussK17))
		ctx.Renderer.BlurK(canvas, ctx.Images[i3], 16+96*2, 16, withKernel(kOpts, GaussK11))
		ctx.Renderer.BlurK(canvas, ctx.Images[i3], 16, 16+96*1, withKernel(kOpts, GaussK5))
		ctx.Renderer.BlurK(canvas, ctx.Images[i2], 16+96*1, 16+96*1, withKernel(kOpts, GaussK13))
		ctx.Renderer.BlurK(canvas, ctx.Images[i1], 16+96*2, 16+96*1, withKernel(kOpts, GaussK9))
	}

	app := NewTestApp(updater, drawer)
	app.Renderer.SetColorF32(1, 0, 1, 1)
	img1 := app.Renderer.NewFilledRect(33, 33)
	app.Renderer.SetColorF32(0, 1, 1, 1)
	img2 := app.Renderer.NewFilledRect(50, 50)
	app.Renderer.SetColorF32(1, 1, 0, 1)
	img3 := app.Renderer.NewFilledRect(67, 67)
	app.Renderer.SetColorF32(1, 1, 1, 1)
	app.Images = append(app.Images, img1, img2, img3)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestApplyBlurVogel$ . -count 1
func TestApplyBlurVogel(t *testing.T) {
	radius := float32(64.0)
	sampling := 8
	downscale := DownscaleNone
	updater := func(ctx TestAppCtx) {
		downscale = updateParam(ctx, ebiten.KeyD, downscale, DownscaleNone, DownscaleX16, 1)
		sampling = updateParamMult(ctx, ebiten.KeyS, sampling, 8, 64, 2)
		ebiten.SetWindowTitle(ctx.Title() + fmt.Sprintf(" [[S]ampling = %d, [D]ownscaling x%d]", sampling, downscale.Factor()))
	}
	drawer := func(canvas *ebiten.Image, ctx TestAppCtx) {
		const FxRadius = 16
		canvas.Fill(color.RGBA{0, 0, 255, 255})

		seed := float32(1.0)
		if ctx.SpacePressed {
			seed = rand.Float32()
		}

		cw, ch := rectSizeF32(canvas.Bounds())
		ctx.Renderer.Blur(canvas, ctx.Images[0], 16, ch-16-radius*2, FxRadius)
		ctx.Renderer.Blur(canvas, ctx.Images[0], 32+radius*2, ch-16-radius*2, FxRadius)
		ctx.Renderer.SetColor(color.RGBA{0, 0, 255, 255})
		ctx.Renderer.SetTint(1)
		ctx.Renderer.BlurVogel(canvas, ctx.Images[0], 16, ch-16-radius*2, FxRadius, sampling, downscale, seed)
		ctx.Renderer.Blur(canvas, ctx.Images[0], 32+radius*2, ch-16-radius*2, FxRadius)
		ctx.Renderer.SetTint(0)

		ctx.Renderer.BlurVogel(canvas, ctx.Images[1], cw-float32(ctx.Images[1].Bounds().Dx())-16, 16, FxRadius, sampling, downscale, seed)

		lx, ly := ctx.LeftClickF32()
		rx, ry := ctx.RightClickF32()

		ctx.Renderer.SetColor(color.RGBA{255, 0, 0, 255})
		ctx.Renderer.SetTint(float32(ctx.DistAnim(1.0, 1.0)))
		ctx.Renderer.Blur(canvas, ctx.Images[0], lx-radius, ly-radius, FxRadius)

		ctx.Renderer.SetColor(color.RGBA{255, 0, 0, 255})
		ctx.Renderer.BlurVogel(canvas, ctx.Images[0], rx-radius, ry-radius, FxRadius, sampling, downscale, seed)
		ctx.Renderer.SetTint(0)
	}

	app := NewTestApp(updater, drawer)
	circle := app.Renderer.NewFilledCircle(float64(radius))
	rect := ebiten.NewImage(120, 80)
	app.Renderer.StrokeIntRect(rect, rect.Bounds(), 0, 16)
	app.Images = append(app.Images, circle, rect)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestApplyBlurVogelFull$ . -count 1
func TestApplyBlurVogelFull(t *testing.T) {
	const Sampling = 16
	const Downscale = DownscaleNone

	var full *ebiten.Image = ebiten.NewImage(1920, 1080)
	updater := func(TestAppCtx) {}
	drawer := func(canvas *ebiten.Image, ctx TestAppCtx) {
		seed := float32(1.0)
		if ctx.SpacePressed {
			seed = rand.Float32()
		}

		radius := ctx.DistAnim(64, 0.5)
		ctx.Renderer.BlurVogel(canvas, full, 0, 0, float32(radius), Sampling, Downscale, seed)
	}

	app := NewTestApp(updater, drawer)
	for range 256 {
		rand.Float64()
		app.Renderer.SetColorF32(rand.Float32(), rand.Float32(), rand.Float32(), 1.0)
		app.Renderer.ScaleAlphaBy(0.5)
		app.Renderer.FillCircle(full, rand.Float32()*1920, rand.Float32()*1080, 16+rand.Float32()*64)
	}
	app.Renderer.SetColorF32(1, 1, 1, 1)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}
