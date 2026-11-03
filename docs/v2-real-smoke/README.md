# 真窗交审包

构建于 `5a094e1`，Linux X11，窗口带标题栏。

| 文件 | 内容 |
|---|---|
| `e1-idle-1024.png` | 空闲，宽约 1024 |
| `e2-listening-1024.png` | 听写，宽约 1024 |
| `e3-idle-720.png` | 空闲，宽 720 |
| `e4-settings.png` | 设置 |
| `e5-notes-or-absent.png` | 纪要入口或叠层 |
| `e6-error-model.png` | 模型缺失 |
| `h2-download-click.png` | 点「下载模型」 |
| `e9-primary-states.png` | 主钮四态 |
| `pulse-ref.png` | Pulse Overview 参照 |
| `e7-pulse-hearline-idle.png` | Pulse \| Hearline 空闲并排 |
| `e8-pulse-hearline-listening.png` | Pulse \| Hearline 听写并排 |

- 这台机器没有 Pulse / 麦克风音源，「开始听」是停用态。
- 听写来自文件 WAV。停止钮是 accent `#C2410C`。
- H2：点「下载模型」后开始下载 GitHub 上的 sherpa-onnx SenseVoice 包。按钮变为不可点。未下完的文件在 `/tmp/no-such-hearline-model/`。
- E7 / E8 是 [pulse.egoist.dev](https://pulse.egoist.dev/) 的 Overview 与 Hearline 并排合成。
