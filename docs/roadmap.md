# 开发路线图

## 阶段 0: 基础设施（第 1-2 周） ✅

- [x] 项目结构搭建、CI/CD
- [x] 基础数学库（Vec2, Rect, Color, Mat3, Edges, Corners）
- [x] 平台层接口定义（platform.Platform, platform.Window）
- [x] 渲染后端接口定义（render.Backend, render.CommandBuffer）
- [x] 事件系统基础类型（event.Event, event.Type, event.Key, event.MouseButton）
- [x] 单元测试覆盖率 90%+（math 99.4%, event 100%, render 100%）

## 阶段 1: 最小可用（第 3-8 周）

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
- [x] 单元测试覆盖率 90%+（font 95.1%, atlas 96.5%, segment 98.3%, textrender 91.2%, freetype 92.2%）

### 核心 ✅

- [x] 元素树（创建、挂载、卸载、递归销毁）
- [x] 事件分发（冒泡/捕获）— W3C 三阶段模型（capture → target → bubble）
- [x] 焦点管理（SetFocused）
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
- [x] 单元测试覆盖率 94.5%

### 组件 P0（基础 + 布局 + 核心输入）

- [ ] Text 文本
- [ ] Button 按钮
- [ ] Icon 图标
- [ ] Input 输入框
- [ ] Div / ScrollView 容器
- [ ] Grid 栅格（Row + Col）
- [ ] Layout 布局（Header/Content/Footer/Aside）
- [ ] Space 间距
- [ ] Popup 基础弹出层
- [ ] Tooltip 文字提示
- [ ] ConfigProvider 全局配置

### 里程碑: 能在 Windows 上用 Vulkan 渲染一个包含文本、按钮、输入框的简单界面

## 阶段 2: 核心完善（第 9-16 周）

### 平台

- [x] Windows IME 基础支持（IMM32 组合/候选）
- [ ] Windows IME 完整支持（TSF）
- [ ] Linux 平台层（X11）
- [ ] Linux 输入和基础 IME

### 渲染

- [ ] OpenGL 3.3 后端
- [ ] 渲染命令批处理优化
- [ ] 纹理图集动态扩展和 LRU

### 布局

- [ ] HTML 解析器
- [ ] CSS 选择器匹配
- [ ] Grid（CSS Grid）布局算法
- [x] Absolute 定位（已在 layout 引擎实现）
- [ ] Fixed 定位
- [ ] overflow: scroll/auto

### 组件 P0（输入 + 反馈 + 数据展示）

- [ ] TextArea 多行文本框
- [ ] Checkbox / Radio / Switch
- [ ] Select 选择器
- [ ] Form / FormItem 表单
- [ ] Image 图片
- [ ] Tag 标签
- [ ] Tabs 选项卡
- [ ] Dialog 对话框
- [ ] Message 全局提示
- [ ] MessageBox 消息弹出框
- [ ] Progress 进度条
- [ ] Loading 加载
- [ ] Empty 空状态

### API

- [ ] 声明式 API 完善
- [ ] HTML+CSS 加载 API
- [ ] 数据绑定系统
- [ ] 主题系统

### 里程碑: 完整的表单应用 demo，跨 Windows/Linux，支持中文输入

## 阶段 3: 游戏集成 + P0 游戏组件（第 17-24 周）

### 集成

- [ ] 嵌入模式 API（共享 Vulkan 上下文）
- [ ] 嵌入模式 API（共享 OpenGL 上下文）
- [ ] 命令导出模式
- [ ] 输入事件桥接与穿透
- [ ] 多层 UI 管理（HUD/Dialog/Chat）

### 即时模式

- [ ] IMContext 基础框架
- [ ] 即时模式基础 Widget（Text/Button/Slider/Checkbox）
- [ ] 即时模式面板和布局
- [ ] 即时模式与声明式混合使用

### 游戏 P0 组件

- [ ] HUD 抬头显示层
- [ ] HealthBar 生命/资源条
- [ ] Hotbar 快捷栏
- [ ] CooldownMask 冷却遮罩
- [ ] Inventory 背包
- [ ] ChatBox 聊天框
- [ ] FloatingText 浮动文字
- [ ] ItemTooltip 物品提示
- [ ] NotificationToast 通知浮窗
- [ ] DragDrop 拖放系统

### 性能

- [ ] 渲染命令缓存（静态 UI 不重新生成）
- [ ] 对象池
- [ ] 性能 Profiler 集成

### 里程碑: 完整的游戏 UI demo（HUD + 背包 + 聊天 + 菜单）

## 阶段 4: 平台扩展 + P1 组件（第 25-34 周）

### 平台

- [ ] macOS 平台层（Cocoa via purego）
- [ ] macOS IME 支持
- [ ] Linux Wayland 支持
- [ ] Linux Fcitx/IBus IME
- [ ] Android 基础支持
- [ ] iOS 基础支持
- [ ] DirectX 11 后端（Windows）

### P1 通用组件

- [ ] Link 链接
- [ ] Divider 分割线
- [ ] InputNumber 数字输入框
- [ ] Slider 滑块
- [ ] ColorPicker 颜色选择器
- [ ] DatePicker / TimePicker 日期时间选择器
- [ ] Badge 徽标
- [ ] Avatar 头像
- [ ] Card 卡片
- [ ] List 列表
- [ ] Popover 气泡弹出框
- [ ] Collapse 折叠面板
- [ ] Menu 导航菜单
- [ ] Breadcrumb 面包屑
- [ ] Pagination 分页
- [ ] Notification 消息通知
- [ ] Drawer 抽屉
- [ ] Alert 警告提示
- [ ] VirtualList 虚拟列表
- [ ] Splitter 分割面板
- [ ] Panel 面板
- [ ] SubWindow 子窗口
- [ ] ContextMenu 右键菜单
- [ ] Portal 传送
- [ ] Table 表格
- [ ] Tree 树形控件

### P1 游戏组件

- [ ] Minimap 小地图
- [ ] RadialMenu 径向菜单
- [ ] QuestTracker 任务追踪
- [ ] BuffBar 增益/减益栏
- [ ] Nameplate 名牌
- [ ] Scoreboard 计分板
- [ ] DialogueBox NPC 对话框
- [ ] CountdownTimer 倒计时
- [ ] CurrencyDisplay 货币显示
- [ ] TeamFrame 队伍框架
- [ ] TargetFrame 目标框架
- [ ] CastBar 施法条

## 阶段 5: 高级功能 + P2 组件（第 35-44 周）

### 核心功能

- [ ] 动画系统（过渡、关键帧）
- [ ] SVG 路径渲染
- [ ] 富文本（混合字体/颜色/图片）
- [ ] 子窗口停靠系统（Docking）
- [ ] 手柄导航

### P2 通用组件

- [ ] Cascader 级联选择器
- [ ] TreeSelect 树选择
- [ ] Transfer 穿梭框
- [ ] DateRangePicker 日期范围选择器
- [ ] Upload 上传
- [ ] AutoComplete 自动完成
- [ ] TagInput 标签输入框
- [ ] RangeInput 范围输入框
- [ ] Steps 步骤条
- [ ] Anchor 锚点
- [ ] BackTop 回到顶部
- [ ] ImageViewer 图片预览
- [ ] VirtualGrid 虚拟网格
- [ ] MenuBar 菜单栏
- [ ] Affix 固钉
- [ ] Timeline 时间线
- [ ] Swiper 轮播
- [ ] Statistic 统计数值
- [ ] Skeleton 骨架屏
- [ ] Watermark 水印
- [ ] Calendar 日历
- [ ] Comment 评论
- [ ] Descriptions 描述列表
- [ ] Rate 评分
- [ ] Guide 引导
- [ ] Popconfirm 气泡确认框

### P2 游戏组件

- [ ] LootWindow 拾取窗口
- [ ] SkillTree 天赋/技能树

## 阶段 6: 生态与打磨（第 45 周+）

- [ ] 完整文档和示例
- [ ] 可视化 UI 编辑器（可选）
- [ ] 更多内置主题
- [ ] 内置图标集
- [ ] 无障碍（Accessibility）基础支持
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
