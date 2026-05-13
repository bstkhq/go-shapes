package shapes

import (
	"slices"

	"github.com/hajimehoshi/ebiten/v2"
)

// quad must be given in clockwise order starting from top-left.
func (r *Renderer) mapQuad2(target, source *ebiten.Image, quad [4]PointF32) {
	tox, toy := rectOriginF32(target.Bounds())
	for i, pt := range quad {
		r.vertices[i].DstX = tox + pt.X
		r.vertices[i].DstY = toy + pt.Y
	}

	minX, minY, srcWidth, srcHeight := rectOriginSizeF32(source.Bounds())
	r.setSrcRectCoords(minX, minY, minX+srcWidth, minY+srcHeight)
	r.setFlatCustomVAs01(1.0, 1.0)
	r.opts.Images[0] = source
	target.DrawTrianglesShader32(r.vertices[:], r.indices[:], shaderMapBilinear.Load(), &r.opts)
	r.opts.Images[0] = nil
}

// MapQuad4 draws the given source texture into the given quad using 4
// triangles. This will produce noticeable texture projection distortions,
// but it's not as bad as using just two triangles and can work well
// enough in some cases. Otherwise, consider [Renderer.MapQuad]().
//
// quad must be given in clockwise order starting from top-left.
//
// The renderer's color is applied multiplicatively as a color scale;
// set it to white for neutral operation.
func (r *Renderer) MapQuad4(target, source *ebiten.Image, quad [4]PointF32) {
	tox, toy := rectOriginF32(target.Bounds())
	for i, pt := range quad {
		r.vertices[i].DstX = pt.X + tox
		r.vertices[i].DstY = pt.Y + toy
	}
	ctr := quadCenter(quad)
	ctrVert := r.vertices[0]
	ctrVert.DstX = ctr.X + tox
	ctrVert.DstY = ctr.Y + toy

	minX, minY, srcWidth, srcHeight := rectOriginSizeF32(source.Bounds())
	ctrVert.SrcX = minX + srcWidth/2.0
	ctrVert.SrcY = minX + srcHeight/2.0
	r.vertices = append(r.vertices, ctrVert)

	r.setSrcRectCoords(minX, minY, minX+srcWidth, minY+srcHeight)
	r.setFlatCustomVAs01(1.0, 1.0)
	r.opts.Images[0] = source
	r.indices = r.indices[:0]
	r.indices = slices.Grow(r.indices, 12)[:12]
	r.indices[0] = 0
	r.indices[1] = 1
	r.indices[2] = 4
	r.indices[3] = 1
	r.indices[4] = 2
	r.indices[5] = 4
	r.indices[6] = 2
	r.indices[7] = 3
	r.indices[8] = 4
	r.indices[9] = 3
	r.indices[10] = 0
	r.indices[11] = 4
	target.DrawTrianglesShader32(r.vertices[:], r.indices, shaderMapBilinear.Load(), &r.opts)
	r.restoreIndices()
	r.opts.Images[0] = nil
	r.vertices = r.vertices[:4]
}

func quadCenter(quad [4]PointF32) PointF32 {
	sumX := quad[0].X + quad[1].X + quad[2].X + quad[3].X
	sumY := quad[0].Y + quad[1].Y + quad[2].Y + quad[3].Y
	return PointF32{X: sumX / 4.0, Y: sumY / 4.0}
}

// MapProjective draws the given source texture into the given quad.
// This function computes the homography between the quad and the texture
// space, which involves solving an 8x8 equation system. This can be
// somewhat CPU heavy, so avoid drawing more than ~100 elements with it
// if you are not targeting powerful devices.
//
// quad must be given in clockwise order starting from top-left.
//
// The anisotropic flag can be set to true to reduce texture distortion
// at extreme angles, at the price of sampling 8 times instead of 4.
//
// To avoid jaggy edges, it's recommended to have one pixel of transparent
// padding in the source texture.
//
// The renderer's color is applied multiplicatively as a color scale;
// set it to white for neutral operation.
func (r *Renderer) MapQuad(target, source *ebiten.Image, quad [4]PointF32, anisotropic bool) {
	uvQuad := [4]PointF32{{X: 0, Y: 0}, {X: 1, Y: 0}, {X: 1, Y: 1}, {X: 0, Y: 1}}
	homography := computeHomography(quad, uvQuad)

	tox, toy := rectOriginF32(target.Bounds())
	for i, pt := range quad {
		r.vertices[i].DstX = tox + pt.X
		r.vertices[i].DstY = toy + pt.Y
	}

	minX, minY, srcWidth, srcHeight := rectOriginSizeF32(source.Bounds())
	r.setSrcRectCoords(minX, minY, minX+srcWidth, minY+srcHeight)
	r.opts.Uniforms["Homography"] = [9]float32{ // use column-major order
		homography[0], homography[3], homography[6],
		homography[1], homography[4], homography[7],
		homography[2], homography[5], homography[8],
	}
	r.opts.Images[0] = source
	var shader *ebiten.Shader
	if anisotropic {
		shader = shaderMapQuadAnisotropic.Load()
	} else {
		shader = shaderMapQuadBilinear.Load()
	}
	target.DrawTrianglesShader32(r.vertices[:], r.indices[:], shader, &r.opts)
	r.opts.Images[0] = nil
	clear(r.opts.Uniforms)
}
