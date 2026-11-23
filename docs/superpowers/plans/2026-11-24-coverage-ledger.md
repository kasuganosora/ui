# 覆盖账本

对照 `2026-11-24-approved-nsgo-gap.md`。状态只有 `implemented+tested`、`pending`、`platform-interface-only`。没有把假宿主或未接入的原生后端记成完成。

更新日期：2026-11-24。里程碑 1–2 的行在本增量补齐后，下面标成 `implemented+tested` 的每一行都有直接测试，包含原先写在行内的子行为。

## 平台接入（单独列出，不计入完成）

| 项 | 状态 | 原因 |
| --- | --- | --- |
| Win32 UI Automation COM | platform-interface-only | `HandleGetObject` 仍返回未处理 |
| Linux Fcitx/IBus | platform-interface-only | `platform/linux/ime.go` 仍是桩 |
| 操作系统文件对话框 | platform-interface-only | 只有应用注入的 `OpenPicker` 被测试 |
| WebView / Video 真宿主 | platform-interface-only | 没有 SDK，也没有假宿主充完成 |
| 屏幕取色 | platform-interface-only | 没有操作系统授权路径 |
| 终端 VT / libghostty | platform-interface-only | 不拷贝二进制 |

视觉截图测试在本机没有 Vulkan，不能当作绘制集成已验证。

## 表 1（36）

全部 `implemented+tested`。

| # | 能力 | 直接覆盖的测试 |
| --- | --- | --- |
| 1–6 | Button、Link、Checkbox、Radio、Switch、Slider | `TestButtonSpaceActivates`、`TestRadioArrowsAndSingleTabStop`、`TestSliderArrowsAndRange`、`TestAccessibilityMetadata` |
| 7 | Input / TextArea 无障碍 | `TestInputAndTextAreaAccessibility` |
| 8–13 | InputNumber、Select、Tabs、Menu、ContextMenu、Pagination | `TestCheckboxLinkTabsMenuPaginationRate`、`TestSelectFilterAndMultiple` |
| 14–15 | Accordion 与独立 Collapse | `TestCollapseAccordionKeyboard` |
| 16–20 | Tooltip、Popover、Dialog、MessageBox、Drawer | `TestTreeTooltipPopoverDrawer`、`TestDialogEscapeAndFocusRestore`、`TestRootDispatchFocusScopeAndNestedHit` |
| 21–24 | DatePicker、DateRange、TimePicker、Calendar、ColorPicker 预设/Alpha/最近色 | `TestDatePickerSelectsDayAndCleansHitTargets`、`TestTimePickerHit`、`TestColorAlphaAndRecent`、`TestUploadAffixImageViewerDockDrag` |
| 25–26 | Tree 键盘、懒加载、拖拽排序；List 键盘与 VirtualList 滚轮 | `TestTreeLazyAndDragSort`、`TestListKeyboard`、`TestClockAndScrollBreakpoints` |
| 27–33 | Table、Swiper 时钟与拖拽、Rate、TagInput、Form、Notification 与 Message 超时、ScrollView 翻页 | `TestTableSortPageAndWindow`、`TestSwiperDrag`、`TestMessageTimeout`、`TestScrollViewKeyboard`、`TestClockAndScrollBreakpoints` |
| 34–36 | Splitter、DockLayout、DragDrop | `TestUploadAffixImageViewerDockDrag` |

隐藏焦点、销毁作用域后 Tab 能离开对话框：`TestTabSkipsHidden`、`TestDestroyOpenScopeReleasesTab`。嵌套日期格由 `HitTest` 再 `Dispatch`，不是直接调用控件方法：`TestRootDispatchFocusScopeAndNestedHit`。

## 表 1 附（11）

全部 `implemented+tested`：MenuBar、Grid 断点、Upload `OpenPicker`、Popconfirm、Cascader、TreeSelect、ImageViewer、Affix、AutoComplete、`TestVirtualGridWheel`、`TestGamepadUsesTreeFocus`。

## 表 1b（15）

全部 `implemented+tested`。`TestRemainingAccessibilityTypes` 逐个断言 Avatar、AvatarGroup、Badge、Breadcrumb、Image、Progress、Skeleton、Space、Loading、Steps、Text、Card、Icon、Canvas、Portal 的角色和关键名字或数值。Button、Switch、Slider 仍由 `TestAccessibilityMetadata` 覆盖。

## 表 2（20）

全部 pending。计划：`2026-11-24-milestone-3-missing-controls.md`。

## 表 3（19）

全部 pending。计划：`2026-11-24-milestone-4-subsystems.md`。第 15、16、17、19 行在真宿主通过前只记 platform-interface-only。

## 表 4（15）

排除，不实现。

## 表 5

保持，不在本缺口内。
