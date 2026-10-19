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
