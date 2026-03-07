package freetype

import "unsafe"

// Memory reading helpers for accessing FreeType struct fields via raw pointers.

// readPtr reads a pointer-sized value at base+offset.
func readPtr(base uintptr, offset uintptr) uintptr {
	return *(*uintptr)(unsafe.Pointer(base + offset))
}

// readI32 reads a signed 32-bit integer at base+offset.
func readI32(base uintptr, offset uintptr) int32 {
	return *(*int32)(unsafe.Pointer(base + offset))
}

// readU32 reads an unsigned 32-bit integer at base+offset.
func readU32(base uintptr, offset uintptr) uint32 {
	return *(*uint32)(unsafe.Pointer(base + offset))
}

// readU16 reads an unsigned 16-bit integer at base+offset.
func readU16(base uintptr, offset uintptr) uint16 {
	return *(*uint16)(unsafe.Pointer(base + offset))
}

// readU8 reads an unsigned 8-bit integer at base+offset.
func readU8(base uintptr, offset uintptr) uint8 {
	return *(*uint8)(unsafe.Pointer(base + offset))
}

// fix26_6ToFloat converts a FreeType 26.6 fixed-point value to float32.
func fix26_6ToFloat(v int32) float32 {
	return float32(v) / 64.0
}
