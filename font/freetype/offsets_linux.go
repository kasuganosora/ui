//go:build linux && !android

package freetype

// FreeType struct offsets for Linux LP64 (64-bit).
// These are the same layout as Darwin LP64 since both are POSIX 64-bit.
// Key: FT_Pos = long = 8 bytes on 64-bit, verified via CGO offsetof().

const (
	offFaceUnitsPerEM = 136
	offFaceGlyph      = 152
	offFaceSize       = 160
)

const (
	sizeMetricsBase         = 24
	offSizeMetricsAscender  = sizeMetricsBase + 24
	offSizeMetricsDescender = sizeMetricsBase + 32
	offSizeMetricsHeight    = sizeMetricsBase + 40
)

const (
	offSlotMetrics             = 48
	offSlotMetricsWidth        = offSlotMetrics + 0
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
