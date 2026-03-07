//go:build windows && amd64

package freetype

// FreeType struct offsets for Windows x86_64.
// Derived from FreeType 2.x headers for LP64/LLP64 (pointer = 8 bytes).
//
// These offsets are used to read fields from FreeType structs via unsafe pointer
// arithmetic, avoiding CGO entirely.

// FT_FaceRec offsets (from freetype/freetype.h)
const (
	// face->units_per_EM (FT_UShort at offset 104 in FT_FaceRec)
	offFaceUnitsPerEM = 104

	// face->glyph (FT_GlyphSlot pointer at offset 120)
	offFaceGlyph = 120

	// face->size (FT_Size pointer at offset 128)
	offFaceSize = 128
)

// FT_Size_Metrics offsets within FT_SizeRec
// FT_SizeRec: face(8) + generic(16) + metrics(...)
// metrics starts at offset 24 in FT_SizeRec
const (
	sizeMetricsBase = 24

	// FT_Size_Metrics layout:
	// x_ppem(2) + y_ppem(2) + x_scale(4+pad) + y_scale(4+pad) +
	// ascender(FT_Pos=long=4 on Win64) + descender + height + max_advance
	// On Windows x64 (LLP64): FT_Pos = signed long = 4 bytes
	// x_ppem(2) + y_ppem(2) + x_scale(4) + y_scale(4) = 12 bytes before ascender

	offSizeMetricsAscender  = sizeMetricsBase + 12 // FT_Pos (long = 4 bytes on Win64)
	offSizeMetricsDescender = sizeMetricsBase + 16 // FT_Pos
	offSizeMetricsHeight    = sizeMetricsBase + 20 // FT_Pos
)

// FT_GlyphSlotRec offsets
// The metrics sub-struct (FT_Glyph_Metrics) is embedded in the slot.
const (
	// GlyphSlotRec layout up to metrics:
	// library(8) + face(8) + next(8) + glyph_index(4) + generic(16) + metrics(...)
	// = 8+8+8+4+16 = 44, then padding to 48 for alignment
	offSlotMetrics = 48

	// FT_Glyph_Metrics layout (all FT_Pos = long = 4 bytes on Win64 LLP64):
	// width(4) + height(4) + horiBearingX(4) + horiBearingY(4) + horiAdvance(4) +
	// vertBearingX(4) + vertBearingY(4) + vertAdvance(4) = 32 bytes total

	offSlotMetricsWidth        = offSlotMetrics + 0
	offSlotMetricsHeight       = offSlotMetrics + 4
	offSlotMetricsHoriBearingX = offSlotMetrics + 8
	offSlotMetricsHoriBearingY = offSlotMetrics + 12
	offSlotMetricsHoriAdvance  = offSlotMetrics + 16

	// FT_Bitmap within GlyphSlotRec
	// After metrics(32):
	// offset 80: linearHoriAdvance (FT_Fixed=long=4)
	// offset 84: linearVertAdvance (FT_Fixed=4)
	// offset 88: advance (FT_Vector={FT_Pos,FT_Pos}={4,4}=8)
	// offset 96: format (FT_Glyph_Format=unsigned long=4)
	// offset 100: 4 bytes padding (FT_Bitmap needs 8-byte alignment due to pointer member)
	// offset 104: bitmap starts

	offSlotBitmap = 104

	// FT_Bitmap layout:
	// rows(unsigned int = 4) + width(unsigned int = 4) + pitch(int = 4) +
	// buffer(pointer = 8, aligned at +16) + num_grays(unsigned short = 2) +
	// pixel_mode(unsigned char = 1) + palette_mode(unsigned char = 1) + palette(pointer = 8)

	offSlotBitmapRows      = offSlotBitmap + 0  // unsigned int
	offSlotBitmapWidth     = offSlotBitmap + 4  // unsigned int
	offSlotBitmapPitch     = offSlotBitmap + 8  // int
	offSlotBitmapBuffer    = offSlotBitmap + 16 // pointer (8-byte aligned)
	offSlotBitmapPixelMode = offSlotBitmap + 26 // unsigned char (after buffer(8) + num_grays(2))
)
