package shapes

import (
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"
)

// Margins are used in basic shader operations like [Renderer.DrawRectShader]
// and [Renderer.DrawImgShader]().
type Margins struct {
	Left   float32
	Right  float32
	Top    float32
	Bottom float32
}

var NoMargins = Margins{}

// NewMargins returns symmetrical margins with Left=horz, Right=horz,
// Top=vert, Bottom=vert.
func NewMargins(horz, vert float32) Margins {
	return Margins{
		Left:   horz,
		Right:  horz,
		Top:    vert,
		Bottom: vert,
	}
}

// Downscaling is a common technique used in graphics where an effect is not applied at
// full resolution, but on a smaller offscreen. This is done to reduce the amount of pixels
// to process, but it has some downsides:
//   - Non-trivial overhead due to the extra steps (downscale + effect + upscale back), which
//     can be worse than paying for a full resolution operation if processing many small images.
//   - The upscaled effects can look blocky. Different effects respond differently to downscaling,
//     so each case needs to be assessed independently.
type Downscaling uint8

const (
	DownscaleNone Downscaling = iota
	DownscaleX2
	DownscaleX4
	DownscaleX8
	DownscaleX16
)

func (d Downscaling) valid() bool {
	return d >= DownscaleNone && d <= DownscaleX16
}

// Factor returns 1 for [DownscaleNone], 2 for [DownscaleX2], 4 for [DownscaleX4], etc.
func (d Downscaling) Factor() int {
	return int(1 << d)
}

// Origin is a helper type for adjusting draw coordinates based on a
// source image origin anchor. See [Origin.Adjust] for more details.
type Origin PointF32

// Predefined origin variables. See [Origin.Adjust]() for usage examples.
var (
	TL  = Origin{X: 0, Y: 0}     // top-left
	TC  = Origin{X: 0.5, Y: 0}   // top-center
	TR  = Origin{X: 1, Y: 0}     // top-right
	CL  = Origin{X: 0, Y: 0.5}   // center-left
	CTR = Origin{X: 0.5, Y: 0.5} // center
	CR  = Origin{X: 1, Y: 0.5}   // center-right
	BL  = Origin{X: 0, Y: 1}     // bottom-left
	BC  = Origin{X: 0.5, Y: 1}   // bottom-center
	BR  = Origin{X: 1, Y: 1}     // bottom-right
)

// Adjust translates the given reference position from TL (top-left)
// origin to o.
//
// For example, if we want to draw the bottom-right corner of an image
// at X=60, Y=40, we adjust the coordinates like this:
//
//	x, y := shapes.BR.Adjust(src, 60, 40)
//
// See also [TC], [TR], [CTR] and others for more predefined constants.
func (o Origin) Adjust(source *ebiten.Image, x, y float32) (float32, float32) {
	if source == nil {
		return 0, 0
	}
	w, h := rectSizeF32(source.Bounds())
	return x - w*o.X, y - h*o.Y
}

// Flags for operations like [Renderer.DrawAt](). See [Bilinear], [Dithered], etc.
type Flag int

// Operation flags, read the descriptions for details.
const (
	// noFlag is ignored by any function accepting flags.
	noFlag Flag = iota

	// Use bilinear instead of nearest filtering.
	Bilinear

	// Apply dithering.
	Dithered

	// sentinel
	numFlags
)

func (f Flag) String() string {
	switch f {
	case Bilinear:
		return "Bilinear"
	case Dithered:
		return "Dithered"
	default:
		return "Flag#" + strconv.Itoa(int(f))
	}
}

func mapBool[T any](b bool, falseValue, trueValue T) T {
	if b {
		return trueValue
	}
	return falseValue
}
