// Package material provides Google Material Design Icons (2100+ filled icons).
// Icon SVG path data is embedded in the binary and loaded on demand.
package material

import (
	_ "embed"
	"encoding/json"
	"sync"

	"github.com/kasuganosora/ui/icon"
	"github.com/kasuganosora/ui/render"
)

//go:embed icons_data.json
var iconsJSON []byte

type iconEntry struct {
	Name string `json:"n"`
	Path string `json:"p"`
}

var (
	parseOnce sync.Once
	iconMap   map[string]string
)

func parseIcons() {
	parseOnce.Do(func() {
		var entries []iconEntry
		if err := json.Unmarshal(iconsJSON, &entries); err != nil {
			iconMap = make(map[string]string)
			return
		}
		iconMap = make(map[string]string, len(entries))
		for _, e := range entries {
			iconMap[e.Name] = e.Path
		}
	})
}

// Icons returns the full name→SVG path map for all Material Design Icons.
func Icons() map[string]string {
	parseIcons()
	return iconMap
}

// Path returns the SVG path data for a named icon, or empty string if not found.
func Path(name string) string {
	parseIcons()
	return iconMap[name]
}

// Count returns the number of available icons.
func Count() int {
	parseIcons()
	return len(iconMap)
}

// Register loads all Material Design Icons into an icon registry.
func Register(reg *icon.Registry) {
	parseIcons()
	reg.RegisterAll(iconMap)
}

// NewRegistry creates an icon registry pre-loaded with all Material Design Icons.
func NewRegistry(backend render.Backend) *icon.Registry {
	reg := icon.NewRegistry(backend)
	Register(reg)
	return reg
}
