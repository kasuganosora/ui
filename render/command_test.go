package render

import (
	"testing"

	uimath "github.com/kasuganosora/ui/math"
)

func TestCommandBufferDrawRect(t *testing.T) {
	cb := NewCommandBuffer()
	cb.DrawRect(RectCmd{
		Bounds:    uimath.NewRect(10, 20, 100, 50),
		FillColor: uimath.ColorRed,
	}, 0, 1.0)

	if cb.Len() != 1 {
		t.Fatalf("expected 1 command, got %d", cb.Len())
	}
	cmd := cb.Commands()[0]
	if cmd.Type != CmdRect {
		t.Errorf("expected CmdRect, got %v", cmd.Type)
	}
	if cmd.Rect.FillColor != uimath.ColorRed {
		t.Error("fill color mismatch")
	}
}

func TestCommandBufferMultiple(t *testing.T) {
	cb := NewCommandBuffer()
	cb.DrawRect(RectCmd{Bounds: uimath.NewRect(0, 0, 10, 10)}, 0, 1.0)
	cb.DrawText(TextCmd{X: 5, Y: 5, FontSize: 16}, 1, 1.0)
	cb.DrawImage(ImageCmd{Texture: 42, DstRect: uimath.NewRect(0, 0, 50, 50)}, 2, 0.5)

	if cb.Len() != 3 {
		t.Fatalf("expected 3 commands, got %d", cb.Len())
	}
	if cb.Commands()[0].Type != CmdRect {
		t.Error("first should be rect")
	}
	if cb.Commands()[1].Type != CmdText {
		t.Error("second should be text")
	}
	if cb.Commands()[2].Type != CmdImage {
		t.Error("third should be image")
	}
	if cb.Commands()[2].Opacity != 0.5 {
		t.Error("opacity mismatch")
	}
}

func TestCommandBufferReset(t *testing.T) {
	cb := NewCommandBuffer()
	cb.DrawRect(RectCmd{}, 0, 1.0)
	cb.DrawRect(RectCmd{}, 0, 1.0)
	cb.Reset()
	if cb.Len() != 0 {
		t.Errorf("expected 0 after reset, got %d", cb.Len())
	}
}

func TestCommandBufferClip(t *testing.T) {
	cb := NewCommandBuffer()
	cb.PushClip(uimath.NewRect(10, 10, 200, 200))
	if cb.Len() != 1 {
		t.Fatal("expected 1 command")
	}
	cmd := cb.Commands()[0]
	if cmd.Type != CmdClip {
		t.Error("expected CmdClip")
	}
	if cmd.Clip.Bounds.Width != 200 {
		t.Error("clip width mismatch")
	}
}
