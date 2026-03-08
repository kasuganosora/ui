package widget

import (
	"strconv"

	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/layout"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// ProgressTheme controls the progress bar visual style.
type ProgressTheme uint8

const (
	ProgressThemeLine   ProgressTheme = iota // Horizontal bar (default)
	ProgressThemePlump                       // Thick bar with label inside
	ProgressThemeCircle                      // Circular ring
)

// Deprecated aliases for backward compatibility.
const (
	ThemeLine   = ProgressThemeLine
	ThemeCircle = ProgressThemeCircle
	ThemePlump  = ProgressThemePlump
)

// ProgressStatus controls the progress bar color.
type ProgressStatus uint8

const (
	ProgressNormal  ProgressStatus = iota
	ProgressSuccess
	ProgressError
	ProgressActive
	ProgressWarning
)

const progressHeight = float32(8)

// Progress displays a progress bar.
type Progress struct {
	Base
	percentage  float32 // 0-100
	status      ProgressStatus
	theme       ProgressTheme
	label       string  // custom label; if empty and showLabel, shows percentage
	showLabel   bool
	strokeWidth float32 // track/bar thickness
	size        float32 // circle diameter (default 120)
	color       string  // bar color override
	trackColor  string  // track background color override
}

// NewProgress creates a progress bar.
func NewProgress(tree *core.Tree, cfg *Config) *Progress {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	p := &Progress{
		Base:        NewBase(tree, core.TypeCustom, cfg),
		showLabel:   true,
		strokeWidth: 8,
		size:        120,
	}
	p.style.Display = layout.DisplayBlock
	p.style.Height = layout.Px(progressHeight)
	return p
}

func (p *Progress) Percentage() float32    { return p.percentage }
func (p *Progress) Status() ProgressStatus { return p.status }

func (p *Progress) SetPercentage(pct float32) {
	if pct < 0 {
		pct = 0
	}
	if pct > 100 {
		pct = 100
	}
	p.percentage = pct
}

func (p *Progress) SetStatus(s ProgressStatus) { p.status = s }
func (p *Progress) SetTheme(t ProgressTheme)   { p.theme = t }
func (p *Progress) SetLabel(l string)          { p.label = l }
func (p *Progress) SetShowLabel(b bool)        { p.showLabel = b }
func (p *Progress) SetStrokeWidth(w float32)   { p.strokeWidth = w }
func (p *Progress) SetSize(s float32)          { p.size = s }
func (p *Progress) SetColor(c string)          { p.color = c }
func (p *Progress) SetTrackColor(c string)     { p.trackColor = c }

// Deprecated: Use Percentage instead.
func (p *Progress) Percent() float32 { return p.percentage }

// Deprecated: Use SetPercentage instead.
func (p *Progress) SetPercent(pct float32) { p.SetPercentage(pct) }

// Deprecated: Use SetSize instead.
func (p *Progress) SetCircleSize(s float32) { p.size = s }

func (p *Progress) barColor() uimath.Color {
	if p.color != "" {
		return uimath.ColorHex(p.color)
	}
	switch p.status {
	case ProgressSuccess:
		return p.config.SuccessColor
	case ProgressError:
		return p.config.ErrorColor
	case ProgressWarning:
		return p.config.WarningColor
	default:
		return p.config.PrimaryColor
	}
}

func (p *Progress) bgColor() uimath.Color {
	if p.trackColor != "" {
		return uimath.ColorHex(p.trackColor)
	}
	return uimath.ColorHex("#f5f5f5")
}

func (p *Progress) labelText() string {
	if p.label != "" {
		return p.label
	}
	return strconv.Itoa(int(p.percentage)) + "%"
}

func (p *Progress) Draw(buf *render.CommandBuffer) {
	switch p.theme {
	case ThemeCircle:
		p.drawCircle(buf)
	case ThemePlump:
		p.drawPlump(buf)
	default:
		p.drawLine(buf)
	}
	p.DrawChildren(buf)
}

func (p *Progress) drawLine(buf *render.CommandBuffer) {
	bounds := p.Bounds()
	if bounds.IsEmpty() {
		return
	}

	cfg := p.config
	barH := p.strokeWidth
	radius := barH / 2

	// Reserve space for label on the right
	labelW := float32(0)
	if p.showLabel {
		text := p.labelText()
		labelW = float32(len(text))*cfg.FontSize*0.55 + 8
		if cfg.TextRenderer != nil {
			labelW = cfg.TextRenderer.MeasureText(text, cfg.FontSize) + 8
		}
	}
	trackW := bounds.Width - labelW

	trackY := bounds.Y + (bounds.Height-barH)/2

	// Track
	buf.DrawRect(render.RectCmd{
		Bounds:    uimath.NewRect(bounds.X, trackY, trackW, barH),
		FillColor: uimath.ColorHex("#f5f5f5"),
		Corners:   uimath.CornersAll(radius),
	}, 0, 1)

	// Fill
	if p.percentage > 0 {
		fillW := trackW * p.percentage / 100
		if fillW < barH {
			fillW = barH
		}
		buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(bounds.X, trackY, fillW, barH),
			FillColor: p.barColor(),
			Corners:   uimath.CornersAll(radius),
		}, 1, 1)
	}

	// Label at right end
	if p.showLabel {
		text := p.labelText()
		textX := bounds.X + trackW + 4
		textClr := p.barColor()
		if cfg.TextRenderer != nil {
			lh := cfg.TextRenderer.LineHeight(cfg.FontSize)
			cfg.TextRenderer.DrawText(buf, text, textX, bounds.Y+(bounds.Height-lh)/2, cfg.FontSize, labelW, textClr, 1)
		} else {
			tw := float32(len(text)) * cfg.FontSize * 0.55
			th := cfg.FontSize * 1.2
			buf.DrawRect(render.RectCmd{
				Bounds:    uimath.NewRect(textX, bounds.Y+(bounds.Height-th)/2, tw, th),
				FillColor: textClr,
				Corners:   uimath.CornersAll(2),
			}, 2, 1)
		}
	}
}

func (p *Progress) drawCircle(buf *render.CommandBuffer) {
	bounds := p.Bounds()
	if bounds.IsEmpty() {
		return
	}

	cfg := p.config
	diameter := p.size
	sw := p.strokeWidth
	if sw <= 0 {
		sw = 6
	}

	cx := bounds.X + (bounds.Width-diameter)/2
	cy := bounds.Y + (bounds.Height-diameter)/2

	// Track ring: a large rounded-rect circle with border only
	buf.DrawRect(render.RectCmd{
		Bounds:      uimath.NewRect(cx, cy, diameter, diameter),
		FillColor:   uimath.ColorHex("#f5f5f5"),
		BorderColor: uimath.ColorHex("#e8e8e8"),
		BorderWidth: sw,
		Corners:     uimath.CornersAll(diameter / 2),
	}, 0, 1)

	// Colored ring overlay (approximation: full colored border, clipped by opacity = percent/100)
	// Since we can't do arc clipping with DrawRect, we draw a full colored border circle
	// and rely on the visual to show progress state via the percentage label.
	if p.percentage > 0 {
		buf.DrawRect(render.RectCmd{
			Bounds:      uimath.NewRect(cx, cy, diameter, diameter),
			BorderColor: p.barColor(),
			BorderWidth: sw,
			Corners:     uimath.CornersAll(diameter / 2),
		}, 1, p.percentage/100)
	}

	// Percentage label in center
	if p.showLabel {
		text := p.labelText()
		fontSize := cfg.FontSize
		if diameter >= 80 {
			fontSize = cfg.FontSizeLg
		}
		textClr := p.barColor()
		if cfg.TextRenderer != nil {
			tw := cfg.TextRenderer.MeasureText(text, fontSize)
			lh := cfg.TextRenderer.LineHeight(fontSize)
			cfg.TextRenderer.DrawText(buf, text, cx+(diameter-tw)/2, cy+(diameter-lh)/2, fontSize, diameter, textClr, 1)
		} else {
			tw := float32(len(text)) * fontSize * 0.55
			th := fontSize * 1.2
			buf.DrawRect(render.RectCmd{
				Bounds:    uimath.NewRect(cx+(diameter-tw)/2, cy+(diameter-th)/2, tw, th),
				FillColor: textClr,
				Corners:   uimath.CornersAll(2),
			}, 2, 1)
		}
	}
}

func (p *Progress) drawPlump(buf *render.CommandBuffer) {
	bounds := p.Bounds()
	if bounds.IsEmpty() {
		return
	}

	cfg := p.config
	barH := float32(20)
	radius := barH / 2

	trackY := bounds.Y + (bounds.Height-barH)/2

	// Track
	buf.DrawRect(render.RectCmd{
		Bounds:    uimath.NewRect(bounds.X, trackY, bounds.Width, barH),
		FillColor: uimath.ColorHex("#f5f5f5"),
		Corners:   uimath.CornersAll(radius),
	}, 0, 1)

	// Fill
	fillW := bounds.Width * p.percentage / 100
	if p.percentage > 0 {
		if fillW < barH {
			fillW = barH
		}
		buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(bounds.X, trackY, fillW, barH),
			FillColor: p.barColor(),
			Corners:   uimath.CornersAll(radius),
		}, 1, 1)
	}

	// Label inside the bar
	if p.showLabel {
		text := p.labelText()
		fontSize := cfg.FontSizeSm
		textClr := uimath.ColorWhite
		if p.percentage < 10 {
			textClr = p.barColor() // show outside fill area if too narrow
		}
		if cfg.TextRenderer != nil {
			tw := cfg.TextRenderer.MeasureText(text, fontSize)
			lh := cfg.TextRenderer.LineHeight(fontSize)
			// Center text in the filled portion (or at end if filled is wide enough)
			textX := bounds.X + fillW - tw - 4
			if textX < bounds.X+4 {
				textX = bounds.X + fillW + 4
			}
			cfg.TextRenderer.DrawText(buf, text, textX, trackY+(barH-lh)/2, fontSize, bounds.Width, textClr, 1)
		} else {
			tw := float32(len(text)) * fontSize * 0.55
			th := fontSize * 1.2
			textX := bounds.X + fillW - tw - 4
			if textX < bounds.X+4 {
				textX = bounds.X + fillW + 4
			}
			buf.DrawRect(render.RectCmd{
				Bounds:    uimath.NewRect(textX, trackY+(barH-th)/2, tw, th),
				FillColor: textClr,
				Corners:   uimath.CornersAll(2),
			}, 2, 1)
		}
	}
}
