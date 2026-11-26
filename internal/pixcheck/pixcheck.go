// Package pixcheck compares images in tests, e.g. to check that two ways of
// drawing the same thing give the same pixels, or that an output stays above or
// below some expected values. Tests that read images back from Ebitengine must
// do it inside [Run].
package pixcheck

import (
	"fmt"
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

// Metric measures how far a pixel is from the expected one, from 0 (a match) to
// 255. Besides the predefined metrics, tests can define their own. The zero
// value is [MaxChannelDelta].
type Metric struct {
	Name  string
	Delta func(got, want color.RGBA) uint8
}

var (
	// MaxChannelDelta takes the largest difference in any channel.
	MaxChannelDelta = Metric{"max channel delta", func(got, want color.RGBA) uint8 {
		return max(absDiff(got.R, want.R), absDiff(got.G, want.G), absDiff(got.B, want.B), absDiff(got.A, want.A))
	}}

	// AlphaDelta only takes the difference in alpha.
	AlphaDelta = Metric{"alpha delta", func(got, want color.RGBA) uint8 {
		return absDiff(got.A, want.A)
	}}

	// Shortfall takes how far any channel falls below the expected one, so
	// pixels above the expected values match.
	Shortfall = Metric{"shortfall", func(got, want color.RGBA) uint8 {
		return max(below(got.R, want.R), below(got.G, want.G), below(got.B, want.B), below(got.A, want.A))
	}}

	// Excess takes how far any channel goes above the expected one, so pixels
	// below the expected values match.
	Excess = Metric{"excess", func(got, want color.RGBA) uint8 {
		return max(above(got.R, want.R), above(got.G, want.G), above(got.B, want.B), above(got.A, want.A))
	}}
)

// below returns how far value is below limit, or 0 if it isn't.
func below(value, limit uint8) uint8 {
	if value < limit {
		return limit - value
	}
	return 0
}

// above returns how far value is above limit, or 0 if it isn't.
func above(value, limit uint8) uint8 {
	return below(limit, value)
}

func (m Metric) String() string {
	if m.Delta == nil {
		return MaxChannelDelta.Name
	}
	return m.Name
}

// Compare measures how far got is from want. Both must have the same size, but
// can have different origins: pixels are compared by their position relative
// to the origin of each image.
func (m Metric) Compare(got, want *image.RGBA) Diff {
	assertSameSize(got, want)
	delta := m.deltaFunc()
	w, h := got.Rect.Dx(), got.Rect.Dy()
	diff := Diff{Width: w, Height: h}
	var maxDelta uint8
	for y := range h {
		for x := range w {
			gotPixel := got.RGBAAt(got.Rect.Min.X+x, got.Rect.Min.Y+y)
			wantPixel := want.RGBAAt(want.Rect.Min.X+x, want.Rect.Min.Y+y)
			pixelDelta := delta(gotPixel, wantPixel)
			diff.Histogram[pixelDelta]++
			if pixelDelta > maxDelta {
				maxDelta = pixelDelta
				diff.MaxAt = image.Pt(x, y)
			}
		}
	}
	return diff
}

func (m Metric) deltaFunc() func(got, want color.RGBA) uint8 {
	if m.Delta == nil {
		return MaxChannelDelta.Delta
	}
	return m.Delta
}

// Diff summarizes how far an image is from the expected one, as measured by a
// [Metric].
type Diff struct {
	Width, Height int

	// Histogram counts pixels by their delta: Histogram[0] is the number of
	// pixels that match, Histogram[255] the number of pixels as far as they can
	// be, e.g. a channel going all the way from 0 to 255 or the other way
	// around.
	Histogram [256]int

	// MaxAt is the position of the first pixel with the largest delta, in row
	// order, relative to the image origin.
	MaxAt image.Point
}

// MaxDelta returns the largest delta found.
func (d Diff) MaxDelta() uint8 {
	for delta := 255; delta > 0; delta-- {
		if d.Histogram[delta] > 0 {
			return uint8(delta)
		}
	}
	return 0
}

// Over returns how many pixels have a delta over maxDelta.
func (d Diff) Over(maxDelta uint8) int {
	var count int
	for _, n := range d.Histogram[int(maxDelta)+1:] {
		count += n
	}
	return count
}

func (d Diff) String() string {
	over := d.Over(0)
	if over == 0 {
		return "all pixels match"
	}
	return fmt.Sprintf(
		"%d of %d pixels don't match, max delta %d at (%d, %d)",
		over, d.Width*d.Height, d.MaxDelta(), d.MaxAt.X, d.MaxAt.Y,
	)
}

// Rule decides whether an image matches the expected one: Metric measures each
// pixel, and the images match if at most MaxOver pixels go over MaxDelta. The
// zero value requires identical images.
type Rule struct {
	Metric Metric

	// MaxDelta is the largest delta (out of 255) that still counts as a match.
	// A value of 1 absorbs the rounding differences that appear when floating
	// point operations get reordered.
	MaxDelta uint8

	// MaxOver is how many pixels can go over MaxDelta.
	MaxOver int
}

// Accepts reports whether d fits the rule. d must have been measured with the
// rule's metric.
func (r Rule) Accepts(d Diff) bool {
	return d.Over(r.MaxDelta) <= r.MaxOver
}

// Worse reports whether a is worse than b for the rule: more pixels over
// MaxDelta, or as many but with a larger delta. a and b must have been measured
// with the rule's metric.
func (r Rule) Worse(a, b Diff) bool {
	if overA, overB := a.Over(r.MaxDelta), b.Over(r.MaxDelta); overA != overB {
		return overA > overB
	}
	return a.MaxDelta() > b.MaxDelta()
}

func (r Rule) String() string {
	limits := "exact"
	if r.MaxDelta > 0 || r.MaxOver > 0 {
		limits = fmt.Sprintf("max delta %d, max over %d", r.MaxDelta, r.MaxOver)
	}
	return r.Metric.String() + ", " + limits
}

func absDiff(a, b uint8) uint8 {
	if a >= b {
		return a - b
	}
	return b - a
}

func assertSameSize(a, b *image.RGBA) {
	if !a.Rect.Size().Eq(b.Rect.Size()) {
		panic(fmt.Sprintf("pixcheck: can't compare images of different sizes (%v vs %v)", a.Rect.Size(), b.Rect.Size()))
	}
}

// NewImage copies img into a new Ebitengine image whose bounds start at origin.
// Origins other than (0, 0) help check that operations don't assume them.
func NewImage(img *image.RGBA, origin image.Point) *ebiten.Image {
	shifted := *img
	shifted.Rect = img.Rect.Sub(img.Rect.Min).Add(origin)
	return ebiten.NewImageFromImageWithOptions(&shifted, &ebiten.NewImageFromImageOptions{PreserveBounds: true})
}

// Read copies the pixels of img into a new image, with its origin at (0, 0). It
// must be called inside [Run].
func Read(img *ebiten.Image) *image.RGBA {
	bounds := img.Bounds()
	rgba := image.NewRGBA(image.Rect(0, 0, bounds.Dx(), bounds.Dy()))
	img.ReadPixels(rgba.Pix)
	return rgba
}
