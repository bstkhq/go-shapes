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

// hint: ark10px atlas and map generated with https://gist.github.com/tinne26/a986856a78d73e52583e462e7ddd5613

//go:embed assets/ark10px-atlas.png
var ark10pxAtlasBytes []byte
var ark10pxOnce sync.Once
var ark10pxAtlas *ebiten.Image

//go:embed assets/ark10px-map.bin
var ark10pxMap []byte

const (
	fontMapIdxGlyphsPerRow    = 0
	fontMapIdxGlyphFrameWidth = 1
	fontMapIdxAscent          = 2
	fontMapIdxDescent         = 3
	fontMapIdxLineGap         = 4
	fontMapIdxCapHeight       = 5
	fontMapIdxMidHeight       = 6
	fontMapIdxSpaceWidth      = 7
	fontMapIdxFirstGlyphWidth = 8
)

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

func loadArk10pxAtlas() *ebiten.Image {
	ark10pxOnce.Do(loadArk10pxAtlasOnce)
	return ark10pxAtlas
}

func loadArk10pxAtlasOnce() {
	img, err := png.Decode(bytes.NewReader(ark10pxAtlasBytes))
	if err != nil {
		panic(err)
	}

	ark10pxAtlas = ebiten.NewImageFromImage(img)
}
