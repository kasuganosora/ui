# 布局系统

## 概述

布局系统是 GoUI 的核心，提供两种使用方式：

1. **HTML+CSS 声明式** - 用 HTML 结构 + CSS 子集描述布局
2. **Go API 命令式** - 用 Go 代码直接构建元素树和设置样式

两种方式最终都生成同一个元素树，使用同一套布局算法。

## HTML+CSS 子集

### 支持的 HTML 标签

不追求完整 HTML 规范，只支持布局语义相关的标签：

| 标签 | 说明 |
|------|------|
| `<div>` | 通用块容器 |
| `<span>` | 通用行内容器 |
| `<text>` | 文本节点 |
| `<img>` | 图片 |
| `<input>` | 输入框 (text/password/number) |
| `<button>` | 按钮 |
| `<select>` | 下拉选择 |
| `<textarea>` | 多行文本 |
| `<scroll>` | 滚动容器 |
| `<list>` | 虚拟列表 |
| `<window>` | 子窗口 |
| `<canvas>` | 自定义绘制区域 |
| `<template>` | 模板定义（不渲染） |
| `<slot>` | 插槽 |
| `<component>` | 自定义组件引用 |

### 支持的 CSS 属性

只实现布局、盒模型、视觉相关属性：

#### 布局

```css
display: flex | grid | block | inline | none;
position: relative | absolute | fixed | sticky;

/* Flexbox */
flex-direction: row | column | row-reverse | column-reverse;
flex-wrap: nowrap | wrap;
justify-content: flex-start | flex-end | center | space-between | space-around | space-evenly;
align-items: flex-start | flex-end | center | stretch | baseline;
align-self: auto | flex-start | flex-end | center | stretch;
flex-grow: <number>;
flex-shrink: <number>;
flex-basis: <length> | auto;
gap: <length>;
row-gap: <length>;
column-gap: <length>;

/* Grid */
grid-template-columns: <track-list>;
grid-template-rows: <track-list>;
grid-column: <line> / <line>;
grid-row: <line> / <line>;
grid-gap: <length>;

/* 定位 */
top: <length>;
right: <length>;
bottom: <length>;
left: <length>;
z-index: <integer>;
```

#### 盒模型

```css
width: <length> | <percentage> | auto;
height: <length> | <percentage> | auto;
min-width: <length> | <percentage>;
max-width: <length> | <percentage> | none;
min-height: <length> | <percentage>;
max-height: <length> | <percentage> | none;

margin: <length> | <percentage> | auto;
margin-top/right/bottom/left: ...;

padding: <length> | <percentage>;
padding-top/right/bottom/left: ...;

box-sizing: content-box | border-box;
overflow: visible | hidden | scroll | auto;
overflow-x / overflow-y: ...;
```

#### 视觉

```css
background-color: <color>;
background-image: url(...);
background-size: cover | contain | <length>;
background-position: ...;

border: <width> <style> <color>;
border-radius: <length>;
border-top/right/bottom/left: ...;

color: <color>;
font-size: <length>;
font-family: <string>;
font-weight: normal | bold | <number>;
line-height: <number> | <length>;
text-align: left | center | right;
text-overflow: clip | ellipsis;
white-space: normal | nowrap | pre | pre-wrap;
word-break: normal | break-all | break-word;

opacity: <number>;
visibility: visible | hidden;
cursor: default | pointer | text | grab | ...;

/* 变换与过渡 */
transform: translate() | scale() | rotate();
transition: <property> <duration> <easing>;
```

#### 长度单位

```
px    - 像素（默认）
%     - 父元素百分比
em    - 相对当前字体大小
rem   - 相对根字体大小
vw/vh - 视口百分比
fr    - Grid 比例单位
auto  - 自动计算
```

#### 颜色格式

```
#RGB / #RRGGBB / #RRGGBBAA
rgb(r, g, b) / rgba(r, g, b, a)
hsl(h, s%, l%) / hsla(h, s%, l%, a)
命名颜色 (red, blue, transparent, ...)
```

## 布局算法

### Flexbox 布局

实现 CSS Flexbox 规范的核心子集：

1. 确定主轴方向（row/column）
2. 计算子元素的 flex-basis
3. 分配剩余空间（flex-grow）或收缩（flex-shrink）
4. 处理换行（flex-wrap）
5. 对齐（justify-content, align-items）

### Grid 布局

实现 CSS Grid 的常用功能：

1. 解析 grid-template 定义轨道
2. 放置显式定位的项目
3. 自动放置未指定位置的项目
4. 计算轨道尺寸（fr 单位分配）
5. 对齐网格项目

### Flow 布局

简单的块级/行内流式布局，作为默认布局模式。

### Absolute 布局

脱离文档流，根据 top/right/bottom/left 相对于最近定位祖先定位。

## 自适应布局系统

### Grid 栅格

提供 12/24 列栅格系统：

```go
goui.Row(
    goui.Col(goui.Span(8), /* 主内容 */),
    goui.Col(goui.Span(4), /* 侧边栏 */),
)
```

或 HTML：

```html
<div class="row">
    <div class="col" span="8">主内容</div>
    <div class="col" span="4">侧边栏</div>
</div>
```

### 响应式断点

```go
type Breakpoint struct {
    XS  int  // < 576px   手机竖屏
    SM  int  // >= 576px  手机横屏
    MD  int  // >= 768px  平板
    LG  int  // >= 992px  小屏桌面
    XL  int  // >= 1200px 大屏桌面
    XXL int  // >= 1600px 超宽屏
}
```

栅格列支持按断点指定不同跨度：

```html
<div class="col" xs="12" md="8" xl="6">自适应列</div>
```

### Layout 布局容器

预定义常用布局模式：

```go
// 经典 Header-Content-Footer
goui.Layout(goui.LayoutHCF,
    goui.Header(/* ... */),
    goui.Content(/* ... */),
    goui.Footer(/* ... */),
)

// 侧边栏布局
goui.Layout(goui.LayoutSidebar,
    goui.Aside(goui.Width(200), /* ... */),
    goui.Main(/* ... */),
)
```

### Space 间距

统一的间距管理：

```go
goui.Space(goui.DirectionVertical, goui.Gap(8),
    child1,
    child2,
    child3,
)
```

等价于给子元素之间添加统一间距，避免手动管理 margin。

### 子窗口（SubWindow）

支持在 UI 中嵌套独立的子窗口区域：

```go
goui.SubWindow(goui.SubWindowOptions{
    Title:     "属性面板",
    Draggable: true,
    Resizable: true,
    X: 100, Y: 100,
    Width: 300, Height: 400,
},
    /* 子窗口内容 */
)
```

特性：
- 独立的滚动区域
- 可拖拽、可缩放
- 可最小化/最大化/关闭
- 独立的 Z 序管理
- 支持停靠（Docking）到主布局的边缘

## CSS 解析器

### 解析流程

```
CSS 文本 → Tokenizer → Parser → StyleSheet
                                    ↓
HTML 文本 → Tokenizer → Parser → Element Tree
                                    ↓
                          Style Resolution
                          (选择器匹配 + 层叠 + 继承)
                                    ↓
                          Computed Style Tree
                                    ↓
                            Layout Algorithm
                                    ↓
                             Layout Tree
                          (每个节点带位置和尺寸)
```

### 选择器支持

```css
/* 标签选择器 */
div { }

/* 类选择器 */
.classname { }

/* ID 选择器 */
#idname { }

/* 后代选择器 */
.parent .child { }

/* 子选择器 */
.parent > .child { }

/* 伪类 */
:hover { }
:active { }
:focus { }
:disabled { }
:first-child { }
:last-child { }
:nth-child(n) { }

/* 组合 */
div.class#id:hover { }
```

### 样式优先级

遵循 CSS 优先级规则（简化版）：

1. `!important` 声明
2. 内联样式 (style 属性)
3. ID 选择器数量
4. 类/伪类选择器数量
5. 标签选择器数量
6. 源码顺序（后出现的优先）
