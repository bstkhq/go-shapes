package shapes

import (
	"fmt"
	"math"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// go test -run ^TestStrokeCircArc$ . -count 1
func TestStrokeCircArc(t *testing.T) {
	var startRads, endRads float64 = 0.2, RadsBottomRight
	var thickness, radius float64 = 16.0, 96.0

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
		w, h := rectSizeF64(canvas.Bounds())
		cx, cy := w/2.0, h/2.0
		ctx.Renderer.SetColorF32(1.0, 1.0, 1.0, 1.0)
		ctx.Renderer.StrokeCircArc(canvas, cx, cy, radius, startRads, endRads, thickness)

		ctx.Renderer.SetColorF32(0.5, 0.0, 0.5, 0.5)
		ctx.Renderer.FillCircSector(canvas, float32(cx), float32(cy), 0, float32(radius), startRads, endRads, 0)      // reference
		ctx.Renderer.StrokeLine(canvas, 16+thickness, 16+thickness, 16+thickness, 16+max(32, thickness*4), thickness) // for thickness
		ctx.Renderer.FillCircle(canvas, float32(cx-radius), float32(cy), float32(thickness))                          // for thickness
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

// go test -run ^TestDrawEllipse$ . -count 1
func TestDrawEllipse(t *testing.T) {
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

// go test -run ^TestFillCircSector$ . -count 1
func TestFillCircSector(t *testing.T) {
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
		ctx.Renderer.FillCircSector(canvas, cx, cy, inRadius, outRadius, startRads, endRads, float32(-16.0+ctx.DistAnim(32.0, 1.0)))
		ctx.Renderer.SetColorF32(0.0, 0.5, 0.5, 0.5)
		ctx.Renderer.FillCircSector(canvas, cx, cy, inRadius, outRadius, startRads, endRads, 0.0)
	}

	app := NewTestApp(updater, drawer)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestFillCircSectorRounding$ . -count 1
func TestFillCircSectorRounding(t *testing.T) {
	var startRads, aperture float64 = 0, math.Pi / 4.0
	var inRadius, outRadius float32 = 64.0, 128.0
	var rounding float32 = -16.0

	updater := func(ctx TestAppCtx) {
		ebiten.SetWindowTitle(fmt.Sprintf(
			"%s  [startRads: %.02f (S), aperture: %.02f (A), inRadius: %.02f (Q), outRadius: %.02f (W), rounding: %.02f (R)]",
			ctx.Title(), startRads, aperture, inRadius, outRadius, rounding,
		))

		const StartRadsChange, ApertureChange = math.Pi / 12.0, math.Pi / 16.0
		const RadiusChange, RoundingChange = 16.0, 4.0
		startRads = updateParam(ctx, ebiten.KeyS, startRads, 0, 2.0*math.Pi, StartRadsChange)
		aperture = updateParam(ctx, ebiten.KeyA, aperture, 0, 2.0*math.Pi, ApertureChange)
		inRadius = updateParam(ctx, ebiten.KeyQ, inRadius, 0, outRadius, RadiusChange)
		outRadius = updateParam(ctx, ebiten.KeyW, outRadius, inRadius, 384.0, RadiusChange)
		rounding = updateParam(ctx, ebiten.KeyR, rounding, -48.0, 48.0, RoundingChange)
	}
	drawer := func(canvas *ebiten.Image, ctx TestAppCtx) {
		w, h := rectSizeF32(canvas.Bounds())
		cx, cy := w/2.0, h/2.0

		if ctx.SpacePressed {
			ctx.Renderer.SetColorF32(0.5, 0.5, 0.5, 0.5)
			ctx.Renderer.FillCircSector(canvas, cx, cy, inRadius, outRadius, startRads, uradsAddCW(startRads, aperture), 0)
		} else {
			ctx.Renderer.SetColorF32(1.0, 1.0, 1.0, 1.0)
			ctx.Renderer.FillCircSector(canvas, cx, cy, inRadius, outRadius, startRads, uradsAddCW(startRads, aperture), rounding)
		}
	}

	app := NewTestApp(updater, drawer)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestStrokeCircSector$ . -count 1
func TestStrokeCircSector(t *testing.T) {
	updater := func(ctx TestAppCtx) {}
	drawer := func(canvas *ebiten.Image, ctx TestAppCtx) {
		w, h := rectSizeF32(canvas.Bounds())
		cx, cy := w/2.0, h/2.0
		startRads := ctx.ModAnim(2*math.Pi, 1.0)
		endRads := startRads + 0.4 + ctx.DistAnim(1.6, 1.0)

		ctx.Renderer.SetColorF32(0, 0, 0.5, 0.5)
		inRadius, outRadius := float32(48.0), float32(128.0)
		if ctx.SpacePressed {
			inRadius = 0.0
		}
		ctx.Renderer.FillCircSector(canvas, cx, cy, inRadius, outRadius, startRads, endRads, 0.0)

		ctx.Renderer.SetColorF32(1.0, 1.0, 1.0, 1.0)
		thick := float32(ctx.DistAnim(8.0, 1.0))
		ctx.Renderer.StrokeCircSector(canvas, cx, cy, inRadius, outRadius, thick, startRads, endRads, 0.0)
		ctx.Renderer.SetColorF32(0.0, 0.5, 0.5, 0.5)
		ctx.Renderer.StrokeCircSector(canvas, cx, cy, inRadius, outRadius, thick, startRads, endRads, 8.0)
	}

	app := NewTestApp(updater, drawer)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestFillCircWedge$ . -count 1
func TestFillCircWedge(t *testing.T) {
	var inRate, outRate float64 = 0.2, 0.35
	var rotate bool = true

	updater := func(ctx TestAppCtx) {
		ebiten.SetWindowTitle(fmt.Sprintf("%s  [inRate: %.02f (shift + up/down), outRate: %.02f (up/down), rotate: %t (R)]", ctx.Title(), inRate, outRate, rotate))
		if inpututil.IsKeyJustPressed(ebiten.KeyR) {
			rotate = !rotate
		}

		const Change = 0.05
		change := 0.0
		if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) {
			change = Change
		} else if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) {
			change = -Change
		}

		if ebiten.IsKeyPressed(ebiten.KeyShift) {
			inRate = wrap(inRate+change, 0.0, 1.0)
		} else {
			outRate = wrap(outRate+change, 0.0, 1.0)
		}
	}
	drawer := func(canvas *ebiten.Image, ctx TestAppCtx) {
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
		ctx.Renderer.StrokeLine(canvas, float64(cx), float64(cy), float64(cx)+r*cos, float64(cy)+r*sin, 3.0)

		// draw wedge
		ctx.Renderer.SetColorF32(1.0, 1.0, 1.0, 1.0)
		rounding := -16.0 + ctx.DistAnim(32.0, 1.0)
		ctx.Renderer.fillCircWedge(canvas, float64(cx), float64(cy), inRadius, outRadius, centerDir, inRate*2*math.Pi, outRate*2*math.Pi, rounding)
	}

	app := NewTestApp(updater, drawer)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}
