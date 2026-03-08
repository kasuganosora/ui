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

func TestContextNilConfig(t *testing.T) {
	buf := render.NewCommandBuffer()
	ctx := NewContext(buf, nil)
	ctx.Begin("test", 0, 0, 100, 100)
	ctx.Text("test")
	ctx.End()
	if buf.Len() == 0 {
		t.Error("expected render commands with nil config")
	}
}

func TestContextButton(t *testing.T) {
	buf := render.NewCommandBuffer()
	cfg := widget.DefaultConfig()
	ctx := NewContext(buf, cfg)

	ctx.Begin("panel", 0, 0, 200, 300)
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

func TestContextButtonHover(t *testing.T) {
	buf := render.NewCommandBuffer()
	cfg := widget.DefaultConfig()
	ctx := NewContext(buf, cfg)

	ctx.Begin("panel", 0, 0, 200, 300)
	// Mouse inside button bounds, not clicking
	btnY := cfg.SpaceSM + cfg.ButtonHeight/2
	ctx.SetInput(100, btnY, false, false, false)
	clicked := ctx.Button("Test")
	ctx.End()

	if clicked {
		t.Error("button should not be clicked (no click)")
	}
}

func TestContextButtonHoverAndDown(t *testing.T) {
	buf := render.NewCommandBuffer()
	cfg := widget.DefaultConfig()
	ctx := NewContext(buf, cfg)

	ctx.Begin("panel", 0, 0, 200, 300)
	// Mouse inside button bounds, mouse down
	btnY := cfg.SpaceSM + cfg.ButtonHeight/2
	ctx.SetInput(100, btnY, true, false, false)
	ctx.Button("Test")
	ctx.End()
}

func TestContextButtonClick(t *testing.T) {
	buf := render.NewCommandBuffer()
	cfg := widget.DefaultConfig()
	ctx := NewContext(buf, cfg)

	ctx.Begin("panel", 0, 0, 200, 300)
	btnY := cfg.SpaceSM + cfg.ButtonHeight/2
	ctx.SetInput(100, btnY, false, true, false)
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

func TestContextCheckboxClick(t *testing.T) {
	buf := render.NewCommandBuffer()
	cfg := widget.DefaultConfig()
	ctx := NewContext(buf, cfg)

	ctx.Begin("panel", 0, 0, 200, 300)
	// Click on checkbox box area
	lineH := cfg.FontSize * 1.5
	boxY := cfg.SpaceSM + (lineH-16)/2 + 8 // center of 16px box
	ctx.SetInput(cfg.SpaceSM+8, boxY, false, true, false)
	checked := ctx.Checkbox("cb1", "Option A")
	ctx.End()

	if !checked {
		t.Error("checkbox should be checked after click")
	}
}

func TestContextCheckboxToggle(t *testing.T) {
	buf := render.NewCommandBuffer()
	cfg := widget.DefaultConfig()
	ctx := NewContext(buf, cfg)

	lineH := cfg.FontSize * 1.5
	boxY := cfg.SpaceSM + (lineH-16)/2 + 8

	// First click — check
	ctx.Begin("panel", 0, 0, 200, 300)
	ctx.SetInput(cfg.SpaceSM+8, boxY, false, true, false)
	checked := ctx.Checkbox("toggle", "Toggle")
	ctx.End()

	if !checked {
		t.Fatal("expected checked after first click")
	}

	// Second click — uncheck
	buf.Reset()
	ctx.Begin("panel", 0, 0, 200, 300)
	ctx.SetInput(cfg.SpaceSM+8, boxY, false, true, false)
	checked = ctx.Checkbox("toggle", "Toggle")
	ctx.End()

	if checked {
		t.Error("expected unchecked after second click")
	}
}

func TestContextCheckboxHover(t *testing.T) {
	buf := render.NewCommandBuffer()
	cfg := widget.DefaultConfig()
	ctx := NewContext(buf, cfg)

	ctx.Begin("panel", 0, 0, 200, 300)
	lineH := cfg.FontSize * 1.5
	boxY := cfg.SpaceSM + (lineH-16)/2 + 8
	// Hover over checkbox but don't click
	ctx.SetInput(cfg.SpaceSM+8, boxY, false, false, false)
	ctx.Checkbox("hover", "Hover Test")
	ctx.End()
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
	sliderY := cfg.SpaceSM + 10
	ctx.SetInput(100, sliderY, true, false, false)
	val := ctx.Slider("speed", 0, 100)
	ctx.End()

	if val == 0 {
		t.Log("slider value should have changed if mouse is on track")
	}
}

func TestContextSliderDragMiddle(t *testing.T) {
	buf := render.NewCommandBuffer()
	cfg := widget.DefaultConfig()
	ctx := NewContext(buf, cfg)

	w := float32(200)
	padding := cfg.SpaceSM
	trackW := w - padding*2

	ctx.Begin("panel", 0, 0, w, 300)
	sliderY := padding + 10
	// Click at 50% of the track (MouseClicked=true to initiate drag)
	ctx.SetInput(padding+trackW/2, sliderY, true, true, false)
	val := ctx.Slider("mid", 0, 100)
	ctx.End()

	// Should be approximately 50
	if val < 40 || val > 60 {
		t.Errorf("expected ~50 at midpoint, got %g", val)
	}
}

func TestContextSliderEqualMinMax(t *testing.T) {
	buf := render.NewCommandBuffer()
	cfg := widget.DefaultConfig()
	ctx := NewContext(buf, cfg)

	ctx.Begin("panel", 0, 0, 200, 300)
	ctx.SetInput(0, 0, false, false, false)
	val := ctx.Slider("eq", 50, 50)
	ctx.End()

	if val != 50 {
		t.Errorf("expected 50 for equal min/max, got %g", val)
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

func TestContextProgressBarZero(t *testing.T) {
	buf := render.NewCommandBuffer()
	cfg := widget.DefaultConfig()
	ctx := NewContext(buf, cfg)

	ctx.Begin("panel", 0, 0, 200, 300)
	ctx.ProgressBar(0)
	ctx.End()

	// ratio=0 → only track, no fill
	if buf.Len() < 2 {
		t.Errorf("expected at least 2 commands, got %d", buf.Len())
	}
}

func TestContextProgressBarClamp(t *testing.T) {
	buf := render.NewCommandBuffer()
	cfg := widget.DefaultConfig()
	ctx := NewContext(buf, cfg)

	ctx.Begin("panel", 0, 0, 200, 300)
	ctx.ProgressBar(-0.5)  // should clamp to 0
	ctx.ProgressBar(1.5)   // should clamp to 1
	ctx.End()
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

func TestContextSpace(t *testing.T) {
	buf := render.NewCommandBuffer()
	cfg := widget.DefaultConfig()
	ctx := NewContext(buf, cfg)

	ctx.Begin("panel", 0, 0, 200, 300)
	ctx.Space(20)
	ctx.Text("After space")
	ctx.End()
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
