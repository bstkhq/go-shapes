package main

// Based on the original implementation by mcuadros on 7bbef1b607a53790e6bf8b2b350f54de6a4e57db

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"math"
	"slices"

	"github.com/bstkhq/go-shapes"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

const ScreenWidth, ScreenHeight = 1408, 896

const (
	InfoLeft           = "Use right/left arrows to navigate pages\nClick on panels to play with parameters"
	InfoCenter         = "< Page %d/%d >"
	InfoRightUnfocused = "Fullscreen [F11]\nExit [ESC]\n"
	InfoRightFocused   = "Fullscreen [F11]\nClose Panel [ESC]"
)

var (
	ColorBackground = color.RGBA{0, 0, 0, 255}
	ColorPanel      = color.RGBA{22, 28, 42, 255}
	ColorHoverHalo  = color.RGBA{116, 116, 116, 116}
	ColorText       = color.RGBA{255, 255, 255, 255}
	ColorTextMuted  = color.RGBA{144, 144, 144, 255}
)

func main() {
	ebiten.SetWindowTitle("go-shapes/examples/showcase")
	ebiten.SetWindowSize(ScreenWidth, ScreenHeight)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	if err := ebiten.RunGame(NewShowcase()); err != nil {
		log.Fatal(err)
	}
}

func areaSize(area [2]shapes.PointF32) shapes.PointF32 {
	return area[1].Sub(area[0])
}

func areaCenter(area [2]shapes.PointF32) shapes.PointF32 {
	return area[0].Add(areaSize(area).Scale(0.5))
}

func ptInArea(pt shapes.PointF32, area [2]shapes.PointF32) bool {
	return (pt.X >= area[0].X && pt.X <= area[1].X) && (pt.Y >= area[0].Y && pt.Y <= area[1].Y)
}

func padArea(area [2]shapes.PointF32, pad shapes.PointF32) [2]shapes.PointF32 {
	area[0] = area[0].Add(pad)
	area[1] = area[1].Sub(pad)
	return area
}

func coordsInArea(pt shapes.PointF32, area [2]shapes.PointF32) shapes.PointF32 {
	return pt.Sub(area[0]).Div(area[1].Sub(area[0]))
}

func fluidScale(current, baseFrom, baseTo, growth float32) float32 {
	ratio := current / baseFrom
	exponent := math.Log2(float64((1 + growth)))
	multiplier := math.Pow(float64(ratio), exponent)
	return baseTo * float32(multiplier)
}

func clip(img *ebiten.Image, area [2]shapes.PointF32) *ebiten.Image {
	intCeil := func(v float32) int {
		return int(math.Ceil(float64(v)))
	}
	clipRect := image.Rect(int(area[0].X), int(area[0].Y), intCeil(area[1].X), intCeil(area[1].Y))
	return img.SubImage(clipRect).(*ebiten.Image)
}

type Showcase struct {
	Pages    []Page
	Renderer *shapes.Renderer
	index    int
}

func NewShowcase() *Showcase {
	show := Showcase{
		Renderer: shapes.NewRenderer(),
		Pages: []Page{
			pageShapesPoly(),
			// pageShapesCirc(),
			// pageBlurAndGlows(),
			// pageColor(),
			// pageProjectAndWarp(),
			// pageTexturing(), // (Tiling, Noise, Masking, WaveLines)
			// pageMorphology(),
			// pageUtility(), // text, scale, draw shader
		},
	}
	show.Reset()
	return &show
}

func (s *Showcase) Layout(int, int) (int, int) {
	panic("using LayoutF")
}

func (s *Showcase) LayoutF(w, h float64) (float64, float64) {
	scale := ebiten.Monitor().DeviceScaleFactor()
	return w * scale, h * scale
}

func (s *Showcase) Reset() {
	s.index = 0
	for i := range s.Pages {
		s.Pages[i].Reset()
	}
}

func (s *Showcase) Update() error {
	switch {
	case inpututil.IsKeyJustPressed(ebiten.KeyArrowLeft):
		s.Pages[s.index].Reset()
		s.index = (s.index - 1)
		if s.index < 0 {
			s.index = len(s.Pages) - 1
		}
	case inpututil.IsKeyJustPressed(ebiten.KeyArrowRight):
		s.Pages[s.index].Reset()
		s.index = (s.index + 1) % len(s.Pages)
	case inpututil.IsKeyJustPressed(ebiten.KeyF11):
		ebiten.SetFullscreen(!ebiten.IsFullscreen())
	}

	if err := s.Pages[s.index].Update(); err != nil {
		return err
	}

	return nil
}

func (s *Showcase) Draw(canvas *ebiten.Image) {
	const TitlePanelsRatio = 0.2
	subCanvas, region := s.getCanvasRegion(canvas)
	subCanvas.Fill(ColorBackground)

	headerRegion, panelsRegion := region, region
	height := region[1].Y - region[0].Y
	headerRegion[1].Y = headerRegion[0].Y + height*TitlePanelsRatio
	panelsRegion[0].Y = headerRegion[1].Y

	s.drawHeader(subCanvas, headerRegion)
	s.Pages[s.index].DrawPanels(subCanvas, panelsRegion, s.Renderer)
}

func (s *Showcase) drawHeader(target *ebiten.Image, area [2]shapes.PointF32) {
	const TitleScale = 0.24
	const TitleInterspacing = 0.04
	const InfoScale = 0.11
	const Margin = 0.07

	size := area[1].Sub(area[0])
	margin := size.Y * Margin

	// left info
	s.Renderer.SetColor(ColorTextMuted)
	txtOpts := shapes.TextOpts(shapes.Font().ScaleOf(size.Y*InfoScale), shapes.TopLeft)
	txtOpts.Align = shapes.TopLeft.Snap(shapes.CapLine)
	s.Renderer.Text(target, InfoLeft, area[0].X+margin, area[0].Y+margin, txtOpts)

	// right info
	txtOpts.Align = shapes.TopRight.Snap(shapes.CapLine)
	infoRight := InfoRightUnfocused
	if s.Pages[s.index].HasFocusedPanel() {
		infoRight = InfoRightFocused
	}
	s.Renderer.Text(target, infoRight, area[1].X-margin, area[0].Y+margin, txtOpts)

	// center title and info
	titleOpts := shapes.TextOpts(shapes.Font().ScaleOf(size.Y*TitleScale), shapes.TopCenter)
	title := s.Pages[s.index].Title
	_, titleHeight := s.Renderer.TextSize(title, titleOpts)
	_, subHeight := s.Renderer.TextSize(InfoCenter, txtOpts)
	comboHeight := titleHeight + subHeight + size.Y*TitleInterspacing

	ctr := area[0].Add(size.Scale(0.5))
	txtOpts.Align = shapes.BottomCenter
	s.Renderer.Text(target, fmt.Sprintf(InfoCenter, s.index+1, len(s.Pages)), ctr.X, ctr.Y+comboHeight/2.0, txtOpts)

	s.Renderer.SetColor(ColorText)
	s.Renderer.Text(target, title, ctr.X, ctr.Y-comboHeight/2.0, titleOpts)
}

// getCanvasRegion returns a subregion of canvas that preserves the showcase
// intended aspect ratio. the returned points are the precise (min, max) area
func (s *Showcase) getCanvasRegion(canvas *ebiten.Image) (*ebiten.Image, [2]shapes.PointF32) {
	bounds := canvas.Bounds()
	cw, ch := bounds.Dx(), bounds.Dy()
	canvasSize := shapes.PtF32(cw, ch)
	aspectRatio := canvasSize.X / canvasSize.Y
	targetAspectRatio := float32(ScreenWidth) / float32(ScreenHeight)

	var offset shapes.PointF32
	if aspectRatio < targetAspectRatio {
		offset.Y = (canvasSize.Y - canvasSize.X/targetAspectRatio) / 2.0
	} else {
		offset.X = (canvasSize.X - canvasSize.Y*targetAspectRatio) / 2.0
	}

	area := padArea([2]shapes.PointF32{shapes.PtF32(0, 0), canvasSize}, offset)
	return clip(canvas, area), area
}

type Page struct {
	Title  string
	Panels []Panel
	focus  int
}

func (p *Page) Reset() {
	p.focus = -1
	for i := range p.Panels {
		p.Panels[i].Reset()
	}
}

func (p *Page) HasFocusedPanel() bool {
	return p.focus >= 0
}

func (p *Page) Update() error {
	escape := inpututil.IsKeyJustPressed(ebiten.KeyEscape)
	back := escape || inpututil.IsKeyJustPressed(ebiten.KeyBackspace) || inpututil.IsKeyJustPressed(ebiten.KeyEnter)

	hovering := false
	click := inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft)
	if p.HasFocusedPanel() {
		p.Panels[p.focus].Update(true)
		if back || (click && !p.Panels[p.focus].Hovered) {
			p.focus = -1
		}
	} else {
		for i := range p.Panels {
			p.Panels[i].Update(false)
			if p.Panels[i].Hovered {
				hovering = true
				if click {
					p.focus = i
					p.Panels[i].Hovered = false
				}
			}
		}

		if escape {
			return ebiten.Termination
		}
	}

	if hovering {
		ebiten.SetCursorShape(ebiten.CursorShapePointer)
	} else {
		ebiten.SetCursorShape(ebiten.CursorShapeDefault)
	}

	return nil
}

func (p *Page) DrawPanels(target *ebiten.Image, area [2]shapes.PointF32, renderer *shapes.Renderer) {
	const Interspacing = 0.08

	bounds := target.Bounds()
	itsp := float32(min(bounds.Dx(), bounds.Dy())) * Interspacing

	offset := shapes.PtF32(itsp, itsp).Scale(0.5)
	if p.HasFocusedPanel() {
		p.Panels[p.focus].SetArea(padArea(area, offset))
		p.Panels[p.focus].Draw(target, renderer)
	} else {
		size := areaSize(area)
		panelSize := shapes.PtF32((size.X-5.0*itsp)/4.0, (size.Y-3.0*itsp)/3.0)
		panelDelta := panelSize.Add(offset)
		baseCoords := area[0].Add(offset)
		for i := range p.Panels {
			cell := shapes.PtF32(i%4, i/3)
			coords := baseCoords.Add(panelDelta.Mul(cell))
			p.Panels[i].SetArea([2]shapes.PointF32{coords, coords.Add(panelSize)})
			p.Panels[i].Draw(target, renderer)
		}
	}
}

type Panel struct {
	Title      string
	Renderer   func(*ebiten.Image, RenderContext)
	Parameters []Parameter

	liveParams map[string]*Parameter
	foreground bool
	area       [2]shapes.PointF32 // origin, end

	Hovered bool
}

func (p *Panel) Reset() {
	p.foreground = false
	clear(p.liveParams)
	if p.liveParams == nil {
		p.liveParams = make(map[string]*Parameter)
	}

	for i := range p.Parameters {
		newParam := p.Parameters[i].Clone()
		p.liveParams[newParam.Name] = &newParam
	}
}

func (p *Panel) SetArea(area [2]shapes.PointF32) {
	p.area = area // origin, end
}

func (p *Panel) Update(foreground bool) {
	p.foreground = foreground
	for _, param := range p.liveParams {
		param.Update(foreground)
	}

	x, y := ebiten.CursorPosition()
	p.Hovered = ptInArea(shapes.PtF32(x, y), p.area)
}

type RenderContext struct {
	Area       [2]shapes.PointF32 // origin, end
	Renderer   *shapes.Renderer
	Parameters map[string]*Parameter
}

func (p *Panel) Draw(target *ebiten.Image, renderer *shapes.Renderer) {
	const ParamsRatio = 0.4
	const ParamsInterspacing = 0.02
	const HighlightRadius = 0.225

	// surface
	size := areaSize(p.area)
	rounding := fluidScale(size.Y, 100, 8, 0.75)
	if p.Hovered && !p.foreground {
		renderer.SetColor(ColorHoverHalo)
		renderer.FillRectSoft(target, p.area[0].X, p.area[0].Y, size.X, size.Y, 0, size.Y*HighlightRadius)
	}
	renderer.SetColor(ColorPanel)
	renderer.FillRect(target, p.area[0].X, p.area[0].Y, size.X, size.Y, -rounding)

	// title
	margin := fluidScale(size.Y, 100, 8, 0.5)
	workArea := padArea(p.area, shapes.PtF32(margin, margin))
	workSize := areaSize(workArea)
	titleArea := workArea
	titleArea[1].Y = workArea[0].Y + fluidScale(workSize.Y, 100, 16, 0.5)
	p.drawTitle(target, renderer, titleArea)
	workArea[0].Y = titleArea[1].Y

	// params
	// ...

	// effect
	ctx := RenderContext{
		Area:       workArea,
		Renderer:   renderer,
		Parameters: p.liveParams,
	}
	p.Renderer(target, ctx)
}

func (p *Panel) drawTitle(target *ebiten.Image, renderer *shapes.Renderer, titleArea [2]shapes.PointF32) {
	renderer.SetColor(ColorText)
	scale := shapes.Font().ScaleOf(areaSize(titleArea).Y * 1.15)
	opts := shapes.TextOpts(scale, shapes.CenterLeft)
	renderer.Text(target, p.Title, titleArea[0].X, titleArea[0].Y+areaSize(titleArea).Y/2, opts)
}

type Parameter struct {
	Name string
	Hint string

	Value   float64
	Anim    float64
	Osc     float64 // 0...2*math.Pi
	Current float64

	Fmt func(float64) string
	Min float64
	Max float64

	OptionsOn []bool
	Options   []string
	MultiOpt  bool

	area    [2]shapes.PointF32
	hovered bool
}

func NewOptions(name, hint string, multi bool, opts ...string) Parameter {
	return Parameter{Name: name, Hint: hint, MultiOpt: multi, Options: opts, OptionsOn: make([]bool, len(opts))}
}

func NewSlider(name, hint string, min, max float64, fmt func(float64) string) Parameter {
	return Parameter{Name: name, Hint: hint, Min: min, Max: max, Fmt: fmt}
}

func (p Parameter) Clone() Parameter {
	o := p
	o.OptionsOn = slices.Clone(p.OptionsOn)
	o.Options = slices.Clone(p.Options)
	return o
}

func (p Parameter) AtValue(value, anim float64) Parameter {
	p.Value = value
	p.Anim = anim
	return p
}

func (p Parameter) AtOptions(opts ...string) Parameter {
	on := make(map[string]struct{})
	for _, opt := range opts {
		on[opt] = struct{}{}
	}
	for i, opt := range p.Options {
		_, present := on[opt]
		p.OptionsOn[i] = present
	}
	return p
}

func (p *Parameter) Update(foreground bool) {
	if p.Anim > 0 {
		p.Osc += (0.001 / p.Anim)
		p.Current = p.Value + math.Sin(p.Osc)*p.Anim
		p.Current = min(max(p.Current, p.Min), p.Max)
	} else {
		p.Current = p.Value
	}

	if !foreground {
		return
	}

	x, y := ebiten.CursorPosition()
	cursor := shapes.PtF32(x, y)
	p.hovered = ptInArea(cursor, p.area)
	// ...
	coordsInArea(cursor, p.area)
}

func (p *Parameter) Draw(target *ebiten.Image, area [2]shapes.PointF32) {
	p.area = area
	switch {
	case len(p.Options) > 0:
		p.drawOptions(target)
	default:
		p.drawSlider(target)
	}
}

func (p *Parameter) drawOptions(target *ebiten.Image) {
	// off/on quads, same as options
}

func (p *Parameter) drawSlider(target *ebiten.Image) {}

// ---- pages ----

func pageShapesPoly() Page {
	return Page{
		Title: "Polygonal Shapes",
		Panels: []Panel{
			{Title: "StrokeLine", Renderer: renderShapesPolyLine, Parameters: paramsShapesPolyLine()},
			{Title: "StrokeLine2", Renderer: renderShapesPolyLine, Parameters: paramsShapesPolyLine()},
		},
	}
}

// ---- render functions ----

func paramsShapesPolyLine() []Parameter {
	return []Parameter{
		NewOptions("Flags", "", true, shapes.ColorAABB.String()),
		NewSlider("Value", "", 0.0, 1.0, nil),
	}
}

func renderShapesPolyLine(target *ebiten.Image, ctx RenderContext) {
	ctr, size := areaCenter(ctx.Area), areaSize(ctx.Area)
	th := size.Y * 0.1
	o := ctr.Sub(size.Scale(0.4))
	f := ctr.Add(size.Scale(0.4))
	ctx.Renderer.SetColorF32(1, 1, 1, 1)
	ctx.Renderer.StrokeLine(target, o, f, th)
}
