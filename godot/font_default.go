package godot

import (
	"os"
	"runtime"

	"github.com/kasuganosora/ui/font"
	"github.com/kasuganosora/ui/font/ttf"
)

// systemFontCandidates lists well-known system font locations across the
// platforms the headless runtime is embedded on. All paths are public OS
// defaults; no project-specific locations are referenced.
func systemFontCandidates() []string {
	switch runtime.GOOS {
	case "darwin":
		return []string{
			"/System/Library/Fonts/PingFang.ttc",
			"/Library/Fonts/Arial Unicode.ttf",
			"/System/Library/Fonts/STHeiti Light.ttc",
			"/System/Library/Fonts/Helvetica.ttc",
		}
	case "windows":
		return []string{
			`C:\Windows\Fonts\msyh.ttc`,
			`C:\Windows\Fonts\simhei.ttf`,
			`C:\Windows\Fonts\segoeui.ttf`,
			`C:\Windows\Fonts\arial.ttf`,
		}
	case "android":
		return []string{
			"/system/fonts/NotoSansCJK-Regular.ttc",
			"/system/fonts/Roboto-Regular.ttf",
		}
	default:
		return []string{
			"/usr/share/fonts/opentype/noto/NotoSansCJK-Regular.ttc",
			"/usr/share/fonts/truetype/wqy/wqy-zenhei.ttc",
			"/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf",
			"/usr/share/fonts/truetype/liberation2/LiberationSans-Regular.ttf",
		}
	}
}

// systemFontPath returns the first readable system font, or "" when the host
// has no usable font installed (common in minimal game runtime images).
func systemFontPath() string {
	for _, p := range systemFontCandidates() {
		if st, err := os.Stat(p); err == nil && !st.IsDir() {
			return p
		}
	}
	return ""
}

// resolveFontEngine returns a real-outline font engine together with the font
// path the caller should register. The pure Go engine rasterizes actual glyph
// outlines; the placeholder engine is only used when no font is installed, so
// headless frames show readable text instead of solid blocks by default.
//
// A nil engine with a non-empty path means "use ttf.New()"; the font itself is
// loaded once by the Manager via RegisterFile, avoiding a duplicate parse of
// large system font collections.
func resolveFontEngine(explicit string) (font.Engine, string) {
	path := explicit
	if path == "" {
		path = systemFontPath()
	}
	if path == "" {
		return newMockEngine(), ""
	}
	if st, err := os.Stat(path); err != nil || st.IsDir() {
		return newMockEngine(), ""
	}
	return ttf.New(), path
}
