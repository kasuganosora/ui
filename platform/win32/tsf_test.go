//go:build windows

package win32

import (
	"testing"
	"unsafe"
)

func TestTSFManagerNil(t *testing.T) {
	var tsf *TSFManager
	// All methods should be safe on nil receiver
	if tsf.IsActive() {
		t.Error("nil TSFManager should not be active")
	}
	tsf.Release()             // should not panic
	tsf.SetCandidateWindowPosition(0, 0) // should not panic
	tsf.ClearFocusedEdit()    // should not panic
}

func TestGUIDLayout(t *testing.T) {
	// Verify GUID struct is exactly 16 bytes (COM ABI requirement)
	if got := unsafe.Sizeof(GUID{}); got != 16 {
		t.Errorf("GUID size = %d, want 16", got)
	}
}

func TestGUIDValues(t *testing.T) {
	// Verify well-known CLSID_TF_ThreadMgr Data1 field
	if CLSID_TF_ThreadMgr.Data1 != 0x529a9e6b {
		t.Errorf("CLSID_TF_ThreadMgr.Data1 = 0x%x, want 0x529a9e6b", CLSID_TF_ThreadMgr.Data1)
	}
	// Verify well-known IID_ITfThreadMgr Data1 field
	if IID_ITfThreadMgr.Data1 != 0xaa80e801 {
		t.Errorf("IID_ITfThreadMgr.Data1 = 0x%x, want 0xaa80e801", IID_ITfThreadMgr.Data1)
	}
}

func TestTSFManagerZeroValue(t *testing.T) {
	// A zero-value TSFManager (not nil, but not initialized) should not be active
	tsf := &TSFManager{}
	if tsf.IsActive() {
		t.Error("zero-value TSFManager should not be active")
	}
	// Release on uninitialized should be safe
	tsf.Release()
}

func TestComReleaseZero(t *testing.T) {
	// comRelease with zero pointer should be a safe no-op
	comRelease(0) // should not panic
}
