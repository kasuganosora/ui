# 游戏引擎集成

## 集成模式

GoUI 提供三种集成模式：

### 模式 1: 独立窗口模式（Standalone）

GoUI 自己创建窗口和渲染上下文，适合：
- 独立应用程序
- 游戏内工具（独立窗口的编辑器）

```go
app := goui.NewApp(goui.AppOptions{
    Title: "My App",
    Width: 1280, Height: 720,
    Backend: goui.BackendVulkan,
})
app.Mount(rootElement)
app.Run() // 阻塞，内部运行事件循环
```

### 模式 2: 嵌入模式（Embedded）

GoUI 使用游戏引擎提供的 GPU 上下文渲染，适合：
- 游戏内 UI（HUD、菜单、背包等）
- 需要与 3D 场景混合渲染

```go
ui := goui.NewEmbedded(goui.EmbedOptions{
    Width:  1920,
    Height: 1080,
    // 共享 Vulkan 资源
    Vulkan: &goui.VulkanContext{
        Instance:       instance,
        PhysicalDevice: physDevice,
        Device:         device,
        Queue:          queue,
        QueueFamily:    queueFamilyIdx,
        RenderPass:     renderPass,
        CommandPool:    commandPool,
    },
})

// 在游戏循环中
func GameUpdate(dt float64) {
    // 1. 转发输入事件
    for _, ev := range gameInput.Events() {
        ui.InjectEvent(convertEvent(ev))
    }

    // 2. 更新 UI
    ui.Update(dt)
}

func GameRender(cmd VkCommandBuffer) {
    // 3. 在游戏渲染的合适阶段执行 UI 渲染
    ui.RenderTo(cmd)
}
```

### 模式 3: 命令导出模式（Command Export）

GoUI 只输出渲染命令，由游戏引擎的渲染系统自行执行：

```go
ui := goui.NewHeadless(goui.HeadlessOptions{
    Width: 1920, Height: 1080,
})

func GameUpdate(dt float64) {
    ui.InjectEvent(events...)
    ui.Update(dt)

    // 获取渲染命令列表
    commands := ui.Commands()

    for _, cmd := range commands {
        switch cmd.Type {
        case goui.CmdRect:
            myEngine.DrawRect(cmd.Rect, cmd.Color, cmd.Radius)
        case goui.CmdText:
            myEngine.DrawText(cmd.Text, cmd.Position, cmd.Font, cmd.Color)
        case goui.CmdImage:
            myEngine.DrawImage(cmd.TextureID, cmd.SrcRect, cmd.DstRect)
        case goui.CmdClip:
            myEngine.SetScissor(cmd.ClipRect)
        }
    }
}
```

## 事件桥接

### 输入事件转换

游戏引擎的输入事件需要转换为 GoUI 事件：

```go
func convertEvent(gameEvent GameEvent) goui.Event {
    switch gameEvent.Type {
    case GAME_MOUSE_MOVE:
        return goui.Event{
            Type:   goui.EventMouseMove,
            MouseX: gameEvent.X,
            MouseY: gameEvent.Y,
        }
    case GAME_KEY_DOWN:
        return goui.Event{
            Type: goui.EventKeyDown,
            Key:  convertKey(gameEvent.KeyCode),
            Modifiers: goui.Modifiers{
                Ctrl:  gameEvent.Ctrl,
                Shift: gameEvent.Shift,
                Alt:   gameEvent.Alt,
            },
        }
    // ...
    }
}
```

### 输入穿透

当鼠标/触摸不在 UI 元素上时，事件应穿透到游戏层：

```go
// 检查坐标是否命中 UI 元素
if ui.HitTest(mouseX, mouseY) {
    ui.InjectEvent(mouseEvent)
    // 不传递给游戏
} else {
    // 传递给游戏逻辑
    game.HandleInput(mouseEvent)
}
```

## 典型游戏 UI 架构

```
┌──────────────────────────────────┐
│          Game Screen             │
│                                  │
│  ┌───────────────────────────┐   │
│  │    HUD Layer (固定覆盖)    │   │
│  │  ┌─────┐         ┌─────┐ │   │
│  │  │血条  │         │小地图│ │   │
│  │  └─────┘         └─────┘ │   │
│  │                           │   │
│  │              ┌──────────┐ │   │
│  │              │ 快捷栏   │ │   │
│  │              └──────────┘ │   │
│  └───────────────────────────┘   │
│                                  │
│  ┌───────────────────────────┐   │
│  │  Dialog Layer (弹出层)     │   │
│  │  ┌─────────────────────┐  │   │
│  │  │  背包 / 技能 / 商店  │  │   │
│  │  │                     │  │   │
│  │  └─────────────────────┘  │   │
│  └───────────────────────────┘   │
│                                  │
│  ┌───────────────────────────┐   │
│  │  Chat Layer               │   │
│  │  ┌─────────────────────┐  │   │
│  │  │       ChatBox       │  │   │
│  │  └─────────────────────┘  │   │
│  └───────────────────────────┘   │
└──────────────────────────────────┘
```

```go
// 游戏 UI 层级管理
gameUI := goui.NewEmbedded(opts)

// HUD 层 - 始终显示
hudLayer := goui.Layer(goui.ZIndex(100))
hudLayer.Add(
    HealthBar(player),
    Minimap(world),
    Hotbar(player.Skills),
)

// 对话层 - 按需显示
dialogLayer := goui.Layer(goui.ZIndex(200))

func OpenInventory() {
    dialogLayer.Add(InventoryPanel(player.Inventory))
}

// 聊天层
chatLayer := goui.Layer(goui.ZIndex(150))
chatLayer.Add(ChatBox(chatService))

gameUI.Mount(hudLayer, chatLayer, dialogLayer)
```

## 性能考虑

### 渲染预算

游戏 UI 通常需要在 1-2ms 内完成（60fps 下每帧 16.6ms，UI 不应占用太多）：

- 布局计算: < 0.5ms（脏区域优化）
- 命令生成: < 0.3ms
- GPU 渲染: < 1ms（批处理优化）

### 优化策略

1. **缓存静态 UI** - HUD 中不变的部分缓存渲染命令
2. **脏标记** - 只重算变化的部分
3. **合批** - 减少 draw call
4. **离屏渲染** - 复杂子窗口离屏渲染到纹理，只在内容变化时更新
5. **延迟布局** - 不可见的 UI 不参与布局计算

### 内存管理

- 对象池复用（Element、Event、RenderCommand）
- 纹理图集减少 GPU 内存碎片
- 虚拟列表只渲染可见项（万级列表 < 1MB）
