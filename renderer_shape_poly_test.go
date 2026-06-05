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

// go test -run ^TestDrawShapes$ . -count 1
func TestDrawShapes(t *testing.T) {
	updater := func(TestAppCtx) {}
	drawer := func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Fill(color.Black)
		ctx.Renderer.SetColor(color.RGBA{255, 255, 255, 255})
		lcx, lcy := ctx.LeftClickF64()
		rcx, rcy := ctx.RightClickF64()
		ctx.Renderer.StrokeLine(canvas, lcx, lcy, rcx, rcy, 6.0)

		x, y := float64(160), float64(40)
		var points [3]PointF32
		points[0] = PointF32{X: float32(x), Y: float32(y)}
		points[1] = points[0].AddXY(30, 10)
		points[2] = points[0].AddXY(16, 50)
		ctx.Renderer.FillTriangle(canvas, points, 0)

		x, y = float64(80), float64(260)
		ctx.Renderer.SetColor(color.RGBA{240, 48, 48, 255})
		points[0] = PointF32{X: float32(x), Y: float32(y)}
		points[1] = points[0].AddXY(70, -20)
		points[2] = points[0].AddXY(114, 80)
		ctx.Renderer.FillTriangle(canvas, points, 0)
		ctx.Renderer.SetColor(color.RGBA{255, 255, 255, 255})
		ctx.Renderer.FillTriangle(canvas, points, -8)
		x, y = float64(200), float64(300)
		points[1] = PointF32{X: float32(x), Y: float32(y)}
		points[0] = points[1].AddXY(70, -20)
		points[2] = points[1].AddXY(114, 80)
		ctx.Renderer.StrokeTriangle(canvas, points, 4, -8)
		v := uint8(32 + ctx.DistAnim(196-32, 1.0))
		ctx.Renderer.SetColor(color.RGBA{v, 0, 0, v})
		ctx.Renderer.StrokeTriangle(canvas, points, -4, 0)

		rads := ctx.RadsAnim(1.0)
		ctx.Renderer.FillHexagon(canvas, 540, 80, 60, 0, float32(rads))

		rounding := float32(ctx.DistAnim(48.0, 1.0))
		ctx.Renderer.SetColorF32(0, 0.5, 1.0, 1.0)
		ctx.Renderer.FillHexagon(canvas, 100, 400, 60, rounding, float32(rads))
		ctx.Renderer.ScaleAlphaBy(0.5)
		ctx.Renderer.FillHexagon(canvas, 540, 80, 60, rounding, float32(rads))
	}

	app := NewTestApp(updater, drawer)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestStrokeTriangle$ . -count 1
func TestStrokeTriangle(t *testing.T) {
	var thick float32
	updater := func(TestAppCtx) {
		thick = 8.0
		if ebiten.IsKeyPressed(ebiten.KeyT) {
			thick = -thick
		}
	}
	drawer := func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Fill(color.Black)
		const a = 0.5

		cw, ch := rectSizeF32(canvas.Bounds())
		cw3, ch3 := cw/3.0, ch/3.0

		var points [3][3]PointF32
		for x := range 3 {
			points[x][0] = PointF32{X: 32.0 + float32(x)*cw3, Y: 32.0}
			points[x][1] = points[x][0].AddXY(cw3*0.7, ch3*0.2)
			points[x][2] = points[x][0].AddXY(cw3*0.35, ch3*0.6)
		}

		r := float32(ctx.DistAnim(24.0, 1.0))
		if ctx.SpacePressed {
			r *= 3.0
		}
		ctx.Renderer.SetColorF32(a, a, a, a)
		ctx.Renderer.FillTriangle(canvas, points[0], 0.0) // no rounding
		ctx.Renderer.FillTriangle(canvas, points[1], r)   // outer rounding
		ctx.Renderer.FillTriangle(canvas, points[2], -r)  // inner rounding
		ctx.Renderer.SetColorF32(a, 0, a, a)
		ctx.Renderer.StrokeTriangle(canvas, points[0], thick, 0.0)
		ctx.Renderer.StrokeTriangle(canvas, points[1], thick, r)
		ctx.Renderer.StrokeTriangle(canvas, points[2], thick, -r)
	}

	app := NewTestApp(updater, drawer)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestFillTriangles$ . -count 1
func TestFillTriangles(t *testing.T) {
	updater := func(ctx TestAppCtx) {}
	drawer := func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Fill(color.Black)

		cw, ch := rectSizeF32(canvas.Bounds())
		outRounding := float32(ctx.DistAnim(16.0, 1.0))
		inRounding := -float32(ctx.DistAnim(48.0, 1.0))

		var pointsL [3]PointF32
		pointsL[0] = PointF32{X: 24, Y: 24}
		pointsL[1] = pointsL[0].AddXY(108, 12)
		pointsL[2] = pointsL[0].AddXY(23, 48)
		var pointsM [3]PointF32
		pointsM[0] = PointF32{X: cw / 2, Y: 24}
		pointsM[1] = pointsM[0].AddXY(64, 24)
		pointsM[2] = pointsM[0].AddXY(-32, 56)

		// first row, filled
		ctx.Renderer.SetColorF32(1, 1, 1, 1)
		ctx.Renderer.FillTriangle(canvas, pointsL, outRounding)
		ctx.Renderer.SetColorF32(0.3, 0, 0.3, 0.3)
		ctx.Renderer.FillCircle(canvas, pointsL[0].X, pointsL[0].Y, outRounding)
		ctx.Renderer.FillCircle(canvas, pointsL[1].X, pointsL[1].Y, outRounding)
		ctx.Renderer.FillCircle(canvas, pointsL[2].X, pointsL[2].Y, outRounding)

		ctx.Renderer.SetColorF32(1, 1, 1, 1)
		ctx.Renderer.FillTriangle(canvas, pointsM, inRounding)
		ctx.Renderer.SetColorF32(0.3, 0, 0.3, 0.3)
		ctx.Renderer.StrokeTriangle(canvas, pointsM, -3.0, inRounding)

		// second row, balanced stroke
		const CVOffset = -24
		for i, p := range pointsL {
			pointsL[i] = p.AddXY(0, -24+ch/2+CVOffset)
		}
		ctx.Renderer.SetColorF32(1, 1, 1, 1)
		ctx.Renderer.StrokeTriangle(canvas, pointsL, 8.0, outRounding)
		ctx.Renderer.SetColorF32(0.3, 0, 0.3, 0.3)
		ctx.Renderer.FillCircle(canvas, pointsL[0].X, pointsL[0].Y, outRounding)
		ctx.Renderer.FillCircle(canvas, pointsL[1].X, pointsL[1].Y, outRounding)
		ctx.Renderer.FillCircle(canvas, pointsL[2].X, pointsL[2].Y, outRounding)

		for i, p := range pointsM {
			pointsM[i] = p.AddXY(0, -24+ch/2+CVOffset)
		}
		ctx.Renderer.SetColorF32(1, 1, 1, 1)
		ctx.Renderer.StrokeTriangle(canvas, pointsM, 8.0, inRounding)

		// third row, inner stroke
		const LVOffset = -60
		for i, p := range pointsL {
			pointsL[i] = p.AddXY(0, ch/2+LVOffset)
		}
		ctx.Renderer.SetColorF32(1, 1, 1, 1)
		ctx.Renderer.StrokeTriangle(canvas, pointsL, -8.0, outRounding)
		ctx.Renderer.SetColorF32(0.3, 0, 0.3, 0.3)
		ctx.Renderer.FillCircle(canvas, pointsL[0].X, pointsL[0].Y, outRounding)
		ctx.Renderer.FillCircle(canvas, pointsL[1].X, pointsL[1].Y, outRounding)
		ctx.Renderer.FillCircle(canvas, pointsL[2].X, pointsL[2].Y, outRounding)

		for i, p := range pointsM {
			pointsM[i] = p.AddXY(0, ch/2+LVOffset)
		}
		ctx.Renderer.SetColorF32(1, 1, 1, 1)
		ctx.Renderer.StrokeTriangle(canvas, pointsM, -8.0, inRounding)
	}

	app := NewTestApp(updater, drawer)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestFillHexagons$ . -count 1
func TestFillHexagons(t *testing.T) {
	const Pad, MinRadius, MaxRadius = 16, 48, 64
	const MinApothem = MinRadius * Sqrt3Div2
	manRounding := float32(0.0)

	updater := func(ctx TestAppCtx) {
		// manual control
		ebiten.SetWindowTitle(fmt.Sprintf("%s  [rounding: %.02f (up/down), apothem: %.02f]", ctx.Title(), manRounding, MinApothem))
		if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) {
			manRounding += 1.0
		} else if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) {
			manRounding -= 1.0
		}
	}
	drawer := func(canvas *ebiten.Image, ctx TestAppCtx) {
		cw, ch := rectSizeF32(canvas.Bounds())

		radius := MinRadius + float32(ctx.DistAnim(MaxRadius-MinRadius, 1.0))
		rads := float32(ctx.RadsAnim(1.0))

		// radius bounded hexagon, no roundness
		ctx.Renderer.SetColorF32(1.0, 1.0, 1.0, 1.0)
		ctx.Renderer.FillCircle(canvas, Pad+MaxRadius, Pad+MaxRadius, radius)
		ctx.Renderer.SetColorF32(1.0, 0.0, 1.0, 1.0)
		ctx.Renderer.ScaleAlphaBy(0.666)
		ctx.Renderer.FillHexagon(canvas, Pad+MaxRadius, Pad+MaxRadius, radius, 0.0, rads)

		// radius bounded hexagon with animated roundness
		roundness := float32(ctx.DistAnim(MaxRadius+16, 0.5))
		ctx.Renderer.SetColorF32(1.0, 1.0, 1.0, 1.0)
		ctx.Renderer.FillCircle(canvas, cw/2, Pad+MaxRadius, radius)
		ctx.Renderer.SetColorF32(0.0, 0.0, 1.0, 1.0)
		ctx.Renderer.ScaleAlphaBy(0.666)
		ctx.Renderer.FillCircle(canvas, cw/2, Pad+MaxRadius, roundness)
		ctx.Renderer.SetColorF32(1.0, 0.0, 1.0, 1.0)
		ctx.Renderer.ScaleAlphaBy(0.666)
		ctx.Renderer.FillHexagon(canvas, cw/2, Pad+MaxRadius, radius, roundness, rads)

		// apothem bounded hexagon, no rounding
		apothem := radius * Sqrt3Div2
		ctx.Renderer.SetColorF32(1.0, 1.0, 1.0, 1.0)
		ctx.Renderer.FillCircle(canvas, Pad+MaxRadius, ch/2, apothem)
		ctx.Renderer.SetColorF32(1.0, 0.0, 1.0, 1.0)
		ctx.Renderer.ScaleAlphaBy(0.666)
		ctx.Renderer.FillHexagonApothem(canvas, Pad+MaxRadius, ch/2, apothem, 0.0, rads)

		// apothem bounded hexagon, outwards rounding
		rounding := float32(ctx.DistAnim(24.0, 1.0))
		ctx.Renderer.SetColorF32(0.0, 0.0, 1.0, 1.0)
		ctx.Renderer.ScaleAlphaBy(0.666)
		ctx.Renderer.FillCircle(canvas, cw/2, ch/2, apothem+rounding)
		ctx.Renderer.SetColorF32(1.0, 1.0, 1.0, 1.0)
		ctx.Renderer.FillCircle(canvas, cw/2, ch/2, apothem)
		ctx.Renderer.SetColorF32(0.0, 0.0, 1.0, 1.0)
		ctx.Renderer.ScaleAlphaBy(0.666)
		ctx.Renderer.FillCircle(canvas, cw/2, ch/2, rounding)
		ctx.Renderer.SetColorF32(1.0, 0.0, 1.0, 1.0)
		ctx.Renderer.ScaleAlphaBy(0.666)
		ctx.Renderer.FillHexagonApothem(canvas, cw/2, ch/2, apothem, rounding, rads)

		// apothem bounded hexagon, inwards rounding
		inRounding := float32(ctx.DistAnim(MinApothem, 0.5))
		ctx.Renderer.SetColorF32(1.0, 1.0, 1.0, 1.0)
		ctx.Renderer.FillCircle(canvas, cw-16-MaxRadius, ch/2, apothem)
		ctx.Renderer.SetColorF32(0.0, 0.0, 1.0, 1.0)
		ctx.Renderer.ScaleAlphaBy(0.666)
		ctx.Renderer.FillCircle(canvas, cw-16-MaxRadius, ch/2, inRounding)
		ctx.Renderer.SetColorF32(1.0, 0.0, 1.0, 1.0)
		ctx.Renderer.ScaleAlphaBy(0.666)
		ctx.Renderer.FillHexagonApothem(canvas, cw-16-MaxRadius, ch/2, apothem, -inRounding, rads)

		ctx.Renderer.SetColorF32(1.0, 1.0, 1.0, 1.0)
		ctx.Renderer.FillCircle(canvas, 16+MaxRadius, ch-16-MaxRadius, MinApothem)
		ctx.Renderer.SetColorF32(0.0, 0.0, 1.0, 1.0)
		ctx.Renderer.ScaleAlphaBy(0.666)
		if manRounding < 0 {
			ctx.Renderer.FillCircle(canvas, 16+MaxRadius, ch-16-MaxRadius, -manRounding)
		} else {
			ctx.Renderer.FillCircle(canvas, 16+MaxRadius, ch-16-MaxRadius, MinApothem+manRounding)
		}
		ctx.Renderer.SetColorF32(1.0, 0.0, 1.0, 1.0)
		ctx.Renderer.ScaleAlphaBy(0.666)
		ctx.Renderer.FillHexagonApothem(canvas, 16+MaxRadius, ch-16-MaxRadius, MinApothem, manRounding, 0.0)
	}

	app := NewTestApp(updater, drawer)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestFillRect$ . -count 1
func TestFillRect(t *testing.T) {
	updater := func(ctx TestAppCtx) {}
	drawer := func(canvas *ebiten.Image, ctx TestAppCtx) {
		lc := ctx.LeftClickF32()
		rc := ctx.RightClickF32()

		ctx.Renderer.SetColorF32(1.0, 1.0, 1.0, 1.0)
		w1, h1 := float32(128), float32(48)
		w2, h2 := float32(48), float32(128)
		ctx.Renderer.FillRect(canvas, lc.X-w1/2, lc.Y-h1/2, w1, h1, -float32(ctx.DistAnim(float64(min(w1, h1))/2.0, 1.0)))
		ctx.Renderer.FillRect(canvas, rc.X-w2/2, rc.Y-h2/2, w2, h2, float32(ctx.DistAnim(float64(min(w1, h1))/2.0, 1.0)))

		ctx.Renderer.SetColorF32(0.2, 0.0, 0.2, 0.2)
		ctx.Renderer.FillCircle(canvas, lc.X, lc.Y, max(w1, h1)/2.0)
		ctx.Renderer.FillCircle(canvas, rc.X, rc.Y, max(w2, h2)/2.0)

		cw, ch := rectSizeF32(canvas.Bounds())
		ctx.Renderer.SetColorF32(0.5, 0.5, 0.5, 0.5)
		ctx.Renderer.FillRect(canvas, 16, ch-16, 128, -128, 0)
		ctx.Renderer.FillRect(canvas, 32, ch-32, 128-32, -(128 - 32), float32(ctx.DistAnim(16, 1.0)))

		ctx.Renderer.FillCircle(canvas, 164+32, ch-16-32, 32)
		ctx.Renderer.FillCircle(canvas, 164+128-32, ch-16-128+32, 32)
		ctx.Renderer.FillRect(canvas, 164, ch-16, 128, -128, -float32(ctx.DistAnim(32, 1.0)))

		collapseRounding := -float32(ctx.DistAnim(196.0, 0.5))
		const CRW, CRH = 128, 96
		ctx.Renderer.FillRect(canvas, cw-CRW-16, 16, CRW, CRH, collapseRounding)
		ctx.Renderer.SetColorF32(0.2, 0.0, 0.2, 0.2)
		ctx.Renderer.FillCircle(canvas, cw-CRW/2-16, 16+CRH/2, min(abs(collapseRounding), CRW/2, CRH/2))
	}

	app := NewTestApp(updater, drawer)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestFillRectPrecise$ . -count 1
func TestFillRectPrecise(t *testing.T) {
	updater := func(ctx TestAppCtx) {}
	drawer := func(canvas *ebiten.Image, ctx TestAppCtx) {
		ctx.Renderer.SetColorF32(1.0, 0.0, 1.0, 1.0)
		ctx.Renderer.FillIntRect(canvas, image.Rect(0, 0, 258, 258), 0)
		ctx.DrawAtF32(canvas, ctx.Images[0], 1, 1)
		ctx.DrawAtF32(canvas, ctx.Images[1], 2, 2)

		ctx.DrawAtF32(canvas, ctx.Images[2], 270, 1)
		ctx.DrawAtF32(canvas, ctx.Images[3], 271, 2)
	}

	app := NewTestApp(updater, drawer)
	box := ebiten.NewImage(256, 256)
	app.Renderer.FillRect(box, 1, 1, 254, 254, 0)
	box2 := ebiten.NewImage(254, 254)
	app.Renderer.SetColorF32(1.0, 0, 0, 1.0)
	app.Renderer.FillRect(box2, 1, 1, 252, 252, 0)

	box3 := ebiten.NewImage(256, 256)
	app.Renderer.SetColorF32(1.0, 1.0, 1.0, 1.0)
	app.Renderer.FillRect(box3, 1, 1, 254, 254, -6.0)
	box4 := ebiten.NewImage(254, 254)
	app.Renderer.SetColorF32(1.0, 0, 0, 1.0)
	app.Renderer.FillRect(box4, 1, 1, 252, 252, -6.0)

	app.Renderer.SetColorF32(1.0, 1.0, 1.0, 1.0)

	app.Images = append(app.Images, box, box2, box3, box4)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestStrokeIntRect$ . -count 1
func TestStrokeIntRect(t *testing.T) {
	updater := func(ctx TestAppCtx) {}
	drawer := func(canvas *ebiten.Image, ctx TestAppCtx) {
		lcx, lcy := ctx.LeftClick.X, ctx.LeftClick.Y
		ctx.Renderer.SetColor(color.RGBA{255, 255, 255, 255})
		ctx.Renderer.FillIntRect(canvas, RectWithSize(lcx, lcy, 200, 50), 0)
		ctx.Renderer.SetColor(color.RGBA{0, 255, 0, 255})
		ctx.Renderer.StrokeIntRect(canvas, RectWithSize(lcx-1, lcy-1, 200+2, 50+2), 1, 0)

		ctx.Renderer.SetColor(color.RGBA{0, 128, 0, 128})
		ctx.Renderer.StrokeIntRect(canvas, RectWithSize(lcx, lcy, 200, 50), 0, 1)

		rcx, rcy := ctx.RightClick.X, ctx.RightClick.Y
		ctx.Renderer.SetColor(color.RGBA{240, 0, 240, 255}, 0, 2)
		ctx.Renderer.StrokeIntRect(canvas, RectWithSize(rcx, rcy, 100, 50), 4, 4)

		ctx.Renderer.SetColor(color.RGBA{64, 128, 64, 128})
		ctx.Renderer.FillIntRect(canvas, RectWithSize(rcx, rcy, 100, 50), 0)

		ctx.Renderer.SetColor(color.RGBA{255, 0, 0, 255}, 0)
		ctx.Renderer.SetColor(color.RGBA{0, 255, 0, 255}, 1)
		ctx.Renderer.SetColor(color.RGBA{0, 0, 255, 255}, 2)
		ctx.Renderer.SetColor(color.RGBA{0, 255, 255, 255}, 3)
		ctx.Renderer.StrokeIntRect(canvas, RectWithSize(lcx, rcy, 80, 50), 8, 8)
	}

	app := NewTestApp(updater, drawer)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestStrokeRect$ . -count 1
func TestStrokeRect(t *testing.T) {
	updater := func(ctx TestAppCtx) {}
	drawer := func(canvas *ebiten.Image, ctx TestAppCtx) {
		lc := ctx.LeftClickF32()
		ctx.Renderer.SetColor(color.RGBA{255, 255, 255, 255})
		ctx.Renderer.FillRect(canvas, lc.X, lc.Y, 200, 50, 16)
		ctx.Renderer.SetColor(color.RGBA{0, 255, 0, 255})
		ctx.Renderer.StrokeRect(canvas, lc.X, lc.Y, 200, 50, 0, 2, 16)

		ctx.Renderer.SetColor(color.RGBA{128, 0, 0, 128})
		ctx.Renderer.StrokeRect(canvas, lc.X, lc.Y, 200, 50, 2, 0, 16)

		rc := ctx.RightClickF32()
		ctx.Renderer.SetColor(color.RGBA{240, 0, 240, 255}, 0, 2)
		ctx.Renderer.StrokeRect(canvas, rc.X, rc.Y, 100, 50, 4, 4, 25)

		a := uint8(ctx.DistAnim(144.0, 1.0))
		ctx.Renderer.SetColor(color.RGBA{a, a, a, a})
		ctx.Renderer.FillIntRect(canvas, RectWithSize(int(rc.X), int(rc.Y), 100, 50), 0)

		ctx.Renderer.SetColor(color.RGBA{255, 0, 0, 255}, 0)
		ctx.Renderer.SetColor(color.RGBA{0, 255, 0, 255}, 1)
		ctx.Renderer.SetColor(color.RGBA{0, 0, 255, 255}, 2)
		ctx.Renderer.SetColor(color.RGBA{0, 255, 255, 255}, 3)
		extra := float32(ctx.DistAnim(16, 1.0))
		subRounding := float32(ctx.DistAnim(20, 1.0))
		ctx.Renderer.StrokeRect(canvas, lc.X, rc.Y, 80+extra, 50, 8, 8, 25-subRounding)

		w, h := rectSizeF32(canvas.Bounds())
		ctx.Renderer.SetColor(color.RGBA{255, 255, 255, 255})
		ctx.Renderer.StrokeRect(canvas, w-16, 16, -200, 64, 0, 8, 32)
		ctx.Renderer.SetColor(color.RGBA{128, 0, 0, 128})
		ctx.Renderer.FillCircle(canvas, w-16-32+8, 16+32-8, 32)

		ctx.Renderer.SetColor(color.RGBA{255, 255, 255, 255})
		ctx.Renderer.StrokeRect(canvas, 16, h-16, 96, -64, 0, 8, -32.0)
		ctx.Renderer.SetColor(color.RGBA{128, 0, 0, 128})
		ctx.Renderer.FillCircle(canvas, 16+32, h-16-32, 32)
	}

	app := NewTestApp(updater, drawer)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestFillRectRounding$ . -count 1
func TestFillRectRounding(t *testing.T) {
	var inThick, outThick float32
	updater := func(ctx TestAppCtx) {
		inThick, outThick = float32(8.0), float32(8.0)
		if ebiten.IsKeyPressed(ebiten.KeyI) {
			inThick = 0.0
		}
		if ebiten.IsKeyPressed(ebiten.KeyP) {
			outThick = 0.0
		}
	}
	drawer := func(canvas *ebiten.Image, ctx TestAppCtx) {
		const rw, rh, ra = 196, 128, 0.5
		ctx.Renderer.SetColorF32(ra, ra, ra, ra)

		cw, ch := rectSizeF32(canvas.Bounds())
		rounding := float32(-16.0 + ctx.DistAnim(32.0, 1.0))
		ctx.Renderer.FillRect(canvas, cw/2-rw-16, ch/2-rh/2, rw, rh, rounding)
		ctx.Renderer.StrokeRect(canvas, cw/2+16, ch/2-rh/2, rw, rh, inThick, outThick, rounding)
	}

	app := NewTestApp(updater, drawer)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestFillQuad$ . -count 1
func TestFillQuad(t *testing.T) {
	regions := [][2]PointF32{ // origin and sizes
		{PtF32(0.0, 0.0), PtF32(0.333, 0.5)},
		{PtF32(0.333, 0.0), PtF32(0.333, 0.5)},
		{PtF32(0.666, 0.0), PtF32(0.333, 0.5)},
		{PtF32(0.0, 0.5), PtF32(0.333, 0.5)},
		{PtF32(0.333, 0.5), PtF32(0.333, 0.5)},
		{PtF32(0.666, 0.5), PtF32(0.333, 0.5)},
	}
	quads := [][4]PointF32{
		{PtF32(0.1, 0.1), PtF32(0.9, 0.9), PtF32(0.4, 0.6), PtF32(0.6, 0.4)}, // self-intersecting (doesn't have to render properly, but don't panic)
		{PtF32(0.5, 0.1), PtF32(0.8, 0.8), PtF32(0.5, 0.4), PtF32(0.2, 0.8)}, // arrow quad (concave)
		{PtF32(0.2, 0.1), PtF32(0.25, 0.2), PtF32(0.75, 0.75), PtF32(0.25, 0.9)},
		{PtF32(0.2, 0.9), PtF32(0.7, 0.8), PtF32(0.8, 0.2), PtF32(0.1, 0.1)}, // CCW quad
		{PtF32(0.1, 0.1), PtF32(0.9, 0.1), PtF32(0.1, 0.9), PtF32(0.9, 0.9)}, // symmetric bow-tie (self-intersecting)
		{PtF32(0.1, 0.4), PtF32(0.8, 0.4), PtF32(0.9, 0.6), PtF32(0.2, 0.6)}, // short parallelogram
	}
	if len(regions) != len(quads) {
		panic("each quad must have a region defined")
	}

	rounding := float32(0.0)
	animRounding := false
	updater := func(ctx TestAppCtx) {
		rounding = updateParam(ctx, ebiten.KeyR, rounding, -100.0, +100.0, 2.0)
		animRounding = updateToggle(ctx, ebiten.KeyA, animRounding)
	}
	drawer := func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Fill(backTestColor)
		ctx.Renderer.SetColorF32(1, 1, 1, 1)
		info := fmt.Sprintf("Rounding: %.02f [R]\nAnim Rounding: %t [A]", rounding, animRounding)
		ctx.Renderer.Text(canvas, info, 12, 12, TextOpts(1.0, TopLeft.Snap(CapLine)))

		if ctx.SpacePressed {
			ctx.Renderer.Options().Blend = ebiten.BlendClear
		}
		roundingOffset := float32(0.0)
		if animRounding {
			roundingOffset = -8.0 + float32(ctx.DistAnim(16.0, 1.0))
		}
		w, h := rectSizeF32(canvas.Bounds())
		size := PtF32(w, h)
		for i, region := range regions {
			quad := quads[i]
			for v := range 4 {
				quad[v] = (region[0].Mul(size)).Add(quad[v].Mul(size).Mul(region[1]))
			}
			ctx.Renderer.FillQuad(canvas, quad, rounding+roundingOffset)
		}
		ctx.Renderer.Options().Blend = ebiten.BlendSourceOver
	}

	app := NewTestApp(updater, drawer)
	app.Renderer.Warnings.SetHandler(func(warning Warning, value any, alreadySeen bool) {
		if warning != WarnSelfIntersectingGeom { // ignore self-intersections during visual testing
			panicHandlerFunc(warning, value, alreadySeen)
		}
	})
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestFillQuadSoft$ . -count 1
func TestFillQuadSoft(t *testing.T) {
	updater := func(ctx TestAppCtx) {}
	drawer := func(canvas *ebiten.Image, ctx TestAppCtx) {
		lc := ctx.LeftClickF32()
		rc := ctx.RightClickF32()

		w, h := rectSizeF32(canvas.Bounds())
		quad := [4]PointF32{
			{X: lc.X, Y: lc.Y},
			{X: w/2.0 + w/4.0, Y: h/2.0 - h/4.0},
			{X: rc.X, Y: rc.Y},
			{X: w/2.0 - w/4.0, Y: h/2.0 + h/4.0},
		}
		thickening := float32(ctx.DistAnim(48.0, 1.0))
		softEdge := float32(64.0)
		if ctx.SpacePressed {
			softEdge = 0
		}
		ctx.Renderer.FillQuadSoft(canvas, quad, thickening, softEdge)
	}

	app := NewTestApp(updater, drawer)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestFillRectSoft$ . -count 1
func TestFillRectSoft(t *testing.T) {
	const W, H = 128, 64
	roundingSign := 0
	softEdgeSign := 0

	updater := func(ctx TestAppCtx) {
		roundingSign = updateParam(ctx, ebiten.KeyR, roundingSign, -1, 1, 1)
		softEdgeSign = updateParam(ctx, ebiten.KeyS, softEdgeSign, -1, 1, 1)
	}
	drawer := func(canvas *ebiten.Image, ctx TestAppCtx) {
		cw, ch := rectSizeF32(canvas.Bounds())

		rounding := float32(roundingSign) * float32(ctx.DistAnim(16.0, 1.000))
		softEdge := float32(softEdgeSign) * float32(ctx.DistAnim(16.0, 0.666))

		ctx.Renderer.SetColorF32(1, 1, 1, 1)
		info := fmt.Sprintf(
			"Rounding: %.02f [R]\nSoft Edge: %.02f [S]",
			rounding, softEdge,
		)
		ctx.Renderer.Text(canvas, info, 8, 8, TextOpts(1.0, TopLeft.Snap(CapLine)))

		rect := image.Rect(0, 0, W, H)
		ol := CTR.AdjustXY(rect, cw*0.25, ch*0.25)
		or := CTR.AdjustXY(rect, cw*0.75, ch*0.25)
		ctx.Renderer.FillRectSoft(canvas, ol.X, ol.Y, W, H, rounding, softEdge)

		if softEdge >= 0 {
			roundCeil := max(int(math.Ceil(float64(rounding))), 0)
			tmpRect := ctx.Renderer.UnsafeTemp(0, W+roundCeil*2, H+roundCeil*2, true)
			ctx.Renderer.FillRect(tmpRect, float32(roundCeil), float32(roundCeil), W, H, rounding)
			or := CTR.AdjustXY(tmpRect, cw*0.75, ch*0.25)
			ctx.Renderer.Blur(canvas, tmpRect, or.X, or.Y, softEdge)

			cmpX, cmpY := cw*0.15, ch*0.65
			if ctx.SpacePressed {
				ctx.Renderer.Blur(canvas, tmpRect, cmpX-float32(roundCeil), cmpY-float32(roundCeil), softEdge)
			} else {
				ctx.Renderer.FillRectSoft(canvas, cmpX, cmpY, W, H, rounding, softEdge)
			}
		}

		ctx.Renderer.SetColorF32(0.25, 0, 0, 0.25)
		ctx.Renderer.FillRect(canvas, ol.X, ol.Y, W, H, rounding)
		ctx.Renderer.FillRect(canvas, or.X, or.Y, W, H, rounding)
	}

	app := NewTestApp(updater, drawer)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}
