package shapes

import (
	"fmt"
	"image"
	"image/color"
	"slices"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
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
	r.JFMapBoundary(dst, src, 4, 0.001, 1.0, false, false)

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
	r.JFMapBoundary(dst, src, 257, 0.001, 1.0, false, false)

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
	app := NewTestApp(func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Fill(color.Black)

		bw, bh := rectSizeF32(canvas.Bounds())
		w, h := rectSizeF32(ctx.Images[0].Bounds())
		ctx.DrawAtF32(canvas, ctx.Images[0], bw/4-w/2, bh/4-h/2)
		r := float32(ctx.DistAnim(32.0, 1.0))
		ctx.Renderer.ApplyExpansion(canvas, ctx.Images[0], bw-bw/4-w/2, bh/4-h/2, r)
		ctx.Renderer.JFMExpand(canvas, ctx.Images[0], nil, bw/4-w/2, bh-bh/4-h/2, r, AAMargin*8)
	})

	const BaseRadius = 128
	circle := ebiten.NewImage(BaseRadius*2, BaseRadius*2)
	app.Renderer.GradientRadial(circle, BaseRadius, BaseRadius, color.RGBA{0, 196, 255, 255}, color.RGBA{0, 0, 0, 0}, BaseRadius*0.25, BaseRadius*0.75, BaseRadius, -1, 3.0)
	app.Renderer.SetBlend(ebiten.BlendSourceAtop)
	app.Renderer.Gradient(circle, nil, 0, 0, color.RGBA{196, 64, 0, 196}, color.RGBA{64, 16, 0, 64}, -1.0, DirRadsBLTR, 0.75)
	app.Renderer.SetBlend(ebiten.BlendSourceOver)
	app.Images = append(app.Images, circle)
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
		r := float32(ctx.DistAnim(32.0, 1.0))
		ctx.Renderer.ApplyErosion(canvas, ctx.Images[0], bw-bw/4-w/2, bh/4-h/2, r)
		ctx.Renderer.JFMErode(canvas, ctx.Images[0], nil, bw/4-w/2, bh-bh/4-h/2, r, AAMargin)
	})

	const BaseRadius = 128
	circle := ebiten.NewImage(BaseRadius*2, BaseRadius*2)
	app.Renderer.GradientRadial(circle, BaseRadius, BaseRadius, color.RGBA{0, 196, 255, 255}, color.RGBA{0, 0, 0, 0}, BaseRadius*0.25, BaseRadius*0.75, BaseRadius, -1, 3.0)
	app.Renderer.SetBlend(ebiten.BlendSourceAtop)
	app.Renderer.Gradient(circle, nil, 0, 0, color.RGBA{196, 64, 0, 196}, color.RGBA{64, 16, 0, 64}, -1.0, DirRadsBLTR, 0.75)
	app.Renderer.SetBlend(ebiten.BlendSourceOver)
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
