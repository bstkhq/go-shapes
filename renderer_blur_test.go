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

// go test -run ^TestApplyBlur$ . -count 1
func TestApplyBlur(t *testing.T) {
	radius := float32(64.0)
	fxRadius := float32(16.0)
	app := NewTestApp(func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Fill(color.RGBA{0, 0, 255, 255})

		lx, ly := ctx.LeftClickF32()
		ctx.Renderer.SetColorF32(0, 0, 0, 1.0)
		ctx.Renderer.DrawCircle(canvas, lx, ly, radius+fxRadius)
		ctx.Renderer.SetColor(color.RGBA{0, 0, 255, 255})
		modRadius := float32(ctx.DistAnim(float64(fxRadius), 1.0))
		ctx.Renderer.ApplyBlur(canvas, ctx.Images[0], lx-radius, ly-radius, modRadius, 1.0)

		rx, ry := ctx.RightClickF32()
		ctx.Renderer.SetColor(color.RGBA{255, 0, 0, 255})
		clrMix := float32(ctx.DistAnim(1.0, 1.0))
		ctx.Renderer.ApplyBlur(canvas, ctx.Images[0], rx-radius, ry-radius, fxRadius, clrMix)
	})
	app.Images = append(app.Images, app.Renderer.NewCircle(float64(radius)))
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestApplyBlur2$ . -count 1
func TestApplyBlur2(t *testing.T) {
	radius := float32(64.0)
	fxRadius := float32(16.0)
	app := NewTestApp(func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Fill(color.Black)

		lx, ly := ctx.LeftClickF32()
		ctx.Renderer.SetColor(color.RGBA{0, 0, 255, 255})
		r := float32(ctx.DistAnim(float64(fxRadius), 1.0))
		ctx.Renderer.ApplyBlur2(canvas, ctx.Images[0], lx-radius, ly-radius, r, 1.0)
		ctx.Renderer.ApplyBlur(canvas, ctx.Images[0], lx+radius, ly-radius, r, 1.0)
		if ebiten.IsKeyPressed(ebiten.KeySpace) {
			// NOTE: there are still differences between blur and blur2, as can be
			// seen here, though they are fairly small. I tested many things, but
			// I can't see where it comes from. Floating point precision loss is the
			// most likely candidate, alongside gamma/linearization, but even individual
			// horz/vert blurs have differences, which is the suspicious part. short on
			// both directions, slightly offset on vertical (see TestApplyDirBlur)
			ctx.Renderer.SetBlend(ebiten.BlendXor)
			ctx.Renderer.ApplyBlur2(canvas, ctx.Images[0], lx+radius, ly-radius, r, 1.0)
			ctx.Renderer.SetBlend(ebiten.BlendSourceOver)
		}

		rx, ry := ctx.RightClickF32()
		ctx.Renderer.SetColor(color.RGBA{255, 0, 0, 255})
		clrMix := float32(ctx.DistAnim(1.0, 1.0))
		ctx.Renderer.ApplyBlur2(canvas, ctx.Images[0], rx-radius, ry-radius, fxRadius, clrMix)
	})
	app.Images = append(app.Images, app.Renderer.NewCircle(float64(radius)))
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestApplyDirBlur$ . -count 1
func TestApplyDirBlur(t *testing.T) {
	radius := float32(64.0)
	fxRadius := float32(16.0)
	app := NewTestApp(func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Fill(color.Black)

		lx, ly := ctx.LeftClickF32()
		ctx.Renderer.SetColor(color.RGBA{0, 0, 255, 255})
		r := float32(ctx.DistAnim(float64(fxRadius), 1.0))
		ctx.Renderer.ApplyVertBlur(canvas, ctx.Images[0], lx-radius, ly-radius, r, 1.0)

		rx, ry := ctx.RightClickF32()
		ctx.Renderer.SetColor(color.RGBA{255, 0, 0, 255})
		ctx.Renderer.ApplyHorzBlur(canvas, ctx.Images[0], rx-radius, ry-radius, fxRadius, 0.0)

		canvas.SubImage(image.Rect(480-8, 96-16, 480+80-8, 96+16)).(*ebiten.Image).Fill(color.RGBA{0, 255, 0, 255})
		ctx.Renderer.SetColor(color.RGBA{255, 255, 255, 255})
		if ebiten.IsKeyPressed(ebiten.KeySpace) {
			// see notes on TestApplyBlur2
			ctx.Renderer.ApplyVertBlur(canvas, ctx.Images[1], 480, 96, 15.5, 1.0)
			ctx.Renderer.SetBlend(ebiten.BlendXor)
			ctx.Renderer.ApplyBlur2(canvas, ctx.Images[2], 480, 96, 15.5, 1.0)
			ctx.Renderer.SetBlend(ebiten.BlendSourceOver)
		} else {
			ctx.Renderer.ApplyVertBlur(canvas, ctx.Images[1], 480, 96, 15.5, 1.0)
		}
	})

	rect := ebiten.NewImage(80, 80)
	rect.Fill(color.RGBA{255, 0, 0, 255})
	rect2 := ebiten.NewImage(80, 80)
	app.Renderer.SetColorF32(1, 0, 0, 1)
	app.Renderer.DrawIntArea(rect2, 0, 20, 80, 40)
	app.Renderer.DrawIntArea(rect2, 20, 0, 40, 80)
	app.Renderer.SetColorF32(1, 1, 1, 1)
	app.Images = append(app.Images, app.Renderer.NewCircle(float64(radius)), rect, rect2)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestApplyBlurK$ . -count 1
func TestApplyBlurK(t *testing.T) {
	radius := float32(64.0)
	dscale := DownscaleX4
	bicubic := false
	app := NewTestApp(func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Fill(color.Black)
		ebiten.SetWindowTitle(fmt.Sprintf("%s [downscaling: x%d (D), bicubic = %t (B)]", ctx.Title(), dscale.Factor(), bicubic))

		if ctx.NewInput {
			if inpututil.IsKeyJustPressed(ebiten.KeyD) {
				dscale += 1
				if dscale > DownscaleX16 {
					dscale = DownscaleNone
				}
			} else if inpututil.IsKeyJustPressed(ebiten.KeyB) {
				bicubic = !bicubic
			}
		}

		lx, ly := ctx.LeftClickF32()
		ctx.Renderer.SetColor(color.RGBA{255, 0, 255, 255})
		if ebiten.IsKeyPressed(ebiten.KeySpace) {
			ctx.Renderer.ApplyBlur2(canvas, ctx.Images[0], lx-radius, ly-radius, 16.0, 1.0)
		} else {
			kOpts := KernelOptions{Downscaling: dscale, HorzKernel: GaussK17, VertKernel: GaussK5, ColorMix: 1.0}
			kOpts.Scaling = &ScaleOptions{Bicubic: bicubic}
			ctx.Renderer.ApplyBlurK(canvas, ctx.Images[0], lx-radius, ly-radius, kOpts)
		}
		ctx.Renderer.SetColor(color.RGBA{128, 0, 128, 128})
		ctx.Renderer.DrawCircle(canvas, lx, ly, radius)

		rx, ry := ctx.RightClickF32()
		if ebiten.IsKeyPressed(ebiten.KeySpace) {
			ctx.Renderer.ApplyBlur2(canvas, ctx.Images[0], rx-radius, ry-radius, 16.0, 1.0)
		} else {
			kOpts := KernelOptions{Downscaling: dscale, HorzKernel: GaussK11, VertKernel: GaussK11, ColorMix: 1.0}
			kOpts.Scaling = &ScaleOptions{Bicubic: bicubic}
			ctx.Renderer.ApplyBlurK(canvas, ctx.Images[0], rx-radius, ry-radius, kOpts)
		}

		_, ch := rectSizeF32(canvas.Bounds())
		k := GaussK17
		if ebiten.IsKeyPressed(ebiten.KeySpace) {
			ctx.Renderer.ApplyBlur(canvas, ctx.Images[0], 16, ch-16-radius*2, float32(k.Radius()), 1.0)
		} else {
			kOpts := KernelOptions{Downscaling: DownscaleNone, HorzKernel: k, VertKernel: k, ColorMix: 1.0}
			ctx.Renderer.ApplyBlurK(canvas, ctx.Images[0], 16, ch-16-radius*2, kOpts)
		}
	})
	app.Images = append(app.Images, app.Renderer.NewCircle(float64(radius)))
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestApplyBlurKernLoop$ . -count 1
func TestApplyBlurKernLoop(t *testing.T) {
	radius := float32(64.0)
	app := NewTestApp(func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Fill(color.Black)

		lx, ly := ctx.LeftClickF32()
		ctx.Renderer.SetColor(color.RGBA{255, 0, 255, 255})
		const mink, maxk = GaussK3, GaussK17

		delta := uint64(maxk - mink)
		loop := (120 * delta)
		t := float64(ctx.Ticks%loop) / float64(loop)
		kern := GaussKernel(math.Round(lerp(float64(mink), float64(maxk), t)))

		kOpts := KernelOptions{Downscaling: DownscaleX4, HorzKernel: kern, VertKernel: kern, ColorMix: 1.0}
		ctx.Renderer.ApplyBlurK(canvas, ctx.Images[0], lx-radius, ly-radius, kOpts)

		rx, ry := ctx.RightClickF32()
		ctx.Renderer.ApplyBlur2(canvas, ctx.Images[0], rx-radius, ry-radius, min(float32(kern.Radius())*4.0, 16.0), 1.0)

		ebiten.SetWindowTitle(fmt.Sprintf("kern size: %d, radius: %d", kern.Size(), kern.Radius()))
	})
	app.Images = append(app.Images, app.Renderer.NewCircle(float64(radius)))
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestApplyBlurKernBleed$ . -count 1
func TestApplyBlurKernBleed(t *testing.T) {
	app := NewTestApp(func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Fill(color.Black)

		i1, i2, i3 := 0, 1, 2
		if ebiten.IsKeyPressed(ebiten.KeySpace) {
			i1, i2, i3 = i2, i3, i1
		}
		kOpts := KernelOptions{Downscaling: DownscaleX4, ColorMix: 1.0}
		withKernel := func(opts KernelOptions, kernel GaussKernel) KernelOptions {
			opts.HorzKernel = kernel
			opts.VertKernel = kernel
			return opts
		}
		ctx.Renderer.ApplyBlurK(canvas, ctx.Images[i1], 16, 16, withKernel(kOpts, GaussK9))
		ctx.Renderer.ApplyBlurK(canvas, ctx.Images[i2], 16+96*1, 16, withKernel(kOpts, GaussK17))
		ctx.Renderer.ApplyBlurK(canvas, ctx.Images[i3], 16+96*2, 16, withKernel(kOpts, GaussK11))
		ctx.Renderer.ApplyBlurK(canvas, ctx.Images[i3], 16, 16+96*1, withKernel(kOpts, GaussK5))
		ctx.Renderer.ApplyBlurK(canvas, ctx.Images[i2], 16+96*1, 16+96*1, withKernel(kOpts, GaussK13))
		ctx.Renderer.ApplyBlurK(canvas, ctx.Images[i1], 16+96*2, 16+96*1, withKernel(kOpts, GaussK9))
	})
	app.Renderer.SetColorF32(1, 0, 1, 1)
	img1 := app.Renderer.NewRect(33, 33)
	app.Renderer.SetColorF32(0, 1, 1, 1)
	img2 := app.Renderer.NewRect(50, 50)
	app.Renderer.SetColorF32(1, 1, 0, 1)
	img3 := app.Renderer.NewRect(67, 67)
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
	app := NewTestApp(func(canvas *ebiten.Image, ctx TestAppCtx) {
		const FxRadius = 16
		canvas.Fill(color.RGBA{0, 0, 255, 255})

		if ctx.NewInput {
			if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
				sampling *= 2
				if sampling > 64 {
					sampling = 8
				}
			}
			if inpututil.IsKeyJustPressed(ebiten.KeyD) {
				downscale += 1
				if downscale > DownscaleX16 {
					downscale = DownscaleNone
				}
			}
		}
		seed := float32(1.0)
		if ebiten.IsKeyPressed(ebiten.KeyN) {
			seed = rand.Float32()
		}
		ebiten.SetWindowTitle(ctx.Title() + fmt.Sprintf(" [sampling = %d, downscaling x%d]", sampling, downscale.Factor()))

		cw, ch := rectSizeF32(canvas.Bounds())
		ctx.Renderer.ApplyBlur(canvas, ctx.Images[0], 16, ch-16-radius*2, FxRadius, 1.0)
		ctx.Renderer.ApplyBlur(canvas, ctx.Images[0], 32+radius*2, ch-16-radius*2, FxRadius, 1.0)
		ctx.Renderer.SetColor(color.RGBA{0, 0, 255, 255})
		ctx.Renderer.ApplyBlurVogel(canvas, ctx.Images[0], 16, ch-16-radius*2, FxRadius, 0.0, sampling, downscale, seed)
		ctx.Renderer.ApplyBlur(canvas, ctx.Images[0], 32+radius*2, ch-16-radius*2, FxRadius, 0.0)

		ctx.Renderer.ApplyBlurVogel(canvas, ctx.Images[1], cw-float32(ctx.Images[1].Bounds().Dx())-16, 16, FxRadius, 1.0, sampling, downscale, seed)

		ctx.Renderer.SetColor(color.RGBA{255, 0, 0, 255})
		clrMix := float32(ctx.DistAnim(1.0, 1.0))

		lx, ly := ctx.LeftClickF32()
		ctx.Renderer.ApplyBlur(canvas, ctx.Images[0], lx-radius, ly-radius, FxRadius, clrMix)

		rx, ry := ctx.RightClickF32()
		ctx.Renderer.SetColor(color.RGBA{255, 0, 0, 255})
		ctx.Renderer.ApplyBlurVogel(canvas, ctx.Images[0], rx-radius, ry-radius, FxRadius, clrMix, sampling, downscale, seed)

	})

	circle := app.Renderer.NewCircle(float64(radius))
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

	app := NewTestApp(func(canvas *ebiten.Image, ctx TestAppCtx) {
		seed := float32(1.0)
		if ebiten.IsKeyPressed(ebiten.KeySpace) {
			seed = rand.Float32()
		}

		radius := ctx.DistAnim(64, 0.5)
		ctx.Renderer.ApplyBlurVogel(canvas, full, 0, 0, float32(radius), 1.0, Sampling, Downscale, seed)
	})

	app.Renderer.GradientRadialDither(full, 1920/2, 960, color.RGBA{0, 64, 196, 255}, color.RGBA{0, 128, 255, 255}, 320, 960, Float32Inf(), 1.0)
	app.Renderer.GradientDither(full, 0, 0, 1920, 1080, color.RGBA{64, 0, 64, 64}, color.RGBA{32, 0, 0, 64}, DirRadsBLTR, 1.0)
	for range 256 {
		rand.Float64()
		app.Renderer.SetColorF32(rand.Float32(), rand.Float32(), rand.Float32(), 1.0)
		app.Renderer.ScaleAlphaBy(0.5)
		app.Renderer.DrawCircle(full, rand.Float32()*1920, rand.Float32()*1080, 16+rand.Float32()*64)
	}
	app.Renderer.SetColorF32(1, 1, 1, 1)

	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}
