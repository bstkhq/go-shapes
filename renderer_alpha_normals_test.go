package shapes

import (
	"fmt"
	"image/color"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

// go test -run ^TestAlphaNormals$ . -count 1
func TestAlphaNormals(t *testing.T) {
	const Margin = 16
	imgIndex := 0
	var normals, viz []*ebiten.Image

	updater := func(ctx TestAppCtx) {
		imgIndex = updateParam(ctx, ebiten.KeySpace, imgIndex, 0, len(ctx.Images)-1, 1)
	}
	drawer := func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Fill(color.Black)
		if normals == nil {
			for _, img := range ctx.Images {
				w, h := img.Bounds().Dx(), img.Bounds().Dy()
				normals = append(normals, ebiten.NewImage(w, h))
				viz = append(viz, ebiten.NewImage(w, h))
				ctx.Renderer.AlphaNormalsCompute(normals[len(normals)-1], img)
				ctx.Renderer.AlphaNormalsViz(viz[len(viz)-1], normals[len(normals)-1], 0, 0)
			}
		}

		source := ctx.Images[imgIndex]
		w, h := rectSizeF32(source.Bounds())
		cw, ch := rectSizeF32(canvas.Bounds())
		y := (ch - h) / 2
		ctx.DrawAtF32(canvas, source, cw/2-w-Margin, y)
		ctx.DrawAtF32(canvas, viz[imgIndex], cw/2+Margin, y)

		ctx.Renderer.SetColor(color.RGBA{255, 255, 255, 255})
		info := fmt.Sprintf(
			"Left: source, right: alpha normals (hue = normal direction, brightness = bright inside, dim outside)\n"+
				"Shape: %d/%d [Space]",
			imgIndex+1, len(ctx.Images),
		)
		ctx.Renderer.Text(canvas, info, 8, 8, TextOpts(1.0, TopLeft.Snap(CapLine)))
	}

	app := NewTestApp(updater, drawer)
	app.Images = append(app.Images, jfmShapes(app.Renderer)...)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}
