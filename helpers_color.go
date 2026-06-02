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

func interpQuadColor(tl, tr, br, bl [4]float32, origin, size, iCoords PointF32) [4]float32 {
	const epsilon = 1e-6

	size.X = maxMagnitude(size.X, epsilon)
	size.Y = maxMagnitude(size.Y, epsilon)
	u := clamp((iCoords.X-origin.X)/size.X, 0, 1)
	v := clamp((iCoords.Y-origin.Y)/size.Y, 0, 1)

	var out [4]float32
	for i := range 4 {
		top := lerp(tl[i], tr[i], u)
		bottom := lerp(bl[i], br[i], u)
		out[i] = top + v*(bottom-top)
	}

	return out
}

// interpTriColor interpolates a 4-channel color within a triangle using barycentric coordinates
func interpTriColor(a, b, c [4]float32, aCoords, bCoords, cCoords, iCoords PointF32) [4]float32 {
	denom := (bCoords.Y-cCoords.Y)*(aCoords.X-cCoords.X) + (cCoords.X-bCoords.X)*(aCoords.Y-cCoords.Y)

	const epsilon = 1e-6
	if denom < epsilon && denom > -epsilon {
		return a // degenerate case
	}

	// determine weights by solving linear system where:
	// wa+wb+wc=1, iCoords=wa*aCoords+wb*bCoords+wc*cCoords
	ciDiff := iCoords.Sub(cCoords)
	wa := ((bCoords.Y-cCoords.Y)*ciDiff.X + (cCoords.X-bCoords.X)*ciDiff.Y) / denom
	wb := ((cCoords.Y-aCoords.Y)*ciDiff.X + (aCoords.X-cCoords.X)*ciDiff.Y) / denom
	wc := 1 - wa - wb

	var out [4]float32
	for i := range 4 {
		out[i] = wa*a[i] + wb*b[i] + wc*c[i]
	}

	return out
}

// interpTriQuadColor interpolates the color of a point within a quad composed
// by two triangles (upper right and lower left). this can be used to match
// colors as closely as possible the default interpolation if triangles are
// modified for hull adjustments
func interpTriQuadColor(tl, tr, br, bl [4]float32, origin, size, iCoords PointF32) [4]float32 {
	// normalize origin and perform half plane check
	iCoords = iCoords.Sub(origin)
	cross := size.X*iCoords.Y - size.Y*iCoords.X

	if cross <= 0 { // upper-right
		return interpTriColor(tl, tr, br, PtF32(0, 0), PtF32(size.X, 0), size, iCoords)
	} else { // lower-left
		return interpTriColor(tl, br, bl, PtF32(0, 0), size, PtF32(0, size.Y), iCoords)
	}
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

// precondition: vertices contains the 4 AABB-defining vertices in clockwise
// order: TL, TR, BR, BL
func setupHullVertices(vertices []ebiten.Vertex, hull []PointF32) []ebiten.Vertex {
	if len(vertices) != 4 {
		panic("expected 4 vertices")
	}

	// TODO

	return vertices
}
