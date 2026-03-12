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
