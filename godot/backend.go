package godot

import (
	"fmt"
	"image"
	"image/color"
	"sort"

	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/platform"
	"github.com/kasuganosora/ui/render"
)

// SoftwareBackend implements render.Backend using a CPU-side RGBA pixel buffer.
// It rasterizes UI commands to an in-memory image that can be read by the host
// (Godot, tests, or any embedding environment).
type SoftwareBackend struct {
	width    int // framebuffer width (physical pixels)
	height   int // framebuffer height (physical pixels)
	dpiScale float32
	pixels   []byte // RGBA, row-major, top-left origin, len = width*height*4

	// Scissor state (physical pixels)
	scissor uimath.Rect

	// Texture storage
	textures   map[render.TextureHandle]*softTexture
	nextHandle render.TextureHandle
}

type softTexture struct {
	width  int
	height int
	format render.TextureFormat
	data   []byte // RGBA (converted on store for R8/BGRA8)
}

// NewSoftwareBackend creates a software rasterizer backend.
func NewSoftwareBackend() *SoftwareBackend {
	return &SoftwareBackend{
		textures:   make(map[render.TextureHandle]*softTexture),
		nextHandle: 1,
		dpiScale:   1.0,
	}
}

func (b *SoftwareBackend) Init(window platform.Window) error {
	w, h := window.FramebufferSize()
	b.dpiScale = window.DPIScale()
	b.resize(w, h)
	return nil
}

func (b *SoftwareBackend) BeginFrame() {
	// Clear to transparent black
	for i := range b.pixels {
		b.pixels[i] = 0
	}
	b.scissor = uimath.NewRect(0, 0, float32(b.width), float32(b.height))
}

func (b *SoftwareBackend) EndFrame() {}

func (b *SoftwareBackend) Resize(width, height int) {
	b.resize(width, height)
}

func (b *SoftwareBackend) resize(w, h int) {
	if w <= 0 || h <= 0 {
		return
	}
	b.width = w
	b.height = h
	b.pixels = make([]byte, w*h*4)
	b.scissor = uimath.NewRect(0, 0, float32(w), float32(h))
}

func (b *SoftwareBackend) Submit(buf *render.CommandBuffer) {
	// Process main commands (already in submission order; sort by z-order)
	cmds := make([]render.Command, len(buf.Commands()))
	copy(cmds, buf.Commands())
	sort.SliceStable(cmds, func(i, j int) bool {
		return cmds[i].ZOrder < cmds[j].ZOrder
	})
	b.executeCommands(cmds)

	// Process overlays (on top, no clip)
	if overlays := buf.Overlays(); len(overlays) > 0 {
		b.scissor = uimath.NewRect(0, 0, float32(b.width), float32(b.height))
		ovr := make([]render.Command, len(overlays))
		copy(ovr, overlays)
		sort.SliceStable(ovr, func(i, j int) bool {
			return ovr[i].ZOrder < ovr[j].ZOrder
		})
		b.executeCommands(ovr)
	}
}

func (b *SoftwareBackend) executeCommands(cmds []render.Command) {
	dpi := b.dpiScale
	for _, cmd := range cmds {
		switch cmd.Type {
		case render.CmdClip:
			if cmd.Clip != nil {
				b.scissor = uimath.NewRect(
					cmd.Clip.Bounds.X*dpi,
					cmd.Clip.Bounds.Y*dpi,
					cmd.Clip.Bounds.Width*dpi,
					cmd.Clip.Bounds.Height*dpi,
				)
				// Clamp to framebuffer
				b.scissor = b.scissor.Intersection(
					uimath.NewRect(0, 0, float32(b.width), float32(b.height)),
				)
			}
		case render.CmdRect:
			if cmd.Rect != nil {
				b.drawRect(cmd.Rect, cmd.Opacity, dpi)
			}
		case render.CmdText:
			if cmd.Text != nil {
				b.drawText(cmd.Text, cmd.Opacity, dpi)
			}
		case render.CmdImage:
			if cmd.Image != nil {
				b.drawImage(cmd.Image, cmd.Opacity, dpi)
			}
		case render.CmdShadow:
			if cmd.Shadow != nil {
				b.drawShadow(cmd.Shadow, cmd.Opacity, dpi)
			}
		}
	}
}

// drawRect fills a rectangle with alpha blending.
func (b *SoftwareBackend) drawRect(rc *render.RectCmd, opacity float32, dpi float32) {
	bounds := rc.Bounds
	px := int(bounds.X * dpi)
	py := int(bounds.Y * dpi)
	pw := int(bounds.Width * dpi)
	ph := int(bounds.Height * dpi)

	fc := rc.FillColor
	if !fc.IsTransparent() {
		a := fc.A * opacity
		r8, g8, b8, _ := fc.RGBA8()
		b.fillRect(px, py, pw, ph, r8, g8, b8, a)
	}

	// Border
	if rc.BorderWidth > 0 && !rc.BorderColor.IsTransparent() {
		bw := int(rc.BorderWidth * dpi)
		if bw < 1 {
			bw = 1
		}
		a := rc.BorderColor.A * opacity
		r8, g8, b8, _ := rc.BorderColor.RGBA8()
		// Top
		b.fillRect(px, py, pw, bw, r8, g8, b8, a)
		// Bottom
		b.fillRect(px, py+ph-bw, pw, bw, r8, g8, b8, a)
		// Left
		b.fillRect(px, py+bw, bw, ph-bw*2, r8, g8, b8, a)
		// Right
		b.fillRect(px+pw-bw, py+bw, bw, ph-bw*2, r8, g8, b8, a)
	}
}

// drawText blits glyph atlas regions with the text color.
func (b *SoftwareBackend) drawText(tc *render.TextCmd, opacity float32, dpi float32) {
	tex := b.textures[tc.Atlas]
	if tex == nil {
		return
	}

	cr, cg, cb, _ := tc.Color.RGBA8()
	a := tc.Color.A * opacity

	for _, g := range tc.Glyphs {
		dx := int(g.X * dpi)
		dy := int(g.Y * dpi)
		dw := int(g.Width * dpi)
		dh := int(g.Height * dpi)

		// UV to pixel coordinates in atlas
		su0 := int(g.U0 * float32(tex.width))
		sv0 := int(g.V0 * float32(tex.height))
		su1 := int(g.U1 * float32(tex.width))
		sv1 := int(g.V1 * float32(tex.height))
		sw := su1 - su0
		sh := sv1 - sv0
		if sw <= 0 || sh <= 0 || dw <= 0 || dh <= 0 {
			continue
		}

		for py := 0; py < dh; py++ {
			// Map destination Y to source Y
			sy := sv0 + py*sh/dh
			if sy < 0 || sy >= tex.height {
				continue
			}
			for px := 0; px < dw; px++ {
				sx := su0 + px*sw/dw
				if sx < 0 || sx >= tex.width {
					continue
				}

				// Read atlas alpha (stored as RGBA, alpha is coverage)
				tidx := (sy*tex.width + sx) * 4
				if tidx+3 >= len(tex.data) {
					continue
				}
				coverage := float32(tex.data[tidx+3]) / 255.0
				if coverage <= 0 {
					continue
				}

				ga := a * coverage
				b.blendPixel(dx+px, dy+py, cr, cg, cb, ga)
			}
		}
	}
}

// drawImage blits a textured rectangle.
func (b *SoftwareBackend) drawImage(ic *render.ImageCmd, opacity float32, dpi float32) {
	tex := b.textures[ic.Texture]
	if tex == nil {
		return
	}

	dst := ic.DstRect
	dx := int(dst.X * dpi)
	dy := int(dst.Y * dpi)
	dw := int(dst.Width * dpi)
	dh := int(dst.Height * dpi)
	if dw <= 0 || dh <= 0 {
		return
	}

	src := ic.SrcRect
	tint := ic.Tint
	if tint.A <= 0 {
		tint = uimath.NewColor(1, 1, 1, 1) // default white tint
	}

	for py := 0; py < dh; py++ {
		// Map to source texture
		ty := src.Y + float32(py)/float32(dh)*src.Height
		sy := int(ty * float32(tex.height))
		if sy < 0 || sy >= tex.height {
			continue
		}
		for px := 0; px < dw; px++ {
			tx := src.X + float32(px)/float32(dw)*src.Width
			sx := int(tx * float32(tex.width))
			if sx < 0 || sx >= tex.width {
				continue
			}

			tidx := (sy*tex.width + sx) * 4
			if tidx+3 >= len(tex.data) {
				continue
			}
			sr := float32(tex.data[tidx]) / 255.0
			sg := float32(tex.data[tidx+1]) / 255.0
			sb := float32(tex.data[tidx+2]) / 255.0
			sa := float32(tex.data[tidx+3]) / 255.0

			// Apply tint
			fr := sr * tint.R
			fg := sg * tint.G
			fb := sb * tint.B
			fa := sa * tint.A * opacity

			r8 := uint8(clampf(fr, 0, 1) * 255)
			g8 := uint8(clampf(fg, 0, 1) * 255)
			b8 := uint8(clampf(fb, 0, 1) * 255)
			b.blendPixel(dx+px, dy+py, r8, g8, b8, fa)
		}
	}
}

// drawShadow renders a simplified box shadow (solid color, no blur for v1).
func (b *SoftwareBackend) drawShadow(sc *render.ShadowCmd, opacity float32, dpi float32) {
	// Expand bounds by spread + offset
	sx := sc.Bounds.X + sc.OffsetX - sc.SpreadRadius
	sy := sc.Bounds.Y + sc.OffsetY - sc.SpreadRadius
	sw := sc.Bounds.Width + sc.SpreadRadius*2
	sh := sc.Bounds.Height + sc.SpreadRadius*2

	px := int(sx * dpi)
	py := int(sy * dpi)
	pw := int(sw * dpi)
	ph := int(sh * dpi)

	// Use shadow color with reduced alpha (approximation of blur)
	a := sc.Color.A * opacity * 0.5
	r8, g8, b8, _ := sc.Color.RGBA8()
	b.fillRect(px, py, pw, ph, r8, g8, b8, a)
}

// fillRect fills a solid rectangle with alpha blending, respecting the scissor.
func (b *SoftwareBackend) fillRect(x, y, w, h int, r, g, bv uint8, alpha float32) {
	if alpha <= 0 {
		return
	}

	// Clip to scissor
	sx0 := int(b.scissor.X)
	sy0 := int(b.scissor.Y)
	sx1 := int(b.scissor.X + b.scissor.Width)
	sy1 := int(b.scissor.Y + b.scissor.Height)

	x0 := max(x, sx0)
	y0 := max(y, sy0)
	x1 := min(x+w, sx1)
	y1 := min(y+h, sy1)

	// Clip to framebuffer
	x0 = max(x0, 0)
	y0 = max(y0, 0)
	x1 = min(x1, b.width)
	y1 = min(y1, b.height)

	if x0 >= x1 || y0 >= y1 {
		return
	}

	if alpha >= 1.0 {
		// Opaque fast path
		for py := y0; py < y1; py++ {
			row := py * b.width * 4
			for px := x0; px < x1; px++ {
				idx := row + px*4
				b.pixels[idx] = r
				b.pixels[idx+1] = g
				b.pixels[idx+2] = bv
				b.pixels[idx+3] = 255
			}
		}
	} else {
		a16 := uint16(alpha * 255)
		ia16 := 255 - a16
		pr := uint16(r)
		pg := uint16(g)
		pb := uint16(bv)
		for py := y0; py < y1; py++ {
			row := py * b.width * 4
			for px := x0; px < x1; px++ {
				idx := row + px*4
				dr := uint16(b.pixels[idx])
				dg := uint16(b.pixels[idx+1])
				db := uint16(b.pixels[idx+2])
				da := uint16(b.pixels[idx+3])
				b.pixels[idx] = uint8((pr*a16 + dr*ia16) / 255)
				b.pixels[idx+1] = uint8((pg*a16 + dg*ia16) / 255)
				b.pixels[idx+2] = uint8((pb*a16 + db*ia16) / 255)
				na := da + a16 - da*a16/255
				if na > 255 {
					na = 255
				}
				b.pixels[idx+3] = uint8(na)
			}
		}
	}
}

// blendPixel blends a single pixel with alpha blending, respecting the scissor.
func (b *SoftwareBackend) blendPixel(x, y int, r, g, bv uint8, alpha float32) {
	if alpha <= 0 {
		return
	}
	// Scissor check
	fx, fy := float32(x), float32(y)
	if fx < b.scissor.X || fx >= b.scissor.X+b.scissor.Width ||
		fy < b.scissor.Y || fy >= b.scissor.Y+b.scissor.Height {
		return
	}
	if x < 0 || x >= b.width || y < 0 || y >= b.height {
		return
	}

	idx := (y*b.width + x) * 4
	if alpha >= 1.0 {
		b.pixels[idx] = r
		b.pixels[idx+1] = g
		b.pixels[idx+2] = bv
		b.pixels[idx+3] = 255
		return
	}

	a16 := uint16(alpha * 255)
	ia16 := 255 - a16
	dr := uint16(b.pixels[idx])
	dg := uint16(b.pixels[idx+1])
	db := uint16(b.pixels[idx+2])
	da := uint16(b.pixels[idx+3])
	b.pixels[idx] = uint8((uint16(r)*a16 + dr*ia16) / 255)
	b.pixels[idx+1] = uint8((uint16(g)*a16 + dg*ia16) / 255)
	b.pixels[idx+2] = uint8((uint16(bv)*a16 + db*ia16) / 255)
	na := da + a16 - da*a16/255
	if na > 255 {
		na = 255
	}
	b.pixels[idx+3] = uint8(na)
}

// --- Texture management ---

func (b *SoftwareBackend) CreateTexture(desc render.TextureDesc) (render.TextureHandle, error) {
	if desc.Width <= 0 || desc.Height <= 0 {
		return render.InvalidTexture, fmt.Errorf("invalid texture size %dx%d", desc.Width, desc.Height)
	}

	handle := b.nextHandle
	b.nextHandle++

	rgba := convertToRGBA(desc.Data, desc.Width, desc.Height, desc.Format)

	b.textures[handle] = &softTexture{
		width:  desc.Width,
		height: desc.Height,
		format: desc.Format,
		data:   rgba,
	}
	return handle, nil
}

func (b *SoftwareBackend) UpdateTexture(handle render.TextureHandle, region uimath.Rect, data []byte) error {
	tex := b.textures[handle]
	if tex == nil {
		return fmt.Errorf("texture %d not found", handle)
	}

	rx := int(region.X)
	ry := int(region.Y)
	rw := int(region.Width)
	rh := int(region.Height)

	// Convert incoming data to RGBA
	rgba := convertToRGBA(data, rw, rh, tex.format)

	// Copy into texture
	for y := 0; y < rh; y++ {
		ty := ry + y
		if ty < 0 || ty >= tex.height {
			continue
		}
		for x := 0; x < rw; x++ {
			tx := rx + x
			if tx < 0 || tx >= tex.width {
				continue
			}
			si := (y*rw + x) * 4
			di := (ty*tex.width + tx) * 4
			if si+3 < len(rgba) && di+3 < len(tex.data) {
				copy(tex.data[di:di+4], rgba[si:si+4])
			}
		}
	}
	return nil
}

func (b *SoftwareBackend) DestroyTexture(handle render.TextureHandle) {
	delete(b.textures, handle)
}

func (b *SoftwareBackend) MaxTextureSize() int { return 4096 }
func (b *SoftwareBackend) DPIScale() float32   { return b.dpiScale }

// ReadPixels returns the current framebuffer as an RGBA image.
func (b *SoftwareBackend) ReadPixels() (*image.RGBA, error) {
	img := image.NewRGBA(image.Rect(0, 0, b.width, b.height))
	copy(img.Pix, b.pixels)
	return img, nil
}

func (b *SoftwareBackend) Destroy() {
	b.pixels = nil
	b.textures = nil
}

// Pixels returns a direct reference to the RGBA pixel buffer.
// Length = width * height * 4 bytes (RGBA, row-major, top-left origin).
// The slice is only valid until the next BeginFrame or Resize.
func (b *SoftwareBackend) Pixels() []byte {
	return b.pixels
}

// FramebufferSize returns the current framebuffer dimensions in physical pixels.
func (b *SoftwareBackend) FramebufferSize() (int, int) {
	return b.width, b.height
}

// --- Helpers ---

// convertToRGBA converts pixel data from any supported format to RGBA.
func convertToRGBA(data []byte, w, h int, format render.TextureFormat) []byte {
	n := w * h
	rgba := make([]byte, n*4)

	switch format {
	case render.TextureFormatRGBA8:
		if len(data) >= n*4 {
			copy(rgba, data[:n*4])
		}
	case render.TextureFormatBGRA8:
		if len(data) >= n*4 {
			for i := 0; i < n; i++ {
				si := i * 4
				rgba[si] = data[si+2]   // R ← B
				rgba[si+1] = data[si+1] // G
				rgba[si+2] = data[si]   // B ← R
				rgba[si+3] = data[si+3] // A
			}
		}
	case render.TextureFormatR8:
		// Single channel → replicate to RGBA (alpha = coverage)
		if len(data) >= n {
			for i := 0; i < n; i++ {
				v := data[i]
				rgba[i*4] = 255
				rgba[i*4+1] = 255
				rgba[i*4+2] = 255
				rgba[i*4+3] = v
			}
		}
	}

	return rgba
}

func clampf(v, lo, hi float32) float32 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// Ensure SoftwareBackend and HeadlessWindow implement their interfaces.
var _ render.Backend = (*SoftwareBackend)(nil)
var _ platform.Window = (*HeadlessWindow)(nil)

// Ensure HeadlessPlatform satisfies platform.Platform at compile time.
func init() {
	var _ platform.Platform = (*HeadlessPlatform)(nil)
}

// pixelAt returns the RGBA color at the given framebuffer coordinates.
func (b *SoftwareBackend) pixelAt(x, y int) color.RGBA {
	if x < 0 || x >= b.width || y < 0 || y >= b.height {
		return color.RGBA{}
	}
	idx := (y*b.width + x) * 4
	return color.RGBA{
		R: b.pixels[idx],
		G: b.pixels[idx+1],
		B: b.pixels[idx+2],
		A: b.pixels[idx+3],
	}
}
