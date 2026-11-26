package shapes

import (
	"image"
	"slices"
	"testing"

	"github.com/bstkhq/go-shapes/internal/pixcheck"
	"github.com/hajimehoshi/ebiten/v2"
)

// ---- TestMorphBordersAsPadding and TestMorphClampContinuesBorders ----

// Without Clamp, every morphological operation must give the same result as
// drawing the source with transparent padding around it, so the area past its
// borders behaves exactly like transparent pixels.
func TestMorphBordersAsPadding(t *testing.T) {
	const Pad = 8 // more than any of the operations reaches

	report := pixcheck.Report{Rule: pixcheck.Rule{MaxDelta: 1}}
	pixcheck.Run(t, func() {
		r := NewRenderer()
		// shapes cut by the borders, so that the borders matter
		for _, shape := range []testShape{bandShape, cornerShape} {
			for _, op := range morphOps {
				source := pixcheck.NewImage(shape.img, image.Pt(3, 5))
				w, h := source.Bounds().Dx()+2*Pad, source.Bounds().Dy()+2*Pad
				want := ebiten.NewImage(w, h)
				op.draw(r, want, padded(source, Pad), 0, 0)
				got := ebiten.NewImage(w, h)
				op.draw(r, got, source, Pad, Pad)
				report.Compare(shape.name+", "+op.name, got, want)
			}
		}
	})
	report.Check(t)
}

// With Clamp, every operation that supports it must give the same result as
// drawing a source that continues past its borders, repeating the pixels on
// them. Only the source area is compared: past the borders, the JFM operations
// follow the distances to the boundaries inside instead, and MorphOutline
// draws nothing.
func TestMorphClampContinuesBorders(t *testing.T) {
	const Pad = 8 // more than any of the operations reaches

	report := pixcheck.Report{Rule: pixcheck.Rule{MaxDelta: 1}}
	pixcheck.Run(t, func() {
		r := NewRenderer()
		// shapes cut by the borders, so that the borders matter
		for _, shape := range []testShape{bandShape, cornerShape} {
			for _, op := range morphOpsWithFlags {
				w, h := shape.img.Rect.Dx(), shape.img.Rect.Dy()
				want := ebiten.NewImage(w, h)
				continued := pixcheck.NewImage(repeatBorders(shape.img, Pad), image.Point{})
				op.draw(r, want, continued, -Pad, -Pad)
				got := ebiten.NewImage(w, h)
				op.draw(r, got, pixcheck.NewImage(shape.img, image.Pt(3, 5)), 0, 0, Clamp)
				report.Compare(shape.name+", "+op.name, got, want)
			}
		}
	})
	report.Check(t)
}

// morphOp draws a morphological operation of source onto target at (ox, oy),
// for TestMorphBordersAsPadding and TestMorphClampContinuesBorders.
type morphOp struct {
	name string
	draw func(r *Renderer, target, source *ebiten.Image, ox, oy float32, flags ...Flag)
}

// morphOpsWithFlags are the operations that support Clamp, checked by
// TestMorphClampContinuesBorders.
var morphOpsWithFlags = []morphOp{
	{"MorphErosion", func(r *Renderer, target, source *ebiten.Image, ox, oy float32, flags ...Flag) {
		r.MorphErosion(target, source, ox, oy, 2, 0.5, flags...)
	}},
	{"MorphOutline", func(r *Renderer, target, source *ebiten.Image, ox, oy float32, flags ...Flag) {
		r.MorphOutline(target, source, ox, oy, 3, 0.5, flags...)
	}},
	{"JFMErode", func(r *Renderer, target, source *ebiten.Image, ox, oy float32, flags ...Flag) {
		r.JFMErode(target, source, ox, oy, 2.5, JFMap{}, flags...)
	}},
	{"JFMErode, normals", func(r *Renderer, target, source *ebiten.Image, ox, oy float32, flags ...Flag) {
		r.JFMErode(target, source, ox, oy, 2.5, JFMap{Normals: alphaNormals(r, source, flags...)}, flags...)
	}},
	{"JFMOutline", func(r *Renderer, target, source *ebiten.Image, ox, oy float32, flags ...Flag) {
		r.JFMOutline(target, source, ox, oy, 5, JFMap{}, flags...)
	}},
	{"JFMOutline, normals", func(r *Renderer, target, source *ebiten.Image, ox, oy float32, flags ...Flag) {
		r.JFMOutline(target, source, ox, oy, 5, JFMap{Normals: alphaNormals(r, source, flags...)}, flags...)
	}},
}

// morphOps are all the morphological operations, checked by
// TestMorphBordersAsPadding.
var morphOps = slices.Concat(morphOpsWithFlags, []morphOp{
	withoutFlags("MorphExpansion", func(r *Renderer, target, source *ebiten.Image, ox, oy float32) {
		r.MorphExpansion(target, source, ox, oy, 2, 0.5)
	}),
	withoutFlags("MorphExpansionRect", func(r *Renderer, target, source *ebiten.Image, ox, oy float32) {
		r.MorphExpansionRect(target, source, ox, oy, 2)
	}),
	withoutFlags("JFMExpand", func(r *Renderer, target, source *ebiten.Image, ox, oy float32) {
		r.JFMExpand(target, source, ox, oy, 2.5, JFMap{})
	}),
	withoutFlags("JFMExpand, normals", func(r *Renderer, target, source *ebiten.Image, ox, oy float32) {
		r.JFMExpand(target, source, ox, oy, 2.5, JFMap{Normals: alphaNormals(r, source)})
	}),
})

// withoutFlags adapts an operation without flags to a morphOp, which panics if
// it gets any.
func withoutFlags(name string, draw func(r *Renderer, target, source *ebiten.Image, ox, oy float32)) morphOp {
	return morphOp{name, func(r *Renderer, target, source *ebiten.Image, ox, oy float32, flags ...Flag) {
		if len(flags) > 0 {
			panic(name + " doesn't support flags")
		}
		draw(r, target, source, ox, oy)
	}}
}

// repeatBorders returns img with padding around it that repeats the pixels on
// its borders, for TestMorphClampContinuesBorders.
func repeatBorders(img *image.RGBA, padding int) *image.RGBA {
	w, h := img.Rect.Dx(), img.Rect.Dy()
	out := image.NewRGBA(image.Rect(0, 0, w+2*padding, h+2*padding))
	for y := range out.Rect.Dy() {
		for x := range out.Rect.Dx() {
			sx, sy := clamp(x-padding, 0, w-1), clamp(y-padding, 0, h-1)
			out.SetRGBA(x, y, img.RGBAAt(img.Rect.Min.X+sx, img.Rect.Min.Y+sy))
		}
	}
	return out
}
