package font

import "testing"

func TestClassifyPunct(t *testing.T) {
	tests := []struct {
		r    rune
		want punctType
	}{
		{'「', punctOpening},
		{'（', punctOpening},
		{'《', punctOpening},
		{'【', punctOpening},
		{'\u201C', punctOpening}, // "
		{'」', punctClosing},
		{'）', punctClosing},
		{'》', punctClosing},
		{'】', punctClosing},
		{'\u201D', punctClosing}, // "
		{'、', punctMiddle},
		{'。', punctMiddle},
		{'，', punctMiddle},
		{'：', punctMiddle},
		{'；', punctMiddle},
		{'！', punctMiddle},
		{'？', punctMiddle},
		{'A', punctNone},
		{'中', punctNone},
		{' ', punctNone},
	}
	for _, tc := range tests {
		got := classifyPunct(tc.r)
		if got != tc.want {
			t.Errorf("classifyPunct(%q) = %d, want %d", tc.r, got, tc.want)
		}
	}
}

func TestCompressPunctuationLineStart(t *testing.T) {
	// 「你好」— opening at line start should be compressed
	runes := []rune{'「', '你', '好', '」'}
	advances := []float32{10, 10, 10, 10}
	comps := compressPunctuation(runes, advances, 5, true)

	// Should have compression for '「' at index 0 (line-start opening)
	hasLineStart := false
	for _, c := range comps {
		if c.index == 0 && c.trimLeft > 0 {
			hasLineStart = true
		}
	}
	if !hasLineStart {
		t.Error("expected line-start compression for opening punctuation")
	}
}

func TestCompressPunctuationAdjacentClosingOpening(t *testing.T) {
	// 」「 — closing followed by opening should compress
	runes := []rune{'好', '」', '「', '世'}
	advances := []float32{10, 10, 10, 10}
	comps := compressPunctuation(runes, advances, 5, true)

	hasAdjacentComp := false
	for _, c := range comps {
		if c.index == 2 && c.trimLeft > 0 { // Opening '「' gets left trimmed
			hasAdjacentComp = true
		}
	}
	if !hasAdjacentComp {
		t.Error("expected adjacent closing+opening compression")
	}
}

func TestCompressPunctuationAdjacentClosingClosing(t *testing.T) {
	// 」） — two closing punctuation adjacent
	runes := []rune{'好', '」', '）'}
	advances := []float32{10, 10, 10}
	comps := compressPunctuation(runes, advances, 5, true)

	hasComp := false
	for _, c := range comps {
		if c.index == 1 && c.trimRight > 0 { // First closing gets right trimmed
			hasComp = true
		}
	}
	if !hasComp {
		t.Error("expected adjacent closing+closing compression")
	}
}

func TestCompressPunctuationAdjacentMiddleOpening(t *testing.T) {
	// 。「 — middle followed by opening
	runes := []rune{'好', '。', '「', '世'}
	advances := []float32{10, 10, 10, 10}
	comps := compressPunctuation(runes, advances, 5, true)

	hasComp := false
	for _, c := range comps {
		if c.index == 2 && c.trimLeft > 0 {
			hasComp = true
		}
	}
	if !hasComp {
		t.Error("expected middle+opening compression")
	}
}

func TestCompressPunctuationNoPunctuation(t *testing.T) {
	runes := []rune{'你', '好', '世', '界'}
	advances := []float32{10, 10, 10, 10}
	comps := compressPunctuation(runes, advances, 5, true)

	if len(comps) != 0 {
		t.Errorf("expected no compression for plain text, got %d", len(comps))
	}
}

func TestCompressPunctuationEmpty(t *testing.T) {
	comps := compressPunctuation(nil, nil, 5, true)
	if comps != nil {
		t.Error("expected nil for empty input")
	}
}

func TestApplyPunctCompression(t *testing.T) {
	// 「你好」 with line-start compression
	glyphs := []PositionedGlyph{
		{X: 0, Advance: 10},  // 「
		{X: 10, Advance: 10}, // 你
		{X: 20, Advance: 10}, // 好
		{X: 30, Advance: 10}, // 」
	}
	comps := []punctCompression{
		{index: 0, trimLeft: 5}, // Line-start opening
	}

	totalTrim := applyPunctCompression(glyphs, comps)
	if totalTrim != 5 {
		t.Errorf("expected total trim 5, got %f", totalTrim)
	}
	if glyphs[0].X != -5 {
		t.Errorf("expected glyph[0].X = -5, got %f", glyphs[0].X)
	}
	if glyphs[1].X != 5 {
		t.Errorf("expected glyph[1].X = 5, got %f", glyphs[1].X)
	}
}

func TestApplyPunctCompressionRightTrim(t *testing.T) {
	// 」） with right trim on first closing
	glyphs := []PositionedGlyph{
		{X: 0, Advance: 10},  // 」
		{X: 10, Advance: 10}, // ）
	}
	comps := []punctCompression{
		{index: 0, trimRight: 5},
	}

	totalTrim := applyPunctCompression(glyphs, comps)
	if totalTrim != 5 {
		t.Errorf("expected total trim 5, got %f", totalTrim)
	}
	if glyphs[0].Advance != 5 {
		t.Errorf("expected glyph[0].Advance = 5, got %f", glyphs[0].Advance)
	}
	if glyphs[1].X != 5 {
		t.Errorf("expected glyph[1].X = 5, got %f", glyphs[1].X)
	}
}

func TestShaperCJKPunctCompression(t *testing.T) {
	// Integration test: shape text with CJK punctuation
	engine := newMockEngine()
	mgr := NewManager(engine)
	id, _ := mgr.Register("Test", WeightRegular, StyleNormal, nil)
	shaper := NewShaper(mgr)

	// 「你好」世界 — opening '「' at line start should be compressed
	runs := shaper.Shape("「你好」世界", ShapeOptions{FontID: id, FontSize: 16})
	if len(runs) != 1 {
		t.Fatalf("expected 1 run, got %d", len(runs))
	}

	run := runs[0]
	if len(run.Glyphs) != 6 {
		t.Fatalf("expected 6 glyphs, got %d", len(run.Glyphs))
	}

	// The first glyph (「) should have its X shifted left due to compression
	// Without compression: X=0. With compression: X = -halfEm = -8.0
	if run.Glyphs[0].X >= 0 {
		t.Errorf("expected negative X for line-start opening punct, got %f", run.Glyphs[0].X)
	}

	// Second glyph should be shifted left too
	noCompX := float32(16 * 0.6) // one glyph advance for mock engine
	if run.Glyphs[1].X >= noCompX {
		t.Errorf("expected glyph[1].X < %f (shifted by compression), got %f", noCompX, run.Glyphs[1].X)
	}
}

func TestShaperAdjacentPunctCompression(t *testing.T) {
	engine := newMockEngine()
	mgr := NewManager(engine)
	id, _ := mgr.Register("Test", WeightRegular, StyleNormal, nil)
	shaper := NewShaper(mgr)

	// 好」「世 — closing + opening pair should compress
	runs := shaper.Shape("好」「世", ShapeOptions{FontID: id, FontSize: 16})
	if len(runs) != 1 {
		t.Fatalf("expected 1 run, got %d", len(runs))
	}

	run := runs[0]
	// Without compression: each glyph at advance intervals
	// With compression: the gap between 」 and 「 should be reduced
	advance := float32(16 * 0.6)
	noCompWidth := 4 * advance // 4 glyphs without compression

	// Actual width should be less due to compression
	lastGlyph := run.Glyphs[len(run.Glyphs)-1]
	actualWidth := lastGlyph.X + lastGlyph.Advance
	if actualWidth >= noCompWidth {
		t.Errorf("expected compressed width < %f, got %f", noCompWidth, actualWidth)
	}
}
