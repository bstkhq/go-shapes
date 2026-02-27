package shapes

import (
	"fmt"
	"image"
	"math"
	"math/rand/v2"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type BaseTestApp struct{}

func (*BaseTestApp) Layout(w, h int) (int, int) {
	return w, h
}
func (*BaseTestApp) Update() error {
	if ebiten.IsKeyPressed(ebiten.KeyEscape) {
		return ebiten.Termination
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyF11) {
		ebiten.SetFullscreen(!ebiten.IsFullscreen())
	}
	return nil
}
func (*BaseTestApp) Draw(*ebiten.Image) {}

type TestAppCtx struct {
	Renderer   *Renderer
	Images     []*ebiten.Image
	Ticks      uint64
	LeftClick  image.Point
	RightClick image.Point
	NewInput   bool // if using inpututil.IsKeyJustPressed inside draw, check for this too
}

func (ctx *TestAppCtx) LeftClickF32() (x, y float32) {
	return float32(ctx.LeftClick.X), float32(ctx.LeftClick.Y)
}
func (ctx *TestAppCtx) RightClickF32() (x, y float32) {
	return float32(ctx.RightClick.X), float32(ctx.RightClick.Y)
}
func (ctx *TestAppCtx) LeftClickF64() (x, y float64) {
	return float64(ctx.LeftClick.X), float64(ctx.LeftClick.Y)
}
func (ctx *TestAppCtx) RightClickF64() (x, y float64) {
	return float64(ctx.RightClick.X), float64(ctx.RightClick.Y)
}
func (ctx *TestAppCtx) RadsAnim(speedFactor float64) float64 {
	return math.Pi * math.Sin(float64(ctx.Ticks)*0.01*speedFactor)
}
func (ctx *TestAppCtx) ModAnim(maxValue, speedFactor float64) float64 {
	return math.Mod((float64(ctx.Ticks) * 0.02 * speedFactor), maxValue)
}
func (ctx *TestAppCtx) DistAnim(maxDist, speedFactor float64) float64 {
	value := maxDist * (math.Sin(float64(ctx.Ticks)*0.02*speedFactor) + 1.0) / 2.0
	return snapEdges(value, 0.0, maxDist, 0.001)
}
func (ctx *TestAppCtx) Title() string {
	return fmt.Sprintf(
		"Clicks L(%d, %d) R(%d, %d) [%.02f FPS]",
		ctx.LeftClick.X, ctx.LeftClick.Y, ctx.RightClick.X, ctx.RightClick.Y, ebiten.ActualFPS(),
	)
}

func (ctx *TestAppCtx) DrawAtF32(target, image *ebiten.Image, ox, oy float32) {
	var opts ebiten.DrawImageOptions
	opts.GeoM.Translate(float64(ox), float64(oy))
	target.DrawImage(image, &opts)
}

func (ctx *TestAppCtx) DrawWithAlphaAtF32(target, image *ebiten.Image, alpha, ox, oy float32) {
	var opts ebiten.DrawImageOptions
	opts.GeoM.Translate(float64(ox), float64(oy))
	opts.ColorScale.ScaleAlpha(alpha)
	target.DrawImage(image, &opts)
}

type TestApp struct {
	BaseTestApp
	TestAppCtx
	drawer func(canvas *ebiten.Image, ctx TestAppCtx)

	origin    image.Point
	offscreen *ebiten.Image
}

func NewTestApp(drawer func(canvas *ebiten.Image, ctx TestAppCtx), images ...*ebiten.Image) *TestApp {
	var app TestApp
	app.Images = images
	app.LeftClick = image.Pt((128*4)/3, 128)
	app.RightClick = image.Pt((320*4)/3, 320)
	app.Renderer = NewRenderer()
	app.Renderer.Warnings.SetHandler(NewWarningPanicHandler())
	app.drawer = drawer
	return &app
}

func (app *TestApp) Update() error {
	app.NewInput = true
	app.Ticks += 1
	left := ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)
	right := ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight)
	if left || right {
		x, y := ebiten.CursorPosition()
		if left {
			app.LeftClick = image.Pt(x, y)
		} else {
			app.RightClick = image.Pt(x, y)
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyO) {
		if app.origin.Eq(image.Pt(0, 0)) {
			offset := 32
			if rand.Float64() < 0.5 {
				offset = -offset
			}

			app.origin = image.Pt(offset, offset)
		} else {
			app.origin = image.Pt(0, 0)
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyV) {
		ebiten.SetVsyncEnabled(!ebiten.IsVsyncEnabled())
	}

	ebiten.SetWindowTitle(app.Title())
	return app.BaseTestApp.Update()
}

func (app *TestApp) Draw(canvas *ebiten.Image) {
	if !app.origin.Eq(image.Pt(0, 0)) {
		shiftedBounds := canvas.Bounds().Add(app.origin)
		if app.offscreen == nil || !app.offscreen.Bounds().Eq(shiftedBounds) {
			app.offscreen = ebiten.NewImageWithOptions(shiftedBounds, nil)
		}
		app.offscreen.Clear()
		app.drawer(app.offscreen, app.TestAppCtx)
		canvas.DrawImage(app.offscreen, nil)
	} else {
		app.drawer(canvas, app.TestAppCtx)
	}
	app.NewInput = false // input will be stale if calling draw again (e.g. high refresh rate displays)
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
