package game

import (
	"testing"

	"github.com/kasuganosora/ui/core"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
	"github.com/kasuganosora/ui/widget"
)

func BenchmarkHealthBarDraw(b *testing.B) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	hb := NewHealthBar(tree, cfg)
	hb.SetCurrent(75)
	hb.SetMax(100)
	tree.SetLayout(hb.ElementID(), core.LayoutResult{
		Bounds: uimath.NewRect(0, 0, 200, 20),
	})
	buf := render.NewCommandBuffer()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		hb.Draw(buf)
	}
}

func BenchmarkInventoryDraw(b *testing.B) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	inv := NewInventory(tree, 6, 8, cfg)
	inv.SetTitle("Inventory")
	tree.SetLayout(inv.ElementID(), core.LayoutResult{
		Bounds: uimath.NewRect(0, 0, 450, 400),
	})

	// Fill half the slots
	for i := 0; i < 24; i++ {
		inv.SetItem(i, &ItemData{
			ID:       itoa(i),
			Name:     "Item " + itoa(i),
			Quantity: i + 1,
			Rarity:   ItemRarity(i % 5),
		})
	}
	buf := render.NewCommandBuffer()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		inv.Draw(buf)
	}
}

func BenchmarkHotbarDraw(b *testing.B) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	hb := NewHotbar(tree, 10, cfg)
	hb.SetSelected(3)
	tree.SetLayout(hb.ElementID(), core.LayoutResult{
		Bounds: uimath.NewRect(0, 0, 600, 60),
	})

	for i := 0; i < 10; i++ {
		hb.SetSlot(i, HotbarSlot{
			Label:     "Skill" + itoa(i),
			Keybind:   itoa(i),
			Cooldown:  float32(i) * 0.1,
			Available: true,
		})
	}
	buf := render.NewCommandBuffer()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		hb.Draw(buf)
	}
}

func BenchmarkChatBoxDraw(b *testing.B) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	cb := NewChatBox(tree, cfg)
	tree.SetLayout(cb.ElementID(), core.LayoutResult{
		Bounds: uimath.NewRect(0, 0, 350, 250),
	})

	for i := 0; i < 50; i++ {
		cb.AddMessage(ChatMessage{
			Sender:  "Player" + itoa(i%10),
			Text:    "Message content " + itoa(i),
			Channel: "world",
		})
	}
	buf := render.NewCommandBuffer()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		cb.Draw(buf)
	}
}

func BenchmarkScoreboardDraw(b *testing.B) {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	sb := NewScoreboard(tree, cfg)
	sb.SetVisible(true)
	tree.SetLayout(sb.ElementID(), core.LayoutResult{
		Bounds: uimath.NewRect(0, 0, 400, 600),
	})

	for i := 0; i < 16; i++ {
		sb.AddEntry(ScoreEntry{
			Name:   "Player" + itoa(i),
			Score:  1000 - i*50,
			Kills:  10 + i,
			Deaths: i,
			Team:   i % 2,
		})
	}
	buf := render.NewCommandBuffer()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		sb.Draw(buf)
	}
}
