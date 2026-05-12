package shapes

import (
	"fmt"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// go test -run ^TestText$ . -count 1
func TestText(t *testing.T) {
	fmt.Printf("Debug font info:\n")
	fmt.Printf(">> Glyphs per row: %d\n", ark10pxMap[fontMapIdxGlyphsPerRow])
	fmt.Printf(">> Glyph frame width: %d\n", ark10pxMap[fontMapIdxGlyphFrameWidth])
	fmt.Printf(">> Ascent: %d\n", ark10pxMap[fontMapIdxAscent])
	fmt.Printf(">> Descent: %d\n", ark10pxMap[fontMapIdxDescent])
	fmt.Printf(">> LineGap: %d\n", ark10pxMap[fontMapIdxLineGap])
	fmt.Printf(">> CapHeight: %d\n", ark10pxMap[fontMapIdxCapHeight])
	fmt.Printf(">> MidHeight: %d\n", ark10pxMap[fontMapIdxMidHeight])
	fmt.Printf(">> Space width: %d\n", ark10pxMap[fontMapIdxSpaceWidth])

	vertAlignIndex, horzAlignIndex, anchorIndex := 0, 0, 0
	align := TopLeft
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

		vAlign := []TextAlign{vertStart, vertMiddle, vertEnd}[vertAlignIndex]
		hAlign := []TextAlign{horzStart, horzMiddle, horzEnd}[horzAlignIndex]
		align = vAlign | hAlign
		if vAlign != vertMiddle {
			align |= TextAlign([]LineAnchor{Top, CapLine, MidLine, Baseline, Bottom}[anchorIndex])
		}
		ebiten.SetWindowTitle(fmt.Sprintf("%s [Scale: %.02f, Align: %s (up/down, left/right, A, A+Shift)]", ctx.Title(), scale, align.String()))
	}
	drawer := func(canvas *ebiten.Image, ctx TestAppCtx) {
		opts := TextOpts(scale, align)
		cbounds := canvas.Bounds()
		cw, ch := cbounds.Dx(), cbounds.Dy()
		cx, cy := cbounds.Min.X+cw/2, cbounds.Min.Y+ch/2
		ctx.Renderer.SetColorF32(0.4, 0.4, 0.4, 0.4)
		ctx.Renderer.FillIntRect(canvas, RectWithSize(0, cy, cw, 1), 0)
		ctx.Renderer.FillIntRect(canvas, RectWithSize(cx, 0, 1, ch), 0)
		ctx.Renderer.SetColorF32(1.0, 1.0, 1.0, 1.0, 0, 1)
		ctx.Renderer.SetColorF32(0.0, 0.5, 1.0, 1.0, 2, 3)
		ctx.Renderer.Text(canvas, "Hellö World!\n¡HELLO WORLD!\nMORE CONTENT", float32(cx), float32(cy), opts)

		abounds := ark10pxAtlas.Bounds()
		ctx.DrawAtF32(canvas, ark10pxAtlas, float32(cbounds.Max.X-8-abounds.Dx()), float32(cbounds.Max.Y-8-abounds.Dy()))
	}

	app := NewTestApp(updater, drawer)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}
