package widget

import (
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/platform"
	"github.com/kasuganosora/ui/render"
)

// TextDrawer renders text into a command buffer.
// When set on Config, widgets use real glyph commands instead of placeholder rects.
type TextDrawer interface {
	DrawText(buf *render.CommandBuffer, text string, x, y, fontSize, maxWidth float32, color uimath.Color, opacity float32)
	// LineHeight returns the font line height (ascent + descent) for the given size.
	// Used by widgets for accurate vertical centering.
	LineHeight(fontSize float32) float32
	// MeasureText returns the width of the given text at the specified font size.
	MeasureText(text string, fontSize float32) float32
}

// ZTextDrawer is a TextDrawer that can place glyphs on a chosen z-order.
// Fills under a label use a lower z so the glyphs stay visible.
type ZTextDrawer interface {
	TextDrawer
	DrawTextZ(buf *render.CommandBuffer, text string, x, y, fontSize, maxWidth float32, color uimath.Color, opacity float32, z int32)
}

// TrackingTextDrawer is an optional extension for size-specific letter
// spacing (§15). Widgets keep working with a plain TextDrawer; only hosts
// that implement this interface get optical tracking. Keeping it optional
// avoids breaking the many embedded/test drawers in the repository.
type TrackingTextDrawer interface {
	DrawTextTracked(buf *render.CommandBuffer, text string, x, y, fontSize, maxWidth float32, color uimath.Color, opacity float32, z int32, tracking float32)
}

// CaptionTextDrawer is an optional extension for small-text hierarchy (§15).
// At 12px a regular-weight CJK outline has almost no fully covered pixels,
// so glyphs read as ~#A3A3A3 grey even when the theme color is much darker.
// Hosts that can select a medium-weight face use it for captions; hosts that
// cannot simply fall back to emitText. The weight is a property of the host's
// font registration, not a widget parameter, so no widget learns about font
// IDs or platform font files.
type CaptionTextDrawer interface {
	DrawCaptionText(buf *render.CommandBuffer, text string, x, y, fontSize, maxWidth float32, color uimath.Color, opacity float32, z int32, tracking float32)
}

// OverlayWindowHost owns OS-level auxiliary windows for floating widgets such
// as tooltips. It is implemented by the application shell, which keeps the
// widget package independent of the root package.
//
// draw is invoked once per frame with the Config bound to that auxiliary
// window's own renderer. Texture handles are local to a render backend, so a
// floating window must not reuse the main window's glyph atlas or icons.
// Coordinates passed to OpenOverlay are in the main window's client space;
// the host converts them to screen coordinates.
type OverlayWindowHost interface {
	OpenOverlay(key Widget, draw func(cfg *Config, buf *render.CommandBuffer), x, y, w, h float32) error
	CloseOverlay(key Widget)
}

// Glyph z-orders for widgets that paint an opaque fill with a non-zero z.
// Backends submit commands by ascending ZOrder, so a label emitted at the
// default z=0 is hidden by any fill with a higher z. Each value is exactly
// one above the highest fill inside that widget; keeping them local (instead
// of one global "always on top" value) preserves overlay/modal layering.
const (
	alertTextZ   int32 = 2  // icon dot fill: z=1
	tableTextZ   int32 = 3  // header/stripe fills: z=1, dividers: z=2
	stepsTextZ   int32 = 4  // dot fills: z=2, icon fragments: z=3
	badgeTextZ   int32 = 6  // badge pill fill: z=5
	messageTextZ int32 = 23 // pill: z=20, icon dot: z=21, close marks: z=22
)

// Glyph z-orders for overlay, dropdown and modal widgets. These states are
// usually inactive in a normal frame, so the same "background at z>0 but label
// at z=0" defect stayed hidden until the state was activated. Each value is
// one above the highest opaque fill that widget emits; staying local keeps a
// modal from painting over an even higher overlay layer.
const (
	sliderMarkTextZ   int32 = 3  // slider track/ticks: up to z=2
	dropdownTextZ     int32 = 43 // calendar/dropdown panels: up to z=42
	tooltipTextZ      int32 = 22 // tooltip / slider bubble: up to z=21
	dialogTextZ       int32 = 14 // dialog / message box: up to z=13
	drawerTextZ       int32 = 20 // drawer: up to z=19
	popoverTextZ      int32 = 42 // popover: up to z=41
	popconfirmTextZ   int32 = 54 // popconfirm: up to z=53
	notificationTextZ int32 = 55 // notification: up to z=54
	dockFloatTextZ    int32 = 53 // floating dock panel: up to z=52
	menuTextZ         int32 = 63 // context menu: up to z=62
	guideTextZ        int32 = 84 // guide card: up to z=83
	subwindowTextZ    int32 = 33 // sub window: up to z=32
)

// emitText draws with an explicit z when the host text drawer supports it,
// and falls back to the plain DrawText call for simpler implementations.
func emitText(cfg *Config, buf *render.CommandBuffer, text string, x, y, size, maxW float32, color uimath.Color, opacity float32, z int32) {
	if cfg == nil || cfg.TextRenderer == nil || text == "" {
		return
	}
	// §15: pass the size-specific tracking when the host renderer supports it.
	if tracked, ok := cfg.TextRenderer.(TrackingTextDrawer); ok {
		tracked.DrawTextTracked(buf, text, x, y, size, maxW, color, opacity, z, cfg.Tracking(size))
		return
	}
	if zdraw, ok := cfg.TextRenderer.(ZTextDrawer); ok {
		zdraw.DrawTextZ(buf, text, x, y, size, maxW, color, opacity, z)
		return
	}
	cfg.TextRenderer.DrawText(buf, text, x, y, size, maxW, color, opacity)
}

// emitOverlayText draws text into the overlay command list. Backends submit
// Commands() and Overlays() as two independently z-sorted batches, overlays
// last. A widget that paints its panel with DrawOverlay therefore hides any
// glyph left in Commands(), no matter how high that glyph's z-order is. This
// helper lifts the glyph commands into the same list (the same mechanism
// Select already uses for its dropdown labels).
func emitOverlayText(cfg *Config, buf *render.CommandBuffer, text string, x, y, size, maxW float32, color uimath.Color, opacity float32, z int32) {
	if cfg == nil || cfg.TextRenderer == nil || text == "" {
		return
	}
	if buf.LayerBase() != 0 {
		// place() already routes commands emitted inside a layer into Overlays().
		emitText(cfg, buf, text, x, y, size, maxW, color, opacity, z)
		return
	}
	before := buf.Len()
	emitText(cfg, buf, text, x, y, size, maxW, color, opacity, z)
	buf.MoveToOverlay(before, z)
}

// emitCaptionText draws a small label with the caption weight when the host
// supports it. Size, tracking and leading stay in Config so Dynamic Type
// scales them together (§15: weight + size + leading as a set).
func emitCaptionText(cfg *Config, buf *render.CommandBuffer, text string, x, y, size, maxW float32, color uimath.Color, opacity float32, z int32) {
	if cfg == nil || cfg.TextRenderer == nil || text == "" {
		return
	}
	if caption, ok := cfg.TextRenderer.(CaptionTextDrawer); ok {
		caption.DrawCaptionText(buf, text, x, y, size, maxW, color, opacity, z, cfg.Tracking(size))
		return
	}
	emitText(cfg, buf, text, x, y, size, maxW, color, opacity, z)
}

// Size represents component size variants.
type Size uint8

const (
	SizeMedium Size = iota // default
	SizeSmall
	SizeLarge
)

// Status represents input validation state.
type Status uint8

const (
	StatusDefault Status = iota
	StatusSuccess
	StatusWarning
	StatusError
)

// Config holds global UI configuration and theme colors.
// Acts as a ConfigProvider — pass it when creating widgets.
type Config struct {
	// Theme colors
	PrimaryColor     uimath.Color
	TextColor        uimath.Color
	BgColor          uimath.Color
	BorderColor      uimath.Color
	DisabledColor    uimath.Color
	HoverColor       uimath.Color
	ActiveColor      uimath.Color
	FocusBorderColor uimath.Color
	ErrorColor       uimath.Color
	WarningColor     uimath.Color
	SuccessColor     uimath.Color

	// Semantic colors for hierarchy. Secondary text must stay legible on
	// floating surfaces; using DisabledColor for body copy drops contrast
	// below readable levels. Mask dims the background behind modal tasks.
	TextSecondaryColor uimath.Color
	MaskColor          uimath.Color

	// Floating-surface material. A real blurred shadow separates a panel from
	// its background without a hard 1px outline; heavier values read as
	// thicker surfaces.
	ShadowColor uimath.Color
	ShadowY     float32
	ShadowBlur  float32

	// Accessibility (apple-design §14). The three signals are independent.
	// ReducedMotion freezes decorative loops at a legible static state; it
	// never disables functional state changes such as auto-dismiss timers.
	// ReducedTransparency makes translucent materials (scrim, hairline,
	// shadow) frostier and more solid. MoreContrast gives floating surfaces a
	// defined, contrasting outline instead of relying on shadow alone.
	ReducedMotion       bool
	ReducedTransparency bool
	MoreContrast        bool

	// Typography
	FontID     uint32
	FontSize   float32
	FontSizeSm float32
	FontSizeLg float32
	LineHeight float32
	// TextScale is the Dynamic Type multiplier. SetTextScale rescales the
	// three type sizes, line height and spacing tokens together so layout
	// grows with the text instead of clipping it (§15).
	TextScale float32

	// AnimateOverlays enables the anchored enter/exit transition (§7) for
	// floating widgets. It is off for Config values created directly so
	// headless renderers and visual regression frames stay deterministic;
	// the GUI host turns it on. ReducedMotion still removes the spatial travel.
	AnimateOverlays bool

	// Text renderer (optional — falls back to placeholder rects if nil)
	TextRenderer TextDrawer

	// Window reference for IME positioning and context menus (optional)
	Window platform.Window

	// Platform reference for clipboard operations (optional)
	Platform platform.Platform

	// OverlayHost owns auxiliary OS windows for floating widgets (optional).
	// When nil, floating widgets fall back to drawing inside the main tree.
	OverlayHost OverlayWindowHost

	// Spacing
	SpaceXS float32
	SpaceSM float32
	SpaceMD float32
	SpaceLG float32
	SpaceXL float32

	// Border
	BorderRadius float32
	BorderWidth  float32

	// Component defaults
	ButtonHeight float32
	InputHeight  float32
	IconSize     float32

	// Rendering backend for texture creation (optional — needed for image loading)
	Backend render.Backend

	// NetClient handles HTTP/HTTPS and data: URL image loading.
	// If nil, <img src> only loads local file paths.
	NetClient NetFetcher

	// Icon registry for SVG icon lookup by name (e.g., Material Design Icons)
	IconRegistry IconLookup
}

// IconLookup provides icon texture lookup by name and size.
type IconLookup interface {
	Get(name string, size int) (render.TextureHandle, bool)
	Has(name string) bool
}

// NetFetcher loads raw bytes from URLs (http/https/data:).
// Implemented by net.Client — set Config.NetClient to enable remote image loading.
type NetFetcher interface {
	Fetch(rawURL string) ([]byte, error)
	FetchAsync(rawURL string, cb func([]byte, error))
}

// DrawMDIcon draws a Material Design Icon at the given position.
// Returns true if the icon was drawn, false if no registry or icon not found.
func (c *Config) DrawMDIcon(buf *render.CommandBuffer, name string, x, y, size float32, clr uimath.Color, zOrder int32, opacity float32) bool {
	if c.IconRegistry == nil {
		return false
	}
	pixelSize := int(size)
	if pixelSize < 1 {
		pixelSize = 1
	}
	tex, ok := c.IconRegistry.Get(name, pixelSize)
	if !ok {
		return false
	}
	buf.DrawImage(render.ImageCmd{
		Texture: tex,
		SrcRect: uimath.NewRect(0, 0, 1, 1),
		DstRect: uimath.NewRect(x, y, size, size),
		Tint:    clr,
	}, zOrder, opacity)
	return true
}

// DrawMDIconOverlay is like DrawMDIcon but draws to the overlay layer.
func (c *Config) DrawMDIconOverlay(buf *render.CommandBuffer, name string, x, y, size float32, clr uimath.Color, zOrder int32, opacity float32) bool {
	if c.IconRegistry == nil {
		return false
	}
	pixelSize := int(size)
	if pixelSize < 1 {
		pixelSize = 1
	}
	tex, ok := c.IconRegistry.Get(name, pixelSize)
	if !ok {
		return false
	}
	before := buf.Len()
	buf.DrawImage(render.ImageCmd{
		Texture: tex,
		SrcRect: uimath.NewRect(0, 0, 1, 1),
		DstRect: uimath.NewRect(x, y, size, size),
		Tint:    clr,
	}, zOrder, opacity)
	buf.MoveToOverlay(before, zOrder)
	return true
}

// DefaultConfig returns a default configuration with light theme defaults.
func DefaultConfig() *Config {
	cfg := &Config{
		// brand-color-7 (#0052d9)
		PrimaryColor:       uimath.ColorHex("#0052d9"),
		TextColor:          uimath.RGBA(0, 0, 0, 0.9),
		BgColor:            uimath.ColorWhite,
		BorderColor:        uimath.ColorHex("#dcdcdc"),
		DisabledColor:      uimath.RGBA(0, 0, 0, 0.26),
		HoverColor:         uimath.ColorHex("#366ef4"),
		ActiveColor:        uimath.ColorHex("#003cab"),
		FocusBorderColor:   uimath.ColorHex("#0052d9"),
		ErrorColor:         uimath.ColorHex("#d54941"),
		WarningColor:       uimath.ColorHex("#e37318"),
		SuccessColor:       uimath.ColorHex("#2ba471"),
		TextSecondaryColor: uimath.RGBA(0, 0, 0, 0.6),
		MaskColor:          uimath.RGBA(0, 0, 0, 0.45),
		ShadowColor:        uimath.RGBA(0, 0, 0, 0.12),
		ShadowY:            4,
		ShadowBlur:         16,

		FontID:     0,
		FontSize:   14,
		FontSizeSm: 12,
		FontSizeLg: 16,
		LineHeight: 1.5,
		TextScale:  1,

		SpaceXS: 4,
		SpaceSM: 8,
		SpaceMD: 16,
		SpaceLG: 24,
		SpaceXL: 32,

		// radius-medium = 6
		BorderRadius: 6,
		BorderWidth:  1,

		ButtonHeight: 32,
		InputHeight:  32,
		IconSize:     16,
	}
	cfg.ApplyAccessibility(AccessibilityFromEnv())
	return cfg
}

// Hairline returns a low-contrast separator derived from the theme border.
// Section dividers should recede instead of reading as a hard 1px outline,
// and deriving the color keeps dark themes consistent without a new token.
// Reduced transparency / more contrast keep the full border so the divider
// no longer depends on a faint alpha overlay (§14).
func (c *Config) Hairline() uimath.Color {
	clr := c.BorderColor
	if c.ReducedTransparency || c.MoreContrast {
		return clr
	}
	clr.A *= 0.45
	return clr
}

// EffectiveMaskColor dims the background behind a modal task. The theme's
// translucent scrim stays for normal rendering; reduced transparency and
// higher contrast both ask for a frostier, near-solid surface (§14).
func (c *Config) EffectiveMaskColor() uimath.Color {
	m := c.MaskColor
	if (c.ReducedTransparency || c.MoreContrast) && m.A < 0.75 {
		m.A = 0.75
	}
	return m
}

// PanelBorderColor and PanelBorderWidth describe the outline of a floating
// surface. They are zero normally (depth comes from the shadow) and become a
// defined, contrasting edge under prefers-contrast: more (§14). Splitting the
// two lets each RectCmd add both fields without introducing a local variable.
func (c *Config) PanelBorderColor() uimath.Color {
	if !c.MoreContrast {
		return uimath.Color{}
	}
	return c.TextColor
}

func (c *Config) PanelBorderWidth() float32 {
	if !c.MoreContrast {
		return 0
	}
	if c.BorderWidth > 0 {
		return c.BorderWidth
	}
	return 1
}

// FloatingShadow builds the SDF shadow for a floating panel. Material weight
// follows surface size (apple-design §12): sheets, dialogs and subwindows use
// heavy=true so they read as thicker than chips, menus and small popovers.
// Reduced transparency / more contrast make the shadow itself more solid so
// depth no longer relies on a faint translucent overlay (§14). Every floating
// panel must pair this with a BgColor surface in the SAME command batch —
// backends z-sort Commands() and Overlays() independently.
func (c *Config) FloatingShadow(bounds uimath.Rect, heavy bool) render.ShadowCmd {
	blur := c.ShadowBlur
	if heavy {
		blur *= 2
	}
	color := c.ShadowColor
	if c.ReducedTransparency || c.MoreContrast {
		color.A *= 2
		if color.A > 1 {
			color.A = 1
		}
	}
	return render.ShadowCmd{
		Bounds:     bounds,
		Corners:    uimath.CornersAll(c.BorderRadius),
		OffsetY:    c.ShadowY,
		BlurRadius: blur,
		Color:      color,
	}
}

// SizeHeight returns the component height for a given size.
func (c *Config) SizeHeight(s Size) float32 {
	switch s {
	case SizeSmall:
		return 24
	case SizeLarge:
		return 40
	default:
		return 32
	}
}

// SizeFontSize returns the font size for a given component size.
func (c *Config) SizeFontSize(s Size) float32 {
	switch s {
	case SizeSmall:
		return c.FontSizeSm
	case SizeLarge:
		return c.FontSizeLg
	default:
		return c.FontSize
	}
}

// Apple §15 — optical sizing. Tracking and leading are size-specific: large
// display text wants negative tracking and tight leading, small captions want
// positive tracking and looser leading. A single fixed value is wrong
// somewhere in the hierarchy.
//
// LeadingRatio is anchored at 1.2 for the 14px body size (the framework's
// existing placeholder leading), then tightens as the type grows and loosens
// as it shrinks. That keeps every existing body-size layout pixel-identical
// while making captions and headings optically correct.
func (c *Config) LeadingRatio(size float32) float32 {
	switch {
	case size <= 10:
		return 1.4
	case size <= 12:
		return 1.32
	case size <= 14:
		return 1.2
	case size <= 18:
		return 1.15
	case size <= 24:
		return 1.1
	default:
		return 1.05
	}
}

// Leading returns the line box height for a given size.
func (c *Config) Leading(size float32) float32 {
	if c == nil {
		return size * 1.2
	}
	return size * c.LeadingRatio(size)
}

// Tracking returns the letter-spacing for a given size, in logical pixels.
// Small captions get a small positive bump for legibility; large headings get
// negative tracking so they do not read as spaced-out letters.
func (c *Config) Tracking(size float32) float32 {
	switch {
	case size <= 10:
		return size * 0.03
	case size <= 12:
		return size * 0.02
	case size <= 16:
		return 0
	case size <= 24:
		return -size * 0.01
	default:
		return -size * 0.02
	}
}

// CaptionSize is the small-text size after Dynamic Type scaling. It never
// drops below the 12px legibility floor, so shrinking the type scale cannot
// make a caption unreadable.
//
// SetTextScale already multiplies FontSizeSm in place: TextScale records the
// applied multiplier, it is not a second factor to apply here. Multiplying
// twice would turn a 2x setting into a 4x caption and overflow every row.
func (c *Config) CaptionSize() float32 {
	if c == nil {
		return 12
	}
	size := c.FontSizeSm
	if size < 12 {
		size = 12
	}
	return size
}

// SetTextScale applies a Dynamic Type multiplier to the whole type system,
// including spacing tokens. Apple §15 asks layout to scale *with* the text;
// scaling only the three font sizes would let a larger caption overflow the
// fixed padding around it.
func (c *Config) SetTextScale(scale float32) {
	if c == nil || scale <= 0 {
		return
	}
	old := c.TextScale
	if old <= 0 {
		old = 1
	}
	ratio := scale / old
	c.FontSize *= ratio
	c.FontSizeSm *= ratio
	c.FontSizeLg *= ratio
	c.LineHeight *= ratio
	c.SpaceXS *= ratio
	c.SpaceSM *= ratio
	c.SpaceMD *= ratio
	c.SpaceLG *= ratio
	c.SpaceXL *= ratio
	c.TextScale = scale
}

// StatusBorderColor returns the border color for a validation status.
func (c *Config) StatusBorderColor(s Status) uimath.Color {
	switch s {
	case StatusSuccess:
		return c.SuccessColor
	case StatusWarning:
		return c.WarningColor
	case StatusError:
		return c.ErrorColor
	default:
		return c.BorderColor
	}
}
