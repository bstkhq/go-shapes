package shapes

import "github.com/hajimehoshi/ebiten/v2"

func (r *Renderer) studyWaveFuncs(target *ebiten.Image, widthFactor, halfAmplitude float32) {
	r.setFlatCustomVAs01(widthFactor, halfAmplitude)
	tox, toy, tw, th := rectOriginSizeF32(target.Bounds())
	r.DrawRectShader(target, tox, toy, tw, th, NoMargins, RegionExact, shaderStudyWaveFuncs.Load())
}
