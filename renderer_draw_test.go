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
		lx, ly := ctx.LeftClickF64()
		rx, ry := ctx.RightClickF64()
		ctx.Renderer.StrokeLine(canvas, lx, ly, rx, ry, 6.0)
		ctx.Renderer.FillCircle(canvas, 540, 80, 60)

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
		ctx.Renderer.FillHexagon(canvas, 80, 400, 60, 0, float32(rads))

		rounding := float32(ctx.DistAnim(48.0, 1.0))
		ctx.Renderer.SetColorF32(0, 0.5, 1.0, 1.0)
		ctx.Renderer.FillHexagon(canvas, 420, 400, 60, rounding, float32(rads))
		ctx.Renderer.ScaleAlphaBy(0.5)
		ctx.Renderer.FillHexagon(canvas, 80, 400, 60, rounding, float32(rads))
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

// go test -run ^TestDrawArea$ . -count 1
func TestFillArea(t *testing.T) {
	updater := func(ctx TestAppCtx) {}
	drawer := func(canvas *ebiten.Image, ctx TestAppCtx) {
		lx, ly := ctx.LeftClickF32()
		rx, ry := ctx.RightClickF32()

		ctx.Renderer.SetColorF32(1.0, 1.0, 1.0, 1.0)
		w1, h1 := float32(128), float32(48)
		w2, h2 := float32(48), float32(128)
		ctx.Renderer.FillRect(canvas, lx-w1/2, ly-h1/2, w1, h1, -float32(ctx.DistAnim(float64(min(w1, h1))/2.0, 1.0)))
		ctx.Renderer.FillRect(canvas, rx-w2/2, ry-h2/2, w2, h2, float32(ctx.DistAnim(float64(min(w1, h1))/2.0, 1.0)))

		ctx.Renderer.SetColorF32(0.2, 0.0, 0.2, 0.2)
		ctx.Renderer.FillCircle(canvas, lx, ly, max(w1, h1)/2.0)
		ctx.Renderer.FillCircle(canvas, rx, ry, max(w2, h2)/2.0)

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

// go test -run ^TestDrawAreaPrecise$ . -count 1
func TestDrawAreaPrecise(t *testing.T) {
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

// go test -run ^TestStrokeIntArea$ . -count 1
func TestStrokeIntArea(t *testing.T) {
	updater := func(ctx TestAppCtx) {}
	drawer := func(canvas *ebiten.Image, ctx TestAppCtx) {
		lx, ly := ctx.LeftClick.X, ctx.LeftClick.Y
		ctx.Renderer.SetColor(color.RGBA{255, 255, 255, 255})
		ctx.Renderer.FillIntRect(canvas, RectWithSize(lx, ly, 200, 50), 0)
		ctx.Renderer.SetColor(color.RGBA{0, 255, 0, 255})
		ctx.Renderer.StrokeIntRect(canvas, RectWithSize(lx-1, ly-1, 200+2, 50+2), 1, 0, 0)

		ctx.Renderer.SetColor(color.RGBA{0, 128, 0, 128})
		ctx.Renderer.StrokeIntRect(canvas, RectWithSize(lx, ly, 200, 50), 0, 1, 0)

		rx, ry := ctx.RightClick.X, ctx.RightClick.Y
		ctx.Renderer.SetColor(color.RGBA{240, 0, 240, 255}, 0, 2)
		ctx.Renderer.StrokeIntRect(canvas, RectWithSize(rx, ry, 100, 50), 4, 4, 0)

		ctx.Renderer.SetColor(color.RGBA{64, 128, 64, 128})
		ctx.Renderer.FillIntRect(canvas, RectWithSize(rx, ry, 100, 50), 0)

		ctx.Renderer.SetColor(color.RGBA{255, 0, 0, 255}, 0)
		ctx.Renderer.SetColor(color.RGBA{0, 255, 0, 255}, 1)
		ctx.Renderer.SetColor(color.RGBA{0, 0, 255, 255}, 2)
		ctx.Renderer.SetColor(color.RGBA{0, 255, 255, 255}, 3)
		ctx.Renderer.StrokeIntRect(canvas, RectWithSize(lx, ry, 80, 50), 8, 8, 0)
	}

	app := NewTestApp(updater, drawer)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestStrokeArea$ . -count 1
func TestStrokeArea(t *testing.T) {
	updater := func(ctx TestAppCtx) {}
	drawer := func(canvas *ebiten.Image, ctx TestAppCtx) {
		lx, ly := ctx.LeftClickF32()
		ctx.Renderer.SetColor(color.RGBA{255, 255, 255, 255})
		ctx.Renderer.FillRect(canvas, lx, ly, 200, 50, 16)
		ctx.Renderer.SetColor(color.RGBA{0, 255, 0, 255})
		ctx.Renderer.StrokeRect(canvas, lx, ly, 200, 50, 0, 2, 16)

		ctx.Renderer.SetColor(color.RGBA{128, 0, 0, 128})
		ctx.Renderer.StrokeRect(canvas, lx, ly, 200, 50, 2, 0, 16)

		rx, ry := ctx.RightClickF32()
		ctx.Renderer.SetColor(color.RGBA{240, 0, 240, 255}, 0, 2)
		ctx.Renderer.StrokeRect(canvas, rx, ry, 100, 50, 4, 4, 25)

		a := uint8(ctx.DistAnim(144.0, 1.0))
		ctx.Renderer.SetColor(color.RGBA{a, a, a, a})
		ctx.Renderer.FillIntRect(canvas, RectWithSize(int(rx), int(ry), 100, 50), 0)

		ctx.Renderer.SetColor(color.RGBA{255, 0, 0, 255}, 0)
		ctx.Renderer.SetColor(color.RGBA{0, 255, 0, 255}, 1)
		ctx.Renderer.SetColor(color.RGBA{0, 0, 255, 255}, 2)
		ctx.Renderer.SetColor(color.RGBA{0, 255, 255, 255}, 3)
		extra := float32(ctx.DistAnim(16, 1.0))
		subRounding := float32(ctx.DistAnim(20, 1.0))
		ctx.Renderer.StrokeRect(canvas, lx, ry, 80+extra, 50, 8, 8, 25-subRounding)

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

// go test -run ^TestAreaRounding$ . -count 1
func TestAreaRounding(t *testing.T) {
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

// go test -run ^TestStrokeCircle$ . -count 1
func TestStrokeCircle(t *testing.T) {
	updater := func(ctx TestAppCtx) {}
	drawer := func(canvas *ebiten.Image, ctx TestAppCtx) {
		const Radius = 72

		lx, ly := ctx.LeftClickF32()
		rx, ry := ctx.RightClickF32()

		ctx.Renderer.SetColorF32(1.0, 1.0, 1.0, 1.0)
		ctx.Renderer.ScaleAlphaBy(0.666)
		ctx.Renderer.FillCircle(canvas, lx, ly, Radius)
		ctx.Renderer.FillCircle(canvas, rx, ry, Radius)

		const MaxThickness = 16
		thick := float32(ctx.DistAnim(MaxThickness, 1.0))
		ctx.Renderer.ScaleAlphaBy(0.666)
		ctx.Renderer.FillCircle(canvas, lx, ly, Radius-MaxThickness)
		ctx.Renderer.FillCircle(canvas, rx, ry, Radius-MaxThickness)

		ctx.Renderer.SetColorF32(1.0, 0.0, 1.0, 1.0)
		ctx.Renderer.ScaleAlphaBy(0.666)
		ctx.Renderer.StrokeCircle(canvas, lx, ly, Radius, thick)
		ctx.Renderer.StrokeCircle(canvas, rx, ry, Radius, -thick)

		thick2 := float32(ctx.DistAnim(32.0, 1.0))
		ctx.Renderer.StrokeCircle(canvas, lx, ry, 16, -thick2)
		ctx.Renderer.StrokeCircle(canvas, rx, ly, thick2-8.0, 16.0)
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
		lx, ly := ctx.LeftClickF32()
		rx, ry := ctx.RightClickF32()

		ctx.Renderer.SetColorF32(0.5, 0.5, 0.5, 0.5)
		ctx.Renderer.FillCircle(canvas, lx, ly, 64.0)
		ctx.Renderer.FillCircle(canvas, rx, ry, 32.0)

		ctx.Renderer.SetColorF32(1.0, 1.0, 1.0, 1.0)
		ctx.Renderer.FillEllipse(canvas, lx, ly, 24.0, 64.0, ctx.RadsAnim(1.0))
		ctx.Renderer.FillEllipse(canvas, rx, ry, 32.0, 16.0, 0)

		ctx.Renderer.SetColorF32(0.0, 0.5, 0.5, 0.5)
		ctx.Renderer.FillCircle(canvas, lx, ly, 24.0)
		ctx.Renderer.FillCircle(canvas, rx, ry, 16.0)
	}

	app := NewTestApp(updater, drawer)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestDrawCircSector$ . -count 1
func TestDrawCircSector(t *testing.T) {
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

// go test -run ^TestDrawCircSectorRounding$ . -count 1
func TestDrawCircSectorRounding(t *testing.T) {
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

// go test -run ^TestFillQuad$ . -count 1
func TestFillQuad(t *testing.T) {
	updater := func(ctx TestAppCtx) {}
	drawer := func(canvas *ebiten.Image, ctx TestAppCtx) {
		lx, ly := ctx.LeftClickF32()
		rx, ry := ctx.RightClickF32()

		w, h := rectSizeF32(canvas.Bounds())
		quad := [4]PointF32{
			{X: lx, Y: ly},
			{X: w/2.0 + w/4.0, Y: h/2.0 - h/4.0},
			{X: rx, Y: ry},
			{X: w/2.0 - w/4.0, Y: h/2.0 + h/4.0},
		}
		thickening := float32(ctx.DistAnim(48.0, 1.0))
		ctx.Renderer.FillQuad(canvas, quad, thickening)
	}

	app := NewTestApp(updater, drawer)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestDrawQuadSoft$ . -count 1
func TestDrawQuadSoft(t *testing.T) {
	updater := func(ctx TestAppCtx) {}
	drawer := func(canvas *ebiten.Image, ctx TestAppCtx) {
		lx, ly := ctx.LeftClickF32()
		rx, ry := ctx.RightClickF32()

		w, h := rectSizeF32(canvas.Bounds())
		quad := [4]PointF32{
			{X: lx, Y: ly},
			{X: w/2.0 + w/4.0, Y: h/2.0 - h/4.0},
			{X: rx, Y: ry},
			{X: w/2.0 - w/4.0, Y: h/2.0 + h/4.0},
		}
		thickening := float32(ctx.DistAnim(48.0, 1.0))
		ctx.Renderer.FillQuadSoft(canvas, quad, thickening, 64.0)
	}

	app := NewTestApp(updater, drawer)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestDrawAreaSoft$ . -count 1
func TestDrawAreaSoft(t *testing.T) {
	updater := func(ctx TestAppCtx) {}
	drawer := func(canvas *ebiten.Image, ctx TestAppCtx) {
		cw, ch := rectSizeF32(canvas.Bounds())

		rounding := float32(-32 + ctx.DistAnim(48, 1.0))
		rounding2 := float32(-32 + ctx.DistAnim(48, 0.777))
		soft := float32(-16.0 + ctx.DistAnim(32, 1.0))
		ctx.Renderer.SetColorF32(1, 1, 1, 1)
		ctx.Renderer.FillRectSoft(canvas, cw/3-64, ch/3-32, 128, 64, rounding, soft)
		ctx.Renderer.FillRectSoft(canvas, cw-cw/3-64, ch-ch/3-32, 128, 64, rounding2, 0.0)

		const brW, brH = 128, 96
		ox, oy := cw-cw/3-64, ch/3-32
		brSoft := float32(ctx.DistAnim(16.0, 1.0))
		if ctx.SpacePressed {
			tmp := ctx.Renderer.UnsafeTemp(0, brW+96, brH+96, true)
			ctx.Renderer.FillRect(tmp, 96/2, 96/2, brW, brH, rounding)
			ctx.Renderer.ApplyBlur(canvas, tmp, ox-96/2, oy-96/2, brSoft)
		} else {
			ctx.Renderer.FillRectBlur(canvas, ox, oy, brW, brH, rounding, brSoft)
		}

		if ctx.SpacePressed {
			ctx.Renderer.SetColorF32(1, 0, 0, 1)
			ctx.Renderer.ScaleAlphaBy(0.333)
			ctx.Renderer.FillRect(canvas, cw/3-64, ch/3-32, 128, 64, rounding)
			ctx.Renderer.FillRect(canvas, cw-cw/3-64, ch-ch/3-32, 128, 64, rounding2)
		}
	}

	app := NewTestApp(updater, drawer)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}
