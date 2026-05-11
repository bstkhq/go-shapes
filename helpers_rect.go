package shapes

import (
	"image"
)

// RectWithSize is a helper for image.Rectangle initialization
// which receives the rect origin and size (unlike image.Rect, which
// takes origin and end).
func RectWithSize(ox, oy int, w, h int) image.Rectangle {
	return image.Rect(ox, oy, ox+w, oy+h)
}

func RectPointsF32(bounds image.Rectangle) (min, max PointF32) {
	return PtF32(float32(bounds.Min.X), float32(bounds.Min.Y)), PtF32(float32(bounds.Max.X), float32(bounds.Max.Y))
}

func rectSize(bounds image.Rectangle) (w, h int) {
	return bounds.Dx(), bounds.Dy()
}

func rectOriginSize(bounds image.Rectangle) (ox, oy, w, h int) {
	return bounds.Min.X, bounds.Min.Y, bounds.Dx(), bounds.Dy()
}

func rectOriginSizeF32(bounds image.Rectangle) (ox, oy, w, h float32) {
	return float32(bounds.Min.X), float32(bounds.Min.Y), float32(bounds.Dx()), float32(bounds.Dy())
}

func rectOriginF32(bounds image.Rectangle) (ox, oy float32) {
	return float32(bounds.Min.X), float32(bounds.Min.Y)
}

func rectSizeF32(bounds image.Rectangle) (w, h float32) {
	return float32(bounds.Dx()), float32(bounds.Dy())
}

func rectSizeF64(bounds image.Rectangle) (w, h float64) {
	return float64(bounds.Dx()), float64(bounds.Dy())
}

func rightBorder(bounds image.Rectangle, borderSize int) image.Rectangle {
	bounds.Min.X = bounds.Max.X - borderSize
	return bounds
}

func bottomBorder(bounds image.Rectangle, borderSize int) image.Rectangle {
	bounds.Min.Y = bounds.Max.Y - borderSize
	return bounds
}

// right border without overlapping the bottom
func clockwiseRightBorder(bounds image.Rectangle, borderSize int) image.Rectangle {
	bounds = rightBorder(bounds, borderSize)
	bounds.Max.Y -= borderSize
	return bounds
}
