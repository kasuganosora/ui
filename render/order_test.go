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
	// z-order moves draws across clip commands. Each draw keeps the scissor
	// that was active when it was emitted, including a draw that had no clip.
	full := uimath.NewRect(0, 0, 1e6, 1e6)
	scope := map[float32]uimath.Rect{
		1: full,
		2: uimath.NewRect(1, 0, 10, 10),
		3: uimath.NewRect(1, 0, 10, 10),
		4: uimath.NewRect(2, 0, 10, 10),
		5: uimath.NewRect(2, 0, 10, 10),
	}
	wantZ := []float32{3, 5, 2, 1, 4}
	if draws := drawOpacities(got); len(draws) != len(wantZ) {
		t.Fatalf("draws %v", draws)
	} else {
		for i, id := range wantZ {
			if draws[i] != id {
				t.Fatalf("draw[%d] %v want %v", i, draws[i], id)
			}
			if clip := clipBefore(t, got, id); clip != scope[id] {
				t.Fatalf("draw %v clip %v want %v", id, clip, scope[id])
			}
		}
	}
	if in[2].Opacity != 2 {
		t.Fatal("sort mutated the caller slice")
	}
}

func TestOrderedForPaintOrdersLayersAcrossClips(t *testing.T) {
	full := uimath.NewRect(0, 0, 1e6, 1e6)
	lowClip := uimath.NewRect(80, 60, 120, 100)
	inner := uimath.NewRect(90, 70, 40, 30)
	rect := func(z int32, id float32) Command {
		return Command{Type: CmdRect, ZOrder: z, Opacity: id, Rect: &RectCmd{}}
	}
	clip := func(b uimath.Rect) Command {
		return Command{Type: CmdClip, Clip: &ClipCmd{Bounds: b}}
	}
	in := []Command{
		rect(20010, 1), // higher layer, emitted first
		rect(20011, 8), // same higher layer, same relative order
		clip(lowClip),
		rect(10003, 2), // lower clipped content
		clip(full),
		rect(10011, 3), // lower sibling outside that clip
		rect(10011, 9), // equal z stays in emission order
		clip(lowClip),
		rect(10005, 4),
		clip(inner),
		rect(10001, 5), // nested clip, lowest z
		clip(lowClip),
		rect(10004, 6),
		clip(full),
		rect(20040, 7), // higher-layer popup, emitted last
	}
	got := OrderedForPaint(in)
	want := []float32{5, 2, 6, 4, 3, 9, 1, 8, 7}
	if draws := drawOpacities(got); len(draws) != len(want) {
		t.Fatalf("draws %v", draws)
	} else {
		for i, id := range want {
			if draws[i] != id {
				t.Fatalf("draw[%d] %v want %v in %v", i, draws[i], id, draws)
			}
		}
	}
	if clipBefore(t, got, 2) != lowClip || clipBefore(t, got, 5) != inner || clipBefore(t, got, 6) != lowClip {
		t.Fatal("lower draws lost the clip they were emitted under")
	}
	if clipBefore(t, got, 1) != full || clipBefore(t, got, 7) != full || clipBefore(t, got, 3) != full {
		t.Fatal("higher layer or an unclipped sibling inherited the lower clip")
	}
	if in[0].Opacity != 1 {
		t.Fatal("sort mutated the caller slice")
	}
}

func TestCommandBufferNestedClipFollowsDraw(t *testing.T) {
	buf := NewCommandBuffer()
	outer := uimath.NewRect(10, 10, 100, 80)
	inner := uimath.NewRect(50, 20, 100, 100)
	buf.PushClip(outer)
	buf.DrawRect(RectCmd{}, 5, 1)
	buf.PushClipZ(inner, 2)
	buf.DrawRect(RectCmd{}, 8, 2)
	buf.PopClipZ(9)
	buf.DrawRect(RectCmd{}, 3, 3)
	buf.PopClip()
	buf.DrawRect(RectCmd{}, 4, 4)
	got := OrderedForPaint(buf.Commands())
	want := []float32{3, 4, 1, 2}
	if draws := drawOpacities(got); len(draws) != len(want) {
		t.Fatalf("draws %v", draws)
	} else {
		for i, id := range want {
			if draws[i] != id {
				t.Fatalf("draw[%d] %v want %v", i, draws[i], id)
			}
		}
	}
	cross := outer.Intersection(inner)
	if clipBefore(t, got, 2) != cross {
		t.Fatalf("inner clip %v want %v", clipBefore(t, got, 2), cross)
	}
	if clipBefore(t, got, 1) != outer || clipBefore(t, got, 3) != outer {
		t.Fatal("outer clip did not survive the pop of the inner clip")
	}
	if clipBefore(t, got, 4) != uimath.NewRect(0, 0, 1e6, 1e6) {
		t.Fatal("command after the outer pop kept a clip")
	}
}

func drawOpacities(cmds []Command) []float32 {
	var out []float32
	for _, cmd := range cmds {
		if cmd.Type != CmdClip {
			out = append(out, cmd.Opacity)
		}
	}
	return out
}

func clipBefore(t *testing.T, cmds []Command, opacity float32) uimath.Rect {
	t.Helper()
	var cur uimath.Rect
	seen := false
	for _, cmd := range cmds {
		if cmd.Type == CmdClip && cmd.Clip != nil {
			cur = cmd.Clip.Bounds
			seen = true
			continue
		}
		if cmd.Opacity == opacity {
			if !seen {
				t.Fatalf("draw %v has no clip command before it", opacity)
			}
			return cur
		}
	}
	t.Fatalf("missing draw %v", opacity)
	return uimath.Rect{}
}
