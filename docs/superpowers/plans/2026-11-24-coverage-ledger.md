# 覆盖账本

对照 `2026-11-24-approved-nsgo-gap.md` 和独立审查 `bc-1b418616-ee44-5cfe-bc4d-bacde8db4310`。状态：`host-verified`、`partial`、`pending`、`platform-interface-only`。只走控件方法、不经过宿主 `HandleEvent`/`HitTest` 的测试不算 `host-verified`。无障碍名字在进程内，不表示系统读屏。`accessibility.Collect` 不会自动跑。

## 平台接入（不算完成）

| 项 | 状态 |
| --- | --- |
| Win32 UI Automation | platform-interface-only |
| Linux Fcitx/IBus | platform-interface-only |
| 操作系统文件对话框 | platform-interface-only |
| WebView / Video 真宿主 | platform-interface-only |
| 屏幕取色 | platform-interface-only |
| 终端 VT | platform-interface-only |
| Vulkan 视觉包 | 本机 `VK_ERROR_UNKNOWN`，不能当作绘制集成已验证 |

## 审查项

| 项 | 结果 | 证据 |
| --- | --- | --- |
| P0 宿主键盘/点击焦点 | 已修，host-verified | `TestEmbeddedTabMovesBetweenButtons`：修前点击焦点为 0。`core.HostInput` 接到 embedded、headless、godot、app |
| P0 隐藏按钮仍响应 Space | 已修，host-verified | `TestEmbeddedHiddenDisabledLoadingAndTagInput`：修前 hidden button activated |
| P0 TagInput 不可聚焦 | 已修，host-verified | 同一测试，宿主点击后 Enter 生成标签 |
| P0 弹层在控件外点不中 | 已修，host-verified | `TestHeadlessPopupHitsOutsideWidget`：修前 hour 0。时间、范围日期、级联、树选择、Alpha、自动完成都走 Headless `HandleEvent` |
| P0 对话框 Enter、关闭后仍可点、Select Escape 和残留遮罩 | 已修，host-verified | `TestEmbeddedDialogLifecycleAndSelectEscape` |
| P0 嵌套作用域被错误弹出 | 已修 | `PopFocusScopeOwned`。该测试先关上层再确认下层 |
| P1 禁用/加载不进 Tab | 已修，host-verified | 嵌入测试跳过 loading 按钮和禁用链接 |
| P1 范围滑块指针只改 `value` | 已修 | `TestRangePointerRadioResetDestroyAndSelectEmpty` 用 `MouseDown` 断言 `RangeValues`。这是 dispatcher，不是宿主 |
| P1 `RadioGroup.SetValue` 不同步 Tab | 已修 | 同上，直接查 `IsFocusable` |
| P1 销毁日期控件留下命中格 | 已修 | 同上，`Destroy` 后元素数回到创建前 |
| P1 Select 无匹配仍画全部选项 | 已修 | 同上，点空结果不改变值 |
| P1 表格滚动后 hover 行号 | 已修 | 同上，`MouseMove` 后 `HoveredRow()==5` |
| P1 手柄忽略模态范围 | 已修 | 同上，`Navigate` 不能聚焦对话框外按钮。宿主未把空焦点方向键打到手柄上的用例另见 fallback |
| P1 `SetFocused` 不发 Focus | 已修，host-verified | `TestEmbeddedDrawerCaptureAndTooltipFocus`：点击后 tooltip 出现，没有手工派发 Focus |
| P1 拖出抽屉失败 | 已修，host-verified | 同一测试，`MouseMove` 到控件外仍关闭 |
| P2 List 只用局部 `e.Y` | 已修 | dispatcher 点击只带 `GlobalY` 时选中第二项 |
| P2 `Form.Reset` 只清输入框 | 部分 | 现在也清 Checkbox、Select、Slider。没有覆盖每一个控件类型 |
| P2 树跨父节点拖拽 | 不支持 | `swapIn` 只交换同一父节点下的节点。根节点之间的交换有测试。跨父节点没有实现 |
| P2 Enter 留下旧选中 | 已修 | Enter 会清掉其他 `Selected` |
| P2 关闭的 Popover 仍可聚焦 | 已修 | 构造时不可聚焦，`Open` 才取焦点 |

## 表 1 / 1 附 / 1b

没有整表标成完成。只有上面写了 host-verified 的路径算宿主集成已测。其余行为仍是 `partial`：单测或 dispatcher 通过，但没有再走一遍 app/embedded/headless/godot 的 `HandleEvent`。

无障碍：Button、Switch、Slider、Input、TextArea 和表 1b 的角色断言都在。Slider 和 Image 在没有单独标签时名字仍是 `slider` / `image`。这不是系统读屏。

## 表 2（20）与表 3（19）

全部 pending。里程碑 3 还没开始。
