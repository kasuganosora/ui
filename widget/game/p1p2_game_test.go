package game

import (
	"testing"

	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/event"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
	"github.com/kasuganosora/ui/widget"
)

func newTree() *core.Tree    { return core.NewTree() }
func newCfg() *widget.Config { return widget.DefaultConfig() }

// mockTextDrawer implements widget.TextDrawer for testing text-rendering branches.
type mockTextDrawer struct{}

func (m *mockTextDrawer) DrawText(buf *render.CommandBuffer, text string, x, y, fontSize, maxWidth float32, color uimath.Color, opacity float32) {
	buf.DrawRect(render.RectCmd{
		Bounds:    uimath.NewRect(x, y, maxWidth, fontSize*1.2),
		FillColor: color,
	}, 99, opacity)
}
func (m *mockTextDrawer) LineHeight(fontSize float32) float32 { return fontSize * 1.2 }
func (m *mockTextDrawer) MeasureText(text string, fontSize float32) float32 {
	return float32(len(text)) * fontSize * 0.5
}

func newCfgWithText() *widget.Config {
	cfg := widget.DefaultConfig()
	cfg.TextRenderer = &mockTextDrawer{}
	return cfg
}

func setBounds(tree *core.Tree, w widget.Widget, x, y, width, height float32) {
	tree.SetLayout(w.ElementID(), core.LayoutResult{
		Bounds: uimath.NewRect(x, y, width, height),
	})
}

// ============================================================
// BuffBar
// ============================================================

func TestBuffBarNilConfig(t *testing.T) {
	bb := NewBuffBar(newTree(), nil)
	if bb == nil {
		t.Fatal("expected non-nil BuffBar")
	}
}

func TestBuffBarAddRemoveClear(t *testing.T) {
	tree := newTree()
	cfg := newCfg()
	bb := NewBuffBar(tree, cfg)

	bb.AddBuff(Buff{ID: "str", Label: "Strength", Stacks: 3, Type: BuffPositive})
	bb.AddBuff(Buff{ID: "poison", Label: "Poison", Duration: 10, Type: BuffNegative})
	if len(bb.Buffs()) != 2 {
		t.Errorf("expected 2 buffs, got %d", len(bb.Buffs()))
	}

	bb.RemoveBuff("str")
	if len(bb.Buffs()) != 1 {
		t.Error("expected 1 buff after remove")
	}

	// Remove non-existent
	bb.RemoveBuff("nonexistent")
	if len(bb.Buffs()) != 1 {
		t.Error("expected still 1 buff after removing nonexistent")
	}

	bb.ClearBuffs()
	if len(bb.Buffs()) != 0 {
		t.Error("expected 0 buffs after clear")
	}
}

func TestBuffBarSetters(t *testing.T) {
	bb := NewBuffBar(newTree(), newCfg())
	bb.SetIconSize(48)
	bb.SetGap(8)
	bb.SetMaxIcons(32)
}

func TestBuffBarDrawNoBounds(t *testing.T) {
	bb := NewBuffBar(newTree(), newCfg())
	bb.AddBuff(Buff{ID: "x", Label: "X", Type: BuffPositive})
	buf := render.NewCommandBuffer()
	bb.Draw(buf)
	// No bounds set => early return
	if buf.Len() != 0 {
		t.Error("expected no commands without bounds")
	}
}

func TestBuffBarDrawWithBounds(t *testing.T) {
	tree := newTree()
	bb := NewBuffBar(tree, newCfg())
	setBounds(tree, bb, 0, 0, 400, 40)

	bb.AddBuff(Buff{ID: "str", Label: "Strength", Stacks: 3, Type: BuffPositive})
	bb.AddBuff(Buff{ID: "poison", Label: "Poison", Duration: 10, Type: BuffNegative})
	bb.AddBuff(Buff{ID: "shield", Label: "Shield", Duration: 0, Type: BuffPositive})

	buf := render.NewCommandBuffer()
	bb.Draw(buf)
	if buf.Len() == 0 {
		t.Error("expected render commands from BuffBar")
	}
}

func TestBuffBarDrawMaxIcons(t *testing.T) {
	tree := newTree()
	bb := NewBuffBar(tree, newCfg())
	setBounds(tree, bb, 0, 0, 400, 40)
	bb.SetMaxIcons(2)

	for i := 0; i < 5; i++ {
		bb.AddBuff(Buff{ID: itoa(i), Label: "B", Type: BuffPositive})
	}

	buf := render.NewCommandBuffer()
	bb.Draw(buf)
	if buf.Len() == 0 {
		t.Error("expected render commands")
	}
}

func TestBuffBarDrawDurationOverlay(t *testing.T) {
	tree := newTree()
	bb := NewBuffBar(tree, newCfg())
	setBounds(tree, bb, 0, 0, 400, 40)

	// Duration < 30 triggers cooldown sweep
	bb.AddBuff(Buff{ID: "dot", Label: "DoT", Duration: 15, Type: BuffNegative})
	buf := render.NewCommandBuffer()
	bb.Draw(buf)
	if buf.Len() < 2 {
		t.Error("expected at least 2 commands (background + duration overlay)")
	}
}

// ============================================================
// CastBar
// ============================================================

func TestCastBarNilConfig(t *testing.T) {
	cb := NewCastBar(newTree(), nil)
	if cb == nil {
		t.Fatal("expected non-nil CastBar")
	}
}

func TestCastBarStartTickComplete(t *testing.T) {
	tree := newTree()
	cfg := newCfg()
	cb := NewCastBar(tree, cfg)

	cb.StartCast("Fireball", 2.0)
	if !cb.IsVisible() {
		t.Error("should be visible during cast")
	}
	if cb.SpellName() != "Fireball" {
		t.Errorf("expected spell 'Fireball', got %q", cb.SpellName())
	}

	cb.Tick(1.0)
	if cb.Progress() < 0.49 || cb.Progress() > 0.51 {
		t.Errorf("expected ~0.5 progress, got %g", cb.Progress())
	}

	completed := false
	cb.OnComplete(func() { completed = true })
	cb.Tick(1.5)
	if !completed {
		t.Error("expected complete callback")
	}
	if cb.IsVisible() {
		t.Error("should be hidden after completion")
	}
}

func TestCastBarInterrupt(t *testing.T) {
	cb := NewCastBar(newTree(), newCfg())
	cb.StartCast("Heal", 3.0)

	interrupted := false
	cb.OnInterrupt(func() { interrupted = true })
	cb.Interrupt()
	if !interrupted {
		t.Error("expected interrupt callback")
	}
	if cb.IsVisible() {
		t.Error("should be hidden after interrupt")
	}
}

func TestCastBarInterruptNoCallback(t *testing.T) {
	cb := NewCastBar(newTree(), newCfg())
	cb.StartCast("Heal", 3.0)
	cb.Interrupt() // no callback set, should not panic
}

func TestCastBarTickWhenNotVisible(t *testing.T) {
	cb := NewCastBar(newTree(), newCfg())
	cb.Tick(1.0) // not visible, should be no-op
	if cb.Progress() != 0 {
		t.Error("progress should remain 0")
	}
}

func TestCastBarTickZeroCastTime(t *testing.T) {
	cb := NewCastBar(newTree(), newCfg())
	cb.StartCast("Instant", 0)
	cb.Tick(1.0) // castTime <= 0, should be no-op
}

func TestCastBarSetters(t *testing.T) {
	cb := NewCastBar(newTree(), newCfg())
	cb.SetColor(uimath.ColorHex("#ff0000"))
	cb.SetSize(300, 30)
}

func TestCastBarDrawNotVisible(t *testing.T) {
	cb := NewCastBar(newTree(), newCfg())
	buf := render.NewCommandBuffer()
	cb.Draw(buf)
	if buf.Len() != 0 || len(buf.Overlays()) != 0 {
		t.Error("expected no commands when not visible")
	}
}

func TestCastBarDrawVisible(t *testing.T) {
	cb := NewCastBar(newTree(), newCfg())
	cb.StartCast("Fireball", 2.0)
	cb.Tick(0.5)

	buf := render.NewCommandBuffer()
	cb.Draw(buf)
	if len(buf.Overlays()) == 0 {
		t.Error("expected overlay commands from CastBar")
	}
}

func TestCastBarDrawWithBounds(t *testing.T) {
	tree := newTree()
	cb := NewCastBar(tree, newCfg())
	setBounds(tree, cb, 10, 10, 300, 30)
	cb.StartCast("Heal", 2.0)
	cb.Tick(0.5)

	buf := render.NewCommandBuffer()
	cb.Draw(buf)
	if len(buf.Overlays()) == 0 {
		t.Error("expected overlay commands from CastBar with bounds")
	}
}

func TestCastBarDrawZeroProgress(t *testing.T) {
	cb := NewCastBar(newTree(), newCfg())
	cb.StartCast("Fireball", 2.0)
	// progress is 0, no fill bar
	buf := render.NewCommandBuffer()
	cb.Draw(buf)
	if len(buf.Overlays()) == 0 {
		t.Error("expected at least background overlay")
	}
}

func TestCastBarCompleteNoCallback(t *testing.T) {
	cb := NewCastBar(newTree(), newCfg())
	cb.StartCast("Heal", 1.0)
	cb.Tick(2.0) // completes without callback set
	if cb.IsVisible() {
		t.Error("should be hidden after completion")
	}
}

// ============================================================
// ChatBox
// ============================================================

func TestChatBoxNilConfig(t *testing.T) {
	cb := NewChatBox(newTree(), nil)
	if cb == nil {
		t.Fatal("expected non-nil ChatBox")
	}
}

func TestChatBoxAddClearMessages(t *testing.T) {
	cb := NewChatBox(newTree(), newCfg())
	cb.AddMessage(ChatMessage{Sender: "A", Text: "Hello"})
	cb.AddMessage(ChatMessage{Sender: "B", Text: "World"})
	if len(cb.Messages()) != 2 {
		t.Errorf("expected 2, got %d", len(cb.Messages()))
	}
	cb.ClearMessages()
	if len(cb.Messages()) != 0 {
		t.Error("expected 0 after clear")
	}
}

func TestChatBoxAutoScroll(t *testing.T) {
	cb := NewChatBox(newTree(), newCfg())
	cb.SetMaxVisible(2)
	for i := 0; i < 5; i++ {
		cb.AddMessage(ChatMessage{Sender: "U", Text: itoa(i)})
	}
	// scrollY should be 3 (5 - 2)
}

func TestChatBoxScrollUpDown(t *testing.T) {
	cb := NewChatBox(newTree(), newCfg())
	cb.SetMaxVisible(2)
	for i := 0; i < 5; i++ {
		cb.AddMessage(ChatMessage{Sender: "U", Text: itoa(i)})
	}
	// scrollY = 3
	cb.ScrollUp()
	cb.ScrollUp()
	cb.ScrollUp()
	cb.ScrollUp() // should not go below 0
	cb.ScrollDown()
	cb.ScrollDown()
	cb.ScrollDown()
	cb.ScrollDown() // should not go above max
}

func TestChatBoxScrollDownEmpty(t *testing.T) {
	cb := NewChatBox(newTree(), newCfg())
	cb.SetMaxVisible(10)
	cb.AddMessage(ChatMessage{Sender: "U", Text: "x"})
	cb.ScrollDown() // max < 0 => clamped to 0
	cb.ScrollUp()   // scrollY already 0
}

func TestChatBoxSetters(t *testing.T) {
	cb := NewChatBox(newTree(), newCfg())
	cb.SetSize(400, 300)
	cb.SetMaxVisible(20)
	cb.OnSend(func(text string) {})
	if cb.InputText() != "" {
		t.Error("expected empty input text")
	}
}

func TestChatBoxDrawNoBounds(t *testing.T) {
	cb := NewChatBox(newTree(), newCfg())
	cb.AddMessage(ChatMessage{Sender: "A", Text: "Hi"})
	buf := render.NewCommandBuffer()
	cb.Draw(buf)
	// Uses fallback bounds
	if buf.Len() == 0 {
		t.Error("expected commands with fallback bounds")
	}
}

func TestChatBoxDrawWithBounds(t *testing.T) {
	tree := newTree()
	cb := NewChatBox(tree, newCfg())
	setBounds(tree, cb, 0, 0, 350, 250)
	cb.AddMessage(ChatMessage{Sender: "A", Text: "Hi", Color: uimath.ColorHex("#ff0000")})
	cb.AddMessage(ChatMessage{Sender: "B", Text: "There"}) // default color
	buf := render.NewCommandBuffer()
	cb.Draw(buf)
	if buf.Len() == 0 {
		t.Error("expected commands from ChatBox")
	}
}

// ============================================================
// FloatingText
// ============================================================

func TestFloatingTextNilConfig(t *testing.T) {
	ft := NewFloatingText(newTree(), "test", 0, 0, uimath.ColorWhite, nil)
	if ft == nil {
		t.Fatal("expected non-nil FloatingText")
	}
}

func TestFloatingTextSetters(t *testing.T) {
	ft := NewFloatingText(newTree(), "test", 0, 0, uimath.ColorWhite, newCfg())
	ft.SetPosition(10, 20)
	ft.SetText("new text")
	ft.SetColor(uimath.ColorHex("#ff0000"))
}

func TestFloatingTextDraw(t *testing.T) {
	ft := NewFloatingText(newTree(), "-50", 100, 200, uimath.ColorHex("#ff0000"), newCfg())
	buf := render.NewCommandBuffer()
	ft.Draw(buf)
	if buf.Len() == 0 {
		t.Error("expected render commands")
	}
}

// ============================================================
// ItemTooltip
// ============================================================

func TestItemTooltipNilConfig(t *testing.T) {
	tt := NewItemTooltip(newTree(), nil)
	if tt == nil {
		t.Fatal("expected non-nil ItemTooltip")
	}
}

func TestItemTooltipNotVisible(t *testing.T) {
	tt := NewItemTooltip(newTree(), newCfg())
	buf := render.NewCommandBuffer()
	tt.Draw(buf)
	if buf.Len() != 0 || len(buf.Overlays()) != 0 {
		t.Error("expected no commands when not visible")
	}
}

func TestItemTooltipVisibleNoItem(t *testing.T) {
	tt := NewItemTooltip(newTree(), newCfg())
	tt.SetVisible(true)
	buf := render.NewCommandBuffer()
	tt.Draw(buf) // visible but no item => return
	if len(buf.Overlays()) != 0 {
		t.Error("expected no overlays with nil item")
	}
}

func TestItemTooltipDraw(t *testing.T) {
	tt := NewItemTooltip(newTree(), newCfg())
	tt.SetItem(&ItemData{Name: "Sword", Rarity: RarityLegendary})
	tt.SetVisible(true)
	tt.SetPosition(50, 50)

	if !tt.IsVisible() {
		t.Error("should be visible")
	}
	buf := render.NewCommandBuffer()
	tt.Draw(buf)
	if len(buf.Overlays()) == 0 {
		t.Error("expected overlay commands")
	}
}

// ============================================================
// NotificationToast
// ============================================================

func TestNotificationToastNilConfig(t *testing.T) {
	nt := NewNotificationToast(newTree(), "test", nil)
	if nt == nil {
		t.Fatal("expected non-nil")
	}
}

func TestNotificationToastSetters(t *testing.T) {
	nt := NewNotificationToast(newTree(), "test", newCfg())
	nt.SetText("new text")
	nt.SetPosition(10, 20)
	nt.SetVisible(false)
	if nt.IsVisible() {
		t.Error("should be hidden")
	}
}

func TestNotificationToastDrawNotVisible(t *testing.T) {
	nt := NewNotificationToast(newTree(), "test", newCfg())
	nt.SetVisible(false)
	buf := render.NewCommandBuffer()
	nt.Draw(buf)
	if len(buf.Overlays()) != 0 {
		t.Error("expected no overlays when not visible")
	}
}

func TestNotificationToastAllTypes(t *testing.T) {
	types := []MessageType{ToastInfo, ToastSuccess, ToastWarning, ToastError}
	for _, mt := range types {
		nt := NewNotificationToast(newTree(), "msg", newCfg())
		nt.SetToastType(mt)
		buf := render.NewCommandBuffer()
		nt.Draw(buf)
		if len(buf.Overlays()) == 0 {
			t.Errorf("expected overlays for toast type %d", mt)
		}
	}
}

// ============================================================
// CountdownTimer
// ============================================================

func TestCountdownTimerNilConfig(t *testing.T) {
	ct := NewCountdownTimer(newTree(), nil)
	if ct == nil {
		t.Fatal("expected non-nil")
	}
}

func TestCountdownTimerBasic(t *testing.T) {
	ct := NewCountdownTimer(newTree(), newCfg())
	ct.SetSeconds(65)
	ct.SetLabel("Time Left")
	ct.SetColor(uimath.ColorHex("#ffffff"))
	ct.SetFontSize(24)

	ct.Tick(10)
	if ct.Seconds() != 55 {
		t.Errorf("expected 55, got %g", ct.Seconds())
	}
	if ct.IsExpired() {
		t.Error("should not be expired")
	}
}

func TestCountdownTimerExpire(t *testing.T) {
	ct := NewCountdownTimer(newTree(), newCfg())
	ct.SetSeconds(1)
	expired := false
	ct.OnExpire(func() { expired = true })
	ct.Tick(2)
	if !expired {
		t.Error("expected expire callback")
	}
	if !ct.IsExpired() {
		t.Error("should be expired")
	}
	if ct.Seconds() != 0 {
		t.Error("seconds should be 0")
	}
}

func TestCountdownTimerTickAlreadyExpired(t *testing.T) {
	ct := NewCountdownTimer(newTree(), newCfg())
	ct.SetSeconds(0)
	ct.Tick(5) // no-op when already expired
	if ct.Seconds() != 0 {
		t.Error("should still be 0")
	}
}

func TestCountdownTimerExpireNoCallback(t *testing.T) {
	ct := NewCountdownTimer(newTree(), newCfg())
	ct.SetSeconds(1)
	ct.Tick(2) // no callback set
}

func TestCountdownTimerFormatTime(t *testing.T) {
	ct := NewCountdownTimer(newTree(), newCfg())

	ct.SetSeconds(65)
	got := ct.formatTime()
	if got != "1:05" {
		t.Errorf("expected '1:05', got %q", got)
	}

	ct.SetSeconds(9)
	got = ct.formatTime()
	if got != "9" {
		t.Errorf("expected '9', got %q", got)
	}

	ct.SetSeconds(0)
	got = ct.formatTime()
	if got != "0" {
		t.Errorf("expected '0', got %q", got)
	}

	ct.SetSeconds(120)
	got = ct.formatTime()
	if got != "2:00" {
		t.Errorf("expected '2:00', got %q", got)
	}
}

func TestCountdownTimerDrawNoBounds(t *testing.T) {
	ct := NewCountdownTimer(newTree(), newCfg())
	ct.SetSeconds(30)
	buf := render.NewCommandBuffer()
	ct.Draw(buf)
	// No bounds => early return
	if buf.Len() != 0 {
		t.Error("expected no commands without bounds")
	}
}

func TestCountdownTimerDrawWithBounds(t *testing.T) {
	tree := newTree()
	ct := NewCountdownTimer(tree, newCfg())
	setBounds(tree, ct, 0, 0, 200, 100)
	ct.SetSeconds(30)
	buf := render.NewCommandBuffer()
	ct.Draw(buf)
	// No TextRenderer, so no visible output
}

func TestCountdownTimerDrawLowTime(t *testing.T) {
	tree := newTree()
	ct := NewCountdownTimer(tree, newCfg())
	setBounds(tree, ct, 0, 0, 200, 100)
	ct.SetSeconds(5) // < 10, flash red
	buf := render.NewCommandBuffer()
	ct.Draw(buf)
}

func TestPad2(t *testing.T) {
	if pad2(5) != "05" {
		t.Errorf("pad2(5) = %q, want '05'", pad2(5))
	}
	if pad2(10) != "10" {
		t.Errorf("pad2(10) = %q, want '10'", pad2(10))
	}
	if pad2(0) != "00" {
		t.Errorf("pad2(0) = %q, want '00'", pad2(0))
	}
}

// ============================================================
// CurrencyDisplay
// ============================================================

func TestCurrencyDisplayNilConfig(t *testing.T) {
	cd := NewCurrencyDisplay(newTree(), nil)
	if cd == nil {
		t.Fatal("expected non-nil")
	}
}

func TestCurrencyDisplayAddSetClear(t *testing.T) {
	cd := NewCurrencyDisplay(newTree(), newCfg())
	cd.AddCurrency(CurrencyEntry{Symbol: "G", Amount: 1500})
	cd.AddCurrency(CurrencyEntry{Symbol: "S", Amount: 45})
	if len(cd.Currencies()) != 2 {
		t.Errorf("expected 2, got %d", len(cd.Currencies()))
	}

	cd.SetAmount(0, 2000)
	if cd.Currencies()[0].Amount != 2000 {
		t.Error("expected updated amount")
	}

	// Out of bounds
	cd.SetAmount(-1, 999)
	cd.SetAmount(99, 999)

	cd.SetGap(16)
	cd.ClearCurrencies()
	if len(cd.Currencies()) != 0 {
		t.Error("expected 0 after clear")
	}
}

func TestCurrencyDisplayDrawNoBounds(t *testing.T) {
	cd := NewCurrencyDisplay(newTree(), newCfg())
	cd.AddCurrency(CurrencyEntry{Symbol: "G", Amount: 100})
	buf := render.NewCommandBuffer()
	cd.Draw(buf)
	// No bounds => early return
}

func TestCurrencyDisplayDrawWithBounds(t *testing.T) {
	tree := newTree()
	cd := NewCurrencyDisplay(tree, newCfg())
	setBounds(tree, cd, 0, 0, 300, 40)
	cd.AddCurrency(CurrencyEntry{Symbol: "G", Amount: 100, Color: uimath.ColorHex("#ffd700")})
	cd.AddCurrency(CurrencyEntry{Symbol: "S", Amount: 45}) // default color (A == 0)
	buf := render.NewCommandBuffer()
	cd.Draw(buf)
	// No TextRenderer, so falls through to else branch
}

func TestFormatAmount(t *testing.T) {
	tests := []struct {
		in   int
		want string
	}{
		{0, "0"},
		{999, "999"},
		{1000, "1.0K"},
		{1500, "1.5K"},
		{999999, "999.9K"},
		{1000000, "1.0M"},
		{2500000, "2.5M"},
	}
	for _, tt := range tests {
		got := formatAmount(tt.in)
		if got != tt.want {
			t.Errorf("formatAmount(%d) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

// ============================================================
// DialogueBox
// ============================================================

func TestDialogueBoxNilConfig(t *testing.T) {
	db := NewDialogueBox(newTree(), nil)
	if db == nil {
		t.Fatal("expected non-nil")
	}
}

func TestDialogueBoxShowHide(t *testing.T) {
	db := NewDialogueBox(newTree(), newCfg())
	db.Show("Guard", "Halt!")
	if !db.IsVisible() {
		t.Error("should be visible")
	}
	if db.Speaker() != "Guard" {
		t.Error("expected 'Guard'")
	}
	if db.Text() != "Halt!" {
		t.Error("expected 'Halt!'")
	}
	db.Hide()
	if db.IsVisible() {
		t.Error("should be hidden")
	}
}

func TestDialogueBoxChoices(t *testing.T) {
	db := NewDialogueBox(newTree(), newCfg())
	db.SetChoices([]DialogueChoice{
		{Text: "Yes"},
		{Text: "No"},
	})
	if len(db.Choices()) != 2 {
		t.Errorf("expected 2, got %d", len(db.Choices()))
	}
	db.ClearChoices()
	if len(db.Choices()) != 0 {
		t.Error("expected 0 after clear")
	}
}

func TestDialogueBoxSetters(t *testing.T) {
	db := NewDialogueBox(newTree(), newCfg())
	db.SetSpeaker("NPC")
	db.SetText("Hello")
	db.SetPortrait(0)
	db.SetSize(600, 200)
	db.OnAdvance(func() {})
}

func TestDialogueBoxDrawNotVisible(t *testing.T) {
	db := NewDialogueBox(newTree(), newCfg())
	buf := render.NewCommandBuffer()
	db.Draw(buf)
	if len(buf.Overlays()) != 0 {
		t.Error("expected no overlays when not visible")
	}
}

func TestDialogueBoxDrawVisible(t *testing.T) {
	db := NewDialogueBox(newTree(), newCfg())
	db.Show("Guard", "Halt!")
	buf := render.NewCommandBuffer()
	db.Draw(buf)
	if len(buf.Overlays()) == 0 {
		t.Error("expected overlay commands")
	}
}

func TestDialogueBoxDrawWithBounds(t *testing.T) {
	tree := newTree()
	db := NewDialogueBox(tree, newCfg())
	setBounds(tree, db, 10, 10, 500, 160)
	db.Show("NPC", "Welcome!")
	buf := render.NewCommandBuffer()
	db.Draw(buf)
	if len(buf.Overlays()) == 0 {
		t.Error("expected overlay commands with bounds")
	}
}

func TestDialogueBoxDrawWithChoices(t *testing.T) {
	db := NewDialogueBox(newTree(), newCfg())
	db.Show("NPC", "Choose:")
	db.SetChoices([]DialogueChoice{
		{Text: "Option A"},
		{Text: "Option B"},
	})
	buf := render.NewCommandBuffer()
	db.Draw(buf)
	if len(buf.Overlays()) == 0 {
		t.Error("expected overlays")
	}
}

// ============================================================
// HUD
// ============================================================

func TestHUDNilConfig(t *testing.T) {
	h := NewHUD(newTree(), nil)
	if h == nil {
		t.Fatal("expected non-nil HUD")
	}
}

func TestHUDAddElementAndDraw(t *testing.T) {
	tree := newTree()
	h := NewHUD(tree, newCfg())

	hb := NewHealthBar(tree, newCfg())
	h.AddElement(hb, AnchorTopLeft, 10, 10)

	buf := render.NewCommandBuffer()
	h.Draw(buf)
	// HealthBar draws with fallback bounds
	if buf.Len() == 0 {
		t.Error("expected commands from HUD Draw")
	}
}

func TestHUDLayoutElements(t *testing.T) {
	tree := newTree()
	h := NewHUD(tree, newCfg())

	hb := NewHealthBar(tree, newCfg())
	// Test all anchor positions
	h.AddElement(hb, AnchorTopLeft, 0, 0)
	h.AddElement(hb, AnchorTopCenter, 0, 0)
	h.AddElement(hb, AnchorTopRight, 0, 0)
	h.AddElement(hb, AnchorMiddleLeft, 0, 0)
	h.AddElement(hb, AnchorMiddleCenter, 0, 0)
	h.AddElement(hb, AnchorMiddleRight, 0, 0)
	h.AddElement(hb, AnchorBottomLeft, 0, 0)
	h.AddElement(hb, AnchorBottomCenter, 0, 0)
	h.AddElement(hb, AnchorBottomRight, 0, 0)

	h.LayoutElements(1920, 1080)
}

// ============================================================
// HealthBar
// ============================================================

func TestHealthBarNilConfig(t *testing.T) {
	hb := NewHealthBar(newTree(), nil)
	if hb == nil {
		t.Fatal("expected non-nil")
	}
}

func TestHealthBarRatio(t *testing.T) {
	hb := NewHealthBar(newTree(), newCfg())

	hb.SetCurrent(75)
	hb.SetMax(100)
	if hb.Ratio() != 0.75 {
		t.Errorf("expected 0.75, got %g", hb.Ratio())
	}

	if hb.Current() != 75 {
		t.Error("expected current 75")
	}
	if hb.Max() != 100 {
		t.Error("expected max 100")
	}

	hb.SetCurrent(0)
	if hb.Ratio() != 0 {
		t.Error("expected 0")
	}

	hb.SetCurrent(150)
	if hb.Ratio() != 1 {
		t.Error("expected clamped to 1")
	}

	hb.SetMax(0)
	if hb.Ratio() != 0 {
		t.Error("expected 0 with max=0")
	}

	hb.SetMax(-1)
	if hb.Ratio() != 0 {
		t.Error("expected 0 with max=-1")
	}

	hb.SetMax(100)
	hb.SetCurrent(-10)
	if hb.Ratio() != 0 {
		t.Error("expected 0 with negative current")
	}
}

func TestHealthBarSetters(t *testing.T) {
	hb := NewHealthBar(newTree(), newCfg())
	hb.SetBarColor(uimath.ColorHex("#ff0000"))
	hb.SetBgColor(uimath.ColorHex("#333333"))
	hb.SetShowText(true)
	hb.SetSize(300, 30)
}

func TestHealthBarDrawNoBounds(t *testing.T) {
	hb := NewHealthBar(newTree(), newCfg())
	buf := render.NewCommandBuffer()
	hb.Draw(buf)
	// Uses fallback bounds
	if buf.Len() == 0 {
		t.Error("expected commands with fallback bounds")
	}
}

func TestHealthBarDrawWithBounds(t *testing.T) {
	tree := newTree()
	hb := NewHealthBar(tree, newCfg())
	setBounds(tree, hb, 0, 0, 200, 20)
	hb.SetCurrent(50)
	hb.SetMax(100)
	buf := render.NewCommandBuffer()
	hb.Draw(buf)
	if buf.Len() < 2 {
		t.Error("expected at least background + fill")
	}
}

func TestHealthBarDrawZeroRatio(t *testing.T) {
	tree := newTree()
	hb := NewHealthBar(tree, newCfg())
	setBounds(tree, hb, 0, 0, 200, 20)
	hb.SetCurrent(0)
	hb.SetMax(100)
	buf := render.NewCommandBuffer()
	hb.Draw(buf)
	// Only background, no fill
}

// ============================================================
// Hotbar
// ============================================================

func TestHotbarNilConfig(t *testing.T) {
	hb := NewHotbar(newTree(), 4, nil)
	if hb == nil {
		t.Fatal("expected non-nil")
	}
}

func TestHotbarBasic(t *testing.T) {
	hb := NewHotbar(newTree(), 8, newCfg())
	if hb.SlotCount() != 8 {
		t.Errorf("expected 8, got %d", hb.SlotCount())
	}
	if hb.Selected() != -1 {
		t.Error("expected -1 initial selection")
	}

	hb.SetSlot(0, HotbarSlot{Label: "Sword", Available: true, Keybind: "1"})
	slot := hb.GetSlot(0)
	if slot.Label != "Sword" {
		t.Error("expected Sword")
	}

	// Out of bounds
	hb.SetSlot(-1, HotbarSlot{})
	hb.SetSlot(99, HotbarSlot{})
	empty := hb.GetSlot(-1)
	if empty.Label != "" {
		t.Error("expected empty slot for out of bounds")
	}
	empty = hb.GetSlot(99)
	if empty.Label != "" {
		t.Error("expected empty slot for out of bounds")
	}

	hb.SetSelected(2)
	if hb.Selected() != 2 {
		t.Error("expected 2")
	}
	hb.SetSlotSize(64)
	hb.SetGap(8)
}

func TestHotbarDrawNoBounds(t *testing.T) {
	hb := NewHotbar(newTree(), 4, newCfg())
	buf := render.NewCommandBuffer()
	hb.Draw(buf)
	// No bounds => early return
}

func TestHotbarDrawWithBounds(t *testing.T) {
	tree := newTree()
	hb := NewHotbar(tree, 4, newCfg())
	setBounds(tree, hb, 0, 0, 400, 60)
	hb.SetSelected(1)
	hb.SetSlot(2, HotbarSlot{Cooldown: 0.5, Available: true})
	buf := render.NewCommandBuffer()
	hb.Draw(buf)
	if buf.Len() == 0 {
		t.Error("expected commands")
	}
}

// ============================================================
// CooldownMask
// ============================================================

func TestCooldownMaskNilConfig(t *testing.T) {
	cm := NewCooldownMask(newTree(), nil)
	if cm == nil {
		t.Fatal("expected non-nil")
	}
}

func TestCooldownMaskDraw(t *testing.T) {
	tree := newTree()
	cm := NewCooldownMask(tree, newCfg())
	cm.SetRatio(0.5)
	if cm.Ratio() != 0.5 {
		t.Error("expected 0.5")
	}

	// No bounds => no draw
	buf := render.NewCommandBuffer()
	cm.Draw(buf)
	if buf.Len() != 0 {
		t.Error("expected no commands without bounds")
	}

	// With bounds
	setBounds(tree, cm, 0, 0, 48, 48)
	buf = render.NewCommandBuffer()
	cm.Draw(buf)
	if buf.Len() == 0 {
		t.Error("expected commands with bounds and ratio > 0")
	}
}

func TestCooldownMaskDrawZeroRatio(t *testing.T) {
	tree := newTree()
	cm := NewCooldownMask(tree, newCfg())
	cm.SetRatio(0)
	setBounds(tree, cm, 0, 0, 48, 48)
	buf := render.NewCommandBuffer()
	cm.Draw(buf)
	// ratio <= 0, early return
}

// ============================================================
// Inventory
// ============================================================

func TestInventoryNilConfig(t *testing.T) {
	inv := NewInventory(newTree(), 4, 6, nil)
	if inv == nil {
		t.Fatal("expected non-nil")
	}
}

func TestInventoryBasic(t *testing.T) {
	inv := NewInventory(newTree(), 4, 6, newCfg())
	if inv.Rows() != 4 {
		t.Error("expected 4 rows")
	}
	if inv.Cols() != 6 {
		t.Error("expected 6 cols")
	}

	inv.SetTitle("Bag")
	if inv.Title() != "Bag" {
		t.Error("expected 'Bag'")
	}

	item := &ItemData{ID: "sword", Name: "Iron Sword", Quantity: 1, Rarity: RarityCommon}
	inv.SetItem(0, item)
	got := inv.GetItem(0)
	if got == nil || got.Name != "Iron Sword" {
		t.Error("expected item in slot 0")
	}

	// Out of bounds
	inv.SetItem(-1, item)
	inv.SetItem(999, item)

	removed := inv.RemoveItem(0)
	if removed == nil {
		t.Error("expected removed item")
	}
	if inv.GetItem(0) != nil {
		t.Error("should be empty")
	}

	// Remove from empty slot
	removed = inv.RemoveItem(0)
	if removed != nil {
		t.Error("expected nil from empty slot")
	}

	inv.SetItem(1, item)
	inv.ClearAll()
	if inv.GetItem(1) != nil {
		t.Error("should be empty after clear")
	}
}

func TestInventorySetters(t *testing.T) {
	inv := NewInventory(newTree(), 2, 3, newCfg())
	inv.SetSlotSize(64)
	inv.SetGap(8)
	inv.OnDrop(func(i int, d any) {})
	inv.OnSelect(func(i int, item *ItemData) {})
}

func TestInventoryDrawNoBounds(t *testing.T) {
	inv := NewInventory(newTree(), 2, 3, newCfg())
	buf := render.NewCommandBuffer()
	inv.Draw(buf)
	// No bounds => early return
}

func TestInventoryDrawWithBounds(t *testing.T) {
	tree := newTree()
	inv := NewInventory(tree, 2, 3, newCfg())
	setBounds(tree, inv, 0, 0, 300, 200)
	inv.SetTitle("Inventory")

	inv.SetItem(0, &ItemData{Name: "Sword", Rarity: RarityRare, Quantity: 1, Icon: 0})
	inv.SetItem(1, &ItemData{Name: "Shield", Rarity: RarityEpic, Quantity: 5, Icon: 0})

	buf := render.NewCommandBuffer()
	inv.Draw(buf)
	if buf.Len() == 0 {
		t.Error("expected commands")
	}
}

func TestInventoryDrawNoTitle(t *testing.T) {
	tree := newTree()
	inv := NewInventory(tree, 2, 3, newCfg())
	setBounds(tree, inv, 0, 0, 300, 200)
	buf := render.NewCommandBuffer()
	inv.Draw(buf)
	if buf.Len() == 0 {
		t.Error("expected commands")
	}
}

func TestRarityColor(t *testing.T) {
	rarities := []ItemRarity{RarityCommon, RarityUncommon, RarityRare, RarityEpic, RarityLegendary}
	for _, r := range rarities {
		c := rarityColor(r)
		if c.A == 0 {
			t.Errorf("rarity %d should have alpha", r)
		}
	}
}

// ============================================================
// LootWindow
// ============================================================

func TestLootWindowNilConfig(t *testing.T) {
	lw := NewLootWindow(newTree(), nil)
	if lw == nil {
		t.Fatal("expected non-nil")
	}
}

func TestLootWindowBasic(t *testing.T) {
	lw := NewLootWindow(newTree(), newCfg())
	lw.AddItem(LootItem{Item: &ItemData{Name: "Gold", Rarity: RarityCommon}, Quantity: 10})
	lw.AddItem(LootItem{Item: &ItemData{Name: "Sword", Rarity: RarityEpic}, Quantity: 1})
	if len(lw.Items()) != 2 {
		t.Errorf("expected 2, got %d", len(lw.Items()))
	}

	lw.SetTitle("Treasure")
	if lw.Title() != "Treasure" {
		t.Error("expected 'Treasure'")
	}
	lw.SetWidth(300)
}

func TestLootWindowOpenClose(t *testing.T) {
	lw := NewLootWindow(newTree(), newCfg())
	lw.Open()
	if !lw.IsVisible() {
		t.Error("should be visible")
	}
	closed := false
	lw.OnClose(func() { closed = true })
	lw.Close()
	if lw.IsVisible() {
		t.Error("should be hidden")
	}
	if !closed {
		t.Error("expected close callback")
	}
}

func TestLootWindowCloseNoCallback(t *testing.T) {
	lw := NewLootWindow(newTree(), newCfg())
	lw.Open()
	lw.Close() // no callback set
}

func TestLootWindowLootItem(t *testing.T) {
	lw := NewLootWindow(newTree(), newCfg())
	lw.AddItem(LootItem{Item: &ItemData{Name: "Gold"}, Quantity: 10})
	lw.AddItem(LootItem{Item: &ItemData{Name: "Sword"}, Quantity: 1})

	looted := -1
	lw.OnLoot(func(i int) { looted = i })
	lw.LootItem(0)
	if looted != 0 {
		t.Error("expected loot callback for index 0")
	}
	if !lw.Items()[0].Claimed {
		t.Error("item should be claimed")
	}

	// Loot already claimed item
	looted = -1
	lw.LootItem(0)
	if looted != -1 {
		t.Error("should not loot already claimed")
	}

	// Out of bounds
	lw.LootItem(-1)
	lw.LootItem(99)
}

func TestLootWindowLootItemNoCallback(t *testing.T) {
	lw := NewLootWindow(newTree(), newCfg())
	lw.AddItem(LootItem{Item: &ItemData{Name: "Gold"}, Quantity: 10})
	lw.LootItem(0) // no callback set
}

func TestLootWindowLootAll(t *testing.T) {
	lw := NewLootWindow(newTree(), newCfg())
	lw.AddItem(LootItem{Item: &ItemData{Name: "A"}, Quantity: 1})
	lw.AddItem(LootItem{Item: &ItemData{Name: "B"}, Quantity: 1})
	lw.AddItem(LootItem{Item: &ItemData{Name: "C"}, Quantity: 1})

	count := 0
	lw.OnLoot(func(i int) { count++ })
	lw.LootAll()
	if count != 3 {
		t.Errorf("expected 3 loot callbacks, got %d", count)
	}
	for _, item := range lw.Items() {
		if !item.Claimed {
			t.Error("all items should be claimed")
		}
	}

	// LootAll again - no more callbacks since all claimed
	count = 0
	lw.LootAll()
	if count != 0 {
		t.Error("should not loot already claimed items")
	}
}

func TestLootWindowLootAllNoCallback(t *testing.T) {
	lw := NewLootWindow(newTree(), newCfg())
	lw.AddItem(LootItem{Item: &ItemData{Name: "A"}, Quantity: 1})
	lw.LootAll() // no callback set
}

func TestLootWindowClearItems(t *testing.T) {
	lw := NewLootWindow(newTree(), newCfg())
	lw.AddItem(LootItem{Item: &ItemData{Name: "A"}, Quantity: 1})
	lw.ClearItems()
	if len(lw.Items()) != 0 {
		t.Error("expected 0 after clear")
	}
}

func TestLootWindowDrawNotVisible(t *testing.T) {
	lw := NewLootWindow(newTree(), newCfg())
	lw.AddItem(LootItem{Item: &ItemData{Name: "A"}, Quantity: 1})
	buf := render.NewCommandBuffer()
	lw.Draw(buf) // not visible
	if len(buf.Overlays()) != 0 {
		t.Error("expected no overlays when not visible")
	}
}

func TestLootWindowDrawEmpty(t *testing.T) {
	lw := NewLootWindow(newTree(), newCfg())
	lw.Open()
	buf := render.NewCommandBuffer()
	lw.Draw(buf) // visible but empty
	if len(buf.Overlays()) != 0 {
		t.Error("expected no overlays with empty items")
	}
}

func TestLootWindowDrawVisible(t *testing.T) {
	lw := NewLootWindow(newTree(), newCfg())
	lw.AddItem(LootItem{Item: &ItemData{Name: "Gold", Rarity: RarityCommon}, Quantity: 10})
	lw.AddItem(LootItem{Item: &ItemData{Name: "Sword", Rarity: RarityEpic}, Quantity: 1, Claimed: true})
	lw.Open()
	buf := render.NewCommandBuffer()
	lw.Draw(buf)
	if len(buf.Overlays()) == 0 {
		t.Error("expected overlays from loot window")
	}
}

func TestLootWindowDrawWithBounds(t *testing.T) {
	tree := newTree()
	lw := NewLootWindow(tree, newCfg())
	setBounds(tree, lw, 10, 10, 220, 300)
	lw.AddItem(LootItem{Item: &ItemData{Name: "Gold"}, Quantity: 10})
	lw.Open()
	buf := render.NewCommandBuffer()
	lw.Draw(buf)
	if len(buf.Overlays()) == 0 {
		t.Error("expected overlays")
	}
}

// ============================================================
// Minimap
// ============================================================

func TestMinimapNilConfig(t *testing.T) {
	m := NewMinimap(newTree(), nil)
	if m == nil {
		t.Fatal("expected non-nil")
	}
}

func TestMinimapBasic(t *testing.T) {
	m := NewMinimap(newTree(), newCfg())
	m.SetPlayerPos(100, 200)
	m.SetZoom(2)
	m.SetSize(200)
	m.SetCircular(false)
	m.SetTexture(0)
	m.SetPlayerRotation(1.5)

	m.AddMarker(MinimapMarker{X: 110, Y: 210, Label: "NPC"})
	if len(m.Markers()) != 1 {
		t.Error("expected 1 marker")
	}
	m.ClearMarkers()
	if len(m.Markers()) != 0 {
		t.Error("expected 0 markers")
	}
}

func TestMinimapDrawNoBounds(t *testing.T) {
	m := NewMinimap(newTree(), newCfg())
	m.SetPlayerPos(100, 200)
	m.AddMarker(MinimapMarker{X: 110, Y: 210, Label: "NPC", Size: 6})
	buf := render.NewCommandBuffer()
	m.Draw(buf)
	// Uses fallback bounds
	if len(buf.Overlays()) == 0 {
		t.Error("expected overlays")
	}
}

func TestMinimapDrawWithBounds(t *testing.T) {
	tree := newTree()
	m := NewMinimap(tree, newCfg())
	setBounds(tree, m, 0, 0, 200, 200)
	m.SetPlayerPos(100, 200)
	m.AddMarker(MinimapMarker{X: 110, Y: 210, Label: "NPC", Size: 0}) // size 0 -> default 4
	buf := render.NewCommandBuffer()
	m.Draw(buf)
	if len(buf.Overlays()) == 0 {
		t.Error("expected overlays")
	}
}

func TestMinimapDrawSquare(t *testing.T) {
	m := NewMinimap(newTree(), newCfg())
	m.SetCircular(false)
	buf := render.NewCommandBuffer()
	m.Draw(buf)
	if len(buf.Overlays()) == 0 {
		t.Error("expected overlays for square minimap")
	}
}

func TestMinimapDrawMarkerOutOfBounds(t *testing.T) {
	tree := newTree()
	m := NewMinimap(tree, newCfg())
	setBounds(tree, m, 0, 0, 180, 180)
	m.SetPlayerPos(0, 0)
	m.SetZoom(1)
	// Marker far away - should be clipped
	m.AddMarker(MinimapMarker{X: 9999, Y: 9999, Label: "Far"})
	buf := render.NewCommandBuffer()
	m.Draw(buf)
	// Background + border + player dot, but marker should be clipped
}

// ============================================================
// Nameplate
// ============================================================

func TestNameplateNilConfig(t *testing.T) {
	np := NewNameplate(newTree(), "Test", nil)
	if np == nil {
		t.Fatal("expected non-nil")
	}
}

func TestNameplateBasic(t *testing.T) {
	np := NewNameplate(newTree(), "Dragon", newCfg())
	if np.Name() != "Dragon" {
		t.Error("expected 'Dragon'")
	}
	np.SetName("Boss")
	np.SetTitle("The Destroyer")
	np.SetLevel(50)
	np.SetHP(50, 100)
	np.SetType(NameplateHostile)
	np.SetPosition(200, 100)
	np.SetBarSize(120, 8)
	np.SetVisible(false)
	if np.IsVisible() {
		t.Error("should be hidden")
	}
	np.SetVisible(true)
}

func TestNameplateDrawNotVisible(t *testing.T) {
	np := NewNameplate(newTree(), "Test", newCfg())
	np.SetVisible(false)
	buf := render.NewCommandBuffer()
	np.Draw(buf)
	if len(buf.Overlays()) != 0 {
		t.Error("expected no overlays when not visible")
	}
}

func TestNameplateDrawVisible(t *testing.T) {
	np := NewNameplate(newTree(), "Dragon", newCfg())
	np.SetHP(50, 100)
	np.SetType(NameplateHostile)
	np.SetPosition(200, 100)
	buf := render.NewCommandBuffer()
	np.Draw(buf)
	if len(buf.Overlays()) == 0 {
		t.Error("expected overlays")
	}
}

func TestNameplateDrawAllTypes(t *testing.T) {
	types := []NameplateType{NameplateFriendly, NameplateHostile, NameplateNeutral, NameplatePlayer}
	for _, nt := range types {
		np := NewNameplate(newTree(), "Test", newCfg())
		np.SetType(nt)
		np.SetHP(50, 100)
		buf := render.NewCommandBuffer()
		np.Draw(buf)
		if len(buf.Overlays()) == 0 {
			t.Errorf("expected overlays for type %d", nt)
		}
	}
}

func TestNameplateHPRatioEdgeCases(t *testing.T) {
	np := NewNameplate(newTree(), "Test", newCfg())

	np.SetHP(0, 0)
	buf := render.NewCommandBuffer()
	np.Draw(buf) // hpMax <= 0, ratio = 0, no fill bar

	np.SetHP(-10, 100)
	buf = render.NewCommandBuffer()
	np.Draw(buf) // negative hp, clamped to 0

	np.SetHP(200, 100)
	buf = render.NewCommandBuffer()
	np.Draw(buf) // over max, clamped to 1
}

func TestNameplateHPColors(t *testing.T) {
	// Low HP (< 30%) -> red
	np := NewNameplate(newTree(), "Test", newCfg())
	np.SetHP(20, 100)
	buf := render.NewCommandBuffer()
	np.Draw(buf)

	// Medium HP (30-60%) -> yellow
	np.SetHP(45, 100)
	buf = render.NewCommandBuffer()
	np.Draw(buf)

	// High HP (> 60%) -> green
	np.SetHP(80, 100)
	buf = render.NewCommandBuffer()
	np.Draw(buf)
}

func TestNameplateColor(t *testing.T) {
	types := []NameplateType{NameplateFriendly, NameplateHostile, NameplateNeutral, NameplatePlayer}
	for _, nt := range types {
		c := nameplateColor(nt)
		if c.A == 0 {
			t.Errorf("nameplateColor(%d) should have alpha", nt)
		}
	}
}

// ============================================================
// QuestTracker
// ============================================================

func TestQuestTrackerNilConfig(t *testing.T) {
	qt := NewQuestTracker(newTree(), nil)
	if qt == nil {
		t.Fatal("expected non-nil")
	}
}

func TestQuestTrackerBasic(t *testing.T) {
	qt := NewQuestTracker(newTree(), newCfg())
	qt.AddQuest(Quest{
		Title:  "Slay the Dragon",
		Active: true,
		Objectives: []QuestObjective{
			{Text: "Kill dragons", Current: 2, Required: 5},
		},
	})
	qt.AddQuest(Quest{
		Title:  "Find the Key",
		Active: false,
		Objectives: []QuestObjective{
			{Text: "Search dungeon", Completed: true},
		},
	})
	if len(qt.Quests()) != 2 {
		t.Error("expected 2 quests")
	}

	qt.RemoveQuest(0)
	if len(qt.Quests()) != 1 {
		t.Error("expected 1 quest")
	}

	// Remove out of bounds
	qt.RemoveQuest(-1)
	qt.RemoveQuest(99)
	if len(qt.Quests()) != 1 {
		t.Error("expected still 1 quest")
	}

	qt.ClearQuests()
	if len(qt.Quests()) != 0 {
		t.Error("expected 0")
	}
}

func TestQuestTrackerSetters(t *testing.T) {
	qt := NewQuestTracker(newTree(), newCfg())
	qt.SetWidth(300)
	qt.SetMaxQuests(10)
}

func TestQuestTrackerDrawNoBounds(t *testing.T) {
	qt := NewQuestTracker(newTree(), newCfg())
	qt.AddQuest(Quest{
		Title:  "Quest",
		Active: true,
		Objectives: []QuestObjective{
			{Text: "Obj", Current: 1, Required: 3},
			{Text: "Done", Completed: true},
		},
	})
	buf := render.NewCommandBuffer()
	qt.Draw(buf)
	// Uses fallback bounds, no TextRenderer -> uses else branches
}

func TestQuestTrackerDrawWithBounds(t *testing.T) {
	tree := newTree()
	qt := NewQuestTracker(tree, newCfg())
	setBounds(tree, qt, 0, 0, 250, 400)
	qt.AddQuest(Quest{
		Title:  "Active Quest",
		Active: true,
		Objectives: []QuestObjective{
			{Text: "Kill", Current: 2, Required: 5},
			{Text: "Collect", Completed: true, Required: 0},
		},
	})
	qt.AddQuest(Quest{
		Title:  "Inactive",
		Active: false,
	})
	buf := render.NewCommandBuffer()
	qt.Draw(buf)
}

func TestQuestTrackerDrawMaxQuests(t *testing.T) {
	tree := newTree()
	qt := NewQuestTracker(tree, newCfg())
	setBounds(tree, qt, 0, 0, 250, 400)
	qt.SetMaxQuests(2)
	for i := 0; i < 5; i++ {
		qt.AddQuest(Quest{Title: "Q" + itoa(i), Active: true})
	}
	buf := render.NewCommandBuffer()
	qt.Draw(buf)
}

// ============================================================
// RadialMenu
// ============================================================

func TestRadialMenuNilConfig(t *testing.T) {
	rm := NewRadialMenu(newTree(), nil)
	if rm == nil {
		t.Fatal("expected non-nil")
	}
}

func TestRadialMenuBasic(t *testing.T) {
	rm := NewRadialMenu(newTree(), newCfg())
	rm.AddItem(RadialMenuItem{Label: "Attack"})
	rm.AddItem(RadialMenuItem{Label: "Defend"})
	rm.AddItem(RadialMenuItem{Label: "Flee", Disabled: true})
	if len(rm.Items()) != 3 {
		t.Error("expected 3 items")
	}

	rm.Show(200, 200)
	if !rm.IsVisible() {
		t.Error("should be visible")
	}
	if rm.Hovered() != -1 {
		t.Error("expected -1 hovered after show")
	}

	rm.SetHovered(1)
	if rm.Hovered() != 1 {
		t.Error("expected 1")
	}

	rm.Hide()
	if rm.IsVisible() {
		t.Error("should be hidden")
	}
}

func TestRadialMenuSetters(t *testing.T) {
	rm := NewRadialMenu(newTree(), newCfg())
	rm.SetRadius(150)
	rm.SetInnerRadius(50)
	rm.OnSelect(func(i int) {})
}

func TestRadialMenuClearItems(t *testing.T) {
	rm := NewRadialMenu(newTree(), newCfg())
	rm.AddItem(RadialMenuItem{Label: "X"})
	rm.ClearItems()
	if len(rm.Items()) != 0 {
		t.Error("expected 0")
	}
}

func TestRadialMenuSelect(t *testing.T) {
	rm := NewRadialMenu(newTree(), newCfg())

	clicked := false
	rm.AddItem(RadialMenuItem{Label: "Attack", OnClick: func() { clicked = true }})
	rm.AddItem(RadialMenuItem{Label: "Disabled", Disabled: true})

	selected := -1
	rm.OnSelect(func(i int) { selected = i })

	rm.Show(200, 200)
	rm.Select(0)
	if !clicked {
		t.Error("expected OnClick callback")
	}
	if selected != 0 {
		t.Error("expected OnSelect callback with 0")
	}
	if rm.IsVisible() {
		t.Error("should be hidden after select")
	}

	// Select disabled item
	rm.Show(200, 200)
	selected = -1
	rm.Select(1)
	if selected != -1 {
		t.Error("should not select disabled item")
	}

	// Out of bounds
	rm.Show(200, 200)
	rm.Select(-1)
	rm.Select(99)
}

func TestRadialMenuSelectNoCallbacks(t *testing.T) {
	rm := NewRadialMenu(newTree(), newCfg())
	rm.AddItem(RadialMenuItem{Label: "Attack"}) // no OnClick
	rm.Show(200, 200)
	rm.Select(0) // should not panic
}

func TestRadialMenuDrawNotVisible(t *testing.T) {
	rm := NewRadialMenu(newTree(), newCfg())
	rm.AddItem(RadialMenuItem{Label: "X"})
	buf := render.NewCommandBuffer()
	rm.Draw(buf)
	if len(buf.Overlays()) != 0 {
		t.Error("expected no overlays when not visible")
	}
}

func TestRadialMenuDrawEmpty(t *testing.T) {
	rm := NewRadialMenu(newTree(), newCfg())
	rm.Show(200, 200)
	buf := render.NewCommandBuffer()
	rm.Draw(buf)
	if len(buf.Overlays()) != 0 {
		t.Error("expected no overlays when empty")
	}
}

func TestRadialMenuDrawVisible(t *testing.T) {
	rm := NewRadialMenu(newTree(), newCfg())
	rm.AddItem(RadialMenuItem{Label: "Attack"})
	rm.AddItem(RadialMenuItem{Label: "Defend"})
	rm.AddItem(RadialMenuItem{Label: "Flee", Disabled: true})
	rm.Show(200, 200)
	rm.SetHovered(0)

	buf := render.NewCommandBuffer()
	rm.Draw(buf)
	if len(buf.Overlays()) == 0 {
		t.Error("expected overlays")
	}
}

func TestSinCosApprox(t *testing.T) {
	// sinApprox(0) should be ~0
	if s := sinApprox(0); s < -0.01 || s > 0.01 {
		t.Errorf("sinApprox(0) = %g, want ~0", s)
	}
	// cosApprox(0) should be ~1
	if c := cosApprox(0); c < 0.9 || c > 1.1 {
		t.Errorf("cosApprox(0) = %g, want ~1", c)
	}
	// sinApprox(PI/2) should be ~1
	if s := sinApprox(1.5708); s < 0.9 || s > 1.1 {
		t.Errorf("sinApprox(PI/2) = %g, want ~1", s)
	}
	// Test large values (normalization)
	_ = sinApprox(100)
	_ = sinApprox(-100)
}

// ============================================================
// Scoreboard
// ============================================================

func TestScoreboardNilConfig(t *testing.T) {
	sb := NewScoreboard(newTree(), nil)
	if sb == nil {
		t.Fatal("expected non-nil")
	}
}

func TestScoreboardBasic(t *testing.T) {
	sb := NewScoreboard(newTree(), newCfg())
	sb.AddEntry(ScoreEntry{Name: "Alice", Score: 100, Kills: 5, Deaths: 2})
	sb.AddEntry(ScoreEntry{Name: "Bob", Score: 200, Kills: 8, Deaths: 1})
	if len(sb.Entries()) != 2 {
		t.Error("expected 2 entries")
	}

	sb.SortByScore()
	if sb.Entries()[0].Name != "Bob" {
		t.Error("expected Bob first")
	}

	sb.SetTitle("Match Results")
	sb.SetWidth(500)
	sb.SetVisible(true)
	if !sb.IsVisible() {
		t.Error("should be visible")
	}

	sb.ClearEntries()
	if len(sb.Entries()) != 0 {
		t.Error("expected 0 after clear")
	}
}

func TestScoreboardSortEmpty(t *testing.T) {
	sb := NewScoreboard(newTree(), newCfg())
	sb.SortByScore() // should not panic
}

func TestScoreboardSortSingle(t *testing.T) {
	sb := NewScoreboard(newTree(), newCfg())
	sb.AddEntry(ScoreEntry{Name: "Solo", Score: 100})
	sb.SortByScore()
	if sb.Entries()[0].Name != "Solo" {
		t.Error("expected Solo")
	}
}

func TestScoreboardDrawNotVisible(t *testing.T) {
	sb := NewScoreboard(newTree(), newCfg())
	sb.AddEntry(ScoreEntry{Name: "A", Score: 100})
	buf := render.NewCommandBuffer()
	sb.Draw(buf)
	if len(buf.Overlays()) != 0 {
		t.Error("expected no overlays when not visible")
	}
}

func TestScoreboardDrawEmpty(t *testing.T) {
	sb := NewScoreboard(newTree(), newCfg())
	sb.SetVisible(true)
	buf := render.NewCommandBuffer()
	sb.Draw(buf)
	if len(buf.Overlays()) != 0 {
		t.Error("expected no overlays when empty")
	}
}

func TestScoreboardDrawVisible(t *testing.T) {
	sb := NewScoreboard(newTree(), newCfg())
	sb.AddEntry(ScoreEntry{Name: "Alice", Score: 100, Kills: 5, Deaths: 2})
	sb.AddEntry(ScoreEntry{Name: "Bob", Score: 200, Kills: 8, Deaths: 1})
	sb.AddEntry(ScoreEntry{Name: "Charlie", Score: 50, Kills: 2, Deaths: 5})
	sb.SetVisible(true)
	buf := render.NewCommandBuffer()
	sb.Draw(buf)
	if len(buf.Overlays()) == 0 {
		t.Error("expected overlays")
	}
}

func TestScoreboardDrawWithBounds(t *testing.T) {
	tree := newTree()
	sb := NewScoreboard(tree, newCfg())
	setBounds(tree, sb, 10, 10, 400, 300)
	sb.AddEntry(ScoreEntry{Name: "A", Score: 100})
	sb.SetVisible(true)
	buf := render.NewCommandBuffer()
	sb.Draw(buf)
	if len(buf.Overlays()) == 0 {
		t.Error("expected overlays with bounds")
	}
}

// ============================================================
// SkillTree
// ============================================================

func TestSkillTreeNilConfig(t *testing.T) {
	st := NewSkillTree(newTree(), nil)
	if st == nil {
		t.Fatal("expected non-nil")
	}
}

func TestSkillTreeBasic(t *testing.T) {
	st := NewSkillTree(newTree(), newCfg())
	st.SetPoints(5)
	if st.Points() != 5 {
		t.Error("expected 5 points")
	}

	st.AddNode(&SkillNode{ID: "fireball", Name: "Fireball", State: SkillAvailable, Level: 0, MaxLevel: 3, Cost: 1})
	st.AddNode(&SkillNode{ID: "meteor", Name: "Meteor", State: SkillLocked, Level: 0, MaxLevel: 1, Cost: 3, Requires: []string{"fireball"}})

	if len(st.Nodes()) != 2 {
		t.Error("expected 2 nodes")
	}

	fb := st.FindNode("fireball")
	if fb == nil {
		t.Fatal("expected to find fireball")
	}

	// Find non-existent
	if st.FindNode("nonexistent") != nil {
		t.Error("expected nil for nonexistent")
	}
}

func TestSkillTreeUnlock(t *testing.T) {
	st := NewSkillTree(newTree(), newCfg())
	st.SetPoints(10)

	st.AddNode(&SkillNode{ID: "fireball", Name: "Fireball", State: SkillAvailable, Level: 0, MaxLevel: 3, Cost: 1})
	st.AddNode(&SkillNode{ID: "meteor", Name: "Meteor", State: SkillLocked, Level: 0, MaxLevel: 1, Cost: 3, Requires: []string{"fireball"}})

	unlocked := ""
	st.OnUnlock(func(id string) { unlocked = id })

	ok := st.UnlockNode("fireball")
	if !ok {
		t.Error("should unlock fireball")
	}
	if unlocked != "fireball" {
		t.Error("expected unlock callback for fireball")
	}
	if st.Points() != 9 {
		t.Errorf("expected 9 points, got %d", st.Points())
	}
	fb := st.FindNode("fireball")
	if fb.Level != 1 {
		t.Error("expected level 1")
	}
	if fb.State != SkillUnlocked {
		t.Error("expected SkillUnlocked state")
	}

	// Unlock meteor (fireball is unlocked, prereq met)
	ok = st.UnlockNode("meteor")
	if !ok {
		t.Error("should unlock meteor")
	}
	meteor := st.FindNode("meteor")
	if meteor.State != SkillMaxed {
		t.Error("expected SkillMaxed since maxLevel=1")
	}

	// Try to unlock maxed node
	ok = st.UnlockNode("meteor")
	if ok {
		t.Error("should not unlock maxed node")
	}
}

func TestSkillTreeUnlockNotEnoughPoints(t *testing.T) {
	st := NewSkillTree(newTree(), newCfg())
	st.SetPoints(0)
	st.AddNode(&SkillNode{ID: "a", Name: "A", State: SkillAvailable, Cost: 1, MaxLevel: 1})
	ok := st.UnlockNode("a")
	if ok {
		t.Error("should not unlock with 0 points")
	}
}

func TestSkillTreeUnlockMissingPrereq(t *testing.T) {
	st := NewSkillTree(newTree(), newCfg())
	st.SetPoints(10)
	st.AddNode(&SkillNode{ID: "ultimate", Name: "Ultimate", State: SkillLocked, Level: 0, MaxLevel: 1, Cost: 1, Requires: []string{"nonexistent"}})
	ok := st.UnlockNode("ultimate")
	if ok {
		t.Error("should not unlock with missing prereq")
	}
}

func TestSkillTreeUnlockLockedPrereq(t *testing.T) {
	st := NewSkillTree(newTree(), newCfg())
	st.SetPoints(10)
	st.AddNode(&SkillNode{ID: "a", Name: "A", State: SkillLocked, Level: 0, MaxLevel: 1, Cost: 1})
	st.AddNode(&SkillNode{ID: "b", Name: "B", State: SkillLocked, Level: 0, MaxLevel: 1, Cost: 1, Requires: []string{"a"}})
	ok := st.UnlockNode("b")
	if ok {
		t.Error("should not unlock with locked prereq")
	}
}

func TestSkillTreeUnlockNonexistentNode(t *testing.T) {
	st := NewSkillTree(newTree(), newCfg())
	st.SetPoints(10)
	ok := st.UnlockNode("nope")
	if ok {
		t.Error("should not unlock nonexistent node")
	}
}

func TestSkillTreeSetters(t *testing.T) {
	st := NewSkillTree(newTree(), newCfg())
	st.SetSelected("fireball")
	if st.Selected() != "fireball" {
		t.Error("expected 'fireball'")
	}
	st.SetNodeSize(64)
	st.SetScroll(10, 20)
	st.OnSelect(func(id string) {})
}

func TestSkillTreeClearNodes(t *testing.T) {
	st := NewSkillTree(newTree(), newCfg())
	st.AddNode(&SkillNode{ID: "a"})
	st.ClearNodes()
	if len(st.Nodes()) != 0 {
		t.Error("expected 0 nodes")
	}
}

func TestSkillTreeDrawNoBounds(t *testing.T) {
	st := NewSkillTree(newTree(), newCfg())
	st.AddNode(&SkillNode{ID: "a", Name: "A", State: SkillAvailable, MaxLevel: 3})
	buf := render.NewCommandBuffer()
	st.Draw(buf)
	// No bounds => early return
}

func TestSkillTreeDrawWithBounds(t *testing.T) {
	tree := newTree()
	st := NewSkillTree(tree, newCfg())
	setBounds(tree, st, 0, 0, 600, 400)
	st.SetSelected("fireball")

	st.AddNode(&SkillNode{ID: "fireball", Name: "Fireball", State: SkillAvailable, X: 10, Y: 10, Level: 0, MaxLevel: 3, Cost: 1})
	st.AddNode(&SkillNode{ID: "meteor", Name: "Meteor", State: SkillLocked, X: 10, Y: 100, Level: 0, MaxLevel: 1, Cost: 3, Requires: []string{"fireball"}})
	st.AddNode(&SkillNode{ID: "maxed", Name: "Max", State: SkillMaxed, X: 200, Y: 10, Level: 3, MaxLevel: 3})
	st.AddNode(&SkillNode{ID: "unlocked", Name: "Unlocked", State: SkillUnlocked, X: 200, Y: 100, Level: 1, MaxLevel: 3, Requires: []string{"fireball"}})

	buf := render.NewCommandBuffer()
	st.Draw(buf)
	if buf.Len() == 0 {
		t.Error("expected commands")
	}
}

func TestSkillTreeDrawNodeOffScreen(t *testing.T) {
	tree := newTree()
	st := NewSkillTree(tree, newCfg())
	setBounds(tree, st, 0, 0, 100, 100)

	st.AddNode(&SkillNode{ID: "far", Name: "Far", State: SkillAvailable, X: 9999, Y: 9999, MaxLevel: 1})
	buf := render.NewCommandBuffer()
	st.Draw(buf)
	// Node is off-screen, should be skipped
}

func TestSkillNodeColor(t *testing.T) {
	states := []SkillNodeState{SkillLocked, SkillAvailable, SkillUnlocked, SkillMaxed}
	for _, s := range states {
		c := skillNodeColor(s)
		if c.A == 0 {
			t.Errorf("skillNodeColor(%d) should have alpha", s)
		}
	}
}

func TestMinFAbsF(t *testing.T) {
	if minF(3, 5) != 3 {
		t.Error("expected 3")
	}
	if minF(5, 3) != 3 {
		t.Error("expected 3")
	}
	if absF(-5) != 5 {
		t.Error("expected 5")
	}
	if absF(5) != 5 {
		t.Error("expected 5")
	}
	if absF(0) != 0 {
		t.Error("expected 0")
	}
}

// ============================================================
// UnitFrame - TeamFrame
// ============================================================

func TestTeamFrameNilConfig(t *testing.T) {
	tf := NewTeamFrame(newTree(), nil)
	if tf == nil {
		t.Fatal("expected non-nil")
	}
}

func TestTeamFrameBasic(t *testing.T) {
	tf := NewTeamFrame(newTree(), newCfg())
	tf.SetMembers([]UnitFrameData{
		{Name: "Tank", HP: 800, HPMax: 1000, MP: 200, MPMax: 200},
		{Name: "Healer", HP: 500, HPMax: 500, MP: 400, MPMax: 500},
	})
	if len(tf.Members()) != 2 {
		t.Error("expected 2 members")
	}

	tf.UpdateMember(0, UnitFrameData{Name: "Tank", HP: 600, HPMax: 1000, Level: 50})
	if tf.Members()[0].HP != 600 {
		t.Error("expected updated HP")
	}

	// Out of bounds update
	tf.UpdateMember(-1, UnitFrameData{})
	tf.UpdateMember(99, UnitFrameData{})
}

func TestTeamFrameSetters(t *testing.T) {
	tf := NewTeamFrame(newTree(), newCfg())
	tf.SetFrameSize(200, 60)
	tf.SetGap(8)
	tf.SetMaxSlots(10)
}

func TestTeamFrameDrawNoBounds(t *testing.T) {
	tf := NewTeamFrame(newTree(), newCfg())
	tf.SetMembers([]UnitFrameData{{Name: "A", HP: 100, HPMax: 100}})
	buf := render.NewCommandBuffer()
	tf.Draw(buf)
	// No bounds => early return
}

func TestTeamFrameDrawWithBounds(t *testing.T) {
	tree := newTree()
	tf := NewTeamFrame(tree, newCfg())
	setBounds(tree, tf, 0, 0, 200, 300)
	tf.SetMembers([]UnitFrameData{
		{Name: "Tank", HP: 800, HPMax: 1000, MP: 200, MPMax: 200, Level: 50},
		{Name: "Healer", HP: 500, HPMax: 500, MP: 400, MPMax: 500, Dead: true},
	})
	buf := render.NewCommandBuffer()
	tf.Draw(buf)
	if len(buf.Overlays()) == 0 {
		t.Error("expected overlays from TeamFrame")
	}
}

func TestTeamFrameDrawMaxSlots(t *testing.T) {
	tree := newTree()
	tf := NewTeamFrame(tree, newCfg())
	setBounds(tree, tf, 0, 0, 200, 500)
	tf.SetMaxSlots(2)
	tf.SetMembers([]UnitFrameData{
		{Name: "A", HP: 100, HPMax: 100},
		{Name: "B", HP: 100, HPMax: 100},
		{Name: "C", HP: 100, HPMax: 100},
		{Name: "D", HP: 100, HPMax: 100},
	})
	buf := render.NewCommandBuffer()
	tf.Draw(buf)
}

// ============================================================
// UnitFrame - TargetFrame
// ============================================================

func TestTargetFrameNilConfig(t *testing.T) {
	tf := NewTargetFrame(newTree(), nil)
	if tf == nil {
		t.Fatal("expected non-nil")
	}
}

func TestTargetFrameBasic(t *testing.T) {
	tf := NewTargetFrame(newTree(), newCfg())
	if tf.IsVisible() {
		t.Error("should not be visible initially")
	}
	if tf.Target() != nil {
		t.Error("expected nil target initially")
	}

	tf.SetTarget(&UnitFrameData{Name: "Boss", HP: 5000, HPMax: 10000, Level: 99})
	if !tf.IsVisible() {
		t.Error("should be visible")
	}
	if tf.Target() == nil {
		t.Error("expected non-nil target")
	}

	tf.ClearTarget()
	if tf.IsVisible() {
		t.Error("should be hidden")
	}
}

func TestTargetFrameSetters(t *testing.T) {
	tf := NewTargetFrame(newTree(), newCfg())
	tf.SetSize(250, 70)
}

func TestTargetFrameDrawNotVisible(t *testing.T) {
	tf := NewTargetFrame(newTree(), newCfg())
	buf := render.NewCommandBuffer()
	tf.Draw(buf)
	if len(buf.Overlays()) != 0 {
		t.Error("expected no overlays when not visible")
	}
}

func TestTargetFrameDrawVisible(t *testing.T) {
	tf := NewTargetFrame(newTree(), newCfg())
	tf.SetTarget(&UnitFrameData{Name: "Boss", HP: 5000, HPMax: 10000, Level: 99, MP: 1000, MPMax: 2000})
	buf := render.NewCommandBuffer()
	tf.Draw(buf)
	if len(buf.Overlays()) == 0 {
		t.Error("expected overlays from TargetFrame")
	}
}

func TestTargetFrameDrawWithBounds(t *testing.T) {
	tree := newTree()
	tf := NewTargetFrame(tree, newCfg())
	setBounds(tree, tf, 10, 10, 200, 56)
	tf.SetTarget(&UnitFrameData{Name: "Boss", HP: 5000, HPMax: 10000, Level: 99})
	buf := render.NewCommandBuffer()
	tf.Draw(buf)
	if len(buf.Overlays()) == 0 {
		t.Error("expected overlays with bounds")
	}
}

func TestTargetFrameDrawDead(t *testing.T) {
	tf := NewTargetFrame(newTree(), newCfg())
	tf.SetTarget(&UnitFrameData{Name: "Dead", HP: 0, HPMax: 100, Dead: true})
	buf := render.NewCommandBuffer()
	tf.Draw(buf)
}

func TestTargetFrameDrawZeroMP(t *testing.T) {
	tf := NewTargetFrame(newTree(), newCfg())
	tf.SetTarget(&UnitFrameData{Name: "Warrior", HP: 100, HPMax: 100, MP: 0, MPMax: 0})
	buf := render.NewCommandBuffer()
	tf.Draw(buf)
}

func TestTargetFrameDrawOverMax(t *testing.T) {
	tf := NewTargetFrame(newTree(), newCfg())
	tf.SetTarget(&UnitFrameData{Name: "Buffed", HP: 200, HPMax: 100, MP: 500, MPMax: 200})
	buf := render.NewCommandBuffer()
	tf.Draw(buf)
	// ratio > 1 clamped to 1
}

// ============================================================
// Helper functions (itoa, formatFloat)
// ============================================================

func TestItoaExtended(t *testing.T) {
	tests := []struct {
		in   int
		want string
	}{
		{0, "0"},
		{1, "1"},
		{42, "42"},
		{100, "100"},
		{-5, "-5"},
		{-100, "-100"},
		{999999, "999999"},
	}
	for _, tt := range tests {
		got := itoa(tt.in)
		if got != tt.want {
			t.Errorf("itoa(%d) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestFormatFloat(t *testing.T) {
	got := formatFloat(100)
	if got != "100" {
		t.Errorf("formatFloat(100) = %q, want '100'", got)
	}
	got = formatFloat(0)
	if got != "0" {
		t.Errorf("formatFloat(0) = %q, want '0'", got)
	}
	// Non-integer values should show one decimal place
	got = formatFloat(50.5)
	if got != "50.5" {
		t.Errorf("formatFloat(50.5) = %q, want '50.5'", got)
	}
	got = formatFloat(-3.7)
	if got != "-3.7" {
		t.Errorf("formatFloat(-3.7) = %q, want '-3.7'", got)
	}
}

// ============================================================
// DrawResourceBar (coverage for shared helper)
// ============================================================

func TestDrawResourceBar(t *testing.T) {
	buf := render.NewCommandBuffer()
	// Normal case
	drawResourceBar(buf, 0, 0, 100, 8, 50, 100, uimath.ColorHex("#52c41a"))
	if len(buf.Overlays()) < 2 {
		t.Error("expected at least 2 overlays (background + fill)")
	}

	// Zero max
	buf = render.NewCommandBuffer()
	drawResourceBar(buf, 0, 0, 100, 8, 50, 0, uimath.ColorHex("#52c41a"))
	if len(buf.Overlays()) != 1 {
		t.Error("expected 1 overlay (background only)")
	}

	// Zero current
	buf = render.NewCommandBuffer()
	drawResourceBar(buf, 0, 0, 100, 8, 0, 100, uimath.ColorHex("#52c41a"))
	if len(buf.Overlays()) != 1 {
		t.Error("expected 1 overlay (background only)")
	}

	// Over max (ratio > 1 clamped)
	buf = render.NewCommandBuffer()
	drawResourceBar(buf, 0, 0, 100, 8, 200, 100, uimath.ColorHex("#52c41a"))
	if len(buf.Overlays()) < 2 {
		t.Error("expected 2 overlays")
	}
}

// ============================================================
// DrawUnitFrame (coverage for shared helper)
// ============================================================

func TestDrawUnitFrame(t *testing.T) {
	cfg := newCfgWithText()
	buf := render.NewCommandBuffer()

	// With level
	drawUnitFrame(buf, cfg, 0, 0, 200, 56, UnitFrameData{Name: "Boss", Level: 99, HP: 500, HPMax: 1000, MP: 200, MPMax: 400})
	if len(buf.Overlays()) == 0 {
		t.Error("expected overlays")
	}

	// Without level (Level == 0)
	buf = render.NewCommandBuffer()
	drawUnitFrame(buf, cfg, 0, 0, 200, 56, UnitFrameData{Name: "NPC", HP: 100, HPMax: 100})

	// Dead
	buf = render.NewCommandBuffer()
	drawUnitFrame(buf, cfg, 0, 0, 200, 56, UnitFrameData{Name: "Dead", Dead: true, HP: 0, HPMax: 100})
}

// ============================================================
// TextRenderer coverage tests — exercise text-rendering branches
// ============================================================

func TestBuffBarDrawWithText(t *testing.T) {
	tree := newTree()
	bb := NewBuffBar(tree, newCfgWithText())
	setBounds(tree, bb, 0, 0, 400, 40)
	bb.AddBuff(Buff{ID: "str", Label: "Strength", Stacks: 3, Type: BuffPositive})
	bb.AddBuff(Buff{ID: "dot", Label: "DoT", Duration: 15, Type: BuffNegative, Stacks: 1})
	buf := render.NewCommandBuffer()
	bb.Draw(buf)
	if buf.Len() == 0 {
		t.Error("expected commands")
	}
}

func TestCastBarDrawWithText(t *testing.T) {
	cb := NewCastBar(newTree(), newCfgWithText())
	cb.StartCast("Fireball", 2.0)
	cb.Tick(0.5)
	buf := render.NewCommandBuffer()
	cb.Draw(buf)
	if len(buf.Overlays()) == 0 {
		t.Error("expected overlays")
	}
}

func TestChatBoxDrawWithText(t *testing.T) {
	tree := newTree()
	cb := NewChatBox(tree, newCfgWithText())
	setBounds(tree, cb, 0, 0, 350, 250)
	cb.SetMaxVisible(5)
	for i := 0; i < 8; i++ {
		cb.AddMessage(ChatMessage{Sender: "U" + itoa(i), Text: "msg " + itoa(i), Channel: "world"})
	}
	// Also test with custom color and empty color
	cb.AddMessage(ChatMessage{Sender: "Colored", Text: "hi", Color: uimath.ColorHex("#ff0000")})
	// Set input text via field (no public setter, test the draw branch)
	cb.inputText = "Hello world"
	buf := render.NewCommandBuffer()
	cb.Draw(buf)
	if buf.Len() == 0 {
		t.Error("expected commands with text renderer")
	}
}

func TestCountdownTimerDrawWithText(t *testing.T) {
	tree := newTree()
	ct := NewCountdownTimer(tree, newCfgWithText())
	setBounds(tree, ct, 0, 0, 200, 100)

	// Normal time with label
	ct.SetSeconds(65)
	ct.SetLabel("Time Left")
	buf := render.NewCommandBuffer()
	ct.Draw(buf)
	if buf.Len() == 0 {
		t.Error("expected commands")
	}

	// Low time (< 10, red flash)
	ct.SetSeconds(5)
	buf = render.NewCommandBuffer()
	ct.Draw(buf)

	// With custom font size
	ct.SetFontSize(32)
	ct.SetSeconds(30)
	buf = render.NewCommandBuffer()
	ct.Draw(buf)
}

func TestCurrencyDisplayDrawWithText(t *testing.T) {
	tree := newTree()
	cd := NewCurrencyDisplay(tree, newCfgWithText())
	setBounds(tree, cd, 0, 0, 300, 40)
	cd.AddCurrency(CurrencyEntry{Symbol: "G", Amount: 1500, Color: uimath.ColorHex("#ffd700")})
	cd.AddCurrency(CurrencyEntry{Symbol: "S", Amount: 45}) // default color
	buf := render.NewCommandBuffer()
	cd.Draw(buf)
	if buf.Len() == 0 {
		t.Error("expected commands with text renderer")
	}
}

func TestDialogueBoxDrawWithText(t *testing.T) {
	tree := newTree()
	db := NewDialogueBox(tree, newCfgWithText())
	setBounds(tree, db, 10, 10, 500, 160)

	// With speaker, text, no choices => "click to continue" hint
	db.Show("Guard", "Halt! Who goes there?")
	buf := render.NewCommandBuffer()
	db.Draw(buf)
	if len(buf.Overlays()) == 0 {
		t.Error("expected overlays")
	}

	// With choices
	db.SetChoices([]DialogueChoice{
		{Text: "Friend"},
		{Text: "Foe"},
	})
	buf = render.NewCommandBuffer()
	db.Draw(buf)
}

func TestDialogueBoxOnAdvanceHandler(t *testing.T) {
	tree := newTree()
	db := NewDialogueBox(tree, newCfgWithText())
	advanced := false
	db.OnAdvance(func() { advanced = true })
	db.Show("NPC", "Hello")
	// No choices set, so click should trigger advance
	d := core.NewDispatcher(tree)
	d.DispatchToTarget(db.ElementID(), &event.Event{Type: event.MouseClick})
	if !advanced {
		t.Error("expected advance callback on click with no choices")
	}

	// With choices, advance should NOT fire
	advanced = false
	db.SetChoices([]DialogueChoice{{Text: "OK"}})
	d.DispatchToTarget(db.ElementID(), &event.Event{Type: event.MouseClick})
	if advanced {
		t.Error("should not advance when choices exist")
	}
}

func TestHealthBarDrawWithText(t *testing.T) {
	tree := newTree()
	hb := NewHealthBar(tree, newCfgWithText())
	setBounds(tree, hb, 0, 0, 200, 20)
	hb.SetCurrent(75)
	hb.SetMax(100)
	hb.SetShowText(true)
	buf := render.NewCommandBuffer()
	hb.Draw(buf)
	if buf.Len() == 0 {
		t.Error("expected commands with text")
	}
}

func TestHotbarDrawWithText(t *testing.T) {
	tree := newTree()
	hb := NewHotbar(tree, 4, newCfgWithText())
	setBounds(tree, hb, 0, 0, 400, 60)
	hb.SetSlot(0, HotbarSlot{Label: "Sword", Keybind: "1", Available: true})
	hb.SetSlot(1, HotbarSlot{Cooldown: 0.5, Keybind: "2"})
	hb.SetSelected(0)
	buf := render.NewCommandBuffer()
	hb.Draw(buf)
	if buf.Len() == 0 {
		t.Error("expected commands")
	}
}

func TestInventoryDrawWithText(t *testing.T) {
	tree := newTree()
	inv := NewInventory(tree, 2, 3, newCfgWithText())
	setBounds(tree, inv, 0, 0, 300, 200)
	inv.SetTitle("Inventory")
	inv.SetItem(0, &ItemData{Name: "Sword", Rarity: RarityRare, Quantity: 5})
	inv.SetItem(1, &ItemData{Name: "Shield", Rarity: RarityEpic, Quantity: 1})
	buf := render.NewCommandBuffer()
	inv.Draw(buf)
	if buf.Len() == 0 {
		t.Error("expected commands with text renderer")
	}
}

func TestLootWindowDrawWithText(t *testing.T) {
	tree := newTree()
	lw := NewLootWindow(tree, newCfgWithText())
	setBounds(tree, lw, 0, 0, 220, 300)
	lw.AddItem(LootItem{Item: &ItemData{Name: "Gold", Rarity: RarityCommon}, Quantity: 10})
	lw.AddItem(LootItem{Item: &ItemData{Name: "Sword", Rarity: RarityEpic}, Quantity: 1, Claimed: true})
	lw.Open()
	buf := render.NewCommandBuffer()
	lw.Draw(buf)
	if len(buf.Overlays()) == 0 {
		t.Error("expected overlays with text renderer")
	}
}

func TestNameplateDrawWithText(t *testing.T) {
	np := NewNameplate(newTree(), "Dragon", newCfgWithText())
	np.SetHP(50, 100)
	np.SetType(NameplateHostile)
	np.SetPosition(200, 100)
	buf := render.NewCommandBuffer()
	np.Draw(buf)
	if len(buf.Overlays()) == 0 {
		t.Error("expected overlays with text")
	}
}

func TestQuestTrackerDrawWithText(t *testing.T) {
	tree := newTree()
	qt := NewQuestTracker(tree, newCfgWithText())
	setBounds(tree, qt, 0, 0, 250, 400)
	qt.AddQuest(Quest{
		Title:  "Active Quest",
		Active: true,
		Objectives: []QuestObjective{
			{Text: "Kill", Current: 2, Required: 5},
			{Text: "Done", Completed: true, Required: 0},
		},
	})
	qt.AddQuest(Quest{Title: "Inactive", Active: false})
	buf := render.NewCommandBuffer()
	qt.Draw(buf)
	if buf.Len() == 0 {
		t.Error("expected commands with text renderer")
	}
}

func TestRadialMenuDrawWithText(t *testing.T) {
	rm := NewRadialMenu(newTree(), newCfgWithText())
	rm.AddItem(RadialMenuItem{Label: "Attack"})
	rm.AddItem(RadialMenuItem{Label: "Defend"})
	rm.AddItem(RadialMenuItem{Label: "Flee", Disabled: true})
	rm.Show(200, 200)
	rm.SetHovered(0)
	buf := render.NewCommandBuffer()
	rm.Draw(buf)
	if len(buf.Overlays()) == 0 {
		t.Error("expected overlays with text")
	}
}

func TestScoreboardDrawWithText(t *testing.T) {
	tree := newTree()
	sb := NewScoreboard(tree, newCfgWithText())
	setBounds(tree, sb, 0, 0, 400, 300)
	sb.AddEntry(ScoreEntry{Name: "Alice", Score: 100, Kills: 5, Deaths: 2})
	sb.AddEntry(ScoreEntry{Name: "Bob", Score: 200, Kills: 8, Deaths: 1})
	sb.AddEntry(ScoreEntry{Name: "Charlie", Score: 50, Kills: 2, Deaths: 5})
	sb.SetVisible(true)
	buf := render.NewCommandBuffer()
	sb.Draw(buf)
	if len(buf.Overlays()) == 0 {
		t.Error("expected overlays with text")
	}
}

func TestSkillTreeDrawWithText(t *testing.T) {
	tree := newTree()
	st := NewSkillTree(tree, newCfgWithText())
	setBounds(tree, st, 0, 0, 600, 400)
	st.SetSelected("fireball")
	st.AddNode(&SkillNode{ID: "fireball", Name: "Fireball", State: SkillAvailable, X: 10, Y: 10, Level: 0, MaxLevel: 3, Cost: 1})
	st.AddNode(&SkillNode{ID: "meteor", Name: "Meteor", State: SkillLocked, X: 10, Y: 100, Level: 0, MaxLevel: 1, Cost: 3, Requires: []string{"fireball"}})
	st.AddNode(&SkillNode{ID: "maxed", Name: "Max", State: SkillMaxed, X: 200, Y: 10, Level: 3, MaxLevel: 3})
	st.AddNode(&SkillNode{ID: "unlocked", Name: "Unlocked", State: SkillUnlocked, X: 200, Y: 100, Level: 1, MaxLevel: 3, Requires: []string{"fireball"}})
	buf := render.NewCommandBuffer()
	st.Draw(buf)
	if buf.Len() == 0 {
		t.Error("expected commands with text renderer")
	}
}

func TestTeamFrameDrawWithText(t *testing.T) {
	tree := newTree()
	tf := NewTeamFrame(tree, newCfgWithText())
	setBounds(tree, tf, 0, 0, 200, 300)
	tf.SetMembers([]UnitFrameData{
		{Name: "Tank", HP: 800, HPMax: 1000, MP: 200, MPMax: 200, Level: 50},
		{Name: "Healer", HP: 500, HPMax: 500, MP: 400, MPMax: 500, Dead: true},
	})
	buf := render.NewCommandBuffer()
	tf.Draw(buf)
	if len(buf.Overlays()) == 0 {
		t.Error("expected overlays with text")
	}
}

func TestTargetFrameDrawWithText(t *testing.T) {
	tf := NewTargetFrame(newTree(), newCfgWithText())
	tf.SetTarget(&UnitFrameData{Name: "Boss", HP: 5000, HPMax: 10000, Level: 99, MP: 1000, MPMax: 2000})
	buf := render.NewCommandBuffer()
	tf.Draw(buf)
	if len(buf.Overlays()) == 0 {
		t.Error("expected overlays with text")
	}
}

func TestItemTooltipDrawWithText(t *testing.T) {
	tt := NewItemTooltip(newTree(), newCfgWithText())
	tt.SetItem(&ItemData{Name: "Dragon Sword", Rarity: RarityLegendary})
	tt.SetVisible(true)
	tt.SetPosition(50, 50)
	buf := render.NewCommandBuffer()
	tt.Draw(buf)
	if len(buf.Overlays()) == 0 {
		t.Error("expected overlays with text")
	}
}

func TestNotificationToastDrawWithText(t *testing.T) {
	nt := NewNotificationToast(newTree(), "Item acquired!", newCfgWithText())
	nt.SetToastType(ToastSuccess)
	buf := render.NewCommandBuffer()
	nt.Draw(buf)
	if len(buf.Overlays()) == 0 {
		t.Error("expected overlays with text")
	}
}

func TestFloatingTextDrawWithText(t *testing.T) {
	ft := NewFloatingText(newTree(), "-50", 100, 200, uimath.ColorHex("#ff0000"), newCfgWithText())
	buf := render.NewCommandBuffer()
	ft.Draw(buf)
	// With TextRenderer, uses DrawText path
}
