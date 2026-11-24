package render

import "testing"

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
