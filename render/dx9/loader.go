//go:build windows

package dx9

import (
	"fmt"
	"syscall"
)

// Loader manages DLL loading for DX9 COM calls (zero-CGO).
type Loader struct {
	d3d9 *syscall.LazyDLL

	direct3DCreate9 *syscall.LazyProc
}

// NewLoader creates a new DX9 loader, loading d3d9.dll.
func NewLoader() (*Loader, error) {
	l := &Loader{
		d3d9: syscall.NewLazyDLL("d3d9.dll"),
	}
	l.direct3DCreate9 = l.d3d9.NewProc("Direct3DCreate9")

	if err := l.d3d9.Load(); err != nil {
		return nil, fmt.Errorf("dx9: failed to load d3d9.dll: %w", err)
	}
	return l, nil
}

// D3D SDK version for DX9 (32 = D3D_SDK_VERSION for DX9)
const D3D_SDK_VERSION = 32
