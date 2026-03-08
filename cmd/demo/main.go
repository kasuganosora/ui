//go:build windows

// Demo showcases the GoUI widget library in a TDesign-style component documentation layout.
// Click a sidebar item to view that component's demo.
// Run: go run ./cmd/demo
package main

import (
	"fmt"
	"os"

	ui "github.com/kasuganosora/ui"
	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/layout"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/widget"
)

// nav maps sidebar button IDs to their content section IDs.
var nav = []struct{ btn, section string }{
	{"nav-button", "sec-button"},
	{"nav-link", "sec-link"},
	{"nav-input", "sec-input"},
	{"nav-textarea", "sec-textarea"},
	{"nav-inputnumber", "sec-inputnumber"},
	{"nav-select", "sec-select"},
	{"nav-checkbox", "sec-checkbox"},
	{"nav-radio", "sec-radio"},
	{"nav-switch", "sec-switch"},
	{"nav-slider", "sec-slider"},
	{"nav-rate", "sec-rate"},
	{"nav-tag", "sec-tag"},
	{"nav-badge", "sec-badge"},
	{"nav-avatar", "sec-avatar"},
	{"nav-progress", "sec-progress"},
	{"nav-table", "sec-table"},
	{"nav-list", "sec-list"},
	{"nav-card", "sec-card"},
	{"nav-statistic", "sec-statistic"},
	{"nav-collapse", "sec-collapse"},
	{"nav-timeline", "sec-timeline"},
	{"nav-tabs", "sec-tabs"},
	{"nav-menu", "sec-menu"},
	{"nav-breadcrumb", "sec-breadcrumb"},
	{"nav-pagination", "sec-pagination"},
	{"nav-steps", "sec-steps"},
	{"nav-alert", "sec-alert"},
	{"nav-message", "sec-message"},
	{"nav-loading", "sec-loading"},
	{"nav-empty", "sec-empty"},
	{"nav-divider", "sec-divider"},
	{"nav-grid", "sec-grid"},
	{"nav-space", "sec-space"},
	{"nav-panel", "sec-panel"},
	{"nav-tooltip", "sec-tooltip"},
}

func main() {
	app, err := ui.NewApp(ui.AppOptions{
		Title:  "GoUI — 组件库文档",
		Width:  1280,
		Height: 800,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	defer app.Destroy()

	doc := app.LoadHTML(demoHTML)

	// Setup page navigation: click sidebar → show one section
	setupNavigation(doc)

	// Setup programmatic widgets
	setupDemoWidgets(doc, app.Tree(), app.Config())

	// Setup event bindings
	setupEventBindings(doc)

	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

// setupNavigation wires sidebar buttons to show/hide content sections.
func setupNavigation(doc *ui.Document) {
	// Collect all section widgets
	sections := make([]widget.Widget, 0, len(nav))
	for _, n := range nav {
		if w := doc.QueryByID(n.section); w != nil {
			sections = append(sections, w)
		}
	}

	showSection := func(targetID string) {
		for _, n := range nav {
			w := doc.QueryByID(n.section)
			if w == nil {
				continue
			}
			s := w.Style()
			if n.section == targetID {
				s.Display = layout.DisplayBlock
			} else {
				s.Display = layout.DisplayNone
			}
			w.SetStyle(s)
		}
	}

	// Initially show only Button section
	showSection("sec-button")

	// Wire up click handlers
	for _, n := range nav {
		sid := n.section
		doc.OnClick(n.btn, func() {
			showSection(sid)
		})
	}
}

// setupEventBindings wires up event handlers for widgets created via HTML.
func setupEventBindings(doc *ui.Document) {
	doc.OnClick("btn-primary", func() { fmt.Println("[Button] 主要按钮 clicked") })
	doc.OnClick("btn-danger", func() { fmt.Println("[Button] 危险按钮 clicked") })
	doc.OnClick("btn-secondary", func() { fmt.Println("[Button] 次要按钮 clicked") })
	doc.OnClick("btn-text", func() { fmt.Println("[Button] 文字按钮 clicked") })
	doc.OnClick("btn-link", func() { fmt.Println("[Button] 链接按钮 clicked") })
	doc.OnChange("input-basic", func(v string) { fmt.Printf("[Input] = %q\n", v) })
	doc.OnChange("textarea-basic", func(v string) { fmt.Printf("[TextArea] len=%d\n", len(v)) })
	doc.OnToggle("cb-apple", func(v bool) { fmt.Printf("[Checkbox] 苹果 = %v\n", v) })
	doc.OnToggle("cb-banana", func(v bool) { fmt.Printf("[Checkbox] 香蕉 = %v\n", v) })
	doc.OnToggle("sw-basic", func(v bool) { fmt.Printf("[Switch] = %v\n", v) })

	if sel, ok := doc.QueryByID("select-basic").(*widget.Select); ok {
		sel.SetOptions([]widget.SelectOption{
			{Label: "北京", Value: "beijing"},
			{Label: "上海", Value: "shanghai"},
			{Label: "广州", Value: "guangzhou"},
			{Label: "深圳", Value: "shenzhen"},
			{Label: "杭州", Value: "hangzhou"},
		})
		sel.SetValue("beijing")
		sel.OnChange(func(v string) { fmt.Printf("[Select] = %s\n", v) })
	}
}

// setupDemoWidgets creates programmatic widgets and inserts them into placeholder containers.
func setupDemoWidgets(doc *ui.Document, tree *core.Tree, cfg *widget.Config) {
	type container interface {
		widget.Widget
		AppendChild(widget.Widget)
	}
	get := func(id string) container {
		w := doc.QueryByID(id)
		if w == nil {
			return nil
		}
		if c, ok := w.(container); ok {
			return c
		}
		return nil
	}

	// --- Divider ---
	if c := get("demo-divider"); c != nil {
		d := widget.NewDivider(tree, cfg)
		d.SetColor(uimath.ColorHex("#e8e8e8"))
		d.SetContent("分割线文字")
		c.AppendChild(d)
	}

	// --- Slider ---
	if c := get("demo-slider"); c != nil {
		s := widget.NewSlider(tree, cfg)
		s.SetMin(0)
		s.SetMax(100)
		s.SetValue(40)
		s.SetStep(1)
		s.OnChange(func(v float32) { fmt.Printf("[Slider] = %.0f\n", v) })
		c.AppendChild(s)
	}

	// --- InputNumber ---
	if c := get("demo-inputnumber"); c != nil {
		n := widget.NewInputNumber(tree, cfg)
		n.SetValue(5)
		n.SetMin(0)
		n.SetMax(100)
		n.SetStep(1)
		n.OnChange(func(v float64) { fmt.Printf("[InputNumber] = %.0f\n", v) })
		c.AppendChild(n)
	}

	// --- Rate ---
	if c := get("demo-rate"); c != nil {
		r := widget.NewRate(tree, cfg)
		r.SetValue(3)
		r.SetCount(5)
		r.OnChange(func(v float32) { fmt.Printf("[Rate] = %.1f\n", v) })
		c.AppendChild(r)
	}

	// --- Badge ---
	if c := get("demo-badge"); c != nil {
		b1 := widget.NewBadge(tree, cfg)
		b1.SetCount(8)
		c.AppendChild(b1)
		b2 := widget.NewBadge(tree, cfg)
		b2.SetCount(99)
		b2.SetMaxCount(99)
		c.AppendChild(b2)
		b3 := widget.NewBadge(tree, cfg)
		b3.SetDot(true)
		c.AppendChild(b3)
	}

	// --- Avatar ---
	if c := get("demo-avatar"); c != nil {
		a1 := widget.NewAvatar(tree, cfg)
		a1.SetText("U")
		a1.SetSize(widget.SizeMedium)
		a1.SetBgColor(uimath.ColorHex("#1677ff"))
		c.AppendChild(a1)
		a2 := widget.NewAvatar(tree, cfg)
		a2.SetText("A")
		a2.SetSize(widget.SizeMedium)
		a2.SetShape(widget.AvatarSquare)
		a2.SetBgColor(uimath.ColorHex("#00a870"))
		c.AppendChild(a2)
		a3 := widget.NewAvatar(tree, cfg)
		a3.SetText("B")
		a3.SetSize(widget.SizeSmall)
		a3.SetBgColor(uimath.ColorHex("#ed7b2f"))
		c.AppendChild(a3)
	}

	// --- Alert ---
	if c := get("demo-alert"); c != nil {
		a1 := widget.NewAlert(tree, "这是一条信息提示", cfg)
		a1.SetTheme(widget.AlertThemeInfo)
		c.AppendChild(a1)
		a2 := widget.NewAlert(tree, "操作已成功完成", cfg)
		a2.SetTheme(widget.AlertThemeSuccess)
		c.AppendChild(a2)
		a3 := widget.NewAlert(tree, "请注意检查配置", cfg)
		a3.SetTheme(widget.AlertThemeWarning)
		c.AppendChild(a3)
		a4 := widget.NewAlert(tree, "发生了一个错误", cfg)
		a4.SetTheme(widget.AlertThemeError)
		a4.SetCloseBtn(true)
		c.AppendChild(a4)
	}

	// --- Statistic ---
	if c := get("demo-statistic"); c != nil {
		s1 := widget.NewStatistic(tree, "用户总数", "28,846", cfg)
		s1.SetColor(uimath.ColorHex("#1677ff"))
		c.AppendChild(s1)
		s2 := widget.NewStatistic(tree, "成交金额", "568.08", cfg)
		s2.SetPrefix("¥")
		s2.SetSuffix("万")
		c.AppendChild(s2)
		s3 := widget.NewStatistic(tree, "满意度", "98.2", cfg)
		s3.SetSuffix("%")
		s3.SetColor(uimath.ColorHex("#00a870"))
		c.AppendChild(s3)
	}

	// --- Table ---
	if c := get("demo-table"); c != nil {
		t := widget.NewTable(tree, []widget.TableColumn{
			{Title: "序号", Width: 60, Align: widget.TableAlignCenter},
			{Title: "名称", Width: 0},
			{Title: "状态", Width: 80, Align: widget.TableAlignCenter},
			{Title: "更新时间", Width: 150},
		}, cfg)
		t.SetRows([][]string{
			{"1", "Button 按钮组件", "稳定", "2026-03-01"},
			{"2", "Input 输入框组件", "稳定", "2026-03-02"},
			{"3", "Table 表格组件", "Beta", "2026-03-05"},
			{"4", "Tree 树形组件", "Beta", "2026-03-06"},
			{"5", "Form 表单组件", "开发中", "2026-03-08"},
		})
		t.SetStripe(true)
		t.SetBordered(true)
		c.AppendChild(t)
	}

	// --- Tabs ---
	if c := get("demo-tabs"); c != nil {
		mk := func(text string) *widget.Div {
			d := widget.NewDiv(tree, cfg)
			l := widget.NewText(tree, text, cfg)
			l.SetColor(uimath.ColorHex("#666666"))
			d.AppendChild(l)
			return d
		}
		tabs := widget.NewTabs(tree, []widget.TabPanel{
			{Value: "tab1", Label: "基础组件", Content: mk("选项卡一：包含 Button、Input 等基础组件。")},
			{Value: "tab2", Label: "数据展示", Content: mk("选项卡二：数据展示组件，如 Table、List。")},
			{Value: "tab3", Label: "反馈", Content: mk("选项卡三：反馈组件，如 Alert、Message。")},
		}, cfg)
		tabs.OnChange(func(k string) { fmt.Printf("[Tabs] = %s\n", k) })
		c.AppendChild(tabs)
	}

	// --- Steps ---
	if c := get("demo-steps"); c != nil {
		s := widget.NewSteps(tree, cfg)
		s.AddStep(widget.StepItem{Title: "提交需求", Content: "填写需求信息"})
		s.AddStep(widget.StepItem{Title: "审核中", Content: "等待管理员审核"})
		s.AddStep(widget.StepItem{Title: "开发中", Content: "功能开发阶段"})
		s.AddStep(widget.StepItem{Title: "已完成", Content: "项目上线"})
		s.SetCurrent(1)
		c.AppendChild(s)
	}

	// --- Breadcrumb ---
	if c := get("demo-breadcrumb"); c != nil {
		bc := widget.NewBreadcrumb(tree, cfg)
		bc.SetOptions([]widget.BreadcrumbItem{
			{Content: "首页", Href: "/"},
			{Content: "组件库", Href: "/components"},
			{Content: "数据展示", Href: "/components/data"},
			{Content: "Table 表格"},
		})
		bc.OnClick(func(i int, href string) { fmt.Printf("[Breadcrumb] %d → %s\n", i, href) })
		c.AppendChild(bc)
	}

	// --- Pagination ---
	if c := get("demo-pagination"); c != nil {
		p := widget.NewPagination(tree, cfg)
		p.SetTotal(200)
		p.SetPageSize(10)
		p.SetCurrent(3)
		p.OnChange(func(info widget.PaginationPageInfo) { fmt.Printf("[Pagination] page = %d\n", info.Current) })
		c.AppendChild(p)
	}

	// --- Collapse ---
	if c := get("demo-collapse"); c != nil {
		mk := func(text string) *widget.Div {
			d := widget.NewDiv(tree, cfg)
			l := widget.NewText(tree, text, cfg)
			l.SetColor(uimath.ColorHex("#666666"))
			d.AppendChild(l)
			return d
		}
		col := widget.NewCollapse(tree, cfg)
		col.SetPanels([]widget.CollapsePanel{
			{Header: "面板一：基本信息", Value: "p1", Content: mk("折叠面板第一项的内容，可以放置任意组件。")},
			{Header: "面板二：详细描述", Value: "p2", Content: mk("折叠面板第二项的内容，默认关闭。")},
			{Header: "面板三：更多设置", Value: "p3", Content: mk("折叠面板第三项的内容。")},
		})
		col.SetBordered(true)
		col.Toggle("p1")
		c.AppendChild(col)
	}

	// --- Timeline ---
	if c := get("demo-timeline"); c != nil {
		tl := widget.NewTimeline(tree, cfg)
		tl.AddItem(widget.TimelineItem{Label: "2026-03-08", Content: "项目初始化", DotColor: "primary"})
		tl.AddItem(widget.TimelineItem{Label: "2026-03-05", Content: "基础组件完成", DotColor: "primary"})
		tl.AddItem(widget.TimelineItem{Label: "2026-03-03", Content: "发现性能问题", DotColor: "warning"})
		tl.AddItem(widget.TimelineItem{Label: "2026-03-01", Content: "立项启动", DotColor: "default"})
		c.AppendChild(tl)
	}

	// --- List ---
	if c := get("demo-list"); c != nil {
		l := widget.NewList(tree, cfg)
		l.SetItems([]widget.ListItem{
			{Title: "列表主内容", Description: "列表内容列表内容", Actions: []string{"操作1", "操作2", "操作3"}},
			{Title: "列表主内容", Description: "列表内容列表内容", Actions: []string{"操作1", "操作2", "操作3"}},
			{Title: "列表主内容", Description: "列表内容列表内容", Actions: []string{"操作1", "操作2", "操作3"}},
		})
		l.SetBordered(true)
		l.OnSelect(func(i int) { fmt.Printf("[List] selected = %d\n", i) })
		c.AppendChild(l)
	}

	// --- Card ---
	if c := get("demo-card"); c != nil {
		card := widget.NewCard(tree, cfg)
		card.SetTitle("卡片标题")
		card.SetBordered(true)
		card.SetShadow(true)
		t := widget.NewText(tree, "卡片组件用于展示一组相关的信息，支持标题、内容和底部操作区。", cfg)
		t.SetColor(uimath.ColorHex("#666666"))
		card.AppendChild(t)
		c.AppendChild(card)
	}

	// --- Panel ---
	if c := get("demo-panel"); c != nil {
		p := widget.NewPanel(tree, "面板标题", cfg)
		p.SetBordered(true)
		p.SetBgColor(uimath.ColorHex("#fafafa"))
		t := widget.NewText(tree, "面板是一个带标题的容器组件，用于对页面内容进行分组。", cfg)
		t.SetColor(uimath.ColorHex("#666666"))
		p.AppendChild(t)
		c.AppendChild(p)
	}

	// --- Menu ---
	if c := get("demo-menu"); c != nil {
		m := widget.NewMenu(tree, cfg)
		m.SetItems([]widget.MenuItem{
			{Value: "home", Content: "首页"},
			{Value: "components", Content: "组件库", Children: []widget.MenuItem{
				{Value: "basic", Content: "基础组件"},
				{Value: "form", Content: "表单组件"},
				{Value: "data", Content: "数据展示"},
			}},
			{Value: "docs", Content: "文档中心"},
			{Value: "about", Content: "关于", Disabled: true},
		})
		m.SetValue("basic")
		m.OnChange(func(k string) { fmt.Printf("[Menu] = %s\n", k) })
		c.AppendChild(m)
	}
}

const demoHTML = `
<style>
	:root {
		--bg-sidebar: #001529;
		--bg-content: #f0f2f5;
		--bg-card: #ffffff;
		--bg-header: #001529;
		--text-title: #1a1a1a;
		--text-body: #666666;
		--border: #e8e8e8;
	}

	layout { background-color: var(--bg-content); }
	header { background-color: var(--bg-header); }
	aside { background-color: var(--bg-sidebar); }
	main { background-color: var(--bg-content); padding: 24px; }

	.section-title { font-size: 22px; color: var(--text-title); }
	.section-desc { font-size: 13px; color: var(--text-body); }

	.demo-card {
		background-color: var(--bg-card);
		border-radius: 6px;
		border-width: 1px;
		border-color: var(--border);
		padding: 20px;
	}
	.demo-label { font-size: 12px; color: var(--text-body); }

	.grid-col-1 { background-color: #e6f4ff; border-radius: 4px; padding: 8px; }
	.grid-col-2 { background-color: #bae0ff; border-radius: 4px; padding: 8px; }
	.grid-col-3 { background-color: #91caff; border-radius: 4px; padding: 8px; }
	.grid-col-4 { background-color: #69b1ff; border-radius: 4px; padding: 8px; }
	.grid-text { color: #0958d9; font-size: 13px; }

	.cat-title { color: #ffffff60; font-size: 12px; }
	.logo { color: #ffffff; font-size: 18px; }
</style>

<layout>
	<header height="48">
		<span class="logo">GoUI 组件库</span>
	</header>

	<div style="display: flex; flex-direction: row; flex-grow: 1">
		<aside width="180" style="padding: 0px">
			<div id="sidebar-scroll" style="padding: 12px">
				<span class="cat-title">基础</span>
				<button id="nav-button" variant="text">Button 按钮</button>
				<button id="nav-link" variant="text">Link 链接</button>

				<span class="cat-title">输入</span>
				<button id="nav-input" variant="text">Input 输入框</button>
				<button id="nav-textarea" variant="text">TextArea 文本域</button>
				<button id="nav-inputnumber" variant="text">InputNumber 数字</button>
				<button id="nav-select" variant="text">Select 选择器</button>
				<button id="nav-checkbox" variant="text">Checkbox 多选</button>
				<button id="nav-radio" variant="text">Radio 单选</button>
				<button id="nav-switch" variant="text">Switch 开关</button>
				<button id="nav-slider" variant="text">Slider 滑块</button>
				<button id="nav-rate" variant="text">Rate 评分</button>

				<span class="cat-title">数据展示</span>
				<button id="nav-tag" variant="text">Tag 标签</button>
				<button id="nav-badge" variant="text">Badge 徽标</button>
				<button id="nav-avatar" variant="text">Avatar 头像</button>
				<button id="nav-progress" variant="text">Progress 进度</button>
				<button id="nav-table" variant="text">Table 表格</button>
				<button id="nav-list" variant="text">List 列表</button>
				<button id="nav-card" variant="text">Card 卡片</button>
				<button id="nav-statistic" variant="text">Statistic 统计</button>
				<button id="nav-collapse" variant="text">Collapse 折叠</button>
				<button id="nav-timeline" variant="text">Timeline 时间线</button>

				<span class="cat-title">导航</span>
				<button id="nav-tabs" variant="text">Tabs 选项卡</button>
				<button id="nav-menu" variant="text">Menu 菜单</button>
				<button id="nav-breadcrumb" variant="text">Breadcrumb 面包屑</button>
				<button id="nav-pagination" variant="text">Pagination 分页</button>
				<button id="nav-steps" variant="text">Steps 步骤条</button>

				<span class="cat-title">反馈</span>
				<button id="nav-alert" variant="text">Alert 警告</button>
				<button id="nav-message" variant="text">Message 消息</button>
				<button id="nav-loading" variant="text">Loading 加载</button>
				<button id="nav-empty" variant="text">Empty 空状态</button>

				<span class="cat-title">布局</span>
				<button id="nav-divider" variant="text">Divider 分割线</button>
				<button id="nav-grid" variant="text">Grid 栅格</button>
				<button id="nav-space" variant="text">Space 间距</button>
				<button id="nav-panel" variant="text">Panel 面板</button>
				<button id="nav-tooltip" variant="text">Tooltip 提示</button>
			</div>
		</aside>

		<main id="content">
			<!-- ===== Button ===== -->
			<div id="sec-button">
				<h2 class="section-title">Button 按钮</h2>
				<span class="section-desc">按钮用于开启一个闭环的操作任务，如"删除"对象、"购买"商品等。</span>
				<div class="demo-card">
					<span class="demo-label">基础按钮</span>
					<space gap="12">
						<button id="btn-primary">主要按钮</button>
						<button id="btn-danger" style="background-color: #e34d59">危险按钮</button>
						<button id="btn-secondary" variant="secondary">次要按钮</button>
						<button id="btn-text" variant="text">文字按钮</button>
						<button id="btn-link" variant="link">链接按钮</button>
					</space>
					<span class="demo-label">按钮状态</span>
					<space gap="12">
						<button disabled>禁用状态</button>
						<button variant="secondary" disabled>禁用次要</button>
					</space>
				</div>
			</div>

			<!-- ===== Link ===== -->
			<div id="sec-link">
				<h2 class="section-title">Link 链接</h2>
				<span class="section-desc">文字超链接，适用于页面跳转、锚点定位等场景。</span>
				<div class="demo-card">
					<space gap="16">
						<a href="https://github.com">默认链接</a>
						<a href="#">另一个链接</a>
					</space>
				</div>
			</div>

			<!-- ===== Input ===== -->
			<div id="sec-input">
				<h2 class="section-title">Input 输入框</h2>
				<span class="section-desc">用于承载用户信息录入的文本输入框。</span>
				<div class="demo-card">
					<span class="demo-label">基础输入框</span>
					<input id="input-basic" placeholder="请输入内容..."/>
					<span class="demo-label">带默认值</span>
					<input value="默认值内容"/>
					<span class="demo-label">禁用状态</span>
					<input value="不可编辑" disabled/>
				</div>
			</div>

			<!-- ===== TextArea ===== -->
			<div id="sec-textarea">
				<h2 class="section-title">TextArea 多行文本框</h2>
				<span class="section-desc">多行纯文本编辑框，适用于评论、备注等场景。</span>
				<div class="demo-card">
					<textarea id="textarea-basic" placeholder="请输入详细描述..." rows="4"></textarea>
				</div>
			</div>

			<!-- ===== InputNumber ===== -->
			<div id="sec-inputnumber">
				<h2 class="section-title">InputNumber 数字输入框</h2>
				<span class="section-desc">通过点击或键盘输入内容，仅允许输入数字格式的输入框。</span>
				<div class="demo-card"><div id="demo-inputnumber"></div></div>
			</div>

			<!-- ===== Select ===== -->
			<div id="sec-select">
				<h2 class="section-title">Select 选择器</h2>
				<span class="section-desc">用于收集用户提供的信息，需从预设的一组选项中选择。</span>
				<div class="demo-card">
					<span class="demo-label">基础选择器</span>
					<select id="select-basic"></select>
				</div>
			</div>

			<!-- ===== Checkbox ===== -->
			<div id="sec-checkbox">
				<h2 class="section-title">Checkbox 多选框</h2>
				<span class="section-desc">多选框代表从一组选项中选中若干选项。</span>
				<div class="demo-card">
					<space gap="16">
						<checkbox id="cb-apple" checked>苹果</checkbox>
						<checkbox id="cb-banana">香蕉</checkbox>
						<checkbox>橙子</checkbox>
						<checkbox disabled>禁用</checkbox>
					</space>
				</div>
			</div>

			<!-- ===== Radio ===== -->
			<div id="sec-radio">
				<h2 class="section-title">Radio 单选框</h2>
				<span class="section-desc">单选框代表从一组互斥的选项中仅选择一个选项。</span>
				<div class="demo-card">
					<space gap="16">
						<radio group="city" checked>北京</radio>
						<radio group="city">上海</radio>
						<radio group="city">广州</radio>
						<radio group="city" disabled>深圳（禁用）</radio>
					</space>
				</div>
			</div>

			<!-- ===== Switch ===== -->
			<div id="sec-switch">
				<h2 class="section-title">Switch 开关</h2>
				<span class="section-desc">用于两个互斥选项之间的切换。</span>
				<div class="demo-card">
					<space gap="16">
						<switch id="sw-basic" checked></switch>
						<switch></switch>
						<switch disabled></switch>
					</space>
				</div>
			</div>

			<!-- ===== Slider ===== -->
			<div id="sec-slider">
				<h2 class="section-title">Slider 滑块</h2>
				<span class="section-desc">通过滑动滑块在一个范围内获取特定值。</span>
				<div class="demo-card"><div id="demo-slider"></div></div>
			</div>

			<!-- ===== Rate ===== -->
			<div id="sec-rate">
				<h2 class="section-title">Rate 评分</h2>
				<span class="section-desc">用于对事物进行评级操作。</span>
				<div class="demo-card"><div id="demo-rate"></div></div>
			</div>

			<!-- ===== Tag ===== -->
			<div id="sec-tag">
				<h2 class="section-title">Tag 标签</h2>
				<span class="section-desc">标签常用于标记、分类和选择。</span>
				<div class="demo-card">
					<space gap="8">
						<tag>默认标签</tag>
						<tag type="success">成功</tag>
						<tag type="warning">警告</tag>
						<tag type="error">错误</tag>
						<tag type="processing">处理中</tag>
					</space>
				</div>
			</div>

			<!-- ===== Badge ===== -->
			<div id="sec-badge">
				<h2 class="section-title">Badge 徽标</h2>
				<span class="section-desc">出现在图标或文字右上角的徽标数字，用于信息提示。</span>
				<div class="demo-card"><space gap="24" id="demo-badge"></space></div>
			</div>

			<!-- ===== Avatar ===== -->
			<div id="sec-avatar">
				<h2 class="section-title">Avatar 头像</h2>
				<span class="section-desc">用于头像展示。</span>
				<div class="demo-card"><space gap="16" id="demo-avatar"></space></div>
			</div>

			<!-- ===== Progress ===== -->
			<div id="sec-progress">
				<h2 class="section-title">Progress 进度条</h2>
				<span class="section-desc">展示操作的当前进度。</span>
				<div class="demo-card">
					<span class="demo-label">65%</span>
					<progress percent="65"></progress>
					<span class="demo-label">30%</span>
					<progress percent="30"></progress>
					<span class="demo-label">100%</span>
					<progress percent="100"></progress>
				</div>
			</div>

			<!-- ===== Table ===== -->
			<div id="sec-table">
				<h2 class="section-title">Table 表格</h2>
				<span class="section-desc">展示行列数据。</span>
				<div class="demo-card"><div id="demo-table"></div></div>
			</div>

			<!-- ===== List ===== -->
			<div id="sec-list">
				<h2 class="section-title">List 列表</h2>
				<span class="section-desc">以列表的形式展示同类别的内容。</span>
				<div class="demo-card"><div id="demo-list"></div></div>
			</div>

			<!-- ===== Card ===== -->
			<div id="sec-card">
				<h2 class="section-title">Card 卡片</h2>
				<span class="section-desc">最基础的卡片容器，可承载文字、图片等内容。</span>
				<div class="demo-card"><div id="demo-card"></div></div>
			</div>

			<!-- ===== Statistic ===== -->
			<div id="sec-statistic">
				<h2 class="section-title">Statistic 统计数值</h2>
				<span class="section-desc">突出展示某个或某组数字、带描述的统计类数据。</span>
				<div class="demo-card"><space gap="48" id="demo-statistic"></space></div>
			</div>

			<!-- ===== Collapse ===== -->
			<div id="sec-collapse">
				<h2 class="section-title">Collapse 折叠面板</h2>
				<span class="section-desc">可以折叠或展开的内容区域。</span>
				<div class="demo-card"><div id="demo-collapse"></div></div>
			</div>

			<!-- ===== Timeline ===== -->
			<div id="sec-timeline">
				<h2 class="section-title">Timeline 时间线</h2>
				<span class="section-desc">按照时间顺序展示事件信息。</span>
				<div class="demo-card"><div id="demo-timeline"></div></div>
			</div>

			<!-- ===== Tabs ===== -->
			<div id="sec-tabs">
				<h2 class="section-title">Tabs 选项卡</h2>
				<span class="section-desc">选项卡切换组件，用于在同一位置展示不同内容。</span>
				<div class="demo-card"><div id="demo-tabs"></div></div>
			</div>

			<!-- ===== Menu ===== -->
			<div id="sec-menu">
				<h2 class="section-title">Menu 导航菜单</h2>
				<span class="section-desc">用于页面或功能间的导航切换。</span>
				<div class="demo-card"><div id="demo-menu"></div></div>
			</div>

			<!-- ===== Breadcrumb ===== -->
			<div id="sec-breadcrumb">
				<h2 class="section-title">Breadcrumb 面包屑</h2>
				<span class="section-desc">面包屑导航用于告诉用户当前页面在系统层级结构中的位置。</span>
				<div class="demo-card"><div id="demo-breadcrumb"></div></div>
			</div>

			<!-- ===== Pagination ===== -->
			<div id="sec-pagination">
				<h2 class="section-title">Pagination 分页</h2>
				<span class="section-desc">用于数据量较大时，采用分页的方式快速访问数据。</span>
				<div class="demo-card"><div id="demo-pagination"></div></div>
			</div>

			<!-- ===== Steps ===== -->
			<div id="sec-steps">
				<h2 class="section-title">Steps 步骤条</h2>
				<span class="section-desc">引导用户按照流程完成任务的导航条。</span>
				<div class="demo-card"><div id="demo-steps"></div></div>
			</div>

			<!-- ===== Alert ===== -->
			<div id="sec-alert">
				<h2 class="section-title">Alert 警告提示</h2>
				<span class="section-desc">警告提示，展示需要关注的信息。</span>
				<div class="demo-card"><div id="demo-alert"></div></div>
			</div>

			<!-- ===== Message ===== -->
			<div id="sec-message">
				<h2 class="section-title">Message 全局提示</h2>
				<span class="section-desc">轻量级的全局反馈，常用于操作结果提示。</span>
				<div class="demo-card">
					<space gap="12">
						<message>普通消息</message>
						<message type="success">操作成功</message>
						<message type="warning">请注意</message>
						<message type="error">发生了一个错误</message>
					</space>
				</div>
			</div>

			<!-- ===== Loading ===== -->
			<div id="sec-loading">
				<h2 class="section-title">Loading 加载</h2>
				<span class="section-desc">用于表示数据加载等待状态。</span>
				<div class="demo-card"><loading tip="正在加载数据..."></loading></div>
			</div>

			<!-- ===== Empty ===== -->
			<div id="sec-empty">
				<h2 class="section-title">Empty 空状态</h2>
				<span class="section-desc">空状态时的展示占位图。</span>
				<div class="demo-card"><empty></empty></div>
			</div>

			<!-- ===== Divider ===== -->
			<div id="sec-divider">
				<h2 class="section-title">Divider 分割线</h2>
				<span class="section-desc">分隔内容区域，对页面内容进行分类。</span>
				<div class="demo-card"><div id="demo-divider"></div></div>
			</div>

			<!-- ===== Grid ===== -->
			<div id="sec-grid">
				<h2 class="section-title">Grid 栅格</h2>
				<span class="section-desc">二十四列栅格系统，用于快速搭建页面布局。</span>
				<div class="demo-card">
					<span class="demo-label">四等分</span>
					<row gutter="16">
						<col span="6"><div class="grid-col-1"><span class="grid-text">col-6</span></div></col>
						<col span="6"><div class="grid-col-2"><span class="grid-text">col-6</span></div></col>
						<col span="6"><div class="grid-col-3"><span class="grid-text">col-6</span></div></col>
						<col span="6"><div class="grid-col-4"><span class="grid-text">col-6</span></div></col>
					</row>
					<span class="demo-label">不等分</span>
					<row gutter="16">
						<col span="8"><div class="grid-col-1"><span class="grid-text">col-8</span></div></col>
						<col span="16"><div class="grid-col-2"><span class="grid-text">col-16</span></div></col>
					</row>
					<span class="demo-label">三栏布局</span>
					<row gutter="16">
						<col span="4"><div class="grid-col-1"><span class="grid-text">col-4</span></div></col>
						<col span="16"><div class="grid-col-2"><span class="grid-text">col-16</span></div></col>
						<col span="4"><div class="grid-col-3"><span class="grid-text">col-4</span></div></col>
					</row>
				</div>
			</div>

			<!-- ===== Space ===== -->
			<div id="sec-space">
				<h2 class="section-title">Space 间距</h2>
				<span class="section-desc">设置组件之间的间距。</span>
				<div class="demo-card">
					<span class="demo-label">gap=8</span>
					<space gap="8">
						<button>按钮一</button>
						<button variant="secondary">按钮二</button>
						<button variant="text">按钮三</button>
					</space>
					<span class="demo-label">gap=24</span>
					<space gap="24">
						<button>按钮一</button>
						<button variant="secondary">按钮二</button>
						<button variant="text">按钮三</button>
					</space>
				</div>
			</div>

			<!-- ===== Panel ===== -->
			<div id="sec-panel">
				<h2 class="section-title">Panel 面板</h2>
				<span class="section-desc">面板是一种容器组件，可以用于放置内容。</span>
				<div class="demo-card"><div id="demo-panel"></div></div>
			</div>

			<!-- ===== Tooltip ===== -->
			<div id="sec-tooltip">
				<h2 class="section-title">Tooltip 文字提示</h2>
				<span class="section-desc">鼠标悬停时显示提示信息。</span>
				<div class="demo-card">
					<button variant="secondary">悬停查看提示</button>
					<tooltip>这是一个工具提示 Tooltip!</tooltip>
				</div>
			</div>

		</main>
	</div>
</layout>
`
