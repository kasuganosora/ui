# 输入与 IME

## 输入事件体系

### 事件类型

```go
type EventType int

const (
    // 鼠标
    EventMouseMove       EventType = iota
    EventMouseDown
    EventMouseUp
    EventMouseClick
    EventMouseDoubleClick
    EventMouseWheel
    EventMouseEnter
    EventMouseLeave

    // 键盘
    EventKeyDown
    EventKeyUp
    EventKeyPress          // 产生字符的按键

    // IME
    EventIMECompositionStart
    EventIMECompositionUpdate
    EventIMECompositionEnd
    EventIMECandidateOpen
    EventIMECandidateClose
    EventIMECandidateChange

    // 触摸
    EventTouchStart
    EventTouchMove
    EventTouchEnd
    EventTouchCancel

    // 手柄
    EventGamepadButtonDown
    EventGamepadButtonUp
    EventGamepadAxis

    // 拖放
    EventDragStart
    EventDragMove
    EventDragEnd
    EventDrop

    // 焦点
    EventFocus
    EventBlur

    // 窗口
    EventResize
    EventDPIChange
    EventClose
)
```

### 事件结构

```go
type Event struct {
    Type      EventType
    Target    Component      // 事件目标元素
    Timestamp time.Time
    Consumed  bool           // 是否已被消费

    // 鼠标相关
    MouseX, MouseY   float32  // 相对于窗口
    LocalX, LocalY   float32  // 相对于目标元素
    Button           MouseButton
    WheelDX, WheelDY float32

    // 键盘相关
    Key       Key
    Modifiers Modifiers      // Ctrl/Shift/Alt/Super
    Char      rune           // KeyPress 产生的字符

    // IME 相关
    IME       *IMEEvent

    // 触摸相关
    Touches   []Touch

    // 手柄相关
    Gamepad   *GamepadEvent
}

type Modifiers struct {
    Ctrl  bool
    Shift bool
    Alt   bool
    Super bool  // Win/Cmd
}
```

### 事件分发机制

采用 W3C 事件模型（简化版）：

```
捕获阶段（从根到目标）
    Root → Parent → ... → Target
目标阶段
    Target（执行 handler）
冒泡阶段（从目标到根）
    Target → ... → Parent → Root
```

```go
// 在冒泡阶段监听（默认）
button.On(goui.EventMouseClick, func(e *Event) {
    // 处理点击
})

// 在捕获阶段监听
parent.OnCapture(goui.EventMouseClick, func(e *Event) {
    // 在子元素之前拦截
})

// 阻止冒泡
button.On(goui.EventMouseClick, func(e *Event) {
    e.StopPropagation()
})
```

## IME 完整支持

IME 支持是本库的重点功能，确保中文、日文、韩文等输入法的完整体验。

### IME 事件流

```
用户开始输入 (按键触发 IME)
    ↓
IMECompositionStart          ← 开始组合
    ↓
IMECompositionUpdate × N     ← 组合文本变化（拼音/假名等）
    ↓ (可选)
IMECandidateOpen             ← 候选窗打开
IMECandidateChange × N       ← 候选列表更新
    ↓
IMECompositionEnd            ← 用户选择候选，提交文本
IMECandidateClose            ← 候选窗关闭
```

### IME 事件数据

```go
type IMEEvent struct {
    // 组合文本
    CompositionText  string     // 当前组合中的文本（如 "zhong"）
    CompositionStart int        // 组合文本在输入框中的起始位置
    CompositionEnd   int        // 组合文本在输入框中的结束位置

    // 候选
    Candidates     []string    // 候选列表（如 ["中", "钟", "终", ...]）
    CandidatePage  int         // 当前候选页
    CandidateTotal int         // 候选总页数
    SelectedIndex  int         // 当前高亮的候选索引

    // 提交
    CommittedText string       // 最终提交的文本
}
```

### 各平台 IME 实现

#### Windows (TSF / IMM32)

```go
// Windows IME 实现要点
type WindowsIME struct {
    hwnd      HWND
    immCtx    HIMC
    // TSF (Text Services Framework) for modern IME
    threadMgr ITfThreadMgr
    context   ITfContext
}

// 关键消息处理
// WM_IME_STARTCOMPOSITION  → 设置候选窗位置
// WM_IME_COMPOSITION       → 获取组合文本
// WM_IME_ENDCOMPOSITION    → 获取提交文本
// WM_IME_NOTIFY            → 候选窗状态变化
```

需要处理的关键点：
- 通过 `ImmSetCompositionWindow` 设置候选窗跟随光标
- 通过 `ImmGetCompositionString` 获取 GCS_COMPSTR / GCS_RESULTSTR
- 处理 TSF 用于支持现代输入法（如微软拼音、搜狗）
- 处理 DPI 缩放对候选窗位置的影响

#### Linux (IBus / Fcitx / XIM)

```go
// Linux 需要同时支持多种 IME 框架
type LinuxIME struct {
    // X11 + XIM
    xim XIM
    xic XIC

    // Wayland + text-input protocol
    textInput *WlTextInput

    // DBus 接口（IBus/Fcitx）
    dbusConn *dbus.Conn
}
```

需要处理的关键点：
- X11: 通过 XIM 协议与输入法通信
- Wayland: 使用 `zwp_text_input_v3` 协议
- 设置预编辑区域位置（preedit area）
- 处理 XFilterEvent 过滤 IME 事件

#### macOS (NSTextInputClient)

```go
// macOS IME 通过 NSTextInputClient 协议
// 需要 CGO 或 purego 调用 Objective-C
```

需要处理的关键点：
- 实现 `NSTextInputClient` 协议
- `insertText:replacementRange:` 处理提交
- `setMarkedText:selectedRange:replacementRange:` 处理组合
- `firstRectForCharacterRange:actualRange:` 返回光标位置

#### Android (InputConnection)

- 通过 JNI 实现 `InputConnection` 接口
- 处理 `composingText` 和 `commitText`

#### iOS (UITextInput)

- 实现 `UITextInput` 协议
- 处理 `markedTextRange` 和 `insertText`

### IME 候选窗定位

自绘候选窗（当系统候选窗不合适时）：

```go
type IMECandidateWindow struct {
    Visible    bool
    Position   Vec2         // 跟随输入光标位置
    Candidates []string
    Selected   int
    PageSize   int
    Page       int
}
```

定位策略：
1. 获取当前输入光标在屏幕上的位置
2. 候选窗默认在光标下方
3. 如果下方空间不足，移到上方
4. 如果右侧超出屏幕，左移
5. 考虑 DPI 缩放

## 焦点管理

### 焦点导航

```go
// Tab 键导航
type FocusManager struct {
    current   Component
    tabOrder  []Component  // 按 tabindex 排序
}

// Tab / Shift+Tab 切换焦点
// Enter / Space 激活当前焦点元素
// 方向键在特定组件内导航（如菜单、列表）
```

### 焦点可见性

- 键盘导航时显示焦点环（focus ring）
- 鼠标点击时不显示焦点环
- 游戏手柄导航时显示焦点高亮

### 游戏手柄导航

游戏场景下支持手柄导航 UI：

```go
type GamepadNavigation struct {
    // 方向键/摇杆 → 移动焦点到最近的可聚焦元素
    // A/Cross     → 确认/激活
    // B/Circle    → 返回/取消
    // LB/RB       → 切换标签页
    // Start       → 打开/关闭菜单
}
```

空间导航算法：
1. 从当前焦点元素出发
2. 根据方向（上下左右）筛选候选元素
3. 计算方向权重 + 距离权重
4. 选择最佳候选元素

## 快捷键系统

```go
// 注册全局快捷键
app.Shortcut(goui.Key{Ctrl: true, Key: KeyS}, func() {
    // 保存
})

// 组件级快捷键
input.Shortcut(goui.Key{Ctrl: true, Key: KeyA}, func() {
    input.SelectAll()
})

// 快捷键上下文（如编辑模式 vs 浏览模式）
ctx := app.ShortcutContext("editor")
ctx.Bind(goui.Key{Key: KeyDelete}, deleteSelected)
ctx.Enable()  // 启用此上下文
ctx.Disable() // 禁用此上下文
```
