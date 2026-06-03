package shapes

import (
	"fmt"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

// go test -run ^TestOffsetQuad$ . -count 1
func TestOffsetQuad(t *testing.T) {
	var points [4]PointF32

	rounding := float32(0.0)
	var drag pointsDragger
	updater := func(ctx TestAppCtx) {
		rounding = updateParam(ctx, ebiten.KeyR, rounding, -150.0, 150.0, 2.0)
		drag.Update(points[:])
	}

	firstDraw := true
	drawer := func(canvas *ebiten.Image, ctx TestAppCtx) {
		if firstDraw {
			w, h := rectSizeF32(canvas.Bounds())
			points[0] = PtF32(w*0.25, h*0.2)
			points[1] = PtF32(w*0.8, h*0.3)
			points[2] = PtF32(w*0.75, h*0.75)
			points[3] = PtF32(w*0.2, h*0.8)
			firstDraw = false
		}
		ctx.Renderer.SetColorF32(1, 1, 1, 1)
		out, shape, leftoverOffset := offsetQuad(points, rounding)
		info := fmt.Sprintf("Press and drag the points\nRounding: %.02f [R]\nShape: %s\nLeftover offset: %.02f", rounding, shape, leftoverOffset)
		ctx.Renderer.Text(canvas, info, 12, 12, TextOpts(1.0, TopLeft.Snap(CapLine)))

		for _, p := range points {
			ctx.Renderer.FillCircle(canvas, p.X, p.Y, 3.0)
		}

		ctx.Renderer.SetColorF32(0, 0.5, 0, 0.5)
		for _, p := range out[:shape.NumPoints()] {
			ctx.Renderer.FillCircle(canvas, p.X, p.Y, 3.0)
		}
	}

	app := NewTestApp(updater, drawer)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestShrinkTriangle$ . -count 1
func TestShrinkTriangle(t *testing.T) {
	var points [3]PointF32

	offset := float32(0.0)
	var drag pointsDragger
	updater := func(ctx TestAppCtx) {
		offset = updateParam(ctx, ebiten.KeyO, offset, 0, 100.0, 2.0)
		drag.Update(points[:])
	}

	firstDraw := true
	drawer := func(canvas *ebiten.Image, ctx TestAppCtx) {
		if firstDraw {
			w, h := rectSizeF32(canvas.Bounds())
			points[0] = PtF32(w*0.4, h*0.25)
			points[1] = PtF32(w*0.8, h*0.75)
			points[2] = PtF32(w*0.2, h*0.9)
			firstDraw = false
		}
		p1, p2, p3, shape, leftoverOffset := shrinkTriangle(points[0], points[1], points[2], offset)

		ctx.Renderer.SetColorF32(1, 1, 1, 1)
		info := fmt.Sprintf("Press and drag the points\nShrink offset: %.02f [O]\nShape: %s\nLeftover offset: %.02f", offset, shape.String(), leftoverOffset)
		ctx.Renderer.Text(canvas, info, 12, 12, TextOpts(1.0, TopLeft.Snap(CapLine)))

		for _, p := range points {
			ctx.Renderer.FillCircle(canvas, p.X, p.Y, 3.0)
		}

		ctx.Renderer.SetColorF32(0, 0.5, 0, 0.5)
		for _, p := range []PointF32{p1, p2, p3}[:shape.NumPoints()] {
			ctx.Renderer.FillCircle(canvas, p.X, p.Y, 3.0)
		}
	}

	app := NewTestApp(updater, drawer)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

type pointsDragger struct {
	wasPressed  bool
	prevPressed int
}

func (pd *pointsDragger) Update(points []PointF32) {
	if !ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		pd.wasPressed = false
		pd.prevPressed = -1
		return
	}

	dragPtIndex := -1
	nearestPtDist := Float32Inf()
	x, y := ebiten.CursorPosition()
	p := PtF32(float32(x), float32(y))
	for i, point := range points {
		l := point.Sub(p).Length()
		if l < float32(nearestPtDist) && l < 16 {
			nearestPtDist = l
			dragPtIndex = i
		}
	}

	if pd.wasPressed {
		dragPtIndex = pd.prevPressed
	} else {
		pd.prevPressed = dragPtIndex
	}
	if dragPtIndex != -1 {
		points[dragPtIndex] = p
	}
	pd.wasPressed = true
}
