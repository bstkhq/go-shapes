package shapes

import (
	"image/color"
	"math"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

// go test -run ^TestApplyExpansion$ . -count 1
func TestApplyExpansion(t *testing.T) {
	const radius, expansion = 64.0, 16.0

	updater := func(TestAppCtx) {}
	drawer := func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Fill(color.Black)

		lx, ly := ctx.LeftClickF32()
		ctx.Renderer.SetColor(color.RGBA{0, 128, 128, 255})
		ctx.Renderer.DrawCircle(canvas, lx, ly, radius+expansion)
		ctx.Renderer.SetColor(color.RGBA{128, 0, 0, 128})
		x := float32(ctx.DistAnim(float64(expansion), 1.0))
		ctx.Renderer.ApplyExpansion(canvas, ctx.Images[0], lx-radius, ly-radius, x)
	}

	app := NewTestApp(updater, drawer)
	app.Images = append(app.Images, app.Renderer.NewCircle(float64(radius)))
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestApplyExpansionRect$ . -count 1
func TestApplyExpansionRect(t *testing.T) {
	const Radius = 64.0
	const Expansion = 16.0

	updater := func(TestAppCtx) {}
	drawer := func(canvas *ebiten.Image, ctx TestAppCtx) {
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
	}

	app := NewTestApp(updater, drawer)
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
	const radius, erosion = 96.0, 16.0

	updater := func(TestAppCtx) {}
	drawer := func(canvas *ebiten.Image, ctx TestAppCtx) {
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
	}

	app := NewTestApp(updater, drawer)
	circle := app.Renderer.NewCircle(float64(radius))
	triangle := ebiten.NewImage(256, 164)
	var points [3]PointF32
	points[0] = PointF32{X: 16, Y: 16}
	points[1] = points[0].AddXY(224, 0)
	points[2] = points[0].AddXY(0, 132)
	app.Renderer.StrokeTriangle(triangle, points, erosion*2, 0)
	app.Images = append(app.Images, circle, triangle)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestApplyOutline$ . -count 1
func TestApplyOutline(t *testing.T) {
	const radius, thick = 64.0, 8.0

	updater := func(TestAppCtx) {}
	drawer := func(canvas *ebiten.Image, ctx TestAppCtx) {
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
		if ctx.SpacePressed {
			ctx.DrawAtF32(canvas, ctx.Images[0], rx-radius, ry-radius)
		} else {
			ctx.Renderer.ApplyOutline(canvas, ctx.Images[0], rx-radius, ry-radius, thick)
		}
	}

	app := NewTestApp(updater, drawer)
	app.Images = append(app.Images, app.Renderer.NewCircle(float64(radius)))
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestApplyGlow2$ . -count 1
func TestApplyGlow2(t *testing.T) {
	updater := func(TestAppCtx) {}
	drawer := func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Fill(color.Black)

		lx, ly := ctx.LeftClickF32()
		ctx.DrawAtF32(canvas, ctx.Images[0], lx, ly)
		ctx.Renderer.ApplyGlow2(canvas, ctx.Images[0], lx, ly, 16, 16, 0.4, 0.7)

		rx, ry := ctx.RightClickF32()
		ctx.DrawAtF32(canvas, ctx.Images[0], rx, ry)
		ctx.Renderer.SetColor(color.RGBA{255, 192, 192, 255})
		ctx.Renderer.SetTint(1)
		dynRadius := float32(ctx.DistAnim(6, 2.0))
		ctx.Renderer.ApplyGlow2(canvas, ctx.Images[0], rx, ry, 10+dynRadius, 16, 0.5, 0.6)
		ctx.Renderer.SetTint(0)
	}

	app := NewTestApp(updater, drawer)
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
	updater := func(TestAppCtx) {}
	drawer := func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Fill(color.Black)

		lx, ly := ctx.LeftClickF32()
		ctx.DrawAtF32(canvas, ctx.Images[0], lx, ly)
		ctx.Renderer.ApplyHorzGlow(canvas, ctx.Images[0], lx, ly, 16, 0.4, 0.5)

		rx, ry := ctx.RightClickF32()
		ctx.DrawAtF32(canvas, ctx.Images[0], rx, ry)
		ctx.Renderer.SetColor(color.RGBA{255, 192, 192, 255})
		ctx.Renderer.SetTint(1)
		dynRadius := float32(ctx.DistAnim(6, 2.0))
		ctx.Renderer.ApplyHorzGlow(canvas, ctx.Images[0], rx, ry, 10+dynRadius, 0.5, 0.6)
		ctx.Renderer.SetTint(0)
	}

	app := NewTestApp(updater, drawer)
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
	updater := func(TestAppCtx) {}
	drawer := func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Fill(color.White)

		lx, ly := ctx.LeftClickF32()
		ctx.DrawAtF32(canvas, ctx.Images[0], lx, ly)
		ctx.Renderer.ApplyDarkHorzGlow(canvas, ctx.Images[0], lx, ly, 16, 0.5, 0.01)
		ctx.DrawAtF32(canvas, ctx.Images[0], lx, ly+120)

		rx, ry := ctx.RightClickF32()
		ctx.DrawAtF32(canvas, ctx.Images[1], rx, ry)
		dynRadius := float32(ctx.DistAnim(16, 4.0))
		ctx.Renderer.SetColor(color.RGBA{64, 0, 0, 255})
		ctx.Renderer.SetTint(1)
		ctx.Renderer.ApplyDarkHorzGlow(canvas, ctx.Images[1], rx, ry, dynRadius, 1, 0.5)
		ctx.Renderer.SetTint(0)
		_, h := rectSizeF32(ctx.Images[1].Bounds())
		ctx.Renderer.ApplyHorzBlur(canvas, ctx.Images[1], rx, ry-h-h/16, dynRadius)
		ctx.DrawAtF32(canvas, ctx.Images[1], rx, ry-h-h/16)
	}

	const s, m = 96, 16
	app := NewTestApp(updater, drawer)
	cross := ebiten.NewImage(s, s)
	app.Renderer.SetColor(color.RGBA{0, 0, 128, 255})
	app.Renderer.DrawLine(cross, m, m, s-m, s-m, m/2)
	app.Renderer.DrawLine(cross, s-m, m, m, s-m, m/2)
	img := ebiten.NewImage(s, s)
	app.Renderer.SetColor(color.RGBA{0, 0, 0, 255})
	gradientOpts := GradientOpts(color.RGBA{255, 255, 255, 255}, color.RGBA{128, 0, 0, 255}, false)
	app.Renderer.Gradient(img, gradientOpts, math.Pi/2)
	app.Images = append(app.Images, img, cross)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestApplyGlowK$ . -count 1
func TestApplyGlowK(t *testing.T) {
	updater := func(TestAppCtx) {}
	drawer := func(canvas *ebiten.Image, ctx TestAppCtx) {
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
		if ctx.SpacePressed {
			ctx.Renderer.ApplyGlow2(canvas, ctx.Images[0], lx, ly, gRad(hkern), gRad(vkern), 0.2, 0.8)
		} else {
			kOpts := KernelOptions{Downscaling: DownscaleX4, HorzKernel: hkern, VertKernel: vkern}
			ctx.Renderer.ApplyGlowK(canvas, ctx.Images[0], lx, ly, 0.2, 0.8, kOpts)
		}

		rx, ry := ctx.RightClickF32()
		ctx.DrawAtF32(canvas, ctx.Images[0], rx, ry)
		ctx.Renderer.SetColor(color.RGBA{255, 192, 192, 255})
		hkern, vkern = GaussK5, GaussK15
		ctx.Renderer.SetTint(1.0)
		if ctx.SpacePressed {
			ctx.Renderer.ApplyGlow2(canvas, ctx.Images[0], rx, ry, gRad(hkern), gRad(vkern), 0.5, 0.6)
		} else {
			kOpts := KernelOptions{Downscaling: DownscaleX4, HorzKernel: hkern, VertKernel: vkern}
			ctx.Renderer.ApplyGlowK(canvas, ctx.Images[0], rx, ry, 0.5, 0.6, kOpts)
		}
		ctx.Renderer.SetTint(0)
	}

	const s, m = 96, 16
	app := NewTestApp(updater, drawer)
	tri := ebiten.NewImage(s, s)
	app.Renderer.SetColor(color.RGBA{96, 240, 240, 255})
	var points = [3]PointF32{{X: m, Y: s - m}, {X: s / 2, Y: m}, {X: s - m, Y: s - m}}
	app.Renderer.DrawTriangle(tri, points, 0)
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
	updater := func(TestAppCtx) {}
	drawer := func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Fill(color.Black)

		i1, i2, i3 := 0, 1, 2
		if ctx.SpacePressed {
			i1, i2, i3 = i2, i3, i1
		}
		const st, et = 0.0, 0.5
		const down = DownscaleX4
		ctx.Renderer.ApplyGlowK(canvas, ctx.Images[i1], 16, 16, st, et, KernelOpts(down, GaussK9))
		ctx.Renderer.ApplyGlowK(canvas, ctx.Images[i2], 16+96*1, 16, st, et, KernelOpts(down, GaussK17))
		ctx.Renderer.ApplyGlowK(canvas, ctx.Images[i3], 16+96*2, 16, st, et, KernelOpts(down, GaussK11))
		ctx.Renderer.ApplyGlowK(canvas, ctx.Images[i3], 16, 16+96*1, st, et, KernelOpts(down, GaussK5))
		ctx.Renderer.ApplyGlowK(canvas, ctx.Images[i2], 16+96*1, 16+96*1, st, et, KernelOpts(down, GaussK13))
		ctx.Renderer.ApplyGlowK(canvas, ctx.Images[i1], 16+96*2, 16+96*1, st, et, KernelOpts(down, GaussK9))
	}

	app := NewTestApp(updater, drawer)
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
	updater := func(TestAppCtx) {}
	drawer := func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Fill(color.Black)

		lx, ly := ctx.LeftClickF32()
		ctx.DrawAtF32(canvas, ctx.Images[0], lx, ly)
		if !ctx.SpacePressed {
			loThresh := 0.1 + ctx.DistAnim(0.4, 1.0)
			kOpts := KernelOpts(DownscaleX4, GaussK7)
			ctx.Renderer.ApplyColorGlowK(canvas, ctx.Images[0], lx, ly, RGBF32(color.RGBA{255, 255, 0, 255}), float32(loThresh), 1.0, kOpts)
		}
	}

	app := NewTestApp(updater, drawer)
	circ := app.Renderer.NewCircle(96.0)
	gradientOpts := GradientOpts(color.RGBA{255, 255, 0, 255}, color.RGBA{255, 0, 255, 255}, false)
	app.Renderer.Options().Blend = ebiten.BlendSourceIn
	app.Renderer.Gradient(circ, gradientOpts, DirRadsLTR)
	app.Renderer.Options().Blend = ebiten.BlendSourceOver
	app.Images = append(app.Images, circ)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestScanlinesSharp . -count 1
func TestScanlinesSharp(t *testing.T) {
	updater := func(TestAppCtx) {}
	drawer := func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Fill(color.White)
		const darkThick, clearThick = 3, 1
		offset := float32(ctx.ModAnim(darkThick+clearThick, 1.0))
		ctx.Renderer.ApplyScanlinesSharp(canvas, darkThick, clearThick, 0.05, offset)
	}

	app := NewTestApp(updater, drawer)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestWaveLines$ . -count 1
func TestWaveLines(t *testing.T) {
	const LineThick = 6.0

	offset := float32(0.0)
	updater := func(TestAppCtx) {}
	drawer := func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Fill(color.White)
		offset += 0.666
		minFillRate := float32(0.2)
		maxFillRate := float32(0.8)
		radsOffset := ctx.ModAnim(2*math.Pi, 0.2)
		if ctx.SpacePressed {
			radsOffset = 0.0
		}
		ctx.Renderer.ApplyWaveLines(canvas, LineThick, minFillRate, maxFillRate, 16.0, offset, DirRadsLTR+radsOffset)
	}

	app := NewTestApp(updater, drawer)
	app.Renderer.SetColorF32(0, 0, 0, 0.2, 0)
	app.Renderer.SetColorF32(0, 0.1, 0, 0.2, 1)
	app.Renderer.SetColorF32(0, 0.1, 0.1, 0.2, 3)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}
