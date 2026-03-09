# Feed Timeline Demo — Twitter 风格信息流

一个模仿 Twitter 时间线的 demo，用于测试 GoUI 绘制复杂 UI 的能力。每 10 秒自动插入一条新推文，全部使用 HTML+CSS 声明布局。

```
go run ./cmd/feed
```

![Feed Timeline](screenshot.png)

---

## 架构概览

```
┌──────────────────────────────────────────────┐
│                   App                         │
│  ┌──────────────────────────────────────────┐ │
│  │ HTML+CSS Parser → Widget Tree            │ │
│  │  <header>  → Header widget               │ │
│  │  <main>    → Content widget (scrollable) │ │
│  │   └ tweet  → tweetWidget (custom)        │ │
│  │   └ tweet  → tweetWidget                 │ │
│  │   └ ...                                  │ │
│  └──────────────────────────────────────────┘ │
│                                               │
│  OnLayout(tree, root, w, h)                   │
│    → 计算每个 widget 的 Bounds                 │
│    → Content.SetContentHeight() 驱动滚动       │
│                                               │
│  每 10s: addRandomTweet()                     │
│    → goroutine 创建 tweetWidget               │
│    → tree.MarkDirty() 触发重绘                 │
└──────────────────────────────────────────────┘
```

---

## 实现步骤详解

### 第一步：定义 HTML 结构

使用 GoUI 的 HTML 解析器声明页面结构。`<style>` 块内的 CSS 会被自动解析并应用到对应的元素：

```html
<div>
  <header id="feed-header">
    <span>首页</span>
  </header>
  <main id="feed-content">
    <!-- 推文会被动态插入到这里 -->
  </main>
</div>
```

**关键点：**
- `<div>` 作为根容器，`display: flex; flex-direction: column` 实现纵向排列
- `<header>` 固定高度 53px，模仿 Twitter 顶栏
- `<main>` 使用 `flex-grow: 1` 填充剩余空间，`overflow: scroll` 启用滚动
- GoUI 的 `<main>` 标签自动映射为 `Content` widget，内置滚动条

### 第二步：定义 CSS 样式

GoUI 支持标准 CSS 子集，包括 Flexbox 布局、颜色、字号、边距等：

```css
/* Twitter 深色主题 */
div { background: #15202B; }

/* 推文卡片：横向 flex，头像在左，内容在右 */
.tweet {
  display: flex;
  flex-direction: row;
  padding: 12px 16px;
  border-bottom: 1px solid #38444D;
  gap: 12px;
}

/* 操作栏：均匀分布 */
.tweet-actions {
  display: flex;
  flex-direction: row;
  justify-content: space-between;
}
```

**支持的 CSS 属性：**

| 类别 | 属性 |
|------|------|
| 布局 | `display`, `flex-direction`, `flex-wrap`, `flex-grow`, `flex-shrink`, `flex-basis` |
| 对齐 | `justify-content`, `align-items`, `align-self`, `gap` |
| 尺寸 | `width`, `height`, `min-*`, `max-*` (px, %, em, rem) |
| 间距 | `margin`, `padding` (四方向简写) |
| 定位 | `position`, `top`, `right`, `bottom`, `left`, `z-index` |
| 视觉 | `color`, `background`, `border-*`, `border-radius`, `box-shadow`, `opacity` |
| 文字 | `font-size` |
| 溢出 | `overflow: visible/hidden/scroll/auto` |

### 第三步：创建自定义 Widget

每条推文是一个自定义 `tweetWidget`，嵌入 `widget.Base` 并实现 `Draw()` 方法：

```go
type tweetWidget struct {
    widget.Base              // 嵌入基础组件（提供 ElementID, Bounds, Style 等）
    avatarColor uimath.Color // 头像颜色
    name        string       // 用户名
    handle      string       // @handle
    text        string       // 推文内容
    // ...
}
```

**Draw() 绘制流程：**

```
Draw(buf *render.CommandBuffer)
  │
  ├─ 1. DrawRect: 底部分割线 (1px, #38444D)
  │
  ├─ 2. DrawRect: 圆形头像 (CornersAll = radius/2)
  │     └ DrawText: 名字首字母（居中）
  │
  ├─ 3. DrawText: 用户名 + @handle + 时间
  │
  ├─ 4. DrawText: 推文正文
  │
  └─ 5. 操作栏 (MDI 图标 + 数字)
        ├ chat_bubble_outline + 回复数
        ├ repeat + 转发数 (绿色)
        ├ favorite_border + 点赞数 (粉色)
        └ share 分享图标
```

**核心 API：**
- `buf.DrawRect(RectCmd{...})` — 绘制矩形/圆角矩形
- `cfg.TextRenderer.DrawText(buf, text, x, y, fontSize, maxW, color, opacity)` — 绘制文字
- `cfg.DrawMDIcon(buf, "icon_name", x, y, size, color, z, opacity)` — 绘制 Material Design 图标

### 第四步：自定义布局函数

GoUI 的 `App` 支持通过 `SetOnLayout` 注入自定义布局逻辑。每一帧渲染前，布局函数会被调用：

```go
app.SetOnLayout(func(tree *core.Tree, root widget.Widget, w, h float32) {
    // 1. 设置根元素和 Header 的 Bounds
    // 2. 计算 Content 区域（总高度 - Header 高度）
    // 3. 遍历推文列表，依次排列（减去 scrollY 偏移）
    // 4. 设置 Content.ContentHeight 驱动滚动条
})
```

**布局计算核心：**

```
Window (600 x 800)
├── Header: (0, 0, 600, 53)           ← 固定高度
└── Content: (0, 53, 600, 747)        ← 填充剩余
    ├── Tweet 0: (0, 53-scrollY, 600, 120)
    ├── Tweet 1: (0, 173-scrollY, 600, 120)
    ├── Tweet 2: (0, 293-scrollY, 600, 120)
    └── ...
    ContentHeight = N * 120            ← 驱动滚动范围
```

`tree.SetLayout(elementID, LayoutResult{Bounds: rect})` 将计算结果写入元素树，渲染时 `widget.Bounds()` 读取这些值。

### 第五步：动态插入推文

使用 goroutine 每 10 秒创建新推文并插入：

```go
go func() {
    for {
        time.Sleep(10 * time.Second)
        addRandomTweet(tree, cfg, container) // 创建 widget + AppendChild
        tree.MarkDirty(tree.Root())          // 触发重绘
    }
}()
```

**关键点：**
- `container.AppendChild(tweet)` — 将 widget 添加到父容器的子列表
- `tree.AppendChild(parentID, childID)` — 在元素树中建立父子关系
- `tree.MarkDirty(root)` — 标记脏节点，主循环下一帧会重新布局和渲染
- GoUI 主循环检测 `tree.NeedsRender()` 来决定是否重绘

### 第六步：滚动机制

`Content` widget 内置滚动支持：

```
用户滚动鼠标滚轮
    ↓
App.handleEvent() 捕获 MouseWheel
    ↓
content.HandleWheel(dy)  →  scrollY += dy * 40
    ↓
下一帧 layoutFeed() 被调用
    ↓
每条推文的 Y = contentY - scrollY + offset
    ↓
Content.Draw() 调用 buf.PushClip(bounds) 裁剪溢出内容
    ↓
绘制可见推文 + 滚动条
```

---

## 渲染管线

```
每帧 (≈1ms @1000fps)
  │
  ├─ PollEvents()          ← Win32 消息
  ├─ handleEvent()         ← 路由到 widget
  │
  ├─ OnLayout(tree, root, w, h)   ← 计算所有 Bounds
  │
  ├─ BeginFrame()          ← GPU 开始
  ├─ root.Draw(buf)        ← 递归绘制所有 widget
  │   ├─ Header.Draw()
  │   └─ Content.Draw()
  │       ├─ PushClip()    ← 裁剪区域
  │       ├─ tweet[0].Draw()
  │       ├─ tweet[1].Draw()
  │       └─ PopClip()
  │
  ├─ TextRenderer.Upload() ← 字形纹理上传 GPU
  ├─ Backend.Submit(buf)   ← 提交渲染命令
  └─ EndFrame()            ← 呈现到屏幕
```

---

## 使用的 GoUI 特性

| 特性 | 用途 |
|------|------|
| HTML+CSS 解析 | 声明式 UI 结构 |
| Flexbox 布局 | Header/Content 纵向排列 |
| Content 滚动 | 推文列表无限滚动 |
| 自定义 Widget | tweetWidget 自由绘制 |
| Material Design Icons | 回复、转发、点赞、分享图标 |
| TextRenderer | 多语言文字渲染（中/英/日） |
| 动态内容 | goroutine + MarkDirty 插入新推文 |
| GPU 加速 | Vulkan/DX11/DX9/OpenGL 自动选择 |

---

## 扩展方向

- **图片推文**: 使用 `Img` widget 加载网络图片或本地图片
- **点赞动画**: 点击心形图标时播放缩放动画
- **引用推文**: 嵌套 tweetWidget 实现 Quote Tweet
- **主题切换**: 利用 CSS Variables 切换亮色/暗色主题
- **虚拟滚动**: 对超长列表使用 VirtualList 只渲染可见项
