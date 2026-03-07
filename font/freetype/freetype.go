// Package freetype provides a font.Engine implementation using FreeType,
// loaded dynamically via syscall (zero CGO).
//
// On Windows, it loads freetype.dll or libfreetype-6.dll.
// On Linux, it loads libfreetype.so.6.
package freetype

import (
	"fmt"
	"sync"
	"unsafe"

	"github.com/kasuganosora/ui/font"
)

// FreeType constants
const (
	ftLoadDefault       = 0
	ftLoadRender        = 4
	ftLoadNoBitmap      = 8
	ftLoadForceAutohint = 0x20

	ftRenderModeNormal = 0
	ftRenderModeLight  = 1
	ftRenderModeSDF    = 6

	ftKerningDefault = 0

	ftPixelModeGray = 2
	ftPixelModeSDF  = 8
)

// Engine implements font.Engine using FreeType.
type Engine struct {
	mu  sync.Mutex
	lib ftLibrary
	ldr *loader

	faces          map[font.ID]*faceEntry
	nextID         font.ID
	pixelsPerPoint float32 // converts point sizes to pixels; default 96/72
	dpiScale       float32 // DPI scale factor for converting metrics to logical units
}

type faceEntry struct {
	face ftFace
	data []byte // Pinned: FT_New_Memory_Face does not copy data
}

// New creates a new FreeType engine.
// Returns an error if FreeType cannot be loaded.
func New() (*Engine, error) {
	ldr, err := newLoader()
	if err != nil {
		return nil, fmt.Errorf("freetype: %w", err)
	}

	var lib ftLibrary
	ret := ftCall(ldr.ftInitFreeType, uintptr(unsafe.Pointer(&lib)))
	if ret != 0 {
		ldr.close()
		return nil, fmt.Errorf("freetype: FT_Init_FreeType failed: %d", ret)
	}

	return &Engine{
		lib:            lib,
		ldr:            ldr,
		faces:          make(map[font.ID]*faceEntry),
		nextID:         1,
		pixelsPerPoint: 96.0 / 72.0,
		dpiScale:       1.0,
	}, nil
}

func (e *Engine) LoadFont(data []byte) (font.ID, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	// Keep a copy so the data stays pinned
	pinned := make([]byte, len(data))
	copy(pinned, data)

	var face ftFace
	ret := ftCall(e.ldr.ftNewMemoryFace,
		uintptr(e.lib),
		uintptr(unsafe.Pointer(&pinned[0])),
		uintptr(len(pinned)),
		0, // face_index
		uintptr(unsafe.Pointer(&face)),
	)
	if ret != 0 {
		return font.InvalidFontID, fmt.Errorf("freetype: FT_New_Memory_Face failed: %d", ret)
	}

	id := e.nextID
	e.nextID++
	e.faces[id] = &faceEntry{face: face, data: pinned}
	return id, nil
}

func (e *Engine) LoadFontFile(path string) (font.ID, error) {
	data, err := readFile(path)
	if err != nil {
		return font.InvalidFontID, fmt.Errorf("freetype: read font file: %w", err)
	}
	return e.LoadFont(data)
}

func (e *Engine) UnloadFont(id font.ID) {
	e.mu.Lock()
	defer e.mu.Unlock()

	entry, ok := e.faces[id]
	if !ok {
		return
	}
	ftCall(e.ldr.ftDoneFace, uintptr(entry.face))
	delete(e.faces, id)
}

func (e *Engine) FontMetrics(id font.ID, size float32) font.Metrics {
	e.mu.Lock()
	defer e.mu.Unlock()

	entry := e.faces[id]
	if entry == nil {
		return font.Metrics{}
	}

	e.setPixelSize(entry.face, size)

	// Read from face->size->metrics (26.6 fixed point)
	sizePtr := readPtr(uintptr(entry.face), offFaceSize)
	if sizePtr == 0 {
		return font.Metrics{}
	}

	ascender := fix26_6ToFloat(readI32(sizePtr, offSizeMetricsAscender))
	descender := fix26_6ToFloat(readI32(sizePtr, offSizeMetricsDescender))
	height := fix26_6ToFloat(readI32(sizePtr, offSizeMetricsHeight))
	unitsPerEM := float32(readU16(uintptr(entry.face), offFaceUnitsPerEM))

	// FreeType returns physical pixels; convert to logical by dividing by dpiScale
	s := e.dpiScale
	return font.Metrics{
		Ascent:     ascender / s,
		Descent:    -descender / s, // FreeType descender is negative
		LineHeight: height / s,
		UnitsPerEm: unitsPerEM,
	}
}

func (e *Engine) GlyphIndex(id font.ID, r rune) font.GlyphID {
	e.mu.Lock()
	defer e.mu.Unlock()

	entry := e.faces[id]
	if entry == nil {
		return 0
	}

	ret, _, _ := syscallN(e.ldr.ftGetCharIndex,
		uintptr(entry.face),
		uintptr(r),
	)
	return font.GlyphID(ret)
}

func (e *Engine) GlyphMetrics(id font.ID, glyph font.GlyphID, size float32) font.GlyphMetrics {
	e.mu.Lock()
	defer e.mu.Unlock()

	entry := e.faces[id]
	if entry == nil {
		return font.GlyphMetrics{}
	}

	e.setPixelSize(entry.face, size)

	ret := ftCall(e.ldr.ftLoadGlyph,
		uintptr(entry.face),
		uintptr(glyph),
		uintptr(ftLoadDefault),
	)
	if ret != 0 {
		return font.GlyphMetrics{}
	}

	// Read glyph slot metrics (26.6 fixed point)
	slot := readPtr(uintptr(entry.face), offFaceGlyph)
	if slot == 0 {
		return font.GlyphMetrics{}
	}

	s := e.dpiScale
	return font.GlyphMetrics{
		Width:    fix26_6ToFloat(readI32(slot, offSlotMetricsWidth)) / s,
		Height:   fix26_6ToFloat(readI32(slot, offSlotMetricsHeight)) / s,
		BearingX: fix26_6ToFloat(readI32(slot, offSlotMetricsHoriBearingX)) / s,
		BearingY: fix26_6ToFloat(readI32(slot, offSlotMetricsHoriBearingY)) / s,
		Advance:  fix26_6ToFloat(readI32(slot, offSlotMetricsHoriAdvance)) / s,
	}
}

func (e *Engine) RasterizeGlyph(id font.ID, glyph font.GlyphID, size float32, sdf bool) (font.GlyphBitmap, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	entry := e.faces[id]
	if entry == nil {
		return font.GlyphBitmap{}, fmt.Errorf("freetype: font %d not found", id)
	}

	e.setPixelSize(entry.face, size)

	loadFlags := uintptr(ftLoadDefault)
	ret := ftCall(e.ldr.ftLoadGlyph,
		uintptr(entry.face),
		uintptr(glyph),
		loadFlags,
	)
	if ret != 0 {
		return font.GlyphBitmap{}, fmt.Errorf("freetype: FT_Load_Glyph failed: %d", ret)
	}

	slot := readPtr(uintptr(entry.face), offFaceGlyph)

	renderMode := uintptr(ftRenderModeNormal)
	if sdf {
		renderMode = uintptr(ftRenderModeSDF)
	}

	ret = ftCall(e.ldr.ftRenderGlyph, slot, renderMode)
	if ret != 0 {
		return font.GlyphBitmap{}, fmt.Errorf("freetype: FT_Render_Glyph failed: %d", ret)
	}

	// Read bitmap from glyph slot
	rows := int(readU32(slot, offSlotBitmapRows))
	width := int(readU32(slot, offSlotBitmapWidth))
	pitch := int(readI32(slot, offSlotBitmapPitch))
	bufPtr := readPtr(slot, offSlotBitmapBuffer)
	pixelMode := readU8(slot, offSlotBitmapPixelMode)

	if rows == 0 || width == 0 || bufPtr == 0 {
		return font.GlyphBitmap{Width: 0, Height: 0, SDF: sdf}, nil
	}

	// Copy bitmap data
	data := make([]byte, width*rows)
	absPitch := pitch
	if absPitch < 0 {
		absPitch = -absPitch
	}

	for y := 0; y < rows; y++ {
		src := bufPtr + uintptr(y*pitch)
		dst := y * width
		for x := 0; x < width; x++ {
			data[dst+x] = *(*byte)(unsafe.Pointer(src + uintptr(x)))
		}
	}

	isSDF := pixelMode == ftPixelModeSDF || sdf
	return font.GlyphBitmap{
		Width:  width,
		Height: rows,
		Data:   data,
		SDF:    isSDF,
	}, nil
}

func (e *Engine) Kerning(id font.ID, left, right font.GlyphID, size float32) float32 {
	e.mu.Lock()
	defer e.mu.Unlock()

	entry := e.faces[id]
	if entry == nil {
		return 0
	}

	e.setPixelSize(entry.face, size)

	var delta [2]int32 // FT_Vector {x, y}
	ret := ftCall(e.ldr.ftGetKerning,
		uintptr(entry.face),
		uintptr(left),
		uintptr(right),
		uintptr(ftKerningDefault),
		uintptr(unsafe.Pointer(&delta[0])),
	)
	if ret != 0 {
		return 0
	}
	return fix26_6ToFloat(delta[0]) / e.dpiScale
}

func (e *Engine) HasGlyph(id font.ID, r rune) bool {
	return e.GlyphIndex(id, r) != 0
}

func (e *Engine) Destroy() {
	e.mu.Lock()
	defer e.mu.Unlock()

	for id, entry := range e.faces {
		ftCall(e.ldr.ftDoneFace, uintptr(entry.face))
		delete(e.faces, id)
	}

	if e.lib != 0 {
		ftCall(e.ldr.ftDoneFreeType, uintptr(e.lib))
		e.lib = 0
	}

	if e.ldr != nil {
		e.ldr.close()
		e.ldr = nil
	}
}

func (e *Engine) SetDPIScale(scale float32) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if scale <= 0 {
		scale = 1.0
	}
	e.dpiScale = scale
	e.pixelsPerPoint = scale * 96.0 / 72.0
}

func (e *Engine) setPixelSize(face ftFace, sizePoints float32) {
	pixels := uint32(sizePoints*e.pixelsPerPoint + 0.5)
	if pixels < 1 {
		pixels = 1
	}
	ftCall(e.ldr.ftSetPixelSizes,
		uintptr(face),
		0,
		uintptr(pixels),
	)
}

// Ensure Engine implements font.Engine at compile time.
var _ font.Engine = (*Engine)(nil)
