package pixcheck

import "image"

// AlphaSum returns the total alpha of img, in pixels: a fully opaque pixel adds
// 1. It measures the area of the shapes in an image.
func AlphaSum(img *image.RGBA) float64 {
	var sum float64
	for y := img.Rect.Min.Y; y < img.Rect.Max.Y; y++ {
		for x := img.Rect.Min.X; x < img.Rect.Max.X; x++ {
			sum += float64(img.RGBAAt(x, y).A) / 255
		}
	}
	return sum
}

// AlphaCentroid returns the center of the shapes in img, weighting each pixel
// center by its alpha, relative to the image origin. If img is fully
// transparent, it returns (0, 0).
func AlphaCentroid(img *image.RGBA) (x, y float64) {
	var sum, sumX, sumY float64
	for py := img.Rect.Min.Y; py < img.Rect.Max.Y; py++ {
		for px := img.Rect.Min.X; px < img.Rect.Max.X; px++ {
			alpha := float64(img.RGBAAt(px, py).A) / 255
			sum += alpha
			sumX += alpha * (float64(px-img.Rect.Min.X) + 0.5)
			sumY += alpha * (float64(py-img.Rect.Min.Y) + 0.5)
		}
	}
	if sum == 0 {
		return 0, 0
	}
	return sumX / sum, sumY / sum
}
