package shapes

import (
	"fmt"
	"image"

	"github.com/hajimehoshi/ebiten/v2"
)

type offscreen struct {
	image  *ebiten.Image
	parent *ebiten.Image

	maxWidth    int // none if <= 0
	maxHeight   int // none if <= 0
	extraMargin int
}

// maxWidth and maxHeight <= 0 indicate no size limit
func newOffscreen(maxWidth, maxHeight, extraMargin int) offscreen {
	if extraMargin < 0 {
		panic("extraMargin < 0")
	}
	if (maxWidth <= 0) != (maxHeight <= 0) {
		panic("both maxWidth and maxHeight must be either > 0 (both limited) or <= 0 (both unlimited)")
	}
	return offscreen{
		maxWidth:    maxWidth,
		maxHeight:   maxHeight,
		extraMargin: extraMargin,
	}
}

// the returned bool indicates whether the offscreen is clear or not.
// if clear is requested, then this will always be true. if clear is not
// requested, this sometimes can still be true.
func (off *offscreen) WithSize(w, h int, clear bool) (*ebiten.Image, bool) {
	hasSizeLimits := (off.maxWidth > 0)
	if hasSizeLimits && (w > off.maxWidth || h > off.maxHeight) {
		panic(fmt.Sprintf("requested offscreen of size %dx%d, but maxWidth/maxHeight are %dx%d", w, h, off.maxWidth, off.maxHeight))
	}

	nw, nh := w+off.extraMargin, h+off.extraMargin
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
	bounds := off.parent.Bounds()
	currWidth, currHeight := bounds.Dx(), bounds.Dy()
	if currWidth > w && currHeight > h {
		off.parent.SubImage(image.Rect(0, 0, w+1, h+1)).(*ebiten.Image).Clear()
	} else if currWidth > w {
		off.parent.SubImage(image.Rect(0, 0, w+1, h)).(*ebiten.Image).Clear()
	} else if currHeight > h {
		off.parent.SubImage(image.Rect(0, 0, w, h+1)).(*ebiten.Image).Clear()
	} else {
		off.parent.Clear()
	}
}

func newUnmanagedImage(w, h int) *ebiten.Image {
	opts := ebiten.NewImageOptions{Unmanaged: true}
	return ebiten.NewImageWithOptions(image.Rect(0, 0, w, h), &opts)
}
