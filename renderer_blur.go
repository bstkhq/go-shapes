package shapes

import "github.com/hajimehoshi/ebiten/v2"

// TODO: implement new blurs on this file and progressively refactor

// Downscaling can be used with some operations
type Downscaling uint8

func (d Downscaling) Valid() bool {
	return d >= DownscaleNone && d <= DownscaleX8
}

func (d Downscaling) Factor() int {
	return int((d + 1) << d)
}

// UpscalingFilter returns the recommended filter shader for re-upscaling
// after the downscale. Returns nil for DownscaleNone.
func (d Downscaling) UpscalingFilter() *ebiten.Shader {
	switch d {
	case DownscaleNone:
		return nil
	case DownscaleX2:
		ensureShaderBilinearLoaded()
		return shaderBilinear
	default:
		return nil // TODO: bicubic
		// ensureShaderBicubicLoaded()
		// return shaderBicubic
	}
}

const (
	DownscaleNone Downscaling = iota
	DownscaleX2
	DownscaleX4
	DownscaleX8
)

// ApplyBlurVogel applies a gaussian blur using numSamples distributed with a vogel disk.
//
// Common numSamples values:
//   - 16: low quality, but fast and practical in many scenarios.
//   - 32: medium quality and useful for a wide variety of blur effects.
//   - 64: maximum allowed value.
//
// This function uses one internal offscreen (#0) if downscaling != DownscaleNone, and target and mask
// can be on the same internal atlas.
func (r *Renderer) ApplyBlurVogel(target *ebiten.Image, mask *ebiten.Image, ox, oy, radius float32, numSamples int, downscaling Downscaling) {
	// TODO: precompute vogel disk?
}

// ApplyBlur2D
func (r *Renderer) ApplyBlur2D(target *ebiten.Image, mask *ebiten.Image, ox, oy, radius float32, downscaling Downscaling) {
	// TODO: precompute vogel disk?
}
