package shapes

import (
	"image"
	"image/color"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

// go test -run ^TestTileRectsGrid$ . -count 1
func TestTileRectsGrid(t *testing.T) {
	var xOffset, yOffset float32
	updater := func(TestAppCtx) {
		xOffset += 0.5
		yOffset += 0.25
	}
	drawer := func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Fill(color.Black)
		ctx.Renderer.SetColorF32(0.7, 0, 0, 0.7)
		if ctx.SpacePressed {
			xOffset, yOffset = 0, 0
		}
		wAnim := float32(ctx.DistAnim(12.0, 1.0))
		ctx.Renderer.TileRectsGrid(canvas, 24+wAnim, 24, 32, 48, xOffset, yOffset)
		Paint(canvas, image.Rect(0, 0, 32, 48), [4]float32{0.5, 0.5, 0.5, 0.5}, ebiten.BlendSourceOver)
	}
	app := NewTestApp(updater, drawer)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestTileDotsHex$ . -count 1
func TestTileDotsHex(t *testing.T) {
	updater := func(TestAppCtx) {}
	drawer := func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Fill(color.Black)
		ctx.Renderer.SetColor(color.RGBA{0, 192, 0, 255})
		ctx.Renderer.TileDotsHex(canvas, 8, 24, 0, 0)
		ctx.Renderer.SetColor(color.RGBA{192, 0, 192, 255})
		xOffset := float32(ctx.DistAnim(6, 1.0))
		yOffset := float32(ctx.DistAnim(6, 0.5))
		ctx.Renderer.TileDotsHex(canvas, 4, 12, xOffset, yOffset)

		h := 24.0 * Sqrt3Div2
		Paint(canvas, image.Rect(12, 0, 12+24, int(h)), [4]float32{0.5, 0.5, 0.5, 0.5}, ebiten.BlendSourceOver)
		Paint(canvas, image.Rect(0, int(h), 24, int(h*2)), [4]float32{0.5, 0.5, 0.5, 0.5}, ebiten.BlendSourceOver)
	}
	app := NewTestApp(updater, drawer)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestTileDotsGrid$ . -count 1
func TestTileDotsGrid(t *testing.T) {
	updater := func(TestAppCtx) {}
	drawer := func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Fill(color.Black)
		ctx.Renderer.SetColor(color.RGBA{220, 64, 32, 255})
		ctx.Renderer.TileDotsGrid(canvas, 8, 24, 0, 0)

		ctx.Renderer.SetColor(color.RGBA{32, 64, 250, 255})
		ctx.Renderer.TileDotsGrid(canvas, 5, 24, 12, 12)

		Paint(canvas, image.Rect(0, 0, 24, 24), [4]float32{0.5, 0.5, 0.5, 0.5}, ebiten.BlendSourceOver)
	}
	app := NewTestApp(updater, drawer)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestTileTriUpGrid$ . -count 1
func TestTileTriUpGrid(t *testing.T) {
	const inSize = 24
	updater := func(TestAppCtx) {}
	drawer := func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Fill(color.Black)

		ctx.Renderer.SetColorF32(0, 0.3, 0.3, 0.5)
		ctx.Renderer.TileRectsGrid(canvas, inSize, inSize, 32, 32*Sqrt3Div2, 0, 0)

		ctx.Renderer.SetColorF32(1.0, 0, 1.0, 1.0)
		ctx.Renderer.TileTriUpGrid(canvas, inSize, 32, 0, 0)
	}

	app := NewTestApp(updater, drawer)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestTileTriHex$ . -count 1
func TestTileTriHex(t *testing.T) {
	const minInSize = 12
	const maxInSize = 30

	updater := func(TestAppCtx) {}
	drawer := func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Fill(color.Black)

		xOff, yOff := float32(ctx.DistAnim(64, 0.5)), float32(ctx.DistAnim(32, 0.5))
		dist := ctx.DistAnim(maxInSize-minInSize, 1.0)
		ctx.Renderer.SetColorF32(1.0, 0, 1.0, 1.0)
		ctx.Renderer.TileTriHex(canvas, minInSize+float32(dist), 32, xOff, yOff)
	}

	app := NewTestApp(updater, drawer)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}
