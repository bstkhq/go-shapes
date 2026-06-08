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

//go:embed shaders/kern_vert_finish.kage
var shaderKernVertFinishSrc []byte
var shaderKernVertFinish = shaderRef{src: shaderKernVertFinishSrc}

//go:embed shaders/draw_tint_nearest.kage
var shaderDrawTintNearestSrc []byte
var shaderDrawTintNearest = shaderRef{src: shaderDrawTintNearestSrc}

//go:embed shaders/draw_tint_bilinear.kage
var shaderDrawTintBilinearSrc []byte
var shaderDrawTintBilinear = shaderRef{src: shaderDrawTintBilinearSrc}

//go:embed shaders/shapes/poly/rect.kage
var shaderRectSrc []byte
var shaderRect = shaderRef{src: shaderRectSrc}

//go:embed shaders/shapes/poly/rect_soft_in.kage
var shaderRectSoftInSrc []byte
var shaderRectSoftIn = shaderRef{src: shaderRectSoftInSrc}

//go:embed shaders/shapes/poly/rect_soft_blur.kage
var shaderRectSoftBlurSrc []byte
var shaderRectSoftBlur = shaderRef{src: shaderRectSoftBlurSrc}

//go:embed shaders/shapes/poly/stroke_rect.kage
var shaderStrokeRectSrc []byte
var shaderStrokeRect = shaderRef{src: shaderStrokeRectSrc}

//go:embed shaders/shapes/poly/line.kage
var shaderLineSrc []byte
var shaderLine = shaderRef{src: shaderLineSrc}

//go:embed shaders/shapes/poly/triangle.kage
var shaderTriangleSrc []byte
var shaderTriangle = shaderRef{src: shaderTriangleSrc}

//go:embed shaders/shapes/poly/hexagon.kage
var shaderHexagonSrc []byte
var shaderHexagon = shaderRef{src: shaderHexagonSrc}

//go:embed shaders/shapes/poly/quad.kage
var shaderQuadSrc []byte
var shaderQuad = shaderRef{src: shaderQuadSrc}

//go:embed shaders/shapes/poly/quad_self_intersect.kage
var shaderQuadSelfIntersectSrc []byte
var shaderQuadSelfIntersect = shaderRef{src: shaderQuadSelfIntersectSrc}

//go:embed shaders/shapes/poly/quad_soft_in.kage
var shaderQuadSoftInSrc []byte
var shaderQuadSoftIn = shaderRef{src: shaderQuadSoftInSrc}

//go:embed shaders/shapes/poly/quad_soft_blur.kage
var shaderQuadSoftBlurSrc []byte
var shaderQuadSoftBlur = shaderRef{src: shaderQuadSoftBlurSrc}

//go:embed shaders/shapes/circ/arc.kage
var shaderArcSrc []byte
var shaderArc = shaderRef{src: shaderArcSrc}

//go:embed shaders/shapes/circ/circle.kage
var shaderCircleSrc []byte
var shaderCircle = shaderRef{src: shaderCircleSrc}

//go:embed shaders/shapes/circ/stroke_circle.kage
var shaderStrokeCircleSrc []byte
var shaderStrokeCircle = shaderRef{src: shaderStrokeCircleSrc}

//go:embed shaders/shapes/circ/radial_sector.kage
var shaderRadialSectorSrc []byte
var shaderRadialSector = shaderRef{src: shaderRadialSectorSrc}

//go:embed shaders/shapes/circ/radial_sector_segment.kage
var shaderRadialSectorSegmentSrc []byte
var shaderRadialSectorSegment = shaderRef{src: shaderRadialSectorSegmentSrc}

//go:embed shaders/shapes/circ/circular_sector_inner.kage
var shaderCircularSectorInnerSrc []byte
var shaderCircularSectorInner = shaderRef{src: shaderCircularSectorInnerSrc}

//go:embed shaders/shapes/circ/ellipse.kage
var shaderEllipseSrc []byte
var shaderEllipse = shaderRef{src: shaderEllipseSrc}

//go:embed shaders/mask/alpha_mask_radial.kage
var shaderAlphaMaskRadialSrc []byte
var shaderAlphaMaskRadial = shaderRef{src: shaderAlphaMaskRadialSrc}

//go:embed shaders/mask/mask.kage
var shaderMaskSrc []byte
var shaderMask = shaderRef{src: shaderMaskSrc}

//go:embed shaders/mask/mask_at.kage
var shaderMaskAtSrc []byte
var shaderMaskAt = shaderRef{src: shaderMaskAtSrc}

//go:embed shaders/mask/mask_threshold.kage
var shaderMaskThresholdSrc []byte
var shaderMaskThreshold = shaderRef{src: shaderMaskThresholdSrc}

//go:embed shaders/mask/mask_horz.kage
var shaderMaskHorzSrc []byte
var shaderMaskHorz = shaderRef{src: shaderMaskHorzSrc}

//go:embed shaders/mask/mask_circ.kage
var shaderMaskCircSrc []byte
var shaderMaskCirc = shaderRef{src: shaderMaskCircSrc}

//go:embed shaders/morph/expansion.kage
var shaderMorphExpansionSrc []byte
var shaderMorphExpansion = shaderRef{src: shaderMorphExpansionSrc}

//go:embed shaders/morph/expansion_rect_vert.kage
var shaderMorphExpansionRectVertSrc []byte
var shaderMorphExpansionRectVert = shaderRef{src: shaderMorphExpansionRectVertSrc}

//go:embed shaders/morph/expansion_rect_horz.kage
var shaderMorphExpansionRectHorzSrc []byte
var shaderMorphExpansionRectHorz = shaderRef{src: shaderMorphExpansionRectHorzSrc}

//go:embed shaders/morph/erosion.kage
var shaderMorphErosionSrc []byte
var shaderMorphErosion = shaderRef{src: shaderMorphErosionSrc}

//go:embed shaders/morph/outline.kage
var shaderMorphOutlineSrc []byte
var shaderMorphOutline = shaderRef{src: shaderMorphOutlineSrc}

//go:embed shaders/blur/naive.kage
var shaderBlurNaiveSrc []byte
var shaderBlurNaive = shaderRef{src: shaderBlurNaiveSrc}

//go:embed shaders/blur/horz.kage
var shaderBlurHorzSrc []byte
var shaderBlurHorz = shaderRef{src: shaderBlurHorzSrc}

//go:embed shaders/blur/vert.kage
var shaderBlurVertSrc []byte
var shaderBlurVert = shaderRef{src: shaderBlurVertSrc}

//go:embed shaders/blur/horz_kern.kage
var shaderBlurHorzKernSrc []byte
var shaderBlurHorzKern = shaderRef{src: shaderBlurHorzKernSrc}

//go:embed shaders/blur/vogel.kage
var shaderBlurVogelSrc []byte
var shaderBlurVogel = shaderRef{src: shaderBlurVogelSrc}

//go:embed shaders/glow/horz.kage
var shaderGlowHorzSrc []byte
var shaderGlowHorz = shaderRef{src: shaderGlowHorzSrc}

//go:embed shaders/glow/vert.kage
var shaderGlowVertSrc []byte
var shaderGlowVert = shaderRef{src: shaderGlowVertSrc}

//go:embed shaders/glow/horz_kern.kage
var shaderGlowHorzKernSrc []byte
var shaderGlowHorzKern = shaderRef{src: shaderGlowHorzKernSrc}

//go:embed shaders/glow/dark_vert.kage
var shaderGlowDarkVertSrc []byte
var shaderGlowDarkVert = shaderRef{src: shaderGlowDarkVertSrc}

//go:embed shaders/glow/dark_horz.kage
var shaderGlowDarkHorzSrc []byte
var shaderGlowDarkHorz = shaderRef{src: shaderGlowDarkHorzSrc}

//go:embed shaders/glow/dark_horz_kern.kage
var shaderGlowDarkHorzKernSrc []byte
var shaderGlowDarkHorzKern = shaderRef{src: shaderGlowDarkHorzKernSrc}

//go:embed shaders/glow/color_horz.kage
var shaderGlowColorHorzSrc []byte
var shaderGlowColorHorz = shaderRef{src: shaderGlowColorHorzSrc}

//go:embed shaders/glow/color_vert.kage
var shaderGlowColorVertSrc []byte
var shaderGlowColorVert = shaderRef{src: shaderGlowColorVertSrc}

//go:embed shaders/glow/color_horz_kern.kage
var shaderGlowColorHorzKernSrc []byte
var shaderGlowColorHorzKern = shaderRef{src: shaderGlowColorHorzKernSrc}

//go:embed shaders/color/gradient.kage
var shaderGradientSrc []byte
var shaderGradient = shaderRef{src: shaderGradientSrc}

//go:embed shaders/color/gradient_radial.kage
var shaderGradientRadialSrc []byte
var shaderGradientRadial = shaderRef{src: shaderGradientRadialSrc}

//go:embed shaders/color/colorize_lightness.kage
var shaderColorizeByLightnessSrc []byte
var shaderColorizeByLightness = shaderRef{src: shaderColorizeByLightnessSrc}

//go:embed shaders/color/oklab_shift.kage
var shaderOklabShiftSrc []byte
var shaderOklabShift = shaderRef{src: shaderOklabShiftSrc}

//go:embed shaders/color/mix.kage
var shaderColorMixSrc []byte
var shaderColorMix = shaderRef{src: shaderColorMixSrc}

//go:embed shaders/color/mix_bilinear.kage
var shaderColorMixBilinearSrc []byte
var shaderColorMixBilinear = shaderRef{src: shaderColorMixBilinearSrc}

//go:embed shaders/project/quad_bilinear.kage
var shaderMapQuadBilinearSrc []byte
var shaderMapQuadBilinear = shaderRef{src: shaderMapQuadBilinearSrc}

//go:embed shaders/project/quad_anisotropic.kage
var shaderMapQuadAnisotropicSrc []byte
var shaderMapQuadAnisotropic = shaderRef{src: shaderMapQuadAnisotropicSrc}

//go:embed shaders/project/bilinear.kage
var shaderMapBilinearSrc []byte
var shaderMapBilinear = shaderRef{src: shaderMapBilinearSrc}

//go:embed shaders/project/warp_barrel.kage
var shaderWarpBarrelSrc []byte
var shaderWarpBarrel = shaderRef{src: shaderWarpBarrelSrc}

//go:embed shaders/project/warp_pincushion_quad.kage
var shaderWarpPincushionQuadSrc []byte
var shaderWarpPincushionQuad = shaderRef{src: shaderWarpPincushionQuadSrc}

//go:embed shaders/project/warp_arc.kage
var shaderWarpArcSrc []byte
var shaderWarpArc = shaderRef{src: shaderWarpArcSrc}

//go:embed shaders/noise/white.kage
var shaderNoiseSrc []byte
var shaderNoise = shaderRef{src: shaderNoiseSrc}

//go:embed shaders/noise/golden.kage
var shaderNoiseGoldenSrc []byte
var shaderNoiseGolden = shaderRef{src: shaderNoiseGoldenSrc}

//go:embed shaders/tile/rects_grid.kage
var shaderTileRectsGridSrc []byte
var shaderTileRectsGrid = shaderRef{src: shaderTileRectsGridSrc}

//go:embed shaders/tile/dots_grid.kage
var shaderTileDotsGridSrc []byte
var shaderTileDotsGrid = shaderRef{src: shaderTileDotsGridSrc}

//go:embed shaders/tile/dots_hex.kage
var shaderTileDotsHexSrc []byte
var shaderTileDotsHex = shaderRef{src: shaderTileDotsHexSrc}

//go:embed shaders/tile/tri_up_grid.kage
var shaderTileTriUpGridSrc []byte
var shaderTileTriUpGrid = shaderRef{src: shaderTileTriUpGridSrc}

//go:embed shaders/tile/tri_hex.kage
var shaderTileTriHexSrc []byte
var shaderTileTriHex = shaderRef{src: shaderTileTriHexSrc}

//go:embed shaders/morph/jfm_pass.kage
var shaderJFMPassSrc []byte
var shaderJFMPass = shaderRef{src: shaderJFMPassSrc}

//go:embed shaders/morph/jfm_init_fill.kage
var shaderJFMInitFillSrc []byte
var shaderJFMInitFill = shaderRef{src: shaderJFMInitFillSrc}

//go:embed shaders/morph/jfm_init_boundary.kage
var shaderJFMInitBoundarySrc []byte
var shaderJFMInitBoundary = shaderRef{src: shaderJFMInitBoundarySrc}

//go:embed shaders/morph/jfm_heat.kage
var shaderJFMHeatSrc []byte
var shaderJFMHeat = shaderRef{src: shaderJFMHeatSrc}

//go:embed shaders/morph/jfm_expansion.kage
var shaderJFMExpansionSrc []byte
var shaderJFMExpansion = shaderRef{src: shaderJFMExpansionSrc}

//go:embed shaders/morph/jfm_erosion.kage
var shaderJFMErosionSrc []byte
var shaderJFMErosion = shaderRef{src: shaderJFMErosionSrc}

//go:embed shaders/misc/halftone_tri.kage
var shaderHalftoneTriSrc []byte
var shaderHalftoneTri = shaderRef{src: shaderHalftoneTriSrc}

//go:embed shaders/misc/dither_matrix4.kage
var shaderDitherMat4Src []byte
var shaderDitherMat4 = shaderRef{src: shaderDitherMat4Src}

//go:embed shaders/misc/scanlines_sharp.kage
var shaderScanlinesSharpSrc []byte
var shaderScanlinesSharp = shaderRef{src: shaderScanlinesSharpSrc}

//go:embed shaders/misc/wave_lines.kage
var shaderWaveLinesSrc []byte
var shaderWaveLines = shaderRef{src: shaderWaveLinesSrc}

//go:embed shaders/misc/text_bilinear.kage
var shaderTextBilinearSrc []byte
var shaderTextBilinear = shaderRef{src: shaderTextBilinearSrc}

//go:embed shaders/misc/study_wave_funcs.kage
var shaderStudyWaveFuncsSrc []byte
var shaderStudyWaveFuncs = shaderRef{src: shaderStudyWaveFuncsSrc}
