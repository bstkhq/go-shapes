package shapes

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
