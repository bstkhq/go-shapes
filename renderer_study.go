package shapes

import "github.com/hajimehoshi/ebiten/v2"

func (r *Renderer) studyWaveFuncs(target *ebiten.Image, widthFactor, halfAmplitude float32) {
	r.setFlatCustomVAs01(widthFactor, halfAmplitude)
	tw, th := rectSizeF32(target.Bounds())
	r.DrawRectShader(target, 0, 0, tw, th, NoMargins, shaderStudyWaveFuncs.Load())
}
