# 覆盖账本

对照 `2026-11-24-approved-nsgo-gap.md`。状态只有三种：`implemented+tested`、`pending`、`platform-interface-only`。假实现和未接入的原生后端不算完成。

更新日期：2026-11-24。本文件在里程碑 1–2 落地后填写结果。开始实现前，表 1、表 1 附、表 1b、表 2、表 3 全部为 `pending`。

## 平台接入（单独列出，不计入完成）

| 项 | 状态 | 原因 |
| --- | --- | --- |
| Win32 UI Automation COM | platform-interface-only | `HandleGetObject` 仍返回未处理。控件元数据不经过系统读屏 |
| Linux Fcitx/IBus | platform-interface-only | `platform/linux/ime.go` 仍是桩 |
| 操作系统文件对话框 | platform-interface-only | Upload 只接受应用注入的 `OpenPicker` |
| WebView / Video 真宿主 | platform-interface-only | 里程碑 4 才允许接口；当前无 SDK |
| 屏幕取色 | platform-interface-only | 需要操作系统授权，不在里程碑 1–2 |
| 终端 VT / libghostty | platform-interface-only | 默认不拷贝二进制，里程碑 4 才做边界 |

## 表 1（36）

实现前全部 `pending`。

## 表 1 附（11）

实现前全部 `pending`。

## 表 1b（15）

实现前全部 `pending`。无障碍属性落地后改为 `implemented+tested`。

## 表 2（20）

全部 `pending`。见里程碑 3 计划。

## 表 3（19）

全部 `pending`。其中第 15、16、17、19 行在未来实现时，默认边界完成也只记 `platform-interface-only`，直到真宿主测试在对应系统上通过。

## 表 4（15）

排除，不实现。

## 表 5

保持，不在本缺口内。
