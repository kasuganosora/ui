package ui

import (
	"errors"
	"testing"

	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/godot"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// halfFail is a shader program that errors on the right half of uv.
// A correct backend paints none of its pixels when any sample fails.
type halfFail struct{}

func (halfFail) Eval(uv uimath.Vec2, _ map[string]float32) (uimath.Color, error) {
	if uv.X > 0.5 {
		return uimath.Color{}, errors.New("half")
	}
	return uimath.RGB(0, 1, 0), nil
}

func TestShaderCommandPaintsVaryingPixels(t *testing.T) {
	sh, err := core.CompileShader("out = vec4(uv.x, uv.y, gain, 1)")
	if err != nil {
		t.Fatal(err)
	}
	be := godot.NewSoftwareBackend()
	be.Resize(80, 48)
	buf := render.NewCommandBuffer()
	buf.DrawRect(render.RectCmd{
		Bounds:    uimath.NewRect(0, 0, 80, 48),
		FillColor: uimath.RGB(1, 1, 1),
	}, 0, 1)
	buf.DrawShader(render.ShaderCmd{
		Bounds:   uimath.NewRect(10, 10, 40, 20),
		Program:  sh,
		Uniforms: map[string]float32{"gain": 0.5},
	}, 1, 1)
	be.BeginFrame()
	be.Submit(buf)
	lr, lg, lb, la := pixAt(be.Pixels(), 80, 11, 20)
	rr, rg, rb, ra := pixAt(be.Pixels(), 80, 48, 20)
	if la == 0 || ra == 0 || lr >= rr || lb < 100 || rb < 100 || absInt(int(lb)-int(rb)) > 8 {
		t.Fatalf("varying shader left %d %d %d %d right %d %d %d %d", lr, lg, lb, la, rr, rg, rb, ra)
	}
	// The test must not have pre-evaluated a solid color. The two samples differ.
	if lr+20 > rr {
		t.Fatalf("shader collapsed to a solid rect %d %d", lr, rr)
	}

	// Clip keeps the shader inside the scissor, including a nested empty result.
	buf.Reset()
	buf.DrawRect(render.RectCmd{Bounds: uimath.NewRect(0, 0, 80, 48), FillColor: uimath.RGB(1, 0, 0)}, 0, 1)
	buf.PushClip(uimath.NewRect(10, 10, 40, 20))
	buf.PushClip(uimath.NewRect(30, 10, 12, 20))
	buf.DrawShader(render.ShaderCmd{Bounds: uimath.NewRect(10, 10, 40, 20), Program: sh, Uniforms: map[string]float32{"gain": 0}}, 2, 1)
	buf.PopClip()
	buf.PopClip()
	buf.PushClip(uimath.NewRect(0, 0, 0, 0))
	buf.DrawShader(render.ShaderCmd{Bounds: uimath.NewRect(10, 10, 40, 20), Program: sh, Uniforms: map[string]float32{"gain": 0}}, 3, 1)
	buf.PopClip()
	be.BeginFrame()
	be.Submit(buf)
	or, og, ob, oa := pixAt(be.Pixels(), 80, 16, 20)
	ir, ig, ib, ia := pixAt(be.Pixels(), 80, 34, 20)
	if oa == 0 || or < 200 || og > 20 || ob > 20 {
		t.Fatalf("outside nested clip %d %d %d %d", or, og, ob, oa)
	}
	if ia == 0 || ig < 40 || ib > 20 {
		t.Fatalf("inside nested clip %d %d %d %d", ir, ig, ib, ia)
	}

	// A higher layer paints over the shader even when the shader command is emitted later.
	buf.Reset()
	buf.SetLayerBase(10000)
	buf.DrawRect(render.RectCmd{Bounds: uimath.NewRect(30, 10, 20, 20), FillColor: uimath.RGB(0, 0, 1)}, 1, 1)
	buf.SetLayerBase(0)
	buf.DrawShader(render.ShaderCmd{
		Bounds:   uimath.NewRect(10, 10, 40, 20),
		Program:  sh,
		Uniforms: map[string]float32{"gain": 0},
	}, 1, 1)
	be.BeginFrame()
	be.Submit(buf)
	cr, cg, cb, ca := pixAt(be.Pixels(), 80, 36, 20)
	if ca == 0 || cb < 200 || cr > 20 || cg > 20 {
		t.Fatalf("layer cover %d %d %d %d", cr, cg, cb, ca)
	}

	// Alpha composites over the rect underneath. It does not replace it.
	buf.Reset()
	buf.DrawRect(render.RectCmd{Bounds: uimath.NewRect(10, 10, 40, 20), FillColor: uimath.RGB(1, 1, 1)}, 0, 1)
	fade, err := core.CompileShader("out = vec4(0, 0, 1, 0.5)")
	if err != nil {
		t.Fatal(err)
	}
	buf.DrawShader(render.ShaderCmd{Bounds: uimath.NewRect(10, 10, 40, 20), Program: fade}, 1, 1)
	be.BeginFrame()
	be.Submit(buf)
	ar, ag, ab, aa := pixAt(be.Pixels(), 80, 20, 20)
	if aa == 0 || ar < 80 || ar > 180 || ab < 200 || ag < 80 {
		t.Fatalf("alpha composite %d %d %d %d", ar, ag, ab, aa)
	}

	// A program that fails on part of the rect leaves the underlay untouched.
	buf.Reset()
	buf.DrawRect(render.RectCmd{Bounds: uimath.NewRect(10, 10, 40, 20), FillColor: uimath.RGB(1, 0, 0)}, 0, 1)
	buf.DrawShader(render.ShaderCmd{Bounds: uimath.NewRect(10, 10, 40, 20), Program: halfFail{}}, 1, 1)
	be.BeginFrame()
	be.Submit(buf)
	fr, fg, fb, fa := pixAt(be.Pixels(), 80, 14, 20)
	if fa == 0 || fr < 200 || fg > 20 || fb > 20 {
		t.Fatalf("failed shader painted %d %d %d %d", fr, fg, fb, fa)
	}

	// Replacing the program on the next submit replaces the pixels.
	green, err := core.CompileShader("out = vec4(0, 1, 0, 1)")
	if err != nil {
		t.Fatal(err)
	}
	buf.Reset()
	buf.DrawShader(render.ShaderCmd{Bounds: uimath.NewRect(10, 10, 40, 20), Program: green}, 1, 1)
	be.BeginFrame()
	be.Submit(buf)
	gr, gg, gb, ga := pixAt(be.Pixels(), 80, 14, 20)
	if ga == 0 || gg < 200 || gr > 20 || gb > 20 {
		t.Fatalf("replacement %d %d %d %d", gr, gg, gb, ga)
	}
}

func absInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
