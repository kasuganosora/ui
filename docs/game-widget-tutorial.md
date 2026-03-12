# GoUI 游戏控件与 CSS 布局教程

本教程基于 `cmd/game/main.go` 演示项目，介绍如何使用 GoUI 的游戏控件库 (`widget/game`) 构建一个完整的 RPG 风格 HUD 界面。

## 目录

1. [游戏UI框架概述](#1-游戏ui框架概述)
2. [CSS布局系统](#2-css布局系统)
3. [定位辅助函数](#3-定位辅助函数)
4. [游戏控件详解](#4-游戏控件详解)
5. [拖拽系统](#5-拖拽系统)
6. [布局缓存优化](#6-布局缓存优化)
7. [动画更新](#7-动画更新)

---

## 1. 游戏UI框架概述

GoUI 的 `widget/game` 包提供了一整套游戏 HUD 控件，专为 RPG、MMORPG 等游戏界面设计。所有控件继承自 `widget.Base`，支持：

- **CSS 布局定位** —— 通过 `SetStyle()` 参与 `position:absolute` 布局
- **自定义绘制** —— 每个控件实现 `Draw(buf *render.CommandBuffer)` 进行 GPU 渲染
- **零 CGO** —— 与框架一致，Windows 上完全通过 syscall 调用

### 基本结构

```go
app, _ := ui.NewApp(ui.AppOptions{
    Title:   "GoUI — Game UI Demo",
    Width:   1280,
    Height:  800,
    Backend: ui.BackendAuto,
})
defer app.Destroy()

tree := app.Tree()
cfg := app.Config()

// 暗色 RPG 主题
cfg.BgColor = uimath.ColorHex("#0a0e17")
cfg.TextColor = uimath.ColorHex("#c8ccd0")
```

### 包导入

```go
import (
    ui "github.com/kasuganosora/ui"
    "github.com/kasuganosora/ui/core"
    "github.com/kasuganosora/ui/event"
    "github.com/kasuganosora/ui/layout"
    uimath "github.com/kasuganosora/ui/math"
    "github.com/kasuganosora/ui/widget"
    "github.com/kasuganosora/ui/widget/game"
)
```

---

## 2. CSS布局系统

游戏 HUD 采用 **HTML + CSS 绝对定位** 模式。根容器是一个 `position:relative` 的全屏 `<div>`，所有 HUD 元素通过 `position:absolute` 精确放置在屏幕各个位置。

### 创建根布局

```go
doc := app.LoadHTML(`<div style="position:relative; width:100%; height:100%; background:#0a0e17;">
  <div id="chat-messages" style="position:absolute; bottom:110px; left:10px; width:340px; height:172px; overflow:auto;"></div>
  <input id="chat-input" style="position:absolute; bottom:82px; left:10px; width:340px; height:28px;" placeholder="输入消息..." />
</div>`)
rootDiv := doc.Root.Children()[0].(*widget.Div)
```

**要点：**
- 根 `<div>` 设置 `position:relative`，作为所有绝对定位子元素的包含块
- `width:100%; height:100%` 使根容器填满整个视口
- HTML 中可以直接放置框架内建控件（如 `<div>` 滚动容器、`<input>` 输入框）
- 游戏控件通过 `rootDiv.AppendChild()` 动态添加

### 样式系统

每个游戏控件通过 `SetStyle()` 设置 CSS 样式，参与 `CSSLayout` 布局计算：

```go
hpBar.SetStyle(layout.Style{
    Position: layout.PositionAbsolute,
    Left:     layout.Px(20),
    Top:      layout.Px(20),
    Width:    layout.Px(220),
    Height:   layout.Px(22),
})
rootDiv.AppendChild(hpBar)
```

支持的定位属性：
- `Position`: `PositionAbsolute` / `PositionRelative`
- `Left` / `Right` / `Top` / `Bottom`: 支持 `Px()` 像素值和 `Pct()` 百分比
- `Width` / `Height`: 元素尺寸
- `Margin`: 用于配合百分比定位实现居中效果

---

## 3. 定位辅助函数

demo 中定义了 8 个定位辅助函数，覆盖屏幕九宫格的所有关键位置。这些函数返回 `layout.Style`，简化了绝对定位的样式构建。

### absTopLeft — 左上角定位

```go
absTopLeft := func(left, top, w, h float32) layout.Style {
    return layout.Style{
        Position: layout.PositionAbsolute,
        Left:     layout.Px(left),
        Top:      layout.Px(top),
        Width:    layout.Px(w),
        Height:   layout.Px(h),
    }
}
// 用法：血条放在左上角 (20, 20)
hpBar.SetStyle(absTopLeft(20, 20, 220, 22))
```

### absTopRight — 右上角定位

```go
absTopRight := func(right, top, w, h float32) layout.Style {
    return layout.Style{
        Position: layout.PositionAbsolute,
        Right:    layout.Px(right),
        Top:      layout.Px(top),
        Width:    layout.Px(w),
        Height:   layout.Px(h),
    }
}
// 用法：小地图放在右上角
minimap.SetStyle(absTopRight(20, 20, 160, 160))
```

### absTopCenter — 顶部居中

通过 `left:50%` + 负 `margin-left` 实现水平居中：

```go
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
// 用法：货币显示在顶部居中
currency.SetStyle(absTopCenter(8, 300, 24, 0))
// 目标框架偏左 140px
target.SetStyle(absTopCenter(68, 220, 52, -140))
```

### absBottomCenter — 底部居中

```go
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
// 用法：快捷栏在底部居中
hotbar.SetStyle(absBottomCenter(20, hotbarW, 52, 0))
```

### absBottomLeft — 左下角定位

```go
absBottomLeft := func(left, bottom, w, h float32) layout.Style {
    return layout.Style{
        Position: layout.PositionAbsolute,
        Left:     layout.Px(left),
        Bottom:   layout.Px(bottom),
        Width:    layout.Px(w),
        Height:   layout.Px(h),
    }
}
// 用法：聊天框在左下角
chat.SetStyle(absBottomLeft(10, 82, 340, 200))
```

### absMiddleRight — 右侧垂直居中

通过 `top:50%` + 负 `margin-top` 实现垂直居中：

```go
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
// 用法：背包在右侧居中
inv.SetStyle(absMiddleRight(20, invW, invH, 0))
```

### absMiddleLeft — 左侧垂直居中

```go
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
// 用法：技能树在左侧居中偏下
skillTree.SetStyle(absMiddleLeft(20, 320, 200, 60))
```

### absMiddleCenter — 屏幕正中

双轴居中，支持额外偏移量：

```go
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
// 用法：计分板偏左上方
scoreboard.SetStyle(absMiddleCenter(360, 220, -190, -60))
// 拾取窗口偏右
loot.SetStyle(absMiddleCenter(200, 180, 100, -40))
```

### 九宫格定位速查

| 位置 | 函数 | 参数 |
|------|------|------|
| 左上 | `absTopLeft` | `(left, top, w, h)` |
| 上中 | `absTopCenter` | `(top, w, h, offsetX)` |
| 右上 | `absTopRight` | `(right, top, w, h)` |
| 左中 | `absMiddleLeft` | `(left, w, h, offsetY)` |
| 正中 | `absMiddleCenter` | `(w, h, offsetX, offsetY)` |
| 右中 | `absMiddleRight` | `(right, w, h, offsetY)` |
| 左下 | `absBottomLeft` | `(left, bottom, w, h)` |
| 下中 | `absBottomCenter` | `(bottom, w, h, offsetX)` |
| 右下 | — | 组合 `Right` + `Bottom` |

---

## 4. 游戏控件详解

### 4.1 HealthBar（血条/资源条）

通用的水平资源条，可用于 HP、MP、XP 等任何百分比条。

**API：**
- `NewHealthBar(tree, cfg)` —— 创建资源条
- `SetCurrent(v)` / `SetMax(v)` —— 设置当前值和最大值
- `SetBarColor(c)` —— 设置填充颜色
- `SetBgColor(c)` —— 设置背景颜色
- `SetShowText(v)` —— 是否显示 "当前/最大" 文字
- `SetSize(w, h)` —— 设置条的尺寸

```go
// HP 条 — 绿色
hpBar := game.NewHealthBar(tree, cfg)
hpBar.SetCurrent(780)
hpBar.SetMax(1200)
hpBar.SetBarColor(uimath.ColorHex("#52c41a"))
hpBar.SetShowText(true)
hpBar.SetSize(220, 22)
hpBar.SetStyle(absTopLeft(20, 20, 220, 22))
rootDiv.AppendChild(hpBar)

// MP 条 — 蓝色
mpBar := game.NewHealthBar(tree, cfg)
mpBar.SetCurrent(350)
mpBar.SetMax(600)
mpBar.SetBarColor(uimath.ColorHex("#1890ff"))
mpBar.SetShowText(true)
mpBar.SetSize(220, 18)
mpBar.SetStyle(absTopLeft(20, 48, 220, 18))
rootDiv.AppendChild(mpBar)

// XP 条 — 紫色，更细
xpBar := game.NewHealthBar(tree, cfg)
xpBar.SetCurrent(4200)
xpBar.SetMax(8500)
xpBar.SetBarColor(uimath.ColorHex("#a335ee"))
xpBar.SetShowText(true)
xpBar.SetSize(220, 12)
xpBar.SetStyle(absTopLeft(20, 72, 220, 12))
rootDiv.AppendChild(xpBar)
```

### 4.2 BuffBar（Buff/Debuff 栏）

显示一行 buff/debuff 图标，支持正面/负面区分、持续时间和层数显示。

**API：**
- `NewBuffBar(tree, cfg)` —— 创建 buff 栏
- `SetIconSize(s)` —— 图标尺寸
- `SetGap(g)` —— 图标间距
- `SetMaxIcons(m)` —— 最大显示数量
- `AddBuff(b)` —— 添加 buff
- `RemoveBuff(id)` —— 按 ID 移除
- `ClearBuffs()` —— 清空所有

**Buff 数据结构：**
```go
type Buff struct {
    ID       string
    Icon     render.TextureHandle  // 图标纹理（0 = 使用占位）
    Label    string                // 短文字标签
    Duration float32               // 剩余秒数，0 = 永久
    Stacks   int                   // 层数
    Type     BuffType              // BuffPositive 或 BuffNegative
}
```

```go
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
```

正面 buff 显示蓝色边框，负面 debuff 显示红色边框。当 `Duration < 30` 时，图标上方会出现倒计时遮罩。

### 4.3 Hotbar（快捷栏）

一行可操作的技能/物品快捷槽，支持冷却遮罩、按键绑定显示、选中高亮。

**API：**
- `NewHotbar(tree, numSlots, cfg)` —— 创建快捷栏
- `SetSlotSize(s)` —— 槽位尺寸
- `SetGap(g)` —— 槽位间距
- `SetSelected(i)` —— 设置选中槽位
- `SetSlot(i, slot)` —— 设置槽位数据

**HotbarSlot 数据结构：**
```go
type HotbarSlot struct {
    Icon      render.TextureHandle  // 技能图标
    Label     string
    Cooldown  float32               // 0-1，冷却剩余比例
    Keybind   string                // 按键绑定文字，如 "1"、"Q"
    Available bool                  // 是否可用
}
```

```go
hotbar := game.NewHotbar(tree, 10, cfg)
hotbar.SetSlotSize(52)
hotbar.SetGap(4)
hotbar.SetSelected(0)
for i := 0; i < 10; i++ {
    hotbar.SetSlot(i, game.HotbarSlot{
        Keybind:   fmt.Sprintf("%d", (i+1)%10),
        Available: true,
    })
}
// 第3个槽位：65% 冷却中
hotbar.SetSlot(2, game.HotbarSlot{Keybind: "3", Cooldown: 0.65, Available: true})
// 第6个槽位：30% 冷却中，不可用
hotbar.SetSlot(5, game.HotbarSlot{Keybind: "6", Cooldown: 0.3, Available: false})

hotbarW := float32(10 * (52 + 4))
hotbar.SetStyle(absBottomCenter(20, hotbarW, 52, 0))
rootDiv.AppendChild(hotbar)
```

### 4.4 CastBar（施法条）

显示技能施法进度，支持开始施法、中断、完成回调。

**API：**
- `NewCastBar(tree, cfg)` —— 创建施法条
- `SetSize(w, h)` —— 尺寸
- `SetColor(c)` —— 进度条颜色
- `StartCast(spellName, castTime)` —— 开始施法
- `Tick(dt)` —— 每帧更新（传入帧间隔秒数）
- `Interrupt()` —— 中断施法
- `OnComplete(fn)` —— 施法完成回调
- `OnInterrupt(fn)` —— 中断回调

```go
castBar := game.NewCastBar(tree, cfg)
castBar.SetSize(280, 22)
castBar.SetColor(uimath.ColorHex("#ffd700"))
castBar.StartCast("火球术", 3.0)
castBar.Tick(1.5)  // 模拟已过 1.5 秒，进度 50%
castBar.SetStyle(absBottomCenter(82, 280, 22, 0))
rootDiv.AppendChild(castBar)

// 每帧在 OnLayout 回调中推进
castBar.Tick(0.008)
```

### 4.5 Minimap（小地图）

圆形或方形小地图，显示玩家位置和标记点。

**API：**
- `NewMinimap(tree, cfg)` —— 创建小地图
- `SetSize(s)` —— 尺寸（正方形）
- `SetCircular(c)` —— 圆形模式
- `SetPlayerPos(x, y)` —— 玩家在地图上的位置
- `SetPlayerRotation(r)` —— 玩家朝向（弧度）
- `SetZoom(z)` —— 缩放倍率
- `AddMarker(marker)` —— 添加地图标记
- `ClearMarkers()` —— 清空标记

**MinimapMarker 数据结构：**
```go
type MinimapMarker struct {
    X, Y  float32      // 地图坐标
    Color uimath.Color // 标记颜色
    Size  float32      // 标记大小
    Label string       // 可选标签
}
```

```go
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
```

标记点坐标相对于 `PlayerPos` 计算偏移，超出地图范围的标记会被裁剪。

### 4.6 QuestTracker（任务追踪器）

显示当前活跃任务及其目标进度。

**API：**
- `NewQuestTracker(tree, cfg)` —— 创建任务追踪器
- `SetWidth(w)` —— 宽度
- `SetMaxQuests(m)` —— 最大显示任务数
- `AddQuest(q)` —— 添加任务
- `RemoveQuest(index)` —— 移除任务
- `ClearQuests()` —— 清空

**数据结构：**
```go
type Quest struct {
    Title      string
    Objectives []QuestObjective
    Active     bool
}

type QuestObjective struct {
    Text      string
    Current   int
    Required  int
    Completed bool
}
```

```go
quest := game.NewQuestTracker(tree, cfg)
quest.SetWidth(230)
quest.AddQuest(game.Quest{
    Title:  "讨伐暗影领主",
    Active: true,
    Objectives: []game.QuestObjective{
        {Text: "击败暗影守卫", Current: 3, Required: 5},
        {Text: "收集暗影碎片", Current: 7, Required: 10},
        {Text: "到达暗影塔顶层"},
    },
})
quest.AddQuest(game.Quest{
    Title:  "商人的委托",
    Active: true,
    Objectives: []game.QuestObjective{
        {Text: "收集铁矿石", Current: 12, Required: 12, Completed: true},
        {Text: "交给铁匠铺"},
    },
})
quest.SetStyle(absTopRight(20, 200, 230, 200))
rootDiv.AppendChild(quest)
```

已完成的目标会显示灰色带 `✓` 前缀，有数量要求的目标显示 `(当前/要求)` 计数。

### 4.7 CurrencyDisplay（货币显示）

水平排列显示多种货币值。

**API：**
- `NewCurrencyDisplay(tree, cfg)` —— 创建货币显示
- `SetGap(g)` —— 货币项间距
- `AddCurrency(c)` —— 添加货币类型
- `SetAmount(index, amount)` —— 更新指定货币数量
- `ClearCurrencies()` —— 清空

**CurrencyEntry 数据结构：**
```go
type CurrencyEntry struct {
    Icon   render.TextureHandle  // 货币图标
    Symbol string                // 文字符号，如 "G"、"◆"
    Amount int
    Color  uimath.Color
}
```

```go
currency := game.NewCurrencyDisplay(tree, cfg)
currency.SetGap(16)
currency.AddCurrency(game.CurrencyEntry{Symbol: "G", Amount: 12580, Color: uimath.ColorHex("#ffd700")})
currency.AddCurrency(game.CurrencyEntry{Symbol: "◆", Amount: 350, Color: uimath.ColorHex("#44aaff")})
currency.AddCurrency(game.CurrencyEntry{Symbol: "★", Amount: 28, Color: uimath.ColorHex("#ff8800")})
currency.SetStyle(absTopCenter(8, 300, 24, 0))
rootDiv.AppendChild(currency)
```

大数值自动格式化：`12580` 显示为 `12.5K`，`1500000` 显示为 `1.5M`。

### 4.8 CountdownTimer（倒计时器）

显示倒计时，支持标签和到期回调，低于 10 秒时文字变红闪烁。

**API：**
- `NewCountdownTimer(tree, cfg)` —— 创建倒计时
- `SetSeconds(s)` —— 设置剩余秒数
- `SetLabel(l)` —— 设置标签文字
- `SetColor(c)` —— 设置颜色
- `SetFontSize(s)` —— 自定义字号
- `Tick(dt)` —— 每帧更新
- `OnExpire(fn)` —— 到期回调
- `IsExpired()` —— 是否已到期

```go
countdown := game.NewCountdownTimer(tree, cfg)
countdown.SetSeconds(185)  // 3分05秒
countdown.SetLabel("Boss 刷新")
countdown.SetColor(uimath.ColorWhite)
countdown.SetStyle(absTopCenter(36, 200, 30, 0))
rootDiv.AppendChild(countdown)

// 每帧更新
countdown.Tick(0.016)
```

时间格式自动处理：`>= 60s` 显示为 `M:SS`，`< 60s` 只显示秒数。

### 4.9 TargetFrame（目标框架）

显示当前选中目标的名称、等级、HP/MP。

**API：**
- `NewTargetFrame(tree, cfg)` —— 创建目标框架
- `SetSize(w, h)` —— 尺寸
- `SetTarget(data)` —— 设置目标数据
- `ClearTarget()` —— 清除目标

**UnitFrameData 数据结构（与 TeamFrame 共享）：**
```go
type UnitFrameData struct {
    Name     string
    Level    int
    HP       float32
    HPMax    float32
    MP       float32
    MPMax    float32
    Class    string
    Portrait render.TextureHandle
    Dead     bool
}
```

```go
target := game.NewTargetFrame(tree, cfg)
target.SetSize(220, 52)
target.SetTarget(&game.UnitFrameData{
    Name:  "暗影领主·莫德雷克",
    Level: 62,
    HP:    185000, HPMax: 500000,
    MP:    80000,  MPMax: 80000,
    Class: "Boss",
})
target.SetStyle(absTopCenter(68, 220, 52, -140))
rootDiv.AppendChild(target)
```

### 4.10 TeamFrame（队伍框架）

显示队伍成员列表，每个成员有独立的血蓝条。

**API：**
- `NewTeamFrame(tree, cfg)` —— 创建队伍框架
- `SetFrameSize(w, h)` —— 单个成员框的尺寸
- `SetGap(g)` —— 成员之间的间距
- `SetMaxSlots(m)` —— 最大显示人数
- `SetMembers(members)` —— 批量设置成员数据
- `UpdateMember(index, data)` —— 更新单个成员

```go
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
```

死亡成员（`Dead: true`）名字显示为灰色。

### 4.11 ChatBox（聊天框）

聊天框由三部分组成：
1. **ChatBox 控件** —— 绘制背景区域
2. **HTML `<div>`** —— 框架内建的可滚动容器，显示消息
3. **HTML `<input>`** —— 框架内建的输入框

```go
// 1. ChatBox 背景
chat := game.NewChatBox(tree, cfg)
chat.SetSize(340, 200)
chat.SetMaxVisible(8)
chat.SetStyle(absBottomLeft(10, 82, 340, 200))
rootDiv.AppendChild(chat)

// 2. 从 HTML 获取滚动容器和输入框
chatMsgDiv := doc.QueryByID("chat-messages").(*widget.Div)
chatInput := doc.QueryByID("chat-input").(*widget.Input)
chatInput.SetBorderless(true)

// 3. 发送消息的辅助函数
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

// 添加初始消息
addChatMsg("系统", "欢迎来到暗影之境！", uimath.ColorHex("#ffd700"))
addChatMsg("骑士", "队伍已就绪，准备进攻北塔", uimath.ColorHex("#44aaff"))

// 4. 滚轮滚动
chatMsgDiv.On(event.MouseWheel, func(e *event.Event) {
    chatMsgDiv.ScrollTo(0, chatMsgDiv.ScrollY()-e.WheelDY*30)
    // ... 边界限制 ...
})

// 5. 回车发送
chatInput.OnEnter(func(text string) {
    if text != "" {
        addChatMsg("你", text, uimath.ColorHex("#ffffff"))
        chatInput.SetValue("")
    }
})
```

**ChatMessage 数据结构：**
```go
type ChatMessage struct {
    Sender  string
    Text    string
    Color   uimath.Color  // 发送者名字颜色
    Channel string        // 频道，如 "world"、"party"、"whisper"
}
```

### 4.12 Nameplate（铭牌）

浮动在游戏场景实体上方的名称和血条指示器。

**API：**
- `NewNameplate(tree, name, cfg)` —— 创建铭牌
- `SetLevel(l)` —— 等级
- `SetHP(current, max)` —— 血量
- `SetType(t)` —— 类型：`NameplateFriendly` / `NameplateHostile` / `NameplateNeutral` / `NameplatePlayer`
- `SetBarSize(w, h)` —— 血条尺寸
- `SetPosition(x, y)` —— 在场景中的位置
- `SetVisible(v)` —— 可见性

```go
// 敌方 — 红色
np1 := game.NewNameplate(tree, "暗影守卫", cfg)
np1.SetLevel(55)
np1.SetHP(3200, 5000)
np1.SetType(game.NameplateHostile)
np1.SetBarSize(100, 6)
np1.SetPosition(480, 350)
np1.SetVisible(true)
np1.SetStyle(absTopLeft(480, 350, 100, 30))
rootDiv.AppendChild(np1)

// 中立 — 黄色
np2 := game.NewNameplate(tree, "旅行商人", cfg)
np2.SetType(game.NameplateNeutral)

// 友方 — 绿色
np3 := game.NewNameplate(tree, "神圣牧师", cfg)
np3.SetType(game.NameplateFriendly)
```

血条颜色随 HP 比例变化：`>= 60%` 绿色，`30%-60%` 黄色，`< 30%` 红色。

### 4.13 Inventory（背包）

网格式背包，支持物品稀有度颜色、数量显示、拖拽交换。

**API：**
- `NewInventory(tree, rows, cols, cfg)` —— 创建背包（行数 x 列数）
- `SetTitle(t)` —— 标题
- `SetSlotSize(s)` —— 格子尺寸
- `SetGap(g)` —— 格子间距
- `SetItem(slot, item)` —— 放置物品
- `GetItem(slot)` —— 获取物品
- `RemoveItem(slot)` —— 取出物品
- `ClearAll()` —— 清空
- `OnSelect(fn)` —— 点击槽位回调
- `OnDrop(fn)` —— 拖放完成回调
- `HandleMouseDown/Move/Up(x, y)` —— 鼠标事件处理

**ItemData 和 ItemRarity：**
```go
type ItemData struct {
    ID       string
    Name     string
    Icon     render.TextureHandle
    Quantity int
    Rarity   ItemRarity
}

const (
    RarityCommon    ItemRarity = iota  // 灰色
    RarityUncommon                      // 绿色 #1eff00
    RarityRare                          // 蓝色 #0070dd
    RarityEpic                          // 紫色 #a335ee
    RarityLegendary                     // 橙色 #ff8000
)
```

```go
inv := game.NewInventory(tree, 5, 6, cfg)
inv.SetTitle("背包")
inv.SetSlotSize(44)
inv.SetGap(3)

inv.SetItem(0, &game.ItemData{ID: "sword", Name: "铁剑", Quantity: 1, Rarity: game.RarityUncommon})
inv.SetItem(1, &game.ItemData{ID: "potion", Name: "治疗药水", Quantity: 5, Rarity: game.RarityCommon})
inv.SetItem(5, &game.ItemData{ID: "shield", Name: "龙鳞盾牌", Quantity: 1, Rarity: game.RarityEpic})
inv.SetItem(8, &game.ItemData{ID: "scroll", Name: "远古卷轴", Quantity: 1, Rarity: game.RarityLegendary})

invW := float32(6*(44+3)) + 20
invH := float32(5*(44+3)) + 40
inv.SetStyle(absMiddleRight(20, invW, invH, 0))
rootDiv.AppendChild(inv)
```

背包内置物品拖拽功能：按住物品拖动到另一个格子会交换两个格子的物品。

### 4.14 SkillTree（技能树）

显示技能节点网络，节点之间有前置依赖连线。

**API：**
- `NewSkillTree(tree, cfg)` —— 创建技能树
- `SetPoints(p)` —— 设置可用技能点
- `SetNodeSize(s)` —— 节点尺寸
- `AddNode(node)` —— 添加节点
- `FindNode(id)` —— 查找节点
- `UnlockNode(id)` —— 解锁节点（自动检查前置和点数）
- `OnUnlock(fn)` —— 解锁回调
- `OnSelect(fn)` —— 选择回调
- `SetScroll(x, y)` —— 滚动偏移

**SkillNode 数据结构：**
```go
type SkillNode struct {
    ID          string
    Name        string
    Description string
    Icon        render.TextureHandle
    X, Y        float32           // 树空间中的位置
    State       SkillNodeState    // Locked/Available/Unlocked/Maxed
    Level       int
    MaxLevel    int
    Cost        int               // 所需技能点
    Requires    []string          // 前置技能 ID 列表
}
```

```go
skillTree := game.NewSkillTree(tree, cfg)
skillTree.SetPoints(3)
skillTree.SetNodeSize(44)

skillTree.AddNode(&game.SkillNode{
    ID: "fireball", Name: "火球", X: 100, Y: 20,
    State: game.SkillUnlocked, Level: 3, MaxLevel: 5, Cost: 1,
})
skillTree.AddNode(&game.SkillNode{
    ID: "firewall", Name: "火墙", X: 50, Y: 80,
    State: game.SkillAvailable, Level: 0, MaxLevel: 3, Cost: 1,
    Requires: []string{"fireball"},
})
skillTree.AddNode(&game.SkillNode{
    ID: "meteor", Name: "陨石", X: 150, Y: 80,
    State: game.SkillLocked, Level: 0, MaxLevel: 1, Cost: 3,
    Requires: []string{"fireball"},
})

skillTree.SetStyle(absMiddleLeft(20, 320, 200, 60))
rootDiv.AppendChild(skillTree)
```

节点颜色按状态区分：
- `SkillLocked` — 灰色
- `SkillAvailable` — 蓝色
- `SkillUnlocked` — 绿色
- `SkillMaxed` — 金色

前置依赖线在未解锁时为灰色，已解锁时为浅绿色。

### 4.15 Scoreboard（计分板）

表格式战场/PvP 积分榜，显示名称、分数、击杀、死亡。

**API：**
- `NewScoreboard(tree, cfg)` —— 创建计分板
- `SetTitle(t)` —— 标题
- `SetWidth(w)` —— 宽度
- `SetVisible(v)` —— 可见性（通常按 Tab 键切换）
- `AddEntry(e)` —— 添加条目
- `SortByScore()` —— 按分数降序排序
- `ClearEntries()` —— 清空

```go
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
```

### 4.16 LootWindow（拾取窗口）

显示可拾取物品列表，支持单个拾取和全部拾取。

**API：**
- `NewLootWindow(tree, cfg)` —— 创建拾取窗口
- `SetTitle(t)` —— 标题
- `SetWidth(w)` —— 宽度
- `AddItem(item)` —— 添加拾取项
- `Open()` / `Close()` —— 打开/关闭
- `LootItem(index)` —— 拾取单个物品
- `LootAll()` —— 拾取全部
- `OnLoot(fn)` —— 拾取回调
- `OnClose(fn)` —— 关闭回调

```go
loot := game.NewLootWindow(tree, cfg)
loot.SetTitle("暗影宝箱")
loot.SetWidth(200)
loot.AddItem(game.LootItem{
    Item: &game.ItemData{Name: "暗影之刃", Rarity: game.RarityEpic, Quantity: 1},
})
loot.AddItem(game.LootItem{
    Item:     &game.ItemData{Name: "金币", Rarity: game.RarityCommon, Quantity: 1},
    Quantity: 250,
})
loot.AddItem(game.LootItem{
    Item:     &game.ItemData{Name: "暗影精华", Rarity: game.RarityRare, Quantity: 1},
    Quantity: 3,
})
loot.Open()
loot.SetStyle(absMiddleCenter(200, 180, 100, -40))
rootDiv.AppendChild(loot)
```

### 4.17 DialogueBox（对话框）

NPC 对话窗口，支持说话者名称、对话文本、玩家选项。

**API：**
- `NewDialogueBox(tree, cfg)` —— 创建对话框
- `SetSize(w, h)` —— 尺寸
- `Show(speaker, text)` —— 显示对话
- `Hide()` —— 隐藏
- `SetChoices(choices)` —— 设置玩家选项
- `ClearChoices()` —— 清空选项
- `OnAdvance(fn)` —— 无选项时点击前进的回调

```go
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
```

选项文字自动添加序号前缀（`1. 接受任务`），无选项时显示 `▼` 提示点击继续。

### 4.18 FloatingText（浮动文字）

用于伤害数字、经验获取等临时浮动文本。

**API：**
- `NewFloatingText(tree, text, x, y, color, cfg)` —— 创建浮动文字
- `SetPosition(x, y)` —— 更新位置
- `SetText(t)` —— 更新文字
- `SetColor(c)` —— 更新颜色

```go
dmgText := game.NewFloatingText(tree, "-1234", 400, 300, uimath.ColorHex("#ff4444"), cfg)
rootDiv.AppendChild(dmgText)

// 在动画循环中让文字上浮
dmgText.SetPosition(400, 300 - t*30)  // t 为帧时间
```

### 4.19 ItemTooltip（物品提示）

鼠标悬停时显示物品详细信息的浮窗。

**API：**
- `NewItemTooltip(tree, cfg)` —— 创建物品提示
- `SetItem(item)` —— 设置显示的物品
- `SetPosition(x, y)` —— 设置显示位置（通常跟随鼠标）
- `SetVisible(v)` —— 显示/隐藏

```go
tooltip := game.NewItemTooltip(tree, cfg)
tooltip.SetItem(&game.ItemData{Name: "暗影之刃", Rarity: game.RarityEpic})
tooltip.SetPosition(mouseX + 10, mouseY + 10)
tooltip.SetVisible(true)
rootDiv.AppendChild(tooltip)
```

提示框边框颜色跟随物品稀有度。

### 4.20 NotificationToast（通知提示）

短暂显示的系统通知，支持信息/成功/警告/错误四种类型。

**API：**
- `NewNotificationToast(tree, text, cfg)` —— 创建通知
- `SetText(t)` —— 更新文字
- `SetToastType(t)` —— 类型：`ToastInfo` / `ToastSuccess` / `ToastWarning` / `ToastError`
- `SetPosition(x, y)` —— 位置
- `SetVisible(v)` —— 显示/隐藏

```go
toast := game.NewNotificationToast(tree, "任务已完成！", cfg)
toast.SetToastType(game.ToastSuccess)
toast.SetPosition(500, 60)
toast.SetVisible(true)
rootDiv.AppendChild(toast)
```

每种类型在左侧显示不同颜色的指示条：蓝色(Info)、绿色(Success)、黄色(Warning)、红色(Error)。

---

## 5. 拖拽系统

demo 实现了两级拖拽：**面板拖拽**（移动整个窗口）和**物品拖拽**（背包内物品交换）。

### 5.1 面板拖拽

通过在根 Div 上监听鼠标事件，结合 `BringChildToFront` 实现窗口管理。

```go
// 可拖拽面板列表
type draggablePanel struct {
    w      widget.Widget
    titleH float32  // 标题栏高度。0 = 整个控件可拖拽
}
panels := []draggablePanel{
    {inv, 28},        // 背包 — 只能拖标题栏
    {scoreboard, 36}, // 计分板 — 只能拖标题栏
    {loot, 32},       // 拾取窗口 — 只能拖标题栏
    {dialogue, 30},   // 对话框 — 只能拖标题栏
    {skillTree, 0},   // 技能树 — 整个区域可拖
    {hpBar, 0},       // 血条 — 整个区域可拖
    // ... 更多面板 ...
}

var drag dragState
offsets := make(map[core.ElementID]*dragOffset)
```

**MouseDown — 判定命中并置顶：**

```go
rootDiv.On(event.MouseDown, func(e *event.Event) {
    // 物品拖拽优先
    if inv.HandleMouseDown(e.GlobalX, e.GlobalY) {
        return
    }
    // 逆序检查（最上层优先）
    for i := len(panels) - 1; i >= 0; i-- {
        p := panels[i]
        b := boundsOf(p.w)
        if /* 鼠标在面板内 */ {
            if p.titleH > 0 && e.GlobalY > b.Y+p.titleH {
                continue  // 点击在内容区，不是标题栏
            }
            // 置顶！
            rootDiv.BringChildToFront(p.w)
            // 记录拖拽起始状态 ...
        }
    }
})
```

**关键：`BringChildToFront`**

调用 `rootDiv.BringChildToFront(p.w)` 将点击的面板移到子元素末尾，使其渲染在最上层。同时需要同步更新 `panels` 切片中的顺序，保持命中测试与渲染顺序一致。

**MouseMove — 实时移动：**

```go
rootDiv.On(event.MouseMove, func(e *event.Event) {
    if !drag.active { return }
    dx := e.GlobalX - drag.startX
    dy := e.GlobalY - drag.startY
    newX := drag.origX + dx
    newY := drag.origY + dy
    b := boundsOf(drag.target)
    tree.SetLayout(drag.target.ElementID(), core.LayoutResult{
        Bounds: uimath.NewRect(newX, newY, b.Width, b.Height),
    })
})
```

使用 `tree.SetLayout()` 直接更新元素位置，绕过 CSS 布局引擎，实现流畅的拖拽响应。

**MouseUp — 持久化偏移：**

```go
rootDiv.On(event.MouseUp, func(e *event.Event) {
    if drag.active {
        eid := drag.target.ElementID()
        if offsets[eid] == nil {
            offsets[eid] = &dragOffset{}
        }
        // 累积拖拽偏移
        offsets[eid].dx += e.GlobalX - drag.startX
        offsets[eid].dy += e.GlobalY - drag.startY
        drag.active = false
    }
})
```

松开鼠标时将偏移存入 `offsets` map，在每帧布局后重新应用（见第 6 节）。

### 5.2 物品拖拽

背包的物品拖拽是内建功能，通过 `HandleMouseDown/Move/Up` 方法实现：

```go
rootDiv.On(event.MouseDown, func(e *event.Event) {
    // 物品拖拽优先于面板拖拽
    if inv.HandleMouseDown(e.GlobalX, e.GlobalY) {
        return
    }
    // ... 面板拖拽 ...
})

rootDiv.On(event.MouseMove, func(e *event.Event) {
    if inv.HandleMouseMove(e.GlobalX, e.GlobalY) {
        return
    }
    // ... 面板拖拽 ...
})

rootDiv.On(event.MouseUp, func(e *event.Event) {
    inv.HandleMouseUp(e.GlobalX, e.GlobalY)
    // ... 面板拖拽 ...
})
```

拖拽物品时：
- 原格子变空
- 物品图标跟随鼠标显示（半透明）
- 目标格子高亮（金色边框）
- 松开时两个格子的物品交换

---

## 6. 布局缓存优化

每帧重新构建布局树并计算 CSS 布局的开销较大（约 10ms）。`CSSLayoutCache` 通过缓存避免不必要的重算：

```go
layoutCache := ui.NewCSSLayoutCache()

app.SetOnLayout(func(tree *core.Tree, root widget.Widget, w, h float32) {
    // 缓存布局 — 仅在树结构或视口变化时重新计算
    layoutCache.Layout(tree, root, w, h, cfg)

    // 布局后应用持久化拖拽偏移
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
})
```

**工作流程：**
1. `layoutCache.Layout()` 检查是否有 `DirtyLayout` 标记或视口尺寸变化
2. 如果没有变化，跳过整个布局计算（< 1ms）
3. 如果有变化，执行完整的 CSS 布局计算，并缓存结果
4. 布局完成后，从 `offsets` map 读取各面板的拖拽偏移，覆盖到计算结果上

**为什么需要在布局后应用偏移？**

因为 CSS 布局每次重算都会把元素恢复到样式定义的位置。拖拽偏移必须在布局完成后叠加，否则会被覆盖。

可以通过 `layoutCache.Invalidate()` 强制下一帧重新计算布局。

---

## 7. 动画更新

动画在 `SetOnLayout` 回调中更新，每帧执行。注意区分**布局变化**和**绘制变化**：

```go
frameN := 0

app.SetOnLayout(func(tree *core.Tree, root widget.Widget, w, h float32) {
    layoutCache.Layout(tree, root, w, h, cfg)
    // ... 拖拽偏移 ...

    frameN++
    t := float32(frameN) * 0.016

    // 推进施法条进度
    castBar.Tick(0.008)

    // HP/MP 呼吸动画
    hpBar.SetCurrent(float32(780 + 50*math.Sin(float64(t*0.8))))
    mpBar.SetCurrent(float32(350 + 80*math.Sin(float64(t*0.5))))

    // Boss HP 缓慢下降
    bossHP := float32(185000) - float32(frameN)*10
    if bossHP < 50000 {
        bossHP = 185000
        frameN = 0
    }
    target.SetTarget(&game.UnitFrameData{
        Name: "暗影领主·莫德雷克", Level: 62,
        HP: bossHP, HPMax: 500000, MP: 80000, MPMax: 80000, Class: "Boss",
    })

    // 倒计时递减
    countdown.Tick(0.016)

    // 标记绘制脏 — 不触发布局重算！
    tree.MarkDirty(tree.Root())
})
```

**关键点：**

- `tree.MarkDirty()` 只标记**绘制脏**（需要重绘），不会触发 CSS 布局重算
- 动画只修改控件的数据属性（如 `SetCurrent`、`Tick`），不改变布局树结构
- 因此 `layoutCache` 每帧只做位置偏移的应用，不做完整布局计算
- 这确保了即使有大量动画，帧率依然保持流畅

### 常见动画模式

| 模式 | 方法 | 示例 |
|------|------|------|
| 进度推进 | `Tick(dt)` | 施法条、倒计时 |
| 值变化 | `SetCurrent(v)` | HP/MP 动画 |
| 数据更新 | `SetTarget(data)` | 目标框架 |
| 位置移动 | `SetPosition(x, y)` | 浮动文字上升 |
| 可见性切换 | `SetVisible(v)` | 通知弹出/消失 |

---

## 附录：控件速查表

| 控件 | 包路径 | 创建函数 | 典型位置 |
|------|--------|----------|----------|
| HealthBar | `widget/game` | `NewHealthBar(tree, cfg)` | 左上角 |
| BuffBar | `widget/game` | `NewBuffBar(tree, cfg)` | 血条下方 |
| Hotbar | `widget/game` | `NewHotbar(tree, n, cfg)` | 底部居中 |
| CastBar | `widget/game` | `NewCastBar(tree, cfg)` | 快捷栏上方 |
| Minimap | `widget/game` | `NewMinimap(tree, cfg)` | 右上角 |
| QuestTracker | `widget/game` | `NewQuestTracker(tree, cfg)` | 小地图下方 |
| CurrencyDisplay | `widget/game` | `NewCurrencyDisplay(tree, cfg)` | 顶部居中 |
| CountdownTimer | `widget/game` | `NewCountdownTimer(tree, cfg)` | 顶部居中 |
| TargetFrame | `widget/game` | `NewTargetFrame(tree, cfg)` | 顶部偏左 |
| TeamFrame | `widget/game` | `NewTeamFrame(tree, cfg)` | 左侧 |
| ChatBox | `widget/game` | `NewChatBox(tree, cfg)` | 左下角 |
| Nameplate | `widget/game` | `NewNameplate(tree, name, cfg)` | 场景绝对定位 |
| Inventory | `widget/game` | `NewInventory(tree, r, c, cfg)` | 右侧居中 |
| SkillTree | `widget/game` | `NewSkillTree(tree, cfg)` | 左侧居中 |
| Scoreboard | `widget/game` | `NewScoreboard(tree, cfg)` | 屏幕中央 |
| LootWindow | `widget/game` | `NewLootWindow(tree, cfg)` | 屏幕中央 |
| DialogueBox | `widget/game` | `NewDialogueBox(tree, cfg)` | 底部居中 |
| FloatingText | `widget/game` | `NewFloatingText(tree, ...)` | 场景绝对定位 |
| ItemTooltip | `widget/game` | `NewItemTooltip(tree, cfg)` | 跟随鼠标 |
| NotificationToast | `widget/game` | `NewNotificationToast(tree, text, cfg)` | 屏幕上方 |
