package layout

import (
	"math"
	"testing"
)

const epsilon = 0.01

func approx(a, b float32) bool {
	return float32(math.Abs(float64(a-b))) < epsilon
}

func expectResult(t *testing.T, e *Engine, id NodeID, name string, x, y, w, h float32) {
	t.Helper()
	r := e.GetResult(id)
	if !approx(r.X, x) || !approx(r.Y, y) || !approx(r.Width, w) || !approx(r.Height, h) {
		t.Errorf("%s: expected (%.1f, %.1f, %.1f, %.1f), got (%.1f, %.1f, %.1f, %.1f)",
			name, x, y, w, h, r.X, r.Y, r.Width, r.Height)
	}
}

// === Value tests ===

func TestValuePx(t *testing.T) {
	v := Px(100)
	resolved, ok := v.Resolve(500)
	if !ok || resolved != 100 {
		t.Errorf("Px(100) should resolve to 100, got %v", resolved)
	}
}

func TestValuePercent(t *testing.T) {
	v := Pct(50)
	resolved, ok := v.Resolve(400)
	if !ok || resolved != 200 {
		t.Errorf("Pct(50) of 400 should be 200, got %v", resolved)
	}
}

func TestValueAuto(t *testing.T) {
	v := Auto
	_, ok := v.Resolve(500)
	if ok {
		t.Error("Auto should not resolve")
	}
	if !v.IsAuto() {
		t.Error("Auto.IsAuto() should be true")
	}
}

func TestValueZero(t *testing.T) {
	v := Zero
	resolved, ok := v.Resolve(500)
	if !ok || resolved != 0 {
		t.Errorf("Zero should resolve to 0, got %v", resolved)
	}
}

// === Style tests ===

func TestDefaultStyle(t *testing.T) {
	s := DefaultStyle()
	if s.Display != DisplayBlock {
		t.Error("default display should be block")
	}
	if s.FlexShrink != 1 {
		t.Error("default flex-shrink should be 1")
	}
	if s.AlignItems != AlignStretch {
		t.Error("default align-items should be stretch")
	}
}

func TestStyleIsRow(t *testing.T) {
	s := Style{FlexDirection: FlexDirectionRow}
	if !s.IsRow() {
		t.Error("Row should be row")
	}
	s.FlexDirection = FlexDirectionColumn
	if s.IsRow() {
		t.Error("Column should not be row")
	}
}

func TestStyleIsReverse(t *testing.T) {
	s := Style{FlexDirection: FlexDirectionRowReverse}
	if !s.IsReverse() {
		t.Error("RowReverse should be reverse")
	}
	s.FlexDirection = FlexDirectionRow
	if s.IsReverse() {
		t.Error("Row should not be reverse")
	}
}

func TestStyleGaps(t *testing.T) {
	s := Style{FlexDirection: FlexDirectionRow, Gap: 10, RowGap: 20, ColumnGap: 30}
	if s.MainGap() != 30 {
		t.Errorf("row main gap should be column gap 30, got %v", s.MainGap())
	}
	if s.CrossGap() != 20 {
		t.Errorf("row cross gap should be row gap 20, got %v", s.CrossGap())
	}

	s.FlexDirection = FlexDirectionColumn
	if s.MainGap() != 20 {
		t.Errorf("column main gap should be row gap 20, got %v", s.MainGap())
	}
	if s.CrossGap() != 30 {
		t.Errorf("column cross gap should be column gap 30, got %v", s.CrossGap())
	}
}

func TestStyleGapFallback(t *testing.T) {
	s := Style{FlexDirection: FlexDirectionRow, Gap: 10}
	if s.MainGap() != 10 {
		t.Errorf("should fallback to Gap, got %v", s.MainGap())
	}
	if s.CrossGap() != 10 {
		t.Errorf("should fallback to Gap, got %v", s.CrossGap())
	}
}

// === Resolve tests ===

func TestResolveSize(t *testing.T) {
	v := resolveSize(Px(100), 500, Auto, Auto)
	if v != 100 {
		t.Errorf("expected 100, got %v", v)
	}
}

func TestResolveSizeWithMin(t *testing.T) {
	v := resolveSize(Px(50), 500, Px(100), Auto)
	if v != 100 {
		t.Errorf("expected min clamped to 100, got %v", v)
	}
}

func TestResolveSizeWithMax(t *testing.T) {
	v := resolveSize(Px(200), 500, Auto, Px(100))
	if v != 100 {
		t.Errorf("expected max clamped to 100, got %v", v)
	}
}

func TestResolveEdges(t *testing.T) {
	ev := EdgeValues{
		Top:    Px(10),
		Right:  Pct(10),
		Bottom: Px(20),
		Left:   Px(5),
	}
	top, right, bottom, left := resolveEdges(ev, 200)
	if top != 10 || right != 20 || bottom != 20 || left != 5 {
		t.Errorf("edges: expected (10, 20, 20, 5), got (%.0f, %.0f, %.0f, %.0f)",
			top, right, bottom, left)
	}
}

// === Engine basic tests ===

func TestEngineNewAndClear(t *testing.T) {
	e := New()
	if e.NodeCount() != 0 {
		t.Error("new engine should have 0 nodes")
	}
	e.AddNode(DefaultStyle())
	if e.NodeCount() != 1 {
		t.Error("should have 1 node")
	}
	e.Clear()
	if e.NodeCount() != 0 {
		t.Error("cleared engine should have 0 nodes")
	}
}

func TestResultBounds(t *testing.T) {
	r := Result{X: 10, Y: 20, Width: 100, Height: 50}
	b := r.Bounds()
	if b.X != 10 || b.Y != 20 || b.Width != 100 || b.Height != 50 {
		t.Error("Bounds mismatch")
	}
}

// === Block layout tests ===

func TestBlockSingleChild(t *testing.T) {
	e := New()
	root := e.AddNode(Style{Display: DisplayBlock, Width: Px(400), Height: Px(300)})
	child := e.AddNode(Style{Display: DisplayBlock, Height: Px(100)})
	e.SetChildren(root, []NodeID{child})
	e.AddRoot(root)
	e.Compute(400, 300)

	expectResult(t, e, root, "root", 0, 0, 400, 300)
	expectResult(t, e, child, "child", 0, 0, 400, 100)
}

func TestBlockMultipleChildren(t *testing.T) {
	e := New()
	root := e.AddNode(Style{Display: DisplayBlock, Width: Px(400), Height: Px(600)})
	a := e.AddNode(Style{Display: DisplayBlock, Height: Px(100)})
	b := e.AddNode(Style{Display: DisplayBlock, Height: Px(150)})
	c := e.AddNode(Style{Display: DisplayBlock, Height: Px(200)})
	e.SetChildren(root, []NodeID{a, b, c})
	e.AddRoot(root)
	e.Compute(400, 600)

	expectResult(t, e, a, "a", 0, 0, 400, 100)
	expectResult(t, e, b, "b", 0, 100, 400, 150)
	expectResult(t, e, c, "c", 0, 250, 400, 200)
}

func TestBlockWithPadding(t *testing.T) {
	e := New()
	root := e.AddNode(Style{
		Display: DisplayBlock,
		Width:   Px(400), Height: Px(300),
		Padding: EdgeValues{Top: Px(10), Right: Px(20), Bottom: Px(10), Left: Px(20)},
	})
	child := e.AddNode(Style{Display: DisplayBlock, Height: Px(100)})
	e.SetChildren(root, []NodeID{child})
	e.AddRoot(root)
	e.Compute(400, 300)

	// Child should be inset by padding
	expectResult(t, e, child, "child", 20, 10, 360, 100)
}

func TestBlockWithMargin(t *testing.T) {
	e := New()
	root := e.AddNode(Style{Display: DisplayBlock, Width: Px(400), Height: Px(300)})
	child := e.AddNode(Style{
		Display: DisplayBlock, Height: Px(100),
		Margin: EdgeValues{Top: Px(10), Left: Px(20), Right: Px(20)},
	})
	e.SetChildren(root, []NodeID{child})
	e.AddRoot(root)
	e.Compute(400, 300)

	expectResult(t, e, child, "child", 20, 10, 360, 100)
}

func TestBlockAutoHeight(t *testing.T) {
	e := New()
	root := e.AddNode(Style{Display: DisplayBlock, Width: Px(400)}) // height = auto
	a := e.AddNode(Style{Display: DisplayBlock, Height: Px(100)})
	b := e.AddNode(Style{Display: DisplayBlock, Height: Px(150)})
	e.SetChildren(root, []NodeID{a, b})
	e.AddRoot(root)
	e.Compute(400, 600)

	r := e.GetResult(root)
	if !approx(r.Height, 250) {
		t.Errorf("auto height: expected 250, got %.1f", r.Height)
	}
}

func TestBlockAutoHeightWithPadding(t *testing.T) {
	e := New()
	root := e.AddNode(Style{
		Display: DisplayBlock, Width: Px(400),
		Padding: EdgeValues{Top: Px(10), Bottom: Px(10)},
	})
	child := e.AddNode(Style{Display: DisplayBlock, Height: Px(100)})
	e.SetChildren(root, []NodeID{child})
	e.AddRoot(root)
	e.Compute(400, 600)

	r := e.GetResult(root)
	if !approx(r.Height, 120) {
		t.Errorf("auto height with padding: expected 120, got %.1f", r.Height)
	}
}

func TestBlockPercentWidth(t *testing.T) {
	e := New()
	root := e.AddNode(Style{Display: DisplayBlock, Width: Px(400), Height: Px(300)})
	child := e.AddNode(Style{Display: DisplayBlock, Width: Pct(50), Height: Px(100)})
	e.SetChildren(root, []NodeID{child})
	e.AddRoot(root)
	e.Compute(400, 300)

	expectResult(t, e, child, "child", 0, 0, 200, 100)
}

func TestBlockDisplayNone(t *testing.T) {
	e := New()
	root := e.AddNode(Style{Display: DisplayBlock, Width: Px(400), Height: Px(300)})
	a := e.AddNode(Style{Display: DisplayBlock, Height: Px(100)})
	hidden := e.AddNode(Style{Display: DisplayNone, Height: Px(100)})
	b := e.AddNode(Style{Display: DisplayBlock, Height: Px(100)})
	e.SetChildren(root, []NodeID{a, hidden, b})
	e.AddRoot(root)
	e.Compute(400, 300)

	expectResult(t, e, a, "a", 0, 0, 400, 100)
	// b should be right after a, not after hidden
	expectResult(t, e, b, "b", 0, 100, 400, 100)
}

// === Flexbox layout tests ===

func TestFlexRowEqualChildren(t *testing.T) {
	e := New()
	root := e.AddNode(Style{
		Display: DisplayFlex, FlexDirection: FlexDirectionRow,
		Width: Px(300), Height: Px(100),
	})
	a := e.AddNode(Style{FlexBasis: Px(100)})
	b := e.AddNode(Style{FlexBasis: Px(100)})
	c := e.AddNode(Style{FlexBasis: Px(100)})
	e.SetChildren(root, []NodeID{a, b, c})
	e.AddRoot(root)
	e.Compute(300, 100)

	expectResult(t, e, a, "a", 0, 0, 100, 100)
	expectResult(t, e, b, "b", 100, 0, 100, 100)
	expectResult(t, e, c, "c", 200, 0, 100, 100)
}

func TestFlexRowGrow(t *testing.T) {
	e := New()
	root := e.AddNode(Style{
		Display: DisplayFlex, FlexDirection: FlexDirectionRow,
		Width: Px(300), Height: Px(100),
	})
	a := e.AddNode(Style{FlexBasis: Px(50), FlexGrow: 1})
	b := e.AddNode(Style{FlexBasis: Px(50), FlexGrow: 2})
	e.SetChildren(root, []NodeID{a, b})
	e.AddRoot(root)
	e.Compute(300, 100)

	// Free space = 300 - 100 = 200
	// a gets 200 * 1/3 ≈ 66.67, total = 116.67
	// b gets 200 * 2/3 ≈ 133.33, total = 183.33
	ra := e.GetResult(a)
	rb := e.GetResult(b)
	if !approx(ra.Width, 116.67) {
		t.Errorf("a width: expected ~116.67, got %.2f", ra.Width)
	}
	if !approx(rb.Width, 183.33) {
		t.Errorf("b width: expected ~183.33, got %.2f", rb.Width)
	}
}

func TestFlexRowShrink(t *testing.T) {
	e := New()
	root := e.AddNode(Style{
		Display: DisplayFlex, FlexDirection: FlexDirectionRow,
		Width: Px(200), Height: Px(100),
	})
	a := e.AddNode(Style{FlexBasis: Px(150), FlexShrink: 1})
	b := e.AddNode(Style{FlexBasis: Px(150), FlexShrink: 1})
	e.SetChildren(root, []NodeID{a, b})
	e.AddRoot(root)
	e.Compute(200, 100)

	// Overflow = 300 - 200 = 100
	// Each shrinks by 50
	expectResult(t, e, a, "a", 0, 0, 100, 100)
	expectResult(t, e, b, "b", 100, 0, 100, 100)
}

func TestFlexColumn(t *testing.T) {
	e := New()
	root := e.AddNode(Style{
		Display: DisplayFlex, FlexDirection: FlexDirectionColumn,
		Width: Px(200), Height: Px(300),
	})
	a := e.AddNode(Style{FlexBasis: Px(100)})
	b := e.AddNode(Style{FlexBasis: Px(100)})
	e.SetChildren(root, []NodeID{a, b})
	e.AddRoot(root)
	e.Compute(200, 300)

	expectResult(t, e, a, "a", 0, 0, 200, 100)
	expectResult(t, e, b, "b", 0, 100, 200, 100)
}

func TestFlexColumnGrow(t *testing.T) {
	e := New()
	root := e.AddNode(Style{
		Display: DisplayFlex, FlexDirection: FlexDirectionColumn,
		Width: Px(200), Height: Px(300),
	})
	a := e.AddNode(Style{FlexBasis: Px(50), FlexGrow: 1})
	b := e.AddNode(Style{FlexBasis: Px(50), FlexGrow: 1})
	e.SetChildren(root, []NodeID{a, b})
	e.AddRoot(root)
	e.Compute(200, 300)

	// Free space = 300 - 100 = 200, each grows by 100
	expectResult(t, e, a, "a", 0, 0, 200, 150)
	expectResult(t, e, b, "b", 0, 150, 200, 150)
}

func TestFlexJustifyCenter(t *testing.T) {
	e := New()
	root := e.AddNode(Style{
		Display: DisplayFlex, FlexDirection: FlexDirectionRow,
		JustifyContent: JustifyCenter,
		Width: Px(300), Height: Px(100),
	})
	a := e.AddNode(Style{FlexBasis: Px(50)})
	b := e.AddNode(Style{FlexBasis: Px(50)})
	e.SetChildren(root, []NodeID{a, b})
	e.AddRoot(root)
	e.Compute(300, 100)

	// Free space = 300 - 100 = 200, offset = 100
	expectResult(t, e, a, "a", 100, 0, 50, 100)
	expectResult(t, e, b, "b", 150, 0, 50, 100)
}

func TestFlexJustifySpaceBetween(t *testing.T) {
	e := New()
	root := e.AddNode(Style{
		Display: DisplayFlex, FlexDirection: FlexDirectionRow,
		JustifyContent: JustifySpaceBetween,
		Width: Px(300), Height: Px(100),
	})
	a := e.AddNode(Style{FlexBasis: Px(50)})
	b := e.AddNode(Style{FlexBasis: Px(50)})
	c := e.AddNode(Style{FlexBasis: Px(50)})
	e.SetChildren(root, []NodeID{a, b, c})
	e.AddRoot(root)
	e.Compute(300, 100)

	// Free space = 300 - 150 = 150, spacing = 150/2 = 75
	expectResult(t, e, a, "a", 0, 0, 50, 100)
	expectResult(t, e, b, "b", 125, 0, 50, 100)
	expectResult(t, e, c, "c", 250, 0, 50, 100)
}

func TestFlexJustifyFlexEnd(t *testing.T) {
	e := New()
	root := e.AddNode(Style{
		Display: DisplayFlex, FlexDirection: FlexDirectionRow,
		JustifyContent: JustifyFlexEnd,
		Width: Px(300), Height: Px(100),
	})
	a := e.AddNode(Style{FlexBasis: Px(100)})
	e.SetChildren(root, []NodeID{a})
	e.AddRoot(root)
	e.Compute(300, 100)

	expectResult(t, e, a, "a", 200, 0, 100, 100)
}

func TestFlexJustifySpaceEvenly(t *testing.T) {
	e := New()
	root := e.AddNode(Style{
		Display: DisplayFlex, FlexDirection: FlexDirectionRow,
		JustifyContent: JustifySpaceEvenly,
		Width: Px(300), Height: Px(100),
	})
	a := e.AddNode(Style{FlexBasis: Px(50)})
	b := e.AddNode(Style{FlexBasis: Px(50)})
	e.SetChildren(root, []NodeID{a, b})
	e.AddRoot(root)
	e.Compute(300, 100)

	// Free space = 200, slots = 3, spacing = 200/3 ≈ 66.67
	ra := e.GetResult(a)
	if !approx(ra.X, 66.67) {
		t.Errorf("a.X: expected ~66.67, got %.2f", ra.X)
	}
}

func TestFlexAlignCenter(t *testing.T) {
	e := New()
	root := e.AddNode(Style{
		Display: DisplayFlex, FlexDirection: FlexDirectionRow,
		AlignItems: AlignCenter,
		Width: Px(300), Height: Px(100),
	})
	a := e.AddNode(Style{FlexBasis: Px(100), Height: Px(40)})
	e.SetChildren(root, []NodeID{a})
	e.AddRoot(root)
	e.Compute(300, 100)

	// Cross centered: (100 - 40) / 2 = 30
	expectResult(t, e, a, "a", 0, 30, 100, 40)
}

func TestFlexAlignFlexEnd(t *testing.T) {
	e := New()
	root := e.AddNode(Style{
		Display: DisplayFlex, FlexDirection: FlexDirectionRow,
		AlignItems: AlignFlexEnd,
		Width: Px(300), Height: Px(100),
	})
	a := e.AddNode(Style{FlexBasis: Px(100), Height: Px(40)})
	e.SetChildren(root, []NodeID{a})
	e.AddRoot(root)
	e.Compute(300, 100)

	expectResult(t, e, a, "a", 0, 60, 100, 40)
}

func TestFlexAlignStretch(t *testing.T) {
	e := New()
	root := e.AddNode(Style{
		Display: DisplayFlex, FlexDirection: FlexDirectionRow,
		AlignItems: AlignStretch,
		Width: Px(300), Height: Px(100),
	})
	a := e.AddNode(Style{FlexBasis: Px(100)}) // no explicit height
	e.SetChildren(root, []NodeID{a})
	e.AddRoot(root)
	e.Compute(300, 100)

	// Stretch should fill cross axis
	expectResult(t, e, a, "a", 0, 0, 100, 100)
}

func TestFlexGap(t *testing.T) {
	e := New()
	root := e.AddNode(Style{
		Display: DisplayFlex, FlexDirection: FlexDirectionRow,
		Width: Px(300), Height: Px(100),
		Gap: 10,
	})
	a := e.AddNode(Style{FlexBasis: Px(90)})
	b := e.AddNode(Style{FlexBasis: Px(90)})
	c := e.AddNode(Style{FlexBasis: Px(90)})
	e.SetChildren(root, []NodeID{a, b, c})
	e.AddRoot(root)
	e.Compute(300, 100)

	expectResult(t, e, a, "a", 0, 0, 90, 100)
	expectResult(t, e, b, "b", 100, 0, 90, 100)
	expectResult(t, e, c, "c", 200, 0, 90, 100)
}

func TestFlexWithPadding(t *testing.T) {
	e := New()
	root := e.AddNode(Style{
		Display: DisplayFlex, FlexDirection: FlexDirectionRow,
		Width: Px(300), Height: Px(100),
		Padding: EdgeValues{Top: Px(10), Right: Px(10), Bottom: Px(10), Left: Px(10)},
	})
	a := e.AddNode(Style{FlexBasis: Px(100)})
	e.SetChildren(root, []NodeID{a})
	e.AddRoot(root)
	e.Compute(300, 100)

	// Content area is 280x80, child starts at (10, 10)
	expectResult(t, e, a, "a", 10, 10, 100, 80)
}

func TestFlexWrap(t *testing.T) {
	e := New()
	root := e.AddNode(Style{
		Display: DisplayFlex, FlexDirection: FlexDirectionRow,
		FlexWrap: FlexWrapWrap,
		Width: Px(200), Height: Px(200),
	})
	a := e.AddNode(Style{FlexBasis: Px(120)})
	b := e.AddNode(Style{FlexBasis: Px(120)})
	e.SetChildren(root, []NodeID{a, b})
	e.AddRoot(root)
	e.Compute(200, 200)

	ra := e.GetResult(a)
	rb := e.GetResult(b)
	// a on first line, b wraps to second line
	if ra.Y != 0 {
		t.Errorf("a should be on first line, Y=%.1f", ra.Y)
	}
	if rb.Y == 0 {
		t.Errorf("b should wrap to second line, Y=%.1f", rb.Y)
	}
}

func TestFlexNoChildren(t *testing.T) {
	e := New()
	root := e.AddNode(Style{
		Display: DisplayFlex, FlexDirection: FlexDirectionRow,
		Width: Px(300), Height: Px(100),
	})
	e.SetChildren(root, []NodeID{})
	e.AddRoot(root)
	e.Compute(300, 100)

	expectResult(t, e, root, "root", 0, 0, 300, 100)
}

func TestFlexDisplayNone(t *testing.T) {
	e := New()
	root := e.AddNode(Style{
		Display: DisplayFlex, FlexDirection: FlexDirectionRow,
		Width: Px(300), Height: Px(100),
	})
	a := e.AddNode(Style{FlexBasis: Px(100)})
	hidden := e.AddNode(Style{Display: DisplayNone, FlexBasis: Px(100)})
	b := e.AddNode(Style{FlexBasis: Px(100)})
	e.SetChildren(root, []NodeID{a, hidden, b})
	e.AddRoot(root)
	e.Compute(300, 100)

	expectResult(t, e, a, "a", 0, 0, 100, 100)
	expectResult(t, e, b, "b", 100, 0, 100, 100)
}

// === Absolute positioning tests ===

func TestAbsolutePosition(t *testing.T) {
	e := New()
	root := e.AddNode(Style{Display: DisplayBlock, Width: Px(400), Height: Px(300)})
	child := e.AddNode(Style{
		Display:  DisplayBlock,
		Position: PositionAbsolute,
		Left:     Px(50), Top: Px(30),
		Width: Px(100), Height: Px(80),
	})
	e.SetChildren(root, []NodeID{child})
	e.AddRoot(root)
	e.Compute(400, 300)

	expectResult(t, e, child, "absolute child", 50, 30, 100, 80)
}

func TestAbsoluteInferWidth(t *testing.T) {
	e := New()
	root := e.AddNode(Style{Display: DisplayBlock, Width: Px(400), Height: Px(300)})
	child := e.AddNode(Style{
		Display:  DisplayBlock,
		Position: PositionAbsolute,
		Left: Px(50), Right: Px(50),
		Top: Px(0), Height: Px(100),
	})
	e.SetChildren(root, []NodeID{child})
	e.AddRoot(root)
	e.Compute(400, 300)

	// Width = 400 - 50 - 50 = 300
	expectResult(t, e, child, "infer width", 50, 0, 300, 100)
}

func TestAbsoluteInferHeight(t *testing.T) {
	e := New()
	root := e.AddNode(Style{Display: DisplayBlock, Width: Px(400), Height: Px(300)})
	child := e.AddNode(Style{
		Display:  DisplayBlock,
		Position: PositionAbsolute,
		Top: Px(20), Bottom: Px(30),
		Left: Px(0), Width: Px(100),
	})
	e.SetChildren(root, []NodeID{child})
	e.AddRoot(root)
	e.Compute(400, 300)

	// Height = 300 - 20 - 30 = 250
	expectResult(t, e, child, "infer height", 0, 20, 100, 250)
}

// === Nested layout tests ===

func TestNestedFlexInBlock(t *testing.T) {
	e := New()
	root := e.AddNode(Style{Display: DisplayBlock, Width: Px(400), Height: Px(400)})
	flexRow := e.AddNode(Style{
		Display: DisplayFlex, FlexDirection: FlexDirectionRow,
		Height: Px(100),
	})
	a := e.AddNode(Style{FlexBasis: Px(100), FlexGrow: 1})
	b := e.AddNode(Style{FlexBasis: Px(100), FlexGrow: 1})
	e.SetChildren(flexRow, []NodeID{a, b})
	e.SetChildren(root, []NodeID{flexRow})
	e.AddRoot(root)
	e.Compute(400, 400)

	expectResult(t, e, flexRow, "flex row", 0, 0, 400, 100)
	expectResult(t, e, a, "a", 0, 0, 200, 100)
	expectResult(t, e, b, "b", 200, 0, 200, 100)
}

func TestNestedBlockInFlex(t *testing.T) {
	e := New()
	root := e.AddNode(Style{
		Display: DisplayFlex, FlexDirection: FlexDirectionRow,
		Width: Px(400), Height: Px(300),
	})
	sidebar := e.AddNode(Style{Display: DisplayBlock, FlexBasis: Px(100)})
	sItem := e.AddNode(Style{Display: DisplayBlock, Height: Px(50)})
	e.SetChildren(sidebar, []NodeID{sItem})

	content := e.AddNode(Style{Display: DisplayBlock, FlexGrow: 1})
	e.SetChildren(root, []NodeID{sidebar, content})
	e.AddRoot(root)
	e.Compute(400, 300)

	expectResult(t, e, sidebar, "sidebar", 0, 0, 100, 300)
	expectResult(t, e, sItem, "sItem", 0, 0, 100, 50)
	expectResult(t, e, content, "content", 100, 0, 300, 300)
}

// === MinWidth / MaxWidth constraint tests ===

func TestBlockMinWidth(t *testing.T) {
	e := New()
	root := e.AddNode(Style{Display: DisplayBlock, Width: Px(400), Height: Px(200)})
	child := e.AddNode(Style{
		Display:  DisplayBlock,
		Width:    Px(50),
		MinWidth: Px(100),
		Height:   Px(50),
	})
	e.SetChildren(root, []NodeID{child})
	e.AddRoot(root)
	e.Compute(400, 200)

	r := e.GetResult(child)
	if !approx(r.Width, 100) {
		t.Errorf("min-width: expected 100, got %.1f", r.Width)
	}
}

func TestBlockMaxWidth(t *testing.T) {
	e := New()
	root := e.AddNode(Style{Display: DisplayBlock, Width: Px(400), Height: Px(200)})
	child := e.AddNode(Style{
		Display:  DisplayBlock,
		Width:    Px(300),
		MaxWidth: Px(200),
		Height:   Px(50),
	})
	e.SetChildren(root, []NodeID{child})
	e.AddRoot(root)
	e.Compute(400, 200)

	r := e.GetResult(child)
	if !approx(r.Width, 200) {
		t.Errorf("max-width: expected 200, got %.1f", r.Width)
	}
}

// === Reuse / multiple computes ===

func TestEngineReuse(t *testing.T) {
	e := New()

	// First layout
	root := e.AddNode(Style{Display: DisplayBlock, Width: Px(400), Height: Px(300)})
	child := e.AddNode(Style{Display: DisplayBlock, Height: Px(100)})
	e.SetChildren(root, []NodeID{child})
	e.AddRoot(root)
	e.Compute(400, 300)
	expectResult(t, e, child, "first", 0, 0, 400, 100)

	// Clear and do a different layout
	e.Clear()
	root2 := e.AddNode(Style{Display: DisplayBlock, Width: Px(200), Height: Px(200)})
	child2 := e.AddNode(Style{Display: DisplayBlock, Height: Px(50)})
	e.SetChildren(root2, []NodeID{child2})
	e.AddRoot(root2)
	e.Compute(200, 200)
	expectResult(t, e, child2, "second", 0, 0, 200, 50)
}
