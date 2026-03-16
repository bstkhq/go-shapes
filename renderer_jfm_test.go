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

// go test -run ^TestJFMExpand$ . -count 1
func TestJFMExpand(t *testing.T) {
	const BaseRadius = 128
	const XW, XH = BaseRadius * 2, (BaseRadius * 7) / 4
	const XMargin, XThick = BaseRadius / 3, BaseRadius / 8
	const Circ2Radius = 72

	imgIndex := 0
	app := NewTestApp(func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Fill(color.Black)

		if ctx.NewInput && inpututil.IsKeyJustPressed(ebiten.KeySpace) {
			imgIndex = (imgIndex + 1) % len(ctx.Images)
		}

		bw, bh := rectSizeF32(canvas.Bounds())
		w, h := rectSizeF32(ctx.Images[imgIndex].Bounds())
		ctx.DrawAtF32(canvas, ctx.Images[imgIndex], bw/4-w/2, bh/4-h/2)
		r := float32(ctx.DistAnim(16.0, 1.0))
		if imgIndex == 2 {
			ctx.Renderer.SetColorF32(0.75, 0.5, 1.0, 1.0)
			ctx.Renderer.DrawCircle(canvas, bw-bw/4, bh/4, Circ2Radius+16.0)
			ctx.Renderer.DrawCircle(canvas, bw/4, bh-bh/4, Circ2Radius+16.0)
			ctx.Renderer.DrawCircle(canvas, bw-bw/4, bh-bh/4, Circ2Radius+16.0)
			ctx.Renderer.SetColorF32(1.0, 1.0, 1.0, 1.0)
		}

		ctx.Renderer.SetColorF32(0.5, 0.75, 1.0, 1.0)
		ctx.Renderer.SetTint(float32(ctx.DistAnim(1.0, 1.0)))

		ctx.Renderer.ApplyExpansion(canvas, ctx.Images[imgIndex], bw-bw/4-w/2, bh/4-h/2, r)
		ctx.Renderer.JFMExpand(canvas, ctx.Images[imgIndex], nil, bw/4-w/2, bh-bh/4-h/2, r, false)

		maxDist := ceilF32(r)
		source, jfmap := ctx.Renderer.UnsafeTempDual(1, ctx.Images[imgIndex], int(maxDist), false)
		ctx.Renderer.JFMapFill(jfmap, source, int(maxDist), 0.001, 1.0)
		yShift := float32(-4.0 + ctx.DistAnim(8.0, 1.0))
		ctx.Renderer.JFMExpand(canvas, source, jfmap, bw-bw/4-w/2-maxDist, bh-bh/4-h/2-maxDist+yShift, r, true)
	})

	circle := ebiten.NewImage(BaseRadius*2, BaseRadius*2)
	gradientOpts := GradientOpts(color.RGBA{0, 196, 255, 255}, color.RGBA{0, 0, 0, 0}, false)
	gradientOpts.Bias = 0.5
	app.Renderer.GradientRadial(circle, gradientOpts, BaseRadius, BaseRadius, BaseRadius*0.25, BaseRadius*0.75, BaseRadius)
	app.Renderer.opts.Blend = ebiten.BlendSourceAtop
	gradientOpts = GradientOpts(color.RGBA{196, 64, 0, 196}, color.RGBA{64, 16, 0, 64}, false)
	gradientOpts.Bias = -0.25
	app.Renderer.Gradient(circle, gradientOpts, DirRadsBLTR)
	app.Renderer.opts.Blend = ebiten.BlendSourceOver
	app.Renderer.SetColorF32(0.75, 0.5, 1.0, 1.0)
	xSign := ebiten.NewImage(XW, XH)
	app.Renderer.DrawLine(xSign, XMargin, XMargin, XW-XMargin, XH-XMargin, XThick)
	app.Renderer.DrawLine(xSign, XW-XMargin, XMargin, XMargin, XH-XMargin, XThick)
	circ2 := ebiten.NewImage(Circ2Radius*2, Circ2Radius*2)
	app.Renderer.SetColorF32(0.5, 0.25, 0.75, 1.0)
	app.Renderer.DrawCircle(circ2, Circ2Radius, Circ2Radius, Circ2Radius)
	app.Renderer.SetColorF32(1.0, 1.0, 1.0, 1.0)
	rect := app.Renderer.NewRect(256, 192)
	app.Images = append(app.Images, circle, xSign, circ2, rect)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestJFMExpandSoftMotion$ . -count 1
func TestJFMExpandSoftMotion(t *testing.T) {
	const BaseRadius = 64
	const XW, XH = BaseRadius * 2, (BaseRadius * 7) / 4
	const XMargin, XThick = BaseRadius / 3, BaseRadius / 8

	app := NewTestApp(func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Fill(color.Black)

		bw, bh := rectSizeF32(canvas.Bounds())
		yShift := float32(-4.0 + ctx.DistAnim(8.0, 1.0))
		xShift := float32(-4.0 + ctx.DistAnim(8.0, 0.77))

		const r = 9.0
		smooth := !ebiten.IsKeyPressed(ebiten.KeySpace)
		i0w, i0h := rectSizeF32(ctx.Images[0].Bounds())
		ctx.Renderer.ApplyExpansion(canvas, ctx.Images[0], bw/4-i0w/2+xShift, bh/4-i0h/2+yShift, r)
		ctx.Renderer.JFMExpand(canvas, ctx.Images[0], nil, bw/4-i0w/2+xShift, bh-bh/4-i0h/2+yShift, r, smooth)

		i1w, i1h := rectSizeF32(ctx.Images[1].Bounds())
		ctx.Renderer.ApplyExpansion(canvas, ctx.Images[1], bw-bw/4-i1w/2+xShift, bh/4-i1h/2+yShift, r)
		ctx.Renderer.JFMExpand(canvas, ctx.Images[1], nil, bw-bw/4-i1w/2+xShift, bh-bh/4-i1h/2+yShift, r, smooth)
	})

	circle := ebiten.NewImage(BaseRadius*2, BaseRadius*2)
	app.Renderer.DrawCircle(circle, BaseRadius, BaseRadius, BaseRadius)
	xSign := ebiten.NewImage(XW, XH)
	app.Renderer.DrawLine(xSign, XMargin, XMargin, XW-XMargin, XH-XMargin, XThick)
	app.Renderer.DrawLine(xSign, XW-XMargin, XMargin, XMargin, XH-XMargin, XThick)
	app.Images = append(app.Images, circle, xSign)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestJFMErode$ . -count 1
func TestJFMErode(t *testing.T) {
	app := NewTestApp(func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Fill(color.Black)

		bw, bh := rectSizeF32(canvas.Bounds())
		w, h := rectSizeF32(ctx.Images[0].Bounds())
		ctx.DrawAtF32(canvas, ctx.Images[0], bw/4-w/2, bh/4-h/2)
		r := float32(ctx.DistAnim(16.0, 1.0))
		ctx.Renderer.ApplyErosion(canvas, ctx.Images[0], bw-bw/4-w/2, bh/4-h/2, r)
		ctx.Renderer.JFMErode(canvas, ctx.Images[0], nil, bw/4-w/2, bh-bh/4-h/2, r)
	})

	const BaseRadius = 128
	circle := ebiten.NewImage(BaseRadius*2, BaseRadius*2)
	gradientOpts := GradientOpts(color.RGBA{0, 196, 255, 255}, color.RGBA{0, 0, 0, 0}, false)
	gradientOpts.Bias = 0.5
	app.Renderer.GradientRadial(circle, gradientOpts, BaseRadius, BaseRadius, BaseRadius*0.25, BaseRadius*0.75, BaseRadius)
	gradientOpts = GradientOpts(color.RGBA{196, 64, 0, 196}, color.RGBA{64, 16, 0, 64}, false)
	gradientOpts.Bias = -0.25
	app.Renderer.Gradient(circle, gradientOpts, DirRadsBLTR)
	app.Renderer.opts.Blend = ebiten.BlendSourceOver
	app.Images = append(app.Images, circle)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestJFMHeat$ . -count 1
func TestJFMHeat(t *testing.T) {
	app := NewTestApp(func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Clear()
		px, py := ebiten.CursorPosition()
		rx, ry := ctx.RightClickF32()
		ctx.Renderer.DrawCircle(canvas, float32(px), float32(py), 128.0)
		ctx.Renderer.DrawCircle(canvas, rx, ry, 96.0)
		if !ebiten.IsKeyPressed(ebiten.KeySpace) {
			const MaxDist = 128
			bounds := canvas.Bounds()
			jfmap := ctx.Renderer.UnsafeTemp(1, bounds.Dx(), bounds.Dy(), false)
			ctx.Renderer.JFMapFill(jfmap, canvas, MaxDist, 0.001, 1.0)
			ctx.Renderer.JFMHeat(canvas, jfmap, 0, 0, MaxDist)
		}
	})

	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}
