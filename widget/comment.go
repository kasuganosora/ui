package widget

import (
	"github.com/kasuganosora/ui/core"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
)

// CommentData holds comment display data.
type CommentData struct {
	Author  string
	Content string
	Time    string
	Avatar  string
}

// Comment displays a single comment with author info.
type Comment struct {
	Base
	data     CommentData
	children []*Comment
}

func NewComment(tree *core.Tree, data CommentData, cfg *Config) *Comment {
	if cfg == nil {
		cfg = DefaultConfig()
	}
	return &Comment{
		Base: NewBase(tree, core.TypeCustom, cfg),
		data: data,
	}
}

func (c *Comment) Data() CommentData       { return c.data }
func (c *Comment) SetData(d CommentData)   { c.data = d }
func (c *Comment) Replies() []*Comment      { return c.children }

func (c *Comment) AddReply(reply *Comment) {
	c.children = append(c.children, reply)
}

func (c *Comment) Draw(buf *render.CommandBuffer) {
	bounds := c.Bounds()
	if bounds.IsEmpty() {
		return
	}
	cfg := c.config
	avatarSize := float32(32)

	// Avatar circle
	buf.DrawRect(render.RectCmd{
		Bounds:    uimath.NewRect(bounds.X, bounds.Y, avatarSize, avatarSize),
		FillColor: uimath.RGBA(0, 0, 0, 0.06),
		Corners:   uimath.CornersAll(avatarSize / 2),
	}, 1, 1)

	// Avatar initial
	if cfg.TextRenderer != nil && c.data.Author != "" {
		initial := c.data.Author[:1]
		tw := cfg.TextRenderer.MeasureText(initial, cfg.FontSizeSm)
		lh := cfg.TextRenderer.LineHeight(cfg.FontSizeSm)
		cfg.TextRenderer.DrawText(buf, initial, bounds.X+(avatarSize-tw)/2, bounds.Y+(avatarSize-lh)/2, cfg.FontSizeSm, avatarSize, cfg.TextColor, 1)
	}

	textX := bounds.X + avatarSize + cfg.SpaceSM

	if cfg.TextRenderer != nil {
		lh := cfg.TextRenderer.LineHeight(cfg.FontSizeSm)
		// Author name
		cfg.TextRenderer.DrawText(buf, c.data.Author, textX, bounds.Y, cfg.FontSizeSm, bounds.Width-textX+bounds.X, cfg.TextColor, 1)
		// Time
		if c.data.Time != "" {
			authorW := cfg.TextRenderer.MeasureText(c.data.Author, cfg.FontSizeSm)
			cfg.TextRenderer.DrawText(buf, c.data.Time, textX+authorW+cfg.SpaceSM, bounds.Y, cfg.FontSizeSm, bounds.Width, cfg.DisabledColor, 1)
		}
		// Content
		cfg.TextRenderer.DrawText(buf, c.data.Content, textX, bounds.Y+lh+4, cfg.FontSize, bounds.Width-textX+bounds.X, cfg.TextColor, 1)
	}
}
