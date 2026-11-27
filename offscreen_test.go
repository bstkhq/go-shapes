package shapes

import (
	"image"
	"testing"

	"github.com/bstkhq/go-shapes/internal/pixcheck"
)

// Ebitengine always allocates power of 2 textures. This test verifies that
// offscreens match that behavior for best underlying texture usage.
func TestOffscreenSizes(t *testing.T) {
	requests := []struct {
		w, h       int
		wantParent image.Point
	}{
		{100, 30, image.Pt(128, 32)},
		{120, 31, image.Pt(128, 32)}, // fits, reused
		{8, 8, image.Pt(128, 32)},    // smaller, reused
		{130, 20, image.Pt(256, 32)}, // grows on one axis, keeps the other
		{2048, 2048, image.Pt(2048, 2048)},
	}
	pixcheck.Run(t, func() {
		off := newOffscreen(0, 0)
		for _, req := range requests {
			img, _ := off.WithSize(req.w, req.h, false)
			if got := img.Bounds().Size(); got != image.Pt(req.w, req.h) {
				t.Errorf("request %dx%d: got an image of %v", req.w, req.h, got)
			}
			if got := off.parent.Bounds().Size(); got != req.wantParent {
				t.Errorf("request %dx%d: parent %v, want %v", req.w, req.h, got, req.wantParent)
			}
		}
	})
	for n, want := range map[int]int{0: 16, 1: 16, 16: 16, 17: 32, 1000: 1024, 1024: 1024, 1025: 2048} {
		if got := textureSize(n); got != want {
			t.Errorf("textureSize(%d) = %d, want %d", n, got, want)
		}
	}
}
