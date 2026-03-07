# 组件系统

> 组件分类参考 TDesign 体系，并扩展游戏 UI 场景组件。

## 组件总览

### 一、基础 Basic

| 组件 | 说明 | 优先级 |
|------|------|--------|
| Button 按钮 | 主要/次要/虚框/文字/图标按钮，加载态、禁用态、按钮组 | P0 |
| Icon 图标 | 内置图标集 + 自定义 SVG/图片图标，支持旋转、大小 | P0 |
| Link 链接 | 文字链接，支持下划线、禁用、前后图标 | P1 |

### 二、布局 Layout

| 组件 | 说明 | 优先级 |
|------|------|--------|
| Grid 栅格 | Row + Col 12/24 列栅格，响应式断点（xs/sm/md/lg/xl/xxl） | P0 |
| Layout 布局 | Header / Content / Footer / Aside 经典布局容器 | P0 |
| Space 间距 | 设置子元素间统一间距，水平/垂直方向，支持换行 | P0 |
| Divider 分割线 | 水平/垂直分割线，可带文字 | P1 |

### 三、导航 Navigation

| 组件 | 说明 | 优先级 |
|------|------|--------|
| Menu 导航菜单 | 顶部/侧边导航，支持多级子菜单、折叠、图标 | P1 |
| Tabs 选项卡 | 常规/卡片风格，可增删、拖拽排序、滑动 | P0 |
| Breadcrumb 面包屑 | 路径导航，支持下拉、图标、分隔符自定义 | P1 |
| Pagination 分页 | 页码/简洁/迷你模式，每页条数选择、快速跳转 | P1 |
| Steps 步骤条 | 水平/垂直步骤，支持图标、描述、点击导航 | P2 |
| Anchor 锚点 | 页内锚点导航，滚动联动高亮 | P2 |
| BackTop 回到顶部 | 页面滚动后出现的返回顶部按钮 | P2 |

### 四、输入 Input

| 组件 | 说明 | 优先级 |
|------|------|--------|
| Input 输入框 | 单行文本，前缀/后缀图标、可清除、密码切换、字数统计 | P0 |
| InputNumber 数字输入框 | 数字增减，步长、范围、精度控制 | P1 |
| TextArea 多行文本框 | 自适应高度、字数限制、可拖拽调整 | P0 |
| Select 选择器 | 单选/多选、搜索过滤、远程搜索、分组、自定义选项 | P0 |
| Cascader 级联选择器 | 多级联动选择，支持搜索、懒加载 | P2 |
| TreeSelect 树选择 | 树形结构选择，单选/多选 | P2 |
| Checkbox 多选框 | 单个/多选组、全选/半选态 | P0 |
| Radio 单选框 | 单选按钮/单选按钮组 | P0 |
| Switch 开关 | 开关切换，支持文字/图标、加载态 | P0 |
| Slider 滑块 | 水平/垂直、单滑块/范围、刻度标记、提示气泡 | P1 |
| Transfer 穿梭框 | 左右双栏选择，搜索、全选 | P2 |
| ColorPicker 颜色选择器 | 面板取色、预设色、最近使用色、Alpha 通道 | P1 |
| DatePicker 日期选择器 | 日期/周/月/季/年选择、范围选择 | P1 |
| TimePicker 时间选择器 | 时分秒选择、12/24 小时制、范围选择 | P1 |
| DateRangePicker 日期范围选择器 | 开始-结束日期区间选择，预设快捷区间 | P2 |
| Upload 上传 | 点击/拖拽上传，图片/文件列表、进度、预览 | P2 |
| AutoComplete 自动完成 | 输入时下拉建议，自定义匹配策略 | P2 |
| TagInput 标签输入框 | 输入后生成标签，可删除、数量限制 | P2 |
| RangeInput 范围输入框 | 双输入框组合（如价格区间） | P2 |
| Form 表单 | 表单容器，自动布局、校验规则、错误提示、重置 | P0 |
| FormItem 表单项 | 标签、必填标记、校验状态、帮助文字 | P0 |

### 五、数据展示 Data Display

| 组件 | 说明 | 优先级 |
|------|------|--------|
| Text 文本 | 文本渲染，富文本、省略号、可复制、可选中 | P0 |
| Tag 标签 | 尺寸/颜色/形状变体、可关闭、可选中、可添加 | P0 |
| Badge 徽标 | 数字/圆点/自定义徽标，上限值 | P1 |
| Avatar 头像 | 图片/文字/图标头像、头像组 | P1 |
| Image 图片 | 加载中/失败占位、懒加载、预览放大、图片组 | P0 |
| ImageViewer 图片预览 | 大图预览、缩放、旋转、多图切换 | P2 |
| Card 卡片 | 标题/内容/操作区，可悬浮阴影 | P1 |
| List 列表 | 基础列表、带操作、加载更多、虚拟滚动 | P1 |
| Table 表格 | 固定表头/列、排序、筛选、展开行、可编辑、虚拟滚动、列拖拽宽度 | P1 |
| Tree 树形控件 | 展开/折叠、复选框、懒加载、拖拽排序、可搜索 | P1 |
| Tooltip 文字提示 | 悬浮提示，支持多方向、主题、延迟 | P0 |
| Popover 气泡弹出框 | 悬浮/点击触发，自定义内容弹出框 | P1 |
| Popup 弹出层 | 基础弹出定位层（所有弹出类组件的底层） | P0 |
| Collapse 折叠面板 | 手风琴/多面板展开、嵌套 | P1 |
| Timeline 时间线 | 垂直/水平时间线、自定义节点 | P2 |
| Swiper 轮播 | 图片/内容轮播、自动播放、指示器 | P2 |
| Statistic 统计数值 | 数值高亮展示、前后缀、加载态 | P2 |
| Empty 空状态 | 无数据时的占位展示 | P1 |
| Skeleton 骨架屏 | 内容加载占位、动画 | P2 |
| Watermark 水印 | 全局/区域水印覆盖层 | P2 |
| Calendar 日历 | 面板式日历展示，可标记事件 | P2 |
| Comment 评论 | 评论/回复列表，嵌套结构 | P2 |
| Descriptions 描述列表 | 键值对信息展示，多列布局 | P2 |
| Rate 评分 | 星级/自定义图标评分、半星、只读 | P2 |

### 六、消息提醒 Feedback

| 组件 | 说明 | 优先级 |
|------|------|--------|
| Dialog 对话框 | 模态确认、自定义内容、可拖拽、多层嵌套 | P0 |
| Message 全局提示 | 顶部轻量反馈（成功/警告/错误/信息/加载） | P0 |
| MessageBox 消息弹出框 | 确认/提示/输入弹窗（Dialog 简化版） | P0 |
| Notification 消息通知 | 右上角通知卡片（标题+内容+操作） | P1 |
| Drawer 抽屉 | 侧边滑出面板、多层嵌套 | P1 |
| Progress 进度条 | 线性/环形、百分比、自定义颜色、动态条纹 | P0 |
| Loading 加载 | 全屏/区域加载遮罩、自定义加载图标 | P0 |
| Alert 警告提示 | 页面级别静态警告条（成功/警告/错误/信息） | P1 |
| Guide 引导 | 步骤引导弹窗，高亮目标元素 | P2 |
| Popconfirm 气泡确认框 | 轻量确认操作弹出框 | P2 |

### 七、容器 Container

| 组件 | 说明 | 优先级 |
|------|------|--------|
| ScrollView 滚动容器 | 自定义滚动条样式、横向/纵向、滚动事件 | P0 |
| VirtualList 虚拟列表 | 万级数据虚拟滚动、动态行高、滚动锚定 | P1 |
| VirtualGrid 虚拟网格 | 网格虚拟滚动、瀑布流 | P2 |
| Splitter 分割面板 | 水平/垂直拖拽分割、最小尺寸、多面板嵌套 | P1 |
| Panel 面板 | 可折叠标题栏、工具栏插槽 | P1 |
| SubWindow 子窗口 | 可拖拽/缩放/最小化/最大化/关闭/停靠 | P1 |
| ContextMenu 右键菜单 | 自定义右键菜单，嵌套子菜单、分割线、快捷键标签 | P1 |
| MenuBar 菜单栏 | 应用级顶部菜单栏 | P2 |

### 八、通用工具 Utility

| 组件 | 说明 | 优先级 |
|------|------|--------|
| DragDrop 拖放 | 通用拖放系统，拖拽排序、跨容器拖放、拖拽约束 | P1 |
| Affix 固钉 | 元素钉在可视区域内（类似 position: sticky） | P2 |
| ConfigProvider 全局配置 | 全局主题、语言、尺寸注入 | P0 |
| Portal 传送 | 将子元素渲染到 DOM 树的指定位置 | P1 |

### 九、游戏 UI Game

| 组件 | 说明 | 优先级 |
|------|------|--------|
| HUD 抬头显示层 | 固定覆盖层，不拦截未命中的输入，Z 序管理 | P0 |
| HealthBar 生命/资源条 | 血条/蓝条/经验条，分段、渐变、延迟扣血动画 | P0 |
| Hotbar 快捷栏 | 技能/物品快捷栏，键位提示、冷却遮罩、拖拽排列 | P0 |
| Inventory 背包 | 网格背包系统，物品图标+数量、拖拽移动/交换/堆叠、筛选 | P0 |
| ChatBox 聊天框 | 消息流（支持频道切换）+ 输入框、滚动锁定、消息类型着色 | P0 |
| Minimap 小地图 | 小地图容器，标记点、视锥、可缩放、可点击导航 | P1 |
| FloatingText 浮动文字 | 伤害数字/经验获取等动画浮动文字，可叠加、缓动消失 | P0 |
| CooldownMask 冷却遮罩 | 技能/物品冷却的扇形遮罩动画 | P0 |
| RadialMenu 径向菜单 | 环形选择菜单，手柄摇杆友好 | P1 |
| QuestTracker 任务追踪 | 任务列表面板，目标进度、可折叠、可点击追踪 | P1 |
| BuffBar 增益/减益栏 | Buff/Debuff 图标列表，倒计时、层数、分组 | P1 |
| Nameplate 名牌 | 角色/NPC 头顶名牌（名字+血条+称号），跟随 3D 坐标 | P1 |
| ItemTooltip 物品提示 | 游戏物品详细信息浮窗（名称、品质色、属性、描述、对比） | P0 |
| LootWindow 拾取窗口 | 掉落物品拾取/Roll 点窗口 | P2 |
| SkillTree 天赋/技能树 | 节点连线树形结构，解锁/激活/路径预览 | P2 |
| Scoreboard 计分板 | 玩家/队伍排名列表（实时更新、排序、高亮自己） | P1 |
| DialogueBox 对话框 | NPC 对话气泡/面板，打字机效果、选项分支 | P1 |
| CountdownTimer 倒计时 | 大字倒计时显示（副本/活动/复活等） | P1 |
| NotificationToast 通知浮窗 | 游戏内成就/系统消息轻量浮窗，自动消失、堆叠 | P0 |
| CurrencyDisplay 货币显示 | 金币/钻石等货币图标+数值，变化动画 | P1 |
| TeamFrame 队伍框架 | 队友头像+血条+Buff 列表，可拖拽排列 | P1 |
| TargetFrame 目标框架 | 当前目标信息面板（头像+名字+血条+Buff） | P1 |
| CastBar 施法条 | 施法/引导/读条进度条，可打断标记 | P1 |

## 优先级说明

- **P0** - 第一阶段必须实现，框架基础能力和核心游戏 UI
- **P1** - 第二阶段实现，常用组件和进阶游戏 UI
- **P2** - 第三阶段实现，高级/低频组件

## 组件接口

### 基础接口

```go
type Component interface {
    // 生命周期
    Mount()                    // 挂载到树时调用
    Unmount()                  // 从树移除时调用
    Update(dt float64)         // 每帧更新（动画等）

    // 布局
    Layout(constraints Constraints) Size
    GetChildren() []Component

    // 渲染
    Draw(buf *CommandBuffer)

    // 事件
    HandleEvent(event Event) bool  // 返回 true 表示已消费
}
```

### 状态管理

组件内部状态通过响应式信号（Signal）管理：

```go
// 创建响应式状态
count := goui.Signal(0)

// 读取
fmt.Println(count.Get())

// 修改（自动触发依赖此状态的组件更新）
count.Set(count.Get() + 1)

// 计算属性
doubled := goui.Computed(func() int {
    return count.Get() * 2
})

// 副作用
goui.Effect(func() {
    fmt.Println("count changed:", count.Get())
})
```

### 自定义组件

```go
type MyWidget struct {
    goui.BaseComponent
    // 自定义字段
    label string
    count *goui.Signal[int]
}

func NewMyWidget(label string) *MyWidget {
    w := &MyWidget{
        label: label,
        count: goui.Signal(0),
    }
    return w
}

func (w *MyWidget) Build() goui.Element {
    return goui.Div(
        goui.Style("padding: 16px; background: #333;"),
        goui.Text(w.label),
        goui.Button(
            goui.Text(fmt.Sprintf("Clicked %d times", w.count.Get())),
            goui.OnClick(func() { w.count.Set(w.count.Get() + 1) }),
        ),
    )
}
```

## 主题系统

### 主题变量

```go
type Theme struct {
    // 颜色
    Primary       Color
    PrimaryHover  Color
    PrimaryActive Color
    Secondary     Color
    Success       Color
    Warning       Color
    Danger        Color
    Info          Color

    BgPrimary     Color
    BgSecondary   Color
    BgElevated    Color

    TextPrimary   Color
    TextSecondary Color
    TextDisabled  Color

    Border        Color
    BorderHover   Color

    // 圆角
    RadiusSM float32
    RadiusMD float32
    RadiusLG float32

    // 间距
    SpaceXS float32
    SpaceSM float32
    SpaceMD float32
    SpaceLG float32
    SpaceXL float32

    // 字体
    FontFamily    string
    FontSizeXS    float32
    FontSizeSM    float32
    FontSizeMD    float32
    FontSizeLG    float32
    FontSizeXL    float32

    // 阴影
    ShadowSM BoxShadow
    ShadowMD BoxShadow
    ShadowLG BoxShadow

    // 动画
    TransitionFast     time.Duration
    TransitionNormal   time.Duration
    TransitionSlow     time.Duration
}
```

### 内置主题

- **Dark** - 深色主题（游戏场景默认）
- **Light** - 浅色主题
- **GameHUD** - 游戏 HUD 风格（半透明、发光边框）
- **Minimal** - 极简风格

### 运行时切换

```go
app.SetTheme(goui.ThemeDark)
// 或自定义
app.SetTheme(&goui.Theme{
    Primary: goui.ColorHex("#00ff88"),
    // ...
})
```
