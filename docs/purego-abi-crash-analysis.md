# purego 调用 macOS 原生 API 的 ABI 崩溃分析与修复

> 本文档记录 feed 应用启动时连续遭遇的多次 SIGSEGV 崩溃的排查全流程。  
> 适用场景：通过 purego (zero-CGO) 在 Go 中调用 Objective-C / Metal API。

---

## 背景

本项目使用 [purego](https://github.com/ebitengine/purego) v0.10.0 实现 zero-CGO 调用 macOS Cocoa 和 Metal API。purego 提供两种调用方式：

| 方式 | 用途 | 底层 |
|------|------|------|
| `purego.SyscallN(fn, args...)` | 快速的"原始"调用，所有参数当作 `uintptr` | 直接 CALL，参数按寄存器+栈序排列 |
| `purego.RegisterFunc(&typedFn, fnPtr)` | 注册带类型签名的 Go 函数 | 根据签名自动处理 float→FP 寄存器、struct→栈布局 |

启动 feed 时连续遇到 4 个 SIGSEGV，分属 3 个不同的根因。以下按发现顺序逐一记录。

---

## 崩溃 1: NSWindow 初始化 — `objc_msgSend` vs `objc_msgSend_stret`

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
4. **验证 ABI 差异**：
   - `objc_msgSend`: 返回值在 rax/rdx（≤ 16 字节）
   - `objc_msgSend_stret`: 调用者把返回缓冲区指针放在 `rdi`，`self` → `rsi`，`_cmd` → `rdx`

### 根因

**amd64 架构下，返回 > 16 字节结构体的 Objective-C 方法必须通过 `objc_msgSend_stret` 调用。** 直接用 `objc_msgSend` 会导致寄存器含义错位——runtime 把返回缓冲区指针当成 `self`、把 `self` 当成 `_cmd`——最终写入非法地址。

> arm64 没有此问题——Apple 的 ARM64 ABI 统一用 `objc_msgSend`，大结构体通过 `x8` 寄存器传返回缓冲区。

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

同时把 `initWindowWithContentRect` 改为使用 NSWindow 的 designated initializer `initWithContentRect:styleMask:backing:defer:`，通过 typed wrapper 传递标量参数。

### 经验规则

```
返回结构体大小  amd64 入口             arm64 入口
≤ 16 字节       objc_msgSend           objc_msgSend
> 16 字节       objc_msgSend_stret     objc_msgSend（x8 传返回指针）
```

---

## 崩溃 2: Metal 层 `setWantsLayer:` — NativeHandle 返回了 NSWindow

### 现象

```
NSInvalidArgumentException: -[NSWindow setWantsLayer:]
```

### 排查路径

1. Metal 后端初始化时调用 `setWantsLayer:` 设置 layer-backed view。
2. `NativeHandle()` 返回的是 NSWindow 对象，但 Metal 期望的是 NSView（content view）。
3. NSWindow 不直接响应 `setWantsLayer:`。

### 根因

`NativeHandle()` 返回了 `nswindow`（NSWindow），而 Metal/Vulkan 后端需要的是 `nsview`（NSView，即 content view）。

### 修复

`platform/darwin/window.go`:

```go
func (w *Window) NativeHandle() uintptr {
    return w.nsview  // Metal/Vulkan 需要 NSView
}

// 新增：Cocoa 层操作使用此方法
func (w *Window) NativeWindowHandle() uintptr {
    return w.nswindow
}
```

---

## 崩溃 3: Metal drawable size 为 0 — DPI 查询时序错误

### 现象

```
CAMetalLayer ignoring invalid setDrawableSize width=0 height=0
```

### 排查路径

1. `setDrawableSize` 参数 `fbWidth`/`fbHeight` 为 0。
2. 追溯发现 `fbWidth = int(float64(width) * dpiScale)`，而 `dpiScale = 0`。
3. `queryDPI()` 在 `nswindow` 创建之前调用——此时 `nswindow = 0`，`backingScaleFactor` 返回 0。

### 修复

将 DPI 查询移到窗口创建之后，并添加保护：

```go
// 窗口创建后查询 DPI
w.dpiScale = queryDPI(w.nswindow)
if w.dpiScale <= 0 {
    w.dpiScale = 1.0  // 安全回退
}
```

---

## 崩溃 4: Metal `replaceRegion` — 大结构体参数的 ABI 错误 ⭐

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
// Metal API 中的 C 定义
typedef struct {
    MTLOrigin origin;  // {x, y, z}     — 3 × NSUInteger
    MTLSize   size;    // {w, h, depth}  — 3 × NSUInteger
} MTLRegion;           // 总计 48 字节 (6 × 8)
```

方法签名：
```objc
- (void)replaceRegion:(MTLRegion)region
         mipmapLevel:(NSUInteger)level
           withBytes:(const void *)pixelBytes
         bytesPerRow:(NSUInteger)bytesPerRow;
```

#### 第三步：分析 System V AMD64 ABI 对结构体参数的处理

System V AMD64 ABI 的核心规则：

| 结构体大小 | 分类 | 传递方式 |
|-----------|------|---------|
| ≤ 16 字节 (2 eightbytes) | INTEGER / SSE | 拆散到寄存器 |
| 17–32 字节 (3–4 eightbytes) | MEMORY | 栈上连续传递 |
| > 32 字节 | MEMORY | 栈上连续传递 |

MTLRegion = 48 字节 → **MEMORY class** → 必须在栈上作为连续内存块传递。

#### 第四步：对比 SyscallN 和正确 ABI 的参数布局

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
栈[1] = 1 (size.depth)
栈[2] = 0 (mipmapLevel)
栈[3] = dataPtr
栈[4] = bytesPerRow
```

**关键差异**：SyscallN 把 MTLRegion 的 6 个字段拆散成独立的 `uintptr` 参数，占据了 DX/CX/R8/R9 和部分栈空间；而 ABI 要求 MTLRegion 整体在栈上，寄存器留给后面的标量参数。方法实现从 DX 读 `mipmapLevel`，但实际拿到的是 `origin.x = 0`；从 CX 读 `withBytes` 指针，但拿到的是 `origin.y = 0` → **解引用空指针 → SIGSEGV**。

#### 第五步：验证 purego.RegisterFunc 的处理

阅读 purego 源码 `struct_amd64.go`：

```go
func addStruct(...) (...) {
    pM := postMerger(t)
    if pM != nil {
        // struct > 16 bytes → placeOnStack = true
        return placeStack(...)
    }
    // ...
}
```

`RegisterFunc` 在注册时检测参数类型——如果是 > 16 字节的结构体，会调用 `placeStack` 把结构体作为连续内存块放到栈上，**完全符合 ABI 要求**。

### 根因

**`purego.SyscallN` 只接受 `uintptr` 参数，按整数寄存器/栈位逐个填充。** 对于 > 16 字节的结构体参数（MTLRegion = 48 bytes, MTLScissorRect = 32 bytes），这种"拆散"传递方式违反了 System V AMD64 ABI——导致被调用函数从错误的位置读取参数。

### 修复

`render/metal/metal.go` — 定义结构体和 typed wrapper：

```go
type MTLRegion struct {
    Origin MTLOrigin  // {X, Y, Z uintptr}
    Size   MTLSize    // {Width, Height, Depth uintptr}
}

type MTLScissorRect struct {
    X, Y, Width, Height uintptr
}

var (
    msgSendReplaceRegion  func(obj, sel uintptr, region MTLRegion, level, bytes, bytesPerRow uintptr)
    msgSendSetScissorRect func(obj, sel uintptr, rect MTLScissorRect)
)

// init() 中注册
purego.RegisterFunc(&msgSendReplaceRegion, objc_msgSend)
purego.RegisterFunc(&msgSendSetScissorRect, objc_msgSend)
```

`render/metal/backend.go` — 调用处改为：

```go
region := MTLRegion{
    Origin: MTLOrigin{X: 0, Y: 0, Z: 0},
    Size:   MTLSize{Width: uintptr(desc.Width), Height: uintptr(desc.Height), Depth: 1},
}
msgSendReplaceRegion(tex, selReplaceRegion, region,
    0,  // mipmapLevel
    uintptr(unsafe.Pointer(&desc.Data[0])),
    uintptr(bytesPerRow),
)
```

---

## 崩溃 5: 事件循环 `processPendingEvents` — NSString 类型错误

### 现象

```
SIGSEGV at addr=0x6e7552464380 (非零野指针)
goroutine: platform/darwin.(*Platform).processPendingEvents → darwin.msgSend
```

### 排查路径

1. 崩溃在 `nextEventMatchingMask:untilDate:inMode:dequeue:` 调用处。
2. 检查 `inMode:` 参数——代码传的是 `cstring("kCFRunLoopDefaultMode")`，即裸 C 字符串指针。
3. 但 `inMode:` 参数类型是 `NSRunLoopMode`（即 `NSString*`）。Cocoa runtime 会在这个指针上调用 Objective-C 方法（如 `isEqualToString:`），C 字符串指针不是合法的 ObjC 对象 → 野指针崩溃。
4. 同时发现 `distantPast` 用 `[[NSDate alloc] init]` 创建，得到的是**当前时间**而非远古时间。

### 根因

两个错误叠加：
1. `inMode:` 需要 `NSString*`，但传了 `*byte`（C 字符串指针）
2. `distantPast` 创建方式错误（`alloc+init` ≠ `distantPast` 类方法）

### 修复

`platform/darwin/platform.go`:

```go
// Init 中加载全局常量
if foundation, err := purego.Dlopen(".../Foundation.framework/Foundation", ...); err == nil {
    if sym, err := purego.Dlsym(foundation, "NSDefaultRunLoopMode"); err == nil {
        p.defaultRunLoopMode = *(*id)(unsafe.Pointer(sym))  // 解引用得到 NSString*
    }
}

// processPendingEvents 中
distantPast := msgSend(id(objcClass("NSDate")), objcSelector("distantPast"))
event := msgSend(p.app, selNextEvent,
    0xFFFFFFFF,               // NSEventMaskAny
    uintptr(distantPast),     // [NSDate distantPast]
    uintptr(p.defaultRunLoopMode), // NSDefaultRunLoopMode (NSString*)
    1,                         // dequeue
)
```

---

## 总结：purego 调用原生 API 的避坑清单

### 何时必须用 `RegisterFunc` 而非 `SyscallN`

| 场景 | 平台 | 原因 |
|------|------|------|
| 参数含 float/double | arm64 | float 走 FP 寄存器 (d0-d7)，SyscallN 只填整数寄存器 |
| 参数含 > 16 字节结构体 | amd64 | ABI 要求栈上连续传递，SyscallN 会拆散到寄存器 |
| 返回 > 16 字节结构体 | amd64 | 需要 `objc_msgSend_stret`（隐式返回指针在 rdi） |
| 参数含 > 16 字节结构体 | arm64 | 大结构体也走内存（通过 x8 或栈），RegisterFunc 自动处理 |

### 常见 Metal 结构体大小速查

| 结构体 | 字段 | 大小 | SyscallN 安全？ |
|--------|------|------|:---------------:|
| MTLOrigin | 3 × uint64 | 24 B | ❌ |
| MTLSize | 3 × uint64 | 24 B | ❌ |
| MTLRegion | Origin + Size | 48 B | ❌ |
| MTLScissorRect | 4 × uint64 | 32 B | ❌ |
| MTLViewport | 4 × double + 2 × double | 48 B | ❌ |
| MTLClearColor | 4 × double | 32 B | ❌ (float) |
| CGSize | 2 × double | 16 B | ❌ (float) |
| CGRect / NSRect | 4 × double | 32 B | ❌ (float) |

> **经验法则**：只要参数/返回值涉及结构体或浮点数，一律用 `RegisterFunc`。`SyscallN` 仅适用于纯整数标量参数+返回值的调用。

### Cocoa 类型陷阱

| 误用 | 正确做法 |
|------|---------|
| `cstring("NSDefaultRunLoopMode")` 作为 `inMode:` | 通过 `Dlsym` 加载 `NSDefaultRunLoopMode` 全局 `NSString*` |
| `[[NSDate alloc] init]` 作为 `distantPast` | `[NSDate distantPast]` 类方法 |
| `NativeHandle()` 返回 NSWindow | Metal/Vulkan 需要 NSView（content view） |
| `alloc + init` 创建 NSWindow | 必须用 designated initializer `initWithContentRect:styleMask:backing:defer:` |

---

## 修改文件清单

| 文件 | 变更内容 |
|------|---------|
| `platform/darwin/darwin.go` | 加载 `objc_msgSend_stret`；注册 `msgSendRectReturn` 等 typed wrapper 到 stret 入口 |
| `platform/darwin/window.go` | NSWindow 改用 designated initializer；`NativeHandle()` 返回 nsview；DPI 查询移到窗口创建后 |
| `platform/darwin/ime.go` | `contentRectForFrameRect:` 改用 typed wrapper |
| `platform/darwin/platform.go` | 通过 Dlsym 加载 `NSDefaultRunLoopMode`；`distantPast` 改用类方法 |
| `render/metal/metal.go` | 定义 `MTLRegion`/`MTLScissorRect` 结构体；注册 `msgSendReplaceRegion`/`msgSendSetScissorRect` |
| `render/metal/backend.go` | `CreateTexture`/`UpdateTexture`/`setScissor` 改用 typed wrapper |
