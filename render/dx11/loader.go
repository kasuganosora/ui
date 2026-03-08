//go:build windows

package dx11

import (
	"fmt"
	"syscall"
)

// Loader manages DLL loading for DX11 COM calls (zero-CGO).
type Loader struct {
	d3d11 *syscall.LazyDLL
	dxgi  *syscall.LazyDLL

	d3d11CreateDeviceAndSwapChain *syscall.LazyProc
}

// NewLoader creates a new DX11 loader, loading d3d11.dll.
func NewLoader() (*Loader, error) {
	l := &Loader{
		d3d11: syscall.NewLazyDLL("d3d11.dll"),
		dxgi:  syscall.NewLazyDLL("dxgi.dll"),
	}
	l.d3d11CreateDeviceAndSwapChain = l.d3d11.NewProc("D3D11CreateDeviceAndSwapChain")

	if err := l.d3d11.Load(); err != nil {
		return nil, fmt.Errorf("dx11: failed to load d3d11.dll: %w", err)
	}
	return l, nil
}
