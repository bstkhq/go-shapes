package shapes

import "github.com/hajimehoshi/ebiten/v2"

func (r *Renderer) studyWaveFuncs(target *ebiten.Image, widthFactor, halfAmplitude float32) {
	r.setFlatCustomVAs01(widthFactor, halfAmplitude)
	r.DrawShader(target, 0, 0, shaderStudyWaveFuncs.Load())
}
