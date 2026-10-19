package shapes

import (
	"image"
	"strconv"
	"strings"
	"sync"
)

const notdefOffset = 1 // set to 1 if present, zero otherwise
const fontPadding = 2

type fontMap []byte

func (fm fontMap) atlasGlyphsPerRow() uint8 {
	return fm[0]
}
func (fm fontMap) atlasGlyphFrameWidth() uint8 {
	return fm[1]
}

func (fm fontMap) GlyphAtlasRect(codePoint rune) (image.Rectangle, bool) {
	idx := int(runeRefIndex(codePoint))
	if idx < 0 {
		return image.Rectangle{}, false
	}
	return fm.TextureAtlasRect(idx), true
}

func (fm fontMap) TextureAtlasRect(index int) image.Rectangle {
	h := int(fm.Ascent() + fm.Descent())
	srcX := fontPadding + (index)%int(fm.atlasGlyphsPerRow())*(int(fm.atlasGlyphFrameWidth())+fontPadding)
	srcY := fontPadding + (index)/int(fm.atlasGlyphsPerRow())*(h+fontPadding)
	return image.Rect(srcX, srcY, srcX+int(fm[8+index]), srcY+h)
}

func (fm fontMap) GlyphAdvance(codePoint rune) (uint8, bool) {
	idx := runeRefIndex(codePoint)
	if idx < 0 {
		return 0, false
	}
	return fm[8+idx], true
}

func (fm fontMap) Ascent() uint8 {
	return fm[2]
}
func (fm fontMap) Descent() uint8 {
	return fm[3]
}
func (fm fontMap) LineGap() int8 {
	return int8(fm[4])
}
func (fm fontMap) CapHeight() uint8 {
	return fm[5]
}
func (fm fontMap) MidHeight() uint8 {
	return fm[6]
}
func (fm fontMap) SpaceWidth() uint8 {
	return fm[7]
}

func (fm fontMap) DebugInfo() string {
	var str strings.Builder
	str.WriteString("Font map debug info:")
	str.WriteString("\n>> Glyphs per row: ")
	str.WriteString(strconv.Itoa(int(fm.atlasGlyphsPerRow())))
	str.WriteString("\n>> Glyph frame width: ")
	str.WriteString(strconv.Itoa(int(fm.atlasGlyphFrameWidth())))
	str.WriteString("\n>> Ascent: ")
	str.WriteString(strconv.Itoa(int(fm.Ascent())))
	str.WriteString("\n>> Descent: ")
	str.WriteString(strconv.Itoa(int(fm.Descent())))
	str.WriteString("\n>> LineGap: ")
	str.WriteString(strconv.Itoa(int(fm.LineGap())))
	str.WriteString("\n>> CapHeight: ")
	str.WriteString(strconv.Itoa(int(fm.CapHeight())))
	str.WriteString("\n>> MidHeight: ")
	str.WriteString(strconv.Itoa(int(fm.MidHeight())))
	str.WriteString("\n>> Space width: ")
	str.WriteString(strconv.Itoa(int(fm.SpaceWidth())))
	str.WriteByte('\n')
	return str.String()
}

// NOTE: this could be adapted with font map data if necessary
func runeRefIndex(codePoint rune) int32 {
	if codePoint < 33 {
		return -1
	}
	if codePoint < 127 {
		return codePoint - 33 + notdefOffset
	}
	if codePoint < 161 || codePoint > 255 {
		return -1
	}
	return codePoint - 161 + (127 - 33) + notdefOffset
}

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

// TextAlign defines the horizontal and vertical text alignment for [Renderer.Text]()
// operations.
//
// A single constant can be used to control the horizontal and vertical alignment of the
// "text container". Given a drawing position (x, y):
//   - [TopLeft] aligns the text's top-left corner to (x, y)
//   - [Center] aligns the text's center to (x, y)
//   - [BottomCenter] aligns the text's bottom-center to (x, y)
//
// For finer-grained vertical adjustment on Top* and Bottom* aligns, a [LineAnchor]
// can also be specified. Given a drawing position (x, y):
//   - TopLeft.Snap([Baseline]) aligns the first line's baseline
//     to (x, y)
//   - BottomRight.Snap([Top]) aligns the last line's top to (x, y)
//   - TopCenter.Snap([Top]) is the same as [TopCenter]
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

func (a TextAlign) oyAdjust(y float32, contentLogicHeight int, fontMap fontMap, scale float32) (float32, bool) {
	switch a & vertMask {
	case vertStart:
		return a.anchorTopAdjust(y, fontMap, scale)
	case vertMiddle:
		return y - float32(contentLogicHeight)*scale/2.0, true
	case vertEnd:
		return a.anchorBottomAdjust(y-float32(contentLogicHeight)*scale, fontMap, scale)
	default:
		return y, false
	}
}

func (a TextAlign) anchorTopAdjust(y float32, fontMap fontMap, scale float32) (float32, bool) {
	switch LineAnchor(a & anchorMask) {
	case Top:
		return y, true
	case CapLine:
		return y - float32(fontMap.Ascent()-fontMap.CapHeight())*scale, true
	case MidLine:
		return y - float32(fontMap.Ascent()-fontMap.MidHeight())*scale, true
	case Baseline:
		return y - float32(fontMap.Ascent())*scale, true
	case Bottom:
		return y - float32(fontMap.Ascent()+fontMap.Descent())*scale, true
	default:
		return y, false
	}
}

func (a TextAlign) anchorBottomAdjust(y float32, fontMap fontMap, scale float32) (float32, bool) {
	switch LineAnchor(a & anchorMask) {
	case Top:
		return y + float32(fontMap.Ascent()+fontMap.Descent())*scale, true
	case CapLine:
		return y + float32(fontMap.Descent()+fontMap.CapHeight())*scale, true
	case MidLine:
		return y + float32(fontMap.Descent()+fontMap.MidHeight())*scale, true
	case Baseline:
		return y + float32(fontMap.Descent())*scale, true
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

var fontInfo FontInfo
var fontInfoOnce sync.Once

type FontInfo struct {
	Family     string
	Variant    string
	Ascent     uint8
	Descent    uint8
	LineGap    int8
	CapHeight  uint8
	MidHeight  uint8
	SpaceWidth uint8
}

// Font returns the information of the font used for the [Renderer] text
// operations.
func Font() FontInfo {
	fontInfoOnce.Do(loadFontInfoOnce)
	return fontInfo
}

func loadFontInfoOnce() {
	fontInfo.Family = "Ark Pixel"
	fontInfo.Variant = "10px Prop latin Regular"
	fontInfo.Ascent = ark10pxMap.Ascent()
	fontInfo.Descent = ark10pxMap.Descent()
	fontInfo.LineGap = ark10pxMap.LineGap()
	fontInfo.CapHeight = ark10pxMap.CapHeight()
	fontInfo.MidHeight = ark10pxMap.MidHeight()
	fontInfo.SpaceWidth = ark10pxMap.SpaceWidth()
}

// ScaleOf returns the font scale that matches the given font size.
func (i FontInfo) ScaleOf(size float32) float32 {
	return size / (float32(i.Ascent) + float32(i.Descent))
}

func (i FontInfo) Height() int {
	return int(i.Ascent) + int(i.Descent)
}
