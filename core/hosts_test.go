package core

import (
	"errors"
	"os"
	"strings"
	"testing"

	uimath "github.com/kasuganosora/ui/math"
)

func TestFakeWebViewHistoryAndLifecycle(t *testing.T) {
	host := &FakeWebViewHost{Scripts: map[string]string{"1+1": "2"}}
	view, err := host.Open()
	if err != nil || host.Opened != 1 {
		t.Fatalf("open %v opened %d", err, host.Opened)
	}
	var urls []string
	view.OnNavigate(func(u string) { urls = append(urls, u) })
	if err := view.LoadURL(""); err == nil || view.URL() != "" {
		t.Fatal("empty url was accepted")
	}
	if err := view.LoadURL("https://example.test/a"); err != nil {
		t.Fatal(err)
	}
	if err := view.LoadURL("https://example.test/b"); err != nil {
		t.Fatal(err)
	}
	if view.URL() != "https://example.test/b" || !view.CanBack() || view.CanForward() {
		t.Fatalf("url %q back %v forward %v", view.URL(), view.CanBack(), view.CanForward())
	}
	if err := view.Back(); err != nil || view.URL() != "https://example.test/a" || !view.CanForward() {
		t.Fatalf("back %v url %q", err, view.URL())
	}
	if err := view.Back(); err == nil || view.URL() != "https://example.test/a" {
		t.Fatal("back past the start changed the page")
	}
	if err := view.Forward(); err != nil || view.URL() != "https://example.test/b" {
		t.Fatalf("forward %v %q", err, view.URL())
	}
	if err := view.LoadURL("https://example.test/c"); err != nil {
		t.Fatal(err)
	}
	if err := view.Forward(); err == nil {
		t.Fatal("forward into a truncated history")
	}
	if err := view.Reload(); err != nil || view.URL() != "https://example.test/c" {
		t.Fatalf("reload %v %q", err, view.URL())
	}
	if err := view.LoadHTML("<title>Hi</title><p>body</p>"); err != nil || view.Title() != "Hi" || !strings.Contains(view.HTML(), "body") {
		t.Fatalf("html title %q err %v", view.Title(), err)
	}
	if _, err := view.Eval("1+1"); err != nil {
		t.Fatal(err)
	}
	got, err := view.Eval("1+1")
	if err != nil || got != "2" {
		t.Fatalf("eval %q %v", got, err)
	}
	if _, err := view.Eval("nope"); err == nil || view.URL() == "" {
		t.Fatal("unknown script changed the document")
	}
	if err := view.Close(); err != nil || !view.Closed() {
		t.Fatal(err)
	}
	if err := view.LoadURL("https://example.test/d"); !errors.Is(err, ErrClosed) {
		t.Fatalf("load after close %v", err)
	}
	if err := view.Close(); !errors.Is(err, ErrClosed) {
		t.Fatalf("second close %v", err)
	}
	if len(urls) < 5 || urls[0] != "https://example.test/a" {
		t.Fatalf("navigations %v", urls)
	}
}

func TestFakeVideoPlaybackBounds(t *testing.T) {
	host := &FakeVideoHost{}
	if _, err := host.Open("", 2); err == nil {
		t.Fatal("empty source opened")
	}
	if _, err := host.Open("clip", 0); err == nil {
		t.Fatal("zero duration opened")
	}
	if host.Opened != 0 {
		t.Fatalf("failed opens counted %d", host.Opened)
	}
	v, err := host.Open("clip", 2)
	if err != nil || host.Opened != 1 || v.Duration() != 2 || v.Playing() {
		t.Fatal(err)
	}
	var ended int
	v.OnEnded(func() { ended++ })
	if err := v.Play(); err != nil {
		t.Fatal(err)
	}
	if err := v.Tick(0.5); err != nil || v.Position() != 0.5 {
		t.Fatalf("tick %v pos %v", err, v.Position())
	}
	if err := v.Pause(); err != nil || v.Playing() {
		t.Fatal(err)
	}
	if err := v.Tick(1); err != nil || v.Position() != 0.5 {
		t.Fatal("paused video advanced")
	}
	if err := v.Seek(3); err == nil || v.Position() != 0.5 {
		t.Fatal("seek past the end changed position")
	}
	if err := v.Seek(1); err != nil || v.Position() != 1 {
		t.Fatalf("seek %v %v", err, v.Position())
	}
	if err := v.SetVolume(1.5); err == nil || v.Volume() != 1 {
		t.Fatalf("volume %v %v", err, v.Volume())
	}
	if err := v.Play(); err != nil {
		t.Fatal(err)
	}
	if err := v.Tick(5); err != nil || v.Position() != 2 || v.Playing() || ended != 1 {
		t.Fatalf("end pos %v playing %v ended %d err %v", v.Position(), v.Playing(), ended, err)
	}
	if err := v.Tick(1); err != nil || ended != 1 {
		t.Fatal("ended fired again")
	}
	if err := v.Close(); err != nil {
		t.Fatal(err)
	}
	if err := v.Play(); !errors.Is(err, ErrClosed) {
		t.Fatalf("play after close %v", err)
	}
}

func TestShaderEvalAndFailedCompile(t *testing.T) {
	s, err := CompileShader("out = vec4(uv.x, uv.y, 0, 1)\n")
	if err != nil {
		t.Fatal(err)
	}
	c, err := s.Eval(uimath.Vec2{X: 0.25, Y: 0.5}, nil)
	if err != nil || c.R != 0.25 || c.G != 0.5 || c.B != 0 || c.A != 1 {
		t.Fatalf("uv color %+v %v", c, err)
	}
	gray, err := CompileShader("out = min(uv.x, 0.2)")
	if err != nil {
		t.Fatal(err)
	}
	c, err = gray.Eval(uimath.Vec2{X: 0.8}, nil)
	if err != nil || c.R != 0.2 || c.G != 0.2 || c.B != 0.2 {
		t.Fatalf("min %+v %v", c, err)
	}
	cl, err := CompileShader("out = clamp(uv.x, 0, 0.4)")
	if err != nil {
		t.Fatal(err)
	}
	c, err = cl.Eval(uimath.Vec2{X: 0.9}, nil)
	if err != nil || c.R != 0.4 {
		t.Fatalf("clamp %+v %v", c, err)
	}
	mx, err := CompileShader("out = mix(0, 1, 0.25)")
	if err != nil {
		t.Fatal(err)
	}
	c, err = mx.Eval(uimath.Vec2{}, nil)
	if err != nil || c.R != 0.25 {
		t.Fatalf("mix %+v %v", c, err)
	}
	gain, err := CompileShader("out = uv.x + gain")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := gain.Eval(uimath.Vec2{X: 0.25}, nil); err == nil {
		t.Fatal("missing uniform was accepted")
	}
	c, err = gain.Eval(uimath.Vec2{X: 0.25}, map[string]float32{"gain": 0.5})
	if err != nil || c.R != 0.75 {
		t.Fatalf("uniform %+v %v", c, err)
	}
	if _, err := CompileShader("out ="); err == nil {
		t.Fatal("broken shader compiled")
	}
	c, err = s.Eval(uimath.Vec2{X: 1, Y: 0}, nil)
	if err != nil || c.R != 1 || c.G != 0 {
		t.Fatalf("previous shader changed %+v %v", c, err)
	}
}

func TestFakeEyeDropperDoesNotSample(t *testing.T) {
	denied := NewFakeDropper(false, uimath.RGB(1, 0, 0))
	var drop ColorEyeDropper = denied
	if err := drop.Request(); !errors.Is(err, ErrPermission) || drop.Picking() {
		t.Fatalf("denied %v picking %v", err, drop.Picking())
	}
	if _, ok := drop.Color(); ok {
		t.Fatal("denied dropper returned a color")
	}
	okd := NewFakeDropper(true, uimath.RGB(0, 0, 1))
	if err := okd.Request(); err != nil || !okd.Picking() {
		t.Fatal(err)
	}
	if err := okd.Request(); err == nil || !okd.Picking() {
		t.Fatal("second request replaced the in-flight pick")
	}
	if err := okd.Deliver(uimath.RGB(0, 1, 0)); err != nil || okd.Picking() {
		t.Fatal(err)
	}
	c, ok := okd.Color()
	if !ok || c.G != 1 || c.R != 0 || c.B != 0 {
		t.Fatalf("delivered %+v %v", c, ok)
	}
	again := NewFakeDropper(true, uimath.RGB(1, 0, 0))
	if err := again.Request(); err != nil {
		t.Fatal(err)
	}
	if err := again.Cancel(); err != nil || again.Picking() || again.Err() == nil || !errors.Is(again.Err(), ErrCancelled) {
		t.Fatalf("cancel %v err %v", err, again.Err())
	}
	if _, ok := again.Color(); ok {
		t.Fatal("cancelled pick kept a color")
	}
	if err := again.Deliver(uimath.RGB(1, 1, 1)); err == nil {
		t.Fatal("deliver after cancel")
	}
	idle := NewFakeDropper(true, uimath.RGB(0, 0, 0))
	if err := idle.Deliver(uimath.RGB(1, 1, 1)); err == nil {
		t.Fatal("deliver without a request")
	}
}

func readSource(name string) ([]byte, error) { return os.ReadFile(name) }

func TestHostSourcesStayInterfaceOnly(t *testing.T) {
	files := []string{"webview.go", "video.go", "shader.go", "eyedropper.go"}
	banned := []string{"ebiten.org", "webkit", "webview2", "ffmpeg", "libvlc", "cgdisplay", "xgetimage"}
	for _, name := range files {
		b, err := readSource(name)
		if err != nil {
			t.Fatal(err)
		}
		low := strings.ToLower(string(b))
		for _, word := range banned {
			if strings.Contains(low, strings.ToLower(word)) {
				t.Fatalf("%s contains %s", name, word)
			}
		}
	}
}
