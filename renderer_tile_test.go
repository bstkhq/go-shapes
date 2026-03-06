package shapes

import (
	"image/color"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

// go test -run ^TestTileRectsGrid$ . -count 1
func TestTileRectsGrid(t *testing.T) {
	app := NewTestApp(func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Fill(color.Black)
		ctx.Renderer.SetColorF32(0.9, 0, 0, 1.0)
		ctx.Renderer.TileRectsGrid(canvas, 32, 48, 24, 24, 0, 0)
	})
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestTileDotsHex$ . -count 1
func TestTileDotsHex(t *testing.T) {
	app := NewTestApp(func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Fill(color.Black)
		ctx.Renderer.SetColor(color.RGBA{0, 192, 0, 255})
		ctx.Renderer.TileDotsHex(canvas, 8, 24, 0, 0)
		ctx.Renderer.SetColor(color.RGBA{192, 0, 192, 255})
		xOffset := float32(ctx.DistAnim(6, 1.0))
		yOffset := float32(ctx.DistAnim(6, 0.5))
		ctx.Renderer.TileDotsHex(canvas, 4, 12, xOffset, yOffset)
	})
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestTileDotsGrid$ . -count 1
func TestTileDotsGrid(t *testing.T) {
	app := NewTestApp(func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Fill(color.Black)
		ctx.Renderer.SetColor(color.RGBA{220, 64, 32, 255})
		ctx.Renderer.TileDotsGrid(canvas, 8, 24, 0, 0)

		ctx.Renderer.SetColor(color.RGBA{32, 64, 250, 255})
		ctx.Renderer.TileDotsGrid(canvas, 5, 24, 12, 12)
	})
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestTileTriUpGrid$ . -count 1
func TestTileTriUpGrid(t *testing.T) {
	const inSize = 24
	app := NewTestApp(func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Fill(color.Black)

		ctx.Renderer.SetColorF32(0, 0.3, 0.3, 0.5)
		ctx.Renderer.TileRectsGrid(canvas, 32, 32*0.86602540378, inSize, inSize, 0, 0)

		ctx.Renderer.SetColorF32(1.0, 0, 1.0, 1.0)
		ctx.Renderer.TileTriUpGrid(canvas, 32, inSize, 0, 0)
	})
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestTileTriHex$ . -count 1
func TestTileTriHex(t *testing.T) {
	const minInSize = 12
	const maxInSize = 30
	app := NewTestApp(func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Fill(color.Black)

		xOff, yOff := float32(ctx.DistAnim(64, 0.5)), float32(ctx.DistAnim(32, 0.5))
		dist := ctx.DistAnim(maxInSize-minInSize, 1.0)
		ctx.Renderer.SetColorF32(1.0, 0, 1.0, 1.0)
		ctx.Renderer.TileTriHex(canvas, 32, minInSize+float32(dist), xOff, yOff)
	})
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}
