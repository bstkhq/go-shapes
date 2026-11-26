package pixcheck

import (
	"image"
	"image/color"
	"image/png"
	"os"
)

// DiffImage returns an image that shows the mismatch between got and want:
// black pixels where the images match, red where they don't.
//
// Precondition: images must be the same size.
func (m Metric) DiffImage(got, want *image.RGBA) *image.RGBA {
	const MinDiffRed = 64

	assertSameSize(got, want)
	delta := m.deltaFunc()
	w, h := got.Rect.Dx(), got.Rect.Dy()
	out := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := range h {
		for x := range w {
			gotPixel := got.RGBAAt(got.Rect.Min.X+x, got.Rect.Min.Y+y)
			wantPixel := want.RGBAAt(want.Rect.Min.X+x, want.Rect.Min.Y+y)
			var red uint8
			if pixelDelta := delta(gotPixel, wantPixel); pixelDelta > 0 {
				red = uint8(MinDiffRed + (255-MinDiffRed)*(int(pixelDelta)-1)/254)
			}
			out.SetRGBA(x, y, color.RGBA{red, 0, 0, 255})
		}
	}
	return out
}

// SavePNG writes img to the given path as a PNG. Ebitengine images must be
// passed through [Read] first.
func SavePNG(path string, img *image.RGBA) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	if err := png.Encode(file, img); err != nil {
		file.Close()
		return err
	}
	return file.Close()
}
