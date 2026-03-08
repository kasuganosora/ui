package material

import (
	"testing"
)

func TestCount(t *testing.T) {
	n := Count()
	if n < 2000 {
		t.Errorf("expected at least 2000 icons, got %d", n)
	}
	t.Logf("Material Design Icons: %d icons loaded", n)
}

func TestPath(t *testing.T) {
	// Common icons that should exist
	names := []string{"home", "search", "settings", "delete", "add", "close", "menu", "star", "favorite", "check"}
	for _, name := range names {
		p := Path(name)
		if p == "" {
			t.Errorf("icon %q not found", name)
		}
	}
}

func TestPathNotFound(t *testing.T) {
	p := Path("nonexistent_icon_xyz")
	if p != "" {
		t.Error("expected empty string for nonexistent icon")
	}
}

func TestIcons(t *testing.T) {
	icons := Icons()
	if len(icons) < 2000 {
		t.Errorf("expected at least 2000 icons, got %d", len(icons))
	}

	// Spot check a few icons have non-empty paths
	for _, name := range []string{"home", "search", "settings"} {
		if icons[name] == "" {
			t.Errorf("icon %q has empty path", name)
		}
	}
}
