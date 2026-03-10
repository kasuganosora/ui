# 用 GoUI 写一个 Twitter 风格的 Feed 应用

本文以 `cmd/feed` 为例，带你从零了解如何用 GoUI 库构建一个真实的桌面 App。
这个 Demo 实现了类 Twitter 的时间线界面：标签栏、发帖框、可滚动推文列表、实时新推文推送。

---

## 目录

1. [GoUI 的核心模型](#1-goui-的核心模型)
2. [第一步：创建窗口](#2-第一步创建窗口)
3. [第二步：用 HTML+CSS 描述界面](#3-第二步用-htmlcss-描述界面)
4. [第三步：查询元素，绑定事件](#4-第三步查询元素绑定事件)
5. [第四步：动态添加内容](#5-第四步动态添加内容)
6. [第五步：SetOnLayout — 每帧的核心钩子](#6-第五步setonlayout--每帧的核心钩子)
7. [第六步：后台 goroutine 与线程安全](#7-第六步后台-goroutine-与线程安全)
8. [布局引擎基础](#8-布局引擎基础)
9. [常用 CSS 属性速查](#9-常用-css-属性速查)
10. [常见坑与避坑指南](#10-常见坑与避坑指南)

---

## 1. GoUI 的核心模型

GoUI 采用 **「HTML 描述界面 + Go 控制逻辑」** 的模式，整体分三层：

```
HTML + CSS 字符串
        ↓  app.LoadHTML()
   Widget 树（核心元素树）
        ↓  app.SetOnLayout() 每帧触发
   CSSLayout 计算尺寸/位置 → 渲染到屏幕
```

- **Widget**：界面元素（Div、Text、Button、TextArea…）
- **Tree**：管理所有 Widget 的树，负责事件分发和脏标记
- **CSSLayout**：每帧将 CSS 样式翻译成屏幕坐标，再送给渲染器（Vulkan/DX11）

---

## 2. 第一步：创建窗口

```go
app, err := ui.NewApp(ui.AppOptions{
    Title:  "我的应用",
    Width:  600,
    Height: 860,
})
if err != nil {
    log.Fatal(err)
}
defer app.Destroy()
```

`NewApp` 创建操作系统窗口并初始化渲染后端（Windows 上使用 DX11 或 Vulkan）。

### 主题色配置

GoUI 有一套全局 Config，控制 Button、Input 等控件的默认颜色。
**必须在 `LoadHTML` 之前设置**，因为 Widget 构造时会读取当前值：

```go
cfg := app.Config()
cfg.BgColor        = uimath.ColorHex("#16181C")   // 全局背景
cfg.TextColor      = uimath.ColorHex("#E7E9EA")   // 默认文字色
cfg.PrimaryColor   = uimath.ColorHex("#1D9BF0")   // 主题蓝（按钮等）
cfg.BorderColor    = uimath.ColorHex("#2F3336")   // 边框
cfg.DisabledColor  = uimath.ColorHex("#536471")   // placeholder
```

---

## 3. 第二步：用 HTML+CSS 描述界面

### 3.1 写 HTML 结构

GoUI 解析一个 HTML 字符串，把它变成 Widget 树。支持常见的 HTML 标签：

| 标签 | 对应 Widget | 说明 |
|------|-------------|------|
| `<div>` | Div | 万能容器，支持 flex 布局 |
| `<span>` | Text | 行内文本 |
| `<button>` | Button | 可点击按钮 |
| `<textarea>` | TextArea | 多行输入框 |
| `<input>` | Input | 单行输入框 |
| `<main>` | Content | 可滚动内容区 |
| `<header>` | Header | 页头容器 |

```go
const feedHTML = `
<div class="feed-root">
  <header class="nav-bar">...</header>
  <div class="tab-bar">...</div>
  <div class="compose">
    <div class="compose-avatar"><span class="avatar-initial">我</span></div>
    <textarea id="compose-input" placeholder="有什么新鲜事？"></textarea>
    <button id="compose-post" theme="primary" shape="round">发帖</button>
  </div>
  <main id="timeline" class="feed-timeline"></main>
</div>
`
```

关键点：
- 给需要操作的元素加 **`id`** 属性，方便后续查询
- `<main>` 会自动变成可滚动区域
- `<button>` 支持 `theme="primary"` 和 `shape="round"` 属性

### 3.2 写 CSS 样式

CSS 以字符串形式传入，支持 flexbox 布局、颜色、字体大小等常用属性：

```go
const feedCSS = `
.feed-root {
  display: flex;
  flex-direction: column;
  width: 100%;
  height: 100%;
  background: #000000;
}

.compose {
  display: flex;
  flex-direction: row;
  padding: 12px 16px;
  gap: 12px;
  border-bottom: 1px solid #2F3336;
}

.compose-avatar {
  width: 40px;
  height: 40px;
  border-radius: 20px;         /* 变成圆形 */
  background: #1D9BF0;
  display: flex;
  align-items: center;
  justify-content: center;     /* 内容居中 */
}
`
```

### 3.3 加载 HTML

```go
// 方式一：HTML 里内嵌 <style> 块
doc := app.LoadHTML(feedHTML + "<style>" + feedCSS + "</style>")

// 方式二：HTML 和 CSS 分开传
root := ui.LoadHTMLWithCSS(tree, cfg, htmlStr, cssStr)
```

`doc` 是一个 `Document` 对象，可以用 ID 查询元素。

---

## 4. 第三步：查询元素，绑定事件

### 4.1 按 ID 查询

```go
tree := app.Tree()

// QueryByID 返回 widget.Widget 接口
postBtn := doc.QueryByID("compose-post")
```

### 4.2 类型断言获取具体控件

Widget 是接口，需要断言成具体类型才能调用特有方法：

```go
// Button 的 OnClick
if btn, ok := postBtn.(*widget.Button); ok {
    btn.OnClick(func() {
        fmt.Println("发帖按钮被点击！")
    })
}

// TextArea 的读写
if inp := doc.QueryByID("compose-input"); inp != nil {
    if ta, ok := inp.(*widget.TextArea); ok {
        text := ta.Value()     // 读取内容
        ta.SetValue("")        // 清空
        ta.SetRows(2)          // 设置行数
        ta.SetAutosizeRows(2, 8)  // 自动增高，最多 8 行
    }
}
```

### 4.3 用 tree.AddHandler 绑定事件

对于 Div、Span 等普通元素，用 `tree.AddHandler` 绑定点击等事件：

```go
// 点击 Tab
tabWidget := doc.QueryByID("tab-0")
tree.AddHandler(tabWidget.ElementID(), event.MouseClick, func(e *event.Event) {
    fmt.Println("切换到第一个 Tab")
    // 修改样式
    tree.SetVisible(indicator.ElementID(), true)
    tree.MarkDirty(tree.Root())   // 触发重绘
})
```

### 4.4 动态修改 Widget

```go
// 修改文字颜色
if txt, ok := labelWidget.(*widget.Text); ok {
    txt.SetColor(uimath.ColorHex("#E7E9EA"))
}

// 显示/隐藏元素
tree.SetVisible(elementID, false)

// 标记需要重绘
tree.MarkDirty(tree.Root())
```

---

## 5. 第四步：动态添加内容

Feed Demo 最核心的能力：**运行时动态创建 Widget 并插入树**。

### 5.1 用 HTML 创建一张推文卡片

```go
func makeTweet(td *tweetData) widget.Widget {
    // 1. 拼接这张卡片的 HTML（含数据）
    html := fmt.Sprintf(`
<div class="tweet">
  <div class="tweet-inner">
    <div class="tweet-avatar" style="background:%s;">
      <span class="avatar-initial">%s</span>
    </div>
    <div class="tweet-body">
      <span class="tweet-name">%s</span>
      <span class="tweet-text">%s</span>
    </div>
  </div>
</div>`, avatarColor(td.Name), firstRune(td.Name), td.Name, td.Text)

    // 2. 用共享 CSS 解析成 Widget 树
    root := ui.LoadHTMLWithCSS(tree, cfg, html, feedCSS)

    // 3. 取出实际卡片（跳过包裹层）
    if len(root.Children()) > 0 {
        return root.Children()[0]
    }
    return root
}
```

### 5.2 追加到列表

```go
// timeline 是 <main id="timeline"> 对应的 Content Widget
type feedContainer interface {
    widget.Widget
    AppendChild(widget.Widget)
    PrependChild(widget.Widget)
}
container := timeline.(feedContainer)

// 追加到末尾
container.AppendChild(makeTweet(&allTweets[i]))

// 插入到顶部
container.PrependChild(makeTweet(newTweet))
```

---

## 6. 第五步：SetOnLayout — 每帧的核心钩子

GoUI 的 UI 更新不是自动的，需要在 `SetOnLayout` 里**主动调用** `CSSLayout`：

```go
app.SetOnLayout(func(tree *core.Tree, root widget.Widget, w, h float32) {
    // 1. 先处理待插入的新元素（本帧要显示的）
    for {
        select {
        case td := <-pendingTweetData:
            tw := makeTweet(td)
            container.PrependChild(tw)
        default:
            goto done
        }
    }
done:

    // 2. 执行 CSS 布局（计算所有元素的位置和大小）
    ui.CSSLayout(tree, root, w, h, cfg)

    // 3. 布局完成后可以读取元素坐标（比如 TextArea 自动调高）
    if composeTA != nil {
        composeTA.UpdateAutosizeHeight()
    }

    // 4. 检测滚动到底部，加载更多
    scrollY := contentWidget.ScrollY()
    contentH := contentWidget.ContentHeight()
    bounds := contentWidget.Bounds()
    if maxScroll := contentH - bounds.Height; maxScroll > 0 && scrollY >= maxScroll-120 {
        loadMoreTweets()
    }
})
```

**要点：**
- `SetOnLayout` 每帧都会被调用（窗口刷新时）
- 必须调用 `CSSLayout` 才会更新布局
- 元素的 `Bounds()` 只有在 `CSSLayout` 之后才有正确值
- `tree.MarkDirty(tree.Root())` 强制触发一次重绘

---

## 7. 第六步：后台 goroutine 与线程安全

**重要规则：Widget 只能在主线程操作。**

错误做法（会 crash）：
```go
go func() {
    time.Sleep(10 * time.Second)
    tw := makeTweet(newData)      // ❌ 在 goroutine 里操作 Widget 树
    container.PrependChild(tw)
}()
```

正确做法：goroutine 只发送数据，Widget 创建放在主线程的 `SetOnLayout` 里：

```go
var pendingTweetData = make(chan *tweetData, 32)

// goroutine：只发数据
go func() {
    for {
        time.Sleep(10 * time.Second)
        pendingTweetData <- &allTweets[rand.Intn(len(allTweets))]
    }
}()

// 主线程：在 SetOnLayout 里消费数据、创建 Widget
app.SetOnLayout(func(tree *core.Tree, root widget.Widget, w, h float32) {
    for {
        select {
        case td := <-pendingTweetData:
            tw := makeTweet(td)          // ✅ 主线程里安全创建
            container.PrependChild(tw)
        default:
            goto done
        }
    }
done:
    ui.CSSLayout(tree, root, w, h, cfg)
})
```

---

## 8. 布局引擎基础

GoUI 使用 **CSS Flexbox** 作为主要布局方式，行为和浏览器基本一致。

### 常用 flex 布局

```css
/* 横向排列（默认），内容垂直居中 */
.row {
  display: flex;
  flex-direction: row;
  align-items: center;
  gap: 8px;
}

/* 纵向排列 */
.column {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

/* 占满剩余空间 */
.grow {
  flex-grow: 1;
}

/* 不缩小（防止被压扁） */
.no-shrink {
  flex-shrink: 0;
}
```

### 滚动区域

```css
/* 超出内容可滚动 */
.list {
  overflow: scroll;   /* 或 auto */
  height: 100%;
}

/* 超出裁剪（不显示滚动条） */
.clip {
  overflow: hidden;
}
```

### `border-radius` 画圆形

```css
.avatar {
  width: 40px;
  height: 40px;
  border-radius: 20px;  /* 半径 = 宽高的一半 → 正圆 */
  background: #1D9BF0;
}
```

### `margin` 外边距

```css
/* 四个方向 */
.item { margin: 5px; }
/* 上下 5px，左右 0 */
.divider { margin: 5px 0; }
/* 分开写 */
.box { margin-top: 8px; margin-bottom: 4px; }
```

---

## 9. 常用 CSS 属性速查

| 属性 | 示例 | 说明 |
|------|------|------|
| `display` | `flex` | 启用 flex 布局 |
| `flex-direction` | `row` / `column` | 主轴方向 |
| `align-items` | `center` | 交叉轴对齐 |
| `justify-content` | `space-between` | 主轴分布 |
| `gap` | `12px` | 子元素间距 |
| `flex-grow` | `1` | 拉伸占满剩余空间 |
| `flex-shrink` | `0` | 禁止压缩 |
| `width` / `height` | `40px` / `100%` | 尺寸 |
| `min-width` | `72px` | 最小宽度 |
| `padding` | `12px 16px` | 内边距（上下 12，左右 16）|
| `margin` | `5px 0` | 外边距（上下 5，左右 0）|
| `background` | `#1D9BF0` | 背景色 |
| `color` | `#E7E9EA` | 文字颜色 |
| `font-size` | `15px` | 字体大小 |
| `border-radius` | `20px` | 圆角 |
| `border-bottom` | `1px solid #2F3336` | 下边框 |
| `overflow` | `hidden` / `scroll` | 溢出处理 |
| `white-space` | `nowrap` | 禁止换行 |

---

## 10. 常见坑与避坑指南

### ❶ Config 要在 LoadHTML 之前设置

Button 等控件在构造时读取 Config 颜色。如果先 `LoadHTML` 再改 Config，颜色不会生效。

```go
// ✅ 正确
cfg := app.Config()
cfg.PrimaryColor = uimath.ColorHex("#1D9BF0")
doc := app.LoadHTML(html)

// ❌ 错误
doc := app.LoadHTML(html)
app.Config().PrimaryColor = ...   // 太晚了，Button 已经用默认色构造好了
```

### ❷ Button 没有固有内容宽度

Button 不会自动根据文字内容撑开宽度，在 flex 布局里默认可能很小。
加 `min-width` 保证可见：

```css
.compose-btn {
  min-width: 72px;
  padding: 8px 20px;
}
```

### ❸ Widget 只能在主线程创建

见第 7 节。goroutine 里只能发数据，不能操作 Widget 树。

### ❹ Bounds 只在 CSSLayout 之后有效

如果需要读取元素位置（比如弹出菜单跟随按钮），要在 `SetOnLayout` 里、`CSSLayout` 调用之后读取：

```go
app.SetOnLayout(func(tree *core.Tree, root widget.Widget, w, h float32) {
    ui.CSSLayout(tree, root, w, h, cfg)
    // ✅ 现在读 Bounds 才是最新的
    bounds := someWidget.Bounds()
})
```

### ❺ 动态插入 Widget 后要 MarkDirty

`AppendChild` / `PrependChild` 不会自动触发重绘，需要调用：

```go
container.AppendChild(newCard)
tree.MarkDirty(tree.Root())
```

（在 `SetOnLayout` 里处理的话，下一帧自然会重排，可以省略）

### ❻ overflow:hidden 的高度要足够

如果父容器用了 `overflow: hidden`，但高度由子内容撑开（auto），要确保布局引擎能正确计算出足够的高度（含子元素 `margin`）。GoUI 布局引擎已处理此情况。

---

## 完整运行

```bash
go run ./cmd/feed
```

效果：一个 600×860 的 Twitter 暗色风格 Feed，支持标签切换、发帖、无限滚动加载、每 10 秒自动推送新推文。

---

## 项目结构参考

```
cmd/feed/
├── main.go          # 所有逻辑，约 780 行
│   ├── feedCSS      # 完整 CSS 样式表（常量字符串）
│   ├── feedHTML     # 静态页面骨架 HTML（常量字符串）
│   ├── tweetHTML()  # 动态生成单条推文 HTML
│   ├── makeTweet()  # HTML → Widget 树的工厂函数
│   └── main()       # 初始化、事件绑定、渲染循环
├── tweets.json      # 嵌入的示例推文数据（go:embed）
└── docs/
    └── tutorial.md  # 本文档
```
