package shapes

import "testing"

func TestTextSize(t *testing.T) {
	r := NewRenderer()
	fontMap := TextOptions{}.fontMap()
	ascent := fontMap.Ascent()
	descent := fontMap.Descent()
	lineGap := fontMap.LineGap()
	spaceWidth := fontMap.SpaceWidth()

	w, h := r.TextSize("\n", TextOptions{})
	if w != 0 || h != float32(ascent+descent) {
		t.Fatalf("expected (w, h) = (%.02f, %.02f), got (%.02f, %.02f)", 0.0, float32(ascent+descent), w, h)
	}

	line2H := 2*float32(ascent+descent) + float32(lineGap)
	w, h = r.TextSize("\n\n", TextOptions{})
	if w != 0 || h != line2H {
		t.Fatalf("expected (w, h) = (%.02f, %.02f), got (%.02f, %.02f)", 0.0, line2H, w, h)
	}

	w, h = r.TextSize("\n\n", TextOptions{LineGap: 7})
	if w != 0 || h != float32(2*(ascent+descent)+7) {
		t.Fatalf("expected (w, h) = (%.02f, %.02f), got (%.02f, %.02f)", 0.0, float32(2*(ascent+descent)+7), w, h)
	}

	_, h = r.TextSize("\nA", TextOptions{})
	if h != line2H {
		t.Fatalf("expected h = %.02f, got %.02f", line2H, h)
	}
	_, h = r.TextSize("A\n", TextOptions{})
	if h != line2H {
		t.Fatalf("expected h = %.02f, got %.02f", line2H, h)
	}
	_, h = r.TextSize("A\nBCD", TextOptions{})
	if h != line2H {
		t.Fatalf("expected h = %.02f, got %.02f", line2H, h)
	}

	w, h = r.TextSize(" ", TextOptions{Scale: 1.0})
	if w != float32(spaceWidth) || h != float32(ascent+descent) {
		t.Fatalf("expected (w, h) = (%.02f, %.02f), got (%.02f, %.02f)", float32(spaceWidth), float32(ascent+descent), w, h)
	}

	wm, _ := r.TextSize("M", TextOptions{Scale: 1.0})
	w, _ = r.TextSize("MM", TextOptions{Scale: 1.0})
	if w != wm+1.0+wm {
		t.Fatalf("expected w = %.02f, got %.02f", wm+1.0+wm, w)
	}
	w, _ = r.TextSize("M M", TextOptions{Scale: 1.0})
	if w != wm+float32(spaceWidth)+wm {
		t.Fatalf("expected w = %.02f, got %.02f", wm+float32(spaceWidth)+wm, w)
	}

	s1, s2 := "MMA", "MM\nMMA"
	w1, _ := r.TextSize(s1, TextOptions{})
	w2, h := r.TextSize(s2, TextOptions{})
	if w1 != w2 {
		t.Fatalf("width mismatch between %q (%.02f) and %q (%.02f)", s1, w1, s2, w2)
	}

	ws2, hs2 := r.TextSize(s2, TextOptions{Scale: 1.5})
	if ws2 != w2*1.5 || hs2 != h*1.5 { // note: might be fuzzy for big fonts
		t.Fatalf("expected (w, h) = (%.02f, %.02f), got (%.02f, %.02f)", float32(w2*1.5), float32(h*1.5), ws2, hs2)
	}
}
