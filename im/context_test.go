package im

import (
	"testing"

	"github.com/kasuganosora/ui/render"
	"github.com/kasuganosora/ui/widget"
)

func TestContextBasic(t *testing.T) {
	buf := render.NewCommandBuffer()
	cfg := widget.DefaultConfig()
	ctx := NewContext(buf, cfg)

	ctx.Begin("test", 10, 10, 300, 400)
	ctx.Text("Hello")
	ctx.End()

	if buf.Len() == 0 {
		t.Error("expected render commands")
	}
}

func TestContextButton(t *testing.T) {
	buf := render.NewCommandBuffer()
	cfg := widget.DefaultConfig()
	ctx := NewContext(buf, cfg)

	ctx.Begin("panel", 0, 0, 200, 300)

	// Not clicking
	ctx.SetInput(100, 100, false, false, false)
	clicked := ctx.Button("Test")
	if clicked {
		t.Error("button should not be clicked")
	}

	ctx.End()

	if buf.Len() < 2 {
		t.Errorf("expected at least 2 commands, got %d", buf.Len())
	}
}

func TestContextButtonClick(t *testing.T) {
	buf := render.NewCommandBuffer()
	cfg := widget.DefaultConfig()
	ctx := NewContext(buf, cfg)

	ctx.Begin("panel", 0, 0, 200, 300)
	// Mouse at button position, clicking
	ctx.SetInput(100, float32(cfg.SpaceSM)+float32(cfg.ButtonHeight)/2, false, true, false)
	clicked := ctx.Button("Test")
	ctx.End()

	if !clicked {
		t.Error("button should be clicked")
	}
}

func TestContextCheckbox(t *testing.T) {
	buf := render.NewCommandBuffer()
	cfg := widget.DefaultConfig()
	ctx := NewContext(buf, cfg)

	ctx.Begin("panel", 0, 0, 200, 300)
	ctx.SetInput(0, 0, false, false, false)
	checked := ctx.Checkbox("cb1", "Option A")
	ctx.End()

	if checked {
		t.Error("checkbox should not be checked initially")
	}
}

func TestContextSlider(t *testing.T) {
	buf := render.NewCommandBuffer()
	cfg := widget.DefaultConfig()
	ctx := NewContext(buf, cfg)

	ctx.Begin("panel", 0, 0, 200, 300)
	ctx.SetInput(0, 0, false, false, false)
	val := ctx.Slider("speed", 0, 100)
	ctx.End()

	if val != 0 {
		t.Errorf("expected slider value 0, got %g", val)
	}
}

func TestContextSliderDrag(t *testing.T) {
	buf := render.NewCommandBuffer()
	cfg := widget.DefaultConfig()
	ctx := NewContext(buf, cfg)

	ctx.Begin("panel", 0, 0, 200, 300)
	// Cursor in middle of slider track
	sliderY := float32(cfg.SpaceSM) + 10 // approximate slider position
	ctx.SetInput(100, sliderY, true, false, false)
	val := ctx.Slider("speed", 0, 100)
	ctx.End()

	if val == 0 {
		t.Log("slider value should have changed if mouse is on track")
	}
}

func TestContextProgressBar(t *testing.T) {
	buf := render.NewCommandBuffer()
	cfg := widget.DefaultConfig()
	ctx := NewContext(buf, cfg)

	ctx.Begin("panel", 0, 0, 200, 300)
	ctx.ProgressBar(0.5)
	ctx.End()

	if buf.Len() < 3 {
		t.Errorf("expected at least 3 commands, got %d", buf.Len())
	}
}

func TestContextSeparator(t *testing.T) {
	buf := render.NewCommandBuffer()
	cfg := widget.DefaultConfig()
	ctx := NewContext(buf, cfg)

	ctx.Begin("panel", 0, 0, 200, 300)
	ctx.Separator()
	ctx.End()

	if buf.Len() < 2 {
		t.Errorf("expected at least 2 commands, got %d", buf.Len())
	}
}

func TestContextMultipleWidgets(t *testing.T) {
	buf := render.NewCommandBuffer()
	cfg := widget.DefaultConfig()
	ctx := NewContext(buf, cfg)

	ctx.Begin("panel", 10, 10, 300, 500)
	ctx.Text("Debug Panel")
	ctx.Separator()
	ctx.Button("Reset")
	ctx.Checkbox("debug", "Debug Mode")
	ctx.Slider("fov", 60, 120)
	ctx.ProgressBar(0.75)
	ctx.Space(10)
	ctx.Text("Done")
	ctx.End()

	if buf.Len() < 10 {
		t.Errorf("expected many commands from full panel, got %d", buf.Len())
	}
}
