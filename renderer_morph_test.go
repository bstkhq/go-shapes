package shapes

import (
	"fmt"
	"image/color"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

// go test -run ^TestMorphExpansion$ . -count 1
func TestMorphExpansion(t *testing.T) {
	const (
		Radius       = 64.0
		MaxThickness = 16.0
	)

	thickness := float32(0)
	softEdge := float32(1.333)
	animateThickness := true
	smoothMovement := false
	updater := func(ctx TestAppCtx) {
		thickness = updateParam(ctx, ebiten.KeyR, thickness, 0, MaxThickness, 0.25)
		softEdge = updateParam(ctx, ebiten.KeyE, softEdge, 0, MaxThickness*2, 0.25)
		animateThickness = updateToggle(ctx, ebiten.KeyT, animateThickness)
		smoothMovement = updateToggle(ctx, ebiten.KeyM, smoothMovement)
	}
	drawer := func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Fill(backTestColor)

		currentThickness := thickness
		if animateThickness {
			currentThickness = float32(ctx.DistAnim(MaxThickness, 0.5))
		}
		effectiveSoftEdge := min(softEdge, currentThickness)
		shiftX, shiftY := float32(0), float32(0)
		if smoothMovement {
			shiftX = float32(-4.0 + ctx.DistAnim(8.0, 0.777))
			shiftY = float32(-4.0 + ctx.DistAnim(8.0, 0.5))
		}

		ctx.Renderer.SetColor(color.White)
		info := fmt.Sprintf(
			"Thickness: %.2f [R/Shift+R]\nSoft edge: %.2f (effective: %.2f) [E/Shift+E]\nThickness animation: %t [T]\nSmooth movement: %t [M]",
			currentThickness, softEdge, effectiveSoftEdge, animateThickness, smoothMovement,
		)
		ctx.Renderer.Text(canvas, info, 8, 8, TextOpts(1.0, TopLeft.Snap(CapLine)))

		morphBlend := ebiten.BlendSourceOver
		if ctx.SpacePressed {
			morphBlend = ebiten.BlendCopy
		}
		lc := ctx.LeftClickF32().AddXY(shiftX, shiftY)
		rc := ctx.RightClickF32().AddXY(shiftX, shiftY)

		// Standalone circle.
		ctx.Renderer.Options().Blend = morphBlend
		ctx.Renderer.SetColor(color.White)
		ctx.Renderer.MorphExpansion(canvas, ctx.Images[0], rc.X-Radius, lc.Y-Radius, currentThickness, softEdge)

		// Compare the result against the source and maximum-expansion bounds.
		comparisonY := rc.Y
		ctx.Renderer.Options().Blend = ebiten.BlendSourceOver
		ctx.Renderer.SetColor(color.RGBA{255, 0, 0, 255})
		ctx.Renderer.FillCircle(canvas, lc.X, comparisonY, Radius+MaxThickness)
		ctx.Renderer.SetColor(color.White)
		ctx.Renderer.FillCircle(canvas, lc.X, comparisonY, Radius)
		ctx.Renderer.SetColor(color.RGBA{0, 0, 164, 164})
		ctx.Renderer.MorphExpansion(canvas, ctx.Images[0], lc.X-Radius, comparisonY-Radius, currentThickness, softEdge)

		// Non-circular source.
		ctx.Renderer.Options().Blend = morphBlend
		triW, triH := rectSizeF32(ctx.Images[1].Bounds())
		ctx.Renderer.SetColor(color.RGBA{172, 0, 224, 255})
		ctx.Renderer.MorphExpansion(canvas, ctx.Images[1], rc.X-triW/2.0, rc.Y-triH/2.0, currentThickness, softEdge)
		ctx.Renderer.Options().Blend = ebiten.BlendSourceOver
	}

	app := NewTestApp(updater, drawer)
	circle := app.Renderer.NewFilledCircle(Radius)
	triangle := ebiten.NewImage(160, 128)
	triPoints := [3]PointF32{{X: 12, Y: 12}, {X: 148, Y: 30}, {X: 44, Y: 116}}
	app.Renderer.FillTriangle(triangle, triPoints, 0)
	app.Images = append(app.Images, circle, triangle)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestMorphExpansionRect$ . -count 1
func TestMorphExpansionRect(t *testing.T) {
	const Radius = 64.0
	const Expansion = 32.0

	updater := func(TestAppCtx) {}
	drawer := func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Fill(color.Black)

		lc := ctx.LeftClickF32()
		rc := ctx.RightClickF32()
		ctx.Renderer.SetColor(color.RGBA{255, 0, 0, 255})
		expansion := float32(ctx.DistAnim(Expansion, 1.0))
		ctx.Renderer.MorphExpansionRect(canvas, ctx.Images[0], lc.X-Radius, lc.Y-Radius, expansion)
		ctx.Renderer.MorphExpansionRect(canvas, ctx.Images[1], rc.X-Radius, rc.Y-Radius, expansion)

		ctx.Renderer.SetColor(color.RGBA{255, 0, 0, 255})
		ctx.DrawWithAlphaAtF32(canvas, ctx.Images[0], 0.5, lc.X-Radius, lc.Y-Radius)
		ctx.DrawWithAlphaAtF32(canvas, ctx.Images[3], 0.5, lc.X-Radius-Expansion, lc.Y-Radius-Expansion)
		ctx.DrawWithAlphaAtF32(canvas, ctx.Images[1], 0.5, rc.X-Radius, rc.Y-Radius)
		ctx.DrawWithAlphaAtF32(canvas, ctx.Images[2], 0.5, rc.X-Radius-Expansion, rc.Y-Radius-Expansion)
	}

	app := NewTestApp(updater, drawer)
	img1 := app.Renderer.NewFilledRect(int(Radius*2), int(Radius*2))
	img4 := app.Renderer.NewFilledRect(int((Radius+Expansion)*2), int((Radius+Expansion)*2))
	img2 := app.Renderer.NewFilledCircle(float64(Radius))
	img3 := app.Renderer.NewFilledCircle(float64(Radius + Expansion))
	app.Images = append(app.Images, img1, img2, img3, img4)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestMorphErosion$ . -count 1
func TestMorphErosion(t *testing.T) {
	const (
		Radius       = 96.0
		MaxThickness = 16.0
	)

	thickness := float32(0)
	softEdge := float32(0.666)
	animateThickness := true
	smoothMovement := false
	strokeTriangle := false
	updater := func(ctx TestAppCtx) {
		thickness = updateParam(ctx, ebiten.KeyR, thickness, 0, MaxThickness, 0.25)
		softEdge = updateParam(ctx, ebiten.KeyE, softEdge, 0, MaxThickness*2, 0.25)
		animateThickness = updateToggle(ctx, ebiten.KeyT, animateThickness)
		smoothMovement = updateToggle(ctx, ebiten.KeyM, smoothMovement)
		strokeTriangle = updateToggle(ctx, ebiten.KeyS, strokeTriangle)
	}
	drawer := func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Fill(backTestColor)

		currentThickness := thickness
		if animateThickness {
			currentThickness = float32(ctx.DistAnim(MaxThickness, 0.5))
		}
		effectiveSoftEdge := min(softEdge, currentThickness)
		shiftX, shiftY := float32(0), float32(0)
		if smoothMovement {
			shiftX = float32(-4.0 + ctx.DistAnim(8.0, 0.777))
			shiftY = float32(-4.0 + ctx.DistAnim(8.0, 0.5))
		}

		ctx.Renderer.SetColor(color.White)
		triangleType := "Filled"
		triangleIndex := 1
		if strokeTriangle {
			triangleType = "Stroked"
			triangleIndex = 2
		}
		info := fmt.Sprintf(
			"Thickness: %.2f [R/Shift+R]\nSoft edge: %.2f (effective: %.2f) [E/Shift+E]\nThickness animation: %t [T]\nSmooth movement: %t [M]\nTriangle: %s [S]",
			currentThickness, softEdge, effectiveSoftEdge, animateThickness, smoothMovement, triangleType,
		)
		ctx.Renderer.Text(canvas, info, 8, 8, TextOpts(1.0, TopLeft.Snap(CapLine)))

		morphBlend := ebiten.BlendSourceOver
		if ctx.SpacePressed {
			morphBlend = ebiten.BlendCopy
		}
		lc := ctx.LeftClickF32().AddXY(shiftX, shiftY)
		rc := ctx.RightClickF32().AddXY(shiftX, shiftY)

		// Standalone circle.
		ctx.Renderer.Options().Blend = morphBlend
		ctx.Renderer.SetColor(color.White)
		ctx.Renderer.MorphErosion(canvas, ctx.Images[0], rc.X-Radius, lc.Y-Radius, currentThickness, softEdge)

		// Compare the result against the source and maximum-erosion bounds.
		comparisonY := rc.Y
		ctx.Renderer.Options().Blend = ebiten.BlendSourceOver
		ctx.Renderer.SetColor(color.White)
		ctx.Renderer.FillCircle(canvas, lc.X, comparisonY, Radius)
		ctx.Renderer.SetColor(color.RGBA{255, 0, 0, 255})
		ctx.Renderer.FillCircle(canvas, lc.X, comparisonY, Radius-MaxThickness)
		ctx.Renderer.SetColor(color.RGBA{0, 0, 164, 164})
		ctx.Renderer.MorphErosion(canvas, ctx.Images[0], lc.X-Radius, comparisonY-Radius, currentThickness, softEdge)

		// Non-circular source.
		ctx.Renderer.Options().Blend = morphBlend
		triangle := ctx.Images[triangleIndex]
		triW, triH := rectSizeF32(triangle.Bounds())
		ctx.Renderer.SetColor(color.RGBA{172, 0, 224, 255})
		ctx.Renderer.MorphErosion(canvas, triangle, rc.X-triW/2.0, rc.Y-triH/2.0, currentThickness, softEdge)
		ctx.Renderer.Options().Blend = ebiten.BlendSourceOver
	}

	app := NewTestApp(updater, drawer)
	circle := app.Renderer.NewFilledCircle(Radius)
	filledTriangle := ebiten.NewImage(256, 164)
	strokeTriangleImg := ebiten.NewImage(256, 164)
	var points [3]PointF32
	points[0] = PointF32{X: 16, Y: 16}
	points[1] = points[0].AddXY(224, 0)
	points[2] = points[0].AddXY(0, 132)
	app.Renderer.FillTriangle(filledTriangle, points, 0)
	app.Renderer.StrokeTriangle(strokeTriangleImg, points, MaxThickness*2, 0)
	app.Images = append(app.Images, circle, filledTriangle, strokeTriangleImg)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}

// go test -run ^TestMorphOutline$ . -count 1
func TestMorphOutline(t *testing.T) {
	const (
		Radius       = 64.0
		MinThickness = -8.0
		MaxThickness = 16.0
	)

	colorSets := newTestColorSets(tcsSingle("White", tcWhite), tcsBase4, tcsPastel4, tcsVivid4)
	thickness := float32(8)
	softEdge := float32(0)
	animateThickness := false
	smoothMovement := false
	showSources := false
	updater := func(ctx TestAppCtx) {
		thickness = updateParam(ctx, ebiten.KeyR, thickness, MinThickness, MaxThickness, 0.25)
		softEdge = updateParam(ctx, ebiten.KeyE, softEdge, 0, MaxThickness, 0.25)
		animateThickness = updateToggle(ctx, ebiten.KeyT, animateThickness)
		smoothMovement = updateToggle(ctx, ebiten.KeyM, smoothMovement)
		colorSets.Update(ebiten.KeyS)
		showSources = ebiten.IsKeyPressed(ebiten.KeyC)
	}
	drawer := func(canvas *ebiten.Image, ctx TestAppCtx) {
		canvas.Fill(backTestColor)

		currentThickness := thickness
		if animateThickness {
			currentThickness = MinThickness + float32(ctx.DistAnim(MaxThickness-MinThickness, 0.5))
		}
		inThickness, outThickness := currentThickness/2.0, currentThickness/2.0
		if currentThickness < 0 {
			inThickness, outThickness = -currentThickness, 0
		}
		effectiveSoftEdge := min(softEdge, inThickness)

		xShift, yShift := float32(0), float32(0)
		if smoothMovement {
			xShift = float32(-4.0 + ctx.DistAnim(8.0, 0.777))
			yShift = float32(-4.0 + ctx.DistAnim(8.0, 0.5))
		}

		ctx.Renderer.SetColor(color.White)
		info := fmt.Sprintf(
			"Thickness: %.2f [R/Shift+R] (-8 inner to +16 balanced)\nSoft edge: %.2f (effective: %.2f) [E/Shift+E]\nThickness animation: %t [T]\nColor set: %s [S]\nSources: %t [Hold C]\nSmooth movement: %t [M]",
			currentThickness, softEdge, effectiveSoftEdge, animateThickness, colorSets.Info(), showSources, smoothMovement,
		)
		ctx.Renderer.Text(canvas, info, 8, 8, TextOpts(1.0, TopLeft.Snap(CapLine)))

		lc := ctx.LeftClickF32().AddXY(xShift, yShift)
		ctx.Renderer.SetColor(color.RGBA{0, 255, 0, 255})
		ctx.Renderer.FillCircle(canvas, lc.X, lc.Y, Radius+outThickness+1.0)
		ctx.Renderer.SetColor(color.RGBA{255, 0, 0, 255})
		ctx.Renderer.FillCircle(canvas, lc.X, lc.Y, Radius-inThickness-1.0)

		if ctx.SpacePressed {
			ctx.Renderer.Options().Blend = ebiten.BlendCopy
		}
		ctx.Renderer.SetColor(color.RGBA{0, 0, 255, 255})
		ctx.Renderer.MorphOutline(canvas, ctx.Images[0], lc.X-Radius, lc.Y-Radius, currentThickness, softEdge)

		colorSets.Apply(ctx.Renderer)
		rc := ctx.RightClickF32().AddXY(xShift, yShift)
		shape := ctx.Images[1]
		shapeW, shapeH := rectSizeF32(shape.Bounds())
		shapeOX, shapeOY := lc.X-shapeW/2.0, rc.Y-shapeH/2.0
		if showSources {
			ctx.DrawAtF32(canvas, ctx.Images[0], rc.X-Radius, rc.Y-Radius)
			ctx.DrawAtF32(canvas, shape, shapeOX, shapeOY)
		} else {
			ctx.Renderer.MorphOutline(canvas, ctx.Images[0], rc.X-Radius, rc.Y-Radius, currentThickness, softEdge)
			ctx.Renderer.MorphOutline(canvas, shape, shapeOX, shapeOY, currentThickness, softEdge)
		}
		ctx.Renderer.Options().Blend = ebiten.BlendSourceOver

	}

	app := NewTestApp(updater, drawer)
	circle := app.Renderer.NewFilledCircle(Radius)
	triangle := ebiten.NewImage(160, 128)
	triPoints := [3]PointF32{{X: 12, Y: 12}, {X: 148, Y: 30}, {X: 44, Y: 116}}
	app.Renderer.FillTriangle(triangle, triPoints, 0)
	app.Images = append(app.Images, circle, triangle)
	if err := ebiten.RunGame(app); err != nil {
		t.Fatal(err)
	}
}
