package ui

import (
	"testing"

	"github.com/kasuganosora/ui/core"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

func TestShaderResultPaintsInSoftware(t *testing.T) {
	sh, err := core.CompileShader("out = vec4(0, 0, 1, 1)")
	if err != nil {
		t.Fatal(err)
	}
	c, err := sh.Eval(uimath.Vec2{}, nil)
	if err != nil {
		t.Fatal(err)
	}
	fx := newGlyphFixture(t, 120, 80)
	buf := render.NewCommandBuffer()
	buf.DrawRect(render.RectCmd{Bounds: uimath.NewRect(20, 16, 40, 24), FillColor: c}, 1, 1)
	fx.present(buf)
	r, g, b, a := pixAt(fx.be.Pixels(), 120, 40, 28)
	if a == 0 || r > 20 || g > 20 || b < 200 {
		t.Fatalf("shader rect %d %d %d %d", r, g, b, a)
	}
	r, _, _, a = pixAt(fx.be.Pixels(), 120, 4, 4)
	if a != 0 {
		t.Fatal("shader rect painted outside its bounds")
	}
}
