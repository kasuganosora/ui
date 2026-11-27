package widget

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/render"
)

// reducedMotionCases covers every decorative looping effect in the toolkit.
// Functional timers and progress indicators are intentionally excluded:
// apple-design §14 freezes motion, not state changes.
func reducedMotionCases() []struct {
	name  string
	build func(*core.Tree, *Config) Widget
} {
	return []struct {
		name  string
		build func(*core.Tree, *Config) Widget
	}{
		{"loading dots", func(tree *core.Tree, cfg *Config) Widget {
			return NewLoading(tree, cfg)
		}},
		{"skeleton gradient", func(tree *core.Tree, cfg *Config) Widget {
			w := NewSkeleton(tree, cfg)
			w.animation = AnimationGradient
			return w
		}},
		{"skeleton flashed", func(tree *core.Tree, cfg *Config) Widget {
			w := NewSkeleton(tree, cfg)
			w.animation = AnimationFlashed
			return w
		}},
		{"avatar spinner", func(tree *core.Tree, cfg *Config) Widget {
			w := NewAvatar(tree, cfg)
			w.loadState = AvatarLoading
			return w
		}},
		{"input caret", func(tree *core.Tree, cfg *Config) Widget {
			w := NewInput(tree, cfg)
			tree.SetFocused(w.ElementID(), true)
			return w
		}},
		{"textarea caret", func(tree *core.Tree, cfg *Config) Widget {
			w := NewTextArea(tree, cfg)
			tree.SetFocused(w.ElementID(), true)
			return w
		}},
	}
}

// motionSnapshot serialises the visible geometry/colors of a frame. Command
// payloads are pooled pointers, so a plain struct comparison would only
// compare addresses.
func motionSnapshot(buf *render.CommandBuffer) string {
	var b strings.Builder
	for _, lc := range paintCommands(buf) {
		c := lc.cmd
		fmt.Fprintf(&b, "%d/%d/%.4f/%v|", c.Type, c.ZOrder, c.Opacity, lc.overlay)
		if c.Rect != nil {
			r := c.Rect.Bounds
			fmt.Fprintf(&b, "R=%g,%g,%g,%g/%v/%g|", r.X, r.Y, r.Width, r.Height, c.Rect.FillColor, c.Rect.BorderWidth)
		}
		if c.Shadow != nil {
			fmt.Fprintf(&b, "S=%g|", c.Shadow.BlurRadius)
		}
	}
	return b.String()
}

// TestReducedMotionStopsContinuousRedraws is the core §14 contract: a frozen
// effect must not keep the render loop alive. The normal-motion arm is a
// control — if it stops requesting frames too, this test is vacuous.
func TestReducedMotionStopsContinuousRedraws(t *testing.T) {
	for _, tc := range reducedMotionCases() {
		t.Run(tc.name, func(t *testing.T) {
			for _, reduced := range []bool{true, false} {
				tree := newTestTree()
				cfg := DefaultConfig()
				cfg.ReducedMotion = reduced
				cfg.TextRenderer = &colorProbeDrawer{}

				w := tc.build(tree, cfg)
				setBounds(tree, w, 0, 0, 200, 80)
				buf := render.NewCommandBuffer()

				tree.ClearAllDirty()
				w.Draw(buf)
				// Widget elements are unparented in this fixture, so dirty
				// flags do not propagate to the root; inspect the widget's own
				// paint flag, which is exactly what MarkDirty sets.
				elem := tree.Get(w.ElementID())
				got := elem != nil && elem.IsDirty(core.DirtyPaint)
				want := !reduced
				if got != want {
					t.Errorf("reduced=%v: widget still requests a repaint frame=%v, want %v", reduced, got, want)
				}
			}
		})
	}
}

// TestReducedMotionFramesAreStatic verifies the frozen state is actually
// stable across frames, not merely slower. A reduced-motion indicator must be
// a legible static shape (§14: gentler non-vestibular equivalent).
func TestReducedMotionFramesAreStatic(t *testing.T) {
	for _, tc := range reducedMotionCases() {
		t.Run(tc.name, func(t *testing.T) {
			tree := newTestTree()
			cfg := DefaultConfig()
			cfg.ReducedMotion = true
			cfg.TextRenderer = &colorProbeDrawer{}
			w := tc.build(tree, cfg)
			setBounds(tree, w, 0, 0, 200, 80)

			first := render.NewCommandBuffer()
			w.Draw(first)
			time.Sleep(2 * time.Millisecond)
			second := render.NewCommandBuffer()
			w.Draw(second)

			if a, b := motionSnapshot(first), motionSnapshot(second); a != b {
				t.Errorf("reduced-motion frame changed between draws:\nfirst  = %s\nsecond = %s", a, b)
			}
			if len(paintCommands(first)) == 0 {
				t.Error("reduced motion removed the indicator entirely instead of freezing it")
			}
		})
	}
}

// TestReducedMotionKeepsCaretVisible checks the input/textarea equivalent:
// a steadily visible caret, never a blink that hides the insertion point.
func TestReducedMotionKeepsCaretVisible(t *testing.T) {
	for _, tc := range reducedMotionCases() {
		if !strings.Contains(tc.name, "caret") {
			continue
		}
		t.Run(tc.name, func(t *testing.T) {
			tree := newTestTree()
			cfg := DefaultConfig()
			cfg.ReducedMotion = true
			cfg.TextRenderer = &colorProbeDrawer{}
			w := tc.build(tree, cfg)
			setBounds(tree, w, 0, 0, 200, 80)

			// Force well past the 500ms "always visible after edit" window so
			// only a blink-gated caret could disappear.
			switch v := w.(type) {
			case *Input:
				v.blinkReset = time.Now().Add(-time.Hour).UnixMilli()
			case *TextArea:
				v.blinkReset = time.Now().Add(-time.Hour).UnixMilli()
			}
			// Sample across a full 500ms blink cycle. A single frame could
			// accidentally land on the visible phase and hide a regression,
			// so require the caret on every sample instead.
			for frame := 0; frame < 6; frame++ {
				if frame > 0 {
					time.Sleep(110 * time.Millisecond)
				}
				buf := render.NewCommandBuffer()
				w.Draw(buf)

				carets := 0
				for _, lc := range paintCommands(buf) {
					if lc.cmd.Type == render.CmdRect && lc.cmd.Rect != nil && lc.cmd.Rect.Bounds.Width == 1 {
						carets++
					}
				}
				if carets == 0 {
					t.Errorf("frame %d: caret is not drawn under reduced motion (blink still gates it)", frame)
				}
			}
		})
	}
}

// TestReducedMotionFreezesAnimatedGIF checks content animation too: an
// animated image stays on its first frame instead of looping.
func TestReducedMotionFreezesAnimatedGIF(t *testing.T) {
	for _, reduced := range []bool{true, false} {
		tree := newTestTree()
		cfg := DefaultConfig()
		cfg.ReducedMotion = reduced
		img := NewImg(tree, cfg)
		img.frames = []gifFrame{
			{pixels: []byte{0}, delay: time.Millisecond},
			{pixels: []byte{1}, delay: time.Millisecond},
		}
		img.frameIdx = 0
		img.playing = true
		img.lastUpdate = time.Now().Add(-time.Hour)
		setBounds(tree, img, 0, 0, 100, 100)

		buf := render.NewCommandBuffer()
		img.Draw(buf)
		t.Logf("reduced=%v frameIdx=%d", reduced, img.frameIdx)
		if reduced && img.frameIdx != 0 {
			t.Errorf("reduced motion still advanced GIF to frame %d", img.frameIdx)
		}
		if !reduced && img.frameIdx == 0 {
			t.Error("control arm did not advance the GIF; the freeze assertion is vacuous")
		}
	}
}

// TestReducedMotionDoesNotGateFunctionalTimers guards the boundary: an
// auto-dismiss timer is state, not decoration, and must keep working.
func TestReducedMotionDoesNotGateFunctionalTimers(t *testing.T) {
	for _, reduced := range []bool{true, false} {
		tree := newTestTree()
		cfg := DefaultConfig()
		cfg.ReducedMotion = reduced

		msg := NewMessage(tree, "操作成功", cfg)
		msg.SetVisible(true)
		msg.SetDuration(100)
		msg.Tick(0.2)
		if msg.IsVisible() {
			t.Errorf("reduced=%v: message auto-dismiss was disabled", reduced)
		}

		note := NewNotification(tree, "标题", "内容", cfg)
		note.Show()
		note.SetDuration(100)
		note.Tick(0.2)
		if note.IsVisible() {
			t.Errorf("reduced=%v: notification auto-dismiss was disabled", reduced)
		}
	}
}

func TestReducedMotionFromEnv(t *testing.T) {
	t.Setenv(ReducedMotionEnv, "true")
	if !ReducedMotionFromEnv() {
		t.Fatal("env preference was not read")
	}
	if !DefaultConfig().ReducedMotion {
		t.Fatal("DefaultConfig did not apply the env preference")
	}
	if DefaultConfig().Animate() {
		t.Fatal("Animate() should be false when reduced motion is requested")
	}

	t.Setenv(ReducedMotionEnv, "0")
	if ReducedMotionFromEnv() {
		t.Fatal("\"0\" must not enable reduced motion")
	}
}
