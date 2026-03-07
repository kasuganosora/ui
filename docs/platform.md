# 平台适配

## 平台矩阵

| 功能 | Windows | Linux | macOS | Android | iOS |
|------|---------|-------|-------|---------|-----|
| 窗口管理 | Win32 API | X11 / Wayland | Cocoa | NativeActivity | UIKit |
| 渲染 - Vulkan | MoltenVK 或原生 | 原生 | MoltenVK | 原生 | MoltenVK |
| 渲染 - OpenGL | WGL | GLX / EGL | NSOpenGL | EGL | EAGL |
| 渲染 - DirectX | 原生 | - | - | - | - |
| IME | TSF / IMM32 | IBus/Fcitx/XIM | NSTextInputClient | InputConnection | UITextInput |
| 剪贴板 | Win32 | X11 Selection | NSPasteboard | ClipboardManager | UIPasteboard |
| DPI | SetProcessDpiAwareness | Xrandr / wl_output | NSScreen backingScaleFactor | DisplayMetrics | UIScreen.scale |
| 文件对话框 | IFileDialog | zenity / portal | NSOpenPanel | Intent | UIDocumentPicker |
| 系统主题检测 | Registry | gsettings / portal | NSAppearance | UiModeManager | UITraitCollection |

## 实现策略

### 最低依赖原则

1. **纯 Go 优先** - 核心逻辑（布局、事件、组件）100% 纯 Go
2. **系统调用优于 CGO** - Windows 使用 `syscall.NewLazyDLL`，避免 CGO
3. **CGO 仅用于必需场景** - macOS Cocoa、Linux X11（也可通过 purego 避免）
4. **可选依赖** - 每个平台后端为独立模块，按需引入

### Windows 实现

```go
// 通过 syscall 直接调用 Win32 API，零 CGO
var (
    user32   = syscall.NewLazyDLL("user32.dll")
    kernel32 = syscall.NewLazyDLL("kernel32.dll")
    gdi32    = syscall.NewLazyDLL("gdi32.dll")
    imm32    = syscall.NewLazyDLL("imm32.dll")
    dwmapi   = syscall.NewLazyDLL("dwmapi.dll")
    shcore   = syscall.NewLazyDLL("shcore.dll")

    procCreateWindowExW    = user32.NewProc("CreateWindowExW")
    procDefWindowProcW     = user32.NewProc("DefWindowProcW")
    procGetMessageW        = user32.NewProc("GetMessageW")
    procDispatchMessageW   = user32.NewProc("DispatchMessageW")
    // ...
)
```

优势：
- Windows 上完全零 CGO
- 编译速度快，交叉编译简单
- 无 MinGW 依赖

### Linux 实现

**X11 路径**（通过 purego 或 CGO）：

```go
// purego 方式 - 避免 CGO
libX11, _ := purego.Dlopen("libX11.so.6", purego.RTLD_LAZY)
var XOpenDisplay func(name *byte) uintptr
purego.RegisterLibFunc(&XOpenDisplay, libX11, "XOpenDisplay")
```

**Wayland 路径**：

```go
libWayland, _ := purego.Dlopen("libwayland-client.so.0", purego.RTLD_LAZY)
// ...
```

运行时检测优先使用 Wayland，回退到 X11。

### macOS 实现

通过 purego 调用 Objective-C Runtime：

```go
// 加载 ObjC runtime
libobjc, _ := purego.Dlopen("/usr/lib/libobjc.A.dylib", purego.RTLD_LAZY)
var objc_msgSend func(obj, sel uintptr, args ...uintptr) uintptr
purego.RegisterLibFunc(&objc_msgSend, libobjc, "objc_msgSend")

// 创建 NSWindow
nsWindowClass := objc_getClass("NSWindow")
allocSel := sel_registerName("alloc")
initSel := sel_registerName("initWithContentRect:styleMask:backing:defer:")
// ...
```

### Android 实现

- 使用 Go Mobile 框架 或 NativeActivity
- JNI 调用 Android SDK
- EGL 初始化 OpenGL ES / Vulkan

### iOS 实现

- 通过 CGO + ObjC 桥接
- 或 purego 调用 ObjC Runtime

## 目录结构

```
platform/
    platform.go          // Platform 接口定义
    event.go             // 事件类型定义
    window.go            // Window 接口定义
    windows/
        platform_windows.go    // +build windows
        window_windows.go
        ime_windows.go
        clipboard_windows.go
        dpi_windows.go
        dialog_windows.go
    linux/
        platform_linux.go      // +build linux
        x11.go
        wayland.go
        ime_linux.go
    darwin/
        platform_darwin.go     // +build darwin
        cocoa.go
        ime_darwin.go
    android/
        platform_android.go    // +build android
    ios/
        platform_ios.go        // +build ios
```

## DPI 处理

### DPI 感知

```go
type DPIInfo struct {
    Scale      float64  // 系统缩放比（1.0 = 100%, 1.5 = 150%, 2.0 = 200%）
    DPIX       float64  // 水平 DPI
    DPIY       float64  // 垂直 DPI
    ScalePerMonitor bool // 是否支持多显示器不同缩放
}
```

### 逻辑像素 vs 物理像素

- 所有 API 均使用逻辑像素（与 DPI 无关）
- 布局引擎在逻辑像素空间工作
- 渲染时乘以 DPI 缩放转换为物理像素
- 字体大小自动按 DPI 缩放
- 纹理支持 @2x 高分资源

### 多显示器

```go
// 监听 DPI 变化（窗口拖到不同 DPI 的显示器）
app.OnDPIChange(func(newDPI DPIInfo) {
    // 自动重新布局和重绘
    // 纹理按需重建
})
```

## Vulkan 加载

避免编译时链接 Vulkan：

```go
// 运行时动态加载 Vulkan
var vulkanLib uintptr

func loadVulkan() error {
    switch runtime.GOOS {
    case "windows":
        vulkanLib, _ = purego.Dlopen("vulkan-1.dll", purego.RTLD_LAZY)
    case "linux":
        vulkanLib, _ = purego.Dlopen("libvulkan.so.1", purego.RTLD_LAZY)
    case "darwin":
        vulkanLib, _ = purego.Dlopen("libvulkan.dylib", purego.RTLD_LAZY)
        // 或 MoltenVK
        // vulkanLib, _ = purego.Dlopen("libMoltenVK.dylib", purego.RTLD_LAZY)
    }

    // 加载 vkGetInstanceProcAddr
    purego.RegisterLibFunc(&vkGetInstanceProcAddr, vulkanLib, "vkGetInstanceProcAddr")
    // 通过 vkGetInstanceProcAddr 加载其余函数
    return nil
}
```
