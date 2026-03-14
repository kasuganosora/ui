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
