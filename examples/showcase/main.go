package main

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"math"

	shapes "github.com/erparts/go-shapes"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// The showcase is fully resizable: the layout is derived from the live logical
// screen size (see LayoutF, which folds in the monitor's device scale factor)
// and every panel renders its demo at native resolution, so nothing is drawn at
// a fixed pixel size.
const (
	// Initial window size only; the real layout follows the current screen size.
	initialWindowWidth  = 1480
	initialWindowHeight = 980

	// Reference coordinate box each demo is authored in. It is mapped uniformly
	// into whatever size the panel currently has, so demos stay crisp and keep
	// their proportions at any window size or device scale factor.
	refDemoWidth  = 320
	refDemoHeight = 210
)

// Offscreen indices reserved for the showcase. The renderer only uses indices 0
// and 1 for its internal composition, so anything from 8 upwards is safe to hold
// across renderer calls (see Renderer.UnsafeTemp docs).
const (
	tempPanel = 8
	tempAuxA  = 10
	tempAuxB  = 11
)

type panel struct {
	title string
	// draw renders the demo and returns optional notes (e.g. animated values or
	// hints) rendered at the bottom of the panel. Returning []string keeps adding
	// extra lines trivial.
	draw func(demoCtx) []string
}

type page struct {
	title  string
	panels []panel
}

type Showcase struct {
	renderer *shapes.Renderer
	tick     uint64
	page     int
	pages    []page

	screenW, screenH int

	spriteA *ebiten.Image
	spriteB *ebiten.Image
	maskA   *ebiten.Image
	jfMask  *ebiten.Image
}

func main() {
	ebiten.SetWindowSize(initialWindowWidth, initialWindowHeight)
	ebiten.SetWindowTitle("go-shapes showcase")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	app := NewShowcase()
	if err := ebiten.RunGame(app); err != nil {
		log.Fatal(err)
	}
}

func NewShowcase() *Showcase {
	app := &Showcase{renderer: shapes.NewRenderer()}
	app.makeAssets()
	app.pages = buildPages()
	return app
}

// LayoutF makes the logical resolution track the window size and the monitor's
// device scale factor, so the showcase stays sharp on high-DPI displays and
// fills the window at any size. Implementing LayoutFer means Layout is unused.
func (a *Showcase) LayoutF(outsideWidth, outsideHeight float64) (float64, float64) {
	scale := ebiten.Monitor().DeviceScaleFactor()
	return outsideWidth * scale, outsideHeight * scale
}

func (a *Showcase) Layout(outsideWidth, outsideHeight int) (int, int) {
	return outsideWidth, outsideHeight
}

func (a *Showcase) Update() error {
	a.tick++
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		return ebiten.Termination
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyRight) || inpututil.IsKeyJustPressed(ebiten.KeyPageDown) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		a.page = (a.page + 1) % len(a.pages)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyLeft) || inpututil.IsKeyJustPressed(ebiten.KeyPageUp) {
		a.page = (a.page + len(a.pages) - 1) % len(a.pages)
	}
	return nil
}

func (a *Showcase) Draw(screen *ebiten.Image) {
	bounds := screen.Bounds()
	a.screenW, a.screenH = bounds.Dx(), bounds.Dy()

	a.paintBackground(screen)
	a.drawHeader(screen)

	pg := a.pages[a.page]
	cols, rows := gridDims(len(pg.panels))
	for i, p := range pg.panels {
		a.drawPanel(screen, cols, rows, i, p)
	}
}

func (a *Showcase) gap() int    { return max(8, a.screenW/92) }
func (a *Showcase) header() int { return max(44, a.screenH/18) }

func (a *Showcase) paintBackground(screen *ebiten.Image) {
	screen.Fill(color.RGBA{12, 14, 20, 255})
	spacing := float32(max(18, a.screenW/52))
	a.renderer.SetColor(color.RGBA{20, 24, 31, 255})
	a.renderer.TileDotsGrid(screen, spacing*0.05, spacing, 0, 0)
}

func (a *Showcase) drawHeader(screen *ebiten.Image) {
	title := fmt.Sprintf("go-shapes showcase  •  page %d/%d  •  %s", a.page+1, len(a.pages), a.pages[a.page].title)
	help := "Left/Right or PgUp/PgDn to switch pages • Esc to quit • resize the window freely"
	ebitenutil.DebugPrintAt(screen, title, a.gap(), 12)
	ebitenutil.DebugPrintAt(screen, help, a.gap(), 30)
}

// gridDims picks a roughly square grid that fits n panels.
func gridDims(n int) (cols, rows int) {
	if n <= 0 {
		return 1, 1
	}
	cols = int(math.Ceil(math.Sqrt(float64(n))))
	rows = (n + cols - 1) / cols
	return cols, rows
}

func (a *Showcase) panelRect(cols, rows, idx int) image.Rectangle {
	gap := a.gap()
	top := a.header()
	w := (a.screenW - gap*(cols+1)) / cols
	h := (a.screenH - top - gap*(rows+1)) / rows
	col, row := idx%cols, idx/cols
	x := gap + col*(w+gap)
	y := top + gap + row*(h+gap)
	return image.Rect(x, y, x+w, y+h)
}

func (a *Showcase) drawPanel(screen *ebiten.Image, cols, rows, idx int, p panel) {
	bounds := a.panelRect(cols, rows, idx)
	w, h := bounds.Dx(), bounds.Dy()

	panel := a.renderer.UnsafeTemp(tempPanel, w, h, false)
	panel.Fill(color.RGBA{18, 22, 31, 245})
	a.renderer.SetColor(color.RGBA{55, 68, 84, 255})
	a.renderer.StrokeArea(panel, 0, 0, float32(w), float32(h), 1, 0, float32(min(w, h))/22)

	ebitenutil.DebugPrintAt(panel, p.title, 12, 8)

	// Map the reference demo box uniformly into the panel's content area.
	const top, bottom, padX = 28.0, 34.0, 12.0
	rw := float64(w) - 2*padX
	rh := float64(h) - top - bottom
	k := math.Min(rw/refDemoWidth, rh/refDemoHeight)
	bw, bh := refDemoWidth*k, refDemoHeight*k
	ctx := demoCtx{
		a:   a,
		r:   a.renderer,
		dst: panel,
		bx:  float32(padX + (rw-bw)/2),
		by:  float32(top + (rh-bh)/2),
		k:   float32(k),
	}

	a.renderer.SetColor(color.White)
	a.renderer.SetBlend(ebiten.BlendSourceOver)
	notes := p.draw(ctx)

	for i, note := range notes {
		ebitenutil.DebugPrintAt(panel, note, 12, h-30+i*14)
	}

	var opts ebiten.DrawImageOptions
	opts.GeoM.Translate(float64(bounds.Min.X), float64(bounds.Min.Y))
	screen.DrawImage(panel, &opts)
}

// demoCtx maps the reference demo box into panel coordinates and hands out
// temporary offscreens from the renderer instead of allocating per frame.
type demoCtx struct {
	a         *Showcase
	r         *shapes.Renderer
	dst       *ebiten.Image
	bx, by, k float32
}

func (d demoCtx) x(v float32) float32 { return d.bx + v*d.k }
func (d demoCtx) y(v float32) float32 { return d.by + v*d.k }
func (d demoCtx) s(v float32) float32 { return v * d.k }

func (d demoCtx) xf(v float32) float64 { return float64(d.x(v)) }
func (d demoCtx) yf(v float32) float64 { return float64(d.y(v)) }
func (d demoCtx) sf(v float32) float64 { return float64(d.s(v)) }

func (d demoCtx) xi(v float32) int { return int(d.x(v)) }
func (d demoCtx) yi(v float32) int { return int(d.y(v)) }

func (d demoCtx) pt(x, y float32) shapes.PointF32 {
	return shapes.PointF32{X: d.x(x), Y: d.y(y)}
}

// tempRef returns a temporary offscreen sized from reference dimensions scaled
// to the current panel, so intermediate images scale with the window too.
func (d demoCtx) tempRef(idx int, refW, refH float32, clear bool) *ebiten.Image {
	return d.r.UnsafeTemp(idx, int(refW*d.k), int(refH*d.k), clear)
}

// scaled returns a scaled copy of a static asset, so asset-based demos also
// grow and shrink with the window (dogfooding Renderer.Scale in the process).
func (d demoCtx) scaled(idx int, src *ebiten.Image) *ebiten.Image {
	b := src.Bounds()
	w := int(math.Ceil(float64(b.Dx()) * float64(d.k)))
	h := int(math.Ceil(float64(b.Dy()) * float64(d.k)))
	t := d.r.UnsafeTemp(idx, w, h, true)
	d.r.Scale(t, src, 0, 0, d.k, false)
	return t
}

// blit draws src into the demo at reference coordinates (x, y).
func (d demoCtx) blit(src *ebiten.Image, x, y float32) {
	var opts ebiten.DrawImageOptions
	opts.GeoM.Translate(d.xf(x), d.yf(y))
	d.dst.DrawImage(src, &opts)
}

func (a *Showcase) makeAssets() {
	// Assets are built once at startup, so plain NewImage is appropriate here.
	a.spriteA = ebiten.NewImage(220, 220)
	a.maskA = ebiten.NewImage(220, 220)
	a.jfMask = ebiten.NewImage(220, 220)

	fillChecker(a.spriteA)
	a.renderer.SetColor(color.RGBA{255, 210, 80, 255})
	a.renderer.DrawHexagon(a.spriteA, 110, 110, 70, 18, float32(shapes.RadsBottomRight))
	a.renderer.SetColor(color.RGBA{70, 200, 255, 230})
	a.renderer.DrawRingSector(a.spriteA, 110, 110, 48, 86, shapes.RadsTopLeft, shapes.RadsBottomRight+0.8, 8)
	a.renderer.SetColor(color.RGBA{255, 120, 120, 255})
	a.renderer.StrokeCircle(a.spriteA, 110, 110, 90, 4)

	a.renderer.SetColor(color.White)
	a.renderer.DrawCircle(a.maskA, 110, 110, 72)
	a.renderer.SetColor(color.RGBA{255, 255, 255, 200})
	a.renderer.DrawPie(a.maskA, 110, 110, 96, shapes.RadsLeft+0.3, shapes.RadsTopRight, 8)

	a.renderer.SetColor(color.White)
	a.renderer.DrawTriangle(a.jfMask, 32, 190, 110, 28, 188, 190, 12)
	a.renderer.DrawCircle(a.jfMask, 110, 110, 34)

	a.spriteB = ebiten.NewImage(220, 220)
	a.renderer.GradientRadial(a.spriteB, 110, 110, color.RGBA{76, 164, 255, 255}, color.RGBA{10, 10, 16, 0}, 16, 72, 106, -1, 2.2)
	a.renderer.SetBlend(ebiten.BlendSourceAtop)
	a.renderer.Gradient(a.spriteB, nil, 0, 0, color.RGBA{255, 108, 64, 255}, color.RGBA{123, 26, 184, 255}, 8, shapes.DirRadsTLBR, 1.0)
	a.renderer.SetBlend(ebiten.BlendSourceOver)
	a.renderer.SetColor(color.RGBA{255, 255, 255, 220})
	a.renderer.StrokeTriangle(a.spriteB, 48, 172, 110, 46, 172, 172, 7, 10)
}

func fillChecker(img *ebiten.Image) {
	b := img.Bounds()
	tile := 16
	for y := b.Min.Y; y < b.Max.Y; y += tile {
		for x := b.Min.X; x < b.Max.X; x += tile {
			clr := color.RGBA{36, 42, 52, 255}
			if ((x/tile)+(y/tile))%2 == 0 {
				clr = color.RGBA{28, 33, 42, 255}
			}
			r := image.Rect(x, y, min(x+tile, b.Max.X), min(y+tile, b.Max.Y))
			sub := img.SubImage(r).(*ebiten.Image)
			sub.Fill(clr)
		}
	}
}

func buildPages() []page {
	return []page{
		{title: "Primitives", panels: primitivesPanels()},
		{title: "Rounded rects & sectors", panels: rectsPanels()},
		{title: "Color & gradients", panels: colorPanels()},
		{title: "Tiles & noise", panels: tilePanels()},
		{title: "Masks", panels: maskPanels()},
		{title: "Filters & light", panels: effectPanels()},
		{title: "Mapping & JFM", panels: mappingPanels()},
	}
}

func primitivesPanels() []panel {
	return []panel{
		{"DrawArea / StrokeArea", func(d demoCtx) []string {
			d.r.SetColor(color.RGBA{68, 180, 255, 255})
			d.r.DrawArea(d.dst, d.x(28), d.y(44), d.s(120), d.s(86), d.s(18))
			d.r.SetColor(color.RGBA{255, 204, 84, 255})
			d.r.StrokeArea(d.dst, d.x(170), d.y(38), d.s(120), d.s(98), d.s(5), d.s(8), d.s(24))
			return nil
		}},
		{"DrawLine / DrawCircle", func(d demoCtx) []string {
			d.r.SetColor(color.RGBA{255, 120, 120, 255})
			d.r.DrawLine(d.dst, d.xf(20), d.yf(150), d.xf(140), d.yf(40), d.sf(10))
			d.r.SetColor(color.RGBA{90, 225, 170, 255})
			d.r.DrawCircle(d.dst, d.x(230), d.y(96), d.s(46))
			return nil
		}},
		{"StrokeCircle / DrawRing", func(d demoCtx) []string {
			d.r.SetColor(color.RGBA{114, 217, 255, 255})
			d.r.StrokeCircle(d.dst, d.x(96), d.y(100), d.s(52), d.s(8))
			d.r.SetColor(color.RGBA{255, 193, 84, 255})
			d.r.DrawRing(d.dst, d.x(228), d.y(100), d.s(26), d.s(52))
			return nil
		}},
		{"DrawPie / StrokePie", func(d demoCtx) []string {
			t := 0.7 + 0.6*math.Sin(float64(d.a.tick)*0.03)
			d.r.SetColor(color.RGBA{255, 99, 132, 255})
			d.r.DrawPie(d.dst, d.x(90), d.y(100), d.s(52), shapes.RadsTop, shapes.RadsTop+t*math.Pi, d.s(12))
			d.r.SetColor(color.RGBA{115, 215, 255, 255})
			d.r.StrokePie(d.dst, d.x(220), d.y(100), d.s(54), d.s(12), shapes.RadsLeft, shapes.RadsLeft+t*math.Pi, d.s(8))
			return nil
		}},
		{"DrawEllipse", func(d demoCtx) []string {
			d.r.SetColor(color.RGBA{170, 120, 255, 255})
			d.r.DrawEllipse(d.dst, d.x(160), d.y(104), d.s(108), d.s(50), math.Sin(float64(d.a.tick)*0.02)*0.7)
			return nil
		}},
		{"DrawTriangle / StrokeTriangle", func(d demoCtx) []string {
			d.r.SetColor(color.RGBA{114, 217, 255, 255})
			d.r.DrawTriangle(d.dst, d.xf(24), d.yf(170), d.xf(84), d.yf(56), d.xf(150), d.yf(180), d.sf(10))
			d.r.SetColor(color.RGBA{255, 210, 96, 255})
			d.r.StrokeTriangle(d.dst, d.xf(172), d.yf(170), d.xf(232), d.yf(60), d.xf(296), d.yf(170), d.sf(8), d.sf(10))
			return nil
		}},
		{"DrawHexagon", func(d demoCtx) []string {
			d.r.SetColor(color.RGBA{83, 220, 171, 255})
			d.r.DrawHexagon(d.dst, d.x(160), d.y(104), d.s(74), d.s(16), float32(d.a.tick)*0.01)
			return nil
		}},
		{"DrawQuad / DrawQuadSoft", func(d demoCtx) []string {
			quadA := [4]shapes.PointF32{d.pt(24, 78), d.pt(148, 56), d.pt(170, 172), d.pt(12, 188)}
			quadB := [4]shapes.PointF32{d.pt(168, 84), d.pt(300, 98), d.pt(288, 190), d.pt(164, 178)}
			d.r.SetColor(color.RGBA{255, 104, 104, 255})
			d.r.DrawQuad(d.dst, quadA, 0)
			d.r.SetColor(color.RGBA{100, 180, 255, 255})
			d.r.DrawQuadSoft(d.dst, quadB, d.s(3), d.s(4))
			return nil
		}},
	}
}

func rectsPanels() []panel {
	return []panel{
		{"DrawIntArea / StrokeIntArea", func(d demoCtx) []string {
			d.r.SetColor(color.RGBA{90, 215, 255, 255})
			d.r.DrawIntArea(d.dst, d.xi(28), d.yi(38), int(d.s(122)), int(d.s(88)))
			d.r.SetColor(color.RGBA{255, 194, 84, 255})
			d.r.StrokeIntArea(d.dst, d.xi(172), d.yi(28), int(d.s(116)), int(d.s(106)), int(d.s(6)), int(d.s(8)))
			return nil
		}},
		{"DrawRingSector", func(d demoCtx) []string {
			d.r.SetColor(color.RGBA{170, 120, 255, 255})
			d.r.DrawRingSector(d.dst, d.x(160), d.y(100), d.s(28), d.s(68), shapes.RadsTopRight, shapes.RadsBottomLeft+0.4, d.s(10))
			return nil
		}},
		{"StrokeRingSector", func(d demoCtx) []string {
			d.r.SetColor(color.RGBA{255, 210, 84, 255})
			d.r.StrokeRingSector(d.dst, d.x(160), d.y(100), d.s(30), d.s(72), d.s(10), shapes.RadsLeft+0.2, shapes.RadsTopRight, d.s(8))
			return nil
		}},
		{"DrawPieRate", func(d demoCtx) []string {
			rate := (math.Sin(float64(d.a.tick)*0.03) + 1) * 0.5
			d.r.SetColor(color.RGBA{114, 217, 255, 255})
			d.r.DrawPieRate(d.dst, d.x(160), d.y(100), d.s(60), shapes.RadsTop, rate, d.s(10))
			return []string{fmt.Sprintf("rate = %.2f", rate)}
		}},
		{"StrokePieRate", func(d demoCtx) []string {
			rate := (math.Cos(float64(d.a.tick)*0.04) + 1) * 0.5
			d.r.SetColor(color.RGBA{255, 123, 123, 255})
			d.r.StrokePieRate(d.dst, d.x(160), d.y(100), d.s(64), d.s(14), shapes.RadsRight, rate, d.s(8))
			return []string{fmt.Sprintf("rate = %.2f", rate)}
		}},
		{"Rounded stroke contrast", func(d demoCtx) []string {
			d.r.SetColor(color.RGBA{88, 205, 255, 255})
			d.r.StrokeArea(d.dst, d.x(26), d.y(40), d.s(120), d.s(104), 0, d.s(12), 0)
			d.r.SetColor(color.RGBA{255, 200, 84, 255})
			d.r.StrokeArea(d.dst, d.x(176), d.y(40), d.s(120), d.s(104), 0, d.s(12), d.s(28))
			return []string{"sharp vs rounded corners"}
		}},
	}
}

func colorPanels() []panel {
	return []panel{
		{"SimpleGradient", func(d demoCtx) []string {
			area := d.tempRef(tempAuxA, 262, 150, true)
			d.r.SimpleGradient(area, color.RGBA{45, 164, 255, 255}, color.RGBA{171, 94, 255, 255}, shapes.DirRadsTLBR)
			d.blit(area, 30, 30)
			return nil
		}},
		{"Gradient (masked)", func(d demoCtx) []string {
			mask := d.tempRef(tempAuxA, 220, 130, true)
			d.r.SetColor(color.White)
			d.r.DrawArea(mask, d.s(10), d.s(10), d.s(200), d.s(110), d.s(30))
			d.r.Gradient(d.dst, mask, d.x(48), d.y(30), color.RGBA{255, 190, 80, 255}, color.RGBA{255, 80, 120, 255}, 8, shapes.DirRadsLTR, 1.4)
			return nil
		}},
		{"GradientRadial", func(d demoCtx) []string {
			d.r.GradientRadial(d.dst, d.x(160), d.y(100), color.RGBA{70, 200, 255, 255}, color.RGBA{0, 0, 0, 0}, d.s(10), d.s(66), d.s(82), 8, 2)
			return nil
		}},
		{"ColorizeByLightness", func(d demoCtx) []string {
			// Recolor a shaded sprite purely by its luminance ramp.
			src := d.scaled(tempAuxA, d.a.spriteB)
			d.r.ColorizeByLightness(d.dst, src, d.x(50), d.y(18), color.RGBA{30, 60, 255, 255}, color.RGBA{255, 170, 60, 255}, 0.1, 0.9, 0, 1.0)
			return []string{"luminance -> two-color ramp"}
		}},
		{"OklabShift", func(d demoCtx) []string {
			src := d.scaled(tempAuxA, d.a.spriteB)
			d.r.OklabShift(d.dst, src, d.x(50), d.y(18), 0.15, 0.08, float32(math.Sin(float64(d.a.tick)*0.03))*0.5)
			return []string{"animated hue shift"}
		}},
		{"ColorMix", func(d demoCtx) []string {
			// Two translucent blobs interpolated with mix() rather than composited.
			base := d.tempRef(tempAuxA, 220, 130, true)
			over := d.tempRef(tempAuxB, 220, 130, true)
			d.r.GradientRadial(base, d.s(78), d.s(64), color.RGBA{70, 190, 255, 255}, color.RGBA{70, 190, 255, 0}, 0, d.s(20), d.s(60), 0, 2)
			d.r.GradientRadial(over, d.s(148), d.s(66), color.RGBA{255, 128, 84, 255}, color.RGBA{255, 128, 84, 0}, 0, d.s(20), d.s(60), 0, 2)
			mix := float32((math.Sin(float64(d.a.tick)*0.03) + 1) * 0.5)
			d.r.ColorMix(d.dst, base, over, d.xi(50), d.yi(40), 1.0, mix)
			return []string{fmt.Sprintf("mixLevel = %.2f", mix)}
		}},
		{"DitherMat4", func(d demoCtx) []string {
			mask := d.tempRef(tempAuxA, 220, 130, true)
			d.r.SimpleGradient(mask, color.RGBA{20, 20, 20, 255}, color.RGBA{240, 240, 240, 255}, shapes.DirRadsLTR)
			rgbaColors := []float32{
				0.12, 0.16, 0.24, 1,
				0.26, 0.50, 0.96, 1,
				0.98, 0.74, 0.29, 1,
				0.98, 0.37, 0.46, 1,
			}
			dmat := [16]float32{0, 8, 2, 10, 12, 4, 14, 6, 3, 11, 1, 9, 15, 7, 13, 5}
			d.r.DitherMat4(d.dst, mask, d.x(48), d.y(40), 0, 0, rgbaColors, dmat, 1, 0)
			return nil
		}},
		{"FlatPaint", func(d demoCtx) []string {
			// Any silhouette in a mask, painted with a flat vertex color.
			mask := d.tempRef(tempAuxA, 200, 150, true)
			d.r.SetColor(color.White)
			d.r.DrawTriangle(mask, d.sf(30), d.sf(120), d.sf(100), d.sf(20), d.sf(170), d.sf(120), d.sf(10))
			d.r.DrawCircle(mask, d.s(100), d.s(112), d.s(34))
			d.r.SetColor(color.RGBA{255, 150, 72, 255})
			d.r.FlatPaint(d.dst, mask, d.x(60), d.y(30))
			return []string{"any mask -> flat color"}
		}},
	}
}

func tilePanels() []panel {
	return []panel{
		{"TileDotsGrid", func(d demoCtx) []string {
			area := d.tempRef(tempAuxA, 300, 180, false)
			area.Fill(color.RGBA{12, 16, 24, 255})
			d.r.SetColor(color.RGBA{95, 200, 255, 255})
			d.r.TileDotsGrid(area, d.s(5), d.s(22), 0, 0)
			d.blit(area, 10, 15)
			return nil
		}},
		{"TileDotsHex", func(d demoCtx) []string {
			area := d.tempRef(tempAuxA, 300, 180, false)
			area.Fill(color.RGBA{12, 16, 24, 255})
			d.r.SetColor(color.RGBA{255, 194, 84, 255})
			d.r.TileDotsHex(area, d.s(4), d.s(26), 0, 0)
			d.blit(area, 10, 15)
			return nil
		}},
		{"TileRectsGrid", func(d demoCtx) []string {
			area := d.tempRef(tempAuxA, 300, 180, false)
			area.Fill(color.RGBA{14, 18, 28, 255})
			d.r.SetColor(color.RGBA{83, 220, 171, 255})
			d.r.TileRectsGrid(area, d.s(30), d.s(24), d.s(18), d.s(14), 0, 0)
			d.blit(area, 10, 15)
			return nil
		}},
		{"TileTriUpGrid / TileTriHex", func(d demoCtx) []string {
			left := d.tempRef(tempAuxA, 148, 180, false)
			right := d.tempRef(tempAuxB, 148, 180, false)
			left.Fill(color.RGBA{14, 18, 28, 255})
			right.Fill(color.RGBA{14, 18, 28, 255})
			d.r.SetColor(color.RGBA{255, 118, 118, 255})
			d.r.TileTriUpGrid(left, d.s(22), d.s(11), 0, 0)
			d.r.SetColor(color.RGBA{114, 217, 255, 255})
			d.r.TileTriHex(right, d.s(20), d.s(11), 0, 0)
			d.blit(left, 10, 15)
			d.blit(right, 162, 15)
			return nil
		}},
		{"Noise", func(d demoCtx) []string {
			area := d.tempRef(tempAuxA, 300, 180, false)
			area.Fill(color.RGBA{22, 26, 36, 255})
			d.r.Noise(area, 0.18, 1.0, float32(d.a.tick)*0.01)
			d.blit(area, 10, 15)
			return []string{"animated grain"}
		}},
		{"NoiseGolden", func(d demoCtx) []string {
			area := d.tempRef(tempAuxA, 300, 180, false)
			area.Fill(color.RGBA{22, 26, 36, 255})
			d.r.NoiseGolden(area, d.s(12), 0.22, float32(d.a.tick)*0.01)
			d.blit(area, 10, 15)
			return nil
		}},
	}
}

func maskPanels() []panel {
	return []panel{
		{"Mask", func(d demoCtx) []string {
			sp := d.scaled(tempAuxA, d.a.spriteA)
			mk := d.scaled(tempAuxB, d.a.maskA)
			d.r.Mask(d.dst, sp, mk, d.x(50), d.y(20))
			return []string{"sprite kept only where mask is opaque"}
		}},
		{"MaskThreshold", func(d demoCtx) []string {
			sp := d.scaled(tempAuxA, d.a.spriteA)
			mk := d.scaled(tempAuxB, d.a.maskA)
			reveal := float32((math.Sin(float64(d.a.tick)*0.03) + 1) * 0.5)
			d.r.MaskThreshold(d.dst, sp, mk, reveal, d.x(50), d.y(20))
			return []string{fmt.Sprintf("reveal = %.2f", reveal)}
		}},
		{"MaskCircle", func(d demoCtx) []string {
			sp := d.scaled(tempAuxA, d.a.spriteA)
			soft := float32(18 + 8*math.Sin(float64(d.a.tick)*0.04))
			d.r.MaskCircle(d.dst, sp, d.x(160), d.y(100), d.s(50), d.s(20), d.s(60), d.s(soft))
			return []string{"soft circular spotlight"}
		}},
		{"DrawAlphaMaskCirc", func(d demoCtx) []string {
			d.r.SetColor(color.RGBA{114, 217, 255, 255})
			d.r.DrawAlphaMaskCirc(d.dst, d.x(160), d.y(100), d.s(84), d.s(24), shapes.MaskPatternFlare)
			return []string{"procedural flare pattern"}
		}},
		{"HalftoneTri", func(d demoCtx) []string {
			sp := d.scaled(tempAuxA, d.a.spriteB)
			d.r.HalftoneTri(d.dst, sp, d.x(50), d.y(20), d.s(24), d.s(4), d.s(22), 0, 0)
			return nil
		}},
	}
}

func effectPanels() []panel {
	return []panel{
		{"ApplyExpansion", func(d demoCtx) []string {
			mk := d.scaled(tempAuxA, d.a.maskA)
			thick := float32(12 + 8*math.Sin(float64(d.a.tick)*0.04))
			d.r.SetColor(color.RGBA{255, 194, 84, 255})
			d.r.ApplyExpansion(d.dst, mk, d.x(50), d.y(20), d.s(thick))
			d.blit(mk, 50, 20)
			return nil
		}},
		{"ApplyErosion", func(d demoCtx) []string {
			mk := d.scaled(tempAuxA, d.a.maskA)
			thick := float32(10 + 6*math.Sin(float64(d.a.tick)*0.03))
			d.r.SetColor(color.RGBA{114, 217, 255, 255})
			d.r.ApplyErosion(d.dst, mk, d.x(50), d.y(20), d.s(thick))
			return nil
		}},
		{"ApplyOutline", func(d demoCtx) []string {
			mk := d.scaled(tempAuxA, d.a.maskA)
			sp := d.scaled(tempAuxB, d.a.spriteA)
			d.r.SetColor(color.RGBA{255, 120, 120, 255})
			d.r.ApplyOutline(d.dst, mk, d.x(50), d.y(20), d.s(8))
			d.blit(sp, 50, 20)
			return nil
		}},
		{"ApplyBlur", func(d demoCtx) []string {
			mk := d.scaled(tempAuxA, d.a.maskA)
			d.r.SetColor(color.RGBA{120, 220, 255, 255})
			d.r.ApplyBlur(d.dst, mk, d.x(50), d.y(20), d.s(14), 0)
			return nil
		}},
		{"ApplyBlur2", func(d demoCtx) []string {
			mk := d.scaled(tempAuxA, d.a.maskA)
			d.r.SetColor(color.RGBA{170, 120, 255, 255})
			d.r.ApplyBlur2(d.dst, mk, d.x(50), d.y(20), d.s(12), 0)
			return nil
		}},
		{"ApplyShadow", func(d demoCtx) []string {
			mk := d.scaled(tempAuxA, d.a.maskA)
			sp := d.scaled(tempAuxB, d.a.spriteA)
			d.r.SetColor(color.RGBA{0, 0, 0, 180})
			d.r.ApplyShadow(d.dst, mk, d.x(50), d.y(20), d.s(18), d.s(14), d.s(18), shapes.ClampBottom)
			d.blit(sp, 50, 20)
			return nil
		}},
		{"ApplyZoomShadow", func(d demoCtx) []string {
			mk := d.scaled(tempAuxA, d.a.maskA)
			sp := d.scaled(tempAuxB, d.a.spriteA)
			d.r.SetColor(color.RGBA{0, 0, 0, 180})
			d.r.ApplyZoomShadow(d.dst, mk, d.x(50), d.y(20), d.s(12), d.s(10), 1.3, shapes.ClampBottom)
			d.blit(sp, 50, 20)
			return nil
		}},
		{"ApplySimpleGlow / ApplyGlow", func(d demoCtx) []string {
			mk := d.scaled(tempAuxA, d.a.maskA)
			d.r.SetColor(color.RGBA{255, 210, 84, 255})
			d.r.ApplySimpleGlow(d.dst, mk, d.x(4), d.y(20), d.s(14))
			d.r.SetColor(color.RGBA{114, 217, 255, 255})
			d.r.ApplyGlow(d.dst, mk, d.x(150), d.y(20), d.s(20), d.s(10), 0.15, 0.9, 0)
			d.blit(mk, 4, 20)
			d.blit(mk, 150, 20)
			return nil
		}},
		{"ApplyScanlinesSharp", func(d demoCtx) []string {
			area := d.tempRef(tempAuxA, 300, 180, false)
			d.r.SimpleGradient(area, color.RGBA{60, 160, 255, 255}, color.RGBA{170, 110, 255, 255}, shapes.DirRadsTTB)
			d.r.ApplyScanlinesSharp(area, int(d.s(2)+1), int(d.s(4)+1), 0.55, float32(d.a.tick)*0.04)
			d.blit(area, 10, 15)
			return nil
		}},
		{"ApplyWaveLines", func(d demoCtx) []string {
			area := d.tempRef(tempAuxA, 300, 180, false)
			area.Fill(color.RGBA{12, 16, 24, 255})
			d.r.SetColor(color.RGBA{255, 210, 84, 255})
			d.r.ApplyWaveLines(area, d.s(6), 0.2, 0.8, 6, float32(d.a.tick)*0.03, math.Pi/8)
			d.blit(area, 10, 15)
			return nil
		}},
		{"ApplyExpansionRect", func(d demoCtx) []string {
			mask := d.tempRef(tempAuxA, 160, 110, true)
			d.r.SetColor(color.White)
			d.r.DrawArea(mask, d.s(20), d.s(18), d.s(120), d.s(74), 0)
			d.r.SetColor(color.RGBA{114, 217, 255, 255})
			d.r.ApplyExpansionRect(d.dst, mask, d.x(80), d.y(50), d.s(10))
			d.blit(mask, 80, 50)
			return nil
		}},
	}
}

func mappingPanels() []panel {
	return []panel{
		{"Scale", func(d demoCtx) []string {
			d.r.Scale(d.dst, d.a.spriteB, d.x(42), d.y(18), d.k*0.75, false)
			return []string{"bilinear sampling"}
		}},
		{"Scale (pixelated)", func(d demoCtx) []string {
			d.r.Scale(d.dst, d.a.spriteB, d.x(42), d.y(18), d.k*0.75, true)
			return []string{"nearest-like sampling"}
		}},
		{"MapQuad4", func(d demoCtx) []string {
			quad := [4]shapes.PointF32{d.pt(34, 32), d.pt(266, 18), d.pt(290, 154), d.pt(54, 142)}
			d.r.MapQuad4(d.dst, d.a.spriteA, quad)
			return nil
		}},
		{"MapProjective", func(d demoCtx) []string {
			quad := [4]shapes.PointF32{d.pt(56, 26), d.pt(260, 46), d.pt(280, 154), d.pt(26, 168)}
			d.r.MapProjective(d.dst, d.a.spriteA, quad)
			return []string{"perspective-correct mapping"}
		}},
		{"JFMapFill + JFMHeat", func(d demoCtx) []string {
			js := d.scaled(tempAuxA, d.a.jfMask)
			jb := js.Bounds()
			jfmap := d.r.UnsafeTemp(tempAuxB, jb.Dx(), jb.Dy(), true)
			d.r.JFMapFill(jfmap, js, int(d.s(96)), 0.001, 1)
			d.r.JFMHeat(d.dst, jfmap, d.x(50), d.y(20), d.s(96))
			return []string{"distance field heatmap"}
		}},
		{"JFMExpand", func(d demoCtx) []string {
			js := d.scaled(tempAuxA, d.a.jfMask)
			d.r.SetColor(color.RGBA{255, 194, 84, 255})
			d.r.JFMExpand(d.dst, js, nil, d.x(50), d.y(20), d.s(12), shapes.AAMargin*4)
			d.blit(js, 50, 20)
			return nil
		}},
		{"JFMErode", func(d demoCtx) []string {
			js := d.scaled(tempAuxA, d.a.jfMask)
			d.r.SetColor(color.RGBA{114, 217, 255, 255})
			d.r.JFMErode(d.dst, js, nil, d.x(50), d.y(20), d.s(10), shapes.AAMargin)
			return nil
		}},
		{"JFMapBoundary", func(d demoCtx) []string {
			js := d.scaled(tempAuxA, d.a.jfMask)
			jb := js.Bounds()
			jfmap := d.r.UnsafeTemp(tempAuxB, jb.Dx(), jb.Dy(), true)
			d.r.JFMapBoundary(jfmap, js, int(d.s(96)), 0.001, 1, false, false)
			d.blit(jfmap, 50, 20)
			return nil
		}},
	}
}
