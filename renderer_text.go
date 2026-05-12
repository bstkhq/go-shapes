package shapes

import (
	"math"
	"slices"
	"strconv"
	"unicode/utf8"

	"github.com/hajimehoshi/ebiten/v2"
)

// LineAnchor is one of the components of the [TextAlign], which allows fine grained
// adjustment of the vertical origin for the first or last line during text rendering
// operations.
type LineAnchor uint8

func (a LineAnchor) String() string {
	if str, ok := a.validString(); ok {
		return str
	}
	return "LineAnchor#" + strconv.Itoa(int(a))
}

func (a LineAnchor) validString() (string, bool) {
	switch a {
	case Top:
		return "Top", true
	case CapLine:
		return "CapLine", true
	case MidLine:
		return "MidLine", true
	case Baseline:
		return "Baseline", true
	case Bottom:
		return "Bottom", true
	default:
		return "", false
	}
}

// See [TextAlign] for more details.
const (
	anchorMask = 0b1111

	Top      LineAnchor = 0b1000
	CapLine  LineAnchor = 0b1101
	MidLine  LineAnchor = 0b1011
	Baseline LineAnchor = 0b1001
	Bottom   LineAnchor = 0b0001
)

// TextAlign defines the horizontal and vertical text alignment during text rendering
// operations.
//
// A single constant we be used to control the horizontal and vertical alignment of the
// "text container". Given a drawing position (x, y):
//   - [TopLeft] aligns the text's top-left corner to (x, y)
//   - [Center] aligns the text's center to (x, y)
//   - [BottomCenter] aligns the text's bottom-center to (x, y)
//
// For even finer-grained vertical adjustment on Top* and Bottom* aligns, a line anchor
// can also be specified. Given a drawing position (x, y):
//   - TopLeft.Snap([Baseline]) aligns the first line's baseline
//     to (x, y)
//   - BottomRight.Snap([Top]) aligns the last line's top to (x, y)
//   - TopCenter.Snap([Top]) is the same as TopCenter
//
// Notice that line anchors are ignored for Center* aligns, and they are completely
// optional otherwise.
type TextAlign uint8

// Snap returns a new align with the given [LineAnchor] applied.
//
// Remember that line anchors should only be used on Top* and Bottom*
// aligns, never Center*.
func (a TextAlign) Snap(anchor LineAnchor) TextAlign {
	return (a & 0b1111_0000) | TextAlign(anchor)
}

func (a TextAlign) String() string {
	fallback := func(a TextAlign) string {
		return "TextAlign#" + strconv.Itoa(int(a))
	}

	anchor := (LineAnchor(a) & anchorMask)
	anchorStr, anchorOk := anchor.validString()
	if !anchorOk && (a&vertMask) != vertMiddle {
		return fallback(a)
	}

	var vert string
	var appendAnchor bool
	switch a & vertMask {
	case vertStart:
		if anchor == Bottom {
			vert = "FirstBottom"
		} else {
			vert = anchorStr
		}
	case vertMiddle:
		vert = "Center"
		appendAnchor = true
	case vertEnd:
		vert = "Last" + anchorStr
	default:
		return fallback(a)
	}

	var horz string
	switch a & horzMask {
	case horzStart:
		horz = "Left"
	case horzMiddle:
		horz = "Center"
		if vert == "Center" {
			vert = ""
		}
	case horzEnd:
		horz = "Right"
	default:
		return fallback(a)
	}

	if appendAnchor && anchorStr != "" {
		return vert + horz + "-" + anchorStr
	}
	return vert + horz

}

func (a TextAlign) horzShiftRate() (float32, bool) {
	switch a & horzMask {
	case horzStart:
		return 0.0, true
	case horzMiddle:
		return -0.5, true
	case horzEnd:
		return -1.0, true
	default:
		return 0.0, false
	}
}

func (a TextAlign) oyAdjust(y float32, contentLogicHeight int, scale float32) (float32, bool) {
	switch a & vertMask {
	case vertStart:
		return a.anchorTopAdjust(y, scale)
	case vertMiddle:
		return y - float32(contentLogicHeight)*scale/2.0, true
	case vertEnd:
		return a.anchorBottomAdjust(y-float32(contentLogicHeight)*scale, scale)
	default:
		return y, false
	}
}

func (a TextAlign) anchorTopAdjust(y, scale float32) (float32, bool) {
	switch LineAnchor(a & anchorMask) {
	case Top:
		return y, true
	case CapLine:
		return y - float32(ark10pxMap[fontMapIdxAscent]-ark10pxMap[fontMapIdxCapHeight])*scale, true
	case MidLine:
		return y - float32(ark10pxMap[fontMapIdxAscent]-ark10pxMap[fontMapIdxMidHeight])*scale, true
	case Baseline:
		return y - float32(ark10pxMap[fontMapIdxAscent])*scale, true
	case Bottom:
		return y - float32(ark10pxMap[fontMapIdxAscent]+ark10pxMap[fontMapIdxDescent])*scale, true
	default:
		return y, false
	}
}

func (a TextAlign) anchorBottomAdjust(y, scale float32) (float32, bool) {
	switch LineAnchor(a & anchorMask) {
	case Top:
		return y + float32(ark10pxMap[fontMapIdxAscent]+ark10pxMap[fontMapIdxDescent])*scale, true
	case CapLine:
		return y + float32(ark10pxMap[fontMapIdxDescent]+ark10pxMap[fontMapIdxCapHeight])*scale, true
	case MidLine:
		return y + float32(ark10pxMap[fontMapIdxDescent]+ark10pxMap[fontMapIdxMidHeight])*scale, true
	case Baseline:
		return y + float32(ark10pxMap[fontMapIdxDescent])*scale, true
	case Bottom:
		return y, true
	default:
		return y, false
	}
}

const (
	vertMask   TextAlign = 0b1100_0000
	vertStart  TextAlign = 0b0100_0000
	vertMiddle TextAlign = 0b1100_0000
	vertEnd    TextAlign = 0b1000_0000

	horzMask   TextAlign = 0b0011_0000
	horzStart  TextAlign = 0b0001_0000
	horzMiddle TextAlign = 0b0011_0000
	horzEnd    TextAlign = 0b0010_0000

	TopLeft   TextAlign = vertStart | horzStart | TextAlign(Top)
	TopCenter TextAlign = vertStart | horzMiddle | TextAlign(Top)
	TopRight  TextAlign = vertStart | horzEnd | TextAlign(Top)

	CenterLeft  TextAlign = vertMiddle | horzStart
	Center      TextAlign = vertMiddle | horzMiddle
	CenterRight TextAlign = vertMiddle | horzEnd

	BottomLeft   TextAlign = vertEnd | horzStart | TextAlign(Bottom)
	BottomCenter TextAlign = vertEnd | horzMiddle | TextAlign(Bottom)
	BottomRight  TextAlign = vertEnd | horzEnd | TextAlign(Bottom)
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

	// SpaceWidth controls the text logical space size.
	// A sensible default is used when the value is zero.
	//
	// This is a low-level control only intended for advanced usage.
	SpaceWidth uint8

	// LineGap controls the logical line interspacing size.
	// A sensible default is used when the value is zero.
	//
	// This is a low-level control intended for advanced usage.
	LineGap int8
}

func TextOpts(scale float32, align TextAlign) TextOptions {
	return TextOptions{
		Scale: scale,
		Align: align,
	}
}

// SmoothAnim returns a copy of TextOptions with the smooth animation flag set. By default,
// this flag is not set, and line positions are quantized to the nearest pixel.
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

func (opts TextOptions) lineGap() int8 {
	if opts.LineGap != 0 {
		return opts.LineGap
	}
	return int8(ark10pxMap[fontMapIdxLineGap])
}

func (opts TextOptions) spaceWidth() uint8 {
	if opts.SpaceWidth != 0 {
		return opts.SpaceWidth
	}
	return ark10pxMap[fontMapIdxSpaceWidth]
}

func (opts TextOptions) scale() float32 {
	if opts.Scale > 0.0 {
		return opts.Scale
	}
	return 1.0
}

// Text is a utility method to draw ASCII text with the proportional 10px [Ark pixel
// font]. This is very similar to what ebiten/inpututil does, but this is the proportional
// version and slightly smaller by default. [TextOptions] can be used to control the text
// scale and align.
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
	ascent := ark10pxMap[fontMapIdxAscent]
	descent := ark10pxMap[fontMapIdxDescent]
	lineGap := float32(opts.lineGap()) * scale
	textHeight := lineCount*int(ascent+descent) + int(opts.lineGap())*(lineCount-1)
	horzShiftRate, horzAlignOk := opts.Align.horzShiftRate()
	y, vertAlignOk := opts.Align.oyAdjust(y, textHeight, scale)
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
	glyphsPerRow := int32(ark10pxMap[fontMapIdxGlyphsPerRow])
	glyphFrameWidth := float32(ark10pxMap[fontMapIdxGlyphFrameWidth])
	glyphFrameHeight := float32(ascent + descent)
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
			dy += glyphFrameHeight*scale + lineGap
			dx = 0
			lineGlyphCount = 0
		default:
			index := runeRefIndex(codePoint)
			if index >= 0 {
				dx += pendingLetterGap
				w := float32(ark10pxMap[fontMapIdxFirstGlyphWidth+index])
				ws := w * scale
				srcX := float32(1.0) + float32(index%glyphsPerRow)*(glyphFrameWidth+1)
				srcY := float32(1.0) + float32(index/glyphsPerRow)*(glyphFrameHeight+1)
				offset := 0.1 * scale
				setVertDstCoordsIdx(r.vertices, glyphCount<<2, x+dx-offset, y+dy-offset, x+dx+ws+offset, y+dy+glyphFrameHeight*scale+offset)
				setVertSrcCoordsIdx(r.vertices, glyphCount<<2, srcX-0.1, srcY-0.1, srcX+w+0.1, srcY+glyphFrameHeight+0.1)
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
	r.opts.Images[0] = loadArk10pxAtlas()
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
	ascent := int32(ark10pxMap[fontMapIdxAscent])
	descent := int32(ark10pxMap[fontMapIdxDescent])
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
			index := runeRefIndex(codePoint)
			if index >= 0 {
				lw += pendingLetterGap
				lw += int32(ark10pxMap[fontMapIdxFirstGlyphWidth+index])
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

func runeRefIndex(codePoint rune) int32 {
	if codePoint < 33 {
		return -1
	}
	if codePoint < 127 {
		return codePoint - 33
	}
	if codePoint < 161 || codePoint > 255 {
		return -1
	}
	return codePoint - 161 + (127 - 33)
}
