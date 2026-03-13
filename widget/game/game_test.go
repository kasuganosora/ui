package game

import (
	"testing"

	"github.com/kasuganosora/ui/core"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
	"github.com/kasuganosora/ui/widget"
)

func TestNewHUD(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	hud := NewHUD(tree, cfg)
	if hud == nil {
		t.Fatal("expected HUD")
	}
}

func TestHealthBar(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	hb := NewHealthBar(tree, cfg)

	hb.SetCurrent(75)
	hb.SetMax(100)
	if hb.Ratio() != 0.75 {
		t.Errorf("expected ratio 0.75, got %g", hb.Ratio())
	}

	hb.SetCurrent(0)
	if hb.Ratio() != 0 {
		t.Errorf("expected ratio 0, got %g", hb.Ratio())
	}

	hb.SetCurrent(150)
	if hb.Ratio() != 1 {
		t.Errorf("expected ratio capped at 1, got %g", hb.Ratio())
	}

	buf := render.NewCommandBuffer()
	hb.Draw(buf)
	if buf.Len() < 1 {
		t.Error("expected render commands from HealthBar")
	}
}

func TestHotbar(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	hb := NewHotbar(tree, 8, cfg)

	if hb.SlotCount() != 8 {
		t.Errorf("expected 8 slots, got %d", hb.SlotCount())
	}

	hb.SetSlot(0, HotbarSlot{Label: "Sword", Available: true})
	slot := hb.GetSlot(0)
	if slot.Label != "Sword" {
		t.Errorf("expected label 'Sword', got %q", slot.Label)
	}

	hb.SetSelected(2)
	if hb.Selected() != 2 {
		t.Errorf("expected selected 2, got %d", hb.Selected())
	}
}

func TestCooldownMask(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	cm := NewCooldownMask(tree, cfg)

	cm.SetRatio(0.5)
	if cm.Ratio() != 0.5 {
		t.Errorf("expected 0.5, got %g", cm.Ratio())
	}

	buf := render.NewCommandBuffer()
	cm.Draw(buf)
	// No bounds set, should still produce commands if ratio > 0
}

func TestInventory(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	inv := NewInventory(tree, 4, 6, cfg)

	if inv.Rows() != 4 {
		t.Errorf("expected 4 rows, got %d", inv.Rows())
	}
	if inv.Cols() != 6 {
		t.Errorf("expected 6 cols, got %d", inv.Cols())
	}

	item := &ItemData{ID: "sword", Name: "Iron Sword", Quantity: 1, Rarity: RarityCommon}
	inv.SetItem(0, item)
	got := inv.GetItem(0)
	if got == nil || got.Name != "Iron Sword" {
		t.Error("expected item in slot 0")
	}

	removed := inv.RemoveItem(0)
	if removed == nil {
		t.Error("expected removed item")
	}
	if inv.GetItem(0) != nil {
		t.Error("slot 0 should be empty after remove")
	}
}

func TestInventoryRarityColors(t *testing.T) {
	colors := []ItemRarity{RarityCommon, RarityUncommon, RarityRare, RarityEpic, RarityLegendary}
	for _, r := range colors {
		c := rarityColor(r)
		if c.A == 0 {
			t.Errorf("rarity %d color should have alpha", r)
		}
	}
}

func TestChatBox(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	cb := NewChatBox(tree, cfg)

	cb.AddMessage(ChatMessage{Sender: "Player1", Text: "Hello!"})
	cb.AddMessage(ChatMessage{Sender: "Player2", Text: "Hi there!"})

	if len(cb.Messages()) != 2 {
		t.Errorf("expected 2 messages, got %d", len(cb.Messages()))
	}

	cb.ClearMessages()
	if len(cb.Messages()) != 0 {
		t.Errorf("expected 0 messages after clear, got %d", len(cb.Messages()))
	}
}

func TestFloatingText(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	ft := NewFloatingText(tree, "-50", 100, 200, rarityColor(RarityEpic), cfg)

	buf := render.NewCommandBuffer()
	ft.Draw(buf)
	if buf.Len() == 0 {
		t.Error("expected render commands from FloatingText")
	}
}

func TestItemTooltip(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	tt := NewItemTooltip(tree, cfg)

	tt.SetItem(&ItemData{Name: "Dragon Sword", Rarity: RarityLegendary})
	tt.SetVisible(true)
	tt.SetPosition(100, 100)

	if !tt.IsVisible() {
		t.Error("tooltip should be visible")
	}

	buf := render.NewCommandBuffer()
	tt.Draw(buf)
	// Overlay commands
	if buf.Len() == 0 {
		// Overlay commands go to overlays, not commands
		if buf.Len() == 0 {
			t.Error("expected render commands")
		}
	}
}

func TestNotificationToast(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	nt := NewNotificationToast(tree, "Item acquired!", cfg)
	nt.SetToastType(ToastSuccess)
	nt.SetPosition(10, 10)

	if !nt.IsVisible() {
		t.Error("toast should be visible")
	}

	buf := render.NewCommandBuffer()
	nt.Draw(buf)
	if buf.Len() == 0 {
		t.Error("expected render commands from toast")
	}
}

func TestHUDDrag(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	hud := NewHUD(tree, cfg)

	// Create a draggable panel
	panel := NewInventory(tree, 2, 2, cfg)
	tree.SetLayout(panel.ElementID(), core.LayoutResult{
		Bounds: uimath.NewRect(0, 0, 100, 100),
	})
	hud.AddElementDraggable(panel, AnchorTopLeft, 50, 50)

	// Layout at 800x600
	hud.LayoutElements(800, 600)

	// Panel should be at (50, 50) after layout
	if e := tree.Get(panel.ElementID()); e != nil {
		b := e.Layout().Bounds
		if b.X != 50 || b.Y != 50 {
			t.Errorf("panel at (%g,%g), expected (50,50)", b.X, b.Y)
		}
	}

	// Non-draggable element should not start drag
	bar := NewHealthBar(tree, cfg)
	tree.SetLayout(bar.ElementID(), core.LayoutResult{
		Bounds: uimath.NewRect(0, 0, 200, 20),
	})
	hud.AddElement(bar, AnchorTopLeft, 10, 10)
	hud.LayoutElements(800, 600)

	if hud.HandleMouseDown(15, 15) {
		t.Error("should not drag non-draggable element")
	}

	// Click on draggable panel should start drag
	if !hud.HandleMouseDown(60, 60) {
		t.Fatal("expected drag to start on panel")
	}
	if !hud.IsDragging() {
		t.Error("should be dragging")
	}

	// Drag 30px right, 20px down
	hud.HandleMouseMove(90, 80)
	hud.LayoutElements(800, 600)

	if e := tree.Get(panel.ElementID()); e != nil {
		b := e.Layout().Bounds
		// Offset was 50+30=80, 50+20=70
		if b.X != 80 || b.Y != 70 {
			t.Errorf("panel at (%g,%g) after drag, expected (80,70)", b.X, b.Y)
		}
	}

	// Release
	if !hud.HandleMouseUp() {
		t.Error("expected drag end")
	}
	if hud.IsDragging() {
		t.Error("should not be dragging after release")
	}

	// MouseUp without drag should return false
	if hud.HandleMouseUp() {
		t.Error("expected false when not dragging")
	}
}

func TestHUDBringToFront(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	hud := NewHUD(tree, cfg)

	// Create two draggable panels: A at (50,50), B at (60,60) overlapping
	panelA := NewInventory(tree, 2, 2, cfg)
	tree.SetLayout(panelA.ElementID(), core.LayoutResult{Bounds: uimath.NewRect(0, 0, 100, 100)})
	hud.AddElementDraggable(panelA, AnchorTopLeft, 50, 50)

	panelB := NewInventory(tree, 2, 2, cfg)
	tree.SetLayout(panelB.ElementID(), core.LayoutResult{Bounds: uimath.NewRect(0, 0, 100, 100)})
	hud.AddElementDraggable(panelB, AnchorTopLeft, 60, 60)

	hud.LayoutElements(800, 600)

	// B is last in array → drawn on top. Click at (70,70) hits B (top-most).
	buf := render.NewCommandBuffer()
	hud.Draw(buf)

	// Now click on A's unique region (55, 55) — only A is there
	if !hud.HandleMouseDown(55, 55) {
		t.Fatal("expected to start dragging A")
	}
	hud.HandleMouseUp()

	// After clicking A, A should be the last element (brought to front)
	lastElem := hud.elements[len(hud.elements)-1]
	if lastElem.Widget.ElementID() != panelA.ElementID() {
		t.Error("expected panel A to be brought to front (last in draw order)")
	}
}

func TestItoa(t *testing.T) {
	tests := []struct {
		in   int
		want string
	}{
		{0, "0"},
		{1, "1"},
		{42, "42"},
		{100, "100"},
	}
	for _, tt := range tests {
		got := itoa(tt.in)
		if got != tt.want {
			t.Errorf("itoa(%d) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

// ============================================================
// ChatBox: InputH, InputBounds, MessageBounds, HandleWheel
// ============================================================

func TestChatBoxInputH(t *testing.T) {
	cb := NewChatBox(core.NewTree(), widget.DefaultConfig())
	if cb.InputH() != 28 {
		t.Errorf("expected default InputH 28, got %g", cb.InputH())
	}
}

func TestChatBoxInputBoundsNoBounds(t *testing.T) {
	cb := NewChatBox(core.NewTree(), widget.DefaultConfig())
	ib := cb.InputBounds()
	// Fallback: width=350, height=250, inputH=28
	if ib.Width != 350 || ib.Height != 28 {
		t.Errorf("unexpected InputBounds %v", ib)
	}
	if ib.Y != 250-28 {
		t.Errorf("expected Y=%g, got %g", float32(250-28), ib.Y)
	}
}

func TestChatBoxInputBoundsWithBounds(t *testing.T) {
	tree := core.NewTree()
	cb := NewChatBox(tree, widget.DefaultConfig())
	tree.SetLayout(cb.ElementID(), core.LayoutResult{
		Bounds: uimath.NewRect(10, 20, 400, 300),
	})
	ib := cb.InputBounds()
	if ib.X != 10 || ib.Width != 400 || ib.Height != 28 {
		t.Errorf("unexpected InputBounds %v", ib)
	}
	if ib.Y != 20+300-28 {
		t.Errorf("expected Y=%g, got %g", float32(20+300-28), ib.Y)
	}
}

func TestChatBoxMessageBoundsNoBounds(t *testing.T) {
	cb := NewChatBox(core.NewTree(), widget.DefaultConfig())
	mb := cb.MessageBounds()
	// Fallback: width=350, height=250, inputH=28
	if mb.Width != 350 || mb.Height != 250-28 {
		t.Errorf("unexpected MessageBounds %v", mb)
	}
}

func TestChatBoxMessageBoundsWithBounds(t *testing.T) {
	tree := core.NewTree()
	cb := NewChatBox(tree, widget.DefaultConfig())
	tree.SetLayout(cb.ElementID(), core.LayoutResult{
		Bounds: uimath.NewRect(10, 20, 400, 300),
	})
	mb := cb.MessageBounds()
	if mb.X != 10 || mb.Y != 20 || mb.Width != 400 || mb.Height != 300-28 {
		t.Errorf("unexpected MessageBounds %v", mb)
	}
}

func TestChatBoxHandleWheelNoBounds(t *testing.T) {
	cb := NewChatBox(core.NewTree(), widget.DefaultConfig())
	// No bounds set => returns false
	if cb.HandleWheel(10, 10, -1) {
		t.Error("expected false with no bounds")
	}
}

func TestChatBoxHandleWheelOutside(t *testing.T) {
	tree := core.NewTree()
	cb := NewChatBox(tree, widget.DefaultConfig())
	tree.SetLayout(cb.ElementID(), core.LayoutResult{
		Bounds: uimath.NewRect(100, 100, 200, 200),
	})
	// Outside bounds
	if cb.HandleWheel(50, 50, -1) {
		t.Error("expected false outside bounds")
	}
}

func TestChatBoxHandleWheelInside(t *testing.T) {
	tree := core.NewTree()
	cb := NewChatBox(tree, widget.DefaultConfig())
	tree.SetLayout(cb.ElementID(), core.LayoutResult{
		Bounds: uimath.NewRect(0, 0, 350, 250),
	})
	cb.SetMaxVisible(2)
	for i := 0; i < 5; i++ {
		cb.AddMessage(ChatMessage{Sender: "U", Text: "msg"})
	}
	// Scroll up (deltaY < 0)
	if !cb.HandleWheel(100, 100, -1) {
		t.Error("expected true inside bounds")
	}
	// Scroll down (deltaY > 0)
	if !cb.HandleWheel(100, 100, 1) {
		t.Error("expected true inside bounds")
	}
	// deltaY == 0
	if !cb.HandleWheel(100, 100, 0) {
		t.Error("expected true inside bounds even with zero delta")
	}
}

// ============================================================
// Inventory: Drag & Drop
// ============================================================

func TestInventoryDragDrop(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	inv := NewInventory(tree, 4, 6, cfg)
	inv.SetTitle("Bag")

	// Set bounds so contentOrigin and slotAt work
	tree.SetLayout(inv.ElementID(), core.LayoutResult{
		Bounds: uimath.NewRect(0, 0, 320, 260),
	})

	sword := &ItemData{ID: "sword", Name: "Sword", Quantity: 1, Rarity: RarityRare}
	shield := &ItemData{ID: "shield", Name: "Shield", Quantity: 1, Rarity: RarityEpic}
	inv.SetItem(0, sword)
	inv.SetItem(1, shield)

	if inv.IsDragging() {
		t.Error("should not be dragging initially")
	}

	// contentOrigin: pad=8, titleH=28 (title is set), so origin = (8, 28)
	// slot 0 is at (8, 28), size 48x48 with gap 4
	// Click in slot 0
	if !inv.HandleMouseDown(20, 40) {
		t.Error("expected drag to start on slot 0")
	}
	if !inv.IsDragging() {
		t.Error("should be dragging")
	}

	// Move to slot 1: slot 1 is at (8+52, 28) = (60, 28)
	if !inv.HandleMouseMove(70, 40) {
		t.Error("expected true during drag")
	}

	// Drop on slot 1
	if !inv.HandleMouseUp(70, 40) {
		t.Error("expected drag end")
	}
	if inv.IsDragging() {
		t.Error("should not be dragging after mouse up")
	}

	// Items should be swapped
	if inv.GetItem(0) != shield {
		t.Error("expected shield in slot 0 after swap")
	}
	if inv.GetItem(1) != sword {
		t.Error("expected sword in slot 1 after swap")
	}
}

func TestInventoryDragEmptySlot(t *testing.T) {
	tree := core.NewTree()
	inv := NewInventory(tree, 4, 6, widget.DefaultConfig())
	tree.SetLayout(inv.ElementID(), core.LayoutResult{
		Bounds: uimath.NewRect(0, 0, 320, 260),
	})

	// Click empty slot - should not start drag
	if inv.HandleMouseDown(20, 10) {
		t.Error("should not start drag on empty slot")
	}
}

func TestInventoryDragOutsideGrid(t *testing.T) {
	tree := core.NewTree()
	inv := NewInventory(tree, 4, 6, widget.DefaultConfig())
	tree.SetLayout(inv.ElementID(), core.LayoutResult{
		Bounds: uimath.NewRect(0, 0, 320, 260),
	})

	// Click outside grid area
	if inv.HandleMouseDown(319, 259) {
		t.Error("should not start drag outside grid")
	}
}

func TestInventoryMouseMoveNotDragging(t *testing.T) {
	tree := core.NewTree()
	inv := NewInventory(tree, 4, 6, widget.DefaultConfig())
	tree.SetLayout(inv.ElementID(), core.LayoutResult{
		Bounds: uimath.NewRect(0, 0, 320, 260),
	})
	if inv.HandleMouseMove(50, 50) {
		t.Error("expected false when not dragging")
	}
}

func TestInventoryMouseUpNotDragging(t *testing.T) {
	tree := core.NewTree()
	inv := NewInventory(tree, 4, 6, widget.DefaultConfig())
	tree.SetLayout(inv.ElementID(), core.LayoutResult{
		Bounds: uimath.NewRect(0, 0, 320, 260),
	})
	if inv.HandleMouseUp(50, 50) {
		t.Error("expected false when not dragging")
	}
}

func TestInventoryDragDropSameSlot(t *testing.T) {
	tree := core.NewTree()
	inv := NewInventory(tree, 4, 6, widget.DefaultConfig())
	tree.SetLayout(inv.ElementID(), core.LayoutResult{
		Bounds: uimath.NewRect(0, 0, 320, 260),
	})
	inv.SetItem(0, &ItemData{ID: "sword", Name: "Sword", Quantity: 1})
	inv.HandleMouseDown(20, 10)
	// Drop on same slot
	inv.HandleMouseUp(20, 10)
	// Item should still be in slot 0
	if inv.GetItem(0) == nil {
		t.Error("expected item still in slot 0")
	}
}

func TestInventoryDragWithCallbacks(t *testing.T) {
	tree := core.NewTree()
	inv := NewInventory(tree, 4, 6, widget.DefaultConfig())
	tree.SetLayout(inv.ElementID(), core.LayoutResult{
		Bounds: uimath.NewRect(0, 0, 320, 260),
	})

	item := &ItemData{ID: "sword", Name: "Sword", Quantity: 1, Rarity: RarityRare}
	inv.SetItem(0, item)

	selectCalled := false
	inv.OnSelect(func(i int, it *ItemData) { selectCalled = true })
	dropCalled := false
	inv.OnDrop(func(i int, d any) { dropCalled = true })

	inv.HandleMouseDown(20, 10) // slot 0
	if !selectCalled {
		t.Error("expected select callback")
	}

	// Move and drop on slot 1
	inv.HandleMouseMove(70, 10)
	inv.HandleMouseUp(70, 10)
	if !dropCalled {
		t.Error("expected drop callback")
	}
}

func TestInventoryDrawWithDrag(t *testing.T) {
	tree := core.NewTree()
	inv := NewInventory(tree, 2, 3, widget.DefaultConfig())
	tree.SetLayout(inv.ElementID(), core.LayoutResult{
		Bounds: uimath.NewRect(0, 0, 200, 150),
	})
	inv.SetItem(0, &ItemData{ID: "sword", Name: "Sword", Quantity: 3, Rarity: RarityEpic, Icon: 0})
	inv.SetItem(1, &ItemData{ID: "shield", Name: "Shield", Quantity: 1, Rarity: RarityCommon, Icon: 0})

	// Start drag
	inv.HandleMouseDown(12, 4)
	inv.HandleMouseMove(60, 4)

	buf := render.NewCommandBuffer()
	inv.Draw(buf)
	if buf.Len() == 0 {
		t.Error("expected render commands during drag")
	}
}

func TestInventorySlotAtGap(t *testing.T) {
	tree := core.NewTree()
	inv := NewInventory(tree, 4, 6, widget.DefaultConfig())
	tree.SetLayout(inv.ElementID(), core.LayoutResult{
		Bounds: uimath.NewRect(0, 0, 320, 260),
	})
	// Click in the gap between slot 0 and slot 1
	// slot 0 ends at x = 8 + 48 = 56, gap starts there, slot 1 starts at 56+4=60 (x-wise, but slotAt expects within the slot)
	// Actually need to check: pad=8, slotSize=48, gap=4
	// col = (x - 8) / 52
	// For x=56: col = 48/52 = 0.92 -> int(0.92) = 0
	// sx = 8 + 0*52 = 8; x >= 8 && x < 56 => 56 < 56 is false! So it's in the gap
	inv.SetItem(0, &ItemData{ID: "x", Name: "X"})
	if inv.HandleMouseDown(56, 10) {
		t.Error("should not start drag in gap area")
	}
}

// ============================================================
// HUD: DragHandleH
// ============================================================

func TestHUDDragHandleH(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	hud := NewHUD(tree, cfg)

	panel := NewInventory(tree, 2, 2, cfg)
	tree.SetLayout(panel.ElementID(), core.LayoutResult{
		Bounds: uimath.NewRect(50, 50, 100, 100),
	})
	// Title bar is 30px, so only top 30px should be draggable
	hud.AddElementDraggable(panel, AnchorTopLeft, 50, 50, 30)
	hud.LayoutElements(800, 600)

	// Click in content area (below title bar) should NOT start drag
	if hud.HandleMouseDown(60, 90) {
		t.Error("should not drag from content area when DragHandleH is set")
	}

	// Click in title bar area should start drag
	if !hud.HandleMouseDown(60, 60) {
		t.Error("expected drag to start in title bar area")
	}
	hud.HandleMouseUp()
}

// ============================================================
// HUD: HandleMouseMove when not dragging
// ============================================================

func TestHUDHandleMouseMoveNotDragging(t *testing.T) {
	hud := NewHUD(core.NewTree(), widget.DefaultConfig())
	if hud.HandleMouseMove(100, 100) {
		t.Error("expected false when not dragging")
	}
}

// ============================================================
// Panel: Shadow branch
// ============================================================

func TestPanelDrawShadow(t *testing.T) {
	buf := render.NewCommandBuffer()
	cfg := widget.DefaultConfig()
	p := Panel{
		Title:   "Test",
		Width:   200,
		Height:  100,
		Shadow:  true,
		Padding: 12,
		BgColor: uimath.RGBA(0.1, 0.1, 0.1, 0.9),
		BorderColor: uimath.RGBA(0.5, 0.5, 0.5, 1),
		BorderWidth: 2,
		TitleColor: uimath.ColorWhite,
	}
	r := p.Draw(buf, uimath.Rect{}, cfg, 50)
	if r.PanelW != 200 {
		t.Errorf("expected PanelW 200, got %g", r.PanelW)
	}
	// Shadow should add at least 1 more command
	if buf.Len() < 3 {
		t.Errorf("expected at least 3 commands (shadow+bg+separator), got %d", buf.Len())
	}
}

func TestPanelDrawNoTitle(t *testing.T) {
	buf := render.NewCommandBuffer()
	cfg := widget.DefaultConfig()
	p := Panel{Width: 200, Height: 100, TitleH: -1}
	r := p.Draw(buf, uimath.Rect{}, cfg, 0)
	if r.ContentY != r.PanelY {
		t.Error("no title bar means content starts at panel top")
	}
}

func TestPanelDrawZeroWidthUsesDefault(t *testing.T) {
	buf := render.NewCommandBuffer()
	cfg := widget.DefaultConfig()
	p := Panel{}
	r := p.Draw(buf, uimath.Rect{}, cfg, 0)
	if r.PanelW != 300 {
		t.Errorf("expected default width 300, got %g", r.PanelW)
	}
}

// ============================================================
// DialogueBox: no-speaker draw branch
// ============================================================

func TestDialogueBoxDrawNoSpeaker(t *testing.T) {
	tree := core.NewTree()
	db := NewDialogueBox(tree, widget.DefaultConfig())
	db.Show("", "Some text with no speaker")
	buf := render.NewCommandBuffer()
	db.Draw(buf)
	if buf.Len() == 0 {
		t.Error("expected render commands")
	}
}

// ============================================================
// SkillTree: off-screen node culling
// ============================================================

func TestSkillTreeDrawOffscreen(t *testing.T) {
	tree := core.NewTree()
	st := NewSkillTree(tree, widget.DefaultConfig())
	tree.SetLayout(st.ElementID(), core.LayoutResult{
		Bounds: uimath.NewRect(0, 0, 200, 200),
	})
	// Add a node far off-screen
	st.AddNode(&SkillNode{ID: "far", Name: "Far", X: 9999, Y: 9999, MaxLevel: 3})
	buf := render.NewCommandBuffer()
	st.Draw(buf)
	// The node should be culled (off-screen), so fewer draw calls
}

// ── Window tests ─────────────────────────────────────────────────────────

func TestWindowNew(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	w := NewWindow(tree, "Test", cfg)
	if w.Title() != "Test" {
		t.Fatalf("title = %q, want Test", w.Title())
	}
	if !w.Visible() {
		t.Fatal("window should be visible by default")
	}
	if !w.ShowClose() {
		t.Fatal("close button should be visible by default")
	}
	if w.TitleH() != 28 {
		t.Fatalf("titleH = %g, want 28", w.TitleH())
	}
}

func TestWindowSetters(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	w := NewWindow(tree, "Win", cfg)
	w.SetTitle("New")
	w.SetTitleH(32)
	w.SetVisible(false)
	w.SetShowClose(false)
	w.SetShadow(false)
	w.SetSize(200, 150)
	w.SetBgColor(uimath.ColorHex("#112233"))
	w.SetBorderColor(uimath.ColorHex("#445566"))
	w.SetBorderWidth(2)
	w.SetTitleColor(uimath.ColorHex("#ff0000"))
	w.SetTitleBg(uimath.ColorHex("#00ff00"))
	if w.Title() != "New" {
		t.Error("SetTitle failed")
	}
	if w.TitleH() != 32 {
		t.Error("SetTitleH failed")
	}
	if w.Visible() {
		t.Error("SetVisible failed")
	}
	if w.ShowClose() {
		t.Error("SetShowClose failed")
	}
}

func TestWindowDrawVisible(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	w := NewWindow(tree, "Draw", cfg)
	tree.SetLayout(w.ElementID(), core.LayoutResult{
		Bounds: uimath.NewRect(10, 10, 300, 250),
	})
	buf := render.NewCommandBuffer()
	w.Draw(buf)
	if buf.Len() == 0 {
		t.Error("expected draw commands for visible window")
	}
}

func TestWindowDrawHidden(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	w := NewWindow(tree, "Hidden", cfg)
	w.SetVisible(false)
	buf := render.NewCommandBuffer()
	w.Draw(buf)
	if buf.Len() != 0 {
		t.Error("hidden window should not draw")
	}
}

func TestWindowDrawNoClose(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	w := NewWindow(tree, "NoClose", cfg)
	w.SetShowClose(false)
	tree.SetLayout(w.ElementID(), core.LayoutResult{
		Bounds: uimath.NewRect(0, 0, 200, 200),
	})
	buf := render.NewCommandBuffer()
	w.Draw(buf)
	if buf.Len() == 0 {
		t.Error("expected draw commands")
	}
}

func TestWindowDrawFallbackBounds(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	w := NewWindow(tree, "Fallback", cfg)
	w.SetSize(200, 150)
	// No SetLayout — should use fallback bounds
	buf := render.NewCommandBuffer()
	w.Draw(buf)
	if buf.Len() == 0 {
		t.Error("expected draw commands with fallback bounds")
	}
}

func TestWindowCloseDefault(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	root := widget.NewDiv(tree, cfg)
	tree.SetLayout(root.ElementID(), core.LayoutResult{Bounds: uimath.NewRect(0, 0, 800, 600)})

	w := NewWindow(tree, "Close", cfg)
	wm := NewWindowManager(tree, root)
	wm.Add(w, 0, 0, 200, 200)

	// Click close button area (right side of title bar)
	titleH := w.TitleH()
	btnSize := titleH * 0.6
	cx := 200 - titleH*0.2 - btnSize/2
	cy := titleH / 2
	wm.HandleMouseDown(cx, cy)
	if w.Visible() {
		t.Error("window should be hidden after close click")
	}
}

func TestWindowCloseCallback(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	root := widget.NewDiv(tree, cfg)
	tree.SetLayout(root.ElementID(), core.LayoutResult{Bounds: uimath.NewRect(0, 0, 800, 600)})

	w := NewWindow(tree, "CB", cfg)
	wm := NewWindowManager(tree, root)
	wm.Add(w, 0, 0, 200, 200)

	closed := false
	w.OnClose(func() { closed = true })
	titleH := w.TitleH()
	btnSize := titleH * 0.6
	cx := 200 - titleH*0.2 - btnSize/2
	cy := titleH / 2
	wm.HandleMouseDown(cx, cy)
	if !closed {
		t.Error("OnClose callback should have been called")
	}
	if !w.Visible() {
		t.Error("OnClose callback set — window should remain visible")
	}
}

func TestWindowDrag(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	root := widget.NewDiv(tree, cfg)
	tree.SetLayout(root.ElementID(), core.LayoutResult{Bounds: uimath.NewRect(0, 0, 800, 600)})

	w := NewWindow(tree, "Drag", cfg)
	wm := NewWindowManager(tree, root)
	wm.SetSnapEnabled(false)
	wm.Add(w, 100, 100, 200, 200)
	// Set initial layout so moveElement has something to work with
	tree.SetLayout(w.ElementID(), core.LayoutResult{Bounds: uimath.NewRect(100, 100, 200, 200)})

	// Mouse down on title bar (not on close button)
	wm.HandleMouseDown(120, 110)
	if !wm.IsDragging() {
		t.Fatal("should be dragging after title bar mouse down")
	}

	// Mouse move
	wm.HandleMouseMove(170, 160)
	// PostLayout applies the position
	wm.PostLayout()
	b := w.Bounds()
	if b.X != 150 || b.Y != 150 {
		t.Errorf("bounds after drag move = (%g,%g), want (150,150)", b.X, b.Y)
	}

	// Mouse up
	wm.HandleMouseUp()
	if wm.IsDragging() {
		t.Error("should not be dragging after mouse up")
	}

	// After CSSLayout resets, PostLayout re-applies WM position
	tree.SetLayout(w.ElementID(), core.LayoutResult{Bounds: uimath.NewRect(0, 0, 200, 200)})
	wm.PostLayout()
	b = w.Bounds()
	if b.X != 150 || b.Y != 150 {
		t.Errorf("persistent position after CSSLayout reset = (%g,%g), want (150,150)", b.X, b.Y)
	}
}

func TestWindowPostLayout(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	root := widget.NewDiv(tree, cfg)
	tree.SetLayout(root.ElementID(), core.LayoutResult{Bounds: uimath.NewRect(0, 0, 800, 600)})

	w := NewWindow(tree, "PL", cfg)
	wm := NewWindowManager(tree, root)
	wm.Add(w, 50, 50, 200, 200)
	tree.SetLayout(w.ElementID(), core.LayoutResult{Bounds: uimath.NewRect(50, 50, 200, 200)})

	// No drag — PostLayout should keep position at (50,50)
	wm.PostLayout()
	b := w.Bounds()
	if b.X != 50 || b.Y != 50 {
		t.Errorf("PostLayout bounds = (%g,%g), want (50,50)", b.X, b.Y)
	}

	// Drag: down at (60,55), move to (110,105) → new position = (100,100)
	wm.SetSnapEnabled(false)
	wm.HandleMouseDown(60, 55)
	wm.HandleMouseMove(110, 105)
	wm.HandleMouseUp()

	// Simulate CSSLayout overwriting
	tree.SetLayout(w.ElementID(), core.LayoutResult{Bounds: uimath.NewRect(0, 0, 200, 200)})
	// PostLayout should re-apply WM position
	wm.PostLayout()
	b = w.Bounds()
	if b.X != 100 || b.Y != 100 {
		t.Errorf("PostLayout bounds = (%g,%g), want (100,100)", b.X, b.Y)
	}
}

func TestWindowAutoHeightDrag(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	root := widget.NewDiv(tree, cfg)
	tree.SetLayout(root.ElementID(), core.LayoutResult{Bounds: uimath.NewRect(0, 0, 800, 600)})

	w := NewWindow(tree, "Auto", cfg)
	wm := NewWindowManager(tree, root)
	wm.Add(w, 100, 100, 200, 0) // h=0 = auto-height

	// Simulate CSS layout computing height = 150
	tree.SetLayout(w.ElementID(), core.LayoutResult{Bounds: uimath.NewRect(100, 100, 200, 150)})

	// Click on title bar should start drag
	consumed := wm.HandleMouseDown(150, 110)
	if !consumed {
		t.Error("title bar click should be consumed")
	}
	if !wm.IsDragging() {
		t.Error("should be dragging after title bar click on auto-height window")
	}
	wm.HandleMouseUp()

	// Click in content area (y=200, within 100+150=250)
	consumed = wm.HandleMouseDown(150, 200)
	if !consumed {
		t.Error("content click should be consumed by auto-height window")
	}
	if wm.IsDragging() {
		t.Error("content click should not start drag")
	}

	// Click outside (y=260, beyond 100+150=250)
	wm.HandleMouseUp()
	consumed = wm.HandleMouseDown(150, 260)
	if consumed {
		t.Error("click outside auto-height window should not be consumed")
	}
}

func TestWindowSnapToViewportEdges(t *testing.T) {
	setup := func() (*core.Tree, *widget.Config, *widget.Div, *WindowManager) {
		tree := core.NewTree()
		cfg := widget.DefaultConfig()
		root := widget.NewDiv(tree, cfg)
		tree.SetLayout(root.ElementID(), core.LayoutResult{Bounds: uimath.NewRect(0, 0, 800, 600)})
		wm := NewWindowManager(tree, root)
		return tree, cfg, root, wm
	}

	t.Run("default_enabled", func(t *testing.T) {
		_, _, _, wm := setup()
		if !wm.SnapEnabled() {
			t.Fatal("snap should be on by default")
		}
	})

	t.Run("left_edge", func(t *testing.T) {
		tree, cfg, _, wm := setup()
		w := NewWindow(tree, "L", cfg)
		wm.Add(w, 100, 100, 200, 200)
		tree.SetLayout(w.ElementID(), core.LayoutResult{Bounds: uimath.NewRect(100, 100, 200, 200)})
		// offset = (200-100, 114-100) = (100, 14)
		wm.HandleMouseDown(200, 114)
		wm.HandleMouseMove(105, 114) // newX = 5, within snap → 0
		wm.HandleMouseUp()
		wm.PostLayout()
		if b := w.Bounds(); b.X != 0 {
			t.Errorf("expected snap to left edge, got X=%g", b.X)
		}
	})

	t.Run("top_edge", func(t *testing.T) {
		tree, cfg, _, wm := setup()
		w := NewWindow(tree, "T", cfg)
		wm.Add(w, 100, 100, 200, 200)
		tree.SetLayout(w.ElementID(), core.LayoutResult{Bounds: uimath.NewRect(100, 100, 200, 200)})
		// offset = (100, 14)
		wm.HandleMouseDown(200, 114)
		wm.HandleMouseMove(200, 19) // newY = 19-14 = 5 → snap to 0
		wm.HandleMouseUp()
		wm.PostLayout()
		if b := w.Bounds(); b.Y != 0 {
			t.Errorf("expected snap to top edge, got Y=%g", b.Y)
		}
	})

	t.Run("right_edge", func(t *testing.T) {
		tree, cfg, _, wm := setup()
		w := NewWindow(tree, "R", cfg)
		wm.Add(w, 300, 100, 200, 200)
		tree.SetLayout(w.ElementID(), core.LayoutResult{Bounds: uimath.NewRect(300, 100, 200, 200)})
		// offset = (400-300, 114-100) = (100, 14)
		wm.HandleMouseDown(400, 114)
		wm.HandleMouseMove(705, 114) // newX = 705-100 = 605, gap = 800-605-200 = -5 → snap to 600
		wm.HandleMouseUp()
		wm.PostLayout()
		if b := w.Bounds(); b.X != 600 {
			t.Errorf("expected snap to right edge (X=600), got X=%g", b.X)
		}
	})

	t.Run("bottom_edge", func(t *testing.T) {
		tree, cfg, _, wm := setup()
		w := NewWindow(tree, "B", cfg)
		wm.Add(w, 100, 200, 200, 200)
		tree.SetLayout(w.ElementID(), core.LayoutResult{Bounds: uimath.NewRect(100, 200, 200, 200)})
		// offset = (200-100, 214-200) = (100, 14)
		wm.HandleMouseDown(200, 214)
		wm.HandleMouseMove(200, 409) // newY = 409-14 = 395, gap = 600-395-200 = 5 → snap to 400
		wm.HandleMouseUp()
		wm.PostLayout()
		if b := w.Bounds(); b.Y != 400 {
			t.Errorf("expected snap to bottom edge (Y=400), got Y=%g", b.Y)
		}
	})
}

func TestWindowSnapDisabled(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	root := widget.NewDiv(tree, cfg)
	tree.SetLayout(root.ElementID(), core.LayoutResult{Bounds: uimath.NewRect(0, 0, 800, 600)})

	w := NewWindow(tree, "NoSnap", cfg)
	wm := NewWindowManager(tree, root)
	wm.SetSnapEnabled(false)
	wm.Add(w, 100, 100, 200, 200)
	tree.SetLayout(w.ElementID(), core.LayoutResult{Bounds: uimath.NewRect(100, 100, 200, 200)})

	// Drag near left edge - should NOT snap
	wm.HandleMouseDown(200, 114) // offset = (100, 14)
	wm.HandleMouseMove(105, 114) // newX = 5
	wm.HandleMouseUp()
	wm.PostLayout()
	b := w.Bounds()
	if b.X != 5 {
		t.Errorf("expected X=5 (no snap), got X=%g", b.X)
	}
}

func TestWindowSnapViewportOnly(t *testing.T) {
	// Window-to-window snap is disabled; only viewport edges snap.
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	root := widget.NewDiv(tree, cfg)
	tree.SetLayout(root.ElementID(), core.LayoutResult{Bounds: uimath.NewRect(0, 0, 800, 600)})

	w1 := NewWindow(tree, "A", cfg)
	w2 := NewWindow(tree, "B", cfg)
	wm := NewWindowManager(tree, root)
	wm.Add(w1, 100, 100, 200, 200)
	wm.Add(w2, 400, 100, 200, 200)
	tree.SetLayout(w1.ElementID(), core.LayoutResult{Bounds: uimath.NewRect(100, 100, 200, 200)})
	tree.SetLayout(w2.ElementID(), core.LayoutResult{Bounds: uimath.NewRect(400, 100, 200, 200)})

	// Drag w2 near w1's right edge — should NOT snap (no window-to-window snap)
	wm.HandleMouseDown(500, 114) // offset = (100, 14)
	wm.HandleMouseMove(405, 114) // newX = 405-100 = 305
	wm.HandleMouseUp()
	wm.PostLayout()
	b := w2.Bounds()
	if b.X != 305 {
		t.Errorf("expected no window-to-window snap (X=305), got X=%g", b.X)
	}
}

func TestWindowContentClickNoDrag(t *testing.T) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	root := widget.NewDiv(tree, cfg)
	tree.SetLayout(root.ElementID(), core.LayoutResult{Bounds: uimath.NewRect(0, 0, 800, 600)})

	w := NewWindow(tree, "NoDrag", cfg)
	wm := NewWindowManager(tree, root)
	wm.Add(w, 0, 0, 200, 200)

	// Click below title bar — should NOT start drag
	consumed := wm.HandleMouseDown(100, 100)
	if !consumed {
		t.Error("content area click should be consumed by window")
	}
	if wm.IsDragging() {
		t.Error("clicking content area should not start drag")
	}
}
