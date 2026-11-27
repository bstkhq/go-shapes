package shapes

import (
	"fmt"
	"image"
	"math/bits"

	"github.com/hajimehoshi/ebiten/v2"
)

type offscreen struct {
	image  *ebiten.Image
	parent *ebiten.Image

	maxWidth  int // none if <= 0
	maxHeight int // none if <= 0
}

// maxWidth and maxHeight <= 0 indicate no size limit
func newOffscreen(maxWidth, maxHeight int) offscreen {
	if (maxWidth <= 0) != (maxHeight <= 0) {
		panic("both maxWidth and maxHeight must be either > 0 (both limited) or <= 0 (both unlimited)")
	}
	return offscreen{
		maxWidth:  maxWidth,
		maxHeight: maxHeight,
	}
}

// the returned bool indicates whether the offscreen has to be cleared
// (including an extra 1 px margin) or not. if clear is requested, then this
// will always be true. if clear is not requested, this sometimes can still be
// true.
func (off *offscreen) WithSize(w, h int, clear bool) (*ebiten.Image, bool) {
	hasSizeLimits := (off.maxWidth > 0)
	if hasSizeLimits && (w > off.maxWidth || h > off.maxHeight) {
		panic(fmt.Sprintf("requested offscreen of size %dx%d, but maxWidth/maxHeight are %dx%d", w, h, off.maxWidth, off.maxHeight))
	}

	// Ebitengine allocates power of 2 textures, so we have to use the same
	// for optimal texture usage in case of later resize/growth
	nw, nh := textureSize(w), textureSize(h)
	if hasSizeLimits {
		nw, nh = min(nw, off.maxWidth), min(nh, off.maxHeight)
	}

	if off.image == nil {
		off.parent = newUnmanagedImage(nw, nh)
		off.image = off.parent.SubImage(image.Rect(0, 0, w, h)).(*ebiten.Image)
		return off.image, true
	}

	bounds := off.image.Bounds()
	if bounds.Dx() == w && bounds.Dy() == h {
		if clear {
			off.clearParentFor(w, h)
		}
		return off.image, clear
	}

	bounds = off.parent.Bounds()
	currWidth, currHeight := bounds.Dx(), bounds.Dy()
	if currWidth >= w && currHeight >= h {
		if clear {
			off.clearParentFor(w, h)
		}
		off.image = off.parent.SubImage(image.Rect(0, 0, w, h)).(*ebiten.Image)
	} else {
		if off.parent != nil {
			off.parent.Deallocate()
		}
		off.parent = newUnmanagedImage(max(nw, currWidth), max(nh, currHeight))
		off.image = off.parent.SubImage(image.Rect(0, 0, w, h)).(*ebiten.Image)
		clear = true
	}
	return off.image, clear
}

func (off *offscreen) clearParentFor(w, h int) {
	clearRect := off.parent.Bounds().Intersect(image.Rect(0, 0, w+1, h+1))
	Clear(off.parent, clearRect)
}

// textureSize mimics Ebitengine's behavior: min size or next power of two.
// See github.com/hajimehoshi/ebiten/blob/v2.9.0/internal/graphics/math.go
func textureSize(n int) int {
	if n <= 16 {
		return 16
	}
	return 1 << bits.Len(uint(n-1))
}

func newUnmanagedImage(w, h int) *ebiten.Image {
	opts := ebiten.NewImageOptions{Unmanaged: true}
	return ebiten.NewImageWithOptions(image.Rect(0, 0, w, h), &opts)
}
