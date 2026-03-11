//go:build linux && !android

// Package linux implements the platform.Platform interface for Linux using X11.
// This file contains IME (Input Method Editor) support stubs.
//
// Full XIM (X Input Method) integration requires:
//   - Opening an XIM connection: XOpenIM(dpy, NULL, NULL, NULL)
//   - Creating an XIC per window: XCreateIC(xim, XNInputStyle, XIMPreeditNothing|XIMStatusNothing, XNClientWindow, win, nil)
//   - Calling XSetICValues with XNSpotLocation before each composition session
//   - Filtering events through XFilterEvent before dispatching
//
// Full Fcitx/IBus support additionally requires D-Bus integration.
// These are left as future enhancements.
package linux
