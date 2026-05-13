package shapes

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
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

// KernelOptions are used in kernel-based blurs and glows. These are multi-pass operations
// with fixed radiuses and configurable [Downscaling].
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

// Internal function used by BlurK, GlowK and GlowColorK. It downscales the mask,
// applies a custom horizontal kernel shader, then a standard vertical blur shader, and
// upscales the result back, optionally with a BlendLighter blend. At invokeShader,
// KernelLen and Kernel uniforms have been set, as well as the downscaled source image,
// but other uniforms and custom VAs have to be set during invocation.
//
// This function uses the internal offscreen (#0), and if downscaling also (#1).
// Target and mask can be on the same internal atlas.
func (r *Renderer) applyKernel(target *ebiten.Image, mask *ebiten.Image, ox, oy float32, opts KernelOptions, invokeShader func(downHorzTarget *ebiten.Image), lighterBlend bool) {
	if !opts.Downscaling.valid() {
		panic("invalid downscaling value")
	}
	if !opts.HorzKernel.valid() || !opts.VertKernel.valid() {
		panic("invalid GaussKernel value")
	}
	if mask == nil {
		r.Warnings.report(WarnMissingSourceOpSkipped, mask)
		return
	}

	if opts.Downscaling == DownscaleNone {
		r.applyKernelDirect(target, mask, ox, oy, opts, invokeShader, lighterBlend)
		return
	}

	// measures
	df := opts.Downscaling.Factor()
	maskW64, maskH64 := rectSizeF64(mask.Bounds())
	downW64, downH64 := maskW64/float64(df), maskH64/float64(df)
	halfHorzMargin, halfVertMargin := float64(opts.HorzKernel.Radius()), float64(opts.VertKernel.Radius())
	dkernW64, dkernH64 := downW64+halfHorzMargin+halfHorzMargin, downH64+halfVertMargin+halfVertMargin

	// get offscreens and smart clears
	downImgWidth, downImgHeight := math.Ceil(downW64)+2, math.Ceil(downH64)+2
	dkernImgWidth, dkernImgHeight := math.Ceil(dkernW64)+2, math.Ceil(dkernH64)+2
	dkern, _ := r.getTemp(0, int(dkernImgWidth), int(dkernImgHeight), false) // get first as the biggest offscreen
	down, _ := r.getTemp(0, int(downImgWidth), int(downImgHeight), false)    // shared with dkern
	dkernHorz, _ := r.getTemp(1, int(dkernImgWidth), int(downImgHeight), false)
	preBlend := r.opts.Blend
	r.opts.Blend = ebiten.BlendClear
	r.StrokeIntRect(down, down.Bounds(), 2, 0)
	r.FillIntRect(dkern, clockwiseRightBorder(dkern.Bounds(), 1), 0) // *
	r.FillIntRect(dkern, bottomBorder(dkern.Bounds(), 1), 0)
	r.FillIntRect(dkernHorz, clockwiseRightBorder(dkernHorz.Bounds(), 1), 0)
	r.FillIntRect(dkernHorz, bottomBorder(dkernHorz.Bounds(), 1), 0)
	// * Notice that technically dkern content could be overwritten by operations
	//   on 'down' after the clear, but since kernels can't be zero and 'down' already
	//   has 1 pixel margins, this won't happen in practice. Otherwise the clear should
	//   be delayed until after the horz kernel application

	// downscaling
	r.opts.Blend = ebiten.BlendCopy
	df32 := float32(df)
	r.Scale(down, mask, 1, 1, 1.0/df32, opts.Scaling)

	// apply effect
	r.applyKernelOp(down, dkern, dkernHorz, dkernW64, dkernH64, downW64, downH64, opts, invokeShader)

	// upscale
	if lighterBlend {
		r.opts.Blend = ebiten.BlendLighter
	} else {
		r.opts.Blend = preBlend
	}
	fx, fy := ox+-df32-float32(halfHorzMargin)*df32, oy+-df32-float32(halfVertMargin)*df32
	r.Scale(target, dkern, fx, fy, df32, opts.Scaling)
	if lighterBlend {
		r.opts.Blend = preBlend
	}
}

func (r *Renderer) applyKernelDirect(target, mask *ebiten.Image, ox, oy float32, opts KernelOptions, invokeShader func(horzTarget *ebiten.Image), lighterBlend bool) {
	horzKernelLen := opts.HorzKernel.Size()
	ceilHRadius := float32(horzKernelLen)
	ox32, oy32, w32, h32 := rectOriginSizeF32(mask.Bounds())
	w32 += float32(horzKernelLen + horzKernelLen)
	tmp, _ := r.getTemp(0, int(w32), int(h32), false)
	preBlend := r.opts.Blend

	// apply horz kern shader
	r.setDstRectCoords(0, 0, w32, h32)
	//ox32, oy32 = 0.0, 0.0
	sx := ox32 - ceilHRadius
	r.setSrcRectCoords(sx, oy32, sx+w32, oy32+h32)
	r.opts.Blend = ebiten.BlendCopy
	r.opts.Images[0] = mask
	r.opts.Uniforms["KernelLen"] = opts.HorzKernel.Size()
	r.opts.Uniforms["Kernel"] = gaussKernels[opts.HorzKernel]
	invokeShader(tmp) // set VAs, more uniforms, invoke shader and clear(r.opts.Uniforms) if needed

	ceilVRadius := float32(opts.VertKernel.Radius())
	dx := ox - ceilHRadius
	r.setDstRectCoords(dx, oy-ceilVRadius, dx+w32, oy+h32+ceilVRadius)
	r.setSrcRectCoords(0, -ceilVRadius, w32, h32+ceilVRadius)

	r.opts.Blend = preBlend
	if lighterBlend {
		r.opts.Blend = ebiten.BlendLighter
	}
	r.opts.Uniforms["KernelLen"] = opts.VertKernel.Size()
	r.opts.Uniforms["Kernel"] = gaussKernels[opts.VertKernel]
	r.opts.Images[0] = tmp
	target.DrawTrianglesShader32(r.vertices[:], r.indices[:], shaderKernVertFinish.Load(), &r.opts)
	r.opts.Images[0] = nil
	clear(r.opts.Uniforms)
	r.opts.Blend = preBlend
}

func (r *Renderer) applyKernelOp(down, dkern, dkernHorz *ebiten.Image, dkernW64, dkernH64, downW64, downH64 float64, opts KernelOptions, invokeShader func(downHorzTarget *ebiten.Image)) {
	halfHorzMargin, halfVertMargin := float64(opts.HorzKernel.Radius()), float64(opts.VertKernel.Radius())

	// apply horz kern shader
	r.setDstRectCoords(0, 0, float32(dkernW64)+2, float32(downH64)+2)
	r.setSrcRectCoords(float32(-halfHorzMargin), float32(0), float32(downW64+halfHorzMargin)+2, float32(downH64)+2)
	r.opts.Blend = ebiten.BlendCopy
	r.opts.Images[0] = down
	r.opts.Uniforms["KernelLen"] = opts.HorzKernel.Size()
	r.opts.Uniforms["Kernel"] = gaussKernels[opts.HorzKernel]
	invokeShader(dkernHorz) // set VAs, more uniforms, invoke shader and clear(r.opts.Uniforms) if needed

	// apply vert blur kern
	r.opts.Uniforms["KernelLen"] = opts.VertKernel.Size()
	r.opts.Uniforms["Kernel"] = gaussKernels[opts.VertKernel]
	r.setDstRectCoords(0, 0, float32(dkernW64)+2, float32(dkernH64)+2)
	r.setSrcRectCoords(0, float32(-halfVertMargin), float32(dkernW64)+2, float32(downH64+halfVertMargin)+2)
	r.opts.Images[0] = dkernHorz
	dkern.DrawTrianglesShader32(r.vertices[:], r.indices[:], shaderKernVertFinish.Load(), &r.opts)
	r.opts.Images[0] = nil
	clear(r.opts.Uniforms)
}
