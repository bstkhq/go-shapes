# High quality shaders

There are significant differences between a shader that "seems to work" and a high quality shader. This document goes over a few concepts that must be taken into account when trying to write high quality shaders.

## Bounding

A shader can't be bounded by simply taking mins/maxes of the requested coordinates:
1) Even for axis-aligned rectangular shapes, floor/ceil need to be used to avoid triangle rasterization cutting part of the shape's edge. This is easy to miss with quick visual inspection alone, especially when no movement animation is tested.
2) When it comes to more complex shapes, we must consider hulls. We default to AABBs, but a `Hull` flag allows the optimization of some shapes. See the flag's docs for more details. The short version is that a tight hull can reduce the number of ineffective pixels in 50% when drawing a shape like a triangle, but many other complications have to be balanced: color interpolation can break and hull calculations can be complex (e.g. tight edges might require mitering during padding).
3) Stroked shapes are particularly problematic, as most pixels are ineffective when rendering thin outlines. Cases like these should definitely support hulls.

Implementing hulls for shapes and effects that can heavily benefit from it is recommended, but it should be opt-in in most cases.

## Smoothness under subtle movement

Making shaders render correctly under subtle, slow movement can be difficult. With SDFs this is typically not a problem, as the shape boundary is determined analytically, but when it comes to using images, bilinear interpolation might be required.

Bilinear interpolation has severe performance implications in Ebitengine, though. Having two shaders is quite common, and then the `Bilinear` flag is accepted to choose between the two.

Even in the cases where `Bilinear` is not supported, if float coordinates are supported, the shader should be checked for artifacts and behavior under smooth coordinate changes. A good example of this are blurs. Blurs are too expensive to perform bilinear lookups during the effect itself, but at least they should not break under float coordinates.

## Gamma correction

Gamma correction is also part of smooth movement and rendering. For detailed context, see [ebiten issue#3313](https://github.com/hajimehoshi/ebiten/issues/3313). The quick practical advice is that any shader where boundaries must be well defined should consider applying manual gamma correction, e.g.:
```Go
alpha = pow(alpha, 1.0/2.2)
return color * alpha
```

When writing a shader, you should consider and test this to see if it's relevant or not, if it makes a difference or not.

## Leveraging parameters and renderer configuration

- If rounding parameters are used, support positive for outer rounding (expansion), and negative for inner rounding. This often involves analytical geometry and shape collapse when done properly, and it can present hard cases for color interpolation and tight bounding if hulls are supported.
- Ensure that `Renderer.Tint()` is considered and applied if possible.
- Blending modes should be respected as much as possible, and it must be documented if an operation needs to force a specific blend mode. Blend modes must also be checked before skipping an operation; study `blendSafeToCrop` before assuming an operation can be skipped because "it wouldn't draw anything visible" (the blend itself can lead to visible effects). To debug, it's common to use a colored background and draw with `BlendClear` to check.

## Performance

- Use custom vertex attributes when possible to avoid breaking batching during consecutive calls to the same shader. In case of insufficient custom VAs, prioritize the VAs for the most commonly changing parameters (e.g. for animation).
- Use `sourceCoords` for positioning rather than `targetCoords` when possible. `targetCoords` need to use `imageDstOrigin()` and introduce an extra uniform internally for Ebitengine. `srcImageOrigin()` also introduces a uniform in the same way.
- Theoretically, we could implement batched versions of some multi-step functions. This is currently outside scope.

## Non-zero bounds

- Ensure target origin behavior matches default Ebitengine functions. Tests support pressing `0` to change the origin and check.

## Technical specification, correctness and consistency

All characteristics of a shader must be well defined, consistent and clearly specified. This seems like a silly remark, but it's quite hard to get right in practice:

- The dimensions of a shape must be as precise as possible. While it sounds like rendering a rectangle of size NxM is trivial, when bounds are clamped, edge smoothing is applied and so on, it's easy to get it wrong. For example, edge smoothing must be internal, not around or beyond the boundary.
- The rounding for a rounded rectangle shader must be exact. If you pass a radius of 6 and you overlap a circle of radius 6 on the corner, the two must match exactly. It's surprisingly easy to do this incorrectly, be off by one or two pixels due to padding, bound adjustments or others.
- A blur shader can't just "blur around the given radius". A blur must have a specific standard deviation, and if that's hardcoded, it should be as consistent as possible among different blur methods.
- We use smoothstep over a 1.333 margin as the anti-aliasing margin for most shapes that require it.

## Example case: drawing a triangle

> Uses three vertices to draw a triangle.

This is "incorrect" because edges will be jagged; we are implementing high quality shapes and effects.

> Implements SDF shader and applies it to min(points[:]), max(points[:]).

This is incorrect because at least the rectangle must be floored for the Min point and ceiled for the Max point.

> We are done.

Is the test implemented? Does it look good when moving smoothly? Has gamma correction been applied?

> Yes, it works.

Is rounding implemented? Allowing both signs for outwards and inwards rounding, and handling collapse into a circle when rounding is negative enough?

> Analytical geometry applied, inner rounding done, collapse to circle detected and applied.

Make sure to revise:
- Rounding matches a natural circle.
- Edge antialiasing is 1.333 (consistent with all other shapes), and it's smoothed inwardly (we never exceed the shape mathematical bounds).
- Bounds are being tightened during collapse.
- Target origin is respected.
- Tint is being passed and respected, probably through one of the custom VAs.
- `Hull` flag has been considered as triangles have ~50% ineffective pixels when drawn within an AABB.
