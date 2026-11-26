package pixcheck

import (
	"fmt"
	"image"
	"image/color"
	"strconv"
	"strings"
)

// Parse returns an image with the pixels given in text, one row per line.
// Pixels are separated by spaces, and blank lines and indentation are ignored,
// so they can be written aligned. parsePixel converts each one from text. It
// panics if the text has no pixels, rows have different lengths or parsePixel
// fails.
//
// See [ParseAlpha] for an example.
func Parse(text string, parsePixel func(string) (color.RGBA, error)) *image.RGBA {
	var rows [][]color.RGBA
	for i, line := range strings.Split(text, "\n") {
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		row := make([]color.RGBA, len(fields))
		for x, field := range fields {
			pixel, err := parsePixel(field)
			if err != nil {
				panic(fmt.Sprintf("pixcheck: invalid pixel %q in line %d: %v", field, i+1, err))
			}
			row[x] = pixel
		}
		if len(rows) > 0 && len(row) != len(rows[0]) {
			panic(fmt.Sprintf("pixcheck: line %d has %d pixels, but the first row has %d", i+1, len(row), len(rows[0])))
		}
		rows = append(rows, row)
	}
	if len(rows) == 0 {
		panic("pixcheck: no pixels")
	}

	img := image.NewRGBA(image.Rect(0, 0, len(rows[0]), len(rows)))
	for y, row := range rows {
		for x, pixel := range row {
			img.SetRGBA(x, y, pixel)
		}
	}
	return img
}

// Format returns the pixels of img in the format read by [Parse], with
// formatPixel converting each one to text. Pixels are right-aligned so the
// columns line up.
func Format(img *image.RGBA, formatPixel func(color.RGBA) string) string {
	var cells []string // in row order
	var width int
	for y := img.Rect.Min.Y; y < img.Rect.Max.Y; y++ {
		for x := img.Rect.Min.X; x < img.Rect.Max.X; x++ {
			cell := formatPixel(img.RGBAAt(x, y))
			cells = append(cells, cell)
			width = max(width, len(cell))
		}
	}

	var text strings.Builder
	w, h := img.Rect.Dx(), img.Rect.Dy()
	for y := range h {
		if y > 0 {
			text.WriteByte('\n')
		}
		for x := range w {
			if x > 0 {
				text.WriteByte(' ')
			}
			fmt.Fprintf(&text, "%*s", width, cells[y*w+x])
		}
	}
	return text.String()
}

// ParseAlpha is [Parse] for alpha values from 0 to 255:
//
//	img := pixcheck.ParseAlpha(`
//		000 064 255
//		064 255 255
//	`)
//
// Colors are premultiplied white, so each pixel has R = G = B = A.
func ParseAlpha(text string) *image.RGBA {
	return Parse(text, func(field string) (color.RGBA, error) {
		alpha, err := strconv.ParseUint(field, 10, 8)
		return color.RGBA{uint8(alpha), uint8(alpha), uint8(alpha), uint8(alpha)}, err
	})
}

// FormatAlpha is [Format] for alpha values, with 3 digits each. Other channels
// are ignored.
func FormatAlpha(img *image.RGBA) string {
	return Format(img, formatAlphaPixel)
}

func formatAlphaPixel(pixel color.RGBA) string {
	return fmt.Sprintf("%03d", pixel.A)
}
