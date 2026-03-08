package render

import (
	"testing"

	uimath "github.com/kasuganosora/ui/math"
)

func TestAcquireReleaseRectCmd(t *testing.T) {
	rc := AcquireRectCmd()
	if rc == nil {
		t.Fatal("AcquireRectCmd returned nil")
	}
	rc.Bounds = uimath.NewRect(10, 20, 100, 200)
	rc.BorderWidth = 2.5
	ReleaseRectCmd(rc)

	// After release, the object should be zeroed.
	// Acquire again - may or may not be the same pointer (sync.Pool is best-effort).
	rc2 := AcquireRectCmd()
	if rc2.BorderWidth != 0 || rc2.Bounds != (uimath.Rect{}) {
		t.Error("RectCmd not zeroed after release")
	}
	ReleaseRectCmd(rc2)
}

func TestAcquireReleaseTextCmd(t *testing.T) {
	tc := AcquireTextCmd()
	tc.X = 5
	tc.Y = 10
	tc.FontSize = 16
	tc.Color = uimath.NewColor(1, 0, 0, 1)
	ReleaseTextCmd(tc)

	tc2 := AcquireTextCmd()
	if tc2.X != 0 || tc2.Y != 0 || tc2.FontSize != 0 {
		t.Error("TextCmd not zeroed after release")
	}
	ReleaseTextCmd(tc2)
}

func TestAcquireReleaseImageCmd(t *testing.T) {
	ic := AcquireImageCmd()
	ic.Texture = 42
	ic.Tint = uimath.NewColor(0, 1, 0, 1)
	ReleaseImageCmd(ic)

	ic2 := AcquireImageCmd()
	if ic2.Texture != 0 {
		t.Error("ImageCmd not zeroed after release")
	}
	ReleaseImageCmd(ic2)
}

func TestAcquireReleaseClipCmd(t *testing.T) {
	cc := AcquireClipCmd()
	cc.Bounds = uimath.NewRect(0, 0, 800, 600)
	ReleaseClipCmd(cc)

	cc2 := AcquireClipCmd()
	if cc2.Bounds != (uimath.Rect{}) {
		t.Error("ClipCmd not zeroed after release")
	}
	ReleaseClipCmd(cc2)
}

func TestGlyphSlicePool(t *testing.T) {
	s := AcquireGlyphSlice()
	if len(s) != 0 {
		t.Fatalf("expected length 0, got %d", len(s))
	}
	if cap(s) < 64 {
		t.Fatalf("expected capacity >= 64, got %d", cap(s))
	}

	// Append some glyphs and release.
	s = append(s, GlyphInstance{X: 1, Y: 2}, GlyphInstance{X: 3, Y: 4})
	ReleaseGlyphSlice(s)

	// Re-acquire: length should be 0 but capacity preserved.
	s2 := AcquireGlyphSlice()
	if len(s2) != 0 {
		t.Fatalf("expected length 0 after re-acquire, got %d", len(s2))
	}
	ReleaseGlyphSlice(s2)
}

func TestCommandBufferResetReleasesPool(t *testing.T) {
	cb := NewCommandBuffer()
	cb.DrawRect(RectCmd{BorderWidth: 5}, 0, 1)
	cb.DrawText(TextCmd{FontSize: 12}, 0, 1)
	cb.DrawImage(ImageCmd{Texture: 7}, 0, 1)
	cb.PushClip(uimath.NewRect(0, 0, 100, 100))
	cb.DrawOverlay(RectCmd{BorderWidth: 3}, 10, 1)
	cb.DrawOverlayTextCmd(TextCmd{FontSize: 8}, 10, 1)

	if cb.Len() != 6 {
		t.Fatalf("expected 6 commands, got %d", cb.Len())
	}

	// Reset should release all commands back to pools.
	cb.Reset()

	if len(cb.Commands()) != 0 {
		t.Error("commands not cleared after Reset")
	}
	if len(cb.Overlays()) != 0 {
		t.Error("overlays not cleared after Reset")
	}
}

func TestPooledCommandsHaveCorrectValues(t *testing.T) {
	cb := NewCommandBuffer()
	cb.DrawRect(RectCmd{
		Bounds:      uimath.NewRect(10, 20, 100, 200),
		BorderWidth: 2,
		FillColor:   uimath.NewColor(1, 0, 0, 1),
	}, 5, 0.8)

	cmds := cb.Commands()
	if len(cmds) != 1 {
		t.Fatalf("expected 1 command, got %d", len(cmds))
	}
	c := cmds[0]
	if c.Type != CmdRect {
		t.Fatalf("expected CmdRect, got %d", c.Type)
	}
	if c.ZOrder != 5 {
		t.Errorf("expected zOrder 5, got %d", c.ZOrder)
	}
	if c.Opacity != 0.8 {
		t.Errorf("expected opacity 0.8, got %f", c.Opacity)
	}
	if c.Rect.BorderWidth != 2 {
		t.Errorf("expected BorderWidth 2, got %f", c.Rect.BorderWidth)
	}
	cb.Reset()
}

func BenchmarkDrawRectPooled(b *testing.B) {
	cb := NewCommandBuffer()
	cmd := RectCmd{
		Bounds:      uimath.NewRect(0, 0, 100, 50),
		FillColor:   uimath.NewColor(1, 1, 1, 1),
		BorderWidth: 1,
	}
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		cb.DrawRect(cmd, 0, 1)
		if len(cb.Commands()) >= 256 {
			cb.Reset()
		}
	}
}
