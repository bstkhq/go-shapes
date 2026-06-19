package shapes

import (
	"image"
	"image/color"
	"sync"
	"sync/atomic"

	"github.com/hajimehoshi/ebiten/v2"
)

var paintVerts []ebiten.Vertex
var paintIndices []uint16
var paintOpts *ebiten.DrawTrianglesOptions
var paintInUse uint32

var whiteImgInit sync.Once
var whiteImg *ebiten.Image

func initWhiteImg() {
	whiteImg = ebiten.NewImage(3, 3)
	whiteImg.Fill(color.White)
}

var White [4]float32 = [4]float32{1, 1, 1, 1}
var white [4]float32 = [4]float32{1, 1, 1, 1} // safer internal use, no one can modify

// Paint is a utility method to draw rect over target using the given color and blend
// mode. This is more flexible than ebiten.SubImage(rect).(*ebiten.Image).Fill and
// avoids the subimage creation. See also [Clear].
//
// The rect coordinates are global; to draw the rect at the top left of target, rect.Min
// must match target.Min.
func Paint(target *ebiten.Image, rect image.Rectangle, rgba [4]float32, blend ebiten.Blend) {
	// Paint is implemented using DrawTriangles instead of DrawTrianglesShader due
	// to DrawTriangles being generally friendlier for batching, even though that's
	// not necessarily more lightweight in isolation

	// target bounds are applied here as rect already has global coordinates
	o, f := RectPointsF32(rect)

	whiteImgInit.Do(initWhiteImg)
	if atomic.CompareAndSwapUint32(&paintInUse, 0, 1) {
		if len(paintVerts) == 0 {
			paintVerts = make([]ebiten.Vertex, 4)
			setVertSrcCoords(paintVerts, 1.5, 1.5, 1.5, 1.5)
			paintIndices = []uint16{0, 1, 2, 2, 3, 0}
			paintOpts = &ebiten.DrawTrianglesOptions{}
		}
		paintOpts.Blend = blend
		if blend != ebiten.BlendClear {
			setVertColors(paintVerts, rgba)
		}
		setVertDstCoords(paintVerts, o.X, o.Y, f.X, f.Y)
		target.DrawTriangles(paintVerts[:], paintIndices[:], whiteImg, paintOpts)
		atomic.StoreUint32(&paintInUse, 0)
	} else { // concurrent case, hope for stack allocs
		verts := make([]ebiten.Vertex, 4)
		if blend != ebiten.BlendClear {
			setVertColors(verts, rgba)
		}
		setVertDstCoords(verts, o.X, o.Y, f.X, f.Y)
		setVertSrcCoords(verts, 1.5, 1.5, 1.5, 1.5)
		target.DrawTriangles(verts[:], []uint16{0, 1, 2, 2, 3, 0}[:], whiteImg, &ebiten.DrawTrianglesOptions{Blend: blend})
	}
}

// Clear is equivalent to [Paint](target, rect, [4]float32{1, 1, 1, 1}, ebiten.BlendClear)
func Clear(target *ebiten.Image, rect image.Rectangle) {
	Paint(target, rect, white, ebiten.BlendClear)
}

func setVertexColor(vertex *ebiten.Vertex, r, g, b, a float32) {
	vertex.ColorR = r
	vertex.ColorG = g
	vertex.ColorB = b
	vertex.ColorA = a
}

func setVertColors(verts []ebiten.Vertex, rgba [4]float32) {
	for i := range verts {
		verts[i].ColorR = rgba[0]
		verts[i].ColorG = rgba[1]
		verts[i].ColorB = rgba[2]
		verts[i].ColorA = rgba[3]
	}
}

// assumes clockwise vertex ordering (0 = TL, 1 = TR, 2 = BR, 3 = RL)
func setVertDstCoords(verts []ebiten.Vertex, minX, minY, maxX, maxY float32) {
	verts[0].DstX = minX // TL
	verts[0].DstY = minY // TL
	verts[1].DstX = maxX // TR
	verts[1].DstY = minY // TR
	verts[2].DstX = maxX // BR
	verts[2].DstY = maxY // BR
	verts[3].DstX = minX // BL
	verts[3].DstY = maxY // BL
}

func setVertDstCoordsIdx(verts []ebiten.Vertex, idx int, minX, minY, maxX, maxY float32) {
	verts[idx+0].DstX = minX // TL
	verts[idx+0].DstY = minY // TL
	verts[idx+1].DstX = maxX // TR
	verts[idx+1].DstY = minY // TR
	verts[idx+2].DstX = maxX // BR
	verts[idx+2].DstY = maxY // BR
	verts[idx+3].DstX = minX // BL
	verts[idx+3].DstY = maxY // BL
}

func setVertSrcCoords(verts []ebiten.Vertex, minX, minY, maxX, maxY float32) {
	verts[0].SrcX = minX // TL
	verts[0].SrcY = minY // TL
	verts[1].SrcX = maxX // TR
	verts[1].SrcY = minY // TR
	verts[2].SrcX = maxX // BR
	verts[2].SrcY = maxY // BR
	verts[3].SrcX = minX // BL
	verts[3].SrcY = maxY // BL
}

func setVertSrcCoordsIdx(verts []ebiten.Vertex, idx int, minX, minY, maxX, maxY float32) {
	verts[idx+0].SrcX = minX // TL
	verts[idx+0].SrcY = minY // TL
	verts[idx+1].SrcX = maxX // TR
	verts[idx+1].SrcY = minY // TR
	verts[idx+2].SrcX = maxX // BR
	verts[idx+2].SrcY = maxY // BR
	verts[idx+3].SrcX = minX // BL
	verts[idx+3].SrcY = maxY // BL
}
