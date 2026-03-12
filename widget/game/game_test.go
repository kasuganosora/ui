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
