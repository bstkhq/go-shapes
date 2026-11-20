package shapes

import "github.com/hajimehoshi/ebiten/v2"

func newBWCheckerboard(w, h, border, cellSize int) *ebiten.Image {
	return newBWImage(w, h, border, func(x, y int) bool {
		return (x/cellSize+y/cellSize)%2 == 0
	})
}

func newBWHorzStripes(w, h, border, stripeHeight int) *ebiten.Image {
	return newBWImage(w, h, border, func(_, y int) bool {
		return (y/stripeHeight)%2 == 0
	})
}

func newBWVertStripes(w, h, border, stripeWidth int) *ebiten.Image {
	return newBWImage(w, h, border, func(x, _ int) bool {
		return (x/stripeWidth)%2 == 0
	})
}

// newBWImage creates a black and white image with a transparent border of
// the given size, with white pixels where light reports true.
func newBWImage(w, h, border int, light func(x, y int) bool) *ebiten.Image {
	pix := make([]byte, w*h*4)
	for y := border; y < h-border; y += 1 {
		for x := border; x < w-border; x += 1 {
			i, v := (y*w+x)*4, byte(0)
			if light(x, y) {
				v = 255
			}
			pix[i], pix[i+1], pix[i+2], pix[i+3] = v, v, v, 255
		}
	}
	img := ebiten.NewImage(w, h)
	img.WritePixels(pix)
	return img
}
