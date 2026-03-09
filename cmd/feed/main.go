//go:build windows

// Feed Timeline Demo — A Twitter-like feed to showcase GoUI's complex UI capabilities.
//
// This demo demonstrates:
//   - Pure HTML+CSS layout using CSSLayout engine (flexbox/block flow)
//   - Each tweet built from HTML template with CSS classes
//   - Real Twitter data loaded from embedded JSON (go:embed)
//   - New tweets prepended at the top every 10 seconds
//   - Scroll-to-bottom auto-loads more tweets
//   - Scrollable content area with scrollbar
//
// Run: go run ./cmd/feed
package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	ui "github.com/kasuganosora/ui"
	"github.com/kasuganosora/ui/core"
	"github.com/kasuganosora/ui/widget"
)

// ── Embedded tweet data ────────────────────────────────────────────────

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

// formatCount formats large numbers: 1234 → "1,234", 12345 → "1.2万"
func formatCount(n int) string {
	if n >= 10000 {
		return fmt.Sprintf("%.1f万", float64(n)/10000)
	}
	if n >= 1000 {
		return fmt.Sprintf("%d,%03d", n/1000, n%1000)
	}
	return fmt.Sprintf("%d", n)
}

// parseCreatedAt converts Twitter date format to relative time string.
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
		return fmt.Sprintf("%d分钟", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%d小时", int(d.Hours()))
	case d < 7*24*time.Hour:
		return fmt.Sprintf("%d天", int(d.Hours()/24))
	default:
		return t.Format("01-02")
	}
}

// ── HTML + CSS ─────────────────────────────────────────────────────────

// feedCSS is the shared CSS for the feed and all tweet cards.
const feedCSS = `
/* Root container: full-screen dark column */
.feed-root {
  display: flex;
  flex-direction: column;
  width: 100%;
  height: 100%;
  background: #15202B;
}

/* Header bar */
.feed-header {
  display: flex;
  flex-direction: row;
  align-items: center;
  height: 53px;
  padding-left: 16px;
  padding-right: 16px;
  background: #15202B;
  border-bottom: 1px solid #38444D;
}
.feed-header span {
  font-size: 20px;
  color: #E7E9EA;
}

/* Scrollable timeline */
.feed-timeline {
  flex-grow: 1;
  background: #15202B;
  overflow: scroll;
}

/* Single tweet row */
.tweet {
  display: flex;
  flex-direction: row;
  padding: 12px 16px;
  border-bottom: 1px solid #38444D;
  gap: 12px;
}

/* Avatar: fixed 48x48 circle with placeholder color */
.tweet-avatar {
  width: 48px;
  height: 48px;
  border-radius: 24px;
  flex-shrink: 0;
  background: #1D9BF0;
}

/* Tweet content: fills remaining space */
.tweet-body {
  display: flex;
  flex-direction: column;
  flex-grow: 1;
  gap: 2px;
}

/* Name + handle + time row */
.tweet-meta {
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

/* Tweet text */
.tweet-text {
  font-size: 15px;
  color: #E7E9EA;
}

/* Action buttons row */
.tweet-actions {
  display: flex;
  flex-direction: row;
  margin-top: 4px;
  gap: 60px;
}
.action-reply {
  font-size: 13px;
  color: #71767B;
}
.action-rt {
  font-size: 13px;
  color: #00BA7C;
}
.action-like {
  font-size: 13px;
  color: #F91880;
}
.action-share {
  font-size: 13px;
  color: #71767B;
}
`

// feedHTML is the page skeleton.
const feedHTML = `
<div class="feed-root">
  <header class="feed-header">
    <span>首页</span>
  </header>
  <main class="feed-timeline" id="timeline">
  </main>
</div>
`

// tweetHTML builds HTML for a single tweet.
func tweetHTML(name, handle, timeStr, text string, reply, rt, fav int) string {
	text = strings.ReplaceAll(text, "<", "&lt;")
	text = strings.ReplaceAll(text, ">", "&gt;")
	name = strings.ReplaceAll(name, "<", "&lt;")
	return fmt.Sprintf(`
<div class="tweet">
  <div class="tweet-avatar"></div>
  <div class="tweet-body">
    <div class="tweet-meta">
      <span class="tweet-name">%s</span>
      <span class="tweet-handle">%s · %s</span>
    </div>
    <span class="tweet-text">%s</span>
    <div class="tweet-actions">
      <span class="action-reply">%s</span>
      <span class="action-rt">%s</span>
      <span class="action-like">%s</span>
    </div>
  </div>
</div>`, name, handle, timeStr, text, formatCount(reply), formatCount(rt), formatCount(fav))
}

// loadBatchSize is how many tweets to load when reaching the bottom.
const loadBatchSize = 5

// pendingTweetData is a channel for thread-safe tweet data from goroutines.
// Only data is sent (not widgets) because widget creation touches the tree (map).
var pendingTweetData = make(chan *tweetData, 32)

func main() {
	// Parse embedded tweet data
	if err := json.Unmarshal(tweetsJSON, &allTweets); err != nil {
		fmt.Fprintf(os.Stderr, "parse tweets: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("[Feed] Loaded %d real tweets from embedded data\n", len(allTweets))

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

	doc := app.LoadHTML(feedHTML + "<style>" + feedCSS + "</style>")
	tree := app.Tree()
	cfg := app.Config()

	// Get the timeline container
	timeline := doc.QueryByID("timeline")
	if timeline == nil {
		fmt.Fprintln(os.Stderr, "timeline not found")
		os.Exit(1)
	}
	contentWidget := timeline.(*widget.Content)

	type feedContainer interface {
		widget.Widget
		AppendChild(widget.Widget)
		PrependChild(widget.Widget)
	}
	container := timeline.(feedContainer)

	// Helper: create a tweet widget from data using HTML+CSS
	makeTweet := func(td *tweetData, timeOverride string) widget.Widget {
		t := timeOverride
		if t == "" {
			t = parseCreatedAt(td.Created)
		}
		text := strings.ReplaceAll(td.Text, "\n", " ")
		html := tweetHTML(td.Name, td.Handle, t, text, td.Reply, td.RT, td.Fav)
		tweetRoot := ui.LoadHTMLWithCSS(tree, cfg, html, feedCSS)
		// tweetRoot is a Div wrapping the .tweet div; get the actual .tweet child
		if len(tweetRoot.Children()) > 0 {
			return tweetRoot.Children()[0]
		}
		return tweetRoot
	}

	// Seed initial tweets
	for i := range min(15, len(allTweets)) {
		tw := makeTweet(&allTweets[i], "")
		container.AppendChild(tw)
	}

	// Auto-load more when scrolled to bottom (debounced)
	var lastLoad time.Time
	var onNearBottom func()
	onNearBottom = func() {
		if time.Since(lastLoad) < 500*time.Millisecond {
			return
		}
		lastLoad = time.Now()
		for range loadBatchSize {
			td := &allTweets[rand.Intn(len(allTweets))]
			tw := makeTweet(td, "")
			container.AppendChild(tw)
		}
		tree.MarkDirty(tree.Root())
		fmt.Printf("[Feed] Loaded %d more (total: %d)\n", loadBatchSize, len(contentWidget.Children()))
	}

	// Goroutine: send tweet data every 10s via channel (thread-safe).
	// Widget creation must happen on main thread since it touches tree maps.
	go func() {
		for {
			time.Sleep(10 * time.Second)
			td := &allTweets[rand.Intn(len(allTweets))]
			pendingTweetData <- td
		}
	}()

	// Layout: drain pending tweets + CSSLayout engine
	app.SetOnLayout(func(tree *core.Tree, root widget.Widget, w, h float32) {
		// Drain pending tweet data and create widgets on main thread (avoids concurrent map access)
		for {
			select {
			case td := <-pendingTweetData:
				tw := makeTweet(td, "刚刚")
				container.PrependChild(tw)
				fmt.Println("[Feed] New tweet at top")
			default:
				goto done
			}
		}
	done:
		// Use CSS layout engine for the entire widget tree (with text measurer)
		ui.CSSLayout(tree, root, w, h, cfg)

		// Check if scrolled near bottom for auto-load
		scrollY := contentWidget.ScrollY()
		contentH := contentWidget.ContentHeight()
		bounds := contentWidget.Bounds()
		maxScroll := contentH - bounds.Height
		if maxScroll > 0 && scrollY >= maxScroll-100 && onNearBottom != nil {
			onNearBottom()
		}
	})

	_ = cfg // used by makeTweet closure

	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
