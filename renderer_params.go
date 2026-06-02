package shapes

import (
	"image"
	"strconv"
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

type Bounded interface {
	Bounds() image.Rectangle
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

// Adjust translates the given reference position from TL (top-left) origin to o. Common
// argument types are [*ebiten.Image], [image.Rectangle] and [image.Image].
//
// For example, if we want to draw the bottom-right corner of an image at X=60, Y=40, we
// adjust the coordinates like this:
//
//	x, y := shapes.BR.Adjust(src, shapes.PtF32(60, 40))
//
// See also [TC], [TR], [CTR] and others for more predefined constants.
func (o Origin) Adjust(bounded Bounded, xy PointF32) PointF32 {
	return o.AdjustXY(bounded, xy.X, xy.Y)
}

// AdjustXY is an alternative form of [Origin.Adjust]() that accepts individual
// coordinates instead of a [PointF32].
func (o Origin) AdjustXY(bounded Bounded, x, y float32) PointF32 {
	if bounded == nil {
		return PointF32{}
	}
	w, h := rectSizeF32(bounded.Bounds())
	return PtF32(x-w*o.X, y-h*o.Y)
}

// Flags for operations like [Renderer.DrawAt](). See [Bilinear], [Dithered], etc.
type Flag int

// Operation flags, read the descriptions for details.
const (
	// noFlag is ignored by any function accepting flags.
	noFlag Flag = iota

	// Use bilinear instead of nearest filtering.
	//
	// Bilinear filtering is required when rendering images and effects at
	// non-integer coordinates. This is particularly visible when attempting
	// to smoothly animate image coordinates.
	//
	// Bilinear filtering is expensive, typically requiring at least 4 samples
	// per fragment instead of 1.
	Bilinear

	// Apply dithering.
	//
	// Dithering is critical in slow alpha and color transitions, where large
	// areas of the output share the same color and even small changes become
	// distinguishable color jumps (banding).
	//
	// Dithering in this package is often implemented with blue noise textures.
	Dithered

	// Improves rendering bounds where supported by switching from axis
	// aligned bounding boxes (AABBs) to shape-specific hulls.
	//
	// For example, the AABB of a triangle has a ~50% of ineffective pixels.
	// Using a hull can severely reduce overdraw. For strokes and outlines,
	// the impact is even greater.
	//
	// Unfortunately, hulls have multiple downsides:
	//  - Computing hulls can be non-trivial and increase the number of
	//    triangles to render.
	//  - Color interpolation, although approximated within reason, becomes
	//    notably different from AABB rendering when vertices have different
	//    colors.
	//
	// Therefore, hulls are left as an opt-in optimization tool for stroked
	// and medium to large shapes that support it.
	Hull

	// sentinel
	numFlags
)

func (f Flag) String() string {
	switch f {
	case Bilinear:
		return "Bilinear"
	case Dithered:
		return "Dithered"
	case Hull:
		return "Hull"
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
