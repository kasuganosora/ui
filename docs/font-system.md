# 字体系统

## 概述

GoUI 的字体系统基于 FreeType 构建，提供高质量的字形光栅化和文本排版能力。FreeType 是业界标准的字体引擎，支持 TrueType、OpenType、WOFF 等主流字体格式，且跨平台可用。

## 架构

```
┌─────────────────────────────────────────────────┐
│                   Text API                       │
│  DrawText() / MeasureText() / TextLayout         │
├─────────────────────────────────────────────────┤
│               Text Shaping                       │
│     文本分段 → 字体回退 → 字形映射 → 定位          │
├─────────────────────────────────────────────────┤
│               Font Manager                       │
│    字体注册 / 查询 / 缓存 / 回退链管理             │
├─────────────────────────────────────────────────┤
│             Glyph Rasterizer                     │
│         FreeType → Bitmap / SDF 生成             │
├─────────────────────────────────────────────────┤
│              Glyph Atlas                         │
│       字形纹理图集 / LRU 缓存 / 动态扩展          │
├─────────────────────────────────────────────────┤
│            Render Backend                        │
│       文本管线（SDF Shader / Bitmap Shader）      │
└─────────────────────────────────────────────────┘
```

## FreeType 集成

### 绑定方式

使用 Go 封装 FreeType C 库，通过 CGO 调用：

```go
// font/freetype/freetype.go

// #cgo CFLAGS: -I${SRCDIR}/include
// #cgo LDFLAGS: -lfreetype
// #cgo windows LDFLAGS: -L${SRCDIR}/lib/windows -lfreetype
// #cgo linux LDFLAGS: -lfreetype
// #cgo darwin LDFLAGS: -lfreetype
import "C"
```

为降低外部依赖，将 FreeType 源码静态编译打包：

```
font/
  freetype/
    include/           # FreeType 头文件
    src/               # FreeType 源码（精简编译，只含必要模块）
    lib/
      windows/         # 预编译 .a (MinGW) / .lib (MSVC)
      linux/           # 预编译 .a
      darwin/          # 预编译 .a
    freetype.go        # CGO 绑定
    freetype_static.go # +build static — 静态编译 FreeType 源码
```

### 静态编译选项

为实现零外部依赖，可将 FreeType 源码直接通过 CGO 编译：

```go
// freetype_static.go
// +build static

// #cgo CFLAGS: -I${SRCDIR}/include -DFT2_BUILD_LIBRARY
// #cgo CFLAGS: -DFT_CONFIG_OPTION_ERROR_STRINGS
// 只编译需要的模块，减小体积
// #include "src/base/ftbase.c"
// #include "src/base/ftinit.c"
// #include "src/base/ftsystem.c"
// #include "src/base/ftglyph.c"
// #include "src/base/ftbitmap.c"
// #include "src/truetype/truetype.c"
// #include "src/type1/type1.c"
// #include "src/cff/cff.c"
// #include "src/sfnt/sfnt.c"
// #include "src/psaux/psaux.c"
// #include "src/psnames/psnames.c"
// #include "src/smooth/smooth.c"
// #include "src/autofit/autofit.c"
import "C"
```

### 也可选择纯 Go 备选

对于不想引入 CGO 的场景，提供纯 Go 的字体解析作为备选（功能受限）：

```go
// font/gofont/gofont.go
// +build !cgo

// 使用 golang.org/x/image/font/opentype 解析 TTF/OTF
// 使用 golang.org/x/image/font/sfnt 获取字形轮廓
// 自行光栅化（质量低于 FreeType，但零 CGO）
```

构建标签选择：
- `go build` — 默认使用 FreeType（CGO）
- `go build -tags nocgo` — 使用纯 Go 备选
- `go build -tags static` — 静态编译 FreeType 源码

## FreeType 封装 API

### 核心类型

```go
// FreeType 库实例
type Library struct {
    handle C.FT_Library
}

// 字体面（一个字体文件中的一个 face）
type Face struct {
    handle     C.FT_Face
    library    *Library
    fontData   []byte       // 保持字体数据引用，防止 GC 回收
    unitsPerEM int
    ascender   int
    descender  int
    lineHeight int
}

// 字形信息
type GlyphInfo struct {
    Index     uint32       // 字形索引
    Width     int          // 位图宽度
    Height    int          // 位图高度
    BearingX  int          // 水平偏移
    BearingY  int          // 垂直偏移（基线到顶部）
    Advance   int          // 水平步进
    Bitmap    []byte       // 光栅化位图数据
}

// 字形度量（不含位图，用于布局计算）
type GlyphMetrics struct {
    Index    uint32
    Advance  float32       // 水平步进（像素）
    BearingX float32
    BearingY float32
    Width    float32
    Height   float32
}
```

### 核心操作

```go
// 初始化 FreeType
func NewLibrary() (*Library, error)
func (lib *Library) Destroy()

// 加载字体
func (lib *Library) NewFace(fontData []byte, faceIndex int) (*Face, error)
func (lib *Library) NewFaceFromFile(path string, faceIndex int) (*Face, error)
func (face *Face) Destroy()

// 设置字号
func (face *Face) SetPixelSize(width, height int) error
func (face *Face) SetCharSize(charWidth, charHeight int, hDPI, vDPI int) error

// 字形操作
func (face *Face) GetCharIndex(charcode rune) uint32
func (face *Face) LoadGlyph(glyphIndex uint32, flags LoadFlags) error
func (face *Face) RenderGlyph(renderMode RenderMode) error
func (face *Face) GetGlyphInfo(charcode rune) (*GlyphInfo, error)
func (face *Face) GetGlyphMetrics(charcode rune) (*GlyphMetrics, error)

// 字距调整（Kerning）
func (face *Face) GetKerning(left, right uint32) (x, y float32, err error)
func (face *Face) HasKerning() bool

// 字体信息
func (face *Face) FamilyName() string
func (face *Face) StyleName() string
func (face *Face) Ascender() int           // 上升部（基线到顶部，设计单位）
func (face *Face) Descender() int          // 下降部（基线到底部，负值）
func (face *Face) LineHeight() int         // 行高
func (face *Face) UnitsPerEM() int         // 设计单位
func (face *Face) NumGlyphs() int
func (face *Face) IsBold() bool
func (face *Face) IsItalic() bool
func (face *Face) IsScalable() bool
```

### FreeType 渲染模式

```go
type RenderMode int

const (
    RenderNormal   RenderMode = iota  // 256 级灰度抗锯齿（默认）
    RenderLight                       // 轻量 Hinting
    RenderMono                        // 单色（无抗锯齿）
    RenderLCD                         // 亚像素渲染（LCD RGB）
    RenderLCDV                        // 亚像素渲染（LCD 垂直）
    RenderSDF                         // SDF 模式（FreeType 2.11+）
)
```

## SDF 字形生成

### 两种 SDF 生成路径

**路径 1: FreeType 原生 SDF（推荐，FreeType 2.11+）**

FreeType 2.11 起内置 SDF 光栅化器（`FT_RENDER_MODE_SDF`），直接输出 SDF 位图：

```go
func (face *Face) RenderGlyphSDF(charcode rune, spread int) (*SDFGlyph, error) {
    idx := face.GetCharIndex(charcode)
    face.LoadGlyph(idx, LoadDefault)
    // 使用 FreeType 内置 SDF 渲染
    face.RenderGlyph(RenderSDF)
    // 获取 SDF 位图
    bitmap := face.handle.glyph.bitmap
    // ...
}
```

**路径 2: 自行从轮廓生成 SDF**

对于旧版 FreeType 或需要自定义控制时，从字形轮廓自行生成 SDF：

```go
func GenerateSDF(outline []Contour, width, height, spread int) []byte {
    // 1. 从 FreeType 获取字形轮廓点（贝塞尔曲线控制点）
    // 2. 对输出位图的每个像素，计算到最近轮廓边的有符号距离
    // 3. 将距离值映射到 [0, 255]（128 = 边界，>128 = 内部，<128 = 外部）
    // 4. spread 控制距离场范围（通常 4-8 像素）
}
```

### SDF 参数

```go
type SDFConfig struct {
    Spread      int     // SDF 扩展范围（像素），默认 8
    BaseSize    int     // 基础渲染字号（像素），默认 48
    Padding     int     // 字形周围额外空间，默认 = Spread
    DownScale   int     // 降采样倍率（用于提高精度），默认 1
}
```

说明：
- **Spread**: 距离场覆盖范围，值越大描边/阴影越粗，但纹理越大
- **BaseSize**: SDF 纹理以此字号渲染，之后可缩放到任意字号而不失真
- **Padding**: 确保距离场不被裁切

### SDF 文本渲染 Shader

```glsl
// Fragment Shader (GLSL)
uniform sampler2D u_sdfTexture;
uniform float u_smoothing;      // 抗锯齿平滑量，随字号动态调整
uniform vec4 u_color;
uniform float u_threshold;      // 边界阈值，默认 0.5

// 可选：描边
uniform float u_outlineWidth;
uniform vec4 u_outlineColor;

// 可选：阴影
uniform vec2 u_shadowOffset;
uniform vec4 u_shadowColor;
uniform float u_shadowSmoothing;

void main() {
    float distance = texture(u_sdfTexture, v_texCoord).r;

    // 基础文本
    float alpha = smoothstep(u_threshold - u_smoothing,
                             u_threshold + u_smoothing,
                             distance);

    vec4 color = vec4(u_color.rgb, u_color.a * alpha);

    // 描边
    if (u_outlineWidth > 0.0) {
        float outlineAlpha = smoothstep(u_threshold - u_outlineWidth - u_smoothing,
                                        u_threshold - u_outlineWidth + u_smoothing,
                                        distance);
        vec4 outline = vec4(u_outlineColor.rgb, u_outlineColor.a * outlineAlpha);
        color = mix(outline, color, alpha);
    }

    // 阴影
    if (u_shadowColor.a > 0.0) {
        float shadowDist = texture(u_sdfTexture, v_texCoord - u_shadowOffset).r;
        float shadowAlpha = smoothstep(u_threshold - u_shadowSmoothing,
                                       u_threshold + u_shadowSmoothing,
                                       shadowDist);
        vec4 shadow = vec4(u_shadowColor.rgb, u_shadowColor.a * shadowAlpha);
        color = mix(shadow, color, color.a);
    }

    gl_FragColor = color;
}
```

## Font Manager

### 字体注册与查询

```go
type FontManager struct {
    library  *Library
    fonts    map[string]*FontFamily
    fallback []*Face           // 全局回退链
    cache    *GlyphCache
    atlas    *GlyphAtlas
}

type FontFamily struct {
    Name    string
    Regular *Face
    Bold    *Face
    Italic  *Face
    BoldItalic *Face
}

// 注册字体
func (fm *FontManager) RegisterFont(name string, data []byte, style FontStyle) error
func (fm *FontManager) RegisterFontFile(name string, path string, style FontStyle) error

// 注册回退字体（用于 CJK、Emoji 等）
func (fm *FontManager) AddFallback(data []byte) error

// 查询字体
func (fm *FontManager) GetFace(family string, style FontStyle, size float32) *Face

// 内置默认字体（嵌入二进制，确保始终可用）
//go:embed fonts/default.ttf
var defaultFontData []byte
```

### 字体回退链（Font Fallback）

当主字体缺少某个字形时（如中文字体缺少 Emoji），自动从回退链中查找：

```go
func (fm *FontManager) ResolveGlyph(char rune, family string, style FontStyle) (*Face, uint32) {
    // 1. 查主字体
    face := fm.GetFace(family, style, 0)
    if idx := face.GetCharIndex(char); idx != 0 {
        return face, idx
    }

    // 2. 遍历回退链
    for _, fallback := range fm.fallback {
        if idx := fallback.GetCharIndex(char); idx != 0 {
            return fallback, idx
        }
    }

    // 3. 使用 .notdef（缺字符号 □）
    return face, 0
}
```

典型回退链配置：
```
主字体 (Noto Sans) → CJK 字体 (Noto Sans CJK SC) → Emoji (Noto Color Emoji) → 符号 (Noto Sans Symbols)
```

### 系统字体发现

```go
// 可选功能：扫描系统字体目录
func (fm *FontManager) LoadSystemFonts() error {
    // Windows: C:\Windows\Fonts\
    // Linux:   /usr/share/fonts/, ~/.local/share/fonts/
    // macOS:   /Library/Fonts/, ~/Library/Fonts/
}

// 按名称匹配系统字体
func (fm *FontManager) FindSystemFont(family string) (string, error)
```

## 字形缓存（Glyph Cache）

### 缓存键

```go
type GlyphCacheKey struct {
    FaceID    uint32    // 字体标识
    GlyphIdx  uint32    // 字形索引
    Size      uint16    // 字号（SDF 模式下为基础字号）
    Flags     uint8     // 渲染标志（SDF/Bitmap/LCD）
}
```

### 缓存策略

```go
type GlyphCache struct {
    entries  map[GlyphCacheKey]*CachedGlyph
    lruList  *list.List                       // LRU 链表
    maxSize  int                              // 最大缓存条目数
    atlas    *GlyphAtlas                      // 关联的纹理图集
}

type CachedGlyph struct {
    Metrics  GlyphMetrics
    AtlasX   int           // 在图集中的位置
    AtlasY   int
    AtlasW   int
    AtlasH   int
    AtlasIdx int           // 图集纹理索引（多页图集）
}
```

### LRU 淘汰

CJK 字符集巨大（数万字形），不可能全部缓存，采用 LRU 淘汰：

- 每次访问字形时移到 LRU 头部
- 缓存满时从尾部淘汰
- 被淘汰的字形在图集中标记为可复用区域
- 下次需要时重新光栅化并放入图集

## 字形纹理图集（Glyph Atlas）

### 图集管理

```go
type GlyphAtlas struct {
    textures  []TextureHandle    // 图集纹理列表（支持多页）
    width     int                // 单页宽度（如 1024）
    height    int                // 单页高度（如 1024）
    packer    *RectPacker        // 矩形装箱算法
    format    TexFormat           // R8（灰度/SDF）或 RGBA8（彩色 Emoji）
    dirty     bool               // 是否需要上传到 GPU
    dirtyRect Rect               // 脏区域（最小化上传）
}
```

### 动态增长

```
初始: 512x512 单页
 ↓ 不够用
扩展: 1024x1024 单页（重新打包）
 ↓ 不够用
扩展: 2048x2048 单页
 ↓ 不够用
多页: 2048x2048 × N（添加新页）
```

### 脏区域上传

只上传发生变化的区域到 GPU，而非整个纹理：

```go
func (atlas *GlyphAtlas) Flush(backend RenderBackend) {
    if !atlas.dirty {
        return
    }
    // 只上传 dirtyRect 区域
    backend.UpdateTexture(atlas.textures[dirtyPage], dirtyData, atlas.dirtyRect)
    atlas.dirty = false
}
```

## 文本排版（Text Layout）

### 排版流程

```
输入文本 (string)
    ↓
1. Unicode 分段
   - 按脚本（Latin/CJK/Arabic/...）分段
   - 按字体回退分段（主字体 vs 回退字体）
   - 按方向分段（LTR/RTL）
    ↓
2. 字形映射
   - rune → glyph index（通过 FreeType cmap）
   - 连字处理（fi → ﬁ，可选）
    ↓
3. 字形度量查询
   - 从缓存或 FreeType 获取每个字形的 advance/bearing
   - 查询字距调整（kerning）
    ↓
4. 行断行
   - 按可用宽度断行
   - 中文逐字断行
   - 英文按空格/连字符断行
   - 处理 word-break / white-space CSS 属性
    ↓
5. 行对齐
   - text-align: left/center/right/justify
   - 计算每行的起始 X 坐标
    ↓
6. 输出 GlyphRun
   - 每个字形的精确位置 (x, y)
   - 对应的 atlas 区域
   - 颜色/样式信息
```

### GlyphRun

```go
type GlyphRun struct {
    Glyphs []PositionedGlyph
    Bounds Rect              // 整体包围盒
    Lines  []LineInfo        // 行信息
}

type PositionedGlyph struct {
    X, Y     float32         // 位置（相对于文本块左上角）
    AtlasRect Rect           // 在图集中的 UV 区域
    AtlasPage int            // 图集页索引
    Color    Color           // 字形颜色（富文本）
}

type LineInfo struct {
    StartIndex int           // 行起始字形索引
    EndIndex   int
    Y          float32       // 行基线 Y 坐标
    Width      float32       // 行宽度
    Height     float32       // 行高
}
```

### 文本度量

```go
type TextMetrics struct {
    Width      float32       // 文本总宽度
    Height     float32       // 文本总高度
    LineCount  int           // 行数
    Ascender   float32       // 最大上升部
    Descender  float32       // 最大下降部
    Lines      []LineMetrics
}

type LineMetrics struct {
    Width   float32
    Height  float32
    BaseLine float32
}

// 度量文本
func (fm *FontManager) MeasureText(text string, style TextStyle) TextMetrics
func (fm *FontManager) MeasureTextWidth(text string, style TextStyle) float32
```

## 文本样式

```go
type TextStyle struct {
    FontFamily  string
    FontSize    float32       // 像素
    FontWeight  FontWeight    // Normal(400) / Bold(700) / 100-900
    FontStyle   FontStyle     // Normal / Italic
    Color       Color
    LineHeight  float32       // 行高倍数或像素值
    LetterSpace float32       // 字间距
    TextAlign   TextAlign     // Left / Center / Right / Justify
    WhiteSpace  WhiteSpace    // Normal / NoWrap / Pre / PreWrap
    WordBreak   WordBreak     // Normal / BreakAll / BreakWord
    TextOverflow TextOverflow // Clip / Ellipsis
    MaxLines    int           // 最大行数（0 = 不限）
    // 装饰
    Underline    bool
    Strikethrough bool
    // SDF 效果
    OutlineWidth float32
    OutlineColor Color
    ShadowOffset Vec2
    ShadowColor  Color
    ShadowBlur   float32
}
```

## 东亚语言完整支持

GoUI 将东亚语言（中文/日文/韩文）作为一等公民支持，而非事后补丁。

### 设计原则

1. **CJK 优先测试** - 所有文本功能以中日韩文本作为主要测试用例，而非仅用 ASCII
2. **大字符集友好** - 架构上考虑数万字形的内存和性能问题
3. **IME 深度集成** - 输入法与文本组件紧密配合（详见 input-ime.md）
4. **排版规范遵循** - 遵循 W3C 《中文排版需求》(CLReq)、《日文排版需求》(JLReq) 中的关键规则

### CJK 断行规则

遵循 Unicode Line Breaking Algorithm (UAX #14) 并结合 CJK 特有规则：

```go
type LineBreaker struct {
    locale    Locale           // zh-CN / ja-JP / ko-KR
    strict    bool             // 严格模式（更严格的标点规则）
}

// 断行类型
type BreakOpportunity int
const (
    BreakNone       BreakOpportunity = iota  // 不可断
    BreakAllowed                              // 可断
    BreakMandatory                            // 必须断（\n）
)
```

核心规则：

- **CJK 字符间可断** - 任意两个 CJK 表意文字之间都是合法断行点
- **行首禁则** - 以下字符不能出现在行首：
  ```
  ）》」』】〗〉！，。、；：？！…—·
  ) ] } > ! , . ; : ?
  ー（日文长音）
  っゃゅょぁぃぅぇぉ（日文小假名）
  ```
- **行尾禁则** - 以下字符不能出现在行尾：
  ```
  （《「『【〖〈
  ( [ { <
  ```
- **不可分割** - 以下组合不可在中间断开：
  - 数字与单位（`100元`、`50%`）
  - 省略号 `……`
  - 破折号 `——`
  - 日文连续假名词组（基于词典，可选）

```go
func (lb *LineBreaker) FindBreaks(text []rune) []BreakOpportunity {
    breaks := make([]BreakOpportunity, len(text))
    for i := 0; i < len(text)-1; i++ {
        curr := text[i]
        next := text[i+1]

        // 行首禁则
        if isLineStartProhibited(next) {
            breaks[i] = BreakNone
            continue
        }
        // 行尾禁则
        if isLineEndProhibited(curr) {
            breaks[i] = BreakNone
            continue
        }
        // CJK 字符间默认可断
        if isCJK(curr) || isCJK(next) {
            breaks[i] = BreakAllowed
            continue
        }
        // Latin 字符按空格/连字符断
        // ...
    }
    return breaks
}
```

### 标点挤压（Punctuation Squeezing）

CJK 标点占全角宽度，相邻标点时需要挤压以避免过大空白：

```
未挤压: 他说：「你好。」她说：「再见。」
挤压后: 他说:「你好。」她说:「再见。」
        ↑ 冒号与引号之间的空白被压缩
```

```go
type PunctuationAdjustment struct {
    // 全角标点宽度调整
    HalfWidthPunctuation bool    // 是否启用标点半宽化
    SqueezeConsecutive   bool    // 相邻标点挤压
}

// 标点分类
type PunctuationClass int
const (
    PuncOpening   PunctuationClass = iota  // 开括号类：（「【
    PuncClosing                            // 闭括号类：）」】
    PuncMiddleDot                          // 中间点：·、
    PuncFullStop                           // 句末：。！？
    PuncComma                              // 逗号：，、
    PuncColon                              // 冒号：：；
)

// 挤压规则矩阵
// [前一个标点类型][后一个标点类型] → 是否挤压
var squeezeMatrix = [6][6]bool{
    //             Opening  Closing  MiddleDot  FullStop  Comma  Colon
    /* Opening */  {false,  false,   false,     false,    false, false},
    /* Closing */  {true,   true,    true,      true,     true,  true },
    /* MiddleDot */{true,   false,   false,     false,    false, false},
    /* FullStop */ {true,   true,    false,     false,    false, false},
    /* Comma */    {true,   true,    false,     false,    false, false},
    /* Colon */    {true,   false,   false,     false,    false, false},
}
```

### CJK 与 Latin 混排间距

中文与英文/数字混排时，自动插入适当间距（约 1/4 em）：

```
无间距: GoUI是一个UI库
有间距: GoUI 是一个 UI 库
              ↑          ↑ 自动插入的间距
```

```go
type MixedSpacing struct {
    Enabled     bool
    SpaceWidth  float32   // CJK-Latin 间距，默认 0.25em
}

func needsCJKLatinSpace(prev, curr rune) bool {
    prevCJK := isCJK(prev)
    currCJK := isCJK(curr)
    prevLatin := isLatinOrDigit(prev)
    currLatin := isLatinOrDigit(curr)

    // CJK → Latin 或 Latin → CJK 之间需要间距
    return (prevCJK && currLatin) || (prevLatin && currCJK)
}
```

### 竖排文本（Vertical Writing）

预留竖排支持（日文、古典中文常用）：

```go
type WritingMode int
const (
    WritingHorizontalTB WritingMode = iota  // 水平，从左到右，行从上到下（默认）
    WritingVerticalRL                        // 竖排，从上到下，列从右到左
    WritingVerticalLR                        // 竖排，从上到下，列从左到右
)

// CSS 对应
// writing-mode: horizontal-tb | vertical-rl | vertical-lr;
```

竖排时需要的特殊处理：
- 全角字符旋转 0 度（直接竖排）
- 半角 Latin/数字旋转 90 度（或使用竖排替代字形）
- 标点位置调整（句号在右上角而非右下角）
- 使用 FreeType 的 `FT_LOAD_VERTICAL_LAYOUT` 获取竖排度量

### Ruby 注音（振り仮名 / 拼音）

支持在文字上方/下方显示注音：

```go
type RubyText struct {
    Base       string     // 基础文字（如 "漢字"）
    Annotation string     // 注音（如 "かんじ" 或 "hàn zì"）
    Position   RubyPos    // 上方（默认）/ 下方
}

// HTML 语法
// <ruby>漢字<rt>かんじ</rt></ruby>
// <ruby>汉字<rt>hàn zì</rt></ruby>
```

注音排版规则：
- 注音文字字号通常为基础文字的 50%
- 注音居中对齐于基础文字上方
- 注音溢出时，可悬挂到相邻字符上方（jukugo-ruby 模式）
- 行高需要考虑注音空间

### 着重号（Emphasis Dots）

CJK 文本使用着重号强调而非斜体/粗体：

```css
/* CSS: text-emphasis 属性 */
text-emphasis: filled dot;           /* ● 实心圆（简中默认） */
text-emphasis: open sesame;          /* ﹇ 开芝麻点（日文默认） */
text-emphasis: filled circle #333;   /* 自定义样式和颜色 */
```

```go
type TextEmphasis struct {
    Style EmphasisStyle    // Dot / Circle / Triangle / Sesame / Custom
    Fill  EmphasisFill     // Filled / Open
    Color Color
    Char  rune             // 自定义强调符号
}
```

### 字体回退策略（CJK 专用）

```go
// 内置推荐回退链
var DefaultFallbackChinese = []string{
    "Noto Sans SC",           // Google 开源，覆盖全
    "Microsoft YaHei",        // Windows 中文
    "PingFang SC",            // macOS 中文
    "Source Han Sans SC",     // Adobe 开源
    "WenQuanYi Micro Hei",   // Linux 中文
}

var DefaultFallbackJapanese = []string{
    "Noto Sans JP",
    "Yu Gothic",              // Windows 日文
    "Hiragino Sans",          // macOS 日文
    "Source Han Sans JP",
}

var DefaultFallbackKorean = []string{
    "Noto Sans KR",
    "Malgun Gothic",          // Windows 韩文
    "Apple SD Gothic Neo",    // macOS 韩文
    "Source Han Sans KR",
}
```

系统字体自动检测：
```go
// 启动时自动检测系统 locale 并配置对应的回退链
func (fm *FontManager) AutoConfigureCJK() {
    locale := platform.GetSystemLocale()  // "zh-CN", "ja-JP", "ko-KR"

    switch {
    case strings.HasPrefix(locale, "zh"):
        fm.AddFallbackChain(DefaultFallbackChinese)
    case strings.HasPrefix(locale, "ja"):
        fm.AddFallbackChain(DefaultFallbackJapanese)
    case strings.HasPrefix(locale, "ko"):
        fm.AddFallbackChain(DefaultFallbackKorean)
    }

    // 始终添加 Emoji 回退
    fm.AddFallback(emojiFont)
}
```

### CJK 字形缓存优化

CJK 字符集巨大（CJK Unified Ideographs 有 9 万+字符），需要特殊策略：

```go
type CJKCacheStrategy struct {
    // 预热常用字符
    PreheatCommon    bool      // 启动时预加载常用 3000 字
    CommonCharSet    string    // "gb2312" / "jis-level1" / "ksx1001"

    // 动态加载
    PageSize         int       // 按 Unicode 页加载（每页 256 字符）
    MaxCachedPages   int       // 最大缓存页数

    // Atlas 配置
    AtlasSize        int       // 推荐 2048（CJK 场景）
    MaxAtlasPages    int       // 最大 Atlas 页数
    GlyphSize        int       // 单个 SDF 字形尺寸（推荐 48px base）
}
```

常用字符预热表：
- **中文**: GB2312 一级常用字 3755 字
- **日文**: JIS 第一水准 2965 字 + 平假名 + 片假名
- **韩文**: KS X 1001 常用 2350 字

### Unicode 范围感知

```go
func isCJK(r rune) bool {
    return unicode.Is(unicode.Han, r) ||              // CJK 统一表意文字
        unicode.Is(unicode.Hiragana, r) ||             // 平假名
        unicode.Is(unicode.Katakana, r) ||             // 片假名
        unicode.Is(unicode.Hangul, r) ||               // 韩文
        (r >= 0x3000 && r <= 0x303F) ||                // CJK 符号和标点
        (r >= 0xFF00 && r <= 0xFFEF)                   // 全角字符
}

func isCJKPunctuation(r rune) bool {
    return (r >= 0x3000 && r <= 0x303F) ||             // CJK 符号
        (r >= 0xFE30 && r <= 0xFE4F) ||                // CJK 兼容形式
        (r >= 0xFF01 && r <= 0xFF60)                   // 全角 ASCII
}

func isFullWidth(r rune) bool {
    // East Asian Width 属性为 W 或 F 的字符
    return isCJK(r) || isCJKPunctuation(r)
}
```

## 性能考虑

### 数据流优化

- FreeType 调用开销较大，所有字形数据严格缓存
- 常用 ASCII + 常用 CJK 字符可在启动时预热缓存
- 文本度量结果按内容哈希缓存（相同文本+样式不重复计算）

### 内存控制

| 场景 | 推荐 Atlas 大小 | 推荐缓存上限 |
|------|-----------------|-------------|
| 纯英文应用 | 512x512 单页 | 256 字形 |
| 中英混合应用 | 1024x1024 | 1024 字形 |
| 游戏（多语言） | 2048x2048 × 2 页 | 4096 字形 |
| 重文本应用 | 2048x2048 × 4 页 | 8192 字形 |

### 线程安全

- FreeType `Library` 对象线程安全（FreeType 2.5.1+，需编译时开启）
- `Face` 对象非线程安全，每个线程使用独立 Face 或加锁
- GlyphCache 和 GlyphAtlas 使用读写锁
- 推荐模式：主线程排版，光栅化可并行
