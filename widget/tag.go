package widget

import (
	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/event"
	"github.com/kasuganosora/ui/layout"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// TagTheme controls the tag appearance.
type TagTheme uint8

const (
	TagThemeDefault TagTheme = iota
	TagThemePrimary
	TagThemeWarning
	TagThemeDanger
	TagThemeSuccess
)

// TagShape controls the tag border radius style.
type TagShape uint8

const (
	TagShapeSquare TagShape = iota // default borderRadius
	TagShapeRound                  // height/2 borderRadius (pill)
	TagShapeMark                   // right side rounded, left side flat
)

// TagVariant controls the tag color variant.
type TagVariant uint8

const (
	VariantDark         TagVariant = iota // filled background with white text
	VariantLight                          // light tinted bg with colored text (default look)
	VariantOutline                        // transparent bg with colored border + text
	VariantLightOutline                   // light bg with colored border + text
)

// Tag displays a small labeled tag/badge.
type Tag struct {
	Base
	content  string
	theme    TagTheme
	color    uimath.Color // custom color (zero = use theme default)
	size     Size
	closable bool
	onClose  func()
	onClick  func()
	shape    TagShape
	variant  TagVariant
	disabled bool
	maxWidth float32
	title    string

	closeID core.ElementID // clickable close button element
}

// NewTag creates a tag with the given label.
func NewTag(tree *core.Tree, label string, cfg *Config) *Tag {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	t := &Tag{
		Base:    NewBase(tree, core.TypeCustom, cfg),
		content: label,
		variant: VariantLight,
	}
	t.style.Display = layout.DisplayFlex
	t.style.AlignItems = layout.AlignCenter
	t.style.JustifyContent = layout.JustifyCenter
	t.style.Height = layout.Px(24) // Medium default
	t.style.Padding = layout.EdgeValues{
		Left:  layout.Px(cfg.SpaceSM),
		Right: layout.Px(cfg.SpaceSM),
	}
	tree.SetProperty(t.id, "text", label)

	// Create close button element (hidden until closable is set)
	t.closeID = tree.CreateElement(core.TypeCustom)
	tree.AppendChild(t.id, t.closeID)
	tree.AddHandler(t.closeID, event.MouseClick, func(e *event.Event) {
		if t.closable && !t.disabled && t.onClose != nil {
			t.onClose()
		}
	})

	// Click handler on the tag itself
	tree.AddHandler(t.id, event.MouseClick, func(e *event.Event) {
		if !t.disabled && t.onClick != nil {
			t.onClick()
		}
	})

	return t
}

func (t *Tag) Content() string      { return t.content }
func (t *Tag) Theme() TagTheme      { return t.theme }
func (t *Tag) Size() Size           { return t.size }
func (t *Tag) Closable() bool       { return t.closable }
func (t *Tag) Shape() TagShape      { return t.shape }
func (t *Tag) Variant() TagVariant  { return t.variant }
func (t *Tag) Disabled() bool       { return t.disabled }
func (t *Tag) MaxWidth() float32    { return t.maxWidth }
func (t *Tag) Title() string        { return t.title }

func (t *Tag) SetContent(content string) {
	t.content = content
	t.tree.SetProperty(t.id, "text", content)
}

func (t *Tag) SetTheme(th TagTheme)      { t.theme = th }
func (t *Tag) SetColor(c uimath.Color)   { t.color = c }
func (t *Tag) SetSize(s Size)            { t.size = s }
func (t *Tag) SetClosable(c bool)        { t.closable = c }
func (t *Tag) OnClose(fn func())         { t.onClose = fn }
func (t *Tag) OnClick(fn func())         { t.onClick = fn }
func (t *Tag) SetShape(s TagShape)       { t.shape = s }
func (t *Tag) SetVariant(v TagVariant)   { t.variant = v }
func (t *Tag) SetDisabled(d bool)        { t.disabled = d }
func (t *Tag) SetMaxWidth(w float32)     { t.maxWidth = w }
func (t *Tag) SetTitle(title string)     { t.title = title }

// sizeHeight returns the tag height based on size.
func (t *Tag) sizeHeight() float32 {
	switch t.size {
	case SizeSmall:
		return 20
	case SizeLarge:
		return 28
	default:
		return 24
	}
}

// tagBaseColor returns the primary color for the tag type.
func (t *Tag) tagBaseColor() uimath.Color {
	if t.color != (uimath.Color{}) {
		return t.color
	}
	switch t.theme {
	case TagThemeSuccess:
		return uimath.ColorHex("#2ba471")
	case TagThemeWarning:
		return uimath.ColorHex("#e37318")
	case TagThemeDanger:
		return uimath.ColorHex("#d54941")
	case TagThemePrimary:
		return uimath.ColorHex("#0052d9")
	default:
		return t.config.TextColor
	}
}

func (t *Tag) tagColors() (bg, border, text uimath.Color) {
	base := t.tagBaseColor()

	switch t.variant {
	case VariantDark:
		// Filled background with white text
		return base, base, uimath.ColorWhite
	case VariantOutline:
		// Transparent bg with colored border + text
		return uimath.Color{R: 0, G: 0, B: 0, A: 0}, base, base
	case VariantLightOutline:
		// Light bg with colored border + text
		return uimath.Color{R: base.R, G: base.G, B: base.B, A: 0.1}, base, base
	default: // VariantLight
		// Light tinted background with colored text
		if t.color != (uimath.Color{}) {
			return uimath.Color{R: t.color.R, G: t.color.G, B: t.color.B, A: 0.1},
				t.color, t.color
		}
		switch t.theme {
		case TagThemeSuccess:
			return uimath.ColorHex("#e3f9e9"), uimath.ColorHex("#c6f3d7"), uimath.ColorHex("#2ba471")
		case TagThemeWarning:
			return uimath.ColorHex("#fff1e9"), uimath.ColorHex("#ffd9c2"), uimath.ColorHex("#e37318")
		case TagThemeDanger:
			return uimath.ColorHex("#fff0ed"), uimath.ColorHex("#ffd8d2"), uimath.ColorHex("#d54941")
		case TagThemePrimary:
			return uimath.ColorHex("#f2f3ff"), uimath.ColorHex("#b5c7ff"), uimath.ColorHex("#0052d9")
		default:
			return uimath.ColorHex("#f3f3f3"), uimath.ColorHex("#dcdcdc"), t.config.TextColor
		}
	}
}

func (t *Tag) Draw(buf *render.CommandBuffer) {
	bounds := t.Bounds()
	if bounds.IsEmpty() {
		return
	}

	cfg := t.config
	bgColor, borderColor, textColor := t.tagColors()
	h := t.sizeHeight()
	fontSize := cfg.FontSizeSm

	// Disabled: reduce opacity
	opacity := float32(1)
	if t.disabled {
		opacity = 0.5
	}

	// Compute corners based on shape
	var corners uimath.Corners
	switch t.shape {
	case TagShapeRound:
		corners = uimath.CornersAll(h / 2)
	case TagShapeMark:
		// Left side flat, right side rounded
		r := cfg.BorderRadius / 2
		corners = uimath.Corners{
			TopLeft:     2,
			BottomLeft:  2,
			TopRight:    r,
			BottomRight: r,
		}
	default: // ShapeSquare
		corners = uimath.CornersAll(cfg.BorderRadius / 2)
	}

	// Border width
	bw := float32(1)
	if t.variant == VariantOutline || t.variant == VariantLightOutline {
		bw = 1
	}

	// Background
	buf.DrawRect(render.RectCmd{
		Bounds:      bounds,
		FillColor:   bgColor,
		BorderColor: borderColor,
		BorderWidth: bw,
		Corners:     corners,
	}, 0, opacity)

	// Label
	closeSpace := float32(0)
	if t.closable {
		closeSpace = fontSize + 4 // space for close button
	}
	if t.content != "" {
		if cfg.TextRenderer != nil {
			tw := cfg.TextRenderer.MeasureText(t.content, fontSize)
			lh := cfg.TextRenderer.LineHeight(fontSize)
			availW := bounds.Width - closeSpace
			tx := bounds.X + (availW-tw)/2
			ty := bounds.Y + (bounds.Height-lh)/2
			maxW := availW
			cfg.TextRenderer.DrawText(buf, t.content, tx, ty, fontSize, maxW, textColor, opacity)
		} else {
			textW := float32(len(t.content)) * fontSize * 0.55
			maxW := bounds.Width - cfg.SpaceSM*2 - closeSpace
			if textW > maxW {
				textW = maxW
			}
			textH := fontSize * 1.2
			tx := bounds.X + (bounds.Width-textW-closeSpace)/2
			ty := bounds.Y + (bounds.Height-textH)/2
			buf.DrawRect(render.RectCmd{
				Bounds:    uimath.NewRect(tx, ty, textW, textH),
				FillColor: textColor,
				Corners:   uimath.CornersAll(2),
			}, 1, opacity)
		}
	}

	// Close button (X)
	if t.closable {
		closeSize := float32(12)
		closeX := bounds.X + bounds.Width - cfg.SpaceSM - closeSize
		closeY := bounds.Y + (bounds.Height-closeSize)/2
		hitSize := closeSize + 4 // slightly larger hit area

		// Draw X as two crossing lines
		// Diagonal line 1: top-left to bottom-right
		buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(closeX+closeSize/2-1, closeY, 2, closeSize),
			FillColor: textColor,
		}, 1, opacity)
		// Diagonal line 2: approximate cross with horizontal
		buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(closeX, closeY+closeSize/2-1, closeSize, 2),
			FillColor: textColor,
		}, 1, opacity)

		// Set layout on close element for hit testing
		t.tree.SetLayout(t.closeID, core.LayoutResult{
			Bounds: uimath.NewRect(closeX-2, closeY-2, hitSize, hitSize),
		})
	}

	t.DrawChildren(buf)
}
