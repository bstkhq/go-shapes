package shapes

import (
	"fmt"
	"image/color"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// go test -run ^TestMask$ . -count 1
func TestMask(t *testing.T) {
	app := NewTestApp(func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Fill(color.Black)
		lx, ly := ctx.LeftClickF32()
		ctx.Renderer.Mask(canvas, ctx.Images[0], ctx.Images[1], lx, ly)

		rx, ry := ctx.RightClickF32()
		ctx.Renderer.Mask(canvas, ctx.Images[0], ctx.Images[2], rx, ry)
	})

	circ := app.Renderer.NewCircle(64.0)
	app.Renderer.SetColorF32(0, 0, 0, 0, 1, 2)
	bigRect := app.Renderer.NewRect(256, 128)
	smallRect := app.Renderer.NewRect(16, 8)
	app.Renderer.SetColorF32(1, 1, 1, 1)
	app.Images = append(app.Images, circ, bigRect, smallRect)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestMaskAt$ . -count 1
func TestMaskAt(t *testing.T) {
	app := NewTestApp(func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Fill(color.Black)
		lx, ly := ctx.LeftClickF32()
		dist := float32(ctx.DistAnim(256.0, 1.0))
		ctx.Renderer.MaskAt(canvas, ctx.Images[0], ctx.Images[1], lx+dist, ly, lx, ly)
	})

	circ := app.Renderer.NewCircle(32.0)
	app.Renderer.SetColorF32(0, 0, 0, 0, 1, 2)
	trans := app.Renderer.NewRect(256, 64)
	app.Renderer.SetColorF32(1, 1, 1, 1)
	app.Images = append(app.Images, circ, trans)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestMaskHorz$ . -count 1
func TestMaskHorz(t *testing.T) {
	app := NewTestApp(func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Fill(color.Black)
		lx, ly := ctx.LeftClickF32()
		x, _ := ebiten.CursorPosition()
		ctx.Renderer.MaskHorz(canvas, ctx.Images[0], lx, ly, lx+256/2.0, float32(x))
	})

	rect := app.Renderer.NewRect(256, 64)
	app.Images = append(app.Images, rect)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestMaskCircle$ . -count 1
func TestMaskCircle(t *testing.T) {
	app := NewTestApp(func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Fill(color.Black)
		w, h := rectSizeF32(canvas.Bounds())
		hardRadius := 48.0 + float32(ctx.DistAnim(16.0, 0.25))
		softEdge := float32(ctx.DistAnim(16.0, 1.0))
		ctx.Renderer.MaskCircle(canvas, ctx.Images[0], w/2, h/2, 0, 0, hardRadius, softEdge)
	})

	rect := app.Renderer.NewRect(360, 360)
	app.Images = append(app.Images, rect)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestMaskThreshold$ . -count 1
func TestMaskThreshold(t *testing.T) {
	const Size = 256
	app := NewTestApp(func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Fill(color.Black)
		reveal := -0.1 + float32(ctx.ModAnim(1.2, 0.5))
		w, h := rectSizeF32(canvas.Bounds())
		ctx.Renderer.MaskThreshold(canvas, ctx.Images[0], ctx.Images[1], reveal, w/2-Size/2, h/2-Size/2)
		ebiten.SetWindowTitle(ctx.Title() + fmt.Sprintf(" - reveal %.02f", reveal))
	})

	maskTarget := ebiten.NewImage(Size, Size)
	app.Renderer.Gradient(maskTarget, nil, 0, 0, color.RGBA{0, 0, 0, 0}, color.RGBA{255, 255, 255, 255}, 16, DirRadsLTR, 1.0)
	whiteRect := app.Renderer.NewRect(Size, Size)

	app.Images = append(app.Images, whiteRect, maskTarget)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestAlphaMaskCirc$ . -count 1
func TestAlphaMaskCirc(t *testing.T) {
	const Size = 256
	randomness := float32(0.3)

	app := NewTestApp(func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Fill(color.Black)

		w, h := rectSizeF32(canvas.Bounds())
		ox, oy := w/2-Size/2, h/2-Size/2

		if ctx.NewInput {
			switch {
			case inpututil.IsKeyJustPressed(ebiten.KeySpace):
				lx, ly := ctx.LeftClickF32()
				ctx.Renderer.Options().Blend = ebiten.BlendCopy
				ctx.Renderer.DrawAlphaMaskCirc(ctx.Images[1], lx-ox, ly-oy, Size*1.44, randomness, MaskPatternEllipseCuts)
				ctx.Renderer.Options().Blend = ebiten.BlendSourceOver
			case inpututil.IsKeyJustPressed(ebiten.KeyArrowUp):
				randomness = min(randomness+0.1, 1.0)
			case inpututil.IsKeyJustPressed(ebiten.KeyArrowDown):
				randomness = max(randomness-0.1, 0.0)
			}
		}
		ebiten.SetWindowTitle(ctx.Title() + fmt.Sprintf(" - randomness %.02f", randomness))

		reveal := -0.1 + float32(ctx.ModAnim(2.0, 0.2))
		ctx.Renderer.MaskThreshold(canvas, ctx.Images[0], ctx.Images[1], reveal, ox, oy)
	})

	maskTarget := ebiten.NewImage(Size, Size)
	whiteRect := app.Renderer.NewRect(Size, Size)
	app.Renderer.DrawAlphaMaskCirc(maskTarget, Size/2, Size/2, Size, randomness, MaskPatternDefault)
	app.Images = append(app.Images, whiteRect, maskTarget)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}
