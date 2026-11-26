package shapes

import (
	"image"
	"image/color"
	"math"
)

// jfmModel is a CPU copy of how jfmaps measure distances (see
// jfm_read.kage), that finds the best seed for each pixel by brute force. Its
// formulas must be changed along with the shaders'.
type jfmModel struct {
	jfmap           JFMap       // only the region limits are used
	source, normals *image.RGBA // normals can be nil
}

// padded returns the model of the source with transparent padding around it.
func (m jfmModel) padded(padding int) jfmModel {
	grow := func(img *image.RGBA) *image.RGBA {
		if img == nil {
			return nil
		}
		w, h := img.Rect.Dx(), img.Rect.Dy()
		out := image.NewRGBA(image.Rect(0, 0, w+2*padding, h+2*padding))
		for y := range h {
			for x := range w {
				out.SetRGBA(x+padding, y+padding, img.RGBAAt(img.Rect.Min.X+x, img.Rect.Min.Y+y))
			}
		}
		return out
	}
	return jfmModel{m.jfmap, grow(m.source), grow(m.normals)}
}

// result returns the pixel that an operation draws at the given position, for
// a white renderer color and no tint.
func (m jfmModel) result(op int, pixel image.Point, amount float64) color.RGBA {
	alpha := m.alpha(pixel)
	center := jfmVec{float64(pixel.X) + 0.5, float64(pixel.Y) + 0.5}
	var coverage float64
	switch {
	case op == jfmOpExpand && m.within(alpha):
		coverage = min(1, 0.5-m.regionOffset(alpha)+amount)
	case op == jfmOpExpand:
		coverage = clamp(amount-m.signedDist(center, pixel)+0.5, 0, 1)
	case op == jfmOpErode:
		before := max(0.5-m.signedDist(center, pixel), 0.001)
		coverage = alpha * clamp(before-amount, 0, 1) / min(before, 1)
	case op == jfmOpOutline:
		width, middle := math.Abs(amount), min(amount, 0)*0.5
		coverage = clamp(width*0.5-math.Abs(m.signedDist(center, pixel)-middle)+0.5, 0, 1)
	}
	v := uint8(math.Round(coverage * 255))
	return color.RGBA{v, v, v, v}
}

// signedDist returns the signed distance from the position from to the region
// boundary, using the best seed for the given pixel.
func (m jfmModel) signedDist(from jfmVec, pixel image.Point) float64 {
	alpha := m.alpha(pixel)
	if m.isSeed(pixel) {
		if normal, ok := m.normalsAt(pixel); ok {
			return m.nearestBoundarySign(alpha) * m.normalEdgeDist(from, pixel, normal)
		}
		return m.regionOffset(alpha)
	}

	fromSign := mapBool(m.within(alpha), 1.0, -1.0)
	best := 65536.0
	for y := m.source.Rect.Min.Y; y < m.source.Rect.Max.Y; y++ {
		for x := m.source.Rect.Min.X; x < m.source.Rect.Max.X; x++ {
			if seed := image.Pt(x, y); m.isSeed(seed) {
				best = min(best, m.seedDist(from, seed, fromSign))
			}
		}
	}
	return fromSign * best
}

func (m jfmModel) seedDist(from jfmVec, seed image.Point, fromSign float64) float64 {
	center := jfmVec{float64(seed.X) + 0.5, float64(seed.Y) + 0.5}
	alpha := m.alpha(seed)
	fromAlpha := math.Hypot(from.x-center.x, from.y-center.y) + fromSign*m.regionOffset(alpha)
	normal, ok := m.normalsAt(seed)
	if !ok {
		return fromAlpha
	}
	if dist := fromSign * m.nearestBoundarySign(alpha) * m.normalEdgeDist(from, seed, normal); dist >= 0 {
		return dist
	}
	return fromAlpha
}

func (m jfmModel) normalEdgeDist(from jfmVec, normalOrigin image.Point, normal color.RGBA) float64 {
	dx, dy := float64(normal.R)/255*2-1, float64(normal.G)/255*2-1
	length := math.Hypot(dx, dy)
	dx, dy = dx/length, dy/length
	offset := (float64(normal.B)/255 - 0.5) * 1.5
	tx := from.x - (float64(normalOrigin.X) + 0.5 - dx*offset)
	ty := from.y - (float64(normalOrigin.Y) + 0.5 - dy*offset)
	across := tx*dx + ty*dy
	pastEnd := max(math.Abs(-tx*dy+ty*dx)-0.5, 0)
	dist := math.Hypot(across, pastEnd)
	if across < 0 {
		return -dist
	}
	return dist
}

func (m jfmModel) alpha(pixel image.Point) float64 {
	if !pixel.In(m.source.Rect) {
		return 0
	}
	return float64(m.source.RGBAAt(pixel.X, pixel.Y).A) / 255
}

func (m jfmModel) normalsAt(pixel image.Point) (color.RGBA, bool) {
	if m.normals == nil || !pixel.In(m.normals.Rect) {
		return color.RGBA{}, false
	}
	data := m.normals.RGBAAt(pixel.X, pixel.Y)
	return data, data.A != 0
}

func (m jfmModel) within(alpha float64) bool {
	minAlpha, maxAlpha := m.jfmap.region()
	return float64(minAlpha) <= alpha && alpha <= float64(maxAlpha)
}

func (m jfmModel) isSeed(pixel image.Point) bool {
	if !m.within(m.alpha(pixel)) {
		return false
	}
	for _, step := range []image.Point{{-1, -1}, {0, -1}, {1, -1}, {-1, 0}, {1, 0}, {-1, 1}, {0, 1}, {1, 1}} {
		if !m.within(m.alpha(pixel.Add(step))) {
			return true
		}
	}
	return false
}

func (m jfmModel) nearestBoundarySign(alpha float64) float64 {
	minAlpha32, maxAlpha32 := m.jfmap.region()
	minAlpha, maxAlpha := float64(minAlpha32), float64(maxAlpha32)
	if maxAlpha >= 1 {
		return 1
	}
	if minAlpha <= 0 {
		return -1
	}
	return mapBool(math.Abs(alpha-minAlpha) <= math.Abs(alpha-maxAlpha), -1.0, 1.0)
}

func (m jfmModel) regionOffset(alpha float64) float64 {
	return m.nearestBoundarySign(alpha) * (0.5 - alpha)
}

type jfmVec struct{ x, y float64 }
