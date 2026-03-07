package atlas

import "testing"

func TestPackerSingle(t *testing.T) {
	p := newShelfPacker(256, 256)
	r, ok := p.Pack(10, 10)
	if !ok {
		t.Fatal("Pack failed")
	}
	if r.X != 0 || r.Y != 0 || r.Width != 10 || r.Height != 10 {
		t.Errorf("unexpected region: %+v", r)
	}
}

func TestPackerMultipleOnShelf(t *testing.T) {
	p := newShelfPacker(256, 256)
	r1, _ := p.Pack(10, 10)
	r2, _ := p.Pack(10, 10)

	// Second should be next to first (with padding)
	if r2.X <= r1.X {
		t.Errorf("r2.X (%d) should be > r1.X (%d)", r2.X, r1.X)
	}
	if r2.Y != r1.Y {
		t.Errorf("r2.Y (%d) should equal r1.Y (%d)", r2.Y, r1.Y)
	}
}

func TestPackerNewShelf(t *testing.T) {
	p := newShelfPacker(30, 256)
	// Pack three 10-wide items in a 30-wide atlas
	// First two fit on shelf (10+1+10+1 = 22 <= 30), third needs new shelf
	p.Pack(10, 10)
	p.Pack(10, 10)
	r3, ok := p.Pack(10, 10)
	if !ok {
		t.Fatal("Pack failed for third item")
	}
	if r3.Y == 0 {
		t.Error("third item should be on a new shelf (Y > 0)")
	}
}

func TestPackerFull(t *testing.T) {
	p := newShelfPacker(20, 20)
	// Fill it up
	p.Pack(18, 18)
	_, ok := p.Pack(10, 10)
	if ok {
		t.Error("expected Pack to fail when atlas is full")
	}
}

func TestPackerTooLarge(t *testing.T) {
	p := newShelfPacker(64, 64)
	_, ok := p.Pack(100, 10)
	if ok {
		t.Error("expected failure for item wider than atlas")
	}
}

func TestPackerReset(t *testing.T) {
	p := newShelfPacker(64, 64)
	p.Pack(60, 60)
	_, ok := p.Pack(10, 10)
	if ok {
		t.Error("should be full before reset")
	}
	p.Reset()
	_, ok = p.Pack(10, 10)
	if !ok {
		t.Error("should succeed after reset")
	}
}

func TestPackerOccupancy(t *testing.T) {
	p := newShelfPacker(100, 100)
	if p.Occupancy() != 0 {
		t.Errorf("expected 0 occupancy, got %f", p.Occupancy())
	}
	p.Pack(10, 10)
	if p.Occupancy() <= 0 {
		t.Error("expected positive occupancy after Pack")
	}
}
