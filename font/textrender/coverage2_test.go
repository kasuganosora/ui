package textrender

import (
	"testing"

	"github.com/kasuganosora/ui/font"
	"github.com/kasuganosora/ui/font/atlas"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// Additional tests for 90%+ coverage.

func TestShape(t *testing.T) {
	r, id := setup(t)
	defer r.Destroy()

	runs := r.Shape("Hello", font.ShapeOptions{FontID: id, FontSize: 16})
	if len(runs) != 1 {
		t.Fatalf("expected 1 run, got %d", len(runs))
	}
	if len(runs[0].Glyphs) != 5 {
		t.Errorf("expected 5 glyphs, got %d", len(runs[0].Glyphs))
	}
}

func TestShaper(t *testing.T) {
	r, _ := setup(t)
	defer r.Destroy()

	s := r.Shaper()
	if s == nil {
		t.Error("Shaper() should not return nil")
	}
}

func TestDrawRunEmpty(t *testing.T) {
	r, id := setup(t)
	defer r.Destroy()

	buf := render.NewCommandBuffer()
	// Draw with runs that have no glyphs
	r.DrawRuns(buf, []font.GlyphRun{{FontID: id, FontSize: 16, Glyphs: nil}}, DrawOptions{Opacity: 1})
	if buf.Len() != 0 {
		t.Errorf("expected 0 commands for empty run, got %d", buf.Len())
	}
}

func TestDrawRunsMultiple(t *testing.T) {
	r, id := setup(t)
	defer r.Destroy()

	runs := r.Shape("A\nB", font.ShapeOptions{FontID: id, FontSize: 16})
	buf := render.NewCommandBuffer()
	r.DrawRuns(buf, runs, DrawOptions{
		Color:   uimath.Color{R: 1, A: 1},
		Opacity: 1,
	})
	if buf.Len() != 2 {
		t.Errorf("expected 2 commands for 2 runs, got %d", buf.Len())
	}
}

func TestAtlasAccessor(t *testing.T) {
	r, _ := setup(t)
	defer r.Destroy()

	a := r.Atlas()
	if a == nil {
		t.Error("Atlas() should not return nil")
	}
}

func TestNewWithSDF(t *testing.T) {
	engine := newMockEngine()
	mgr := font.NewManager(engine)
	mgr.Register("Test", font.WeightRegular, font.StyleNormal, nil)
	a := atlas.New(atlas.Options{Width: 128, Height: 128})

	r := New(Options{Manager: mgr, Atlas: a, SDF: true})
	defer r.Destroy()

	if r.sdf != true {
		t.Error("SDF should be enabled")
	}
}
