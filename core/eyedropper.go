package core

import (
	"fmt"

	uimath "github.com/kasuganosora/ui/math"
)

// ColorEyeDropper is the screen-pick contract.
// Request does not capture the screen. A permitted fake stays in the picking
// state until the host calls Deliver or the caller calls Cancel.
type ColorEyeDropper interface {
	Request() error
	Cancel() error
	Picking() bool
	Color() (uimath.Color, bool)
	Err() error
}

// NewFakeDropper builds a dropper. permit is the only way a pick can start.
// held is stored and is not returned; the host must Deliver a color.
func NewFakeDropper(permit bool, held uimath.Color) *FakeDropper {
	return &FakeDropper{permit: permit, held: held}
}

// FakeDropper is the headless pick. Deliver is the host side of the contract.
type FakeDropper struct {
	permit  bool
	held    uimath.Color
	color   uimath.Color
	ready   bool
	picking bool
	err     error
}

func (f *FakeDropper) Request() error {
	if f.picking {
		return fmt.Errorf("%w: busy", ErrState)
	}
	f.ready = false
	f.color = uimath.Color{}
	if !f.permit {
		f.err = ErrPermission
		f.picking = false
		return ErrPermission
	}
	f.err = nil
	f.picking = true
	return nil
}

// Deliver completes an in-flight pick with a host-supplied color.
func (f *FakeDropper) Deliver(c uimath.Color) error {
	if !f.picking {
		return fmt.Errorf("%w: no request", ErrState)
	}
	f.picking = false
	f.color = c
	f.ready = true
	f.err = nil
	_ = f.held
	return nil
}

func (f *FakeDropper) Cancel() error {
	if !f.picking {
		return fmt.Errorf("%w: idle", ErrState)
	}
	f.picking = false
	f.ready = false
	f.color = uimath.Color{}
	f.err = ErrCancelled
	return nil
}

func (f *FakeDropper) Picking() bool { return f.picking }

func (f *FakeDropper) Color() (uimath.Color, bool) {
	if !f.ready {
		return uimath.Color{}, false
	}
	return f.color, true
}

func (f *FakeDropper) Err() error { return f.err }
