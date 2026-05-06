package shapes

import (
	_ "embed"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
)

type shaderRef struct {
	shader *ebiten.Shader
	once   sync.Once
	src    []byte
}

func (s *shaderRef) Load() *ebiten.Shader {
	s.once.Do(s.loadOnce)
	return s.shader
}

func (s *shaderRef) loadOnce() {
	var err error
	s.shader, err = ebiten.NewShader(s.src)
	if err != nil {
		panic(err)
	}
}

//go:embed shaders/default.kage
var shaderDefaultSrc []byte
var shaderDefault = shaderRef{src: shaderDefaultSrc}

//go:embed shaders/bilinear.kage
var shaderBilinearSrc []byte
var shaderBilinear = shaderRef{src: shaderBilinearSrc}

//go:embed shaders/bicubic.kage
var shaderBicubicSrc []byte
var shaderBicubic = shaderRef{src: shaderBicubicSrc}

//go:embed shaders/draw_tint_nearest.kage
var shaderDrawTintNearestSrc []byte
var shaderDrawTintNearest = shaderRef{src: shaderDrawTintNearestSrc}

//go:embed shaders/draw_tint_bilinear.kage
var shaderDrawTintBilinearSrc []byte
var shaderDrawTintBilinear = shaderRef{src: shaderDrawTintBilinearSrc}

//go:embed shaders/rect.kage
var shaderRectSrc []byte
var shaderRect = shaderRef{src: shaderRectSrc}

//go:embed shaders/rect_soft.kage
var shaderRectSoftSrc []byte
var shaderRectSoft = shaderRef{src: shaderRectSoftSrc}

//go:embed shaders/rect_blur.kage
var shaderRectBlurSrc []byte
var shaderRectBlur = shaderRef{src: shaderRectBlurSrc}

//go:embed shaders/stroke_rect.kage
var shaderStrokeRectSrc []byte
var shaderStrokeRect = shaderRef{src: shaderStrokeRectSrc}

//go:embed shaders/line.kage
var shaderLineSrc []byte
var shaderLine = shaderRef{src: shaderLineSrc}

//go:embed shaders/circ_line.kage
var shaderCircLineSrc []byte
var shaderCircLine = shaderRef{src: shaderCircLineSrc}

//go:embed shaders/circle.kage
var shaderCircleSrc []byte
var shaderCircle = shaderRef{src: shaderCircleSrc}

//go:embed shaders/stroke_circle.kage
var shaderStrokeCircleSrc []byte
var shaderStrokeCircle = shaderRef{src: shaderStrokeCircleSrc}

//go:embed shaders/circ_sector.kage
var shaderCircSectorSrc []byte
var shaderCircSector = shaderRef{src: shaderCircSectorSrc}

//go:embed shaders/stroke_circ_sector.kage
var shaderStrokeCircSectorSrc []byte
var shaderStrokeCircSector = shaderRef{src: shaderStrokeCircSectorSrc}

//go:embed shaders/circ_sector_segment.kage
var shaderCircSectorSegmentSrc []byte
var shaderCircSectorSegment = shaderRef{src: shaderCircSectorSegmentSrc}

//go:embed shaders/ellipse.kage
var shaderEllipseSrc []byte
var shaderEllipse = shaderRef{src: shaderEllipseSrc}

//go:embed shaders/triangle.kage
var shaderTriangleSrc []byte
var shaderTriangle = shaderRef{src: shaderTriangleSrc}

//go:embed shaders/hexagon.kage
var shaderHexagonSrc []byte
var shaderHexagon = shaderRef{src: shaderHexagonSrc}

//go:embed shaders/quad.kage
var shaderQuadSrc []byte
var shaderQuad = shaderRef{src: shaderQuadSrc}

//go:embed shaders/alpha_mask_circ.kage
var shaderAlphaMaskCircSrc []byte
var shaderAlphaMaskCirc = shaderRef{src: shaderAlphaMaskCircSrc}

//go:embed shaders/mask.kage
var shaderMaskSrc []byte
var shaderMask = shaderRef{src: shaderMaskSrc}

//go:embed shaders/mask_at.kage
var shaderMaskAtSrc []byte
var shaderMaskAt = shaderRef{src: shaderMaskAtSrc}

//go:embed shaders/mask_horz.kage
var shaderMaskHorzSrc []byte
var shaderMaskHorz = shaderRef{src: shaderMaskHorzSrc}

//go:embed shaders/mask_circle.kage
var shaderMaskCircleSrc []byte
var shaderMaskCircle = shaderRef{src: shaderMaskCircleSrc}

//go:embed shaders/mask_threshold.kage
var shaderMaskThresholdSrc []byte
var shaderMaskThreshold = shaderRef{src: shaderMaskThresholdSrc}

//go:embed shaders/expansion.kage
var shaderExpansionSrc []byte
var shaderExpansion = shaderRef{src: shaderExpansionSrc}

//go:embed shaders/expansion_vert.kage
var shaderExpansionVertSrc []byte
var shaderExpansionVert = shaderRef{src: shaderExpansionVertSrc}

//go:embed shaders/expansion_horz.kage
var shaderExpansionHorzSrc []byte
var shaderExpansionHorz = shaderRef{src: shaderExpansionHorzSrc}

//go:embed shaders/erosion.kage
var shaderErosionSrc []byte
var shaderErosion = shaderRef{src: shaderErosionSrc}

//go:embed shaders/outline.kage
var shaderOutlineSrc []byte
var shaderOutline = shaderRef{src: shaderOutlineSrc}

//go:embed shaders/blur.kage
var shaderBlurSrc []byte
var shaderBlur = shaderRef{src: shaderBlurSrc}

//go:embed shaders/horz_blur.kage
var shaderHorzBlurSrc []byte
var shaderHorzBlur = shaderRef{src: shaderHorzBlurSrc}

//go:embed shaders/vert_blur.kage
var shaderVertBlurSrc []byte
var shaderVertBlur = shaderRef{src: shaderVertBlurSrc}

//go:embed shaders/horz_blur_kern.kage
var shaderHorzBlurKernSrc []byte
var shaderHorzBlurKern = shaderRef{src: shaderHorzBlurKernSrc}

//go:embed shaders/vert_blur_kern.kage
var shaderVertBlurKernSrc []byte
var shaderVertBlurKern = shaderRef{src: shaderVertBlurKernSrc}

//go:embed shaders/blur_vogel.kage
var shaderBlurVogelSrc []byte
var shaderBlurVogel = shaderRef{src: shaderBlurVogelSrc}

//go:embed shaders/glow_first_pass.kage
var shaderGlowFirstPassSrc []byte
var shaderGlowFirstPass = shaderRef{src: shaderGlowFirstPassSrc}

//go:embed shaders/glow_horz.kage
var shaderHorzGlowSrc []byte
var shaderHorzGlow = shaderRef{src: shaderHorzGlowSrc}

//go:embed shaders/glow_horz_dark.kage
var shaderDarkHorzGlowSrc []byte
var shaderDarkHorzGlow = shaderRef{src: shaderDarkHorzGlowSrc}

//go:embed shaders/horz_glow_kern.kage
var shaderHorzGlowKernSrc []byte
var shaderHorzGlowKern = shaderRef{src: shaderHorzGlowKernSrc}

//go:embed shaders/horz_color_glow.kage
var shaderHorzColorGlowSrc []byte
var shaderHorzColorGlow = shaderRef{src: shaderHorzColorGlowSrc}

//go:embed shaders/scanlines_sharp.kage
var shaderScanlinesSharpSrc []byte
var shaderScanlinesSharp = shaderRef{src: shaderScanlinesSharpSrc}

//go:embed shaders/wave_lines.kage
var shaderWaveLinesSrc []byte
var shaderWaveLines = shaderRef{src: shaderWaveLinesSrc}

//go:embed shaders/gradient.kage
var shaderGradientSrc []byte
var shaderGradient = shaderRef{src: shaderGradientSrc}

//go:embed shaders/gradient_radial.kage
var shaderGradientRadialSrc []byte
var shaderGradientRadial = shaderRef{src: shaderGradientRadialSrc}

//go:embed shaders/colorize_lightness.kage
var shaderColorizeByLightnessSrc []byte
var shaderColorizeByLightness = shaderRef{src: shaderColorizeByLightnessSrc}

//go:embed shaders/oklab_shift.kage
var shaderOklabShiftSrc []byte
var shaderOklabShift = shaderRef{src: shaderOklabShiftSrc}

//go:embed shaders/color_mix.kage
var shaderColorMixSrc []byte
var shaderColorMix = shaderRef{src: shaderColorMixSrc}

//go:embed shaders/color_mix_bilinear.kage
var shaderColorMixBilinearSrc []byte
var shaderColorMixBilinear = shaderRef{src: shaderColorMixBilinearSrc}

//go:embed shaders/dither_matrix4.kage
var shaderDitherMat4Src []byte
var shaderDitherMat4 = shaderRef{src: shaderDitherMat4Src}

//go:embed shaders/map_projective.kage
var shaderMapProjectiveSrc []byte
var shaderMapProjective = shaderRef{src: shaderMapProjectiveSrc}

//go:embed shaders/map_projective_ani.kage
var shaderMapProjectiveAniSrc []byte
var shaderMapProjectiveAni = shaderRef{src: shaderMapProjectiveAniSrc}

//go:embed shaders/map_quad4.kage
var shaderMapQuad4Src []byte
var shaderMapQuad4 = shaderRef{src: shaderMapQuad4Src}

//go:embed shaders/warp_barrel.kage
var shaderWarpBarrelSrc []byte
var shaderWarpBarrel = shaderRef{src: shaderWarpBarrelSrc}

//go:embed shaders/warp_pincushion_quad.kage
var shaderWarpPincushionQuadSrc []byte
var shaderWarpPincushionQuad = shaderRef{src: shaderWarpPincushionQuadSrc}

//go:embed shaders/warp_arc.kage
var shaderWarpArcSrc []byte
var shaderWarpArc = shaderRef{src: shaderWarpArcSrc}

//go:embed shaders/noise.kage
var shaderNoiseSrc []byte
var shaderNoise = shaderRef{src: shaderNoiseSrc}

//go:embed shaders/noise_golden.kage
var shaderNoiseGoldenSrc []byte
var shaderNoiseGolden = shaderRef{src: shaderNoiseGoldenSrc}

//go:embed shaders/tile_rects_grid.kage
var shaderTileRectsGridSrc []byte
var shaderTileRectsGrid = shaderRef{src: shaderTileRectsGridSrc}

//go:embed shaders/tile_dots_grid.kage
var shaderTileDotsGridSrc []byte
var shaderTileDotsGrid = shaderRef{src: shaderTileDotsGridSrc}

//go:embed shaders/tile_dots_hex.kage
var shaderTileDotsHexSrc []byte
var shaderTileDotsHex = shaderRef{src: shaderTileDotsHexSrc}

//go:embed shaders/tile_tri_up_grid.kage
var shaderTileTriUpGridSrc []byte
var shaderTileTriUpGrid = shaderRef{src: shaderTileTriUpGridSrc}

//go:embed shaders/tile_tri_hex.kage
var shaderTileTriHexSrc []byte
var shaderTileTriHex = shaderRef{src: shaderTileTriHexSrc}

//go:embed shaders/halftone_tri.kage
var shaderHalftoneTriSrc []byte
var shaderHalftoneTri = shaderRef{src: shaderHalftoneTriSrc}

//go:embed shaders/jfm_pass.kage
var shaderJFMPassSrc []byte
var shaderJFMPass = shaderRef{src: shaderJFMPassSrc}

//go:embed shaders/jfm_init_fill.kage
var shaderJFMInitFillSrc []byte
var shaderJFMInitFill = shaderRef{src: shaderJFMInitFillSrc}

//go:embed shaders/jfm_init_boundary.kage
var shaderJFMInitBoundarySrc []byte
var shaderJFMInitBoundary = shaderRef{src: shaderJFMInitBoundarySrc}

//go:embed shaders/jfm_heat.kage
var shaderJFMHeatSrc []byte
var shaderJFMHeat = shaderRef{src: shaderJFMHeatSrc}

//go:embed shaders/jfm_expansion.kage
var shaderJFMExpansionSrc []byte
var shaderJFMExpansion = shaderRef{src: shaderJFMExpansionSrc}

//go:embed shaders/jfm_erosion.kage
var shaderJFMErosionSrc []byte
var shaderJFMErosion = shaderRef{src: shaderJFMErosionSrc}

//go:embed shaders/study_wave_funcs.kage
var shaderStudyWaveFuncsSrc []byte
var shaderStudyWaveFuncs = shaderRef{src: shaderStudyWaveFuncsSrc}
