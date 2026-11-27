package widget

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// widgetSources returns every production widget source file, including the
// game subpackage. Go test runs with the package directory as cwd.
func widgetSources(t *testing.T) map[string]string {
	t.Helper()
	out := make(map[string]string)
	for _, dir := range []string{".", "game"} {
		entries, err := os.ReadDir(dir)
		if err != nil {
			t.Fatalf("read %s: %v", dir, err)
		}
		for _, e := range entries {
			name := e.Name()
			if e.IsDir() || !strings.HasSuffix(name, ".go") || strings.HasSuffix(name, "_test.go") {
				continue
			}
			data, err := os.ReadFile(filepath.Join(dir, name))
			if err != nil {
				t.Fatalf("read %s: %v", name, err)
			}
			out[filepath.Join(dir, name)] = string(data)
		}
	}
	if len(out) < 100 {
		t.Fatalf("only %d widget sources found; the repo-wide guard would be incomplete", len(out))
	}
	return out
}

// TestNoLegacyMaterialPatternsInWidgetSources is the repo-wide backstop for
// the per-control behaviour tests: every source file, not just the widgets
// that happened to be on the demo landing page, must stay on the material
// system. Patterns here are deliberately narrow to avoid flagging legitimate
// white foreground text, slider knobs, or genuinely disabled states.
func TestNoLegacyMaterialPatternsInWidgetSources(t *testing.T) {
	banned := []struct {
		name string
		re   *regexp.Regexp
	}{
		// A dark rectangle offset by +2/+3 from the panel is a flat fake
		// shadow; use Config.FloatingShadow (a real SDF shadow) instead.
		{"offset fake shadow", regexp.MustCompile(`NewRect\([^)]*\bx\s*\+\s*[23]\s*,\s*\by\s*\+\s*[23]`)},
		// Modal scrims must come from cfg.MaskColor so dark themes work.
		{"hardcoded dark scrim", regexp.MustCompile(`FillColor:\s*uimath\.RGBA\(\s*0\s*,\s*0\s*,\s*0\s*,\s*0\.4`)},
	}
	sources := widgetSources(t)
	for path, src := range sources {
		for _, rule := range banned {
			for _, m := range rule.re.FindAllStringIndex(src, -1) {
				line := strings.Count(src[:m[0]], "\n") + 1
				t.Errorf("%s:%d: %s — %q", path, line, rule.name, strings.TrimSpace(src[m[0]:m[1]]))
			}
		}
	}
}

// TestFloatingShadowAdoptedAcrossWidgetSources keeps the material system from
// silently regressing to a handful of hand-fixed components. New floating
// surfaces should call the shared helper rather than inventing their own.
func TestFloatingShadowAdoptedAcrossWidgetSources(t *testing.T) {
	sources := widgetSources(t)
	users := 0
	for path, src := range sources {
		if strings.Contains(src, "FloatingShadow(") {
			users++
			_ = path
		}
	}
	if users < 20 {
		t.Errorf("only %d source files use Config.FloatingShadow; floating surfaces have drifted away from the shared material", users)
	}
}
