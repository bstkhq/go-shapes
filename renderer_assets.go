package shapes

import (
	"bytes"
	_ "embed"
	"image/png"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
)

//go:embed assets/blue64.png
var blueNoiseBytes []byte
var blueNoiseOnce sync.Once
var blueNoise64RGB *ebiten.Image

func loadBlueNoise64RGB() *ebiten.Image {
	blueNoiseOnce.Do(loadBlueNoise64RGBOnce)
	return blueNoise64RGB
}

func loadBlueNoise64RGBOnce() {
	img, err := png.Decode(bytes.NewReader(blueNoiseBytes))
	if err != nil {
		panic(err)
	}

	blueNoise64RGB = ebiten.NewImageFromImage(img)
}
