package widget_test

import (
	"fmt"
	"testing"

	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/layout"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
	"github.com/kasuganosora/ui/widget"
)

// benchEnv provides shared test infrastructure for widget benchmarks.
type benchEnv struct {
	tree *core.Tree
	cfg  *widget.Config
	buf  *render.CommandBuffer
}

func newBenchEnv() *benchEnv {
	tree := core.NewTree()
	cfg := widget.DefaultConfig()
	buf := render.NewCommandBuffer()
	return &benchEnv{tree: tree, cfg: cfg, buf: buf}
}

// setLayout sets bounds on a widget's element.
func (e *benchEnv) setLayout(w widget.Widget, x, y, width, height float32) {
	e.tree.SetLayout(w.ElementID(), core.LayoutResult{
		Bounds: uimath.NewRect(x, y, width, height),
	})
}

// ---- Widget Creation Benchmarks ----

func BenchmarkNewText(b *testing.B) {
	env := newBenchEnv()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		widget.NewText(env.tree, "Hello World", env.cfg)
	}
}

func BenchmarkNewButton(b *testing.B) {
	env := newBenchEnv()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		widget.NewButton(env.tree, "Click Me", env.cfg)
	}
}

func BenchmarkNewDiv(b *testing.B) {
	env := newBenchEnv()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		widget.NewDiv(env.tree, env.cfg)
	}
}

func BenchmarkNewInput(b *testing.B) {
	env := newBenchEnv()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		widget.NewInput(env.tree, env.cfg)
	}
}

func BenchmarkNewAvatar(b *testing.B) {
	env := newBenchEnv()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		widget.NewAvatar(env.tree, env.cfg)
	}
}

func BenchmarkNewCheckbox(b *testing.B) {
	env := newBenchEnv()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		widget.NewCheckbox(env.tree, "Option", env.cfg)
	}
}

func BenchmarkNewRadio(b *testing.B) {
	env := newBenchEnv()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		widget.NewRadio(env.tree, "Choice", env.cfg)
	}
}

func BenchmarkNewTag(b *testing.B) {
	env := newBenchEnv()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		widget.NewTag(env.tree, "Label", env.cfg)
	}
}

func BenchmarkNewBadge(b *testing.B) {
	env := newBenchEnv()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b := widget.NewBadge(env.tree, env.cfg)
		b.SetCount(99)
	}
}

func BenchmarkNewCard(b *testing.B) {
	env := newBenchEnv()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c := widget.NewCard(env.tree, env.cfg)
		c.SetTitle("Title")
	}
}

func BenchmarkNewProgress(b *testing.B) {
	env := newBenchEnv()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		p := widget.NewProgress(env.tree, env.cfg)
		p.SetPercentage(0.5)
	}
}

func BenchmarkNewSwitch(b *testing.B) {
	env := newBenchEnv()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		widget.NewSwitch(env.tree, env.cfg)
	}
}

func BenchmarkNewSlider(b *testing.B) {
	env := newBenchEnv()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		widget.NewSlider(env.tree, env.cfg)
	}
}

// ---- Draw Benchmarks ----

func BenchmarkDrawText(b *testing.B) {
	env := newBenchEnv()
	t := widget.NewText(env.tree, "Hello World, 你好世界, benchmark text rendering", env.cfg)
	env.setLayout(t, 0, 0, 300, 20)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		env.buf.Reset()
		t.Draw(env.buf)
	}
}

func BenchmarkDrawButton(b *testing.B) {
	env := newBenchEnv()
	btn := widget.NewButton(env.tree, "Click Me", env.cfg)
	env.setLayout(btn, 0, 0, 120, 32)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		env.buf.Reset()
		btn.Draw(env.buf)
	}
}

func BenchmarkDrawDiv(b *testing.B) {
	env := newBenchEnv()
	d := widget.NewDiv(env.tree, env.cfg)
	d.SetBgColor(uimath.ColorHex("#f0f0f0"))
	d.SetBorderColor(uimath.ColorHex("#ddd"))
	d.SetBorderWidth(1)
	d.SetBorderRadius(6)
	env.setLayout(d, 0, 0, 400, 200)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		env.buf.Reset()
		d.Draw(env.buf)
	}
}

func BenchmarkDrawDivWithChildren(b *testing.B) {
	env := newBenchEnv()
	d := widget.NewDiv(env.tree, env.cfg)
	d.SetBgColor(uimath.ColorHex("#f0f0f0"))
	env.setLayout(d, 0, 0, 400, 200)
	for j := 0; j < 10; j++ {
		t := widget.NewText(env.tree, "Child text line", env.cfg)
		env.setLayout(t, 10, float32(j*20), 380, 20)
		d.AppendChild(t)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		env.buf.Reset()
		d.Draw(env.buf)
	}
}

func BenchmarkDrawInput(b *testing.B) {
	env := newBenchEnv()
	inp := widget.NewInput(env.tree, env.cfg)
	inp.SetValue("Some input text here")
	env.setLayout(inp, 0, 0, 200, 32)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		env.buf.Reset()
		inp.Draw(env.buf)
	}
}

func BenchmarkDrawAvatar(b *testing.B) {
	env := newBenchEnv()
	av := widget.NewAvatar(env.tree, env.cfg)
	av.SetContent("Alice")
	av.SetShape(widget.AvatarCircle)
	env.setLayout(av, 0, 0, 40, 40)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		env.buf.Reset()
		av.Draw(env.buf)
	}
}

func BenchmarkDrawCheckbox(b *testing.B) {
	env := newBenchEnv()
	cb := widget.NewCheckbox(env.tree, "Check me", env.cfg)
	env.setLayout(cb, 0, 0, 120, 20)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		env.buf.Reset()
		cb.Draw(env.buf)
	}
}

func BenchmarkDrawTag(b *testing.B) {
	env := newBenchEnv()
	tag := widget.NewTag(env.tree, "Label", env.cfg)
	env.setLayout(tag, 0, 0, 60, 22)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		env.buf.Reset()
		tag.Draw(env.buf)
	}
}

func BenchmarkDrawBadge(b *testing.B) {
	env := newBenchEnv()
	badge := widget.NewBadge(env.tree, env.cfg)
	badge.SetCount(99)
	env.setLayout(badge, 0, 0, 20, 20)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		env.buf.Reset()
		badge.Draw(env.buf)
	}
}

func BenchmarkDrawProgress(b *testing.B) {
	env := newBenchEnv()
	p := widget.NewProgress(env.tree, env.cfg)
	p.SetPercentage(0.75)
	env.setLayout(p, 0, 0, 200, 8)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		env.buf.Reset()
		p.Draw(env.buf)
	}
}

func BenchmarkDrawSwitch(b *testing.B) {
	env := newBenchEnv()
	sw := widget.NewSwitch(env.tree, env.cfg)
	env.setLayout(sw, 0, 0, 44, 22)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		env.buf.Reset()
		sw.Draw(env.buf)
	}
}

func BenchmarkDrawCard(b *testing.B) {
	env := newBenchEnv()
	c := widget.NewCard(env.tree, env.cfg)
	c.SetTitle("Card Title")
	env.setLayout(c, 0, 0, 300, 200)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		env.buf.Reset()
		c.Draw(env.buf)
	}
}

func BenchmarkDrawSlider(b *testing.B) {
	env := newBenchEnv()
	s := widget.NewSlider(env.tree, env.cfg)
	s.SetValue(0.5)
	env.setLayout(s, 0, 0, 200, 20)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		env.buf.Reset()
		s.Draw(env.buf)
	}
}

func BenchmarkDrawAlert(b *testing.B) {
	env := newBenchEnv()
	a := widget.NewAlert(env.tree, "This is an info alert", env.cfg)
	env.setLayout(a, 0, 0, 400, 48)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		env.buf.Reset()
		a.Draw(env.buf)
	}
}

// ---- Viewport Culling Benchmarks ----

func BenchmarkDrawChildren_NoCull(b *testing.B) {
	// All children visible (no clipping)
	env := newBenchEnv()
	parent := widget.NewDiv(env.tree, env.cfg)
	env.setLayout(parent, 0, 0, 400, 1000)
	for j := 0; j < 50; j++ {
		t := widget.NewText(env.tree, "Visible line", env.cfg)
		env.setLayout(t, 10, float32(j*20), 380, 20)
		parent.AppendChild(t)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		env.buf.Reset()
		parent.Draw(env.buf)
	}
}

func BenchmarkDrawChildren_WithCull(b *testing.B) {
	// Only 10 of 50 children visible (clip rect covers first 200px)
	env := newBenchEnv()
	parent := widget.NewDiv(env.tree, env.cfg)
	parent.SetScrollable(true)
	env.setLayout(parent, 0, 0, 400, 200) // viewport is 200px tall
	for j := 0; j < 50; j++ {
		t := widget.NewText(env.tree, "Maybe visible line", env.cfg)
		env.setLayout(t, 10, float32(j*20), 380, 20)
		parent.AppendChild(t)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		env.buf.Reset()
		parent.Draw(env.buf) // PushClip + DrawChildren with culling
	}
}

func BenchmarkDrawChildren_100_WithCull(b *testing.B) {
	// Only 10 of 100 visible
	env := newBenchEnv()
	parent := widget.NewDiv(env.tree, env.cfg)
	parent.SetScrollable(true)
	env.setLayout(parent, 0, 0, 400, 200)
	for j := 0; j < 100; j++ {
		t := widget.NewText(env.tree, "Maybe visible line", env.cfg)
		env.setLayout(t, 10, float32(j*20), 380, 20)
		parent.AppendChild(t)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		env.buf.Reset()
		parent.Draw(env.buf)
	}
}

// ---- Command Buffer Benchmarks ----

func BenchmarkCommandBuffer_DrawRect(b *testing.B) {
	buf := render.NewCommandBuffer()
	cmd := render.RectCmd{
		Bounds:    uimath.NewRect(10, 10, 100, 40),
		FillColor: uimath.ColorHex("#0052d9"),
		Corners:   uimath.CornersAll(6),
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.DrawRect(cmd, 0, 1)
	}
}

func BenchmarkCommandBuffer_Reset(b *testing.B) {
	buf := render.NewCommandBuffer()
	for j := 0; j < 200; j++ {
		buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(0, float32(j*20), 400, 20),
			FillColor: uimath.ColorWhite,
		}, 0, 1)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Reset()
		for j := 0; j < 200; j++ {
			buf.DrawRect(render.RectCmd{
				Bounds:    uimath.NewRect(0, float32(j*20), 400, 20),
				FillColor: uimath.ColorWhite,
			}, 0, 1)
		}
	}
}

func BenchmarkCommandBuffer_PushPopClip(b *testing.B) {
	buf := render.NewCommandBuffer()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.PushClip(uimath.NewRect(0, 0, 400, 200))
		buf.PopClip()
	}
}

// ---- Tree Operations Benchmarks ----

func BenchmarkTreeCreateElement(b *testing.B) {
	tree := core.NewTree()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tree.CreateElement(core.TypeDiv)
	}
}

func BenchmarkTreeSetLayout(b *testing.B) {
	tree := core.NewTree()
	id := tree.CreateElement(core.TypeDiv)
	lr := core.LayoutResult{
		Bounds: uimath.NewRect(10, 20, 300, 200),
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tree.SetLayout(id, lr)
	}
}

func BenchmarkTreeMarkDirty(b *testing.B) {
	tree := core.NewTree()
	parent := tree.CreateElement(core.TypeDiv)
	child := tree.CreateElement(core.TypeDiv)
	tree.AppendChild(parent, child)
	tree.ClearAllDirty()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tree.MarkDirty(child)
	}
}

func BenchmarkTreeNeedsLayout(b *testing.B) {
	tree := core.NewTree()
	tree.ClearAllDirty()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tree.NeedsLayout()
	}
}

// ---- Layout Engine Benchmarks ----

func BenchmarkLayoutStyle_Default(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = layout.DefaultStyle()
	}
}

// ---- Composite Widget Benchmarks ----

func BenchmarkDrawFeedCard(b *testing.B) {
	// Simulates a tweet card: avatar + name + handle + text + action bar
	env := newBenchEnv()
	card := widget.NewDiv(env.tree, env.cfg)
	card.SetBgColor(uimath.ColorWhite)
	card.SetBorderRadius(8)
	env.setLayout(card, 0, 0, 600, 180)

	av := widget.NewAvatar(env.tree, env.cfg)
	av.SetContent("Alice")
	av.SetShape(widget.AvatarCircle)
	env.setLayout(av, 12, 12, 40, 40)
	card.AppendChild(av)

	name := widget.NewText(env.tree, "Alice Johnson", env.cfg)
	env.setLayout(name, 64, 12, 200, 20)
	card.AppendChild(name)

	handle := widget.NewText(env.tree, "@alice · 5m", env.cfg)
	env.setLayout(handle, 64, 34, 200, 16)
	card.AppendChild(handle)

	body := widget.NewText(env.tree, "This is a long tweet with some content that wraps across multiple lines in the feed view.", env.cfg)
	env.setLayout(body, 64, 56, 524, 60)
	card.AppendChild(body)

	for j := 0; j < 4; j++ {
		btn := widget.NewText(env.tree, "0", env.cfg)
		env.setLayout(btn, float32(64+j*120), 130, 80, 20)
		card.AppendChild(btn)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		env.buf.Reset()
		card.Draw(env.buf)
	}
}

func BenchmarkDrawFeedTimeline(b *testing.B) {
	for _, n := range []int{5, 15, 50} {
		b.Run(fmt.Sprintf("%d_cards", n), func(b *testing.B) {
			env := newBenchEnv()
			timeline := widget.NewDiv(env.tree, env.cfg)
			timeline.SetScrollable(true)
			env.setLayout(timeline, 0, 0, 600, 800) // viewport

			for j := 0; j < n; j++ {
				card := widget.NewDiv(env.tree, env.cfg)
				card.SetBgColor(uimath.ColorWhite)
				env.setLayout(card, 0, float32(j*190), 600, 180)

				av := widget.NewAvatar(env.tree, env.cfg)
				av.SetContent("User")
				env.setLayout(av, 12, float32(j*190+12), 40, 40)
				card.AppendChild(av)

				name := widget.NewText(env.tree, "Username", env.cfg)
				env.setLayout(name, 64, float32(j*190+12), 200, 20)
				card.AppendChild(name)

				body := widget.NewText(env.tree, "Tweet body text for benchmark testing purposes.", env.cfg)
				env.setLayout(body, 64, float32(j*190+40), 524, 40)
				card.AppendChild(body)

				timeline.AppendChild(card)
			}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				env.buf.Reset()
				timeline.Draw(env.buf)
			}
		})
	}
}

func BenchmarkDrawTable_10x5(b *testing.B) {
	env := newBenchEnv()
	cols := []widget.TableColumn{
		{Title: "Name", Width: 100},
		{Title: "Age", Width: 60},
		{Title: "Email", Width: 200},
		{Title: "City", Width: 100},
		{Title: "Status", Width: 80},
	}
	tbl := widget.NewTable(env.tree, cols, env.cfg)
	rows := make([][]string, 10)
	for i := range rows {
		rows[i] = []string{"Alice", "25", "alice@example.com", "Beijing", "Active"}
	}
	tbl.SetRows(rows)
	env.setLayout(tbl, 0, 0, 540, 400)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		env.buf.Reset()
		tbl.Draw(env.buf)
	}
}

func BenchmarkDrawList_50Items(b *testing.B) {
	env := newBenchEnv()
	l := widget.NewList(env.tree, env.cfg)
	for j := 0; j < 50; j++ {
		item := widget.NewText(env.tree, "List item", env.cfg)
		env.setLayout(item, 0, float32(j*32), 300, 32)
		l.AppendChild(item)
	}
	env.setLayout(l, 0, 0, 300, 400)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		env.buf.Reset()
		l.Draw(env.buf)
	}
}
