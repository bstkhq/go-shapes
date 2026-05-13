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

// go test -run ^TestText$ . -count 1
func TestText(t *testing.T) {
	fmt.Print(TextOptions{}.fontMap().DebugInfo())

	vertAlignIndex, horzAlignIndex, anchorIndex := 0, 0, 0
	align := TopLeft
	lineGap := int8(0)
	scale := float32(1.0)
	updater := func(ctx TestAppCtx) {
		if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) {
			vertAlignIndex = (vertAlignIndex + 1) % 3
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) {
			vertAlignIndex = (vertAlignIndex + 2) % 3
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyArrowLeft) {
			horzAlignIndex = (horzAlignIndex + 1) % 3
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyArrowRight) {
			horzAlignIndex = (horzAlignIndex + 2) % 3
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyA) {
			if ebiten.IsKeyPressed(ebiten.KeyShift) {
				anchorIndex = (anchorIndex + 4) % 5
			} else {
				anchorIndex = (anchorIndex + 1) % 5
			}
		}
		scale = updateParam(ctx, ebiten.KeyS, scale, 0.5, 6.0, 0.5)
		lineGap = updateParam(ctx, ebiten.KeyG, lineGap, -3, 3, 1)

		vAlign := []TextAlign{vertStart, vertMiddle, vertEnd}[vertAlignIndex]
		hAlign := []TextAlign{horzStart, horzMiddle, horzEnd}[horzAlignIndex]
		align = vAlign | hAlign
		if vAlign != vertMiddle {
			align |= TextAlign([]LineAnchor{Top, CapLine, MidLine, Baseline, Bottom}[anchorIndex])
		}
		ebiten.SetWindowTitle(fmt.Sprintf("%s [Scale: %.02f, Align: %s (up/down, left/right, A, A+Shift), Gap: %d (G)]", ctx.Title(), scale, align.String(), TextOptions{LineGap: lineGap}.lineGap()))
	}
	drawer := func(canvas *ebiten.Image, ctx TestAppCtx) {
		opts := TextOpts(scale, align)
		opts.LineGap = lineGap
		cbounds := canvas.Bounds()
		cw, ch := cbounds.Dx(), cbounds.Dy()
		cx, cy := cw/2, ch/2
		ctx.Renderer.SetColorF32(0.4, 0.4, 0.4, 0.4)
		ctx.Renderer.FillIntRect(canvas, RectWithSize(0, cy, cw, 1), 0)
		ctx.Renderer.FillIntRect(canvas, RectWithSize(cx, 0, 1, ch), 0)
		ctx.Renderer.SetColorF32(1.0, 1.0, 1.0, 1.0, 0, 1)
		ctx.Renderer.SetColorF32(0.0, 0.5, 1.0, 1.0, 2, 3)
		Paint(canvas, image.Rect(cx, cy, cx+8, cy+8), [4]float32{0.5, 0.5, 0.5, 0.5}, ebiten.BlendSourceOver)
		ctx.Renderer.Text(canvas, "Hellö World!\n¡HELLO WORLD!\nMoRE CoNTeNT", float32(cx), float32(cy), opts)

		abounds := ark10pxAtlas.Bounds()
		ctx.DrawAtF32(canvas, ark10pxAtlas, float32(cbounds.Max.X-8-abounds.Dx()), float32(cbounds.Max.Y-8-abounds.Dy()))
	}

	app := NewTestApp(updater, drawer)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestTextAnim$ . -count 1
func TestTextAnim(t *testing.T) {
	updater := func(TestAppCtx) {}
	drawer := func(canvas *ebiten.Image, ctx TestAppCtx) {
		const Text = "Animation Test\nColor and SIZE\n(and position...)"

		canvas.Fill(color.Black)
		w, h := rectSizeF32(canvas.Bounds())
		cx, cy := w/2.0, h/2.0
		scale := float32(1.0 + ctx.DistAnim(2.0, 0.5))
		xoff := float32(8.0 * math.Sin(ctx.RadsAnim(0.777)))
		yoff := float32(8.0 * math.Cos(ctx.RadsAnim(0.636)))

		ctx.Renderer.SetColorF32(float32(ctx.DistAnim(1.0, 0.33)), 0.5, 0.5, 1.0, 0)
		ctx.Renderer.SetColorF32(0.5, 0.75, float32(ctx.DistAnim(1.0, 0.5)), 1.0, 1)
		ctx.Renderer.SetColorF32(0.5, float32(ctx.DistAnim(1.0, 0.4)), 1.0, 1.0, 2)
		ctx.Renderer.SetColorF32(float32(ctx.DistAnim(1.0, 0.17)), 0.5, float32(ctx.DistAnim(1.0, 0.57)), 1.0, 3)
		ctx.Renderer.Text(canvas, Text, cx+xoff, cy+yoff, TextOpts(scale, Center).Quantized(false))
	}

	app := NewTestApp(updater, drawer)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

func TestTextSize(t *testing.T) {
	r := NewRenderer()
	fontMap := TextOptions{}.fontMap()
	ascent := fontMap.Ascent()
	descent := fontMap.Descent()
	lineGap := fontMap.LineGap()
	spaceWidth := fontMap.SpaceWidth()

	w, h := r.TextSize("\n", TextOptions{})
	if w != 0 || h != float32(ascent+descent) {
		t.Fatalf("expected (w, h) = (%.02f, %.02f), got (%.02f, %.02f)", 0.0, float32(ascent+descent), w, h)
	}

	line2H := 2*float32(ascent+descent) + float32(lineGap)
	w, h = r.TextSize("\n\n", TextOptions{})
	if w != 0 || h != line2H {
		t.Fatalf("expected (w, h) = (%.02f, %.02f), got (%.02f, %.02f)", 0.0, line2H, w, h)
	}

	w, h = r.TextSize("\n\n", TextOptions{LineGap: 7})
	if w != 0 || h != float32(2*(ascent+descent)+7) {
		t.Fatalf("expected (w, h) = (%.02f, %.02f), got (%.02f, %.02f)", 0.0, float32(2*(ascent+descent)+7), w, h)
	}

	_, h = r.TextSize("\nA", TextOptions{})
	if h != line2H {
		t.Fatalf("expected h = %.02f, got %.02f", line2H, h)
	}
	_, h = r.TextSize("A\n", TextOptions{})
	if h != line2H {
		t.Fatalf("expected h = %.02f, got %.02f", line2H, h)
	}
	_, h = r.TextSize("A\nBCD", TextOptions{})
	if h != line2H {
		t.Fatalf("expected h = %.02f, got %.02f", line2H, h)
	}

	w, h = r.TextSize(" ", TextOptions{Scale: 1.0})
	if w != float32(spaceWidth) || h != float32(ascent+descent) {
		t.Fatalf("expected (w, h) = (%.02f, %.02f), got (%.02f, %.02f)", float32(spaceWidth), float32(ascent+descent), w, h)
	}

	wm, _ := r.TextSize("M", TextOptions{Scale: 1.0})
	w, _ = r.TextSize("MM", TextOptions{Scale: 1.0})
	if w != wm+1.0+wm {
		t.Fatalf("expected w = %.02f, got %.02f", wm+1.0+wm, w)
	}
	w, _ = r.TextSize("M M", TextOptions{Scale: 1.0})
	if w != wm+float32(spaceWidth)+wm {
		t.Fatalf("expected w = %.02f, got %.02f", wm+float32(spaceWidth)+wm, w)
	}

	s1, s2 := "MMA", "MM\nMMA"
	w1, _ := r.TextSize(s1, TextOptions{})
	w2, h := r.TextSize(s2, TextOptions{})
	if w1 != w2 {
		t.Fatalf("width mismatch between %q (%.02f) and %q (%.02f)", s1, w1, s2, w2)
	}

	ws2, hs2 := r.TextSize(s2, TextOptions{Scale: 1.5})
	if ws2 != w2*1.5 || hs2 != h*1.5 { // note: might be fuzzy for big fonts
		t.Fatalf("expected (w, h) = (%.02f, %.02f), got (%.02f, %.02f)", float32(w2*1.5), float32(h*1.5), ws2, hs2)
	}
}
