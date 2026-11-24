package main

import (
	"os"
	"runtime"
	"testing"
)

// See the matching TestMain in cmd/demo for the full rationale: a single P
// keeps Cocoa on the main thread. GUI test packages must additionally be run
// with `go test -p 1` on macOS to avoid WindowServer contention.
func TestMain(m *testing.M) {
	runtime.GOMAXPROCS(1)
	os.Exit(m.Run())
}
