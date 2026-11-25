package widget

import (
	"fmt"
	"testing"

	"github.com/kasuganosora/ui/core"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// recordingOverlayHost stands in for the application shell's OS window host.
type recordingOverlayHost struct {
	opened int
	closed int
	fail   bool
	key    Widget
	draw   func(*Config, *render.CommandBuffer)
	x, y   float32
	w, h   float32
}

func (h *recordingOverlayHost) OpenOverlay(key Widget, draw func(*Config, *render.CommandBuffer), x, y, w, winH float32) error {
	if h.fail {
		return fmt.Errorf("overlay: no native window available")
	}
	h.opened++
	h.key, h.draw = key, draw
	h.x, h.y, h.w, h.h = x, y, w, winH
	return nil
}

func (h *recordingOverlayHost) CloseOverlay(key Widget) {
	h.closed++
	h.key = key
}

func newOverlayTooltip(t *testing.T) (*core.Tree, *Tooltip, *Config, *recordingOverlayHost) {
	t.Helper()
	tree := newTestTree()
	anchor := tree.CreateElement(core.TypeButton)
	tree.SetLayout(anchor, core.LayoutResult{Bounds: uimath.NewRect(100, 200, 120, 32)})

	cfg := DefaultConfig()
	cfg.Window = &mockWindow{}
	cfg.TextRenderer = &zProbeDrawer{}
	host := &recordingOverlayHost{}
	cfg.OverlayHost = host

	tt := NewTooltip(tree, "提示文本", anchor, cfg)
	return tree, tt, cfg, host
}

// TestTooltipOpensIndependentOverlay verifies that Show() asks the host for a
// real auxiliary window, sized from the label metrics and placed relative to
// the anchor, and that Hide() releases it again.
func TestTooltipOpensIndependentOverlay(t *testing.T) {
	_, tt, cfg, host := newOverlayTooltip(t)

	tt.Show()
	if host.opened != 1 {
		t.Fatalf("OpenOverlay called %d times, want 1", host.opened)
	}
	if !tt.overlayOpen {
		t.Fatal("tooltip did not record an open auxiliary window")
	}
	if host.key != tt {
		t.Fatal("host received the wrong widget key")
	}

	w, h := tt.measure(cfg)
	if host.w != w || host.h != h {
		t.Fatalf("window size = (%v,%v), want (%v,%v)", host.w, host.h, w, h)
	}
	wantX := float32(100) + (120-w)/2
	wantY := float32(200) - h - 8 // PlacementTop
	if host.x != wantX || host.y != wantY {
		t.Fatalf("window position = (%v,%v), want (%v,%v)", host.x, host.y, wantX, wantY)
	}

	// The host paints with its own config/command buffer; the bubble and its
	// label must both be emitted there.
	buf := render.NewCommandBuffer()
	host.draw(cfg, buf)
	var rects, texts int
	for _, c := range buf.Commands() {
		switch c.Type {
		case render.CmdRect:
			rects++
		case render.CmdText:
			texts++
		}
	}
	if rects == 0 || texts == 0 {
		t.Fatalf("overlay draw produced rects=%d texts=%d, want both", rects, texts)
	}

	tt.Hide()
	if host.closed != 1 {
		t.Fatalf("CloseOverlay called %d times, want 1", host.closed)
	}
	if tt.overlayOpen {
		t.Fatal("tooltip still marked as owning an auxiliary window after Hide")
	}
}

// TestTooltipPlacementPositionsOverlay covers all four placement groups.
func TestTooltipPlacementPositionsOverlay(t *testing.T) {
	cases := []struct {
		name      string
		placement PopupPlacement
		wantX     func(w float32) float32
		wantY     func(h float32) float32
	}{
		{"top", PlacementTop,
			func(w float32) float32 { return 100 + (120-w)/2 },
			func(h float32) float32 { return 200 - h - 8 }},
		{"bottom", PlacementBottom,
			func(w float32) float32 { return 100 + (120-w)/2 },
			func(h float32) float32 { return 200 + 32 + 8 }},
		{"left", PlacementLeft,
			func(w float32) float32 { return 100 - w - 8 },
			func(h float32) float32 { return 200 + (32-h)/2 }},
		{"right", PlacementRight,
			func(w float32) float32 { return 100 + 120 + 8 },
			func(h float32) float32 { return 200 + (32-h)/2 }},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, tt, cfg, host := newOverlayTooltip(t)
			tt.SetPlacement(tc.placement)
			tt.Show()
			w, h := tt.measure(cfg)
			if host.x != tc.wantX(w) || host.y != tc.wantY(h) {
				t.Fatalf("position = (%v,%v), want (%v,%v)", host.x, host.y, tc.wantX(w), tc.wantY(h))
			}
		})
	}
}

// TestTooltipOverlaySuppressesInTreeDraw guards against drawing the bubble
// twice: once into the auxiliary window and once into the main window.
func TestTooltipOverlaySuppressesInTreeDraw(t *testing.T) {
	tree, tt, _, host := newOverlayTooltip(t)
	tree.SetLayout(tt.ElementID(), core.LayoutResult{Bounds: uimath.NewRect(0, 0, 120, 30)})

	tt.Show()
	buf := render.NewCommandBuffer()
	tt.Draw(buf)
	if len(buf.Commands()) != 0 || len(buf.Overlays()) != 0 {
		t.Fatalf("main-tree draw emitted %d/%d commands while the auxiliary window is open",
			len(buf.Commands()), len(buf.Overlays()))
	}

	tt.Hide()
	buf.Reset()
	tt.Draw(buf)
	if len(buf.Commands()) != 0 {
		t.Fatal("hidden tooltip still emitted commands")
	}

	// If a later hover cannot get a native window, the tooltip must fall back
	// to in-tree rendering rather than silently disappearing.
	host.fail = true
	tt.Show()
	buf.Reset()
	tt.Draw(buf)
	if len(buf.Commands()) == 0 {
		t.Fatal("legacy in-tree rendering did not resume after the overlay became unavailable")
	}
}

// TestTooltipFallsBackWithoutOverlayHost keeps the tooltip usable in hosts
// that have no OS window support (headless / test harnesses).
func TestTooltipFallsBackWithoutOverlayHost(t *testing.T) {
	tree := newTestTree()
	anchor := tree.CreateElement(core.TypeButton)
	tree.SetLayout(anchor, core.LayoutResult{Bounds: uimath.NewRect(10, 10, 80, 28)})
	cfg := DefaultConfig()
	cfg.Window = &mockWindow{}
	cfg.TextRenderer = &zProbeDrawer{}

	tt := NewTooltip(tree, "Hint", anchor, cfg)
	tree.SetLayout(tt.ElementID(), core.LayoutResult{Bounds: uimath.NewRect(0, 0, 80, 24)})
	tt.Show()
	if tt.overlayOpen {
		t.Fatal("tooltip opened an overlay without a host")
	}
	buf := render.NewCommandBuffer()
	tt.Draw(buf)
	if len(buf.Commands()) == 0 {
		t.Fatal("legacy in-tree rendering produced no commands")
	}
	tt.Hide()
}

// TestTooltipHostFailureFallsBack verifies a failed window creation does not
// leave the tooltip invisible.
func TestTooltipHostFailureFallsBack(t *testing.T) {
	tree, tt, _, host := newOverlayTooltip(t)
	tree.SetLayout(tt.ElementID(), core.LayoutResult{Bounds: uimath.NewRect(0, 0, 120, 30)})
	host.fail = true

	tt.Show()
	if tt.overlayOpen {
		t.Fatal("tooltip should fall back when the host cannot create a window")
	}
	buf := render.NewCommandBuffer()
	tt.Draw(buf)
	if len(buf.Commands()) == 0 {
		t.Fatal("fallback rendering produced no commands")
	}
}
