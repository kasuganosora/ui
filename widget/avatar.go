package widget

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"math"
	"os"
	"strconv"
	"time"
	"unicode"

	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/layout"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// AvatarShape controls the avatar shape.
type AvatarShape uint8

const (
	AvatarCircle AvatarShape = iota
	AvatarRound              // rounded square
	AvatarSquare
)

// AvatarLoadState represents the loading state of an avatar image.
type AvatarLoadState uint8

const (
	AvatarIdle    AvatarLoadState = iota // No src set
	AvatarLoading                        // Loading in progress
	AvatarLoaded                         // Successfully loaded
	AvatarError                          // Load failed
)

// Avatar displays a user avatar (image, icon, or initials).
// Supports remote image loading via src URL with GIF animation.
type Avatar struct {
	Base
	content          string
	alt              string
	src              string               // remote image URL
	icon             render.TextureHandle
	image            render.TextureHandle // directly-set image texture (caller owns)
	imagePath        string               // image URL/path (for reference)
	bgColor          uimath.Color
	size             Size    // SizeSmall=32, SizeMedium=48, SizeLarge=64
	customSize       float32 // if > 0, overrides size enum
	shape            AvatarShape
	hideOnLoadFailed bool
	onError          func()

	// Cached display character (avoids string alloc in Draw)
	displayChar string

	// Remote image loading
	loadState  AvatarLoadState
	srcTexture render.TextureHandle // loaded from src (we own)
	naturalW   int
	naturalH   int

	// GIF animation
	frames     []gifFrame
	frameIdx   int
	loopCount  int
	loopsDone  int
	lastUpdate time.Time
	playing    bool
}

// avatarSizePx returns the pixel dimension for a given avatar size.
func avatarSizePx(s Size) float32 {
	switch s {
	case SizeSmall:
		return 32
	case SizeLarge:
		return 64
	default: // SizeMedium
		return 48
	}
}

func NewAvatar(tree *core.Tree, cfg *Config) *Avatar {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	a := &Avatar{
		Base:       NewBase(tree, core.TypeCustom, cfg),
		bgColor:    uimath.ColorHex("#c6c6c6"),
		size:       SizeMedium,
		shape:      AvatarCircle,
		srcTexture: render.InvalidTexture,
	}
	px := avatarSizePx(a.size)
	a.style.Width = layout.Px(px)
	a.style.Height = layout.Px(px)
	a.style.FlexShrink = 0
	a.style.FlexGrow = 0
	a.style.Display = layout.DisplayBlock
	return a
}

func (a *Avatar) SetContent(t string) {
	a.content = t
	a.displayChar = "" // invalidate cache
}
func (a *Avatar) SetAlt(alt string) {
	a.alt = alt
	a.displayChar = "" // invalidate cache
}
func (a *Avatar) SetIcon(h render.TextureHandle) { a.icon = h }
func (a *Avatar) SetImageTexture(h render.TextureHandle) { a.image = h }
func (a *Avatar) SetImagePath(path string)       { a.imagePath = path }
func (a *Avatar) SetBgColor(c uimath.Color)      { a.bgColor = c }
func (a *Avatar) SetHideOnLoadFailed(v bool)     { a.hideOnLoadFailed = v }
func (a *Avatar) OnError(fn func())              { a.onError = fn }
func (a *Avatar) ImagePath() string              { return a.imagePath }
func (a *Avatar) Content() string                { return a.content }
func (a *Avatar) Shape() AvatarShape             { return a.shape }
func (a *Avatar) Src() string                    { return a.src }
func (a *Avatar) LoadState() AvatarLoadState     { return a.loadState }

// SetSize sets one of the preset sizes (Small=32, Medium=48, Large=64).
func (a *Avatar) SetSize(s Size) {
	a.size = s
	a.customSize = 0
	px := avatarSizePx(s)
	a.style.Width = layout.Px(px)
	a.style.Height = layout.Px(px)
	a.style.FlexShrink = 0
	a.style.FlexGrow = 0
}

// SetCustomSize sets a custom pixel size, overriding the preset.
func (a *Avatar) SetCustomSize(px float32) {
	a.customSize = px
	a.style.Width = layout.Px(px)
	a.style.Height = layout.Px(px)
	a.style.FlexShrink = 0
	a.style.FlexGrow = 0
}

// SetShape sets the avatar shape (Circle, Round, Square).
func (a *Avatar) SetShape(s AvatarShape) { a.shape = s }

// AvatarSize returns the effective pixel size.
func (a *Avatar) AvatarSize() float32 {
	if a.customSize > 0 {
		return a.customSize
	}
	return avatarSizePx(a.size)
}

// Deprecated: Use SetContent instead.
func (a *Avatar) SetText(t string) { a.content = t }

// Deprecated: Use SetImagePath instead.
func (a *Avatar) SetImage(path string) { a.imagePath = path }

// Deprecated: Use ImagePath instead.
func (a *Avatar) Image() string { return a.imagePath }

// SetSrc sets the remote image URL and triggers async loading.
// Supports http/https URLs (via NetClient), data: URLs, and local file paths.
// GIF images are auto-detected and played as animations.
func (a *Avatar) SetSrc(src string) {
	if src == a.src && a.loadState == AvatarLoaded {
		return
	}
	a.src = src
	a.destroySrcTexture()
	a.frames = nil
	a.frameIdx = 0
	a.loopsDone = 0
	a.playing = false

	if src == "" {
		a.loadState = AvatarIdle
		return
	}

	a.loadState = AvatarLoading
	a.tree.MarkDirty(a.id)

	if isRemoteURL(src) {
		a.loadFromURL(src)
	} else {
		a.loadFromFile(src)
	}
}

func (a *Avatar) loadFromURL(src string) {
	nc := a.config.NetClient
	if nc == nil {
		a.loadState = AvatarError
		if a.onError != nil {
			a.onError()
		}
		return
	}
	nc.FetchAsync(src, func(data []byte, err error) {
		if err != nil {
			a.loadState = AvatarError
			if a.onError != nil {
				a.onError()
			}
			a.tree.MarkDirty(a.id)
			return
		}
		a.loadFromBytes(data, src)
		a.tree.MarkDirty(a.id)
	})
}

func (a *Avatar) loadFromFile(path string) {
	f, err := os.Open(path)
	if err != nil {
		a.loadState = AvatarError
		if a.onError != nil {
			a.onError()
		}
		return
	}
	defer f.Close()

	if isGIF(path) {
		a.loadGIFFromReader(f)
		return
	}

	decoded, _, err := image.Decode(f)
	if err != nil {
		a.loadState = AvatarError
		if a.onError != nil {
			a.onError()
		}
		return
	}
	bounds := decoded.Bounds()
	rgba := image.NewRGBA(bounds)
	draw.Draw(rgba, bounds, decoded, bounds.Min, draw.Src)
	a.naturalW = bounds.Dx()
	a.naturalH = bounds.Dy()
	a.createSrcTexture(rgba.Pix, a.naturalW, a.naturalH)
}

func (a *Avatar) loadFromBytes(data []byte, hint string) {
	if isGIF(hint) || looksLikeGIF(data) {
		a.loadGIFFromReader(bytes.NewReader(data))
		return
	}
	decoded, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		a.loadState = AvatarError
		if a.onError != nil {
			a.onError()
		}
		return
	}
	bounds := decoded.Bounds()
	rgba := image.NewRGBA(bounds)
	draw.Draw(rgba, bounds, decoded, bounds.Min, draw.Src)
	a.naturalW = bounds.Dx()
	a.naturalH = bounds.Dy()
	a.createSrcTexture(rgba.Pix, a.naturalW, a.naturalH)
}

func (a *Avatar) loadGIFFromReader(r io.Reader) {
	g, err := gif.DecodeAll(r)
	if err != nil {
		a.loadState = AvatarError
		if a.onError != nil {
			a.onError()
		}
		return
	}
	if len(g.Image) == 0 {
		a.loadState = AvatarError
		if a.onError != nil {
			a.onError()
		}
		return
	}

	canvasW := g.Config.Width
	canvasH := g.Config.Height
	if canvasW == 0 || canvasH == 0 {
		canvasW = g.Image[0].Bounds().Dx()
		canvasH = g.Image[0].Bounds().Dy()
	}
	a.naturalW = canvasW
	a.naturalH = canvasH
	a.loopCount = g.LoopCount

	canvas := image.NewRGBA(image.Rect(0, 0, canvasW, canvasH))
	var prevCanvas *image.RGBA

	a.frames = make([]gifFrame, len(g.Image))
	for i, frame := range g.Image {
		disposal := byte(gif.DisposalNone)
		if i < len(g.Disposal) {
			disposal = g.Disposal[i]
		}
		if disposal == gif.DisposalPrevious {
			prevCanvas = cloneRGBA(canvas)
		}
		draw.Draw(canvas, frame.Bounds(), frame, frame.Bounds().Min, draw.Over)

		pixels := make([]byte, len(canvas.Pix))
		copy(pixels, canvas.Pix)

		delay := time.Duration(100) * time.Millisecond
		if i < len(g.Delay) && g.Delay[i] > 0 {
			delay = time.Duration(g.Delay[i]) * 10 * time.Millisecond
		}
		a.frames[i] = gifFrame{pixels: pixels, delay: delay}

		switch disposal {
		case gif.DisposalBackground:
			bg := image.NewUniform(color.Transparent)
			draw.Draw(canvas, frame.Bounds(), bg, image.Point{}, draw.Src)
		case gif.DisposalPrevious:
			if prevCanvas != nil {
				copy(canvas.Pix, prevCanvas.Pix)
			}
		}
	}

	a.createSrcTexture(a.frames[0].pixels, canvasW, canvasH)
	if len(a.frames) > 1 {
		a.playing = true
		a.lastUpdate = time.Now()
	}
}

func (a *Avatar) createSrcTexture(pixels []byte, w, h int) {
	backend := a.config.Backend
	if backend == nil {
		a.loadState = AvatarLoaded
		return
	}
	tex, err := backend.CreateTexture(render.TextureDesc{
		Width:  w,
		Height: h,
		Format: render.TextureFormatRGBA8,
		Filter: render.TextureFilterLinear,
		Data:   pixels,
	})
	if err != nil {
		a.loadState = AvatarError
		if a.onError != nil {
			a.onError()
		}
		return
	}
	a.srcTexture = tex
	a.loadState = AvatarLoaded
}

func (a *Avatar) uploadFrame(idx int) {
	if idx < 0 || idx >= len(a.frames) || a.srcTexture == render.InvalidTexture {
		return
	}
	backend := a.config.Backend
	if backend == nil {
		return
	}
	region := uimath.NewRect(0, 0, float32(a.naturalW), float32(a.naturalH))
	backend.UpdateTexture(a.srcTexture, region, a.frames[idx].pixels)
}

func (a *Avatar) destroySrcTexture() {
	if a.srcTexture != render.InvalidTexture && a.config.Backend != nil {
		a.config.Backend.DestroyTexture(a.srcTexture)
		a.srcTexture = render.InvalidTexture
	}
}

// Destroy cleans up GPU resources.
func (a *Avatar) Destroy() {
	a.destroySrcTexture()
	a.Base.Destroy()
}

// firstNonSymbolRune returns the first letter or digit from s.
// Falls back to the first rune if no letter/digit found.
func firstNonSymbolRune(s string) string {
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return string(r)
		}
	}
	for _, r := range s {
		return string(r)
	}
	return ""
}

func (a *Avatar) Draw(buf *render.CommandBuffer) {
	px := a.AvatarSize()
	bounds := a.Bounds()
	if bounds.IsEmpty() {
		bounds = uimath.NewRect(0, 0, px, px)
	}
	// Force square bounds — use the configured size, centered within layout bounds.
	if bounds.Width != px || bounds.Height != px {
		cx := bounds.X + bounds.Width/2
		cy := bounds.Y + bounds.Height/2
		bounds = uimath.NewRect(cx-px/2, cy-px/2, px, px)
	}

	fontSize := px * 0.45
	if a.customSize > 0 && a.customSize > 64 {
		fontSize = a.customSize * 0.4
	}

	cfg := a.config
	radius := float32(0)
	switch a.shape {
	case AvatarCircle:
		radius = bounds.Width / 2
	case AvatarRound:
		radius = cfg.BorderRadius
	default: // AvatarSquare
		radius = 0
	}

	// Advance GIF animation
	if a.playing && len(a.frames) > 1 {
		now := time.Now()
		elapsed := now.Sub(a.lastUpdate)
		frame := a.frames[a.frameIdx]
		if elapsed >= frame.delay {
			a.lastUpdate = now
			nextIdx := a.frameIdx + 1
			if nextIdx >= len(a.frames) {
				if a.loopCount == 0 {
					nextIdx = 0
				} else {
					a.loopsDone++
					if a.loopsDone >= a.loopCount {
						a.playing = false
						nextIdx = a.frameIdx
					} else {
						nextIdx = 0
					}
				}
			}
			if nextIdx != a.frameIdx {
				a.frameIdx = nextIdx
				a.uploadFrame(nextIdx)
			}
		}
		if a.playing {
			a.tree.MarkDirty(a.id)
		}
	}

	// Background
	buf.DrawRect(render.RectCmd{
		Bounds:    bounds,
		FillColor: a.bgColor,
		Corners:   uimath.CornersAll(radius),
	}, 0, 1)

	// Determine which texture to render
	tex := a.srcTexture
	if tex == render.InvalidTexture {
		tex = a.image // fallback to directly-set texture
	}

	if tex != render.InvalidTexture {
		// Draw loaded/direct image
		buf.DrawImage(render.ImageCmd{
			Texture: tex,
			DstRect: bounds,
			Tint:    uimath.ColorWhite,
			Corners: uimath.CornersAll(radius),
		}, 1, 1)
	} else if a.loadState == AvatarLoading {
		// Draw loading spinner
		a.drawLoadingSpinner(buf, bounds)
	} else if a.icon != 0 {
		// Icon centered within bounds
		iconSize := px * 0.5
		iconBounds := uimath.NewRect(
			bounds.X+(bounds.Width-iconSize)/2,
			bounds.Y+(bounds.Height-iconSize)/2,
			iconSize, iconSize,
		)
		buf.DrawImage(render.ImageCmd{
			Texture: a.icon,
			DstRect: iconBounds,
			Tint:    uimath.ColorWhite,
		}, 1, 1)
	} else {
		// Text content: extract first non-symbol character
		// Use cached displayChar to avoid string allocation per frame.
		if a.displayChar == "" {
			if a.content != "" {
				a.displayChar = firstNonSymbolRune(a.content)
			}
			if a.displayChar == "" && a.alt != "" {
				a.displayChar = firstNonSymbolRune(a.alt)
			}
		}
		displayChar := a.displayChar

		if displayChar != "" {
			if cfg.TextRenderer != nil {
				lh := cfg.TextRenderer.LineHeight(fontSize)
				tw := cfg.TextRenderer.MeasureText(displayChar, fontSize)
				cfg.TextRenderer.DrawText(buf, displayChar,
					bounds.X+(bounds.Width-tw)/2,
					bounds.Y+(bounds.Height-lh)/2,
					fontSize, bounds.Width, uimath.ColorWhite, 1)
			} else {
				tw := fontSize * 0.6
				th := fontSize
				buf.DrawRect(render.RectCmd{
					Bounds:    uimath.NewRect(bounds.X+(bounds.Width-tw)/2, bounds.Y+(bounds.Height-th)/2, tw, th),
					FillColor: uimath.ColorWhite,
					Corners:   uimath.CornersAll(2),
				}, 1, 1)
			}
		} else {
			// Default: draw a user icon placeholder (simple silhouette)
			a.drawDefaultIcon(buf, bounds, radius)
		}
	}
}

// drawLoadingSpinner draws a rotating dot spinner animation.
func (a *Avatar) drawLoadingSpinner(buf *render.CommandBuffer, bounds uimath.Rect) {
	cx := bounds.X + bounds.Width/2
	cy := bounds.Y + bounds.Height/2
	dotCount := 8
	dotR := bounds.Width * 0.05
	if dotR < 1.5 {
		dotR = 1.5
	}
	ringR := bounds.Width * 0.28

	// Phase rotates over time (one full rotation per 1.2s)
	phase := float64(time.Now().UnixMilli()%1200) / 1200.0 * 2 * math.Pi

	for i := 0; i < dotCount; i++ {
		angle := float64(i)*2*math.Pi/float64(dotCount) - math.Pi/2 + phase
		dx := float32(math.Cos(angle)) * ringR
		dy := float32(math.Sin(angle)) * ringR

		// Trailing opacity: first dot is brightest, last is dimmest
		opacity := 1.0 - float32(i)/float32(dotCount)*0.7

		buf.DrawRect(render.RectCmd{
			Bounds:    uimath.NewRect(cx+dx-dotR, cy+dy-dotR, dotR*2, dotR*2),
			FillColor: uimath.ColorWhite,
			Corners:   uimath.CornersAll(dotR),
		}, 2, opacity)
	}
	a.tree.MarkDirty(a.id) // continue animation
}

// drawDefaultIcon draws a simple user silhouette placeholder.
func (a *Avatar) drawDefaultIcon(buf *render.CommandBuffer, bounds uimath.Rect, radius float32) {
	cfg := a.config
	iconSize := bounds.Width * 0.6
	ix := bounds.X + (bounds.Width-iconSize)/2
	iy := bounds.Y + (bounds.Height-iconSize)/2
	if cfg.DrawMDIcon(buf, "person", ix, iy, iconSize, uimath.ColorWhite, 1, 1) {
		return
	}
	// Fallback: manual silhouette
	cx := bounds.X + bounds.Width/2
	cy := bounds.Y + bounds.Height/2
	px := bounds.Width

	headR := px * 0.18
	buf.DrawRect(render.RectCmd{
		Bounds:    uimath.NewRect(cx-headR, cy-px*0.18-headR, headR*2, headR*2),
		FillColor: uimath.ColorWhite,
		Corners:   uimath.CornersAll(headR),
	}, 1, 1)

	bodyW := px * 0.38
	bodyH := px * 0.22
	bodyY := cy + px*0.05
	buf.DrawRect(render.RectCmd{
		Bounds:    uimath.NewRect(cx-bodyW/2, bodyY, bodyW, bodyH),
		FillColor: uimath.ColorWhite,
		Corners:   uimath.CornersAll(bodyW / 2),
	}, 1, 1)
}

// ── AvatarGroup ─────────────────────────────────────────────────────────────

// CascadingValue controls the stacking direction of avatars.
type CascadingValue uint8

const (
	CascadingRightUp CascadingValue = iota // right avatar on top (default)
	CascadingLeftUp                        // left avatar on top
)

// AvatarGroup holds multiple Avatars and draws them overlapping.
type AvatarGroup struct {
	Base
	avatars        []*Avatar
	max            int            // max visible avatars (0 = show all)
	size           Size           // size applied to all child avatars
	cascading      CascadingValue // stacking direction
	collapseAvatar string         // custom text for the "+N" element
}

func NewAvatarGroup(tree *core.Tree, cfg *Config) *AvatarGroup {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	return &AvatarGroup{
		Base: NewBase(tree, core.TypeCustom, cfg),
		size: SizeMedium,
	}
}

func (g *AvatarGroup) SetMax(m int)                  { g.max = m }
func (g *AvatarGroup) SetSize(s Size)                { g.size = s }
func (g *AvatarGroup) SetCascading(c CascadingValue) { g.cascading = c }
func (g *AvatarGroup) SetCollapseAvatar(text string) { g.collapseAvatar = text }

// Deprecated: Use SetMax instead.
func (g *AvatarGroup) SetMaxCount(m int) { g.max = m }

// Deprecated: spacing is now auto-calculated from size.
func (g *AvatarGroup) SetSpacing(s float32) {}

// AddAvatar appends an avatar to the group.
func (g *AvatarGroup) AddAvatar(a *Avatar) {
	a.SetSize(g.size)
	g.avatars = append(g.avatars, a)
}

// Avatars returns the group's avatar slice.
func (g *AvatarGroup) Avatars() []*Avatar { return g.avatars }

func (g *AvatarGroup) Draw(buf *render.CommandBuffer) {
	bounds := g.Bounds()
	if bounds.IsEmpty() {
		return
	}

	px := avatarSizePx(g.size)
	// Overlap is ~25% of avatar size
	overlap := px * 0.25
	step := px - overlap

	visible := g.avatars
	hasMore := false
	if g.max > 0 && len(visible) > g.max {
		visible = visible[:g.max]
		hasMore = true
	}

	x := bounds.X
	for i, av := range visible {
		avBounds := uimath.NewRect(x, bounds.Y, px, px)
		lo := av.tree.Get(av.id).Layout()
		lo.Bounds = avBounds
		av.tree.SetLayout(av.id, lo)

		// Z-order based on cascading direction
		var z int32
		if g.cascading == CascadingLeftUp {
			z = int32(len(visible) - i)
		} else {
			z = int32(i)
		}

		// Draw border ring (white outline for overlap visibility)
		borderBounds := uimath.NewRect(avBounds.X-1, avBounds.Y-1, px+2, px+2)
		buf.DrawRect(render.RectCmd{
			Bounds:    borderBounds,
			FillColor: uimath.ColorWhite,
			Corners:   uimath.CornersAll((px + 2) / 2),
		}, z, 1)

		av.Draw(buf)
		x += step
	}

	// "+N" collapse indicator
	if hasMore {
		remaining := len(g.avatars) - g.max
		cfg := g.config
		extraBounds := uimath.NewRect(x, bounds.Y, px, px)

		// White border
		borderBounds := uimath.NewRect(extraBounds.X-1, extraBounds.Y-1, px+2, px+2)
		buf.DrawRect(render.RectCmd{
			Bounds:    borderBounds,
			FillColor: uimath.ColorWhite,
			Corners:   uimath.CornersAll((px + 2) / 2),
		}, int32(len(visible)), 1)

		buf.DrawRect(render.RectCmd{
			Bounds:    extraBounds,
			FillColor: uimath.ColorHex("#e8e8e8"),
			Corners:   uimath.CornersAll(px / 2),
		}, int32(len(visible)), 1)

		text := g.collapseAvatar
		if text == "" {
			text = "+" + strconv.Itoa(remaining)
		}
		if cfg.TextRenderer != nil {
			fs := px * 0.38
			lh := cfg.TextRenderer.LineHeight(fs)
			tw := cfg.TextRenderer.MeasureText(text, fs)
			cfg.TextRenderer.DrawText(buf, text,
				extraBounds.X+(px-tw)/2,
				extraBounds.Y+(px-lh)/2,
				fs, px, uimath.ColorHex("#8b8b8b"), 1)
		}
	}
}
