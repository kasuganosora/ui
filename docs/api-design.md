# API 设计

## 概述

GoUI 提供三种 API 风格，满足不同使用场景：

1. **声明式 API** - 用 Go 代码构建元素树（类似 React/SwiftUI）
2. **HTML+CSS API** - 用 HTML 和 CSS 字符串定义 UI
3. **即时模式 API** - 每帧重新描述 UI（类似 Dear ImGui）

三种风格可以混合使用。

## 1. 声明式 API（主要推荐）

### 基本用法

```go
package main

import "github.com/kasuganosora/ui"

func main() {
    app := goui.NewApp(goui.AppOptions{
        Title:  "My App",
        Width:  1280,
        Height: 720,
        Theme:  goui.ThemeDark,
    })

    app.Mount(MyPage())
    app.Run()
}

func MyPage() goui.Element {
    count := goui.Signal(0)

    return goui.Div(
        goui.Class("container"),
        goui.Flex(goui.Column, goui.AlignCenter, goui.Gap(16)),

        goui.Text("Hello GoUI",
            goui.Style("font-size: 24px; color: #fff;"),
        ),

        goui.Button(
            goui.Text(fmt.Sprintf("Count: %d", count.Get())),
            goui.OnClick(func() {
                count.Set(count.Get() + 1)
            }),
        ),

        goui.Input(
            goui.Placeholder("Type something..."),
            goui.OnChange(func(value string) {
                fmt.Println("Input:", value)
            }),
        ),
    )
}
```

### 元素构建函数

```go
// 容器
goui.Div(children ...any) Element
goui.Span(children ...any) Element
goui.ScrollView(children ...any) Element

// 基础
goui.Text(content string, opts ...any) Element
goui.Image(src string, opts ...any) Element
goui.Icon(name string, opts ...any) Element

// 交互
goui.Button(children ...any) Element
goui.Input(opts ...any) Element
goui.TextArea(opts ...any) Element
goui.Checkbox(opts ...any) Element
goui.Radio(opts ...any) Element
goui.Switch(opts ...any) Element
goui.Slider(opts ...any) Element
goui.Select(opts ...any) Element

// 容器
goui.Modal(opts ...any) Element
goui.Drawer(opts ...any) Element
goui.Tabs(opts ...any) Element
goui.Collapse(opts ...any) Element

// 布局
goui.Row(children ...any) Element
goui.Col(children ...any) Element
goui.Space(children ...any) Element
goui.SubWindow(children ...any) Element
```

### 属性设置

使用函数式选项模式：

```go
goui.Button(
    goui.Text("Submit"),
    goui.Class("btn-primary"),
    goui.Style("padding: 8px 16px;"),
    goui.Disabled(isLoading.Get()),
    goui.OnClick(handleSubmit),
    goui.Tooltip("Click to submit"),
)
```

### 条件渲染

```go
goui.Div(
    goui.If(isLoggedIn.Get(),
        goui.Text("Welcome!"),
    ).Else(
        goui.Button(goui.Text("Login")),
    ),
)
```

### 列表渲染

```go
goui.Div(
    goui.For(items.Get(), func(item Item, index int) goui.Element {
        return goui.Div(
            goui.Key(item.ID), // 用于 diff 优化
            goui.Text(item.Name),
        )
    }),
)
```

### 引用（Ref）

```go
inputRef := goui.Ref[*goui.InputComponent]()

goui.Input(
    goui.BindRef(inputRef),
)

// 之后可以直接操作
inputRef.Get().Focus()
inputRef.Get().SetValue("hello")
```

## 2. HTML+CSS API

### 从字符串构建

```go
ui := goui.FromHTML(`
    <div class="container">
        <div class="header">
            <text style="font-size: 24px;">{{title}}</text>
        </div>
        <div class="content">
            <button id="btn-add" class="btn primary">Add Item</button>
            <list id="item-list">
                <template>
                    <div class="item">
                        <text>{{name}}</text>
                        <text class="desc">{{description}}</text>
                    </div>
                </template>
            </list>
        </div>
    </div>
`)

ui.SetCSS(`
    .container {
        display: flex;
        flex-direction: column;
        width: 100%;
        height: 100%;
    }
    .header {
        padding: 16px;
        background: #1a1a2e;
        border-bottom: 1px solid #333;
    }
    .content {
        flex: 1;
        padding: 16px;
        overflow-y: auto;
    }
    .btn.primary {
        background: #4a9eff;
        color: #fff;
        padding: 8px 16px;
        border-radius: 4px;
    }
    .btn.primary:hover {
        background: #3a8eef;
    }
    .item {
        padding: 12px;
        border-bottom: 1px solid #222;
    }
`)

// 绑定数据
ui.Bind("title", "My App")

// 绑定事件
ui.Query("#btn-add").On(goui.EventClick, func(e *goui.Event) {
    // ...
})

// 操作列表
list := ui.Query("#item-list")
list.Append(map[string]any{"name": "Item 1", "description": "First item"})

app.Mount(ui)
```

### 从文件加载

```go
ui, err := goui.LoadHTML("ui/main.html")
ui.LoadCSS("ui/style.css")
```

### 数据绑定

```go
// 单向绑定（数据 → UI）
ui.Bind("username", username)

// 双向绑定（数据 ↔ UI，用于 input 等）
ui.BindTwo("searchText", &searchText)

// 绑定列表数据
ui.BindList("items", &items)

// 计算绑定
ui.BindComputed("fullName", func() string {
    return firstName.Get() + " " + lastName.Get()
})
```

## 3. 即时模式 API

适合游戏 HUD、调试工具等需要每帧动态生成的场景：

```go
app.OnFrame(func(ctx *goui.IMContext) {
    // 面板
    if ctx.BeginPanel("Debug Info", 10, 10, 300, 0) {
        ctx.Text(fmt.Sprintf("FPS: %d", fps))
        ctx.Text(fmt.Sprintf("Entities: %d", entityCount))
        ctx.Separator()

        if ctx.Button("Spawn Entity") {
            spawnEntity()
        }

        if ctx.SliderFloat("Speed", &speed, 0, 100) {
            updateSpeed(speed)
        }

        if ctx.Checkbox("Show Hitboxes", &showHitboxes) {
            toggleHitboxes(showHitboxes)
        }

        ctx.EndPanel()
    }

    // 游戏 HUD
    ctx.SetNextPos(screenW/2, 20, goui.AnchorCenter)
    ctx.HealthBar(player.HP, player.MaxHP, 200, 20)

    ctx.SetNextPos(screenW-10, screenH-10, goui.AnchorBottomRight)
    ctx.Text(fmt.Sprintf("Gold: %d", player.Gold))
})
```

### 即时模式特点

- 无需维护 UI 状态树，每帧重新描述
- 状态由用户代码持有
- 内部自动做帧间 diff 以优化渲染
- 布局由调用顺序和位置 API 决定
- 可与声明式 UI 混合使用

### 即时模式 Widget

```go
// 文本
ctx.Text(text string)
ctx.TextColored(text string, color Color)
ctx.TextWrapped(text string)

// 按钮
ctx.Button(label string) bool                    // 返回是否被点击
ctx.ImageButton(img ImageHandle, w, h float32) bool
ctx.RadioButton(label string, active bool) bool

// 输入
ctx.InputText(label string, buf *string) bool    // 返回是否修改
ctx.InputInt(label string, val *int) bool
ctx.InputFloat(label string, val *float64) bool
ctx.SliderFloat(label string, val *float64, min, max float64) bool
ctx.SliderInt(label string, val *int, min, max int) bool
ctx.ColorEdit(label string, col *Color) bool

// 布局
ctx.SameLine()
ctx.NewLine()
ctx.Indent()
ctx.Unindent()
ctx.Separator()
ctx.Space(size float32)

// 容器
ctx.BeginPanel(title string, x, y, w, h float32) bool
ctx.EndPanel()
ctx.BeginGroup()
ctx.EndGroup()
ctx.BeginChild(id string, w, h float32)
ctx.EndChild()
ctx.BeginTabBar(id string)
ctx.TabItem(label string) bool
ctx.EndTabBar()

// 表格
ctx.BeginTable(id string, columns int)
ctx.TableNextColumn()
ctx.TableNextRow()
ctx.EndTable()

// 树
ctx.TreeNode(label string) bool
ctx.TreePop()

// 弹出层
ctx.BeginPopup(id string) bool
ctx.EndPopup()
ctx.OpenPopup(id string)
ctx.BeginTooltip()
ctx.EndTooltip()

// 绘图
ctx.DrawLine(from, to Vec2, color Color, thickness float32)
ctx.DrawRect(min, max Vec2, color Color)
ctx.DrawCircle(center Vec2, radius float32, color Color)
ctx.DrawImage(img ImageHandle, min, max Vec2)
```

## 4. 样式 API

### 内联样式

```go
goui.Style("padding: 16px; background: #333; border-radius: 8px;")
```

### 样式对象

```go
style := goui.NewStyle().
    Padding(16).
    Background(goui.ColorHex("#333")).
    BorderRadius(8).
    Display(goui.Flex).
    FlexDirection(goui.Column).
    Gap(8)

goui.Div(style, /* children */)
```

### 样式组合

```go
baseStyle := goui.NewStyle().Padding(8).BorderRadius(4)
primaryStyle := baseStyle.Extend().Background(goui.ColorHex("#4a9eff")).Color(goui.White)
dangerStyle := baseStyle.Extend().Background(goui.ColorHex("#ff4a4a")).Color(goui.White)
```

### 动态样式

```go
goui.Div(
    goui.DynamicStyle(func() *goui.StyleBuilder {
        s := goui.NewStyle().Padding(8)
        if isActive.Get() {
            s.Background(goui.ColorHex("#4a9eff"))
        } else {
            s.Background(goui.ColorHex("#333"))
        }
        return s
    }),
)
```

## 5. 动画 API

```go
// 简单过渡
goui.Div(
    goui.Transition("background", 200*time.Millisecond, goui.EaseInOut),
)

// 关键帧动画
fadeIn := goui.Keyframes(
    goui.Frame(0, goui.NewStyle().Opacity(0)),
    goui.Frame(100, goui.NewStyle().Opacity(1)),
)

goui.Div(
    goui.Animation(fadeIn, 300*time.Millisecond),
)

// 命令式动画
anim := goui.Animate(element).
    Duration(500 * time.Millisecond).
    Ease(goui.EaseOutBack).
    To(goui.NewStyle().TranslateX(100).Opacity(1)).
    OnComplete(func() { fmt.Println("done") })

anim.Play()
anim.Pause()
anim.Reverse()
```

## 6. 游戏引擎集成 API

```go
// 在已有的游戏窗口/渲染上下文中创建 UI
ui := goui.NewEmbedded(goui.EmbedOptions{
    // 提供已有的渲染后端接口
    Backend: myGameBackend,
    // 或提供 Vulkan/OpenGL 原生句柄
    VulkanDevice:     device,
    VulkanRenderPass: renderPass,
    // UI 区域
    Viewport: goui.Rect{0, 0, 1920, 1080},
})

// 在游戏循环中调用
func gameLoop() {
    // 传入游戏的输入事件
    ui.HandleInput(gameInputEvents)

    // 更新 UI（布局计算等）
    ui.Update(deltaTime)

    // 获取渲染命令（不立即执行）
    commands := ui.RenderCommands()

    // 在游戏渲染管线的合适阶段执行 UI 渲染
    // 方式1: 让 GoUI 直接渲染
    ui.Render()

    // 方式2: 自行处理渲染命令
    for _, cmd := range commands {
        myEngine.ExecuteUICommand(cmd)
    }
}
```
