package shapes

import (
	"fmt"
	"image/color"
	"math"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

// go test -run ^TestJFMExpand$ . -count 1
func TestJFMExpand(t *testing.T) {
	const MaxDist = 16
	const MapRange = MaxDist + 2 // boundaries can be ~1.5 pixels away from seed centers
	imgIndex, regionIndex := 0, 0
	dist := float32(6)
	motion := false
	mapNormals, opNormals, animDist := true, true, true
	cache := newJFMTestCache(MapRange, MapRange)

	updater := func(ctx TestAppCtx) {
		imgIndex = updateParam(ctx, ebiten.KeySpace, imgIndex, 0, len(ctx.Images)-1, 1)
		dist = updateParam(ctx, ebiten.KeyD, dist, 0, MaxDist, 1)
		animDist = updateToggle(ctx, ebiten.KeyA, animDist)
		motion = updateToggle(ctx, ebiten.KeyM, motion)
		regionIndex = updateParam(ctx, ebiten.KeyR, regionIndex, 0, len(jfmRegions)-1, 1)
		mapNormals = updateToggle(ctx, ebiten.KeyJ, mapNormals)
		opNormals = updateToggle(ctx, ebiten.KeyN, opNormals)
	}
	drawer := func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Fill(color.Black)
		bw, bh := rectSizeF32(canvas.Bounds())
		img := ctx.Images[imgIndex]
		w, h := rectSizeF32(img.Bounds())
		r := dist
		if animDist {
			r = float32(ctx.DistAnim(MaxDist, 1.0))
		}
		mx, my := float32(0.0), float32(0.0)
		if motion {
			mx, my = float32(-4.0+ctx.DistAnim(8.0, 1.0)), float32(-4.0+ctx.DistAnim(8.0, 0.777))
		}

		ctx.DrawAtF32(canvas, img, bw/4-w/2, bh/4-h/2)
		if imgIndex == 2 {
			ctx.Renderer.SetColorF32(0.75, 0.5, 1.0, 1.0)
			ctx.Renderer.FillCircle(canvas, bw-bw/4, bh/4, w/2+16.0)
			ctx.Renderer.FillCircle(canvas, bw-bw/4, bh-bh/4, w/2+16.0)
		}

		ctx.Renderer.SetColorF32(0.5, 0.75, 1.0, 1.0)
		ctx.Renderer.MorphExpansion(canvas, img, bw-bw/4-w/2+mx, bh/4-h/2+my, r, 1.333)

		normals := cache.alphaNormals(ctx.Renderer, img, false)
		region := jfmRegions[regionIndex]
		computeOpts := JFMapComputeOptions{Normals: mapBool(mapNormals, nil, normals), MinAlpha: region.minAlpha, MaxAlpha: region.maxAlpha}
		jfmap := cache.jfmap(ctx.Renderer, img, computeOpts, false)
		jfmap.Normals = mapBool(opNormals, nil, normals)
		ox, oy := bw-bw/4-w/2-MapRange+mx, bh-bh/4-h/2-MapRange+my
		ctx.Renderer.JFMExpand(canvas, cache.source(img), ox, oy, r, jfmap)

		ctx.Renderer.SetColor(color.White)
		minAlpha, maxAlpha := jfmRegion(region.minAlpha, region.maxAlpha)
		controls := fmt.Sprintf(
			"Shape: %d/%d [Space]\nDistance: %.1f [D], animated: %t [A]\nMovement: %t [M]\n\n"+
				"(JFM only controls)\nRegion: %s, alpha in [%.3g, %.3g] [R]\n"+
				"Normals: jfmap %t [J], operation %t [N]",
			imgIndex+1, len(ctx.Images), r, animDist, motion,
			region.name, minAlpha, maxAlpha, mapNormals, opNormals,
		)
		ctx.Renderer.Text(canvas, controls, 8, 8, TextOpts(1.0, TopLeft.Snap(CapLine)))
		panels := "Top-left: source\nTop-right: MorphExpansion\nBottom-right: JFMExpand"
		ctx.Renderer.Text(canvas, panels, 8, bh-8, TextOpts(1.0, BottomLeft.Snap(Baseline)))
	}

	app := NewTestApp(updater, drawer)
	app.Images = append(app.Images, jfmShapes(app.Renderer)...)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestJFMOutline$ . -count 1
func TestJFMOutline(t *testing.T) {
	const MaxThickness = 16
	const MapRange = MaxThickness/2 + 2 // boundaries can be ~1.5 pixels away from seed centers
	imgIndex, regionIndex := 0, 0
	thickness := float32(6)
	clampBorders, motion := false, false
	mapNormals, opNormals, animThickness := true, true, true
	cache := newJFMTestCache(MapRange, MapRange)

	updater := func(ctx TestAppCtx) {
		imgIndex = updateParam(ctx, ebiten.KeySpace, imgIndex, 0, len(ctx.Images)-1, 1)
		thickness = updateParam(ctx, ebiten.KeyT, thickness, 0, MaxThickness, 1)
		animThickness = updateToggle(ctx, ebiten.KeyA, animThickness)
		clampBorders = updateToggle(ctx, ebiten.KeyC, clampBorders)
		motion = updateToggle(ctx, ebiten.KeyM, motion)
		regionIndex = updateParam(ctx, ebiten.KeyR, regionIndex, 0, len(jfmRegions)-1, 1)
		mapNormals = updateToggle(ctx, ebiten.KeyJ, mapNormals)
		opNormals = updateToggle(ctx, ebiten.KeyN, opNormals)
	}
	drawer := func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Fill(color.Black)
		bw, bh := rectSizeF32(canvas.Bounds())
		img := ctx.Images[imgIndex]
		w, h := rectSizeF32(img.Bounds())
		th := thickness
		if animThickness {
			th = float32(ctx.DistAnim(MaxThickness, 1.0))
		}
		mx, my := float32(0.0), float32(0.0)
		if motion {
			mx, my = float32(-4.0+ctx.DistAnim(8.0, 1.0)), float32(-4.0+ctx.DistAnim(8.0, 0.777))
		}

		ctx.DrawAtF32(canvas, img, bw/4-w/2, bh/4-h/2)
		if imgIndex == 2 {
			ctx.Renderer.SetColorF32(0.75, 0.5, 1.0, 1.0)
			ctx.Renderer.FillCircle(canvas, bw-bw/4, bh/4, w/2+16.0)
			ctx.Renderer.FillCircle(canvas, bw-bw/4, bh-bh/4, w/2+16.0)
		}

		var flags []Flag
		if clampBorders {
			flags = append(flags, Clamp)
		}
		ctx.Renderer.SetColorF32(0.5, 0.75, 1.0, 1.0)
		ctx.Renderer.MorphOutline(canvas, img, bw-bw/4-w/2+mx, bh/4-h/2+my, th, 0.666, flags...)

		normals := cache.alphaNormals(ctx.Renderer, img, clampBorders)
		region := jfmRegions[regionIndex]
		computeOpts := JFMapComputeOptions{Normals: mapBool(mapNormals, nil, normals), MinAlpha: region.minAlpha, MaxAlpha: region.maxAlpha}
		jfmap := cache.jfmap(ctx.Renderer, img, computeOpts, clampBorders)
		jfmap.Normals = mapBool(opNormals, nil, normals)
		ox, oy := bw-bw/4-w/2-MapRange+mx, bh-bh/4-h/2-MapRange+my
		ctx.Renderer.JFMOutline(canvas, cache.source(img), ox, oy, th, jfmap)

		ctx.Renderer.SetColor(color.White)
		minAlpha, maxAlpha := jfmRegion(region.minAlpha, region.maxAlpha)
		controls := fmt.Sprintf(
			"Shape: %d/%d [Space]\nThickness: %.1f [T], animated: %t [A]\nClamp: %t [C]\nMovement: %t [M]\n\n"+
				"(JFM only controls)\nRegion: %s, alpha in [%.3g, %.3g] [R]\n"+
				"Normals: jfmap %t [J], operation %t [N]",
			imgIndex+1, len(ctx.Images), th, animThickness, clampBorders, motion,
			region.name, minAlpha, maxAlpha, mapNormals, opNormals,
		)
		ctx.Renderer.Text(canvas, controls, 8, 8, TextOpts(1.0, TopLeft.Snap(CapLine)))
		panels := "Top-left: source\nTop-right: MorphOutline\nBottom-right: JFMOutline"
		ctx.Renderer.Text(canvas, panels, 8, bh-8, TextOpts(1.0, BottomLeft.Snap(Baseline)))
	}

	app := NewTestApp(updater, drawer)
	app.Images = append(app.Images, jfmShapes(app.Renderer)...)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestJFMErode$ . -count 1
func TestJFMErode(t *testing.T) {
	const MaxDist = 16
	const MapRange = MaxDist + 2 // boundaries can be ~1.5 pixels away from seed centers
	imgIndex, regionIndex := 0, 0
	dist := float32(6)
	softEdge := float32(0.666)
	var flags flagList
	motion, mapNormals, opNormals, animDist := false, true, true, true
	cache := newJFMTestCache(0, MapRange)

	updater := func(ctx TestAppCtx) {
		imgIndex = updateParam(ctx, ebiten.KeySpace, imgIndex, 0, len(ctx.Images)-1, 1)
		softEdge = updateParam(ctx, ebiten.KeyS, softEdge, 0.666, 16.0, 1.0)
		flags.UpdateFlag(Clamp, ebiten.KeyC)
		dist = updateParam(ctx, ebiten.KeyD, dist, 0, MaxDist, 1)
		animDist = updateToggle(ctx, ebiten.KeyA, animDist)
		motion = updateToggle(ctx, ebiten.KeyM, motion)
		regionIndex = updateParam(ctx, ebiten.KeyR, regionIndex, 0, len(jfmRegions)-1, 1)
		mapNormals = updateToggle(ctx, ebiten.KeyJ, mapNormals)
		opNormals = updateToggle(ctx, ebiten.KeyN, opNormals)
	}
	drawer := func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Fill(color.Black)
		bw, bh := rectSizeF32(canvas.Bounds())
		img := ctx.Images[imgIndex]
		w, h := rectSizeF32(img.Bounds())
		ctx.DrawAtF32(canvas, img, bw/4-w/2, bh/4-h/2)
		r := dist
		if animDist {
			r = float32(ctx.DistAnim(MaxDist, 1.0))
		}
		mx, my := float32(0.0), float32(0.0)
		if motion {
			mx, my = float32(-4.0+ctx.DistAnim(8.0, 1.0)), float32(-4.0+ctx.DistAnim(8.0, 0.777))
		}
		ctx.Renderer.SetTint(float32(ctx.DistAnim(1.0, 0.5)))

		ctx.Renderer.MorphErosion(canvas, img, bw-bw/4-w/2+mx, bh/4-h/2+my, r, softEdge, flags...)

		normals := cache.alphaNormals(ctx.Renderer, img, flags.Has(Clamp))
		region := jfmRegions[regionIndex]
		computeOpts := JFMapComputeOptions{Normals: mapBool(mapNormals, nil, normals), MinAlpha: region.minAlpha, MaxAlpha: region.maxAlpha}
		jfmap := cache.jfmap(ctx.Renderer, img, computeOpts, flags.Has(Clamp))
		jfmap.Normals = mapBool(opNormals, nil, normals)
		ctx.Renderer.JFMErode(canvas, cache.source(img), bw-bw/4-w/2+mx, bh-bh/4-h/2+my, r, jfmap)
		ctx.Renderer.SetTint(0)

		ctx.Renderer.SetColor(color.White)
		minAlpha, maxAlpha := jfmRegion(region.minAlpha, region.maxAlpha)
		controls := fmt.Sprintf(
			"Shape: %d/%d [Space]\nSoftEdge: %.02f [S] (MorphErosion only)\nClamp: %t [C]\n"+
				"Distance: %.1f [D], animated: %t [A]\nMovement: %t [M]\n\n"+
				"(JFM only controls)\nRegion: %s, alpha in [%.3g, %.3g] [R]\n"+
				"Normals: jfmap %t [J], operation %t [N]",
			imgIndex+1, len(ctx.Images), softEdge, flags.Has(Clamp), r, animDist, motion,
			region.name, minAlpha, maxAlpha, mapNormals, opNormals,
		)
		ctx.Renderer.Text(canvas, controls, 8, 8, TextOpts(1.0, TopLeft.Snap(CapLine)))
		panels := "Top-left: source\nTop-right: MorphErosion\nBottom-right: JFMErode\n" +
			"Tint animates between source and renderer colors"
		ctx.Renderer.Text(canvas, panels, 8, bh-8, TextOpts(1.0, BottomLeft.Snap(Baseline)))
	}

	app := NewTestApp(updater, drawer)
	app.Images = append(app.Images, jfmShapes(app.Renderer)...)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestJFMapViews$ . -count 1
func TestJFMapViews(t *testing.T) {
	clamp, inside, viz := false, false, false
	opacity := float32(1.0)
	updater := func(ctx TestAppCtx) {
		clamp = updateToggle(ctx, ebiten.KeyC, clamp)
		inside = updateToggle(ctx, ebiten.KeyI, inside)
		viz = updateToggle(ctx, ebiten.KeyV, viz)
		opacity = updateParam(ctx, ebiten.KeyO, opacity, 0.25, 1.0, 0.25)
	}
	drawer := func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Clear()
		px, py := ebiten.CursorPosition()
		rc := ctx.RightClickF32()
		ctx.Renderer.SetColor(color.White)
		ctx.Renderer.FillCircle(canvas, float32(px), float32(py), 128.0)
		ctx.Renderer.FillCircle(canvas, rc.X, rc.Y, 96.0)

		if !ctx.SpacePressed {
			const MaxDist = 128
			bounds := canvas.Bounds()
			// the source is read too, so it can't be drawn onto
			source := ctx.Renderer.UnsafeTempCopy(2, canvas, 0, false)
			target := ctx.Renderer.UnsafeTemp(1, bounds.Dx(), bounds.Dy(), false)
			var flags []Flag
			if clamp {
				flags = append(flags, Clamp)
			}
			jfmap := ctx.Renderer.JFMapCompute(target, source, JFMapComputeOptions{MaxDistance: MaxDist}, flags...)
			ctx.Renderer.SetColorF32(opacity, opacity, opacity, opacity) // premultiplied
			if viz {
				ctx.Renderer.JFMapViz(canvas, source, 0, 0, MaxDist, jfmap)
			} else {
				ctx.Renderer.JFMapHeat(canvas, source, 0, 0, MaxDist, inside, jfmap)
			}
		}

		_, bh := rectSizeF32(canvas.Bounds())
		ctx.Renderer.SetColor(color.White)
		controls := fmt.Sprintf(
			"JFMapViz: %t [V] (JFMapHeat otherwise)\nInside: %t [I] (JFMapHeat only)\n"+
				"Opacity: %.2f [O]\nClamp: %t [C]\nShow source: %t [Hold Space]",
			viz, inside, opacity, clamp, ctx.SpacePressed,
		)
		ctx.Renderer.Text(canvas, controls, 8, 8, TextOpts(1.0, TopLeft.Snap(CapLine)))
		panels := "Distances to two circles: one follows the cursor, the other is\n" +
			"placed with right click"
		ctx.Renderer.Text(canvas, panels, 8, bh-8, TextOpts(1.0, BottomLeft.Snap(Baseline)))
	}

	app := NewTestApp(updater, drawer)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

func jfmShapes(r *Renderer) []*ebiten.Image {
	const BaseRadius = 128
	const XW, XH = BaseRadius * 2, (BaseRadius * 7) / 4
	const XMargin, XThick = BaseRadius / 3, BaseRadius / 8
	const Circ2Radius = 72

	circle := ebiten.NewImage(BaseRadius*2, BaseRadius*2)
	gradientOpts := GradientOpts(color.RGBA{0, 196, 255, 255}, color.RGBA{0, 0, 0, 0}, false)
	gradientOpts.Bias = 0.5
	r.GradientRadial(circle, gradientOpts, BaseRadius, BaseRadius, BaseRadius*0.25, BaseRadius*0.75, BaseRadius)
	r.opts.Blend = ebiten.BlendSourceAtop
	gradientOpts = GradientOpts(color.RGBA{196, 64, 0, 196}, color.RGBA{64, 16, 0, 64}, false)
	gradientOpts.Bias = -0.25
	r.Gradient(circle, gradientOpts, DirRadsBLTR)
	r.opts.Blend = ebiten.BlendSourceOver
	r.SetColorF32(0.75, 0.5, 1.0, 1.0)
	xSign := ebiten.NewImage(XW, XH)
	r.StrokeLine(xSign, PtF32(XMargin, XMargin), PtF32(XW-XMargin, XH-XMargin), XThick)
	r.StrokeLine(xSign, PtF32(XW-XMargin, XMargin), PtF32(XMargin, XH-XMargin), XThick)
	circ2 := ebiten.NewImage(Circ2Radius*2, Circ2Radius*2)
	r.SetColorF32(0.5, 0.25, 0.75, 1.0)
	r.FillCircle(circ2, Circ2Radius, Circ2Radius, Circ2Radius)
	r.SetColorF32(0.7, 0.7, 0.7, 1.0)
	// rhombus clipped by the 4 image borders, leaving the corners empty
	rhombus := ebiten.NewImage(256, 192)
	r.FillQuad(rhombus, [4]PointF32{{128, -24}, {288, 96}, {128, 216}, {-32, 96}}, 0)
	// vertical stripes clipped by the top and bottom borders only. they go past
	// the borders, so the border rows aren't antialiased
	stripes := ebiten.NewImage(256, 192)
	for x := float32(16); x < 256-16; x += 48 {
		r.FillRect(stripes, x, -8, 24, 192+16, 0)
	}
	// magic circle: rings of different thicknesses plus a hexagram, with some
	// strokes thinner than 2 pixels
	magic := ebiten.NewImage(256, 256)
	r.SetColorF32(1.0, 0.85, 0.4, 1.0)
	for _, ring := range [][2]float32{{120, 6}, {109, 1.5}, {101, 3}, {40, 2}, {24, 8}, {8, 1}} {
		r.StrokeCircle(magic, 128, 128, ring[0], ring[1])
	}
	for i, thickness := range []float32{3, 1.5} {
		var tri [3]PointF32
		for j := range tri {
			sin, cos := math.Sincos(math.Pi * (float64(i) - 0.5 + float64(j)*2.0/3.0))
			tri[j] = PtF32(128+101*cos, 128+101*sin)
		}
		r.StrokeTriangle(magic, tri, thickness, 0)
	}
	return []*ebiten.Image{circle, xSign, circ2, rhombus, stripes, magic}
}

// jfmTestCache keeps what TestJFMExpand, TestJFMOutline and TestJFMErode
// would otherwise compute on every frame: the sources with transparent padding
// around them, for the operations to draw on, their alpha normals and their
// jfmaps.
type jfmTestCache struct {
	padding, mapRange int
	sources           map[*ebiten.Image]*ebiten.Image
	normals           map[jfmTestNormalsKey]*ebiten.Image
	jfmaps            map[jfmTestMapKey]JFMap
}

type jfmTestNormalsKey struct {
	img   *ebiten.Image
	clamp bool
}

type jfmTestMapKey struct {
	img   *ebiten.Image
	opts  JFMapComputeOptions
	clamp bool
}

func newJFMTestCache(padding, mapRange int) *jfmTestCache {
	return &jfmTestCache{
		padding:  padding,
		mapRange: mapRange,
		sources:  make(map[*ebiten.Image]*ebiten.Image),
		normals:  make(map[jfmTestNormalsKey]*ebiten.Image),
		jfmaps:   make(map[jfmTestMapKey]JFMap),
	}
}

// source returns img with the padding.
func (c *jfmTestCache) source(img *ebiten.Image) *ebiten.Image {
	source, found := c.sources[img]
	if !found {
		source = padded(img, c.padding)
		c.sources[img] = source
	}
	return source
}

// alphaNormals returns the alpha normals of img, with the padding. With clamp,
// they're computed before adding the padding, so that img continues past its
// own borders, not past the padding.
func (c *jfmTestCache) alphaNormals(r *Renderer, img *ebiten.Image, clamp bool) *ebiten.Image {
	key := jfmTestNormalsKey{img, clamp}
	normals, found := c.normals[key]
	if !found {
		normals = padded(alphaNormals(r, img, mapBool(clamp, noFlag, Clamp)), c.padding)
		c.normals[key] = normals
	}
	return normals
}

// jfmap returns a jfmap of img with the padding, computed with opts and the
// cache's map range. With clamp, the padding is left out of the map, so that
// img continues past its own borders.
func (c *jfmTestCache) jfmap(r *Renderer, img *ebiten.Image, opts JFMapComputeOptions, clamp bool) JFMap {
	key := jfmTestMapKey{img, opts, clamp}
	jfmap, found := c.jfmaps[key]
	if !found {
		source := c.source(img)
		target := ebiten.NewImage(source.Bounds().Dx(), source.Bounds().Dy())
		opts.MaxDistance = c.mapRange
		opts.Inset = mapBool(clamp, 0, c.padding)
		jfmap = r.JFMapCompute(target, source, opts, mapBool(clamp, noFlag, Clamp))
		c.jfmaps[key] = jfmap
	}
	return jfmap
}
