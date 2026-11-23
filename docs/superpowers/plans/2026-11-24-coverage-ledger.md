# 覆盖账本

对照 `2026-11-24-approved-nsgo-gap.md`。状态只有 `implemented+tested`、`pending`、`platform-interface-only`。假实现和未接入的原生后端不算完成。

更新日期：2026-11-24。本账本对应里程碑 1–2 增量。测试见该增量的 `widget/m1m2_test.go` 与 `core/focus_test.go`。

## 平台接入（单独列出，不计入完成）

| 项 | 状态 | 原因 |
| --- | --- | --- |
| Win32 UI Automation COM | platform-interface-only | `HandleGetObject` 仍返回未处理。控件元数据不经过系统读屏 |
| Linux Fcitx/IBus | platform-interface-only | `platform/linux/ime.go` 仍是桩 |
| 操作系统文件对话框 | platform-interface-only | Upload 只接受应用注入的 `OpenPicker`，已测。系统对话框没有实现 |
| WebView / Video 真宿主 | platform-interface-only | 里程碑 4，当前无 SDK |
| 屏幕取色 | platform-interface-only | 需要操作系统授权，不在本增量 |
| 终端 VT / libghostty | platform-interface-only | 默认不拷贝二进制，里程碑 4 |

## 表 1（36）

| # | 能力 | 状态 |
| --- | --- | --- |
| 1 | Button（含图标名） | implemented+tested |
| 2 | Link | implemented+tested |
| 3 | Checkbox | implemented+tested |
| 4 | Radio | implemented+tested |
| 5 | Switch | implemented+tested（Space 走与 Checkbox 相同的 `bindActivate`；元数据有单独断言） |
| 6 | Slider 方向键与范围值 | implemented+tested |
| 7 | Input / TextArea | pending（键盘是原有行为；`AccessibleProps` 已写，本增量没有单独断言） |
| 8 | InputNumber | implemented+tested |
| 9 | Select 过滤与多选 | implemented+tested |
| 10 | Tabs | implemented+tested |
| 11 | Menu | implemented+tested |
| 12 | ContextMenu | implemented+tested |
| 13 | Pagination | implemented+tested |
| 14 | Accordion | pending |
| 15 | Collapsible | pending |
| 16 | Tooltip | implemented+tested |
| 17 | Popover | implemented+tested |
| 18 | Dialog 焦点陷阱与 Escape | implemented+tested |
| 19 | MessageBox | implemented+tested |
| 20 | Drawer 滑动关闭 | implemented+tested |
| 21 | DatePicker 与 DateRangePicker | implemented+tested |
| 22 | TimePicker | pending |
| 23 | Calendar | implemented+tested |
| 24 | ColorPicker | implemented+tested 仅预设命中。Alpha 与最近使用色 pending |
| 25 | Tree 键盘、展开、勾选 | implemented+tested。懒加载与拖拽排序 pending |
| 26 | VirtualList 滚轮 | implemented+tested。普通 List 键盘 pending |
| 27 | Table 排序、分页、行窗口 | implemented+tested |
| 28 | Swiper `Tick` | implemented+tested。指针拖拽 pending |
| 29 | Rate | implemented+tested |
| 30 | TagInput | implemented+tested |
| 31 | Form `Validate` / `Reset` | implemented+tested |
| 32 | Notification `Tick` | implemented+tested。`Message` 超时 pending |
| 33 | ScrollView 键盘翻页 | pending |
| 34 | Splitter 最小比例 | implemented+tested（夹紧是原有逻辑，本增量加了回归测试） |
| 35 | DockLayout `OnDock` | implemented+tested。缩放与序列化布局仍属表 3 |
| 36 | DragDropManager 放下回调 | implemented+tested |

按行计：implemented+tested 31 行（1–6、8–13、16–21、23–32、34–36）。pending 5 行（7、14、15、22、33）。24、25、26、28、32 的行内还写了未做的子行为，那些子行为不是整行完成。

## 表 1 附（11）

| 项 | 状态 |
| --- | --- |
| MenuBar | implemented+tested |
| Grid 断点 | implemented+tested |
| Upload `OpenPicker` | implemented+tested |
| Popconfirm Escape | implemented+tested |
| Cascader | implemented+tested |
| TreeSelect | implemented+tested |
| ImageViewer 拖拽平移 | implemented+tested |
| Affix | implemented+tested |
| AutoComplete | implemented+tested |
| VirtualGrid 滚轮 | pending（处理函数已接上，没有单独测试） |
| GamepadNavigator 走新焦点 | pending |

## 表 1b（15）

`AccessibleProps` 已实现。`TestAccessibilityMetadata` 断言了 Button、Switch、Slider。其余类型的方法返回真实字段，但没有逐个断言，因此记 pending，直到有直接断言：Avatar、AvatarGroup、Badge、Breadcrumbs、Image、Progress、Skeleton、Space、Spinner/Loading、Stepper/Steps、Text、Card、Icon、Canvas、Portal。

Button、Switch、Slider 的元数据为 implemented+tested。Progress 的方法已写，仍算在这 15 行的 pending 里，因为它没有单独断言。

## 表 2（20）

全部 pending。计划：`2026-11-24-milestone-3-missing-controls.md`。

## 表 3（19）

全部 pending。计划：`2026-11-24-milestone-4-subsystems.md`。第 15、16、17、19 行将来即使做了默认边界，在真宿主通过前也只记 platform-interface-only。

## 表 4（15）

排除，不实现。

## 表 5

保持，不在本缺口内。
