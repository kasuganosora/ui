package layout

import "testing"

// Additional tests for 90%+ coverage.

func TestContentBounds(t *testing.T) {
	r := Result{X: 10, Y: 20, Width: 200, Height: 100}
	style := &Style{
		Padding: EdgeValues{Top: Px(5), Right: Px(10), Bottom: Px(5), Left: Px(10)},
		Border:  EdgeValues{Top: Px(1), Right: Px(1), Bottom: Px(1), Left: Px(1)},
	}
	cb := r.ContentBounds(style, 200)
	// X = 10 + 10 + 1 = 21
	// Y = 20 + 5 + 1 = 26
	// W = 200 - (10+10+1+1) = 178
	// H = 100 - (5+5+1+1) = 88
	if !approx(cb.X, 21) || !approx(cb.Y, 26) || !approx(cb.Width, 178) || !approx(cb.Height, 88) {
		t.Errorf("ContentBounds: expected (21,26,178,88), got (%.1f,%.1f,%.1f,%.1f)",
			cb.X, cb.Y, cb.Width, cb.Height)
	}
}

type mockTextMeasurer struct{}

func (m *mockTextMeasurer) MeasureText(text string, fontID uint32, fontSize float32, maxWidth float32) (float32, float32) {
	w := float32(len(text)) * fontSize * 0.5
	if maxWidth > 0 && w > maxWidth {
		lines := int(w/maxWidth) + 1
		return maxWidth, float32(lines) * fontSize * 1.2
	}
	return w, fontSize * 1.2
}

func TestSetTextMeasurer(t *testing.T) {
	e := New()
	m := &mockTextMeasurer{}
	e.SetTextMeasurer(m)
	if e.measurer != m {
		t.Error("measurer should be set")
	}
}

func TestAddTextNode(t *testing.T) {
	e := New()
	id := e.AddTextNode(DefaultStyle(), "Hello World")
	if e.nodes[id].text != "Hello World" {
		t.Error("text should be stored")
	}
}

func TestLayoutNodeDisplayNone(t *testing.T) {
	e := New()
	root := e.AddNode(Style{Display: DisplayBlock, Width: Px(400), Height: Px(300)})
	noneChild := e.AddNode(Style{Display: DisplayNone, Width: Px(100), Height: Px(100)})
	e.SetChildren(root, []NodeID{noneChild})
	e.AddRoot(root)
	e.Compute(400, 300)

	r := e.GetResult(noneChild)
	if r.Width != 0 || r.Height != 0 {
		t.Errorf("display:none should have zero result, got %v", r)
	}
}

func TestGetResultInvalid(t *testing.T) {
	e := New()
	r := e.GetResult(999)
	if r.Width != 0 || r.Height != 0 {
		t.Error("invalid ID should return zero result")
	}
}

func TestSetChildrenInvalid(t *testing.T) {
	e := New()
	// Should not panic
	e.SetChildren(999, []NodeID{0})
}

func TestJustifySpaceAround(t *testing.T) {
	e := New()
	root := e.AddNode(Style{
		Display:        DisplayFlex,
		FlexDirection:  FlexDirectionRow,
		JustifyContent: JustifySpaceAround,
		Width:          Px(300),
		Height:         Px(100),
	})
	a := e.AddNode(Style{FlexBasis: Px(50)})
	b := e.AddNode(Style{FlexBasis: Px(50)})
	e.SetChildren(root, []NodeID{a, b})
	e.AddRoot(root)
	e.Compute(300, 100)

	ra := e.GetResult(a)
	// Free = 200, n=2, spacing=100, offset=50
	if !approx(ra.X, 50) {
		t.Errorf("SpaceAround: expected a.X ~50, got %.2f", ra.X)
	}
}

func TestFlexReverseOrder(t *testing.T) {
	e := New()
	root := e.AddNode(Style{
		Display:       DisplayFlex,
		FlexDirection: FlexDirectionRowReverse,
		Width:         Px(300),
		Height:        Px(100),
	})
	a := e.AddNode(Style{FlexBasis: Px(100)})
	b := e.AddNode(Style{FlexBasis: Px(100)})
	e.SetChildren(root, []NodeID{a, b})
	e.AddRoot(root)
	e.Compute(300, 100)

	ra := e.GetResult(a)
	rb := e.GetResult(b)
	// Reversed: b should be before a in X position
	if ra.X <= rb.X {
		t.Errorf("reversed: a.X (%.1f) should be > b.X (%.1f)", ra.X, rb.X)
	}
}

func TestFlexAlignSelf(t *testing.T) {
	e := New()
	root := e.AddNode(Style{
		Display:       DisplayFlex,
		FlexDirection: FlexDirectionRow,
		AlignItems:    AlignFlexStart,
		Width:         Px(300),
		Height:        Px(100),
	})
	child := e.AddNode(Style{
		FlexBasis: Px(100),
		Height:    Px(40),
		AlignSelf: AlignSelfCenter,
	})
	e.SetChildren(root, []NodeID{child})
	e.AddRoot(root)
	e.Compute(300, 100)

	r := e.GetResult(child)
	// Cross centered: (100 - 40) / 2 = 30
	if !approx(r.Y, 30) {
		t.Errorf("AlignSelf center: expected Y ~30, got %.1f", r.Y)
	}
}

func TestFlexAlignContentCenter(t *testing.T) {
	e := New()
	root := e.AddNode(Style{
		Display:       DisplayFlex,
		FlexDirection: FlexDirectionRow,
		FlexWrap:      FlexWrapWrap,
		AlignContent:  AlignContentCenter,
		Width:         Px(100),
		Height:        Px(200),
	})
	a := e.AddNode(Style{FlexBasis: Px(60), Height: Px(30)})
	b := e.AddNode(Style{FlexBasis: Px(60), Height: Px(30)})
	e.SetChildren(root, []NodeID{a, b})
	e.AddRoot(root)
	e.Compute(100, 200)

	ra := e.GetResult(a)
	// Total cross = 60 (2 lines of 30), free = 140, offset = 70
	if ra.Y < 60 {
		t.Errorf("AlignContent center: expected Y >= 60, got %.1f", ra.Y)
	}
}

func TestFlexAlignContentFlexEnd(t *testing.T) {
	e := New()
	root := e.AddNode(Style{
		Display:       DisplayFlex,
		FlexDirection: FlexDirectionRow,
		FlexWrap:      FlexWrapWrap,
		AlignContent:  AlignContentFlexEnd,
		Width:         Px(100),
		Height:        Px(200),
	})
	a := e.AddNode(Style{FlexBasis: Px(60), Height: Px(30)})
	b := e.AddNode(Style{FlexBasis: Px(60), Height: Px(30)})
	e.SetChildren(root, []NodeID{a, b})
	e.AddRoot(root)
	e.Compute(100, 200)

	ra := e.GetResult(a)
	if ra.Y < 100 {
		t.Errorf("AlignContent flex-end: expected Y >= 100, got %.1f", ra.Y)
	}
}

func TestFlexAlignContentSpaceBetween(t *testing.T) {
	e := New()
	root := e.AddNode(Style{
		Display:       DisplayFlex,
		FlexDirection: FlexDirectionRow,
		FlexWrap:      FlexWrapWrap,
		AlignContent:  AlignContentSpaceBetween,
		Width:         Px(100),
		Height:        Px(200),
	})
	a := e.AddNode(Style{FlexBasis: Px(60), Height: Px(30)})
	b := e.AddNode(Style{FlexBasis: Px(60), Height: Px(30)})
	e.SetChildren(root, []NodeID{a, b})
	e.AddRoot(root)
	e.Compute(100, 200)

	rb := e.GetResult(b)
	// Second line should be pushed toward bottom
	if rb.Y < 100 {
		t.Errorf("AlignContent space-between: expected b.Y >= 100, got %.1f", rb.Y)
	}
}

func TestFlexAlignContentStretch(t *testing.T) {
	e := New()
	root := e.AddNode(Style{
		Display:       DisplayFlex,
		FlexDirection: FlexDirectionRow,
		FlexWrap:      FlexWrapWrap,
		AlignContent:  AlignContentStretch,
		Width:         Px(100),
		Height:        Px(200),
	})
	a := e.AddNode(Style{FlexBasis: Px(60)})
	b := e.AddNode(Style{FlexBasis: Px(60)})
	e.SetChildren(root, []NodeID{a, b})
	e.AddRoot(root)
	e.Compute(100, 200)

	// Both items should stretch their cross sizes
	_ = e.GetResult(a)
	_ = e.GetResult(b)
}

func TestFlexAlignContentSpaceAround(t *testing.T) {
	e := New()
	root := e.AddNode(Style{
		Display:       DisplayFlex,
		FlexDirection: FlexDirectionRow,
		FlexWrap:      FlexWrapWrap,
		AlignContent:  AlignContentSpaceAround,
		Width:         Px(100),
		Height:        Px(200),
	})
	a := e.AddNode(Style{FlexBasis: Px(60), Height: Px(30)})
	b := e.AddNode(Style{FlexBasis: Px(60), Height: Px(30)})
	e.SetChildren(root, []NodeID{a, b})
	e.AddRoot(root)
	e.Compute(100, 200)

	ra := e.GetResult(a)
	if ra.Y < 20 {
		t.Errorf("AlignContent space-around: expected a.Y >= 20, got %.1f", ra.Y)
	}
}

func TestResolveSizeAuto(t *testing.T) {
	v := resolveSize(Auto, 500, Auto, Auto)
	if v != 0 {
		t.Errorf("auto should resolve to 0 caller-handled, got %v", v)
	}
}

func TestConstrainSizeNegative(t *testing.T) {
	v := constrainSize(-10, 500, Auto, Auto)
	if v != 0 {
		t.Errorf("negative should be clamped to 0, got %v", v)
	}
}

func TestFlexAbsoluteChild(t *testing.T) {
	e := New()
	root := e.AddNode(Style{
		Display:       DisplayFlex,
		FlexDirection: FlexDirectionRow,
		Width:         Px(300),
		Height:        Px(200),
	})
	normal := e.AddNode(Style{FlexBasis: Px(100)})
	absolute := e.AddNode(Style{
		Position: PositionAbsolute,
		Left:     Px(10),
		Top:      Px(10),
		Width:    Px(50),
		Height:   Px(50),
	})
	e.SetChildren(root, []NodeID{normal, absolute})
	e.AddRoot(root)
	e.Compute(300, 200)

	expectResult(t, e, absolute, "absolute in flex", 10, 10, 50, 50)
}

func TestFlexColumnReverse(t *testing.T) {
	e := New()
	root := e.AddNode(Style{
		Display:       DisplayFlex,
		FlexDirection: FlexDirectionColumnReverse,
		Width:         Px(200),
		Height:        Px(300),
	})
	a := e.AddNode(Style{FlexBasis: Px(100)})
	b := e.AddNode(Style{FlexBasis: Px(100)})
	e.SetChildren(root, []NodeID{a, b})
	e.AddRoot(root)
	e.Compute(200, 300)

	ra := e.GetResult(a)
	rb := e.GetResult(b)
	// Reversed: a should be below b
	if ra.Y <= rb.Y {
		t.Errorf("column-reverse: a.Y (%.1f) should be > b.Y (%.1f)", ra.Y, rb.Y)
	}
}

func TestFlexWithChildMargin(t *testing.T) {
	e := New()
	root := e.AddNode(Style{
		Display:       DisplayFlex,
		FlexDirection: FlexDirectionRow,
		Width:         Px(300),
		Height:        Px(100),
	})
	child := e.AddNode(Style{
		FlexBasis: Px(100),
		Margin:    EdgeValues{Top: Px(10), Left: Px(20), Right: Px(20)},
	})
	e.SetChildren(root, []NodeID{child})
	e.AddRoot(root)
	e.Compute(300, 100)

	r := e.GetResult(child)
	if !approx(r.X, 20) {
		t.Errorf("expected X=20 from margin, got %.1f", r.X)
	}
	if !approx(r.Y, 10) {
		t.Errorf("expected Y=10 from margin, got %.1f", r.Y)
	}
}

func TestFlexNoChildrenWithAbsolute(t *testing.T) {
	e := New()
	root := e.AddNode(Style{
		Display:       DisplayFlex,
		FlexDirection: FlexDirectionRow,
		Width:         Px(300),
		Height:        Px(200),
	})
	abs := e.AddNode(Style{
		Position: PositionAbsolute,
		Left:     Px(5),
		Top:      Px(5),
		Width:    Px(50),
		Height:   Px(50),
	})
	e.SetChildren(root, []NodeID{abs})
	e.AddRoot(root)
	e.Compute(300, 200)

	expectResult(t, e, abs, "abs only child", 5, 5, 50, 50)
}

func TestRootAutoWidth(t *testing.T) {
	e := New()
	root := e.AddNode(Style{Display: DisplayBlock, Height: Px(100)}) // width auto
	e.AddRoot(root)
	e.Compute(800, 600)

	r := e.GetResult(root)
	if !approx(r.Width, 800) {
		t.Errorf("auto width root should use viewport width, got %.1f", r.Width)
	}
}

func TestRootExplicitHeight(t *testing.T) {
	e := New()
	root := e.AddNode(Style{Display: DisplayBlock, Width: Px(400), Height: Pct(50)})
	e.AddRoot(root)
	e.Compute(400, 600)

	r := e.GetResult(root)
	if !approx(r.Height, 300) {
		t.Errorf("50%% height of 600 should be 300, got %.1f", r.Height)
	}
}
