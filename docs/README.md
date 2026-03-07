# GoUI - 跨平台 UI 库设计文档

> 一个为游戏和应用程序设计的跨平台、低依赖 Go UI 库

## 文档索引

| 文档 | 说明 |
|------|------|
| [架构总览](./architecture.md) | 整体架构、分层设计、模块关系 |
| [渲染后端](./rendering-backend.md) | Vulkan / OpenGL / DirectX 后端抽象与实现 |
| [布局系统](./layout-system.md) | HTML+CSS 子集解析、布局算法、自适应系统 |
| [组件系统](./components.md) | 基础组件、复合组件、自定义组件 |
| [字体系统](./font-system.md) | FreeType 集成、SDF 渲染、字形缓存、文本排版、东亚语言支持 |
| [输入与 IME](./input-ime.md) | 键盘/鼠标/触摸输入、IME 事件处理 |
| [API 设计](./api-design.md) | 声明式 API、即时模式 API、样式 API |
| [平台适配](./platform.md) | Windows / Linux / macOS / Android / iOS 适配 |
| [游戏引擎集成](./game-integration.md) | 与游戏引擎集成的方式与最佳实践 |
| [路线图](./roadmap.md) | 开发阶段与里程碑 |
