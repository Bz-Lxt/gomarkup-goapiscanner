# Design Spec — Midnight Radar

面向安全工程师的上线前扫描控制台。视觉记忆点是**雷达扫描线扫过墨水色网格**，而不是通用仪表盘紫渐变。

## Palette

| Token | Hex | 用途 |
|---|---|---|
| ink | `#070B14` | 页面底 |
| panel | `#101826` | 卡片 |
| line | `#1C2A3F` | 分割 |
| phosphor | `#3EE0C5` | 主操作 / 进行中 |
| crit | `#E84855` | 严重 |
| high | `#E88C30` | 高危 |
| mid | `#E6C443` | 中危 |
| low | `#7AA2C8` | 低危 |
| text | `#E7EEF6` | 正文 |
| mute | `#8AA0B5` | 辅助 |

## Typography

- Display: **Oxanium**（标题、严重度、数字）
- Body: **IBM Plex Sans**
- Mono: **IBM Plex Mono**（实时打印流）

## Components

- 全宽主区（禁止页面 `max-w`）
- 严重度胶囊：左边一截色条 + 中文标签
- 缺陷树：接口节点可折叠，叶子带证据摘要
- 打印流：等宽、底部吸附、新行闪一下 phosphor
- 表单错误写在字段下方，保存前统一校验
- 弹窗用自定义 Dialog，禁止 `alert`

## Motion

- 背景雷达 8s 旋转一次
- 进度条 phosphor 流光
- 树节点展开 180ms ease
