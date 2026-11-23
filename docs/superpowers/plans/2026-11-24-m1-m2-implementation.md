# 里程碑 1–2 实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task. 使用 TDD：先写失败测试，再写实现。

**Goal:** 在不改 `Widget` / `NewX` 签名、不换自绘后端、不加第三方依赖的前提下，把已批准设计的里程碑 1 和 2 做成可审查增量。

**Architecture:** `core.Tree` 增加可聚焦标记、Tab 顺序和焦点作用域。控件仍把自己的 `KeyDown` / `MouseClick` 注册到已有元素上。弹出层用短期子元素做命中，关闭时 `DestroyElement`。无障碍只填 `accessibility.Props`，不接 Win32 UIA。

**Tech Stack:** 现有 Go 模块，标准库 `time` 用于日期星期。无新依赖。

## Global Constraints

- 已有构造函数签名不变。新行为用新方法。
- 不复制 nsgo 源码或受保护资源。
- 无头测试必须在无 Vulkan 时通过。
- 平台桩（UIA、系统文件框、IME、WebView、视频、取色器、终端 VT）不实现，账本记为 platform-interface-only。
- 提交说明英文，无 Co-authored-by。

---

### 任务 1：焦点生命周期

**Files:** `core/focus.go`，`core/focus_test.go`，`core/element.go` 的 `SetFocused` / `DestroyElement`

- [ ] 写失败测试：两个可聚焦元素 Tab 循环；禁用元素被跳过；`PushFocusScope` 后 Tab 只在作用域内；销毁当前焦点后 `Focused()` 为空。
- [ ] 实现 `SetFocusable`、`FocusNext`、`FocusPrev`、`RouteKey`、作用域栈。`SetFocused` 仍是唯一写入点。

### 任务 2：键盘激活与对话框陷阱

**Files:** `widget/keys.go`，`widget/m1m2_test.go`，`button.go`，`link.go`，`checkbox.go`，`radio.go`，`switch.go`，`slider.go`，`tabs.go`，`menu.go`，`contextmenu.go`，`pagination.go`，`rate.go`，`menubar.go`，`dialog.go`，`messagebox.go`

- [ ] 测试：按钮 Space 触发且禁用/loading 不触发；单选方向键改变 `Value()` 且只有当前项可聚焦；滑块方向键按 step；对话框 Escape 在开关两种设置下的关闭；打开时焦点在确认按钮，关闭后回到先前焦点。
- [ ] 把 `KeyDown` 接到上述控件。对话框 `Open`/`Close` 推入/弹出作用域。

### 任务 3：弹出层命中与清理

**Files:** `widget/hit.go`，`datepicker.go`，`daterangepicker.go`，`calendar.go`，`colorpicker.go`，`cascader.go`，`treeselect.go`，`treew.go`，`tooltip.go`，`popover.go`，`popconfirm.go`，`drawer.go`

- [ ] 测试：2026-01-15 的格子点击使 `Day()` 为 15，随后元素数回到打开前；颜色预设触发 `OnChange`；级联两级路径；树方向键改变选中；Escape 关闭 tooltip/popover；抽屉拖动超过阈值调用 `Close`。
- [ ] 命中矩形与绘制使用同一几何。关闭或选中后销毁临时元素。

### 任务 4：已公开回调、过滤、范围、时钟

**Files:** `selectw.go`，`table.go`，`form.go`，`slider.go`，`swiper.go`，`notification.go`，`taginput.go`，`inputnumber.go`，`autocomplete.go`，`upload.go`，`imageviewer.go`，`affix.go`，`splitter.go`，`list.go`，`virtualgrid.go`，`grid.go`

- [ ] 测试：`Select` 过滤后只剩匹配项，多选不关闭并写入 `SelectedValues`；表头排序调用 `OnSortChange` 且重排行；`SetPageSize` 后 `NextPage` 调用 `OnPageChange`；未设置分页时 10000 行绘制命令少于 200，并且 `SetScrollY` 后点击命中的是窗口内的行号；`Form.Validate`/`Reset`；范围滑块 `Value()` 为下限；`Swiper.Tick(3)` 与 `Notification.Tick`；虚拟列表滚轮改变 `ScrollY`；`SetViewport(500)` 不用 `lg` 跨度，`SetViewport(1000)` 才用。
- [ ] 实现这些路径。`Widget` 接口不加 `Tick`。

### 任务 5：无障碍元数据

**Files:** `accessibility/collect.go`，`widget/a11y.go`

- [ ] 测试：`Collect` 保留顺序；按钮角色为 `RoleButton`；禁用开关带 `StateDisabled`；滑块 `ValueNow` 等于当前值。
- [ ] 表 1 的交互控件和表 1b 的 15 个类型实现 `AccessibleProps`。不改 `platform/win32/uia.go`。

### 任务 6：验证与账本

- [ ] `go test -count=1 -timeout 180s ./core/ ./widget/ ./widget/game/ ./accessibility/ ./platform/linux/`
- [ ] `go test -race -count=1 -timeout 240s ./core/ ./widget/ ./accessibility/`
- [ ] `go build -o /dev/null .`
- [ ] `go test -count=1 -timeout 300s ./...`，把 `cmd/demo`、`cmd/feed`、`cmd/game` 的 Vulkan 失败与库失败分开记录。
- [ ] 更新 `2026-11-24-coverage-ledger.md`：implemented+tested / pending / platform-interface-only。
