package shapes

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"math"
	"strconv"
	"strings"
	"testing"

	"github.com/bstkhq/go-shapes/internal/pixcheck"
	"github.com/hajimehoshi/ebiten/v2"
)

// Each pixel of a jfmap must point to the seed whose boundary is nearest,
// measured from inside or outside the region depending on the pixel.
func TestJFMapCompute(t *testing.T) {
	// a seed 258 pixels below the top of a 1x260 column, with a max distance
	// of 257: every pixel points to it except the top one, which is too far
	longRange := image.NewRGBA(image.Rect(0, 0, 1, 260))
	for y := range 260 {
		longRange.SetRGBA(0, y, encodeJFMOffset(image.Pt(0, 258-y)))
	}
	longRange.SetRGBA(0, 0, jfmNone)

	cases := []struct {
		name    string
		compute func(r *Renderer) *ebiten.Image // returns the jfmap
		want    *image.RGBA
	}{
		{
			// with Clamp, the rows above and below continue the same, so only
			// the ends of the shape are seeds: x = 2 (alpha 64, its boundary
			// 0.25 outside its center) and x = 5 (alpha 255, its boundary 0.5
			// beyond its center). x = 4 is closer to the boundary at x = 5
			// (1 + 0.5) than to the one at x = 2 (2 - 0.25). bounds don't
			// start at (0, 0)
			name: "row",
			compute: func(r *Renderer) *ebiten.Image {
				source := pixcheck.NewImage(pixcheck.ParseAlpha(`
					000 000 064 255 255 255 000
				`), image.Pt(8, 8))
				size := source.Bounds().Size()
				jfmap := ebiten.NewImageWithOptions(image.Rectangle{Max: size}.Add(image.Pt(16, 16)), nil)
				r.JFMapCompute(jfmap, source, JFMapComputeOptions{MaxDistance: 8}, Clamp)
				return jfmap
			},
			want: pixcheck.Parse(`
				2,0 1,0 0,0 -1,0 1,0 0,0 -1,0
			`, parseJFMOffset),
		},
		{
			// offsets too large to fit in a single byte
			name: "long range",
			compute: func(r *Renderer) *ebiten.Image {
				source := ebiten.NewImage(1, 260)
				source.Set(0, 258, color.White)
				jfmap := ebiten.NewImage(1, 260)
				r.JFMapCompute(jfmap, source, JFMapComputeOptions{MaxDistance: 257})
				return jfmap
			},
			want: longRange,
		},
	}

	report := pixcheck.Report{FormatPixel: formatJFMOffset}
	pixcheck.Run(t, func() {
		r := NewRenderer()
		for _, c := range cases {
			report.CompareRGBA(c.name, pixcheck.Read(c.compute(r)), c.want)
		}
	})
	report.Check(t)
}

// ---- TestJFMapCompute helpers ----

// jfmNone is the value of jfmap pixels without any seed within the maximum
// distance.
var jfmNone = color.RGBA{255, 255, 255, 255}

// encodeJFMOffset encodes an offset to a seed like jfm_pass.kage does: 16 bits
// per axis, the highest one for the sign and the rest for the magnitude.
func encodeJFMOffset(offset image.Point) color.RGBA {
	x, y := encodeJFMAxis(offset.X), encodeJFMAxis(offset.Y)
	return color.RGBA{x[0], x[1], y[0], y[1]}
}

func encodeJFMAxis(value int) [2]uint8 {
	var sign uint8
	if value < 0 {
		sign, value = 0x80, -value
	}
	return [2]uint8{sign | uint8(value>>8), uint8(value)}
}

func decodeJFMOffset(pix color.RGBA) image.Point {
	return image.Pt(decodeJFMAxis(pix.R, pix.G), decodeJFMAxis(pix.B, pix.A))
}

func decodeJFMAxis(hi, lo uint8) int {
	magnitude := int(hi&0x7F)<<8 | int(lo)
	if hi&0x80 != 0 {
		return -magnitude
	}
	return magnitude
}

// parseJFMOffset converts an offset written as "x,y", or "none", into a jfmap
// pixel, for [pixcheck.Parse].
func parseJFMOffset(text string) (color.RGBA, error) {
	if text == "none" {
		return jfmNone, nil
	}
	xText, yText, found := strings.Cut(text, ",")
	x, errX := strconv.Atoi(xText)
	y, errY := strconv.Atoi(yText)
	if !found || errX != nil || errY != nil {
		return color.RGBA{}, errors.New(`expected "x,y" or "none"`)
	}
	return encodeJFMOffset(image.Pt(x, y)), nil
}

// formatJFMOffset is the reverse of [parseJFMOffset], for [pixcheck.Format].
func formatJFMOffset(pixel color.RGBA) string {
	if pixel == jfmNone {
		return "none"
	}
	offset := decodeJFMOffset(pixel)
	return fmt.Sprintf("%d,%d", offset.X, offset.Y)
}

// ---- TestJFMReadBestSeed ----

// The jfmap reads must give the same result as the best seed of the whole
// region, found by brute force, for every pixel and region (see jfmRegions).
// Seeds the map misses would show up as holes, which happened when the map and
// the reads measured distances differently.
//
// Note: the CPU copy uses the same formulas as the shaders, so this checks that
// the map finds the best seed, not that the formulas are right. For regions
// whose limits cut through mid alpha, expansions have a known limitation (see
// TestKnownLimitation_JFMExpandMidAlphaLimits).
func TestJFMReadBestSeed(t *testing.T) {
	configs := []jfmBestSeedConfig{
		{op: jfmOpExpand, normals: false},
		{op: jfmOpExpand, normals: true},
		{op: jfmOpErode, normals: false},
		{op: jfmOpErode, normals: true},
		{op: jfmOpOutline, normals: false},
		{op: jfmOpOutline, normals: true},
	}
	distances := []float32{0.5, 1.75, 3, 4.5}

	report := pixcheck.Report{Rule: pixcheck.Rule{MaxDelta: 2}}
	pixcheck.Run(t, func() {
		r := NewRenderer()
		for _, region := range jfmRegions {
			jfmap := JFMap{MinAlpha: region.minAlpha, MaxAlpha: region.maxAlpha}
			for _, shape := range []testShape{blobShape, cornerShape, fadeShape, thinShape} {
				for _, config := range configs {
					for _, distance := range distances {
						got, want := readWithBestSeed(r, shape, config, distance, jfmap)
						label := fmt.Sprintf("%s, %s, %+v, distance %g", region.name, shape.name, config, distance)
						report.CompareRGBA(label, got, want)
					}
				}
			}
		}
	})
	report.Check(t)
}

type jfmBestSeedConfig struct {
	op      int
	normals bool
}

// readWithBestSeed is a helper for TestJFMReadBestSeed. It returns the result
// of the operation drawn by the renderer (got), with a jfmap generated at draw
// time for the region of jfmap, and the result computed on the CPU with the
// best seed for each pixel (want). The results include the area the operations
// draw around the source, which is the margin of the generated maps:
// distance + 2 pixels for expansions, and thickness / 2 + 2 for outlines.
func readWithBestSeed(r *Renderer, shape testShape, config jfmBestSeedConfig, distance float32, jfmap JFMap) (got, want *image.RGBA) {
	source := pixcheck.NewImage(shape.img, image.Pt(3, 5))
	model := jfmModel{jfmap: jfmap, source: shape.img}
	if config.normals {
		jfmap.Normals = alphaNormals(r, source)
		model.normals = pixcheck.Read(jfmap.Normals)
	}

	var pad int
	switch config.op {
	case jfmOpExpand:
		pad = int(math.Ceil(float64(distance))) + 2
	case jfmOpOutline:
		pad = int(math.Ceil(float64(distance/2))) + 2
	}
	w, h := shape.img.Rect.Dx()+2*pad, shape.img.Rect.Dy()+2*pad
	target := ebiten.NewImage(w, h)
	switch config.op {
	case jfmOpExpand:
		r.JFMExpand(target, source, float32(pad), float32(pad), distance, jfmap)
	case jfmOpErode:
		r.JFMErode(target, source, 0, 0, distance, jfmap)
	case jfmOpOutline:
		r.JFMOutline(target, source, float32(pad), float32(pad), distance, jfmap)
	}

	model = model.padded(pad)
	want = image.NewRGBA(image.Rect(0, 0, w, h))
	for y := range h {
		for x := range w {
			want.SetRGBA(x, y, model.result(config.op, image.Pt(x, y), float64(distance)))
		}
	}
	return pixcheck.Read(target), want
}

// thinShape has features too thin for alpha normals (lines of 1 and 2 pixels)
// and faint pixels, where seeds get no normals data.
var thinShape = testShape{"thin", pixcheck.ParseAlpha(`
	000 000 000 000 000 000 000 000 000 000
	000 255 255 255 255 255 255 255 255 000
	000 000 000 000 000 000 000 000 000 000
	000 000 012 040 000 000 000 000 000 000
	000 000 000 000 000 255 128 000 000 000
	000 000 000 000 000 255 128 000 000 000
	000 024 000 000 000 255 128 000 000 000
	000 024 000 000 000 255 128 000 000 000
	000 000 000 000 000 000 000 000 000 000
	000 000 000 000 000 000 000 000 000 000
`)}

// ---- TestJFMGeneratedMap ----

// The jfmap operations must give the same result with a jfmap generated at draw
// time as with a precomputed one with plenty of range.
func TestJFMGeneratedMap(t *testing.T) {
	configs := []jfmGeneratedConfig{
		{op: jfmOpExpand, clamp: false, normals: false},
		{op: jfmOpExpand, clamp: false, normals: true},
		{op: jfmOpErode, clamp: false, normals: false},
		{op: jfmOpErode, clamp: false, normals: true},
		{op: jfmOpErode, clamp: true, normals: false},
		{op: jfmOpErode, clamp: true, normals: true},
		{op: jfmOpOutline, clamp: false, normals: false},
		{op: jfmOpOutline, clamp: false, normals: true},
		{op: jfmOpOutline, clamp: true, normals: false},
		{op: jfmOpOutline, clamp: true, normals: true},
	}
	distances := []float32{0, 0.5, 1, 1.75, 3, 4.5}

	report := pixcheck.Report{Rule: pixcheck.Rule{MaxDelta: 1}}
	pixcheck.Run(t, func() {
		r := NewRenderer()
		// the band is cut by the borders, so that Clamp matters
		for _, shape := range []testShape{blobShape, bandShape} {
			for _, config := range configs {
				for _, distance := range distances {
					got, want := readWithBothMaps(r, shape, config, distance)
					report.Compare(fmt.Sprintf("%s, %+v, distance %g", shape.name, config, distance), got, want)
				}
			}
		}
	})
	report.Check(t)
}

type jfmGeneratedConfig struct {
	op             int
	clamp, normals bool
}

// readWithBothMaps is a helper for TestJFMGeneratedMap. It draws the
// operation twice and returns both results: want uses a precomputed jfmap, and
// got one generated at draw time.
func readWithBothMaps(r *Renderer, shape testShape, config jfmGeneratedConfig, distance float32) (got, want *ebiten.Image) {
	// generated maps and their padding reach distance + 2, so this is plenty
	// for both the precomputed map and the margin around the results
	margin := int(math.Ceil(float64(distance))) + 8

	source := pixcheck.NewImage(shape.img, image.Pt(3, 5))
	w, h := shape.img.Rect.Dx()+2*margin, shape.img.Rect.Dy()+2*margin
	var flags []Flag
	if config.clamp {
		flags = append(flags, Clamp)
	}
	var normals, paddedNormals *ebiten.Image
	if config.normals {
		normals = alphaNormals(r, source, flags...)
		paddedNormals = padded(normals, margin)
	}
	draw := func(target, source *ebiten.Image, offset float32, jfmap JFMap, flags ...Flag) {
		switch config.op {
		case jfmOpExpand:
			r.JFMExpand(target, source, offset, offset, distance, jfmap)
		case jfmOpErode:
			r.JFMErode(target, source, offset, offset, distance, jfmap, flags...)
		case jfmOpOutline:
			r.JFMOutline(target, source, offset, offset, distance, jfmap, flags...)
		}
	}

	// the precomputed map is padded, which works for all the operations. with
	// Clamp, the inset makes clamping happen at the original borders
	paddedSource := padded(source, margin)
	computeOpts := JFMapComputeOptions{MaxDistance: margin, Inset: mapBool(config.clamp, 0, margin), Normals: paddedNormals}
	precomputed := r.JFMapCompute(ebiten.NewImage(w, h), paddedSource, computeOpts, flags...)
	want = ebiten.NewImage(w, h)
	draw(want, paddedSource, 0, precomputed)

	generated := JFMap{Normals: normals}
	got = ebiten.NewImage(w, h)
	draw(got, source, float32(margin), generated, flags...)
	return got, want
}

// ---- TestJFMNormalsFallback ----

// Seeds fall back to the estimate from their alpha when they have no normals
// data, and when their line puts the pixel being measured on the wrong side of
// the boundary. So normals without any data must give the same results as no
// normals at all. Normals pointing the wrong way must too when expanding, as
// all their lines put the pixels on the wrong side. The other operations also
// measure the pixels that are their own seed, which use their line as is.
func TestJFMNormalsFallback(t *testing.T) {
	const Pad, Offset = 8, 8.25 // fractional, to also measure from off-center points

	ops := []struct {
		name string
		draw func(r *Renderer, target, source *ebiten.Image, jfmap JFMap)
	}{
		{"JFMExpand", func(r *Renderer, target, source *ebiten.Image, jfmap JFMap) {
			r.JFMExpand(target, source, Offset, Offset, 3, jfmap)
		}},
		// small amounts make the pixels that are their own seed matter
		{"JFMErode 0.5", func(r *Renderer, target, source *ebiten.Image, jfmap JFMap) {
			r.JFMErode(target, source, Offset, Offset, 0.5, jfmap)
		}},
		{"JFMErode 2.5", func(r *Renderer, target, source *ebiten.Image, jfmap JFMap) {
			r.JFMErode(target, source, Offset, Offset, 2.5, jfmap)
		}},
		{"JFMOutline 1", func(r *Renderer, target, source *ebiten.Image, jfmap JFMap) {
			r.JFMOutline(target, source, Offset, Offset, 1, jfmap)
		}},
		{"JFMOutline 4", func(r *Renderer, target, source *ebiten.Image, jfmap JFMap) {
			r.JFMOutline(target, source, Offset, Offset, 4, jfmap)
		}},
		{"JFMapHeat outside", func(r *Renderer, target, source *ebiten.Image, jfmap JFMap) {
			r.JFMapHeat(target, source, Offset, Offset, 6, false, jfmap)
		}},
		{"JFMapHeat inside", func(r *Renderer, target, source *ebiten.Image, jfmap JFMap) {
			r.JFMapHeat(target, source, Offset, Offset, 6, true, jfmap)
		}},
		{"JFMapViz", func(r *Renderer, target, source *ebiten.Image, jfmap JFMap) {
			r.JFMapViz(target, source, Offset, Offset, 6, jfmap)
		}},
	}

	var report pixcheck.Report
	pixcheck.Run(t, func() {
		r := NewRenderer()
		for _, shape := range []testShape{blobShape, thinShape} {
			for _, op := range ops {
				source := pixcheck.NewImage(shape.img, image.Pt(3, 5))
				w, h := shape.img.Rect.Dx()+2*Pad, shape.img.Rect.Dy()+2*Pad
				want := ebiten.NewImage(w, h)
				op.draw(r, want, source, JFMap{})

				emptyNormals := JFMap{Normals: ebiten.NewImage(shape.img.Rect.Dx(), shape.img.Rect.Dy())}
				got := ebiten.NewImage(w, h)
				op.draw(r, got, source, emptyNormals)
				report.Compare(shape.name+", "+op.name+", empty normals", got, want)
			}
		}

		// the same lines, with the high and low alpha sides swapped. only on
		// shapes without thin or faint features, whose lines can be wrong for
		// some pixels already, and right for them once swapped
		for _, shape := range []testShape{blobShape, fadeShape, cornerShape} {
			source := pixcheck.NewImage(shape.img, image.Pt(3, 5))
			w, h := shape.img.Rect.Dx()+2*Pad, shape.img.Rect.Dy()+2*Pad
			want := ebiten.NewImage(w, h)
			r.JFMExpand(want, source, Offset, Offset, 3, JFMap{})

			wrongNormals := JFMap{Normals: pixcheck.NewImage(invertedNormals(pixcheck.Read(alphaNormals(r, source))), image.Point{})}
			got := ebiten.NewImage(w, h)
			r.JFMExpand(got, source, Offset, Offset, 3, wrongNormals)
			report.Compare(shape.name+", JFMExpand, wrong normals", got, want)
		}
	})
	report.Check(t)
}

// invertedNormals returns normals with their direction and offset negated, for
// TestJFMNormalsFallback. Each line stays in place, but with its high and low
// alpha sides swapped.
func invertedNormals(normals *image.RGBA) *image.RGBA {
	inverted := image.NewRGBA(normals.Rect)
	for y := normals.Rect.Min.Y; y < normals.Rect.Max.Y; y++ {
		for x := normals.Rect.Min.X; x < normals.Rect.Max.X; x++ {
			if pixel := normals.RGBAAt(x, y); pixel.A != 0 {
				inverted.SetRGBA(x, y, color.RGBA{255 - pixel.R, 255 - pixel.G, 255 - pixel.B, pixel.A})
			}
		}
	}
	return inverted
}

// ---- TestJFMErodeBorders ----

// Without Clamp, JFMErode erodes from the source borders as if the area outside
// were transparent. With Clamp, the borders don't erode at all.
func TestJFMErodeBorders(t *testing.T) {
	// fully opaque, so the borders are the only thing that can erode it
	opaque := pixcheck.ParseAlpha(`
		255 255 255 255 255 255 255 255
		255 255 255 255 255 255 255 255
		255 255 255 255 255 255 255 255
		255 255 255 255 255 255 255 255
		255 255 255 255 255 255 255 255
		255 255 255 255 255 255 255 255
	`)

	// pixel centers are 0.5, 1.5 and 2.5 pixels away from the borders. at
	// those distances, the alpha left is clamp(0.5 + distance - radius)
	cases := []struct {
		name     string
		distance float32
		clamp    bool
		want     *image.RGBA
	}{
		{"radius 0", 0, false, opaque},
		{"radius 0.75", 0.75, false, pixcheck.ParseAlpha(`
			064 064 064 064 064 064 064 064
			064 255 255 255 255 255 255 064
			064 255 255 255 255 255 255 064
			064 255 255 255 255 255 255 064
			064 255 255 255 255 255 255 064
			064 064 064 064 064 064 064 064
		`)},
		{"radius 1.25", 1.25, false, pixcheck.ParseAlpha(`
			000 000 000 000 000 000 000 000
			000 191 191 191 191 191 191 000
			000 191 255 255 255 255 191 000
			000 191 255 255 255 255 191 000
			000 191 191 191 191 191 191 000
			000 000 000 000 000 000 000 000
		`)},
		{"radius 2", 2, false, pixcheck.ParseAlpha(`
			000 000 000 000 000 000 000 000
			000 000 000 000 000 000 000 000
			000 000 255 255 255 255 000 000
			000 000 255 255 255 255 000 000
			000 000 000 000 000 000 000 000
			000 000 000 000 000 000 000 000
		`)},
		{"radius 2, clamp", 2, true, opaque},
	}

	report := pixcheck.Report{Rule: pixcheck.Rule{MaxDelta: 1}}
	// with normals, the borders are measured with the lines of the seeds on
	// them. the 4 corner pixels see a diagonal edge, so the corners are cut
	// like a 45 degree chamfer instead of staying square
	reportNormals := pixcheck.Report{Rule: pixcheck.Rule{MaxDelta: 1, MaxOver: 4}}
	pixcheck.Run(t, func() {
		r := NewRenderer()
		source := pixcheck.NewImage(opaque, image.Pt(3, 5))
		for _, c := range cases {
			var flags []Flag
			if c.clamp {
				flags = append(flags, Clamp)
			}
			jfmap := JFMap{}
			got := ebiten.NewImage(source.Bounds().Dx(), source.Bounds().Dy())
			r.JFMErode(got, source, 0, 0, c.distance, jfmap, flags...)
			report.CompareRGBA(c.name, pixcheck.Read(got), c.want)

			jfmap.Normals = alphaNormals(r, source, flags...)
			gotNormals := ebiten.NewImage(source.Bounds().Dx(), source.Bounds().Dy())
			r.JFMErode(gotNormals, source, 0, 0, c.distance, jfmap, flags...)
			reportNormals.CompareRGBA(c.name+", normals", pixcheck.Read(gotNormals), c.want)
		}
	})
	report.Check(t)
	reportNormals.Check(t)
}

// ---- TestJFMErodeSmooth ----

// JFMErode must start from the source and shrink it steadily as the distance
// grows, without jumps at small distances, with and without normals.
func TestJFMErodeSmooth(t *testing.T) {
	const Step = 0.125

	// each step erodes a band of the same width, so the area lost per step can
	// only change gradually: its ratio to the area lost in the previous step
	// must stay within the range measured for each case, widened by 0.05. the
	// blob varies more, as distances to small shapes come in fewer distinct
	// values, especially without normals
	cases := []struct {
		name               string
		blob, normals      bool
		minRatio, maxRatio float64
	}{
		{"circle", false, false, 0.85, 1.26},
		{"circle, normals", false, true, 0.85, 1.07},
		{"blob", true, false, 0.54, 1.21},
		{"blob, normals", true, true, 0.65, 1.26},
	}

	pixcheck.Run(t, func() {
		r := NewRenderer()
		circle := ebiten.NewImage(32, 32)
		r.FillCircle(circle, 16.3, 15.6, 11)
		blob := pixcheck.NewImage(blobShape.img, image.Pt(3, 5))
		for _, c := range cases {
			source := mapBool(c.blob, circle, blob)
			if areas, ok := erodedAreas(t, r, source, c.normals, Step, c.name); ok {
				checkSteadyLoss(t, areas, Step, c.minRatio, c.maxRatio, c.name)
			}
		}
	})
}

// erodedAreas is a helper for TestJFMErodeSmooth. It returns the areas left
// after eroding source by 0, step, 2*step, ... 16*step, and fails t if erosion
// by 0 doesn't return the source.
func erodedAreas(t *testing.T, r *Renderer, source *ebiten.Image, withNormals bool, step float32, label string) ([]float64, bool) {
	var normals *ebiten.Image
	if withNormals {
		normals = alphaNormals(r, source)
	}
	w, h := source.Bounds().Dx(), source.Bounds().Dy()
	var areas []float64
	for i := range 17 {
		target := ebiten.NewImage(w, h)
		jfmap := JFMap{Normals: normals}
		r.JFMErode(target, source, 0, 0, float32(i)*step, jfmap)
		eroded := pixcheck.Read(target)
		if i == 0 {
			if diff := pixcheck.MaxChannelDelta.Compare(eroded, pixcheck.Read(source)); diff.Over(1) > 0 {
				t.Errorf("%s: erosion by 0 changes it: %s", label, diff)
				return nil, false
			}
		}
		areas = append(areas, pixcheck.AlphaSum(eroded))
	}
	return areas, true
}

// checkSteadyLoss is a helper for TestJFMErodeSmooth. It fails t if a step
// loses no area, or an area outside [minRatio, maxRatio] times the previous
// step's.
func checkSteadyLoss(t *testing.T, areas []float64, step float32, minRatio, maxRatio float64, label string) {
	for i := 1; i+1 < len(areas); i++ {
		lost, nextLost := areas[i-1]-areas[i], areas[i]-areas[i+1]
		if lost <= 0 || nextLost < lost*minRatio || nextLost > lost*maxRatio {
			t.Errorf("%s: area lost per step goes from %.2f to %.2f at distance %g (areas %.2f)", label, lost, nextLost, float32(i)*step, areas)
			return
		}
	}
}

// ---- TestJFMExpandKeepsSource ----

// JFMExpand must keep every pixel of the region at least as opaque as it was,
// growing by the distance, also at fractional positions.
func TestJFMExpandKeepsSource(t *testing.T) {
	configs := []jfmKeepsSourceConfig{
		{offset: 0, normals: false},
		{offset: 0.25, normals: false},
		{offset: 0.75, normals: false},
		{offset: 0, normals: true},
		{offset: 0.25, normals: true},
		{offset: 0.75, normals: true},
	}
	distances := []float32{0, 0.5, 1, 2}

	report := pixcheck.Report{Rule: pixcheck.Rule{Metric: pixcheck.Shortfall, MaxDelta: 1}}
	pixcheck.Run(t, func() {
		r := NewRenderer()
		for _, shape := range []testShape{fadeShape, blobShape} {
			for _, config := range configs {
				for _, distance := range distances {
					got, want := expandOverSource(r, shape, config, distance)
					report.CompareRGBA(fmt.Sprintf("%s, %+v, distance %g", shape.name, config, distance), got, want)
				}
			}
		}
	})
	report.Check(t)
}

// expandOverSource is a helper for TestJFMExpandKeepsSource. It returns the
// lowest result allowed (want), and the expansion of shape drawn at
// config.offset (got).
func expandOverSource(r *Renderer, shape testShape, config jfmKeepsSourceConfig, distance float32) (got, want *image.RGBA) {
	const Margin = 4 // more than the expansions reach
	w, h := shape.img.Rect.Dx()+2*Margin, shape.img.Rect.Dy()+2*Margin
	origin := Margin + config.offset

	source := pixcheck.NewImage(shape.img, image.Pt(3, 5))
	var normals *ebiten.Image
	if config.normals {
		normals = alphaNormals(r, source)
	}
	target := ebiten.NewImage(w, h)
	jfmap := JFMap{Normals: normals}
	r.JFMExpand(target, source, origin, origin, distance, jfmap)
	return pixcheck.Read(target), expansionLowerBound(shape, origin, distance, w, h)
}

type jfmKeepsSourceConfig struct {
	offset  float32
	normals bool
}

// expansionLowerBound returns the lowest result allowed for an expansion of
// shape by distance, drawn at (origin, origin) onto a w x h target: each pixel
// takes the alpha of the source pixel it shows, plus the distance.
func expansionLowerBound(shape testShape, origin, distance float32, w, h int) *image.RGBA {
	lowest := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := range h {
		for x := range w {
			// the source pixel whose area contains this pixel's center
			sx := int(math.Floor(float64(float32(x) + 0.5 - origin)))
			sy := int(math.Floor(float64(float32(y) + 0.5 - origin)))
			if !image.Pt(sx, sy).In(shape.img.Rect) {
				continue
			}
			if alpha := float32(shape.img.RGBAAt(sx, sy).A); alpha > 0 {
				v := uint8(min(255, math.Round(float64(alpha+255*distance))))
				lowest.SetRGBA(x, y, color.RGBA{v, v, v, v})
			}
		}
	}
	return lowest
}

// ---- TestJFMExpandMotion ----

// JFMExpand must follow fractional positions smoothly, with and without
// normals: the result moves with the offset, and its area stays the same.
func TestJFMExpandMotion(t *testing.T) {
	const Distance = 3

	pixcheck.Run(t, func() {
		r := NewRenderer()
		circle := ebiten.NewImage(24, 24)
		r.FillCircle(circle, 12, 12, 7)
		for _, normals := range []*ebiten.Image{nil, alphaNormals(r, circle)} {
			measure := func(offset float32) (area, x, y float64) {
				target := ebiten.NewImage(40, 40)
				jfmap := JFMap{Normals: normals}
				r.JFMExpand(target, circle, 8+offset, 8+offset*0.5, Distance, jfmap)
				result := pixcheck.Read(target)
				x, y = pixcheck.AlphaCentroid(result)
				return pixcheck.AlphaSum(result), x, y
			}
			baseArea, baseX, baseY := measure(0)
			for i := 1; i <= 10; i++ {
				offset := float64(i) * 0.1
				area, x, y := measure(float32(offset))
				if math.Abs(x-baseX-offset) > 0.05 || math.Abs(y-baseY-offset*0.5) > 0.05 {
					t.Errorf("normals %t, offset (%.1f, %.2f): the result moved by (%.3f, %.3f)", normals != nil, offset, offset*0.5, x-baseX, y-baseY)
				}
				if math.Abs(area-baseArea) > baseArea*0.005 {
					t.Errorf("normals %t, offset (%.1f, %.2f): the area changed from %.2f to %.2f", normals != nil, offset, offset*0.5, baseArea, area)
				}
			}
		}
	})
}

// ---- TestJFMExpandAccuracy ----

// Expanding a circle must give about the same as drawing the circle with the
// radius grown by the distance. Unlike the other tests, which check that the
// operations agree with each other or with values worked out by hand, this one
// checks them against shapes drawn independently.
func TestJFMExpandAccuracy(t *testing.T) {
	// measured: the worst pixels are off by 84 without normals and 42 with
	// them, and the areas by up to 1.8% and 0.8%
	reports := map[bool]*pixcheck.Report{
		false: {Rule: pixcheck.Rule{Metric: pixcheck.AlphaDelta, MaxDelta: 100}},
		true:  {Rule: pixcheck.Rule{Metric: pixcheck.AlphaDelta, MaxDelta: 51}},
	}
	maxAreaError := map[bool]float64{false: 0.025, true: 0.012}

	pixcheck.Run(t, func() {
		r := NewRenderer()
		for _, withNormals := range []bool{false, true} {
			for _, radius := range []float32{6.2, 17.6} {
				for _, distance := range []float32{0.5, 1.5, 3.25, 6} {
					got, want := expandCircle(r, radius, distance, withNormals)
					label := fmt.Sprintf("radius %g, distance %g, normals %t", radius, distance, withNormals)
					reports[withNormals].CompareRGBA(label, got, want)
					gotArea, wantArea := pixcheck.AlphaSum(got), pixcheck.AlphaSum(want)
					if math.Abs(gotArea-wantArea) > wantArea*maxAreaError[withNormals] {
						t.Errorf("%s: area %.2f instead of %.2f", label, gotArea, wantArea)
					}
				}
			}
		}
	})
	reports[false].Check(t)
	reports[true].Check(t)
}

// expandCircle is a helper for TestJFMExpandAccuracy. It returns a circle
// expanded by distance (got), and the circle drawn with the radius grown by
// distance (want).
func expandCircle(r *Renderer, radius, distance float32, withNormals bool) (got, want *image.RGBA) {
	const Size, X, Y = 64, 32.3, 31.6

	circle := ebiten.NewImage(Size, Size)
	r.FillCircle(circle, X, Y, radius)
	jfmap := JFMap{}
	if withNormals {
		jfmap.Normals = alphaNormals(r, circle)
	}
	expanded := ebiten.NewImage(Size, Size)
	r.JFMExpand(expanded, circle, 0, 0, distance, jfmap)

	grown := ebiten.NewImage(Size, Size)
	r.FillCircle(grown, X, Y, radius+distance)
	return pixcheck.Read(expanded), pixcheck.Read(grown)
}

// ---- TestKnownLimitation_JFMExpandMidAlphaLimits ----

// Known limitation: JFMExpand misplaces the boundaries of regions whose limits
// cut through mid alpha, like a threshold or a band over a fade, by up to
// about a pixel.
//
// Inside a seed, the boundary is placed by reading the seed's alpha as the
// coverage of a sharp antialiased edge. That's exact on the edges of shapes,
// but at a mid-alpha limit on a fade, the seed's alpha doesn't tell where
// alpha crosses the limit. In this test, where alpha rises by 20/255 per
// pixel, the boundaries end up 0.5-0.7 pixels inside the region, so the pixels
// next to them get far less coverage than expected.
//
// Fixing it would need to store, for each seed, where alpha crosses the limit,
// which the map has no room for if large textures have to be supported.
//
// The test records the current output, so that any change shows up and has
// to be updated on purpose.
func TestKnownLimitation_JFMExpandMidAlphaLimits(t *testing.T) {
	// a ramp from 0 to 1. the map is computed with Clamp, so the rows above
	// and below continue the same, and the region limits are the only
	// boundaries
	ramp := pixcheck.ParseAlpha(`
		000 020 040 060 080 100 120 140 160 180 200 220 240 255
	`)
	threshold := JFMap{MinAlpha: 0.5, MaxAlpha: 1}
	band := JFMap{MinAlpha: 0.2, MaxAlpha: 0.8}
	lowSide := JFMap{MinAlpha: 0, MaxAlpha: 0.5}
	cases := []struct {
		name     string
		region   JFMap
		distance float32
		want     string
	}{
		{"threshold [0.5, 1], distance 0", threshold, 0, `
		000 000 000 000 000 000 000 140 160 180 200 220 240 255`},
		{"threshold [0.5, 1], distance 0.5", threshold, 0.5, `
		000 000 000 000 000 000 012 255 255 255 255 255 255 255`},
		{"band [0.2, 0.8], distance 0", band, 0, `
		000 000 000 060 080 100 120 115 095 075 055 000 000 000`},
		{"band [0.2, 0.8], distance 0.5", band, 0.5, `
		000 000 000 188 207 227 248 243 223 202 183 000 000 000`},
		{"low side [0, 0.5], distance 0", lowSide, 0, `
		255 235 215 195 175 155 135 000 000 000 000 000 000 000`},
		{"low side [0, 0.5], distance 0.5", lowSide, 0.5, `
		255 255 255 255 255 255 255 007 000 000 000 000 000 000`},
	}

	// small differences between GPUs and drivers are allowed
	report := pixcheck.Report{Rule: pixcheck.Rule{MaxDelta: 2}}
	pixcheck.Run(t, func() {
		r := NewRenderer()
		source := pixcheck.NewImage(ramp, image.Pt(3, 5))
		for _, c := range cases {
			computeOpts := JFMapComputeOptions{MaxDistance: 8, MinAlpha: c.region.MinAlpha, MaxAlpha: c.region.MaxAlpha}
			jfmap := r.JFMapCompute(ebiten.NewImage(ramp.Rect.Dx(), ramp.Rect.Dy()), source, computeOpts, Clamp)
			target := ebiten.NewImage(ramp.Rect.Dx(), ramp.Rect.Dy())
			r.JFMExpand(target, source, 0, 0, c.distance, jfmap)
			report.CompareRGBA(c.name, pixcheck.Read(target), pixcheck.ParseAlpha(c.want))
		}
	})
	report.Check(t)
}
