// Package im provides an immediate-mode UI API.
//
// Unlike the retained-mode widget tree, immediate mode rebuilds the UI
// every frame from function calls. This is ideal for debug UIs, game HUDs,
// and rapid prototyping.
//
//	ctx := im.NewContext(buf, cfg)
//	ctx.Begin("panel", 10, 10, 300, 400)
//	if ctx.Button("Click Me") { ... }
//	ctx.Text("Score: 100")
//	val = ctx.Slider("speed", val, 0, 100)
//	ctx.End()
package im

import (
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
	"github.com/kasuganosora/ui/widget"
)

// Context is the immediate-mode UI context.
// Create one per frame and call widget functions on it.
type Context struct {
	buf    *render.CommandBuffer
	config *widget.Config

	// Input state (set before frame)
	MouseX, MouseY float32
	MouseDown      bool
	MouseClicked   bool // true on the frame mouse was pressed
	MouseReleased  bool

	// Internal state
	hotID    string // widget under mouse
	activeID string // widget being interacted with

	// Layout cursor
	cursorX float32
	cursorY float32
	panelX  float32
	panelY  float32
	panelW  float32
	panelH  float32
	padding float32
	spacing float32
	lineH   float32

	// Persistent widget state (keyed by ID)
	sliderValues map[string]float32
	checkValues  map[string]bool
	textValues   map[string]string
}

// NewContext creates an immediate-mode context.
func NewContext(buf *render.CommandBuffer, cfg *widget.Config) *Context {
	if cfg == nil {
		cfg = widget.DefaultConfig()
	}
	return &Context{
		buf:          buf,
		config:       cfg,
		padding:      cfg.SpaceSM,
		spacing:      cfg.SpaceXS,
		lineH:        cfg.FontSize * 1.5,
		sliderValues: make(map[string]float32),
		checkValues:  make(map[string]bool),
		textValues:   make(map[string]string),
	}
}

// SetInput updates the input state for the current frame.
func (ctx *Context) SetInput(mx, my float32, down, clicked, released bool) {
	ctx.MouseX = mx
	ctx.MouseY = my
	ctx.MouseDown = down
	ctx.MouseClicked = clicked
	ctx.MouseReleased = released
}

// Begin starts a panel at the given position.
func (ctx *Context) Begin(id string, x, y, w, h float32) {
	ctx.panelX = x
	ctx.panelY = y
	ctx.panelW = w
	ctx.panelH = h
	ctx.cursorX = x + ctx.padding
	ctx.cursorY = y + ctx.padding

	// Panel background
	ctx.buf.DrawRect(render.RectCmd{
		Bounds:      uimath.NewRect(x, y, w, h),
		FillColor:   uimath.RGBA(0.1, 0.1, 0.1, 0.9),
		BorderColor: uimath.RGBA(0.3, 0.3, 0.3, 1),
		BorderWidth: 1,
		Corners:     uimath.CornersAll(ctx.config.BorderRadius),
	}, 50, 1)
}

// End finishes the current panel.
func (ctx *Context) End() {
	ctx.hotID = ""
}

// Text draws a text label.
func (ctx *Context) Text(text string) {
	x := ctx.cursorX
	y := ctx.cursorY
	h := ctx.lineH

	if ctx.config.TextRenderer != nil {
		lh := ctx.config.TextRenderer.LineHeight(ctx.config.FontSize)
		ctx.config.TextRenderer.DrawText(ctx.buf, text, x, y+(h-lh)/2, ctx.config.FontSize, ctx.panelW-ctx.padding*2, uimath.ColorWhite, 1)
	} else {
		tw := float32(len(text)) * ctx.config.FontSize * 0.55
		th := ctx.config.FontSize * 1.2
		ctx.buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(x, y+(h-th)/2, tw, th),
			FillColor: uimath.ColorWhite,
			Corners:   uimath.CornersAll(2),
		}, 51, 1)
	}

	ctx.cursorY += h + ctx.spacing
}

// Button draws a button and returns true if clicked.
func (ctx *Context) Button(label string) bool {
	x := ctx.cursorX
	y := ctx.cursorY
	w := ctx.panelW - ctx.padding*2
	h := ctx.config.ButtonHeight

	bounds := uimath.NewRect(x, y, w, h)
	hovered := bounds.Contains(uimath.Vec2{X: ctx.MouseX, Y: ctx.MouseY})
	clicked := hovered && ctx.MouseClicked

	// Background
	bgColor := ctx.config.PrimaryColor
	if hovered {
		bgColor = ctx.config.HoverColor
	}
	if hovered && ctx.MouseDown {
		bgColor = ctx.config.ActiveColor
	}

	ctx.buf.DrawRect(render.RectCmd{
		Bounds:    bounds,
		FillColor: bgColor,
		Corners:   uimath.CornersAll(ctx.config.BorderRadius),
	}, 51, 1)

	// Label
	if ctx.config.TextRenderer != nil {
		lh := ctx.config.TextRenderer.LineHeight(ctx.config.FontSize)
		tw := ctx.config.TextRenderer.MeasureText(label, ctx.config.FontSize)
		ctx.config.TextRenderer.DrawText(ctx.buf, label, x+(w-tw)/2, y+(h-lh)/2, ctx.config.FontSize, w, uimath.ColorWhite, 1)
	} else {
		tw := float32(len(label)) * ctx.config.FontSize * 0.55
		th := ctx.config.FontSize * 1.2
		ctx.buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(x+(w-tw)/2, y+(h-th)/2, tw, th),
			FillColor: uimath.ColorWhite,
			Corners:   uimath.CornersAll(2),
		}, 52, 1)
	}

	ctx.cursorY += h + ctx.spacing
	return clicked
}

// Checkbox draws a checkbox and returns the new checked state.
func (ctx *Context) Checkbox(id string, label string) bool {
	checked := ctx.checkValues[id]
	x := ctx.cursorX
	y := ctx.cursorY
	h := ctx.lineH
	boxSize := float32(16)

	boxBounds := uimath.NewRect(x, y+(h-boxSize)/2, boxSize, boxSize)
	hovered := boxBounds.Contains(uimath.Vec2{X: ctx.MouseX, Y: ctx.MouseY})
	if hovered && ctx.MouseClicked {
		checked = !checked
		ctx.checkValues[id] = checked
	}

	// Box
	borderColor := ctx.config.BorderColor
	if hovered {
		borderColor = ctx.config.PrimaryColor
	}
	fillColor := uimath.Color{}
	if checked {
		fillColor = ctx.config.PrimaryColor
		borderColor = ctx.config.PrimaryColor
	}

	ctx.buf.DrawRect(render.RectCmd{
		Bounds:      boxBounds,
		FillColor:   fillColor,
		BorderColor: borderColor,
		BorderWidth: 1,
		Corners:     uimath.CornersAll(3),
	}, 51, 1)

	// Checkmark
	if checked {
		ctx.buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(x+4, y+(h-boxSize)/2+4, boxSize-8, boxSize-8),
			FillColor: uimath.ColorWhite,
			Corners:   uimath.CornersAll(1),
		}, 52, 1)
	}

	// Label
	labelX := x + boxSize + ctx.config.SpaceSM
	if ctx.config.TextRenderer != nil {
		lh := ctx.config.TextRenderer.LineHeight(ctx.config.FontSize)
		ctx.config.TextRenderer.DrawText(ctx.buf, label, labelX, y+(h-lh)/2, ctx.config.FontSize, ctx.panelW-ctx.padding*2-boxSize-ctx.config.SpaceSM, uimath.ColorWhite, 1)
	} else {
		tw := float32(len(label)) * ctx.config.FontSize * 0.55
		th := ctx.config.FontSize * 1.2
		ctx.buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(labelX, y+(h-th)/2, tw, th),
			FillColor: uimath.ColorWhite,
			Corners:   uimath.CornersAll(2),
		}, 52, 1)
	}

	ctx.cursorY += h + ctx.spacing
	return checked
}

// Slider draws a horizontal slider and returns the new value.
func (ctx *Context) Slider(id string, min, max float32) float32 {
	val, ok := ctx.sliderValues[id]
	if !ok {
		val = min
	}

	x := ctx.cursorX
	y := ctx.cursorY
	w := ctx.panelW - ctx.padding*2
	h := float32(20)
	trackH := float32(4)

	// Track
	trackY := y + (h-trackH)/2
	ctx.buf.DrawRect(render.RectCmd{
		Bounds:    uimath.NewRect(x, trackY, w, trackH),
		FillColor: uimath.RGBA(0.3, 0.3, 0.3, 1),
		Corners:   uimath.CornersAll(trackH / 2),
	}, 51, 1)

	// Filled portion
	ratio := float32(0)
	if max > min {
		ratio = (val - min) / (max - min)
	}
	if ratio < 0 {
		ratio = 0
	}
	if ratio > 1 {
		ratio = 1
	}
	fillW := w * ratio
	ctx.buf.DrawRect(render.RectCmd{
		Bounds:    uimath.NewRect(x, trackY, fillW, trackH),
		FillColor: ctx.config.PrimaryColor,
		Corners:   uimath.CornersAll(trackH / 2),
	}, 52, 1)

	// Thumb
	thumbR := float32(8)
	thumbX := x + fillW - thumbR
	thumbY := y + (h-thumbR*2)/2
	ctx.buf.DrawRect(render.RectCmd{
		Bounds:    uimath.NewRect(thumbX, thumbY, thumbR*2, thumbR*2),
		FillColor: uimath.ColorWhite,
		Corners:   uimath.CornersAll(thumbR),
	}, 53, 1)

	// Interaction: use activeID so drag continues even if cursor leaves track
	trackBounds := uimath.NewRect(x, y, w, h)
	mousePos := uimath.Vec2{X: ctx.MouseX, Y: ctx.MouseY}
	if trackBounds.Contains(mousePos) && ctx.MouseClicked {
		ctx.activeID = id
	}
	if ctx.activeID == id {
		if ctx.MouseDown {
			t := (ctx.MouseX - x) / w
			if t < 0 {
				t = 0
			}
			if t > 1 {
				t = 1
			}
			val = min + t*(max-min)
			ctx.sliderValues[id] = val
		} else {
			ctx.activeID = ""
		}
	}

	ctx.cursorY += h + ctx.spacing
	return val
}

// Separator draws a horizontal line.
func (ctx *Context) Separator() {
	x := ctx.cursorX
	y := ctx.cursorY + ctx.spacing
	w := ctx.panelW - ctx.padding*2

	ctx.buf.DrawRect(render.RectCmd{
		Bounds:    uimath.NewRect(x, y, w, 1),
		FillColor: uimath.RGBA(0.4, 0.4, 0.4, 1),
	}, 51, 1)

	ctx.cursorY = y + 1 + ctx.spacing
}

// ProgressBar draws a progress bar.
func (ctx *Context) ProgressBar(ratio float32) {
	x := ctx.cursorX
	y := ctx.cursorY
	w := ctx.panelW - ctx.padding*2
	h := float32(8)

	if ratio < 0 {
		ratio = 0
	}
	if ratio > 1 {
		ratio = 1
	}

	// Track
	ctx.buf.DrawRect(render.RectCmd{
		Bounds:    uimath.NewRect(x, y, w, h),
		FillColor: uimath.RGBA(0.2, 0.2, 0.2, 1),
		Corners:   uimath.CornersAll(h / 2),
	}, 51, 1)

	// Fill
	if ratio > 0 {
		ctx.buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(x, y, w*ratio, h),
			FillColor: ctx.config.PrimaryColor,
			Corners:   uimath.CornersAll(h / 2),
		}, 52, 1)
	}

	ctx.cursorY += h + ctx.spacing
}

// Space adds vertical spacing.
func (ctx *Context) Space(h float32) {
	ctx.cursorY += h
}
