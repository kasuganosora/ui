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
	dcomp *syscall.LazyDLL

	d3d11CreateDeviceAndSwapChain *syscall.LazyProc
	d3d11CreateDevice             *syscall.LazyProc
	createDXGIFactory2            *syscall.LazyProc
	dcompCreateDevice             *syscall.LazyProc
}

// NewLoader creates a new DX11 loader, loading d3d11.dll.
func NewLoader() (*Loader, error) {
	l := &Loader{
		d3d11: syscall.NewLazyDLL("d3d11.dll"),
		dxgi:  syscall.NewLazyDLL("dxgi.dll"),
		dcomp: syscall.NewLazyDLL("dcomp.dll"),
	}
	l.d3d11CreateDeviceAndSwapChain = l.d3d11.NewProc("D3D11CreateDeviceAndSwapChain")
	l.d3d11CreateDevice = l.d3d11.NewProc("D3D11CreateDevice")
	l.createDXGIFactory2 = l.dxgi.NewProc("CreateDXGIFactory2")
	l.dcompCreateDevice = l.dcomp.NewProc("DCompositionCreateDevice2")

	if err := l.d3d11.Load(); err != nil {
		return nil, fmt.Errorf("dx11: failed to load d3d11.dll: %w", err)
	}
	return l, nil
}
