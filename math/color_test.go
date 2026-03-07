package math

import "testing"

func TestColorHex6(t *testing.T) {
	c := ColorHex("#ff8800")
	if !c.Approx(Color{R: 1, G: 136.0 / 255, B: 0, A: 1}, 0.01) {
		t.Errorf("unexpected color: %v", c)
	}
}

func TestColorHex3(t *testing.T) {
	c := ColorHex("#f00")
	if !c.Approx(ColorRed, 0.01) {
		t.Errorf("expected red, got %v", c)
	}
}

func TestColorHex8(t *testing.T) {
	c := ColorHex("#ff000080")
	r, g, b, a := c.RGBA8()
	if r != 255 || g != 0 || b != 0 || a != 128 {
		t.Errorf("expected (255,0,0,128), got (%v,%v,%v,%v)", r, g, b, a)
	}
}

func TestColorRGBA8(t *testing.T) {
	c := ColorRGBA8(255, 128, 0, 255)
	r, g, b, a := c.RGBA8()
	if r != 255 || g != 128 || b != 0 || a != 255 {
		t.Errorf("expected (255,128,0,255), got (%v,%v,%v,%v)", r, g, b, a)
	}
}

func TestColorLerp(t *testing.T) {
	a := ColorBlack
	b := ColorWhite
	mid := a.Lerp(b, 0.5)
	if !mid.Approx(Color{R: 0.5, G: 0.5, B: 0.5, A: 1}, 0.01) {
		t.Errorf("expected gray, got %v", mid)
	}
}

func TestColorWithAlpha(t *testing.T) {
	c := ColorRed.WithAlpha(0.5)
	if c.R != 1 || c.A != 0.5 {
		t.Errorf("expected red with alpha 0.5, got %v", c)
	}
}

func TestColorHexRoundTrip(t *testing.T) {
	c := ColorHex("#3a7bff")
	hex := c.Hex()
	c2 := ColorHex(hex)
	if !c.Approx(c2, 0.01) {
		t.Errorf("hex round trip failed: %v -> %v -> %v", c, hex, c2)
	}
}

func TestColorIsTransparent(t *testing.T) {
	if !ColorTransparent.IsTransparent() {
		t.Error("transparent should be transparent")
	}
	if ColorRed.IsTransparent() {
		t.Error("red should not be transparent")
	}
}
