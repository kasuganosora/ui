// Game UI Demo — Showcases GoUI's game widget library in an RPG-style HUD.
//
// Uses HTML+CSS layout (position:absolute) for all HUD element placement.
// Game widgets participate in CSSLayout via their Style() properties.
//
// Demonstrates:
//   - HUD overlay with CSS-positioned elements (health/mana bars, hotbar, minimap)
//   - Inventory grid with rarity-colored items and drag-and-drop
//   - Chat box with scrollable Div + Input (framework controls) inside a Window
//   - Cast bar with real-time progress
//   - Buff/debuff bar with duration tracking
//   - Quest tracker, unit frames, scoreboard
//   - Skill tree with prerequisites
//   - Nameplate system, currency display, countdown timer
//   - Dialogue box with NPC choices
//   - Draggable Windows with title bar and close button
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

	doc := app.LoadHTML(`<div style="position:relative; width:100%; height:100%; background:#0a0e17;"></div>`)
	rootDiv := doc.Root.Children()[0].(*widget.Div)

	// ── Helper: chromeless draggable wrapper ────────────────────────
	// Wraps a widget in a chromeless Window managed by WindowManager.
	// Returns the Window so caller can reference it.
	var wm *game.WindowManager // forward declare, initialized below

	chromeless := func(child widget.Widget, x, y, w, h float32) *game.Window {
		win := game.NewWindow(tree, "", cfg)
		win.SetChromeless(true)
		win.SetShowClose(false)
		win.SetShadow(false)
		child.SetStyle(layout.Style{Width: layout.Px(w), Height: layout.Px(h)})
		win.AppendChild(child)
		wm.Add(win, x, y, w, h)
		return win
	}

	// chromelessGroup wraps multiple children in a flex-column Div inside a chromeless Window.
	chromelessGroup := func(x, y, w, h float32, children ...widget.Widget) *game.Window {
		container := widget.NewDiv(tree, cfg)
		container.SetStyle(layout.Style{
			Display:       layout.DisplayFlex,
			FlexDirection: layout.FlexDirectionColumn,
			Width:         layout.Px(w),
			Height:        layout.Px(h),
		})
		for _, c := range children {
			container.AppendChild(c)
		}
		win := game.NewWindow(tree, "", cfg)
		win.SetChromeless(true)
		win.SetShowClose(false)
		win.SetShadow(false)
		container.SetStyle(layout.Style{Width: layout.Px(w), Height: layout.Px(h)})
		win.AppendChild(container)
		wm.Add(win, x, y, w, h)
		return win
	}

	// ══════════════════════════════════════════════════════════════════
	// WindowManager — manages all windows and HUD elements
	// ══════════════════════════════════════════════════════════════════

	wm = game.NewWindowManager(tree, rootDiv)

	// ── Player status panel (name + HP/MP/XP) ──────────────────────

	playerName := widget.NewText(tree, "Lv.60 龙骑士·苍", cfg)
	playerName.SetColor(uimath.ColorHex("#ffd700"))
	playerName.SetFontSize(cfg.FontSizeSm)
	playerName.SetStyle(layout.Style{
		Width: layout.Px(220), Height: layout.Px(16),
		Margin: layout.EdgeValues{Bottom: layout.Px(5)},
	})

	hpBar := game.NewHealthBar(tree, cfg)
	hpBar.SetCurrent(780)
	hpBar.SetMax(1200)
	hpBar.SetBarColor(uimath.ColorHex("#52c41a"))
	hpBar.SetShowText(true)
	hpBar.SetSize(220, 22)
	hpBar.SetStyle(layout.Style{
		Width: layout.Px(220), Height: layout.Px(22),
		Margin: layout.EdgeValues{Bottom: layout.Px(5)},
	})

	mpBar := game.NewHealthBar(tree, cfg)
	mpBar.SetCurrent(350)
	mpBar.SetMax(600)
	mpBar.SetBarColor(uimath.ColorHex("#1890ff"))
	mpBar.SetShowText(true)
	mpBar.SetSize(220, 18)
	mpBar.SetStyle(layout.Style{
		Width: layout.Px(220), Height: layout.Px(18),
		Margin: layout.EdgeValues{Bottom: layout.Px(5)},
	})

	xpBar := game.NewHealthBar(tree, cfg)
	xpBar.SetCurrent(4200)
	xpBar.SetMax(8500)
	xpBar.SetBarColor(uimath.ColorHex("#a335ee"))
	xpBar.SetShowText(true)
	xpBar.SetSize(220, 12)
	xpBar.SetStyle(layout.Style{Width: layout.Px(220), Height: layout.Px(12)})

	// 16 name + 5 + 22 hp + 5 + 18 mp + 5 + 12 xp = 83
	chromelessGroup(20, 20, 220, 83, playerName, hpBar, mpBar, xpBar)

	// ── Buff / Debuff bars (separate, below status) ─────────────────

	allBuffs := []game.Buff{
		{ID: "str", Label: "力", Duration: 120, Type: game.BuffPositive},
		{ID: "haste", Label: "速", Duration: 45, Type: game.BuffPositive},
		{ID: "shield", Label: "盾", Duration: 30, Type: game.BuffPositive},
		{ID: "poison", Label: "毒", Duration: 8, Type: game.BuffNegative},
		{ID: "slow", Label: "慢", Duration: 15, Type: game.BuffNegative},
	}

	buffBarPos := game.NewBuffBar(tree, cfg)
	buffBarPos.SetIconSize(28)
	buffBarPos.SetGap(3)
	buffBarPos.SetFilter(game.BuffPositive)
	for _, b := range allBuffs {
		buffBarPos.AddBuff(b)
	}
	buffBarPos.OnCancel(func(id string) {
		buffBarPos.RemoveBuff(id)
		fmt.Printf("[Game] 取消增益: %s\n", id)
	})
	buffBarPos.SetStyle(layout.Style{
		Width: layout.Px(220), Height: layout.Px(28),
		Margin: layout.EdgeValues{Bottom: layout.Px(5)},
	})

	buffBarNeg := game.NewBuffBar(tree, cfg)
	buffBarNeg.SetIconSize(28)
	buffBarNeg.SetGap(3)
	buffBarNeg.SetFilter(game.BuffNegative)
	for _, b := range allBuffs {
		buffBarNeg.AddBuff(b)
	}
	buffBarNeg.SetStyle(layout.Style{Width: layout.Px(220), Height: layout.Px(28)})

	// 28 buffs + 5 + 28 debuffs = 61
	chromelessGroup(20, 108, 220, 61, buffBarPos, buffBarNeg)

	// ── Team frames (left side) ─────────────────────────────────────

	team := game.NewTeamFrame(tree, cfg)
	team.SetFrameSize(170, 44)
	team.SetGap(3)
	team.SetMembers([]game.UnitFrameData{
		{Name: "龙骑士·苍", Level: 60, HP: 11500, HPMax: 15000, MP: 2800, MPMax: 3000, Class: "战士"},
		{Name: "月影法师", Level: 59, HP: 6200, HPMax: 8000, MP: 1200, MPMax: 6000, Class: "法师"},
		{Name: "神圣牧师", Level: 60, HP: 7800, HPMax: 9500, MP: 4500, MPMax: 7000, Class: "牧师"},
		{Name: "暗影猎手", Level: 58, HP: 0, HPMax: 7500, MP: 2000, MPMax: 3500, Class: "猎人", Dead: true},
	})
	chromeless(team, 20, 178, 170, 4*47)

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
	hotbarW := float32(10 * (52 + 4))
	chromeless(hotbar, 640-hotbarW/2, 728, hotbarW, 52)

	// ── Cast bar (above hotbar) ─────────────────────────────────────

	castBar := game.NewCastBar(tree, cfg)
	castBar.SetSize(280, 22)
	castBar.SetColor(uimath.ColorHex("#ffd700"))
	castBar.StartCast("火球术", 3.0)
	castBar.Tick(1.5)
	chromeless(castBar, 640-140, 696, 280, 22)

	// ── Minimap (top-right) ─────────────────────────────────────────

	minimap := game.NewMinimap(tree, cfg)
	minimap.SetSize(160)
	minimap.SetCircular(true)
	minimap.SetPlayerPos(80, 80)
	minimap.SetPlayerRotation(0.4)
	minimap.AddMarker(game.MinimapMarker{X: 40, Y: 50, Color: uimath.ColorHex("#ff4444"), Size: 6, Label: "!"})
	minimap.AddMarker(game.MinimapMarker{X: 120, Y: 30, Color: uimath.ColorHex("#ffdd44"), Size: 5, Label: "?"})
	minimap.AddMarker(game.MinimapMarker{X: 60, Y: 130, Color: uimath.ColorHex("#44aaff"), Size: 5})
	chromeless(minimap, 1100, 20, 160, 160)

	// ── Quest tracker (right, below minimap) ────────────────────────

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
	chromeless(quest, 1030, 200, 230, 200)

	// ── Currency display (top-center) ───────────────────────────────

	currency := game.NewCurrencyDisplay(tree, cfg)
	currency.SetGap(16)
	currency.AddCurrency(game.CurrencyEntry{Symbol: "G", Amount: 12580, Color: uimath.ColorHex("#ffd700")})
	currency.AddCurrency(game.CurrencyEntry{Symbol: "◆", Amount: 350, Color: uimath.ColorHex("#44aaff")})
	currency.AddCurrency(game.CurrencyEntry{Symbol: "★", Amount: 28, Color: uimath.ColorHex("#ff8800")})
	chromeless(currency, 490, 8, 300, 24)

	// ── Countdown timer ─────────────────────────────────────────────

	countdown := game.NewCountdownTimer(tree, cfg)
	countdown.SetSeconds(185)
	countdown.SetLabel("Boss 刷新")
	countdown.SetColor(uimath.ColorWhite)
	chromeless(countdown, 540, 36, 200, 30)

	// ── Target frame (top-center) ───────────────────────────────────

	target := game.NewTargetFrame(tree, cfg)
	target.SetSize(220, 52)
	target.SetTarget(&game.UnitFrameData{
		Name: "暗影领主·莫德雷克", Level: 62, HP: 185000, HPMax: 500000, MP: 80000, MPMax: 80000, Class: "Boss",
	})
	chromeless(target, 390, 68, 220, 52)

	// ── Nameplates (absolute positioned in scene — not draggable) ───

	np1 := game.NewNameplate(tree, "暗影守卫", cfg)
	np1.SetLevel(55)
	np1.SetHP(3200, 5000)
	np1.SetType(game.NameplateHostile)
	np1.SetBarSize(100, 6)
	np1.SetPosition(480, 350)
	np1.SetVisible(true)
	chromeless(np1, 480, 350, 100, 30)

	np2 := game.NewNameplate(tree, "旅行商人", cfg)
	np2.SetLevel(30)
	np2.SetHP(3000, 3000)
	np2.SetType(game.NameplateNeutral)
	np2.SetBarSize(90, 5)
	np2.SetPosition(560, 420)
	np2.SetVisible(true)
	chromeless(np2, 560, 420, 90, 26)

	np3 := game.NewNameplate(tree, "神圣牧师", cfg)
	np3.SetLevel(60)
	np3.SetHP(7800, 9500)
	np3.SetType(game.NameplateFriendly)
	np3.SetBarSize(90, 5)
	np3.SetPosition(400, 440)
	np3.SetVisible(true)
	chromeless(np3, 400, 440, 90, 26)

	// ── Windowed components (with title bar) ────────────────────────

	// ── Chat Window (bottom-left) ────────────────────────────────────

	chatWin := game.NewWindow(tree, "聊天", cfg)
	chatWin.SetTitleH(24)
	chatWinW := float32(340)
	chatWinH := float32(228) // 24 title + 172 messages + 4 gap + 28 input

	// Scrollable message area (Div inside chat window)
	chatMsgDiv := widget.NewDiv(tree, cfg)
	chatMsgDiv.SetStyle(layout.Style{
		Width:    layout.Px(chatWinW),
		Height:   layout.Px(172),
		Overflow: layout.OverflowAuto,
	})
	chatWin.AppendChild(chatMsgDiv)

	// Chat input (inside chat window)
	chatInput := widget.NewInput(tree, cfg)
	chatInput.SetPlaceholder("输入消息...")
	chatInput.SetBorderless(true)
	chatInput.SetStyle(layout.Style{
		Width:  layout.Px(chatWinW),
		Height: layout.Px(28),
	})
	chatWin.AppendChild(chatInput)

	wm.Add(chatWin, 10, 520, chatWinW, chatWinH)

	// Chat message helper
	chat := game.NewChatBox(tree, cfg) // data model only, not appended to tree
	chat.SetSize(chatWinW, 200)
	chat.SetMaxVisible(8)

	lineH := float32(18)
	addChatMsg := func(sender, text string, color uimath.Color) {
		chat.AddMessage(game.ChatMessage{Sender: sender, Text: text, Color: color})
		msgText := widget.NewText(tree, "["+sender+"] "+text, cfg)
		msgText.SetColor(color)
		msgText.SetFontSize(cfg.FontSizeSm)
		chatMsgDiv.AppendChild(msgText)
		n := float32(len(chat.Messages()))
		chatMsgDiv.SetContentHeight(n * lineH)
		maxScroll := n*lineH - 172
		if maxScroll > 0 {
			chatMsgDiv.ScrollTo(0, maxScroll)
		}
	}

	addChatMsg("系统", "欢迎来到暗影之境！", uimath.ColorHex("#ffd700"))
	addChatMsg("骑士", "队伍已就绪，准备进攻北塔", uimath.ColorHex("#44aaff"))
	addChatMsg("法师", "我的蓝不多了", uimath.ColorHex("#44aaff"))
	addChatMsg("世界", "LFG 副本·迷雾深渊 4/5 缺T", uimath.ColorHex("#ffaa00"))
	addChatMsg("治疗", "注意躲地板火！", uimath.ColorHex("#52c41a"))

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

	chatInput.OnEnter(func(text string) {
		if text != "" {
			addChatMsg("你", text, uimath.ColorHex("#ffffff"))
			chatInput.SetValue("")
		}
	})

	// ── Inventory Window (middle-right) ──────────────────────────────

	inv := game.NewInventory(tree, 5, 6, cfg)
	inv.SetTitle("")
	inv.SetEmbedded(true)
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
	invContentW := float32(6*(44+3)) + 20
	invContentH := float32(5*(44+3)) + 12

	invWin := game.NewWindow(tree, "背包", cfg)
	inv.SetStyle(layout.Style{Width: layout.Px(invContentW), Height: layout.Px(invContentH)})
	invWin.AppendChild(inv)
	wm.Add(invWin, 980, 250, invContentW, 0) // auto-height

	// Mouse events: WindowManager handles drag/z-order/close, then inv handles item DnD
	rootDiv.On(event.MouseDown, func(e *event.Event) {
		// Right-click: cancel positive buffs
		if e.Button == event.MouseButtonRight {
			buffBarPos.HandleRightClick(e.GlobalX, e.GlobalY)
			return
		}
		wm.HandleMouseDown(e.GlobalX, e.GlobalY)
		if !wm.IsDragging() {
			inv.HandleMouseDown(e.GlobalX, e.GlobalY)
		}
	})
	rootDiv.On(event.MouseMove, func(e *event.Event) {
		wm.HandleMouseMove(e.GlobalX, e.GlobalY)
		if !wm.IsDragging() {
			inv.HandleMouseMove(e.GlobalX, e.GlobalY)
		}
	})
	rootDiv.On(event.MouseUp, func(e *event.Event) {
		wm.HandleMouseUp()
		inv.HandleMouseUp(e.GlobalX, e.GlobalY)
	})

	// ── Skill Tree Window (middle-left) ──────────────────────────────

	skillTree := game.NewSkillTree(tree, cfg)
	skillTree.SetPoints(3)
	skillTree.SetNodeSize(44)
	skillTree.AddNode(&game.SkillNode{ID: "fireball", Name: "火球", X: 100, Y: 20, State: game.SkillUnlocked, Level: 3, MaxLevel: 5, Cost: 1})
	skillTree.AddNode(&game.SkillNode{ID: "firewall", Name: "火墙", X: 50, Y: 80, State: game.SkillAvailable, Level: 0, MaxLevel: 3, Cost: 1, Requires: []string{"fireball"}})
	skillTree.AddNode(&game.SkillNode{ID: "meteor", Name: "陨石", X: 150, Y: 80, State: game.SkillLocked, Level: 0, MaxLevel: 1, Cost: 3, Requires: []string{"fireball"}})
	skillTree.AddNode(&game.SkillNode{ID: "icebolt", Name: "冰箭", X: 250, Y: 20, State: game.SkillUnlocked, Level: 2, MaxLevel: 5, Cost: 1})
	skillTree.AddNode(&game.SkillNode{ID: "blizzard", Name: "暴风雪", X: 250, Y: 80, State: game.SkillAvailable, Level: 0, MaxLevel: 3, Cost: 2, Requires: []string{"icebolt"}})

	skillWin := game.NewWindow(tree, "天赋树", cfg)
	skillTree.SetStyle(layout.Style{Width: layout.Px(320), Height: layout.Px(200)})
	skillWin.AppendChild(skillTree)
	wm.Add(skillWin, 20, 346, 320, 0) // auto-height

	// ── Scoreboard Window (center) ───────────────────────────────────

	scoreboard := game.NewScoreboard(tree, cfg)
	scoreboard.SetTitle("")
	scoreboard.SetEmbedded(true)
	scoreboard.SetWidth(360)
	scoreboard.AddEntry(game.ScoreEntry{Name: "龙骑士·苍", Score: 28500, Kills: 15, Deaths: 2, Team: 1})
	scoreboard.AddEntry(game.ScoreEntry{Name: "暗影猎手", Score: 22300, Kills: 12, Deaths: 5, Team: 1})
	scoreboard.AddEntry(game.ScoreEntry{Name: "月影法师", Score: 19800, Kills: 8, Deaths: 3, Team: 1})
	scoreboard.AddEntry(game.ScoreEntry{Name: "神圣牧师", Score: 31200, Kills: 2, Deaths: 1, Team: 1})
	scoreboard.AddEntry(game.ScoreEntry{Name: "红莲武士", Score: 21000, Kills: 11, Deaths: 6, Team: 2})
	scoreboard.SortByScore()
	scoreboard.SetVisible(true)

	scoreWin := game.NewWindow(tree, "战场统计", cfg)
	scoreboard.SetStyle(layout.Style{Width: layout.Px(360), Height: layout.Px(220)})
	scoreWin.AppendChild(scoreboard)
	wm.Add(scoreWin, 270, 216, 360, 0) // auto-height

	// ── Loot Window (center) ─────────────────────────────────────────

	loot := game.NewLootWindow(tree, cfg)
	loot.SetTitle("")
	loot.SetEmbedded(true)
	loot.SetWidth(200)
	loot.AddItem(game.LootItem{Item: &game.ItemData{Name: "暗影之刃", Rarity: game.RarityEpic, Quantity: 1}})
	loot.AddItem(game.LootItem{Item: &game.ItemData{Name: "金币", Rarity: game.RarityCommon, Quantity: 1}, Quantity: 250})
	loot.AddItem(game.LootItem{Item: &game.ItemData{Name: "暗影精华", Rarity: game.RarityRare, Quantity: 1}, Quantity: 3})
	loot.Open()

	lootWin := game.NewWindow(tree, "暗影宝箱", cfg)
	loot.SetStyle(layout.Style{Width: layout.Px(200), Height: layout.Px(180)})
	lootWin.AppendChild(loot)
	wm.Add(lootWin, 640, 256, 200, 0) // auto-height

	// ── Dialogue Window (bottom-center) ──────────────────────────────

	dialogue := game.NewDialogueBox(tree, cfg)
	dialogue.SetSize(480, 180)
	dialogue.SetEmbedded(true)
	dialogue.SetChoices([]game.DialogueChoice{
		{Text: "接受任务", OnClick: func() { fmt.Println("[Game] 接受任务") }},
		{Text: "告诉我更多", OnClick: func() { fmt.Println("[Game] 更多信息") }},
		{Text: "还没准备好", OnClick: func() { fmt.Println("[Game] 拒绝") }},
	})
	dialogue.Show("旅行商人·艾瑞克", "旅人，你来得正好。暗影塔的封印正在减弱，你愿意接受这个任务吗？")

	dialogueWin := game.NewWindow(tree, "对话", cfg)
	dialogue.SetStyle(layout.Style{Width: layout.Px(480), Height: layout.Px(180)})
	dialogueWin.AppendChild(dialogue)
	wm.Add(dialogueWin, 400, 352, 480, 0) // auto-height

	// ── Layout + animation ───────────────────────────────────────────

	// cursorPoser is satisfied by win32.Window (and any platform that exposes
	// raw cursor position). Used for smooth drag without WM_MOUSEMOVE coalescing lag.
	type cursorPoser interface{ CursorClientPos() (float32, float32) }

	layoutCache := ui.NewCSSLayoutCache()
	frameN := 0
	app.SetOnLayout(func(tree *core.Tree, root widget.Widget, w, h float32) {
		// During drag: sample the raw hardware cursor position every frame so the
		// window follows the pointer even when WM_MOUSEMOVE hasn't arrived yet.
		// This eliminates the 1–2 frame lag that WM_MOUSEMOVE coalescing causes.
		if wm.IsDragging() {
			if cp, ok := app.Window().(cursorPoser); ok {
				cx, cy := cp.CursorClientPos()
				wm.HandleMouseMove(cx, cy)
			}
		}

		layoutCache.Layout(tree, root, w, h, cfg)

		// Reapply window positions after CSSLayout.
		wm.PostLayout()

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
