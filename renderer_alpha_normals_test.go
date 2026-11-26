package shapes

import (
	"errors"
	"image"
	"image/color"
	"math"
	"strconv"
	"strings"
	"testing"

	"github.com/bstkhq/go-shapes/internal/pixcheck"
	"github.com/hajimehoshi/ebiten/v2"
)

// Checks both parts of the alpha normals (the descent direction and the offset
// to the boundary), on edges at different coverages and angles, and on the
// pixels that must get no data.
func TestAlphaNormalsCompute(t *testing.T) {
	cases := []struct {
		name         string
		clamp        bool
		source, want string // alpha, and normals as "dx,dy,offset" or "none"
	}{
		{
			// a vertical edge covering a quarter of the edge pixel: its
			// boundary is 0.25 pixels before its center, on the high side
			name:  "edge 0.25",
			clamp: true,
			source: `
				255 255 255 064 000 000
				255 255 255 064 000 000
				255 255 255 064 000 000
				255 255 255 064 000 000
			`,
			want: `
				none     none 1,0,-0.5 1,0,0.25  1,0,0.5     none
				none     none 1,0,-0.5 1,0,0.25  1,0,0.5     none
				none     none 1,0,-0.5 1,0,0.25  1,0,0.5     none
				none     none 1,0,-0.5 1,0,0.25  1,0,0.5     none
			`,
		},
		{
			// the same edge, crossing the edge pixel through its center
			name:  "edge 0.5",
			clamp: true,
			source: `
				255 255 255 128 000 000
				255 255 255 128 000 000
				255 255 255 128 000 000
				255 255 255 128 000 000
			`,
			want: `
				none     none 1,0,-0.5    1,0,0  1,0,0.5     none
				none     none 1,0,-0.5    1,0,0  1,0,0.5     none
				none     none 1,0,-0.5    1,0,0  1,0,0.5     none
				none     none 1,0,-0.5    1,0,0  1,0,0.5     none
			`,
		},
		{
			// the same edge, with three quarters of the edge pixel covered
			name:  "edge 0.75",
			clamp: true,
			source: `
				255 255 255 191 000 000
				255 255 255 191 000 000
				255 255 255 191 000 000
				255 255 255 191 000 000
			`,
			want: `
				none      none  1,0,-0.5 1,0,-0.25   1,0,0.5      none
				none      none  1,0,-0.5 1,0,-0.25   1,0,0.5      none
				none      none  1,0,-0.5 1,0,-0.25   1,0,0.5      none
				none      none  1,0,-0.5 1,0,-0.25   1,0,0.5      none
			`,
		},
		{
			// a 45 degree edge crossing the centers of the pixels at half alpha.
			// at the borders, Clamp continues it straight out of them, so it
			// bends there
			name:  "diagonal",
			clamp: true,
			source: `
				255 255 255 255 128 000 000
				255 255 255 128 000 000 000
				255 255 128 000 000 000 000
				255 128 000 000 000 000 000
				128 000 000 000 000 000 000
			`,
			want: `
				           none            none 0.71,0.71,-0.71 0.88,0.48,-0.68    0.92,0.38,0  0.92,0.38,0.66            none
				           none 0.71,0.71,-0.71 0.71,0.71,-0.71     0.71,0.71,0  0.71,0.71,0.71  0.71,0.71,0.71            none
				0.71,0.71,-0.71 0.71,0.71,-0.71     0.71,0.71,0  0.71,0.71,0.71  0.71,0.71,0.71            none            none
				0.48,0.88,-0.68     0.71,0.71,0  0.71,0.71,0.71  0.71,0.71,0.71            none            none            none
				    0.71,0.71,0  0.88,0.48,0.68  0.71,0.71,0.71            none            none            none            none
			`,
		},
		{
			// lines 1, 2 and 3 pixels thick. the pixels on the 1px line are
			// brighter than both sides, so they get no data
			name:  "lines",
			clamp: true,
			source: `
				000 000 000 000 000 000
				255 255 255 255 255 255
				000 000 000 000 000 000
				000 000 000 000 000 000
				000 000 000 000 000 000
				255 255 255 255 255 255
				255 255 255 255 255 255
				000 000 000 000 000 000
				000 000 000 000 000 000
				000 000 000 000 000 000
				255 255 255 255 255 255
				255 255 255 255 255 255
				255 255 255 255 255 255
				000 000 000 000 000 000
				000 000 000 000 000 000
			`,
			want: `
				 0,-1,0.5  0,-1,0.5  0,-1,0.5  0,-1,0.5  0,-1,0.5  0,-1,0.5
				     none      none      none      none      none      none
				  0,1,0.5   0,1,0.5   0,1,0.5   0,1,0.5   0,1,0.5   0,1,0.5
				     none      none      none      none      none      none
				 0,-1,0.5  0,-1,0.5  0,-1,0.5  0,-1,0.5  0,-1,0.5  0,-1,0.5
				0,-1,-0.5 0,-1,-0.5 0,-1,-0.5 0,-1,-0.5 0,-1,-0.5 0,-1,-0.5
				 0,1,-0.5  0,1,-0.5  0,1,-0.5  0,1,-0.5  0,1,-0.5  0,1,-0.5
				  0,1,0.5   0,1,0.5   0,1,0.5   0,1,0.5   0,1,0.5   0,1,0.5
				     none      none      none      none      none      none
				 0,-1,0.5  0,-1,0.5  0,-1,0.5  0,-1,0.5  0,-1,0.5  0,-1,0.5
				0,-1,-0.5 0,-1,-0.5 0,-1,-0.5 0,-1,-0.5 0,-1,-0.5 0,-1,-0.5
				     none      none      none      none      none      none
				 0,1,-0.5  0,1,-0.5  0,1,-0.5  0,1,-0.5  0,1,-0.5  0,1,-0.5
				  0,1,0.5   0,1,0.5   0,1,0.5   0,1,0.5   0,1,0.5   0,1,0.5
				     none      none      none      none      none      none
			`,
		},
		{
			// an edge where alpha changes by 3 steps, too little for data
			name:  "change of 3",
			clamp: true,
			source: `
				000 000 003 003
				000 000 003 003
				000 000 003 003
			`,
			want: `
				none none none none
				none none none none
				none none none none
			`,
		},
		{
			// the same edge changing by 4 steps, the smallest change with data
			name:  "change of 4",
			clamp: true,
			source: `
				000 000 004 004
				000 000 004 004
				000 000 004 004
			`,
			want: `
				none -1,0,0.5 -1,0,-0.5 none
				none -1,0,0.5 -1,0,-0.5 none
				none -1,0,0.5 -1,0,-0.5 none
			`,
		},
		{
			// without Clamp, the area outside counts as transparent, so the
			// borders are edges
			name:  "opaque",
			clamp: false,
			source: `
				255 255 255 255 255
				255 255 255 255 255
				255 255 255 255 255
				255 255 255 255 255
			`,
			want: `
				-0.71,-0.71,-0.71         0,-1,-0.5         0,-1,-0.5         0,-1,-0.5  0.71,-0.71,-0.71
				        -1,0,-0.5              none              none              none          1,0,-0.5
				        -1,0,-0.5              none              none              none          1,0,-0.5
				 -0.71,0.71,-0.71          0,1,-0.5          0,1,-0.5          0,1,-0.5   0.71,0.71,-0.71
			`,
		},
		{
			// with Clamp, the source continues past its borders, so there are
			// no edges
			name:  "opaque, clamp",
			clamp: true,
			source: `
				255 255 255 255 255
				255 255 255 255 255
				255 255 255 255 255
				255 255 255 255 255
			`,
			want: `
				none none none none none
				none none none none none
				none none none none none
				none none none none none
			`,
		},
	}

	report := pixcheck.Report{Rule: pixcheck.Rule{MaxDelta: 2}, FormatPixel: formatNormal}
	pixcheck.Run(t, func() {
		r := NewRenderer()
		for _, c := range cases {
			alpha := pixcheck.ParseAlpha(c.source)
			normals := ebiten.NewImage(alpha.Rect.Dx(), alpha.Rect.Dy())
			var flags []Flag
			if c.clamp {
				flags = append(flags, Clamp)
			}
			r.AlphaNormalsCompute(normals, pixcheck.NewImage(alpha, image.Pt(3, 5)), flags...)
			report.CompareRGBA(c.name, pixcheck.Read(normals), pixcheck.Parse(c.want, parseNormal))
		}
	})
	report.Check(t)
}

// parseNormal converts a normal written as "dx,dy,offset", or "none", into a
// pixel encoded like alpha_normals.kage does, for [pixcheck.Parse].
func parseNormal(text string) (color.RGBA, error) {
	if text == "none" {
		return color.RGBA{}, nil
	}
	fields := strings.Split(text, ",")
	if len(fields) != 3 {
		return color.RGBA{}, errors.New(`expected "dx,dy,offset" or "none"`)
	}
	var values [3]float64
	for i, field := range fields {
		value, err := strconv.ParseFloat(field, 64)
		if err != nil {
			return color.RGBA{}, err
		}
		values[i] = value
	}
	to8 := func(v float64) uint8 { return uint8(math.Round(clamp(v, 0, 1) * 255)) }
	return color.RGBA{to8(values[0]*0.5 + 0.5), to8(values[1]*0.5 + 0.5), to8(values[2]/1.5 + 0.5), 255}, nil
}

// formatNormal is the reverse of [parseNormal], for [pixcheck.Format], with the
// direction normalized and two decimals.
func formatNormal(pixel color.RGBA) string {
	if pixel.A == 0 {
		return "none"
	}
	dx, dy := float64(pixel.R)/255*2-1, float64(pixel.G)/255*2-1
	length := math.Hypot(dx, dy)
	format := func(v float64) string {
		return strconv.FormatFloat(math.Round(v*100)/100+0, 'f', -1, 64) // +0 turns -0 into 0
	}
	return format(dx/length) + "," + format(dy/length) + "," + format((float64(pixel.B)/255-0.5)*1.5)
}
