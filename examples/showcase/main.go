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

const (
	screenWidth    = 1480
	screenHeight   = 980
	panelGap       = 16
	panelCols      = 4
	panelRows      = 3
	panelTextLine1 = 8
	panelTextLine2 = 24
)

type demoPage struct {
	title string
	draw  func(*Showcase, *ebiten.Image)
}

type Showcase struct {
	renderer *shapes.Renderer
	tick     uint64
	page     int
	pages    []demoPage

	spriteA *ebiten.Image
	spriteB *ebiten.Image
	maskA   *ebiten.Image
	jfMask  *ebiten.Image
}

func main() {
	ebiten.SetWindowSize(screenWidth, screenHeight)
	ebiten.SetWindowTitle("go-shapes showcase")
	app := NewShowcase()
	if err := ebiten.RunGame(app); err != nil {
		log.Fatal(err)
	}
}

func NewShowcase() *Showcase {
	app := &Showcase{renderer: shapes.NewRenderer()}
	app.makeAssets()
	app.pages = []demoPage{
		{title: "Primitives", draw: (*Showcase).drawPrimitivesPage},
		{title: "Rects and sectors", draw: (*Showcase).drawRectsPage},
		{title: "Color and tiles", draw: (*Showcase).drawColorPage},
		{title: "Masks and noise", draw: (*Showcase).drawMaskNoisePage},
		{title: "Filters and light", draw: (*Showcase).drawEffectsPage},
		{title: "Mapping and warps", draw: (*Showcase).drawWarpPage},
		{title: "JFM and generated assets", draw: (*Showcase).drawJFMPage},
	}
	return app
}

func (a *Showcase) Layout(_, _ int) (int, int) { return screenWidth, screenHeight }

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
	a.paintBackground(screen)
	a.drawHeader(screen)
	a.pages[a.page].draw(a, screen)
}

func (a *Showcase) paintBackground(screen *ebiten.Image) {
	screen.Fill(color.RGBA{12, 14, 20, 255})
	grid := ebiten.NewImage(screenWidth, screenHeight)
	a.renderer.SetColor(color.RGBA{22, 27, 35, 255})
	a.renderer.TileDotsGrid(grid, 1.4, 28, 0, 0)
	var opts ebiten.DrawImageOptions
	opts.ColorScale.ScaleAlpha(0.45)
	screen.DrawImage(grid, &opts)
}

func (a *Showcase) drawHeader(screen *ebiten.Image) {
	title := fmt.Sprintf("go-shapes showcase  •  page %d/%d  •  %s", a.page+1, len(a.pages), a.pages[a.page].title)
	help := "Left/Right or PgUp/PgDn to switch pages • Esc to quit"
	ebitenutil.DebugPrintAt(screen, title, 20, 16)
	ebitenutil.DebugPrintAt(screen, help, 20, 34)
}

func (a *Showcase) panelRect(col, row int) image.Rectangle {
	top := 70
	w := (screenWidth - panelGap*(panelCols+1)) / panelCols
	h := (screenHeight - top - panelGap*(panelRows+1)) / panelRows
	x := panelGap + col*(w+panelGap)
	y := top + panelGap + row*(h+panelGap)
	return image.Rect(x, y, x+w, y+h)
}

func centeredRect(outer image.Rectangle, w, h int) image.Rectangle {
	x := outer.Min.X + (outer.Dx()-w)/2
	y := outer.Min.Y + (outer.Dy()-h)/2
	return image.Rect(x, y, x+w, y+h)
}

func (a *Showcase) drawPanel(screen *ebiten.Image, col, row int, title string, drawFn func(dst *ebiten.Image, bounds image.Rectangle)) {
	bounds := a.panelRect(col, row)
	panel := ebiten.NewImage(bounds.Dx(), bounds.Dy())
	panel.Fill(color.RGBA{18, 22, 31, 245})

	a.renderer.SetColor(color.RGBA{55, 68, 84, 255})
	a.renderer.StrokeArea(panel, 0, 0, float32(bounds.Dx()), float32(bounds.Dy()), 1, 0, 14)

	content := image.Rect(12, 30, bounds.Dx()-12, bounds.Dy()-12)

	demoW := min(content.Dx()-20, 320)
	demoH := min(content.Dy()-20, 180)
	demo := centeredRect(content, demoW, demoH)

	drawFn(panel, demo)
	ebitenutil.DebugPrintAt(panel, title, 12, 10)

	var opts ebiten.DrawImageOptions
	opts.GeoM.Translate(float64(bounds.Min.X), float64(bounds.Min.Y))
	screen.DrawImage(panel, &opts)
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

func (a *Showcase) makeAssets() {
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

func blit(dst, src *ebiten.Image, x, y float64) {
	var opts ebiten.DrawImageOptions
	opts.GeoM.Translate(x, y)
	dst.DrawImage(src, &opts)
}

func blitAt(dst, src *ebiten.Image, pt image.Point) {
	blit(dst, src, float64(pt.X), float64(pt.Y))
}

func (a *Showcase) drawPrimitivesPage(screen *ebiten.Image) {
	a.drawPanel(screen, 0, 0, "DrawArea / StrokeArea", func(dst *ebiten.Image, b image.Rectangle) {
		a.renderer.SetColor(color.RGBA{68, 180, 255, 255})
		a.renderer.DrawArea(dst, float32(b.Min.X+28), float32(b.Min.Y+44), 120, 86, 18)
		a.renderer.SetColor(color.RGBA{255, 204, 84, 255})
		a.renderer.StrokeArea(dst, float32(b.Min.X+170), float32(b.Min.Y+38), 120, 98, 5, 8, 24)
	})

	a.drawPanel(screen, 1, 0, "DrawLine / DrawCircle", func(dst *ebiten.Image, b image.Rectangle) {
		a.renderer.SetColor(color.RGBA{255, 120, 120, 255})
		a.renderer.DrawLine(dst, float64(b.Min.X+20), float64(b.Min.Y+130), float64(b.Max.X-30), float64(b.Min.Y+30), 10)
		a.renderer.SetColor(color.RGBA{90, 225, 170, 255})
		a.renderer.DrawCircle(dst, float32(b.Min.X+115), float32(b.Min.Y+86), 46)
	})

	a.drawPanel(screen, 2, 0, "StrokeCircle / DrawRing", func(dst *ebiten.Image, b image.Rectangle) {
		a.renderer.SetColor(color.RGBA{114, 217, 255, 255})
		a.renderer.StrokeCircle(dst, float32(b.Min.X+84), float32(b.Min.Y+84), 46, 8)
		a.renderer.SetColor(color.RGBA{255, 193, 84, 255})
		a.renderer.DrawRing(dst, float32(b.Min.X+214), float32(b.Min.Y+84), 26, 52)
	})

	a.drawPanel(screen, 3, 0, "DrawPie / StrokePie", func(dst *ebiten.Image, b image.Rectangle) {
		t := 0.7 + 0.6*math.Sin(float64(a.tick)*0.03)
		a.renderer.SetColor(color.RGBA{255, 99, 132, 255})
		a.renderer.DrawPie(dst, float32(b.Min.X+90), float32(b.Min.Y+86), 52, shapes.RadsTop, shapes.RadsTop+t*math.Pi, 12)
		a.renderer.SetColor(color.RGBA{115, 215, 255, 255})
		a.renderer.StrokePie(dst, float32(b.Min.X+220), float32(b.Min.Y+86), 54, 12, shapes.RadsLeft, shapes.RadsLeft+t*math.Pi, 8)
	})

	a.drawPanel(screen, 0, 1, "DrawEllipse", func(dst *ebiten.Image, b image.Rectangle) {
		a.renderer.SetColor(color.RGBA{170, 120, 255, 255})
		a.renderer.DrawEllipse(dst, float32(b.Min.X+160), float32(b.Min.Y+84), 108, 50, math.Sin(float64(a.tick)*0.02)*0.7)
	})

	a.drawPanel(screen, 1, 1, "DrawTriangle / StrokeTriangle", func(dst *ebiten.Image, b image.Rectangle) {
		a.renderer.SetColor(color.RGBA{114, 217, 255, 255})
		a.renderer.DrawTriangle(dst, 50, 190, 110, 84, 180, 202, 10)
		a.renderer.SetColor(color.RGBA{255, 210, 96, 255})
		a.renderer.StrokeTriangle(dst, 190, 190, 250, 88, 310, 190, 8, 10)
	})

	a.drawPanel(screen, 2, 1, "DrawHexagon", func(dst *ebiten.Image, b image.Rectangle) {
		a.renderer.SetColor(color.RGBA{83, 220, 171, 255})
		a.renderer.DrawHexagon(dst, float32(b.Min.X+160), float32(b.Min.Y+88), 66, 16, float32(a.tick)*0.01)
	})

	a.drawPanel(screen, 3, 1, "DrawQuad / DrawQuadSoft", func(dst *ebiten.Image, b image.Rectangle) {
		quadA := [4]shapes.PointF32{{50, 92}, {174, 70}, {196, 172}, {38, 188}}
		quadB := [4]shapes.PointF32{{194, 98}, {304, 112}, {292, 200}, {170, 190}}
		a.renderer.SetColor(color.RGBA{255, 104, 104, 255})
		a.renderer.DrawQuad(dst, quadA, 0)
		a.renderer.SetColor(color.RGBA{100, 180, 255, 255})
		a.renderer.DrawQuadSoft(dst, quadB, 3, 4)
	})

	a.drawPanel(screen, 0, 2, "NewRect", func(dst *ebiten.Image, b image.Rectangle) {
		blit(dst, a.renderer.NewRect(150, 90), float64(b.Min.X+85), float64(b.Min.Y+40))
	})

	a.drawPanel(screen, 1, 2, "NewCircle", func(dst *ebiten.Image, b image.Rectangle) {
		blit(dst, a.renderer.NewCircle(54), float64(b.Min.X+104), float64(b.Min.Y+32))
	})

	a.drawPanel(screen, 2, 2, "NewRing", func(dst *ebiten.Image, b image.Rectangle) {
		blit(dst, a.renderer.NewRing(34, 64), float64(b.Min.X+94), float64(b.Min.Y+22))
	})

	a.drawPanel(screen, 3, 2, "Angles constants", func(dst *ebiten.Image, b image.Rectangle) {
		cx, cy := float32(b.Min.X+160), float32(b.Min.Y+86)
		a.renderer.SetColor(color.RGBA{90, 100, 120, 255})
		a.renderer.StrokeCircle(dst, cx, cy, 64, 2)
		a.renderer.SetColor(color.RGBA{255, 194, 84, 255})
		a.renderer.DrawLine(dst, float64(cx), float64(cy), float64(cx+64), float64(cy), 3)
		a.renderer.SetColor(color.RGBA{114, 217, 255, 255})
		a.renderer.DrawLine(dst, float64(cx), float64(cy), float64(cx), float64(cy+64), 3)
		ebitenutil.DebugPrintAt(dst, "0 = right", b.Min.X+16, b.Max.Y-panelTextLine2)
		ebitenutil.DebugPrintAt(dst, "pi/2 = bottom", b.Min.X+16, b.Max.Y-panelTextLine1)
	})
}

func (a *Showcase) drawRectsPage(screen *ebiten.Image) {
	a.drawPanel(screen, 0, 0, "DrawIntArea / StrokeIntArea", func(dst *ebiten.Image, b image.Rectangle) {
		a.renderer.SetColor(color.RGBA{90, 215, 255, 255})
		a.renderer.DrawIntArea(dst, b.Min.X+28, b.Min.Y+38, 122, 88)
		a.renderer.SetColor(color.RGBA{255, 194, 84, 255})
		a.renderer.StrokeIntArea(dst, b.Min.X+172, b.Min.Y+28, 116, 106, 6, 8)
	})

	a.drawPanel(screen, 1, 0, "DrawRect / StrokeRect", func(dst *ebiten.Image, b image.Rectangle) {
		a.renderer.SetColor(color.RGBA{83, 220, 171, 255})
		a.renderer.DrawRect(dst, image.Rect(b.Min.X+24, b.Min.Y+34, b.Min.X+146, b.Min.Y+118), 18)
		a.renderer.SetColor(color.RGBA{255, 118, 118, 255})
		a.renderer.StrokeRect(dst, image.Rect(b.Min.X+176, b.Min.Y+30, b.Min.X+300, b.Min.Y+124), 5, 9, 22)
	})

	a.drawPanel(screen, 2, 0, "DrawRingSector", func(dst *ebiten.Image, b image.Rectangle) {
		a.renderer.SetColor(color.RGBA{170, 120, 255, 255})
		a.renderer.DrawRingSector(dst, float32(b.Min.X+160), float32(b.Min.Y+84), 28, 68, shapes.RadsTopRight, shapes.RadsBottomLeft+0.4, 10)
	})

	a.drawPanel(screen, 3, 0, "StrokeRingSector", func(dst *ebiten.Image, b image.Rectangle) {
		a.renderer.SetColor(color.RGBA{255, 210, 84, 255})
		a.renderer.StrokeRingSector(dst, float32(b.Min.X+160), float32(b.Min.Y+84), 30, 72, 10, shapes.RadsLeft+0.2, shapes.RadsTopRight, 8)
	})

	a.drawPanel(screen, 0, 1, "DrawPieRate", func(dst *ebiten.Image, b image.Rectangle) {
		rate := (math.Sin(float64(a.tick)*0.03) + 1) * 0.5
		a.renderer.SetColor(color.RGBA{114, 217, 255, 255})
		a.renderer.DrawPieRate(dst, float32(b.Min.X+160), float32(b.Min.Y+84), 60, shapes.RadsTop, rate, 10)
		ebitenutil.DebugPrintAt(dst, fmt.Sprintf("rate = %.2f", rate), b.Min.X+18, b.Max.Y-panelTextLine1)
	})

	a.drawPanel(screen, 1, 1, "StrokePieRate", func(dst *ebiten.Image, b image.Rectangle) {
		rate := (math.Cos(float64(a.tick)*0.04) + 1) * 0.5
		a.renderer.SetColor(color.RGBA{255, 123, 123, 255})
		a.renderer.StrokePieRate(dst, float32(b.Min.X+160), float32(b.Min.Y+84), 64, 14, shapes.RadsRight, rate, 8)
		ebitenutil.DebugPrintAt(dst, fmt.Sprintf("rate = %.2f", rate), b.Min.X+18, b.Max.Y-panelTextLine1)
	})

	a.drawPanel(screen, 2, 1, "Rounded stroke contrast", func(dst *ebiten.Image, b image.Rectangle) {
		a.renderer.SetColor(color.RGBA{88, 205, 255, 255})
		a.renderer.StrokeArea(dst, float32(b.Min.X+26), float32(b.Min.Y+30), 120, 104, 0, 12, 0)
		a.renderer.SetColor(color.RGBA{255, 200, 84, 255})
		a.renderer.StrokeArea(dst, float32(b.Min.X+176), float32(b.Min.Y+30), 120, 104, 0, 12, 28)
	})

	a.drawPanel(screen, 3, 1, "BlendMultiply", func(dst *ebiten.Image, b image.Rectangle) {
		dst.Fill(color.RGBA{30, 34, 45, 255})
		a.renderer.SetBlend(ebiten.BlendSourceOver)
		a.renderer.SetColor(color.RGBA{255, 110, 110, 220})
		a.renderer.DrawCircle(dst, float32(b.Min.X+136), float32(b.Min.Y+80), 50)
		a.renderer.SetBlend(shapes.BlendMultiply)
		a.renderer.SetColor(color.RGBA{110, 180, 255, 220})
		a.renderer.DrawCircle(dst, float32(b.Min.X+192), float32(b.Min.Y+88), 50)
		a.renderer.SetBlend(ebiten.BlendSourceOver)
	})

	a.drawPanel(screen, 0, 2, "BlendSubtract", func(dst *ebiten.Image, b image.Rectangle) {
		dst.Fill(color.RGBA{50, 70, 100, 255})
		a.renderer.SetColor(color.RGBA{255, 100, 100, 255})
		a.renderer.DrawArea(dst, float32(b.Min.X+44), float32(b.Min.Y+34), 120, 90, 20)
		a.renderer.SetBlend(shapes.BlendSubtract)
		a.renderer.SetColor(color.RGBA{40, 40, 40, 255})
		a.renderer.DrawCircle(dst, float32(b.Min.X+198), float32(b.Min.Y+84), 54)
		a.renderer.SetBlend(ebiten.BlendSourceOver)
	})

	a.drawPanel(screen, 1, 2, "SetCustomVAs (quad shader helper)", func(dst *ebiten.Image, b image.Rectangle) {
		a.renderer.SetColor(color.RGBA{95, 200, 255, 255})
		quad := [4]shapes.PointF32{{48, 72}, {274, 66}, {250, 174}, {70, 180}}
		a.renderer.DrawQuadSoft(dst, quad, 10, 12)
		ebitenutil.DebugPrintAt(dst, "Used internally by many shader paths.", b.Min.X+16, b.Max.Y-panelTextLine1)
	})

	a.drawPanel(screen, 2, 2, "ScaleAlphaBy", func(dst *ebiten.Image, b image.Rectangle) {
		a.renderer.SetColor(color.RGBA{255, 194, 84, 255})
		a.renderer.DrawCircle(dst, float32(b.Min.X+120), float32(b.Min.Y+84), 54)
		a.renderer.ScaleAlphaBy(0.45)
		a.renderer.DrawCircle(dst, float32(b.Min.X+190), float32(b.Min.Y+84), 54)
		a.renderer.SetColor(color.White)
	})

	a.drawPanel(screen, 3, 2, "PointF32 helpers", func(dst *ebiten.Image, b image.Rectangle) {
		p0 := shapes.PointF32{70, 130}
		p1 := shapes.PointF32{255, 42}
		d := p1.Sub(p0)
		u := d.Normalize().Scale(80)
		a.renderer.SetColor(color.RGBA{114, 217, 255, 255})
		a.renderer.DrawLine(dst, float64(p0.X), float64(p0.Y), float64(p1.X), float64(p1.Y), 4)
		a.renderer.SetColor(color.RGBA{255, 210, 84, 255})
		a.renderer.DrawLine(dst, float64(p0.X), float64(p0.Y), float64(p0.X+u.X), float64(p0.Y+u.Y), 7)
		ebitenutil.DebugPrintAt(dst, fmt.Sprintf("len = %.1f", d.Length()), b.Min.X+18, b.Max.Y-panelTextLine1)
	})
}

func (a *Showcase) drawColorPage(screen *ebiten.Image) {
	a.drawPanel(screen, 0, 0, "FlatPaint", func(dst *ebiten.Image, b image.Rectangle) {
		mask := ebiten.NewImage(180, 120)
		a.renderer.SetColor(color.White)
		a.renderer.DrawHexagon(mask, 90, 60, 48, 10, 0)
		a.renderer.SetColor(color.RGBA{255, 150, 72, 255})
		a.renderer.FlatPaint(dst, mask, float32(b.Min.X+70), float32(b.Min.Y+28))
	})

	a.drawPanel(screen, 1, 0, "SimpleGradient", func(dst *ebiten.Image, b image.Rectangle) {
		a.renderer.SimpleGradient(dst.SubImage(image.Rect(b.Min.X+30, b.Min.Y+26, b.Min.X+292, b.Min.Y+144)).(*ebiten.Image), color.RGBA{45, 164, 255, 255}, color.RGBA{171, 94, 255, 255}, shapes.DirRadsTLBR)
	})

	a.drawPanel(screen, 2, 0, "Gradient", func(dst *ebiten.Image, b image.Rectangle) {
		mask := ebiten.NewImage(220, 130)
		a.renderer.SetColor(color.White)
		a.renderer.DrawArea(mask, 10, 10, 200, 110, 30)
		a.renderer.Gradient(dst, mask, float32(b.Min.X+48), float32(b.Min.Y+22), color.RGBA{255, 190, 80, 255}, color.RGBA{255, 80, 120, 255}, 8, shapes.DirRadsLTR, 1.4)
	})

	a.drawPanel(screen, 3, 0, "GradientRadial", func(dst *ebiten.Image, b image.Rectangle) {
		tmp := ebiten.NewImage(220, 150)
		a.renderer.GradientRadial(tmp, 110, 75, color.RGBA{70, 200, 255, 255}, color.RGBA{0, 0, 0, 0}, 10, 60, 74, 8, 2)
		blit(dst, tmp, float64(b.Min.X+48), float64(b.Min.Y+16))
	})

	a.drawPanel(screen, 0, 1, "ColorizeByLightness", func(dst *ebiten.Image, b image.Rectangle) {
		src := ebiten.NewImage(220, 130)
		a.renderer.SimpleGradient(src, color.RGBA{20, 20, 20, 255}, color.RGBA{240, 240, 240, 255}, shapes.DirRadsLTR)
		a.renderer.ColorizeByLightness(dst, src, float32(b.Min.X+48), float32(b.Min.Y+20), color.RGBA{30, 60, 255, 255}, color.RGBA{255, 120, 60, 255}, 0, 1, 8, 1.0)
	})

	a.drawPanel(screen, 1, 1, "OklabShift", func(dst *ebiten.Image, b image.Rectangle) {
		a.renderer.OklabShift(dst, a.spriteB, float32(b.Min.X+50), float32(b.Min.Y+18), 0.15, 0.08, float32(math.Sin(float64(a.tick)*0.03))*0.5)
	})

	a.drawPanel(screen, 2, 1, "ColorMix", func(dst *ebiten.Image, b image.Rectangle) {
		base := ebiten.NewImage(220, 130)
		over := ebiten.NewImage(220, 130)
		base.Fill(color.RGBA{35, 45, 70, 255})
		a.renderer.SetColor(color.RGBA{75, 185, 255, 255})
		a.renderer.DrawCircle(base, 76, 66, 48)
		a.renderer.SetColor(color.RGBA{255, 128, 84, 255})
		a.renderer.DrawHexagon(over, 140, 65, 50, 12, 0.2)
		a.renderer.ColorMix(dst, base, over, b.Min.X+48, b.Min.Y+20, 0.9, 0.55)
	})

	a.drawPanel(screen, 3, 1, "DitherMat4", func(dst *ebiten.Image, b image.Rectangle) {
		mask := ebiten.NewImage(220, 130)
		a.renderer.SimpleGradient(mask, color.RGBA{20, 20, 20, 255}, color.RGBA{240, 240, 240, 255}, shapes.DirRadsLTR)
		rgbaColors := []float32{
			0.12, 0.16, 0.24, 1,
			0.26, 0.50, 0.96, 1,
			0.98, 0.74, 0.29, 1,
			0.98, 0.37, 0.46, 1,
		}
		dmat := [16]float32{0, 8, 2, 10, 12, 4, 14, 6, 3, 11, 1, 9, 15, 7, 13, 5}
		a.renderer.DitherMat4(dst, mask, float32(b.Min.X+48), float32(b.Min.Y+20), 0, 0, rgbaColors, dmat, 1, 0)
	})

	a.drawPanel(screen, 0, 2, "TileDotsGrid", func(dst *ebiten.Image, b image.Rectangle) {
		area := ebiten.NewImage(272, 136)
		area.Fill(color.RGBA{12, 16, 24, 255})
		a.renderer.SetColor(color.RGBA{95, 200, 255, 255})
		a.renderer.TileDotsGrid(area, 5, 20, 0, 0)
		blitAt(dst, area, image.Pt(b.Min.X+24, b.Min.Y+18))
	})

	a.drawPanel(screen, 1, 2, "TileDotsHex", func(dst *ebiten.Image, b image.Rectangle) {
		area := ebiten.NewImage(272, 136)
		area.Fill(color.RGBA{12, 16, 24, 255})
		a.renderer.SetColor(color.RGBA{255, 194, 84, 255})
		a.renderer.TileDotsHex(area, 4, 24, 0, 0)
		blitAt(dst, area, image.Pt(b.Min.X+24, b.Min.Y+18))
	})

	a.drawPanel(screen, 2, 2, "TileRectsGrid", func(dst *ebiten.Image, b image.Rectangle) {
		area := ebiten.NewImage(272, 136)
		area.Fill(color.RGBA{14, 18, 28, 255})
		a.renderer.SetColor(color.RGBA{83, 220, 171, 255})
		a.renderer.TileRectsGrid(area, 28, 22, 16, 12, 0, 0)
		blitAt(dst, area, image.Pt(b.Min.X+24, b.Min.Y+18))
	})

	a.drawPanel(screen, 3, 2, "TileTriUpGrid / TileTriHex", func(dst *ebiten.Image, b image.Rectangle) {
		left := ebiten.NewImage(136, 132)
		right := ebiten.NewImage(136, 132)
		left.Fill(color.RGBA{14, 18, 28, 255})
		right.Fill(color.RGBA{14, 18, 28, 255})
		a.renderer.SetColor(color.RGBA{255, 118, 118, 255})
		a.renderer.TileTriUpGrid(left, 20, 10, 0, 0)
		a.renderer.SetColor(color.RGBA{114, 217, 255, 255})
		a.renderer.TileTriHex(right, 18, 10, 0, 0)
		blitAt(dst, left, image.Pt(b.Min.X+18, b.Min.Y+22))
		blitAt(dst, right, image.Pt(b.Min.X+166, b.Min.Y+22))
	})
}

func (a *Showcase) drawMaskNoisePage(screen *ebiten.Image) {
	a.drawPanel(screen, 0, 0, "Mask", func(dst *ebiten.Image, b image.Rectangle) {
		a.renderer.Mask(dst, a.spriteA, a.maskA, float32(b.Min.X+50), float32(b.Min.Y+20))
	})

	a.drawPanel(screen, 1, 0, "MaskAt", func(dst *ebiten.Image, b image.Rectangle) {
		t := float32(math.Sin(float64(a.tick)*0.03)) * 18
		a.renderer.MaskAt(dst, a.spriteA, a.maskA, float32(b.Min.X+50), float32(b.Min.Y+20), 12+t, -8)
	})

	a.drawPanel(screen, 2, 0, "MaskThreshold", func(dst *ebiten.Image, b image.Rectangle) {
		reveal := float32((math.Sin(float64(a.tick)*0.03) + 1) * 0.5)
		a.renderer.MaskThreshold(dst, a.spriteA, a.maskA, reveal, float32(b.Min.X+50), float32(b.Min.Y+20))
		ebitenutil.DebugPrintAt(dst, fmt.Sprintf("reveal = %.2f", reveal), b.Min.X+16, b.Max.Y-panelTextLine1)
	})

	a.drawPanel(screen, 3, 0, "MaskHorz", func(dst *ebiten.Image, b image.Rectangle) {
		inX := float32((math.Sin(float64(a.tick)*0.03) + 1) * 80)
		a.renderer.MaskHorz(dst, a.spriteA, float32(b.Min.X+50), float32(b.Min.Y+20), inX, inX+30)
	})

	a.drawPanel(screen, 0, 1, "MaskCircle", func(dst *ebiten.Image, b image.Rectangle) {
		soft := float32(18 + 8*math.Sin(float64(a.tick)*0.04))
		a.renderer.MaskCircle(dst, a.spriteA, float32(b.Min.X+160), float32(b.Min.Y+86), 50, 20, 56, soft)
	})

	a.drawPanel(screen, 1, 1, "DrawAlphaMaskCirc", func(dst *ebiten.Image, b image.Rectangle) {
		a.renderer.SetColor(color.RGBA{114, 217, 255, 255})
		a.renderer.DrawAlphaMaskCirc(dst, float32(b.Min.X+160), float32(b.Min.Y+86), 84, 24, shapes.MaskPatternFlare)
	})

	a.drawPanel(screen, 2, 1, "Noise", func(dst *ebiten.Image, b image.Rectangle) {
		area := ebiten.NewImage(264, 128)
		area.Fill(color.RGBA{22, 26, 36, 255})
		a.renderer.Noise(area, 0.18, 1.0, float32(a.tick)*0.01)
		blitAt(dst, area, image.Pt(b.Min.X+28, b.Min.Y+22))
	})

	a.drawPanel(screen, 3, 1, "NoiseGolden", func(dst *ebiten.Image, b image.Rectangle) {
		area := ebiten.NewImage(264, 128)
		area.Fill(color.RGBA{22, 26, 36, 255})
		a.renderer.NoiseGolden(area, 12, 0.22, float32(a.tick)*0.01)
		blitAt(dst, area, image.Pt(b.Min.X+28, b.Min.Y+22))
	})

	a.drawPanel(screen, 0, 2, "HalftoneTri", func(dst *ebiten.Image, b image.Rectangle) {
		a.renderer.HalftoneTri(dst, a.spriteB, float32(b.Min.X+50), float32(b.Min.Y+20), 24, 4, 22, 0, 0)
	})

	a.drawPanel(screen, 1, 2, "Alpha mask patterns", func(dst *ebiten.Image, b image.Rectangle) {
		left := ebiten.NewImage(136, 132)
		right := ebiten.NewImage(136, 132)
		a.renderer.SetColor(color.RGBA{255, 194, 84, 255})
		a.renderer.DrawAlphaMaskCirc(left, 68, 66, 54, 12, shapes.MaskPatternEllipseCuts)
		a.renderer.SetColor(color.RGBA{83, 220, 171, 255})
		a.renderer.DrawAlphaMaskCirc(right, 68, 66, 54, 12, shapes.MaskPatternPhiGrid)
		blitAt(dst, left, image.Pt(b.Min.X+18, b.Min.Y+20))
		blitAt(dst, right, image.Pt(b.Min.X+166, b.Min.Y+20))
	})

	a.drawPanel(screen, 2, 2, "MaskPatternCircMesh", func(dst *ebiten.Image, b image.Rectangle) {
		a.renderer.SetColor(color.RGBA{255, 120, 120, 255})
		a.renderer.DrawAlphaMaskCirc(dst, float32(b.Min.X+160), float32(b.Min.Y+86), 86, 6, shapes.MaskPatternCircMesh)
	})

	a.drawPanel(screen, 3, 2, "Source asset", func(dst *ebiten.Image, b image.Rectangle) {
		blit(dst, a.spriteA, float64(b.Min.X+50), float64(b.Min.Y+20))
	})
}

func (a *Showcase) drawEffectsPage(screen *ebiten.Image) {
	a.drawPanel(screen, 0, 0, "ApplyExpansion", func(dst *ebiten.Image, b image.Rectangle) {
		thick := float32(12 + 8*math.Sin(float64(a.tick)*0.04))
		a.renderer.SetColor(color.RGBA{255, 194, 84, 255})
		a.renderer.ApplyExpansion(dst, a.maskA, float32(b.Min.X+50), float32(b.Min.Y+20), thick)
		blit(dst, a.maskA, float64(b.Min.X+50), float64(b.Min.Y+20))
	})

	a.drawPanel(screen, 1, 0, "ApplyErosion", func(dst *ebiten.Image, b image.Rectangle) {
		thick := float32(10 + 6*math.Sin(float64(a.tick)*0.03))
		a.renderer.SetColor(color.RGBA{114, 217, 255, 255})
		a.renderer.ApplyErosion(dst, a.maskA, float32(b.Min.X+50), float32(b.Min.Y+20), thick)
	})

	a.drawPanel(screen, 2, 0, "ApplyOutline", func(dst *ebiten.Image, b image.Rectangle) {
		a.renderer.SetColor(color.RGBA{255, 120, 120, 255})
		a.renderer.ApplyOutline(dst, a.maskA, float32(b.Min.X+50), float32(b.Min.Y+20), 8)
		blit(dst, a.spriteA, float64(b.Min.X+50), float64(b.Min.Y+20))
	})

	a.drawPanel(screen, 3, 0, "ApplyBlur", func(dst *ebiten.Image, b image.Rectangle) {
		a.renderer.SetColor(color.RGBA{120, 220, 255, 255})
		a.renderer.ApplyBlur(dst, a.maskA, float32(b.Min.X+50), float32(b.Min.Y+20), 14, 0)
		blit(dst, a.maskA, float64(b.Min.X+50), float64(b.Min.Y+20))
	})

	a.drawPanel(screen, 0, 1, "ApplyBlur2", func(dst *ebiten.Image, b image.Rectangle) {
		a.renderer.SetColor(color.RGBA{170, 120, 255, 255})
		a.renderer.ApplyBlur2(dst, a.maskA, float32(b.Min.X+50), float32(b.Min.Y+20), 12, 0)
		blit(dst, a.maskA, float64(b.Min.X+50), float64(b.Min.Y+20))
	})

	a.drawPanel(screen, 1, 1, "ApplyShadow", func(dst *ebiten.Image, b image.Rectangle) {
		a.renderer.SetColor(color.RGBA{0, 0, 0, 180})
		a.renderer.ApplyShadow(dst, a.maskA, float32(b.Min.X+50), float32(b.Min.Y+20), 18, 14, 18, shapes.ClampBottom)
		blit(dst, a.spriteA, float64(b.Min.X+50), float64(b.Min.Y+20))
	})

	a.drawPanel(screen, 2, 1, "ApplyZoomShadow", func(dst *ebiten.Image, b image.Rectangle) {
		a.renderer.SetColor(color.RGBA{0, 0, 0, 180})
		a.renderer.ApplyZoomShadow(dst, a.maskA, float32(b.Min.X+50), float32(b.Min.Y+20), 12, 10, 1.3, shapes.ClampBottom)
		blit(dst, a.spriteA, float64(b.Min.X+50), float64(b.Min.Y+20))
	})

	a.drawPanel(screen, 3, 1, "ApplySimpleGlow / ApplyGlow", func(dst *ebiten.Image, b image.Rectangle) {
		a.renderer.SetColor(color.RGBA{255, 210, 84, 255})
		a.renderer.ApplySimpleGlow(dst, a.maskA, float32(b.Min.X+18), float32(b.Min.Y+20), 14)
		a.renderer.SetColor(color.RGBA{114, 217, 255, 255})
		a.renderer.ApplyGlow(dst, a.maskA, float32(b.Min.X+162), float32(b.Min.Y+20), 20, 10, 0.15, 0.9, 0)
		blit(dst, a.maskA, float64(b.Min.X+18), float64(b.Min.Y+20))
		blit(dst, a.maskA, float64(b.Min.X+162), float64(b.Min.Y+20))
	})

	a.drawPanel(screen, 0, 2, "ApplyHorzGlow / ApplyDarkHorzGlow", func(dst *ebiten.Image, b image.Rectangle) {
		a.renderer.SetColor(color.RGBA{83, 220, 171, 255})
		a.renderer.ApplyHorzGlow(dst, a.maskA, float32(b.Min.X+16), float32(b.Min.Y+22), 22, 0.1, 0.8, 0)
		a.renderer.SetColor(color.RGBA{255, 120, 120, 255})
		a.renderer.ApplyDarkHorzGlow(dst, a.maskA, float32(b.Min.X+168), float32(b.Min.Y+22), 22, 0.8, 0.1, 0)
		blit(dst, a.maskA, float64(b.Min.X+16), float64(b.Min.Y+22))
		blit(dst, a.maskA, float64(b.Min.X+168), float64(b.Min.Y+22))
	})

	a.drawPanel(screen, 1, 2, "ApplyScanlinesSharp", func(dst *ebiten.Image, b image.Rectangle) {
		area := ebiten.NewImage(264, 128)
		a.renderer.SimpleGradient(area, color.RGBA{60, 160, 255, 255}, color.RGBA{170, 110, 255, 255}, shapes.DirRadsTTB)
		a.renderer.ApplyScanlinesSharp(area, 2, 4, 0.55, float32(a.tick)*0.04)
		blitAt(dst, area, image.Pt(b.Min.X+28, b.Min.Y+22))
	})

	a.drawPanel(screen, 2, 2, "ApplyWaveLines", func(dst *ebiten.Image, b image.Rectangle) {
		area := ebiten.NewImage(264, 128)
		area.Fill(color.RGBA{12, 16, 24, 255})
		a.renderer.SetColor(color.RGBA{255, 210, 84, 255})
		a.renderer.ApplyWaveLines(area, 6, 0.2, 0.8, 6, float32(a.tick)*0.03, math.Pi/8)
		blitAt(dst, area, image.Pt(b.Min.X+28, b.Min.Y+22))
	})

	a.drawPanel(screen, 3, 2, "ApplyExpansionRect", func(dst *ebiten.Image, b image.Rectangle) {
		mask := ebiten.NewImage(160, 110)
		mask.Fill(color.Transparent)
		a.renderer.SetColor(color.White)
		a.renderer.DrawArea(mask, 20, 18, 120, 74, 0)
		a.renderer.SetColor(color.RGBA{114, 217, 255, 255})
		a.renderer.ApplyExpansionRect(dst, mask, float32(b.Min.X+78), float32(b.Min.Y+34), 10)
		blit(dst, mask, float64(b.Min.X+78), float64(b.Min.Y+34))
	})
}

func (a *Showcase) drawWarpPage(screen *ebiten.Image) {
	a.drawPanel(screen, 0, 0, "Scale", func(dst *ebiten.Image, b image.Rectangle) {
		a.renderer.Scale(dst, a.spriteB, float32(b.Min.X+42), float32(b.Min.Y+18), 0.75, false)
	})

	a.drawPanel(screen, 1, 0, "Scale (pixelated sampling)", func(dst *ebiten.Image, b image.Rectangle) {
		a.renderer.Scale(dst, a.spriteB, float32(b.Min.X+42), float32(b.Min.Y+18), 0.75, true)
	})

	a.drawPanel(screen, 2, 0, "MapQuad4", func(dst *ebiten.Image, b image.Rectangle) {
		quad := [4]shapes.PointF32{{float32(b.Min.X + 34), float32(b.Min.Y + 32)}, {float32(b.Min.X + 266), float32(b.Min.Y + 18)}, {float32(b.Min.X + 290), float32(b.Min.Y + 154)}, {float32(b.Min.X + 54), float32(b.Min.Y + 142)}}
		a.renderer.MapQuad4(dst, a.spriteA, quad)
	})

	a.drawPanel(screen, 3, 0, "MapProjective", func(dst *ebiten.Image, b image.Rectangle) {
		quad := [4]shapes.PointF32{{float32(b.Min.X + 56), float32(b.Min.Y + 26)}, {float32(b.Min.X + 260), float32(b.Min.Y + 46)}, {float32(b.Min.X + 280), float32(b.Min.Y + 142)}, {float32(b.Min.X + 26), float32(b.Min.Y + 156)}}
		a.renderer.MapProjective(dst, a.spriteA, quad)
	})

	a.drawPanel(screen, 0, 1, "WarpBarrel", func(dst *ebiten.Image, b image.Rectangle) {
		a.renderer.WarpBarrel(dst, a.spriteB, float32(b.Min.X+50), float32(b.Min.Y+20), 0.18, 0.12)
	})

	a.drawPanel(screen, 1, 1, "WarpArc", func(dst *ebiten.Image, b image.Rectangle) {
		a.renderer.WarpArc(dst, a.spriteB, float32(b.Min.X+160), float32(b.Min.Y+162), 120, math.Pi*0.9)
	})

	a.drawPanel(screen, 2, 1, "Map + overlay", func(dst *ebiten.Image, b image.Rectangle) {
		quad := [4]shapes.PointF32{{float32(b.Min.X + 40), float32(b.Min.Y + 30)}, {float32(b.Min.X + 274), float32(b.Min.Y + 22)}, {float32(b.Min.X + 266), float32(b.Min.Y + 154)}, {float32(b.Min.X + 46), float32(b.Min.Y + 146)}}
		a.renderer.MapProjective(dst, a.spriteB, quad)
		a.renderer.SetColor(color.RGBA{255, 255, 255, 140})
		a.renderer.DrawQuadSoft(dst, quad, 0, 2)
	})

	a.drawPanel(screen, 3, 1, "NewSimpleGradient", func(dst *ebiten.Image, b image.Rectangle) {
		grad := a.renderer.NewSimpleGradient(220, 140, color.RGBA{83, 220, 171, 255}, color.RGBA{255, 120, 120, 255}, shapes.DirRadsBLTR)
		blit(dst, grad, float64(b.Min.X+50), float64(b.Min.Y+16))
	})

	a.drawPanel(screen, 0, 2, "Generated asset A", func(dst *ebiten.Image, b image.Rectangle) {
		blit(dst, a.spriteA, float64(b.Min.X+50), float64(b.Min.Y+20))
	})

	a.drawPanel(screen, 1, 2, "Generated asset B", func(dst *ebiten.Image, b image.Rectangle) {
		blit(dst, a.spriteB, float64(b.Min.X+50), float64(b.Min.Y+20))
	})

	a.drawPanel(screen, 2, 2, "Clamping flags", func(dst *ebiten.Image, b image.Rectangle) {
		a.renderer.SetColor(color.RGBA{0, 0, 0, 180})
		a.renderer.ApplyShadow(dst, a.maskA, float32(b.Min.X+50), float32(b.Min.Y+20), 20, 16, 18, shapes.ClampBottom|shapes.ClampLeft)
		blit(dst, a.spriteA, float64(b.Min.X+50), float64(b.Min.Y+20))
		ebitenutil.DebugPrintAt(dst, "ClampBottom | ClampLeft", b.Min.X+18, b.Max.Y-panelTextLine1)
	})

	a.drawPanel(screen, 3, 2, "Shader-backed pipeline", func(dst *ebiten.Image, b image.Rectangle) {
		ebitenutil.DebugPrintAt(dst, "Most drawing/effects in this package", b.Min.X+18, b.Min.Y+52)
		ebitenutil.DebugPrintAt(dst, "are implemented with compact Kage shaders", b.Min.X+18, b.Min.Y+70)
		ebitenutil.DebugPrintAt(dst, "instead of triangle rasterization alone.", b.Min.X+18, b.Min.Y+88)
	})
}

func (a *Showcase) drawJFMPage(screen *ebiten.Image) {
	a.drawPanel(screen, 0, 0, "JFMapFill + JFMHeat", func(dst *ebiten.Image, b image.Rectangle) {
		jfmap := ebiten.NewImage(a.jfMask.Bounds().Dx(), a.jfMask.Bounds().Dy())
		a.renderer.JFMapFill(jfmap, a.jfMask, 96, 0.001, 1)
		a.renderer.JFMHeat(dst, jfmap, float32(b.Min.X+50), float32(b.Min.Y+20), 96)
	})

	a.drawPanel(screen, 1, 0, "JFMExpand", func(dst *ebiten.Image, b image.Rectangle) {
		a.renderer.SetColor(color.RGBA{255, 194, 84, 255})
		a.renderer.JFMExpand(dst, a.jfMask, nil, float32(b.Min.X+50), float32(b.Min.Y+20), 12, shapes.AAMargin*4)
		blit(dst, a.jfMask, float64(b.Min.X+50), float64(b.Min.Y+20))
	})

	a.drawPanel(screen, 2, 0, "JFMErode", func(dst *ebiten.Image, b image.Rectangle) {
		a.renderer.SetColor(color.RGBA{114, 217, 255, 255})
		a.renderer.JFMErode(dst, a.jfMask, nil, float32(b.Min.X+50), float32(b.Min.Y+20), 10, shapes.AAMargin)
	})

	a.drawPanel(screen, 3, 0, "JFMapBoundary", func(dst *ebiten.Image, b image.Rectangle) {
		jfmap := ebiten.NewImage(a.jfMask.Bounds().Dx(), a.jfMask.Bounds().Dy())
		a.renderer.JFMapBoundary(jfmap, a.jfMask, 96, 0.001, 1, false, false)
		blit(dst, jfmap, float64(b.Min.X+50), float64(b.Min.Y+20))
	})

	a.drawPanel(screen, 0, 1, "UnsafeTemp", func(dst *ebiten.Image, b image.Rectangle) {
		tmp := a.renderer.UnsafeTemp(2, 220, 140, true)
		tmp.Fill(color.RGBA{22, 30, 44, 255})
		a.renderer.SetColor(color.RGBA{83, 220, 171, 255})
		a.renderer.DrawHexagon(tmp, 110, 70, 44, 12, 0.2)
		blit(dst, tmp, float64(b.Min.X+50), float64(b.Min.Y+20))
	})

	a.drawPanel(screen, 1, 1, "UnsafeTempCopy", func(dst *ebiten.Image, b image.Rectangle) {
		tmp := a.renderer.UnsafeTempCopy(3, a.spriteA, false)
		blit(dst, tmp, float64(b.Min.X+50), float64(b.Min.Y+20))
	})

	a.drawPanel(screen, 2, 1, "UnsafeTempDual", func(dst *ebiten.Image, b image.Rectangle) {
		left := a.renderer.UnsafeTempCopy(4, a.spriteA, false)
		right := a.renderer.UnsafeTemp(5, a.spriteA.Bounds().Dx(), a.spriteA.Bounds().Dy(), true)
		a.renderer.SetColor(color.RGBA{255, 120, 120, 255})
		a.renderer.ApplyOutline(right, a.maskA, 0, 0, 6)
		blit(dst, left, float64(b.Min.X+18), float64(b.Min.Y+20))
		blit(dst, right, float64(b.Min.X+154), float64(b.Min.Y+20))
	})

	a.drawPanel(screen, 3, 1, "Renderer color state", func(dst *ebiten.Image, b image.Rectangle) {
		c := a.renderer.GetColorF32()
		ebitenutil.DebugPrintAt(dst, fmt.Sprintf("current RGBA = %.2f %.2f %.2f %.2f", c[0], c[1], c[2], c[3]), b.Min.X+18, b.Min.Y+56)
		ebitenutil.DebugPrintAt(dst, "SetColor / SetBlend drive most samples.", b.Min.X+18, b.Min.Y+74)
	})

	a.drawPanel(screen, 0, 2, "Mask source for JFM", func(dst *ebiten.Image, b image.Rectangle) {
		blit(dst, a.jfMask, float64(b.Min.X+50), float64(b.Min.Y+20))
	})

	a.drawPanel(screen, 1, 2, "ApplyVertBlur", func(dst *ebiten.Image, b image.Rectangle) {
		a.renderer.SetColor(color.RGBA{170, 120, 255, 255})
		a.renderer.ApplyVertBlur(dst, a.maskA, float32(b.Min.X+50), float32(b.Min.Y+20), 14, 0)
		blit(dst, a.maskA, float64(b.Min.X+50), float64(b.Min.Y+20))
	})

	a.drawPanel(screen, 2, 2, "ApplyHorzBlur", func(dst *ebiten.Image, b image.Rectangle) {
		a.renderer.SetColor(color.RGBA{83, 220, 171, 255})
		a.renderer.ApplyHorzBlur(dst, a.maskA, float32(b.Min.X+50), float32(b.Min.Y+20), 14, 0)
		blit(dst, a.maskA, float64(b.Min.X+50), float64(b.Min.Y+20))
	})

	a.drawPanel(screen, 3, 2, "Coverage note", func(dst *ebiten.Image, b image.Rectangle) {
		ebitenutil.DebugPrintAt(dst, "Included all public drawing, mask,", b.Min.X+18, b.Min.Y+52)
		ebitenutil.DebugPrintAt(dst, "tile, warp, noise and stable JFM APIs.", b.Min.X+18, b.Min.Y+70)
		ebitenutil.DebugPrintAt(dst, "Skipped JFMOutline/JFMInsetContour because", b.Min.X+18, b.Min.Y+88)
		ebitenutil.DebugPrintAt(dst, "they currently panic as unimplemented.", b.Min.X+18, b.Min.Y+106)
	})
}
