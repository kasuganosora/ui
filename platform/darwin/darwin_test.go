//go:build darwin

package darwin

import (
	"runtime"
	"testing"
	"unsafe"
)

func skipIfNotDarwin(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("Skipping test on non-darwin platform")
	}
}

// ========== 纯逻辑测试(无需窗口) ==========

func TestCstring(t *testing.T) {
	skipIfNotDarwin(t)
	
	tests := []struct {
		name string
		input string
	}{
		{"empty", ""},
		{"simple", "hello"},
		{"unicode", "你好世界"},
		{"with spaces", "hello world"},
		{"with symbols", "test@#$%"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cstr := cstring(tt.input)
			if cstr == nil {
				t.Fatal("cstring returned nil")
			}
			// Verify null terminator
			ptr := unsafe.Pointer(cstr)
			offset := len(tt.input)
			nullByte := *(*byte)(unsafe.Pointer(uintptr(ptr) + uintptr(offset)))
			if nullByte != 0 {
				t.Errorf("cstring not null-terminated: got %d", nullByte)
			}
		})
	}
}

func TestNSStringGoStringRoundTrip(t *testing.T) {
	skipIfNotDarwin(t)
	
	// Ensure Foundation framework is loaded for test
	ensureFoundation()
	
	// Test creating an NSString and converting back to Go string
	testStr := "Hello, 世界!"
	
	// Create NSString
	ns := nsString(testStr)
	if ns == 0 {
		t.Fatal("nsString returned nil")
	}
	defer msgSend(ns, selRelease)
	
	// Convert back to Go string
	result := goString(ns)
	if result != testStr {
		t.Errorf("Round trip failed: expected %q, got %q", testStr, result)
	}
}

func TestGoStringNil(t *testing.T) {
	skipIfNotDarwin(t)

	result := goString(0)
	if result != "" {
		t.Errorf("goString(nil) expected empty string, got %q", result)
	}
}

func TestNSRect(t *testing.T) {
	skipIfNotDarwin(t)

	tests := []struct {
		x, y, w, h float64
	}{
		{0, 0, 100, 200},
		{10.5, 20.5, 300.5, 400.5},
		{-10, -20, 50, 60},
	}

	for _, tt := range tests {
		rect := nsRect(tt.x, tt.y, tt.w, tt.h)
		if rect.Origin.X != tt.x {
			t.Errorf("nsRect X: expected %f, got %f", tt.x, rect.Origin.X)
		}
		if rect.Origin.Y != tt.y {
			t.Errorf("nsRect Y: expected %f, got %f", tt.y, rect.Origin.Y)
		}
		if rect.Size.Width != tt.w {
			t.Errorf("nsRect Width: expected %f, got %f", tt.w, rect.Size.Width)
		}
		if rect.Size.Height != tt.h {
			t.Errorf("nsRect Height: expected %f, got %f", tt.h, rect.Size.Height)
		}
	}
}

func TestNSPoint(t *testing.T) {
	skipIfNotDarwin(t)

	tests := []struct {
		x, y float64
	}{
		{0, 0},
		{100.5, 200.5},
		{-50.5, -100.5},
	}

	for _, tt := range tests {
		pt := nsPoint(tt.x, tt.y)
		if pt.X != tt.x || pt.Y != tt.y {
			t.Errorf("nsPoint: expected (%f,%f), got (%f,%f)", tt.x, tt.y, pt.X, pt.Y)
		}
	}
}

func TestNSSize(t *testing.T) {
	skipIfNotDarwin(t)

	tests := []struct {
		w, h float64
	}{
		{0, 0},
		{100, 200},
		{50.5, 100.5},
	}

	for _, tt := range tests {
		sz := nsSize(tt.w, tt.h)
		if sz.Width != tt.w || sz.Height != tt.h {
			t.Errorf("nsSize: expected (%f,%f), got (%f,%f)", tt.w, tt.h, sz.Width, sz.Height)
		}
	}
}

func TestRectFromPtr(t *testing.T) {
	skipIfNotDarwin(t)

	original := nsRect(10, 20, 300, 400)
	ptr := unsafe.Pointer(&original)
	
	result := rectFromPtr(ptr)
	
	if result.Origin.X != original.Origin.X ||
		result.Origin.Y != original.Origin.Y ||
		result.Size.Width != original.Size.Width ||
		result.Size.Height != original.Size.Height {
		t.Errorf("rectFromPtr mismatch: expected %+v, got %+v", original, result)
	}
}

func TestPointFromPtr(t *testing.T) {
	skipIfNotDarwin(t)

	original := nsPoint(100.5, 200.5)
	ptr := unsafe.Pointer(&original)
	
	result := pointFromPtr(ptr)
	
	if result.X != original.X || result.Y != original.Y {
		t.Errorf("pointFromPtr mismatch: expected %+v, got %+v", original, result)
	}
}

func TestSizeFromPtr(t *testing.T) {
	skipIfNotDarwin(t)

	original := nsSize(300.5, 400.5)
	ptr := unsafe.Pointer(&original)
	
	result := sizeFromPtr(ptr)
	
	if result.Width != original.Width || result.Height != original.Height {
		t.Errorf("sizeFromPtr mismatch: expected %+v, got %+v", original, result)
	}
}

func TestObjcClassCaching(t *testing.T) {
	skipIfNotDarwin(t)

	// Test that objcClass returns same pointer for same class
	cls1 := objcClass("NSString")
	cls2 := objcClass("NSString")
	
	if cls1 == 0 {
		t.Fatal("objcClass returned nil for NSString")
	}
	
	if cls1 != cls2 {
		t.Errorf("objcClass caching failed: got different pointers %p vs %p", 
			unsafe.Pointer(cls1), unsafe.Pointer(cls2))
	}
	
	// Test different classes
	clsArray := objcClass("NSArray")
	if clsArray == 0 {
		t.Fatal("objcClass returned nil for NSArray")
	}
	
	if cls1 == clsArray {
		t.Error("objcClass returned same pointer for different classes")
	}
}

func TestObjcSelector(t *testing.T) {
	skipIfNotDarwin(t)

	sel1 := objcSelector("alloc")
	if sel1 == 0 {
		t.Fatal("objcSelector returned nil for 'alloc'")
	}

	// 测试多个常用 selector
	selectors := []string{
		"init", "release", "retain", "autorelease",
		"stringWithUTF8String:", "length",
		"sharedApplication", "run",
	}

	for _, selName := range selectors {
		sel := objcSelector(selName)
		if sel == 0 {
			t.Errorf("objcSelector(%s) returned nil", selName)
		}
	}
}

func TestMsgSendVariadicArgs(t *testing.T) {
	skipIfNotDarwin(t)

	// Test msgSend with different argument counts
	// We can't really test the actual calls without a full Cocoa environment,
	// but we can ensure the function doesn't panic with different arg counts

	// This is more of a smoke test
	_ = msgSend(0, 0)           // 0 args
	_ = msgSend(0, 0, 1)        // 1 arg
	_ = msgSend(0, 0, 1, 2)     // 2 args
	_ = msgSend(0, 0, 1, 2, 3)  // 3 args
	_ = msgSend(0, 0, 1, 2, 3, 4) // 4 args
	_ = msgSend(0, 0, 1, 2, 3, 4, 5) // 5 args
	_ = msgSend(0, 0, 1, 2, 3, 4, 5, 6, 7) // > 5 args
}

func TestMsgSendFloat64VariadicArgs(t *testing.T) {
	skipIfNotDarwin(t)

	_ = msgSendFloat64(0, 0)
	_ = msgSendFloat64(0, 0, 1)
	_ = msgSendFloat64(0, 0, 1, 2)
	_ = msgSendFloat64(0, 0, 1, 2, 3, 4)
}

func TestNSRange(t *testing.T) {
	skipIfNotDarwin(t)

	tests := []struct {
		location, length uint64
	}{
		{0, 0},
		{10, 20},
		{NSNotFound, 0},
	}

	for _, tt := range tests {
		r := NSRange{Location: tt.location, Length: tt.length}
		if r.Location != tt.location || r.Length != tt.length {
			t.Errorf("NSRange mismatch: expected (%d,%d), got (%d,%d)",
				tt.location, tt.length, r.Location, r.Length)
		}
	}
}

func TestNSNotFoundConstant(t *testing.T) {
	skipIfNotDarwin(t)

	if NSNotFound != 0x7FFFFFFFFFFFFFFF {
		t.Errorf("NSNotFound value incorrect: expected 0x7FFFFFFFFFFFFFFF, got 0x%X", NSNotFound)
	}
}

func TestNSConstantsValues(t *testing.T) {
	skipIfNotDarwin(t)

	// Test some known constant values
	tests := []struct {
		name  string
		value uint64
	}{
		{"NSApplicationActivationPolicyRegular", NSApplicationActivationPolicyRegular},
		{"NSWindowStyleMaskTitled", NSWindowStyleMaskTitled},
		{"NSBackingStoreBuffered", NSBackingStoreBuffered},
		{"NSEventTypeLeftMouseDown", NSEventTypeLeftMouseDown},
		{"NSEventModifierFlagShift", NSEventModifierFlagShift},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Just ensure they're defined (no panic)
			_ = tt.value
		})
	}
}

// Benchmark tests
func BenchmarkNSStringCreation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		ns := nsString("Hello, World!")
		msgSend(ns, selRelease)
	}
}

func BenchmarkCstring(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = cstring("Hello, World!")
	}
}

func BenchmarkObjcClassCached(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = objcClass("NSString")
	}
}

func BenchmarkObjcSelector(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = objcSelector("alloc")
	}
}
