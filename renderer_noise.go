package shapes

import (
	"image/png"

	"github.com/hajimehoshi/ebiten/v2"
)

// Noise draws pseudo-random, hash-based white noise with the current renderer color over
// the given target.
//
// The cycle parameter controls the noise animation. Progressively increasing the cycle
// value from 0 to 1 and looping back to zero will create a continuous, looping animation
// with an organic feel. If you don't need animation, leave cycle to zero to reduce shader
// calculations.
//
// Seed must be in [0, 1].
func (r *Renderer) Noise(target *ebiten.Image, intensity float32, seed, cycle float32) {
	if seed < 0.0 || seed > 1.0 {
		r.Warnings.report(WarnInvalidNoiseSeedClamped, seed)
		seed = clamp(seed, 0.0, 1.0)
	}
	r.setFlatCustomVAs(intensity, seed, cycle, 0.0)
	tox, toy, tw, th := rectOriginSizeF32(target.Bounds())
	r.DrawRectShader(target, tox, toy, tw, th, NoMargins, shaderNoise.Load())
}

// NoiseGolden draws a grid geometric noise with the current renderer color over the
// given target. This noise is based on the golden ratio and it's highly sensitive
// to the scale, producing very different results at different levels. Some interesting
// scales are 0.06, 1.0, 64.0, 93.0 and upwards (patterns start to darken and vanish
// afterwards).
//
// The param t controls the animation pace. Increase t at a rate of 1.0 per second
// for a natural animation rate.
func (r *Renderer) NoiseGolden(target *ebiten.Image, scale, intensity, t float32) {
	r.setFlatCustomVAs(scale, intensity, t, 0)
	tox, toy, tw, th := rectOriginSizeF32(target.Bounds())
	r.DrawRectShader(target, tox, toy, tw, th, NoMargins, shaderNoiseGolden.Load())
}

func (r *Renderer) ensureBlueNoiseLoaded() {
	if r.blueNoise64RGB != nil {
		return
	}

	file, err := assets.Open("assets/blue64.png")
	if err != nil {
		panic(err)
	}

	img, err := png.Decode(file)
	if err != nil {
		panic(err)
	}

	r.blueNoise64RGB = ebiten.NewImageFromImage(img)
}
