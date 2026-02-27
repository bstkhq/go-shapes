package shapes

import (
	"fmt"
	"image"
	"image/color"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// go test -run ^TestScale$ . -count 1
func TestScale(t *testing.T) {
	// TODO: test minification, unclear if I should disallow DstSampling in that case
	app := NewTestApp(func(canvas *ebiten.Image, ctx TestAppCtx) {
		_, _, w, h := rectOriginSize(canvas.Bounds())
		scale := float64(min(w, h)) / 6.0

		var opts ebiten.DrawImageOptions
		opts.GeoM.Scale(scale, scale)
		canvas.DrawImage(ctx.Images[0], &opts)

		opts.Filter = ebiten.FilterLinear
		opts.GeoM.Reset()
		opts.GeoM.Translate(-3, 0)
		opts.GeoM.Scale(scale, scale)
		opts.GeoM.Translate(float64(w), 0)
		canvas.DrawImage(ctx.Images[0], &opts)

		size := float32(3 * scale)
		var scOpts ScaleOptions
		scOpts.Clamp = ebiten.IsKeyPressed(ebiten.KeyC)
		scOpts.DstSampling = ebiten.IsKeyPressed(ebiten.KeyD)

		ctx.Renderer.Scale(canvas, ctx.Images[0], 0, float32(h)-size, float32(scale), &scOpts)
		scOpts.Bicubic = true
		ctx.Renderer.Scale(canvas, ctx.Images[0], float32(w)-size, float32(h)-size, float32(scale), &scOpts)
	})

	sq9 := ebiten.NewImage(3, 3)
	sq9.WritePixels([]byte{
		255, 0, 0, 255 /* */, 0, 128, 255, 255 /* */, 255, 128, 0, 255,
		0, 0, 255, 255 /* */, 255, 0, 255, 255 /* */, 0, 255, 0, 255,
		128, 255, 0, 255 /* */, 255, 255, 0, 255 /* */, 0, 255, 255, 255,
	})
	app.Images = append(app.Images, sq9)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

type flagList []Flag

func newFlagList() flagList {
	return make(flagList, numFlags)
}

func (l flagList) Has(f Flag) bool {
	return l[f] == f
}
func (l flagList) Flip(f Flag) {
	l[f] = -(l[f] - f)
}
func (l flagList) All() []Flag {
	return l
}

// go test -run ^TestDrawAt$ . -count 1
func TestDrawAt(t *testing.T) {
	flags := newFlagList()
	app := NewTestApp(func(canvas *ebiten.Image, ctx TestAppCtx) {
		ebiten.SetWindowTitle(fmt.Sprintf("%s [bilinear: %t, dithered: %t]", ctx.Title(), flags[Bilinear] == Bilinear, flags[Dithered] == Dithered))
		if ctx.NewInput {
			if inpututil.IsKeyJustPressed(ebiten.KeyF) {
				flags.Flip(Bilinear)
			}
			if inpututil.IsKeyJustPressed(ebiten.KeyD) {
				flags.Flip(Dithered)
			}
		}

		lx, ly := ctx.LeftClickF32()
		x, y := CTR.Adjust(ctx.Images[0], lx, ly)
		y += -8.0 + float32(ctx.DistAnim(16, 1.0))
		if !ebiten.IsKeyPressed(ebiten.KeySpace) {
			ctx.Renderer.DrawAt(canvas, ctx.Images[0], x, y, 1.0, flags.All()...)
		} else {
			mark := image.Rectangle{Max: canvas.Bounds().Max}
			mark.Min = mark.Max.Sub(image.Pt(16, 8))
			Paint(canvas, mark, [4]float32{0.5, 0.5, 0, 1.0}, ebiten.BlendSourceOver)

			var opts ebiten.DrawImageOptions
			if flags.Has(Bilinear) {
				opts.Filter = ebiten.FilterLinear
			}
			opts.GeoM.Translate(float64(x), float64(y))
			canvas.DrawImage(ctx.Images[0], &opts)
		}

		rx, ry := ctx.RightClickF32()
		x, y = CTR.Adjust(ctx.Images[1], rx, ry)
		ctx.Renderer.SetTint(0.5 + float32(ctx.DistAnim(0.5, 1.0)))
		alpha := float32(ctx.DistAnim(1.0, 0.333))
		ctx.Renderer.DrawAt(canvas, ctx.Images[1], x, y, alpha, flags.All()...)
		ctx.Renderer.SetTint(0)
	})

	rect := ebiten.NewImage(96, 64)
	app.Renderer.Gradient(rect, nil, 0, 0, color.RGBA{255, 230, 200, 255}, color.RGBA{200, 230, 255, 255}, -1, DirRadsBLTR, 1.0)
	circ := ebiten.NewImage(72, 72)
	app.Renderer.DrawCircle(circ, 72/2, 72/2, 72/2)
	app.Renderer.Options().Blend = ebiten.BlendSourceIn
	app.Renderer.Gradient(circ, nil, 0, 0, color.RGBA{0, 230, 200, 255}, color.RGBA{255, 0, 236, 255}, -1, DirRadsTTB, 1.0)
	app.Renderer.Options().Blend = ebiten.BlendSourceOver
	app.Images = append(app.Images, rect, circ)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}
