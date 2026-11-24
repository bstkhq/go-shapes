package shapes

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
)

// AlphaNormalsCompute computes the alpha descent direction (the direction in
// which alpha decreases fastest) and the boundary offset for each pixel of
// source, and stores them in target, which must be the same size as source.
//
// On antialiased solid shapes, the descent direction is the normal of the shape
// boundary, pointing out of the shape, and the boundary offset is the signed
// distance from the pixel center to that boundary.
//
// This can be used to measure distances to the actual shape boundaries instead
// of to pixel centers, which can improve the results of jumping flood methods.
//
// Even when source has varying opacity, alpha is normalized locally to better
// estimate boundaries. See shaders/analysis/alpha_normals.kage for further
// encoding details and limitations.
//
// Notice that techniques like gamma correction can slightly skew "pixel
// coverage" and boundary detection.
//
// See also [Renderer.AlphaNormalsViz]().
func (r *Renderer) AlphaNormalsCompute(target, source *ebiten.Image) {
	sx, sy, sw, sh := rectOriginSizeF32(source.Bounds())
	tx, ty, tw, th := rectOriginSizeF32(target.Bounds())
	if sw != tw || sh != th {
		panic(fmt.Sprintf("source size != target size (%.0fx%.0f != %.0fx%.0f)", sw, sh, tw, th))
	}

	r.setDstRectCoords(tx, ty, tx+tw, ty+th)
	r.setSrcRectCoords(sx, sy, sx+sw, sy+sh)
	memoBlend := r.opts.Blend
	r.opts.Blend = ebiten.BlendCopy
	r.opts.Images[0] = source
	target.DrawTrianglesShader32(r.vertices[:], r.indices[:], shaderAlphaNormals.Load(), &r.opts)
	r.opts.Images[0] = nil
	r.opts.Blend = memoBlend
}

// AlphaNormalsViz is a debug method to visualize a map computed with
// [Renderer.AlphaNormalsCompute](). Hue shows the alpha descent direction, and
// brightness the side of the boundary the pixel center is on (bright on the
// high alpha side, dim on the low one). Pixels without data are left
// transparent.
func (r *Renderer) AlphaNormalsViz(target, normals *ebiten.Image, ox, oy float32) {
	r.DrawImgShader(target, normals, ox, oy, NoMargins, RegionExact, shaderAlphaNormalsViz.Load())
}
