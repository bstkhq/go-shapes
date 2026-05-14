package shapes

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

// Common blend modes not directly exposed on Ebitengine.
var (
	BlendSubtract = ebiten.Blend{
		BlendFactorSourceRGB:        ebiten.BlendFactorOne,
		BlendFactorSourceAlpha:      ebiten.BlendFactorOne,
		BlendFactorDestinationRGB:   ebiten.BlendFactorOne,
		BlendFactorDestinationAlpha: ebiten.BlendFactorOne,
		BlendOperationRGB:           ebiten.BlendOperationReverseSubtract,
		BlendOperationAlpha:         ebiten.BlendOperationReverseSubtract,
	}
	BlendMultiply = ebiten.Blend{
		BlendFactorSourceRGB:        ebiten.BlendFactorDestinationColor,
		BlendFactorSourceAlpha:      ebiten.BlendFactorDestinationColor,
		BlendFactorDestinationRGB:   ebiten.BlendFactorOneMinusSourceAlpha,
		BlendFactorDestinationAlpha: ebiten.BlendFactorOneMinusSourceAlpha,
		BlendOperationRGB:           ebiten.BlendOperationAdd,
		BlendOperationAlpha:         ebiten.BlendOperationAdd,
	}
	blendDarkGlow = ebiten.Blend{
		BlendFactorSourceRGB:        ebiten.BlendFactorOneMinusDestinationColor,
		BlendFactorSourceAlpha:      ebiten.BlendFactorOneMinusDestinationColor,
		BlendFactorDestinationRGB:   ebiten.BlendFactorOne,
		BlendFactorDestinationAlpha: ebiten.BlendFactorOne,
		BlendOperationRGB:           ebiten.BlendOperationReverseSubtract,
		BlendOperationAlpha:         ebiten.BlendOperationAdd,
	}
)

// RGBAF32 converts a color to float32 format (R, G, B, A).
func RGBAF32(clr color.Color) [4]float32 {
	r, g, b, a := clr.RGBA()
	return [4]float32{float32(r) / 65535.0, float32(g) / 65535.0, float32(b) / 65535.0, float32(a) / 65535.0}
}

// RGBF32 converts the color rgb values to float32 format (no alpha).
func RGBF32(clr color.Color) [3]float32 {
	r, g, b, _ := clr.RGBA()
	return [3]float32{float32(r) / 65535.0, float32(g) / 65535.0, float32(b) / 65535.0}
}

func colorToF64(clr color.Color) [4]float64 {
	r, g, b, a := clr.RGBA()
	return [4]float64{float64(r) / 65535.0, float64(g) / 65535.0, float64(b) / 65535.0, float64(a) / 65535.0}
}

func interpVertexColor(a, b ebiten.Vertex, t float32) (cr, cg, cb, ca float32) {
	return lerp(a.ColorR, b.ColorR, t), lerp(a.ColorG, b.ColorG, t), lerp(a.ColorB, b.ColorB, t), lerp(a.ColorA, b.ColorA, t)
}

func rgbToOklab(rgb [3]float64) [3]float64 {
	linR, linG, linB := linearize(rgb[0]), linearize(rgb[1]), linearize(rgb[2])
	x := math.Pow(0.4122214708*linR+0.5363325363*linG+0.0514459929*linB, 1.0/3.0)
	y := math.Pow(0.2119034982*linR+0.6806995451*linG+0.1073969566*linB, 1.0/3.0)
	z := math.Pow(0.0883024619*linR+0.2817188376*linG+0.6299787005*linB, 1.0/3.0)

	l := 0.2104542553*x + 0.7936177850*y - 0.0040720468*z
	a := 1.9779984951*x - 2.4285922050*y + 0.4505937099*z
	b := 0.0259040371*x + 0.7827717662*y - 0.8086757660*z
	return [3]float64{l, a, b}
}

func toOklab(red, green, blue float64) (l, a, b float64) {
	linR, linG, linB := linearize(red), linearize(green), linearize(blue)
	x := math.Pow(0.4122214708*linR+0.5363325363*linG+0.0514459929*linB, 1.0/3.0)
	y := math.Pow(0.2119034982*linR+0.6806995451*linG+0.1073969566*linB, 1.0/3.0)
	z := math.Pow(0.0883024619*linR+0.2817188376*linG+0.6299787005*linB, 1.0/3.0)

	l = 0.2104542553*x + 0.7936177850*y - 0.0040720468*z
	a = 1.9779984951*x - 2.4285922050*y + 0.4505937099*z
	b = 0.0259040371*x + 0.7827717662*y - 0.8086757660*z
	return l, a, b
}

func linearize(colorChan float64) float64 {
	if colorChan >= 0.04045 {
		return math.Pow((colorChan+0.055)/1.055, 2.4)
	} else {
		return colorChan / 12.92
	}
}
