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

		lc := ctx.LeftClickF32()
		ctx.DrawAtF32(canvas, ctx.Images[1], lc.X, lc.Y)
		ctx.Renderer.SetColorF32(1, 1, 1, 1)
		if ebiten.IsKeyPressed(ebiten.KeyAlt) {
			ctx.Renderer.Blur2(canvas, ctx.Images[1], lc.X, lc.Y, 16, 16)
		} else {
			ctx.Renderer.Glow2(canvas, ctx.Images[1], lc.X, lc.Y, 16, 16, 0.4, 0.7)
		}

		rc := ctx.RightClickF32()
		ctx.DrawAtF32(canvas, ctx.Images[0], rc.X, rc.Y)
		ctx.Renderer.SetColor(color.RGBA{255, 192, 192, 255})
		dynRadius := float32(ctx.DistAnim(6, 2.0))
		vertRadius := float32(16.0)
		if ctx.SpacePressed {
			vertRadius = float32(ctx.DistAnim(16.0, 0.5))
		}
		if ebiten.IsKeyPressed(ebiten.KeyAlt) {
			ctx.Renderer.Blur2(canvas, ctx.Images[0], rc.X, rc.Y, 10+dynRadius, vertRadius)
		} else {
			ctx.Renderer.Glow2(canvas, ctx.Images[0], rc.X, rc.Y, 10+dynRadius, vertRadius, 0.5, 0.6)
		}
	}

	app := NewTestApp(updater, drawer)
	const s, m = 96, 16
	cross := ebiten.NewImage(s, s)
	app.Renderer.SetColor(color.RGBA{96, 240, 240, 255})
	app.Renderer.StrokeLine(cross, PtF32(m, m), PtF32(s-m, s-m), m/2)
	app.Renderer.StrokeLine(cross, PtF32(s-m, m), PtF32(m, s-m), m/2)
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

		lc := ctx.LeftClickF32()
		ctx.DrawAtF32(canvas, ctx.Images[0], lc.X, lc.Y)
		hkern, vkern := GaussK3, GaussK3
		gRad := func(kern GaussKernel) float32 {
			// NOTE: this still doesn't match because the edges are blurred/dilated
			// on downscaling too
			radius := float32(kern.Size()*4-1) / 2.0
			return min(radius, 16.0)
		}
		if ctx.SpacePressed {
			ctx.Renderer.Glow2(canvas, ctx.Images[0], lc.X, lc.Y, gRad(hkern), gRad(vkern), 0.2, 0.8)
		} else {
			kOpts := KernelOptions{Downscaling: DownscaleX4, HorzKernel: hkern, VertKernel: vkern}
			ctx.Renderer.GlowK(canvas, ctx.Images[0], lc.X, lc.Y, 0.2, 0.8, kOpts)
		}

		rc := ctx.RightClickF32()
		ctx.DrawAtF32(canvas, ctx.Images[0], rc.X, rc.Y)
		ctx.Renderer.SetColor(color.RGBA{255, 192, 192, 255})
		hkern, vkern = GaussK5, GaussK15
		ctx.Renderer.SetTint(1.0)
		if ctx.SpacePressed {
			ctx.Renderer.Glow2(canvas, ctx.Images[0], rc.X, rc.Y, gRad(hkern), gRad(vkern), 0.5, 0.6)
		} else {
			kOpts := KernelOptions{Downscaling: DownscaleX4, HorzKernel: hkern, VertKernel: vkern}
			ctx.Renderer.GlowK(canvas, ctx.Images[0], rc.X, rc.Y, 0.5, 0.6, kOpts)
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

		lc := ctx.LeftClickF32()
		ctx.DrawAtF32(canvas, ctx.Images[0], lc.X, lc.Y)
		if !ctx.SpacePressed {
			loThresh := 0.1 + ctx.DistAnim(0.4, 1.0)
			kOpts := KernelOpts(DownscaleX4, GaussK7)
			ctx.Renderer.GlowColorK(canvas, ctx.Images[0], lc.X, lc.Y, RGBF32(color.RGBA{255, 255, 0, 255}), float32(loThresh), 1.0, kOpts)
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

		o0 := CTR.AdjustXY(ctx.Images[0], cw*0.25, ch*0.25)
		o1 := CTR.AdjustXY(ctx.Images[1], cw*0.75, ch*0.25)
		o2 := CTR.AdjustXY(ctx.Images[2], cw*0.25, ch*0.75)
		o3 := CTR.AdjustXY(ctx.Images[3], cw*0.75, ch*0.75)
		hRadius, vRadius := horzRadius, vertRadius
		if motionAnim {
			animOffset := PtF32(float32(ctx.DistAnim(3.0, 1.0)), float32(ctx.DistAnim(3.0, 0.666)))
			o0 = o0.Add(animOffset)
			o1 = o1.Add(animOffset)
			o2 = o2.Add(animOffset)
			o3 = o3.Add(animOffset)
			hRadius = max(hRadius-float32(ctx.DistAnim(1.0, 0.6)), 0)
			vRadius = max(vRadius-float32(ctx.DistAnim(1.0, 0.8)), 0)
		}

		if ctx.SpacePressed {
			ctx.Renderer.SetColor(rendererColors[rendererClrIdx])
			ctx.Renderer.Blur2(canvas, ctx.Images[0], o0.X, o0.Y, hRadius, vRadius)
			ctx.Renderer.Blur2(canvas, ctx.Images[1], o1.X, o1.Y, hRadius, vRadius)
			ctx.Renderer.Blur2(canvas, ctx.Images[2], o2.X, o2.Y, hRadius, vRadius)
			ctx.Renderer.Blur2(canvas, ctx.Images[3], o3.X, o3.Y, hRadius, vRadius)
			return
		}

		ctx.DrawAtF32(canvas, ctx.Images[0], o0.X, o0.Y)
		ctx.DrawAtF32(canvas, ctx.Images[1], o0.X, o0.Y)
		ctx.DrawAtF32(canvas, ctx.Images[2], o0.X, o0.Y)
		ctx.DrawAtF32(canvas, ctx.Images[3], o0.X, o0.Y)
		ctx.Renderer.SetColor(rendererColors[rendererClrIdx])
		if intensityAnim {
			ctx.Renderer.ScaleAlphaBy(float32(ctx.DistAnim(1.0, 1.0)))
		}
		switch mode {
		case 1: // std
			ctx.Renderer.Glow2(canvas, ctx.Images[0], o0.X, o0.Y, hRadius, vRadius, 0.4, 1.0)
			ctx.Renderer.Glow2(canvas, ctx.Images[1], o1.X, o1.Y, hRadius, vRadius, 0.4, 1.0)
			ctx.Renderer.Glow2(canvas, ctx.Images[2], o2.X, o2.Y, hRadius, vRadius, 0.4, 1.0)
			ctx.Renderer.Glow2(canvas, ctx.Images[3], o3.X, o3.Y, hRadius, vRadius, 0.4, 1.0)
		case 2: // color
			rgb := [3]float32{1, 0, 1}
			ctx.Renderer.GlowColor2(canvas, ctx.Images[0], o0.X, o0.Y, hRadius, vRadius, rgb, 0.4, 1.0)
			ctx.Renderer.GlowColor2(canvas, ctx.Images[1], o1.X, o1.Y, hRadius, vRadius, rgb, 0.4, 1.0)
			ctx.Renderer.GlowColor2(canvas, ctx.Images[2], o2.X, o2.Y, hRadius, vRadius, rgb, 0.4, 1.0)
			ctx.Renderer.GlowColor2(canvas, ctx.Images[3], o3.X, o3.Y, hRadius, vRadius, rgb, 0.4, 1.0)
		case 3: // dark
			ctx.Renderer.GlowDark2(canvas, ctx.Images[0], o0.X, o0.Y, hRadius, vRadius, 0.5, 0.25)
			ctx.Renderer.GlowDark2(canvas, ctx.Images[1], o1.X, o1.Y, hRadius, vRadius, 0.5, 0.25)
			ctx.Renderer.GlowDark2(canvas, ctx.Images[2], o2.X, o2.Y, hRadius, vRadius, 0.5, 0.25)
			ctx.Renderer.GlowDark2(canvas, ctx.Images[3], o3.X, o3.Y, hRadius, vRadius, 0.5, 0.25)
		}
	}

	app := NewTestApp(updater, drawer)
	circ := app.Renderer.NewFilledCircle(96.0)
	paintShape(app.Renderer, circ, color.RGBA{0, 255, 255, 255}, color.RGBA{255, 0, 255, 255}, DirRadsLTR)
	darkCirc := app.Renderer.NewFilledCircle(96.0)
	paintShape(app.Renderer, darkCirc, color.RGBA{255, 128, 64, 255}, color.RGBA{0, 0, 0, 255}, DirRadsTLBR)

	cross := ebiten.NewImage(128, 128)
	darkCross := ebiten.NewImage(128, 128)
	app.Renderer.SetColorF32(1.0, 0.75, 0.5, 1.0, 0, 1)
	app.Renderer.StrokeLine(cross, PtF32(128/2.0, 8.0), PtF32(128/2.0, 128.0-8.0), 6.0)
	swapDiagColors(app.Renderer)
	app.Renderer.StrokeLine(cross, PtF32(8.0, 128/2.0), PtF32(128.0-8.0, 128/2.0), 6.0)
	app.Renderer.SetColorF32(0.0, 0.0, 0.0, 1.0, 0, 1)
	app.Renderer.SetColorF32(1.0, 0.0, 1.0, 1.0, 2, 3)
	app.Renderer.StrokeLine(darkCross, PtF32(128/2.0, 8.0), PtF32(128/2.0, 128.0-8.0), 6.0)
	swapDiagColors(app.Renderer)
	app.Renderer.StrokeLine(darkCross, PtF32(8.0, 128/2.0), PtF32(128.0-8.0, 128/2.0), 6.0)
	app.Renderer.SetColorF32(1, 1, 1, 1)

	app.Renderer.SetColorF32(1, 1, 1, 1)
	app.Images = append(app.Images, circ, cross, darkCross, darkCirc)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

func newCross(r *Renderer, size int) *ebiten.Image {
	cross := ebiten.NewImage(size, size)
	memo := r.memorizeColors()
	s := float32(size)
	th := s * 0.05
	r.StrokeLine(cross, PtF32(th*1.25, s/2.0), PtF32(s-th*1.25, s/2.0), th)
	r.StrokeLine(cross, PtF32(s/2.0, th*1.25), PtF32(s/2.0, s-th*1.25), th)
	r.setColors(memo)
	return cross
}

func swapDiagColors(r *Renderer) {
	r1, g1, b1, a1 := r.vertices[1].ColorR, r.vertices[1].ColorG, r.vertices[1].ColorB, r.vertices[1].ColorA
	r3, g3, b3, a3 := r.vertices[3].ColorR, r.vertices[3].ColorG, r.vertices[3].ColorB, r.vertices[3].ColorA
	r.vertices[1].ColorR, r.vertices[1].ColorG, r.vertices[1].ColorB, r.vertices[1].ColorA = r3, g3, b3, a3
	r.vertices[3].ColorR, r.vertices[3].ColorG, r.vertices[3].ColorB, r.vertices[3].ColorA = r1, g1, b1, a1
}

// go test -run ^TestGlowKCompare$ . -count 1
func TestGlowKCompare(t *testing.T) {
	modes := []string{"glow", "glow-color", "glow-dark", "none"}
	mode := 0
	downscaling := DownscaleNone
	horzKern, vertKern := GaussK11, GaussK11
	motionAnim := false
	intensityAnim := false

	updater := func(ctx TestAppCtx) {
		mode = updateParam(ctx, ebiten.KeyM, mode, 0, 3, 1)
		horzKern = updateParam(ctx, ebiten.KeyH, horzKern, GaussK3, GaussK31, 1)
		vertKern = updateParam(ctx, ebiten.KeyU, vertKern, GaussK3, GaussK31, 1)
		downscaling = updateParam(ctx, ebiten.KeyD, downscaling, DownscaleNone, DownscaleX16, 1)
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
		cw, ch := rectSizeF32(canvas.Bounds())
		ctx.Renderer.SetColorF32(0.0, 0.0, 0.0, 1.0, 0, 3)
		ctx.Renderer.SetColorF32(0.0, 0.5, 0.5, 0.5, 1, 2)
		ctx.Renderer.FillIntRect(canvas, canvas.Bounds(), 0)

		info := fmt.Sprintf(
			"\nMode: %s [M]\nDownscaling: x%d\nHorz/Vert Kernels: %d / %d [H / U]\nTint: %.02f [T]\nMotionAnim: %t [A]\nIntensityAnim: %t [I]",
			modes[mode], downscaling.Factor(), horzKern.Radius(), vertKern.Radius(), ctx.Renderer.Tint(), motionAnim, intensityAnim,
		)
		ctx.Renderer.SetColorF32(1, 1, 1, 1)
		ctx.Renderer.Text(canvas, info, 8, 8, TextOpts(1.0, TopLeft.Snap(CapLine)))

		o0 := CTR.AdjustXY(ctx.Images[0], cw*0.25, ch*0.25)
		o1 := CTR.AdjustXY(ctx.Images[1], cw*0.75, ch*0.25)
		o2 := CTR.AdjustXY(ctx.Images[2], cw*0.25, ch*0.75)
		o3 := CTR.AdjustXY(ctx.Images[3], cw*0.75, ch*0.75)
		if motionAnim {
			animOffset := PtF32(float32(ctx.DistAnim(3.0, 1.0)), float32(ctx.DistAnim(3.0, 0.666)))
			o0 = o0.Add(animOffset)
			o1 = o1.Add(animOffset)
			o2 = o2.Add(animOffset)
			o3 = o3.Add(animOffset)
		}

		ctx.DrawAtF32(canvas, ctx.Images[0], o0.X, o0.Y)
		ctx.DrawAtF32(canvas, ctx.Images[1], o1.X, o1.Y)
		ctx.DrawAtF32(canvas, ctx.Images[2], o2.X, o2.Y)
		ctx.DrawAtF32(canvas, ctx.Images[3], o3.X, o3.Y)
		ctx.Renderer.SetColorF32(0, 1, 1, 1)
		if intensityAnim {
			ctx.Renderer.ScaleAlphaBy(float32(ctx.DistAnim(1.0, 1.0)))
		}

		opts := KernelOpts(downscaling, horzKern)
		opts.VertKernel = vertKern
		switch mode {
		case 0: // std
			ctx.Renderer.GlowK(canvas, ctx.Images[0], o0.X, o0.Y, 0.5, 1.0, opts)
			ctx.Renderer.GlowK(canvas, ctx.Images[1], o1.X, o1.Y, 0.5, 1.0, opts)
			ctx.Renderer.GlowK(canvas, ctx.Images[2], o2.X, o2.Y, 0.5, 1.0, opts)
			ctx.Renderer.GlowK(canvas, ctx.Images[3], o3.X, o3.Y, 0.5, 1.0, opts)
		case 1: // color
			rgb := [3]float32{1, 0, 1}
			ctx.Renderer.GlowColorK(canvas, ctx.Images[0], o0.X, o0.Y, rgb, 0.4, 1.0, opts)
			ctx.Renderer.GlowColorK(canvas, ctx.Images[1], o1.X, o1.Y, rgb, 0.4, 1.0, opts)
			ctx.Renderer.GlowColorK(canvas, ctx.Images[2], o2.X, o2.Y, rgb, 0.4, 1.0, opts)
			ctx.Renderer.GlowColorK(canvas, ctx.Images[3], o3.X, o3.Y, rgb, 0.4, 1.0, opts)
		case 2: // dark
			ctx.Renderer.GlowDarkK(canvas, ctx.Images[0], o0.X, o0.Y, 0.5, 0.25, opts)
			ctx.Renderer.GlowDarkK(canvas, ctx.Images[1], o1.X, o1.Y, 0.5, 0.25, opts)
			ctx.Renderer.GlowDarkK(canvas, ctx.Images[2], o2.X, o2.Y, 0.5, 0.25, opts)
			ctx.Renderer.GlowDarkK(canvas, ctx.Images[3], o3.X, o3.Y, 0.5, 0.25, opts)
		}
	}

	app := NewTestApp(updater, drawer)
	circ := app.Renderer.NewFilledCircle(96.0)
	paintShape(app.Renderer, circ, color.RGBA{0, 255, 255, 255}, color.RGBA{255, 0, 255, 255}, DirRadsLTR)
	darkCross := newCross(app.Renderer, 128)
	paintShape(app.Renderer, darkCross, color.RGBA{0, 0, 0, 255}, color.RGBA{0, 192, 255, 255}, DirRadsLTR)
	cross := newCross(app.Renderer, 128)
	paintShape(app.Renderer, cross, color.RGBA{255, 255, 0, 255}, color.RGBA{255, 0, 0, 255}, DirRadsTTB)
	darkCirc := app.Renderer.NewFilledCircle(82.0)
	paintShape(app.Renderer, darkCirc, color.RGBA{255, 172, 64, 255}, color.RGBA{0, 0, 32, 255}, DirRadsTLBR)

	app.Images = append(app.Images, circ, cross, darkCross, darkCirc)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

func paintShape(r *Renderer, shape *ebiten.Image, from, to color.RGBA, dir float64) {
	memo := r.Options().Blend
	r.Options().Blend = ebiten.BlendSourceIn
	gradientOpts := GradientOpts(from, to, false)
	r.Gradient(shape, gradientOpts, dir)
	r.Options().Blend = memo
}
