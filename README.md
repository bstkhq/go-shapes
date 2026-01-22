# shapes
[![Go Reference](https://pkg.go.dev/badge/github.com/erparts/go-shapes.svg)](https://pkg.go.dev/github.com/erparts/go-shapes)

`shapes` is a package for [**Ebitengine**](https://github.com/hajimehoshi/ebiten) (the 2D game library written by Hajime Hoshi) that allows rendering some common shapes and effects in complementary ways to the official [`vector`](github.com/hajimehoshi/ebiten/v2/vector) package. Some examples:

- Smooth circles, triangles, hexagons, ellipses, rings, rects... (all with the option of rounding).
- High quality image outlines, expansions and blurs.
- Simple patterns like dots, rhombuses and triangles.
- Some color functions and gradients.
- Other misc effects like noise, dithering, warps, quad mapping...

Unlike `vector`, `shapes` relies more on simple [Kage shaders](https://github.com/tinne26/kage-desk) instead of raw triangles rasterization. This means rendering tends to be smoother, but extra care has to be taken as changing shaders or some of its parameters will break [batching](https://github.com/tinne26/efficient-ebitengine).

[TODO: simple UI surface example image]

# Credit

Many of the SDFs are based on https://iquilezles.org/articles/distfunctions2d.

# Pending improvements

- Ebitengine still doesn't handle color linearization properly for shaders. A few shaders are compensating this manually, like those involving Oklab color conversions or those requiring smooth edges, but in many significant cases, like blurs, all this is ignored. Once Ebitengine provides linearized colors in Kage, this will have to be homogenized.
- Revise target origins being properly applied everywhere, and document target/source origins behavior clearly.
- All shapes that accept rounding should support a positive rounding value for SDF expansion, and optionally, if inner rounding is allowed, handle it through negative values. Currently, most shapes do inner rounding with positive values, so newer functions can't even be properly adapted. This respects size boundaries but it's a missed opportunity.
