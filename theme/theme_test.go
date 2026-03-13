package theme

import (
	"testing"

	uimath "github.com/kasuganosora/ui/math"
)

func TestLightTheme(t *testing.T) {
	th := Light()
	if th.Name != "light" {
		t.Errorf("expected name 'light', got %q", th.Name)
	}
	if th.Primary == (uimath.Color{}) {
		t.Error("primary color should not be zero")
	}
	if th.FontSizeMD != 14 {
		t.Errorf("expected font size 14, got %g", th.FontSizeMD)
	}
	if th.SpaceMD != 16 {
		t.Errorf("expected space MD 16, got %g", th.SpaceMD)
	}
}

func TestDarkTheme(t *testing.T) {
	th := Dark()
	if th.Name != "dark" {
		t.Errorf("expected name 'dark', got %q", th.Name)
	}
	// Dark bg should be dark
	if th.BgPrimary.R > 0.2 {
		t.Errorf("dark bg primary should be dark, got R=%g", th.BgPrimary.R)
	}
}

func TestToConfig(t *testing.T) {
	th := Light()
	cfg := th.ToConfig()
	if cfg.PrimaryColor != th.Primary {
		t.Error("ToConfig primary color mismatch")
	}
	if cfg.FontSize != th.FontSizeMD {
		t.Error("ToConfig font size mismatch")
	}
	if cfg.BorderRadius != th.RadiusMD {
		t.Error("ToConfig border radius mismatch")
	}
	if cfg.ButtonHeight != th.HeightMD {
		t.Error("ToConfig button height mismatch")
	}
}

func TestAllThemes(t *testing.T) {
	themes := []*Theme{Light(), Dark(), Blue(), Green(), Purple(), Warm(), GameDark(), Compact()}
	names := map[string]bool{}
	for _, th := range themes {
		if th.Name == "" {
			t.Error("theme has empty name")
		}
		if names[th.Name] {
			t.Errorf("duplicate theme name: %s", th.Name)
		}
		names[th.Name] = true
		if th.Primary == (uimath.Color{}) {
			t.Errorf("theme %s: primary color is zero", th.Name)
		}
		if th.FontSizeMD <= 0 {
			t.Errorf("theme %s: invalid font size", th.Name)
		}
		// Verify ToConfig doesn't panic
		cfg := th.ToConfig()
		if cfg.FontSize <= 0 {
			t.Errorf("theme %s: ToConfig font size invalid", th.Name)
		}
		// Verify ToCSSVariables doesn't panic and produces output
		vars := th.ToCSSVariables()
		if len(vars) == 0 {
			t.Errorf("theme %s: ToCSSVariables returned empty", th.Name)
		}
	}
}

func TestCompactSmaller(t *testing.T) {
	light := Light()
	compact := Compact()
	if compact.FontSizeMD >= light.FontSizeMD {
		t.Error("compact font should be smaller than light")
	}
	if compact.SpaceMD >= light.SpaceMD {
		t.Error("compact spacing should be smaller than light")
	}
	if compact.HeightMD >= light.HeightMD {
		t.Error("compact height should be smaller than light")
	}
}

func TestGameDarkIsDark(t *testing.T) {
	gd := GameDark()
	if gd.BgPrimary.R > 0.2 {
		t.Errorf("game-dark bg should be dark, R=%g", gd.BgPrimary.R)
	}
}

func TestLightDarkDiffer(t *testing.T) {
	l := Light()
	d := Dark()
	if l.BgPrimary == d.BgPrimary {
		t.Error("light and dark should have different bg colors")
	}
	if l.TextPrimary == d.TextPrimary {
		t.Error("light and dark should have different text colors")
	}
}
