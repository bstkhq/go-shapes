package shapes

import (
	"image"
	"regexp"
	"strings"
	"testing"

	"github.com/bstkhq/go-shapes/internal/pixcheck"
	"github.com/hajimehoshi/ebiten/v2"
)

// testShape is a small named source image for numerical tests, usually written
// by hand with pixcheck.ParseAlpha. The name is used in failure labels.
type testShape struct {
	name string
	img  *image.RGBA
}

// blobShape is an antialiased blob with a transparent margin.
var blobShape = testShape{"blob", pixcheck.ParseAlpha(`
	000 000 000 000 000 000 000 000 000 000
	000 000 032 128 192 192 128 032 000 000
	000 032 224 255 255 255 255 224 032 000
	000 128 255 255 255 255 255 255 128 000
	000 192 255 255 255 255 255 255 192 000
	000 192 255 255 255 255 255 255 192 000
	000 128 255 255 255 255 255 255 128 000
	000 032 224 255 255 255 255 224 032 000
	000 000 032 128 192 192 128 032 000 000
	000 000 000 000 000 000 000 000 000 000
`)}

// bandShape is an antialiased vertical band cut by the top and bottom borders.
var bandShape = testShape{"band", pixcheck.ParseAlpha(`
	000 064 255 255 255 255 255 192 000 000
	000 064 255 255 255 255 255 192 000 000
	000 064 255 255 255 255 255 192 000 000
	000 064 255 255 255 255 255 192 000 000
	000 064 255 255 255 255 255 192 000 000
	000 064 255 255 255 255 255 192 000 000
	000 064 255 255 255 255 255 192 000 000
	000 064 255 255 255 255 255 192 000 000
	000 064 255 255 255 255 255 192 000 000
	000 064 255 255 255 255 255 192 000 000
`)}

// fadeShape fades out towards the right, with a transparent margin.
var fadeShape = testShape{"fade", pixcheck.ParseAlpha(`
	000 000 000 000 000 000 000 000 000 000
	000 255 255 224 160 096 048 016 000 000
	000 255 255 224 160 096 048 016 000 000
	000 255 255 224 160 096 048 016 000 000
	000 255 255 224 160 096 048 016 000 000
	000 000 000 000 000 000 000 000 000 000
`)}

// cornerShape is an antialiased quarter disc cut by the top and left borders.
var cornerShape = testShape{"corner", pixcheck.ParseAlpha(`
	255 255 255 255 192 064 000 000
	255 255 255 240 160 032 000 000
	255 255 240 192 096 000 000 000
	255 240 192 112 016 000 000 000
	192 160 096 016 000 000 000 000
	064 032 000 000 000 000 000 000
	000 000 000 000 000 000 000 000
	000 000 000 000 000 000 000 000
`)}

// jfmRegions are the regions for jumping flood map tests: the default one, and
// custom ones whose boundaries behave differently. A threshold leaves the faint
// pixels out, a band also has an upper boundary, and the low side is the low
// alpha side of the shapes.
var jfmRegions = []struct {
	name               string
	minAlpha, maxAlpha float32
}{
	{"default", 0, 0},
	{"threshold", 0.5, 1},
	{"band", 0.2, 0.8},
	{"low side", 0, 0.5},
}

// alphaNormals returns the alpha normals of img, computed with the given flags.
func alphaNormals(r *Renderer, img *ebiten.Image, flags ...Flag) *ebiten.Image {
	normals := ebiten.NewImage(img.Bounds().Dx(), img.Bounds().Dy())
	r.AlphaNormalsCompute(normals, img, flags...)
	return normals
}

// padded returns a copy of img with transparent padding around it.
func padded(img *ebiten.Image, padding int) *ebiten.Image {
	w, h := img.Bounds().Dx(), img.Bounds().Dy()
	out := ebiten.NewImage(w+2*padding, h+2*padding)
	var opts ebiten.DrawImageOptions
	opts.Blend = ebiten.BlendCopy
	opts.GeoM.Translate(float64(padding), float64(padding))
	out.DrawImage(img, &opts)
	return out
}

// namedShader is a shader source with a name for test messages.
type namedShader struct {
	name string
	src  []byte
}

// checkSharedShaderFuncs checks that each of the given functions is defined in
// all the given shaders, and is the same as in the first one. Comments,
// whitespace and image indices are ignored, as copies can leave out comments
// and bind their images in a different order.
func checkSharedShaderFuncs(t *testing.T, shaders []namedShader, funcs ...string) {
	t.Helper()
	for _, name := range funcs {
		reference, found := shaderFunc(shaders[0].src, name)
		if !found {
			t.Errorf("%s isn't defined in %s", name, shaders[0].name)
			continue
		}
		for _, shader := range shaders[1:] {
			code, found := shaderFunc(shader.src, name)
			if !found {
				t.Errorf("%s isn't defined in %s", name, shader.name)
			} else if code != reference {
				t.Errorf("%s in %s isn't the same as in %s", name, shader.name, shaders[0].name)
			}
		}
	}
}

// shaderFunc returns the code of the named function in a shader source, without
// comments, whitespace or image indices, for checkSharedShaderFuncs.
func shaderFunc(src []byte, name string) (string, bool) {
	code := regexp.MustCompile(`(?ms)^func ` + regexp.QuoteMeta(name) + `\(.*?^}`).Find(src)
	if code == nil {
		return "", false
	}
	code = regexp.MustCompile(`//[^\n]*`).ReplaceAll(code, nil)
	code = regexp.MustCompile(`imageSrc\d`).ReplaceAll(code, []byte("imageSrcN"))
	return strings.Join(strings.Fields(string(code)), " "), true
}
