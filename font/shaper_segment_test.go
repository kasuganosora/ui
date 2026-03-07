package font

import "testing"

// mockSegmenter implements Segmenter for testing.
type mockSegmenter struct {
	words map[string][]string
}

func newMockSegmenter() *mockSegmenter {
	return &mockSegmenter{
		words: map[string][]string{
			"我来到北京":    {"我", "来到", "北京"},
			"今天天气不错":   {"今天", "天气", "不错"},
			"你好世界":     {"你好", "世界"},
			"我来到北京清华大学": {"我", "来到", "北京", "清华大学"},
		},
	}
}

func (m *mockSegmenter) Segment(text string) []string {
	if words, ok := m.words[text]; ok {
		return words
	}
	// Fallback: single-character segmentation
	var result []string
	for _, r := range text {
		result = append(result, string(r))
	}
	return result
}

func setupShaperWithSegmenter() (Shaper, ID, *mockSegmenter) {
	engine := newMockEngine()
	mgr := NewManager(engine)
	id, _ := mgr.Register("Test", WeightRegular, StyleNormal, nil)
	seg := newMockSegmenter()
	return NewShaper(mgr), id, seg
}

func TestShaperSegmenterBreaking(t *testing.T) {
	shaper, id, seg := setupShaperWithSegmenter()

	// Font size 16, advance per char = 9.6
	// "我来到北京" = 5 chars, total width = 48
	// MaxWidth 30 should force a wrap
	// Without segmenter: can break after any CJK char
	// With segmenter: should break at word boundaries (我/来到/北京)
	runs := shaper.Shape("我来到北京", ShapeOptions{
		FontID:    id,
		FontSize:  16,
		MaxWidth:  30,
		Segmenter: seg,
	})

	if len(runs) < 2 {
		t.Fatalf("expected at least 2 runs with word wrap, got %d", len(runs))
	}

	// First line should break at a word boundary
	firstLineGlyphs := len(runs[0].Glyphs)
	// With segmenter: "我" (1) + "来到" (2) = 3 chars = 28.8 width, fits in 30
	// "我来到北" would be 4 chars = 38.4, doesn't fit
	// So first line should be 3 glyphs (我来到)
	if firstLineGlyphs != 3 {
		t.Errorf("expected first line to have 3 glyphs (word boundary), got %d", firstLineGlyphs)
	}
}

func TestShaperTruncateChar(t *testing.T) {
	shaper, id, _ := setupShaperWithSegmenter()

	// "你好世界" = 4 chars, advance = 9.6 each, total = 38.4
	// MaxWidth 30, MaxLines 1, TruncateChar → truncate + ellipsis
	runs := shaper.Shape("你好世界", ShapeOptions{
		FontID:   id,
		FontSize: 16,
		MaxWidth: 30,
		MaxLines: 1,
		Truncate: TruncateChar,
	})

	if len(runs) != 1 {
		t.Fatalf("expected 1 run with truncation, got %d", len(runs))
	}

	// Should have some visible glyphs + ellipsis character(s)
	glyphs := runs[0].Glyphs
	if len(glyphs) == 0 {
		t.Fatal("expected non-empty glyphs")
	}

	// Last glyph should be the ellipsis '…'
	lastGlyph := glyphs[len(glyphs)-1]
	lastEnd := lastGlyph.X + lastGlyph.Advance
	if lastEnd > 30+0.1 { // small float tolerance
		t.Errorf("truncated line width %f exceeds MaxWidth 30", lastEnd)
	}
}

func TestShaperTruncateWord(t *testing.T) {
	shaper, id, seg := setupShaperWithSegmenter()

	// "你好世界" with segmenter → ["你好", "世界"]
	// MaxWidth 30, MaxLines 1, TruncateWord
	// "你好" = 19.2, + ellipsis "…" = 9.6, total = 28.8, fits in 30
	// "你好世" = 28.8, + ellipsis = 38.4, doesn't fit
	runs := shaper.Shape("你好世界", ShapeOptions{
		FontID:    id,
		FontSize:  16,
		MaxWidth:  30,
		MaxLines:  1,
		Truncate:  TruncateWord,
		Segmenter: seg,
	})

	if len(runs) != 1 {
		t.Fatalf("expected 1 run with word truncation, got %d", len(runs))
	}

	glyphs := runs[0].Glyphs
	// Should be "你好" (2) + "…" (1) = 3 glyphs
	if len(glyphs) != 3 {
		t.Errorf("expected 3 glyphs (你好…), got %d", len(glyphs))
	}
}

func TestShaperTruncateMultiLine(t *testing.T) {
	shaper, id, _ := setupShaperWithSegmenter()

	// "你好\n世界\n测试" = 3 lines
	// MaxLines 2 with truncation → only 2 lines, last line truncated if needed
	runs := shaper.Shape("你好\n世界\n测试", ShapeOptions{
		FontID:   id,
		FontSize: 16,
		MaxWidth: 100,
		MaxLines: 2,
		Truncate: TruncateChar,
	})

	if len(runs) > 2 {
		t.Errorf("expected at most 2 runs with MaxLines=2, got %d", len(runs))
	}
}

func TestShaperNoTruncation(t *testing.T) {
	shaper, id, _ := setupShaperWithSegmenter()

	// No truncation: text fits within MaxWidth
	runs := shaper.Shape("你好", ShapeOptions{
		FontID:   id,
		FontSize: 16,
		MaxWidth: 100,
		MaxLines: 1,
		Truncate: TruncateChar,
	})

	if len(runs) != 1 {
		t.Fatalf("expected 1 run, got %d", len(runs))
	}
	if len(runs[0].Glyphs) != 2 {
		t.Errorf("expected 2 glyphs (no truncation needed), got %d", len(runs[0].Glyphs))
	}
}

func TestShaperCustomEllipsis(t *testing.T) {
	shaper, id, _ := setupShaperWithSegmenter()

	runs := shaper.Shape("你好世界", ShapeOptions{
		FontID:   id,
		FontSize: 16,
		MaxWidth: 30,
		MaxLines: 1,
		Truncate: TruncateChar,
		Ellipsis: "..",
	})

	if len(runs) != 1 {
		t.Fatalf("expected 1 run, got %d", len(runs))
	}

	// Custom ellipsis ".." should be appended
	glyphs := runs[0].Glyphs
	if len(glyphs) < 2 {
		t.Fatal("expected at least 2 glyphs")
	}
}

func TestShaperSegmenterNilFallback(t *testing.T) {
	shaper, id, _ := setupShaperWithSegmenter()

	// Without segmenter, should still work (character-level CJK breaking)
	runs := shaper.Shape("你好世界", ShapeOptions{
		FontID:   id,
		FontSize: 16,
		MaxWidth: 25,
	})

	if len(runs) < 2 {
		t.Fatalf("expected at least 2 runs for wrap, got %d", len(runs))
	}
}
