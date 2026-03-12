// Game UI Demo — Showcases GoUI's game widget library in an RPG-style HUD.
//
// Uses HTML+CSS layout (position:absolute) for all HUD element placement.
// Game widgets participate in CSSLayout via their Style() properties.
//
// Demonstrates:
//   - HUD overlay with CSS-positioned elements (health/mana bars, hotbar, minimap)
//   - Inventory grid with rarity-colored items and drag-and-drop
//   - Chat box with scrollable Div + Input (framework controls)
//   - Cast bar with real-time progress
//   - Buff/debuff bar with duration tracking
//   - Quest tracker, unit frames, scoreboard
//   - Skill tree with prerequisites
//   - Nameplate system, currency display, countdown timer
//   - Dialogue box with NPC choices
//   - Draggable panels with bring-to-front
//
// Run: go run ./cmd/game
package main

import (
	"flag"
	"fmt"
	"math"
	"os"

	ui "github.com/kasuganosora/ui"
	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/event"
	"github.com/kasuganosora/ui/layout"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/widget"
	"github.com/kasuganosora/ui/widget/game"
)

// dragState tracks panel drag operations.
type dragState struct {
	active bool
	target widget.Widget
	startX float32
	startY float32
	// Original computed position (from CSSLayout) at drag start
	origX float32
	origY float32
}

// dragOffset stores accumulated drag offset for a widget (persistent across drags).
type dragOffset struct {
	dx, dy float32
}

func main() {
	backendFlag := flag.String("backend", "auto", "rendering backend: auto, vulkan, dx11, dx9, gl")
	flag.Parse()

	var backend ui.BackendType
	switch *backendFlag {
	case "dx11", "d3d11":
		backend = ui.BackendDX11
	case "dx9", "d3d9":
		backend = ui.BackendDX9
	case "vulkan", "vk":
		backend = ui.BackendVulkan
	case "gl", "opengl":
		backend = ui.BackendOpenGL
	default:
		backend = ui.BackendAuto
	}

	app, err := ui.NewApp(ui.AppOptions{
		Title:   "GoUI — Game UI Demo (RPG HUD)",
		Width:   1280,
		Height:  800,
		Backend: backend,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	defer app.Destroy()

	tree := app.Tree()
	cfg := app.Config()

	// Dark RPG theme
	cfg.BgColor = uimath.ColorHex("#0a0e17")
	cfg.TextColor = uimath.ColorHex("#c8ccd0")
	cfg.PrimaryColor = uimath.ColorHex("#1D9BF0")
	cfg.SuccessColor = uimath.ColorHex("#52c41a")
	cfg.ErrorColor = uimath.ColorHex("#ff4d4f")
	cfg.WarningColor = uimath.ColorHex("#faad14")
	cfg.BorderColor = uimath.ColorHex("#2a2f38")

	// ── HTML + CSS layout ───────────────────────────────────────────
	// Root div is position:relative, fills viewport.
	// All HUD elements use position:absolute with CSS offsets.

	doc := app.LoadHTML(`<div style="position:relative; width:100%; height:100%; background:#0a0e17;">
  <div id="chat-messages" style="position:absolute; bottom:110px; left:10px; width:340px; height:172px; overflow:auto;"></div>
  <input id="chat-input" style="position:absolute; bottom:82px; left:10px; width:340px; height:28px;" placeholder="输入消息..." />
</div>`)
	rootDiv := doc.Root.Children()[0].(*widget.Div)

	// Chat controls from HTML
	chatMsgDiv := doc.QueryByID("chat-messages").(*widget.Div)
	chatInput := doc.QueryByID("chat-input").(*widget.Input)
	chatInput.SetBorderless(true)

	// ── CSS positioning helper ──────────────────────────────────────

	// absTopLeft positions an element at (left, top) with explicit size.
	absTopLeft := func(left, top, w, h float32) layout.Style {
		return layout.Style{
			Position: layout.PositionAbsolute,
			Left:     layout.Px(left),
			Top:      layout.Px(top),
			Width:    layout.Px(w),
			Height:   layout.Px(h),
		}
	}
	// absTopRight positions an element at (right, top) with explicit size.
	absTopRight := func(right, top, w, h float32) layout.Style {
		return layout.Style{
			Position: layout.PositionAbsolute,
			Right:    layout.Px(right),
			Top:      layout.Px(top),
			Width:    layout.Px(w),
			Height:   layout.Px(h),
		}
	}
	// absTopCenter positions an element centered horizontally with top offset.
	// Uses left:50% + margin-left:-(w/2 + extraOffsetX).
	absTopCenter := func(top, w, h, extraOffsetX float32) layout.Style {
		return layout.Style{
			Position: layout.PositionAbsolute,
			Left:     layout.Pct(50),
			Top:      layout.Px(top),
			Width:    layout.Px(w),
			Height:   layout.Px(h),
			Margin: layout.EdgeValues{
				Left: layout.Px(-w/2 + extraOffsetX),
			},
		}
	}
	// absBottomCenter positions an element centered horizontally with bottom offset.
	absBottomCenter := func(bottom, w, h, extraOffsetX float32) layout.Style {
		return layout.Style{
			Position: layout.PositionAbsolute,
			Left:     layout.Pct(50),
			Bottom:   layout.Px(bottom),
			Width:    layout.Px(w),
			Height:   layout.Px(h),
			Margin: layout.EdgeValues{
				Left: layout.Px(-w/2 + extraOffsetX),
			},
		}
	}
	// absBottomLeft positions an element at (left, bottom) with explicit size.
	absBottomLeft := func(left, bottom, w, h float32) layout.Style {
		return layout.Style{
			Position: layout.PositionAbsolute,
			Left:     layout.Px(left),
			Bottom:   layout.Px(bottom),
			Width:    layout.Px(w),
			Height:   layout.Px(h),
		}
	}
	// absMiddleRight positions an element at right edge, vertically centered.
	absMiddleRight := func(right, w, h, extraOffsetY float32) layout.Style {
		return layout.Style{
			Position: layout.PositionAbsolute,
			Right:    layout.Px(right),
			Top:      layout.Pct(50),
			Width:    layout.Px(w),
			Height:   layout.Px(h),
			Margin: layout.EdgeValues{
				Top: layout.Px(-h/2 + extraOffsetY),
			},
		}
	}
	// absMiddleLeft positions an element at left edge, vertically centered.
	absMiddleLeft := func(left, w, h, extraOffsetY float32) layout.Style {
		return layout.Style{
			Position: layout.PositionAbsolute,
			Left:     layout.Px(left),
			Top:      layout.Pct(50),
			Width:    layout.Px(w),
			Height:   layout.Px(h),
			Margin: layout.EdgeValues{
				Top: layout.Px(-h/2 + extraOffsetY),
			},
		}
	}
	// absMiddleCenter positions an element centered both ways with offsets.
	absMiddleCenter := func(w, h, extraOffsetX, extraOffsetY float32) layout.Style {
		return layout.Style{
			Position: layout.PositionAbsolute,
			Left:     layout.Pct(50),
			Top:      layout.Pct(50),
			Width:    layout.Px(w),
			Height:   layout.Px(h),
			Margin: layout.EdgeValues{
				Left: layout.Px(-w/2 + extraOffsetX),
				Top:  layout.Px(-h/2 + extraOffsetY),
			},
		}
	}

	// ── Health / Mana / XP bars (top-left) ───────────────────────────

	hpBar := game.NewHealthBar(tree, cfg)
	hpBar.SetCurrent(780)
	hpBar.SetMax(1200)
	hpBar.SetBarColor(uimath.ColorHex("#52c41a"))
	hpBar.SetShowText(true)
	hpBar.SetSize(220, 22)
	hpBar.SetStyle(absTopLeft(20, 20, 220, 22))
	rootDiv.AppendChild(hpBar)

	mpBar := game.NewHealthBar(tree, cfg)
	mpBar.SetCurrent(350)
	mpBar.SetMax(600)
	mpBar.SetBarColor(uimath.ColorHex("#1890ff"))
	mpBar.SetShowText(true)
	mpBar.SetSize(220, 18)
	mpBar.SetStyle(absTopLeft(20, 48, 220, 18))
	rootDiv.AppendChild(mpBar)

	xpBar := game.NewHealthBar(tree, cfg)
	xpBar.SetCurrent(4200)
	xpBar.SetMax(8500)
	xpBar.SetBarColor(uimath.ColorHex("#a335ee"))
	xpBar.SetShowText(true)
	xpBar.SetSize(220, 12)
	xpBar.SetStyle(absTopLeft(20, 72, 220, 12))
	rootDiv.AppendChild(xpBar)

	// ── Buff bar (below bars) ────────────────────────────────────────

	buffBar := game.NewBuffBar(tree, cfg)
	buffBar.SetIconSize(28)
	buffBar.SetGap(3)
	buffBar.AddBuff(game.Buff{ID: "str", Label: "力", Duration: 120, Type: game.BuffPositive})
	buffBar.AddBuff(game.Buff{ID: "haste", Label: "速", Duration: 45, Type: game.BuffPositive})
	buffBar.AddBuff(game.Buff{ID: "shield", Label: "盾", Duration: 30, Type: game.BuffPositive})
	buffBar.AddBuff(game.Buff{ID: "poison", Label: "毒", Duration: 8, Type: game.BuffNegative})
	buffBar.AddBuff(game.Buff{ID: "slow", Label: "慢", Duration: 15, Type: game.BuffNegative})
	buffBar.SetStyle(absTopLeft(20, 94, 300, 28))
	rootDiv.AppendChild(buffBar)

	// ── Team frames (left side) ──────────────────────────────────────

	team := game.NewTeamFrame(tree, cfg)
	team.SetFrameSize(170, 44)
	team.SetGap(3)
	team.SetMembers([]game.UnitFrameData{
		{Name: "龙骑士·苍", Level: 60, HP: 11500, HPMax: 15000, MP: 2800, MPMax: 3000, Class: "战士"},
		{Name: "月影法师", Level: 59, HP: 6200, HPMax: 8000, MP: 1200, MPMax: 6000, Class: "法师"},
		{Name: "神圣牧师", Level: 60, HP: 7800, HPMax: 9500, MP: 4500, MPMax: 7000, Class: "牧师"},
		{Name: "暗影猎手", Level: 58, HP: 0, HPMax: 7500, MP: 2000, MPMax: 3500, Class: "猎人", Dead: true},
	})
	team.SetStyle(absTopLeft(20, 132, 170, 4*47))
	rootDiv.AppendChild(team)

	// ── Hotbar (bottom-center) ──────────────────────────────────────

	hotbar := game.NewHotbar(tree, 10, cfg)
	hotbar.SetSlotSize(52)
	hotbar.SetGap(4)
	hotbar.SetSelected(0)
	for i := 0; i < 10; i++ {
		hotbar.SetSlot(i, game.HotbarSlot{Keybind: fmt.Sprintf("%d", (i+1)%10), Available: true})
	}
	hotbar.SetSlot(2, game.HotbarSlot{Keybind: "3", Cooldown: 0.65, Available: true})
	hotbar.SetSlot(5, game.HotbarSlot{Keybind: "6", Cooldown: 0.3, Available: false})
	hotbarW := float32(10*(52+4))
	hotbar.SetStyle(absBottomCenter(20, hotbarW, 52, 0))
	rootDiv.AppendChild(hotbar)

	// ── Cast bar (above hotbar) ──────────────────────────────────────

	castBar := game.NewCastBar(tree, cfg)
	castBar.SetSize(280, 22)
	castBar.SetColor(uimath.ColorHex("#ffd700"))
	castBar.StartCast("火球术", 3.0)
	castBar.Tick(1.5) // start at 50%
	castBar.SetStyle(absBottomCenter(82, 280, 22, 0))
	rootDiv.AppendChild(castBar)

	// ── Minimap (top-right) ──────────────────────────────────────────

	minimap := game.NewMinimap(tree, cfg)
	minimap.SetSize(160)
	minimap.SetCircular(true)
	minimap.SetPlayerPos(80, 80)
	minimap.SetPlayerRotation(0.4)
	minimap.AddMarker(game.MinimapMarker{X: 40, Y: 50, Color: uimath.ColorHex("#ff4444"), Size: 6, Label: "!"})
	minimap.AddMarker(game.MinimapMarker{X: 120, Y: 30, Color: uimath.ColorHex("#ffdd44"), Size: 5, Label: "?"})
	minimap.AddMarker(game.MinimapMarker{X: 60, Y: 130, Color: uimath.ColorHex("#44aaff"), Size: 5})
	minimap.SetStyle(absTopRight(20, 20, 160, 160))
	rootDiv.AppendChild(minimap)

	// ── Quest tracker (right, below minimap) ─────────────────────────

	quest := game.NewQuestTracker(tree, cfg)
	quest.SetWidth(230)
	quest.AddQuest(game.Quest{Title: "讨伐暗影领主", Active: true, Objectives: []game.QuestObjective{
		{Text: "击败暗影守卫", Current: 3, Required: 5},
		{Text: "收集暗影碎片", Current: 7, Required: 10},
		{Text: "到达暗影塔顶层"},
	}})
	quest.AddQuest(game.Quest{Title: "商人的委托", Active: true, Objectives: []game.QuestObjective{
		{Text: "收集铁矿石", Current: 12, Required: 12, Completed: true},
		{Text: "交给铁匠铺"},
	}})
	quest.SetStyle(absTopRight(20, 200, 230, 200))
	rootDiv.AppendChild(quest)

	// ── Currency display (top-center) ────────────────────────────────

	currency := game.NewCurrencyDisplay(tree, cfg)
	currency.SetGap(16)
	currency.AddCurrency(game.CurrencyEntry{Symbol: "G", Amount: 12580, Color: uimath.ColorHex("#ffd700")})
	currency.AddCurrency(game.CurrencyEntry{Symbol: "◆", Amount: 350, Color: uimath.ColorHex("#44aaff")})
	currency.AddCurrency(game.CurrencyEntry{Symbol: "★", Amount: 28, Color: uimath.ColorHex("#ff8800")})
	currency.SetStyle(absTopCenter(8, 300, 24, 0))
	rootDiv.AppendChild(currency)

	// ── Countdown timer ──────────────────────────────────────────────

	countdown := game.NewCountdownTimer(tree, cfg)
	countdown.SetSeconds(185)
	countdown.SetLabel("Boss 刷新")
	countdown.SetColor(uimath.ColorWhite)
	countdown.SetStyle(absTopCenter(36, 200, 30, 0))
	rootDiv.AppendChild(countdown)

	// ── Target frame (top-center, offset left) ──────────────────────

	target := game.NewTargetFrame(tree, cfg)
	target.SetSize(220, 52)
	target.SetTarget(&game.UnitFrameData{
		Name: "暗影领主·莫德雷克", Level: 62, HP: 185000, HPMax: 500000, MP: 80000, MPMax: 80000, Class: "Boss",
	})
	target.SetStyle(absTopCenter(68, 220, 52, -140))
	rootDiv.AppendChild(target)

	// ── Chat box background (bottom-left) ────────────────────────────

	chat := game.NewChatBox(tree, cfg)
	chat.SetSize(340, 200)
	chat.SetMaxVisible(8)
	chat.SetStyle(absBottomLeft(10, 82, 340, 200))
	rootDiv.AppendChild(chat)

	// ── Chat messages (framework scrollable Div from HTML) ──────────

	lineH := float32(18)
	addChatMsg := func(sender, text string, color uimath.Color) {
		chat.AddMessage(game.ChatMessage{Sender: sender, Text: text, Color: color})
		msgText := widget.NewText(tree, "["+sender+"] "+text, cfg)
		msgText.SetColor(color)
		msgText.SetFontSize(cfg.FontSizeSm)
		chatMsgDiv.AppendChild(msgText)
		n := float32(len(chat.Messages()))
		chatMsgDiv.SetContentHeight(n * lineH)
		maxScroll := n*lineH - 172 // message area height
		if maxScroll > 0 {
			chatMsgDiv.ScrollTo(0, maxScroll)
		}
	}

	addChatMsg("系统", "欢迎来到暗影之境！", uimath.ColorHex("#ffd700"))
	addChatMsg("骑士", "队伍已就绪，准备进攻北塔", uimath.ColorHex("#44aaff"))
	addChatMsg("法师", "我的蓝不多了", uimath.ColorHex("#44aaff"))
	addChatMsg("世界", "LFG 副本·迷雾深渊 4/5 缺T", uimath.ColorHex("#ffaa00"))
	addChatMsg("治疗", "注意躲地板火！", uimath.ColorHex("#52c41a"))

	// Chat scroll via MouseWheel
	chatMsgDiv.On(event.MouseWheel, func(e *event.Event) {
		chatMsgDiv.ScrollTo(0, chatMsgDiv.ScrollY()-e.WheelDY*30)
		maxScroll := chatMsgDiv.ContentHeight() - 172
		if maxScroll < 0 {
			maxScroll = 0
		}
		sy := chatMsgDiv.ScrollY()
		if sy < 0 {
			sy = 0
		}
		if sy > maxScroll {
			sy = maxScroll
		}
		chatMsgDiv.ScrollTo(0, sy)
	})

	// Chat input enter handler
	chatInput.OnEnter(func(text string) {
		if text != "" {
			addChatMsg("你", text, uimath.ColorHex("#ffffff"))
			chatInput.SetValue("")
		}
	})

	// ── Nameplates (absolute positioned in scene) ────────────────────

	np1 := game.NewNameplate(tree, "暗影守卫", cfg)
	np1.SetLevel(55)
	np1.SetHP(3200, 5000)
	np1.SetType(game.NameplateHostile)
	np1.SetBarSize(100, 6)
	np1.SetPosition(480, 350)
	np1.SetVisible(true)
	np1.SetStyle(absTopLeft(480, 350, 100, 30))
	rootDiv.AppendChild(np1)

	np2 := game.NewNameplate(tree, "旅行商人", cfg)
	np2.SetLevel(30)
	np2.SetHP(3000, 3000)
	np2.SetType(game.NameplateNeutral)
	np2.SetBarSize(90, 5)
	np2.SetPosition(560, 420)
	np2.SetVisible(true)
	np2.SetStyle(absTopLeft(560, 420, 90, 26))
	rootDiv.AppendChild(np2)

	np3 := game.NewNameplate(tree, "神圣牧师", cfg)
	np3.SetLevel(60)
	np3.SetHP(7800, 9500)
	np3.SetType(game.NameplateFriendly)
	np3.SetBarSize(90, 5)
	np3.SetPosition(400, 440)
	np3.SetVisible(true)
	np3.SetStyle(absTopLeft(400, 440, 90, 26))
	rootDiv.AppendChild(np3)

	// ── Inventory (middle-right, draggable) ──────────────────────────

	inv := game.NewInventory(tree, 5, 6, cfg)
	inv.SetTitle("背包")
	inv.SetSlotSize(44)
	inv.SetGap(3)
	for _, it := range []struct {
		slot   int
		name   string
		qty    int
		rarity game.ItemRarity
	}{
		{0, "铁剑", 1, game.RarityUncommon},
		{1, "治疗药水", 5, game.RarityCommon},
		{2, "暗影披风", 1, game.RarityRare},
		{5, "龙鳞盾牌", 1, game.RarityEpic},
		{8, "远古卷轴", 1, game.RarityLegendary},
		{10, "火焰宝石", 3, game.RarityRare},
		{12, "铁矿石", 12, game.RarityCommon},
		{18, "暗影精华", 1, game.RarityEpic},
		{27, "不明碎片", 1, game.RarityLegendary},
	} {
		inv.SetItem(it.slot, &game.ItemData{ID: it.name, Name: it.name, Quantity: it.qty, Rarity: it.rarity})
	}
	invW := float32(6*(44+3)) + 20
	invH := float32(5*(44+3)) + 40
	inv.SetStyle(absMiddleRight(20, invW, invH, 0))
	rootDiv.AppendChild(inv)

	// ── Skill tree (middle-left, draggable) ─────────────────────────

	skillTree := game.NewSkillTree(tree, cfg)
	skillTree.SetPoints(3)
	skillTree.SetNodeSize(44)
	skillTree.AddNode(&game.SkillNode{ID: "fireball", Name: "火球", X: 100, Y: 20, State: game.SkillUnlocked, Level: 3, MaxLevel: 5, Cost: 1})
	skillTree.AddNode(&game.SkillNode{ID: "firewall", Name: "火墙", X: 50, Y: 80, State: game.SkillAvailable, Level: 0, MaxLevel: 3, Cost: 1, Requires: []string{"fireball"}})
	skillTree.AddNode(&game.SkillNode{ID: "meteor", Name: "陨石", X: 150, Y: 80, State: game.SkillLocked, Level: 0, MaxLevel: 1, Cost: 3, Requires: []string{"fireball"}})
	skillTree.AddNode(&game.SkillNode{ID: "icebolt", Name: "冰箭", X: 250, Y: 20, State: game.SkillUnlocked, Level: 2, MaxLevel: 5, Cost: 1})
	skillTree.AddNode(&game.SkillNode{ID: "blizzard", Name: "暴风雪", X: 250, Y: 80, State: game.SkillAvailable, Level: 0, MaxLevel: 3, Cost: 2, Requires: []string{"icebolt"}})
	skillTree.SetStyle(absMiddleLeft(20, 320, 200, 60))
	rootDiv.AppendChild(skillTree)

	// ── Scoreboard (center, draggable) ──────────────────────────────

	scoreboard := game.NewScoreboard(tree, cfg)
	scoreboard.SetTitle("战场统计")
	scoreboard.SetWidth(360)
	scoreboard.AddEntry(game.ScoreEntry{Name: "龙骑士·苍", Score: 28500, Kills: 15, Deaths: 2, Team: 1})
	scoreboard.AddEntry(game.ScoreEntry{Name: "暗影猎手", Score: 22300, Kills: 12, Deaths: 5, Team: 1})
	scoreboard.AddEntry(game.ScoreEntry{Name: "月影法师", Score: 19800, Kills: 8, Deaths: 3, Team: 1})
	scoreboard.AddEntry(game.ScoreEntry{Name: "神圣牧师", Score: 31200, Kills: 2, Deaths: 1, Team: 1})
	scoreboard.AddEntry(game.ScoreEntry{Name: "红莲武士", Score: 21000, Kills: 11, Deaths: 6, Team: 2})
	scoreboard.SortByScore()
	scoreboard.SetVisible(true)
	scoreboard.SetStyle(absMiddleCenter(360, 220, -190, -60))
	rootDiv.AppendChild(scoreboard)

	// ── Loot window (center, draggable) ─────────────────────────────

	loot := game.NewLootWindow(tree, cfg)
	loot.SetTitle("暗影宝箱")
	loot.SetWidth(200)
	loot.AddItem(game.LootItem{Item: &game.ItemData{Name: "暗影之刃", Rarity: game.RarityEpic, Quantity: 1}})
	loot.AddItem(game.LootItem{Item: &game.ItemData{Name: "金币", Rarity: game.RarityCommon, Quantity: 1}, Quantity: 250})
	loot.AddItem(game.LootItem{Item: &game.ItemData{Name: "暗影精华", Rarity: game.RarityRare, Quantity: 1}, Quantity: 3})
	loot.Open()
	loot.SetStyle(absMiddleCenter(200, 180, 100, -40))
	rootDiv.AppendChild(loot)

	// ── Dialogue box (bottom-center, draggable) ─────────────────────

	dialogue := game.NewDialogueBox(tree, cfg)
	dialogue.SetSize(480, 130)
	dialogue.SetChoices([]game.DialogueChoice{
		{Text: "接受任务", OnClick: func() { fmt.Println("[Game] 接受任务") }},
		{Text: "告诉我更多", OnClick: func() { fmt.Println("[Game] 更多信息") }},
		{Text: "还没准备好", OnClick: func() { fmt.Println("[Game] 拒绝") }},
	})
	dialogue.Show("旅行商人·艾瑞克", "旅人，你来得正好。暗影塔的封印正在减弱，你愿意接受这个任务吗？")
	dialogue.SetStyle(absBottomCenter(290, 480, 130, 0))
	rootDiv.AppendChild(dialogue)

	// ── Drag support for HUD panels ─────────────────────────────────

	// Draggable panels and their title bar heights (0 = entire area is draggable)
	type draggablePanel struct {
		w      widget.Widget
		titleH float32
	}
	panels := []draggablePanel{
		{inv, 28},
		{scoreboard, 36},
		{loot, 32},
		{dialogue, 30},
		{skillTree, 0},
		{hpBar, 0},
		{mpBar, 0},
		{xpBar, 0},
		{buffBar, 0},
		{team, 0},
		{hotbar, 0},
		{castBar, 0},
		{minimap, 0},
		{quest, 0},
		{currency, 0},
		{countdown, 0},
		{target, 0},
		{chat, 0},
	}

	var drag dragState
	offsets := make(map[core.ElementID]*dragOffset)

	// Helper to get widget bounds from tree
	boundsOf := func(w widget.Widget) uimath.Rect {
		if elem := tree.Get(w.ElementID()); elem != nil {
			return elem.Layout().Bounds
		}
		return uimath.Rect{}
	}

	rootDiv.On(event.MouseDown, func(e *event.Event) {
		// Item drag takes priority over window drag
		if inv.HandleMouseDown(e.GlobalX, e.GlobalY) {
			return
		}
		// Check panels in reverse order (last = top-most)
		for i := len(panels) - 1; i >= 0; i-- {
			p := panels[i]
			b := boundsOf(p.w)
			if b.IsEmpty() {
				continue
			}
			if e.GlobalX >= b.X && e.GlobalX < b.X+b.Width &&
				e.GlobalY >= b.Y && e.GlobalY < b.Y+b.Height {
				// If titleH is set, only the title bar starts drag
				if p.titleH > 0 && e.GlobalY > b.Y+p.titleH {
					continue
				}
				// Bring to front
				rootDiv.BringChildToFront(p.w)
				// Move panel entry to end of our tracking list too
				el := panels[i]
				copy(panels[i:], panels[i+1:])
				panels[len(panels)-1] = el

				drag.active = true
				drag.target = p.w
				drag.startX = e.GlobalX
				drag.startY = e.GlobalY
				drag.origX = b.X
				drag.origY = b.Y
				return
			}
		}
	})

	rootDiv.On(event.MouseMove, func(e *event.Event) {
		if inv.HandleMouseMove(e.GlobalX, e.GlobalY) {
			return
		}
		if !drag.active {
			return
		}
		dx := e.GlobalX - drag.startX
		dy := e.GlobalY - drag.startY
		newX := drag.origX + dx
		newY := drag.origY + dy
		// Update immediately via tree.SetLayout for responsive dragging
		b := boundsOf(drag.target)
		tree.SetLayout(drag.target.ElementID(), core.LayoutResult{
			Bounds: uimath.NewRect(newX, newY, b.Width, b.Height),
		})
	})

	rootDiv.On(event.MouseUp, func(e *event.Event) {
		inv.HandleMouseUp(e.GlobalX, e.GlobalY)
		if drag.active {
			eid := drag.target.ElementID()
			if offsets[eid] == nil {
				offsets[eid] = &dragOffset{}
			}
			// Store cumulative drag offset (applied after CSSLayout each frame)
			offsets[eid].dx += e.GlobalX - drag.startX
			offsets[eid].dy += e.GlobalY - drag.startY
			drag.active = false
			drag.target = nil
		}
	})

	// ── Layout + animation ───────────────────────────────────────────

	layoutCache := ui.NewCSSLayoutCache()
	frameN := 0
	app.SetOnLayout(func(tree *core.Tree, root widget.Widget, w, h float32) {
		// Cached CSS layout — only rebuilds when tree structure or viewport changes.
		layoutCache.Layout(tree, root, w, h, cfg)

		// Apply persistent drag offsets after CSSLayout
		for eid, off := range offsets {
			if off.dx == 0 && off.dy == 0 {
				continue
			}
			if elem := tree.Get(eid); elem != nil {
				b := elem.Layout().Bounds
				tree.SetLayout(eid, core.LayoutResult{
					Bounds: uimath.NewRect(b.X+off.dx, b.Y+off.dy, b.Width, b.Height),
				})
			}
		}

		// ── Animate ──
		frameN++
		t := float32(frameN) * 0.016

		castBar.Tick(0.008)

		hpBar.SetCurrent(float32(780 + 50*math.Sin(float64(t*0.8))))
		mpBar.SetCurrent(float32(350 + 80*math.Sin(float64(t*0.5))))

		bossHP := float32(185000) - float32(frameN)*10
		if bossHP < 50000 {
			bossHP = 185000
			frameN = 0
		}
		target.SetTarget(&game.UnitFrameData{
			Name: "暗影领主·莫德雷克", Level: 62, HP: bossHP, HPMax: 500000, MP: 80000, MPMax: 80000, Class: "Boss",
		})

		countdown.Tick(0.016)

		// Mark paint-dirty for animations (does NOT trigger layout rebuild).
		tree.MarkDirty(tree.Root())
	})

	fmt.Println("[Game] RPG HUD demo running — press Ctrl+C to exit")

	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
