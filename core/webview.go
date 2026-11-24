package core

import (
	"fmt"
	"strings"
)

// WebView is the page surface a host can implement.
// This package does not open a system web view or fetch the URL.
type WebView interface {
	LoadURL(url string) error
	LoadHTML(html string) error
	URL() string
	Title() string
	HTML() string
	CanBack() bool
	CanForward() bool
	Back() error
	Forward() error
	Reload() error
	Eval(script string) (string, error)
	Close() error
	Closed() bool
	OnNavigate(func(url string))
}

// WebViewHost opens a WebView. The fake host records the call and does not
// create a window.
type WebViewHost interface {
	Open() (WebView, error)
}

// FakeWebViewHost is the headless stand-in. Scripts maps an exact script
// string to a result. Unknown scripts fail and do not change the page.
type FakeWebViewHost struct {
	Scripts map[string]string
	Opened  int
}

// Open returns a view with an empty history.
func (h *FakeWebViewHost) Open() (WebView, error) {
	if h == nil {
		return nil, fmt.Errorf("%w: nil host", ErrState)
	}
	h.Opened++
	return &fakeWebView{host: h, index: -1}, nil
}

type webPage struct {
	url, title, html string
}

type fakeWebView struct {
	host  *FakeWebViewHost
	pages []webPage
	index int
	onNav func(string)
	dead  bool
}

func (v *fakeWebView) LoadURL(url string) error {
	if err := v.alive(); err != nil {
		return err
	}
	if strings.TrimSpace(url) == "" {
		return fmt.Errorf("%w: empty url", ErrState)
	}
	v.push(webPage{url: url})
	return nil
}

func (v *fakeWebView) LoadHTML(html string) error {
	if err := v.alive(); err != nil {
		return err
	}
	v.push(webPage{url: "about:blank", title: htmlTitle(html), html: html})
	return nil
}

func (v *fakeWebView) URL() string {
	if p, ok := v.current(); ok {
		return p.url
	}
	return ""
}

func (v *fakeWebView) Title() string {
	if p, ok := v.current(); ok {
		return p.title
	}
	return ""
}

func (v *fakeWebView) HTML() string {
	if p, ok := v.current(); ok {
		return p.html
	}
	return ""
}

func (v *fakeWebView) CanBack() bool    { return !v.dead && v.index > 0 }
func (v *fakeWebView) CanForward() bool { return !v.dead && v.index >= 0 && v.index < len(v.pages)-1 }

func (v *fakeWebView) Back() error {
	if err := v.alive(); err != nil {
		return err
	}
	if v.index <= 0 {
		return fmt.Errorf("%w: back", ErrState)
	}
	v.index--
	v.emit()
	return nil
}

func (v *fakeWebView) Forward() error {
	if err := v.alive(); err != nil {
		return err
	}
	if v.index < 0 || v.index >= len(v.pages)-1 {
		return fmt.Errorf("%w: forward", ErrState)
	}
	v.index++
	v.emit()
	return nil
}

func (v *fakeWebView) Reload() error {
	if err := v.alive(); err != nil {
		return err
	}
	if _, ok := v.current(); !ok {
		return fmt.Errorf("%w: reload", ErrState)
	}
	v.emit()
	return nil
}

func (v *fakeWebView) Eval(script string) (string, error) {
	if err := v.alive(); err != nil {
		return "", err
	}
	if _, ok := v.current(); !ok {
		return "", fmt.Errorf("%w: no document", ErrState)
	}
	got, ok := v.host.Scripts[script]
	if !ok {
		return "", fmt.Errorf("%w: script", ErrState)
	}
	return got, nil
}

func (v *fakeWebView) Close() error {
	if v.dead {
		return ErrClosed
	}
	v.dead = true
	return nil
}

func (v *fakeWebView) Closed() bool { return v.dead }

func (v *fakeWebView) OnNavigate(fn func(string)) { v.onNav = fn }

func (v *fakeWebView) alive() error {
	if v.dead {
		return ErrClosed
	}
	return nil
}

func (v *fakeWebView) current() (webPage, bool) {
	if v.index < 0 || v.index >= len(v.pages) {
		return webPage{}, false
	}
	return v.pages[v.index], true
}

func (v *fakeWebView) push(p webPage) {
	if v.index >= 0 && v.index < len(v.pages)-1 {
		v.pages = v.pages[:v.index+1]
	}
	v.pages = append(v.pages, p)
	v.index = len(v.pages) - 1
	v.emit()
}

func (v *fakeWebView) emit() {
	if v.onNav != nil {
		v.onNav(v.URL())
	}
}

func htmlTitle(s string) string {
	low := strings.ToLower(s)
	i := strings.Index(low, "<title>")
	if i < 0 {
		return ""
	}
	rest := low[i+len("<title>"):]
	j := strings.Index(rest, "</title>")
	if j < 0 {
		return ""
	}
	return strings.TrimSpace(s[i+len("<title>") : i+len("<title>")+j])
}
