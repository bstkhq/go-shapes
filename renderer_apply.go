package shapes

import (
	"math"

	"github.com/hajimehoshi/ebiten/v2"
)

// Precondition: thickness can't exceed 32.
//
// WARNING: this is a quadratic algorithm on GPU. For large expansions,
// consider [Renderer.ApplyExpansionRect]() or [Renderer.JFMExpansion]()
// instead, but both of those are only useful in specific situations.
func (r *Renderer) ApplyExpansion(target *ebiten.Image, mask *ebiten.Image, ox, oy, thickness float32) {
	if thickness > 32 {
		panic("thickness can't exceed 32")
	}

	srcBounds := mask.Bounds()
	srcWidth, srcHeight := float32(srcBounds.Dx()), float32(srcBounds.Dy())
	ht32 := thickness / 2.0
	dstBounds := target.Bounds()
	dstMinX, dstMinY := float32(dstBounds.Min.X), float32(dstBounds.Min.Y)
	minX, minY := dstMinX+ox-ht32, dstMinY+oy-ht32
	maxX, maxY := dstMinX+ox+srcWidth+ht32, dstMinY+oy+srcHeight+ht32
	r.setDstRectCoords(minX-1, minY-1, maxX+1, maxY+1)

	srcMinX, srcMinY := float32(srcBounds.Min.X), float32(srcBounds.Min.Y)
	srcMaxX, srcMaxY := float32(srcBounds.Max.X), float32(srcBounds.Max.Y)
	r.setSrcRectCoords(srcMinX-ht32-1, srcMinY-ht32-1, srcMaxX+ht32+1, srcMaxY+ht32+1)
	r.setFlatCustomVA0(thickness)

	// draw shader
	r.opts.Images[0] = mask
	ensureShaderExpansionLoaded()
	target.DrawTrianglesShader(r.vertices[:], r.indices[:], shaderExpansion, &r.opts)
	r.opts.Images[0] = nil
}

// ApplyExpansionRect performs double pass expansion with a square kernel.
// This is less general but more efficient than [Renderer.ApplyExpansion]().
//
// Precondition: thickness can't exceed 32.
//
// This function uses one internal offscreen (#0), and target and mask
// can be on the same internal atlas.
func (r *Renderer) ApplyExpansionRect(target *ebiten.Image, mask *ebiten.Image, ox, oy, thickness float32) {
	if thickness > 32 {
		panic("thickness can't exceed 32")
	}

	// first pass (vert)
	thickCeil := float32(math.Ceil(float64(thickness)))
	sx, sy, sw, sh := rectOriginSize(mask.Bounds())
	temp := r.getTemp(0, sw, sh+int(thickCeil)*2.0, false)
	sx32, sy32, sw32, sh32 := float32(sx), float32(sy), float32(sw), float32(sh)
	memoBlend := r.opts.Blend
	r.opts.Blend = ebiten.BlendCopy
	r.setSrcRectCoords(sx32, sy32-thickCeil, sx32+sw32, sy32+sh32+thickCeil)
	r.setDstRectCoords(0, 0, sw32, sh32+thickCeil*2)
	r.setFlatCustomVA0(thickness)
	r.opts.Images[0] = mask
	ensureShaderExpansionVertLoaded()
	temp.DrawTrianglesShader(r.vertices[:], r.indices[:], shaderExpansionVert, &r.opts)
	r.opts.Images[0] = nil

	// second pass (horz)
	r.opts.Blend = memoBlend
	r.setSrcRectCoords(-thickCeil, 0, sw32+thickCeil, sh32+thickCeil*2.0)
	dx, dy := rectOriginF32(target.Bounds())
	ox += dx
	oy += dy
	r.setDstRectCoords(ox-thickCeil, oy-thickCeil, ox+sw32+thickCeil, oy+sh32+thickCeil)
	r.opts.Images[0] = temp
	ensureShaderExpansionHorzLoaded()
	target.DrawTrianglesShader(r.vertices[:], r.indices[:], shaderExpansionHorz, &r.opts)
	r.opts.Images[0] = nil
}

// Precondition: thickness can't exceed 32.
//
// WARNING: this is a quadratic algorithm on GPU. For large erosions,
// consider [Renderer.JFMErosion]() instead.
func (r *Renderer) ApplyErosion(target *ebiten.Image, mask *ebiten.Image, ox, oy, thickness float32) {
	if thickness > 32 {
		panic("thickness can't exceed 32")
	}

	srcBounds := mask.Bounds()
	srcWidth, srcHeight := float32(srcBounds.Dx()), float32(srcBounds.Dy())
	dstBounds := target.Bounds()
	dstMinX, dstMinY := float32(dstBounds.Min.X), float32(dstBounds.Min.Y)
	minX, minY := dstMinX+ox, dstMinY+oy
	maxX, maxY := dstMinX+ox+srcWidth, dstMinY+oy+srcHeight
	r.setDstRectCoords(minX-1, minY-1, maxX+1.0, maxY+1.0)

	srcMinX, srcMinY := float32(srcBounds.Min.X), float32(srcBounds.Min.Y)
	srcMaxX, srcMaxY := float32(srcBounds.Max.X), float32(srcBounds.Max.Y)
	r.setSrcRectCoords(srcMinX-1, srcMinY-1, srcMaxX+1.0, srcMaxY+1.0)
	r.setFlatCustomVA0(thickness)

	// draw shader
	r.opts.Images[0] = mask
	ensureShaderErosionLoaded()
	target.DrawTrianglesShader(r.vertices[:], r.indices[:], shaderErosion, &r.opts)
	r.opts.Images[0] = nil
}

// Precondition: thickness can't exceed 32.
//
// WARNING: this is a quadratic algorithm on GPU. For large outlines,
// consider [Renderer.JFMOutline]() instead.
func (r *Renderer) ApplyOutline(target *ebiten.Image, mask *ebiten.Image, ox, oy, thickness float32) {
	if thickness > 32 {
		panic("thickness can't exceed 32")
	}

	srcBounds := mask.Bounds()
	srcWidth, srcHeight := float32(srcBounds.Dx()), float32(srcBounds.Dy())
	ht32 := thickness / 2.0
	dstBounds := target.Bounds()
	dstMinX, dstMinY := float32(dstBounds.Min.X), float32(dstBounds.Min.Y)
	minX, minY := dstMinX+ox-ht32, dstMinY+oy-ht32
	maxX, maxY := dstMinX+ox+srcWidth+ht32, dstMinY+oy+srcHeight+ht32
	r.setDstRectCoords(minX-1, minY-1, maxX+1, maxY+1)

	srcMinX, srcMinY := float32(srcBounds.Min.X), float32(srcBounds.Min.Y)
	srcMaxX, srcMaxY := float32(srcBounds.Max.X), float32(srcBounds.Max.Y)
	r.setSrcRectCoords(srcMinX-ht32-1, srcMinY-ht32-1, srcMaxX+ht32+1.0, srcMaxY+ht32+1.0)
	r.setFlatCustomVA0(thickness)

	// draw shader
	r.opts.Images[0] = mask
	ensureShaderOutlineLoaded()
	target.DrawTrianglesShader(r.vertices[:], r.indices[:], shaderOutline, &r.opts)
	r.opts.Images[0] = nil
}

// ApplyBlur applies a gaussian blur to the given mask and draws it onto the given target.
// colorMix = 0 will use the renderer's vertex colors; colorMix = 1 will use the original mask colors.
//
// WARNING: this is a quadratic algorithm on GPU. For radiuses above 4, you want to look at
// [Renderer.ApplyBlur2]() or [Renderer.ApplyBlurD4]().
func (r *Renderer) ApplyBlur(target *ebiten.Image, mask *ebiten.Image, ox, oy, radius, colorMix float32) {
	if radius > 32 {
		panic("radius can't exceed 32")
	}
	if radius < 0 {
		panic("radius can't be negative")
	}

	srcBounds := mask.Bounds()
	srcWidth, srcHeight := float32(srcBounds.Dx()), float32(srcBounds.Dy())
	hr32 := radius / 2.0
	dstBounds := target.Bounds()
	dstMinX, dstMinY := float32(dstBounds.Min.X), float32(dstBounds.Min.Y)
	minX, minY := dstMinX+ox-hr32, dstMinY+oy-hr32
	maxX, maxY := dstMinX+ox+srcWidth+hr32, dstMinY+oy+srcHeight+hr32
	r.setDstRectCoords(minX-1, minY-1, maxX+1, maxY+1)

	srcMinX, srcMinY := float32(srcBounds.Min.X), float32(srcBounds.Min.Y)
	srcMaxX, srcMaxY := float32(srcBounds.Max.X), float32(srcBounds.Max.Y)
	r.setSrcRectCoords(srcMinX-hr32-1, srcMinY-hr32-1, srcMaxX+hr32+1.0, srcMaxY+hr32+1.0)
	r.setFlatCustomVAs01(radius, colorMix)

	// draw shader
	r.opts.Images[0] = mask
	ensureShaderBlurLoaded()
	target.DrawTrianglesShader(r.vertices[:], r.indices[:], shaderBlur, &r.opts)
	r.opts.Images[0] = nil
}

// ApplyBlur2 is similar to ApplyBlur, but uses two 1D passes instead of a single 2D pass.
// This greatly reduces the amount of sampled pixels for the shader, and despite breaking
// batching tends to be much more efficient than [Renderer.ApplyBlur]().
//
// This function uses one internal offscreen (#0).
func (r *Renderer) ApplyBlur2(target *ebiten.Image, mask *ebiten.Image, ox, oy, radius, colorMix float32) {
	if radius > 32 {
		panic("radius can't exceed 32")
	}
	if radius < 0 {
		panic("radius can't be negative")
	}

	srcBounds := mask.Bounds()
	w32, h32 := float32(srcBounds.Dx()), float32(srcBounds.Dy())+radius
	w, h := int(w32), int(math.Ceil(float64(h32)))
	tmp := r.getTemp(0, w, h, false)
	preBlend := r.opts.Blend
	r.opts.Blend = ebiten.BlendCopy
	hrCeil := ceilF32(radius / 2.0)
	r.ApplyVertBlur(tmp, mask, 0, hrCeil, radius, 1.0)
	r.opts.Blend = preBlend
	r.ApplyHorzBlur(target, tmp, ox, oy-hrCeil, radius, colorMix)
}

func (r *Renderer) ApplyVertBlur(target *ebiten.Image, mask *ebiten.Image, ox, oy, radius, colorMix float32) {
	if radius > 32 {
		panic("radius can't exceed 32")
	}
	if radius < 0 {
		panic("radius can't be negative")
	}

	srcBounds := mask.Bounds()
	srcWidth, srcHeight := float32(srcBounds.Dx()), float32(srcBounds.Dy())
	dstBounds := target.Bounds()
	dstMinX, dstMinY := float32(dstBounds.Min.X), float32(dstBounds.Min.Y)
	hrCeil := ceilF32(radius / 2.0)
	minX, minY := dstMinX+ox, dstMinY+oy-hrCeil
	maxX, maxY := dstMinX+ox+srcWidth, dstMinY+oy+srcHeight+hrCeil
	r.setDstRectCoords(minX, minY, maxX, maxY)

	srcMinX, srcMinY := float32(srcBounds.Min.X), float32(srcBounds.Min.Y)
	srcMaxX, srcMaxY := float32(srcBounds.Max.X), float32(srcBounds.Max.Y)
	r.setSrcRectCoords(srcMinX, srcMinY-hrCeil, srcMaxX, srcMaxY+hrCeil)
	r.setFlatCustomVAs01(radius, colorMix)

	// draw shader
	r.opts.Images[0] = mask
	ensureShaderVertBlurLoaded()
	target.DrawTrianglesShader(r.vertices[:], r.indices[:], shaderVertBlur, &r.opts)
	r.opts.Images[0] = nil
}

func (r *Renderer) ApplyHorzBlur(target *ebiten.Image, mask *ebiten.Image, ox, oy, radius, colorMix float32) {
	if radius > 32 {
		panic("radius can't exceed 32")
	}
	if radius < 0 {
		panic("radius can't be negative")
	}

	srcBounds := mask.Bounds()
	srcWidth, srcHeight := float32(srcBounds.Dx()), float32(srcBounds.Dy())
	hrCeil := ceilF32(radius / 2.0)
	dstBounds := target.Bounds()
	dstMinX, dstMinY := float32(dstBounds.Min.X), float32(dstBounds.Min.Y)
	minX, minY := dstMinX+ox-hrCeil, dstMinY+oy
	maxX, maxY := dstMinX+ox+srcWidth+hrCeil, dstMinY+oy+srcHeight
	r.setDstRectCoords(minX, minY, maxX, maxY)

	srcMinX, srcMinY := float32(srcBounds.Min.X), float32(srcBounds.Min.Y)
	srcMaxX, srcMaxY := float32(srcBounds.Max.X), float32(srcBounds.Max.Y)
	r.setSrcRectCoords(srcMinX-hrCeil, srcMinY, srcMaxX+hrCeil, srcMaxY)
	r.setFlatCustomVAs01(radius, colorMix)

	// draw shader
	r.opts.Images[0] = mask
	ensureShaderHorzBlurLoaded()
	target.DrawTrianglesShader(r.vertices[:], r.indices[:], shaderHorzBlur, &r.opts)
	r.opts.Images[0] = nil
}

type Clamping uint8

const (
	ClampNone   Clamping = 0b0000
	ClampTop    Clamping = 0b1000
	ClampBottom Clamping = 0b0100
	ClampLeft   Clamping = 0b0010
	ClampRight  Clamping = 0b0001

	ClampTopLeft     Clamping = ClampTop | ClampLeft
	ClampTopRight    Clamping = ClampTop | ClampRight
	ClampBottomLeft  Clamping = ClampBottom | ClampLeft
	ClampBottomRight Clamping = ClampBottom | ClampRight
)

func (r *Renderer) ApplyHardShadow(target *ebiten.Image, mask *ebiten.Image, ox, oy, xOffset, yOffset float32, clamping Clamping) {
	dstBounds, srcBounds := target.Bounds(), mask.Bounds()
	srcWidth, srcHeight := float32(srcBounds.Dx()), float32(srcBounds.Dy())
	dstMinX, dstMinY := float32(dstBounds.Min.X), float32(dstBounds.Min.Y)
	leftOff, topOff := min(0, xOffset), min(0, yOffset)
	rightOff, bottomOff := max(0, xOffset), max(0, yOffset)
	minX, minY := dstMinX+ox+leftOff, dstMinY+oy+topOff
	maxX, maxY := dstMinX+srcWidth+ox+rightOff, dstMinY+srcHeight+oy+bottomOff
	r.setDstRectCoords(minX, minY, maxX, maxY)

	srcMinX, srcMinY := float32(srcBounds.Min.X), float32(srcBounds.Min.Y)
	srcMaxX, srcMaxY := float32(srcBounds.Max.X), float32(srcBounds.Max.Y)
	r.setSrcRectCoords(srcMinX+leftOff, srcMinY+topOff, srcMaxX+rightOff, srcMaxY+bottomOff)
	r.setFlatCustomVAs(xOffset, yOffset, float32(clamping), 0)

	// draw shader
	r.opts.Images[0] = mask
	ensureShaderHardShadowLoaded()
	target.DrawTrianglesShader(r.vertices[:], r.indices[:], shaderHardShadow, &r.opts)
	r.opts.Images[0] = nil
}

func (r *Renderer) ApplyShadow(target *ebiten.Image, mask *ebiten.Image, ox, oy, xOffset, yOffset, radius float32, clamping Clamping) {
	if radius > 32 {
		panic("radius can't exceed 32")
	}
	if radius < 0 {
		panic("radius can't be negative")
	}

	dstBounds, srcBounds := target.Bounds(), mask.Bounds()
	srcWidth, srcHeight := float32(srcBounds.Dx()), float32(srcBounds.Dy())
	dstMinX, dstMinY := float32(dstBounds.Min.X), float32(dstBounds.Min.Y)
	leftOff, topOff := min(0, xOffset), min(0, yOffset)
	rightOff, bottomOff := max(0, xOffset), max(0, yOffset)
	hr32 := radius / 2.0
	topHR32, bottomHR32, leftHR32, rightHR32 := hr32, hr32, hr32, hr32
	if clamping&ClampBottom != 0 {
		bottomHR32 = 0
	}
	if clamping&ClampTop != 0 {
		topHR32 = 0
	}
	if clamping&ClampLeft != 0 {
		leftHR32 = 0
	}
	if clamping&ClampRight != 0 {
		rightHR32 = 0
	}

	minX, minY := dstMinX+ox+leftOff-leftHR32, dstMinY+oy+topOff-topHR32
	maxX, maxY := dstMinX+srcWidth+ox+rightOff+rightHR32, dstMinY+srcHeight+oy+bottomOff+bottomHR32
	r.setDstRectCoords(minX, minY, maxX, maxY)

	srcMinX, srcMinY := float32(srcBounds.Min.X), float32(srcBounds.Min.Y)
	srcMaxX, srcMaxY := float32(srcBounds.Max.X), float32(srcBounds.Max.Y)
	r.setSrcRectCoords(srcMinX+leftOff-leftHR32, srcMinY+topOff-topHR32, srcMaxX+rightOff+rightHR32, srcMaxY+bottomOff+bottomHR32)
	r.setFlatCustomVAs(xOffset, yOffset, radius, float32(clamping))

	// draw shader
	r.opts.Images[0] = mask
	ensureShaderShadowLoaded()
	target.DrawTrianglesShader(r.vertices[:], r.indices[:], shaderShadow, &r.opts)
	r.opts.Images[0] = nil
}

func (r *Renderer) ApplyZoomShadow(target *ebiten.Image, mask *ebiten.Image, ox, oy, xOffset, yOffset, zoom float32, clamping Clamping) {
	if zoom < 1.0 || zoom > 16.0 {
		panic("zoom must be in [1, 16] range")
	}

	dstBounds, srcBounds := target.Bounds(), mask.Bounds()
	srcWidth, srcHeight := float32(srcBounds.Dx()), float32(srcBounds.Dy())
	dstMinX, dstMinY := float32(dstBounds.Min.X), float32(dstBounds.Min.Y)
	leftOff, topOff := min(0, xOffset), min(0, yOffset)
	rightOff, bottomOff := max(0, xOffset), max(0, yOffset)
	var topCut, leftCut, bottomCut, rightCut float32
	topPad, leftPad := srcHeight*0.5*(zoom-1.0), srcWidth*0.5*(zoom-1.0)
	bottomPad, rightPad := topPad, leftPad
	if clamping&ClampBottom != 0 {
		bottomCut = bottomPad / zoom
		bottomPad = 0
	}
	if clamping&ClampTop != 0 {
		topCut = topPad / zoom
		topPad = 0
	}
	if clamping&ClampLeft != 0 {
		leftCut = leftPad / zoom
		leftPad = 0
	}
	if clamping&ClampRight != 0 {
		rightCut = rightPad / zoom
		rightPad = 0
	}

	minX, minY := dstMinX+ox+leftOff-leftPad, dstMinY+oy+topOff-topPad
	maxX, maxY := dstMinX+srcWidth+ox+rightOff+rightPad, dstMinY+srcHeight+oy+bottomOff+bottomPad
	r.setDstRectCoords(minX, minY, maxX, maxY)

	srcMinX, srcMinY := float32(srcBounds.Min.X), float32(srcBounds.Min.Y)
	srcMaxX, srcMaxY := float32(srcBounds.Max.X), float32(srcBounds.Max.Y)
	r.setSrcRectCoords(srcMinX+leftOff/zoom-leftCut, srcMinY+topOff/zoom-topCut, srcMaxX+rightOff/zoom-rightCut, srcMaxY+bottomOff/zoom-bottomCut)
	r.setFlatCustomVAs(xOffset/zoom, yOffset/zoom, float32(clamping), zoom)

	// draw shader
	r.opts.Images[0] = mask
	ensureShaderZoomShadowLoaded()
	target.DrawTrianglesShader(r.vertices[:], r.indices[:], shaderZoomShadow, &r.opts)
	r.opts.Images[0] = nil
}

// ApplySimpleGlow draws the given mask into the target, at the given coordinates, with
// an glow effect added. The effect mix intensity is determined by the renderer's color
// alphas. For finer control, see also [Renderer.ApplyGlow]().
func (r *Renderer) ApplySimpleGlow(target *ebiten.Image, mask *ebiten.Image, ox, oy, radius float32) {
	r.ApplyGlow(target, mask, ox, oy, radius, radius, 0.4, 0.7, 1.0)
}

// ApplyGlow draws a horizontal glow effect for the given mask into the target, at the
// given coordinates. The effect mix intensity is determined by the renderer's color alphas.
//
// Regarding the advanced control parameters:
//   - threshStart and threshEnd indicate the start luminosity threshold at which the glow
//     effect kicks in and the point at which it's fully active. threshStart must be <=
//     threshEnd, and the values must be in [0, 1] range.
//   - colorMix controls the glow's color. If 0, the glow color will be determined fully
//     by the renderer's vertex colors. If 1, the glow color will be determined by the original
//     mask colors. Any values in between will lead to linear interpolation.
//
// Notice that this effect uses an internal offscreen (#0) and two passes, and target and mask
// can be on the same internal atlas.
func (r *Renderer) ApplyGlow(target *ebiten.Image, mask *ebiten.Image, ox, oy, horzRadius, vertRadius, threshStart, threshEnd, colorMix float32) {
	if threshStart > threshEnd {
		panic("threshStart > threshEnd")
	}
	if horzRadius > 32 || vertRadius > 32 {
		panic("radius can't exceed 32")
	}

	srcBounds := mask.Bounds()
	srcWidth, srcHeight := float32(srcBounds.Dx()), float32(srcBounds.Dy())
	w32, h32 := float32(srcWidth), float32(srcHeight)+vertRadius
	w, h := int(w32), int(math.Ceil(float64(h32)))
	tmp := r.getTemp(0, w, h, false)

	hr32 := vertRadius / 2.0
	r.setDstRectCoords(0, 0, w32, h32+2)

	srcMinX, srcMinY := float32(srcBounds.Min.X), float32(srcBounds.Min.Y)
	srcMaxX, srcMaxY := float32(srcBounds.Max.X), float32(srcBounds.Max.Y)
	r.setSrcRectCoords(srcMinX, srcMinY-hr32-1, srcMaxX, srcMaxY+hr32+1.0)
	r.setFlatCustomVAs(vertRadius, threshStart, threshEnd, 1.0)

	// first pass (threshold + vertical blur)
	r.opts.Images[0] = mask
	preBlend := r.opts.Blend
	r.opts.Blend = ebiten.BlendCopy
	ensureShaderGlowFirstPassLoaded()
	tmp.DrawTrianglesShader(r.vertices[:], r.indices[:], shaderGlowFirstPass, &r.opts)
	r.opts.Images[0] = nil

	// second pass
	r.opts.Blend = ebiten.BlendLighter
	r.ApplyHorzBlur(target, tmp, ox, oy-hr32-1.0, horzRadius, colorMix)
	r.opts.Blend = preBlend
}

// ApplyHorzGlow draws a horizontal glow effect for the given mask into the target, at the
// given coordinates. See [Renderer.ApplyGlow]() for additional documentation. Comparedto
// Renderer.ApplyGlow, this effect only applies the glow horizontally and it's much cheaper,
// requiring no offscreen and a single pass.
func (r *Renderer) ApplyHorzGlow(target *ebiten.Image, mask *ebiten.Image, ox, oy, horzRadius, threshStart, threshEnd, colorMix float32) {
	if threshStart > threshEnd {
		panic("threshStart > threshEnd")
	}
	if horzRadius > 32 {
		panic("radius can't exceed 32")
	}

	srcBounds := mask.Bounds()
	srcWidth, srcHeight := float32(srcBounds.Dx()), float32(srcBounds.Dy())

	hr32 := horzRadius / 2.0
	r.setDstRectCoords(ox-hr32-1.0, oy, ox+float32(srcWidth)+hr32+1.0, oy+float32(srcHeight))

	srcMinX, srcMinY := float32(srcBounds.Min.X), float32(srcBounds.Min.Y)
	srcMaxX, srcMaxY := float32(srcBounds.Max.X), float32(srcBounds.Max.Y)
	r.setSrcRectCoords(srcMinX-hr32-1, srcMinY, srcMaxX+hr32+1, srcMaxY)
	r.setFlatCustomVAs(horzRadius, threshStart, threshEnd, colorMix)

	r.opts.Images[0] = mask
	ensureShaderHorzGlowLoaded()
	preBlend := r.opts.Blend
	r.opts.Blend = ebiten.BlendLighter
	target.DrawTrianglesShader(r.vertices[:], r.indices[:], shaderHorzGlow, &r.opts)
	r.opts.Blend = preBlend
	r.opts.Images[0] = nil
}

// ApplyDarkHorzGlow is the "negative" version of [Renderer.ApplyHorzGlow](). Instead of
// using an additive blending effect around high luminosity areas, it uses multiplicative
// blending around dark areas.
//
// Notice that unlike regular glow effects, dark glows expects threshStart >= threshEnd.
func (r *Renderer) ApplyDarkHorzGlow(target *ebiten.Image, mask *ebiten.Image, ox, oy, horzRadius, threshStart, threshEnd, colorMix float32) {
	if threshStart < threshEnd {
		panic("threshStart < threshEnd")
	}
	if horzRadius > 32 {
		panic("radius can't exceed 32")
	}

	srcBounds := mask.Bounds()
	srcWidth, srcHeight := float32(srcBounds.Dx()), float32(srcBounds.Dy())

	hr32 := horzRadius / 2.0
	r.setDstRectCoords(ox-hr32-1.0, oy, ox+float32(srcWidth)+hr32+1.0, oy+float32(srcHeight))

	srcMinX, srcMinY := float32(srcBounds.Min.X), float32(srcBounds.Min.Y)
	srcMaxX, srcMaxY := float32(srcBounds.Max.X), float32(srcBounds.Max.Y)
	r.setSrcRectCoords(srcMinX-hr32-1, srcMinY, srcMaxX+hr32+1, srcMaxY)
	r.setFlatCustomVAs(horzRadius, threshStart, threshEnd, colorMix)

	r.opts.Images[0] = mask
	ensureShaderDarkHorzGlowLoaded()
	preBlend := r.opts.Blend
	r.opts.Blend = BlendMultiply
	//r.opts.Blend = BlendSubtract // also possible with a shader flag, but multiply feels more natural
	target.DrawTrianglesShader(r.vertices[:], r.indices[:], shaderDarkHorzGlow, &r.opts)
	r.opts.Blend = preBlend
	r.opts.Images[0] = nil
}

type GaussKern uint8

const (
	GaussKern3 GaussKern = iota
	GaussKern5
	GaussKern7
	GaussKern9
	GaussKern11
	GaussKern13
	GaussKern15
	GaussKern17
)

func (k GaussKern) Radius() int {
	return (k.Size() - 1) >> 1
}

func (k GaussKern) Size() int {
	ik := int(k)
	return 3 + ik + ik
}

// gaussian kernels are encoded as one-dimension separable filters, with the values
// for the center and a single side (both sides have to be applied in the shader)
var gaussKerns = [][9]float32{ // binomial forms
	{0.5000000, 0.2500000, 0.0000000, 0.0000000, 0.0000000, 0.0000000, 0.0000000, 0.0000000, 0.0000000}, // 3x3
	{0.3750000, 0.2500000, 0.0625000, 0.0000000, 0.0000000, 0.0000000, 0.0000000, 0.0000000, 0.0000000}, // 5x5
	{0.3125000, 0.2343750, 0.0937500, 0.0156250, 0.0000000, 0.0000000, 0.0000000, 0.0000000, 0.0000000}, // 7x7
	{0.2734375, 0.2187500, 0.1093750, 0.0312500, 0.0039063, 0.0000000, 0.0000000, 0.0000000, 0.0000000}, // 9x9
	{0.2460937, 0.2050781, 0.1171875, 0.0439453, 0.0097656, 0.0009765, 0.0000000, 0.0000000, 0.0000000}, // 11x11
	{0.2255859, 0.1933593, 0.1208496, 0.0537109, 0.0161132, 0.0029296, 0.0002441, 0.0000000, 0.0000000}, // 13x13
	{0.2094726, 0.1832885, 0.1221923, 0.0610961, 0.0222167, 0.0055542, 0.0008545, 0.0000611, 0.0000000}, // 15x15
	{0.1963806, 0.1745605, 0.1221924, 0.0666504, 0.0277709, 0.0085449, 0.0018311, 0.0002441, 0.0001526}, // 17x17
}

// var gaussKerns = [][9]float32{ // true gaussians
// 	{0.7869860, 0.1065069, 0.0000000, 0.0000000, 0.0000000, 0.0000000, 0.0000000, 0.0000000, 0.0000000}, // 3x3
// 	{0.4026199, 0.2442013, 0.0544887, 0.0000000, 0.0000000, 0.0000000, 0.0000000, 0.0000000, 0.0000000}, // 5x5
// 	{0.2706821, 0.2167453, 0.1112808, 0.0366328, 0.0000000, 0.0000000, 0.0000000, 0.0000000, 0.0000000}, // 7x7
// 	{0.2041637, 0.1801738, 0.1238315, 0.0662822, 0.0276306, 0.0000000, 0.0000000, 0.0000000, 0.0000000}, // 9x9
// 	{0.1639672, 0.1513608, 0.1190646, 0.0798114, 0.0455890, 0.0221906, 0.0000000, 0.0000000, 0.0000000}, // 11x11
// 	{0.1370228, 0.1296181, 0.1097193, 0.0831085, 0.0563317, 0.0341669, 0.0185440, 0.0000000, 0.0000000}, // 13x13
// 	{0.1176958, 0.1129886, 0.0999667, 0.0815125, 0.0612548, 0.0424232, 0.0270778, 0.0159284, 0.0000000}, // 15x15
// 	{0.1031526, 0.0999789, 0.0910319, 0.0778637, 0.0625652, 0.0472267, 0.0334887, 0.0223083, 0.0139602}, // 17x17
// }

// ApplyBlurD4 is a less flexible form of blur, similar to [Renderer.ApplyBlur2](),
// that downscales the source x4 before applying a gaussian kernel. This blur
// implementation tends to be more efficient than ApplyBlur2 when it comes to less
// powerful hardware and large blur areas (it uses less memory and compute at the
// cost of more steps). When enough resources are available (e.g. most medium-sized
// or small blurs), ApplyBlur2 tends to be slightly more efficient than ApplyBlurD4.
//
// This function uses two internal offscreens (#0, #1), and target and mask can be on
// the same internal atlas.
func (r *Renderer) ApplyBlurD4(target *ebiten.Image, mask *ebiten.Image, ox, oy float32, horzKernel, vertKernel GaussKern, colorMix float32) {
	r.applyKernelD4(target, mask, ox, oy, horzKernel, vertKernel, func(downHorzTarget *ebiten.Image) {
		r.setFlatCustomVA0(colorMix)
		ensureShaderHorzBlurKernLoaded()
		downHorzTarget.DrawTrianglesShader(r.vertices[:], r.indices[:], shaderHorzBlurKern, &r.opts)
	}, false)
}

// ApplyGlowD4 is the multipass downscaling version of [Renderer.ApplyGlow]().
// See [Renderer.ApplyBlurD4]() for further docs and context.
//
// This function uses two internal offscreens (#0, #1), and target and mask can be on
// the same internal atlas.
func (r *Renderer) ApplyGlowD4(target *ebiten.Image, mask *ebiten.Image, ox, oy float32, horzKernel, vertKernel GaussKern, threshStart, threshEnd, colorMix float32) {
	if threshStart > threshEnd {
		panic("threshStart > threshEnd")
	}

	r.applyKernelD4(target, mask, ox, oy, horzKernel, vertKernel, func(downHorzTarget *ebiten.Image) {
		r.setFlatCustomVAs(threshStart, threshEnd, colorMix, 0)
		ensureShaderHorzGlowKernLoaded()
		downHorzTarget.DrawTrianglesShader(r.vertices[:], r.indices[:], shaderHorzGlowKern, &r.opts)
	}, true)
}

// ApplyColorGlowD4 is a color-specific version of [Renderer.ApplyGlowD4](), where glow
// intensity is determined by color similarity instead of lightness.
//
// This function uses two internal offscreens (#0, #1), and target and mask can be on
// the same internal atlas.
func (r *Renderer) ApplyColorGlowD4(target *ebiten.Image, mask *ebiten.Image, ox, oy float32, horzKernel, vertKernel GaussKern, rgb [3]float32, threshStart, threshEnd, colorMix float32) {
	if threshStart > threshEnd {
		panic("threshStart > threshEnd")
	}

	// note: quadratic approximations don't hold up well with threshStart close to zero
	// adjustedThreshStart := 1.0 - (threshEnd * threshEnd / 3.0)
	// adjustedThreshEnd := 1.0 - (threshStart * threshStart / 3.0)

	r.applyKernelD4(target, mask, ox, oy, horzKernel, vertKernel, func(downHorzTarget *ebiten.Image) {
		r.opts.Uniforms["RGB"] = rgb
		r.setFlatCustomVAs(threshStart, threshEnd, colorMix, 0)
		ensureShaderHorzColorGlowLoaded()
		downHorzTarget.DrawTrianglesShader(r.vertices[:], r.indices[:], shaderHorzColorGlow, &r.opts)
		clear(r.opts.Uniforms)
	}, true)
}

// Internal function used by ApplyBlurD4, ApplyGlowD4 and ApplyColorGlowD4. It downscales the
// mask, applies a custom horizontal kernel shader, then a standard vertical blur shader, and
// upscales the result back, optionally with a BlendLighter blend. At invokeShader, KernelLen
// and Kernel uniforms have been set, as well as the downscaled source image, but other uniforms
// and custom VAs have to be set during invocation.
//
// This function uses two internal offscreens (#0, #1), and target and mask can be on
// the same internal atlas.
func (r *Renderer) applyKernelD4(target *ebiten.Image, mask *ebiten.Image, ox, oy float32, horzKernel, vertKernel GaussKern, invokeShader func(downHorzTarget *ebiten.Image), lighterBlend bool) {
	// measures
	const downscaling = 4
	maskBounds := mask.Bounds()
	maskWidth, maskHeight := maskBounds.Dx(), maskBounds.Dy()
	maskW64, maskH64 := float64(maskWidth), float64(maskHeight)
	downW64, downH64 := maskW64/downscaling, maskH64/downscaling
	downImgWidth, downImgHeight := math.Ceil(downW64)+2, math.Ceil(downH64)+2

	halfHorzMargin, halfVertMargin := float64(horzKernel.Radius()), float64(vertKernel.Radius())
	horzMargin, vertMargin := halfHorzMargin*2.0, halfVertMargin*2.0
	dkernW64, dkernH64 := downW64+horzMargin, downH64+vertMargin
	dkernImgWidth, dkernImgHeight := math.Ceil(dkernW64)+2, math.Ceil(dkernH64)+2

	// get offscreens and smart clears
	dkern := r.getTemp(0, int(dkernImgWidth), int(dkernImgHeight), false) // get first as the biggest offscreen
	down := r.getTemp(0, int(downImgWidth), int(downImgHeight), false)    // shared with dkern
	dkernHorz := r.getTemp(1, int(dkernImgWidth), int(downImgHeight), false)
	preBlend := r.opts.Blend
	r.opts.Blend = ebiten.BlendClear
	r.StrokeIntRect(down, down.Bounds(), 0, 2)
	r.DrawIntRect(dkern, clockwiseRightBorder(dkern.Bounds(), 1)) // *
	r.DrawIntRect(dkern, bottomBorder(dkern.Bounds(), 1))
	r.DrawIntRect(dkernHorz, clockwiseRightBorder(dkernHorz.Bounds(), 1))
	r.DrawIntRect(dkernHorz, bottomBorder(dkernHorz.Bounds(), 1))
	// * Notice that technically dkern content could be overwritten by operations
	//   on 'down' after the clear, but since kernels can't be zero and 'down' already
	//   has 1 pixel margins, this won't happen in practice. Otherwise the clear should
	//   be delayed until after the horz kernel application

	// downscaling
	r.opts.Blend = ebiten.BlendCopy
	r.Scale(down, mask, 1, 1, 1.0/downscaling, false)

	// apply horz kern shader
	r.setDstRectCoords(0, 0, float32(dkernW64)+2, float32(downH64)+2)
	r.setSrcRectCoords(float32(-halfHorzMargin), float32(0), float32(downW64+halfHorzMargin)+2, float32(downH64)+2)
	r.opts.Blend = ebiten.BlendCopy
	r.opts.Images[0] = down
	r.opts.Uniforms["KernelLen"] = horzKernel.Size()
	r.opts.Uniforms["Kernel"] = gaussKerns[horzKernel]
	invokeShader(dkernHorz) // set VAs, more uniforms, invoke shader and clear(r.opts.Uniforms) if needed

	// apply vert blur kern
	r.opts.Uniforms["KernelLen"] = vertKernel.Size()
	r.opts.Uniforms["Kernel"] = gaussKerns[vertKernel]
	r.setDstRectCoords(0, 0, float32(dkernW64)+2, float32(dkernH64)+2)
	r.setSrcRectCoords(0, float32(-halfVertMargin), float32(dkernW64)+2, float32(downH64+halfVertMargin)+2)
	r.opts.Images[0] = dkernHorz
	ensureShaderVertBlurKernLoaded()
	dkern.DrawTrianglesShader(r.vertices[:], r.indices[:], shaderVertBlurKern, &r.opts)
	r.opts.Images[0] = nil
	clear(r.opts.Uniforms)

	// upscale
	if lighterBlend {
		r.opts.Blend = ebiten.BlendLighter
	} else {
		r.opts.Blend = preBlend
	}
	fx, fy := ox+float32(-downscaling-halfHorzMargin*downscaling), oy+float32(-downscaling-halfVertMargin*downscaling)
	r.Scale(target, dkern, fx, fy, downscaling, false)
	if lighterBlend {
		r.opts.Blend = preBlend
	}
}

func (r *Renderer) ApplyScanlinesSharp(target *ebiten.Image, darkThick, clearThick int, intensity, offset float32) {
	r.setFlatCustomVAs(float32(darkThick), float32(clearThick), intensity, offset)
	ensureShaderScanlinesSharpLoaded()
	r.DrawShader(target, 0, 0, shaderScanlinesSharp)
}

func (r *Renderer) ApplyWaveLines(target *ebiten.Image, lineThick, minFillRate, maxFillRate, linesPerOsc, offset float32, dirRadians float64) {
	if minFillRate > maxFillRate {
		panic("minFillRate > maxFillRate")
	}
	if minFillRate < 0 {
		panic("minFillRate < 0")
	}
	if maxFillRate == 0 {
		return
	}
	if maxFillRate > 1.0 {
		panic("maxFillRate > 1.0")
	}

	minFillThick := minFillRate * lineThick
	maxFillThick := maxFillRate * lineThick
	waveLen := linesPerOsc * lineThick
	r.opts.Uniforms["Offset"] = float32(math.Mod(float64(offset), float64(waveLen)))
	drs, drc := math.Sincos(dirRadians)
	hypot := math.Hypot(drs, drc)
	drs, drc = drs/hypot, drc/hypot
	r.opts.Uniforms["DirRadsSin"] = float32(drs)
	r.opts.Uniforms["DirRadsCos"] = float32(drc)
	r.setFlatCustomVAs(lineThick, minFillThick, maxFillThick, waveLen)
	ensureShaderWaveLinesLoaded()
	r.DrawShader(target, 0, 0, shaderWaveLines)
	clear(r.opts.Uniforms)
}
