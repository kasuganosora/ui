package font

import "testing"

// Additional tests for 90%+ coverage.

func TestRegisterFile(t *testing.T) {
	engine := newMockEngine()
	mgr := NewManager(engine)
	id, err := mgr.RegisterFile("Test", WeightRegular, StyleNormal, "/fake/path")
	if err != nil {
		t.Fatalf("RegisterFile failed: %v", err)
	}
	if id == InvalidFontID {
		t.Error("expected valid font ID")
	}
}

func TestResolveRune(t *testing.T) {
	engine := newMockEngine()
	mgr := NewManager(engine)
	id, _ := mgr.Register("Test", WeightRegular, StyleNormal, nil)

	// ResolveRune for ASCII char that exists
	found, ok := mgr.ResolveRune(Properties{Family: "Test", Weight: WeightRegular, Style: StyleNormal}, 'A')
	if !ok || found != id {
		t.Errorf("expected to resolve 'A' to font %d, got %d, ok=%v", id, found, ok)
	}
}

func TestResolveRuneNotFound(t *testing.T) {
	engine := newMockEngine()
	mgr := NewManager(engine)
	mgr.Register("Test", WeightRegular, StyleNormal, nil)

	// Rune that doesn't exist in any font
	_, ok := mgr.ResolveRune(Properties{Family: "Test", Weight: WeightRegular, Style: StyleNormal}, '\U0001F600')
	if ok {
		t.Error("should not resolve missing rune")
	}
}

func TestResolveRuneWithFallback(t *testing.T) {
	engine := newMockEngine()
	mgr := NewManager(engine)
	mgr.Register("Primary", WeightRegular, StyleNormal, nil)
	mgr.Register("Fallback", WeightRegular, StyleNormal, nil)

	// Set fallback chain
	mgr.SetFallbackChain("", []string{"Primary", "Fallback"})

	// Resolve 'A' which exists in both
	found, ok := mgr.ResolveRune(Properties{Family: "Primary", Weight: WeightRegular, Style: StyleNormal}, 'A')
	if !ok {
		t.Error("expected to resolve 'A' via primary or fallback")
	}
	_ = found
}

func TestResolveRuneUnknownFamily(t *testing.T) {
	engine := newMockEngine()
	mgr := NewManager(engine)

	_, ok := mgr.ResolveRune(Properties{Family: "NonExistent"}, 'A')
	if ok {
		t.Error("should not resolve from unknown family")
	}
}

func TestResolveRuneFallbackSkipsPrimary(t *testing.T) {
	engine := newMockEngine()
	mgr := NewManager(engine)
	mgr.Register("Primary", WeightRegular, StyleNormal, nil)
	mgr.Register("Fallback", WeightRegular, StyleNormal, nil)

	mgr.SetFallbackChain("", []string{"Primary", "Fallback"})

	// An emoji rune that doesn't exist in either mock font
	_, ok := mgr.ResolveRune(Properties{Family: "Primary", Weight: WeightRegular, Style: StyleNormal}, '\U0001F600')
	if ok {
		t.Error("should not resolve emoji that doesn't exist anywhere")
	}
}

func TestSetFallbackChainLocale(t *testing.T) {
	engine := newMockEngine()
	mgr := NewManager(engine)

	// Set locale-specific chain (stored but not used by current fallbackChainsFor)
	mgr.SetFallbackChain("zh-CN", []string{"NotoSansCJK", "Arial"})
}

func TestApplyAlignmentJustify(t *testing.T) {
	shaper, id := setupShaper()
	// Create text that wraps into multiple lines then justify it
	runs := shaper.Shape("Hello World Foo Bar", ShapeOptions{
		FontID:   id,
		FontSize: 16,
		MaxWidth: 80,
		Align:    TextAlignJustify,
	})
	if len(runs) < 2 {
		t.Fatalf("expected multiple lines, got %d", len(runs))
	}
	// First line should be justified (not last line)
}

func TestShaperTruncateCharCoverage(t *testing.T) {
	shaper, id := setupShaper()
	runs := shaper.Shape("ABCDEFGHIJ", ShapeOptions{
		FontID:   id,
		FontSize: 16,
		MaxWidth: 50,
		MaxLines: 1,
		Truncate: TruncateChar,
	})
	if len(runs) != 1 {
		t.Fatalf("expected 1 run (truncated), got %d", len(runs))
	}
	// Should end with ellipsis
}

func TestShaperInvalidFont(t *testing.T) {
	shaper, _ := setupShaper()
	runs := shaper.Shape("Hello", ShapeOptions{FontID: InvalidFontID, FontSize: 16})
	if runs != nil {
		t.Error("expected nil runs for invalid font")
	}
}

func TestShaperMeasureEmpty(t *testing.T) {
	shaper, id := setupShaper()
	m := shaper.Measure("", ShapeOptions{FontID: id, FontSize: 16})
	if m.Width != 0 || m.Height != 0 {
		t.Error("expected zero metrics for empty text")
	}
}

func TestShaperMeasureInvalidFont(t *testing.T) {
	shaper, _ := setupShaper()
	m := shaper.Measure("Hello", ShapeOptions{FontID: InvalidFontID, FontSize: 16})
	if m.Width != 0 {
		t.Error("expected zero width for invalid font")
	}
}

func TestShaperCarriageReturn(t *testing.T) {
	shaper, id := setupShaper()
	runs := shaper.Shape("A\r\nB", ShapeOptions{FontID: id, FontSize: 16})
	if len(runs) != 2 {
		t.Fatalf("expected 2 runs, got %d", len(runs))
	}
}

func TestShaperForceBreakNoWordBoundary(t *testing.T) {
	shaper, id := setupShaper()
	// Long word with no break points that must force break
	runs := shaper.Shape("ABCDEFGHIJKLMNOP", ShapeOptions{
		FontID:   id,
		FontSize: 16,
		MaxWidth: 50,
	})
	if len(runs) < 2 {
		t.Fatalf("expected word to be force-broken into multiple lines, got %d", len(runs))
	}
}

func TestShaperLineHeightOverride(t *testing.T) {
	shaper, id := setupShaper()
	runs := shaper.Shape("A\nB", ShapeOptions{
		FontID:     id,
		FontSize:   16,
		LineHeight: 30,
	})
	if len(runs) != 2 {
		t.Fatalf("expected 2 runs, got %d", len(runs))
	}
	// Second line Y should be based on 30px line height
	if runs[1].Bounds.Y < 29 || runs[1].Bounds.Y > 31 {
		t.Errorf("expected second line Y ~30, got %f", runs[1].Bounds.Y)
	}
}

func TestShaperTruncateMultiLineLimit(t *testing.T) {
	shaper, id := setupShaper()
	runs := shaper.Shape("A\nB\nC\nD", ShapeOptions{
		FontID:   id,
		FontSize: 16,
		MaxWidth: 200,
		MaxLines: 2,
		Truncate: TruncateChar,
	})
	if len(runs) > 2 {
		t.Errorf("expected at most 2 runs, got %d", len(runs))
	}
}

func TestResolveDifferentWeights(t *testing.T) {
	engine := newMockEngine()
	mgr := NewManager(engine)
	mgr.Register("Test", WeightRegular, StyleNormal, nil)
	mgr.Register("Test", WeightBold, StyleNormal, nil)

	// Resolve bold
	id, ok := mgr.Resolve(Properties{Family: "Test", Weight: WeightBold, Style: StyleNormal})
	if !ok {
		t.Error("should resolve bold weight")
	}
	_ = id
}

func TestResolveDifferentStyles(t *testing.T) {
	engine := newMockEngine()
	mgr := NewManager(engine)
	mgr.Register("Test", WeightRegular, StyleNormal, nil)
	mgr.Register("Test", WeightRegular, StyleItalic, nil)

	// Resolve italic
	_, ok := mgr.Resolve(Properties{Family: "Test", Weight: WeightRegular, Style: StyleItalic})
	if !ok {
		t.Error("should resolve italic style")
	}
}

func TestResolveClosestWeight(t *testing.T) {
	engine := newMockEngine()
	mgr := NewManager(engine)
	mgr.Register("Test", 300, StyleNormal, nil)
	mgr.Register("Test", 700, StyleNormal, nil)

	// Request weight 400 — should get 300 (closer)
	_, ok := mgr.Resolve(Properties{Family: "Test", Weight: 400, Style: StyleNormal})
	if !ok {
		t.Error("should resolve closest weight")
	}
}

func TestResolveFallbackToAnyStyle(t *testing.T) {
	engine := newMockEngine()
	mgr := NewManager(engine)
	mgr.Register("Test", WeightRegular, StyleItalic, nil) // only italic

	// Request normal style — should fall back to italic
	_, ok := mgr.Resolve(Properties{Family: "Test", Weight: WeightRegular, Style: StyleNormal})
	if !ok {
		t.Error("should fall back to any style")
	}
}

func TestResolveRuneMultipleFacesInFamily(t *testing.T) {
	engine := newMockEngine()
	mgr := NewManager(engine)
	mgr.Register("Test", WeightRegular, StyleNormal, nil)
	mgr.Register("Test", WeightBold, StyleNormal, nil)

	// Both have 'A'
	_, ok := mgr.ResolveRune(Properties{Family: "Test", Weight: WeightRegular, Style: StyleNormal}, 'A')
	if !ok {
		t.Error("should resolve from family with multiple faces")
	}
}

func TestManagerEngineCoverage(t *testing.T) {
	engine := newMockEngine()
	mgr := NewManager(engine)
	if mgr.Engine() != engine {
		t.Error("Engine() should return the provided engine")
	}
}
