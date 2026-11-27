//go:build darwin

package ui

import (
	"os"
	"runtime"
	"testing"
)

// A single P keeps Cocoa on the process main thread for the GUI tests in this
// package (real NSWindow creation and finishLaunching). Matches the TestMain
// used by cmd/demo, cmd/feed and platform/darwin.
func TestMain(m *testing.M) {
	runtime.GOMAXPROCS(1)
	os.Exit(m.Run())
}
