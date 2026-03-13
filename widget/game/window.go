package game

import (
	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/layout"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
	"github.com/kasuganosora/ui/widget"
)

// Window is a sub-window with a title bar and optional close button.
// It is a pure visual container — all interaction (drag, close, z-order)
// is handled by WindowManager.
//
// Usage:
//
//	wm := game.NewWindowManager(tree, rootDiv)
//	win := game.NewWindow(tree, "Inventory", cfg)
//	win.AppendChild(invWidget)
//	wm.Add(win, 100, 100, 300, 250)  // x, y, w, h
type Window struct {
	widget.Base

	title   string
	titleH  float32 // title bar height (default 28)
	width   float32 // fallback size
	height  float32
	visible bool

	// Close button
	showClose bool
	onClose   func()

	// Appearance
	bgColor     uimath.Color
	borderColor uimath.Color
	borderWidth float32
	titleColor  uimath.Color
	titleBg     uimath.Color
	shadow      bool

	// Chromeless mode: no title bar, border, or shadow.
	// Entire window area is draggable. Used for HUD elements.
	chromeless bool
}

// NewWindow creates a window. Add it to a WindowManager for interaction.
func NewWindow(tree *core.Tree, title string, cfg *widget.Config) *Window {
	if cfg == nil {
		cfg = widget.DefaultConfig()
	}
	return &Window{
		Base:      widget.NewBase(tree, core.TypeCustom, cfg),
		title:     title,
		titleH:    28,
		visible:   true,
		showClose: true,
		shadow:    true,
	}
}

// --- Getters / Setters -------------------------------------------------------

func (w *Window) Title() string                    { return w.title }
func (w *Window) SetTitle(t string)                { w.title = t }
func (w *Window) TitleH() float32                  { return w.titleH }
func (w *Window) SetTitleH(h float32)              { w.titleH = h }
func (w *Window) SetSize(width, height float32)    { w.width = width; w.height = height }
func (w *Window) Visible() bool                    { return w.visible }
func (w *Window) SetVisible(v bool)                { w.visible = v }
func (w *Window) ShowClose() bool                  { return w.showClose }
func (w *Window) SetShowClose(v bool)              { w.showClose = v }
func (w *Window) OnClose(fn func())                { w.onClose = fn }
func (w *Window) SetShadow(v bool)                 { w.shadow = v }
func (w *Window) SetBgColor(c uimath.Color)        { w.bgColor = c }
func (w *Window) SetBorderColor(c uimath.Color)    { w.borderColor = c }
func (w *Window) SetBorderWidth(v float32)          { w.borderWidth = v }
func (w *Window) SetTitleColor(c uimath.Color)      { w.titleColor = c }
func (w *Window) SetTitleBg(c uimath.Color)         { w.titleBg = c }
func (w *Window) Chromeless() bool                   { return w.chromeless }
func (w *Window) SetChromeless(v bool)               { w.chromeless = v }

// --- Draw --------------------------------------------------------------------

func (w *Window) Draw(buf *render.CommandBuffer) {
	if !w.visible {
		return
	}
	bounds := w.Bounds()
	if bounds.IsEmpty() {
		if w.width > 0 && w.height > 0 {
			bounds = uimath.NewRect(0, 0, w.width, w.height+w.titleH)
		} else {
			return
		}
	}

	// Chromeless: just draw children, no chrome
	if w.chromeless {
		w.DrawChildren(buf)
		return
	}

	cfg := w.Config()
	radius := cfg.BorderRadius

	bgColor := w.bgColor
	if bgColor.A == 0 {
		bgColor = uimath.RGBA(0.06, 0.06, 0.1, 0.92)
	}
	borderColor := w.borderColor
	if borderColor.A == 0 {
		borderColor = uimath.RGBA(0.35, 0.35, 0.45, 0.8)
	}
	borderW := w.borderWidth
	if borderW == 0 {
		borderW = 1
	}
	titleBg := w.titleBg
	if titleBg.A == 0 {
		titleBg = uimath.RGBA(0.12, 0.12, 0.18, 1)
	}
	titleColor := w.titleColor
	if titleColor.A == 0 {
		titleColor = uimath.ColorHex("#ffd700")
	}

	x, y := bounds.X, bounds.Y
	bw, bh := bounds.Width, bounds.Height

	if w.shadow {
		buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(x+3, y+3, bw, bh),
			FillColor: uimath.RGBA(0, 0, 0, 0.25),
			Corners:   uimath.CornersAll(radius),
		}, 0, 1)
	}

	buf.DrawRect(render.RectCmd{
		Bounds:      uimath.NewRect(x, y, bw, bh),
		FillColor:   bgColor,
		BorderColor: borderColor,
		BorderWidth: borderW,
		Corners:     uimath.CornersAll(radius),
	}, 1, 1)

	buf.DrawRect(render.RectCmd{
		Bounds:    uimath.NewRect(x, y, bw, w.titleH),
		FillColor: titleBg,
		Corners:   uimath.Corners{TopLeft: radius, TopRight: radius},
	}, 2, 1)

	buf.DrawRect(render.RectCmd{
		Bounds:    uimath.NewRect(x, y+w.titleH-1, bw, 1),
		FillColor: borderColor,
	}, 3, 1)

	if w.title != "" && cfg.TextRenderer != nil {
		pad := float32(8)
		lh := cfg.TextRenderer.LineHeight(cfg.FontSize)
		maxW := bw - pad*2
		if w.showClose {
			maxW -= w.titleH
		}
		cfg.TextRenderer.DrawText(buf, w.title, x+pad, y+(w.titleH-lh)/2, cfg.FontSize, maxW, titleColor, 1)
	}

	if w.showClose {
		iconSize := w.titleH * 0.5
		iconX := x + bw - w.titleH*0.5 - iconSize/2
		iconY := y + (w.titleH-iconSize)/2
		closeColor := uimath.RGBA(0.7, 0.7, 0.7, 0.9)
		if !cfg.DrawMDIcon(buf, "close", iconX, iconY, iconSize, closeColor, 4, 1) {
			if cfg.TextRenderer != nil {
				lh := cfg.TextRenderer.LineHeight(cfg.FontSize)
				tx := x + bw - w.titleH*0.5 - cfg.FontSize*0.25
				ty := y + (w.titleH-lh)/2
				cfg.TextRenderer.DrawText(buf, "×", tx, ty, cfg.FontSize, w.titleH, closeColor, 1)
			}
		}
	}

	w.DrawChildren(buf)
}

// ═══════════════════════════════════════════════════════════════════════════
// WindowManager
// ═══════════════════════════════════════════════════════════════════════════

// windowEntry stores a managed window's absolute position and size.
type windowEntry struct {
	win        *Window
	x, y, w, h float32
}

// WindowManager handles window positioning, drag, z-order, and close.
// It does NOT use CSSLayout for window positions — positions are managed
// directly and applied every frame via PostLayout.
type WindowManager struct {
	tree    *core.Tree
	root    widget.Widget // parent container (rootDiv)
	windows []*windowEntry

	// Drag state
	dragging *windowEntry
	dragOX   float32
	dragOY   float32

	// Edge snapping
	snapEnabled  bool
	snapDistance  float32
}

// NewWindowManager creates a window manager.
// root is the parent container (e.g. rootDiv) that owns the windows.
func NewWindowManager(tree *core.Tree, root widget.Widget) *WindowManager {
	return &WindowManager{tree: tree, root: root, snapEnabled: true, snapDistance: 10}
}

// SetSnapEnabled toggles edge snapping (default: on).
func (wm *WindowManager) SetSnapEnabled(v bool) { wm.snapEnabled = v }

// SnapEnabled returns whether edge snapping is on.
func (wm *WindowManager) SnapEnabled() bool { return wm.snapEnabled }

// SetSnapDistance sets the snap threshold in pixels (default: 10).
func (wm *WindowManager) SetSnapDistance(d float32) { wm.snapDistance = d }

// Add registers a window at the given absolute position.
// The window is appended as a child of root.
// If h <= 0, the window auto-sizes based on its content (CSS-computed height).
// It sets the window's CSS style so children are sized and padded below the title bar.
func (wm *WindowManager) Add(win *Window, x, y, w, h float32) {
	wm.windows = append(wm.windows, &windowEntry{win: win, x: x, y: y, w: w, h: h})
	style := layout.Style{
		Position: layout.PositionAbsolute,
		Left:     layout.Px(x),
		Top:      layout.Px(y),
		Width:    layout.Px(w),
	}
	if !win.chromeless {
		style.Padding = layout.EdgeValues{Top: layout.Px(win.titleH)}
	}
	if h > 0 {
		style.Height = layout.Px(h)
	}
	win.SetStyle(style)
	wm.root.(interface{ AppendChild(widget.Widget) }).AppendChild(win)
}

// Remove unregisters a window.
func (wm *WindowManager) Remove(win *Window) {
	for i, e := range wm.windows {
		if e.win == win {
			wm.windows = append(wm.windows[:i], wm.windows[i+1:]...)
			return
		}
	}
}

// Windows returns the list of managed windows (front-to-back order).
func (wm *WindowManager) Windows() []*Window {
	out := make([]*Window, len(wm.windows))
	for i, e := range wm.windows {
		out[i] = e.win
	}
	return out
}

// IsDragging returns true if any window is being dragged.
func (wm *WindowManager) IsDragging() bool { return wm.dragging != nil }

// HandleMouseDown checks if (x,y) hits a window's title bar or close button.
// Checks windows in reverse z-order (last = top-most).
// Returns true if the event was consumed.
func (wm *WindowManager) HandleMouseDown(x, y float32) bool {
	for i := len(wm.windows) - 1; i >= 0; i-- {
		e := wm.windows[i]
		if !e.win.visible {
			continue
		}
		// Check if point is inside window bounds
		eh := wm.actualHeight(e)
		if x < e.x || x >= e.x+e.w || y < e.y || y >= e.y+eh {
			continue
		}
		// Bring to front
		wm.bringToFront(i)
		e = wm.windows[len(wm.windows)-1] // re-fetch after reorder

		// Chromeless: entire area is drag zone
		if e.win.chromeless {
			wm.dragging = e
			wm.dragOX = x - e.x
			wm.dragOY = y - e.y
			return true
		}

		titleH := e.win.titleH
		// Title bar region?
		if y <= e.y+titleH {
			// Close button hit test
			if e.win.showClose {
				btnSize := titleH * 0.6
				btnX := e.x + e.w - titleH*0.2 - btnSize
				btnY := e.y + (titleH-btnSize)/2
				if x >= btnX && x <= btnX+btnSize && y >= btnY && y <= btnY+btnSize {
					if e.win.onClose != nil {
						e.win.onClose()
					} else {
						e.win.visible = false
					}
					return true
				}
			}
			// Start drag
			wm.dragging = e
			wm.dragOX = x - e.x
			wm.dragOY = y - e.y
			return true
		}
		// Click is in window content area — consume so it doesn't pass through
		// to windows below, but don't start drag.
		return true
	}
	return false
}

// HandleMouseMove updates drag position with optional edge snapping.
func (wm *WindowManager) HandleMouseMove(x, y float32) {
	if wm.dragging == nil {
		return
	}
	e := wm.dragging
	newX := x - wm.dragOX
	newY := y - wm.dragOY

	if wm.snapEnabled {
		// Get viewport dimensions from root element layout
		var vpW, vpH float32
		rootElem := wm.tree.Get(wm.root.(widget.Widget).ElementID())
		if rootElem != nil {
			rb := rootElem.Layout().Bounds
			vpW, vpH = rb.Width, rb.Height
		}
		if vpW > 0 && vpH > 0 {
			winW := e.w
			winH := wm.actualHeight(e)
			d := wm.snapDistance

			// Snap left edge
			if newX >= -d && newX <= d {
				newX = 0
			}
			// Snap right edge
			if rightGap := vpW - (newX + winW); rightGap >= -d && rightGap <= d {
				newX = vpW - winW
			}
			// Snap top edge
			if newY >= -d && newY <= d {
				newY = 0
			}
			// Snap bottom edge
			if bottomGap := vpH - (newY + winH); bottomGap >= -d && bottomGap <= d {
				newY = vpH - winH
			}

			// Also snap to other windows' edges
			for _, other := range wm.windows {
				if other == e || !other.win.visible {
					continue
				}
				otherH := wm.actualHeight(other)
				// Snap our left to other's right
				if gap := other.x + other.w - newX; gap >= -d && gap <= d {
					newX = other.x + other.w
				}
				// Snap our right to other's left
				if gap := other.x - (newX + winW); gap >= -d && gap <= d {
					newX = other.x - winW
				}
				// Snap our top to other's bottom
				if gap := other.y + otherH - newY; gap >= -d && gap <= d {
					newY = other.y + otherH
				}
				// Snap our bottom to other's top
				if gap := other.y - (newY + winH); gap >= -d && gap <= d {
					newY = other.y - winH
				}
			}
		}
	}

	e.x = newX
	e.y = newY

	// Sync CSS Left/Top so CSS layout computes correct position,
	// avoiding expensive recursive moveElement in PostLayout.
	s := e.win.Style()
	s.Left = layout.Px(newX)
	s.Top = layout.Px(newY)
	e.win.SetStyle(s)
}

// HandleMouseUp ends drag.
func (wm *WindowManager) HandleMouseUp() {
	wm.dragging = nil
}

// actualHeight returns the effective height for a window entry.
// For auto-height windows (h <= 0), reads the CSS-computed height from layout.
func (wm *WindowManager) actualHeight(e *windowEntry) float32 {
	if e.h > 0 {
		return e.h
	}
	elem := wm.tree.Get(e.win.ElementID())
	if elem != nil && elem.Layout().Bounds.Height > 0 {
		return elem.Layout().Bounds.Height
	}
	return 0
}

// PostLayout applies window positions after CSSLayout.
// Call this every frame from SetOnLayout.
func (wm *WindowManager) PostLayout() {
	for _, e := range wm.windows {
		if !e.win.visible {
			continue
		}
		wm.setElementBounds(e.win.ElementID(), e.x, e.y, e.w, wm.actualHeight(e))
	}
}

// setElementBounds sets absolute position and size for an element,
// then moves all descendants by the same delta.
func (wm *WindowManager) setElementBounds(id core.ElementID, x, y, w, h float32) {
	elem := wm.tree.Get(id)
	if elem == nil {
		return
	}
	old := elem.Layout().Bounds
	dx := x - old.X
	dy := y - old.Y
	// Skip if position and size already match (CSS Left/Top computed correctly)
	if dx == 0 && dy == 0 && old.Width == w && old.Height == h {
		return
	}
	// Set window's own bounds directly (position + size)
	wm.tree.SetLayout(id, core.LayoutResult{
		Bounds: uimath.NewRect(x, y, w, h),
	})
	// Move all children by the same delta
	if dx != 0 || dy != 0 {
		for _, child := range elem.ChildIDs() {
			moveElement(wm.tree, child, dx, dy)
		}
	}
}

// bringToFront moves window at index i to the end (top of z-order).
func (wm *WindowManager) bringToFront(i int) {
	if i >= len(wm.windows)-1 {
		return
	}
	e := wm.windows[i]
	copy(wm.windows[i:], wm.windows[i+1:])
	wm.windows[len(wm.windows)-1] = e
	// Also reorder in widget tree for draw order
	wm.root.(interface{ BringChildToFront(widget.Widget) }).BringChildToFront(e.win)
}

// moveElement offsets an element and all its descendants by (dx, dy).
func moveElement(tree *core.Tree, id core.ElementID, dx, dy float32) {
	elem := tree.Get(id)
	if elem == nil {
		return
	}
	b := elem.Layout().Bounds
	tree.SetLayout(id, core.LayoutResult{
		Bounds: uimath.NewRect(b.X+dx, b.Y+dy, b.Width, b.Height),
	})
	for _, child := range elem.ChildIDs() {
		moveElement(tree, child, dx, dy)
	}
}
