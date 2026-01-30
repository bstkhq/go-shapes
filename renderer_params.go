package shapes

// Downscaling is a common technique used in graphics where an effect is not applied at full
// resolution, but rather a smaller offscreen. The main upside is that less pixels to process
// typically mean less GPU load.
//
// Downscaling also has important weaknesses and downsides:
//   - Non-trivial overhead due to the extra steps (downscale + effect + upscale back), which
//     can be worse than paying for a full resolution operation if processing many small images.
//   - The upscaled effects can look blocky. This depends on the effect type and the downscaling
//     level; downscaling is not a silver bullet.
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

func (d Downscaling) Factor() int {
	return int(1 << d)
}

// Clamping is a bitset used to express which borders to clamp in some operations
// and effects (e.g. [Renderer.ApplyHardShadow]()).
type Clamping uint8

const (
	ClampNone   Clamping = 0b0000
	ClampTop    Clamping = 0b1000
	ClampBottom Clamping = 0b0100
	ClampLeft   Clamping = 0b0010
	ClampRight  Clamping = 0b0001

	ClampTopLeft     Clamping = ClampTop | ClampLeft
	ClampTopRight    Clamping = ClampTop | ClampRight
	ClampBottomLeft  Clamping = ClampBottom | ClampLeft
	ClampBottomRight Clamping = ClampBottom | ClampRight
)

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
)

func (k GaussKernel) Radius() int {
	return (k.Size() - 1) >> 1
}

func (k GaussKernel) Size() int {
	ik := int(k)
	return 3 + ik + ik
}

func (k GaussKernel) valid() bool {
	return k >= GaussK3 && k <= GaussK17
}

// gaussian kernels are encoded as one-dimension separable filters, with the values
// for the center and a single side (both sides have to be applied in the shader)
var gaussKernels = [][9]float32{ // binomial forms
	{0.5000000, 0.2500000, 0.0000000, 0.0000000, 0.0000000, 0.0000000, 0.0000000, 0.0000000, 0.0000000}, // 3x3
	{0.3750000, 0.2500000, 0.0625000, 0.0000000, 0.0000000, 0.0000000, 0.0000000, 0.0000000, 0.0000000}, // 5x5
	{0.3125000, 0.2343750, 0.0937500, 0.0156250, 0.0000000, 0.0000000, 0.0000000, 0.0000000, 0.0000000}, // 7x7
	{0.2734375, 0.2187500, 0.1093750, 0.0312500, 0.0039063, 0.0000000, 0.0000000, 0.0000000, 0.0000000}, // 9x9
	{0.2460937, 0.2050781, 0.1171875, 0.0439453, 0.0097656, 0.0009765, 0.0000000, 0.0000000, 0.0000000}, // 11x11
	{0.2255859, 0.1933593, 0.1208496, 0.0537109, 0.0161132, 0.0029296, 0.0002441, 0.0000000, 0.0000000}, // 13x13
	{0.2094726, 0.1832885, 0.1221923, 0.0610961, 0.0222167, 0.0055542, 0.0008545, 0.0000611, 0.0000000}, // 15x15
	{0.1963806, 0.1745605, 0.1221924, 0.0666504, 0.0277709, 0.0085449, 0.0018311, 0.0002441, 0.0001526}, // 17x17
}

// var gaussKerns = [][9]float32{ // true gaussians
// 	{0.7869860, 0.1065069, 0.0000000, 0.0000000, 0.0000000, 0.0000000, 0.0000000, 0.0000000, 0.0000000}, // 3x3
// 	{0.4026199, 0.2442013, 0.0544887, 0.0000000, 0.0000000, 0.0000000, 0.0000000, 0.0000000, 0.0000000}, // 5x5
// 	{0.2706821, 0.2167453, 0.1112808, 0.0366328, 0.0000000, 0.0000000, 0.0000000, 0.0000000, 0.0000000}, // 7x7
// 	{0.2041637, 0.1801738, 0.1238315, 0.0662822, 0.0276306, 0.0000000, 0.0000000, 0.0000000, 0.0000000}, // 9x9
// 	{0.1639672, 0.1513608, 0.1190646, 0.0798114, 0.0455890, 0.0221906, 0.0000000, 0.0000000, 0.0000000}, // 11x11
// 	{0.1370228, 0.1296181, 0.1097193, 0.0831085, 0.0563317, 0.0341669, 0.0185440, 0.0000000, 0.0000000}, // 13x13
// 	{0.1176958, 0.1129886, 0.0999667, 0.0815125, 0.0612548, 0.0424232, 0.0270778, 0.0159284, 0.0000000}, // 15x15
// 	{0.1031526, 0.0999789, 0.0910319, 0.0778637, 0.0625652, 0.0472267, 0.0334887, 0.0223083, 0.0139602}, // 17x17
// }

// KernelOptions are used in kernel-based blur and glows.
type KernelOptions struct {
	Downscaling Downscaling
	HorzKernel  GaussKernel
	VertKernel  GaussKernel

	// ColorMix determines how the renderer and source image colors are mixed during
	// operation. It can smoothly go from 0.0 (use only renderer colors) to 1.0 (use
	// only source image colors)
	ColorMix float32

	// Scaling options can be provided if defaults need to be overridden.
	Scaling *ScaleOptions
}
