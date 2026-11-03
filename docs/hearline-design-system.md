# Hearline 设计系统（v0.2.1 · 目标 9–10 分尺 · T7 色修订）

**进仓路径**：`docs/hearline-design-system.md`  
**门禁**：`docs/hearline-pulse-craft-gate.md`（冲突以门禁为准）  
**对照**：`miaoboai/client-pulse-vcam` + https://pulse.egoist.dev/

**标准声明**：v0.1.x ≈「能打回的七分尺」。**v0.2 起按九–十分写**：令牌完备、全交互态、对比度、动效禁用、Pulse 工艺迁移表；**真窗未勾满前产品分仍按门禁实算（当前约 3）**。

---

## 0. 语言

**Hearline Language = 字幕器物**  
字是唯一主角。四词：静、锋、诚实、本机。

---

## 1. Pulse 工艺迁移表（必须对齐的「做法」，不是抄像素）

| Pulse 做法 | Hearline 对应 |
|---|---|
| 主数字极大、标签极小 | 当前句 type.current；meta/历史降两级 |
| 工具收边栏/角，不抢主数字 | Header 文字链；听时无 Dock |
| 自绘控件，无默认套件皮 | Primary / 字段 / 菜单 / 叠层全自绘 |
| 卡间距与内边距同一节奏 | 只用 space.* / pad.* |
| 领域色克制、一物一色 | accent 只给主钮与听写点 |
| 空态/窄窗/悬停都设计过 | §4 全态表；E 交审含窄窗 720 宽 |

---

## 2. 反模式（出现 = 产品 Fail）

等权描边并排；听写音源行；常驻纪要栏；空壳灰按钮；双层影；框套框；裸默认 Button/Select/Dialog；空态双行喊话；闪烁动画；第二装饰字体。

---

## 3. 令牌（冻结）

### 3.1 色 + 对比度（WCAG AA 最低）

| Token | Hex | 对地上的正文对比 |
|---|---|---|
| ground | `#F3F4F6` | — |
| card | `#FFFFFF` | ink on card ≥ 12:1 |
| ink | `#15171C` | on ground/card ≥ 12:1 |
| muted | `#5A5F6B` | on card ≥ 4.5:1 |
| quiet | `#667085` | on card ≥ 4.5:1（小字） |
| line | `#E8EAEE` | — |
| accent | `#C2410C` | onAccent 对比 ≥ 4.5:1 |
| accent.hover | `#A83410` | |
| accent.pressed | `#8F2C0D` | |
| stage | `#15171C` | |
| onDark | `#F7F8FA` | on stage ≥ 12:1 |
| hist | `#A8ADB8` | on stage ≥ 4.5:1 |
| partial | `#C8CDD6` | on stage ≥ 3:1（大字可） |
| danger | `#D92D20` | |
| onAccent | `#FFFFFF` | |
| focus | `#2F6FED` | 仅键盘焦点环 |

强调色只给：主钮、听写点、危险。

### 3.1.1 T7 对比度冻结（2026-11-04 · 不豁免）

实测旧值未达 AA，**改令牌，不豁免**：

| Token | 旧 | 新 | 对地上 |
|---|---|---|---|
| quiet | `#8B909A` | **`#667085`** | on white≈4.97；on ground≈4.52（不用 `#6B7280`：ground 仅≈4.39） |
| accent | `#E8552B` | **`#C2410C`** | white on accent ≈5.18 |
| accent.hover | `#D44820` | **`#A83410`** | ≈6.64 |
| accent.pressed | `#B83C18` | **`#8F2C0D`** | ≈8.29 |

- Stage 上空态/弱字：**用 `hist` `#A8ADB8`**，禁止再用 `quiet`（quiet 仅浅色壳）。
- `onAccent` 保持 `#FFFFFF`。
- 实现改 `internal/ui/tokens.go` 常量；破窗测试须覆盖 quiet/accent 对比下限。


### 3.2 字族

| 平台 | 主族 | 回退 |
|---|---|---|
| macOS | SF Pro Text | system-ui |
| Windows | Segoe UI Variable | Segoe UI, system-ui |
| Linux | Inter（应用内嵌） | system-ui, sans-serif |

禁止第二装饰字体；禁止衬线正文。高级路径：`ui-monospace` 栈。

### 3.3 字阶 / 字重 / 行高

| Token | Size | Weight | LH | Color |
|---|---|---|---|---|
| type.brand | 20 | 600 | 28 | ink |
| type.sub | 12 | 400 | 16 | quiet |
| type.current | 28（长句 24） | 600 | 36（32） | onDark |
| type.currentMeta | 12 | 500 | 16 | quiet |
| type.hist | 15 | 400 | 22 | hist |
| type.partial | 24–28 | 400 italic | 32–36 | partial |
| type.empty | 13 | 400 | 18 | hist（Stage 上；勿用 quiet） |
| type.textLink | 13 | 500 | 18 | muted→ink |
| type.primaryBtn | 15 | 600 | 20 | onAccent |
| type.body | 14 | 400 | 20 | ink |
| type.caption | 12 | 400 | 16 | muted |
| type.about | 13 | 400 | 20 | muted |

字距：默认 0；品牌可不跟踪。禁止随意字号。

### 3.4 间距

| Token | px |
|---|---|
| space.1 | 8 |
| space.2 | 16 |
| space.3 | 24 |
| space.4 | 32 |
| space.5 | 40 |
| space.half | 4（仅图标微调） |
| pad.window | 16 |
| pad.header.x | 16 |
| pad.header.y | 12（Header 高 48–56） |
| pad.stage.x | ≤16 |
| pad.stage.y | ≤12 |
| gap.block | 16 |
| gap.inline | 8 |
| row.settings | ≥44 |

### 3.5 圆角 / 影 / 描边

| Token | 值 |
|---|---|
| radius.ctrl | 8 |
| radius.stage | 0–12（优先 0） |
| radius.overlay | 12 |
| radius.menu | 8 |
| shadow.card | `0 1px 2px rgba(0,0,0,0.04)` |
| shadow.overlay | `0 4px 12px rgba(0,0,0,0.08)` |
| shadow.none | none（**Stage 默认**） |
| stroke.none | 0 |
| stroke.hair | 1px `#E8EAEE` |
| focus.ring | 2px solid focus；offset 2px |

旧 Stage 厚影 `0 8 24/10%` **作废**。

### 3.6 装潢禁用

X1 渐变　X2 彩虹装饰　X3 闪烁/弹跳　X4 整窗毛玻璃　X5 裸默认控件皮　X6 灰空壳按钮　X7 框套框　X8 第二字体　X9 入场花活动画

### 3.7 动效（九分要求：默认静止）

| 项 | 规则 |
|---|---|
| 页面转场 | 无 |
| 主钮 | 颜色瞬时切换；无缩放弹跳 |
| 听写点 | 静止或透明度 1Hz 内缓变；禁止闪烁 |
| 叠层 | 可选 ≤150ms 透明度；曲线 linear；可关 |
| 字幕 | 无打字机；final 出现可瞬时 |

---

## 4. 组件全态（九分缺态 = Fail）

### 4.1 Primary（开始听 / 停止 / 下载模型）

| 态 | 底 | 字 | 环 |
|---|---|---|---|
| default | accent | onAccent | — |
| hover | accent.hover | onAccent | — |
| pressed | accent.pressed | onAccent | — |
| disabled | `#F0B3A3` 或 accent@40% | onAccent@80% | — |
| focus-visible | 同 default | | focus.ring |

几何：高 ≥36；水平 pad ≥20；radius.ctrl=8；自绘。

### 4.2 Text link（Header）

| 态 | 字色 | 底/边 |
|---|---|---|
| default | muted | 无 |
| hover | ink | 无 |
| pressed | ink | 无 |
| disabled | quiet | 无 |
| focus-visible | ink | focus.ring |

### 4.3 音源字段（仅设置/首次）

高 36；stroke.hair 或底 ground；右 chevron 自绘 16×16。  
态：default / hover（底略深）/ open / disabled / focus-visible。

### 4.4 设置行

行高 ≥44；左 label type.body；右控件；分割 line 或 space.1 留白。  
开关：自绘；开=accent；关=line 轨。

### 4.5 纪要叠层

遮罩 rgba(0,0,0,0.40)；面板 card + shadow.overlay + radius.overlay；宽 min(420, 窗宽−32)。

### 4.6 错误条

绑真 fault Code；主操作唯一且真绑定（ModelMissing→下载模型）。

---

## 5. 两态布局

**空闲**：Header 文字链 + Stage + 唯一「开始听」。  
**听写**：Header 态文案 + Stage 全幅字 + 唯一「停止」。无 Dock。  
**纪要**：叠层；v0.1 无本机档；未配远程无入口。

最小窗宽 720；交审含 720 与 ≥1024。

---

## 6. 隐私文案（冻结）

听写在本机完成，音频不会上传。  
若纪要选用远程服务，只会发送已识别的文字，不会发送音频。  
未开远程：当前纪要未启用远程。

---

## 7. 版本

| Ver | 说明 |
|---|---|
| 0.1–0.1.2 | 七分尺：色板→令牌初稿 |
| **0.2** | 九–十分尺：Pulse 迁移表、对比度、全态、动效禁用、焦点环 |
| **0.2.1** | T7：quiet/accent 加深达 AA；Stage 弱字改用 hist |

