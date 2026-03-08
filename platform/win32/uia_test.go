//go:build windows

package win32

import "testing"

func TestUIAProviderNil(t *testing.T) {
	var p *UIAProvider
	result, handled := p.HandleGetObject(0, 0)
	if handled {
		t.Error("nil provider should not handle messages")
	}
	if result != 0 {
		t.Error("nil provider should return 0")
	}
	// These should not panic
	p.RaiseStructureChanged()
	p.RaiseFocusChanged()
	p.RaisePropertyChanged(0)
	p.SetEnabled(true)
}

func TestUIAProviderDisabled(t *testing.T) {
	p := NewUIAProvider(0)
	p.SetEnabled(false)
	_, handled := p.HandleGetObject(0, 0xFFFFFFFC)
	if handled {
		t.Error("disabled provider should not handle")
	}
}

func TestUIAProviderEnabled(t *testing.T) {
	p := NewUIAProvider(0x1234)
	if !p.enabled {
		t.Error("new provider should be enabled by default")
	}

	// WM_GETOBJECT with non-OBJID_CLIENT lParam should not be handled
	_, handled := p.HandleGetObject(0, 0)
	if handled {
		t.Error("should not handle non-OBJID_CLIENT lParam")
	}

	// WM_GETOBJECT with OBJID_CLIENT lParam
	// Currently returns (0, false) as it's a stub
	result, handled := p.HandleGetObject(0, 0xFFFFFFFC)
	if handled {
		t.Error("stub should not claim to handle yet")
	}
	if result != 0 {
		t.Error("stub should return 0")
	}
}

func TestUIAProviderSetEnabled(t *testing.T) {
	p := NewUIAProvider(0)
	if !p.enabled {
		t.Error("expected enabled by default")
	}
	p.SetEnabled(false)
	if p.enabled {
		t.Error("expected disabled after SetEnabled(false)")
	}
	p.SetEnabled(true)
	if !p.enabled {
		t.Error("expected enabled after SetEnabled(true)")
	}
}
