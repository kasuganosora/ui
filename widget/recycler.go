package widget

import (
	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/layout"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// RecyclerView is a scrollable list that only materializes widgets for visible
// items plus a buffer zone. Off-screen widgets are automatically destroyed to
// prevent unbounded memory growth in infinite-scroll scenarios.
//
// For simple widgets (Div, custom Draw), items are positioned directly.
// For complex CSS-laid-out items (HTML+CSS tweet cards), set a LayoutItem
// callback — RecyclerView captures relative positions at creation and applies
// absolute offsets each frame, so internal flexbox layout works correctly.
//
// Usage:
//
//	rv := widget.NewRecyclerView(tree, cfg)
//	rv.SetItemCount(10000)
//	rv.SetEstimatedItemHeight(120)
//	rv.SetCreateItem(func(index int) widget.Widget { ... })
//	rv.SetDestroyItem(func(index int, w widget.Widget) { w.Destroy() })
type RecyclerView struct {
	Base

	// Data
	itemCount int

	// Sizing
	estimatedHeight float32              // default height estimate per item
	heightFn        func(int) float32    // optional per-item height estimate

	// Widget lifecycle
	createFn  func(index int) Widget          // create & bind a widget for data[index]
	destroyFn func(index int, w Widget)       // cleanup before recycle (must call w.Destroy())

	// Optional: run CSS layout on newly created items. Returns computed height.
	// When set, RecyclerView captures relative element positions at creation
	// and applies absolute offsets during Draw(), enabling full CSS flexbox
	// inside each item (e.g. tweet cards built with LoadHTMLWithCSS).
	layoutFn func(w Widget, width float32) float32

	// Scroll state
	scrollY       float32
	contentHeight float32
	scrollBarDrag bool
	dragStartY    float32
	dragStartScrl float32

	// Active widgets: data index → widget
	active map[int]Widget

	// Per-item layout cache: relative positions captured after layoutFn
	itemLayouts map[int]*itemLayoutCache

	// Measured heights: data index → actual measured height (after layout)
	heights map[int]float32

	// cumY[i] = sum of heights[0..i-1], lazily computed
	cumY      []float32
	cumYValid bool

	// Buffer: how many extra items above/below viewport to keep materialized
	bufferCount int

	// Callbacks
	onScroll    func(scrollY float32)
	onNearEnd   func() // called when scrolled near the bottom (infinite scroll trigger)
	nearEndPx   float32

	// Track if we need to rebuild cumulative heights
	heightsDirty bool
}

// itemLayoutCache stores the relative element positions for one item,
// captured after CSS layout at origin (0,0). During Draw(), these are
// offset by the item's absolute position in the scroll view.
type itemLayoutCache struct {
	entries []layoutEntry
}

type layoutEntry struct {
	id     core.ElementID
	layout core.LayoutResult
}

// NewRecyclerView creates a recycler view.
func NewRecyclerView(tree *core.Tree, cfg *Config) *RecyclerView {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	rv := &RecyclerView{
		Base:            NewBase(tree, core.TypeDiv, cfg),
		estimatedHeight: 100,
		bufferCount:     3,
		active:          make(map[int]Widget),
		itemLayouts:     make(map[int]*itemLayoutCache),
		heights:         make(map[int]float32),
		nearEndPx:       200,
	}
	rv.style.Display = layout.DisplayBlock
	rv.style.FlexGrow = 1
	rv.style.Overflow = layout.OverflowScroll
	return rv
}

// --- Configuration ---

func (rv *RecyclerView) SetItemCount(n int) {
	if rv.itemCount != n {
		rv.itemCount = n
		rv.cumYValid = false
		rv.updateContentHeight()
	}
}

func (rv *RecyclerView) ItemCount() int { return rv.itemCount }

func (rv *RecyclerView) SetEstimatedItemHeight(h float32) {
	rv.estimatedHeight = h
	rv.cumYValid = false
}

// SetHeightFn sets a per-item height estimator. Falls back to estimatedHeight.
func (rv *RecyclerView) SetHeightFn(fn func(index int) float32) {
	rv.heightFn = fn
	rv.cumYValid = false
}

// SetCreateItem sets the factory function that creates a widget for the given data index.
func (rv *RecyclerView) SetCreateItem(fn func(index int) Widget) {
	rv.createFn = fn
}

// SetDestroyItem sets the cleanup function called when a widget is recycled.
// The function MUST call w.Destroy() to clean up the element tree.
func (rv *RecyclerView) SetDestroyItem(fn func(index int, w Widget)) {
	rv.destroyFn = fn
}

// SetLayoutItem sets a function that computes CSS layout for a newly created item.
// The function receives the item widget and available width, runs CSS layout
// (e.g. ui.CSSLayout), and returns the computed height.
// When set, RecyclerView captures per-element relative positions and applies
// absolute offsets during Draw(), so complex items (HTML+CSS) render correctly.
func (rv *RecyclerView) SetLayoutItem(fn func(w Widget, width float32) float32) {
	rv.layoutFn = fn
}

func (rv *RecyclerView) SetBufferCount(n int) { rv.bufferCount = n }
func (rv *RecyclerView) SetOnScroll(fn func(float32)) { rv.onScroll = fn }
func (rv *RecyclerView) SetOnNearEnd(fn func())       { rv.onNearEnd = fn }
func (rv *RecyclerView) SetNearEndThreshold(px float32) { rv.nearEndPx = px }

func (rv *RecyclerView) ScrollY() float32       { return rv.scrollY }
func (rv *RecyclerView) ContentHeight() float32  { return rv.contentHeight }
func (rv *RecyclerView) ActiveCount() int        { return len(rv.active) }

// ScrollTo sets the scroll position.
func (rv *RecyclerView) ScrollTo(y float32) {
	rv.scrollY = y
	rv.clampScroll()
}

// ScrollBy adjusts scroll by delta.
func (rv *RecyclerView) ScrollBy(dy float32) {
	rv.scrollY += dy
	rv.clampScroll()
}

// HandleWheel processes mouse wheel events.
func (rv *RecyclerView) HandleWheel(dy float32) {
	rv.ScrollBy(-dy * 40)
}

// SetMeasuredHeight records the actual height of item[index] after layout.
// This improves scroll accuracy for variable-height items.
func (rv *RecyclerView) SetMeasuredHeight(index int, h float32) {
	if old, ok := rv.heights[index]; !ok || old != h {
		rv.heights[index] = h
		rv.cumYValid = false
	}
}

// --- Internal ---

// captureRelativeLayout walks the item's widget subtree and records all
// element positions (relative to origin 0,0 as computed by layoutFn).
func (rv *RecyclerView) captureRelativeLayout(w Widget) *itemLayoutCache {
	cache := &itemLayoutCache{}
	rv.tree.Walk(w.ElementID(), func(id core.ElementID, depth int) bool {
		if elem := rv.tree.Get(id); elem != nil {
			cache.entries = append(cache.entries, layoutEntry{
				id:     id,
				layout: elem.Layout(),
			})
		}
		return true
	})
	return cache
}

// applyItemOffset sets absolute positions for all elements in an item
// by adding (offsetX, offsetY) to the captured relative positions.
func (rv *RecyclerView) applyItemOffset(cache *itemLayoutCache, offsetX, offsetY float32) {
	for _, entry := range cache.entries {
		l := entry.layout
		l.Bounds.X += offsetX
		l.Bounds.Y += offsetY
		l.ContentBounds.X += offsetX
		l.ContentBounds.Y += offsetY
		rv.tree.SetLayout(entry.id, l)
	}
}

func (rv *RecyclerView) itemHeight(index int) float32 {
	if h, ok := rv.heights[index]; ok {
		return h
	}
	if rv.heightFn != nil {
		return rv.heightFn(index)
	}
	return rv.estimatedHeight
}

func (rv *RecyclerView) ensureCumY() {
	if rv.cumYValid && len(rv.cumY) == rv.itemCount+1 {
		return
	}
	if cap(rv.cumY) >= rv.itemCount+1 {
		rv.cumY = rv.cumY[:rv.itemCount+1]
	} else {
		rv.cumY = make([]float32, rv.itemCount+1)
	}
	y := float32(0)
	for i := 0; i < rv.itemCount; i++ {
		rv.cumY[i] = y
		y += rv.itemHeight(i)
	}
	rv.cumY[rv.itemCount] = y
	rv.cumYValid = true
}

func (rv *RecyclerView) updateContentHeight() {
	rv.ensureCumY()
	if rv.itemCount > 0 {
		rv.contentHeight = rv.cumY[rv.itemCount]
	} else {
		rv.contentHeight = 0
	}
}

func (rv *RecyclerView) clampScroll() {
	bounds := rv.Bounds()
	maxScroll := rv.contentHeight - bounds.Height
	if maxScroll < 0 {
		maxScroll = 0
	}
	if rv.scrollY < 0 {
		rv.scrollY = 0
	}
	if rv.scrollY > maxScroll {
		rv.scrollY = maxScroll
	}
}

// visibleRange returns the [startIdx, endIdx) of items that should be materialized.
func (rv *RecyclerView) visibleRange() (int, int) {
	bounds := rv.Bounds()
	if bounds.IsEmpty() || rv.itemCount == 0 {
		return 0, 0
	}

	rv.ensureCumY()

	// Binary search for first visible item
	viewTop := rv.scrollY
	viewBottom := rv.scrollY + bounds.Height

	start := rv.searchIndex(viewTop)
	end := rv.searchIndex(viewBottom) + 1

	// Apply buffer
	start -= rv.bufferCount
	end += rv.bufferCount

	if start < 0 {
		start = 0
	}
	if end > rv.itemCount {
		end = rv.itemCount
	}
	return start, end
}

// searchIndex finds the item index at the given scroll Y position (binary search).
func (rv *RecyclerView) searchIndex(y float32) int {
	lo, hi := 0, rv.itemCount
	for lo < hi {
		mid := (lo + hi) / 2
		if rv.cumY[mid+1] <= y {
			lo = mid + 1
		} else {
			hi = mid
		}
	}
	if lo >= rv.itemCount {
		lo = rv.itemCount - 1
	}
	if lo < 0 {
		lo = 0
	}
	return lo
}

// Reconcile creates/destroys widgets to match the visible range.
// Call this after scrolling or after changing itemCount.
func (rv *RecyclerView) Reconcile() {
	if rv.createFn == nil {
		return
	}

	rv.updateContentHeight()
	rv.clampScroll()

	start, end := rv.visibleRange()

	bounds := rv.Bounds()
	availW := bounds.Width

	// Destroy widgets outside visible range
	for idx, w := range rv.active {
		if idx < start || idx >= end {
			if rv.destroyFn != nil {
				rv.destroyFn(idx, w)
			} else {
				w.Destroy()
			}
			delete(rv.active, idx)
			delete(rv.itemLayouts, idx)
		}
	}

	// Create widgets for newly visible items
	for i := start; i < end; i++ {
		if _, exists := rv.active[i]; !exists {
			w := rv.createFn(i)
			rv.active[i] = w

			// Run CSS layout and capture relative positions
			if rv.layoutFn != nil && availW > 0 {
				h := rv.layoutFn(w, availW)
				if h > 0 {
					rv.heights[i] = h
					rv.cumYValid = false
				}
				rv.itemLayouts[i] = rv.captureRelativeLayout(w)
			}
		}
	}

	// Recompute cumulative heights if changed
	if !rv.cumYValid {
		rv.updateContentHeight()
	}

	// Check near-end trigger for infinite scroll
	if rv.onNearEnd != nil {
		bounds := rv.Bounds()
		maxScroll := rv.contentHeight - bounds.Height
		if maxScroll > 0 && rv.scrollY >= maxScroll-rv.nearEndPx {
			rv.onNearEnd()
		}
	}

	if rv.onScroll != nil {
		rv.onScroll(rv.scrollY)
	}
}

// Draw renders only the active (visible) widgets at their correct positions.
func (rv *RecyclerView) Draw(buf *render.CommandBuffer) {
	bounds := rv.Bounds()
	if bounds.IsEmpty() {
		return
	}

	buf.PushClip(bounds)

	rv.ensureCumY()

	// Position and draw each active widget
	start, end := rv.visibleRange()
	for i := start; i < end; i++ {
		w, ok := rv.active[i]
		if !ok {
			continue
		}

		itemY := bounds.Y + rv.cumY[i] - rv.scrollY
		itemH := rv.itemHeight(i)

		// Skip if entirely off-screen (belt + suspenders with clip)
		if itemY+itemH < bounds.Y || itemY > bounds.Y+bounds.Height {
			continue
		}

		// Apply layout: either offset captured CSS positions or set root directly
		if cache, ok := rv.itemLayouts[i]; ok {
			rv.applyItemOffset(cache, bounds.X, itemY)
		} else {
			rv.tree.SetLayout(w.ElementID(), core.LayoutResult{
				Bounds: uimath.NewRect(bounds.X, itemY, bounds.Width, itemH),
			})
		}

		w.Draw(buf)
	}

	// Draw scrollbar
	rv.drawScrollbar(buf, bounds)

	buf.PopClip()
}

func (rv *RecyclerView) drawScrollbar(buf *render.CommandBuffer, bounds uimath.Rect) {
	if rv.contentHeight <= bounds.Height {
		return
	}

	const barW = 6
	trackX := bounds.X + bounds.Width - barW - 2

	// Track
	buf.DrawRect(render.RectCmd{
		Bounds:    uimath.NewRect(trackX, bounds.Y+2, barW, bounds.Height-4),
		FillColor: uimath.RGBA(0, 0, 0, 0.05),
		Corners:   uimath.CornersAll(barW / 2),
	}, 10, 1)

	// Thumb
	trackH := bounds.Height - 4
	ratio := bounds.Height / rv.contentHeight
	if ratio > 1 {
		ratio = 1
	}
	thumbH := trackH * ratio
	if thumbH < 20 {
		thumbH = 20
	}
	maxScroll := rv.contentHeight - bounds.Height
	scrollRatio := float32(0)
	if maxScroll > 0 {
		scrollRatio = rv.scrollY / maxScroll
	}
	thumbY := bounds.Y + 2 + (trackH-thumbH)*scrollRatio

	thumbColor := uimath.RGBA(0, 0, 0, 0.25)
	if rv.scrollBarDrag {
		thumbColor = uimath.RGBA(0, 0, 0, 0.45)
	}
	buf.DrawRect(render.RectCmd{
		Bounds:    uimath.NewRect(trackX, thumbY, barW, thumbH),
		FillColor: thumbColor,
		Corners:   uimath.CornersAll(barW / 2),
	}, 11, 1)
}

// HandleScrollBarDown starts a scrollbar thumb drag.
func (rv *RecyclerView) HandleScrollBarDown(globalY float32) bool {
	if rv.contentHeight <= rv.Bounds().Height {
		return false
	}
	bounds := rv.Bounds()
	thumbY, thumbH := rv.thumbRect(bounds)
	if globalY >= thumbY && globalY <= thumbY+thumbH {
		rv.scrollBarDrag = true
		rv.dragStartY = globalY
		rv.dragStartScrl = rv.scrollY
		return true
	}
	return false
}

// HandleScrollBarMove updates scroll during drag.
func (rv *RecyclerView) HandleScrollBarMove(globalY float32) {
	if !rv.scrollBarDrag {
		return
	}
	bounds := rv.Bounds()
	trackH := bounds.Height - 4
	thumbH := rv.thumbHeight(bounds)
	maxThumbY := trackH - thumbH
	if maxThumbY <= 0 {
		return
	}
	dy := globalY - rv.dragStartY
	maxScroll := rv.contentHeight - bounds.Height
	scrollDelta := dy * maxScroll / maxThumbY
	rv.scrollY = rv.dragStartScrl + scrollDelta
	rv.clampScroll()
}

// HandleScrollBarUp ends a scrollbar drag.
func (rv *RecyclerView) HandleScrollBarUp() {
	rv.scrollBarDrag = false
}

func (rv *RecyclerView) thumbHeight(bounds uimath.Rect) float32 {
	if rv.contentHeight <= 0 {
		return 0
	}
	ratio := bounds.Height / rv.contentHeight
	if ratio > 1 {
		ratio = 1
	}
	h := (bounds.Height - 4) * ratio
	if h < 20 {
		h = 20
	}
	return h
}

func (rv *RecyclerView) thumbRect(bounds uimath.Rect) (y, h float32) {
	h = rv.thumbHeight(bounds)
	trackH := bounds.Height - 4
	maxThumbY := trackH - h
	maxScroll := rv.contentHeight - bounds.Height
	if maxScroll <= 0 {
		return bounds.Y + 2, h
	}
	ratio := rv.scrollY / maxScroll
	return bounds.Y + 2 + maxThumbY*ratio, h
}

// DestroyAll destroys all active widgets. Call when the RecyclerView is removed.
func (rv *RecyclerView) DestroyAll() {
	for idx, w := range rv.active {
		if rv.destroyFn != nil {
			rv.destroyFn(idx, w)
		} else {
			w.Destroy()
		}
	}
	rv.active = make(map[int]Widget)
	rv.itemLayouts = make(map[int]*itemLayoutCache)
}

// Destroy cleans up the RecyclerView and all its active widgets.
func (rv *RecyclerView) Destroy() {
	rv.DestroyAll()
	rv.Base.Destroy()
}
