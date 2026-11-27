package widget

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/kasuganosora/ui/core"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// typographyProbe implements every optional text capability so a test can
// tell exactly which path emitText / emitCaptionText chose.
type typographyProbe struct {
	calls []typographyCall
}

type typographyCall struct {
	kind     string // plain, z, tracked, caption
	text     string
	size     float32
	tracking float32
	z        int32
}

func (p *typographyProbe) record(kind, text string, size, tracking float32, z int32) {
	p.calls = append(p.calls, typographyCall{kind: kind, text: text, size: size, tracking: tracking, z: z})
}

func (p *typographyProbe) DrawText(_ *render.CommandBuffer, text string, _, _, fontSize, _ float32, _ uimath.Color, _ float32) {
	p.record("plain", text, fontSize, 0, 0)
}
func (p *typographyProbe) DrawTextZ(_ *render.CommandBuffer, text string, _, _, fontSize, _ float32, _ uimath.Color, _ float32, z int32) {
	p.record("z", text, fontSize, 0, z)
}
func (p *typographyProbe) DrawTextTracked(_ *render.CommandBuffer, text string, _, _, fontSize, _ float32, _ uimath.Color, _ float32, z int32, tracking float32) {
	p.record("tracked", text, fontSize, tracking, z)
}
func (p *typographyProbe) DrawCaptionText(_ *render.CommandBuffer, text string, _, _, fontSize, _ float32, _ uimath.Color, _ float32, z int32, tracking float32) {
	p.record("caption", text, fontSize, tracking, z)
}
func (p *typographyProbe) LineHeight(size float32) float32 { return size * 1.2 }
func (p *typographyProbe) MeasureText(text string, size float32) float32 {
	return float32(len([]rune(text))) * size * 0.6
}

func (p *typographyProbe) last(t *testing.T) typographyCall {
	t.Helper()
	if len(p.calls) == 0 {
		t.Fatal("text drawer was never called")
	}
	return p.calls[len(p.calls)-1]
}

func approx(a, b, eps float32) bool {
	d := a - b
	if d < 0 {
		d = -d
	}
	return d <= eps
}

// §15: leading must change with size, anchored at the existing 14px body
// value so ordinary layout stays pixel-identical.
func TestLeadingRatioIsSizeSpecific(t *testing.T) {
	cfg := DefaultConfig()
	sizes := []float32{8, 10, 12, 14, 18, 24, 36, 72}
	prev := float32(0)
	for _, size := range sizes {
		ratio := cfg.LeadingRatio(size)
		if ratio <= 0 {
			t.Fatalf("non-positive leading ratio at %gpx", size)
		}
		if prev != 0 && ratio > prev {
			t.Errorf("leading ratio grew from %g to %g as text got bigger: must tighten", prev, ratio)
		}
		if !approx(cfg.Leading(size), size*ratio, 1e-4) {
			t.Errorf("Leading(%g) does not match LeadingRatio", size)
		}
		prev = ratio
	}
	if cfg.LeadingRatio(14) != 1.2 {
		t.Errorf("14px body leading = %g, want the existing 1.2 anchor", cfg.LeadingRatio(14))
	}
	if cfg.LeadingRatio(12) <= 1.2 || cfg.LeadingRatio(36) >= 1.2 {
		t.Error("captions must be looser and headings tighter than body copy")
	}
	if got := (*Config)(nil).Leading(12); !approx(got, 14.4, 1e-4) {
		t.Errorf("nil config fallback leading = %g, want 14.4", got)
	}
}

// §15: small text gets positive tracking, large display text gets negative.
func TestTrackingIsSizeSpecific(t *testing.T) {
	cfg := DefaultConfig()
	for _, size := range []float32{10, 12} {
		if cfg.Tracking(size) <= 0 {
			t.Errorf("Tracking(%g) = %g, want positive for captions", size, cfg.Tracking(size))
		}
	}
	for _, size := range []float32{14, 16} {
		if cfg.Tracking(size) != 0 {
			t.Errorf("Tracking(%g) = %g, want neutral body tracking", size, cfg.Tracking(size))
		}
	}
	if cfg.Tracking(24) >= 0 || cfg.Tracking(72) >= cfg.Tracking(24) {
		t.Errorf("large text tracking did not go negative / tighten: 24=%g 72=%g",
			cfg.Tracking(24), cfg.Tracking(72))
	}
}

// Dynamic Type scales the whole type system, and captions keep a readable
// floor instead of shrinking with the user's scale.
func TestCaptionSizeAndDynamicTypeScale(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.CaptionSize() != 12 {
		t.Fatalf("default caption size = %g, want 12", cfg.CaptionSize())
	}
	cfg.SetTextScale(2)
	if cfg.CaptionSize() != 24 {
		t.Errorf("caption did not scale with Dynamic Type: %g", cfg.CaptionSize())
	}
	cfg.SetTextScale(0.5)
	if cfg.CaptionSize() != 12 {
		t.Errorf("caption fell below the 12px legibility floor: %g", cfg.CaptionSize())
	}
	if got := (*Config)(nil).CaptionSize(); got != 12 {
		t.Errorf("nil config caption size = %g", got)
	}
}

func TestSetTextScaleScalesTypeAndSpacing(t *testing.T) {
	cfg := DefaultConfig()
	before := struct {
		fs, sm, lg, lh       float32
		xs, sm2, md, lg2, xl float32
	}{
		cfg.FontSize, cfg.FontSizeSm, cfg.FontSizeLg, cfg.LineHeight,
		cfg.SpaceXS, cfg.SpaceSM, cfg.SpaceMD, cfg.SpaceLG, cfg.SpaceXL,
	}
	cfg.SetTextScale(1.5)
	if !approx(cfg.FontSize, before.fs*1.5, 1e-3) || !approx(cfg.FontSizeSm, before.sm*1.5, 1e-3) ||
		!approx(cfg.FontSizeLg, before.lg*1.5, 1e-3) || !approx(cfg.LineHeight, before.lh*1.5, 1e-3) {
		t.Errorf("type sizes did not scale together: %+v", cfg)
	}
	if !approx(cfg.SpaceXS, before.xs*1.5, 1e-3) || !approx(cfg.SpaceMD, before.md*1.5, 1e-3) ||
		!approx(cfg.SpaceXL, before.xl*1.5, 1e-3) {
		t.Error("spacing tokens stayed fixed; a larger font would overflow its padding")
	}
	if cfg.TextScale != 1.5 {
		t.Errorf("TextScale = %g, want 1.5", cfg.TextScale)
	}

	cfg.SetTextScale(0) // invalid scales must be ignored
	if cfg.TextScale != 1.5 {
		t.Error("invalid text scale changed the configuration")
	}
	cfg.SetTextScale(1)
	if !approx(cfg.FontSize, before.fs, 1e-3) || !approx(cfg.FontSizeSm, before.sm, 1e-3) ||
		!approx(cfg.SpaceMD, before.md, 1e-3) {
		t.Error("returning to scale 1 did not restore the original type system")
	}
}

func TestEmitTextPrefersTrackingDrawer(t *testing.T) {
	cfg := DefaultConfig()
	probe := &typographyProbe{}
	cfg.TextRenderer = probe
	buf := render.NewCommandBuffer()

	emitText(cfg, buf, "Heading", 0, 0, 36, 200, cfg.TextColor, 1, 3)
	call := probe.last(t)
	if call.kind != "tracked" {
		t.Fatalf("emitText chose %q, want the tracking-aware path", call.kind)
	}
	if call.tracking != cfg.Tracking(36) || call.tracking >= 0 {
		t.Errorf("tracking = %g, want the size-specific negative value", call.tracking)
	}
	if call.z != 3 {
		t.Errorf("z = %d, want 3", call.z)
	}
}

func TestEmitCaptionTextPrefersCaptionDrawer(t *testing.T) {
	cfg := DefaultConfig()
	probe := &typographyProbe{}
	cfg.TextRenderer = probe
	buf := render.NewCommandBuffer()
	size := cfg.CaptionSize()

	emitCaptionText(cfg, buf, "分类标题", 0, 0, size, 200, cfg.TextColor, 1, 1)
	call := probe.last(t)
	if call.kind != "caption" {
		t.Fatalf("emitCaptionText chose %q, want the caption-weight path", call.kind)
	}
	if !approx(call.size, size, 1e-4) || call.tracking != cfg.Tracking(size) {
		t.Errorf("caption typography = size %g tracking %g, want %g/%g",
			call.size, call.tracking, size, cfg.Tracking(size))
	}
	if call.tracking <= 0 {
		t.Error("caption lost its positive legibility tracking")
	}
}

// A host with no caption capability must degrade to the regular tracked path
// rather than silently dropping the label.
func TestEmitCaptionTextDegradesWithoutCaptionDrawer(t *testing.T) {
	cfg := DefaultConfig()
	trackingProbe := &trackingProbeShim{}
	cfg.TextRenderer = trackingProbe
	buf := render.NewCommandBuffer()
	emitCaptionText(cfg, buf, "分类标题", 0, 0, 12, 200, cfg.TextColor, 1, 1)
	if trackingProbe.kind != "tracked" {
		t.Fatalf("fallback chose %q, want the regular tracked path", trackingProbe.kind)
	}
}

type trackingProbeShim struct {
	kind string
}

func (p *trackingProbeShim) DrawText(_ *render.CommandBuffer, _ string, _, _, _, _ float32, _ uimath.Color, _ float32) {
	p.kind = "plain"
}
func (p *trackingProbeShim) DrawTextZ(_ *render.CommandBuffer, _ string, _, _, _, _ float32, _ uimath.Color, _ float32, _ int32) {
	p.kind = "z"
}
func (p *trackingProbeShim) DrawTextTracked(_ *render.CommandBuffer, _ string, _, _, _, _ float32, _ uimath.Color, _ float32, _ int32, _ float32) {
	p.kind = "tracked"
}
func (p *trackingProbeShim) LineHeight(size float32) float32 { return size * 1.2 }
func (p *trackingProbeShim) MeasureText(text string, size float32) float32 {
	return float32(len([]rune(text))) * size * 0.6
}

// The widget that originally exposed the #A3A3A3 caption problem must use the
// whole §15 set: Dynamic Type caption size, size-aware leading, tracking and
// the caption-weight drawer.
func TestListGroupHeaderUsesCaptionTypography(t *testing.T) {
	tree := newTestTree()
	cfg := DefaultConfig()
	probe := &typographyProbe{}
	cfg.TextRenderer = probe

	list := NewList(tree, cfg)
	list.SetItems([]ListItem{
		{Value: "1", Group: "稳定开发中", Title: "第一项"},
		{Value: "2", Group: "稳定开发中", Title: "第二项"},
	})
	tree.SetLayout(list.ElementID(), core.LayoutResult{Bounds: uimath.NewRect(0, 0, 240, 200)})

	buf := render.NewCommandBuffer()
	list.Draw(buf)

	var caption *typographyCall
	for i := range probe.calls {
		if probe.calls[i].kind == "caption" && probe.calls[i].text == "稳定开发中" {
			caption = &probe.calls[i]
			break
		}
	}
	if caption == nil {
		t.Fatalf("group header did not use the caption drawer; calls = %+v", probe.calls)
	}
	if !approx(caption.size, cfg.CaptionSize(), 1e-4) {
		t.Errorf("caption size = %g, want CaptionSize() %g", caption.size, cfg.CaptionSize())
	}
	if caption.tracking != cfg.Tracking(cfg.CaptionSize()) {
		t.Errorf("caption tracking = %g, want %g", caption.tracking, cfg.Tracking(cfg.CaptionSize()))
	}

	cfg.SetTextScale(2)
	probe.calls = nil
	buf.Reset()
	list.Draw(buf)
	for _, c := range probe.calls {
		if c.kind == "caption" && c.text == "稳定开发中" {
			if !approx(c.size, cfg.CaptionSize(), 1e-4) || cfg.CaptionSize() <= 12 {
				t.Errorf("Dynamic Type did not reach the group header: size %g", c.size)
			}
			return
		}
	}
	t.Fatal("group header disappeared after Dynamic Type scaling")
}

// Source guard: fixed 1.2 leading must not creep back in. The two icon-size
// expressions in steps.go are dimensions, not line boxes, and are whitelisted.
func TestTypographySourceGuard(t *testing.T) {
	fixed := regexp.MustCompile(`[\w.]+\s*\*\s*1\.2`)
	allowed := map[string]bool{
		"iconSize := size * 1.2": true,
		"return size * 1.2":      true, // nil-Config fallback in config.go
	}
	violations := []string{}
	err := filepath.WalkDir(".", func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		for i, line := range strings.Split(string(data), "\n") {
			if !fixed.MatchString(line) || allowed[strings.TrimSpace(line)] {
				continue
			}
			violations = append(violations, path+":"+strconv.Itoa(i+1)+": "+strings.TrimSpace(line))
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk: %v", err)
	}
	if len(violations) != 0 {
		t.Errorf("fixed leading literals are back (%d):\n%s", len(violations), strings.Join(violations, "\n"))
	}

	listSrc, err := os.ReadFile("list.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"emitCaptionText", "cfg.CaptionSize()", "cfg.Leading(fs)"} {
		if !strings.Contains(string(listSrc), want) {
			t.Errorf("list.go no longer uses %q for group captions", want)
		}
	}
}
