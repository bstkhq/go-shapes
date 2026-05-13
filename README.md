# shapes
[![Go Reference](https://pkg.go.dev/badge/github.com/erparts/go-shapes.svg)](https://pkg.go.dev/github.com/erparts/go-shapes)

`shapes` is a package for [**Ebitengine**](https://github.com/hajimehoshi/ebiten) (the 2D game library written by Hajime Hoshi) that allows rendering some common shapes and effects in complementary ways to the official [`vector`](https://pkg.go.dev/github.com/hajimehoshi/ebiten/v2/vector) package. Some examples:

- Smooth circles, triangles, hexagons, ellipses, rings, rects... with rounding options.
- High quality image blurs, glows and outlines.
- Oklab gradients and color functions.
- Utility image, shader and scaling methods.
- Tiling patterns with dots, triangles and rectangles.
- Many other effects like noise, warps, dithering, quad mapping, text...

Unlike `vector`, `shapes` relies more on [Kage shaders](https://github.com/tinne26/kage-desk) instead of raw triangles rasterization. This means rendering tends to be smoother, but extra care has to be taken as changing shaders or some of its parameters will break [batching](https://github.com/tinne26/efficient-ebitengine).

[TODO: simple UI surface example image]

# Credit

- Many of the SDFs are based on https://iquilezles.org/articles/distfunctions2d.
- The text rendering debug function uses an atlas derived from the [Ark Pixel Font](https://ark-pixel-font.takwolf.com), by TakWolf (OFL-1.1).

# Limitations

- Ebitengine still doesn't handle color linearization properly for shaders. A few shaders are compensating this manually, like those involving Oklab color conversions or those requiring smooth edges, but in many significant cases, like blurs, all this is ignored. Once Ebitengine provides linearized colors in Kage, this will have to be homogenized.
