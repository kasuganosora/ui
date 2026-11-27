package ui

import (
	"os"
	"testing"

	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/font"
	"github.com/kasuganosora/ui/font/ttf"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
	"github.com/kasuganosora/ui/widget"
)

const captionTestFont = "/System/Library/Fonts/PingFang.ttc"

// The optional weight capability must be wired all the way from the engine to
// the widget-facing drawer, without disturbing the default family's regular
// face lookup.
func TestConfigureCaptionFontRegistersMediumFace(t *testing.T) {
	if _, err := os.Stat(captionTestFont); err != nil {
		t.Skipf("system CJK collection not available: %v", err)
	}
	mgr := font.NewManager(ttf.New())
	regular, err := mgr.RegisterFile("Default", font.WeightRegular, font.StyleNormal, captionTestFont)
	if err != nil {
		t.Fatalf("register regular: %v", err)
	}
	adapter := &textDrawerAdapter{fontID: regular}

	id := ConfigureCaptionFont(mgr, captionTestFont, adapter)
	if id == font.InvalidFontID {
		t.Fatal("caption face was not registered")
	}
	if id == regular || adapter.captionFontID != id {
		t.Errorf("caption drawer not wired: id=%d regular=%d adapter=%d", id, regular, adapter.captionFontID)
	}
	resolved, ok := mgr.Resolve(font.Properties{
		Family: "Caption", Weight: font.WeightMedium, Style: font.StyleNormal,
	})
	if !ok || resolved != id {
		t.Errorf("Caption/Medium resolved to %d (ok=%v), want %d", resolved, ok, id)
	}
	body, ok := mgr.Resolve(font.Properties{
		Family: "Default", Weight: font.WeightRegular, Style: font.StyleNormal,
	})
	if !ok || body != regular {
		t.Errorf("body text no longer resolves to the regular face: %d (ok=%v)", body, ok)
	}
}

type stubTextDrawer struct{}

func (stubTextDrawer) DrawText(_ *render.CommandBuffer, _ string, _, _, _, _ float32, _ uimath.Color, _ float32) {
}
func (stubTextDrawer) DrawTextZ(_ *render.CommandBuffer, _ string, _, _, _, _ float32, _ uimath.Color, _ float32, _ int32) {
}
func (stubTextDrawer) LineHeight(size float32) float32 { return size * 1.2 }
func (stubTextDrawer) MeasureText(text string, size float32) float32 {
	return float32(len([]rune(text))) * size * 0.6
}

// Every degradation path is a visual downgrade to the regular face, never an
// error and never a partially configured drawer.
func TestConfigureCaptionFontDegradesSafely(t *testing.T) {
	mgr := font.NewManager(ttf.New())
	adapter := &textDrawerAdapter{}
	if got := ConfigureCaptionFont(nil, captionTestFont, adapter); got != font.InvalidFontID {
		t.Errorf("nil manager returned %d", got)
	}
	if got := ConfigureCaptionFont(mgr, "", adapter); got != font.InvalidFontID {
		t.Errorf("empty font path returned %d", got)
	}
	if got := ConfigureCaptionFont(mgr, "/no/such/font.ttc", adapter); got != font.InvalidFontID {
		t.Errorf("missing font file returned %d", got)
	}
	if adapter.captionFontID != font.InvalidFontID {
		t.Error("failed registration left a stale caption font on the drawer")
	}
	if got := ConfigureCaptionFont(mgr, captionTestFont, stubTextDrawer{}); got != font.InvalidFontID {
		t.Errorf("foreign drawer returned %d", got)
	}
}

// §15 Dynamic Type is an App-level switch: it must reach the shared Config
// and must not double-apply the scale to captions.
func TestSetTextScaleReachesSharedConfig(t *testing.T) {
	app := &App{cfg: widget.DefaultConfig(), tree: core.NewTree()}
	app.SetTextScale(2)
	if app.TextScale() != 2 || app.cfg.FontSize != 28 || app.cfg.FontSizeSm != 24 {
		t.Fatalf("scale did not reach Config: scale=%g body=%g small=%g",
			app.TextScale(), app.cfg.FontSize, app.cfg.FontSizeSm)
	}
	if app.cfg.CaptionSize() != 24 {
		t.Errorf("caption double-scaled to %g, want 24", app.cfg.CaptionSize())
	}
	if app.cfg.SpaceMD != 32 {
		t.Errorf("spacing did not scale: SpaceMD=%g", app.cfg.SpaceMD)
	}
	app.SetTextScale(0)
	if app.TextScale() != 2 {
		t.Error("invalid scale was accepted")
	}
	app.SetTextScale(1)
	if app.cfg.FontSize != 14 || app.cfg.CaptionSize() != 12 || app.cfg.SpaceMD != 16 {
		t.Error("returning to scale 1 did not restore the type system")
	}
}
