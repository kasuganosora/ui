# 渲染后端

## 后端优先级

| 后端 | 优先级 | 状态 | 说明 |
|------|--------|------|------|
| Vulkan | 主要 | 首期开发 | 跨平台首选，性能最优 |
| OpenGL 3.3+ / ES 3.0 | 次要 | 首期开发 | 兼容老硬件和 WebGL |
| DirectX 11/12 | 计划 | 二期 | Windows 原生支持 |
| Metal | 计划 | 二期 | macOS/iOS 原生支持 |
| Software | 后备 | 三期 | 无 GPU 环境兜底 |

## 渲染抽象接口

所有后端实现同一个 `RenderBackend` 接口。框架核心只依赖这个抽象，不直接引用任何图形 API。

### 后端注册与选择

```go
// 编译时通过 build tags 控制包含哪些后端
// go build -tags vulkan
// go build -tags opengl
// go build -tags dx11

func NewBackend(preferred BackendType) (RenderBackend, error) {
    // 按优先级尝试: preferred → Vulkan → OpenGL → Software
}
```

### 构建标签隔离

```
render/
  backend.go          // RenderBackend 接口定义
  command.go          // 渲染命令定义
  atlas.go            // 纹理图集管理
  vulkan/
    backend_vulkan.go      // +build vulkan
    pipeline_vulkan.go
    texture_vulkan.go
  opengl/
    backend_opengl.go      // +build opengl
    shader_opengl.go
    texture_opengl.go
  dx11/
    backend_dx11.go        // +build dx11,windows
  software/
    backend_sw.go          // 始终编译，作为兜底
```

## Vulkan 后端设计

### 管线架构

```
UI 渲染管线 (单 Pass)
├── 矩形/圆角矩形 Pipeline
│   ├── Vertex Shader: 2D 变换 + UV
│   └── Fragment Shader: 圆角 SDF + 颜色/渐变 + 边框
├── 文本 Pipeline
│   ├── Vertex Shader: 字形定位
│   └── Fragment Shader: SDF 采样 + 抗锯齿
├── 图片 Pipeline
│   ├── Vertex Shader: 2D 变换 + UV
│   └── Fragment Shader: 纹理采样 + 九宫格 + 色调
└── 自定义 Path Pipeline
    ├── Vertex Shader: 路径顶点
    └── Fragment Shader: 抗锯齿描边/填充
```

### 批处理优化

1. **排序** - 按纹理/管线/Z 序对渲染命令排序
2. **合批** - 相同管线+纹理的连续命令合并为一次 draw call
3. **实例化** - 同类型图元使用实例化绘制
4. **动态顶点缓冲** - 每帧上传所有 UI 顶点到一个大的动态 VBO

### 资源管理

- Descriptor Set 池化复用
- 纹理使用 staging buffer 异步上传
- 帧间资源通过 frame-in-flight 索引管理，避免 GPU/CPU 竞争

## OpenGL 后端设计

### 兼容性目标

- 桌面: OpenGL 3.3 Core Profile
- 移动: OpenGL ES 3.0
- Web: WebGL 2.0（通过 wasm 编译时）

### 与 Vulkan 的差异

- 使用 VAO + VBO 替代 Vulkan Buffer
- 使用 Uniform 替代 Push Constants
- 使用 FBO 实现离屏渲染
- 裁剪使用 glScissor

## 游戏引擎集成模式

当嵌入游戏引擎时，不创建自己的 GPU 设备和交换链，而是：

```go
type ExternalBackendConfig struct {
    // Vulkan 集成
    VkDevice     uintptr  // 外部提供的 VkDevice
    VkQueue      uintptr  // 外部提供的 VkQueue
    VkRenderPass uintptr  // 外部提供的 RenderPass

    // OpenGL 集成
    SharedContext uintptr  // 共享的 GL Context

    // 通用
    FramebufferWidth  int
    FramebufferHeight int
    SampleCount       int
}

func NewExternalBackend(config ExternalBackendConfig) (RenderBackend, error)
```

关键点：
- 不创建窗口和交换链
- 使用引擎提供的 Device/Context
- 在引擎的 RenderPass 中作为子 Pass 执行
- 尊重引擎的帧同步机制
- 提供回调让引擎控制渲染时机

## 着色器管理

### 预编译 SPIR-V

Vulkan 着色器预编译为 SPIR-V 字节码，嵌入到 Go 二进制中：

```go
//go:embed shaders/rect.vert.spv
var rectVertShader []byte

//go:embed shaders/rect.frag.spv
var rectFragShader []byte
```

### OpenGL 着色器

GLSL 源码嵌入，运行时编译：

```go
//go:embed shaders/rect.vert.glsl
var rectVertGLSL string

//go:embed shaders/rect.frag.glsl
var rectFragGLSL string
```

## sRGB 色彩空间与线性混合

### 问题背景

UI 渲染中文字和半透明元素的质量很大程度取决于 alpha 混合发生在哪个色彩空间。在 gamma 空间（UNORM framebuffer）中做混合会导致：

- 文字边缘偏细/偏暗，出现锯齿感
- 半透明叠加颜色不准确
- 细线条（如 1px border）可能出现断裂

正确做法是在**线性空间**中做 alpha 混合，由 GPU 自动处理 sRGB 编解码。

### 实现方案

**Vulkan：** 使用 `VK_FORMAT_B8G8R8A8_SRGB` swapchain 格式。GPU 在混合时自动将 framebuffer 内容从 sRGB 解码为线性，混合完成后再编码回 sRGB。

**DX11：** swap chain 使用 `DXGI_FORMAT_R8G8B8A8_UNORM`，但创建 RTV 时指定 `DXGI_FORMAT_R8G8B8A8_UNORM_SRGB`（跨格式 RTV）。效果与 Vulkan 一致。

> **关键要求：** DX11 必须使用 `DXGI_SWAP_EFFECT_FLIP_DISCARD` 而不是 `DXGI_SWAP_EFFECT_DISCARD`。旧的 DISCARD 模式不支持跨格式 RTV，会导致 sRGB 编解码不生效，进而出现断线和颜色错误。`FLIP_DISCARD` 需要 Windows 10+。

### 颜色输入转换

由于 framebuffer 启用了 sRGB 编码，GPU 会对 pixel shader 输出做 linear → sRGB 转换。因此 CSS 颜色值（本身就是 sRGB 空间的）需要先转为线性空间再传给 shader，否则会被双重编码导致颜色偏淡。

转换在顶点着色器中完成：

```glsl
// GLSL (Vulkan)
vec3 srgbToLinear(vec3 c) {
    return mix(c / 12.92, pow((c + 0.055) / 1.055, vec3(2.4)), step(0.04045, c));
}
fragColor = vec4(srgbToLinear(inColor.rgb), inColor.a);
```

```hlsl
// HLSL (DX11)
float3 srgbToLinear(float3 c) {
    return c <= 0.04045 ? c / 12.92 : pow((c + 0.055) / 1.055, 2.4);
}
output.color = float4(srgbToLinear(input.color.rgb), input.color.a);
```

Alpha 通道不需要转换（它不参与 sRGB 编解码）。

### DX11 采样器注意事项

字形 atlas 使用 R8_UNORM 单通道纹理，只有 1 个 mip level。采样器必须使用 `D3D11_FILTER_MIN_MAG_LINEAR_MIP_POINT`（而不是 `MIN_MAG_MIP_LINEAR`），并设置 `MaxLOD = 0`。否则 MIP_LINEAR 会尝试在 mip 0 和不存在的 mip 1 之间插值，导致 coverage 值变暗，文字发虚。

Vulkan 侧通过 `maxLod = 1.0` 在 sampler 中限制了 mip 范围，不受此问题影响。

## 纹理图集

### 字形 Atlas

- 使用 SDF 技术渲染字形
- 单通道 R8 格式，节省显存
- 动态增长：初始 512x512，按需扩展到 2048x2048
- LRU 淘汰不常用字形（对 CJK 字符尤为重要）

### 图标/图片 Atlas

- RGBA8 格式
- 使用 MaxRects 装箱算法
- 支持运行时动态添加
- 九宫格信息与 atlas 区域一同存储
