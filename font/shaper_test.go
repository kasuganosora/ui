package font

import "testing"

func setupShaper() (Shaper, ID) {
	engine := newMockEngine()
	mgr := NewManager(engine)
	id, _ := mgr.Register("Test", WeightRegular, StyleNormal, nil)
	return NewShaper(mgr), id
}

func TestShaperShapeEmpty(t *testing.T) {
	shaper, id := setupShaper()
	runs := shaper.Shape("", ShapeOptions{FontID: id, FontSize: 16})
	if len(runs) != 0 {
		t.Errorf("expected 0 runs, got %d", len(runs))
	}
}

func TestShaperShapeSingleLine(t *testing.T) {
	shaper, id := setupShaper()
	runs := shaper.Shape("Hello", ShapeOptions{FontID: id, FontSize: 16})
	if len(runs) != 1 {
		t.Fatalf("expected 1 run, got %d", len(runs))
	}
	if len(runs[0].Glyphs) != 5 {
		t.Errorf("expected 5 glyphs, got %d", len(runs[0].Glyphs))
	}
}

func TestShaperShapeNewline(t *testing.T) {
	shaper, id := setupShaper()
	runs := shaper.Shape("A\nB", ShapeOptions{FontID: id, FontSize: 16})
	if len(runs) != 2 {
		t.Fatalf("expected 2 runs, got %d", len(runs))
	}
	if len(runs[0].Glyphs) != 1 {
		t.Errorf("line 1: expected 1 glyph, got %d", len(runs[0].Glyphs))
	}
	if len(runs[1].Glyphs) != 1 {
		t.Errorf("line 2: expected 1 glyph, got %d", len(runs[1].Glyphs))
	}
}

func TestShaperShapeWordWrap(t *testing.T) {
	shaper, id := setupShaper()
	// Font size 16, advance = 16*0.6 = 9.6 per char
	// "Hello World" = 11 chars, total ~105.6
	// MaxWidth 60 should force a wrap
	runs := shaper.Shape("Hello World", ShapeOptions{
		FontID:   id,
		FontSize: 16,
		MaxWidth: 60,
	})
	if len(runs) < 2 {
		t.Fatalf("expected at least 2 runs for word wrap, got %d", len(runs))
	}
}

func TestShaperMeasureSingleLine(t *testing.T) {
	shaper, id := setupShaper()
	m := shaper.Measure("ABC", ShapeOptions{FontID: id, FontSize: 20})
	// 3 chars * 20*0.6 = 36
	expectedW := float32(3 * 20 * 0.6)
	if m.Width < expectedW-1 || m.Width > expectedW+1 {
		t.Errorf("expected width ~%.1f, got %.1f", expectedW, m.Width)
	}
	if m.LineCount != 1 {
		t.Errorf("expected 1 line, got %d", m.LineCount)
	}
}

func TestShaperMeasureMultiLine(t *testing.T) {
	shaper, id := setupShaper()
	m := shaper.Measure("A\nB\nC", ShapeOptions{FontID: id, FontSize: 16})
	if m.LineCount != 3 {
		t.Errorf("expected 3 lines, got %d", m.LineCount)
	}
}

func TestShaperMeasureWordWrap(t *testing.T) {
	shaper, id := setupShaper()
	m := shaper.Measure("Hello World", ShapeOptions{
		FontID:   id,
		FontSize: 16,
		MaxWidth: 60,
	})
	if m.LineCount < 2 {
		t.Errorf("expected at least 2 lines for word wrap, got %d", m.LineCount)
	}
}

func TestShaperAlignCenter(t *testing.T) {
	shaper, id := setupShaper()
	runs := shaper.Shape("Hi", ShapeOptions{
		FontID:   id,
		FontSize: 16,
		MaxWidth: 200,
		Align:    TextAlignCenter,
	})
	if len(runs) != 1 {
		t.Fatalf("expected 1 run, got %d", len(runs))
	}
	// "Hi" = 2 chars * 9.6 = 19.2 wide, centered in 200 -> offset ~90.4
	if runs[0].Glyphs[0].X < 80 {
		t.Errorf("expected centered X > 80, got %.1f", runs[0].Glyphs[0].X)
	}
}

func TestShaperAlignRight(t *testing.T) {
	shaper, id := setupShaper()
	runs := shaper.Shape("Hi", ShapeOptions{
		FontID:   id,
		FontSize: 16,
		MaxWidth: 200,
		Align:    TextAlignRight,
	})
	if len(runs) != 1 {
		t.Fatalf("expected 1 run, got %d", len(runs))
	}
	// "Hi" = 19.2 wide, right-aligned in 200 -> first glyph X ~180.8
	if runs[0].Glyphs[0].X < 170 {
		t.Errorf("expected right-aligned X > 170, got %.1f", runs[0].Glyphs[0].X)
	}
}

func TestShaperCJK(t *testing.T) {
	shaper, id := setupShaper()
	runs := shaper.Shape("你好世界", ShapeOptions{FontID: id, FontSize: 16})
	if len(runs) != 1 {
		t.Fatalf("expected 1 run, got %d", len(runs))
	}
	if len(runs[0].Glyphs) != 4 {
		t.Errorf("expected 4 glyphs, got %d", len(runs[0].Glyphs))
	}
}

func TestIsBreakOpportunity(t *testing.T) {
	if !isBreakOpportunity(' ') {
		t.Error("space should be a break opportunity")
	}
	if !isBreakOpportunity('-') {
		t.Error("hyphen should be a break opportunity")
	}
	if isBreakOpportunity('A') {
		t.Error("'A' should not be a break opportunity")
	}
}

func TestIsCJK(t *testing.T) {
	if !isCJK('中') {
		t.Error("'中' should be CJK")
	}
	if !isCJK('日') {
		t.Error("'日' should be CJK")
	}
	if isCJK('A') {
		t.Error("'A' should not be CJK")
	}
	if !isCJK('あ') {
		t.Error("'あ' (hiragana) should be CJK")
	}
	if !isCJK('カ') {
		t.Error("'カ' (katakana) should be CJK")
	}
}
