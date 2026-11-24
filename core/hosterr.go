package core

import "errors"

// Sentinel errors for the interface-only hosts. They do not report an OS status.
var (
	ErrClosed     = errors.New("core: closed")
	ErrPermission = errors.New("core: permission")
	ErrCancelled  = errors.New("core: cancelled")
	ErrRange      = errors.New("core: range")
	ErrState      = errors.New("core: state")
)
