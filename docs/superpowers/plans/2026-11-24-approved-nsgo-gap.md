# 对照 nsgo 补齐并优化组件 — 已批准设计

> 批准时间：2026-11-24（Asia/Shanghai）。用户原话：「按方案分阶段开始实现，平台接入尚未完成的部分单独列出。」
>
> 本文件是 2026-11-23 修订设计的批准快照。里程碑 1–2 的具体任务在 `2026-11-24-m1-m2-implementation.md`。里程碑 3、4 各自另有计划，不在本增量里实现。平台原生接入只记在覆盖账本的「仅接口」列，不记为完成。

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.
>
> 修订说明：上一版把「目标 `docs/components.md` 里没有的类型」默认排除了。用户要求缺少的组件和需要优化的地方都要补上。本版保留原来的五批行为修复，并把缺失的通用控件放进里程碑 3。图表、编辑器、公式、图形、终端等 UI 子系统放进里程碑 4，带依赖、复杂度和许可证约束，不再写成范围外。

**Goal:** 在目标自己的自绘栈和现有公开 API 上，增量补齐 nsgo 有而目标没有的通用控件，并修好目标已有控件的不完整行为。不把 nsgo 整仓移植过来，也不换成原生视图后端。

**Architecture:** 新控件仍是 `widget.Widget`：`NewX(tree, cfg)`、`Draw` 进 `render.CommandBuffer`、事件走 `core.Tree`。nsgo 的 `base`/`ui` 只提供行为契约（键盘、焦点、命中、受控值）。实现按目标现有控件的写法重写。

**Tech Stack:** Go 1.25 模块（本环境 `go1.27.3` 已能编译）。渲染保持 Vulkan / OpenGL / DirectX / Metal / 软件。平台保持 Win32 / X11 / Wayland / Cocoa / Android 骨架 / Godot。不引入 CGO 作为新设计。

## 请用户确认的唯一范围

批准按里程碑 1 → 2 → 3 → 4 做完本文件列出的全部 UI 能力：里程碑 1–2 是原来的五批（已有控件的行为），里程碑 3 是表 2 的 20 个缺失通用控件（含 Command/CommandDialog、OtpInput、HoverCard、HoldToConfirm、Meter、Toolbar、ToggleGroup、ToggleButton、ImageCompare、SuggestList，以及按自有目录实现的 EmojiPicker），里程碑 4 是表 3 的 19 个子系统，并采用文末「产品默认」而不再单独砍掉某一类控件。

不批准时，需要改的是这份默认，而不是重新发现一遍。

## Global Constraints

- 只改 `github.com/kasuganosora/ui`。`github.com/nsgo-dev/nsgo` 只读，不修改、不推送。
- 保持 `widget.Widget`、已有 `NewX` 签名、HTML 标签名和方法语义。新能力用新类型或新方法。
- 不复制 nsgo 源码、`emoji.bin`、Lucide 包、Ghostty 二进制、New Computer Modern 字体、Taffy 夹具或 shader。行为按本文件重写。
- 新测试在无 GPU 环境通过。不把 Vulkan 变成新测试的前提。
- 日后提交说明用英文，不加 `Co-authored-by` 或任何归属尾注。
- 报告和代码不写密钥或生产基础设施。

## 钉住的版本

| 仓库 | 默认分支 | 提交 | 树 | 提交时间 |
| --- | --- | --- | --- | --- |
| `kasuganosora/ui`（工作区 `/workspace`） | `master` | `deefbf9bb4e31016a78f40e29697c8efbef206a2` | `70e84427e636e24547e35455481c0c0cc6e79829` | 2026-04-12 |
| `nsgo-dev/nsgo`（只读 `/tmp/nsgo`） | `main` | `2e9c12b7bea1f9139b7a4a12fb2471dddeeefa79` | `3b6951958df67f90162776a987a7fd3e26295237` | 2026-11-23 |

nsgo 历史只加深了 25 个提交，用来确认 `main` 仍在改控件，不是全历史审计。

## 计数规则

一行是一个应用能单独构造的 UI 能力。`base` 与 `ui` 的同一控件算一行。子零件（`MenuItem`、`Tab`、`AccordionItem`、`ToolbarSeparator`、`Option`、`ColorWell`、`AlertAction`）不算行。主题结构、locale 生成物、测试文件不算行。

数字是下面各表的行数，不是估计。

| 表 | 行数 | 含义 |
| --- | ---: | --- |
| 表 1 | 36 | nsgo 有对应物，目标已有类型，行为不完整。里程碑 1–2 |
| 表 1 附 | 11 | 只有目标有的半成品，五批里已经要修，不计入 36 |
| 表 1b | 15 | nsgo 有对应物，主路径已经可用，里程碑 2 只补无障碍属性 |
| 表 2 | 20 | nsgo 有、目标没有类型的通用控件。里程碑 3。不再因为目标目录没写就排除 |
| 表 3 | 19 | 仍是 UI、但依赖重或带许可证约束的子系统。里程碑 4，仍在范围内 |
| 表 4 | 15 | 不是应用内 UI。排除，并写明原因 |
| 表 5 | 18 + 25 | 目标独有、nsgo 没有对应控件。保持不动 |

nsgo UI 能力合计：36 + 15 + 20 + 19 = **90**。90 不包含表 1 附、表 4、表 5。

构造函数扫描（另一套单位，不要和 90 相加）：`widget` 包 95 个 `New*`，其中 `NewBase`、`NewImageData` 不是控件；`widget/game` 25 个 `New*`。

## 许可证

- nsgo：Apache-2.0，版权 2026 ONEO。`NOTICE` 另外指向 NSH、Lucide（ISC）、emojibase（MIT）、CLDR（Unicode License v3）、ProseMirror/Tiptap（MIT）、Ghostty（MIT）、limu/Pele（MIT）、New Computer Modern Math（GUST / LPPL 1.3c）、Taffy 夹具、Kage/Ebiten 等。
- 目标：根目录没有 `LICENSE` / `NOTICE`。
- 里程碑 1–3 的默认路径不复制上述资产，因此不给目标仓库加 Apache 声明。
- 里程碑 4 里若某行的产品默认被改成「带上第三方作品」，那一行才需要相应的版权声明。默认不改。

### EmojiPicker 的资源策略

纳入里程碑 3，不排除。

- 不复制 `base/internal/emoji/emoji.bin`。该文件来自 emojibase-data（MIT）和 CLDR 注释（Unicode License v3）。
- 字形走目标已有的彩色 emoji 字体绘制，不内嵌图片。
- 目录是仓库里手写的 `emoji_catalog.go`：Unicode 标量值，加上为本仓库写的简短中英标签。搜索只匹配这些标签和字符本身。这是原创文案，不带 emojibase 关键词。
- 肤色和 ZWJ 序列只收录目录里写明的项。
- 以后若用户给仓库选定许可证并同意声明，可以另加 `go generate`，在 NOTICE 中保留 emojibase MIT 与 Unicode License v3，再换上完整关键词。那不是里程碑 3 的默认。

## 「组件」在两边分别指什么

目标的组件是 `widget` 与 `widget/game` 的保留式控件，外加 `html.go` 标签、`builder.go` 和 `im` 即时模式。

nsgo 的组件是 `base` 的无样式 `core.Node` 和 `ui` 的主题外观。复杂 UI 在子包：`datatable`、`dock`、`plot`、`toast`、`editor`、`diff`、`richeditor`、`textview`、`terminal`、`formula`、`diagram`、`flow`、`camera`、`commitgraph`、`filepicker`、`chat`。nsgo 的 `widget` 包是主屏幕小组件和 Live Activity，不是控件库。

## 架构

目标是自绘引擎：帧循环、`core.Tree`、flex/grid、控件 `Draw`、多渲染后端。文档把 P0–P2 标成完成，完成的是类型和一部分鼠标绘制。`event.KeyDown` 只出现在 `input.go`、`textarea.go`、`taginput.go`、`gamepad.go`。`accessibility.Accessible` 没有控件实现。`platform/win32/uia.go` 的 `HandleGetObject` 对客户区返回 `(0, false)`。

nsgo 是原生壳：Go 做反应图和布局，平台控件绘制。本设计不改用那条绘制路径。

## 基线（只读，未改源码）

`go version` → `go1.27.3 linux/amd64`。模块为 `go 1.25.0`。

```text
go test -count=1 -timeout 300s ./...
```

退出码 **1**。库包为 `ok`。失败仅 `cmd/demo`、`cmd/feed`、`cmd/game` 的视觉测试：`vkCreateInstance` 返回 `VK_ERROR_UNKNOWN`。无测试文件：`cmd/pet`、`godot/cmd/gdext`、`net`、`platform`、`platform/linux/wayland`。Win32 / Cocoa / Metal / DirectX 在 Linux 上不参与 `./...`。

```text
go build -o /dev/null .
go build -o /dev/null ./widget/ ./widget/game/ ./core/ ./layout/ ./render/ ./accessibility/
```

两条退出码都是 **0**。nsgo 的测试没有跑。

## 表 1 — 已有类型，行为不完整（36）

这些行由里程碑 1–2 的五批覆盖。括号里是被并进同一行的 nsgo 零件，不另计。

| # | nsgo | 目标 | 要补的行为 |
| --- | --- | --- | --- |
| 1 | `Button`（含 `IconButton`、`Press`） | `Button` | Space/Enter 激活；`SetIcon`；方形图标按钮用已有 `SetShape` |
| 2 | `Link` | `Link` | 键盘激活 |
| 3 | `Checkbox` | `Checkbox` | Space 切换 |
| 4 | `Radio` / `RadioGroup` | `Radio` / `RadioGroup` | 一组一个 Tab 停靠，方向键移动 |
| 5 | `Switch` | `Switch` | Space 切换 |
| 6 | `Slider` | `Slider` | 方向键按 step；`SetRange` 目前只存布尔，补 `SetRangeValues` |
| 7 | `Input` / `Field` | `Input` / `TextArea` | 键盘已有；补无障碍名字和焦点可见性 |
| 8 | `NumberInput` | `InputNumber` | 上下键步进 |
| 9 | `Select` | `Select` | 上下/Enter；`filterable` 真正过滤；`multiple` 写入 `SelectedValues` |
| 10 | `Tabs` | `Tabs` | 方向键、Home/End |
| 11 | `Menu`（含 `Submenu`） | `Menu` | 方向键打开和触发 |
| 12 | `ContextMenu` | `ContextMenu` | 键盘；桌面右键保持 |
| 13 | `Pagination` | `Pagination` | 键盘换页 |
| 14 | `Accordion` | `Collapse` | 键盘展开；与下一行共用实现 |
| 15 | `Collapsible` | `Collapse` | 单面板键盘展开 |
| 16 | `Tooltip` | `Tooltip` | 焦点显示，Escape 关闭 |
| 17 | `Popover` | `Popover` | Escape 与焦点归还 |
| 18 | `Dialog` | `Dialog` | `closeOnEscKeydown` 今天没有 `KeyDown` 读取它；补焦点陷阱 |
| 19 | `AlertDialog` | `MessageBox` | 与 Dialog 同一套键 |
| 20 | `Sheet` | `Drawer` | 底部方位已有标签；补滑动关闭。不另建 Sheet 类型 |
| 21 | `DatePicker` | `DatePicker` / `DateRangePicker` | 点击今天只翻转 `open`。点日才改日期；范围第二次点击写结束，结束早于开始则交换 |
| 22 | `TimeField` | `TimePicker` | 点时/分改值 |
| 23 | `Calendar` | `Calendar` | 点日改值 |
| 24 | `ColorPicker` | `ColorPicker` | 点预设调用 `SelectColor`；`SetAlphaEnabled` 默认 false。取色器见表 3 |
| 25 | `Tree` | `Tree` | 方向键展开/选择；复选为新方法 |
| 26 | `List` | `List` / `VirtualList` | 键盘；虚拟列表接滚轮。动态行高不改单一 `SetItemHeight` 契约 |
| 27 | `Table` | `Table` | `OnSortChange` / `OnPageChange` 今天没有调用点。表头排序；`SetPageSize` 未调用时不截断行 |
| 28 | `Carousel` | `Swiper` | `Tick(dt)` 驱动已有 `autoplay` 标志；指针拖拽 |
| 29 | `Rating` | `Rate` | 左右键改分 |
| 30 | `TagInput` | `TagInput` | 回车后保持焦点；退格删除最后一枚 |
| 31 | `Form` | `Form` / `FormItem` | `Validate` / `Reset`。分组外观见表 2 |
| 32 | `Toast` | `Message` / `Notification` | `Tick` 超时移除。弹簧堆叠不移植 |
| 33 | `ScrollView` | `Div` 滚动 | 键盘翻页 |
| 34 | `Resizable` | `Splitter` | 最小尺寸夹紧 |
| 35 | `dock.Area` 的停靠/标签 | `DockLayout` | 为现有 `DockPanel` / `OnDock` 补无头测试。缩放和可序列化布局见表 3 |
| 36 | `dragdrop` | `DragDropManager` | 无头测试覆盖现有源/目标回调；树和标签拖拽复用它 |

### 表 1 附 — 五批里顺手修的目标独有半成品（11）

不计入 90。`MenuBar`、`Grid` 断点、`Upload` 的注入式 `OpenPicker`、`Popconfirm`、`Cascader`、`TreeSelect`、`ImageViewer`、`Affix`、`AutoComplete` 的键盘与点击、`VirtualGrid` 滚轮、`GamepadNavigator` 能走到表 1 的可聚焦控件。

`Upload` 不实现操作系统文件框。完整文件浏览器是表 3 的 `FilePicker`。

`Grid`：`SetSpanAt(breakpoint, span)` 与 `SetViewport(width)`。未调用 `SetViewport` 时 `Span()` 与现在一致。断点宽度 0、576、768、992、1200、1600。

### 表 1b — 主路径已可用，只补无障碍属性（15）

`Avatar`、`AvatarGroup`、`Badge`、`Breadcrumbs`、`Image`、`Progress`（含 `ProgressRing` → 已有 `ProgressThemeCircle`）、`Skeleton`、`Space`（覆盖 `HStack` / `Wrap` / `Spacer`）、`Spinner` → `Loading`、`Stepper` → `Steps`、`Text`、`Cell` → `Card`、`Icon`、`Canvas`、`Portal`。

## 表 2 — 缺失的通用控件（20）

里程碑 3。全部新建 `widget` 类型，并加对应 HTML 标签。依赖里程碑 1 的焦点和里程碑 2 的弹出层命中。

| # | 新类型 | 行为验收（无头） |
| --- | --- | --- |
| 1 | `Command` + `CommandDialog` | 输入缩小列表；上下移动；Enter 触发；Escape 关对话框并还焦 |
| 2 | `OtpInput` | 定长；输入、退格、粘贴填格；满长调用 `OnComplete` |
| 3 | `HoverCard` | 悬停延迟打开；焦点打开；Escape 和外部点击关闭 |
| 4 | `HoldToConfirm` | 按住达到时长触发；提前松开回退且不触发 |
| 5 | `Meter` | 0–1 的值与标记位；与 `Progress` 类型分开 |
| 6 | `Toolbar` | 一个 Tab 停靠；方向键只移动焦点，不激活 |
| 7 | `ToggleGroup` | 单选或多项；方向键移动；与 `RadioGroup` 类型分开 |
| 8 | `ToggleButton` | 按下保持 `Selected()`；Space 切换 |
| 9 | `ImageCompare` | 0–100 的分割位置；拖拽与方向键按 5 改位置 |
| 10 | `SuggestList` | 锚在插入符矩形；上下/Enter 选择；不拥有输入框的值 |
| 11 | `EmojiPicker` | 使用上文自有目录；分类、搜索、选择回调；虚拟网格只画可见格 |
| 12 | `Combobox` | 可编辑文本加列表；与 `Select`、`AutoComplete` 并存，不改它们的签名 |
| 13 | `Scrollbar` | 绑定滚动偏移；拖拽滑块改变偏移 |
| 14 | `Nav` + `NavHeader` | 压入/弹出页面；返回键弹出；标题跟随栈顶 |
| 15 | `FormSection` / `FormRow` / `FormField` | 分组行布局；校验仍走表 1 的 `Form.Validate` |
| 16 | `PlainShell` | 顶栏加内容槽 |
| 17 | `Sidebar` | 项与可折叠分组；方向键 |
| 18 | `SidebarShell` | 侧栏加内容 |
| 19 | `TabShell` | 标签切换内容页 |
| 20 | `Settings` / `SettingsDialog` | 分类列表打开分组行；复用 15 和 17 |

## 表 3 — 扩展子系统（19）

里程碑 4，按表顺序做。每一行都是 UI，不再标成范围外。复杂度是相对表 2 的实现量，不是工期。

| # | 子系统 | 依赖 | 复杂度 | 许可证与默认 |
| --- | --- | --- | --- | --- |
| 1 | `Camera` 平移缩放平面 | 现有指针与 `Tick` | 中 | 无第三方作品。干净重写 |
| 2 | `plot` 图表（折线、面积、柱、点、饼、坐标轴；桑基和树图作为同一包的后半） | 目标 `Canvas` | 高 | 不复制 d3 或 nsgo 源码。干净重写 |
| 3 | `DataTable` 列移动、列宽、左右固定列、大行窗口 | 表 1 的 `Table` 窗口 | 高 | 无。不移植 nsgo 槽位协议，沿用目标行数据 |
| 4 | `TextView` Markdown 子集（标题、列表、代码、链接） | `RichText` 跨度 | 中高 | 默认不新增 goldmark。手写子集。若改用 goldmark（MIT）再加声明 |
| 5 | `flow` 流程图 | 表 3 的 Camera | 高 | 分层布局干净重写 |
| 6 | `filepicker` 文件浏览器 | 应用提供的目录列举回调 | 中 | 无。不调用操作系统对话框 API，除非产品默认改掉 |
| 7 | `commitgraph` 提交图 | `List` 行 | 中 | 无。它是 Git 历史视图，仍然是 UI |
| 8 | `diagram` Mermaid 式图 | `Canvas` | 高 | 默认不依赖 limu（MIT）。只做能测的小布局。完整 Mermaid 需另准 limu |
| 9 | `editor` 代码编辑 | 表 1 的文本输入、虚拟行 | 很高 | 默认不新增 chroma。简单分词。不复制 nsgo `editor` |
| 10 | `diff` | 表 3 的 editor 位置模型 | 高 | 无额外资产 |
| 11 | `richeditor` | 文档模型，可复用 editor 的选区 | 很高 | 不复制 ProseMirror/Tiptap 移植。干净重写才不需要 MIT 声明 |
| 12 | `formula` | `Canvas`；字体由调用方提供 | 高 | 默认不嵌入 NewCMMath（GUST/LPPL）。不复制 untie |
| 13 | `ui/chat` 会话界面 | `List`、`Input`、气泡 | 高 | 无。这是界面，不是模型服务 |
| 14 | Dock 的缩放与可序列化布局树 | 表 1 的 `DockLayout` | 中 | 无 |
| 15 | `terminal` | 伪终端与 VT | 极高 | 不复制 nsgo 里的 libghostty 二进制。默认只做字节流视图和键盘路由接口；完整仿真要另准自建或引入 MIT 库并写声明 |
| 16 | `core.WebView` | 平台 WebView | 高 | 先定目标侧接口和无头假实现。真 WebKit/WebView2 等到对应 SDK 出现在环境里 |
| 17 | `core.Video` | 平台播放器 | 高 | 与 WebView 同一约束 |
| 18 | `core.Shader` | 渲染后端 | 高 | 不复制 Kage/Ebiten。干净重写 |
| 19 | `ColorEyeDropper` | 屏幕取色权限 | 中 | 无资产。默认不在里程碑 3，因为要操作系统授权 |

## 表 4 — 排除的基础设施（15）

这些不是应用里的控件。排除原因如下。

| # | 参考内容 | 排除原因 |
| --- | --- | --- |
| 1 | `tsx` 运行时、打包器、编译器、签名包 | 应用投递格式，不是控件 |
| 2 | `cmd/nsgo` 的打包、签名、安装器 | 发布工具 |
| 3 | `updater` | 应用更新服务 |
| 4 | `cloud` | 存储服务 |
| 5 | `gameservices` | 平台账号与成就服务 |
| 6 | `audio` | 播放服务，没有对应视图契约要在本设计里重做 |
| 7 | `microphone` | 采集服务 |
| 8 | nsgo `widget` 包（主屏幕小组件、Live Activity） | 操作系统扩展生命周期，不是 `widget.Widget` |
| 9 | 用 AppKit/GTK/Direct2D/DOM/小程序视图替换绘制 | 会换掉目标渲染栈 |
| 10 | 小程序与 HarmonyOS 宿主 | 另一套宿主，不是控件 |
| 11 | `automation` MCP 与 `nsgotest` | 代理和测试驱动，不是界面 |
| 12 | CLDR 代码生成和 YAML 文案目录 | 工具链。控件继续用现有中文默认和 setter |
| 13 | 用 Lucide 换掉 Material 图标 | 目标已有图标集。新控件用现有 `Icon` 或几何图形 |
| 14 | 蓝牙、推送、通讯录、日历、定位 | 设备与账号服务；nsgo 文档里有的还没建成 |
| 15 | 把 `desktop` 的托盘、全局热键、窗口服务再移植一套 | 目标 `platform` 已负责窗口。重复一层平台，不是缺控件 |

## 表 5 — 目标独有，保持（18 个通用构造 + 25 个游戏构造）

通用：`Alert`、`Anchor`、`BackTop`、`Comment`、`Descriptions`、`Guide`、`Panel`、`RangeInput`、`RichText`、`Statistic`、`SubWindow`、`Tag`、`Timeline`、`Transfer`、`Watermark`、`Empty`、`Layout`（`Header`/`Content`/`Footer`/`Aside` 是它的零件）、`Div`。

游戏：`NewBuffBar`、`NewCastBar`、`NewChatBox`、`NewFloatingText`、`NewItemTooltip`、`NewNotificationToast`、`NewCountdownTimer`、`NewCurrencyDisplay`、`NewDialogueBox`、`NewHUD`、`NewHealthBar`、`NewHotbar`、`NewCooldownMask`、`NewInventory`、`NewLootWindow`、`NewMinimap`、`NewNameplate`、`NewQuestTracker`、`NewRadialMenu`、`NewScoreboard`、`NewSkillTree`、`NewTeamFrame`、`NewTargetFrame`、`NewWindow`、`NewWindowManager`。共 25 个，与扫描一致。`ui/chat` 在表 3，不并进游戏聊天框。

## 里程碑

### 里程碑 1 — 原批次 1 和 2

依赖：无。

- [ ] **批次 1。** `core` 增加树序 `FocusNext`/`FocusPrev`，仍只用 `SetFocused` 写焦点。`widget/keys.go` 提供 Enter/Space 和方向键。接到表 1 的 1–13、18、19、29，以及表 1 附的 `MenuBar`。Dialog 读取已有 `closeOnEscKeydown`。
- [ ] **批次 2。** `widget/hit.go` 为打开期间的格子创建元素，关闭和 `Destroy` 时删除。接到表 1 的 16、17、20–25，以及表 1 附的 `Cascader`、`TreeSelect`、`Popconfirm`、`Calendar` 已含在 23。

验收：

```text
go test -count=1 -timeout 180s ./core/ ./widget/ ./widget/game/
```

退出码 0。覆盖：按钮 Space、单选方向键、滑块 step、对话框 Escape 的开/关、Select 的 Enter、日期格子的 `Day()`、颜色预设的 `OnChange`、级联两级路径。

### 里程碑 2 — 原批次 3、4、5

依赖：里程碑 1。

- [ ] **批次 3。** 表 1 的 6（范围值）、8、9（过滤与多选）、27、28、30、31、32、34、35、36，以及表 1 附的 `AutoComplete`、`Upload`、`ImageViewer`、`Affix`。`Swiper.Tick` 与 `Notification.Tick` 由应用调用，不加进 `Widget` 接口。`Value()` 在范围滑块上仍返回下限。
- [ ] **批次 4。** 表 1 的 26、27 的可视行窗口（万行 `Draw` 的矩形命令少于 200），表 1 附的 `VirtualGrid` 滚轮和 `Grid` 断点。
- [ ] **批次 5。** 表 1 的交互控件和表 1b 的 15 个实现 `Accessible`。`accessibility.Collect` 接收调用方给出的切片。不改 Win32 UIA 桩的返回值。

验收：

```text
go test -count=1 -timeout 180s ./accessibility/ ./widget/ ./widget/game/ ./platform/linux/
go build -o /dev/null .
```

退出码都是 0。

### 里程碑 3 — 表 2 的 20 个新控件

依赖：里程碑 2 的焦点、命中、虚拟网格和无障碍收集。

- [ ] **3a 叶子：** Meter、HoldToConfirm、ToggleButton、ImageCompare、OtpInput、Scrollbar。
- [ ] **3b 组合：** ToggleGroup、Toolbar、HoverCard、Combobox、SuggestList、Command、CommandDialog。
- [ ] **3c：** EmojiPicker 与自有目录。
- [ ] **3d 外壳：** Nav、FormSection/FormRow/FormField、PlainShell、Sidebar、SidebarShell、TabShell、Settings。

每个新类型带一个无头测试，断言上表「行为验收」那一列。HTML 标签用小写类型名，不改已有标签。

验收：`go test -count=1 -timeout 180s ./widget/` 退出码 0，且表 2 每一行至少有一个测试函数名含该类型。

### 里程碑 4 — 表 3 的 19 个子系统

依赖：里程碑 3 完成后再按表 3 的序号做。1→2 和 1→5 可以分叉，但 5 依赖 1，10 依赖 9，14 依赖里程碑 2 的 Dock 测试。15、16、17、12、8 在产品默认下停在该行写明的边界，并有测试证明边界（例如终端测试只喂字节流、公式测试使用调用方字体、WebView 测试使用假宿主）。

每一行合并前跑：

```text
go test -count=1 -timeout 300s ./widget/ ./accessibility/ ./core/ ./layout/
go build -o /dev/null .
```

`go test ./...` 在没有 Vulkan 时仍会因三个视觉包失败。总结里单独报告这个退出码，不把它当成回归。

## 产品默认

确认「唯一范围」时，这些默认一并生效。要改的话在同一次回复里指出编号。

| # | 默认 |
| --- | --- |
| D1 | Emoji 用自有短标签，不引入 emojibase/CLDR 数据 |
| D2 | 公式不嵌入 NewCMMath，不把 LPPL 带进仓库 |
| D3 | 终端不复制 libghostty 二进制；完整 VT 另说才做 |
| D4 | 图表、流程图、编辑器、富文本、示意图走干净重写，不新增 chroma、goldmark、limu、untie |
| D5 | WebView、Video、取色器先做接口和假实现，不在当前无 SDK 的环境里接真宿主 |
| D6 | 中文默认文案（「确定」「请选择」）保持 |
| D7 | 不改 gin 依赖，不加 LICENSE/CI，除非另一次明确要求 |
| D8 | 日期、颜色、级联的点击改为真正选中 |

## 实现时的提交

确认之后再写代码。一批或表 3 的一行一个英文提交，例如 `Add keyboard activation for buttons and dialogs`。不要 `Co-authored-by`。本设计文件在确认前不要提交。
