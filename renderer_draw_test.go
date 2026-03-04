package shapes

import (
	"fmt"
	"image/color"
	"math"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// go test -run ^TestDrawShapes$ . -count 1
func TestDrawShapes(t *testing.T) {
	app := NewTestApp(func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Fill(color.Black)
		ctx.Renderer.SetColor(color.RGBA{255, 255, 255, 255})
		lx, ly := ctx.LeftClickF64()
		rx, ry := ctx.RightClickF64()
		ctx.Renderer.DrawLine(canvas, lx, ly, rx, ry, 6.0)
		ctx.Renderer.DrawCircle(canvas, 540, 80, 60)

		x, y := float64(160), float64(40)
		var points [3]PointF32
		points[0] = PointF32{X: float32(x), Y: float32(y)}
		points[1] = points[0].AddXY(30, 10)
		points[2] = points[0].AddXY(16, 50)
		ctx.Renderer.DrawTriangle(canvas, points, 0)

		x, y = float64(80), float64(260)
		ctx.Renderer.SetColor(color.RGBA{240, 48, 48, 255})
		points[0] = PointF32{X: float32(x), Y: float32(y)}
		points[1] = points[0].AddXY(70, -20)
		points[2] = points[0].AddXY(114, 80)
		ctx.Renderer.DrawTriangle(canvas, points, 0)
		ctx.Renderer.SetColor(color.RGBA{255, 255, 255, 255})
		ctx.Renderer.DrawTriangle(canvas, points, -8)
		x, y = float64(200), float64(300)
		points[1] = PointF32{X: float32(x), Y: float32(y)}
		points[0] = points[1].AddXY(70, -20)
		points[2] = points[1].AddXY(114, 80)
		ctx.Renderer.StrokeTriangle(canvas, points, 4, -8)
		v := uint8(32 + ctx.DistAnim(196-32, 1.0))
		ctx.Renderer.SetColor(color.RGBA{v, 0, 0, v})
		ctx.Renderer.StrokeTriangle(canvas, points, -4, 0)

		rads := ctx.RadsAnim(1.0)
		ctx.Renderer.DrawHexagon(canvas, 80, 400, 60, 0, float32(rads))

		rounding := float32(ctx.DistAnim(48.0, 1.0))
		ctx.Renderer.SetColorF32(0, 0.5, 1.0, 1.0)
		ctx.Renderer.DrawHexagon(canvas, 420, 400, 60, rounding, float32(rads))
		ctx.Renderer.ScaleAlphaBy(0.5)
		ctx.Renderer.DrawHexagon(canvas, 80, 400, 60, rounding, float32(rads))

	})
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestDrawTriangles$ . -count 1
func TestDrawTriangles(t *testing.T) {
	app := NewTestApp(func(canvas *ebiten.Image, ctx TestAppCtx) {
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
		ctx.Renderer.DrawTriangle(canvas, pointsL, outRounding)
		ctx.Renderer.SetColorF32(0.3, 0, 0.3, 0.3)
		ctx.Renderer.DrawCircle(canvas, pointsL[0].X, pointsL[0].Y, outRounding)
		ctx.Renderer.DrawCircle(canvas, pointsL[1].X, pointsL[1].Y, outRounding)
		ctx.Renderer.DrawCircle(canvas, pointsL[2].X, pointsL[2].Y, outRounding)

		ctx.Renderer.SetColorF32(1, 1, 1, 1)
		ctx.Renderer.DrawTriangle(canvas, pointsM, inRounding)
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
		ctx.Renderer.DrawCircle(canvas, pointsL[0].X, pointsL[0].Y, outRounding)
		ctx.Renderer.DrawCircle(canvas, pointsL[1].X, pointsL[1].Y, outRounding)
		ctx.Renderer.DrawCircle(canvas, pointsL[2].X, pointsL[2].Y, outRounding)

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
		ctx.Renderer.DrawCircle(canvas, pointsL[0].X, pointsL[0].Y, outRounding)
		ctx.Renderer.DrawCircle(canvas, pointsL[1].X, pointsL[1].Y, outRounding)
		ctx.Renderer.DrawCircle(canvas, pointsL[2].X, pointsL[2].Y, outRounding)

		for i, p := range pointsM {
			pointsM[i] = p.AddXY(0, ch/2+LVOffset)
		}
		ctx.Renderer.SetColorF32(1, 1, 1, 1)
		ctx.Renderer.StrokeTriangle(canvas, pointsM, -8.0, inRounding)
	})

	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestDrawHexagons$ . -count 1
func TestDrawHexagons(t *testing.T) {
	manRounding := float32(0.0)
	app := NewTestApp(func(canvas *ebiten.Image, ctx TestAppCtx) {
		cw, ch := rectSizeF32(canvas.Bounds())

		const Pad, MinRadius, MaxRadius = 16, 48, 64
		radius := MinRadius + float32(ctx.DistAnim(MaxRadius-MinRadius, 1.0))
		rads := float32(ctx.RadsAnim(1.0))

		// radius bounded hexagon, no roundness
		ctx.Renderer.SetColorF32(1.0, 1.0, 1.0, 1.0)
		ctx.Renderer.DrawCircle(canvas, Pad+MaxRadius, Pad+MaxRadius, radius)
		ctx.Renderer.SetColorF32(1.0, 0.0, 1.0, 1.0)
		ctx.Renderer.ScaleAlphaBy(0.666)
		ctx.Renderer.DrawHexagon(canvas, Pad+MaxRadius, Pad+MaxRadius, radius, 0.0, rads)

		// radius bounded hexagon with animated roundness
		roundness := float32(ctx.DistAnim(MaxRadius+16, 0.5))
		ctx.Renderer.SetColorF32(1.0, 1.0, 1.0, 1.0)
		ctx.Renderer.DrawCircle(canvas, cw/2, Pad+MaxRadius, radius)
		ctx.Renderer.SetColorF32(0.0, 0.0, 1.0, 1.0)
		ctx.Renderer.ScaleAlphaBy(0.666)
		ctx.Renderer.DrawCircle(canvas, cw/2, Pad+MaxRadius, roundness)
		ctx.Renderer.SetColorF32(1.0, 0.0, 1.0, 1.0)
		ctx.Renderer.ScaleAlphaBy(0.666)
		ctx.Renderer.DrawHexagon(canvas, cw/2, Pad+MaxRadius, radius, roundness, rads)

		// apothem bounded hexagon, no rounding
		apothem := radius * Sqrt3Div2
		ctx.Renderer.SetColorF32(1.0, 1.0, 1.0, 1.0)
		ctx.Renderer.DrawCircle(canvas, Pad+MaxRadius, ch/2, apothem)
		ctx.Renderer.SetColorF32(1.0, 0.0, 1.0, 1.0)
		ctx.Renderer.ScaleAlphaBy(0.666)
		ctx.Renderer.DrawHexagonApothem(canvas, Pad+MaxRadius, ch/2, apothem, 0.0, rads)

		// apothem bounded hexagon, outwards rounding
		rounding := float32(ctx.DistAnim(24.0, 1.0))
		ctx.Renderer.SetColorF32(0.0, 0.0, 1.0, 1.0)
		ctx.Renderer.ScaleAlphaBy(0.666)
		ctx.Renderer.DrawCircle(canvas, cw/2, ch/2, apothem+rounding)
		ctx.Renderer.SetColorF32(1.0, 1.0, 1.0, 1.0)
		ctx.Renderer.DrawCircle(canvas, cw/2, ch/2, apothem)
		ctx.Renderer.SetColorF32(0.0, 0.0, 1.0, 1.0)
		ctx.Renderer.ScaleAlphaBy(0.666)
		ctx.Renderer.DrawCircle(canvas, cw/2, ch/2, rounding)
		ctx.Renderer.SetColorF32(1.0, 0.0, 1.0, 1.0)
		ctx.Renderer.ScaleAlphaBy(0.666)
		ctx.Renderer.DrawHexagonApothem(canvas, cw/2, ch/2, apothem, rounding, rads)

		// apothem bounded hexagon, inwards rounding
		const MinApothem = MinRadius * Sqrt3Div2
		inRounding := float32(ctx.DistAnim(MinApothem, 0.5))
		ctx.Renderer.SetColorF32(1.0, 1.0, 1.0, 1.0)
		ctx.Renderer.DrawCircle(canvas, cw-16-MaxRadius, ch/2, apothem)
		ctx.Renderer.SetColorF32(0.0, 0.0, 1.0, 1.0)
		ctx.Renderer.ScaleAlphaBy(0.666)
		ctx.Renderer.DrawCircle(canvas, cw-16-MaxRadius, ch/2, inRounding)
		ctx.Renderer.SetColorF32(1.0, 0.0, 1.0, 1.0)
		ctx.Renderer.ScaleAlphaBy(0.666)
		ctx.Renderer.DrawHexagonApothem(canvas, cw-16-MaxRadius, ch/2, apothem, -inRounding, rads)

		// manual control
		ebiten.SetWindowTitle(fmt.Sprintf("%s  [rounding: %.02f (up/down), apothem: %.02f]", ctx.Title(), manRounding, MinApothem))
		if ctx.NewInput {
			if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) {
				manRounding += 1.0
			} else if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) {
				manRounding -= 1.0
			}
		}

		ctx.Renderer.SetColorF32(1.0, 1.0, 1.0, 1.0)
		ctx.Renderer.DrawCircle(canvas, 16+MaxRadius, ch-16-MaxRadius, MinApothem)
		ctx.Renderer.SetColorF32(0.0, 0.0, 1.0, 1.0)
		ctx.Renderer.ScaleAlphaBy(0.666)
		if manRounding < 0 {
			ctx.Renderer.DrawCircle(canvas, 16+MaxRadius, ch-16-MaxRadius, -manRounding)
		} else {
			ctx.Renderer.DrawCircle(canvas, 16+MaxRadius, ch-16-MaxRadius, MinApothem+manRounding)
		}
		ctx.Renderer.SetColorF32(1.0, 0.0, 1.0, 1.0)
		ctx.Renderer.ScaleAlphaBy(0.666)
		ctx.Renderer.DrawHexagonApothem(canvas, 16+MaxRadius, ch-16-MaxRadius, MinApothem, manRounding, 0.0)
	})

	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestDrawArea$ . -count 1
func TestDrawArea(t *testing.T) {
	app := NewTestApp(func(canvas *ebiten.Image, ctx TestAppCtx) {
		lx, ly := ctx.LeftClickF32()
		rx, ry := ctx.RightClickF32()

		ctx.Renderer.SetColorF32(1.0, 1.0, 1.0, 1.0)
		w1, h1 := float32(128), float32(48)
		w2, h2 := float32(48), float32(128)
		ctx.Renderer.DrawArea(canvas, lx-w1/2, ly-h1/2, w1, h1, -float32(ctx.DistAnim(float64(min(w1, h1))/2.0, 1.0)))
		ctx.Renderer.DrawArea(canvas, rx-w2/2, ry-h2/2, w2, h2, float32(ctx.DistAnim(float64(min(w1, h1))/2.0, 1.0)))

		ctx.Renderer.SetColorF32(0.2, 0.0, 0.2, 0.2)
		ctx.Renderer.DrawCircle(canvas, lx, ly, max(w1, h1)/2.0)
		ctx.Renderer.DrawCircle(canvas, rx, ry, max(w2, h2)/2.0)

		cw, ch := rectSizeF32(canvas.Bounds())
		ctx.Renderer.SetColorF32(0.5, 0.5, 0.5, 0.5)
		ctx.Renderer.DrawArea(canvas, 16, ch-16, 128, -128, 0)
		ctx.Renderer.DrawArea(canvas, 32, ch-32, 128-32, -(128 - 32), float32(ctx.DistAnim(16, 1.0)))

		ctx.Renderer.DrawCircle(canvas, 164+32, ch-16-32, 32)
		ctx.Renderer.DrawCircle(canvas, 164+128-32, ch-16-128+32, 32)
		ctx.Renderer.DrawArea(canvas, 164, ch-16, 128, -128, -float32(ctx.DistAnim(32, 1.0)))

		collapseRounding := -float32(ctx.DistAnim(196.0, 0.5))
		const CRW, CRH = 128, 96
		ctx.Renderer.DrawArea(canvas, cw-CRW-16, 16, CRW, CRH, collapseRounding)
		ctx.Renderer.SetColorF32(0.2, 0.0, 0.2, 0.2)
		ctx.Renderer.DrawCircle(canvas, cw-CRW/2-16, 16+CRH/2, min(abs(collapseRounding), CRW/2, CRH/2))
	})
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestDrawAreaPrecise$ . -count 1
func TestDrawAreaPrecise(t *testing.T) {
	app := NewTestApp(func(canvas *ebiten.Image, ctx TestAppCtx) {
		ctx.Renderer.SetColorF32(1.0, 0.0, 1.0, 1.0)
		ctx.Renderer.DrawIntArea(canvas, 0, 0, 258, 258)
		ctx.DrawAtF32(canvas, ctx.Images[0], 1, 1)
		ctx.DrawAtF32(canvas, ctx.Images[1], 2, 2)

		ctx.DrawAtF32(canvas, ctx.Images[2], 270, 1)
		ctx.DrawAtF32(canvas, ctx.Images[3], 271, 2)
	})

	box := ebiten.NewImage(256, 256)
	app.Renderer.DrawArea(box, 1, 1, 254, 254, 0)
	box2 := ebiten.NewImage(254, 254)
	app.Renderer.SetColorF32(1.0, 0, 0, 1.0)
	app.Renderer.DrawArea(box2, 1, 1, 252, 252, 0)

	box3 := ebiten.NewImage(256, 256)
	app.Renderer.SetColorF32(1.0, 1.0, 1.0, 1.0)
	app.Renderer.DrawArea(box3, 1, 1, 254, 254, -6.0)
	box4 := ebiten.NewImage(254, 254)
	app.Renderer.SetColorF32(1.0, 0, 0, 1.0)
	app.Renderer.DrawArea(box4, 1, 1, 252, 252, -6.0)

	app.Renderer.SetColorF32(1.0, 1.0, 1.0, 1.0)

	app.Images = append(app.Images, box, box2, box3, box4)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestStrokeIntArea$ . -count 1
func TestStrokeIntArea(t *testing.T) {
	app := NewTestApp(func(canvas *ebiten.Image, ctx TestAppCtx) {
		lx, ly := ctx.LeftClick.X, ctx.LeftClick.Y
		ctx.Renderer.SetColor(color.RGBA{255, 255, 255, 255})
		ctx.Renderer.DrawIntArea(canvas, lx, ly, 200, 50)
		ctx.Renderer.SetColor(color.RGBA{0, 255, 0, 255})
		ctx.Renderer.StrokeIntArea(canvas, lx-1, ly-1, 200+2, 50+2, 1, 0)

		ctx.Renderer.SetColor(color.RGBA{0, 128, 0, 128})
		ctx.Renderer.StrokeIntArea(canvas, lx, ly, 200, 50, 0, 1)

		rx, ry := ctx.RightClick.X, ctx.RightClick.Y
		ctx.Renderer.SetColor(color.RGBA{240, 0, 240, 255}, 0, 2)
		ctx.Renderer.StrokeIntArea(canvas, rx, ry, 100, 50, 4, 4)

		ctx.Renderer.SetColor(color.RGBA{64, 128, 64, 128})
		ctx.Renderer.DrawIntArea(canvas, rx, ry, 100, 50)

		ctx.Renderer.SetColor(color.RGBA{255, 0, 0, 255}, 0)
		ctx.Renderer.SetColor(color.RGBA{0, 255, 0, 255}, 1)
		ctx.Renderer.SetColor(color.RGBA{0, 0, 255, 255}, 2)
		ctx.Renderer.SetColor(color.RGBA{0, 255, 255, 255}, 3)
		ctx.Renderer.StrokeIntArea(canvas, lx, ry, 80, 50, 8, 8)
	})
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestStrokeArea$ . -count 1
func TestStrokeArea(t *testing.T) {
	app := NewTestApp(func(canvas *ebiten.Image, ctx TestAppCtx) {
		lx, ly := ctx.LeftClickF32()
		ctx.Renderer.SetColor(color.RGBA{255, 255, 255, 255})
		ctx.Renderer.DrawArea(canvas, lx, ly, 200, 50, 16)
		ctx.Renderer.SetColor(color.RGBA{0, 255, 0, 255})
		ctx.Renderer.StrokeArea(canvas, lx, ly, 200, 50, 0, 2, 16)

		ctx.Renderer.SetColor(color.RGBA{128, 0, 0, 128})
		ctx.Renderer.StrokeArea(canvas, lx, ly, 200, 50, 2, 0, 16)

		rx, ry := ctx.RightClickF32()
		ctx.Renderer.SetColor(color.RGBA{240, 0, 240, 255}, 0, 2)
		ctx.Renderer.StrokeArea(canvas, rx, ry, 100, 50, 4, 4, 25)

		a := uint8(ctx.DistAnim(144.0, 1.0))
		ctx.Renderer.SetColor(color.RGBA{a, a, a, a})
		ctx.Renderer.DrawIntArea(canvas, int(rx), int(ry), 100, 50)

		ctx.Renderer.SetColor(color.RGBA{255, 0, 0, 255}, 0)
		ctx.Renderer.SetColor(color.RGBA{0, 255, 0, 255}, 1)
		ctx.Renderer.SetColor(color.RGBA{0, 0, 255, 255}, 2)
		ctx.Renderer.SetColor(color.RGBA{0, 255, 255, 255}, 3)
		extra := float32(ctx.DistAnim(16, 1.0))
		subRounding := float32(ctx.DistAnim(20, 1.0))
		ctx.Renderer.StrokeArea(canvas, lx, ry, 80+extra, 50, 8, 8, 25-subRounding)

		w, h := rectSizeF32(canvas.Bounds())
		ctx.Renderer.SetColor(color.RGBA{255, 255, 255, 255})
		ctx.Renderer.StrokeArea(canvas, w-16, 16, -200, 64, 0, 8, 32)
		ctx.Renderer.SetColor(color.RGBA{128, 0, 0, 128})
		ctx.Renderer.DrawCircle(canvas, w-16-32+8, 16+32-8, 32)

		ctx.Renderer.SetColor(color.RGBA{255, 255, 255, 255})
		ctx.Renderer.StrokeArea(canvas, 16, h-16, 96, -64, 0, 8, -32.0)
		ctx.Renderer.SetColor(color.RGBA{128, 0, 0, 128})
		ctx.Renderer.DrawCircle(canvas, 16+32, h-16-32, 32)
	})
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestStrokeCircle$ . -count 1
func TestStrokeCircle(t *testing.T) {
	app := NewTestApp(func(canvas *ebiten.Image, ctx TestAppCtx) {
		const Radius = 72

		lx, ly := ctx.LeftClickF32()
		rx, ry := ctx.RightClickF32()

		ctx.Renderer.SetColorF32(1.0, 1.0, 1.0, 1.0)
		ctx.Renderer.ScaleAlphaBy(0.666)
		ctx.Renderer.DrawCircle(canvas, lx, ly, Radius)
		ctx.Renderer.DrawCircle(canvas, rx, ry, Radius)

		const MaxThickness = 16
		thick := float32(ctx.DistAnim(MaxThickness, 1.0))
		ctx.Renderer.ScaleAlphaBy(0.666)
		ctx.Renderer.DrawCircle(canvas, lx, ly, Radius-MaxThickness)
		ctx.Renderer.DrawCircle(canvas, rx, ry, Radius-MaxThickness)

		ctx.Renderer.SetColorF32(1.0, 0.0, 1.0, 1.0)
		ctx.Renderer.ScaleAlphaBy(0.666)
		ctx.Renderer.StrokeCircle(canvas, lx, ly, Radius, thick)
		ctx.Renderer.StrokeCircle(canvas, rx, ry, Radius, -thick)

		thick2 := float32(ctx.DistAnim(32.0, 1.0))
		ctx.Renderer.StrokeCircle(canvas, lx, ry, 16, -thick2)
		ctx.Renderer.StrokeCircle(canvas, rx, ly, thick2-8.0, 16.0)
	})
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestDrawEllipse$ . -count 1
func TestDrawEllipse(t *testing.T) {
	app := NewTestApp(func(canvas *ebiten.Image, ctx TestAppCtx) {
		lx, ly := ctx.LeftClickF32()
		rx, ry := ctx.RightClickF32()

		ctx.Renderer.SetColorF32(0.5, 0.5, 0.5, 0.5)
		ctx.Renderer.DrawCircle(canvas, lx, ly, 64.0)
		ctx.Renderer.DrawCircle(canvas, rx, ry, 32.0)

		ctx.Renderer.SetColorF32(1.0, 1.0, 1.0, 1.0)
		ctx.Renderer.DrawEllipse(canvas, lx, ly, 24.0, 64.0, ctx.RadsAnim(1.0))
		ctx.Renderer.DrawEllipse(canvas, rx, ry, 32.0, 16.0, 0)

		ctx.Renderer.SetColorF32(0.0, 0.5, 0.5, 0.5)
		ctx.Renderer.DrawCircle(canvas, lx, ly, 24.0)
		ctx.Renderer.DrawCircle(canvas, rx, ry, 16.0)
	})
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestDrawRing$ . -count 1
func TestDrawRing(t *testing.T) {
	app := NewTestApp(func(canvas *ebiten.Image, ctx TestAppCtx) {
		lx, ly := ctx.LeftClickF32()
		rx, ry := ctx.RightClickF32()

		ctx.Renderer.SetColorF32(1.0, 1.0, 1.0, 1.0)
		ctx.Renderer.DrawCircle(canvas, lx, ly, 64.0)
		ctx.Renderer.DrawCircle(canvas, rx, ry, 48.0)

		ctx.Renderer.SetColorF32(1.0, 0.0, 1.0, 1.0, 0, 1)
		ctx.Renderer.SetColorF32(0.5, 1.0, 0.5, 1.0, 2, 3)
		ctx.Renderer.DrawRing(canvas, lx, ly, 65.0, 67.0)
		ctx.Renderer.DrawRing(canvas, rx, ry, 48.0-4, 48+0)

		ctx.Renderer.SetColorF32(0.5, 0.5, 0.5, 0.5)
		ctx.Renderer.DrawCircle(canvas, rx, ry, 48.0)
	})
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestDrawRingSector$ . -count 1
func TestDrawRingSector(t *testing.T) {
	app := NewTestApp(func(canvas *ebiten.Image, ctx TestAppCtx) {
		w, h := rectSizeF32(canvas.Bounds())
		cx, cy := w/2.0, h/2.0
		startRads := ctx.ModAnim(2*math.Pi, 1.0)
		endRads := startRads + 0.4 + ctx.DistAnim(1.6, 1.0)
		ctx.Renderer.SetColorF32(1.0, 1.0, 1.0, 1.0)
		ctx.Renderer.DrawRingSector(canvas, cx, cy, 48, 128, startRads, endRads, 0.0)
		ctx.Renderer.SetColorF32(0.0, 0.5, 0.5, 0.5)
		ctx.Renderer.DrawRingSector(canvas, cx, cy, 48, 128, startRads, endRads, 8.0)
	})
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestStrokeRingSector$ . -count 1
func TestStrokeRingSector(t *testing.T) {
	app := NewTestApp(func(canvas *ebiten.Image, ctx TestAppCtx) {
		w, h := rectSizeF32(canvas.Bounds())
		cx, cy := w/2.0, h/2.0
		startRads := ctx.ModAnim(2*math.Pi, 1.0)
		endRads := startRads + 0.4 + ctx.DistAnim(1.6, 1.0)

		ctx.Renderer.SetColorF32(0, 0, 0.5, 0.5)
		ctx.Renderer.DrawRingSector(canvas, cx, cy, 48, 128, startRads, endRads, 0.0)

		ctx.Renderer.SetColorF32(1.0, 1.0, 1.0, 1.0)
		thick := float32(ctx.DistAnim(8.0, 1.0))
		ctx.Renderer.StrokeRingSector(canvas, cx, cy, 48, 128, thick, startRads, endRads, 0.0)
		ctx.Renderer.SetColorF32(0.0, 0.5, 0.5, 0.5)
		ctx.Renderer.StrokeRingSector(canvas, cx, cy, 48, 128, thick, startRads, endRads, 8.0)
	})
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestDrawPie$ . -count 1
func TestDrawPie(t *testing.T) {
	app := NewTestApp(func(canvas *ebiten.Image, ctx TestAppCtx) {
		w, h := rectSizeF32(canvas.Bounds())
		cx, cy := w/2.0, h/2.0
		rate := -0.01 + ctx.DistAnim(1.02, 1.0)

		ctx.Renderer.SetColorF32(1.0, 1.0, 1.0, 1.0)
		ctx.Renderer.DrawPieRate(canvas, cx, cy, 96.0, RadsRight, rate, 6.0)

		ctx.Renderer.SetColorF32(0.0, 1.0, 0.0, 1.0)
		ctx.Renderer.DrawPie(canvas, cx, cy, 64.0, RadsRight+rate, RadsBottom, 3.0)
	})
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestStrokePie$ . -count 1
func TestStrokePie(t *testing.T) {
	app := NewTestApp(func(canvas *ebiten.Image, ctx TestAppCtx) {
		w, h := rectSizeF32(canvas.Bounds())
		cx, cy := w/2.0, h/2.0
		rate := -0.01 + ctx.DistAnim(1.02, 1.0)
		thick := float32(4.0)

		ctx.Renderer.SetColorF32(1.0, 1.0, 1.0, 1.0)
		ctx.Renderer.StrokePieRate(canvas, cx, cy, 96.0, thick, RadsRight, rate, 6.0)

		ctx.Renderer.SetColorF32(0.0, 1.0, 0.0, 1.0)
		ctx.Renderer.StrokePie(canvas, cx, cy, 64.0, thick, RadsRight+rate, RadsBottom, 3.0)
	})
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestDrawQuad$ . -count 1
func TestDrawQuad(t *testing.T) {
	app := NewTestApp(func(canvas *ebiten.Image, ctx TestAppCtx) {
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
		ctx.Renderer.DrawQuad(canvas, quad, thickening)
	})
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestDrawQuadSoft$ . -count 1
func TestDrawQuadSoft(t *testing.T) {
	app := NewTestApp(func(canvas *ebiten.Image, ctx TestAppCtx) {
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
		ctx.Renderer.DrawQuadSoft(canvas, quad, thickening, 64.0)
	})
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestDrawAreaSoft$ . -count 1
func TestDrawAreaSoft(t *testing.T) {
	app := NewTestApp(func(canvas *ebiten.Image, ctx TestAppCtx) {
		cw, ch := rectSizeF32(canvas.Bounds())

		rounding := float32(-32 + ctx.DistAnim(48, 1.0))
		rounding2 := float32(-32 + ctx.DistAnim(48, 0.777))
		soft := float32(-16.0 + ctx.DistAnim(32, 1.0))
		ctx.Renderer.SetColorF32(1, 1, 1, 1)
		ctx.Renderer.DrawAreaSoft(canvas, cw/3-64, ch/3-32, 128, 64, rounding, soft)
		ctx.Renderer.DrawAreaSoft(canvas, cw-cw/3-64, ch-ch/3-32, 128, 64, rounding2, 0.0)

		const brW, brH = 128, 96
		ox, oy := cw-cw/3-64, ch/3-32
		brSoft := float32(ctx.DistAnim(16.0, 1.0))
		if ebiten.IsKeyPressed(ebiten.KeySpace) {
			tmp := ctx.Renderer.UnsafeTemp(0, brW+96, brH+96, true)
			ctx.Renderer.DrawArea(tmp, 96/2, 96/2, brW, brH, rounding)
			ctx.Renderer.ApplyBlur(canvas, tmp, ox-96/2, oy-96/2, brSoft)
		} else {
			ctx.Renderer.DrawAreaBlur(canvas, ox, oy, brW, brH, rounding, brSoft)
		}

		if ebiten.IsKeyPressed(ebiten.KeySpace) {
			ctx.Renderer.SetColorF32(1, 0, 0, 1)
			ctx.Renderer.ScaleAlphaBy(0.333)
			ctx.Renderer.DrawArea(canvas, cw/3-64, ch/3-32, 128, 64, rounding)
			ctx.Renderer.DrawArea(canvas, cw-cw/3-64, ch-ch/3-32, 128, 64, rounding2)
		}
	})

	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}
