package shapes

import (
	"slices"
	"testing"
)

// Kage shaders can't import code from each other, so shaders that need the
// same helper functions keep their own copies. This test checks the copies
// that different operations rely on to give consistent results.
func TestSharedShaderCode(t *testing.T) {
	// all the bilinear operations must align pixels and clamp at the borders
	// in the same way, or some of them end up shifted by half a pixel
	t.Run("bilinear", func(t *testing.T) {
		bilinear := namedShader{"Bilinear", shaderBilinearSrc}
		drawTintBilinear := namedShader{"DrawTintBilinear", shaderDrawTintBilinearSrc}
		colorMixBilinear := namedShader{"ColorMixBilinear", shaderColorMixBilinearSrc}
		textBilinear := namedShader{"TextBilinear", shaderTextBilinearSrc}
		mask := namedShader{"Mask", shaderMaskSrc}
		maskThreshold := namedShader{"MaskThreshold", shaderMaskThresholdSrc}
		masks := []namedShader{
			mask, {"MaskAt", shaderMaskAtSrc}, {"MaskCirc", shaderMaskCircSrc},
			{"MaskHorz", shaderMaskHorzSrc}, maskThreshold,
		}
		warps := []namedShader{
			{"WarpArc", shaderWarpArcSrc}, {"WarpBarrel", shaderWarpBarrelSrc},
			{"WarpPincushionQuad", shaderWarpPincushionQuadSrc},
		}
		mapBilinear := namedShader{"MapBilinear", shaderMapBilinearSrc}
		filteredMaps := []namedShader{
			{"MapAffineAnisotropic", shaderMapAffineAnisotropicSrc},
			{"MapQuadBilinear", shaderMapQuadBilinearSrc},
			{"MapQuadAnisotropic", shaderMapQuadAnisotropicSrc},
		}
		maps := slices.Concat([]namedShader{mapBilinear}, filteredMaps)

		mixing := slices.Concat([]namedShader{bilinear, drawTintBilinear, colorMixBilinear}, masks, maps, warps)
		srcIndexing := slices.Concat([]namedShader{drawTintBilinear, colorMixBilinear}, masks, maps, warps)
		dstIndexing := slices.Concat([]namedShader{bilinear, textBilinear, mask, maskThreshold}, maps)
		clampedRawSampling := slices.Concat([]namedShader{bilinear}, maps)
		srcSampling := slices.Concat([]namedShader{drawTintBilinear}, masks, warps)

		checkSharedShaderFuncs(t, mixing, "bilinearMix")
		checkSharedShaderFuncs(t, srcIndexing, "srcBilinearIndicesAndBlend")
		checkSharedShaderFuncs(t, dstIndexing, "dstBilinearIndicesAndBlend")
		checkSharedShaderFuncs(t, []namedShader{bilinear, colorMixBilinear}, "rawBilinear0")
		checkSharedShaderFuncs(t, clampedRawSampling, "clampedRawBilinear0", "clampSourcePixelIndices")
		checkSharedShaderFuncs(t, srcSampling, "srcBilinear0")
		checkSharedShaderFuncs(t, filteredMaps, "clampedSrcBilinear0", "clampedDstBilinear0")
	})

	// the same Oklab color must become the same RGB color in all the color
	// operations
	t.Run("oklab", func(t *testing.T) {
		colorizeByLightness := namedShader{"ColorizeByLightness", shaderColorizeByLightnessSrc}
		oklabShift := namedShader{"OklabShift", shaderOklabShiftSrc}
		checkSharedShaderFuncs(t, []namedShader{
			{"Gradient", shaderGradientSrc}, {"GradientRadial", shaderGradientRadialSrc},
			colorizeByLightness, oklabShift,
		}, "OklabToRGB", "OklabToLinRGB", "unlinearize")
		checkSharedShaderFuncs(t, []namedShader{colorizeByLightness, oklabShift}, "linearize")
	})

	// the jfm shaders must measure distances and decode offsets in the same
	// way, or the seeds stored in the maps aren't the nearest ones for the
	// shaders that read them (the reference is jfm_read.kage)
	t.Run("jfm", func(t *testing.T) {
		read := namedShader{"JFMRead", shaderJFMReadSrc}
		heat := namedShader{"JFMHeat", shaderJFMHeatSrc}
		viz := namedShader{"JFMViz", shaderJFMVizSrc}
		pass := namedShader{"JFMPass", shaderJFMPassSrc}
		init := namedShader{"JFMInit", shaderJFMInitSrc}
		checkSharedShaderFuncs(t, []namedShader{read, heat, viz}, "signedDistance")
		checkSharedShaderFuncs(t, []namedShader{read, heat, viz, pass},
			"seedDist", "regionOffset", "normalEdgeDist", "nearestBoundarySign",
			"jfaComputeOffsetToSeed", "jfaDecodeAxisOffsetToSeed")
		checkSharedShaderFuncs(t, []namedShader{read, heat, viz, pass, init}, "within")
	})

	// all the soft shapes must fade their edges with the same curve
	t.Run("soft edges", func(t *testing.T) {
		checkSharedShaderFuncs(t, []namedShader{
			{"CircleSoftBlur", shaderCircleSoftBlurSrc}, {"LineSoftBlur", shaderLineSoftBlurSrc},
			{"QuadSoftBlur", shaderQuadSoftBlurSrc}, {"RectSoftBlur", shaderRectSoftBlurSrc},
		}, "erfApproxSDF")
	})
}
