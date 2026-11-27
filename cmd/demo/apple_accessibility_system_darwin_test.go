//go:build darwin

package main

import (
	"testing"

	ui "github.com/kasuganosora/ui"
	"github.com/kasuganosora/ui/widget"
)

// TestNewAppAdoptsAccessibilityOnRealGUIStartup is the end-to-end proof that
// the startup wiring runs inside a genuine AppKit + Metal App, not only in a
// unit-test fake. Environment variables are controlled so the expectation is
// deterministic whatever the host's System Settings happen to be.
//
// Single test body: App.Init locks the OS thread and bootstraps Cocoa, so
// this must not be split into t.Run subtests.
func TestNewAppAdoptsAccessibilityOnRealGUIStartup(t *testing.T) {
	t.Setenv(widget.ReducedMotionEnv, "1")
	t.Setenv(widget.ReducedTransparencyEnv, "0") // explicit off must win
	t.Setenv(widget.MoreContrastEnv, "yes")

	app, err := ui.NewApp(ui.AppOptions{
		Title:  "GoUI Accessibility Startup",
		Width:  320,
		Height: 240,
	})
	if err != nil {
		t.Fatalf("NewApp on real GUI: %v", err)
	}
	defer app.Destroy()

	got := app.Accessibility()
	t.Logf("real GUI startup accessibility: %+v", got)
	if !got.ReducedMotion {
		t.Error("explicit reduced-motion was not applied at startup")
	}
	if got.ReducedTransparency {
		t.Error("explicit reduced-transparency=false did not override the environment default")
	}
	if !got.MoreContrast {
		t.Error("more-contrast was not applied at startup")
	}
	if app.TextScale() != 1 {
		t.Errorf("startup unexpectedly changed Dynamic Type: %g", app.TextScale())
	}
}
