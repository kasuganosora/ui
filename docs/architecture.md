# 架构总览

## 设计目标

1. **跨平台** - Windows、Linux、macOS、Android、iOS
2. **最低系统依赖** - 不依赖 CGO（纯 Go 实现核心逻辑），渲染后端通过可选的 CGO 或系统调用接入
3. **游戏友好** - 可嵌入任意游戏引擎，不抢占渲染循环，支持共享 GPU 上下文
4. **独立可用** - 也可作为独立应用的 UI 框架使用，自带窗口管理
5. **灵活性** - 同时提供声明式（HTML+CSS）和命令式（即时模式 API）两种使用方式

## 分层架构

```
┌─────────────────────────────────────────────────┐
│                  Application                     │
│         (用户代码 / 游戏引擎集成层)                │
├─────────────────────────────────────────────────┤
│                Component Layer                   │
│    Button / Input / Select / Scroll / Modal ...  │
├─────────────────────────────────────────────────┤
│                 Layout Engine                     │
│   HTML+CSS Parser → Style Resolver → Layout Algo │
│   Grid / Flexbox / Absolute / Flow               │
├─────────────────────────────────────────────────┤
│               Core Layer (goui)                  │
│  Element Tree │ Event System │ Animation │ State │
├─────────────────────────────────────────────────┤
│             Rendering Abstraction                │
│          RenderCommand → RenderBackend           │
├──────────┬──────────┬──────────┬────────────────┤
│  Vulkan  │  OpenGL  │ DirectX  │   Software     │
│ (primary)│(secondary│(planned) │  (fallback)    │
│          │          │          │                │
├──────────┴──────────┴──────────┴────────────────┤
│              Platform Layer                      │
│  Window │ Input │ IME │ Clipboard │ DPI │ Timer  │
├──────────┬──────────┬──────────┬────────────────┤
│ Windows  │  Linux   │  macOS   │  Mobile        │
│ (Win32)  │(X11/Way.)│ (Cocoa)  │ (Android/iOS)  │
└──────────┴──────────┴──────────┴────────────────┘
```

## 核心模块

### 1. Platform Layer（平台层）

最底层，负责操作系统交互。每个平台一个独立实现：

- **窗口管理** - 创建/销毁/调整窗口，全屏切换
- **输入采集** - 键盘、鼠标、触摸、手柄输入的原始事件
- **IME 管理** - 输入法候选窗定位、组合文本事件、提交事件
- **系统集成** - 剪贴板、DPI 感知、系统主题检测、文件对话框
- **计时器** - 高精度帧计时

接口定义：

```go
type Platform interface {
    CreateWindow(opts WindowOptions) (Window, error)
    PollEvents() []Event
    GetClipboard() string
    SetClipboard(text string)
    GetDPI() float64
    Terminate()
}

type Window interface {
    Size() (width, height int)
    SetSize(width, height int)
    SetTitle(title string)
    SetFullscreen(fullscreen bool)
    ShouldClose() bool
    GetNativeHandle() uintptr  // HWND / X11 Window / NSWindow
    SwapBuffers()
}
```

### 2. Rendering Abstraction（渲染抽象层）

将所有 UI 绘制操作抽象为统一的渲染命令流：

```go
type RenderBackend interface {
    Init(window Window) error
    BeginFrame()
    EndFrame()
    Resize(width, height int)
    Destroy()

    // 基础绘制
    DrawRect(rect Rect, style DrawStyle)
    DrawRoundedRect(rect Rect, radii [4]float32, style DrawStyle)
    DrawText(text string, pos Vec2, style TextStyle) TextMetrics
    DrawImage(img ImageHandle, rect Rect, opts ImageDrawOptions)
    DrawLine(from, to Vec2, style LineStyle)
    DrawPath(path Path, style DrawStyle)

    // 裁剪与变换
    PushClip(rect Rect)
    PopClip()
    PushTransform(mat Mat3)
    PopTransform()

    // 资源管理
    CreateTexture(data []byte, width, height int, format TexFormat) TextureHandle
    UpdateTexture(handle TextureHandle, data []byte, region Rect)
    DestroyTexture(handle TextureHandle)

    // 字体
    LoadFont(data []byte) (FontHandle, error)
    MeasureText(text string, style TextStyle) TextMetrics
}
```

### 2.5 Font System（字体系统）

基于 FreeType 的完整字体引擎，东亚语言一等公民支持：

- **FreeType 集成** - 通过 CGO 绑定 FreeType，支持 TrueType/OpenType/WOFF
- **SDF 渲染** - FreeType 2.11+ 原生 SDF 光栅化，分辨率无关
- **字形缓存** - LRU 缓存 + 纹理图集，CJK 大字符集友好
- **文本排版** - CJK 断行规则、标点挤压、中英混排间距
- **字体回退** - 自动回退链（CJK → Emoji → 符号），系统字体发现
- **注音/着重号** - Ruby 注音、着重号等东亚排版特性

详见 [字体系统文档](./font-system.md)。

### 3. Core Layer（核心层）

UI 框架的核心逻辑，纯 Go 实现：

- **Element Tree** - UI 元素树，类似 DOM
- **Event System** - 冒泡/捕获事件分发，焦点管理
- **State Management** - 响应式状态管理，脏标记优化
- **Animation** - 缓动函数、关键帧动画、过渡动画
- **Theming** - 主题系统，支持运行时切换

### 4. Layout Engine（布局引擎）

支持 HTML+CSS 子集的布局计算：

- **CSS Parser** - 解析 CSS 子集（布局相关属性）
- **Style Resolver** - 样式层叠、继承、计算
- **Layout Algorithm** - Flexbox、Grid、Flow、Absolute 布局算法
- **自适应系统** - 响应式断点、百分比尺寸、min/max 约束

### 5. Component Layer（组件层）

基于核心层构建的可复用组件库。

## 实现规范：严格遵循 DDD（领域驱动设计）

**所有模块的实现必须严格按照 DDD（Domain-Driven Design）原则进行设计和编码。** 这是项目的硬性要求，不是可选建议。

### 核心 DDD 概念在本项目中的映射

#### 限界上下文（Bounded Context）

每个顶层包对应一个限界上下文，上下文之间通过明确定义的接口通信，禁止跨上下文直接引用内部类型：

| 限界上下文 | 包路径 | 核心领域 |
|-----------|--------|---------|
| 平台 | `platform/` | 窗口管理、输入采集、系统交互 |
| 渲染 | `render/` | 渲染命令、GPU 资源、绘制管线 |
| 字体 | `font/` | 字形光栅化、文本排版、字形缓存 |
| 核心 | `core/` | 元素树、事件系统、状态管理 |
| 布局 | `layout/` | 样式解析、布局算法、尺寸计算 |
| 组件 | `widget/` | UI 组件逻辑、组件生命周期 |
| 主题 | `theme/` | 主题定义、样式令牌 |
| 动画 | `anim/` | 缓动、关键帧、过渡 |

#### 实体（Entity）与值对象（Value Object）

- **实体** - 有唯一标识、有生命周期的对象。例如：`Element`（有 ID 的 UI 元素）、`Window`（系统窗口）、`FontFace`（已加载的字体面）
- **值对象** - 无标识、不可变、通过值比较的对象。例如：`Vec2`、`Rect`、`Color`、`TextStyle`、`DrawStyle`、`GlyphMetrics`

```go
// 值对象示例 - 不可变，通过值比较
type Color struct {
    R, G, B, A float32
}

// 实体示例 - 有唯一 ID，有生命周期
type Element struct {
    id       ElementID  // 唯一标识
    // ...
}
```

#### 聚合根（Aggregate Root）

每个聚合有且只有一个聚合根，外部只能通过聚合根操作聚合内的对象：

- `ElementTree` 是元素聚合的根 —— 外部不直接操作子元素，而是通过树的方法
- `GlyphAtlas` 是字形纹理聚合的根 —— 外部不直接操作纹理页，而是通过 Atlas 分配
- `CommandBuffer` 是渲染命令聚合的根 —— 外部不直接操作命令数组

```go
// 正确：通过聚合根操作
tree.AppendChild(parentID, childElement)
atlas.AllocateGlyph(glyphKey, bitmap)

// 错误：绕过聚合根直接操作内部
parent.children = append(parent.children, child)  // 禁止
atlas.pages[0].regions = append(...)               // 禁止
```

#### 领域服务（Domain Service）

不属于任何单个实体的业务逻辑封装为领域服务：

- `LayoutService` - 布局计算（涉及元素树 + 样式 + 约束）
- `EventDispatcher` - 事件分发（涉及元素树 + 焦点 + 事件）
- `TextShaper` - 文本排版（涉及字体 + 断行规则 + 度量）
- `HitTester` - 命中测试（涉及元素树 + 布局结果）

#### 仓储（Repository）

资源的持久化和查询通过仓储模式抽象：

```go
// 字体仓储
type FontRepository interface {
    Register(name string, data []byte) (FontID, error)
    FindByName(name string) (FontFace, error)
    FindFallback(r rune) (FontFace, error)
}

// 纹理仓储
type TextureRepository interface {
    Create(width, height int, format TexFormat) (TextureHandle, error)
    Update(handle TextureHandle, region Rect, data []byte) error
    Delete(handle TextureHandle)
}
```

#### 领域事件（Domain Event）

模块间通过领域事件解耦，而非直接调用：

```go
// 领域事件示例
type ElementMounted struct { ElementID ElementID }
type ElementUnmounted struct { ElementID ElementID }
type StyleChanged struct { ElementID ElementID; Properties []string }
type LayoutDirty struct { ElementID ElementID }
type ThemeChanged struct { ThemeName string }
type FocusChanged struct { OldID, NewID ElementID }
```

#### 防腐层（Anti-Corruption Layer）

与外部系统（操作系统 API、GPU 驱动、FreeType 等）的交互必须通过防腐层隔离，确保领域模型不被外部概念污染：

```go
// 防腐层示例：FreeType 绑定
// font/freetype/ 包是防腐层，将 C 的 FreeType API 转换为 Go 领域概念
// 领域层 (font/) 只依赖 Go 接口，不知道 FreeType 的存在

// 防腐层示例：Win32 平台
// platform/windows/ 将 Win32 HWND/MSG 转换为领域的 Window/Event
// core/ 只依赖 platform.Platform 接口
```

### DDD 实施检查清单

实现每个模块时，必须确认以下要点：

1. **包边界 = 限界上下文边界** —— 包的公开 API 就是上下文的契约
2. **依赖方向单向** —— 上层依赖下层接口，下层不知道上层存在；同层之间通过事件或共享接口通信
3. **值对象不可变** —— `Vec2`、`Color`、`Rect` 等创建后不修改，需要变化时创建新实例
4. **实体通过 ID 引用** —— 跨聚合引用使用 ID 而非指针，防止聚合边界泄漏
5. **聚合根守护不变量** —— 所有修改聚合内部状态的操作都必须经过聚合根的方法，由聚合根保证一致性
6. **领域逻辑不泄漏到应用层** —— 布局计算、事件分发、样式解析等逻辑在领域层，应用层只做编排
7. **防腐层隔离外部依赖** —— CGO 绑定、系统调用、第三方库调用都在防腐层内，领域模型保持纯净
8. **通用语言（Ubiquitous Language）** —— 代码中的命名与设计文档保持一致，不引入文档中未定义的概念

## 实现规范：高可测试性设计

**所有模块必须具备高可测试性，能够在无窗口、无 GPU、无操作系统的环境下进行完整的单元测试。** 这是与 DDD 同级的硬性要求。

### 核心原则

1. **UI 层完全可观测** —— 通过独立接口可以查询任意时刻整个 UI 树的完整状态，无需真实渲染
2. **组件可独立测试** —— 每个组件可以脱离窗口系统和渲染后端单独创建、操作、断言
3. **所有外部依赖可替换** —— Platform、RenderBackend、FontRasterizer 等均通过接口注入，测试时使用 Mock 实现

### 测试基础设施

#### TestContext：无头测试环境

提供一个零依赖的测试上下文，不需要窗口、GPU 或字体文件：

```go
// 创建无头测试环境 —— 纯内存，零外部依赖
ctx := goui.NewTestContext(goui.TestOptions{
    Width:  800,
    Height: 600,
})
defer ctx.Destroy()

// 挂载 UI
ctx.Mount(myComponent)

// 驱动一帧更新（布局 + 状态计算，不渲染）
ctx.Update()
```

#### UIAccessor：UI 树查询接口

独立于渲染的查询接口，可以检索所有组件信息：

```go
type UIAccessor interface {
    // === 元素查询 ===

    // 通过 ID 查找元素
    FindByID(id string) (ElementInfo, bool)
    // 通过类型查找所有匹配元素
    FindByType(typeName string) []ElementInfo
    // CSS 选择器查询（单个）
    QuerySelector(selector string) (ElementInfo, bool)
    // CSS 选择器查询（所有）
    QuerySelectorAll(selector string) []ElementInfo
    // 获取整个元素树的快照
    Snapshot() TreeSnapshot

    // === 元素状态 ===

    // 获取元素的计算后样式（布局结果）
    GetComputedLayout(id string) (LayoutRect, bool)
    // 获取元素的计算后样式属性
    GetComputedStyle(id string) (ComputedStyle, bool)
    // 获取元素的当前可见文本内容
    GetTextContent(id string) (string, bool)
    // 获取元素的所有属性
    GetProperties(id string) (map[string]any, bool)
    // 检查元素是否可见（display != none, 在视口内等）
    IsVisible(id string) bool
    // 检查元素是否处于焦点
    IsFocused(id string) bool
    // 检查元素是否被禁用
    IsDisabled(id string) bool

    // === 树结构 ===

    // 获取子元素列表
    Children(id string) []ElementInfo
    // 获取父元素
    Parent(id string) (ElementInfo, bool)
    // 获取元素到根节点的路径
    AncestorPath(id string) []ElementInfo

    // === 统计信息 ===

    // 元素总数
    ElementCount() int
    // 按类型统计元素数量
    ElementCountByType() map[string]int
}

// 元素快照 —— 值对象，包含测试时需要的所有信息
type ElementInfo struct {
    ID         string
    Type       string              // "Button", "Input", "Div", ...
    Classes    []string
    Layout     LayoutRect          // 计算后的位置和尺寸
    Visible    bool
    Focused    bool
    Disabled   bool
    Text       string              // 可见文本内容
    Properties map[string]any      // 组件特有属性
    ChildCount int
}
```

#### UIOperator：UI 操作接口

模拟用户交互，无需真实输入设备：

```go
type UIOperator interface {
    // === 鼠标操作 ===

    Click(id string) error                         // 点击元素（居中位置）
    ClickAt(x, y float32) error                    // 点击指定坐标
    DoubleClick(id string) error                   // 双击
    RightClick(id string) error                    // 右键点击
    Hover(id string) error                         // 悬停（触发 mouseenter/mouseover）
    MouseDown(id string, button MouseButton) error // 按下鼠标键
    MouseUp(id string, button MouseButton) error   // 释放鼠标键
    DragTo(fromID, toID string) error              // 从一个元素拖拽到另一个

    // === 键盘操作 ===

    TypeText(id string, text string) error         // 在输入框中输入文本
    PressKey(key Key, modifiers ...Modifier) error // 按下按键
    KeySequence(keys ...Key) error                 // 按键序列

    // === 焦点操作 ===

    Focus(id string) error                         // 设置焦点
    Blur(id string) error                          // 移除焦点
    TabForward() error                             // Tab 切换到下一个
    TabBackward() error                            // Shift+Tab 切换到上一个

    // === 滚动操作 ===

    ScrollTo(id string, x, y float32) error        // 滚动到指定位置
    ScrollBy(id string, dx, dy float32) error      // 相对滚动

    // === 组件特定操作 ===

    SetValue(id string, value any) error           // 设置输入值（Input/Select/Slider 等）
    Toggle(id string) error                        // 切换状态（Checkbox/Switch）
    SelectOption(id string, option string) error   // Select 选择选项
    ExpandNode(id string, nodeID string) error     // 展开树节点
    CloseDialog(id string) error                   // 关闭对话框
}
```

#### RenderInspector：渲染命令检查

不需要 GPU，直接检查生成的渲染命令：

```go
type RenderInspector interface {
    // 获取当前帧的所有渲染命令
    Commands() []RenderCommand
    // 按类型过滤渲染命令
    CommandsByType(cmdType CommandType) []RenderCommand
    // 获取指定元素生成的渲染命令
    CommandsForElement(id string) []RenderCommand
    // 统计 draw call 数量
    DrawCallCount() int
    // 统计顶点数量
    VertexCount() int
    // 获取纹理引用列表（用于验证纹理图集使用）
    TextureRefs() []TextureHandle
}
```

### Mock 实现

所有外部依赖提供开箱即用的 Mock：

```go
// 测试用 Mock —— 在 goui/testing 包中提供
type MockPlatform struct { ... }     // 无窗口、纯内存的平台层
type MockRenderer struct { ... }     // 只收集命令、不执行 GPU 操作的渲染后端
type MockFontEngine struct { ... }   // 返回固定度量值的字体引擎（不需要 FreeType）
type MockClipboard struct { ... }    // 内存剪贴板
type MockTimer struct { ... }        // 可手动推进的计时器

// MockFontEngine 使用固定宽度度量，便于布局断言
func NewMockFontEngine(charWidth, lineHeight float32) *MockFontEngine
```

### 测试示例

```go
func TestButtonClick(t *testing.T) {
    ctx := goui.NewTestContext(goui.TestOptions{Width: 400, Height: 300})
    defer ctx.Destroy()

    clicked := false
    ctx.Mount(goui.Button(
        goui.ID("btn"),
        goui.Text("Submit"),
        goui.OnClick(func() { clicked = true }),
    ))
    ctx.Update()

    // 查询：检查按钮存在且可见
    btn, ok := ctx.Accessor().FindByID("btn")
    assert.True(t, ok)
    assert.Equal(t, "Submit", btn.Text)
    assert.True(t, btn.Visible)
    assert.False(t, btn.Disabled)

    // 操作：模拟点击
    err := ctx.Operator().Click("btn")
    assert.NoError(t, err)
    ctx.Update()

    // 断言：回调被触发
    assert.True(t, clicked)
}

func TestFormValidation(t *testing.T) {
    ctx := goui.NewTestContext(goui.TestOptions{Width: 600, Height: 400})
    defer ctx.Destroy()

    ctx.Mount(LoginForm())
    ctx.Update()

    // 不填写直接提交
    ctx.Operator().Click("submit-btn")
    ctx.Update()

    // 检查错误提示出现
    errors := ctx.Accessor().QuerySelectorAll(".form-error")
    assert.Equal(t, 2, len(errors))
    assert.Equal(t, "用户名不能为空", errors[0].Text)
    assert.Equal(t, "密码不能为空", errors[1].Text)

    // 填写用户名
    ctx.Operator().TypeText("username-input", "admin")
    ctx.Operator().TypeText("password-input", "123456")
    ctx.Operator().Click("submit-btn")
    ctx.Update()

    // 错误提示消失
    errors = ctx.Accessor().QuerySelectorAll(".form-error")
    assert.Equal(t, 0, len(errors))
}

func TestInventoryDragDrop(t *testing.T) {
    ctx := goui.NewTestContext(goui.TestOptions{Width: 800, Height: 600})
    defer ctx.Destroy()

    inv := NewInventory(8, 4) // 8x4 背包
    inv.SetItem(0, 0, &Item{Name: "Sword", Count: 1})
    ctx.Mount(inv.Component())
    ctx.Update()

    // 验证物品在 slot(0,0)
    slot00 := ctx.Accessor().FindByID("slot-0-0")
    assert.Equal(t, "Sword", slot00.Properties["item_name"])

    // 拖拽到 slot(1,0)
    ctx.Operator().DragTo("slot-0-0", "slot-1-0")
    ctx.Update()

    // 验证物品已移动
    slot00, _ = ctx.Accessor().FindByID("slot-0-0")
    assert.Nil(t, slot00.Properties["item_name"])
    slot10, _ := ctx.Accessor().FindByID("slot-1-0")
    assert.Equal(t, "Sword", slot10.Properties["item_name"])
}

func TestLayoutComputation(t *testing.T) {
    ctx := goui.NewTestContext(goui.TestOptions{Width: 400, Height: 300})
    defer ctx.Destroy()

    ctx.Mount(goui.Div(
        goui.ID("container"),
        goui.Style("display: flex; width: 400px; gap: 10px;"),
        goui.Div(goui.ID("a"), goui.Style("flex: 1;")),
        goui.Div(goui.ID("b"), goui.Style("flex: 2;")),
    ))
    ctx.Update()

    // 直接断言布局计算结果
    layoutA, _ := ctx.Accessor().GetComputedLayout("a")
    layoutB, _ := ctx.Accessor().GetComputedLayout("b")
    assert.InDelta(t, 126.67, layoutA.Width, 1.0)  // (400-10) / 3
    assert.InDelta(t, 253.33, layoutB.Width, 1.0)  // (400-10) * 2/3
}

func TestRenderCommands(t *testing.T) {
    ctx := goui.NewTestContext(goui.TestOptions{Width: 400, Height: 300})
    defer ctx.Destroy()

    ctx.Mount(goui.Button(goui.ID("btn"), goui.Text("OK")))
    ctx.Update()

    // 检查渲染命令（不需要 GPU）
    inspector := ctx.RenderInspector()
    cmds := inspector.CommandsForElement("btn")
    assert.True(t, len(cmds) >= 2) // 至少有背景矩形 + 文本

    rectCmds := inspector.CommandsByType(goui.CmdRect)
    textCmds := inspector.CommandsByType(goui.CmdText)
    assert.True(t, len(rectCmds) > 0)
    assert.True(t, len(textCmds) > 0)
}
```

### 可测试性检查清单

实现每个模块时，必须确认：

1. **可无头运行** —— 所有逻辑可以在 `TestContext` 中运行，不依赖窗口/GPU/字体文件
2. **状态可观测** —— 组件的所有用户可感知状态都能通过 `UIAccessor` 查询到
3. **交互可模拟** —— 所有用户交互都能通过 `UIOperator` 触发，效果与真实输入一致
4. **渲染可检查** —— 生成的渲染命令可以通过 `RenderInspector` 检查，不需要像素比对
5. **时间可控制** —— 动画、定时器、延迟等时间相关逻辑使用可注入的 `Clock` 接口，测试中可手动推进
6. **依赖可注入** —— 每个依赖外部资源的组件都通过接口注入，测试包中提供对应 Mock
7. **单元测试独立** —— 每个组件的测试完全独立，不依赖其他测试的执行顺序或全局状态

## 实现规范：主题系统与组件自定义

**每个游戏、每个应用都有自己独特的视觉风格。** 主题系统是本 UI 库的核心特性之一，必须深度贯穿所有组件。不是简单的"换颜色"，而是从设计令牌、组件样式到渲染行为的全链路可定制。

### 设计原则

1. **零默认样式假设** —— 组件不硬编码任何视觉样式，所有外观由主题驱动
2. **逐层覆盖** —— 全局主题 → 区域主题 → 组件级主题 → 实例级样式，逐层精细覆盖
3. **运行时热切换** —— 主题可在运行时整体切换或局部替换，无需重建 UI 树
4. **组合优于继承** —— 主题通过令牌组合而非类继承扩展，避免深层嵌套陷阱
5. **类型安全** —— 设计令牌有明确类型（Color、Size、Font 等），编译期发现拼写错误

### 三层架构

```
┌──────────────────────────────────────────────┐
│            Instance Style Override             │
│   button.Style(Variant("danger"), Size("lg")) │  ← 单个实例的样式覆盖
├──────────────────────────────────────────────┤
│           Component Style Rules                │
│   ButtonStyle { Default, Hover, Active, ... }  │  ← 组件状态样式规则
├──────────────────────────────────────────────┤
│              Design Tokens                     │
│   Colors / Typography / Spacing / Shadows ...  │  ← 原子级设计令牌
└──────────────────────────────────────────────┘
```

### 第一层：设计令牌（Design Token）

设计令牌是主题的最小单元，定义基础视觉原子。所有组件样式最终引用令牌，不硬编码具体值：

```go
// 设计令牌定义 —— 值对象
type Token struct {
    Key   string  // 令牌路径，如 "color.primary"
    Value any     // 具体值：Color / float32 / string / Edges 等
}

// 主题 —— 一组设计令牌的集合
type Theme struct {
    name   string
    tokens map[string]any      // 扁平化的令牌映射
    parent *Theme              // 可选的父主题（用于继承缺省值）
}

// 令牌类别
type Tokens struct {
    // === 颜色系统 ===
    Colors struct {
        // 语义色（组件引用这些，而非具体色值）
        Primary       Color   // 主色调
        PrimaryHover  Color   // 主色调悬停态
        PrimaryActive Color   // 主色调按下态
        Secondary     Color
        Success       Color
        Warning       Color
        Danger        Color

        // 表面色
        BgPrimary     Color   // 主背景
        BgSecondary   Color   // 次级背景（卡片、面板）
        BgElevated    Color   // 浮层背景（弹窗、下拉）

        // 文本色
        TextPrimary   Color   // 主文本
        TextSecondary Color   // 次要文本
        TextDisabled  Color   // 禁用态文本
        TextInverse   Color   // 反色文本（用于深色按钮上的白字等）

        // 边框色
        Border        Color
        BorderFocus   Color

        // 自定义扩展（游戏场景常用）
        Custom map[string]Color // "hp_bar", "mana_bar", "rarity_epic" ...
    }

    // === 字体系统 ===
    Typography struct {
        FontFamily     string   // 默认字体族
        FontFamilyCJK  string   // CJK 字体族
        FontFamilyMono string   // 等宽字体族

        // 字号梯度
        FontSizeXS  float32     // 10
        FontSizeSM  float32     // 12
        FontSizeMD  float32     // 14
        FontSizeLG  float32     // 18
        FontSizeXL  float32     // 24
        FontSize2XL float32     // 32

        LineHeight  float32     // 默认行高倍数

        // 字重
        FontWeightNormal int
        FontWeightBold   int
    }

    // === 间距系统 ===
    Spacing struct {
        XS  float32  // 4
        SM  float32  // 8
        MD  float32  // 12
        LG  float32  // 16
        XL  float32  // 24
        XXL float32  // 32
    }

    // === 圆角 ===
    Radius struct {
        None  float32  // 0
        SM    float32  // 2
        MD    float32  // 4
        LG    float32  // 8
        XL    float32  // 12
        Full  float32  // 9999（完全圆角）
    }

    // === 阴影 ===
    Shadows struct {
        SM  Shadow
        MD  Shadow
        LG  Shadow
        XL  Shadow
    }

    // === 动画 ===
    Motion struct {
        DurationFast   time.Duration  // 100ms
        DurationNormal time.Duration  // 200ms
        DurationSlow   time.Duration  // 300ms
        EasingDefault  EasingFunc
        EasingEnter    EasingFunc
        EasingExit     EasingFunc
    }

    // === 层级（z-order） ===
    ZIndex struct {
        Dropdown int32
        Modal    int32
        Popover  int32
        Toast    int32
        Tooltip  int32
    }
}
```

### 第二层：组件样式（Component Style）

每个组件定义自己的样式结构，引用设计令牌。组件样式描述每个**状态**下的外观：

```go
// 组件样式通过 StyleFunc 从主题派生
// 这是组件与主题之间的契约
type StyleFunc[S any] func(theme *Theme) S

// 以 Button 为例
type ButtonStyle struct {
    // 各状态对应的样式
    Default  ButtonStateStyle
    Hover    ButtonStateStyle
    Active   ButtonStateStyle
    Focused  ButtonStateStyle
    Disabled ButtonStateStyle

    // 尺寸变体
    SizeSmall  ButtonSizeStyle
    SizeMedium ButtonSizeStyle
    SizeLarge  ButtonSizeStyle
}

type ButtonStateStyle struct {
    BgColor     Color
    TextColor   Color
    BorderColor Color
    BorderWidth float32
    Shadow      Shadow
}

type ButtonSizeStyle struct {
    Height   float32
    Padding  Edges
    FontSize float32
    Radius   float32
    IconSize float32
}

// 组件注册默认样式派生函数
var DefaultButtonStyleFunc = func(t *Theme) ButtonStyle {
    return ButtonStyle{
        Default: ButtonStateStyle{
            BgColor:     t.Color("color.primary"),
            TextColor:   t.Color("color.text.inverse"),
            BorderColor: ColorTransparent,
            BorderWidth: 0,
        },
        Hover: ButtonStateStyle{
            BgColor:   t.Color("color.primary.hover"),
            TextColor: t.Color("color.text.inverse"),
        },
        Disabled: ButtonStateStyle{
            BgColor:   t.Color("color.bg.secondary"),
            TextColor: t.Color("color.text.disabled"),
        },
        SizeMedium: ButtonSizeStyle{
            Height:   t.Size("spacing.xl") + t.Size("spacing.sm"),
            Padding:  EdgesSymmetric(t.Size("spacing.sm"), t.Size("spacing.md")),
            FontSize: t.Size("typography.font_size.md"),
            Radius:   t.Size("radius.md"),
        },
        // ...
    }
}
```

### 第三层：实例级覆盖

单个组件实例可以直接覆盖样式，不影响同类其他实例：

```go
// 常规按钮 —— 使用主题默认样式
Button(Text("普通按钮"))

// 变体按钮 —— 使用预定义变体
Button(Text("危险操作"), Variant("danger"))
Button(Text("次要按钮"), Variant("outline"))

// 自定义覆盖 —— 仅此实例
Button(
    Text("自定义"),
    StyleOverride(func(s *ButtonStyle) {
        s.Default.BgColor = ColorHex("#ff6b00")
        s.Default.Radius = 20
        s.Hover.BgColor = ColorHex("#ff8533")
    }),
)

// 游戏场景：完全自定义外观的技能按钮
Button(
    Text(""),
    StyleOverride(func(s *ButtonStyle) {
        s.Default.BgColor = ColorTransparent
        s.Default.BorderWidth = 2
        s.Default.BorderColor = theme.Custom["rarity_epic"]
    }),
    BackgroundImage(skillIcon),
    CooldownOverlay(remainingTime),
)
```

### 主题管理器

```go
// ThemeManager 是主题系统的聚合根
type ThemeManager struct {
    active     *Theme                          // 当前活跃主题
    themes     map[string]*Theme               // 已注册的主题
    overrides  map[ElementID]*ThemeOverride     // 区域级主题覆盖
    styleCache map[cacheKey]any                 // 已计算样式缓存
    listeners  []func(ThemeChangedEvent)        // 主题变更监听
}

// 注册主题
func (m *ThemeManager) Register(name string, theme *Theme)

// 切换全局主题（触发全 UI 树重绘）
func (m *ThemeManager) SetActive(name string)

// 区域级主题覆盖（子树使用不同主题）
func (m *ThemeManager) SetOverride(elementID ElementID, themeName string)

// 获取元素生效的主题（考虑区域覆盖和继承链）
func (m *ThemeManager) ResolveTheme(elementID ElementID) *Theme
```

### 区域主题（Scoped Theme）

UI 树的不同子树可以使用不同主题，常见于：

- 游戏中不同面板风格不同（背包用暗色、商店用亮色）
- 应用中嵌入的预览区域使用对比主题
- 设置面板中的主题预览

```go
// 整体暗色主题
App(
    Theme("dark"),

    // 顶部导航使用全局主题
    Navbar(...),

    // 主内容区使用全局主题
    Content(...),

    // 侧边面板局部使用亮色主题
    ThemeScope("light",
        SettingsPanel(...),
    ),

    // 游戏 HUD 使用自定义主题
    ThemeScope("game-hud",
        HPBar(...),
        SkillBar(...),
        Minimap(...),
    ),
)
```

### 组件变体系统

组件通过变体（Variant）提供预定义的风格方案，变体定义在主题中而非组件中：

```go
// 主题中注册 Button 的变体
theme.RegisterVariant("button", "danger", func(base ButtonStyle) ButtonStyle {
    base.Default.BgColor = theme.Color("color.danger")
    base.Hover.BgColor = theme.Color("color.danger.hover")
    return base
})

theme.RegisterVariant("button", "outline", func(base ButtonStyle) ButtonStyle {
    base.Default.BgColor = ColorTransparent
    base.Default.BorderWidth = 1
    base.Default.BorderColor = theme.Color("color.primary")
    base.Default.TextColor = theme.Color("color.primary")
    return base
})

theme.RegisterVariant("button", "ghost", func(base ButtonStyle) ButtonStyle {
    base.Default.BgColor = ColorTransparent
    base.Default.BorderWidth = 0
    base.Hover.BgColor = theme.Color("color.primary").WithAlpha(0.1)
    return base
})
```

### 游戏 UI 主题场景示例

```go
// 赛博朋克风格主题
cyberTheme := NewTheme("cyberpunk",
    WithColors(Colors{
        Primary:      ColorHex("#00f0ff"),
        PrimaryHover: ColorHex("#33f3ff"),
        BgPrimary:    ColorHex("#0a0a1a"),
        BgSecondary:  ColorHex("#1a1a3a"),
        BgElevated:   ColorHex("#2a2a4a"),
        TextPrimary:  ColorHex("#e0e0ff"),
        Border:       ColorHex("#00f0ff33"),
        Custom: map[string]Color{
            "neon_pink":   ColorHex("#ff2d95"),
            "neon_green":  ColorHex("#39ff14"),
            "rarity_common":    ColorHex("#808080"),
            "rarity_rare":      ColorHex("#4169e1"),
            "rarity_epic":      ColorHex("#9400d3"),
            "rarity_legendary": ColorHex("#ffa500"),
        },
    }),
    WithTypography(Typography{
        FontFamily:    "Rajdhani",
        FontFamilyCJK: "Noto Sans SC",
    }),
    WithRadius(Radius{
        SM: 0, MD: 2, LG: 4, // 锐利的赛博风
    }),
    WithShadows(Shadows{
        MD: Shadow{Color: ColorHex("#00f0ff40"), BlurRadius: 8}, // 霓虹发光
    }),
)

// 中式仙侠风格主题
xianxiaTheme := NewTheme("xianxia",
    WithColors(Colors{
        Primary:     ColorHex("#c8a864"),
        BgPrimary:   ColorHex("#1a0f0a"),
        BgSecondary: ColorHex("#2a1f1a"),
        TextPrimary: ColorHex("#e8d5a3"),
        Custom: map[string]Color{
            "hp_bar":    ColorHex("#8b0000"),
            "mp_bar":    ColorHex("#1e3a5f"),
            "gold_text": ColorHex("#ffd700"),
        },
    }),
    WithTypography(Typography{
        FontFamily:    "LXGW WenKai",
        FontFamilyCJK: "LXGW WenKai",
    }),
    WithRadius(Radius{
        SM: 0, MD: 0, LG: 0, // 无圆角，中式硬朗
    }),
)

// 商务应用主题
businessTheme := NewTheme("business",
    WithParent(DefaultLightTheme), // 继承默认亮色主题
    WithColors(Colors{
        Primary: ColorHex("#1677ff"),
    }),
)
```

### 主题与样式的 DDD 映射

| DDD 概念 | 主题系统中的对应 |
|----------|----------------|
| 值对象 | Token、Color、Shadow、ButtonStyle、ButtonStateStyle |
| 实体 | Theme（有名称标识）|
| 聚合根 | ThemeManager（管理所有主题的注册、切换、解析）|
| 领域服务 | StyleResolver（根据主题 + 组件状态计算最终样式）|
| 领域事件 | ThemeChanged、TokenOverridden |
| 仓储 | ThemeRepository（可从文件/网络加载主题定义）|

### 主题可测试性

```go
func TestThemeSwitching(t *testing.T) {
    ctx := goui.NewTestContext(goui.TestOptions{Width: 400, Height: 300})
    defer ctx.Destroy()

    ctx.SetTheme(darkTheme)
    ctx.Mount(Button(ID("btn"), Text("OK")))
    ctx.Update()

    // 验证暗色主题下的按钮样式
    style, _ := ctx.Accessor().GetComputedStyle("btn")
    assert.Equal(t, darkTheme.Color("color.primary"), style.BgColor)

    // 切换到亮色主题
    ctx.SetTheme(lightTheme)
    ctx.Update()

    // 验证样式已更新
    style, _ = ctx.Accessor().GetComputedStyle("btn")
    assert.Equal(t, lightTheme.Color("color.primary"), style.BgColor)
}

func TestComponentVariant(t *testing.T) {
    ctx := goui.NewTestContext(goui.TestOptions{Width: 400, Height: 300})
    defer ctx.Destroy()

    ctx.Mount(Button(ID("btn"), Text("删除"), Variant("danger")))
    ctx.Update()

    style, _ := ctx.Accessor().GetComputedStyle("btn")
    assert.Equal(t, ctx.Theme().Color("color.danger"), style.BgColor)
}

func TestScopedTheme(t *testing.T) {
    ctx := goui.NewTestContext(goui.TestOptions{Width: 400, Height: 300})
    defer ctx.Destroy()

    ctx.SetTheme(lightTheme)
    ctx.Mount(Div(
        Button(ID("outer-btn"), Text("外部")),
        ThemeScope("dark",
            Button(ID("inner-btn"), Text("内部")),
        ),
    ))
    ctx.Update()

    outerStyle, _ := ctx.Accessor().GetComputedStyle("outer-btn")
    innerStyle, _ := ctx.Accessor().GetComputedStyle("inner-btn")
    // 外部按钮使用亮色主题，内部按钮使用暗色主题
    assert.NotEqual(t, outerStyle.BgColor, innerStyle.BgColor)
}
```

### 主题系统检查清单

实现主题相关功能时，必须确认：

1. **组件零硬编码** —— 组件不包含任何硬编码的颜色、字号、间距值，全部来自主题令牌
2. **令牌全覆盖** —— 每个组件用到的视觉属性都能追溯到一个设计令牌
3. **切换即生效** —— 调用 `SetTheme()` 后，所有组件在下一帧自动使用新主题，无需手动刷新
4. **变体可扩展** —— 用户可以通过 `RegisterVariant()` 为任意组件添加新变体，不修改组件源码
5. **区域可隔离** —— `ThemeScope` 内的组件使用独立主题，不影响外部，嵌套 Scope 正确继承
6. **缓存可失效** —— 主题切换时自动清除样式缓存，不出现视觉残留
7. **游戏友好** —— 支持完全自定义的视觉风格，不受"标准 UI"审美约束
8. **可序列化** —— 主题定义可从 JSON/TOML 等配置文件加载，支持热重载

## 关键设计决策

### 渲染命令缓冲

UI 层不直接调用图形 API，而是生成渲染命令列表（Command Buffer）。这带来几个好处：

1. 渲染后端可以对命令进行批处理优化（合并同纹理绘制调用）
2. 可以在单独的线程中执行渲染
3. 游戏引擎集成时可以将命令注入引擎的渲染管线

```go
type RenderCommand struct {
    Type    CommandType
    Clip    Rect
    ZOrder  int
    // union-like fields
    Rect    *RectCommand
    Text    *TextCommand
    Image   *ImageCommand
    Path    *PathCommand
}

type CommandBuffer struct {
    commands []RenderCommand
}
```

### 纹理图集（Texture Atlas）

所有小图片、图标、字形自动打包到纹理图集中，减少绘制调用切换：

- 字体字形 Atlas（SDF 渲染）
- 图标 Atlas
- 九宫格图片 Atlas

### 脏区域追踪

只重新布局和重绘发生变化的区域：

- 元素标记脏位（布局脏 / 绘制脏）
- 向上传播脏标记到祖先节点
- 布局阶段只重算脏子树
- 绘制阶段只重绘脏区域覆盖的元素

### 文本渲染

采用 SDF（Signed Distance Field）字体渲染：

- 分辨率无关，缩放不模糊
- GPU 友好，一个纹理适配多种字号
- 支持描边、阴影等效果
- 纯 Go 实现字形光栅化（生成 SDF 纹理）

## 线程模型

```
主线程 (UI 线程)          渲染线程 (可选)
┌──────────────┐        ┌──────────────┐
│ PollEvents() │        │              │
│      ↓       │        │              │
│ Event Dispatch│        │              │
│      ↓       │        │              │
│ State Update │        │              │
│      ↓       │        │              │
│ Layout Calc  │        │              │
│      ↓       │        │              │
│ Build CmdBuf ├───────→│ Execute Cmds │
│      ↓       │        │      ↓       │
│ (next frame) │        │ SwapBuffers  │
└──────────────┘        └──────────────┘
```

- UI 逻辑（事件处理、布局计算、命令生成）在主线程
- 渲染执行可以在单独线程（双缓冲命令列表）
- 游戏集成模式下，不创建渲染线程，由游戏主循环驱动
