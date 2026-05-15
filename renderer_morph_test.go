package shapes

import (
	"fmt"
	"image"
	"image/color"
	"slices"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// go test -run ^TestMorphExpansion$ . -count 1
func TestMorphExpansion(t *testing.T) {
	const radius, expansion = 64.0, 16.0

	updater := func(TestAppCtx) {}
	drawer := func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Fill(color.Black)

		lc := ctx.LeftClickF32()
		ctx.Renderer.SetColor(color.RGBA{0, 128, 128, 255})
		ctx.Renderer.FillCircle(canvas, lc.X, lc.Y, radius+expansion)
		ctx.Renderer.SetColor(color.RGBA{128, 0, 0, 128})
		x := float32(ctx.DistAnim(float64(expansion), 1.0))
		ctx.Renderer.MorphExpansion(canvas, ctx.Images[0], lc.X-radius, lc.Y-radius, x)
	}

	app := NewTestApp(updater, drawer)
	app.Images = append(app.Images, app.Renderer.NewFilledCircle(float64(radius)))
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestMorphExpansionRect$ . -count 1
func TestMorphExpansionRect(t *testing.T) {
	const Radius = 64.0
	const Expansion = 16.0

	updater := func(TestAppCtx) {}
	drawer := func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Fill(color.Black)

		lc := ctx.LeftClickF32()
		rc := ctx.RightClickF32()
		ctx.Renderer.SetColor(color.RGBA{255, 0, 0, 255})
		expansion := float32(ctx.DistAnim(Expansion, 1.0))
		ctx.Renderer.MorphExpansionRect(canvas, ctx.Images[0], lc.X-Radius, lc.Y-Radius, expansion)
		ctx.Renderer.MorphExpansionRect(canvas, ctx.Images[1], rc.X-Radius, rc.Y-Radius, expansion)

		ctx.Renderer.SetColor(color.RGBA{255, 0, 0, 255})
		ctx.DrawWithAlphaAtF32(canvas, ctx.Images[0], 0.5, lc.X-Radius, lc.Y-Radius)
		ctx.DrawWithAlphaAtF32(canvas, ctx.Images[3], 0.5, lc.X-Radius-Expansion, lc.Y-Radius-Expansion)
		ctx.DrawWithAlphaAtF32(canvas, ctx.Images[1], 0.5, rc.X-Radius, rc.Y-Radius)
		ctx.DrawWithAlphaAtF32(canvas, ctx.Images[2], 0.5, rc.X-Radius-Expansion, rc.Y-Radius-Expansion)
	}

	app := NewTestApp(updater, drawer)
	img1 := app.Renderer.NewFilledRect(int(Radius*2), int(Radius*2))
	img4 := app.Renderer.NewFilledRect(int((Radius+Expansion)*2), int((Radius+Expansion)*2))
	img2 := app.Renderer.NewFilledCircle(float64(Radius))
	img3 := app.Renderer.NewFilledCircle(float64(Radius + Expansion))
	app.Images = append(app.Images, img1, img2, img3, img4)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestMorphErosion$ . -count 1
func TestMorphErosion(t *testing.T) {
	const radius, erosion = 96.0, 16.0

	updater := func(TestAppCtx) {}
	drawer := func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Fill(color.Black)

		lc := ctx.LeftClickF32()
		ctx.Renderer.SetColor(color.RGBA{255, 255, 255, 255})
		ctx.Renderer.FillCircle(canvas, lc.X, lc.Y, radius)

		ctx.Renderer.SetColor(color.RGBA{255, 0, 0, 255})
		ctx.Renderer.FillCircle(canvas, lc.X, lc.Y, radius-erosion)

		r := float32(ctx.DistAnim(float64(erosion), 1.0))
		ctx.Renderer.SetColor(color.RGBA{0, 0, 164, 164})
		ctx.Renderer.MorphErosion(canvas, ctx.Images[0], lc.X-radius, lc.Y-radius, r)

		rc := ctx.RightClickF32()
		ctx.Renderer.SetColor(color.RGBA{172, 0, 224, 255})
		ctx.Renderer.MorphErosion(canvas, ctx.Images[1], rc.X-128, rc.Y-82, r)
	}

	app := NewTestApp(updater, drawer)
	circle := app.Renderer.NewFilledCircle(float64(radius))
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

// go test -run ^TestMorphOutline$ . -count 1
func TestMorphOutline(t *testing.T) {
	const radius, thick = 64.0, 8.0

	updater := func(TestAppCtx) {}
	drawer := func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Fill(color.Black)

		lc := ctx.LeftClickF32()
		ctx.Renderer.SetColor(color.RGBA{0, 255, 0, 255})
		ctx.Renderer.FillCircle(canvas, lc.X, lc.Y, radius+thick/2+1.0)
		ctx.Renderer.SetColor(color.RGBA{255, 0, 0, 255})
		ctx.Renderer.FillCircle(canvas, lc.X, lc.Y, radius-thick/2-1.0)

		ctx.Renderer.SetColor(color.RGBA{0, 0, 255, 255})
		ctx.Renderer.MorphOutline(canvas, ctx.Images[0], lc.X-radius, lc.Y-radius, thick)

		rc := ctx.RightClickF32()
		ctx.Renderer.SetColor(color.RGBA{255, 255, 255, 255})
		if ctx.SpacePressed {
			ctx.DrawAtF32(canvas, ctx.Images[0], rc.X-radius, rc.Y-radius)
		} else {
			ctx.Renderer.MorphOutline(canvas, ctx.Images[0], rc.X-radius, rc.Y-radius, thick)
		}
	}

	app := NewTestApp(updater, drawer)
	app.Images = append(app.Images, app.Renderer.NewFilledCircle(float64(radius)))
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

const negBit = 0x80

func jfmDebugPrint(t *testing.T, out *image.RGBA) {
	const DisplayMode string = "coords" // "coords", "rgba" or "dual"
	const DebugPrint bool = false

	decodeAxisOffsetToSeed := func(a, b int) int {
		hi, lo := a, b
		magnitude := ((hi & 0x7F) << 8) | lo
		sign := 1 - ((hi >> 7) << 1)
		return sign * magnitude
	}
	fmtRGBA := func(rgba color.RGBA) string {
		x := decodeAxisOffsetToSeed(int(rgba.R), int(rgba.G))
		y := decodeAxisOffsetToSeed(int(rgba.B), int(rgba.A))
		switch DisplayMode {
		case "coords":
			return fmt.Sprintf("[%+04d %+04d]", x, y)
		case "rgba":
			return fmt.Sprintf("[%03d %03d %03d %03d]", rgba.R, rgba.G, rgba.B, rgba.A)
		case "dual":
			return fmt.Sprintf("[%+04d %+04d](%03d %03d %03d %03d)", x, y, rgba.R, rgba.G, rgba.B, rgba.A)
		default:
			panic("invalid display mode '" + DisplayMode + "'")
		}
	}
	if DebugPrint {
		for y := range out.Rect.Max.Y {
			fmt.Printf("row#%d ", y)
			for x := range out.Rect.Max.X {
				fmt.Printf("%s ", fmtRGBA(out.RGBAAt(x, y)))
			}
			fmt.Printf("\n")
		}
		t.Fatalf("debug print")
	}
}

// go test -run ^TestJFMCompute$ . -count 1
func TestJFMCompute(t *testing.T) {
	r := NewRenderer()
	src := ebiten.NewImage(9, 9)
	// top-left hollow rectangle
	src.Set(0, 0, color.White)
	src.Set(1, 0, color.White)
	src.Set(2, 0, color.White)
	src.Set(0, 1, color.White)
	src.Set(2, 1, color.White)
	src.Set(0, 2, color.White)
	src.Set(1, 2, color.White)
	src.Set(2, 2, color.White)

	dst := ebiten.NewImage(9, 9)
	r.JFMapBoundary(dst, src, 4, 0.001, 1.0, BoundaryMode{})

	out := image.NewRGBA(image.Rect(0, 0, 9, 9))
	if err := ebiten.RunGame(&testOutputWriter{subject: dst, out: out.Pix}); err != nil {
		t.Fatal(err)
	}
	jfmDebugPrint(t, out)

	expectedOut := image.NewRGBA(image.Rect(0, 0, 9, 9))
	for i := range expectedOut.Pix {
		expectedOut.Pix[i] = 255
	}
	expectedOut.SetRGBA(0, 0, color.RGBA{0, 0, 0, 0})
	expectedOut.SetRGBA(1, 0, color.RGBA{0, 0, 0, 0})
	expectedOut.SetRGBA(2, 0, color.RGBA{0, 0, 0, 0})
	expectedOut.SetRGBA(0, 1, color.RGBA{0, 0, 0, 0})
	expectedOut.SetRGBA(2, 1, color.RGBA{0, 0, 0, 0})
	expectedOut.SetRGBA(0, 2, color.RGBA{0, 0, 0, 0})
	expectedOut.SetRGBA(1, 2, color.RGBA{0, 0, 0, 0})
	expectedOut.SetRGBA(2, 2, color.RGBA{0, 0, 0, 0})

	expectedOut.SetRGBA(1, 1, color.RGBA{negBit, 1, 0, 0})
	for c := range 3 {
		for i := range 4 {
			expectedOut.SetRGBA(c, 3+i, color.RGBA{0, 0, negBit, uint8(i + 1)})
			expectedOut.SetRGBA(3+i, c, color.RGBA{negBit, uint8(i + 1), 0, 0})
		}
	}
	low := [][]color.RGBA{
		{{negBit, 1, negBit, 1}, {negBit, 2, negBit, 1}, {negBit, 3, negBit, 1}},
		{{negBit, 1, negBit, 2}, {negBit, 2, negBit, 2}, {negBit, 3, negBit, 2}},
		{{negBit, 1, negBit, 3}, {negBit, 2, negBit, 3}},
	}
	for y, row := range low {
		for x, value := range row {
			expectedOut.SetRGBA(3+x, 3+y, value)
		}
	}

	if !slices.Equal(out.Pix, expectedOut.Pix) {
		t.Fatalf("expected slices.Equal(expected, out)\nexpected:\n%v\nout:\n%v", expectedOut.Pix, out.Pix)
	}
}

// go test -run ^TestJFMCompute2$ . -count 1
func TestJFMCompute2(t *testing.T) {
	r := NewRenderer()
	src := ebiten.NewImage(1, 260)
	dst := ebiten.NewImage(1, 260)
	src.Set(0, 258, color.White)
	r.JFMapBoundary(dst, src, 257, 0.001, 1.0, BoundaryMode{})

	out := image.NewRGBA(image.Rect(0, 0, 1, 260))
	if err := ebiten.RunGame(&testOutputWriter{subject: dst, out: out.Pix}); err != nil {
		t.Fatal(err)
	}
	jfmDebugPrint(t, out)

	expectedOut := image.NewRGBA(image.Rect(0, 0, 1, 260))
	for i := range 260 {
		n := 258 - i
		if n < 256 {
			expectedOut.SetRGBA(0, i, color.RGBA{0, 0, 0, uint8(n)})
		} else {
			expectedOut.SetRGBA(0, i, color.RGBA{0, 0, 1, uint8(n - 256)})
		}
	}
	expectedOut.SetRGBA(0, 258, color.RGBA{0, 0, 0, 0})       // seed
	expectedOut.SetRGBA(0, 259, color.RGBA{0, 0, negBit, 1})  // after seed
	expectedOut.SetRGBA(0, 0, color.RGBA{255, 255, 255, 255}) // outside range

	if !slices.Equal(out.Pix, expectedOut.Pix) {
		t.Fatalf("expected slices.Equal(expected, out)\nexpected:\n%v\nout:\n%v", expectedOut.Pix, out.Pix)
	}
}

type testOutputWriter struct {
	ticks   int
	subject *ebiten.Image
	out     []byte
}

func (t *testOutputWriter) Draw(*ebiten.Image) {}
func (t *testOutputWriter) Layout(w, h int) (int, int) {
	return w, h
}
func (t *testOutputWriter) Update() error {
	t.ticks += 1
	if t.ticks == 32 {
		t.subject.ReadPixels(t.out)
	}
	if t.ticks >= 64 {
		return ebiten.Termination
	}
	return nil
}

func jfmShapes(r *Renderer) []*ebiten.Image {
	const BaseRadius = 128
	const XW, XH = BaseRadius * 2, (BaseRadius * 7) / 4
	const XMargin, XThick = BaseRadius / 3, BaseRadius / 8
	const Circ2Radius = 72

	circle := ebiten.NewImage(BaseRadius*2, BaseRadius*2)
	gradientOpts := GradientOpts(color.RGBA{0, 196, 255, 255}, color.RGBA{0, 0, 0, 0}, false)
	gradientOpts.Bias = 0.5
	r.GradientRadial(circle, gradientOpts, BaseRadius, BaseRadius, BaseRadius*0.25, BaseRadius*0.75, BaseRadius)
	r.opts.Blend = ebiten.BlendSourceAtop
	gradientOpts = GradientOpts(color.RGBA{196, 64, 0, 196}, color.RGBA{64, 16, 0, 64}, false)
	gradientOpts.Bias = -0.25
	r.Gradient(circle, gradientOpts, DirRadsBLTR)
	r.opts.Blend = ebiten.BlendSourceOver
	r.SetColorF32(0.75, 0.5, 1.0, 1.0)
	xSign := ebiten.NewImage(XW, XH)
	r.StrokeLine(xSign, XMargin, XMargin, XW-XMargin, XH-XMargin, XThick)
	r.StrokeLine(xSign, XW-XMargin, XMargin, XMargin, XH-XMargin, XThick)
	circ2 := ebiten.NewImage(Circ2Radius*2, Circ2Radius*2)
	r.SetColorF32(0.5, 0.25, 0.75, 1.0)
	r.FillCircle(circ2, Circ2Radius, Circ2Radius, Circ2Radius)
	r.SetColorF32(1.0, 1.0, 1.0, 1.0)
	rect := r.NewFilledRect(256, 192)
	return []*ebiten.Image{circle, xSign, circ2, rect}
}

// go test -run ^TestJFMExpand$ . -count 1
func TestJFMExpand(t *testing.T) {
	imgIndex := 0
	outline := false
	updater := func(ctx TestAppCtx) {
		imgIndex = updateParam(ctx, ebiten.KeySpace, imgIndex, 0, len(ctx.Images)-1, 1)
		if inpututil.IsKeyJustPressed(ebiten.KeyL) {
			outline = !outline
		}
	}
	drawer := func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Fill(color.Black)
		bw, bh := rectSizeF32(canvas.Bounds())
		w, h := rectSizeF32(ctx.Images[imgIndex].Bounds())
		ctx.DrawAtF32(canvas, ctx.Images[imgIndex], bw/4-w/2, bh/4-h/2)
		r := float32(ctx.DistAnim(16.0, 1.0))
		if imgIndex == 2 {
			ctx.Renderer.SetColorF32(0.75, 0.5, 1.0, 1.0)
			ctx.Renderer.FillCircle(canvas, bw-bw/4, bh/4, w/2+16.0)
			ctx.Renderer.FillCircle(canvas, bw/4, bh-bh/4, w/2+16.0)
			ctx.Renderer.FillCircle(canvas, bw-bw/4, bh-bh/4, w/2+16.0)
			ctx.Renderer.SetColorF32(1.0, 1.0, 1.0, 1.0)
		}

		ctx.Renderer.SetColorF32(0.5, 0.75, 1.0, 1.0)
		ctx.Renderer.SetTint(float32(ctx.DistAnim(1.0, 1.0)))

		if outline {
			ctx.Renderer.MorphOutline(canvas, ctx.Images[imgIndex], bw-bw/4-w/2, bh/4-h/2, r)
		} else {
			ctx.Renderer.MorphExpansion(canvas, ctx.Images[imgIndex], bw-bw/4-w/2, bh/4-h/2, r)
		}
		ctx.Renderer.JFMExpand(canvas, ctx.Images[imgIndex], nil, bw/4-w/2, bh-bh/4-h/2, r, outline, false)

		maxDist := ceilF32(r)
		source, jfmap := ctx.Renderer.UnsafeTempDual(1, ctx.Images[imgIndex], int(maxDist), false)
		if outline {
			ctx.Renderer.JFMapBoundary(jfmap, source, int(maxDist), 0.001, 1.0, BoundaryMode{})
		} else {
			ctx.Renderer.JFMapFill(jfmap, source, int(maxDist), 0.001, 1.0)
		}
		ctx.Renderer.JFMExpand(canvas, source, jfmap, bw-bw/4-w/2-maxDist, bh-bh/4-h/2-maxDist, r, outline, true)
	}

	app := NewTestApp(updater, drawer)
	app.Images = append(app.Images, jfmShapes(app.Renderer)...)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestJFMExpandSoftMotion$ . -count 1
func TestJFMExpandSoftMotion(t *testing.T) {
	imgIndex := 0
	updater := func(ctx TestAppCtx) {
		imgIndex = updateParam(ctx, ebiten.KeySpace, imgIndex, 0, len(ctx.Images)-1, 1)
	}
	drawer := func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Fill(color.Black)

		bw, bh := rectSizeF32(canvas.Bounds())
		yShift := float32(-4.0 + ctx.DistAnim(8.0, 1.0))
		xShift := float32(-4.0 + ctx.DistAnim(8.0, 0.77))

		const r = 9.0
		iw, ih := rectSizeF32(ctx.Images[imgIndex].Bounds())
		ctx.Renderer.MorphExpansion(canvas, ctx.Images[imgIndex], bw/4-iw/2+xShift, bh/4-ih/2+yShift, r)
		ctx.Renderer.JFMExpand(canvas, ctx.Images[imgIndex], nil, bw/4-iw/2+xShift, bh-bh/4-ih/2+yShift, r, false, false)

		ctx.Renderer.MorphExpansion(canvas, ctx.Images[imgIndex], bw-bw/4-iw/2+xShift, bh/4-ih/2+yShift, r)
		ctx.Renderer.JFMExpand(canvas, ctx.Images[imgIndex], nil, bw-bw/4-iw/2+xShift, bh-bh/4-ih/2+yShift, r, false, true)
	}

	app := NewTestApp(updater, drawer)
	app.Images = append(app.Images, jfmShapes(app.Renderer)...)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestJFMErode$ . -count 1
func TestJFMErode(t *testing.T) {
	imgIndex := 0
	motion := false

	updater := func(ctx TestAppCtx) {
		imgIndex = updateParam(ctx, ebiten.KeySpace, imgIndex, 0, len(ctx.Images)-1, 1)
		if inpututil.IsKeyJustPressed(ebiten.KeyM) {
			motion = !motion
		}
	}
	drawer := func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Fill(color.Black)

		bw, bh := rectSizeF32(canvas.Bounds())
		w, h := rectSizeF32(ctx.Images[imgIndex].Bounds())
		ctx.DrawAtF32(canvas, ctx.Images[imgIndex], bw/4-w/2, bh/4-h/2)
		r := float32(ctx.DistAnim(16.0, 1.0))
		tint := float32(ctx.DistAnim(1.0, 0.5))
		ctx.Renderer.SetTint(tint)
		mx, my := float32(0.0), float32(0.0)
		if motion {
			mx, my = float32(-4.0+ctx.DistAnim(8.0, 1.0)), float32(-4.0+ctx.DistAnim(8.0, 0.777))
			r = 6.0
		}
		ctx.Renderer.MorphErosion(canvas, ctx.Images[imgIndex], bw-bw/4-w/2+mx, bh/4-h/2+my, r)
		ctx.Renderer.JFMErode(canvas, ctx.Images[imgIndex], nil, bw/4-w/2+mx, bh-bh/4-h/2+my, r, false)
		ctx.Renderer.JFMErode(canvas, ctx.Images[imgIndex], nil, bw-bw/4-w/2+mx, bh-bh/4-h/2+my, r, true)

		// ctx.Renderer.JFMErode(canvas, ctx.Images[imgIndex], nil, bw/4-w/2, bh-bh/4-h/2, r, true)
		ctx.Renderer.SetTint(0)
	}

	app := NewTestApp(updater, drawer)
	app.Images = append(app.Images, jfmShapes(app.Renderer)...)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestJFMHeat$ . -count 1
func TestJFMHeat(t *testing.T) {
	mode := 0
	clamp, outer := false, false
	updater := func(ctx TestAppCtx) {
		mode = updateParam(ctx, ebiten.KeyM, mode, 0, 1, +1)
		switch {
		case inpututil.IsKeyJustPressed(ebiten.KeyC):
			clamp = !clamp
		case inpututil.IsKeyJustPressed(ebiten.KeyR):
			outer = !outer
		}
	}
	drawer := func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Clear()
		px, py := ebiten.CursorPosition()
		rc := ctx.RightClickF32()
		ctx.Renderer.FillCircle(canvas, float32(px), float32(py), 128.0)
		ctx.Renderer.FillCircle(canvas, rc.X, rc.Y, 96.0)
		if ebiten.IsKeyPressed(ebiten.KeySpace) {
			return
		}

		const MaxDist = 128
		bounds := canvas.Bounds()
		jfmap := ctx.Renderer.UnsafeTemp(1, bounds.Dx(), bounds.Dy(), false)
		switch mode {
		case 0:
			ebiten.SetWindowTitle(fmt.Sprintf("%s [[M]ode: %s]", ctx.Title(), "fill"))
			ctx.Renderer.JFMapFill(jfmap, canvas, MaxDist, 0.001, 1.0)
		case 1:
			// TODO: apparently clamp = false can fail in some contexts
			ebiten.SetWindowTitle(fmt.Sprintf("%s [[M]ode: %s, oute[R]: %t, [C]lamp: %t]", ctx.Title(), "boundary", outer, clamp))
			ctx.Renderer.JFMapBoundary(jfmap, canvas, MaxDist, 0.001, 1.0, BoundaryMode{Outer: outer, Clamp: clamp})
		}
		ctx.Renderer.JFMHeat(canvas, jfmap, 0, 0, MaxDist)
	}

	app := NewTestApp(updater, drawer)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}
