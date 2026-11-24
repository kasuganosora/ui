package godot

import (
	"testing"

	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/event"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/widget"
)

func TestUI_NewAndDestroy(t *testing.T) {
	ui, err := NewUI(UIOptions{Width: 400, Height: 300})
	if err != nil {
		t.Fatal(err)
	}
	defer ui.Destroy()

	if ui.Tree() == nil {
		t.Fatal("tree is nil")
	}
	if ui.Config() == nil {
		t.Fatal("config is nil")
	}
}

func TestUI_SetRootAndFrame(t *testing.T) {
	ui, err := NewUI(UIOptions{Width: 400, Height: 300})
	if err != nil {
		t.Fatal(err)
	}
	defer ui.Destroy()

	root := widget.NewDiv(ui.Tree(), ui.Config())
	root.SetBgColor(uimath.NewColor(1, 0, 0, 1))
	ui.SetRoot(root)

	ui.Frame(0.016)

	pixels := ui.Pixels()
	if len(pixels) == 0 {
		t.Fatal("no pixels after frame")
	}

	fw, fh := ui.FramebufferSize()
	if fw != 400 || fh != 300 {
		t.Errorf("framebuffer size: %dx%d, want 400x300", fw, fh)
	}
}

func TestUI_Resize(t *testing.T) {
	ui, err := NewUI(UIOptions{Width: 400, Height: 300})
	if err != nil {
		t.Fatal(err)
	}
	defer ui.Destroy()

	root := widget.NewDiv(ui.Tree(), ui.Config())
	ui.SetRoot(root)

	ui.Resize(800, 600)
	ui.Frame(0.016)

	fw, fh := ui.FramebufferSize()
	if fw != 800 || fh != 600 {
		t.Errorf("after resize: %dx%d, want 800x600", fw, fh)
	}
}

func TestUI_InjectClick(t *testing.T) {
	ui, err := NewUI(UIOptions{Width: 400, Height: 300})
	if err != nil {
		t.Fatal(err)
	}
	defer ui.Destroy()

	root := widget.NewDiv(ui.Tree(), ui.Config())
	root.SetBgColor(uimath.NewColor(0, 0, 1, 1))
	ui.SetRoot(root)

	clicked := false
	ui.Tree().AddHandler(root.ElementID(), event.MouseClick, func(e *event.Event) {
		clicked = true
	})

	// Set layout so hit testing works
	ui.Tree().SetLayout(root.ElementID(), core.LayoutResult{
		Bounds: uimath.NewRect(0, 0, 400, 300),
	})

	// Inject click
	ui.InjectMouseClick(200, 150, event.MouseButtonLeft)
	ui.Frame(0.016)

	if !clicked {
		t.Error("click handler was not called")
	}
}

func TestInjectMouseClickActivatesOnce(t *testing.T) {
	ui, err := NewUI(UIOptions{Width: 400, Height: 300})
	if err != nil {
		t.Fatal(err)
	}
	defer ui.Destroy()
	btn := widget.NewButton(ui.Tree(), "go", ui.Config())
	ui.SetRoot(btn)
	ui.Tree().SetLayout(ui.Tree().Root(), core.LayoutResult{Bounds: uimath.NewRect(0, 0, 400, 300)})
	ui.Tree().SetLayout(btn.ElementID(), core.LayoutResult{Bounds: uimath.NewRect(10, 10, 80, 32)})
	n := 0
	btn.OnClick(func() { n++ })
	ui.InjectMouseClick(20, 20, event.MouseButtonLeft)
	ui.Frame(0.016)
	if n != 1 {
		t.Fatalf("activations %d", n)
	}
	other := widget.NewButton(ui.Tree(), "other", ui.Config())
	ui.Tree().AppendChild(ui.Tree().Root(), other.ElementID())
	ui.Tree().SetLayout(other.ElementID(), core.LayoutResult{Bounds: uimath.NewRect(120, 10, 80, 32)})
	ui.InjectEvent(event.Event{Type: event.MouseDown, GlobalX: 20, GlobalY: 20})
	ui.InjectEvent(event.Event{Type: event.MouseUp, GlobalX: 140, GlobalY: 20})
	ui.Frame(0.016)
	if n != 1 {
		t.Fatalf("up on another control synthesized a click, activations %d", n)
	}
}

func TestDownUpFrameThenStandaloneClick(t *testing.T) {
	ui, err := NewUI(UIOptions{Width: 400, Height: 200})
	if err != nil {
		t.Fatal(err)
	}
	defer ui.Destroy()
	root := widget.NewDiv(ui.Tree(), ui.Config())
	ui.SetRoot(root)
	a := widget.NewButton(ui.Tree(), "a", ui.Config())
	b := widget.NewButton(ui.Tree(), "b", ui.Config())
	root.AppendChild(a)
	root.AppendChild(b)
	ui.Tree().SetLayout(a.ElementID(), core.LayoutResult{Bounds: uimath.NewRect(10, 10, 80, 32)})
	ui.Tree().SetLayout(b.ElementID(), core.LayoutResult{Bounds: uimath.NewRect(120, 10, 80, 32)})
	na, nb := 0, 0
	a.OnClick(func() { na++ })
	b.OnClick(func() { nb++ })
	ui.InjectEvent(event.Event{Type: event.MouseDown, GlobalX: 20, GlobalY: 20})
	ui.Frame(0.016)
	ui.InjectEvent(event.Event{Type: event.MouseUp, GlobalX: 20, GlobalY: 20})
	ui.Frame(0.016)
	ui.InjectEvent(event.Event{Type: event.MouseClick, GlobalX: 140, GlobalY: 20})
	ui.Frame(0.016)
	ui.InjectMouseClick(20, 20, event.MouseButtonLeft)
	ui.Frame(0.016)
	if na != 2 || nb != 1 {
		t.Fatalf("activations a %d b %d", na, nb)
	}
}

func TestSoftwarePaintsOpenedDialogAboveDate(t *testing.T) {
	ui, err := NewUI(UIOptions{Width: 800, Height: 600})
	if err != nil {
		t.Fatal(err)
	}
	defer ui.Destroy()
	root := widget.NewDiv(ui.Tree(), ui.Config())
	ui.SetRoot(root)
	dlg := widget.NewDialog(ui.Tree(), "early", ui.Config())
	dp := widget.NewDatePicker(ui.Tree(), ui.Config())
	dp.SetDate(2026, 1, 1)
	root.AppendChild(dlg)
	root.AppendChild(dp)
	ui.Tree().SetLayout(dlg.ElementID(), core.LayoutResult{Bounds: uimath.NewRect(0, 0, 800, 600)})
	ui.Tree().SetLayout(dp.ElementID(), core.LayoutResult{Bounds: uimath.NewRect(10, 10, 200, 32)})
	dp.SetOpen(true)
	dlg.Open()
	ui.Frame(0.016)
	r, _, _, _ := pixelAt(ui, 154, 162)
	if r > 200 {
		t.Fatalf("software raster left the date popup above the later dialog, red %d", r)
	}

	ui2, err := NewUI(UIOptions{Width: 800, Height: 600})
	if err != nil {
		t.Fatal(err)
	}
	defer ui2.Destroy()
	root2 := widget.NewDiv(ui2.Tree(), ui2.Config())
	ui2.SetRoot(root2)
	host := widget.NewDialog(ui2.Tree(), "host", ui2.Config())
	inner := widget.NewDatePicker(ui2.Tree(), ui2.Config())
	inner.SetDate(2026, 1, 1)
	root2.AppendChild(host)
	host.AppendChild(inner)
	ui2.Tree().SetLayout(host.ElementID(), core.LayoutResult{Bounds: uimath.NewRect(0, 0, 800, 600)})
	ui2.Tree().SetLayout(inner.ElementID(), core.LayoutResult{Bounds: uimath.NewRect(30, 40, 200, 32)})
	host.Open()
	inner.SetOpen(true)
	ui2.Frame(0.016)
	r, _, _, _ = pixelAt(ui2, 30+4*32+16, 40+32+4+36+2*32+16)
	if r < 240 {
		t.Fatalf("software raster hid the dialog-owned calendar, red %d", r)
	}
}

func pixelAt(ui *UI, x, y int) (byte, byte, byte, byte) {
	w, h := ui.FramebufferSize()
	px := ui.Pixels()
	if x < 0 || y < 0 || x >= w || y >= h {
		return 0, 0, 0, 0
	}
	i := (y*w + x) * 4
	return px[i], px[i+1], px[i+2], px[i+3]
}

func TestUI_InjectMouseMove(t *testing.T) {
	ui, err := NewUI(UIOptions{Width: 400, Height: 300})
	if err != nil {
		t.Fatal(err)
	}
	defer ui.Destroy()

	root := widget.NewDiv(ui.Tree(), ui.Config())
	ui.SetRoot(root)

	// Set layout so hit testing works
	ui.Tree().SetLayout(root.ElementID(), core.LayoutResult{
		Bounds: uimath.NewRect(0, 0, 400, 300),
	})

	entered := false
	ui.Tree().AddHandler(root.ElementID(), event.MouseEnter, func(e *event.Event) {
		entered = true
	})

	ui.InjectMouseMove(200, 150)
	ui.Frame(0.016)

	if !entered {
		t.Error("mouse enter handler was not called")
	}
}

func TestUI_MultipleInstances(t *testing.T) {
	ui1, err := NewUI(UIOptions{Width: 200, Height: 200})
	if err != nil {
		t.Fatal(err)
	}
	defer ui1.Destroy()

	ui2, err := NewUI(UIOptions{Width: 300, Height: 300})
	if err != nil {
		t.Fatal(err)
	}
	defer ui2.Destroy()

	r1 := widget.NewDiv(ui1.Tree(), ui1.Config())
	r1.SetBgColor(uimath.NewColor(1, 0, 0, 1))
	ui1.SetRoot(r1)

	r2 := widget.NewDiv(ui2.Tree(), ui2.Config())
	r2.SetBgColor(uimath.NewColor(0, 1, 0, 1))
	ui2.SetRoot(r2)

	ui1.Frame(0.016)
	ui2.Frame(0.016)

	fw1, fh1 := ui1.FramebufferSize()
	fw2, fh2 := ui2.FramebufferSize()

	if fw1 != 200 || fh1 != 200 {
		t.Errorf("ui1 size: %dx%d", fw1, fh1)
	}
	if fw2 != 300 || fh2 != 300 {
		t.Errorf("ui2 size: %dx%d", fw2, fh2)
	}
}

func TestExport_CreateDestroyInstance(t *testing.T) {
	id, err := CreateInstance(400, 300, 1.0)
	if err != nil {
		t.Fatal(err)
	}
	if id <= 0 {
		t.Fatal("invalid instance id")
	}

	ui := GetInstance(id)
	if ui == nil {
		t.Fatal("instance not found")
	}

	InstanceFrame(id, 0.016)
	InstanceResize(id, 800, 600)
	InstanceFrame(id, 0.016)

	fw, fh := InstanceFramebufferSize(id)
	if fw != 800 || fh != 600 {
		t.Errorf("instance size after resize: %dx%d", fw, fh)
	}

	DestroyInstance(id)

	if GetInstance(id) != nil {
		t.Error("instance still exists after destroy")
	}
}

func TestExport_InjectEvents(t *testing.T) {
	id, err := CreateInstance(400, 300, 1.0)
	if err != nil {
		t.Fatal(err)
	}
	defer DestroyInstance(id)

	// These should not panic
	InstanceInjectMouseMove(id, 100, 100)
	InstanceInjectMouseClick(id, 100, 100, 0)
	InstanceInjectScroll(id, 100, 100, 0, -10)
	InstanceInjectKeyDown(id, int(event.KeyA), 0)
	InstanceInjectKeyUp(id, int(event.KeyA), 0)
	InstanceInjectChar(id, 'A')
	InstanceFrame(id, 0.016)
}

func BenchmarkUI_Frame(b *testing.B) {
	ui, err := NewUI(UIOptions{Width: 800, Height: 600})
	if err != nil {
		b.Fatal(err)
	}
	defer ui.Destroy()

	root := widget.NewDiv(ui.Tree(), ui.Config())
	root.SetBgColor(uimath.NewColor(0.1, 0.1, 0.2, 1))

	for i := 0; i < 10; i++ {
		child := widget.NewDiv(ui.Tree(), ui.Config())
		child.SetBgColor(uimath.NewColor(0.3, 0.5, 0.8, 0.9))
		root.AppendChild(child)
	}
	ui.SetRoot(root)
	ui.Frame(0.016) // warm up

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ui.Frame(0.016)
	}
}
