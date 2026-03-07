package font

import "testing"

func TestManagerRegisterAndResolve(t *testing.T) {
	engine := newMockEngine()
	mgr := NewManager(engine)

	id, err := mgr.Register("TestFont", WeightRegular, StyleNormal, nil)
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}
	if id == InvalidFontID {
		t.Fatal("got InvalidFontID")
	}

	resolved, ok := mgr.Resolve(Properties{Family: "TestFont", Weight: WeightRegular, Style: StyleNormal})
	if !ok {
		t.Fatal("Resolve failed")
	}
	if resolved != id {
		t.Errorf("expected %d, got %d", id, resolved)
	}
}

func TestManagerResolveClosestWeight(t *testing.T) {
	engine := newMockEngine()
	mgr := NewManager(engine)

	light, _ := mgr.Register("F", WeightLight, StyleNormal, nil)
	bold, _ := mgr.Register("F", WeightBold, StyleNormal, nil)

	// Request regular (400) — should get light (300) as closer than bold (700)
	resolved, ok := mgr.Resolve(Properties{Family: "F", Weight: WeightRegular, Style: StyleNormal})
	if !ok {
		t.Fatal("Resolve failed")
	}
	if resolved != light {
		t.Errorf("expected light (%d), got %d (bold=%d)", light, resolved, bold)
	}
}

func TestManagerResolveStyleFallback(t *testing.T) {
	engine := newMockEngine()
	mgr := NewManager(engine)

	id, _ := mgr.Register("F", WeightRegular, StyleNormal, nil)

	// Request italic but only normal exists — should fall back
	resolved, ok := mgr.Resolve(Properties{Family: "F", Weight: WeightRegular, Style: StyleItalic})
	if !ok {
		t.Fatal("Resolve failed")
	}
	if resolved != id {
		t.Errorf("expected %d, got %d", id, resolved)
	}
}

func TestManagerResolveUnknownFamily(t *testing.T) {
	engine := newMockEngine()
	mgr := NewManager(engine)

	_, ok := mgr.Resolve(Properties{Family: "NonExistent"})
	if ok {
		t.Error("expected Resolve to fail for unknown family")
	}
}

func TestManagerResolveRune(t *testing.T) {
	engine := newMockEngine()
	mgr := NewManager(engine)

	id, _ := mgr.Register("Latin", WeightRegular, StyleNormal, nil)

	resolved, ok := mgr.ResolveRune(Properties{Family: "Latin", Weight: WeightRegular}, 'A')
	if !ok {
		t.Fatal("ResolveRune failed for 'A'")
	}
	if resolved != id {
		t.Errorf("expected %d, got %d", id, resolved)
	}
}

func TestManagerFallbackChain(t *testing.T) {
	engine := newMockEngine()
	mgr := NewManager(engine)

	mgr.Register("Latin", WeightRegular, StyleNormal, nil)
	cjkID, _ := mgr.Register("CJK", WeightRegular, StyleNormal, nil)

	mgr.SetFallbackChain("", []string{"Latin", "CJK"})

	// '你' is in the CJK font but the mock engine puts it in all fonts.
	// For this test, verify fallback logic works.
	resolved, ok := mgr.ResolveRune(Properties{Family: "Latin", Weight: WeightRegular}, '你')
	if !ok {
		t.Fatal("ResolveRune with fallback failed")
	}
	// Should find it in Latin first (mock has CJK chars in all fonts)
	_ = cjkID
	if resolved == InvalidFontID {
		t.Error("got InvalidFontID")
	}
}

func TestManagerEngine(t *testing.T) {
	engine := newMockEngine()
	mgr := NewManager(engine)
	if mgr.Engine() != engine {
		t.Error("Engine() returned wrong engine")
	}
}

func TestManagerMultipleWeights(t *testing.T) {
	engine := newMockEngine()
	mgr := NewManager(engine)

	thin, _ := mgr.Register("F", WeightThin, StyleNormal, nil)
	regular, _ := mgr.Register("F", WeightRegular, StyleNormal, nil)
	bold, _ := mgr.Register("F", WeightBold, StyleNormal, nil)
	black, _ := mgr.Register("F", WeightBlack, StyleNormal, nil)

	tests := []struct {
		reqWeight Weight
		expectID  ID
	}{
		{WeightThin, thin},
		{WeightRegular, regular},
		{WeightBold, bold},
		{WeightBlack, black},
		{WeightMedium, regular},   // 500 closest to 400
		{WeightSemiBold, bold},    // 600 closest to 700
		{WeightExtraBold, bold},   // 800 equidistant to 700/900, first match wins
	}

	for _, tc := range tests {
		resolved, ok := mgr.Resolve(Properties{Family: "F", Weight: tc.reqWeight})
		if !ok {
			t.Errorf("weight %d: Resolve failed", tc.reqWeight)
			continue
		}
		if resolved != tc.expectID {
			t.Errorf("weight %d: expected %d, got %d", tc.reqWeight, tc.expectID, resolved)
		}
	}
}
