package shapes

import (
	"fmt"
	"image"
	"slices"
	"strings"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/bstkhq/go-shapes/internal/pixcheck"
)

func TestWarnings(t *testing.T) {
	tests := []struct {
		Warnings Warning
		Expected []Warning
	}{
		{Warnings: 0b0000_0000, Expected: []Warning{}},
		{Warnings: 0b0100_1000, Expected: []Warning{0b0000_1000, 0b0100_0000}},
		{Warnings: 0b1000_0011, Expected: []Warning{0b0000_0001, 0b0000_0010, 0b1000_0000}},
	}

	for i, test := range tests {
		found := slices.Collect(allWarnings(test.Warnings))
		if !slices.Equal(found, test.Expected) {
			t.Fatalf("on test#%d (0x%08x), expected %v, found %v", i, test.Warnings, test.Expected, found)
		}
	}
}

func TestWarningMessages(t *testing.T) {
	var w Warning = 1
	for w < warnSentinel {
		if strings.HasPrefix(w.Message(), unknownWarningPrefix) {
			t.Fatalf("missing warning message: %s", w.Message())
		}
		w <<= 1
	}
}

func TestReadOptInFlag(t *testing.T) {
	const OptIn = Clamp
	cases := []struct {
		flags    []Flag
		want     Flag
		warnings []flagWarning
	}{
		// empty flags
		{nil, noFlag, nil},
		{[]Flag{}, noFlag, nil},

		// noFlag is ignored
		{[]Flag{noFlag}, noFlag, nil},

		// opt-in flag
		{[]Flag{Clamp}, Clamp, nil},
		{[]Flag{noFlag, Clamp, noFlag}, Clamp, nil},

		// repeated flags
		{[]Flag{Clamp, Clamp}, Clamp, []flagWarning{{WarnRepeatedFlag, Clamp}}},
		{[]Flag{Clamp, Clamp, Clamp}, Clamp, []flagWarning{{WarnRepeatedFlag, Clamp}, {WarnRepeatedFlag, Clamp}}},

		// invalid flags
		{[]Flag{Bilinear}, noFlag, []flagWarning{{WarnInvalidFlag, Bilinear}}},
		{[]Flag{Bilinear, Clamp}, Clamp, []flagWarning{{WarnInvalidFlag, Bilinear}}},

		// mixed
		{[]Flag{Clamp, Bilinear, Clamp}, Clamp, []flagWarning{{WarnInvalidFlag, Bilinear}, {WarnRepeatedFlag, Clamp}}},
	}

	var r Renderer
	var warnings []flagWarning
	r.Warnings.SetHandler(func(warning Warning, value any, _ bool) {
		warnings = append(warnings, flagWarning{warning, value.(Flag)})
	})
	for _, c := range cases {
		got := r.readOptInFlag(OptIn, c.flags...)
		if got != c.want || !slices.Equal(warnings, c.warnings) {
			t.Errorf("flags %v: got %v with warnings %v, want %v with warnings %v", c.flags, got, warnings, c.want, c.warnings)
		}
		r.Warnings.Reset()
		warnings = nil
	}
}

func TestReadOptInFlagPair(t *testing.T) {
	const OptInA, OptInB = Bilinear, Dithered
	cases := []struct {
		flags        []Flag
		wantA, wantB Flag
		warnings     []flagWarning
	}{
		// empty flags
		{nil, noFlag, noFlag, nil},
		{[]Flag{}, noFlag, noFlag, nil},

		// noFlag is ignored
		{[]Flag{noFlag}, noFlag, noFlag, nil},

		// opt-in flags
		{[]Flag{Bilinear}, Bilinear, noFlag, nil},
		{[]Flag{Dithered}, noFlag, Dithered, nil},
		{[]Flag{noFlag, Dithered}, noFlag, Dithered, nil},
		{[]Flag{Dithered, Bilinear}, Bilinear, Dithered, nil},

		// repeated flags
		{[]Flag{Bilinear, Bilinear}, Bilinear, noFlag, []flagWarning{{WarnRepeatedFlag, Bilinear}}},
		{[]Flag{Dithered, Bilinear, Dithered}, Bilinear, Dithered, []flagWarning{{WarnRepeatedFlag, Dithered}}},

		// invalid flags
		{[]Flag{Clamp}, noFlag, noFlag, []flagWarning{{WarnInvalidFlag, Clamp}}},

		// mixed
		{[]Flag{Bilinear, Clamp, Bilinear}, Bilinear, noFlag, []flagWarning{{WarnInvalidFlag, Clamp}, {WarnRepeatedFlag, Bilinear}}},
	}

	var r Renderer
	var warnings []flagWarning
	r.Warnings.SetHandler(func(warning Warning, value any, _ bool) {
		warnings = append(warnings, flagWarning{warning, value.(Flag)})
	})
	for _, c := range cases {
		gotA, gotB := r.readOptInFlagPair(OptInA, OptInB, c.flags...)
		if gotA != c.wantA || gotB != c.wantB || !slices.Equal(warnings, c.warnings) {
			t.Errorf("flags %v: got %v and %v with warnings %v, want %v and %v with warnings %v",
				c.flags, gotA, gotB, warnings, c.wantA, c.wantB, c.warnings)
		}
		r.Warnings.Reset()
		warnings = nil
	}
}

// helper for flag reading tests
type flagWarning struct {
	warning Warning
	flag    Flag
}

// Operations skip missing (nil) and empty source images: nil ones report
// WarnMissingSourceOpSkipped, and empty ones return quietly, without drawing.
// Operations that compute data for later use panic on nil sources instead, as
// skipping would leave their target with stale data.
func TestSkippedSources(t *testing.T) {
	quad := [4]PointF32{{0, 0}, {20, 0}, {20, 20}, {0, 20}}
	dither := []float32{0, 0, 0, 1, 1, 1, 1, 1}
	// each call draws onto target, with img in the position being checked.
	// operations that require matching sizes get a target of img's size
	ops := []struct {
		name string
		draw func(r *Renderer, target, img, full *ebiten.Image)
	}{
		{"Scale", func(r *Renderer, tg, img, _ *ebiten.Image) { r.Scale(tg, img, 0, 0, 1, nil) }},
		{"DrawAt", func(r *Renderer, tg, img, _ *ebiten.Image) { r.DrawAt(tg, img, 0, 0, 1) }},
		{"DrawImgShader", func(r *Renderer, tg, img, _ *ebiten.Image) {
			r.DrawImgShader(tg, img, 0, 0, NoMargins, RegionExact, shaderDefault.Load())
		}},
		{"AlphaNormalsCompute", func(r *Renderer, tg, img, _ *ebiten.Image) { r.AlphaNormalsCompute(sizedLike(tg, img), img) }},
		{"AlphaNormalsViz", func(r *Renderer, tg, img, _ *ebiten.Image) { r.AlphaNormalsViz(tg, img, 0, 0) }},
		{"OklabShift", func(r *Renderer, tg, img, _ *ebiten.Image) { r.OklabShift(tg, img, 0, 0, 0, 0, 0) }},
		{"ColorizeByLightness", func(r *Renderer, tg, img, _ *ebiten.Image) {
			r.ColorizeByLightness(tg, img, GradientOptions{}, 0, 0, 0, 1)
		}},
		{"ColorMix base", func(r *Renderer, tg, img, full *ebiten.Image) {
			r.ColorMix(tg, img, sizedLike(full, img), 0, 0, 1, 0.5)
		}},
		{"ColorMix over", func(r *Renderer, tg, img, full *ebiten.Image) {
			r.ColorMix(tg, sizedLike(full, img), img, 0, 0, 1, 0.5)
		}},
		{"DitherMat4", func(r *Renderer, tg, img, _ *ebiten.Image) {
			r.DitherMat4(tg, img, 0, 0, 0, 0, dither, [16]float32{}, 0, 0)
		}},
		{"MapQuad4", func(r *Renderer, tg, img, _ *ebiten.Image) { r.MapQuad4(tg, img, quad, MapOptions{}) }},
		{"MapQuad", func(r *Renderer, tg, img, _ *ebiten.Image) { r.MapQuad(tg, img, quad, MapOptions{}) }},
		{"Mask source", func(r *Renderer, tg, img, full *ebiten.Image) { r.Mask(tg, img, full, 0, 0) }},
		{"Mask mask", func(r *Renderer, tg, img, full *ebiten.Image) { r.Mask(tg, full, img, 0, 0) }},
		{"MaskAt source", func(r *Renderer, tg, img, full *ebiten.Image) { r.MaskAt(tg, img, full, 0, 0, 0, 0) }},
		{"MaskAt mask", func(r *Renderer, tg, img, full *ebiten.Image) { r.MaskAt(tg, full, img, 0, 0, 0, 0) }},
		{"MaskThreshold source", func(r *Renderer, tg, img, full *ebiten.Image) { r.MaskThreshold(tg, img, full, 0.5, 0, 0) }},
		{"MaskThreshold mask", func(r *Renderer, tg, img, full *ebiten.Image) { r.MaskThreshold(tg, full, img, 0.5, 0, 0) }},
		{"MaskHorz", func(r *Renderer, tg, img, _ *ebiten.Image) { r.MaskHorz(tg, img, 0, 0, 0, 10) }},
		{"MaskCirc", func(r *Renderer, tg, img, _ *ebiten.Image) { r.MaskCirc(tg, img, 0, 0, 10, 10, 5, 2) }},
		{"HalftoneTri", func(r *Renderer, tg, img, _ *ebiten.Image) { r.HalftoneTri(tg, img, 0, 0, 8, 1, 4, 0, 0) }},
		{"WarpBarrel", func(r *Renderer, tg, img, _ *ebiten.Image) { r.WarpBarrel(tg, img, 0, 0, 0.1, 0.1) }},
		{"WarpArc", func(r *Renderer, tg, img, _ *ebiten.Image) { r.WarpArc(tg, img, 16, 16, 10, 1, 2) }},
		{"Blur", func(r *Renderer, tg, img, _ *ebiten.Image) { r.Blur(tg, img, 0, 0, 2) }},
		{"Blur2", func(r *Renderer, tg, img, _ *ebiten.Image) { r.Blur2(tg, img, 0, 0, 2, 2) }},
		{"BlurK", func(r *Renderer, tg, img, _ *ebiten.Image) {
			r.BlurK(tg, img, 0, 0, KernelOptions{Downscaling: DownscaleX2})
		}},
		{"BlurVogel", func(r *Renderer, tg, img, _ *ebiten.Image) { r.BlurVogel(tg, img, 0, 0, 2, 16, DownscaleNone, 0) }},
		{"Glow2", func(r *Renderer, tg, img, _ *ebiten.Image) { r.Glow2(tg, img, 0, 0, 2, 2, 0, 1) }},
		{"MorphExpansionRect", func(r *Renderer, tg, img, _ *ebiten.Image) { r.MorphExpansionRect(tg, img, 0, 0, 2) }},
		{"MorphErosion", func(r *Renderer, tg, img, _ *ebiten.Image) { r.MorphErosion(tg, img, 0, 0, 2, 0) }},
		{"JFMapCompute", func(r *Renderer, tg, img, _ *ebiten.Image) {
			r.JFMapCompute(sizedLike(tg, img), img, JFMapComputeOptions{MaxDistance: 4})
		}},
		{"JFMExpand", func(r *Renderer, tg, img, _ *ebiten.Image) { r.JFMExpand(tg, img, 0, 0, 2, JFMap{}) }},
		{"JFMErode", func(r *Renderer, tg, img, _ *ebiten.Image) { r.JFMErode(tg, img, 0, 0, 2, JFMap{}) }},
	}

	panicsOnNil := map[string]bool{"AlphaNormalsCompute": true, "JFMapCompute": true}

	pixcheck.Run(t, func() {
		full := ebiten.NewImage(32, 32)
		full.Fill(image.White)
		empty := full.SubImage(image.Rect(4, 4, 4, 10)).(*ebiten.Image)
		for _, op := range ops {
			for _, img := range []*ebiten.Image{nil, empty} {
				r := NewRenderer()
				target := ebiten.NewImage(32, 32)
				label := fmt.Sprintf("%s, %s", op.name, map[bool]string{true: "nil", false: "empty"}[img == nil])
				panicked := drawRecovering(func() { op.draw(r, target, img, full) })
				if img == nil && panicsOnNil[op.name] {
					if panicked == "" {
						t.Errorf("%s: didn't panic", label)
					}
					continue
				}
				if panicked != "" {
					t.Errorf("%s: %s", label, panicked)
					continue
				}
				warnings := slices.Collect(r.Warnings.All())
				wantWarnings := []Warning{}
				if img == nil {
					wantWarnings = []Warning{WarnMissingSourceOpSkipped}
				}
				if !slices.Equal(warnings, wantWarnings) {
					t.Errorf("%s: warnings %v, want %v", label, warnings, wantWarnings)
				}
				if pixcheck.AlphaSum(pixcheck.Read(target)) != 0 {
					t.Errorf("%s: drew on the target", label)
				}
			}
		}
	})
}

// sizedLike is a helper for TestSkippedSources. It returns img, or an empty
// image if like is empty, for operations that require matching sizes.
func sizedLike(img, like *ebiten.Image) *ebiten.Image {
	if like != nil && like.Bounds().Empty() {
		return img.SubImage(image.Rectangle{}).(*ebiten.Image)
	}
	return img
}

// drawRecovering is a helper for TestSkippedSources. It returns the panic
// message of draw, if any.
func drawRecovering(draw func()) (panicked string) {
	defer func() {
		if err := recover(); err != nil {
			panicked = fmt.Sprint(err)
		}
	}()
	draw()
	return ""
}
