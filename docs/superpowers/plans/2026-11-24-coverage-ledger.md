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
| GPU 着色像素 | 只有软件后端逐像素执行。OpenGL、Vulkan、Metal、DX9、DX11 的 `Submit` 不把错误返回给调用方 |
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

下面 10 项在修前对 `fc2d56fd` 是红的。修后走 embedded/headless 的 `HandleEvent`/`HitTest`，或 godot `InjectMouseClick`。写这段时里程碑 1–2 仍是 partial，表 2/3 尚未开始。当前范围见文末交接。

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

## 表 2（20）

里程碑 3。行为走 `core.HostInput.Handle`（与 embedded、headless、godot、app 相同的指针和键盘路径）或 `HitTest`。`Tick` 由帧宿主调用，和 `Message.Tick` 一样，测试里直接调用。HTML 标签在 `TestHTMLMilestone3Meter`。Builder 增加了 `Meter` 和 `ToggleButton`，其余用已有的 `Custom`。没有操作系统读屏。

| 行 | 类型 | 状态 | 测试 |
| --- | --- | --- | --- |
| 1 | Command / CommandDialog | host-verified | `TestCommandHostFilterAndRun`、`TestCommandDialogHostEscapeRestoresFocus`。`TestCommandRootClickRunsFilteredRow`：查询条只有 28px 时，选项是浮层，根上点击和软件光栅都能看到第一项。窗口最多 6 行，方向键滚动选中 |
| 2 | OtpInput | host-verified | `TestOtpInputHostTypeBackspacePaste`。`TestOtpCompleteFiresOnce`：满长只完成一次，多余字符不再次回调，删掉再填会再完成 |
| 3 | HoverCard | host-verified | `TestHoverCardHostDelayFocusEscape` 覆盖延迟和已聚焦时的 Escape。`TestHoverCardEscapeWithoutStealingFocus` 覆盖悬停打开后、焦点仍在输入框时的 Escape 和外部点击。悬停层走 `PushTransient`，不移动焦点 |
| 4 | HoldToConfirm | host-verified | `TestHoldToConfirmHostHoldAndRelease`。提前松开不触发；禁用不再累计 |
| 5 | Meter | host-verified | `TestMeterHostValueAndMarks`。类型不是 `Progress` |
| 6 | Toolbar | host-verified | `TestToolbarHostArrowsDoNotActivate` |
| 7 | ToggleGroup | host-verified | `TestToggleGroupHostSingleAndMulti` |
| 8 | ToggleButton | host-verified | `TestToggleButtonHostSpace` |
| 9 | ImageCompare | host-verified | `TestImageCompareHostDragAndKeys` 覆盖分割位置。`TestImageCompareClipsTwoImages` 用两张 `Image` 纹理做裁剪对比，没有纹理时仍只是色块 |
| 10 | SuggestList | host-verified | `TestSuggestListHostDoesNotOwnInput` |
| 11 | EmojiPicker | host-verified | `TestEmojiPickerHostSearchAndVirtualCells`。目录是手写 `emoji_catalog.go`，只画可见格 |
| 12 | Combobox | host-verified | `TestComboboxHostEditAndChoose`。不改 `Select` / `AutoComplete` |
| 13 | Scrollbar | host-verified | `TestScrollbarHostDrag` |
| 14 | Nav / NavHeader | host-verified | `TestNavHostPushPopTitle` |
| 15 | FormSection / FormRow / FormField | host-verified | `TestFormSectionHostRowsDoNotReplaceValidate` 确认校验仍归 `Form.Validate`。`TestFormRowLaysOutFields` 确认一行两个字段有宽度且不重叠 |
| 16 | PlainShell | host-verified | `TestPlainShellHostSlots` 检查槽。`TestPlainAndSidebarShellLayout` 把内容排在 36px 顶栏下面，并能命中 |
| 17 | Sidebar | host-verified | `TestSidebarHostGroupsAndArrows` |
| 18 | SidebarShell | host-verified | `TestSidebarShellHostSlots`。`TestPlainAndSidebarShellLayout` 左栏右内容，点击内容按钮能聚焦 |
| 19 | TabShell | host-verified | `TestTabShellHostSwitch` |
| 20 | Settings / SettingsDialog | host-verified | `TestSettingsHostCategoryAndDialog`。`TestSettingsCategoryOpensLiveRows`：分类拥有 `FormSection` 行和输入框，指针与方向键切换后控件可输入。对话框使用层和作用域 |

## 浮层排序

绘制顺序用 `render.OrderedForPaint`：每条绘制记住发出时的绝对剪刀，再按 ZOrder 稳定排序（图层基数已经在 ZOrder 里），最后按剪刀变化重建裁剪命令。软件光栅覆盖了邻居裁剪、嵌套相交，以及更高图层先画仍然盖住更低图层。OpenGL、Vulkan、Metal、DX9、DX11 的提交调用同一个函数；这些原生后端没有帧图像。`StableZOrder` 仍是不含裁剪的纯排序。`TestOrderedForPaintKeepsClipScopes`、`TestOrderedForPaintOrdersLayersAcrossClips`、`TestCommandBufferNestedClipFollowsDraw`、`TestHigherLayerCoversEarlierDraw`。

HTML 的 `<nav>` 仍是通用容器。页面栈用 `<navstack>`。`TestHTMLNavKeepsChildrenAndFollowingMeter` 检查嵌套段落、后续 meter/togglebutton，以及带子节点的 `formsection` 不会吃掉下一个兄弟。

## 表 3（19）

19 行都在各自写出的边界内独立通过。汇总见文末交接。

| 行 | 状态 | 证据 |
| --- | --- | --- |
| 1 Camera | approved-bounded | 审查第 10 轮通过 `8e6efffa`。平移、缩放、命中、视口裁剪、图层顺序。`TestCameraHostPanZoomAndHit`、`TestCameraClipsViewportAndAnchorsGrid`、`TestCameraClipDoesNotCoverOtherControls`、`TestHigherLayerCoversEarlierDraw`。一个平面和点标记，不是场景图 |
| 2 plot | approved-bounded | 同一轮通过。柱、折线中点、面积、散点只命中点、扇形、坐标轴、桑基直线加权边、单层树图切片。`TestPlotHostBarsAxesAndPie`、`TestPlotLinesAreaSankeyPie`、`TestPlotScatterHitIsThePoint`。树图不是 squarified。桑基不是丝带 |
| 3 DataTable | host-verified，边界写明 | 列先按序号分成左固定、中间、右固定，每列只画自己的区域，再和表头或表体、外框相交。`TestDataTableFrozenColumnsOwnTheirRegions`：宽 120 的中间列在 `scrollX` 40 时几何上横跨三段，左右固定表头仍是绿色，中间字形的剪刀停在中间视口。左右单独固定、横纵同时滚动、拖宽和按序号移动都在。没有列组，也没有单元格里的控件 |
| 4 TextView | host-verified，边界写明 | 标题、列表、行内代码、围栏代码和链接标签都是字形，不是色条。围栏代码可以有底色。`TestTextViewPaintsGlyphs` 检查这些字符串进了文本命令，长行换到下一行，两个链接的命中框分开，换文档后旧标题消失。`TestTextViewMarkdownLinksAndCode` 仍检查宿主点击。没有表格、图片、强调语法，也不使用 goldmark |
| 5 flow | host-verified，边界写明 | 节点名字是字形。`NodeScreen` 连宽高一起乘缩放。`TestFlowScalesNodesAndPaintsNames` 在平移后检查 2 倍和 0.5 倍：绘制矩形、内部命中、外部拒绝、宿主点中、边接到缩放后的节点边缘、超出视口被裁掉。`TestFlowLayersUseCamera` 仍覆盖分层、平移、删除。自环不改层。只在环里的节点共用一个附加层，不是完整环布局 |
| 6 filepicker | host-verified，边界写明 | `TestFilePickerNavigatesInjectedList`：目录名和路径是字形，Enter 进入目录，Backspace 回到父级，文件 Enter 只回调路径。列举失败画出错误文本。窗口外的名字不画，滚动后的行不盖住路径条。数据来自注入的 `List`，控件里没有操作系统读写。没有新建、删除或重命名 |
| 7 commitgraph | host-verified，边界写明 | `TestCommitGraphLanesAndRows`：第一父提交留在原泳道，旁支另开一列。主题是字形。视口下面的提交不画，滚下来才出现。点击和方向键改选中。调用方提供提交，不是 git 命令 |
| 8 diagram | host-verified，边界写明 | `TestDiagramBoxArrowLayout`：`box` 和 `->` 排成从左到右的层，标签是字形。`subgraph` 被忽略。视口外的节点不画，滚到视口里才画。不是 Mermaid，不依赖 limu |
| 9 editor | host-verified，边界写明 | `TestEditorRuneCaretAndGlyphs` 覆盖符文光标、退格「你」、字形和视口滚动。`TestEditorDeleteSelectionPreservesNewlines`：选中中间行 `cd` 后退格得到 `ab\n\nef`。`TestEditorBoundaryEditsMatchRuneModel` 用符文切片对照：文档开头退格、文末 Delete、空粘贴都不改文本也不发 `OnChange`。首行、末行、反向选区、跨行、停在列 0、空选区、替换成 `X`、行首合并、Delete、Enter 和粘贴各只触发一次 `OnChange`。没有 chroma，没有自动换行 |
| 10 diff | host-verified，边界写明 | `TestDiffViewSideBySide`：左右列各自裁剪，`old`/`new` 的字形不越过中线。`n` 从相同行跳到替换再跳到新增。点击用 `TextPos` 列偏移。视口外的变更不画。这是行级 LCS，不是词级差异，也不是三方合并 |
| 11 richeditor | host-verified，边界写明 | `TestRichEditorMarksAndGlyphs`、`TestRichEditorMarkTable`。文档是 UTF-8 加符文区间标记。开头退格、文末 Delete、空粘贴不改文本也不发 `OnChange`。选中 `ab\ncd\nef` 里的 `cd` 得到 `ab\n\nef`，未选中的换行还在。插入和删除会平移、裁短或丢掉标记。粗体是同一字面再画一次，代码有底色但仍是字形，链接是主色字形。点击用测量宽度，视口外的行不画。没有图片、表格、HTML 粘贴，也不是 ProseMirror/Tiptap。不自动换行。`RichEditor.Draw` 和 `Editor.Draw` 一样不读树上的可见性：父级决定要不要调用 `Draw`（`TabShell` 只画当前页，`LayerManager` 跳过不可见层），命中和焦点走 `ChainAllows`。这是现有约定，没有改 `Draw` |
| 12 formula | host-verified，边界写明 | `TestFormulaLayoutGlyphsAndHit`。语法是四则、`frac`、`sup`、`sub`、`sqrt` 和括号。分子在分母上面，上标抬高，下标降低，根号内容是字形。点击分母得到 `den`。宽公式在视口外的原子不画，滚进来才画。源字符串本身不作为一条文本画出。字体用调用方的 `TextDrawer`。没有积分、矩阵或完整 TeX，也不嵌入 NewCMMath |
| 13 chat | host-verified，边界写明 | `TestChatAuthorAndBodyStayApart`：作者和正文的字形框不相交，作者在上，别人靠左，自己的多行靠右。长名字按测量宽度换行。视口外的消息不画。点正文用同一套行几何命中。列表和输入框仍分开裁剪。`TestChatReentrantSendKeepsSnapshot`：回调里再按 `Q` 或再按 Enter，记录仍是按下 Enter 前的 `A`，只调用一次 `Send`；成功后草稿只剩回调期间的新输入。发送失败保留原草稿加新输入。`Append` 不调用发送。空 Enter 不发送。没有网络 |
| 14 Dock 序列化 | host-verified，边界写明 | `TestDockRestoreGeometryAndIDs`。序列化的是现有 `DockLayout` 的平铺边缘：每侧一块、中间页签编号、浮动矩形和缩放。JSON 恢复编号和几何，同编号的内容指针还在。重复编号、空编号、负数尺寸、未知当前页签、缩放不大于 0、坏 JSON 都拒绝，当前布局不动。标题是字形。这不是嵌套分割树。`Draw` 不画 `Content`。点击页签不改变当前页签 |
| 15 terminal | host-verified，边界写明 | `TestTerminalBytesKeysAndScroll`。`Write` 追加 UTF-8，半个符文先留着，拼齐再显示。`ESC [2J` 留在文本里，不清屏。按键只交给 `OnBytes`，不写进画面，也不启动进程。视口外的行不画。不是 VT，没有 libghostty，没有 PTY |
| 16 WebView | platform-interface-only | `TestFakeWebViewHistoryAndLifecycle`。`FakeWebViewHost.Open` 只增加计数。`LoadURL` 记历史，空地址失败且不改当前页。后退、前进、截断前进栈、刷新和 `<title>` 读取都在。`Eval` 只查宿主给的脚本表，未知脚本不改页面。关闭后加载和再次关闭返回 `ErrClosed`。没有系统 WebView，也不发起网络请求 |
| 17 Video | platform-interface-only | `TestFakeVideoPlaybackBounds`。空源或时长不大于 0 不会打开。播放时 `Tick` 推进位置，暂停后不动。越界 `Seek` 和音量不改原值。放到结尾只触发一次 `OnEnded` 并暂停。关闭后再播放返回 `ErrClosed`。没有解码器 |
| 18 Shader | 软件逐像素路径已独立通过 | 审查第 23 轮通过 `c39708b6af693b796c1376b5a5ff11a3be7bc906`（tree `d3fd81db78ff5ad700bab563269ea2c002a75e46`）。`DrawShader` 把 `core.Shader` 放进命令缓冲，软件后端对裁剪后的每个像素调用 `Eval`。`TestShaderCommandPaintsVaryingPixels` 不先求值再画纯色矩形。左右像素的 `uv.x` 不同，`gain` 在两侧相同。嵌套剪刀只留下内层，空剪刀不落笔，更高图层盖住着色结果，半透明和底下的矩形合成。右半失败的程序一个像素都不写。下一帧换程序后旧颜色消失。OpenGL、Vulkan、Metal、DX9、DX11 的 `Submit` 没有把 `ErrShaderGPUUnsupported` 返回给调用方：命令循环执行 `_ = render.ShaderFallback()`，丢掉这个错误，跳过着色命令，然后继续后面的命令。Vulkan 在跳过前先 `flushBatch`。这五个后端不画出着色像素，也没有帧图像 |
| 19 ColorEyeDropper | platform-interface-only | `TestFakeEyeDropperDoesNotSample`。未授权时 `Request` 返回 `ErrPermission`，没有颜色。授权后停在拾取中，直到宿主 `Deliver` 一个颜色。重复请求不替换进行中的拾取。取消后没有颜色，再 `Deliver` 失败。没有屏幕采样 |

悬停层和模态：`TestModalEscapeSkipsCoveredHover`。被更高模态盖住的 transient 不再吃掉 Escape。没有模态时，悬停打开的卡片仍能在不持有焦点的情况下被 Escape 关掉。

## 交接

代码证据钉在 `c39708b6af693b796c1376b5a5ff11a3be7bc906`，tree `d3fd81db78ff5ad700bab563269ea2c002a75e46`。其后的提交只改 `docs/`。相对该提交，代码文件没有差异。

里程碑 1–2 是 partial。`Form.Reset` 只清 Input、TextArea、Checkbox、Select、Slider 这五类。树拖拽只交换同一父节点。硬件 `GamepadButtonDown` 没有接到 app、embedded、godot。无障碍名字和 `accessibility.Collect` 在进程内，不是 Win32 UI Automation，也不会自动跑。

里程碑 3 的表 2 共 20 行是 host-verified。没有操作系统读屏。ImageCompare 没有纹理时仍是色块。

里程碑 4 的表 3 共 19 行在已批准边界内独立通过：

- Camera 是一个平面。plot 的树图是单层切片，桑基是直线加权边。
- diagram 只有 `box` 和 `->`。editor 没有 chroma，也没有自动换行。diff 是行级 LCS。
- richeditor 没有图片、表格或 HTML 粘贴；`Draw` 不读树上的可见性。formula 是小语法，不嵌入 NewCMMath。
- Dock 是现有平铺边缘的缩放和序列化。不是嵌套分割树。`Draw` 不画 `Content`。点击页签不切换。
- terminal 是 UTF-8 字节流和按键路由，不是 VT。
- WebView、Video、ColorEyeDropper 是接口和假宿主。
- Shader 的软件逐像素路径在上面的提交通过。五个 GPU 后端不画着色像素。

没有接入：系统 WebView、视频解码器、屏幕取色、GPU 着色像素、完整 VT、Win32 UI Automation、Linux IME、操作系统文件对话框。

在该代码上重跑：

```text
go test -count=1 -timeout 180s ./widget ./core ./render ./accessibility ./layout .
go test -race -count=1 -timeout 300s ./core ./widget ./accessibility ./render .
go build -o /dev/null .
```

三条退出码都是 0。

```text
go test -count=1 -timeout 600s ./...
```

退出码 1。失败只有 16 个 `VK_ERROR_UNKNOWN`：`cmd/demo` 11 个（`TestVisualEmoji` 和 10 个 `TestVisual*`），`cmd/feed` 2 个，`cmd/game` 3 个。没有原生 GPU 帧。
