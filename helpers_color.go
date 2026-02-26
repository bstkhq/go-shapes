package shapes

import (
	"image/color"

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
)

func RGBAF32(clr color.Color) [4]float32 {
	r, g, b, a := clr.RGBA()
	return [4]float32{float32(r) / 65535.0, float32(g) / 65535.0, float32(b) / 65535.0, float32(a) / 65535.0}
}

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
