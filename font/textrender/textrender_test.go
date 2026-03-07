package textrender

import (
	"testing"

	"github.com/kasuganosora/ui/font"
	"github.com/kasuganosora/ui/font/atlas"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// mockEngine for textrender tests (monospace, all glyphs identical)
type mockEngine struct {
	fonts  map[font.ID]*mockFont
	nextID font.ID
}

type mockFont struct {
	glyphs map[rune]font.GlyphID
	nextG  font.GlyphID
}

func newMockEngine() *mockEngine {
	return &mockEngine{fonts: make(map[font.ID]*mockFont), nextID: 1}
}

func (e *mockEngine) LoadFont(data []byte) (font.ID, error) {
	id := e.nextID
	e.nextID++
	f := &mockFont{glyphs: make(map[rune]font.GlyphID), nextG: 1}
	for r := rune(32); r < 127; r++ {
		f.glyphs[r] = f.nextG
		f.nextG++
	}
	for _, r := range "你好世界" {
		f.glyphs[r] = f.nextG
		f.nextG++
	}
	e.fonts[id] = f
	return id, nil
}

func (e *mockEngine) LoadFontFile(path string) (font.ID, error) { return e.LoadFont(nil) }
func (e *mockEngine) UnloadFont(id font.ID)                     { delete(e.fonts, id) }

func (e *mockEngine) FontMetrics(id font.ID, size float32) font.Metrics {
	return font.Metrics{Ascent: size * 0.8, Descent: size * 0.2, LineHeight: size * 1.2, UnitsPerEm: 1000}
}

func (e *mockEngine) GlyphIndex(id font.ID, r rune) font.GlyphID {
	if f := e.fonts[id]; f != nil {
		return f.glyphs[r]
	}
	return 0
}

func (e *mockEngine) GlyphMetrics(id font.ID, glyph font.GlyphID, size float32) font.GlyphMetrics {
	adv := size * 0.6
	return font.GlyphMetrics{Width: adv, Height: size, BearingX: 0, BearingY: size * 0.8, Advance: adv}
}

func (e *mockEngine) RasterizeGlyph(id font.ID, glyph font.GlyphID, size float32, sdf bool) (font.GlyphBitmap, error) {
	w, h := int(size*0.6), int(size)
	if w < 1 {
		w = 1
	}
	if h < 1 {
		h = 1
	}
	return font.GlyphBitmap{Width: w, Height: h, Data: make([]byte, w*h), SDF: sdf}, nil
}

func (e *mockEngine) Kerning(id font.ID, l, r font.GlyphID, size float32) float32 { return 0 }
func (e *mockEngine) HasGlyph(id font.ID, r rune) bool {
	if f := e.fonts[id]; f != nil {
		_, ok := f.glyphs[r]
		return ok
	}
	return false
}
func (e *mockEngine) Destroy() {}

func setup(t *testing.T) (*Renderer, font.ID) {
	t.Helper()
	engine := newMockEngine()
	mgr := font.NewManager(engine)
	id, err := mgr.Register("Test", font.WeightRegular, font.StyleNormal, nil)
	if err != nil {
		t.Fatal(err)
	}
	a := atlas.New(atlas.Options{Width: 256, Height: 256})
	r := New(Options{Manager: mgr, Atlas: a})
	return r, id
}

func TestDrawTextEmitsCommands(t *testing.T) {
	r, id := setup(t)
	defer r.Destroy()

	buf := render.NewCommandBuffer()
	r.DrawText(buf, "Hello", DrawOptions{
		ShapeOpts: font.ShapeOptions{FontID: id, FontSize: 16},
		Color:     uimath.Color{R: 1, G: 1, B: 1, A: 1},
		Opacity:   1,
	})

	cmds := buf.Commands()
	if len(cmds) == 0 {
		t.Fatal("expected at least 1 command")
	}
	if cmds[0].Type != render.CmdText {
		t.Errorf("expected CmdText, got %d", cmds[0].Type)
	}
	if cmds[0].Text == nil {
		t.Fatal("TextCmd should not be nil")
	}
	if len(cmds[0].Text.Glyphs) != 5 {
		t.Errorf("expected 5 glyph instances, got %d", len(cmds[0].Text.Glyphs))
	}
}

func TestDrawTextEmpty(t *testing.T) {
	r, id := setup(t)
	defer r.Destroy()

	buf := render.NewCommandBuffer()
	r.DrawText(buf, "", DrawOptions{
		ShapeOpts: font.ShapeOptions{FontID: id, FontSize: 16},
	})

	if buf.Len() != 0 {
		t.Errorf("expected 0 commands for empty text, got %d", buf.Len())
	}
}

func TestDrawTextMultiLine(t *testing.T) {
	r, id := setup(t)
	defer r.Destroy()

	buf := render.NewCommandBuffer()
	r.DrawText(buf, "A\nB", DrawOptions{
		ShapeOpts: font.ShapeOptions{FontID: id, FontSize: 16},
		Opacity:   1,
	})

	cmds := buf.Commands()
	if len(cmds) != 2 {
		t.Fatalf("expected 2 commands for 2 lines, got %d", len(cmds))
	}
}

func TestDrawTextCJK(t *testing.T) {
	r, id := setup(t)
	defer r.Destroy()

	buf := render.NewCommandBuffer()
	r.DrawText(buf, "你好", DrawOptions{
		ShapeOpts: font.ShapeOptions{FontID: id, FontSize: 24},
		Opacity:   1,
	})

	cmds := buf.Commands()
	if len(cmds) != 1 {
		t.Fatalf("expected 1 command, got %d", len(cmds))
	}
	if len(cmds[0].Text.Glyphs) != 2 {
		t.Errorf("expected 2 glyph instances, got %d", len(cmds[0].Text.Glyphs))
	}
}

func TestEnsureGlyphCaching(t *testing.T) {
	r, id := setup(t)
	defer r.Destroy()

	// Draw same text twice — atlas should cache
	buf := render.NewCommandBuffer()
	r.DrawText(buf, "A", DrawOptions{
		ShapeOpts: font.ShapeOptions{FontID: id, FontSize: 16},
		Opacity:   1,
	})

	count1 := r.Atlas().GlyphCount()

	r.DrawText(buf, "A", DrawOptions{
		ShapeOpts: font.ShapeOptions{FontID: id, FontSize: 16},
		Opacity:   1,
	})

	count2 := r.Atlas().GlyphCount()
	if count2 != count1 {
		t.Errorf("glyph count should not increase on re-draw: %d -> %d", count1, count2)
	}
}

func TestUploadNilBackend(t *testing.T) {
	r, id := setup(t)
	defer r.Destroy()

	buf := render.NewCommandBuffer()
	r.DrawText(buf, "Test", DrawOptions{
		ShapeOpts: font.ShapeOptions{FontID: id, FontSize: 16},
		Opacity:   1,
	})

	// Upload with nil backend should not panic
	err := r.Upload()
	if err != nil {
		t.Errorf("Upload should not error with nil backend: %v", err)
	}
}

func TestMeasureText(t *testing.T) {
	r, id := setup(t)
	defer r.Destroy()

	m := r.Measure("ABC", font.ShapeOptions{FontID: id, FontSize: 20})
	expectedW := float32(3 * 20 * 0.6)
	if m.Width < expectedW-1 || m.Width > expectedW+1 {
		t.Errorf("expected width ~%.1f, got %.1f", expectedW, m.Width)
	}
}

func TestDrawOptionsOrigin(t *testing.T) {
	r, id := setup(t)
	defer r.Destroy()

	buf := render.NewCommandBuffer()
	r.DrawText(buf, "A", DrawOptions{
		ShapeOpts: font.ShapeOptions{FontID: id, FontSize: 16},
		OriginX:   100,
		OriginY:   200,
		Opacity:   1,
	})

	cmds := buf.Commands()
	if len(cmds) == 0 || cmds[0].Text == nil || len(cmds[0].Text.Glyphs) == 0 {
		t.Fatal("expected glyph instances")
	}

	g := cmds[0].Text.Glyphs[0]
	if g.X < 100 {
		t.Errorf("expected glyph X >= 100 (origin offset), got %f", g.X)
	}
	if g.Y < 180 { // origin 200 - bearingY (~12.8)
		t.Errorf("expected glyph Y around 187, got %f", g.Y)
	}
}

func TestBeginFrame(t *testing.T) {
	r, _ := setup(t)
	defer r.Destroy()
	// Should not panic
	r.BeginFrame()
	r.BeginFrame()
}
