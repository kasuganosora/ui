package font

// mockEngine is a test double for font.Engine.
// It simulates a monospace font where every glyph has the same advance.
type mockEngine struct {
	fonts    map[ID]*mockFont
	nextID   ID
}

type mockFont struct {
	glyphs map[rune]GlyphID
	nextG  GlyphID
}

func newMockEngine() *mockEngine {
	return &mockEngine{
		fonts:  make(map[ID]*mockFont),
		nextID: 1,
	}
}

func (e *mockEngine) LoadFont(data []byte) (ID, error) {
	id := e.nextID
	e.nextID++
	f := &mockFont{
		glyphs: make(map[rune]GlyphID),
		nextG:  1,
	}
	// Pre-populate ASCII glyphs
	for r := rune(32); r < 127; r++ {
		f.glyphs[r] = f.nextG
		f.nextG++
	}
	// Add CJK characters and punctuation
	for _, r := range "你好世界中文测试我来到北京清华大学今天气不错是一个人「」（）、。，…" {
		f.glyphs[r] = f.nextG
		f.nextG++
	}
	e.fonts[id] = f
	return id, nil
}

func (e *mockEngine) LoadFontFile(path string) (ID, error) {
	return e.LoadFont(nil)
}

func (e *mockEngine) UnloadFont(id ID) {
	delete(e.fonts, id)
}

func (e *mockEngine) FontMetrics(id ID, size float32) Metrics {
	return Metrics{
		Ascent:     size * 0.8,
		Descent:    size * 0.2,
		LineHeight: size * 1.2,
		UnitsPerEm: 1000,
	}
}

func (e *mockEngine) GlyphIndex(id ID, r rune) GlyphID {
	f := e.fonts[id]
	if f == nil {
		return 0
	}
	return f.glyphs[r]
}

func (e *mockEngine) GlyphMetrics(id ID, glyph GlyphID, size float32) GlyphMetrics {
	// Monospace: advance = size * 0.6
	advance := size * 0.6
	return GlyphMetrics{
		Width:    advance,
		Height:   size,
		BearingX: 0,
		BearingY: size * 0.8,
		Advance:  advance,
	}
}

func (e *mockEngine) RasterizeGlyph(id ID, glyph GlyphID, size float32, sdf bool) (GlyphBitmap, error) {
	w := int(size * 0.6)
	h := int(size)
	if w < 1 { w = 1 }
	if h < 1 { h = 1 }
	return GlyphBitmap{
		Width:  w,
		Height: h,
		Data:   make([]byte, w*h),
		SDF:    sdf,
	}, nil
}

func (e *mockEngine) Kerning(id ID, left, right GlyphID, size float32) float32 {
	return 0 // No kerning in mock
}

func (e *mockEngine) HasGlyph(id ID, r rune) bool {
	f := e.fonts[id]
	if f == nil {
		return false
	}
	_, ok := f.glyphs[r]
	return ok
}

func (e *mockEngine) HasColorGlyphs(ID) bool { return false }
func (e *mockEngine) SetDPIScale(float32)          {}
func (e *mockEngine) Destroy()                     {}
