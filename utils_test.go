package shapes

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"math/rand/v2"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// color that's dark but distinguishable against BlendClears used to
// test bounds
var backTestColor = color.RGBA{48, 0, 128, 255}

func setTestMultiColors(renderer *Renderer) {
	renderer.SetColorF32(1.0, 0.3, 0.3, 1.0, 0)
	renderer.SetColorF32(0.3, 1.0, 0.75, 1.0, 1)
	renderer.SetColorF32(0.5, 0.3, 1.0, 1.0, 2)
	renderer.SetColorF32(0.75, 1.0, 0.3, 1.0, 3)
}

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
	Renderer     *Renderer
	Images       []*ebiten.Image
	Ticks        uint64
	LeftClick    image.Point
	RightClick   image.Point
	SpacePressed bool

	slomo uint8
}

func (ctx *TestAppCtx) LeftClickF32() PointF32 {
	return PtF32(ctx.LeftClick.X, ctx.LeftClick.Y)
}
func (ctx *TestAppCtx) RightClickF32() PointF32 {
	return PtF32(ctx.RightClick.X, ctx.RightClick.Y)
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

func updateToggle(ctx TestAppCtx, key ebiten.Key, value bool) bool {
	if inpututil.IsKeyJustPressed(key) {
		return !value
	}
	return value
}

func updateParam[T float32 | float64 | ~int | ~uint8 | ~int8](ctx TestAppCtx, key ebiten.Key, value, minValue, maxValue, delta T) T {
	if !inpututil.IsKeyJustPressed(key) {
		return value
	}

	if ebiten.IsKeyPressed(ebiten.KeyShift) {
		value -= delta
	} else {
		value += delta
	}
	value = wrap(value, minValue, maxValue)
	return value
}

func updateParamMult[T float32 | float64 | ~int](ctx TestAppCtx, key ebiten.Key, value, minValue, maxValue, factor T) T {
	if !inpututil.IsKeyJustPressed(key) {
		return value
	}

	if ebiten.IsKeyPressed(ebiten.KeyShift) {
		value /= factor
	} else {
		value *= factor
	}
	value = wrap(value, minValue, maxValue)
	return value
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
	updater func(ctx TestAppCtx)
	drawer  func(canvas *ebiten.Image, ctx TestAppCtx)

	origin    image.Point
	offscreen *ebiten.Image
}

func NewTestApp(updater func(TestAppCtx), drawer func(canvas *ebiten.Image, ctx TestAppCtx), images ...*ebiten.Image) *TestApp {
	var app TestApp
	app.Images = images
	app.LeftClick = image.Pt((128*4)/3, 128)
	app.RightClick = image.Pt((320*4)/3, 320)
	app.Renderer = NewRenderer()
	app.Renderer.Warnings.SetHandler(NewWarningPanicHandler())
	app.updater = updater
	app.drawer = drawer
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	return &app
}

func (app *TestApp) Update() error {
	if !ebiten.IsKeyPressed(ebiten.KeyPeriod) {
		if ebiten.IsKeyPressed(ebiten.KeyComma) {
			app.slomo += 1
			if app.slomo > 4 {
				app.slomo = 0
				app.Ticks += 1
			}
		} else {
			app.Ticks += 1
		}
	}

	app.SpacePressed = ebiten.IsKeyPressed(ebiten.KeySpace)
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

	if inpututil.IsKeyJustPressed(ebiten.KeyDigit0) {
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
	if err := app.BaseTestApp.Update(); err != nil {
		return err
	}
	app.updater(app.TestAppCtx)
	return nil
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
}

type flagList []Flag

func (l *flagList) Has(f Flag) bool {
	for _, flag := range *l {
		if f == flag {
			return true
		}
	}
	return false
}

func (l *flagList) Flip(f Flag) {
	if !l.Unset(f) {
		l.Set(f)
	}
}

func (l *flagList) Unset(f Flag) bool {
	s := *l
	for i, flag := range s {
		if flag == f {
			s[len(s)-1], s[i] = s[i], s[len(s)-1]
			s = s[:len(s)-1]
			*l = s
			return true
		}
	}
	return false
}

func (l *flagList) Set(f Flag) {
	if !l.Has(f) {
		*l = append(*l, f)
	}
}

func (l *flagList) UpdateFlag(f Flag, key ebiten.Key) {
	if inpututil.IsKeyJustPressed(key) {
		l.Flip(f)
	}
}

func wrap[Float ~float32 | ~float64 | ~int | ~uint8 | ~int8](x, lo, hi Float) Float {
	if x < lo {
		return hi
	} else if x > hi {
		return lo
	}
	return x
}
