# 开发路线图

## 阶段 0: 基础设施（第 1-2 周） ✅

- [x] 项目结构搭建、CI/CD
- [x] 基础数学库（Vec2, Rect, Color, Mat3, Edges, Corners）
- [x] 平台层接口定义（platform.Platform, platform.Window）
- [x] 渲染后端接口定义（render.Backend, render.CommandBuffer）
- [x] 事件系统基础类型（event.Event, event.Type, event.Key, event.MouseButton）
- [x] 单元测试覆盖率 90%+（math 99.4%, event 100%, render 92.4%）

## 阶段 1: 最小可用（第 3-8 周） ✅

### 平台层 - Windows ✅

- [x] Win32 窗口创建/管理（零 CGO，syscall.NewLazyDLL）
- [x] 键盘/鼠标输入采集（WndProc 全消息翻译）
- [x] DPI 感知（Per-Monitor DPI Aware）
- [x] 基础剪贴板（文本，Unicode + CJK + Emoji）
- [x] IME 支持（IMM32 组合文本、候选窗口定位）
- [x] 单元测试覆盖率 90.7%

### 渲染 - Vulkan ✅

- [x] Vulkan 动态加载（vulkan-1.dll, vkGetInstanceProcAddr 链）
- [x] Vulkan 初始化（实例、设备、交换链）
- [x] 矩形绘制管线（SDF 圆角、边框）— 管线结构已创建
- [x] SPIR-V 着色器编译（.glsl → .spv, go:embed 嵌入）
- [x] SDF 文本渲染管线（text.frag SDF 着色器 + 纹理描述符集）
- [x] 图片/纹理绘制管线（textured.vert/frag 着色器 + 管线）
- [x] 纹理系统（CreateTexture/UpdateTexture GPU 资源、staging buffer、布局转换）
- [x] 描述符集基础设施（布局、池、分配、combined image sampler）
- [x] 裁剪（Scissor）— CmdClip 支持
- [x] 渲染命令稳定排序（sort.SliceStable，修复光标闪烁导致 UI 元素消失）

### 截图与验收测试基础设施 ✅

- [x] Backend.ReadPixels() 接口 — 从 GPU 帧缓冲区读回 RGBA 像素
- [x] Vulkan ReadPixels 实现（swapchain image → staging buffer → host 内存映射, BGRA→RGBA 转换）
- [x] render/capture 包 — 截图/视觉回归测试工具集
  - [x] Screenshot(backend) — 截取当前帧
  - [x] SavePNG / LoadPNG — PNG 文件读写
  - [x] Compare(a, b, threshold) — 像素级图像比对（均值误差/最大误差/差异像素数/差异图）
  - [x] MatchesGolden(img, goldenPath, threshold) — 黄金图对比（首次运行自动 bootstrap）
  - [x] PSNR(a, b) — 峰值信噪比计算
  - [x] MustMatchGolden(t, backend, golden, threshold) — 测试断言（失败时自动保存 actual + diff 图）
  - [x] UpdateGolden(backend, path) — 更新黄金图
  - [x] AssertImageEqual(t, expected, actual, threshold) — 图像等值断言

### 字体系统 ✅

- [x] FreeType 动态链接引擎（零 CGO，syscall 加载 freetype.dll）
- [x] 字形光栅化（FreeType SDF / Bitmap）
- [x] 字形纹理图集（shelf bin-packing、LRU 淘汰、脏区域上传）
- [x] 字体系统接口定义（Engine, Manager, Shaper）
- [x] 字体管理器实现（注册、CSS 权重匹配、回退链）
- [x] 基础文本排版器（左到右排版、自动换行、CJK 断行、对齐）
- [x] 东亚文本排版完善（标点挤压）— 行首/相邻标点压缩（W3C CLREQ/JLREQ）
- [x] 中文分词引擎（结巴算法，DAG + DP，纯 Go 零依赖）
- [x] 分词感知换行与截断（词边界断行、字/词级截断 + 省略号）
- [x] TextRenderer 集成层（shaper → atlas → render 命令桥接）
- [x] DPI 感知字体渲染（pt→px 按 DPI 缩放，基础设施层处理）
- [x] 字形图集双线性过滤（消除文本锯齿，匹配 Windows 原生质量）
- [x] 单元测试覆盖率 90%+（font 95.1%, atlas 96.5%, segment 98.3%, textrender 91.2%, freetype 92.2%）

### 核心 ✅

- [x] 元素树（创建、挂载、卸载、递归销毁）
- [x] 事件分发（冒泡/捕获）— W3C 三阶段模型（capture → target → bubble）
- [x] 焦点管理（SetFocused，自动互斥——同一时刻仅一个元素获焦）
- [x] 脏标记与增量更新（DirtyLayout / DirtyPaint / DirtyStyle）
- [x] 命中测试（HitTest 深度优先）

### 布局 ✅

- [x] CSS 属性类型（Display, Position, Flex*, Align*, Justify*, Value/Unit）
- [x] Flexbox 布局算法（grow/shrink、wrap、6 种 justify、5 种 align、gap）
- [x] Block 流式布局（垂直堆叠、auto 高度）
- [x] 盒模型（margin/padding/border）
- [x] 百分比和 auto 尺寸
- [x] 绝对定位（left/top/right/bottom 推断宽高）
- [x] min/max 约束
- [x] 单元测试覆盖率 87.2%+

### 组件 P0（基础 + 布局 + 核心输入） ✅

- [x] Text 文本
- [x] Button 按钮（Primary/Secondary/Text/Link 四种变体，hover/press/disabled 状态，全变体视觉反馈）
- [x] Icon 图标（纹理 + 着色）
- [x] Input 输入框（键盘输入、光标闪烁、删除、placeholder、焦点、禁用、文本选择、剪贴板、右键菜单、IME 定位）
- [x] Div / ScrollView 容器（背景、边框、圆角、裁剪滚动）
- [x] Grid 栅格（Row + Col，24 列系统，gutter/offset）
- [x] Layout 布局（Header/Content/Footer/Aside）
- [x] Space 间距（水平/垂直、可配置 gap）
- [x] Popup 基础弹出层（8 种 placement，锚点定位）
- [x] Tooltip 文字提示（hover 自动显示/隐藏）
- [x] ConfigProvider 全局配置（主题色、字体、间距、圆角）
- [x] Widget 基础架构（Widget 接口、Base 组合、生命周期、事件绑定）
- [x] 单元测试覆盖率 94.1%+（widget 94.1%, widget/game 99.5%）

### 验收演示程序 ✅

- [x] cmd/demo — Windows + Vulkan 完整渲染演示
  - 经典后台布局（Header + Aside + Content + Footer）
  - 按钮组（5 种变体）、输入框组（含 placeholder/disabled）
  - 24 列栅格、Tooltip、主题配色
  - 事件路由（鼠标命中测试 + 焦点管理 + 键盘分发）
  - 无 GPU 纯逻辑测试（40 个 widget，40 条渲染命令）

### 里程碑: 能在 Windows 上用 Vulkan 渲染一个包含文本、按钮、输入框的简单界面 ✅

## 阶段 2: 核心完善（第 9-16 周）

### 平台

- [x] Windows IME 完善支持（IMM32 组合/候选、CFS_EXCLUDE 精准定位、WM_IME_STARTCOMPOSITION 时机）
- [x] Windows 原生右键菜单（TrackPopupMenu、ClientToScreen）
- [x] Windows IME 完整支持（TSF）
- [ ] Linux 平台层（X11）
- [ ] Linux 输入和基础 IME

### 渲染

- [x] SDF 抗锯齿四边形扩展修复（圆角矩形 AA 品质提升）
- [x] 叠加渲染层（Overlay layer，弹出层/下拉菜单独立渲染）
- [x] 脏标记驱动渲染循环（无变化时不重绘，Loading 等动画组件自标记）
- [x] OpenGL 3.3 后端
- [x] 渲染命令批处理优化
- [x] 纹理图集动态扩展和 LRU

### 核心

- [x] 焦点管理完善（ClearFocus、点击空白区域自动失焦）
- [x] 脏标记公共 API（MarkDirty / NeedsRender / ClearAllDirty）
- [x] 光标管理（IBeam 文本光标、Hand 可点击光标、自定义描边光标）

### HTML+CSS 模板驱动 UI 系统（详见 [design-html-css.md](design-html-css.md)）

用 HTML+CSS 定义布局，Go 侧注入数据 + 绑定事件，降低使用门槛。

- [x] CSS 解析器 + 选择器匹配（`css/` 包，选择器引擎，级联特异性）
- [x] HTML 解析器增强（更多标签、`<style>` 块、CSS 属性全覆盖）
- [x] HTML 全控件标签支持（47 种控件标签：card/tabs/table/menu/collapse/dialog/drawer/slider/alert/badge/avatar 等）
- [x] background 完整支持（linear-gradient 解析与渲染、Div 渐变背景、background-image/size/repeat/position CSS 属性）
- [x] 透明度体系（opacity CSS 属性解析、值钳位、applyVisualProps 应用）
- [x] 盒模型完善（box-sizing、box-shadow CSS 属性解析存储）
- [x] 盒模型高级（margin collapse、border-style/单边 shorthand、box-shadow 解析与 Div 渲染）
- [x] 事件绑定（`doc.OnClick/OnChange/OnToggle`，Go 侧回调绑定）
- [x] 事件委托与声明式绑定（`doc.On(selector, event, handler)`、`doc.QueryAll(selector)`）
- [x] 数据绑定与模板（`{{}}` 插值、`data-if`、`data-model` 双向绑定、`SetData/GetData`）
- [x] 查询与动态操作（`QueryByID/QueryByClass/QueryByTag/QueryAll`、`AppendChild/RemoveChild`、`SetStyle`）
- [x] Document 生命周期（`Dispose()` 清理绑定与控件树）
- [x] CSS 变量系统（`:root` 变量解析、`var(--name, fallback)` 解析替换）
- [x] CSS 变量主题切换（`app.SetTheme()` / `doc.SetTheme()` 换主题 = 注入 `--ui-*` CSS 变量集）
- [x] Theme → CSS 变量导出（`Theme.ToCSSVariables()` 输出 `--ui-*` 变量映射）
- [x] 零样板 App 入口（`ui.NewApp` + `app.LoadHTML` + `app.Run`）

### 布局

- [x] Grid（CSS Grid）布局算法（template-columns/rows, fr/px/pct/auto, gap, 显式/自动放置, span）
- [x] Absolute 定位（已在 layout 引擎实现）
- [x] Relative 定位（流内偏移，不影响兄弟布局）
- [x] Flex 容器 auto 高度（Column/Row 按内容自适应尺寸）
- [x] Flex 子项内容自适应（intrinsic sizing，递归测量子树）
- [x] Flex Order 属性（稳定排序，保持源码顺序）
- [x] Fixed 定位（相对视口定位，脱离正常流）
- [x] overflow: scroll/auto（内容尺寸追踪，OverflowAuto 类型）

### 组件 P0（输入 + 反馈 + 数据展示）

- [x] TextArea 多行文本框（多行编辑、选择、剪贴板、上下键导航）
- [x] Checkbox / Radio / Switch（勾选、单选组互斥、开关切换）
- [x] Select 选择器（下拉面板、选项高亮、禁用项）
- [x] Form / FormItem 表单（标签+控件布局、必填标记、错误提示）
- [x] Image 图片（纹理显示、Fit 模式、着色）
- [x] Tag 标签（5 种预设类型 + 自定义颜色）
- [x] Tabs 选项卡（多 Tab 切换、活跃指示器、hover 反馈）
- [x] Dialog 对话框（模态遮罩、居中面板、标题 + 关闭按钮）
- [x] Message 全局提示（4 种类型：info/success/warning/error）
- [x] 组件渲染修复（Checkbox 勾选、Switch 裁剪、Select 箭头、Loading 弹跳动画）
- [x] MessageBox 消息弹出框（模态确认、5 种类型、OK/Cancel 按钮）
- [x] Progress 进度条（百分比 + 3 种状态颜色）
- [x] Loading 加载（三点指示器 + 提示文本）
- [x] Empty 空状态（图标占位 + 描述文字）

### API

- [x] 声明式 API（Builder 模式，fluent 链式构建 widget 树）
- [x] HTML+CSS 加载 API（简易 HTML 解析 → widget 树，内联样式支持）
- [x] 数据绑定系统（State[T] 响应式容器、Watch/Bind/Computed、ListState）
- [x] 主题系统（token 化设计、Light/Dark 预设、Theme → Config 桥接）

### 里程碑: 完整的表单应用 demo，跨 Windows/Linux，支持中文输入

## 阶段 3: 游戏集成 + P0 游戏组件（第 17-24 周） ✅

### 集成

- [x] 嵌入模式 API（EmbeddedUI：共享渲染后端、事件桥接、层管理）
- [x] 嵌入模式 API（共享 OpenGL 上下文）
- [x] 命令导出模式（HeadlessUI：无头渲染、ExportJSON 序列化命令）
- [x] 输入事件桥接与穿透（HandleEvent 统一入口，鼠标/键盘/IME 路由）
- [x] 多层 UI 管理（LayerManager：Base/HUD/Dialog/Chat/Tooltip 五层，z-order 排序）

### 即时模式

- [x] IMContext 基础框架（im 包，每帧重建 UI，面板/光标布局系统）
- [x] 即时模式基础 Widget（Text/Button/Slider/Checkbox/ProgressBar/Separator）
- [x] 即时模式面板和布局（Begin/End 面板，自动垂直布局，padding/spacing）
- [x] 即时模式与声明式混合使用（EmbeddedUI.Layers 组合 im.Context 输出）

### 游戏 P0 组件

- [x] HUD 抬头显示层（9 点锚定、元素管理、视口自适应）
- [x] HealthBar 生命/资源条（当前/最大值、颜色、文本、圆角）
- [x] Hotbar 快捷栏（N 槽位、选中高亮、冷却叠加、快捷键标签）
- [x] CooldownMask 冷却遮罩（比例遮罩叠加层）
- [x] Inventory 背包（行×列网格、物品放置/移除、稀有度边框、数量徽标）
- [x] ChatBox 聊天框（消息列表、滚动、频道颜色、输入区）
- [x] FloatingText 浮动文字（位置、颜色、动态文本）
- [x] ItemTooltip 物品提示（稀有度边框、叠加层渲染）
- [x] NotificationToast 通知浮窗（4 种类型、颜色条、叠加层）
- [x] DragDrop 拖放系统（DragSource/DropTarget 注册、阈值检测、Accept/OnDrop 回调）

### 性能

- [x] 渲染命令缓存（静态 UI 不重新生成）
- [x] 对象池
- [x] 性能 Profiler 集成

### 里程碑: 完整的游戏 UI demo（HUD + 背包 + 聊天 + 菜单） ✅

## 阶段 4: 平台扩展 + P1 组件（第 25-34 周）

### 平台

- [ ] macOS 平台层（Cocoa via purego）
- [ ] macOS IME 支持
- [ ] Linux Wayland 支持
- [ ] Linux Fcitx/IBus IME
- [ ] Android 基础支持
- [ ] iOS 基础支持
- [x] DirectX 11 后端（Windows）

### P1 通用组件 ✅

- [x] Link 链接（可点击文本 + 下划线，hover 颜色）
- [x] Divider 分割线（水平/垂直，居中文本）
- [x] InputNumber 数字输入框（+/- 按钮，min/max/step）
- [x] Slider 滑块（轨道 + 滑块，拖拽交互，步进吸附）
- [x] ColorPicker 颜色选择器（色板按钮 + 下拉预设网格）
- [x] DatePicker / TimePicker 日期时间选择器（日历网格叠加层）
- [x] Badge 徽标（计数/圆点，定位于子组件）
- [x] Avatar 头像（圆形/方形，文本首字母/图片）
- [x] Card 卡片（标题栏 + 内容区 + 边框）
- [x] List 列表（可滚动项目列表）
- [x] Popover 气泡弹出框（锚定浮动内容面板）
- [x] Collapse 折叠面板（展开/收起，手风琴模式）
- [x] Menu 导航菜单（垂直菜单，子菜单，选中高亮）
- [x] Breadcrumb 面包屑（分隔符连接导航链接）
- [x] Pagination 分页（页码按钮 + 前/后翻页）
- [x] Notification 消息通知（堆叠式 toast，4 种类型）
- [x] Drawer 抽屉（边缘滑出面板 + 遮罩）
- [x] Alert 警告提示（彩色横幅，4 种类型）
- [x] VirtualList 虚拟列表（仅渲染可见项）
- [x] Splitter 分割面板（可拖拽分割条）
- [x] Panel 面板（标题 + 边框容器）
- [x] SubWindow 子窗口（可拖拽浮动窗口）
- [x] ContextMenu 右键菜单（菜单项 + 分隔线）
- [x] Portal 传送（根层叠加渲染）
- [x] Table 表格（列头 + 行数据，条纹/边框）
- [x] Tree 树形控件（展开/收起节点，选中，搜索）

### P1 游戏组件 ✅

- [x] Minimap 小地图（圆形/方形地图、标记点、玩家指示器、缩放）
- [x] RadialMenu 径向菜单（环形菜单项、高亮、选中回调）
- [x] QuestTracker 任务追踪（任务列表、目标进度、完成标记）
- [x] BuffBar 增益/减益栏（图标行、持续时间、叠层数、正/负类型）
- [x] Nameplate 名牌（浮动名字+血条、友好/敌对/中立颜色）
- [x] Scoreboard 计分板（玩家列表、分数/击杀/死亡、排序）
- [x] DialogueBox NPC 对话框（说话人、文本、选项分支、推进回调）
- [x] CountdownTimer 倒计时（分:秒显示、Tick 更新、到期回调、低时闪红）
- [x] CurrencyDisplay 货币显示（多币种、符号+数量、K/M 格式化）
- [x] TeamFrame 队伍框架（队员列表、HP/MP 双条、等级、死亡状态）
- [x] TargetFrame 目标框架（选中目标显示、HP/MP、SetTarget/ClearTarget）
- [x] CastBar 施法条（施法进度、技能名、完成/打断回调）

## 阶段 5: 高级功能 + P2 组件（第 35-44 周） ✅

### 核心功能 ✅

- [x] 动画系统（过渡、关键帧）— anim 包，Tween/Keyframe/Spring/Sequence，18 种缓动函数
- [x] SVG 路径渲染 — ParseSVGPath 全指令集（M/L/H/V/C/S/Q/T/A/Z），Bezier 边界计算，弧线端点→中心转换
- [x] 富文本（混合字体/颜色/图片）— RichText widget，Span 模型，内联样式混排
- [x] 子窗口停靠系统（Docking）— DockManager，四向停靠区，拖拽分割
- [x] 手柄导航 — Gamepad 输入映射，方向导航，焦点遍历
- [x] Canvas 画布 — W3C Canvas 2D Context API 风格，软件光栅化 → GPU 纹理上传，平台无关

### P2 通用组件 ✅

- [x] Cascader 级联选择器（多级联动下拉面板）
- [x] TreeSelect 树选择（下拉树形结构选择）
- [x] Transfer 穿梭框（双列表移动项目）
- [x] DateRangePicker 日期范围选择器（起止日期选择）
- [x] Upload 上传（拖拽上传区域，文件管理）
- [x] AutoComplete 自动完成（输入联想下拉建议）
- [x] TagInput 标签输入框（标签添加/删除，换行）
- [x] RangeInput 范围输入框（双滑块范围选择）
- [x] Steps 步骤条（导航进度指示器）
- [x] Anchor 锚点（页面节锚定导航侧栏）
- [x] BackTop 回到顶部（浮动返回按钮）
- [x] ImageViewer 图片预览（缩放/平移）
- [x] VirtualGrid 虚拟网格（仅渲染可见单元格）
- [x] MenuBar 菜单栏（水平应用菜单）
- [x] Affix 固钉（滚动固定内容）
- [x] Timeline 时间线（垂直事件列表，状态颜色）
- [x] Swiper 轮播（面板切换 + 指示点）
- [x] Statistic 统计数值（标题 + 大号数字）
- [x] Skeleton 骨架屏（加载占位符动画）
- [x] Watermark 水印（重复文本覆盖）
- [x] Calendar 日历（完整月历网格）
- [x] Comment 评论（头像 + 作者 + 内容 + 回复）
- [x] Descriptions 描述列表（标签-值对列表）
- [x] Rate 评分（星形评分输入）
- [x] Guide 引导（分步引导覆盖层、遮罩高亮、步骤卡片）
- [x] Popconfirm 气泡确认框（锚定确认弹出）

### P2 游戏组件 ✅

- [x] LootWindow 拾取窗口（物品列表、稀有度边框、拾取/全部拾取、已拾取标记）
- [x] SkillTree 天赋/技能树（节点网络、前置依赖连线、解锁/等级、技能点管理）

### 质量保障 ✅

- [x] 全模块代码审查（SVG 平滑曲线反射修复、响应式状态观察者 ID 化、JSON 结构标签修复、滑块拖拽 activeID、HUD 锚定定位）
- [x] 单元测试覆盖率 90%+（render 92.4%, anim 96.8%, widget 94.1%, widget/game 99.5%, im 91.0%, root 98.5%, theme 100%, core 89.4%, canvas 99.5%）
- [x] 全部 22 个包测试通过

### 里程碑: 动画、SVG 路径、富文本、停靠系统、手柄导航全部完成，P2 组件齐备 ✅

## 阶段 6: 生态与打磨（第 45 周+）

- [ ] 完整文档和示例
- [ ] 可视化 UI 编辑器（可选）
- [ ] 更多内置主题
- [ ] 内置图标集
- [x] 无障碍（Accessibility）基础支持
- [ ] 性能基准测试
- [ ] 稳定 API，发布 v1.0

## 项目目录结构（规划）

```
github.com/kasuganosora/ui/
├── docs/                    # 设计文档
├── cmd/
│   └── examples/            # 示例程序
│       ├── hello/           # 最简示例
│       ├── form/            # 表单 demo
│       ├── game-hud/        # 游戏 HUD demo
│       └── inventory/       # 背包系统 demo
├── math/                    # 数学库 (Vec2, Rect, Color, Mat3)
├── platform/                # 平台抽象层
│   ├── windows/
│   ├── linux/
│   ├── darwin/
│   ├── android/
│   └── ios/
├── render/                  # 渲染抽象层
│   ├── vulkan/
│   ├── opengl/
│   ├── dx11/
│   └── software/
├── font/                    # 字体系统
│   ├── freetype/            # FreeType CGO 绑定
│   ├── gofont/              # 纯 Go 备选（无 CGO）
│   ├── shaper/              # 文本排版与断行
│   ├── atlas/               # 字形纹理图集
│   └── cache/               # 字形缓存
├── css/                     # CSS 解析器
├── html/                    # HTML 解析器
├── layout/                  # 布局引擎 (Flexbox, Grid, Flow)
├── core/                    # 核心层 (Element Tree, Event, State)
├── widget/                  # 组件库
│   ├── basic/               # 基础: Button, Icon, Link
│   ├── layout/              # 布局: Grid, Layout, Space, Divider
│   ├── navigation/          # 导航: Menu, Tabs, Breadcrumb, Pagination, Steps, Anchor
│   ├── input/               # 输入: Input, TextArea, Select, Checkbox, Radio, Switch, Slider, ...
│   ├── display/             # 展示: Text, Tag, Badge, Avatar, Image, Card, Table, Tree, ...
│   ├── feedback/            # 反馈: Dialog, Message, Notification, Drawer, Progress, Loading, ...
│   ├── container/           # 容器: ScrollView, VirtualList, Splitter, SubWindow, ContextMenu
│   ├── utility/             # 工具: DragDrop, Affix, ConfigProvider, Portal
│   └── game/                # 游戏: HUD, HealthBar, Hotbar, Inventory, ChatBox, Minimap, ...
├── theme/                   # 主题系统
├── anim/                    # 动画系统
├── im/                      # 即时模式 API
├── app.go                   # 独立模式入口
├── embedded.go              # 嵌入模式入口
├── headless.go              # 命令导出模式入口
└── goui.go                  # 公共 API 导出
```
