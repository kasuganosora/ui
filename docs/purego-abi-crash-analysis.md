# macOS 平台移植：purego ABI 问题分析与修复全记录

> 本文档完整记录了 UI 框架移植到 macOS 平台时遇到的所有问题的排查流程、根因分析和修复方案。  
> 适用场景：通过 purego (zero-CGO) 在 Go 中调用 Objective-C / Metal / FreeType API。

---

## 目录

1. [背景](#背景)
2. [问题总览](#问题总览)
3. [崩溃 1: NSWindow 初始化 — stret ABI](#崩溃-1-nswindow-初始化--stret-abi)
4. [崩溃 2: Metal 层 setWantsLayer — NativeHandle 错误](#崩溃-2-metal-层-setwantslayer--nativehandle-错误)
5. [崩溃 3: Metal drawable size 为 0 — DPI 时序](#崩溃-3-metal-drawable-size-为-0--dpi-时序)
6. [崩溃 4: Metal replaceRegion — 大结构体参数 ABI](#崩溃-4-metal-replaceregion--大结构体参数-abi)
7. [崩溃 5: 事件循环 — NSString 类型错误](#崩溃-5-事件循环--nsstring-类型错误)
8. [窗体不可见 — float64 返回值 + 前台激活 + 透明度](#窗体不可见--float64-返回值--前台激活--透明度)
9. [字体全是方块 — FreeType 加载 + 结构体偏移量](#字体全是方块--freetype-加载--结构体偏移量)
10. [按钮点击无响应 — NSPoint 返回值 ABI](#按钮点击无响应--nspoint-返回值-abi)
11. [总结：避坑清单](#总结避坑清单)
12. [修改文件清单](#修改文件清单)

---

## 背景

本项目使用 [purego](https://github.com/ebitengine/purego) v0.10.0 实现 zero-CGO 调用 macOS Cocoa 和 Metal API。purego 提供两种调用方式：

| 方式 | 用途 | 底层 |
|------|------|------|
| `purego.SyscallN(fn, args...)` | 快速的"原始"调用，所有参数当作 `uintptr` | 直接 CALL，参数按寄存器+栈序排列 |
| `purego.RegisterFunc(&typedFn, fnPtr)` | 注册带类型签名的 Go 函数 | 根据签名自动处理 float→FP 寄存器、struct→栈布局 |

移植过程中连续遇到 **7 个不同问题**，分属多个根因类别。以下按发现顺序逐一记录。

---

## 问题总览

| # | 现象 | 根因类别 | 关键文件 |
|---|------|---------|---------|
| 1 | SIGSEGV: NSWindow 初始化 | stret ABI（返回大结构体） | `darwin.go`, `window.go` |
| 2 | NSInvalidArgumentException | NativeHandle 返回类型错误 | `window.go` |
| 3 | drawable size = 0 | DPI 查询时序错误 | `window.go` |
| 4 | SIGSEGV: replaceRegion | 大结构体参数 ABI | `metal.go`, `backend.go` |
| 5 | SIGSEGV: processPendingEvents | NSString/NSDate 类型错误 | `platform.go` |
| 6 | 窗体不可见 | float64 返回值 ABI + 前台激活 + 透明度 | `darwin.go`, `window.go`, `backend.go` |
| 7 | 字体全是方块 | FreeType 未加载 + 偏移量错误 | `loader_darwin.go`, `offsets_*.go` |
| 8 | 按钮点击无响应 | NSPoint 返回值 ABI | `platform.go`, `darwin.go` |

---

## 崩溃 1: NSWindow 初始化 — stret ABI

### 现象

```
SIGSEGV at addr=0x0
goroutine: platform/darwin.initWindowWithContentRect → darwin.msgSend
```

### 排查路径

1. **定位崩溃点**：`initWindowWithContentRect` 内调用 `msgSend(nswindow, selFrame, ...)`，返回 `NSRect`。
2. **确认 NSRect 大小**：`NSRect = {NSPoint{x,y}, NSSize{w,h}}` = 4 × float64 = **32 字节**。
3. **查阅 purego 源码** (`objc/objc_runtime_darwin.go`)：
   - 泛型函数 `Send[T]` 内部判断：若返回类型 > `maxRegAllocStructSize` (amd64 = 16 字节)，自动切换到 `objc_msgSend_stret`。
   - 但我们自己的 `msgSend` wrapper 始终用 `objc_msgSend`，没有区分。

### 根因

**amd64 架构下，返回 > 16 字节结构体的 ObjC 方法必须通过 `objc_msgSend_stret` 调用。** 直接用 `objc_msgSend` 会导致寄存器含义错位：runtime 把返回缓冲区指针当成 `self`、把 `self` 当成 `_cmd`——最终写入非法地址。

> arm64 没有此问题——Apple ARM64 ABI 统一用 `objc_msgSend`，大结构体通过 `x8` 寄存器传返回缓冲区。

### 修复

`platform/darwin/darwin.go`:

```go
// 加载 objc_msgSend_stret（仅 amd64）
if runtime.GOARCH == "amd64" {
    objc_msgSend_stret, _ = purego.Dlsym(objc, "objc_msgSend_stret")
}

// 返回 NSRect 的 wrapper 注册到 stret 入口
msgSendStretOrNormal := objc_msgSend
if runtime.GOARCH == "amd64" {
    msgSendStretOrNormal = objc_msgSend_stret
}
purego.RegisterFunc(&msgSendRectReturn, msgSendStretOrNormal)
```

### 经验规则

```
返回结构体大小  amd64 入口             arm64 入口
≤ 16 字节       objc_msgSend           objc_msgSend
> 16 字节       objc_msgSend_stret     objc_msgSend（x8 传返回指针）
```

---

## 崩溃 2: Metal 层 setWantsLayer — NativeHandle 错误

### 现象

```
NSInvalidArgumentException: -[NSWindow setWantsLayer:]
```

### 排查路径

1. Metal 后端初始化时调用 `setWantsLayer:` 设置 layer-backed view。
2. `NativeHandle()` 返回的是 NSWindow 对象，但 Metal 期望 NSView（content view）。
3. NSWindow 不直接响应 `setWantsLayer:`。

### 根因

`NativeHandle()` 返回了 `nswindow`（NSWindow），而 Metal/Vulkan 后端需要的是 `nsview`（NSView，即 content view）。

### 修复

`platform/darwin/window.go`:

```go
func (w *Window) NativeHandle() uintptr {
    return w.nsview  // Metal/Vulkan 需要 NSView
}

func (w *Window) NativeWindowHandle() uintptr {
    return w.nswindow  // Cocoa 层操作使用此方法
}
```

---

## 崩溃 3: Metal drawable size 为 0 — DPI 时序

### 现象

```
CAMetalLayer ignoring invalid setDrawableSize width=0 height=0
```

### 排查路径

1. `setDrawableSize` 参数 `fbWidth`/`fbHeight` 为 0。
2. 追溯发现 `fbWidth = int(float64(width) * dpiScale)`，而 `dpiScale = 0`。
3. `queryDPI()` 在 `nswindow` 创建之前调用——`nswindow = 0`，`backingScaleFactor` 返回 0。

### 修复

将 DPI 查询移到窗口创建之后，并添加保护：

```go
w.dpiScale = queryDPI(w.nswindow)
if w.dpiScale <= 0 {
    w.dpiScale = 1.0  // 安全回退
}
```

---

## 崩溃 4: Metal replaceRegion — 大结构体参数 ABI ⭐

> 这是最隐蔽、最有教育意义的一个。

### 现象

```
SIGSEGV at addr=0x0
goroutine: render/metal.(*MetalBackend).CreateTexture
  → backend.go:1063 (replaceRegion 调用处)
  → purego.SyscallN
```

### 排查路径

#### 第一步：定位崩溃代码

```go
// backend.go:1063 — 原始代码
msgSend(tex, selReplaceRegion,
    0, 0, 0,                           // origin x,y,z
    uintptr(desc.Width), uintptr(desc.Height), 1, // size w,h,d
    0,                                  // mipmapLevel
    uintptr(unsafe.Pointer(&desc.Data[0])),
    uintptr(bytesPerRow),
)
```

#### 第二步：分析 MTLRegion 结构

```c
typedef struct {
    MTLOrigin origin;  // {x, y, z}     — 3 × NSUInteger
    MTLSize   size;    // {w, h, depth}  — 3 × NSUInteger
} MTLRegion;           // 总计 48 字节 (6 × 8)
```

#### 第三步：System V AMD64 ABI 结构体参数规则

| 结构体大小 | 分类 | 传递方式 |
|-----------|------|---------|
| ≤ 16 字节 | INTEGER / SSE | 拆散到寄存器 |
| > 16 字节 | MEMORY | 栈上连续传递 |

MTLRegion = 48 字节 → **MEMORY class** → 必须在栈上连续传递。

#### 第四步：对比 SyscallN 和正确 ABI

```
SyscallN 实际传递（错误）:            正确 ABI 布局:
─────────────────────────          ─────────────────────
DI = tex (self)                    DI = tex (self)
SI = sel (_cmd)                    SI = sel (_cmd)
DX = 0 (origin.x) ← 错位!         DX = mipmapLevel
CX = 0 (origin.y)                 CX = withBytes ptr
R8 = 0 (origin.z)                 R8 = bytesPerRow
R9 = 0x400 (size.width)           栈[0..5] = MTLRegion (48 bytes, 连续)
栈[0] = 0x400 (size.height)
...
```

**关键差异**：SyscallN 把 MTLRegion 的 6 个字段拆散为独立 `uintptr` 参数，占据了寄存器和栈空间；而 ABI 要求 MTLRegion 整体在栈上，寄存器留给后续标量参数。方法实现从 DX 读 `mipmapLevel`，拿到的是 `origin.x = 0`；从 CX 读 `withBytes` 指针，拿到的是 `origin.y = 0` → 解引用空指针 → SIGSEGV。

### 修复

`render/metal/metal.go` — 定义结构体和 typed wrapper：

```go
type MTLRegion struct {
    Origin MTLOrigin
    Size   MTLSize
}

var msgSendReplaceRegion func(obj, sel uintptr, region MTLRegion, level, bytes, bytesPerRow uintptr)

purego.RegisterFunc(&msgSendReplaceRegion, objc_msgSend)
```

---

## 崩溃 5: 事件循环 — NSString 类型错误

### 现象

```
SIGSEGV at addr=0x6e7552464380 (非零野指针)
goroutine: platform/darwin.(*Platform).processPendingEvents
```

### 排查路径

1. 崩溃在 `nextEventMatchingMask:untilDate:inMode:dequeue:` 调用处。
2. `inMode:` 参数传了 `cstring("kCFRunLoopDefaultMode")`（裸 C 字符串指针），但该参数类型是 `NSRunLoopMode`（即 `NSString*`）。
3. Cocoa runtime 在此指针上调用 ObjC 方法 → 野指针崩溃。
4. 同时 `distantPast` 用 `[[NSDate alloc] init]` 创建（当前时间），应该用 `[NSDate distantPast]`。

### 根因

两个错误叠加：
1. `inMode:` 需要 `NSString*`，但传了裸 C 字符串
2. `distantPast` 创建方式错误

### 修复

```go
// Init 中通过 Dlsym 加载全局常量
if sym, err := purego.Dlsym(foundation, "NSDefaultRunLoopMode"); err == nil {
    p.defaultRunLoopMode = *(*id)(unsafe.Pointer(sym))
}

// processPendingEvents 中
distantPast := msgSend(id(objcClass("NSDate")), objcSelector("distantPast"))
event := msgSend(p.app, selNextEvent,
    0xFFFFFFFF,
    uintptr(distantPast),
    uintptr(p.defaultRunLoopMode),
    1,
)
```

---

## 窗体不可见 — float64 返回值 + 前台激活 + 透明度

> 应用启动后完全看不到窗口，由 3 个独立问题叠加导致。

### 问题 6a: backingScaleFactor 返回值 ABI

**现象**：`dpiScale` 值异常（不是预期的 2.0），导致各种尺寸计算错乱。

**根因**：`backingScaleFactor` 返回 `CGFloat`（float64），在 xmm0（amd64）/ d0（arm64）寄存器中。`SyscallN` 只读 `rax`（整数寄存器），拿到的是垃圾值。

**修复**：新增 `msgSendFloat64Return` typed wrapper：

```go
var msgSendFloat64Return func(obj id, sel SEL) float64
purego.RegisterFunc(&msgSendFloat64Return, objc_msgSend)
```

### 问题 6b: 非 bundle 应用缺少前台激活

**现象**：窗口创建成功但不可见。

**根因**：非 `.app` bundle 的 CLI 程序调用 `finishLaunching` 后没有 `activate` → 应用停留在后台 → 窗口不显示。

**修复**：`Init()` 中 `finishLaunching` 后立即调用 `activateIgnoringOtherApps:`；`SetVisible` 中增加 `orderFrontRegardless`。

### 问题 6c: 窗口全透明

**现象**：窗口已在前台但看不见内容。

**根因**：`setOpaque:NO` + clear color (0,0,0,0) → 窗口内容全透明。

**修复**：`setOpaque:YES` + 白色背景色 + Metal clear color alpha=1.0。

---

## 字体全是方块 — FreeType 加载 + 结构体偏移量

> 窗体可见后所有文字都显示为白色方块（tofu/□）。

### 问题 7a: Darwin FreeType 动态加载未实现

**现象**：所有字符渲染为白色矩形方块。

**根因**：`font/freetype/loader_darwin.go` 中 `newLoader()` 直接返回 `errNotSupported` → 回退到 mockEngine → 所有 glyph 渲染为空白矩形。

**修复**：参照 Linux 实现，使用 `purego.Dlopen` 加载 `libfreetype.dylib`：

```go
func newLoader() (*loader, error) {
    // 按顺序尝试多个路径
    paths := []string{
        "/opt/homebrew/lib/libfreetype.dylib",
        "/usr/local/lib/libfreetype.dylib",
    }
    for _, path := range paths {
        handle, err := purego.Dlopen(path, purego.RTLD_LAZY)
        if err == nil {
            return loadSymbols(handle)
        }
    }
    return nil, errNotSupported
}
```

### 问题 7b: FreeType 结构体偏移量全部错误

**现象**：FreeType 加载成功但读取结构体字段时 SIGSEGV（从错误地址读内存）。

**排查方法**：编写 CGO 小程序，用 `offsetof()` 精确获取 FreeType 2.14.2 在 macOS LP64 上的实际偏移量。

**根因**：`FT_Pos` 在 64-bit 系统上是 `long`（8 字节），但代码中的偏移量按 4 字节计算。偏移量差异巨大：

| 关键字段 | 代码中值 | 实际值 | 差异原因 |
|---------|---------|--------|---------|
| `face->units_per_EM` | 104 | **136** | 前面字段含 FT_Pos/FT_Long |
| `face->glyph` | 120 | **152** | 同上 |
| `face->size` | 128 | **160** | 同上 |
| `size_metrics->ascender` | 36 | **48** | FT_Pos = 8 bytes |
| `size_metrics->descender` | 40 | **56** | 同上 |
| `size_metrics->height` | 44 | **64** | 同上 |
| `slot->metrics.height` | 52 | **56** | FT_Pos = 8 bytes |
| `slot->bitmap` | 104 | **152** | 前面 FT_Glyph_Metrics 成员都是 8 bytes |

**修复**（两步）：

1. **修正偏移量** — `offsets_darwin.go` 和 `offsets_linux.go`（两者都是 LP64，偏移量相同）：

```go
const (
    offFaceUnitsPerEM = 136
    offFaceGlyph      = 152
    offFaceSize       = 160
)

const (
    sizeMetricsBase         = 24
    offSizeMetricsAscender  = sizeMetricsBase + 24
    offSizeMetricsDescender = sizeMetricsBase + 32
    offSizeMetricsHeight    = sizeMetricsBase + 40
)

const (
    offSlotMetrics             = 48
    offSlotMetricsWidth        = offSlotMetrics + 0   // FT_Pos (8 bytes each)
    offSlotMetricsHeight       = offSlotMetrics + 8
    offSlotMetricsHoriBearingX = offSlotMetrics + 16
    offSlotMetricsHoriBearingY = offSlotMetrics + 24
    offSlotMetricsHoriAdvance  = offSlotMetrics + 32
    offSlotBitmap              = 152
    // bitmap 内部字段是 unsigned int (4 bytes)，不受影响
)
```

2. **新增 `readLong()` 读取 FT_Pos 字段** — `helpers.go`：

```go
// readLong 读取 FT_Pos/FT_Long (LP64 上 8 字节) 并截断为 int32
func readLong(base uintptr, offset uintptr) int32 {
    v := *(*int64)(unsafe.Pointer(base + offset))
    return int32(v)
}
```

`freetype.go` 中所有读取 `FT_Pos` 字段的 `readI32` 改为 `readLong`：

```go
ascender := fix26_6ToFloat(readLong(sizePtr, offSizeMetricsAscender))
descender := fix26_6ToFloat(readLong(sizePtr, offSizeMetricsDescender))
height := fix26_6ToFloat(readLong(sizePtr, offSizeMetricsHeight))
```

### 偏移量验证方法

编写 CGO 小程序自动获取偏移量（可复用于其他平台/FreeType 版本）：

```go
/*
#include <ft2build.h>
#include FT_FREETYPE_H
#include <stddef.h>
#include <stdio.h>

void print_offsets() {
    printf("units_per_EM = %zu\n", offsetof(FT_FaceRec, units_per_EM));
    printf("glyph = %zu\n", offsetof(FT_FaceRec, glyph));
    printf("size = %zu\n", offsetof(FT_FaceRec, size));
    // ... 其他字段
}
*/
import "C"
```

编译运行：`CGO_CFLAGS="-I/usr/local/include/freetype2" CGO_LDFLAGS="-L/usr/local/lib -lfreetype" go run offset_check.go`

---

## 按钮点击无响应 — NSPoint 返回值 ABI

> 窗体正常显示、字体正常渲染后，发现所有按钮点击无反应。

### 现象

点击按钮无任何响应，添加调试日志后发现 `[CLICK]` 事件根本没有触发。

### 排查路径

1. **事件链分析**：`Platform.PollEvents` → `processPendingEvents` → `convertAndStoreEvent` → 读取 `locationInWindow` → `handleMouse` → `HitTest` → widget callback。
2. **定位关键代码**：

```go
// platform.go - 原始代码
loc := msgSend(nsevent, selLocationInWindow)
location = pointFromPtr(unsafe.Pointer(&loc))
```

3. **确认 ABI 问题**：`locationInWindow` 返回 `NSPoint`（2 × float64 = 16 字节）。在 System V ABI 中，2 个 float64 成员的结构体通过 **xmm0 + xmm1**（浮点寄存器）返回，而 `SyscallN` 只读 `rax`（整数寄存器）→ 坐标值为垃圾 → hit test 命中空区域 → 事件不传递到任何 widget。

### 根因

```
locationInWindow 返回 NSPoint{x: float64, y: float64}

正确返回路径:  xmm0 = x, xmm1 = y
SyscallN 读取: rax = 垃圾值（整数寄存器，与浮点寄存器无关）
```

这与 [backingScaleFactor 的问题](#问题-6a-backingscalefactor-返回值-abi) 本质相同——**浮点返回值在 FP 寄存器中，SyscallN 只能读整数寄存器**。

### 修复

`platform/darwin/darwin.go` — 新增 `msgSendPointReturn` typed wrapper：

```go
// NSPoint (2× float64) 返回值在 xmm0 + xmm1 (amd64) / d0 + d1 (arm64)
var msgSendPointReturn func(obj id, sel SEL) NSPoint
purego.RegisterFunc(&msgSendPointReturn, objc_msgSend)
```

`platform/darwin/platform.go` — 使用 typed wrapper：

```go
// 之前（错误）
loc := msgSend(nsevent, selLocationInWindow)
location = pointFromPtr(unsafe.Pointer(&loc))

// 之后（正确）
location = msgSendPointReturn(nsevent, selLocationInWindow)
```

---

## 总结：避坑清单

### 何时必须用 RegisterFunc 而非 SyscallN

| 场景 | 平台 | 原因 |
|------|------|------|
| 返回 float/double | 全平台 | 浮点返回值在 xmm0/d0，SyscallN 只读 rax/x0 |
| 返回含 float 的结构体 (≤16B) | 全平台 | NSPoint 等走 xmm0+xmm1/d0+d1 |
| 返回 > 16 字节结构体 | amd64 | 需要 `objc_msgSend_stret`（隐式返回指针在 rdi） |
| 参数含 > 16 字节结构体 | amd64 | ABI 要求栈上连续传递，SyscallN 会拆散到寄存器 |
| 参数含 float/double | arm64 | float 参数走 FP 寄存器 (d0-d7) |

### 常见 Metal/Cocoa 结构体大小速查

| 结构体 | 字段 | 大小 | SyscallN 安全？ |
|--------|------|------|:---------------:|
| NSPoint/CGPoint | 2 × float64 | 16 B | ❌ (float 走 FP 寄存器) |
| NSSize/CGSize | 2 × float64 | 16 B | ❌ (float) |
| CGRect / NSRect | 4 × float64 | 32 B | ❌ (float + 大结构体) |
| MTLOrigin | 3 × uint64 | 24 B | ❌ (> 16B) |
| MTLSize | 3 × uint64 | 24 B | ❌ (> 16B) |
| MTLRegion | Origin + Size | 48 B | ❌ |
| MTLScissorRect | 4 × uint64 | 32 B | ❌ |
| MTLViewport | 6 × double | 48 B | ❌ |
| MTLClearColor | 4 × double | 32 B | ❌ |

> **经验法则**：只要参数或返回值涉及结构体或浮点数，一律用 `RegisterFunc`。`SyscallN` 仅适用于纯整数标量参数+纯整数标量返回值的调用。

### FreeType 偏移量陷阱

| 陷阱 | 说明 |
|------|------|
| `FT_Pos` 大小因平台而异 | LP64 (macOS/Linux 64-bit): 8 bytes; LLP64 (Windows 64-bit): 4 bytes |
| 偏移量不能手算 | 必须用 `offsetof()` 验证，结构体可能有 padding |
| 读取函数要匹配 | LP64 上 FT_Pos 用 `readLong`（8 bytes→int32），不能用 `readI32`（4 bytes） |

### Cocoa 类型陷阱

| 误用 | 正确做法 |
|------|---------|
| `cstring("NSDefaultRunLoopMode")` 作为 `inMode:` | 通过 `Dlsym` 加载 `NSDefaultRunLoopMode` 全局 `NSString*` |
| `[[NSDate alloc] init]` 作为 `distantPast` | `[NSDate distantPast]` 类方法 |
| `NativeHandle()` 返回 NSWindow | Metal/Vulkan 需要 NSView（content view） |
| 不调用 `activateIgnoringOtherApps:` | 非 bundle CLI 应用必须手动激活前台 |

---

## 修改文件清单

| 文件 | 变更内容 |
|------|---------|
| `platform/darwin/darwin.go` | 加载 `objc_msgSend_stret`；注册 `msgSendRectReturn`、`msgSendFloat64Return`、`msgSendPointReturn` 等 typed wrapper |
| `platform/darwin/window.go` | NSWindow 改用 designated initializer；`NativeHandle()` 返回 nsview；DPI 查询移到窗口创建后；`setOpaque:YES` + 白色背景 |
| `platform/darwin/platform.go` | 通过 Dlsym 加载 `NSDefaultRunLoopMode`；`distantPast` 改用类方法；`locationInWindow` 改用 `msgSendPointReturn`；启动时 `activateIgnoringOtherApps:` |
| `platform/darwin/ime.go` | `contentRectForFrameRect:` 改用 typed wrapper |
| `render/metal/metal.go` | 定义 `MTLRegion`/`MTLScissorRect` 结构体；注册 `msgSendReplaceRegion`/`msgSendSetScissorRect` |
| `render/metal/backend.go` | `CreateTexture`/`UpdateTexture`/`setScissor` 改用 typed wrapper；clear color alpha=1.0 |
| `font/freetype/loader_darwin.go` | 实现 Darwin FreeType 动态加载（purego.Dlopen） |
| `font/freetype/offsets_darwin.go` | 修正所有偏移量为 LP64 实际值（通过 CGO offsetof 验证） |
| `font/freetype/offsets_linux.go` | 同步修正 LP64 偏移量 |
| `font/freetype/helpers.go` | 新增 `readLong()` 读取 FT_Pos（LP64 上 8 字节） |
| `font/freetype/freetype.go` | FT_Pos 字段读取从 `readI32` 改为 `readLong` |
