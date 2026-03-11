//go:build android

package freetype

// FreeType struct offsets for Android (ARM64/x86_64 LP64).
// Same layout as Linux LP64.

const (
	offFaceUnitsPerEM = 104
	offFaceGlyph      = 120
	offFaceSize       = 128
)

const (
	sizeMetricsBase         = 24
	offSizeMetricsAscender  = sizeMetricsBase + 12
	offSizeMetricsDescender = sizeMetricsBase + 16
	offSizeMetricsHeight    = sizeMetricsBase + 20
)

const (
	offSlotMetrics             = 48
	offSlotMetricsWidth        = offSlotMetrics + 0
	offSlotMetricsHeight       = offSlotMetrics + 4
	offSlotMetricsHoriBearingX = offSlotMetrics + 8
	offSlotMetricsHoriBearingY = offSlotMetrics + 12
	offSlotMetricsHoriAdvance  = offSlotMetrics + 16
	offSlotBitmap              = 104
	offSlotBitmapRows          = offSlotBitmap + 0
	offSlotBitmapWidth         = offSlotBitmap + 4
	offSlotBitmapPitch         = offSlotBitmap + 8
	offSlotBitmapBuffer        = offSlotBitmap + 16
	offSlotBitmapPixelMode     = offSlotBitmap + 26
)
