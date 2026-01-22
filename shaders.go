package shapes

import (
	_ "embed"

	"github.com/hajimehoshi/ebiten/v2"
)

//go:embed shaders/default.kage
var shaderDefaultSrc []byte

//go:embed shaders/bilinear.kage
var shaderBilinearSrc []byte

//go:embed shaders/rect.kage
var shaderRectSrc []byte

//go:embed shaders/stroke_rect.kage
var shaderStrokeRectSrc []byte

//go:embed shaders/line.kage
var shaderLineSrc []byte

//go:embed shaders/circle.kage
var shaderCircleSrc []byte

//go:embed shaders/stroke_circle.kage
var shaderStrokeCircleSrc []byte

//go:embed shaders/ring.kage
var shaderRingSrc []byte

//go:embed shaders/ring_sector.kage
var shaderRingSectorSrc []byte

//go:embed shaders/stroke_ring_sector.kage
var shaderStrokeRingSectorSrc []byte

//go:embed shaders/pie.kage
var shaderPieSrc []byte

//go:embed shaders/stroke_pie.kage
var shaderStrokePieSrc []byte

//go:embed shaders/ellipse.kage
var shaderEllipseSrc []byte

//go:embed shaders/triangle.kage
var shaderTriangleSrc []byte

//go:embed shaders/hexagon.kage
var shaderHexagonSrc []byte

//go:embed shaders/quad.kage
var shaderQuadSrc []byte

//go:embed shaders/alpha_mask_circ.kage
var shaderAlphaMaskCircSrc []byte

//go:embed shaders/mask.kage
var shaderMaskSrc []byte

//go:embed shaders/mask_at.kage
var shaderMaskAtSrc []byte

//go:embed shaders/mask_horz.kage
var shaderMaskHorzSrc []byte

//go:embed shaders/mask_circle.kage
var shaderMaskCircleSrc []byte

//go:embed shaders/mask_threshold.kage
var shaderMaskThresholdSrc []byte

//go:embed shaders/expansion.kage
var shaderExpansionSrc []byte

//go:embed shaders/expansion_vert.kage
var shaderExpansionVertSrc []byte

//go:embed shaders/expansion_horz.kage
var shaderExpansionHorzSrc []byte

//go:embed shaders/erosion.kage
var shaderErosionSrc []byte

//go:embed shaders/outline.kage
var shaderOutlineSrc []byte

//go:embed shaders/blur.kage
var shaderBlurSrc []byte

//go:embed shaders/horz_blur.kage
var shaderHorzBlurSrc []byte

//go:embed shaders/vert_blur.kage
var shaderVertBlurSrc []byte

//go:embed shaders/horz_blur_kern.kage
var shaderHorzBlurKernSrc []byte

//go:embed shaders/vert_blur_kern.kage
var shaderVertBlurKernSrc []byte

//go:embed shaders/glow_first_pass.kage
var shaderGlowFirstPassSrc []byte

//go:embed shaders/glow_horz.kage
var shaderHorzGlowSrc []byte

//go:embed shaders/glow_horz_dark.kage
var shaderDarkHorzGlowSrc []byte

//go:embed shaders/horz_glow_kern.kage
var shaderHorzGlowKernSrc []byte

//go:embed shaders/horz_color_glow.kage
var shaderHorzColorGlowSrc []byte

//go:embed shaders/shadow.kage
var shaderShadowSrc []byte

//go:embed shaders/hard_shadow.kage
var shaderHardShadowSrc []byte

//go:embed shaders/zoom_shadow.kage
var shaderZoomShadowSrc []byte

//go:embed shaders/scanlines_sharp.kage
var shaderScanlinesSharpSrc []byte

//go:embed shaders/wave_lines.kage
var shaderWaveLinesSrc []byte

//go:embed shaders/flat_paint.kage
var shaderFlatPaintSrc []byte

//go:embed shaders/gradient.kage
var shaderGradientSrc []byte

//go:embed shaders/gradient_dither.kage
var shaderGradientDitherSrc []byte

//go:embed shaders/gradient_radial.kage
var shaderGradientRadialSrc []byte

//go:embed shaders/colorize_lightness.kage
var shaderColorizeByLightnessSrc []byte

//go:embed shaders/oklab_shift.kage
var shaderOklabShiftSrc []byte

//go:embed shaders/color_mix.kage
var shaderColorMixSrc []byte

//go:embed shaders/dither_matrix4.kage
var shaderDitherMat4Src []byte

//go:embed shaders/map_projective.kage
var shaderMapProjectiveSrc []byte

//go:embed shaders/map_quad4.kage
var shaderMapQuad4Src []byte

//go:embed shaders/warp_barrel.kage
var shaderWarpBarrelSrc []byte

//go:embed shaders/warp_pincushion_quad.kage
var shaderWarpPincushionQuadSrc []byte

//go:embed shaders/warp_arc.kage
var shaderWarpArcSrc []byte

//go:embed shaders/noise.kage
var shaderNoiseSrc []byte

//go:embed shaders/noise_golden.kage
var shaderNoiseGoldenSrc []byte

//go:embed shaders/tile_rects_grid.kage
var shaderTileRectsGridSrc []byte

//go:embed shaders/tile_dots_grid.kage
var shaderTileDotsGridSrc []byte

//go:embed shaders/tile_dots_hex.kage
var shaderTileDotsHexSrc []byte

//go:embed shaders/tile_tri_up_grid.kage
var shaderTileTriUpGridSrc []byte

//go:embed shaders/tile_tri_hex.kage
var shaderTileTriHexSrc []byte

//go:embed shaders/halftone_tri.kage
var shaderHalftoneTriSrc []byte

//go:embed shaders/jfm_pass.kage
var shaderJFMPassSrc []byte

//go:embed shaders/jfm_init_fill.kage
var shaderJFMInitFillSrc []byte

//go:embed shaders/jfm_init_boundary.kage
var shaderJFMInitBoundarySrc []byte

//go:embed shaders/jfm_heat.kage
var shaderJFMHeatSrc []byte

//go:embed shaders/jfm_expansion.kage
var shaderJFMExpansionSrc []byte

//go:embed shaders/jfm_erosion.kage
var shaderJFMErosionSrc []byte

//go:embed shaders/study_wave_funcs.kage
var shaderStudyWaveFuncsSrc []byte

var shaderDefault *ebiten.Shader
var shaderBilinear *ebiten.Shader
var shaderRect *ebiten.Shader
var shaderStrokeRect *ebiten.Shader
var shaderLine *ebiten.Shader
var shaderCircle *ebiten.Shader
var shaderStrokeCircle *ebiten.Shader
var shaderRing *ebiten.Shader
var shaderRingSector *ebiten.Shader
var shaderStrokeRingSector *ebiten.Shader
var shaderPie *ebiten.Shader
var shaderStrokePie *ebiten.Shader
var shaderEllipse *ebiten.Shader
var shaderTriangle *ebiten.Shader
var shaderHexagon *ebiten.Shader
var shaderQuad *ebiten.Shader
var shaderAlphaMaskCirc *ebiten.Shader
var shaderMask *ebiten.Shader
var shaderMaskAt *ebiten.Shader
var shaderMaskHorz *ebiten.Shader
var shaderMaskCircle *ebiten.Shader
var shaderMaskThreshold *ebiten.Shader
var shaderExpansion *ebiten.Shader
var shaderExpansionVert *ebiten.Shader
var shaderExpansionHorz *ebiten.Shader
var shaderErosion *ebiten.Shader
var shaderOutline *ebiten.Shader
var shaderBlur *ebiten.Shader
var shaderHorzBlur *ebiten.Shader
var shaderVertBlur *ebiten.Shader
var shaderHorzBlurKern *ebiten.Shader
var shaderVertBlurKern *ebiten.Shader
var shaderGlowFirstPass *ebiten.Shader
var shaderHorzGlow *ebiten.Shader
var shaderDarkHorzGlow *ebiten.Shader
var shaderHorzGlowKern *ebiten.Shader
var shaderHorzColorGlow *ebiten.Shader
var shaderShadow *ebiten.Shader
var shaderHardShadow *ebiten.Shader
var shaderZoomShadow *ebiten.Shader
var shaderScanlinesSharp *ebiten.Shader
var shaderWaveLines *ebiten.Shader
var shaderFlatPaint *ebiten.Shader
var shaderGradient *ebiten.Shader
var shaderGradientDither *ebiten.Shader
var shaderGradientRadial *ebiten.Shader
var shaderColorizeByLightness *ebiten.Shader
var shaderOklabShift *ebiten.Shader
var shaderColorMix *ebiten.Shader
var shaderDitherMat4 *ebiten.Shader
var shaderMapQuad4 *ebiten.Shader
var shaderMapProjective *ebiten.Shader
var shaderWarpBarrel *ebiten.Shader
var shaderWarpPincushionQuad *ebiten.Shader
var shaderWarpArc *ebiten.Shader
var shaderNoise *ebiten.Shader
var shaderNoiseGolden *ebiten.Shader
var shaderTileRectsGrid *ebiten.Shader
var shaderTileDotsGrid *ebiten.Shader
var shaderTileDotsHex *ebiten.Shader
var shaderTileTriUpGrid *ebiten.Shader
var shaderTileTriHex *ebiten.Shader
var shaderHalftoneTri *ebiten.Shader
var shaderJFMPass *ebiten.Shader
var shaderJFMInitFill *ebiten.Shader
var shaderJFMInitBoundary *ebiten.Shader
var shaderJFMHeat *ebiten.Shader
var shaderJFMExpansion *ebiten.Shader
var shaderJFMErosion *ebiten.Shader

var shaderStudyWaveFuncs *ebiten.Shader

func mustCompile(src []byte) *ebiten.Shader {
	shader, err := ebiten.NewShader(src)
	if err != nil {
		panic(err)
	}
	return shader
}

func ensureShaderDefaultLoaded() {
	if shaderDefault == nil {
		shaderDefault = mustCompile(shaderDefaultSrc)
	}
}

func ensureShaderBilinearLoaded() {
	if shaderBilinear == nil {
		shaderBilinear = mustCompile(shaderBilinearSrc)
	}
}

func ensureShaderRectLoaded() {
	if shaderRect == nil {
		shaderRect = mustCompile(shaderRectSrc)
	}
}

func ensureShaderStrokeRectLoaded() {
	if shaderStrokeRect == nil {
		shaderStrokeRect = mustCompile(shaderStrokeRectSrc)
	}
}

func ensureShaderLineLoaded() {
	if shaderLine == nil {
		shaderLine = mustCompile(shaderLineSrc)
	}
}

func ensureShaderCircleLoaded() {
	if shaderCircle == nil {
		shaderCircle = mustCompile(shaderCircleSrc)
	}
}

func ensureShaderStrokeCircleLoaded() {
	if shaderStrokeCircle == nil {
		shaderStrokeCircle = mustCompile(shaderStrokeCircleSrc)
	}
}

func ensureShaderRingLoaded() {
	if shaderRing == nil {
		shaderRing = mustCompile(shaderRingSrc)
	}
}

func ensureShaderRingSectorLoaded() {
	if shaderRingSector == nil {
		shaderRingSector = mustCompile(shaderRingSectorSrc)
	}
}

func ensureShaderStrokeRingSectorLoaded() {
	if shaderStrokeRingSector == nil {
		shaderStrokeRingSector = mustCompile(shaderStrokeRingSectorSrc)
	}
}

func ensureShaderPieLoaded() {
	if shaderPie == nil {
		shaderPie = mustCompile(shaderPieSrc)
	}
}

func ensureShaderStrokePieLoaded() {
	if shaderStrokePie == nil {
		shaderStrokePie = mustCompile(shaderStrokePieSrc)
	}
}

func ensureShaderEllipseLoaded() {
	if shaderEllipse == nil {
		shaderEllipse = mustCompile(shaderEllipseSrc)
	}
}

func ensureShaderTriangleLoaded() {
	if shaderTriangle == nil {
		shaderTriangle = mustCompile(shaderTriangleSrc)
	}
}

func ensureShaderHexagonLoaded() {
	if shaderHexagon == nil {
		shaderHexagon = mustCompile(shaderHexagonSrc)
	}
}

func ensureShaderQuadLoaded() {
	if shaderQuad == nil {
		shaderQuad = mustCompile(shaderQuadSrc)
	}
}

func ensureShaderAlphaMaskCircLoaded() {
	if shaderAlphaMaskCirc == nil {
		shaderAlphaMaskCirc = mustCompile(shaderAlphaMaskCircSrc)
	}
}

func ensureShaderMaskLoaded() {
	if shaderMask == nil {
		shaderMask = mustCompile(shaderMaskSrc)
	}
}

func ensureShaderMaskAtLoaded() {
	if shaderMaskAt == nil {
		shaderMaskAt = mustCompile(shaderMaskAtSrc)
	}
}

func ensureShaderMaskHorzLoaded() {
	if shaderMaskHorz == nil {
		shaderMaskHorz = mustCompile(shaderMaskHorzSrc)
	}
}

func ensureShaderMaskCircleLoaded() {
	if shaderMaskCircle == nil {
		shaderMaskCircle = mustCompile(shaderMaskCircleSrc)
	}
}

func ensureShaderMaskThresholdLoaded() {
	if shaderMaskThreshold == nil {
		shaderMaskThreshold = mustCompile(shaderMaskThresholdSrc)
	}
}

func ensureShaderExpansionLoaded() {
	if shaderExpansion == nil {
		shaderExpansion = mustCompile(shaderExpansionSrc)
	}
}

func ensureShaderExpansionVertLoaded() {
	if shaderExpansionVert == nil {
		shaderExpansionVert = mustCompile(shaderExpansionVertSrc)
	}
}

func ensureShaderExpansionHorzLoaded() {
	if shaderExpansionHorz == nil {
		shaderExpansionHorz = mustCompile(shaderExpansionHorzSrc)
	}
}

func ensureShaderErosionLoaded() {
	if shaderErosion == nil {
		shaderErosion = mustCompile(shaderErosionSrc)
	}
}

func ensureShaderOutlineLoaded() {
	if shaderOutline == nil {
		shaderOutline = mustCompile(shaderOutlineSrc)
	}
}

func ensureShaderBlurLoaded() {
	if shaderBlur == nil {
		shaderBlur = mustCompile(shaderBlurSrc)
	}
}

func ensureShaderHorzBlurLoaded() {
	if shaderHorzBlur == nil {
		shaderHorzBlur = mustCompile(shaderHorzBlurSrc)
	}
}

func ensureShaderVertBlurLoaded() {
	if shaderVertBlur == nil {
		shaderVertBlur = mustCompile(shaderVertBlurSrc)
	}
}

func ensureShaderHorzBlurKernLoaded() {
	if shaderHorzBlurKern == nil {
		shaderHorzBlurKern = mustCompile(shaderHorzBlurKernSrc)
	}
}

func ensureShaderVertBlurKernLoaded() {
	if shaderVertBlurKern == nil {
		shaderVertBlurKern = mustCompile(shaderVertBlurKernSrc)
	}
}

func ensureShaderGlowFirstPassLoaded() {
	if shaderGlowFirstPass == nil {
		shaderGlowFirstPass = mustCompile(shaderGlowFirstPassSrc)
	}
}

func ensureShaderHorzGlowLoaded() {
	if shaderHorzGlow == nil {
		shaderHorzGlow = mustCompile(shaderHorzGlowSrc)
	}
}

func ensureShaderDarkHorzGlowLoaded() {
	if shaderDarkHorzGlow == nil {
		shaderDarkHorzGlow = mustCompile(shaderDarkHorzGlowSrc)
	}
}

func ensureShaderHorzGlowKernLoaded() {
	if shaderHorzGlowKern == nil {
		shaderHorzGlowKern = mustCompile(shaderHorzGlowKernSrc)
	}
}

func ensureShaderHorzColorGlowLoaded() {
	if shaderHorzColorGlow == nil {
		shaderHorzColorGlow = mustCompile(shaderHorzColorGlowSrc)
	}
}

func ensureShaderShadowLoaded() {
	if shaderShadow == nil {
		shaderShadow = mustCompile(shaderShadowSrc)
	}
}

func ensureShaderHardShadowLoaded() {
	if shaderHardShadow == nil {
		shaderHardShadow = mustCompile(shaderHardShadowSrc)
	}
}

func ensureShaderZoomShadowLoaded() {
	if shaderZoomShadow == nil {
		shaderZoomShadow = mustCompile(shaderZoomShadowSrc)
	}
}

func ensureShaderScanlinesSharpLoaded() {
	if shaderScanlinesSharp == nil {
		shaderScanlinesSharp = mustCompile(shaderScanlinesSharpSrc)
	}
}

func ensureShaderWaveLinesLoaded() {
	if shaderWaveLines == nil {
		shaderWaveLines = mustCompile(shaderWaveLinesSrc)
	}
}

func ensureShaderFlatPaintLoaded() {
	if shaderFlatPaint == nil {
		shaderFlatPaint = mustCompile(shaderFlatPaintSrc)
	}
}

func ensureShaderGradientLoaded() {
	if shaderGradient == nil {
		shaderGradient = mustCompile(shaderGradientSrc)
	}
}

func ensureShaderGradientDitherLoaded() {
	if shaderGradientDither == nil {
		shaderGradientDither = mustCompile(shaderGradientDitherSrc)
	}
}

func ensureShaderGradientRadialLoaded() {
	if shaderGradientRadial == nil {
		shaderGradientRadial = mustCompile(shaderGradientRadialSrc)
	}
}

func ensureShaderColorizeByLightnessLoaded() {
	if shaderColorizeByLightness == nil {
		shaderColorizeByLightness = mustCompile(shaderColorizeByLightnessSrc)
	}
}

func ensureShaderOklabShiftLoaded() {
	if shaderOklabShift == nil {
		shaderOklabShift = mustCompile(shaderOklabShiftSrc)
	}
}

func ensureShaderColorMixLoaded() {
	if shaderColorMix == nil {
		shaderColorMix = mustCompile(shaderColorMixSrc)
	}
}

func ensureShaderDitherMat4Loaded() {
	if shaderDitherMat4 == nil {
		shaderDitherMat4 = mustCompile(shaderDitherMat4Src)
	}
}

func ensureShaderMapQuad4Loaded() {
	if shaderMapQuad4 == nil {
		shaderMapQuad4 = mustCompile(shaderMapQuad4Src)
	}
}

func ensureShaderMapProjectiveLoaded() {
	if shaderMapProjective == nil {
		shaderMapProjective = mustCompile(shaderMapProjectiveSrc)
	}
}

func ensureShaderWarpBarrelLoaded() {
	if shaderWarpBarrel == nil {
		shaderWarpBarrel = mustCompile(shaderWarpBarrelSrc)
	}
}

func ensureShaderWarpPincushionQuadLoaded() {
	if shaderWarpPincushionQuad == nil {
		shaderWarpPincushionQuad = mustCompile(shaderWarpPincushionQuadSrc)
	}
}

func ensureShaderWarpArcLoaded() {
	if shaderWarpArc == nil {
		shaderWarpArc = mustCompile(shaderWarpArcSrc)
	}
}

func ensureShaderNoiseLoaded() {
	if shaderNoise == nil {
		shaderNoise = mustCompile(shaderNoiseSrc)
	}
}

func ensureShaderNoiseGoldenLoaded() {
	if shaderNoiseGolden == nil {
		shaderNoiseGolden = mustCompile(shaderNoiseGoldenSrc)
	}
}

func ensureShaderTileRectsGridLoaded() {
	if shaderTileRectsGrid == nil {
		shaderTileRectsGrid = mustCompile(shaderTileRectsGridSrc)
	}
}

func ensureShaderTileDotsGridLoaded() {
	if shaderTileDotsGrid == nil {
		shaderTileDotsGrid = mustCompile(shaderTileDotsGridSrc)
	}
}

func ensureShaderTileDotsHexLoaded() {
	if shaderTileDotsHex == nil {
		shaderTileDotsHex = mustCompile(shaderTileDotsHexSrc)
	}
}

func ensureShaderTileTriUpGridLoaded() {
	if shaderTileTriUpGrid == nil {
		shaderTileTriUpGrid = mustCompile(shaderTileTriUpGridSrc)
	}
}

func ensureShaderTileTriHexLoaded() {
	if shaderTileTriHex == nil {
		shaderTileTriHex = mustCompile(shaderTileTriHexSrc)
	}
}

func ensureShaderHalftoneTriLoaded() {
	if shaderHalftoneTri == nil {
		shaderHalftoneTri = mustCompile(shaderHalftoneTriSrc)
	}
}

func ensureShaderJFMPassLoaded() {
	if shaderJFMPass == nil {
		shaderJFMPass = mustCompile(shaderJFMPassSrc)
	}
}

func ensureShaderJFMInitFillLoaded() {
	if shaderJFMInitFill == nil {
		shaderJFMInitFill = mustCompile(shaderJFMInitFillSrc)
	}
}

func ensureShaderJFMInitBoundaryLoaded() {
	if shaderJFMInitBoundary == nil {
		shaderJFMInitBoundary = mustCompile(shaderJFMInitBoundarySrc)
	}
}

func ensureShaderJFMHeatLoaded() {
	if shaderJFMHeat == nil {
		shaderJFMHeat = mustCompile(shaderJFMHeatSrc)
	}
}

func ensureShaderJFMExpansionLoaded() {
	if shaderJFMExpansion == nil {
		shaderJFMExpansion = mustCompile(shaderJFMExpansionSrc)
	}
}

func ensureShaderJFMErosionLoaded() {
	if shaderJFMErosion == nil {
		shaderJFMErosion = mustCompile(shaderJFMErosionSrc)
	}
}

func ensureShaderStudyWaveFuncsLoaded() {
	if shaderStudyWaveFuncs == nil {
		shaderStudyWaveFuncs = mustCompile(shaderStudyWaveFuncsSrc)
	}
}
