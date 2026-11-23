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

## 第二次审查（fc2d56fd 仍阻断基础）

下面 10 项在修前对 `fc2d56fd` 是红的。修后走 embedded/headless 的 `HandleEvent`/`HitTest`，或 godot `InjectMouseClick`。里程碑 1–2 仍是 partial，表 2/3 仍 pending。

| # | 结果 | 测试 |
| --- | --- | --- |
| 1 乱序关闭后焦点留在上层对话框 | 已修，host-verified | `TestOutOfOrderDialogCloseKeepsUpperScope`。`Dialog.Close` / `MessageBox.Close` 只在焦点确实在被关层内时恢复，且 prior 必须仍在当前作用域里、可见且启用 |
| 2 第二次 Escape 关掉模态 | 已修，host-verified | `TestSecondEscapeClosesDialogAfterSelect`。`bindArrows` 只有真正处理了 Escape 才 `StopPropagation` |
| 3 Select 选项点击和按下/抬起 | 已修，host-verified | `TestSelectOptionClickAndEmptyBackdrop` 前半。只在自身 target 阶段切换；`Draw` 之后直接 `MouseClick` 和 `MouseDown`+`MouseUp` 都选中第一项并关闭 |
| 4 空结果点击只关一次 | 已修，host-verified | 同一测试后半。遮罩是 Select 的子节点，点击停止冒泡，不会再把下拉打开 |
| 5 后开的模态挡住更早的日期/选择浮层 | 已修，host-verified | 打开时 `PushLayer`，不依赖插入顺序。`TestEarlyDialogOpenedLateBlocksDate` 先建对话框再打开；`TestDialogOpenOrderBeatsConstructionOrder` 反转两个对话框的打开顺序；`TestSelectBackdropShieldedByLaterDialog`；`TestInnerDateStaysAboveOwnDialog` 内部浮层仍在自己的模态之上。关闭后回到上一层。`TestLaterModalBlocksEarlierOverlay` 只覆盖“后插入”这一种顺序，不能单独当证据。绘制和命中共用层号：`TestLaterOpenedDialogPaintsAboveEarlierDate`、`TestInnerDatePaintsAboveOwnDialog`，软件光栅 `TestSoftwarePaintsOpenedDialogAboveDate` |
| 6 按下+抬起只点一次，抬到别处不点 | 已修，host-verified | 约定在生产者：`InjectMouseClick` 只发 Down+Up，宿主对同一目标合成一次 Click。单独的 `MouseClick` 总会送达，包括换目标和跨帧。`TestInjectMouseClickActivatesOnce`、`TestStandaloneClickAfterDownUp`、`TestDownUpFrameThenStandaloneClick`。A 上按下、B 上抬起不产生 Click |
| 7 取消/失焦结束拖拽 | 已修，host-verified | `TestTouchCancelAndWindowBlurClearDrawerDrag`。`TouchCancel` 和 `WindowBlur` 对捕获目标发 `MouseUp` 且不合成 Click |
| 8 禁用祖先挡住子节点 | 已修，host-verified | `TestEmbeddedDisabledAncestorBlocksChild`，dispatcher 对照是 `TestDisabledAncestorBlocksChild`。命中、Tab、程序 `SetFocused`、鼠标和键盘派发都看整条祖先链 |
| 9 焦点重入最终只通知 B 一次 | 已修 | `TestFocusReentryDeliversSecondFocus`。`SetFocused` 在通知期间把后续焦点排队，上限 32 步，不再直接丢掉 |
| 10 手柄第一次方向键停在当前项 | 已修，host-verified | `TestEmbeddedGamepadFallbackFocusesCurrent`，对照 `TestGamepadFirstDirectionFocusesCurrent`。硬件 `GamepadButtonDown` 仍没有接到 app/embedded/godot，这是接受的平台边界；方向键 fallback 走 `SetKeyFallback` + `RouteKey` |

空白处点击：点到没有可聚焦祖先、而且不在当前焦点控件内部时，`HostInput` 清掉焦点。点在已聚焦输入框内部不会清，随后的 IME 仍经 `RouteKey` 送给它。`TestBlankClickClearsInputButNotItsOwnPress`。

`Form.Reset` 覆盖 Input、TextArea、Checkbox、Select、Slider 这五类，其余类型仍未接。树拖拽仍只交换同一父节点。

## 表 1 / 1 附 / 1b

没有整表标成完成。只有上面写了 host-verified 的路径算宿主集成已测。其余行为仍是 `partial`：单测或 dispatcher 通过，但没有再走一遍 app/embedded/headless/godot 的 `HandleEvent`。

无障碍：Button、Switch、Slider、Input、TextArea 和表 1b 的角色断言都在。Slider 和 Image 在没有单独标签时名字仍是 `slider` / `image`。这不是系统读屏。

## 表 2（20）与表 3（19）

全部 pending。里程碑 3 还没开始。
