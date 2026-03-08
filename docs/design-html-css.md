# HTML+CSS 模板驱动 UI 系统 — 设计文档

## 目标

用 HTML+CSS 定义 UI 结构和样式，Go 侧动态注入数据、绑定事件、增删元素。
使用者**不需要懂 Go widget 体系**，只需要会写 HTML+CSS 就能搭出完整 UI。
Go 代码只做两件事：**提供数据** 和 **响应事件**。

## 目标用法

### 最简用法 — 3 行 Go + 1 个 HTML 文件

```go
app := ui.NewApp("我的工具", 800, 600)
doc := app.LoadFile("ui/main.html")   // HTML 里写 <style>，不用单独 CSS 文件
doc.Set("version", "1.0.0")           // 数据注入
doc.On("#quit-btn", "click", func() { app.Close() })
app.Run()
```

```html
<!-- ui/main.html — 前端工程师/设计师直接可写 -->
<style>
  .container { display: flex; flex-direction: column; gap: 12px; padding: 16px; }
  .header    { font-size: 24px; color: #333; }
  .btn       { background: #1890ff; color: white; border-radius: 6px; padding: 8px 16px; }
  .btn:hover { background: #40a9ff; }
</style>
<div class="container">
  <span class="header">我的工具 v{{version}}</span>
  <div data-for="item in items">
    <span>{{item.name}}</span>
    <button class="btn" data-on-click="select" data-id="{{item.id}}">选择</button>
  </div>
  <button class="btn" id="quit-btn">退出</button>
</div>
```

### 完整 API 示例

```go
doc := ui.LoadDocument(tree, cfg, htmlString, cssString)

// 绑定事件
doc.On("#login-btn", "click", func() { doLogin() })
doc.On(".item", "click", func(e ui.DocEvent) { selectItem(e.Data("id")) })

// 注入数据 → 自动更新对应 DOM
doc.Set("username", "Luna")           // {{username}} → "Luna"
doc.Set("items", []Item{...})         // data-for="item in items" → 生成列表
doc.Set("showPanel", true)            // data-if="showPanel" → 显示/隐藏

// 动态追加/删除
doc.Append("items", newItem)
doc.RemoveAt("items", 2)

// 查询元素
btn := doc.Query("#submit")           // 返回 widget.Widget
allCards := doc.QueryAll(".card")     // 返回 []widget.Widget
```

## 设计原则：最低门槛

### 门槛梯度

| 层级 | 面向谁 | 需要掌握 | API |
|------|--------|---------|-----|
| L0 静态界面 | 设计师 / 前端 | HTML+CSS | `app.LoadFile("ui.html")` |
| L1 动态数据 | 应用开发者 | +模板语法 `{{}}` | `doc.Set(key, val)` |
| L2 事件交互 | 应用开发者 | +CSS 选择器 | `doc.On(sel, event, fn)` |
| L3 精细控制 | 高级开发者 | +Go widget API | `doc.Query(sel).(*widget.Button).SetVariant(...)` |
| L4 自定义组件 | 框架扩展者 | +widget.Base 体系 | 实现 Widget 接口，在 HTML 中用自定义标签 |

1. **入门**：写 HTML+CSS → `app.LoadFile()` → 能看到界面（零 Go widget 知识）
2. **交互**：`doc.On()` 绑事件 + `doc.Set()` 改数据 → 动态界面（只需知道选择器）
3. **进阶**：`doc.Query()` 拿 widget 引用 → 调 Go API 做精细控制（按需学习）

### 开发者友好

- HTML/CSS 语法报错带行号和上下文提示
- 未匹配的选择器打 warning 日志（帮查拼写错误）
- `data-*` 未绑定的字段打 warning（帮查数据遗漏）
- 开发模式下文件热重载（改完 HTML 立刻看效果，不用重编译）

---

## 实现规划

### 第一步：CSS 解析器 + 选择器匹配

#### CSS 词法/语法解析器（`css/` 包）

- 规则解析：`selector { property: value; }` → `[]Rule{Selector, []Declaration}`
- 支持 `<style>` 块和内联 `style=""` 属性（已有基础）
- 值解析：px/em/rem/%/auto/色值/简写展开（margin/padding/border/flex/background）

#### 选择器引擎

- 基础选择器：标签 `div`、类 `.class`、ID `#id`、通配 `*`
- 组合器：后代 `A B`、子元素 `A > B`、相邻兄弟 `A + B`、通用兄弟 `A ~ B`
- 伪类：`:hover`、`:focus`、`:active`、`:disabled`、`:first-child`、`:last-child`、`:nth-child(n)`
- 多选择器组：`h1, h2, h3 { ... }`

#### 级联与特异性（Specificity）

- 按 (inline, #id, .class, tag) 四元组计算优先级
- `!important` 支持
- 继承规则（color、font-* 等可继承属性向下传播）
- 计算属性缓存 + 脏标记失效

### 第二步：HTML 解析器增强

#### 完善 HTML 解析（增强现有 `html.go`）

- 新增标签：`<ul>/<ol>/<li>`、`<a>`、`<table>/<tr>/<td>/<th>`、`<form>/<label>/<select>/<option>/<textarea>`
- `<style>` 块提取 → CSS 解析器
- `<link rel="stylesheet">` 外部样式表加载（文件系统/嵌入）
- `id`/`class` 属性 → 元素树元数据（class 已有基础）
- 嵌套 `<div>` 递归解析（已有）

#### HTML → Widget 映射完善

- 语义化标签映射：`<ul>` → 垂直列表、`<table>` → Table widget、`<a>` → Link widget
- 表单标签映射：`<select>` → Select widget、`<textarea>` → TextArea widget
- CSS 属性 → layout.Style 全覆盖（现仅支持 display/flex-direction/gap/width/height/padding）
- 补充：margin、position、top/right/bottom/left、flex-grow/shrink/basis、align-*、justify-*、overflow、min/max-*、border、font-*、color、opacity、z-index

#### background 完整属性支持（高优先级）

游戏和软件个性化 UI 的核心依赖 — 皮肤、主题化、视觉风格几乎全靠 background 实现。

| 属性 | 说明 | 示例 |
|------|------|------|
| `background-color` | 纯色背景 | `#1a1a2e`、`rgba(0,0,0,0.8)` |
| `background-image: url()` | 图片背景 | `url(bg.png)`、`url(embed:assets/panel.png)` |
| `background-size` | 缩放模式 | `cover`、`contain`、`100px 200px`、`100% auto` |
| `background-position` | 定位 | `center`、`top right`、`10px 20px`、`50% 50%` |
| `background-repeat` | 平铺 | `no-repeat`、`repeat`、`repeat-x`、`repeat-y` |
| `background-clip` | 裁剪区域 | `border-box`、`padding-box`、`content-box` |
| `background-origin` | 定位参考 | `border-box`、`padding-box`、`content-box` |
| `background` | 简写 | `#333 url(bg.png) center/cover no-repeat` |

渐变支持：

| 语法 | 说明 | 典型用途 |
|------|------|---------|
| `linear-gradient(to bottom, #1a1a2e, #16213e)` | 线性渐变 | 面板背景、标题栏 |
| `radial-gradient(circle, #e94560, transparent)` | 径向渐变 | 技能光环、高亮效果 |
| `linear-gradient(45deg, #0f3460 0%, #533483 50%, #e94560 100%)` | 多色渐变 | 品质边框、稀有度效果 |

多背景叠加（游戏 UI 常见需求）：

```css
/* 游戏面板：底色 + 纹理 + 渐变边缘 */
.game-panel {
  background:
    linear-gradient(to bottom, rgba(0,0,0,0) 80%, rgba(0,0,0,0.8) 100%),
    url(panel-texture.png) center/cover no-repeat,
    #1a1a2e;
}

/* 技能按钮：冷却遮罩 + 图标 + 底色 */
.skill-slot {
  background:
    linear-gradient(to top, rgba(0,0,0,0.7) var(--cooldown), transparent var(--cooldown)),
    url(skill-icon.png) center/contain no-repeat,
    #2a2a3a;
}

/* 稀有度边框发光 */
.item-legendary {
  background: radial-gradient(ellipse at center, rgba(255,165,0,0.15), transparent 70%);
  border: 1px solid #ffa500;
}
```

实现要点：
- 渲染层新增 `DrawGradient` 命令（线性/径向）→ Vulkan fragment shader 实现
- 多背景按声明顺序从后往前叠加绘制（最后声明的在最底层）
- `background-size: cover/contain` 复用 Image widget 的 Fit 逻辑
- 图片背景走纹理系统（`render.TextureHandle`），支持 `embed.FS` 嵌入资源
- 渐变 color-stop 解析：位置（px/%/auto）+ 颜色，插值在 GPU 端完成

#### 透明度与半透明渲染（高优先级）

游戏 UI 大量依赖半透明效果（面板底色、遮罩、淡入淡出）。行为必须与 Chrome 一致，降低认知负担。

| 属性 | 说明 | Chrome 行为（需对齐） |
|------|------|----------------------|
| `opacity: 0-1` | 整个元素（含子树）透明度 | 创建合成层，子元素继承父透明度，不单独混合 |
| `background: rgba(r,g,b,a)` | 仅背景半透明 | 只影响背景色，子元素文字不受影响 |
| `color: rgba(r,g,b,a)` | 文字半透明 | 仅文字，不影响背景和子元素 |
| `visibility: hidden` | 隐藏但占位 | 保留布局空间，不绘制，子元素可 `visibility: visible` 覆盖 |
| `display: none` | 移除布局 | 完全不参与布局和绘制 |

关键区别 — `opacity` vs `rgba` alpha：

```css
/* opacity: 整个面板（含子文字）都变半透明 */
.panel { opacity: 0.7; }

/* rgba alpha: 只有背景半透明，文字保持不透明 — 游戏 HUD 最常用 */
.panel { background: rgba(0, 0, 0, 0.7); color: white; }
```

实现要点：
- `opacity` 需要离屏渲染（render-to-texture）：先将子树绘制到临时纹理，再以指定透明度合成到主帧缓冲
- `rgba` alpha 直接在顶点/片段着色器中处理，无需离屏（当前 SDF rect shader 已支持）
- 混合顺序：半透明元素必须从后往前绘制（painter's algorithm），当前 z-order 排序已满足
- `opacity: 0` 仍响应事件（与 Chrome 一致），`visibility: hidden` 和 `display: none` 不响应
- `opacity` 可与 `anim` 包的 Tween 联动实现淡入淡出过渡

#### 盒模型完善（与 Chrome 对齐）

目标：盒模型行为与 Chrome 完全一致，前端开发者零学习成本。

**box-sizing**（核心）：

```css
/* Chrome 默认 content-box，但现代 CSS 普遍用 border-box */
/* 建议：全局默认 border-box（更直觉），同时支持 content-box */
*, *::before, *::after { box-sizing: border-box; }
```

| box-sizing | width 包含 | Chrome 行为 |
|-----------|-----------|------------|
| `content-box` | 仅内容 | width = 内容宽度，padding/border 额外加 |
| `border-box` | 内容+padding+border | width = 总宽度，内容区自动缩减 |

**完整盒模型属性**：

| 属性 | 当前状态 | 需补充 |
|------|---------|--------|
| `margin` | 已支持（四向） | margin 合并（相邻块级垂直 margin collapse） |
| `padding` | 已支持（四向） | 简写解析：`padding: 10px 20px`（上下/左右） |
| `border` | 已支持（width/color） | `border-style`（solid/dashed/dotted/none）、单边 `border-top` 等 |
| `border-radius` | 已支持（统一值） | 四角独立：`border-radius: 10px 0 0 10px`、`/` 椭圆圆角 |
| `box-sizing` | 未实现 | content-box / border-box（默认 border-box） |
| `outline` | 未实现 | 不占布局空间的外边框（焦点可见性） |
| `box-shadow` | 未实现 | `offset-x offset-y blur spread color`、`inset`、多阴影 |

**margin collapse 规则**（与 Chrome 对齐）：

```
  Block A: margin-bottom: 20px
  Block B: margin-top: 30px
  → 实际间距: max(20, 30) = 30px（非 50px）

  例外（不 collapse）：
  - flex/grid 容器的子元素
  - 有 border/padding 隔开的父子
  - float / absolute / inline-block 元素
```

**简写属性解析**（必须与 Chrome 一致）：

```css
/* 1 值 → 四边 */
margin: 10px;                    /* top=right=bottom=left=10 */

/* 2 值 → 上下、左右 */
margin: 10px 20px;               /* top=bottom=10, left=right=20 */

/* 3 值 → 上、左右、下 */
margin: 10px 20px 30px;          /* top=10, left=right=20, bottom=30 */

/* 4 值 → 上、右、下、左（顺时针） */
margin: 10px 20px 30px 40px;     /* top=10, right=20, bottom=30, left=40 */

/* border 简写 */
border: 1px solid #ccc;          /* width style color */
border-left: 2px dashed red;     /* 单边覆盖 */
```

**box-shadow**（视觉表现力关键属性）：

```css
/* 基础阴影 */
.card { box-shadow: 0 2px 8px rgba(0,0,0,0.15); }

/* 多层阴影（深度感） */
.elevated {
  box-shadow:
    0 1px 3px rgba(0,0,0,0.12),
    0 4px 16px rgba(0,0,0,0.08);
}

/* 内阴影（凹陷效果 — 游戏输入框常用） */
.inset-panel { box-shadow: inset 0 2px 6px rgba(0,0,0,0.5); }

/* 发光效果（游戏选中/稀有度） */
.glow { box-shadow: 0 0 12px 4px rgba(0,150,255,0.6); }
```

实现要点：
- box-shadow 渲染：扩展 SDF rect shader，增加 shadow pass（高斯模糊近似）
- inset shadow：在背景之上、内容之下绘制
- 多阴影：按声明顺序叠加绘制
- box-sizing 在布局引擎 `layout.Compute()` 中处理：border-box 时 width 减去 padding+border 得到内容区
- margin collapse 在 block flow 布局中实现（flex/grid 跳过）

### 第三步：事件绑定系统

#### 选择器事件绑定 API

- `doc.On(selector, eventType, handler)` — 按 CSS 选择器绑定事件
- `doc.Off(selector, eventType)` — 解绑
- 事件委托：父节点监听，选择器匹配冒泡目标
- 内联事件属性：`<button data-on-click="handleLogin">` → Go 回调注册表

#### 事件上下文

- `DocEvent` 结构：Target（触发元素）、Data(key)（读取 data-* 属性）、StopPropagation/PreventDefault
- 表单事件：`change`、`input`、`submit`

### 第四步：数据绑定与模板渲染

#### 文本插值

- `{{expression}}` 语法解析（HTML 文本节点和属性值中）
- 点路径访问：`{{user.name}}`、`{{items[0].title}}`
- 管道过滤器（可选）：`{{price | currency}}`
- 绑定 `State[T]` / `ListState[T]` → 值变化自动重渲染对应节点

#### 条件渲染

- `data-if="condition"` — 真值时渲染，假值时移除子树
- `data-else` — 配合 data-if
- 条件变化 → 自动挂载/卸载对应 widget 子树

#### 列表渲染

- `data-for="item in items"` — 遍历数据生成重复子树
- key 追踪：`data-key="item.id"` — 复用已有节点，最小化 DOM diff
- `ListState.Append/RemoveAt/Set` → 增量更新（非全量重建）

#### 双向绑定

- `data-model="fieldName"` — Input/TextArea/Select/Checkbox 值与数据源双向同步
- 输入事件 → `doc.Set(field, newValue)` → 联动更新其他引用处

### 第五步：查询与动态操作 API

#### 元素查询

- `doc.Query(selector) widget.Widget` — 返回首个匹配
- `doc.QueryAll(selector) []widget.Widget` — 返回所有匹配
- `doc.GetByID(id) widget.Widget` — 快速 ID 查找（O(1) 索引）

#### 动态 DOM 操作

- `doc.AppendHTML(parentSelector, htmlFragment)` — 解析并追加子节点
- `doc.Remove(selector)` — 移除匹配元素
- `doc.SetAttribute(selector, key, value)` — 修改属性 → 重计算样式
- `doc.AddClass/RemoveClass/ToggleClass(selector, className)` — 类操作 → 触发 CSS 重匹配

#### 样式动态操作

- `doc.SetStyle(selector, property, value)` — 运行时修改内联样式
- 类名变化 → CSS 规则重匹配 → 布局/绘制脏标记

### 第六步：Document 生命周期与热更新

#### Document 对象

- `LoadDocument(tree, cfg, html, css) *Document` — 一次性解析 HTML+CSS → widget 树
- `doc.Reload(html, css)` — 热更新：diff 新旧树，最小化重建
- `doc.Dispose()` — 释放所有 widget、解绑事件、清理数据

#### 文件监听热重载（开发模式）

- 监听 .html/.css 文件变化 → 自动 `doc.Reload()`
- 保留组件状态（输入框文本、滚动位置、选中状态）

### 第七步：零样板 App 入口

- `ui.NewApp(title, w, h)` 一行启动窗口（封装 Tree + Config + Backend + Window）
- `app.LoadFile(path)` / `app.LoadHTML(html)` — 从文件或字符串加载
- `app.Embed(htmlFS)` — 支持 `embed.FS` 嵌入 HTML/CSS 资源（单二进制分发）
- `app.Run()` — 阻塞式主循环（内部处理事件、布局、渲染）
- `app.SetTheme("dark"/"light")` — 一行切换主题

---

## 技术架构

```
                    用户 Go 代码
                         |
                   ui.NewApp / LoadDocument
                         |
            +------------+------------+
            |                         |
     HTML 解析器                 CSS 解析器
     (html.go 增强)             (css/ 新包)
            |                         |
            v                         v
     Widget 树构建            选择器匹配 + 级联
     (标签→widget映射)        (特异性计算)
            |                         |
            +--------- 合并 ----------+
                         |
                   Document 对象
                   /    |    \
           数据绑定  事件绑定  查询API
           {{}}      On()    Query()
           data-for  Off()   QueryAll()
           data-if           AddClass()
           data-model        SetStyle()
```

## 与现有系统的关系

| 现有模块 | 关系 |
|---------|------|
| `html.go` (LoadHTML) | 增强为完整 HTML 解析器，保持向后兼容 |
| `builder.go` (Builder) | 并行存在，Builder 面向纯 Go 用户 |
| `bind.go` (State/ListState) | Document 内部使用 State 驱动数据绑定 |
| `layout/` (Style) | CSS 属性解析结果映射到 layout.Style |
| `core/` (Tree, Element) | Document 底层仍使用 core.Tree 管理元素 |
| `widget/` | HTML 标签映射到已有 widget，无需重新实现 |
| `theme/` | `app.SetTheme()` 桥接主题系统 |
| `event/` | `doc.On()` 底层使用 event 系统分发 |

## CSS 驱动主题系统

CSS 本身就是最自然的主题机制 — 换主题 = 换一份 CSS，零 Go 代码改动。

### 设计

```
themes/
├── light.css      # 浅色主题
├── dark.css       # 深色主题
├── game-rpg.css   # 游戏风格
└── custom.css     # 用户自定义
```

每份主题 CSS 只需定义 CSS 变量（自定义属性）：

```css
/* themes/light.css */
:root {
  --primary: #1890ff;
  --bg: #ffffff;
  --text: #333333;
  --border: #d9d9d9;
  --radius: 6px;
  --font-size: 14px;
}

/* themes/dark.css */
:root {
  --primary: #177ddc;
  --bg: #141414;
  --text: #ffffffd9;
  --border: #434343;
  --radius: 6px;
  --font-size: 14px;
}
```

业务 CSS 引用变量，自动跟随主题：

```css
.btn {
  background: var(--primary);
  color: var(--text);
  border-radius: var(--radius);
}
.card {
  background: var(--bg);
  border: 1px solid var(--border);
}
```

### API

```go
// 加载时指定主题
doc := app.LoadFile("ui/main.html", ui.WithTheme("themes/dark.css"))

// 运行时切换 — 重新应用 CSS 变量，触发样式重算
app.SetTheme("themes/light.css")

// 内置预设（不需要额外文件）
app.SetTheme(ui.ThemeLight)
app.SetTheme(ui.ThemeDark)

// 读取/修改单个变量
doc.SetVar("--primary", "#ff4d4f")  // 动态改主色
```

### 与现有 theme/ 包的关系

| 现有方式 | CSS 主题方式 | 关系 |
|---------|------------|------|
| `theme.Light/Dark` Go 结构体 | `light.css/dark.css` CSS 文件 | CSS 为上层，内部桥接到 Config |
| `widget.Config` 字段 | CSS 变量 `var(--xxx)` | CSS 变量解析后写入 Config 对应字段 |
| `theme.Apply(t, cfg)` Go 调用 | `app.SetTheme("x.css")` | CSS 方式对使用者更友好 |

内置的 `ui.ThemeLight` / `ui.ThemeDark` 预设实际上是嵌入的 CSS 文件（`embed.FS`），
与自定义 CSS 主题走同一条路径，无特殊逻辑。

### 实现要点

- CSS 变量解析：`:root { --name: value }` → 全局变量表
- `var(--name)` 在属性值解析时替换
- `var(--name, fallback)` 支持默认值
- `app.SetTheme()` → 重新解析 CSS → 更新变量表 → 标记全树样式脏 → 下帧重绘
- `doc.SetVar()` → 修改单个变量 → 仅脏标记引用该变量的节点

---

## 不做的事情

- 不实现完整的 Web 浏览器（不支持 JavaScript、不支持网络请求）
- 不追求 100% CSS 规范兼容（优先覆盖布局和视觉属性，忽略 print/animation/@media 等）
- 不做虚拟 DOM diff（直接操作 widget 树，用脏标记增量更新）
- 不实现 Shadow DOM / Web Components（用自定义标签注册替代）
