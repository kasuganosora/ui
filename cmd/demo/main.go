//go:build windows

// Demo showcases the GoUI widget library using the HTML+CSS template approach.
// Run: go run ./cmd/demo
package main

import (
	"fmt"
	"os"

	ui "github.com/kasuganosora/ui"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/widget"
)

func main() {
	app, err := ui.NewApp(ui.AppOptions{
		Title:  "GoUI Demo — 组件演示平台",
		Width:  960,
		Height: 640,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	defer app.Destroy()

	doc := app.LoadHTML(demoHTML)

	// Wire up aside border (not expressible in CSS yet)
	if aside := doc.QueryByTag("aside"); len(aside) > 0 {
		if a, ok := aside[0].(*widget.Aside); ok {
			a.SetBorderRight(1, uimath.ColorHex("#e8e8e8"))
		}
	}

	// Event bindings
	doc.OnClick("btn-primary", func() { fmt.Println("[点击] 主要按钮") })
	doc.OnClick("btn-secondary", func() { fmt.Println("[点击] 次要按钮") })
	doc.OnClick("btn-text", func() { fmt.Println("[点击] 文字按钮") })
	doc.OnClick("btn-link", func() { fmt.Println("[点击] 链接按钮") })

	doc.OnChange("input-name", func(v string) { fmt.Printf("[输入] 姓名 = %q\n", v) })
	doc.OnChange("textarea", func(v string) { fmt.Printf("[多行] len=%d\n", len(v)) })

	doc.OnToggle("cb-a", func(v bool) { fmt.Printf("[复选] A = %v\n", v) })
	doc.OnToggle("sw-1", func(v bool) { fmt.Printf("[开关] = %v\n", v) })

	// Select needs options set programmatically
	if sel, ok := doc.QueryByID("city-select").(*widget.Select); ok {
		sel.SetOptions([]widget.SelectOption{
			{Label: "北京", Value: "beijing"},
			{Label: "上海", Value: "shanghai"},
			{Label: "广州", Value: "guangzhou"},
			{Label: "深圳（禁用）", Value: "shenzhen", Disabled: true},
		})
		sel.SetValue("beijing")
		sel.OnChange(func(v string) { fmt.Printf("[选择] = %s\n", v) })
	}

	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

const demoHTML = `
<style>
	:root {
		--primary-bg: #001529;
		--aside-bg: #f7f8fa;
		--content-bg: #ffffff;
		--title-color: #1a1a1a;
		--col-text: #0050b3;
		--footer-text: #ffffff80;
	}

	layout { background-color: #f0f2f5; }
	header { background-color: var(--primary-bg); }
	footer { background-color: var(--primary-bg); }

	.title { color: white; font-size: 20px; }
	.body { display: flex; flex-direction: row; flex-grow: 1; background-color: white; }
	main { background-color: var(--content-bg); }

	.section-title { font-size: 24px; color: var(--title-color); }

	.col-1 { background-color: #e6f7ff; border-radius: 4px; }
	.col-2 { background-color: #bae7ff; border-radius: 4px; }
	.col-3 { background-color: #91d5ff; border-radius: 4px; }
	.col-4 { background-color: #69c0ff; border-radius: 4px; }
	.col-text { color: var(--col-text); }

	.footer-text { color: var(--footer-text); font-size: 12px; }
</style>

<layout>
	<header height="56">
		<span class="title">组件演示平台</span>
	</header>

	<div class="body">
		<aside width="220" style="background-color: #f7f8fa">
			<button id="menu-dashboard" variant="text">仪表盘</button>
			<button id="menu-components" variant="text">组件库</button>
			<button id="menu-settings" variant="text">系统设置</button>
			<button id="menu-about" variant="text">关于我们</button>
		</aside>

		<main id="content">
			<span class="section-title">基础组件展示</span>

			<space gap="12">
				<button id="btn-primary">主要按钮</button>
				<button id="btn-secondary" variant="secondary">次要按钮</button>
				<button id="btn-text" variant="text">文字按钮</button>
				<button id="btn-link" variant="link">链接按钮</button>
				<button disabled>禁用按钮</button>
			</space>

			<span>输入框：</span>
			<input id="input-name" placeholder="请输入您的姓名..."/>
			<input placeholder="请输入电子邮箱..."/>
			<input value="已禁用的输入框" disabled/>

			<span>栅格布局（二十四列）：</span>
			<row gutter="16">
				<col span="6"><div class="col-1"><span class="col-text">第一列</span></div></col>
				<col span="6"><div class="col-2"><span class="col-text">第二列</span></div></col>
				<col span="6"><div class="col-3"><span class="col-text">第三列</span></div></col>
				<col span="6"><div class="col-4"><span class="col-text">第四列</span></div></col>
			</row>

			<button id="btn-tooltip" variant="secondary">悬停查看提示信息</button>
			<tooltip>这是一个工具提示！</tooltip>

			<span>复选框 &amp; 开关：</span>
			<space gap="16">
				<checkbox id="cb-a" checked>选项A</checkbox>
				<checkbox>选项B</checkbox>
				<checkbox disabled>禁用</checkbox>
				<switch id="sw-1" checked></switch>
				<switch disabled></switch>
			</space>

			<span>单选按钮：</span>
			<space gap="16">
				<radio group="fruit" checked>苹果</radio>
				<radio group="fruit">香蕉</radio>
				<radio group="fruit">橙子</radio>
			</space>

			<span>标签：</span>
			<space gap="8">
				<tag>默认</tag>
				<tag type="success">成功</tag>
				<tag type="warning">警告</tag>
				<tag type="error">错误</tag>
				<tag type="processing">处理中</tag>
			</space>

			<span>进度条：</span>
			<progress percent="65"></progress>

			<span>多行输入框：</span>
			<textarea id="textarea" placeholder="请输入多行文本..." rows="3"></textarea>

			<span>下拉选择：</span>
			<select id="city-select"></select>

			<span>消息通知：</span>
			<space gap="12">
				<message>普通消息</message>
				<message type="success">操作成功</message>
				<message type="warning">请注意</message>
				<message type="error">出错了</message>
			</space>

			<span>空状态：</span>
			<empty></empty>

			<span>加载中：</span>
			<loading tip="正在加载..."></loading>
		</main>
	</div>

	<footer>
		<span class="footer-text">GoUI v0.1 — 零CGO跨平台界面库</span>
	</footer>
</layout>
`
