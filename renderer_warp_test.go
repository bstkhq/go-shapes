package shapes

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"strings"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

// go test -run ^TestWarpBarrel$ . -count 1
func TestWarpBarrel(t *testing.T) {
	const W, H = 480, 360
	warp := float32(0.5)
	pincushion := false
	texIndex := 0
	var texNames []string
	offscreen := ebiten.NewImage(W, H)
	updater := func(ctx TestAppCtx) {
		texIndex = updateParam(ctx, ebiten.KeyT, texIndex, 0, len(ctx.Images)-1, 1)
		pincushion = updateToggle(ctx, ebiten.KeyB, pincushion)
		warp = updateParam(ctx, ebiten.KeyW, warp, 0.1, 1.5, 0.1)
	}
	drawer := func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Fill(backTestColor)
		ox, oy, cw, _ := rectOriginSizeF32(canvas.Bounds())

		// tile the texture over the "offscreen"
		offscreen.Clear()
		tile := ctx.Images[texIndex]
		tw, th := tile.Bounds().Dx(), tile.Bounds().Dy()
		for y := 0; y < H; y += th {
			for x := 0; x < W; x += tw {
				var opts ebiten.DrawImageOptions
				opts.GeoM.Translate(float64(x), float64(y))
				offscreen.DrawImage(tile, &opts)
			}
		}

		horzWarp, vertWarp := warp/2, warp
		if pincushion {
			horzWarp, vertWarp = -horzWarp, -vertWarp
		}
		ctx.Renderer.WarpBarrel(canvas, offscreen, ox+(cw-W)/2, oy+64, horzWarp, vertWarp)

		info := fmt.Sprintf("Texture: %s [T]\n%s: %.1f [B, W / Shift+W]",
			texNames[texIndex], mapBool(pincushion, "Barrel", "Pincushion"), warp)
		ctx.Renderer.SetColorF32A(tcWhite)
		ctx.Renderer.Text(canvas, info, ox+8, oy+8, TextOpts(1, TopLeft.Snap(CapLine)))
	}

	app := NewTestApp(updater, drawer)
	app.Images, texNames = warpBarrelTestTiles(app.Renderer, W, H)
	ebiten.SetWindowSize(960, 520)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// warpBarrelTestTiles returns the tiles that TestWarpBarrel repeats over its
// "screen", which has the given size.
func warpBarrelTestTiles(r *Renderer, w, h int) ([]*ebiten.Image, []string) {
	gradient := ebiten.NewImage(w, h)
	tcsBase4.Apply(r)
	r.FillIntRect(gradient, image.Rect(0, 0, w, h), 0)

	scanlines := newBWHorzStripes(64, 64, 0, 1)
	checker := newBWCheckerboard(64, 64, 0, 2)
	arcTextures, _ := warpArcTestTextures(r)
	dithered := arcTextures[0]
	stripes := arcTextures[1].SubImage(image.Rect(2, 2, 510, 62)).(*ebiten.Image) // skip the transparent border

	images := []*ebiten.Image{gradient, scanlines, dithered, stripes, checker}
	names := []string{"Gradient", "Scanlines 1px", "Dithered gradient", "Stripes and text", "Checker 2px"}
	return images, names
}

// go test -run ^TestWarpArc$ . -count 1
func TestWarpArc(t *testing.T) {
	const MinRadius, MaxRadius = 40, 330
	outRadius := float32(96)
	rads := -math.Pi / 2
	taps := 1
	texIndex := 0
	var texNames []string
	rotating, curving := false, false
	var curvePhase float64
	updater := func(ctx TestAppCtx) {
		taps = updateParam(ctx, ebiten.KeyA, taps, 1, 4, 1)
		texIndex = updateParam(ctx, ebiten.KeyT, texIndex, 0, len(ctx.Images)-1, 1)
		rotating = updateToggle(ctx, ebiten.KeyM, rotating)
		curving = updateToggle(ctx, ebiten.KeyC, curving)
		outRadius = updateParamMult(ctx, ebiten.KeyR, outRadius, MinRadius, MaxRadius, 1.125)
		if rotating {
			rads = normURads(rads + 0.004)
		}
		if curving {
			curvePhase += 0.006
			outRadius = MinRadius + (MaxRadius-MinRadius)*float32(0.5-0.5*math.Cos(curvePhase))
		}
	}
	drawer := func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Fill(backTestColor)
		ox, oy, cw, ch := rectOriginSizeF32(canvas.Bounds())
		src := ctx.Images[texIndex]
		sw, sh := rectSizeF32(src.Bounds())

		var opts ebiten.DrawImageOptions
		scale := min(1, (cw-16)/sw)
		opts.GeoM.Scale(float64(scale), float64(scale))
		opts.GeoM.Translate(float64(ox+(cw-sw*scale)/2), float64(oy+40))
		canvas.DrawImage(src, &opts)
		ctx.Renderer.WarpArc(canvas, src, ox+cw/2, oy+ch/2+40, outRadius, rads, taps)

		// source pixels per destination pixel along the arc, at the outer and inner edges
		span := min(sw/outRadius, 2*math.Pi)
		kRange := fmt.Sprintf("%.2f - inf", sw/(span*outRadius))
		if inRadius := outRadius - sh; inRadius >= 1 {
			kRange = fmt.Sprintf("%.2f - %.2f", sw/(span*outRadius), sw/(span*inRadius))
		}
		info := fmt.Sprintf("Taps: %d (%d samples) [A]  Texture: %s [T]  Src px per dst px: %s\n"+
			"Radius: %.0f [R / Shift+R]  Animate radius: %t [C]  Rotate: %t [M]",
			taps, 4*taps, texNames[texIndex], kRange, outRadius, curving, rotating)
		ctx.Renderer.SetColorF32A(tcWhite)
		ctx.Renderer.Text(canvas, info, ox+8, oy+8, TextOpts(1, TopLeft.Snap(CapLine)))
	}

	app := NewTestApp(updater, drawer)
	app.Images, texNames = warpArcTestTextures(app.Renderer)
	ebiten.SetWindowSize(960, 800)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// warpArcTestTextures returns the source strips for TestWarpArc.
func warpArcTestTextures(r *Renderer) ([]*ebiten.Image, []string) {
	dithered := ebiten.NewImage(512, 64)
	mask := ebiten.NewImage(512, 64)
	gradientOpts := GradientOpts(color.RGBA{0, 0, 0, 255}, color.RGBA{255, 255, 255, 255}, false)
	r.Gradient(mask, gradientOpts, DirRadsRTL)
	r.DitherMat4(dithered, mask, 0, 0, 0, 0, PaletteBW, DitherGlitch, 0.0, 0.0)
	r.SetColorF32(0, 0.5, 0, 0.5)
	r.FillIntRect(dithered, dithered.Bounds(), 0)
	r.SetColorF32(0.5, 0.0, 0.5, 0.5)
	r.FillIntRect(dithered, image.Rect(0, 0, 512/8, 64/16), 0)

	stripes := newBWVertStripes(512, 64, 2, 2)
	Paint(stripes, image.Rect(2, 23, 510, 41), tcBlack, ebiten.BlendCopy)
	r.SetColorF32A(tcWhite)
	r.Text(stripes, "WarpArc  /  thin stripes and small text  /  WarpArc", 256, 27, TextOpts(1, TopCenter.Snap(CapLine)))

	checker := newBWCheckerboard(512, 64, 2, 4)

	// wider than the circumference for most radii, so it wraps the whole circle
	long := newBWImage(1600, 40, 2, func(x, y int) bool { return y < 16 && x%4 < 2 })
	text := strings.Repeat("long strip, squeezed around the whole circle  /  ", 8)
	r.SetColorF32A(tcWhite)
	r.Text(long, text, 800, 22, TextOpts(1, TopCenter.Snap(CapLine)))

	images := []*ebiten.Image{dithered, stripes, checker, long}
	names := []string{"Dithered gradient", "Stripes and text", "Checker 4px", "Long strip"}
	return images, names
}
