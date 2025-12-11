package shapes

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

// go test -run ^TestStudyWaveFuncs$ . -count 1
func TestStudyWaveFuncs(t *testing.T) {
	app := NewTestApp(func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Fill(color.Black)
		ctx.Renderer.studyWaveFuncs(canvas, 1.0, 8.0)
	})
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestStudyRadians$ . -count 1
func TestStudyRadians(t *testing.T) {
	const PointRadius = 3.5
	const Dist = 96.0

	rads := -math.Pi
	app := NewTestApp(func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Fill(color.Black)
		w, h := rectSizeF32(canvas.Bounds())
		cx, cy := w/2.0, h/2.0
		ctx.Renderer.DrawCircle(canvas, cx, cy, PointRadius)

		sin, cos := math.Sincos(normURads(rads))
		ctx.Renderer.DrawCircle(canvas, cx+float32(Dist*cos), cy+float32(Dist*sin), PointRadius)
		rads += 0.01
		ebiten.SetWindowTitle(fmt.Sprintf("%s - rads = %02f", ctx.Title(), rads))
	})
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestStudyBounds$ . -count 1
func TestStudyBounds(t *testing.T) {
	// tldr:
	//  - when drawing from a subimage, its origin will be considered the top-left
	//  - when drawing into a subimage, its origin will affect where the source is drawn
	app := NewTestApp(func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Fill(color.Black)

		lx, ly := ctx.LeftClickF32()
		ctx.DrawAtF32(canvas, ctx.Images[0], lx, ly)

		rx, ry := ctx.RightClickF32()
		ctx.DrawAtF32(canvas, ctx.Images[1], rx, ry)
	})

	a := ebiten.NewImage(128, 128)
	a.Fill(color.RGBA{255, 0, 0, 255})

	b := ebiten.NewImageWithOptions(image.Rect(-64, -64, 64, 64), nil)
	b.Fill(color.RGBA{0, 128, 0, 128})
	b.SubImage(image.Rect(0, 0, 64, 64)).(*ebiten.Image).Fill(color.RGBA{0, 255, 0, 255})

	sub := ebiten.NewImageWithOptions(image.Rect(-16, -16, 16, 16), nil)
	sub.Fill(color.RGBA{0, 0, 128, 128})
	sub.SubImage(image.Rect(-16, -16, 0, 0)).(*ebiten.Image).Fill(color.RGBA{0, 0, 255, 255})
	a.DrawImage(sub, nil)
	b.DrawImage(sub, nil)

	app.Images = append(app.Images, a, b)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}
