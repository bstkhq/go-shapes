package shapes

import (
	"fmt"
	"image"
	"image/color"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// go test -run ^TestGlow2$ . -count 1
func TestGlow2(t *testing.T) {
	updater := func(ctx TestAppCtx) {
		if inpututil.IsKeyJustPressed(ebiten.KeyT) {
			ctx.Renderer.SetTint(1.0 - ctx.Renderer.Tint())
		}
	}
	drawer := func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Fill(color.Black)

		lx, ly := ctx.LeftClickF32()
		ctx.DrawAtF32(canvas, ctx.Images[1], lx, ly)
		ctx.Renderer.SetColorF32(1, 1, 1, 1)
		if ebiten.IsKeyPressed(ebiten.KeyAlt) {
			ctx.Renderer.Blur2(canvas, ctx.Images[1], lx, ly, 16, 16)
		} else {
			ctx.Renderer.Glow2(canvas, ctx.Images[1], lx, ly, 16, 16, 0.4, 0.7)
		}

		rx, ry := ctx.RightClickF32()
		ctx.DrawAtF32(canvas, ctx.Images[0], rx, ry)
		ctx.Renderer.SetColor(color.RGBA{255, 192, 192, 255})
		dynRadius := float32(ctx.DistAnim(6, 2.0))
		vertRadius := float32(16.0)
		if ctx.SpacePressed {
			vertRadius = float32(ctx.DistAnim(16.0, 0.5))
		}
		if ebiten.IsKeyPressed(ebiten.KeyAlt) {
			ctx.Renderer.Blur2(canvas, ctx.Images[0], rx, ry, 10+dynRadius, vertRadius)
		} else {
			ctx.Renderer.Glow2(canvas, ctx.Images[0], rx, ry, 10+dynRadius, vertRadius, 0.5, 0.6)
		}
	}

	app := NewTestApp(updater, drawer)
	const s, m = 96, 16
	cross := ebiten.NewImage(s, s)
	app.Renderer.SetColor(color.RGBA{96, 240, 240, 255})
	app.Renderer.StrokeLine(cross, m, m, s-m, s-m, m/2)
	app.Renderer.StrokeLine(cross, s-m, m, m, s-m, m/2)
	app.Renderer.SetColor(color.RGBA{96, 120, 0, 255}, 2, 3)
	circ := app.Renderer.NewFilledCircle(96)
	app.Images = append(app.Images, cross, circ)
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
			ctx.Renderer.Glow2(canvas, ctx.Images[0], lx, ly, gRad(hkern), gRad(vkern), 0.2, 0.8)
		} else {
			kOpts := KernelOptions{Downscaling: DownscaleX4, HorzKernel: hkern, VertKernel: vkern}
			ctx.Renderer.GlowK(canvas, ctx.Images[0], lx, ly, 0.2, 0.8, kOpts)
		}

		rx, ry := ctx.RightClickF32()
		ctx.DrawAtF32(canvas, ctx.Images[0], rx, ry)
		ctx.Renderer.SetColor(color.RGBA{255, 192, 192, 255})
		hkern, vkern = GaussK5, GaussK15
		ctx.Renderer.SetTint(1.0)
		if ctx.SpacePressed {
			ctx.Renderer.Glow2(canvas, ctx.Images[0], rx, ry, gRad(hkern), gRad(vkern), 0.5, 0.6)
		} else {
			kOpts := KernelOptions{Downscaling: DownscaleX4, HorzKernel: hkern, VertKernel: vkern}
			ctx.Renderer.GlowK(canvas, ctx.Images[0], rx, ry, 0.5, 0.6, kOpts)
		}
		ctx.Renderer.SetTint(0)
	}

	const s, m = 96, 16
	app := NewTestApp(updater, drawer)
	tri := ebiten.NewImage(s, s)
	app.Renderer.SetColor(color.RGBA{96, 240, 240, 255})
	var points = [3]PointF32{{X: m, Y: s - m}, {X: s / 2, Y: m}, {X: s - m, Y: s - m}}
	app.Renderer.FillTriangle(tri, points, 0)
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
		ctx.Renderer.GlowK(canvas, ctx.Images[i1], 16, 16, st, et, KernelOpts(down, GaussK9))
		ctx.Renderer.GlowK(canvas, ctx.Images[i2], 16+96*1, 16, st, et, KernelOpts(down, GaussK17))
		ctx.Renderer.GlowK(canvas, ctx.Images[i3], 16+96*2, 16, st, et, KernelOpts(down, GaussK11))
		ctx.Renderer.GlowK(canvas, ctx.Images[i3], 16, 16+96*1, st, et, KernelOpts(down, GaussK5))
		ctx.Renderer.GlowK(canvas, ctx.Images[i2], 16+96*1, 16+96*1, st, et, KernelOpts(down, GaussK13))
		ctx.Renderer.GlowK(canvas, ctx.Images[i1], 16+96*2, 16+96*1, st, et, KernelOpts(down, GaussK9))
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

// go test -run ^TestGlowColorK$ . -count 1
func TestGlowColorK(t *testing.T) {
	updater := func(TestAppCtx) {}
	drawer := func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Fill(color.Black)

		lx, ly := ctx.LeftClickF32()
		ctx.DrawAtF32(canvas, ctx.Images[0], lx, ly)
		if !ctx.SpacePressed {
			loThresh := 0.1 + ctx.DistAnim(0.4, 1.0)
			kOpts := KernelOpts(DownscaleX4, GaussK7)
			ctx.Renderer.GlowColorK(canvas, ctx.Images[0], lx, ly, RGBF32(color.RGBA{255, 255, 0, 255}), float32(loThresh), 1.0, kOpts)
		}
	}

	app := NewTestApp(updater, drawer)
	circ := app.Renderer.NewFilledCircle(96.0)
	gradientOpts := GradientOpts(color.RGBA{255, 255, 0, 255}, color.RGBA{255, 0, 255, 255}, false)
	app.Renderer.Options().Blend = ebiten.BlendSourceIn
	app.Renderer.Gradient(circ, gradientOpts, DirRadsLTR)
	app.Renderer.Options().Blend = ebiten.BlendSourceOver
	app.Images = append(app.Images, circ)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestGlowCompare$ . -count 1
func TestGlowCompare(t *testing.T) {
	modes := []string{"none", "std", "color", "dark"}
	mode := 0
	horzRadius := float32(16.0)
	vertRadius := float32(16.0)

	motionAnim := false
	intensityAnim := false
	backColors := []color.RGBA{{255, 255, 255, 255}, {128, 128, 128, 128}, {128, 64, 0, 128}}
	rendererColors := []color.RGBA{{255, 255, 255, 255}, {64, 192, 255, 255}, {192, 192, 192, 192}, {192, 192, 0, 192}}
	backColorIdx := 0
	rendererClrIdx := 0

	updater := func(ctx TestAppCtx) {
		mode = updateParam(ctx, ebiten.KeyM, mode, 0, 3, 1)
		backColorIdx = updateParam(ctx, ebiten.KeyB, backColorIdx, 0, 2, 1)
		rendererClrIdx = updateParam(ctx, ebiten.KeyR, rendererClrIdx, 0, 3, 1)
		horzRadius = float32(int(updateParam(ctx, ebiten.KeyH, horzRadius, 0, 32, 1.0)))
		vertRadius = float32(int(updateParam(ctx, ebiten.KeyU, vertRadius, 0, 32, 1.0)))

		if inpututil.IsKeyJustPressed(ebiten.KeyA) {
			motionAnim = !motionAnim
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyT) {
			ctx.Renderer.SetTint(1.0 - ctx.Renderer.Tint())
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyI) {
			intensityAnim = !intensityAnim
		}
	}
	drawer := func(canvas *ebiten.Image, ctx TestAppCtx) {
		cwi, chi := rectSize(canvas.Bounds())
		cw, ch := rectSizeF32(canvas.Bounds())
		Paint(canvas, image.Rect(0, 0, cwi, chi/2), [4]float32{0, 0, 0, 1}, ebiten.BlendCopy)
		Paint(canvas, image.Rect(0, chi/2, cwi, chi), RGBAF32(backColors[backColorIdx]), ebiten.BlendCopy)
		ctx.Renderer.SetColorF32(1.0, 1.0, 1.0, 1.0)

		info := fmt.Sprintf(
			"Back color: %v [B]\nRenderer color: %v [R]\nMode: %s [M]\nHorz/Vert Radius: %.02f / %.02f [H / U]\nTint: %.02f [T]\nMotionAnim: %t [A]\nIntensityAnim: %t [I]",
			backColors[backColorIdx], rendererColors[rendererClrIdx], modes[mode], horzRadius, vertRadius, ctx.Renderer.Tint(), motionAnim, intensityAnim,
		)
		ctx.Renderer.Text(canvas, info, 8, 8, TextOpts(1.0, TopLeft.Snap(CapLine)))

		x0, y0 := CTR.Adjust(ctx.Images[0], cw*0.25, ch*0.25)
		x1, y1 := CTR.Adjust(ctx.Images[1], cw*0.75, ch*0.25)
		x2, y2 := CTR.Adjust(ctx.Images[2], cw*0.25, ch*0.75)
		x3, y3 := CTR.Adjust(ctx.Images[3], cw*0.75, ch*0.75)
		hRadius, vRadius := horzRadius, vertRadius
		if motionAnim {
			xAnim, yAnim := float32(ctx.DistAnim(3.0, 1.0)), float32(ctx.DistAnim(3.0, 0.666))
			x0, x1, x2, x3 = x0+xAnim, x1+xAnim, x2+xAnim, x3+xAnim
			y0, y1, y2, y3 = y0+yAnim, y1+yAnim, y2+yAnim, y3+yAnim
			hRadius -= float32(ctx.DistAnim(1.0, 0.6))
			vRadius -= float32(ctx.DistAnim(1.0, 0.8))
		}

		if ctx.SpacePressed {
			ctx.Renderer.SetColor(rendererColors[rendererClrIdx])
			ctx.Renderer.Blur2(canvas, ctx.Images[0], x0, y0, hRadius, vRadius)
			ctx.Renderer.Blur2(canvas, ctx.Images[1], x1, y1, hRadius, vRadius)
			ctx.Renderer.Blur2(canvas, ctx.Images[2], x2, y2, hRadius, vRadius)
			ctx.Renderer.Blur2(canvas, ctx.Images[3], x3, y3, hRadius, vRadius)
			return
		}

		ctx.DrawAtF32(canvas, ctx.Images[0], x0, y0)
		ctx.DrawAtF32(canvas, ctx.Images[1], x1, y1)
		ctx.DrawAtF32(canvas, ctx.Images[2], x2, y2)
		ctx.DrawAtF32(canvas, ctx.Images[3], x3, y3)
		ctx.Renderer.SetColor(rendererColors[rendererClrIdx])
		if intensityAnim {
			ctx.Renderer.ScaleAlphaBy(float32(ctx.DistAnim(1.0, 1.0)))
		}
		switch mode {
		case 1: // std
			ctx.Renderer.Glow2(canvas, ctx.Images[0], x0, y0, hRadius, vRadius, 0.4, 1.0)
			ctx.Renderer.Glow2(canvas, ctx.Images[1], x1, y1, hRadius, vRadius, 0.4, 1.0)
			ctx.Renderer.Glow2(canvas, ctx.Images[2], x2, y2, hRadius, vRadius, 0.4, 1.0)
			ctx.Renderer.Glow2(canvas, ctx.Images[3], x3, y3, hRadius, vRadius, 0.4, 1.0)
		case 2: // color
			rgb := [3]float32{1, 0, 1}
			ctx.Renderer.GlowColor2(canvas, ctx.Images[0], x0, y0, hRadius, vRadius, rgb, 0.4, 1.0)
			ctx.Renderer.GlowColor2(canvas, ctx.Images[1], x1, y1, hRadius, vRadius, rgb, 0.4, 1.0)
			ctx.Renderer.GlowColor2(canvas, ctx.Images[2], x2, y2, hRadius, vRadius, rgb, 0.4, 1.0)
			ctx.Renderer.GlowColor2(canvas, ctx.Images[3], x3, y3, hRadius, vRadius, rgb, 0.4, 1.0)
		case 3: // dark
			if ebiten.IsKeyPressed(ebiten.KeyD) {
				ctx.Renderer.Options().Blend = BlendSubtract
			}
			if ebiten.IsKeyPressed(ebiten.KeyF) {
				ctx.Renderer.Options().Blend = BlendMultiply
			}
			ctx.Renderer.GlowDark2(canvas, ctx.Images[0], x0, y0, hRadius, vRadius, 0.5, 0.25)
			ctx.Renderer.GlowDark2(canvas, ctx.Images[1], x1, y1, hRadius, vRadius, 0.5, 0.25)
			ctx.Renderer.GlowDark2(canvas, ctx.Images[2], x2, y2, hRadius, vRadius, 0.5, 0.25)
			ctx.Renderer.GlowDark2(canvas, ctx.Images[3], x3, y3, hRadius, vRadius, 0.5, 0.25)
			ctx.Renderer.Options().Blend = ebiten.BlendSourceOver
		}
	}

	app := NewTestApp(updater, drawer)
	circ := app.Renderer.NewFilledCircle(96.0)
	darkCirc := app.Renderer.NewFilledCircle(96.0)
	app.Renderer.Options().Blend = ebiten.BlendSourceIn
	gradientOpts := GradientOpts(color.RGBA{0, 255, 255, 255}, color.RGBA{255, 0, 255, 255}, false)
	app.Renderer.Gradient(circ, gradientOpts, DirRadsLTR)
	gradientOpts = GradientOpts(color.RGBA{255, 128, 64, 255}, color.RGBA{0, 0, 0, 255}, false)
	app.Renderer.Gradient(darkCirc, gradientOpts, DirRadsTLBR)
	app.Renderer.Options().Blend = ebiten.BlendSourceOver

	cross := ebiten.NewImage(128, 128)
	darkCross := ebiten.NewImage(128, 128)
	app.Renderer.SetColorF32(1.0, 0.75, 0.5, 1.0, 0, 1)
	app.Renderer.StrokeLine(cross, 128/2.0, 8.0, 128/2.0, 128.0-8.0, 6.0)
	swapDiagColors(app.Renderer)
	app.Renderer.StrokeLine(cross, 8.0, 128/2.0, 128.0-8.0, 128/2.0, 6.0)
	app.Renderer.SetColorF32(0.0, 0.0, 0.0, 1.0, 0, 1)
	app.Renderer.SetColorF32(1.0, 0.0, 1.0, 1.0, 2, 3)
	app.Renderer.StrokeLine(darkCross, 128/2.0, 8.0, 128/2.0, 128.0-8.0, 6.0)
	swapDiagColors(app.Renderer)
	app.Renderer.StrokeLine(darkCross, 8.0, 128/2.0, 128.0-8.0, 128/2.0, 6.0)
	app.Renderer.SetColorF32(1, 1, 1, 1)
	app.Images = append(app.Images, circ, cross, darkCross, darkCirc)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

func swapDiagColors(r *Renderer) {
	r1, g1, b1, a1 := r.vertices[1].ColorR, r.vertices[1].ColorG, r.vertices[1].ColorB, r.vertices[1].ColorA
	r3, g3, b3, a3 := r.vertices[3].ColorR, r.vertices[3].ColorG, r.vertices[3].ColorB, r.vertices[3].ColorA
	r.vertices[1].ColorR, r.vertices[1].ColorG, r.vertices[1].ColorB, r.vertices[1].ColorA = r3, g3, b3, a3
	r.vertices[3].ColorR, r.vertices[3].ColorG, r.vertices[3].ColorB, r.vertices[3].ColorA = r1, g1, b1, a1
}
