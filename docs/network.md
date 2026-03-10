# Network Layer

GoUI 内置网络层（`net/` 包）统一处理所有远程资源加载需求，包括 HTTP/HTTPS 图片、data: URL、代理配置和内存缓存。

---

## 快速上手

App 启动时自动创建默认 network client，开箱即用：

```go
app, _ := ui.NewApp(ui.AppOptions{...})

// HTML 中的 <img src="https://..."> 直接生效，无需额外配置
doc := app.LoadHTML(`<img src="https://example.com/avatar.png" width="48" height="48">`)
```

---

## 代理配置

```go
// 设置全局 HTTP 代理（须在加载任何图片前调用）
app.SetProxy("http://127.0.0.1:7890")   // HTTP 代理
app.SetProxy("socks5://127.0.0.1:1080") // SOCKS5 代理（需 Go 1.21+）
app.SetProxy("")                         // 取消代理，改用环境变量
```

未调用 `SetProxy` 时，client 默认读取系统环境变量 `HTTP_PROXY` / `HTTPS_PROXY`。

---

## 支持的 URL 格式

| 格式 | 示例 | 说明 |
|------|------|------|
| HTTP/HTTPS | `https://cdn.example.com/img.png` | 标准网络图片 |
| data: URL (base64) | `data:image/png;base64,iVBORw...` | 内联图片数据 |
| data: URL (text) | `data:text/plain,hello%20world` | URL 编码文本 |
| 本地文件路径 | `./assets/logo.png` | 不经过 network client |

---

## 直接使用 `net.Client`

```go
import uinet "github.com/kasuganosora/ui/net"

// 创建自定义 client
client := uinet.New(uinet.Options{
    Proxy:   "http://127.0.0.1:7890",
    Timeout: 15 * time.Second,
})

// 同步获取
data, err := client.Fetch("https://example.com/image.png")

// 异步获取（不阻塞主线程）
// 注意：cb 运行在 goroutine 中，禁止在 cb 内创建或修改 widget
client.FetchAsync("https://example.com/image.png", func(data []byte, err error) {
    if err != nil { return }
    // 仅保存数据，在 SetOnLayout 回调中处理 widget
})

// 解析 data: URL
data, err = uinet.ParseDataURL("data:image/png;base64,iVBORw0KGgo...")

// 判断是否为远程 URL
if uinet.IsRemoteURL(src) { ... }

// 缓存管理
client.Invalidate("https://example.com/image.png") // 删除单项
client.Clear()                                       // 清空全部缓存
```

---

## 自定义 NetFetcher

`widget.Config.NetClient` 接受任何实现 `widget.NetFetcher` 接口的对象：

```go
type NetFetcher interface {
    Fetch(rawURL string) ([]byte, error)
    FetchAsync(rawURL string, cb func([]byte, error))
}
```

可以用来接入其他 HTTP 库、添加请求头（如 Authorization）、实现磁盘缓存等：

```go
type AuthedFetcher struct {
    base   *uinet.Client
    token  string
}

func (f *AuthedFetcher) Fetch(url string) ([]byte, error) {
    // 自定义逻辑，如添加 Authorization header
    req, _ := http.NewRequest("GET", url, nil)
    req.Header.Set("Authorization", "Bearer "+f.token)
    ...
}

func (f *AuthedFetcher) FetchAsync(url string, cb func([]byte, error)) {
    go func() { data, err := f.Fetch(url); cb(data, err) }()
}

// 注入到 App
app.Config().NetClient = &AuthedFetcher{base: uinet.Default, token: "my-token"}
```

---

## Img widget 与网络图片

`widget.Img` / `<img>` 标签在 src 为 HTTP/HTTPS/data URL 时自动异步加载：

```html
<!-- 远程 PNG，加载期间显示灰色占位符 -->
<img src="https://cdn.example.com/avatar.png" width="40" height="40">

<!-- data URL（base64 内联图片）-->
<img src="data:image/png;base64,iVBORw0KGgo..." width="32" height="32">
```

加载状态：
- `ImgStateLoading` — 正在加载，显示灰色占位矩形
- `ImgStateLoaded`  — 加载完成，显示图片
- `ImgStateError`   — 失败，显示破图图标 + alt 文字

```go
img := widget.NewImg(tree, cfg)
img.SetSrc("https://example.com/photo.jpg")
img.OnLoad(func() { fmt.Println("loaded!") })
img.OnError(func(err error) { fmt.Println("error:", err) })
```

---

## 线程安全注意事项

`net.Client.FetchAsync` 在独立 goroutine 中运行，**回调函数中禁止操作 widget tree**（并发 map 写入会 panic）。正确模式：

```go
// ❌ 错误：在 goroutine 中创建 widget
client.FetchAsync(url, func(data []byte, err error) {
    img := widget.NewImg(tree, cfg) // PANIC: concurrent map write
})

// ✅ 正确：将数据发到 channel，在 SetOnLayout 回调（主线程）中处理
var pending = make(chan []byte, 8)

client.FetchAsync(url, func(data []byte, err error) {
    if err == nil { pending <- data }
})

app.SetOnLayout(func(tree *core.Tree, root widget.Widget, w, h float32) {
    select {
    case data := <-pending:
        img := widget.NewImg(tree, cfg)
        img.LoadFromBytes(data) // 在主线程安全操作
    default:
    }
    ui.CSSLayout(tree, root, w, h, cfg)
})
```

> `SetOnLayout` 回调在每帧渲染前由主线程调用，是修改 widget tree 的安全时机。

---

## 缓存行为

- **内存缓存**：HTTP 响应按 URL 缓存在 `sync.Map` 中，进程生命周期内有效
- **data: URL**：每次调用都解码，不缓存（通常内联使用，解码开销可忽略）
- **本地文件**：不经过 net.Client，由 `os.Open` 直接读取，不缓存
- 显式清除：`client.Invalidate(url)` 或 `client.Clear()`

当前版本不支持持久化磁盘缓存（ETag / Last-Modified / TTL）。如有需要，可通过自定义 `NetFetcher` 实现。

---

## 与 App 集成概览

```
AppOptions → ui.NewApp()
                └─ uinet.New(Options{}) → cfg.NetClient
                                              │
html.go: <img src="https://...">             │
    └─ widget.NewImg(tree, cfg)              │
           └─ img.SetSrc(url)               │
                  └─ isRemoteURL? ──yes──→ cfg.NetClient.FetchAsync()
                                             └─ goroutine: HTTP GET
                                                   └─ chan pendingImage
                                                         └─ SetOnLayout (主线程)
                                                               └─ img.loadFromBytes()
                                                                     └─ backend.CreateTexture()
```
