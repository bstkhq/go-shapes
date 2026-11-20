package shapes

import (
	"fmt"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

// go test -run ^TestScale$ . -count 1
func TestScale(t *testing.T) {
	updater := func(TestAppCtx) {}
	drawer := func(canvas *ebiten.Image, ctx TestAppCtx) {
		ox, oy, w, h := rectOriginSize(canvas.Bounds())
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

		ctx.Renderer.SetColorF32A(tcWhite)
		info := "Top: Ebitengine nearest / linear\nBottom: go-shapes bilinear / bicubic\nHold [C] to clamp, hold [D] for DstSampling"
		ctx.Renderer.Text(canvas, info, float32(ox+8), float32(oy+8), TextOpts(1, TopLeft.Snap(CapLine)))
	}

	app := NewTestApp(updater, drawer)
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

// go test -run ^TestScaleFiltering$ . -count 1
func TestScaleFiltering(t *testing.T) {
	const W, H = 192, 128
	const MinScalePercent, MaxScalePercent = 10, 500
	var opts ScaleOptions
	scalePercent := 100
	texIndex := 0
	texNames := []string{"Checkerboard", "Vertical stripes", "Text", "Opaque edges"}
	smoothMovement, smoothScaling := false, false
	updater := func(ctx TestAppCtx) {
		opts.Bicubic = updateToggle(ctx, ebiten.KeyB, opts.Bicubic)
		opts.DstSampling = updateToggle(ctx, ebiten.KeyD, opts.DstSampling)
		texIndex = updateParam(ctx, ebiten.KeyT, texIndex, 0, len(ctx.Images)-1, 1)
		scalePercent = updateParam(ctx, ebiten.KeyZ, scalePercent, MinScalePercent, MaxScalePercent, 5)
		smoothMovement = updateToggle(ctx, ebiten.KeyM, smoothMovement)
		smoothScaling = updateToggle(ctx, ebiten.KeyA, smoothScaling)
	}
	drawer := func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Fill(backTestColor)
		ox, oy, w, h := rectOriginSizeF32(canvas.Bounds())
		src := ctx.Images[texIndex]
		sw, sh := rectSizeF32(src.Bounds())
		currentScalePercent := float32(scalePercent)
		if smoothScaling {
			lo := float32(max(MinScalePercent, scalePercent-30))
			hi := float32(min(MaxScalePercent, scalePercent+30))
			currentScalePercent = lerp(lo, hi, float32(ctx.DistAnim(1, 0.5)))
		}
		scale := currentScalePercent / 100
		x := ox + (w-sw*scale)/2
		y := oy + (h-sh*scale)/2 + 24
		if smoothMovement {
			x += float32(ctx.DistAnim(4, 0.7)) - 2
			y += float32(ctx.DistAnim(4, 1.0)) - 2
		}

		opts.Clamp = ebiten.IsKeyPressed(ebiten.KeyC)
		blend := ebiten.BlendSourceOver
		if ctx.SpacePressed {
			blend = ebiten.BlendCopy
		}
		filter := mapBool(opts.Bicubic, "go-shapes bilinear", "go-shapes bicubic")
		if ebiten.IsKeyPressed(ebiten.KeyE) {
			drawOpts := ebiten.DrawImageOptions{Filter: ebiten.FilterLinear, Blend: blend}
			drawOpts.GeoM.Scale(float64(scale), float64(scale))
			drawOpts.GeoM.Translate(float64(x), float64(y))
			canvas.DrawImage(src, &drawOpts) // notice: positions are quantized by adjustDestinationPixel
			filter = "Ebitengine bilinear"
		} else {
			ctx.Renderer.Options().Blend = blend
			ctx.Renderer.Scale(canvas, src, x, y, scale, &opts)
			ctx.Renderer.Options().Blend = ebiten.BlendSourceOver
		}

		info := fmt.Sprintf("Image: %s [T]\n"+
			"Scale: %.1f%% [Z / Shift+Z]  Smooth scaling: %t [A]\n"+
			"Filter: %s [B: bilinear/bicubic, hold E: Ebitengine]\n"+
			"Clamp: %t [hold C]  DstSampling: %t [D]  Smooth movement: %t [M]",
			texNames[texIndex], currentScalePercent, smoothScaling, filter, opts.Clamp, opts.DstSampling, smoothMovement)
		ctx.Renderer.SetColorF32A(tcWhite)
		ctx.Renderer.Text(canvas, info, ox+8, oy+8, TextOpts(1, TopLeft.Snap(CapLine)))
	}

	app := NewTestApp(updater, drawer)
	const TextW, TextH = 160, 112
	text := ebiten.NewImage(TextW, TextH)
	app.Renderer.SetColorF32A(tcWhite)
	app.Renderer.Text(text, "Small text 0123456789\nSphinx of black quartz", TextW/2, TextH/2, TextOpts(1, Center))
	edges := ebiten.NewImage(W, H)
	Paint(edges, edges.Bounds(), tcWhite, ebiten.BlendCopy)
	Paint(edges, edges.Bounds().Inset(1), tcAzure, ebiten.BlendCopy)
	app.Images = []*ebiten.Image{
		newBWCheckerboard(W, H, 0, 1),
		newBWVertStripes(W, H, 0, 2),
		text,
		edges,
	}

	ebiten.SetWindowSize(1040, 720)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}
