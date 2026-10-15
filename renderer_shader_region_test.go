package shapes

import (
	"image"
	"image/color"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

// TestExpandShaderAxis covers cases that DrawImgShader's 1:1 readback test
// cannot: non-1:1 source scaling and negative fractional coordinates. The
// former must preserve the original scale, while the latter must round away
// from the destination region rather than truncate towards zero.
func TestExpandShaderAxis(t *testing.T) {
	tests := []struct {
		name                   string
		dstO, dstF, srcO, srcF float32
		wantDstO, wantDstF     float32
		wantSrcO, wantSrcF     float32
	}{
		{
			name: "positive coordinates with 2x source scale",
			dstO: 1.25, dstF: 4.75, srcO: 10, srcF: 17,
			wantDstO: 1, wantDstF: 5, wantSrcO: 9.5, wantSrcF: 17.5,
		},
		{
			name: "negative coordinates with 2x source scale",
			dstO: -3.75, dstF: -1.25, srcO: 10, srcF: 15,
			wantDstO: -4, wantDstF: -1, wantSrcO: 9.5, wantSrcF: 15.5,
		},
	}

	for _, test := range tests {
		dstO, dstF, srcO, srcF := expandShaderAxis(test.dstO, test.dstF, test.srcO, test.srcF)
		if dstO != test.wantDstO || dstF != test.wantDstF || srcO != test.wantSrcO || srcF != test.wantSrcF {
			t.Errorf("%s: got (%v, %v, %v, %v), want (%v, %v, %v, %v)",
				test.name, dstO, dstF, srcO, srcF,
				test.wantDstO, test.wantDstF, test.wantSrcO, test.wantSrcF,
			)
		}
	}
}

// TestDrawImgShaderRegionModes verifies the modes through actual GPU triangle
// coverage. Both draws use the same fractional placement and non-zero-origin
// source. RegionExact only covers the requested rectangle; RegionExpanded also
// evaluates the right and bottom boundary pixels. The shader paints extrapolated
// coordinates magenta, making source stretching fail the expected image too.
func TestDrawImgShaderRegionModes(t *testing.T) {
	const safeCopyShader = `//kage:unit pixels
package main

func Fragment(_ vec4, sourceCoords vec2, _ vec4) vec4 {
	clr := imageSrc0At(sourceCoords)
	if clr.a == 0 {
		return vec4(1, 0, 1, 1)
	}
	return clr
}
`

	shader, err := ebiten.NewShader([]byte(safeCopyShader))
	if err != nil {
		t.Fatal(err)
	}

	const SourceOX, SourceOY = 8, 12
	source := ebiten.NewImageWithOptions(image.Rect(SourceOX, SourceOY, SourceOX+2, SourceOY+2), nil)
	source.Set(SourceOX+0, SourceOY+0, color.RGBA{255, 0, 0, 255})
	source.Set(SourceOX+1, SourceOY+0, color.RGBA{0, 255, 0, 255})
	source.Set(SourceOX+0, SourceOY+1, color.RGBA{0, 0, 255, 255})
	source.Set(SourceOX+1, SourceOY+1, color.RGBA{255, 255, 255, 255})

	target := ebiten.NewImage(10, 3)
	r := NewRenderer()
	r.DrawImgShader(target, source, 1.25, 0.25, NoMargins, RegionExact, shader)
	r.DrawImgShader(target, source, 6.25, 0.25, NoMargins, RegionExpanded, shader)

	readback := NewReadTestApp(target)
	if err := ebiten.RunGame(readback); err != nil {
		t.Fatal(err)
	}

	expected := image.NewRGBA(target.Bounds())
	for _, ox := range []int{1, 6} {
		expected.SetRGBA(ox+0, 0, color.RGBA{255, 0, 0, 255})
		expected.SetRGBA(ox+1, 0, color.RGBA{0, 255, 0, 255})
		expected.SetRGBA(ox+0, 1, color.RGBA{0, 0, 255, 255})
		expected.SetRGBA(ox+1, 1, color.RGBA{255, 255, 255, 255})
	}
	for y := range 2 {
		expected.SetRGBA(8, y, color.RGBA{255, 0, 255, 255})
	}
	for x := 6; x <= 8; x += 1 {
		expected.SetRGBA(x, 2, color.RGBA{255, 0, 255, 255})
	}

	for y := range target.Bounds().Dy() {
		for x := range target.Bounds().Dx() {
			got, want := readback.RGBA.RGBAAt(x, y), expected.RGBAAt(x, y)
			if got != want {
				t.Errorf("pixel (%d, %d): got %v, want %v", x, y, got, want)
			}
		}
	}
}
