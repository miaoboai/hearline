# Hearline 九分标注稿（落地 · v2.1）

图：
- `docs/hearline-v2-idle-annotated.jpg` — 空闲
- `docs/hearline-v2-listening-annotated.jpg` — 听写

门禁：`docs/hearline-pulse-craft-gate.md` **v2.1**（含 **Q1–Q3 九分必须项**）  
令牌：`docs/hearline-design-system.md` v0.2

## 看图注意（以文字为准）

听写标注图里 Header「设置/导出」若画成描边块 = **错例**；实现必须做成 **无底无边文字链**（Q1/A2）。其余标注（无 Dock、字阶、一主钮、边距）照做。

## 空闲必须

| 项 | 值 |
|---|---|
| Header | 48–56 高；文字链；0 描边块 |
| Stage | 近全幅；边距 ≤16/≤12；圆角 ≤12；shadow.none |
| 主钮 | 仅「开始听」；h≥36；r=8；accent；自绘 |
| 禁止 | Dock / Select / 刷新 / 更多 / 浮黑卡 |

## 听写必须

| 项 | 值 |
|---|---|
| 主操作 | 仅「停止」 |
| 字 | 当前 ≥24/600；历史 15/400 |
| 禁止 | 音源行、Dock、常驻纪要栏 |

## Q 九分必须（七分交 = Fail）

1. 同栈自绘精度（无默认皮）  
2. 并排 3 秒分不出更业余；≥2 人「像套件」= Fail  
3. 令牌进代码常量 + 破窗测试  

