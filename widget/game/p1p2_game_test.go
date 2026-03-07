package game

import (
	"testing"

	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/render"
	"github.com/kasuganosora/ui/widget"
)

func newTree() *core.Tree       { return core.NewTree() }
func newCfg() *widget.Config    { return widget.DefaultConfig() }

func TestMinimap(t *testing.T) {
	tree := newTree()
	cfg := newCfg()
	m := NewMinimap(tree, cfg)
	m.SetPlayerPos(100, 200)
	m.SetZoom(2)
	m.AddMarker(MinimapMarker{X: 110, Y: 210, Label: "NPC"})
	if len(m.Markers()) != 1 {
		t.Errorf("expected 1 marker")
	}
	m.ClearMarkers()
	if len(m.Markers()) != 0 {
		t.Error("expected 0 markers after clear")
	}
	buf := render.NewCommandBuffer()
	m.Draw(buf)
	if len(buf.Overlays()) == 0 {
		t.Error("expected overlay commands from minimap")
	}
}

func TestRadialMenu(t *testing.T) {
	tree := newTree()
	cfg := newCfg()
	rm := NewRadialMenu(tree, cfg)
	rm.AddItem(RadialMenuItem{Label: "Attack"})
	rm.AddItem(RadialMenuItem{Label: "Defend"})
	rm.AddItem(RadialMenuItem{Label: "Flee", Disabled: true})
	if len(rm.Items()) != 3 {
		t.Errorf("expected 3 items")
	}
	rm.Show(200, 200)
	if !rm.IsVisible() {
		t.Error("should be visible")
	}
	rm.Hide()
	if rm.IsVisible() {
		t.Error("should be hidden")
	}
}

func TestQuestTracker(t *testing.T) {
	tree := newTree()
	cfg := newCfg()
	qt := NewQuestTracker(tree, cfg)
	qt.AddQuest(Quest{
		Title:  "Slay the Dragon",
		Active: true,
		Objectives: []QuestObjective{
			{Text: "Kill dragons", Current: 2, Required: 5},
		},
	})
	if len(qt.Quests()) != 1 {
		t.Errorf("expected 1 quest")
	}
	qt.RemoveQuest(0)
	if len(qt.Quests()) != 0 {
		t.Error("expected 0 quests after remove")
	}
}

func TestBuffBar(t *testing.T) {
	tree := newTree()
	cfg := newCfg()
	bb := NewBuffBar(tree, cfg)
	bb.AddBuff(Buff{ID: "str", Label: "Strength", Stacks: 3, Type: BuffPositive})
	bb.AddBuff(Buff{ID: "poison", Label: "Poison", Duration: 10, Type: BuffNegative})
	if len(bb.Buffs()) != 2 {
		t.Errorf("expected 2 buffs")
	}
	bb.RemoveBuff("str")
	if len(bb.Buffs()) != 1 {
		t.Error("expected 1 buff after remove")
	}
	bb.ClearBuffs()
	if len(bb.Buffs()) != 0 {
		t.Error("expected 0 buffs after clear")
	}
}

func TestNameplate(t *testing.T) {
	tree := newTree()
	cfg := newCfg()
	np := NewNameplate(tree, "Dragon", cfg)
	np.SetHP(50, 100)
	np.SetType(NameplateHostile)
	np.SetPosition(200, 100)
	if np.Name() != "Dragon" {
		t.Errorf("expected name 'Dragon'")
	}
	buf := render.NewCommandBuffer()
	np.Draw(buf)
	if len(buf.Overlays()) == 0 {
		t.Error("expected overlay commands from nameplate")
	}
}

func TestScoreboard(t *testing.T) {
	tree := newTree()
	cfg := newCfg()
	sb := NewScoreboard(tree, cfg)
	sb.AddEntry(ScoreEntry{Name: "Alice", Score: 100, Kills: 5, Deaths: 2})
	sb.AddEntry(ScoreEntry{Name: "Bob", Score: 200, Kills: 8, Deaths: 1})
	sb.SortByScore()
	if sb.Entries()[0].Name != "Bob" {
		t.Error("expected Bob first after sort")
	}
	sb.SetVisible(true)
	buf := render.NewCommandBuffer()
	sb.Draw(buf)
	if len(buf.Overlays()) == 0 {
		t.Error("expected overlay commands from scoreboard")
	}
}

func TestDialogueBox(t *testing.T) {
	tree := newTree()
	cfg := newCfg()
	db := NewDialogueBox(tree, cfg)
	db.Show("Guard", "Halt! Who goes there?")
	if !db.IsVisible() {
		t.Error("should be visible")
	}
	if db.Speaker() != "Guard" {
		t.Error("expected speaker 'Guard'")
	}
	db.SetChoices([]DialogueChoice{
		{Text: "I am a friend"},
		{Text: "None of your business"},
	})
	if len(db.Choices()) != 2 {
		t.Errorf("expected 2 choices")
	}
	db.Hide()
	if db.IsVisible() {
		t.Error("should be hidden")
	}
}

func TestCountdownTimer(t *testing.T) {
	tree := newTree()
	cfg := newCfg()
	ct := NewCountdownTimer(tree, cfg)
	ct.SetSeconds(65)
	ct.SetLabel("Time Left")
	ct.Tick(10)
	if ct.Seconds() != 55 {
		t.Errorf("expected 55 seconds, got %g", ct.Seconds())
	}
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
}

func TestCurrencyDisplay(t *testing.T) {
	tree := newTree()
	cfg := newCfg()
	cd := NewCurrencyDisplay(tree, cfg)
	cd.AddCurrency(CurrencyEntry{Symbol: "G", Amount: 1500})
	cd.AddCurrency(CurrencyEntry{Symbol: "S", Amount: 45})
	if len(cd.Currencies()) != 2 {
		t.Errorf("expected 2 currencies")
	}
	cd.SetAmount(0, 2000)
	if cd.Currencies()[0].Amount != 2000 {
		t.Error("expected updated amount")
	}
}

func TestTeamFrame(t *testing.T) {
	tree := newTree()
	cfg := newCfg()
	tf := NewTeamFrame(tree, cfg)
	tf.SetMembers([]UnitFrameData{
		{Name: "Tank", HP: 800, HPMax: 1000, MP: 200, MPMax: 200},
		{Name: "Healer", HP: 500, HPMax: 500, MP: 400, MPMax: 500},
	})
	if len(tf.Members()) != 2 {
		t.Errorf("expected 2 members")
	}
	tf.UpdateMember(0, UnitFrameData{Name: "Tank", HP: 600, HPMax: 1000, Level: 50})
	if tf.Members()[0].HP != 600 {
		t.Error("expected updated HP")
	}
}

func TestTargetFrame(t *testing.T) {
	tree := newTree()
	cfg := newCfg()
	tf := NewTargetFrame(tree, cfg)
	if tf.IsVisible() {
		t.Error("should not be visible initially")
	}
	tf.SetTarget(&UnitFrameData{Name: "Boss", HP: 5000, HPMax: 10000, Level: 99})
	if !tf.IsVisible() {
		t.Error("should be visible after setting target")
	}
	tf.ClearTarget()
	if tf.IsVisible() {
		t.Error("should be hidden after clear")
	}
}

func TestCastBar(t *testing.T) {
	tree := newTree()
	cfg := newCfg()
	cb := NewCastBar(tree, cfg)
	cb.StartCast("Fireball", 2.0)
	if !cb.IsVisible() {
		t.Error("should be visible during cast")
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

	// Test interrupt
	cb.StartCast("Heal", 3.0)
	interrupted := false
	cb.OnInterrupt(func() { interrupted = true })
	cb.Interrupt()
	if !interrupted {
		t.Error("expected interrupt callback")
	}
}

func TestLootWindow(t *testing.T) {
	tree := newTree()
	cfg := newCfg()
	lw := NewLootWindow(tree, cfg)
	lw.AddItem(LootItem{Item: &ItemData{Name: "Gold Coin", Rarity: RarityCommon}, Quantity: 10})
	lw.AddItem(LootItem{Item: &ItemData{Name: "Epic Sword", Rarity: RarityEpic}, Quantity: 1})
	if len(lw.Items()) != 2 {
		t.Errorf("expected 2 items")
	}
	lw.Open()
	if !lw.IsVisible() {
		t.Error("should be visible")
	}
	looted := -1
	lw.OnLoot(func(i int) { looted = i })
	lw.LootItem(0)
	if looted != 0 {
		t.Error("expected loot callback for index 0")
	}
	if !lw.Items()[0].Claimed {
		t.Error("item should be claimed")
	}
	lw.Close()
	if lw.IsVisible() {
		t.Error("should be hidden after close")
	}
}

func TestSkillTree(t *testing.T) {
	tree := newTree()
	cfg := newCfg()
	st := NewSkillTree(tree, cfg)
	st.SetPoints(5)
	st.AddNode(&SkillNode{ID: "fireball", Name: "Fireball", State: SkillAvailable, Level: 0, MaxLevel: 3, Cost: 1})
	st.AddNode(&SkillNode{ID: "meteor", Name: "Meteor", State: SkillLocked, Level: 0, MaxLevel: 1, Cost: 3, Requires: []string{"fireball"}})

	// Unlock fireball
	ok := st.UnlockNode("fireball")
	if !ok {
		t.Error("should unlock fireball")
	}
	if st.Points() != 4 {
		t.Errorf("expected 4 points, got %d", st.Points())
	}
	fb := st.FindNode("fireball")
	if fb.Level != 1 {
		t.Errorf("expected level 1")
	}

	// Can't unlock meteor yet (fireball not maxed but unlocked, so prereq met)
	ok = st.UnlockNode("meteor")
	if !ok {
		t.Error("should unlock meteor since fireball is unlocked")
	}

	// Try with locked prereqs
	st.AddNode(&SkillNode{ID: "ultimate", Name: "Ultimate", State: SkillLocked, Level: 0, MaxLevel: 1, Cost: 1, Requires: []string{"nonexistent"}})
	ok = st.UnlockNode("ultimate")
	if ok {
		t.Error("should not unlock with missing prereq")
	}
}
