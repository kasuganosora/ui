// Game UI Demo — Showcases GoUI's game widget library in an RPG-style HUD.
//
// Demonstrates:
//   - HUD overlay with anchored elements (health/mana bars, hotbar, minimap)
//   - Inventory grid with rarity-colored items
//   - Chat box with message history
//   - Cast bar with real-time progress
//   - Buff/debuff bar with duration tracking
//   - Quest tracker, unit frames, scoreboard
//   - Skill tree with prerequisites
//   - Nameplate system, currency display, countdown timer
//   - Dialogue box with NPC choices
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
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/widget"
	"github.com/kasuganosora/ui/widget/game"
)

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

	// Use a minimal root; height:100% doesn't resolve correctly as CSS layout root,
	// so we override the root bounds explicitly in SetOnLayout below.
	doc := app.LoadHTML(`<div style="background:#0a0e17;"></div>`)
	rootDiv := doc.Root.Children()[0].(*widget.Div)

	// ── HUD ──────────────────────────────────────────────────────────

	hud := game.NewHUD(tree, cfg)

	// ── Health / Mana / XP bars (top-left) ───────────────────────────

	hpBar := game.NewHealthBar(tree, cfg)
	hpBar.SetCurrent(780)
	hpBar.SetMax(1200)
	hpBar.SetBarColor(uimath.ColorHex("#52c41a"))
	hpBar.SetShowText(true)
	hpBar.SetSize(220, 22)
	tree.SetLayout(hpBar.ElementID(), core.LayoutResult{Bounds: uimath.NewRect(0, 0, 220, 22)})
	hud.AddElementDraggable(hpBar, game.AnchorTopLeft, 20, 20)

	mpBar := game.NewHealthBar(tree, cfg)
	mpBar.SetCurrent(350)
	mpBar.SetMax(600)
	mpBar.SetBarColor(uimath.ColorHex("#1890ff"))
	mpBar.SetShowText(true)
	mpBar.SetSize(220, 18)
	tree.SetLayout(mpBar.ElementID(), core.LayoutResult{Bounds: uimath.NewRect(0, 0, 220, 18)})
	hud.AddElementDraggable(mpBar, game.AnchorTopLeft, 20, 48)

	xpBar := game.NewHealthBar(tree, cfg)
	xpBar.SetCurrent(4200)
	xpBar.SetMax(8500)
	xpBar.SetBarColor(uimath.ColorHex("#a335ee"))
	xpBar.SetShowText(true)
	xpBar.SetSize(220, 12)
	tree.SetLayout(xpBar.ElementID(), core.LayoutResult{Bounds: uimath.NewRect(0, 0, 220, 12)})
	hud.AddElementDraggable(xpBar, game.AnchorTopLeft, 20, 72)

	// ── Buff bar (below bars) ────────────────────────────────────────

	buffBar := game.NewBuffBar(tree, cfg)
	buffBar.SetIconSize(28)
	buffBar.SetGap(3)
	buffBar.AddBuff(game.Buff{ID: "str", Label: "力", Duration: 120, Type: game.BuffPositive})
	buffBar.AddBuff(game.Buff{ID: "haste", Label: "速", Duration: 45, Type: game.BuffPositive})
	buffBar.AddBuff(game.Buff{ID: "shield", Label: "盾", Duration: 30, Type: game.BuffPositive})
	buffBar.AddBuff(game.Buff{ID: "poison", Label: "毒", Duration: 8, Type: game.BuffNegative})
	buffBar.AddBuff(game.Buff{ID: "slow", Label: "慢", Duration: 15, Type: game.BuffNegative})
	tree.SetLayout(buffBar.ElementID(), core.LayoutResult{Bounds: uimath.NewRect(0, 0, 300, 28)})
	hud.AddElementDraggable(buffBar, game.AnchorTopLeft, 20, 94)

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
	tree.SetLayout(hotbar.ElementID(), core.LayoutResult{Bounds: uimath.NewRect(0, 0, 10*(52+4), 52)})
	hud.AddElementDraggable(hotbar, game.AnchorBottomCenter, 0, -20)

	// ── Cast bar (above hotbar) ──────────────────────────────────────

	castBar := game.NewCastBar(tree, cfg)
	castBar.SetSize(280, 22)
	castBar.SetColor(uimath.ColorHex("#ffd700"))
	castBar.StartCast("火球术", 3.0)
	castBar.Tick(1.5) // start at 50%
	tree.SetLayout(castBar.ElementID(), core.LayoutResult{Bounds: uimath.NewRect(0, 0, 280, 22)})
	hud.AddElementDraggable(castBar, game.AnchorBottomCenter, 0, -82)

	// ── Minimap (top-right) ──────────────────────────────────────────

	minimap := game.NewMinimap(tree, cfg)
	minimap.SetSize(160)
	minimap.SetCircular(true)
	minimap.SetPlayerPos(80, 80)
	minimap.SetPlayerRotation(0.4)
	minimap.AddMarker(game.MinimapMarker{X: 40, Y: 50, Color: uimath.ColorHex("#ff4444"), Size: 6, Label: "!"})
	minimap.AddMarker(game.MinimapMarker{X: 120, Y: 30, Color: uimath.ColorHex("#ffdd44"), Size: 5, Label: "?"})
	minimap.AddMarker(game.MinimapMarker{X: 60, Y: 130, Color: uimath.ColorHex("#44aaff"), Size: 5})
	tree.SetLayout(minimap.ElementID(), core.LayoutResult{Bounds: uimath.NewRect(0, 0, 160, 160)})
	hud.AddElementDraggable(minimap, game.AnchorTopRight, -20, 20)

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
	tree.SetLayout(quest.ElementID(), core.LayoutResult{Bounds: uimath.NewRect(0, 0, 230, 200)})
	hud.AddElementDraggable(quest, game.AnchorTopRight, -20, 200)

	// ── Currency display (top-center) ────────────────────────────────

	currency := game.NewCurrencyDisplay(tree, cfg)
	currency.SetGap(16)
	currency.AddCurrency(game.CurrencyEntry{Symbol: "G", Amount: 12580, Color: uimath.ColorHex("#ffd700")})
	currency.AddCurrency(game.CurrencyEntry{Symbol: "◆", Amount: 350, Color: uimath.ColorHex("#44aaff")})
	currency.AddCurrency(game.CurrencyEntry{Symbol: "★", Amount: 28, Color: uimath.ColorHex("#ff8800")})
	tree.SetLayout(currency.ElementID(), core.LayoutResult{Bounds: uimath.NewRect(0, 0, 300, 24)})
	hud.AddElementDraggable(currency, game.AnchorTopCenter, 0, 8)

	// ── Countdown timer ──────────────────────────────────────────────

	countdown := game.NewCountdownTimer(tree, cfg)
	countdown.SetSeconds(185)
	countdown.SetLabel("Boss 刷新")
	countdown.SetColor(uimath.ColorWhite)
	tree.SetLayout(countdown.ElementID(), core.LayoutResult{Bounds: uimath.NewRect(0, 0, 200, 30)})
	hud.AddElementDraggable(countdown, game.AnchorTopCenter, 0, 36)

	// ── Target frame (top-center) ────────────────────────────────────

	target := game.NewTargetFrame(tree, cfg)
	target.SetSize(220, 52)
	target.SetTarget(&game.UnitFrameData{
		Name: "暗影领主·莫德雷克", Level: 62, HP: 185000, HPMax: 500000, MP: 80000, MPMax: 80000, Class: "Boss",
	})
	tree.SetLayout(target.ElementID(), core.LayoutResult{Bounds: uimath.NewRect(0, 0, 220, 52)})
	hud.AddElementDraggable(target, game.AnchorTopCenter, -140, 68)

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
	tree.SetLayout(team.ElementID(), core.LayoutResult{Bounds: uimath.NewRect(0, 0, 170, 4*47)})
	hud.AddElementDraggable(team, game.AnchorTopLeft, 20, 132)

	// ── Chat box (bottom-left) ──────────────────────────────────────

	chat := game.NewChatBox(tree, cfg)
	chat.SetSize(340, 200)
	chat.SetMaxVisible(8)
	tree.SetLayout(chat.ElementID(), core.LayoutResult{Bounds: uimath.NewRect(0, 0, 340, 200)})
	hud.AddElementDraggable(chat, game.AnchorBottomLeft, 10, -82)

	// Chat message area — scrollable Div with Text children (framework controls)
	chatMsgDiv := widget.NewDiv(tree, cfg)
	chatMsgDiv.SetBgColor(uimath.RGBA(0, 0, 0, 0)) // transparent, ChatBox draws background
	chatMsgDiv.SetScrollable(true)
	rootDiv.AppendChild(chatMsgDiv)

	// Helper to add a chat message as a Text widget
	lineH := float32(18) // approximate line height for small font
	addChatMsg := func(sender, text string, color uimath.Color) {
		chat.AddMessage(game.ChatMessage{Sender: sender, Text: text, Color: color})
		msgText := widget.NewText(tree, "["+sender+"] "+text, cfg)
		msgText.SetColor(color)
		msgText.SetFontSize(cfg.FontSizeSm)
		chatMsgDiv.AppendChild(msgText)
		// Update content height for scroll
		n := float32(len(chat.Messages()))
		chatMsgDiv.SetContentHeight(n * lineH)
		// Auto-scroll to bottom
		mb := chat.MessageBounds()
		maxScroll := n*lineH - mb.Height
		if maxScroll > 0 {
			chatMsgDiv.ScrollTo(0, maxScroll)
		}
	}

	addChatMsg("系统", "欢迎来到暗影之境！", uimath.ColorHex("#ffd700"))
	addChatMsg("骑士", "队伍已就绪，准备进攻北塔", uimath.ColorHex("#44aaff"))
	addChatMsg("法师", "我的蓝不多了", uimath.ColorHex("#44aaff"))
	addChatMsg("世界", "LFG 副本·迷雾深渊 4/5 缺T", uimath.ColorHex("#ffaa00"))
	addChatMsg("治疗", "注意躲地板火！", uimath.ColorHex("#52c41a"))

	// Chat message area scroll via MouseWheel
	chatMsgDiv.On(event.MouseWheel, func(e *event.Event) {
		chatMsgDiv.ScrollTo(0, chatMsgDiv.ScrollY()-e.WheelDY*30)
		mb := chat.MessageBounds()
		maxScroll := chatMsgDiv.ContentHeight() - mb.Height
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

	// Chat input (real Input widget for keyboard support)
	chatInput := widget.NewInput(tree, cfg)
	chatInput.SetPlaceholder("输入消息...")
	chatInput.SetBorderless(true)
	chatInput.OnEnter(func(text string) {
		if text != "" {
			addChatMsg("你", text, uimath.ColorHex("#ffffff"))
			chatInput.SetValue("")
		}
	})
	rootDiv.AppendChild(chatInput)

	// ── Nameplates (absolute positioned in scene) ────────────────────
	// Added to HUD BEFORE draggable panels so panels draw on top of them.

	np1 := game.NewNameplate(tree, "暗影守卫", cfg)
	np1.SetLevel(55)
	np1.SetHP(3200, 5000)
	np1.SetType(game.NameplateHostile)
	np1.SetBarSize(100, 6)
	np1.SetPosition(480, 350)
	np1.SetVisible(true)
	tree.SetLayout(np1.ElementID(), core.LayoutResult{Bounds: uimath.NewRect(0, 0, 100, 30)})
	hud.AddElement(np1, game.AnchorTopLeft, 0, 0)

	np2 := game.NewNameplate(tree, "旅行商人", cfg)
	np2.SetLevel(30)
	np2.SetHP(3000, 3000)
	np2.SetType(game.NameplateNeutral)
	np2.SetBarSize(90, 5)
	np2.SetPosition(560, 420)
	np2.SetVisible(true)
	tree.SetLayout(np2.ElementID(), core.LayoutResult{Bounds: uimath.NewRect(0, 0, 90, 26)})
	hud.AddElement(np2, game.AnchorTopLeft, 0, 0)

	np3 := game.NewNameplate(tree, "神圣牧师", cfg)
	np3.SetLevel(60)
	np3.SetHP(7800, 9500)
	np3.SetType(game.NameplateFriendly)
	np3.SetBarSize(90, 5)
	np3.SetPosition(400, 440)
	np3.SetVisible(true)
	tree.SetLayout(np3.ElementID(), core.LayoutResult{Bounds: uimath.NewRect(0, 0, 90, 26)})
	hud.AddElement(np3, game.AnchorTopLeft, 0, 0)

	// ── Inventory (center-right) ─────────────────────────────────────

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
	tree.SetLayout(inv.ElementID(), core.LayoutResult{Bounds: uimath.NewRect(0, 0, invW, invH)})
	hud.AddElementDraggable(inv, game.AnchorMiddleRight, -20, 0, 28) // titleH=28

	// ── Skill tree (middle-left) ─────────────────────────────────────

	skillTree := game.NewSkillTree(tree, cfg)
	skillTree.SetPoints(3)
	skillTree.SetNodeSize(44)
	skillTree.AddNode(&game.SkillNode{ID: "fireball", Name: "火球", X: 100, Y: 20, State: game.SkillUnlocked, Level: 3, MaxLevel: 5, Cost: 1})
	skillTree.AddNode(&game.SkillNode{ID: "firewall", Name: "火墙", X: 50, Y: 80, State: game.SkillAvailable, Level: 0, MaxLevel: 3, Cost: 1, Requires: []string{"fireball"}})
	skillTree.AddNode(&game.SkillNode{ID: "meteor", Name: "陨石", X: 150, Y: 80, State: game.SkillLocked, Level: 0, MaxLevel: 1, Cost: 3, Requires: []string{"fireball"}})
	skillTree.AddNode(&game.SkillNode{ID: "icebolt", Name: "冰箭", X: 250, Y: 20, State: game.SkillUnlocked, Level: 2, MaxLevel: 5, Cost: 1})
	skillTree.AddNode(&game.SkillNode{ID: "blizzard", Name: "暴风雪", X: 250, Y: 80, State: game.SkillAvailable, Level: 0, MaxLevel: 3, Cost: 2, Requires: []string{"icebolt"}})
	tree.SetLayout(skillTree.ElementID(), core.LayoutResult{Bounds: uimath.NewRect(0, 0, 320, 200)})
	hud.AddElementDraggable(skillTree, game.AnchorMiddleLeft, 20, 60)

	// ── Scoreboard (center) ──────────────────────────────────────────

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
	tree.SetLayout(scoreboard.ElementID(), core.LayoutResult{Bounds: uimath.NewRect(0, 0, 360, 220)})
	hud.AddElementDraggable(scoreboard, game.AnchorMiddleCenter, -190, -60, 36) // headerH=36

	// ── Loot window ──────────────────────────────────────────────────

	loot := game.NewLootWindow(tree, cfg)
	loot.SetTitle("暗影宝箱")
	loot.SetWidth(200)
	loot.AddItem(game.LootItem{Item: &game.ItemData{Name: "暗影之刃", Rarity: game.RarityEpic, Quantity: 1}})
	loot.AddItem(game.LootItem{Item: &game.ItemData{Name: "金币", Rarity: game.RarityCommon, Quantity: 1}, Quantity: 250})
	loot.AddItem(game.LootItem{Item: &game.ItemData{Name: "暗影精华", Rarity: game.RarityRare, Quantity: 1}, Quantity: 3})
	loot.Open()
	tree.SetLayout(loot.ElementID(), core.LayoutResult{Bounds: uimath.NewRect(0, 0, 200, 180)})
	hud.AddElementDraggable(loot, game.AnchorMiddleCenter, 100, -40, 32) // titleH=32

	// ── Dialogue box ─────────────────────────────────────────────────

	dialogue := game.NewDialogueBox(tree, cfg)
	dialogue.SetSize(480, 130)
	dialogue.SetChoices([]game.DialogueChoice{
		{Text: "接受任务", OnClick: func() { fmt.Println("[Game] 接受任务") }},
		{Text: "告诉我更多", OnClick: func() { fmt.Println("[Game] 更多信息") }},
		{Text: "还没准备好", OnClick: func() { fmt.Println("[Game] 拒绝") }},
	})
	dialogue.Show("旅行商人·艾瑞克", "旅人，你来得正好。暗影塔的封印正在减弱，你愿意接受这个任务吗？")
	tree.SetLayout(dialogue.ElementID(), core.LayoutResult{Bounds: uimath.NewRect(0, 0, 480, 130)})
	hud.AddElementDraggable(dialogue, game.AnchorBottomCenter, 0, -290, 30) // titleH=30

	// ── Attach HUD to document root ─────────────────────────────────

	rootDiv.AppendChild(hud)

	// ── Drag support for HUD panels ──────────────────────────────────

	rootDiv.On(event.MouseDown, func(e *event.Event) {
		// Item drag takes priority over window drag
		if inv.HandleMouseDown(e.GlobalX, e.GlobalY) {
			return
		}
		hud.HandleMouseDown(e.GlobalX, e.GlobalY)
	})
	rootDiv.On(event.MouseMove, func(e *event.Event) {
		if inv.HandleMouseMove(e.GlobalX, e.GlobalY) {
			return
		}
		if hud.HandleMouseMove(e.GlobalX, e.GlobalY) {
			// Re-layout HUD elements immediately so the panel follows the cursor
			w, h := float32(0), float32(0)
			if elem := tree.Get(rootDiv.ElementID()); elem != nil {
				b := elem.Layout().Bounds
				w, h = b.Width, b.Height
			}
			if w > 0 && h > 0 {
				hud.LayoutElements(w, h)
			}
		}
	})
	rootDiv.On(event.MouseUp, func(e *event.Event) {
		inv.HandleMouseUp(e.GlobalX, e.GlobalY)
		hud.HandleMouseUp()
	})

	// ── Layout + animation ───────────────────────────────────────────

	frameN := 0
	app.SetOnLayout(func(tree *core.Tree, root widget.Widget, w, h float32) {
		ui.CSSLayout(tree, root, w, h, cfg)

		// Force root + rootDiv to fill viewport (CSSLayout can't resolve height:auto for game overlay)
		tree.SetLayout(root.ElementID(), core.LayoutResult{
			Bounds: uimath.NewRect(0, 0, w, h),
		})
		tree.SetLayout(rootDiv.ElementID(), core.LayoutResult{
			Bounds: uimath.NewRect(0, 0, w, h),
		})

		// HUD positions all anchored elements (including nameplates)
		hud.LayoutElements(w, h)

		// Position chat message area and input at the chat box bounds
		mb := chat.MessageBounds()
		tree.SetLayout(chatMsgDiv.ElementID(), core.LayoutResult{
			Bounds: mb,
		})
		// Layout message Text children vertically within the scrollable Div
		msgY := -chatMsgDiv.ScrollY()
		for _, child := range chatMsgDiv.Children() {
			tree.SetLayout(child.ElementID(), core.LayoutResult{
				Bounds: uimath.NewRect(mb.X+4, mb.Y+msgY, mb.Width-12, lineH),
			})
			msgY += lineH
		}
		ib := chat.InputBounds()
		tree.SetLayout(chatInput.ElementID(), core.LayoutResult{
			Bounds: ib,
		})

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

		tree.MarkDirty(tree.Root())
	})

	fmt.Println("[Game] RPG HUD demo running — press Ctrl+C to exit")

	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
