package main

import (
	"os"
	"runtime"
	"testing"
)

// AppKit requires Cocoa calls (finishLaunching, NSWindow creation and event
// polling) on the process main thread. With more than one P, Go may schedule a
// test goroutine onto a non-main OS thread, which makes finishLaunching stall
// and NSWindow/event calls throw. A single P pins every test to the main
// thread. GUI test packages must additionally be run with `go test -p 1` on
// macOS so they do not bootstrap Cocoa concurrently (WindowServer contention).
func TestMain(m *testing.M) {
	runtime.GOMAXPROCS(1)
	os.Exit(m.Run())
}
