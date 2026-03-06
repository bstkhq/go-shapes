package shapes

import (
	"image/color"
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

func NewMargins(horz, vert float32) Margins {
	return Margins{
		Left:   horz,
		Right:  horz,
		Top:    vert,
		Bottom: vert,
	}
}

// Downscaling is a common technique used in graphics where an effect is not applied at full
// resolution, but a smaller offscreen. This is done to reduce the amount of pixels to process,
// but it has some downsides:
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
// See [TC], [TR], [CTR] and company for predefined constants.
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

// GaussKernel is a precomputed gaussian kernel used in certain blur and glow operations.
// Kernels are unidimensional and always used in two passes. Multi-pass shaders are often
// used to avoid quadratic sampling, but they also have more overhead due to breaking
// batching, changing shaders, depending on previous results and so on.
type GaussKernel uint8

const (
	GaussK3 GaussKernel = iota
	GaussK5
	GaussK7
	GaussK9
	GaussK11
	GaussK13
	GaussK15
	GaussK17
	GaussK19
	GaussK21
	GaussK23
	GaussK25
	GaussK27
	GaussK29
	GaussK31
)

// Radius returns the kernel radius. For example, GaussK3 has radius 1,
// Gauss15 has radius 7, etc.
func (k GaussKernel) Radius() int {
	return (k.Size() - 1) >> 1
}

// Size returns the size of the kernel. For example, GaussK3 has size 3,
// GaussK15 has size 15, etc.
func (k GaussKernel) Size() int {
	ik := int(k)
	return 3 + ik + ik
}

func (k GaussKernel) valid() bool {
	return k >= GaussK3 && k <= GaussK31
}

// gaussian kernels are encoded as one-dimension separable filters, with the values
// for the center and a single side (both sides have to be applied in the shader)
// see internal/gen/gauss_kernels.go for generation

var gaussSig3Kernels = [][16]float32{ // sigma = radius/3.0
	{0.97826492, 0.01086754, 0.00000000, 0.00000000, 0.00000000, 0.00000000, 0.00000000, 0.00000000, 0.00000000, 0.00000000, 0.00000000, 0.00000000, 0.00000000, 0.00000000, 0.00000000, 0.00000000}, // 3x3
	{0.59825683, 0.19422555, 0.00664603, 0.00000000, 0.00000000, 0.00000000, 0.00000000, 0.00000000, 0.00000000, 0.00000000, 0.00000000, 0.00000000, 0.00000000, 0.00000000, 0.00000000, 0.00000000}, // 5x5
	{0.39905028, 0.24203623, 0.05400558, 0.00443305, 0.00000000, 0.00000000, 0.00000000, 0.00000000, 0.00000000, 0.00000000, 0.00000000, 0.00000000, 0.00000000, 0.00000000, 0.00000000, 0.00000000}, // 7x7
	{0.29937241, 0.22597815, 0.09719199, 0.02381792, 0.00332573, 0.00000000, 0.00000000, 0.00000000, 0.00000000, 0.00000000, 0.00000000, 0.00000000, 0.00000000, 0.00000000, 0.00000000, 0.00000000}, // 9x9
	{0.23955941, 0.20009684, 0.11660608, 0.04740850, 0.01344761, 0.00266126, 0.00000000, 0.00000000, 0.00000000, 0.00000000, 0.00000000, 0.00000000, 0.00000000, 0.00000000, 0.00000000, 0.00000000}, // 11x11
	{0.19967563, 0.17621312, 0.12110939, 0.06482519, 0.02702316, 0.00877313, 0.00221820, 0.00000000, 0.00000000, 0.00000000, 0.00000000, 0.00000000, 0.00000000, 0.00000000, 0.00000000, 0.00000000}, // 13x13
	{0.17118068, 0.15616027, 0.11855449, 0.07490263, 0.03938291, 0.01723257, 0.00627515, 0.00190165, 0.00000000, 0.00000000, 0.00000000, 0.00000000, 0.00000000, 0.00000000, 0.00000000, 0.00000000}, // 15x15
	{0.14980486, 0.13963348, 0.11307864, 0.07956076, 0.04863452, 0.02582960, 0.01191840, 0.00477799, 0.00166418, 0.00000000, 0.00000000, 0.00000000, 0.00000000, 0.00000000, 0.00000000, 0.00000000}, // 17x17
	{0.13317600, 0.12597909, 0.10663900, 0.08077532, 0.05475029, 0.03320773, 0.01802341, 0.00875346, 0.00380424, 0.00147945, 0.00000000, 0.00000000, 0.00000000, 0.00000000, 0.00000000, 0.00000000}, // 19x19
	{0.11987063, 0.11459602, 0.10012436, 0.07995093, 0.05834730, 0.03891629, 0.02372224, 0.01321580, 0.00672891, 0.00313119, 0.00133164, 0.00000000, 0.00000000, 0.00000000, 0.00000000, 0.00000000}, // 21x21
	{0.10898277, 0.10500414, 0.09391870, 0.07798208, 0.06010833, 0.04301022, 0.02856970, 0.01761720, 0.01008475, 0.00535908, 0.00264371, 0.00121069, 0.00000000, 0.00000000, 0.00000000, 0.00000000}, // 23x23
	{0.09990836, 0.09683450, 0.08816882, 0.07541479, 0.06059748, 0.04574138, 0.03243549, 0.02160670, 0.01352113, 0.00794866, 0.00438967, 0.00227733, 0.00110988, 0.00000000, 0.00000000, 0.00000000}, // 25x25
	{0.09222910, 0.08980571, 0.08291093, 0.07257574, 0.06023420, 0.04739872, 0.03536405, 0.02501666, 0.01677909, 0.01067037, 0.00643373, 0.00367805, 0.00199363, 0.00102457, 0.00000000, 0.00000000}, // 27x27
	{0.08564620, 0.08370223, 0.07813109, 0.06965762, 0.05931593, 0.04824273, 0.03747576, 0.02780525, 0.01970430, 0.01333685, 0.00862191, 0.00532367, 0.00313962, 0.00176848, 0.00095144, 0.00000000}, // 29x29
	{0.07994048, 0.07835755, 0.07379436, 0.06677190, 0.05804870, 0.04848635, 0.03891121, 0.03000255, 0.02222644, 0.01582012, 0.01081877, 0.00710844, 0.00448744, 0.00272177, 0.00158611, 0.00088806}, // 31x31
}

var gaussKernels = gaussSig3Kernels

// KernelOptions are used in kernel-based blur and glows.
//
// Scaling is optional and can be nil, but for downscaling factors of [DownscaleX8] and above
// using bicubic scaling is heavily recommended. At [DownscaleX4] the decision depends more
// on the desired quality, effect and performance constraints.
type KernelOptions struct {
	Downscaling Downscaling
	HorzKernel  GaussKernel
	VertKernel  GaussKernel

	// Scaling options can be provided if defaults need to be overridden.
	Scaling *ScaleOptions
}

// KernelOpts creates KernelOptions with the same kernel for both axes.
func KernelOpts(downscaling Downscaling, kernel GaussKernel) KernelOptions {
	return KernelOptions{
		Downscaling: downscaling,
		HorzKernel:  kernel,
		VertKernel:  kernel,
	}
}

type GradientOptions struct {
	// Starting gradient color.
	From [4]float64

	// End gradient color.
	To [4]float64

	// Steps controls the number of colors in the gradient.
	//  - Steps <= 0 specifies a continuous gradient (no color limit).
	//  - Steps > 0 specifies a stepped gradient.
	Steps int

	// Dither determines whether the gradient will have dithering applied.
	//
	// Dithering is only necessary on subtle gradients or alpha transitions,
	// where very similar colors can cause banding or flickering.
	Dither bool

	// Bias is a value in [-1, 1] that controls the color interpolation:
	//  - Zero generates a linear gradient (both colors have the same presence).
	//  - Negative values give the start color more presence.
	//  - Positive values give the end color more presence.
	//
	// The interpolation is based on Schlick's bias function.
	Bias float32
}

// GradientOpts creates GradientOptions for a continuous gradient.
func GradientOpts(from, to color.Color, dither bool) GradientOptions {
	return GradientOptions{
		From:   colorToF64(from),
		To:     colorToF64(to),
		Dither: dither,
	}
}

// StepGradientOpts creates GradientOptions for a stepped gradient.
func StepGradientOpts(from, to color.Color, steps int) GradientOptions {
	return GradientOptions{
		From:  colorToF64(from),
		To:    colorToF64(to),
		Steps: steps,
	}
}

func mapBool[T any](b bool, falseValue, trueValue T) T {
	if b {
		return trueValue
	}
	return falseValue
}
