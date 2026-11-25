package shapes

import (
	"slices"
	"strings"
	"testing"
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
