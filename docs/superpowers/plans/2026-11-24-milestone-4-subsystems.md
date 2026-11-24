# 里程碑 4 计划（尚未执行）

> 已批准的 19 个子系统。本增量不实现。每个子系统单独开工，不在这里写实现细节。顺序、依赖和许可证默认以已批准设计表 3 为准。

**Goal:** 在里程碑 3 之后，按表 3 的序号做干净重写。平台真接入停在接口和假实现，直到环境里有对应 SDK。

**Architecture:** 仍画进 `render.CommandBuffer`。调用方提供字体、文件系统回调或字节流。不嵌入 NewCMMath、libghostty、limu、chroma、goldmark、untie。

**Tech Stack:** 现有模块。表 3 第 16–19 行只增加接口，不链接系统 WebView 或播放器。

## Global Constraints

- 产品默认 D1–D5 继续有效。
- 假宿主测试通过不等于平台接入完成。账本必须写成 platform-interface-only。
- 无 Co-authored-by。

## 任务

- [x] 1 Camera。已批准的有边界能力：平移、绕光标缩放、世界坐标命中、视口裁剪不泄漏、图层顺序不依赖 Draw 顺序。不是场景图。
- [x] 2 plot。已批准的有边界能力：柱、折线中点、面积、散点只命中点、饼、坐标轴。树图仍是单层切片，不是 squarified。桑基仍是直线加权边，不是丝带。
- [x] 3 DataTable。列移动、列宽、左右固定列、可见行窗口。表头和单元格是字形。整表裁在边界内。没有列组和单元格控件。
- [x] 4 TextView。手写 Markdown 子集，标题、列表、代码和链接都是字形。没有表格、图片和强调语法。
- [x] 5 flow。Camera 上的分层布局。节点矩形随缩放变化，名字是字形。自环不改层。成环节点共用一个附加层。
- [x] 6 filepicker。注入的只读目录列举。不调用操作系统对话框，也不改文件。
- [x] 7 commitgraph。调用方给出的提交行和泳道。不是 git 客户端。
- [x] 8 diagram。`box` 和 `->` 的小布局。不是 Mermaid，不依赖 limu。
- [x] 9 editor。符文光标、选区、滚动、命中和字形。关键字只着色 func/return/if/var/for/type/package。没有 chroma。
- [x] 10 diff。并排行差异，沿用编辑器的行/列位置。不是词级差异，也不是三方合并。
- [x] 11 richeditor。符文选区和标记。粗体是同一字面再画一次，代码和链接是字形。不是 ProseMirror/Tiptap。
- [x] 12 formula。调用方字体上的小语法：四则、分数、上下标、根号。不是 TeX，不嵌入 NewCMMath。
- [x] 13 chat。作者在正文上方，按测量宽度换行。发送先记下草稿，成功后只追加这份文本。注入的发送接口，不开网络。
- [x] 14 Dock 序列化。现有平铺边缘布局的编号、几何和缩放。不是嵌套分割树。内容指针按编号保留，Draw 不画内容。点击页签不切换。
- [x] 15 terminal。UTF-8 字节流和按键路由。不解释 VT，不启动 shell 或 PTY，不带 libghostty。
- [x] 16 WebView。接口和假宿主。不打开系统 WebView，也不抓取 URL。
- [x] 17 Video。接口和假宿主。时钟推进，不解码。
- [x] 18 Shader。CPU 上的小程序。不编译到图形 API。
- [x] 19 ColorEyeDropper。接口和假宿主。不采样屏幕。

按已批准设计表 3 的序号继续。1 完成才能做 5；9 完成才能做 10。第 12 行停在上面的小语法，不嵌入数学字体。第 15 行停在字节流和按键路由。第 16、17、19 行停在接口和假宿主。第 18 行是 CPU 求值，不是 GPU 着色器。

验收命令沿用已批准设计里程碑 4 一节。`go test ./...` 的 Vulkan 视觉失败单独报告。

## 绘制顺序（审查第 9 轮）

图层顺序不能依赖 `Draw` 的调用顺序。`OrderedForPaint` 在规范化时做三件事：按命令发出顺序记下每条绘制当时的绝对剪刀（`PushClip` 已经把嵌套裁剪相交过），丢掉作为屏障的裁剪命令，再按 `ZOrder` 稳定排序。图层基数已经加在 `ZOrder` 上（`rank * 10000`），所以更高的模态排在后面，同层里控件自己的 z 和相同 z 的发出顺序保持不变。排序后再按相邻命令的剪刀差异重建绝对 `CmdClip`。六个后端都消费这个结果，并且把剪刀设成绝对值。软件光栅验证像素。OpenGL、Vulkan、Metal、DX9、DX11 没有帧图像。

测试：`TestOrderedForPaintKeepsClipScopes` 保留邻居和嵌套剪刀；`TestOrderedForPaintOrdersLayersAcrossClips` 让更高 z 穿过更低的裁剪；`TestCommandBufferNestedClipFollowsDraw` 检查相交后的内层矩形；`TestHigherLayerCoversEarlierDraw` 用更高对话框先画、更低层的 Camera、ImageCompare、Command 后画，软件像素和宿主 `HitTest` 都落在更高层。
