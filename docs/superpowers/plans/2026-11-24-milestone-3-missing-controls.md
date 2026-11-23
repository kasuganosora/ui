# 里程碑 3 计划（尚未执行）

> 已批准，但本增量不实现。完成后才能做里程碑 4。细节行为以 `2026-11-24-approved-nsgo-gap.md` 表 2 的 20 行为准，这里不重复展开每个控件的内部结构。

**Goal:** 新增 20 个通用控件，全部是 `widget.Widget`，并带无头测试和 HTML 标签。

**Architecture:** 复用里程碑 1–2 的焦点、命中格、`Tick` 和 `AccessibleProps`。不新增第三方依赖。Emoji 只用手写短标签，不复制 `emoji.bin`。

**Tech Stack:** 与目标仓库现有栈相同。

## Global Constraints

- 不改已有 `NewX` 签名。
- 中文默认文案保持。
- 无 Co-authored-by。

## 任务

- [ ] 3a：`Meter`、`HoldToConfirm`、`ToggleButton`、`ImageCompare`、`OtpInput`、`Scrollbar`。每个类型一个失败测试再实现。验收见已批准设计表 2 第 5、4、8、9、2、13 行。
- [ ] 3b：`ToggleGroup`、`Toolbar`、`HoverCard`、`Combobox`、`SuggestList`、`Command`、`CommandDialog`。依赖 3a 的切换按钮和里程碑 1 的弹出层。
- [ ] 3c：`EmojiPicker` 与 `emoji_catalog.go` 手写目录。依赖虚拟网格。
- [ ] 3d：`Nav`、`FormSection`/`FormRow`/`FormField`、`PlainShell`、`Sidebar`、`SidebarShell`、`TabShell`、`Settings`。

验收：`go test -count=1 -timeout 180s ./widget/` 为 0，且表 2 每一行都有含类型名的测试。
