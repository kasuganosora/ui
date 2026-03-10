//go:build windows

// Feed Timeline Demo — A Twitter/X-style feed to showcase GoUI's complex UI capabilities.
//
// Demonstrates:
//   - Pure HTML+CSS layout (flexbox, overflow:hidden, white-space:nowrap, text-overflow:ellipsis)
//   - Tabs, compose box, tweet cards with avatar, meta, text, actions
//   - Per-user consistent avatar colours (FNV hash)
//   - Retweet attribution headers
//   - Scroll-to-bottom auto-loads more tweets
//   - New tweets prepended every 10 s (goroutine → channel → main thread)
//
// Run: go run ./cmd/feed
package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"math/rand"
	"os"
	"strings"
	"time"

	ui "github.com/kasuganosora/ui"
	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/devtools"
	"github.com/kasuganosora/ui/event"
	uimath "github.com/kasuganosora/ui/math"
	"github.com/kasuganosora/ui/widget"
)

// ── Embedded tweet data ────────────────────────────────────────────────────

//go:embed tweets.json
var tweetsJSON []byte

type tweetData struct {
	Name    string `json:"name"`
	Handle  string `json:"handle"`
	Text    string `json:"text"`
	Fav     int    `json:"fav"`
	RT      int    `json:"rt"`
	Reply   int    `json:"reply"`
	Created string `json:"created"`
}

var allTweets []tweetData

// ── Avatar colours ─────────────────────────────────────────────────────────

// avatarPalette matches Twitter's avatar placeholder colours.
var avatarPalette = []string{
	"#1D9BF0", "#00BA7C", "#F91880", "#FF7700",
	"#794BC4", "#FF6C02", "#2B7DF0", "#DC1F4E",
	"#0DD3BB", "#8B98A5",
}

func avatarColor(name string) string {
	h := fnv.New32a()
	_, _ = h.Write([]byte(name))
	return avatarPalette[int(h.Sum32())%len(avatarPalette)]
}

// firstRune returns the first Unicode character of s as a string.
func firstRune(s string) string {
	for _, r := range s {
		return string(r)
	}
	return "?"
}

// ── Helpers ────────────────────────────────────────────────────────────────

func formatCount(n int) string {
	switch {
	case n >= 10000:
		return fmt.Sprintf("%.1f万", float64(n)/10000)
	case n >= 1000:
		return fmt.Sprintf("%.1fK", float64(n)/1000)
	default:
		return fmt.Sprintf("%d", n)
	}
}

func parseCreatedAt(s string) string {
	t, err := time.Parse("Mon Jan 02 15:04:05 -0700 2006", s)
	if err != nil {
		return s
	}
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "刚刚"
	case d < time.Hour:
		return fmt.Sprintf("%d分", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%d时", int(d.Hours()))
	case d < 7*24*time.Hour:
		return fmt.Sprintf("%d天", int(d.Hours()/24))
	default:
		return t.Format("1月2日")
	}
}

func htmlEsc(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	return s
}

// ── CSS ────────────────────────────────────────────────────────────────────

// feedCSS is the shared stylesheet for the entire feed page.
const feedCSS = `
/* ─── root ─────────────────────────────────────── */
.feed-root {
  display: flex;
  flex-direction: column;
  width: 100%;
  height: 100%;
  background: #000000;
}

/* ─── header / tabs ────────────────────────────── */
.feed-header {
  display: flex;
  flex-direction: column;
  background: rgba(0,0,0,0.9);
  border-bottom: 1px solid #2F3336;
}
.feed-tabs {
  display: flex;
  flex-direction: row;
  height: 53px;
  align-items: center;
}
.tab-item {
  flex-grow: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  height: 53px;
  justify-content: center;
  gap: 0;
}
.tab-item-active {
  flex-grow: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  height: 53px;
  justify-content: center;
  gap: 0;
}
.tab-label {
  font-size: 15px;
  color: #71767B;
}
.tab-label-active {
  font-size: 15px;
  color: #E7E9EA;
}
.tab-indicator {
  width: 56px;
  height: 3px;
  background: #1D9BF0;
  border-radius: 2px;
  margin-top: 4px;
}
.tab-indicator-hidden {
  width: 56px;
  height: 3px;
  border-radius: 2px;
  margin-top: 4px;
}

/* ─── compose box ───────────────────────────────── */
.compose {
  display: flex;
  flex-direction: row;
  padding: 12px 16px 8px 16px;
  border-bottom: 1px solid #2F3336;
  background: #000000;
  gap: 12px;
}
.compose-avatar {
  width: 40px;
  height: 40px;
  border-radius: 20px;
  flex-shrink: 0;
  background: #1D9BF0;
  display: flex;
  align-items: center;
  justify-content: center;
}
.compose-body {
  display: flex;
  flex-direction: column;
  flex-grow: 1;
  gap: 4px;
  overflow: hidden;
  min-width: 0;
}
.compose-audience {
  display: flex;
  flex-direction: row;
  align-items: center;
  border: 1px solid #1D9BF0;
  border-radius: 12px;
  padding: 2px 10px;
  height: 24px;
}
.compose-audience-text {
  font-size: 13px;
  color: #1D9BF0;
}
.compose-input {
  font-size: 20px;
  color: #E7E9EA;
  background: transparent;
  border: none;
  width: 100%;
}
.compose-reply-perm {
  display: flex;
  flex-direction: row;
  align-items: center;
  padding: 4px 0;
  gap: 4px;
}
.compose-reply-icon {
  font-size: 13px;
  color: #1D9BF0;
}
.compose-reply-text {
  font-size: 13px;
  color: #1D9BF0;
}
.compose-divider {
  height: 1px;
  background: #2F3336;
  margin: 5px 0;
}
.compose-footer {
  display: flex;
  flex-direction: row;
  justify-content: space-between;
  align-items: center;
}
.compose-toolbar {
  display: flex;
  flex-direction: row;
  gap: 0;
  align-items: center;
  flex-shrink: 1;
  overflow: hidden;
}
.toolbar-icon {
  width: 32px;
  height: 32px;
  flex-shrink: 0;
}
.compose-btn {
  padding: 8px 20px;
  height: 36px;
  min-width: 72px;
  flex-shrink: 0;
}

/* ─── scrollable timeline ───────────────────────── */
.feed-timeline {
  flex-grow: 1;
  overflow: scroll;
}

/* ─── retweet attribution header ───────────────── */
.rt-header {
  display: flex;
  flex-direction: row;
  padding: 8px 16px 0 68px;
  align-items: center;
  gap: 6px;
}
.rt-icon {
  font-size: 13px;
  color: #71767B;
}
.rt-name {
  font-size: 13px;
  color: #71767B;
}

/* ─── tweet card ────────────────────────────────── */
.tweet {
  display: flex;
  flex-direction: column;
  border-bottom: 1px solid #2F3336;
}
.tweet-inner {
  display: flex;
  flex-direction: row;
  padding: 12px 16px;
  gap: 12px;
}

/* ─── avatar ────────────────────────────────────── */
.tweet-avatar {
  width: 40px;
  height: 40px;
  border-radius: 20px;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
}
.avatar-initial {
  color: #FFFFFF;
  font-size: 18px;
}

/* ─── tweet content column ──────────────────────── */
.tweet-body {
  display: flex;
  flex-direction: column;
  flex-grow: 1;
  gap: 3px;
  min-width: 0;
}

/* name + handle + time — single line with ellipsis */
.tweet-meta {
  display: flex;
  flex-direction: row;
  align-items: center;
  gap: 4px;
  overflow: hidden;
}
.tweet-name {
  font-size: 15px;
  color: #E7E9EA;
  white-space: nowrap;
  text-overflow: ellipsis;
  overflow: hidden;
  min-width: 0;
}
.tweet-sep {
  font-size: 15px;
  color: #71767B;
  white-space: nowrap;
}
.tweet-handle {
  font-size: 15px;
  color: #71767B;
  white-space: nowrap;
}
.tweet-dot {
  font-size: 15px;
  color: #71767B;
  white-space: nowrap;
}
.tweet-time {
  font-size: 15px;
  color: #71767B;
  white-space: nowrap;
}

/* tweet text body — full height, no cap */
.tweet-text-wrap {
  overflow: hidden;
}
.tweet-text {
  font-size: 15px;
  color: #E7E9EA;
}

/* action row */
.tweet-actions {
  display: flex;
  flex-direction: row;
  margin-top: 6px;
  gap: 40px;
  align-items: center;
}
.action-reply { font-size: 13px; color: #71767B; }
.action-rt    { font-size: 13px; color: #71767B; }
.action-like  { font-size: 13px; color: #71767B; }
.action-views { font-size: 13px; color: #71767B; }
`

// ── HTML skeleton ──────────────────────────────────────────────────────────

const feedHTML = `
<div class="feed-root">
  <div class="feed-header">
    <div class="feed-tabs">
      <div id="tab-0" class="tab-item-active">
        <span id="tab-label-0" class="tab-label-active">为你推荐</span>
        <div id="tab-ind-0" class="tab-indicator"></div>
      </div>
      <div id="tab-1" class="tab-item">
        <span id="tab-label-1" class="tab-label">正在关注</span>
        <div id="tab-ind-1" class="tab-indicator-hidden"></div>
      </div>
      <div id="tab-2" class="tab-item">
        <span id="tab-label-2" class="tab-label">メイド情報局</span>
        <div id="tab-ind-2" class="tab-indicator-hidden"></div>
      </div>
      <div id="tab-3" class="tab-item">
        <span id="tab-label-3" class="tab-label">墙外弱智吧</span>
        <div id="tab-ind-3" class="tab-indicator-hidden"></div>
      </div>
    </div>
  </div>
  <div class="compose">
    <div class="compose-avatar"><span class="avatar-initial">我</span></div>
    <div class="compose-body">
      <div class="compose-audience">
        <span class="compose-audience-text">每个人 ↓</span>
      </div>
      <textarea id="compose-input" class="compose-input" placeholder="有什么新鲜事？" rows="2"></textarea>
      <div class="compose-reply-perm">
        <span class="compose-reply-icon">①</span>
        <span class="compose-reply-text">所有人可以回复</span>
      </div>
      <div class="compose-divider"></div>
      <div class="compose-footer">
        <div class="compose-toolbar">
          <button id="tb-img"      class="toolbar-icon" variant="text" theme="primary" shape="round">⊞</button>
          <button id="tb-gif"      class="toolbar-icon" variant="text" theme="primary" shape="round">GIF</button>
          <button id="tb-poll"     class="toolbar-icon" variant="text" theme="primary" shape="round">○</button>
          <button id="tb-thread"   class="toolbar-icon" variant="text" theme="primary" shape="round">≡</button>
          <button id="tb-emoji"    class="toolbar-icon" variant="text" theme="primary" shape="round">☺</button>
          <button id="tb-schedule" class="toolbar-icon" variant="text" theme="primary" shape="round">⊡</button>
          <button id="tb-loc"      class="toolbar-icon" variant="text" theme="primary" shape="round">◎</button>
        </div>
        <button id="compose-post" class="compose-btn" theme="primary" shape="round">发帖</button>
      </div>
    </div>
  </div>
  <main id="timeline" class="feed-timeline"></main>
</div>
`

// ── Tweet HTML builder ─────────────────────────────────────────────────────

// tweetHTML builds a single tweet card as HTML.
// retweetedBy: if non-empty, shows a "X 已转推" header above the card.
func tweetHTML(td *tweetData, timeStr, retweetedBy string) string {
	name := htmlEsc(td.Name)
	handle := htmlEsc(td.Handle)
	text := htmlEsc(strings.ReplaceAll(td.Text, "\n", " "))
	color := avatarColor(td.Name)
	initial := firstRune(td.Name)

	// Estimate view count as a multiple of fav (Twitter-style)
	views := td.Fav * (7 + int(fnv.New32a().Sum32())%8)
	if views == 0 {
		views = td.Reply*50 + td.RT*30 + 200
	}

	rtHeader := ""
	if retweetedBy != "" {
		rtHeader = fmt.Sprintf(`
  <div class="rt-header">
    <span class="rt-icon">↺</span>
    <span class="rt-name">%s 已转推</span>
  </div>`, htmlEsc(retweetedBy))
	}

	return fmt.Sprintf(`
<div class="tweet">%s
  <div class="tweet-inner">
    <div class="tweet-avatar" style="background:%s;"><span class="avatar-initial">%s</span></div>
    <div class="tweet-body">
      <div class="tweet-meta">
        <span class="tweet-name">%s</span>
        <span class="tweet-sep">·</span>
        <span class="tweet-handle">%s</span>
        <span class="tweet-dot">·</span>
        <span class="tweet-time">%s</span>
      </div>
      <div class="tweet-text-wrap">
        <span class="tweet-text">%s</span>
      </div>
      <div class="tweet-actions">
        <span class="action-reply">○ %s</span>
        <span class="action-rt">↺ %s</span>
        <span class="action-like">♡ %s</span>
        <span class="action-views">↗ %s</span>
      </div>
    </div>
  </div>
</div>`,
		rtHeader,
		color, initial,
		name, handle, timeStr,
		text,
		formatCount(td.Reply), formatCount(td.RT), formatCount(td.Fav), formatCount(views),
	)
}

// ── Action button binding ──────────────────────────────────────────────────

// actionNames maps action-row indices to human-readable labels.
var actionNames = []string{"reply", "rt", "like", "views"}

// bindTweetActions walks the tweet widget tree to find the action row
// (a Div with exactly 4 Text children) and registers click handlers.
func bindTweetActions(t *core.Tree, w widget.Widget, td *tweetData, depth int) {
	if depth > 6 {
		return
	}
	children := w.Children()
	// Count direct Text children — tweet-actions has exactly 4
	textKids := 0
	for _, c := range children {
		if _, ok := c.(*widget.Text); ok {
			textKids++
		}
	}
	if textKids == 4 && len(children) == 4 {
		// Found the action row — bind each Text icon
		for i, child := range children {
			idx := i
			name := td.Name
			t.AddHandler(child.ElementID(), event.MouseClick, func(e *event.Event) {
				fmt.Printf("[Feed] %s on tweet by %s\n", actionNames[idx], name)
			})
		}
		return
	}
	for _, c := range children {
		bindTweetActions(t, c, td, depth+1)
	}
}

// ── Constants ─────────────────────────────────────────────────────────────

const loadBatchSize = 5

// pendingTweetData carries tweet data from the goroutine to the main thread.
var pendingTweetData = make(chan *tweetData, 32)

// ── Main ───────────────────────────────────────────────────────────────────

func main() {
	if err := json.Unmarshal(tweetsJSON, &allTweets); err != nil {
		fmt.Fprintf(os.Stderr, "parse tweets: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("[Feed] Loaded %d tweets\n", len(allTweets))

	dt := devtools.NewServer(devtools.Options{
		Addr:    ":9222",
		AppName: "GoUI Feed Timeline",
	})
	go dt.Start()

	app, err := ui.NewApp(ui.AppOptions{
		Title:    "GoUI Feed Timeline",
		Width:    600,
		Height:   860,
		DevTools: dt,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
	defer app.Destroy()

	// Apply Twitter dark theme BEFORE LoadHTML so widget constructors get the correct colors.
	cfg := app.Config()
	cfg.BgColor = uimath.ColorHex("#16181C")
	cfg.TextColor = uimath.ColorHex("#E7E9EA")
	cfg.BorderColor = uimath.ColorHex("#2F3336")
	cfg.PrimaryColor = uimath.ColorHex("#1D9BF0")
	cfg.FocusBorderColor = uimath.ColorHex("#1D9BF0")
	cfg.DisabledColor = uimath.ColorHex("#536471") // placeholder color

	doc := app.LoadHTML(feedHTML + "<style>" + feedCSS + "</style>")
	tree := app.Tree()

	timeline := doc.QueryByID("timeline")
	if timeline == nil {
		fmt.Fprintln(os.Stderr, "timeline not found")
		os.Exit(1)
	}
	contentWidget := timeline.(*widget.Content)

	// Tab labels and their IDs for switching.
	tabLabels := []string{"为你推荐", "正在关注", "メイド情報局", "墙外弱智吧"}
	tabWidgets := make([]widget.Widget, len(tabLabels))
	tabIndicators := make([]widget.Widget, len(tabLabels))
	tabLabelWidgets := make([]widget.Widget, len(tabLabels))
	for i := range tabLabels {
		tabWidgets[i] = doc.QueryByID(fmt.Sprintf("tab-%d", i))
		tabIndicators[i] = doc.QueryByID(fmt.Sprintf("tab-ind-%d", i))
		tabLabelWidgets[i] = doc.QueryByID(fmt.Sprintf("tab-label-%d", i))
	}
	activeTab := 0

	setTabLabelColor := func(w widget.Widget, c uimath.Color) {
		if txt, ok := w.(*widget.Text); ok {
			txt.SetColor(c)
		}
	}

	switchTab := func(idx int) {
		if idx == activeTab {
			return
		}
		// Hide indicator + dim label for old tab
		if ind := tabIndicators[activeTab]; ind != nil {
			tree.SetVisible(ind.ElementID(), false)
		}
		if lbl := tabLabelWidgets[activeTab]; lbl != nil {
			setTabLabelColor(lbl, uimath.ColorHex("#71767B"))
		}
		activeTab = idx
		// Show indicator + brighten label for new tab
		if ind := tabIndicators[idx]; ind != nil {
			tree.SetVisible(ind.ElementID(), true)
		}
		if lbl := tabLabelWidgets[idx]; lbl != nil {
			setTabLabelColor(lbl, uimath.ColorHex("#E7E9EA"))
		}
		tree.MarkDirty(tree.Root())
		fmt.Printf("[Feed] Switched to tab: %s\n", tabLabels[idx])
	}

	for i := range tabLabels {
		w := tabWidgets[i]
		if w == nil {
			continue
		}
		tree.AddHandler(w.ElementID(), event.MouseClick, func(e *event.Event) {
			switchTab(i)
		})
	}

	// applyCSS() overwrites the widget's style including the height set by SetRows.
	// Re-apply rows after HTML loading to restore the correct height, and enable autosize.
	var composeTA *widget.TextArea
	if inp := doc.QueryByID("compose-input"); inp != nil {
		if ta, ok := inp.(*widget.TextArea); ok {
			ta.SetRows(2)
			ta.SetAutosizeRows(2, 8)
			composeTA = ta
		}
	}

	// Toolbar icon click handlers
	toolbarActions := []struct{ id, label string }{
		{"tb-img", "图片"},
		{"tb-gif", "GIF"},
		{"tb-poll", "投票"},
		{"tb-thread", "话题串"},
		{"tb-emoji", "表情"},
		{"tb-schedule", "定时发送"},
		{"tb-loc", "位置"},
	}
	for _, act := range toolbarActions {
		if w := doc.QueryByID(act.id); w != nil {
			label := act.label
			tree.AddHandler(w.ElementID(), event.MouseClick, func(e *event.Event) {
				fmt.Printf("[Feed] 工具栏: %s\n", label)
			})
		}
	}

	// Compose post button
	if postBtn := doc.QueryByID("compose-post"); postBtn != nil {
		if btn, ok := postBtn.(*widget.Button); ok {
			btn.OnClick(func() {
				if inp := doc.QueryByID("compose-input"); inp != nil {
					if ta, ok := inp.(*widget.TextArea); ok {
						text := ta.Value()
						if text != "" {
							fmt.Printf("[Feed] Post: %s\n", text)
							ta.SetValue("")
						}
					}
				}
			})
		}
	}

	type feedContainer interface {
		widget.Widget
		AppendChild(widget.Widget)
		PrependChild(widget.Widget)
	}
	container := timeline.(feedContainer)

	// makeTweet converts tweetData to a widget subtree via HTML+CSS.
	// retweetedBy: non-empty = show RT attribution header (e.g. index % 5 == 0).
	makeTweet := func(td *tweetData, timeOverride, retweetedBy string) widget.Widget {
		t := timeOverride
		if t == "" {
			t = parseCreatedAt(td.Created)
		}
		html := tweetHTML(td, t, retweetedBy)
		root := ui.LoadHTMLWithCSS(tree, cfg, html, feedCSS)
		var tw widget.Widget = root
		if len(root.Children()) > 0 {
			tw = root.Children()[0]
		}
		bindTweetActions(tree, tw, td, 0)
		return tw
	}

	// Seed initial tweets (first 15, every 5th shows RT attribution).
	count := min(15, len(allTweets))
	for i := range count {
		var rtBy string
		if i > 0 && i%5 == 0 {
			rtBy = allTweets[(i+1)%len(allTweets)].Name
		}
		tw := makeTweet(&allTweets[i], "", rtBy)
		container.AppendChild(tw)
	}

	// Auto-load more tweets when scrolled near bottom.
	var lastLoad time.Time
	onNearBottom := func() {
		if time.Since(lastLoad) < 500*time.Millisecond {
			return
		}
		lastLoad = time.Now()
		for range loadBatchSize {
			td := &allTweets[rand.Intn(len(allTweets))]
			tw := makeTweet(td, "", "")
			container.AppendChild(tw)
		}
		tree.MarkDirty(tree.Root())
		fmt.Printf("[Feed] +%d tweets (total: %d)\n", loadBatchSize, len(contentWidget.Children()))
	}

	// Goroutine: send new tweet data every 10 s.
	// Widget creation must happen on the main thread.
	go func() {
		for {
			time.Sleep(10 * time.Second)
			pendingTweetData <- &allTweets[rand.Intn(len(allTweets))]
		}
	}()

	// Layout callback: drain channel, then CSS layout.
	app.SetOnLayout(func(tree *core.Tree, root widget.Widget, w, h float32) {
		for {
			select {
			case td := <-pendingTweetData:
				tw := makeTweet(td, "刚刚", "")
				container.PrependChild(tw)
				fmt.Println("[Feed] New tweet prepended")
			default:
				goto done
			}
		}
	done:
		ui.CSSLayout(tree, root, w, h, cfg)

		// Autosize compose textarea (must run after layout so bounds.Width is known).
		if composeTA != nil {
			composeTA.UpdateAutosizeHeight()
		}

		// Auto-load near bottom
		scrollY := contentWidget.ScrollY()
		contentH := contentWidget.ContentHeight()
		bounds := contentWidget.Bounds()
		if maxScroll := contentH - bounds.Height; maxScroll > 0 && scrollY >= maxScroll-120 {
			onNearBottom()
		}
	})

	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
