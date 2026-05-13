package shapes

import (
	"math"
	"slices"
	"unicode/utf8"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	textFlagSmoothAnim = 0b0000_0001
)

// TextOptions are used in [Renderer.Text]() and [Renderer.TextSize]().
type TextOptions struct {
	// Scale controls the text scale.
	// Values <= 0 default to scale = 1.
	Scale float32

	// Align controls the text positioning relative to the drawing (x,y)
	// coordinates. If unset, [TopLeft] will be used as the default.
	Align TextAlign

	// 0b0000_000S
	//  - S: smooth animation flag. When set, line origin should not be quantized.
	// Other flags might include monospace, font, nearest, etc.
	flags uint8

	// SpaceWidth controls the font's logical space width.
	// A sensible default is used when the value is zero.
	//
	// This is a low-level control for advanced usage.
	SpaceWidth uint8

	// LineGap controls the logical line interspacing.
	// A sensible default is used when the value is zero.
	//
	// This is a low-level control for advanced usage.
	LineGap int8
}

func TextOpts(scale float32, align TextAlign) TextOptions {
	return TextOptions{
		Scale: scale,
		Align: align,
	}
}

// Quantized returns a copy of TextOptions with the quantization flag set to the
// given value. By default , which causes line positions to be quantized to the
// nearest pixel. As a general guideline:
//   - Static text should snap to nearest pixel to avoid blurriness on centered aligns.
//   - Animated text (moving or zooming) should use SmoothAnim to prevent motion jitter.
func (opts TextOptions) SmoothAnim() TextOptions {
	opts.flags |= textFlagSmoothAnim
	return opts
}

// if smooth animation flag is not set, the given float is quantized to the nearest integer.
func (opts TextOptions) quantize(f float32) float32 {
	if opts.flags&textFlagSmoothAnim == textFlagSmoothAnim {
		return f
	}
	return float32(math.Round(float64(f)))
}

func (opts TextOptions) fontMap() fontMap {
	return ark10pxMap
}

func (opts TextOptions) fontAtlas() *ebiten.Image {
	return loadArk10pxAtlas()
}

func (opts TextOptions) lineGap() int8 {
	if opts.LineGap != 0 {
		return opts.LineGap
	}
	return opts.fontMap().LineGap()
}

func (opts TextOptions) spaceWidth() uint8 {
	if opts.SpaceWidth != 0 {
		return opts.SpaceWidth
	}
	return opts.fontMap().SpaceWidth()
}

func (opts TextOptions) scale() float32 {
	if opts.Scale > 0.0 {
		return opts.Scale
	}
	return 1.0
}

// Text is a utility method to draw ASCII and Latin-1 Supplement text with the proportional
// 10px [Ark pixel font]. This is very similar to what ebiten/inpututil does, but this is
// the proportional version and slightly smaller by default. [TextOptions] can be used to
// control the text scale and align.
//
// Unknown glyphs are silently skipped. The text data is stored in binary format, and the
// atlas image is not initialized unless text functions are used.
//
// [Ark pixel font]: https://ark-pixel-font.takwolf.com/
func (r *Renderer) Text(target *ebiten.Image, text string, x, y float32, opts TextOptions) {
	// get basic glyph and line counts
	glyphCount, lineCount := 0, 0
	for _, codePoint := range text {
		if codePoint == '\n' {
			lineCount += 1
		} else if codePoint > ' ' {
			glyphCount += 1
		}
	}

	if glyphCount == 0 {
		return
	}

	if r, _ := utf8.DecodeLastRuneInString(text); r != '\n' {
		lineCount += 1
	}

	// set up vertices and indices (might slightly overshoot)
	// TODO: consider MaxVertexCount?
	scale := opts.scale()
	r.setFlatCustomVAs01(1.0/scale, 1.0/scale)
	growth := glyphCount*4 - len(r.vertices)
	r.vertices = slices.Grow(r.vertices, max(growth, 0))[:glyphCount*4]
	fastPatternFill(r.vertices[4:], r.vertices[:4])
	r.indices = r.indices[:0]
	r.indices = slices.Grow(r.indices, glyphCount*6)[:glyphCount*6]
	iI, iV := 0, uint32(0)
	for range glyphCount {
		r.indices[iI+0] = iV
		r.indices[iI+1] = iV + 1
		r.indices[iI+2] = iV + 2
		r.indices[iI+3] = iV
		r.indices[iI+4] = iV + 2
		r.indices[iI+5] = iV + 3
		iV += 4
		iI += 6
	}

	// compute align offsets
	if opts.Align == 0 {
		opts.Align = TopLeft
	}
	fontMap := opts.fontMap()
	lineGap := float32(opts.lineGap()) * scale
	textHeight := lineCount*int(fontMap.Ascent()+fontMap.Descent()) + int(opts.lineGap())*(lineCount-1)
	horzShiftRate, horzAlignOk := opts.Align.horzShiftRate()
	y, vertAlignOk := opts.Align.oyAdjust(y, textHeight, fontMap, scale)
	y = opts.quantize(y)
	if !horzAlignOk || !vertAlignOk {
		r.Warnings.report(WarnInvalidTextAlign, opts.Align)
		if !horzAlignOk {
			opts.Align = (opts.Align & ^horzMask) | horzStart
		}
		if !vertAlignOk {
			opts.Align = vertStart | TextAlign(Top)
		}
	}

	// iterate text
	var dx, dy float32
	scaledGlyphHeight := float32(fontMap.Ascent()+fontMap.Descent()) * scale
	glyphCount = 0 // reset and compute accurately
	lineGlyphCount := 0
	spaceWidth := float32(opts.spaceWidth()) * scale

	pendingLetterGap := float32(0)
	for _, codePoint := range text {
		switch codePoint {
		case ' ':
			dx += spaceWidth
			pendingLetterGap = 0
		case '\n':
			r.lineApplyHorzShift(opts.quantize(horzShiftRate*dx), glyphCount-lineGlyphCount, lineGlyphCount)
			pendingLetterGap = 0
			dy += scaledGlyphHeight + lineGap
			dx = 0
			lineGlyphCount = 0
		default:
			if rect, ok := fontMap.GlyphAtlasRect(codePoint); ok {
				dx += pendingLetterGap
				ws := float32(rect.Dx()) * scale
				setVertDstCoordsIdx(r.vertices, glyphCount<<2, x+dx, y+dy, x+dx+ws, y+dy+scaledGlyphHeight)
				o, f := RectPointsF32(rect)
				setVertSrcCoordsIdx(r.vertices, glyphCount<<2, o.X, o.Y, f.X, f.Y)
				glyphCount += 1
				lineGlyphCount += 1
				dx += ws
				pendingLetterGap = scale
			} else {
				// missing glyph (should use notdef), or control (should skip)
				// TODO: map notdef explicitly?
			}
		}
	}
	r.lineApplyHorzShift(opts.quantize(horzShiftRate*dx), glyphCount-lineGlyphCount, lineGlyphCount)

	// draw
	r.opts.Images[0] = opts.fontAtlas()
	r.vertices = r.vertices[:glyphCount<<2]
	r.indices = r.indices[:glyphCount*6]
	target.DrawTrianglesShader32(r.vertices, r.indices, shaderTextBilinear.Load(), &r.opts)
	r.opts.Images[0] = nil
	r.vertices = r.vertices[:4]
	r.restoreIndices()
}

func (r *Renderer) lineApplyHorzShift(shift float32, startLineGlyph, lineGlyphCount int) {
	if shift == 0 || lineGlyphCount == 0 {
		return
	}

	for i := range lineGlyphCount * 4 {
		r.vertices[startLineGlyph*4+i].DstX += shift
	}
}

func fastPatternFill[T any](slice []T, pattern []T) {
	copy(slice, pattern)
	done := len(pattern)
	for done < len(slice) {
		copy(slice[done:], slice[:done])
		done *= 2
	}
}

// TextSize returns the size of the given text.
//
// See [Renderer.Text]() for more details on the font used.
func (r *Renderer) TextSize(text string, opts TextOptions) (width, height float32) {
	if len(text) == 0 {
		return 0, 0
	}

	var w, lw, h int32
	spaceWidth := int32(opts.spaceWidth())
	lineGap := int32(opts.lineGap())
	fontMap := opts.fontMap()
	ascent := int32(fontMap.Ascent())
	descent := int32(fontMap.Descent())
	pendingLineGap := int32(0)
	pendingLetterGap := int32(0)
	for _, codePoint := range text {
		h += pendingLineGap
		pendingLineGap = 0

		switch codePoint {
		case ' ':
			lw += spaceWidth
			pendingLetterGap = 0
		case '\n':
			pendingLetterGap = 0
			h += ascent + descent
			w = max(w, lw)
			lw = 0
			pendingLineGap = lineGap
		default:
			advance, ok := fontMap.GlyphAdvance(codePoint)
			if ok {
				lw += pendingLetterGap
				lw += int32(advance)
				pendingLetterGap = 1
			} else {
				// missing glyph (should use notdef), or control (should skip)
			}
		}
	}

	w = max(w, lw)
	if w != 0 {
		h += pendingLineGap
		h += ascent + descent
	}

	scale := opts.scale()
	return float32(w) * scale, float32(h) * scale
}
