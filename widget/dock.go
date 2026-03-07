package widget

import (
	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/event"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// DockPosition indicates where a panel is docked.
type DockPosition uint8

const (
	DockLeft   DockPosition = iota
	DockRight
	DockTop
	DockBottom
	DockCenter // Tabbed center area
	DockFloat  // Floating (undocked)
)

// DockPanel represents a single dockable panel.
type DockPanel struct {
	ID       string
	Title    string
	Content  Widget // The widget to display inside the panel
	Position DockPosition
	// For floating panels
	FloatX float32
	FloatY float32
	FloatW float32
	FloatH float32
	// Size when docked (width for left/right, height for top/bottom)
	DockSize float32
	Visible  bool
}

// DockLayout manages a docking layout with panels that can be docked to edges or floated.
type DockLayout struct {
	Base
	panels       []*DockPanel
	centerTabs   []string // panel IDs in center tab area
	activeTab    int      // index into centerTabs
	splitterSize float32
	onDock       func(panelID string, pos DockPosition)
	onUndock     func(panelID string)

	// Drag state for splitters
	dragSplitter DockPosition
	dragging     bool
	dragStart    float32
	dragOrigSize float32
}

// NewDockLayout creates a new docking layout manager.
func NewDockLayout(tree *core.Tree, cfg *Config) *DockLayout {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	dl := &DockLayout{
		Base:         NewBase(tree, core.TypeCustom, cfg),
		splitterSize: 4,
	}
	tree.AddHandler(dl.id, event.MouseDown, dl.onMouseDown)
	tree.AddHandler(dl.id, event.MouseMove, dl.onMouseMove)
	tree.AddHandler(dl.id, event.MouseUp, dl.onMouseUp)
	return dl
}

// Panels returns all registered panels.
func (dl *DockLayout) Panels() []*DockPanel { return dl.panels }

// ActiveTab returns the active center tab index.
func (dl *DockLayout) ActiveTab() int { return dl.activeTab }

// SetActiveTab sets the active center tab.
func (dl *DockLayout) SetActiveTab(i int) { dl.activeTab = i }

// OnDock sets a callback for when a panel is docked.
func (dl *DockLayout) OnDock(fn func(string, DockPosition)) { dl.onDock = fn }

// OnUndock sets a callback for when a panel is undocked.
func (dl *DockLayout) OnUndock(fn func(string)) { dl.onUndock = fn }

// AddPanel registers a new dockable panel.
func (dl *DockLayout) AddPanel(panel *DockPanel) {
	if panel.DockSize <= 0 {
		panel.DockSize = 200
	}
	if panel.FloatW <= 0 {
		panel.FloatW = 300
	}
	if panel.FloatH <= 0 {
		panel.FloatH = 200
	}
	panel.Visible = true
	dl.panels = append(dl.panels, panel)
	if panel.Position == DockCenter {
		dl.centerTabs = append(dl.centerTabs, panel.ID)
	}
}

// RemovePanel removes a panel by ID.
func (dl *DockLayout) RemovePanel(id string) {
	for i, p := range dl.panels {
		if p.ID == id {
			dl.panels = append(dl.panels[:i], dl.panels[i+1:]...)
			break
		}
	}
	for i, tid := range dl.centerTabs {
		if tid == id {
			dl.centerTabs = append(dl.centerTabs[:i], dl.centerTabs[i+1:]...)
			if dl.activeTab >= len(dl.centerTabs) && dl.activeTab > 0 {
				dl.activeTab--
			}
			break
		}
	}
}

// FindPanel finds a panel by ID.
func (dl *DockLayout) FindPanel(id string) *DockPanel {
	for _, p := range dl.panels {
		if p.ID == id {
			return p
		}
	}
	return nil
}

// DockPanel docks a panel to a position.
func (dl *DockLayout) DockPanel(id string, pos DockPosition) {
	p := dl.FindPanel(id)
	if p == nil {
		return
	}
	oldPos := p.Position
	p.Position = pos

	// Remove from center tabs if was center
	if oldPos == DockCenter {
		for i, tid := range dl.centerTabs {
			if tid == id {
				dl.centerTabs = append(dl.centerTabs[:i], dl.centerTabs[i+1:]...)
				break
			}
		}
	}

	// Add to center tabs if now center
	if pos == DockCenter {
		dl.centerTabs = append(dl.centerTabs, id)
	}

	if dl.onDock != nil {
		dl.onDock(id, pos)
	}
}

// UndockPanel makes a panel floating.
func (dl *DockLayout) UndockPanel(id string) {
	p := dl.FindPanel(id)
	if p == nil {
		return
	}
	// Save current position for float
	p.Position = DockFloat
	for i, tid := range dl.centerTabs {
		if tid == id {
			dl.centerTabs = append(dl.centerTabs[:i], dl.centerTabs[i+1:]...)
			break
		}
	}
	if dl.onUndock != nil {
		dl.onUndock(id)
	}
}

// computeRegions calculates the layout regions for each dock position.
func (dl *DockLayout) computeRegions() (left, right, top, bottom, center uimath.Rect) {
	bounds := dl.Bounds()
	if bounds.IsEmpty() {
		return
	}

	x, y := bounds.X, bounds.Y
	w, h := bounds.Width, bounds.Height

	// Calculate sizes
	var leftW, rightW, topH, bottomH float32
	for _, p := range dl.panels {
		if !p.Visible {
			continue
		}
		switch p.Position {
		case DockLeft:
			leftW = p.DockSize
		case DockRight:
			rightW = p.DockSize
		case DockTop:
			topH = p.DockSize
		case DockBottom:
			bottomH = p.DockSize
		}
	}

	sp := dl.splitterSize
	top = uimath.NewRect(x+leftW+sp, y, w-leftW-rightW-sp*2, topH)
	bottom = uimath.NewRect(x+leftW+sp, y+h-bottomH, w-leftW-rightW-sp*2, bottomH)
	left = uimath.NewRect(x, y, leftW, h)
	right = uimath.NewRect(x+w-rightW, y, rightW, h)
	center = uimath.NewRect(
		x+leftW+sp,
		y+topH+sp,
		w-leftW-rightW-sp*2,
		h-topH-bottomH-sp*2,
	)
	return
}

func (dl *DockLayout) onMouseDown(e *event.Event) {
	bounds := dl.Bounds()
	if bounds.IsEmpty() {
		return
	}
	left, right, top, bottom, _ := dl.computeRegions()
	mx, my := e.GlobalX, e.GlobalY
	sp := dl.splitterSize

	// Check splitter hit areas
	if left.Width > 0 {
		splitter := uimath.NewRect(left.Right(), left.Y, sp, left.Height)
		if splitter.Contains(uimath.NewVec2(mx, my)) {
			dl.dragging = true
			dl.dragSplitter = DockLeft
			dl.dragStart = mx
			dl.dragOrigSize = left.Width
			return
		}
	}
	if right.Width > 0 {
		splitter := uimath.NewRect(right.X-sp, right.Y, sp, right.Height)
		if splitter.Contains(uimath.NewVec2(mx, my)) {
			dl.dragging = true
			dl.dragSplitter = DockRight
			dl.dragStart = mx
			dl.dragOrigSize = right.Width
			return
		}
	}
	if top.Height > 0 {
		splitter := uimath.NewRect(top.X, top.Bottom(), top.Width, sp)
		if splitter.Contains(uimath.NewVec2(mx, my)) {
			dl.dragging = true
			dl.dragSplitter = DockTop
			dl.dragStart = my
			dl.dragOrigSize = top.Height
			return
		}
	}
	if bottom.Height > 0 {
		splitter := uimath.NewRect(bottom.X, bottom.Y-sp, bottom.Width, sp)
		if splitter.Contains(uimath.NewVec2(mx, my)) {
			dl.dragging = true
			dl.dragSplitter = DockBottom
			dl.dragStart = my
			dl.dragOrigSize = bottom.Height
			return
		}
	}
}

func (dl *DockLayout) onMouseMove(e *event.Event) {
	if !dl.dragging {
		return
	}
	for _, p := range dl.panels {
		if p.Position != dl.dragSplitter || !p.Visible {
			continue
		}
		switch dl.dragSplitter {
		case DockLeft:
			delta := e.GlobalX - dl.dragStart
			p.DockSize = dl.dragOrigSize + delta
			if p.DockSize < 50 {
				p.DockSize = 50
			}
		case DockRight:
			delta := dl.dragStart - e.GlobalX
			p.DockSize = dl.dragOrigSize + delta
			if p.DockSize < 50 {
				p.DockSize = 50
			}
		case DockTop:
			delta := e.GlobalY - dl.dragStart
			p.DockSize = dl.dragOrigSize + delta
			if p.DockSize < 50 {
				p.DockSize = 50
			}
		case DockBottom:
			delta := dl.dragStart - e.GlobalY
			p.DockSize = dl.dragOrigSize + delta
			if p.DockSize < 50 {
				p.DockSize = 50
			}
		}
		break
	}
}

func (dl *DockLayout) onMouseUp(e *event.Event) {
	dl.dragging = false
}

// Draw renders the dock layout.
func (dl *DockLayout) Draw(buf *render.CommandBuffer) {
	bounds := dl.Bounds()
	if bounds.IsEmpty() {
		return
	}
	cfg := dl.config
	left, right, top, bottom, center := dl.computeRegions()
	sp := dl.splitterSize

	// Draw dock regions
	for _, p := range dl.panels {
		if !p.Visible {
			continue
		}
		var region uimath.Rect
		switch p.Position {
		case DockLeft:
			region = left
		case DockRight:
			region = right
		case DockTop:
			region = top
		case DockBottom:
			region = bottom
		case DockCenter:
			continue // drawn separately
		case DockFloat:
			dl.drawFloating(buf, p, cfg)
			continue
		}
		if region.IsEmpty() {
			continue
		}
		dl.drawDockedPanel(buf, p, region, cfg)
	}

	// Draw center tabbed area
	if !center.IsEmpty() && len(dl.centerTabs) > 0 {
		dl.drawCenterTabs(buf, center, cfg)
	}

	// Draw splitters
	splitterColor := uimath.RGBA(0, 0, 0, 0.06)
	if left.Width > 0 {
		buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(left.Right(), left.Y, sp, left.Height),
			FillColor: splitterColor,
		}, 3, 1)
	}
	if right.Width > 0 {
		buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(right.X-sp, right.Y, sp, right.Height),
			FillColor: splitterColor,
		}, 3, 1)
	}
	if top.Height > 0 {
		buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(top.X, top.Bottom(), top.Width, sp),
			FillColor: splitterColor,
		}, 3, 1)
	}
	if bottom.Height > 0 {
		buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(bottom.X, bottom.Y-sp, bottom.Width, sp),
			FillColor: splitterColor,
		}, 3, 1)
	}
}

func (dl *DockLayout) drawDockedPanel(buf *render.CommandBuffer, p *DockPanel, region uimath.Rect, cfg *Config) {
	headerH := float32(28)
	// Background
	buf.DrawRect(render.RectCmd{
		Bounds:      region,
		FillColor:   cfg.BgColor,
		BorderColor: cfg.BorderColor,
		BorderWidth: 1,
	}, 1, 1)
	// Header
	buf.DrawRect(render.RectCmd{
		Bounds:    uimath.NewRect(region.X, region.Y, region.Width, headerH),
		FillColor: uimath.RGBA(0, 0, 0, 0.03),
	}, 2, 1)
	if cfg.TextRenderer != nil {
		lh := cfg.TextRenderer.LineHeight(cfg.FontSizeSm)
		cfg.TextRenderer.DrawText(buf, p.Title, region.X+cfg.SpaceSM, region.Y+(headerH-lh)/2, cfg.FontSizeSm, region.Width-cfg.SpaceSM*2, cfg.TextColor, 1)
	}
}

func (dl *DockLayout) drawCenterTabs(buf *render.CommandBuffer, region uimath.Rect, cfg *Config) {
	tabH := float32(28)
	// Tab bar background
	buf.DrawRect(render.RectCmd{
		Bounds:    uimath.NewRect(region.X, region.Y, region.Width, tabH),
		FillColor: uimath.RGBA(0, 0, 0, 0.03),
	}, 2, 1)

	// Tab buttons
	tabX := region.X
	tabW := float32(100)
	for i, tid := range dl.centerTabs {
		p := dl.FindPanel(tid)
		if p == nil {
			continue
		}
		isActive := i == dl.activeTab
		bgColor := uimath.ColorTransparent
		if isActive {
			bgColor = cfg.BgColor
		}
		buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(tabX, region.Y, tabW, tabH),
			FillColor: bgColor,
		}, 3, 1)
		if isActive {
			// Active indicator
			buf.DrawRect(render.RectCmd{
				Bounds:    uimath.NewRect(tabX, region.Y+tabH-2, tabW, 2),
				FillColor: cfg.PrimaryColor,
			}, 4, 1)
		}
		if cfg.TextRenderer != nil {
			lh := cfg.TextRenderer.LineHeight(cfg.FontSizeSm)
			cfg.TextRenderer.DrawText(buf, p.Title, tabX+cfg.SpaceXS, region.Y+(tabH-lh)/2, cfg.FontSizeSm, tabW-cfg.SpaceXS*2, cfg.TextColor, 1)
		}
		tabX += tabW
	}

	// Content area
	contentRect := uimath.NewRect(region.X, region.Y+tabH, region.Width, region.Height-tabH)
	buf.DrawRect(render.RectCmd{
		Bounds:      contentRect,
		FillColor:   cfg.BgColor,
		BorderColor: cfg.BorderColor,
		BorderWidth: 1,
	}, 1, 1)
}

func (dl *DockLayout) drawFloating(buf *render.CommandBuffer, p *DockPanel, cfg *Config) {
	headerH := float32(28)
	// Shadow
	buf.DrawOverlay(render.RectCmd{
		Bounds:    uimath.NewRect(p.FloatX+3, p.FloatY+3, p.FloatW, p.FloatH),
		FillColor: uimath.RGBA(0, 0, 0, 0.15),
		Corners:   uimath.CornersAll(cfg.BorderRadius),
	}, 50, 1)
	// Window body
	buf.DrawOverlay(render.RectCmd{
		Bounds:      uimath.NewRect(p.FloatX, p.FloatY, p.FloatW, p.FloatH),
		FillColor:   cfg.BgColor,
		BorderColor: cfg.BorderColor,
		BorderWidth: 1,
		Corners:     uimath.CornersAll(cfg.BorderRadius),
	}, 51, 1)
	// Title bar
	buf.DrawOverlay(render.RectCmd{
		Bounds: uimath.NewRect(p.FloatX, p.FloatY, p.FloatW, headerH),
		FillColor: uimath.RGBA(0, 0, 0, 0.03),
		Corners: uimath.Corners{
			TopLeft:  cfg.BorderRadius,
			TopRight: cfg.BorderRadius,
		},
	}, 52, 1)
	if cfg.TextRenderer != nil {
		lh := cfg.TextRenderer.LineHeight(cfg.FontSizeSm)
		cfg.TextRenderer.DrawText(buf, p.Title, p.FloatX+cfg.SpaceSM, p.FloatY+(headerH-lh)/2, cfg.FontSizeSm, p.FloatW-cfg.SpaceSM*2, cfg.TextColor, 1)
	}
}
