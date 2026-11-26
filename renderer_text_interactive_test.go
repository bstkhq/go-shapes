//go:build interactive

package shapes

import (
	"fmt"
	"image"
	"math"
	"testing"
	"unicode/utf8"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// go test -tags interactive -run ^TestText$ . -count 1
func TestText(t *testing.T) {
	fmt.Print(TextOptions{}.fontMap().DebugInfo())

	vertAlignIndex, horzAlignIndex, anchorIndex := 0, 0, 0
	align := TopLeft
	skipMissing := false
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
		skipMissing = updateToggle(ctx, ebiten.KeyM, skipMissing)

		vAlign := []TextAlign{vertStart, vertMiddle, vertEnd}[vertAlignIndex]
		hAlign := []TextAlign{horzStart, horzMiddle, horzEnd}[horzAlignIndex]
		align = vAlign | hAlign
		if vAlign != vertMiddle {
			align |= TextAlign([]LineAnchor{Top, CapLine, MidLine, Baseline, Bottom}[anchorIndex])
		}
	}

	drawer := func(canvas *ebiten.Image, ctx TestAppCtx) {
		info := fmt.Sprintf(
			"Scale: %.02f [S]\nAlign: %s (up/down, left/right, A, A+Shift)\nGap: %d (G)\nSkipMissing: %t [M]",
			scale, align.String(), TextOptions{LineGap: lineGap}.lineGap(), skipMissing,
		)
		ctx.Renderer.SetColorF32A(tcWhite)
		ctx.Renderer.Text(canvas, info, 8, 8, TextOpts(1.0, TopLeft.Snap(CapLine)))

		opts := TextOpts(scale, align).SkipMissing(skipMissing)
		opts.LineGap = lineGap
		cbounds := canvas.Bounds()
		cw, ch := cbounds.Dx(), cbounds.Dy()
		cx, cy := cw/2, ch/2
		ctx.Renderer.SetColorF32A(ScaleAlphaF32A(tcWhite, 0.4))
		ctx.Renderer.FillIntRect(canvas, RectWithSize(0, cy, cw, 1), 0)
		ctx.Renderer.FillIntRect(canvas, RectWithSize(cx, 0, 1, ch), 0)
		setColorsTB(ctx.Renderer, tcWhite, tcAzure)
		Paint(canvas, image.Rect(cx, cy, cx+8, cy+8), [4]float32{0.5, 0.5, 0.5, 0.5}, ebiten.BlendSourceOver)
		text := "Hellö World!\n¡HELLO WORLD!\nMoRE" + string(utf8.RuneError) + "_CoNTeNT"
		ctx.Renderer.Text(canvas, text, float32(cx), float32(cy), opts)

		abounds := ark10pxAtlas.Bounds()
		ctx.DrawAtF32(canvas, ark10pxAtlas, float32(cbounds.Max.X-8-abounds.Dx()), float32(cbounds.Max.Y-8-abounds.Dy()))
	}

	app := NewTestApp(updater, drawer)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -tags interactive -run ^TestTextAnim$ . -count 1
func TestTextAnim(t *testing.T) {
	updater := func(TestAppCtx) {}
	drawer := func(canvas *ebiten.Image, ctx TestAppCtx) {
		const Text = "Animation Test\nColor and SIZE\n(and position...)"

		canvas.Fill(backTestColor)
		w, h := rectSizeF32(canvas.Bounds())
		cx, cy := w/2.0, h/2.0
		scale := float32(1.0 + ctx.DistAnim(2.0, 0.5))
		xoff := float32(8.0 * math.Sin(ctx.RadsAnim(0.777)))
		yoff := float32(8.0 * math.Cos(ctx.RadsAnim(0.636)))

		ctx.Renderer.SetColorF32(float32(ctx.DistAnim(1.0, 0.33)), 0.5, 0.5, 1.0, 0)
		ctx.Renderer.SetColorF32(0.5, 0.75, float32(ctx.DistAnim(1.0, 0.5)), 1.0, 1)
		ctx.Renderer.SetColorF32(0.5, float32(ctx.DistAnim(1.0, 0.4)), 1.0, 1.0, 2)
		ctx.Renderer.SetColorF32(float32(ctx.DistAnim(1.0, 0.17)), 0.5, float32(ctx.DistAnim(1.0, 0.57)), 1.0, 3)
		if ctx.SpacePressed {
			ctx.Renderer.Options().Blend = ebiten.BlendCopy
		}
		ctx.Renderer.Text(canvas, Text, cx+xoff, cy+yoff, TextOpts(scale, Center).Quantized(false))
		ctx.Renderer.Options().Blend = ebiten.BlendSourceOver
	}

	app := NewTestApp(updater, drawer)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}
