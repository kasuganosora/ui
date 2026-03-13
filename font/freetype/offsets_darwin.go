//go:build darwin

package freetype

// FreeType struct offsets for darwin LP64 (64-bit).
// Verified against FreeType 2.14.2 on macOS via CGO offsetof().
// Key difference from 32-bit: FT_Pos = long = 8 bytes (not 4).

const (
	offFaceFlags      = 16 // face->face_flags (FT_Long — LP64: long = 8 bytes)
	offFaceUnitsPerEM = 136
	offFaceGlyph      = 152
	offFaceSize       = 160
)

const (
	sizeMetricsBase         = 24
	offSizeMetricsAscender  = sizeMetricsBase + 24 // FT_Pos (8 bytes each: x_ppem(2)+y_ppem(2)+pad(4)+x_scale(8)+y_scale(8) = 24)
	offSizeMetricsDescender = sizeMetricsBase + 32
	offSizeMetricsHeight    = sizeMetricsBase + 40
)

const (
	offSlotMetrics             = 48
	offSlotMetricsWidth        = offSlotMetrics + 0  // FT_Pos (8 bytes)
	offSlotMetricsHeight       = offSlotMetrics + 8
	offSlotMetricsHoriBearingX = offSlotMetrics + 16
	offSlotMetricsHoriBearingY = offSlotMetrics + 24
	offSlotMetricsHoriAdvance  = offSlotMetrics + 32
	offSlotBitmap              = 152
	offSlotBitmapRows          = offSlotBitmap + 0
	offSlotBitmapWidth         = offSlotBitmap + 4
	offSlotBitmapPitch         = offSlotBitmap + 8
	offSlotBitmapBuffer        = offSlotBitmap + 16
	offSlotBitmapPixelMode     = offSlotBitmap + 26
)
