package shapes

import (
	"fmt"
	"image/color"
	"math"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

// go test -run ^TestStrokeArc$ . -count 1
func TestStrokeArc(t *testing.T) {
	var startRads, endRads float64 = 0.2, RadsBottomRight
	var thickness, radius float32 = 16.0, 96.0

	updater := func(ctx TestAppCtx) {
		ebiten.SetWindowTitle(fmt.Sprintf(
			"%s  [startRads: %.02f (S), endRads: %.02f (E), thickness: %.02f (T), radius: %.02f (R)]",
			ctx.Title(), startRads, endRads, thickness, radius,
		))
		startRads = updateParam(ctx, ebiten.KeyS, startRads, 0, 2*math.Pi, math.Pi/6.0)
		endRads = updateParam(ctx, ebiten.KeyE, endRads, 0, 2*math.Pi, math.Pi/6.0)
		thickness = updateParam(ctx, ebiten.KeyT, thickness, 0, 32.0, 2.0)
		radius = updateParam(ctx, ebiten.KeyR, radius, 0, 256.0, 16.0)
	}
	drawer := func(canvas *ebiten.Image, ctx TestAppCtx) {
		w, h := rectSizeF32(canvas.Bounds())
		cx, cy := w/2.0, h/2.0
		ctx.Renderer.SetColorF32(1.0, 1.0, 1.0, 1.0)
		ctx.Renderer.StrokeArc(canvas, float64(cx), float64(cy), float64(radius), startRads, endRads, float64(thickness))

		ctx.Renderer.SetColorF32(0.5, 0.0, 0.5, 0.5)
		ctx.Renderer.FillRadialSector(canvas, float32(cx), float32(cy), 0, float32(radius), startRads, endRads, 0)                  // reference
		ctx.Renderer.StrokeLine(canvas, PtF32(16+thickness, 16+thickness), PtF32(16+thickness, 16+max(32, thickness*4)), thickness) // for thickness
		ctx.Renderer.FillCircle(canvas, cx-radius, cy, thickness)                                                                   // for thickness
	}

	app := NewTestApp(updater, drawer)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestStrokeCircle$ . -count 1
func TestStrokeCircle(t *testing.T) {
	updater := func(ctx TestAppCtx) {}
	drawer := func(canvas *ebiten.Image, ctx TestAppCtx) {
		const Radius = 72

		lc := ctx.LeftClickF32()
		rc := ctx.RightClickF32()

		ctx.Renderer.SetColorF32(1.0, 1.0, 1.0, 1.0)
		ctx.Renderer.ScaleAlphaBy(0.666)
		ctx.Renderer.FillCircle(canvas, lc.X, lc.Y, Radius)
		ctx.Renderer.FillCircle(canvas, rc.X, rc.Y, Radius)

		const MaxThickness = 16
		thick := float32(ctx.DistAnim(MaxThickness, 1.0))
		ctx.Renderer.ScaleAlphaBy(0.666)
		ctx.Renderer.FillCircle(canvas, lc.X, lc.Y, Radius-MaxThickness)
		ctx.Renderer.FillCircle(canvas, rc.X, rc.Y, Radius-MaxThickness)

		ctx.Renderer.SetColorF32(1.0, 0.0, 1.0, 1.0)
		ctx.Renderer.ScaleAlphaBy(0.666)
		ctx.Renderer.StrokeCircle(canvas, lc.X, lc.Y, Radius, thick)
		ctx.Renderer.StrokeCircle(canvas, rc.X, rc.Y, Radius, -thick)

		thick2 := float32(ctx.DistAnim(32.0, 1.0))
		ctx.Renderer.StrokeCircle(canvas, lc.X, rc.Y, 16, -thick2)
		ctx.Renderer.StrokeCircle(canvas, rc.X, lc.Y, thick2-8.0, 16.0)
	}

	app := NewTestApp(updater, drawer)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestFillEllipse$ . -count 1
func TestFillEllipse(t *testing.T) {
	updater := func(ctx TestAppCtx) {}
	drawer := func(canvas *ebiten.Image, ctx TestAppCtx) {
		lc := ctx.LeftClickF32()
		rc := ctx.RightClickF32()

		ctx.Renderer.SetColorF32(0.5, 0.5, 0.5, 0.5)
		ctx.Renderer.FillCircle(canvas, lc.X, lc.Y, 64.0)
		ctx.Renderer.FillCircle(canvas, rc.X, rc.Y, 32.0)

		ctx.Renderer.SetColorF32(1.0, 1.0, 1.0, 1.0)
		ctx.Renderer.FillEllipse(canvas, lc.X, lc.Y, 24.0, 64.0, ctx.RadsAnim(1.0))
		ctx.Renderer.FillEllipse(canvas, rc.X, rc.Y, 32.0, 16.0, 0)

		ctx.Renderer.SetColorF32(0.0, 0.5, 0.5, 0.5)
		ctx.Renderer.FillCircle(canvas, lc.X, lc.Y, 24.0)
		ctx.Renderer.FillCircle(canvas, rc.X, rc.Y, 16.0)
	}

	app := NewTestApp(updater, drawer)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestFillCircularSectorInner$ . -count 1
func TestFillCircularSectorInner(t *testing.T) {
	var animRounding, animSpin, debugBounds bool
	updater := func(ctx TestAppCtx) {
		animRounding = updateToggle(ctx, ebiten.KeyR, animRounding)
		animSpin = updateToggle(ctx, ebiten.KeyS, animSpin)
		debugBounds = updateToggle(ctx, ebiten.KeyB, debugBounds)
	}
	drawer := func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Fill(color.RGBA{0, 0, 128, 255})
		w, h := rectSizeF32(canvas.Bounds())
		cx, cy := w/2.0, h/2.0

		startRads, endRads := -math.Pi/6, math.Pi/6
		radius := 120.0
		rounding := -8.0
		if animRounding {
			rounding -= ctx.DistAnim(48.0, 1.0)
		}
		if animSpin {
			shift := math.Pi/2.0 - ctx.DistAnim(math.Pi, 0.666)
			startRads = uradsAddCW(startRads, shift)
			endRads = uradsAddCW(endRads, shift)
		}
		ctx.Renderer.SetColorF32(1.0, 1.0, 1.0, 1.0)
		if debugBounds {
			ctx.Renderer.Options().Blend = ebiten.BlendCopy
		}
		ctx.Renderer.fillCircularSector(canvas, cx, cy, radius, startRads, endRads, rounding)
		ctx.Renderer.Options().Blend = ebiten.BlendSourceOver
		ctx.Renderer.SetColorF32(0.0, 0.5, 0.5, 0.5)
		ctx.Renderer.FillRadialSector(canvas, cx, cy, 0, float32(radius), startRads, endRads, 0.0)
	}

	app := NewTestApp(updater, drawer)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestFillRadialSector$ . -count 1
func TestFillRadialSector(t *testing.T) {
	updater := func(ctx TestAppCtx) {}
	drawer := func(canvas *ebiten.Image, ctx TestAppCtx) {
		w, h := rectSizeF32(canvas.Bounds())
		cx, cy := w/2.0, h/2.0
		startRads := ctx.ModAnim(2*math.Pi, 1.0)
		endRads := startRads + 0.4 + ctx.DistAnim(1.6, 1.0)
		inRadius, outRadius := float32(48.0), float32(128.0)
		if ctx.SpacePressed {
			inRadius = 0.0
		}

		ctx.Renderer.SetColorF32(1.0, 1.0, 1.0, 1.0)
		ctx.Renderer.FillRadialSector(canvas, cx, cy, inRadius, outRadius, startRads, endRads, float32(-16.0+ctx.DistAnim(32.0, 1.0)))
		ctx.Renderer.SetColorF32(0.0, 0.5, 0.5, 0.5)
		ctx.Renderer.FillRadialSector(canvas, cx, cy, inRadius, outRadius, startRads, endRads, 0.0)
	}

	app := NewTestApp(updater, drawer)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestFillRadialSectorRounding$ . -count 1
func TestFillRadialSectorRounding(t *testing.T) {
	var startRads, aperture float64 = 0, math.Pi / 4.0
	var inRadius, outRadius float32 = 64.0, 128.0
	var rounding float32 = -16.0
	var animAperture, animRounding bool

	updater := func(ctx TestAppCtx) {
		const StartRadsChange, ApertureChange = math.Pi / 12.0, math.Pi / 32.0
		const RadiusChange = 16.0
		startRads = updateParam(ctx, ebiten.KeyS, startRads, 0, 2.0*math.Pi, StartRadsChange)
		aperture = updateParam(ctx, ebiten.KeyA, aperture, 0, 2.0*math.Pi, ApertureChange)
		inRadius = updateParam(ctx, ebiten.KeyQ, inRadius, 0, outRadius, RadiusChange)
		outRadius = updateParam(ctx, ebiten.KeyW, outRadius, inRadius, 384.0, RadiusChange)
		roundingChange := float32(4.0)
		if (rounding > 0 && rounding < 24) || rounding < -16 {
			roundingChange = float32(2.0)
		}
		if (rounding > 0 && rounding < 10) || rounding < -30 {
			roundingChange = float32(1.0)
		}
		rounding = updateParam(ctx, ebiten.KeyR, rounding, -48.0, 48.0, roundingChange)
		animAperture = updateToggle(ctx, ebiten.KeyE, animAperture)
		animRounding = updateToggle(ctx, ebiten.KeyT, animRounding)
	}
	drawer := func(canvas *ebiten.Image, ctx TestAppCtx) {
		ctx.Renderer.SetColorF32(1.0, 1.0, 1.0, 1.0)
		info := fmt.Sprintf(
			"startRads: %.02f [S]\naperture: %.02f [A]\nin/out radius: %.02f / %.02f [Q / W]\nrounding: %.02f [R]\nAnim aperture/rounding: %t, %t [E / T]",
			startRads, aperture, inRadius, outRadius, rounding, animAperture, animRounding,
		)
		ctx.Renderer.Text(canvas, info, 8, 8, TextOpts(1.0, TopLeft.Snap(CapLine)))

		var apertureAnim float64
		if animAperture {
			apertureAnim = -math.Pi/8.0 + ctx.DistAnim(math.Pi/4.0, 1.0)
		}

		var roundingAnim float32
		if animRounding {
			roundingAnim = float32(-8.0 + ctx.DistAnim(16.0, 1.0))
		}

		w, h := rectSizeF32(canvas.Bounds())
		cx, cy := w/2.0, h/2.0

		endRads := uradsAddCW(startRads, max(aperture+apertureAnim, 0))
		if ctx.SpacePressed {
			ctx.Renderer.SetColorF32(0.5, 0.5, 0.5, 0.5)
			ctx.Renderer.FillRadialSector(canvas, cx, cy, inRadius, outRadius, startRads, endRads, 0)
		} else {
			ctx.Renderer.SetColorF32(1.0, 1.0, 1.0, 1.0)
			ctx.Renderer.FillRadialSector(canvas, cx, cy, inRadius, outRadius, startRads, endRads, rounding+roundingAnim)
		}
	}

	app := NewTestApp(updater, drawer)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestFillRadialWedge$ . -count 1
func TestFillRadialWedge(t *testing.T) {
	var inRate, outRate float64 = 0.2, 0.35
	var rotate bool = true
	var showBounds bool

	updater := func(ctx TestAppCtx) {
		rotate = updateToggle(ctx, ebiten.KeyR, rotate)
		showBounds = updateToggle(ctx, ebiten.KeyB, showBounds)
		inRate = updateParam(ctx, ebiten.KeyI, inRate, 0.0, 1.0, 0.05)
		outRate = updateParam(ctx, ebiten.KeyO, outRate, 0.0, 1.0, 0.05)
	}
	drawer := func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Fill(backTestColor)

		info := fmt.Sprintf(
			"In/OutRate: %.02f / %.02f [I/O]\nRotate: %t [R]\nShow bounds: %t [B]",
			inRate, outRate, rotate, showBounds,
		)
		ctx.Renderer.SetColorF32(1, 1, 1, 1)
		ctx.Renderer.Text(canvas, info, 12, 12, TextOpts(1.0, TopLeft.Snap(CapLine)))

		w, h := rectSizeF32(canvas.Bounds())
		cx, cy := w/2.0, h/2.0
		inRadius, outRadius := 128.0, 192.0

		// draw guidelines
		ctx.Renderer.SetColorF32(0.4, 0.4, 0.4, 1.0)
		ctx.Renderer.StrokeCircle(canvas, cx, cy, float32(inRadius), 3.0)
		ctx.Renderer.StrokeCircle(canvas, cx, cy, float32(outRadius), 3.0)
		centerDir := ctx.ModAnim(2*math.Pi, 0.5)
		if !rotate {
			centerDir = 0.0
		}
		sin, cos := math.Sincos(centerDir)
		r := outRadius + 32.0
		ctx.Renderer.StrokeLine(canvas, PtF32(cx, cy), PtF32(cx+float32(r*cos), cy+float32(r*sin)), 3.0)

		// draw wedge
		ctx.Renderer.SetColorF32(1.0, 1.0, 1.0, 1.0)
		rounding := -16.0 + ctx.DistAnim(32.0, 1.0)
		if showBounds {
			ctx.Renderer.Options().Blend = ebiten.BlendCopy
		}
		ctx.Renderer.fillRadialWedge(canvas, float64(cx), float64(cy), inRadius, outRadius, centerDir, inRate*2*math.Pi, outRate*2*math.Pi, rounding)
		ctx.Renderer.Options().Blend = ebiten.BlendSourceOver
	}

	app := NewTestApp(updater, drawer)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}
