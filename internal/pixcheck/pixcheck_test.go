package pixcheck

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"slices"
	"testing"
)

func TestCompare(t *testing.T) {
	// empty images match
	got := image.NewRGBA(image.Rect(0, 0, 8, 8))
	want := image.NewRGBA(image.Rect(0, 0, 8, 8))
	if diff := MaxChannelDelta.Compare(got, want); diff.Over(0) != 0 || diff.String() != "all pixels match" {
		t.Fatalf("empty images should match, got %s", diff)
	}

	// mismatches are counted by delta, and the max delta is found at the
	// first pixel with it
	want.Set(2, 3, color.RGBA{0, 0, 1, 1}) // delta 1
	want.Set(5, 1, color.RGBA{9, 0, 0, 9}) // delta 9
	got.Set(6, 6, color.RGBA{0, 9, 0, 9})  // delta 9
	diff := Metric{}.Compare(got, want)    // zero value, same as MaxChannelDelta
	if diff.Over(0) != 3 || diff.Over(1) != 2 || diff.Over(9) != 0 {
		t.Fatalf("expected 3, 2 and 0 pixels over 0, 1 and 9, got %d, %d and %d", diff.Over(0), diff.Over(1), diff.Over(9))
	}
	if diff.MaxDelta() != 9 || diff.MaxAt != image.Pt(5, 1) {
		t.Fatalf("expected max delta 9 at (5, 1), got %d at %v", diff.MaxDelta(), diff.MaxAt)
	}

	// rules allow up to MaxOver pixels over MaxDelta
	if !(Rule{MaxDelta: 1, MaxOver: 2}).Accepts(diff) || (Rule{MaxDelta: 1, MaxOver: 1}).Accepts(diff) {
		t.Fatalf("wrong rule checks for %s", diff)
	}

	// origins don't change the result
	moved := *got
	moved.Rect = got.Rect.Add(image.Pt(10, -3))
	if movedDiff := MaxChannelDelta.Compare(&moved, want); movedDiff != diff {
		t.Fatalf("expected the same result with got moved, got %s instead of %s", movedDiff, diff)
	}
}

func TestMetrics(t *testing.T) {
	cases := []struct {
		metric           Metric
		pixel, reference color.RGBA
		wantDelta        uint8
	}{
		// MaxChannelDelta: largest difference in any channel
		{MaxChannelDelta, color.RGBA{10, 20, 30, 40}, color.RGBA{10, 20, 30, 40}, 0},
		{MaxChannelDelta, color.RGBA{10, 20, 30, 40}, color.RGBA{15, 20, 30, 40}, 5},
		{MaxChannelDelta, color.RGBA{10, 20, 30, 40}, color.RGBA{0, 25, 30, 40}, 10},

		// AlphaDelta: difference in alpha only
		{AlphaDelta, color.RGBA{0, 0, 0, 40}, color.RGBA{90, 90, 90, 40}, 0},
		{AlphaDelta, color.RGBA{0, 0, 0, 40}, color.RGBA{0, 0, 0, 50}, 10},

		// Shortfall: only channels below the expected ones count
		{Shortfall, color.RGBA{10, 20, 30, 40}, color.RGBA{0, 0, 0, 0}, 0},
		{Shortfall, color.RGBA{10, 20, 30, 40}, color.RGBA{15, 0, 0, 0}, 5},
		{Shortfall, color.RGBA{10, 20, 30, 40}, color.RGBA{0, 0, 40, 30}, 10},

		// Excess: only channels above the expected ones count
		{Excess, color.RGBA{10, 20, 30, 40}, color.RGBA{90, 90, 90, 90}, 0},
		{Excess, color.RGBA{10, 20, 30, 40}, color.RGBA{5, 90, 90, 90}, 5},
		{Excess, color.RGBA{10, 20, 30, 40}, color.RGBA{90, 90, 20, 50}, 10},
	}
	for _, c := range cases {
		if delta := c.metric.Delta(c.pixel, c.reference); delta != c.wantDelta {
			t.Errorf("%s: expected delta %d for %v against %v, got %d", c.metric, c.wantDelta, c.pixel, c.reference, delta)
		}
	}
}

func TestReport(t *testing.T) {
	report := Report{Rule: Rule{MaxDelta: 1}}
	got := ParseAlpha(`000 000`)
	cases := []struct {
		label string
		want  *image.RGBA
	}{
		{"pass", ParseAlpha(`001 000`)},
		{"1 over, by 200", ParseAlpha(`200 000`)},
		{"match", ParseAlpha(`000 000`)},
		{"2 over, up to 3", ParseAlpha(`003 003`)},
		{"2 over, up to 5", ParseAlpha(`005 003`)},
	}
	for _, c := range cases {
		report.CompareRGBA(c.label, got, c.want)
	}

	// failures are listed in order
	var labels []string
	for _, result := range report.failed() {
		labels = append(labels, result.Label)
	}
	if expected := []string{"1 over, by 200", "2 over, up to 3", "2 over, up to 5"}; !slices.Equal(labels, expected) {
		t.Fatalf("expected failures %q, got %q", expected, labels)
	}

	// the worst failure has the most pixels over the max delta, then the
	// largest delta, and keeps its images
	if report.worst.Label != "2 over, up to 5" || report.worst.got != got || report.worst.want != cases[4].want {
		t.Fatalf("expected the worst failure to be %q with its images, got %q", "2 over, up to 5", report.worst.Label)
	}
}

func TestWindow(t *testing.T) {
	cases := []struct {
		size, centerAt image.Point
		wantWindow     image.Rectangle
	}{
		// small images are shown whole
		{image.Pt(7, 5), image.Pt(5, 1), image.Rect(0, 0, 7, 5)},

		// centered on the pixel
		{image.Pt(64, 64), image.Pt(32, 32), image.Rect(25, 25, 40, 40)},
		{image.Pt(64, 3), image.Pt(40, 1), image.Rect(33, 0, 48, 3)},

		// moved inside by the borders
		{image.Pt(64, 64), image.Pt(2, 62), image.Rect(0, 49, 15, 64)},
	}
	for _, c := range cases {
		if win := window(c.size, c.centerAt, 15); win != c.wantWindow {
			t.Errorf("size %v, centered at %v: expected window %v, got %v", c.size, c.centerAt, c.wantWindow, win)
		}
	}
}

func TestParseFormatAlpha(t *testing.T) {
	// indentation and blank lines are ignored
	img := ParseAlpha(`

		0 64 255
		  128 7   1

	`)
	if img.Rect != image.Rect(0, 0, 3, 2) || img.RGBAAt(1, 0) != (color.RGBA{64, 64, 64, 64}) {
		t.Fatalf("wrong image: %v, %v at (1, 0)", img.Rect, img.RGBAAt(1, 0))
	}

	// values are formatted aligned, with 3 digits
	const Expected = "000 064 255\n128 007 001"
	if text := FormatAlpha(img); text != Expected {
		t.Fatalf("expected:\n%s\ngot:\n%s", Expected, text)
	}

	// invalid text panics
	for _, invalid := range []string{"", "1 2\n3", "256", "-1", "0x10", "a"} {
		expectPanic(t, func() { ParseAlpha(invalid) }, "ParseAlpha(%q)", invalid)
	}
}

func TestAlphaStats(t *testing.T) {
	img := ParseAlpha(`
		000 000 000 000
		000 255 255 000
		000 000 255 000
	`)

	// area, in pixels
	if sum := AlphaSum(img); sum != 3 {
		t.Errorf("expected an alpha sum of 3, got %g", sum)
	}

	// center, relative to the image origin. pixel centers are (1.5, 1.5),
	// (2.5, 1.5) and (2.5, 2.5)
	if x, y := AlphaCentroid(img); math.Abs(x-6.5/3) > 1e-9 || math.Abs(y-5.5/3) > 1e-9 {
		t.Errorf("expected the centroid at (%g, %g), got (%g, %g)", 6.5/3, 5.5/3, x, y)
	}
	sub := img.SubImage(image.Rect(1, 1, 4, 3)).(*image.RGBA)
	if x, y := AlphaCentroid(sub); math.Abs(x-3.5/3) > 1e-9 || math.Abs(y-2.5/3) > 1e-9 {
		t.Errorf("expected the sub-image centroid at (%g, %g), got (%g, %g)", 3.5/3, 2.5/3, x, y)
	}

	// fully transparent images give (0, 0)
	if x, y := AlphaCentroid(image.NewRGBA(image.Rect(0, 0, 2, 2))); x != 0 || y != 0 {
		t.Errorf("expected (0, 0) for a transparent image, got (%g, %g)", x, y)
	}
}

// expectPanic fails t if fn doesn't panic. The description of the call is
// formatted with fmt.Sprintf.
func expectPanic(t *testing.T, fn func(), format string, args ...any) {
	t.Helper()
	defer func() {
		if recover() == nil {
			t.Errorf("expected a panic from %s", fmt.Sprintf(format, args...))
		}
	}()
	fn()
}
