package render

import (
	"testing"

	uimath "github.com/kasuganosora/ui/math"
)

func TestStableZOrderKeepsEqualRanksInInsertionOrder(t *testing.T) {
	cmds := []Command{
		{ZOrder: 20, Opacity: 1},
		{ZOrder: 5, Opacity: 2},
		{ZOrder: 20, Opacity: 3},
		{ZOrder: 5, Opacity: 4},
	}
	got := StableZOrder(cmds)
	if cmds[0].Opacity != 1 {
		t.Fatal("sort mutated the caller slice")
	}
	want := []float32{2, 4, 1, 3}
	for i, opacity := range want {
		if got[i].Opacity != opacity {
			t.Fatalf("order[%d] opacity %v want %v", i, got[i].Opacity, opacity)
		}
	}
}

func TestOrderedForPaintKeepsClipScopes(t *testing.T) {
	clip := func(id float32) Command {
		return Command{Type: CmdClip, ZOrder: 0, Opacity: id, Clip: &ClipCmd{Bounds: uimath.NewRect(id, 0, 10, 10)}}
	}
	rect := func(z int32, id float32) Command {
		return Command{Type: CmdRect, ZOrder: z, Opacity: id, Rect: &RectCmd{}}
	}
	in := []Command{
		rect(5, 1),
		clip(1),
		rect(4, 2),
		rect(1, 3),
		clip(2),
		rect(9, 4),
		rect(2, 5),
	}
	got := OrderedForPaint(in)
	want := []float32{1, 1, 3, 2, 2, 5, 4}
	if len(got) != len(want) {
		t.Fatalf("len %d", len(got))
	}
	for i, opacity := range want {
		if got[i].Opacity != opacity {
			t.Fatalf("cmd[%d] opacity %v want %v", i, got[i].Opacity, opacity)
		}
	}
	if in[2].Opacity != 2 {
		t.Fatal("sort mutated the caller slice")
	}
}
