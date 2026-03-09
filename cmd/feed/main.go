//go:build windows

// Feed Timeline Demo — A Twitter-like feed to showcase GoUI's complex UI capabilities.
//
// This demo demonstrates:
//   - HTML+CSS layout with flexbox
//   - Dynamic content insertion (new tweet every 10 seconds)
//   - Scrollable content area
//   - Avatar, text, icons, action buttons
//   - CSS styling (colors, spacing, borders, hover)
//
// Run: go run ./cmd/feed
package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	ui "github.com/kasuganosora/ui"
	"github.com/kasuganosora/ui/core"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/render"
	"github.com/kasuganosora/ui/widget"
)

// ── Random tweet data ──────────────────────────────────────────────────

var users = []struct {
	name, handle string
	avatarColor  uimath.Color
}{
	{"猫月ユキ", "@yukinkzk", uimath.ColorHex("#FFB6C1")},
	{"Linus Torvalds", "@Linus__Torvalds", uimath.ColorHex("#4A90D9")},
	{"尚硅谷", "@atguigu", uimath.ColorHex("#FF6B35")},
	{"GoUI Official", "@GoUI_dev", uimath.ColorHex("#2BA471")},
	{"Miku Hatsune", "@cfm_miku", uimath.ColorHex("#39C5BB")},
	{"Rust Lang", "@rustlang", uimath.ColorHex("#DEA584")},
	{"小岛秀夫", "@HIDEO_KOJIMA_EN", uimath.ColorHex("#8B5CF6")},
	{"Elon Musk", "@elonmusk", uimath.ColorHex("#1DA1F2")},
	{"阮一峰", "@ruaborntree", uimath.ColorHex("#E8590C")},
	{"React", "@reactjs", uimath.ColorHex("#61DAFB")},
}

var tweets = []string{
	"今天天气真好，适合写代码 ☀️",
	"Just released v2.0! Check out the new features 🚀",
	"Go 语言的错误处理确实需要改进，但 error wrapping 已经好很多了",
	"每次看到自己三个月前写的代码都想重构…",
	"Zero-CGO is the way. Pure Go, pure performance. 💪",
	"刚在 GitHub 上发现一个很棒的 UI 库，渲染引擎支持 Vulkan/DX11/OpenGL",
	"Flexbox is still the best layout system. Change my mind.",
	"今日のコーディングは楽しかったです！新しい機能を実装しました",
	"软路由里的内存比它的主板还值钱了（ 还有王法吗，还有法律吗！",
	"CSS-in-Go might sound crazy, but it works surprisingly well",
	"周末在家学习 Vulkan，感觉比 OpenGL 复杂但更合理",
	"The best code is no code. The second best is Go code.",
	"对不起，我又在凌晨3点推送了一个 breaking change 🙈",
	"This GPU-accelerated UI framework runs at 144fps. Not bad for Go!",
	"新建了一个 Twitter 风格的 feed，用纯 Go 渲染的",
	"半夜饿了，外卖还是泡面？这是个问题",
	"SDF font rendering + linear-space blending = crisp text at any size",
	"今天的 PR 终于合并了，开心到起飞 🎉",
	"Does anyone else rewrite their side project from scratch every 6 months?",
	"レンダリングパイプラインをゼロから作り直しました。大変でしたが満足しています",
}

var timeAgo = []string{
	"1分前", "3分前", "5分前", "10分前", "15分前", "30分前",
	"1小时前", "2小时前", "3小时前", "5小时前", "8小时前", "12小时前",
	"昨天", "2天前", "3天前",
}

// ── HTML template ──────────────────────────────────────────────────────

const feedHTML = `
<div>
  <header id="feed-header">
    <span>首页</span>
  </header>
  <main id="feed-content">
  </main>
</div>
<style>
div {
  display: flex;
  flex-direction: column;
  width: 100%;
  height: 100%;
  background: #15202B;
}

header {
  display: flex;
  flex-direction: row;
  align-items: center;
  height: 53px;
  padding-left: 16px;
  padding-right: 16px;
  background: #15202B;
  border-bottom: 1px solid #38444D;
}

header span {
  font-size: 20px;
  color: #E7E9EA;
}

main {
  flex-grow: 1;
  background: #15202B;
  overflow: scroll;
}

.tweet {
  display: flex;
  flex-direction: row;
  padding: 12px 16px;
  border-bottom: 1px solid #38444D;
  gap: 12px;
}

.tweet-avatar {
  width: 48px;
  height: 48px;
  border-radius: 24px;
  flex-shrink: 0;
}

.tweet-body {
  display: flex;
  flex-direction: column;
  flex-grow: 1;
  gap: 4px;
}

.tweet-header {
  display: flex;
  flex-direction: row;
  gap: 4px;
  align-items: center;
}

.tweet-name {
  font-size: 15px;
  color: #E7E9EA;
}

.tweet-handle {
  font-size: 13px;
  color: #71767B;
}

.tweet-time {
  font-size: 13px;
  color: #71767B;
}

.tweet-text {
  font-size: 15px;
  color: #E7E9EA;
}

.tweet-actions {
  display: flex;
  flex-direction: row;
  justify-content: space-between;
  padding-right: 80px;
  margin-top: 8px;
}

.action-btn {
  font-size: 13px;
  color: #71767B;
}
</style>
`

func main() {
	app, err := ui.NewApp(ui.AppOptions{
		Title:  "GoUI Feed Timeline",
		Width:  600,
		Height: 800,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	defer app.Destroy()

	doc := app.LoadHTML(feedHTML)
	tree := app.Tree()
	cfg := app.Config()

	// Get the content area for inserting tweets
	content := doc.QueryByID("feed-content")
	if content == nil {
		fmt.Fprintln(os.Stderr, "feed-content not found")
		os.Exit(1)
	}
	container, ok := content.(interface {
		widget.Widget
		AppendChild(widget.Widget)
	})
	if !ok {
		fmt.Fprintln(os.Stderr, "content is not a container")
		os.Exit(1)
	}

	// Use custom layout that supports our feed
	app.SetOnLayout(func(tree *core.Tree, root widget.Widget, w, h float32) {
		layoutFeed(tree, root, w, h)
	})

	// Seed initial tweets
	for i := 0; i < 8; i++ {
		addRandomTweet(tree, cfg, container)
	}

	// Insert a new tweet every 10 seconds via a goroutine + MarkDirty
	go func() {
		for {
			time.Sleep(10 * time.Second)
			addRandomTweet(tree, cfg, container)
			tree.MarkDirty(tree.Root())
			fmt.Println("[Feed] New tweet inserted")
		}
	}()

	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

// addRandomTweet creates a tweet widget and prepends it to the container.
func addRandomTweet(tree *core.Tree, cfg *widget.Config, container interface {
	widget.Widget
	AppendChild(widget.Widget)
}) {
	user := users[rand.Intn(len(users))]
	text := tweets[rand.Intn(len(tweets))]
	age := timeAgo[rand.Intn(len(timeAgo))]

	replies := rand.Intn(200)
	retweets := rand.Intn(500)
	likes := rand.Intn(2000)

	tweet := newTweetWidget(tree, cfg, user.name, user.handle, user.avatarColor, text, age, replies, retweets, likes)
	container.AppendChild(tweet)
	tree.AppendChild(container.ElementID(), tweet.ElementID())
}

// ── Tweet widget construction ──────────────────────────────────────────

type tweetWidget struct {
	widget.Base
	avatarColor uimath.Color
	name        string
	handle      string
	text        string
	time        string
	replies     int
	retweets    int
	likes       int
	cfg         *widget.Config
}

func newTweetWidget(tree *core.Tree, cfg *widget.Config, name, handle string, avatarClr uimath.Color, text, timeStr string, replies, retweets, likes int) *tweetWidget {
	tw := &tweetWidget{
		Base:        widget.NewBase(tree, core.TypeDiv, cfg),
		avatarColor: avatarClr,
		name:        name,
		handle:      handle,
		text:        text,
		time:        timeStr,
		replies:     replies,
		retweets:    retweets,
		likes:       likes,
		cfg:         cfg,
	}
	return tw
}

func (tw *tweetWidget) Draw(buf *render.CommandBuffer) {
	bounds := tw.Bounds()
	if bounds.IsEmpty() {
		return
	}

	cfg := tw.cfg
	pad := float32(12)
	gap := float32(12)
	avatarSize := float32(48)

	// Bottom border
	buf.DrawRect(render.RectCmd{
		Bounds:    uimath.NewRect(bounds.X, bounds.Y+bounds.Height-1, bounds.Width, 1),
		FillColor: uimath.ColorHex("#38444D"),
	}, 0, 1)

	// Avatar circle
	ax := bounds.X + pad
	ay := bounds.Y + pad
	buf.DrawRect(render.RectCmd{
		Bounds:    uimath.NewRect(ax, ay, avatarSize, avatarSize),
		FillColor: tw.avatarColor,
		Corners:   uimath.CornersAll(avatarSize / 2),
	}, 0, 1)

	// Avatar initial letter
	if cfg.TextRenderer != nil {
		initial := string([]rune(tw.name)[0])
		fontSize := float32(20)
		lh := cfg.TextRenderer.LineHeight(fontSize)
		iw := cfg.TextRenderer.MeasureText(initial, fontSize)
		cfg.TextRenderer.DrawText(buf, initial,
			ax+(avatarSize-iw)/2, ay+(avatarSize-lh)/2,
			fontSize, avatarSize, uimath.ColorWhite, 1)
	}

	// Text area
	textX := ax + avatarSize + gap
	textW := bounds.Width - pad*2 - avatarSize - gap

	if cfg.TextRenderer != nil {
		// Header line: Name · @handle · time
		nameY := bounds.Y + pad
		nameColor := uimath.ColorHex("#E7E9EA")
		handleColor := uimath.ColorHex("#71767B")
		nameFontSize := float32(15)
		smallFontSize := float32(13)

		cx := textX
		cfg.TextRenderer.DrawText(buf, tw.name, cx, nameY, nameFontSize, textW, nameColor, 1)
		cx += cfg.TextRenderer.MeasureText(tw.name, nameFontSize) + 4

		handleStr := tw.handle + " · " + tw.time
		cfg.TextRenderer.DrawText(buf, handleStr, cx, nameY+1, smallFontSize, textW-(cx-textX), handleColor, 1)

		// Tweet text
		textY := nameY + cfg.TextRenderer.LineHeight(nameFontSize) + 4
		cfg.TextRenderer.DrawText(buf, tw.text, textX, textY, nameFontSize, textW, nameColor, 1)

		// Action buttons row
		actY := textY + cfg.TextRenderer.LineHeight(nameFontSize) + 10
		actionColor := uimath.ColorHex("#71767B")
		actionSpacing := textW / 4

		// Reply
		cfg.DrawMDIcon(buf, "chat_bubble_outline", textX, actY, 16, actionColor, 0, 1)
		cfg.TextRenderer.DrawText(buf, fmt.Sprintf("%d", tw.replies), textX+20, actY+1, smallFontSize, 60, actionColor, 1)

		// Retweet
		rtX := textX + actionSpacing
		cfg.DrawMDIcon(buf, "repeat", rtX, actY, 16, uimath.ColorHex("#00BA7C"), 0, 1)
		cfg.TextRenderer.DrawText(buf, fmt.Sprintf("%d", tw.retweets), rtX+20, actY+1, smallFontSize, 60, uimath.ColorHex("#00BA7C"), 1)

		// Like
		likeX := textX + actionSpacing*2
		cfg.DrawMDIcon(buf, "favorite_border", likeX, actY, 16, uimath.ColorHex("#F91880"), 0, 1)
		cfg.TextRenderer.DrawText(buf, fmt.Sprintf("%d", tw.likes), likeX+20, actY+1, smallFontSize, 60, uimath.ColorHex("#F91880"), 1)

		// Share
		shareX := textX + actionSpacing*3
		cfg.DrawMDIcon(buf, "share", shareX, actY, 16, actionColor, 0, 1)
	}
}

// ── Custom layout ──────────────────────────────────────────────────────

func layoutFeed(tree *core.Tree, root widget.Widget, w, h float32) {
	tree.SetLayout(root.ElementID(), core.LayoutResult{
		Bounds: uimath.NewRect(0, 0, w, h),
	})

	children := root.Children()
	if len(children) == 0 {
		return
	}

	// Root div
	rootDiv := children[0]
	tree.SetLayout(rootDiv.ElementID(), core.LayoutResult{
		Bounds: uimath.NewRect(0, 0, w, h),
	})

	divChildren := rootDiv.Children()
	if len(divChildren) < 2 {
		return
	}

	// Header
	headerH := float32(53)
	header := divChildren[0]
	tree.SetLayout(header.ElementID(), core.LayoutResult{
		Bounds: uimath.NewRect(0, 0, w, headerH),
	})
	// Layout header children (span)
	for _, child := range header.Children() {
		tree.SetLayout(child.ElementID(), core.LayoutResult{
			Bounds: uimath.NewRect(16, 0, w-32, headerH),
		})
	}

	// Content area (main)
	contentY := headerH
	contentH := h - headerH
	contentW := divChildren[1]
	tree.SetLayout(contentW.ElementID(), core.LayoutResult{
		Bounds: uimath.NewRect(0, contentY, w, contentH),
	})

	// Get scroll offset
	scrollY := float32(0)
	if c, ok := contentW.(*widget.Content); ok {
		scrollY = c.ScrollY()
	}

	// Layout tweets
	tweetH := float32(120) // estimated height per tweet
	tweetChildren := contentW.Children()
	totalH := float32(0)
	cy := contentY - scrollY

	for _, child := range tweetChildren {
		tree.SetLayout(child.ElementID(), core.LayoutResult{
			Bounds: uimath.NewRect(0, cy, w, tweetH),
		})
		cy += tweetH
		totalH += tweetH
	}

	// Set content height for scrolling
	if c, ok := contentW.(*widget.Content); ok {
		c.SetContentHeight(totalH)
		c.ScrollBy(0) // clamp
	}
}

