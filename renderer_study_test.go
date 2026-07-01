package shapes

import (
	"fmt"
	"image/color"
	"math"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

// go test -run ^TestStudyWaveFuncs$ . -count 1
func TestStudyWaveFuncs(t *testing.T) {
	updater := func(TestAppCtx) {}
	drawer := func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Fill(color.RGBA{128, 0, 0, 255})
		ctx.Renderer.studyWaveFuncs(canvas, 1.0, 8.0)
	}

	app := NewTestApp(updater, drawer)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestStudyRadians$ . -count 1
func TestStudyRadians(t *testing.T) {
	const PointRadius = 3.5
	const Dist = 96.0

	rads := -math.Pi
	updater := func(TestAppCtx) { rads += 0.01 }
	drawer := func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Fill(color.Black)
		w, h := rectSizeF32(canvas.Bounds())
		cx, cy := w/2.0, h/2.0
		ctx.Renderer.FillCircle(canvas, cx, cy, PointRadius)

		sin, cos := math.Sincos(normURads(rads))
		ctx.Renderer.FillCircle(canvas, cx+float32(Dist*cos), cy+float32(Dist*sin), PointRadius)
		ebiten.SetWindowTitle(fmt.Sprintf("%s [rads: %02f]", ctx.Title(), rads))
	}

	app := NewTestApp(updater, drawer)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestStudyRotation$ . -count 1
func TestStudyRotation(t *testing.T) {
	const CircRadius = 64.0
	const PointRadius = 3.5
	const AngleDelta = math.Pi * 1.618033988749895

	rotation := 0.0
	updater := func(ctx TestAppCtx) {
		if ebiten.IsKeyPressed(ebiten.KeyR) {
			delta := mapBool(ebiten.IsKeyPressed(ebiten.KeyShift), 0.03, -0.03)
			rotation = uradsAddCW(rotation, delta)
		}
	}
	drawer := func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Fill(color.Black)
		w, h := rectSizeF32(canvas.Bounds())

		const Rows, Cols = 2, 2
		rowHeight := h / Rows
		colWidth := w / Cols

		ref := PtF32(CircRadius, 0)
		for y := range Rows {
			cy := float32(y)*rowHeight + rowHeight/2.0
			for x := range Cols {
				cx := float32(x)*colWidth + colWidth/2.0
				ctx.Renderer.SetColorF32(0.3, 0.3, 0.3, 0.3)
				ctx.Renderer.FillCircle(canvas, cx, cy, CircRadius)

				ctx.Renderer.SetColorF32(1, 1, 1, 1)
				p := ref.Rotate(rotation)
				ctx.Renderer.FillCircle(canvas, cx, cy, PointRadius)
				ctx.Renderer.FillCircle(canvas, cx+p.X, cy+p.Y, PointRadius)
				ref = ref.Rotate(AngleDelta)
			}
		}
	}

	app := NewTestApp(updater, drawer)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestStudyGaussian$ . -count 1
func TestStudyGaussian(t *testing.T) {
	const Radius = 64.0
	const H = 192.0
	const Sigma = Radius / 3.0
	const Sigma2 = 2.0 * Sigma * Sigma
	const MarkerRadius = 2.0

	updater := func(TestAppCtx) {}
	drawer := func(canvas *ebiten.Image, ctx TestAppCtx) {
		w, h := rectSizeF64(canvas.Bounds())
		canvas.Fill(color.Black)

		gaussian := func(x float64) float32 {
			return float32(H * math.Exp(-(x*x)/Sigma2))
		}

		cx, by := float32(w*0.5), float32(h*0.666)
		ctx.Renderer.FillCircle(canvas, cx, by-gaussian(0.0), MarkerRadius)
		for i := range 30 {
			x := float64(i+1) * (Radius * 0.05)
			y := gaussian(x)
			ctx.Renderer.FillCircle(canvas, cx+float32(x), by-y, MarkerRadius)
			ctx.Renderer.FillCircle(canvas, cx-float32(x), by-y, MarkerRadius)
		}
	}

	app := NewTestApp(updater, drawer)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestStudyColorInterpolation$ . -count 1
func TestStudyColorInterpolation(t *testing.T) {
	triangles := [][3]PointF32{
		{PtF32(0.25, 0.25), PtF32(0.75, 0.25), PtF32(0.25, 0.75)},
		{PtF32(0.50, 0.75), PtF32(0.90, 0.10), PtF32(0.90, 0.90)},
		{PtF32(0.10, 0.25), PtF32(0.90, 0.75), PtF32(0.25, 0.90)},
		{PtF32(0.50, 0.25), PtF32(0.75, 0.75), PtF32(0.25, 0.75)},
		{PtF32(0.50, 0.75), PtF32(0.75, 0.25), PtF32(0.25, 0.25)},
		{PtF32(0.50, 0.75), PtF32(0.75, 0.25), PtF32(0.80, 0.25)},
	}
	triIndex := 0

	indices := [3]uint32{0, 1, 2}
	var verts [3]ebiten.Vertex
	for i := range verts {
		verts[i].SrcX = 1.5
		verts[i].SrcY = 1.5
	}
	white := ebiten.NewImage(3, 3)
	white.Fill(color.RGBA{255, 255, 255, 255})

	updater := func(ctx TestAppCtx) {
		triIndex = updateParam(ctx, ebiten.KeyT, triIndex, 0, len(triangles)-1, 1)
	}
	drawer := func(canvas *ebiten.Image, ctx TestAppCtx) {
		// draw 3 rects, with space for text on the left
		// draw with color. simple FillRect based on points.
		// then have multiple triangles to switch through
		// within the bounds
		w, h := rectSizeF32(canvas.Bounds())
		canvas.Fill(color.Black)

		ctx.Renderer.SetColorF32(1, 1, 1, 1)
		info := fmt.Sprintf("Triangle: %d [T]\nDark Overlay [Hold Space]\nTri/Quad Interp [Hold Q]", triIndex)
		ctx.Renderer.Text(canvas, info, 12, 12, TextOpts(1.0, TopLeft.Snap(CapLine)))

		tlClr := [4]float32{1, 0, 0, 1}
		trClr := [4]float32{0, 1, 1, 1}
		brClr := [4]float32{0, 0, 1, 1}
		blClr := [4]float32{1, 1, 0, 1}
		ctx.Renderer.SetColorF32A(tlClr, 0)
		ctx.Renderer.SetColorF32A(trClr, 1)
		ctx.Renderer.SetColorF32A(brClr, 2)
		ctx.Renderer.SetColorF32A(blClr, 3)

		origins := []PointF32{PtF32(w*0.25+8, 16), PtF32(16, h*0.25+8), PtF32(w*0.25+8, h*0.25+8)}
		sizes := []PointF32{PtF32(w*0.75-8-16, h*0.25-8-16), PtF32(w*0.25-8-16, h*0.75-8-16), PtF32(w*0.75-8-16, h*0.75-8-16)}
		for i := range origins {
			ctx.Renderer.FillRect(canvas, origins[i].X, origins[i].Y, sizes[i].X, sizes[i].Y, 0)
		}

		// dark overlay on space press
		if ctx.SpacePressed {
			ctx.Renderer.SetColorF32(0, 0, 0, 0.15)
			for i := range origins {
				ctx.Renderer.FillRect(canvas, origins[i].X, origins[i].Y, sizes[i].X, sizes[i].Y, 0)
			}
		}

		tri := triangles[triIndex]
		clrInterpFunc := interpTriQuadColor
		if ebiten.IsKeyPressed(ebiten.KeyQ) {
			clrInterpFunc = interpQuadColor
		}
		for i := range origins {
			for v := range 3 {
				p := sizes[i].Mul(tri[v])
				verts[v].DstX = origins[i].X + p.X
				verts[v].DstY = origins[i].Y + p.Y
				clr := clrInterpFunc(tlClr, trClr, brClr, blClr, PtF32(0, 0), sizes[i], p)
				verts[v].ColorR = clr[0]
				verts[v].ColorG = clr[1]
				verts[v].ColorB = clr[2]
				verts[v].ColorA = clr[3]
			}
			canvas.DrawTriangles32(verts[:], indices[:], white, nil)
		}
	}

	app := NewTestApp(updater, drawer)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}
