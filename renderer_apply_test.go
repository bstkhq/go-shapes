package shapes

import (
	"image"
	"image/color"
	"math"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

// go test -run ^TestApplyExpansion$ . -count 1
func TestApplyExpansion(t *testing.T) {
	radius := float32(64.0)
	expansion := float32(16.0)
	app := NewTestApp(func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Fill(color.Black)

		lx, ly := ctx.LeftClickF32()
		ctx.Renderer.SetColor(color.RGBA{0, 128, 128, 255})
		ctx.Renderer.DrawCircle(canvas, lx, ly, radius+expansion)
		ctx.Renderer.SetColor(color.RGBA{128, 0, 0, 128})
		x := float32(ctx.DistAnim(float64(expansion), 1.0))
		ctx.Renderer.ApplyExpansion(canvas, ctx.Images[0], lx-radius, ly-radius, x)
	})
	app.Images = append(app.Images, app.Renderer.NewCircle(float64(radius)))
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestApplyExpansionRect$ . -count 1
func TestApplyExpansionRect(t *testing.T) {
	const Radius = 64.0
	const Expansion = 16.0
	app := NewTestApp(func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Fill(color.Black)

		lx, ly := ctx.LeftClickF32()
		rx, ry := ctx.RightClickF32()
		ctx.Renderer.SetColor(color.RGBA{255, 0, 0, 255})
		expansion := float32(ctx.DistAnim(Expansion, 1.0))
		ctx.Renderer.ApplyExpansionRect(canvas, ctx.Images[0], lx-Radius, ly-Radius, expansion)
		ctx.Renderer.ApplyExpansionRect(canvas, ctx.Images[1], rx-Radius, ry-Radius, expansion)

		ctx.Renderer.SetColor(color.RGBA{255, 0, 0, 255})
		ctx.DrawWithAlphaAtF32(canvas, ctx.Images[0], 0.5, lx-Radius, ly-Radius)
		ctx.DrawWithAlphaAtF32(canvas, ctx.Images[3], 0.5, lx-Radius-Expansion, ly-Radius-Expansion)
		ctx.DrawWithAlphaAtF32(canvas, ctx.Images[1], 0.5, rx-Radius, ry-Radius)
		ctx.DrawWithAlphaAtF32(canvas, ctx.Images[2], 0.5, rx-Radius-Expansion, ry-Radius-Expansion)
	})
	img1 := app.Renderer.NewRect(int(Radius*2), int(Radius*2))
	img4 := app.Renderer.NewRect(int((Radius+Expansion)*2), int((Radius+Expansion)*2))
	img2 := app.Renderer.NewCircle(float64(Radius))
	img3 := app.Renderer.NewCircle(float64(Radius + Expansion))
	app.Images = append(app.Images, img1, img2, img3, img4)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestApplyErosion$ . -count 1
func TestApplyErosion(t *testing.T) {
	radius := float32(96.0)
	erosion := float32(16.0)
	app := NewTestApp(func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Fill(color.Black)

		lx, ly := ctx.LeftClickF32()
		ctx.Renderer.SetColor(color.RGBA{255, 255, 255, 255})
		ctx.Renderer.DrawCircle(canvas, lx, ly, radius)

		ctx.Renderer.SetColor(color.RGBA{255, 0, 0, 255})
		ctx.Renderer.DrawCircle(canvas, lx, ly, radius-erosion)

		r := float32(ctx.DistAnim(float64(erosion), 1.0))
		ctx.Renderer.SetColor(color.RGBA{0, 0, 164, 164})
		ctx.Renderer.ApplyErosion(canvas, ctx.Images[0], lx-radius, ly-radius, r)

		rx, ry := ctx.RightClickF32()
		ctx.Renderer.SetColor(color.RGBA{172, 0, 224, 255})
		ctx.Renderer.ApplyErosion(canvas, ctx.Images[1], rx-128, ry-82, r)
	})

	circle := app.Renderer.NewCircle(float64(radius))
	triangle := ebiten.NewImage(256, 164)
	app.Renderer.StrokeTriangle(triangle, 16, 16, 256-16, 16, 16, 164-16, float64(erosion)*2, 0)
	app.Images = append(app.Images, circle, triangle)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestApplyOutline$ . -count 1
func TestApplyOutline(t *testing.T) {
	radius := float32(64.0)
	thick := float32(8.0)
	app := NewTestApp(func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Fill(color.Black)

		lx, ly := ctx.LeftClickF32()
		ctx.Renderer.SetColor(color.RGBA{0, 255, 0, 255})
		ctx.Renderer.DrawCircle(canvas, lx, ly, radius+thick/2+1.0)
		ctx.Renderer.SetColor(color.RGBA{255, 0, 0, 255})
		ctx.Renderer.DrawCircle(canvas, lx, ly, radius-thick/2-1.0)

		ctx.Renderer.SetColor(color.RGBA{0, 0, 255, 255})
		ctx.Renderer.ApplyOutline(canvas, ctx.Images[0], lx-radius, ly-radius, thick)

		rx, ry := ctx.RightClickF32()
		ctx.Renderer.SetColor(color.RGBA{255, 255, 255, 255})
		if ebiten.IsKeyPressed(ebiten.KeySpace) {
			ctx.DrawAtF32(canvas, ctx.Images[0], rx-radius, ry-radius)
		} else {
			ctx.Renderer.ApplyOutline(canvas, ctx.Images[0], rx-radius, ry-radius, thick)
		}
	})
	app.Images = append(app.Images, app.Renderer.NewCircle(float64(radius)))
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestApplyHardShadow$ . -count 1
func TestApplyHardShadow(t *testing.T) {
	radius := float32(64.0)
	app := NewTestApp(func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Fill(color.Black)

		lx, ly := ctx.LeftClickF32()
		ctx.Renderer.SetColor(color.RGBA{0, 128, 128, 128})
		ctx.Renderer.ApplyHardShadow(canvas, ctx.Images[0], lx-radius, ly-radius, 0, 8.0, ClampLeft)
		ctx.Renderer.SetColor(color.RGBA{255, 255, 255, 255})
		ctx.Renderer.DrawCircle(canvas, lx, ly, radius)

		rx, ry := ctx.RightClickF32()
		ctx.Renderer.SetColor(color.RGBA{128, 128, 128, 128})
		ctx.Renderer.ApplyHardShadow(canvas, ctx.Images[0], rx-radius, ry-radius, 8.0, 0.0, ClampNone)
		ctx.Renderer.SetColor(color.RGBA{255, 255, 255, 255})
		ctx.Renderer.DrawCircle(canvas, rx, ry, radius)
	})
	app.Images = append(app.Images, app.Renderer.NewCircle(float64(radius)))
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestApplyShadow$ . -count 1
func TestApplyShadow(t *testing.T) {
	app := NewTestApp(func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Fill(color.Black)

		lx, ly := ctx.LeftClickF32()
		ctx.Renderer.SetColor(color.RGBA{0, 128, 128, 128})
		ctx.Renderer.ApplyShadow(canvas, ctx.Images[0], lx, ly, 0, -16.0, 4.0, ClampBottom)
		ctx.DrawAtF32(canvas, ctx.Images[0], lx, ly)

		rx, ry := ctx.RightClickF32()
		ctx.Renderer.SetColor(color.RGBA{128, 128, 128, 128})
		ctx.Renderer.ApplyShadow(canvas, ctx.Images[0], rx, ry, -12.0, -12.0, 9.0, ClampBottom)
		ctx.DrawAtF32(canvas, ctx.Images[0], rx, ry)

		mx, my := min(lx, rx), max(ly, ry)
		const MaxRadius = 16.0
		r := float32(ctx.DistAnim(MaxRadius, 1.0))
		ctx.Renderer.SetColorF32(1.0, 0.0, 0.0, 1.0)
		ctx.Renderer.DrawCircle(canvas, mx+64, my+64, 64+r)
		ctx.Renderer.SetColor(color.RGBA{0, 196, 196, 196})
		ctx.Renderer.ApplyShadow(canvas, ctx.Images[1], mx, my, 0, 0, r, ClampNone)
		ctx.DrawAtF32(canvas, ctx.Images[1], mx, my)

		mx, my = max(lx, rx), min(ly, ry)
		ctx.Renderer.SetColorF32(1.0, 1.0, 1.0, 1.0)
		ctx.Renderer.ApplyBlur(canvas, ctx.Images[1], mx-32, my, r, 1.0)
		ctx.Renderer.ApplyShadow(canvas, ctx.Images[1], mx+32, my, 0, 0, r, ClampNone)
	})
	circle := app.Renderer.NewCircle(64.0)
	circBounds := circle.Bounds()
	halfCircle := circle.SubImage(image.Rect(0, 0, circBounds.Dx(), circBounds.Dy()/2)).(*ebiten.Image)
	app.Images = append(app.Images, halfCircle, circle)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestApplyZoomShadow$ . -count 1
func TestApplyZoomShadow(t *testing.T) {
	app := NewTestApp(func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Fill(color.Black)

		lx, ly := ctx.LeftClickF32()
		ctx.Renderer.SetColor(color.RGBA{0, 128, 128, 128})
		ctx.Renderer.ApplyZoomShadow(canvas, ctx.Images[0], lx, ly, 0, 0, 2.0, ClampNone)
		ctx.DrawAtF32(canvas, ctx.Images[0], lx, ly)

		rx, ry := ctx.RightClickF32()

		ctx.Renderer.SetColor(color.RGBA{128, 0, 128, 128})
		ctx.Renderer.ApplyZoomShadow(canvas, ctx.Images[1], rx, ry, 0, 0, 1.2, ClampBottom)
		ctx.DrawAtF32(canvas, ctx.Images[1], rx, ry)
	})
	circle := app.Renderer.NewCircle(64.0)
	circBounds := circle.Bounds()
	halfCircle := circle.SubImage(image.Rect(0, 0, circBounds.Dx(), circBounds.Dy()/2)).(*ebiten.Image)
	app.Images = append(app.Images, circle, halfCircle)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestApplyGlow2$ . -count 1
func TestApplyGlow2(t *testing.T) {
	app := NewTestApp(func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Fill(color.Black)

		lx, ly := ctx.LeftClickF32()
		ctx.DrawAtF32(canvas, ctx.Images[0], lx, ly)
		ctx.Renderer.ApplyGlow2(canvas, ctx.Images[0], lx, ly, 16, 16, 0.4, 0.7, 1.0)

		rx, ry := ctx.RightClickF32()
		ctx.DrawAtF32(canvas, ctx.Images[0], rx, ry)
		ctx.Renderer.SetColor(color.RGBA{255, 192, 192, 255})
		dynRadius := float32(ctx.DistAnim(6, 2.0))
		ctx.Renderer.ApplyGlow2(canvas, ctx.Images[0], rx, ry, 10+dynRadius, 16, 0.5, 0.6, 0.0)
	})
	const s, m = 96, 16
	cross := ebiten.NewImage(s, s)
	app.Renderer.SetColor(color.RGBA{96, 240, 240, 255})
	app.Renderer.DrawLine(cross, m, m, s-m, s-m, m/2)
	app.Renderer.DrawLine(cross, s-m, m, m, s-m, m/2)
	app.Images = append(app.Images, cross)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestApplyHorzGlow$ . -count 1
func TestApplyHorzGlow(t *testing.T) {
	app := NewTestApp(func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Fill(color.Black)

		lx, ly := ctx.LeftClickF32()
		ctx.DrawAtF32(canvas, ctx.Images[0], lx, ly)
		ctx.Renderer.ApplyHorzGlow(canvas, ctx.Images[0], lx, ly, 16, 0.4, 0.5, 1.0)

		rx, ry := ctx.RightClickF32()
		ctx.DrawAtF32(canvas, ctx.Images[0], rx, ry)
		ctx.Renderer.SetColor(color.RGBA{255, 192, 192, 255})
		dynRadius := float32(ctx.DistAnim(6, 2.0))
		ctx.Renderer.ApplyHorzGlow(canvas, ctx.Images[0], rx, ry, 10+dynRadius, 0.5, 0.6, 0.0)
	})
	const s, m = 96, 16
	cross := ebiten.NewImage(s, s)
	app.Renderer.SetColor(color.RGBA{96, 240, 240, 255})
	app.Renderer.DrawLine(cross, m, m, s-m, s-m, m/2)
	app.Renderer.DrawLine(cross, s-m, m, m, s-m, m/2)
	app.Images = append(app.Images, cross)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestApplyDarkHorzGlow$ . -count 1
func TestApplyDarkHorzGlow(t *testing.T) {
	app := NewTestApp(func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Fill(color.White)

		lx, ly := ctx.LeftClickF32()
		ctx.DrawAtF32(canvas, ctx.Images[0], lx, ly)
		ctx.Renderer.ApplyDarkHorzGlow(canvas, ctx.Images[0], lx, ly, 16, 0.5, 0.01, 1.0)
		ctx.DrawAtF32(canvas, ctx.Images[0], lx, ly+120)

		rx, ry := ctx.RightClickF32()
		ctx.DrawAtF32(canvas, ctx.Images[1], rx, ry)
		dynRadius := float32(ctx.DistAnim(16, 4.0))
		ctx.Renderer.SetColor(color.RGBA{64, 0, 0, 255})
		ctx.Renderer.ApplyDarkHorzGlow(canvas, ctx.Images[1], rx, ry, dynRadius, 1, 0.5, 0.0)
		_, h := rectSizeF32(ctx.Images[1].Bounds())
		ctx.Renderer.ApplyHorzBlur(canvas, ctx.Images[1], rx, ry-h-h/16, dynRadius, 1)
		ctx.DrawAtF32(canvas, ctx.Images[1], rx, ry-h-h/16)
	})
	const s, m = 96, 16
	cross := ebiten.NewImage(s, s)
	app.Renderer.SetColor(color.RGBA{0, 0, 128, 255})
	app.Renderer.DrawLine(cross, m, m, s-m, s-m, m/2)
	app.Renderer.DrawLine(cross, s-m, m, m, s-m, m/2)
	img := ebiten.NewImage(s, s)
	app.Renderer.SetColor(color.RGBA{0, 0, 0, 255})
	app.Renderer.SimpleGradient(img, color.RGBA{255, 255, 255, 255}, color.RGBA{128, 0, 0, 255}, math.Pi/2)
	app.Images = append(app.Images, img, cross)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestApplyGlowK$ . -count 1
func TestApplyGlowK(t *testing.T) {
	app := NewTestApp(func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Fill(color.Black)

		lx, ly := ctx.LeftClickF32()
		ctx.DrawAtF32(canvas, ctx.Images[0], lx, ly)
		hkern, vkern := GaussK3, GaussK3
		gRad := func(kern GaussKernel) float32 {
			// NOTE: this still doesn't match because the edges are blurred/dilated
			// on downscaling too
			radius := float32(kern.Size()*4-1) / 2.0
			return min(radius, 16.0)
		}
		if ebiten.IsKeyPressed(ebiten.KeySpace) {
			ctx.Renderer.ApplyGlow2(canvas, ctx.Images[0], lx, ly, gRad(hkern), gRad(vkern), 0.2, 0.8, 1.0)
		} else {
			kOpts := KernelOptions{Downscaling: DownscaleX4, HorzKernel: hkern, VertKernel: vkern, ColorMix: 1.0}
			ctx.Renderer.ApplyGlowK(canvas, ctx.Images[0], lx, ly, 0.2, 0.8, kOpts)
		}

		rx, ry := ctx.RightClickF32()
		ctx.DrawAtF32(canvas, ctx.Images[0], rx, ry)
		ctx.Renderer.SetColor(color.RGBA{255, 192, 192, 255})
		hkern, vkern = GaussK5, GaussK15
		if ebiten.IsKeyPressed(ebiten.KeySpace) {
			ctx.Renderer.ApplyGlow2(canvas, ctx.Images[0], rx, ry, gRad(hkern), gRad(vkern), 0.5, 0.6, 0.0)
		} else {
			kOpts := KernelOptions{Downscaling: DownscaleX4, HorzKernel: hkern, VertKernel: vkern}
			ctx.Renderer.ApplyGlowK(canvas, ctx.Images[0], rx, ry, 0.5, 0.6, kOpts)
		}
	})
	const s, m = 96, 16
	tri := ebiten.NewImage(s, s)
	app.Renderer.SetColor(color.RGBA{96, 240, 240, 255})
	app.Renderer.DrawTriangle(tri, m, s-m, s/2, m, s-m, s-m, 0)
	app.Images = append(app.Images, tri)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestApplyGlowKBleed$ . -count 1
// Notice: some bleeding edge cases are quite difficult to reproduce
// and haven't been able to catch them through tests yet, only live
// code in more complex projects.
func TestApplyGlowKBleed(t *testing.T) {
	app := NewTestApp(func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Fill(color.Black)

		i1, i2, i3 := 0, 1, 2
		if ebiten.IsKeyPressed(ebiten.KeySpace) {
			i1, i2, i3 = i2, i3, i1
		}
		const st, et = 0.0, 0.5
		kOpts := KernelOptions{Downscaling: DownscaleX4, ColorMix: 1.0}
		withKernel := func(opts KernelOptions, kernel GaussKernel) KernelOptions {
			opts.HorzKernel = kernel
			opts.VertKernel = kernel
			return opts
		}
		ctx.Renderer.ApplyGlowK(canvas, ctx.Images[i1], 16, 16, st, et, withKernel(kOpts, GaussK9))
		ctx.Renderer.ApplyGlowK(canvas, ctx.Images[i2], 16+96*1, 16, st, et, withKernel(kOpts, GaussK17))
		ctx.Renderer.ApplyGlowK(canvas, ctx.Images[i3], 16+96*2, 16, st, et, withKernel(kOpts, GaussK11))
		ctx.Renderer.ApplyGlowK(canvas, ctx.Images[i3], 16, 16+96*1, st, et, withKernel(kOpts, GaussK5))
		ctx.Renderer.ApplyGlowK(canvas, ctx.Images[i2], 16+96*1, 16+96*1, st, et, withKernel(kOpts, GaussK13))
		ctx.Renderer.ApplyGlowK(canvas, ctx.Images[i1], 16+96*2, 16+96*1, st, et, withKernel(kOpts, GaussK9))
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

// go test -run ^TestApplyColorGlowK$ . -count 1
func TestApplyColorGlowK(t *testing.T) {
	app := NewTestApp(func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Fill(color.Black)

		lx, ly := ctx.LeftClickF32()
		ctx.DrawAtF32(canvas, ctx.Images[0], lx, ly)
		if !ebiten.IsKeyPressed(ebiten.KeySpace) {
			loThresh := 0.1 + ctx.DistAnim(0.4, 1.0)
			kOpts := KernelOptions{Downscaling: DownscaleX4, HorzKernel: GaussK7, VertKernel: GaussK7, ColorMix: 1.0}
			ctx.Renderer.ApplyColorGlowK(canvas, ctx.Images[0], lx, ly, RGBF32(color.RGBA{255, 255, 0, 255}), float32(loThresh), 1.0, kOpts)
		}
	})

	circ := app.Renderer.NewCircle(96.0)
	img := ebiten.NewImage(circ.Bounds().Dx(), circ.Bounds().Dy())
	app.Renderer.Gradient(img, circ, 0, 0, color.RGBA{255, 255, 0, 255}, color.RGBA{255, 0, 255, 255}, -1, 0, 1.0)
	app.Images = append(app.Images, img)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestScanlinesSharp . -count 1
func TestScanlinesSharp(t *testing.T) {
	app := NewTestApp(func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Fill(color.White)
		const darkThick, clearThick = 3, 1
		offset := float32(ctx.ModAnim(darkThick+clearThick, 1.0))
		ctx.Renderer.ApplyScanlinesSharp(canvas, darkThick, clearThick, 0.05, offset)
	})
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestWaveLines$ . -count 1
func TestWaveLines(t *testing.T) {
	const LineThick = 6.0

	offset := float32(0.0)
	app := NewTestApp(func(canvas *ebiten.Image, ctx TestAppCtx) {
		ebiten.SetFullscreen(true)
		canvas.Fill(color.White)
		offset += 0.666
		minFillRate := float32(0.2)
		maxFillRate := float32(0.8)
		radsOffset := ctx.ModAnim(2*math.Pi, 0.2)
		if ebiten.IsKeyPressed(ebiten.KeySpace) {
			radsOffset = 0.0
		}
		ctx.Renderer.ApplyWaveLines(canvas, LineThick, minFillRate, maxFillRate, 16.0, offset, DirRadsLTR+radsOffset)
	})

	app.Renderer.SetColorF32(0, 0, 0, 0.2, 0)
	app.Renderer.SetColorF32(0, 0.1, 0, 0.2, 1)
	app.Renderer.SetColorF32(0, 0.1, 0.1, 0.2, 3)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}
