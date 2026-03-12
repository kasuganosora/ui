package game

import (
	"github.com/kasuganosora/ui/core"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
	"github.com/kasuganosora/ui/widget"
)

// MinimapMarker represents a point on the minimap.
type MinimapMarker struct {
	X, Y  float32
	Color uimath.Color
	Size  float32
	Label string
}

// Minimap displays a small overhead map with markers.
type Minimap struct {
	widget.Base
	texture    render.TextureHandle
	size       float32
	circular   bool
	playerX    float32
	playerY    float32
	playerRot  float32
	zoom       float32
	markers    []MinimapMarker
	borderW    float32
}

func NewMinimap(tree *core.Tree, cfg *widget.Config) *Minimap {
	if cfg == nil {
		cfg = widget.DefaultConfig()
	}
	return &Minimap{
		Base:    widget.NewBase(tree, core.TypeCustom, cfg),
		size:    180,
		circular: true,
		zoom:    1,
		borderW: 2,
	}
}

func (m *Minimap) SetTexture(t render.TextureHandle) { m.texture = t }
func (m *Minimap) SetSize(s float32)                 { m.size = s }
func (m *Minimap) SetCircular(c bool)                { m.circular = c }
func (m *Minimap) SetZoom(z float32)                 { m.zoom = z }
func (m *Minimap) SetPlayerPos(x, y float32)         { m.playerX = x; m.playerY = y }
func (m *Minimap) SetPlayerRotation(r float32)       { m.playerRot = r }

func (m *Minimap) AddMarker(marker MinimapMarker) {
	m.markers = append(m.markers, marker)
}

func (m *Minimap) ClearMarkers() {
	m.markers = m.markers[:0]
}

func (m *Minimap) Markers() []MinimapMarker { return m.markers }

func (m *Minimap) Draw(buf *render.CommandBuffer) {
	bounds := m.Bounds()
	if bounds.IsEmpty() {
		bounds = uimath.NewRect(0, 0, m.size, m.size)
	}
	r := m.size / 2

	// Background
	corners := uimath.CornersAll(0)
	if m.circular {
		corners = uimath.CornersAll(r)
	}
	buf.DrawRect(render.RectCmd{
		Bounds:    uimath.NewRect(bounds.X, bounds.Y, m.size, m.size),
		FillColor: uimath.RGBA(0.1, 0.1, 0.1, 0.8),
		Corners:   corners,
	}, 20, 1)

	// Border
	buf.DrawRect(render.RectCmd{
		Bounds:      uimath.NewRect(bounds.X, bounds.Y, m.size, m.size),
		BorderColor: uimath.RGBA(0.6, 0.6, 0.6, 0.8),
		BorderWidth: m.borderW,
		Corners:     corners,
	}, 21, 1)

	// Player indicator (center dot)
	cx := bounds.X + r
	cy := bounds.Y + r
	pSize := float32(6)
	buf.DrawRect(render.RectCmd{
		Bounds:    uimath.NewRect(cx-pSize/2, cy-pSize/2, pSize, pSize),
		FillColor: uimath.ColorHex("#00ff00"),
		Corners:   uimath.CornersAll(pSize / 2),
	}, 22, 1)

	// Markers
	for _, mk := range m.markers {
		dx := (mk.X - m.playerX) * m.zoom
		dy := (mk.Y - m.playerY) * m.zoom
		mx := cx + dx
		my := cy + dy
		// Clamp to minimap bounds
		if mx < bounds.X || mx > bounds.X+m.size || my < bounds.Y || my > bounds.Y+m.size {
			continue
		}
		ms := mk.Size
		if ms <= 0 {
			ms = 4
		}
		buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(mx-ms/2, my-ms/2, ms, ms),
			FillColor: mk.Color,
			Corners:   uimath.CornersAll(ms / 2),
		}, 23, 1)
	}
}
