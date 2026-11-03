# Hearline

本地实时字幕与同声传译（桌面优先）。灵感来自 [Cicada / 知了](https://github.com/wquguru/cicada) 的产品形态，但是**全新项目**：壳层用 [MyGo](https://github.com/egoist/mygo)，识别走本机模型。翻译 / 纪要的架构是本地优先，并可选手动切到 OpenAI 兼容远程，不把设计锁死在「只能本地」。

> 不是 Cicada 的 fork，也不是 NSGo / MyGo 本身。参考仓库 `miaoboai/client-pulse-vcam` 当前不可读，v0.1 按 MyGo 原生窗口（`cmd/` + `internal/`）自己搭，不拷贝别人的业务代码。

## 目标

- 任意会议、直播：本机听系统声 → 实时字幕（可双语）
- 可选：按住说话做同声传译（虚拟麦回灌）
- 可选：把已经确定的句子整理成纪要（要点 / 待办 / 术语）
- 隐私只写在设置和关于，各出现一次：「听写在本机完成，音频不会上传。」启用远程纪要时，下一句是「若纪要选用远程服务，只会发送已识别的文字，不会发送音频。」未启用时，下一句是「当前纪要未启用远程」

## 技术选型（已定）

| 层 | 选择 |
|---|---|
| 桌面 UI 壳 | **MyGo** 原生 UI（Go + GPU 绘制，不用 Electron / Flutter / Tauri / Wails） |
| ASR | 默认 **SenseVoice-Small Q8**（sherpa-onnx `model.int8.onnx`，约 229MB，纯 CPU）。接口上预留 **Qwen3-ASR**，v0.1 不加载它 |
| 翻译 / 纪要（LLM） | **本地优先**。v0.1 的 llama.cpp 还没链上，纪要按钮会停用并写明原因。配好 OpenAI 兼容的 `base_url` 和 `model` 之后，纪要只把已确定的文字发给远程。字幕环本身不调用 LLM |
| 采集 | **Linux：PulseAudio，含 PipeWire 的 pulse 兼容层**（`.monitor` 为系统声）。macOS / Windows 在 v0.1 明确返回未实现，WAV 三个平台都能播 |
| 传译输出 | 各端虚拟麦（BlackHole / VB-Cable 等，驱动不随包分发）。不属于 v0.1 |

## LLM 配置（本地优先）

翻译与纪要走同一套 LLM 后端，配置在 `~/.config/hearline/config.json`（权限 `0600`）：

1. **默认：本地** — `llm.backend` 为 `local`。不要求云端 key。v0.1 只保存 `llm.local.model_path` 和 `runtime`（默认 `llama.cpp`），推理器还没接上。窗口里的「纪要」是灰色的，旁边写着「本机模型尚未接入」。不会点进去再告诉你还没做。
2. **可选：远程 OpenAI 兼容** — `llm.backend` 设为 `remote`，填写：
   - `base_url`（例如 `https://api.openai.com/v1` 或任意兼容网关）
   - `api_key`（本地网关可以留空）
   - `model`（远端模型名）
   地址或模型名缺了，按钮仍是灰色，原因是「先在设置里填写远程服务」。配齐之后，「纪要」会打开整理页，只发送已经确定的句子。
3. **切换粒度** — v0.1 是全局一个后端。
4. **安全** — key 只在这份用户文件里，`hearline` 打印配置时会打成 `***`，不进字幕，也不进导出的 Markdown。
5. **失败** — 远程整理失败时界面写「这次没整理出来」，错误里的 API key 会换成 `***`，不会打到标准错误。不会用摘录冒充模型纪要。`llm.fallback_to_local` 只有在本机推理器真正链上、并且能单独进内存时才会走；v0.1 没链上，远程失败就停在远程失败。识别器还在内存里时，本机模型也不会被加载。

ASR 与 LLM 独立。v0.1 的识别始终是本机 CPU。远程只在你启用纪要、并且点了整理之后，才会收到文字。整理是单线程的：新的一次会取消上一次，只保留最新的确定句，先要 JSON，再写成 Markdown。

导出不依赖纪要。顶栏「导出」把确定的句子写成文本（`.txt`）或字幕（`.srt`），默认 `~/Hearline-字幕.txt`，扩展名随格式换。纪要 Markdown 要等真正整理出结果才可点。半句不计数、不导出、不送去整理。还没有确定的句子时，导出是灰色的。

示例：

```json
{
  "asr": {
    "engine": "sensevoice",
    "language": "auto",
    "vad": "mid",
    "use_itn": true,
    "num_threads": 0,
    "provider": "cpu",
    "chunk_seconds": 8
  },
  "llm": {
    "backend": "local",
    "fallback_to_local": true,
    "local": { "runtime": "llama.cpp", "model_path": "" },
    "remote": { "base_url": "", "api_key": "", "model": "" }
  }
}
```

`hearline config init` 会按 `0600` 写出这份默认文件。`num_threads` 为 0 时自动选择，并且**上限是 4**（8 核及以上用 4，4 核用 2），避免 16GiB 机器上 ONNX Runtime 把激活内存乘太大。`provider` 只接受 `cpu`。

`asr.vad` 是能量切句的灵敏度：`high`（灵敏）、`mid`（标准）、`low`（稳妥）。留空等同标准。它会跟着环境噪声调整门槛，避免系统声里的持续底噪被当成一句话，也避免安静之后的轻声切不出来。只有显式设置了固定门槛时才不用这套跟踪。

## v0.1 在 Linux CPU 上怎么跑

目标机器：x86_64，纯 CPU（验证环境是 8 核 Xeon、16GiB、无 NVIDIA）。SenseVoice 用 int8，按句切块（默认最长 8 秒，静音约 0.45 秒切句），不把整段会议一次性送进模型。

依赖：

- Go **1.27.3+**（MyGo 的要求）
- gcc（sherpa-onnx 是 cgo）
- 窗口才需要 GTK 3 和 WebKitGTK；无显示器的验证不需要
- 实时采集才需要 PulseAudio 或带 `pipewire-pulse` 的 PipeWire

```bash
# 若系统 go 低于 1.27：
curl -fsSL -o /tmp/go.tgz https://go.dev/dl/go1.27.3.linux-amd64.tar.gz
sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf /tmp/go.tgz
export PATH=/usr/local/go/bin:$PATH
go version

sudo apt-get update
sudo apt-get install -y build-essential ca-certificates
# 只要开窗口：
# sudo apt-get install -y libgtk-3-0 libglib2.0-0 libgdk-pixbuf-2.0-0 libcairo2 libwebkit2gtk-4.1-0 fonts-noto-cjk
# 只要采系统声：
# sudo apt-get install -y libpulse0 pipewire-pulse

cd /path/to/hearline
make build
make demo
```

`make demo` 会：

1. 若 `models/sherpa-onnx-sense-voice-zh-en-ja-ko-yue-int8-2024-07-17/` 里没有 Q8 模型，就从 sherpa-onnx 的官方 int8 包下载（压缩包约 156MB，权重 `model.int8.onnx` 约 229MB）
2. 用包内 `test_wavs/zh.wav` 和 `en.wav` 做 CPU 解码
3. 把确定的句子写到 `/tmp/hearline-subs.txt`，纪要没启用时把同一批句子写到 `/tmp/hearline-notes.md`（标准错误会写「纪要未启用，已导出字幕」）
4. 把界面画成 `/tmp/hearline-subs.png`，并在同目录写出空舞台、正在听、模型缺失、设置、纪要几帧（不需要显示器）

仓库里的参考画面（无显示器渲染，不是点过的窗口）：

- `docs/v0.1-empty.png` 就绪空舞台：「选一个音源，开始听」
- `docs/v0.1-listening.png` 正在听，历史一句加淡色半句，顶栏 `正在听 · 00:12`
- `docs/v0.1-subtitles.png` 已停止，字幕留在舞台上，可以清空
- `docs/v0.1-error.png` 找不到识别模型，动作是「下载模型」
- `docs/v0.1-settings.png` 设置：识别、纪要、关于与隐私、置顶、导出路径
- `docs/v0.1-notes.png` 纪要空态；本机模型未接入时按钮停用

模型地址：

`https://github.com/k2-fsa/sherpa-onnx/releases/download/asr-models/sherpa-onnx-sense-voice-zh-en-ja-ko-yue-int8-2024-07-17.tar.bz2`

也可以手动：

```bash
./bin/hearline download-model --dir "$PWD/models/sherpa-onnx-sense-voice-zh-en-ja-ko-yue-int8-2024-07-17"
./bin/hearline transcribe --wav models/sherpa-onnx-sense-voice-zh-en-ja-ko-yue-int8-2024-07-17/test_wavs/zh.wav
./bin/hearline notes --wav models/sherpa-onnx-sense-voice-zh-en-ja-ko-yue-int8-2024-07-17/test_wavs/zh.wav --out /tmp/zh.md
./bin/hearline sources
./bin/hearline          # 有显示器时打开窗口
```

查找模型的顺序：`--model-dir`、`HEARLINE_MODEL_DIR`、仓库里的 `models/...`、用户缓存 `~/Library` 或 `~/.cache/hearline/sensevoice-small-q8`。

窗口是一块舞台，不是多页调试台。顶栏只有 Hearline、会话状态（准备中 / 正在听 · 时长 / 已停止 · 句数 / 出错）、纪要、设置、导出。不再把 SenseVoice、线程、LLM 胶囊放在顶栏。停靠栏是音源、刷新、开始听 / 停止；WAV 收在「更多」里，只显示文件名。舞台在空、加载、出错时还是同一块深色区域。停止后字幕留着，可以清空。

SenseVoice 是离线模型，所以字幕按句出现，连续说话最长约 8 秒出一行。直播时半句只解码最近约 3 秒，并且队列里最多留一条半句；确定的句子不会被丢掉。解码队列有上限，只有一个工作线程。半句是斜体淡色，不进句数。线程数夹在 1 到 4。

快捷键（焦点在输入框时，空格留给输入）：空格开始或停止，Ctrl+E（macOS 为 ⌘E）导出 SRT，Ctrl+Shift+C 复制当前句，Ctrl+, 打开设置，Esc 关闭设置或纪要；面板都关着时 Esc 停止。正在听时关窗口，会先问「停止并退出？」。

出错只给短中文和能做的动作：找不到模型是「下载模型」，模型损坏是「重试」，音源听不到是「改音源」，文件读不了是「重选」。没有权限、识别跟不上，只提示，不放一个假按钮。

`go test -short ./...` 不加载模型。`make test-asr` 会真正加载 Q8 并检查中文样例里有「点」和「下午」。

第一次 `make build` 会通过 Go 模块下载 `sherpa-onnx-go-linux` 里的 CPU 版 `libsherpa-onnx-c-api.so` 和 `libonnxruntime.so`，并用 rpath 指到模块缓存。在**将要运行的那台 Linux** 上编译，不要把别的机器编出来的二进制直接拷走。

### macOS / Windows

- 识别代码与 Linux 共用 sherpa-onnx 的 Go 绑定（macOS / Windows 用各自的预编译库）。本仓库的验证在 Linux 上完成。
- 系统声：macOS 计划用 Core Audio process tap，Windows 计划用 WASAPI loopback。v0.1 调用采集会得到明确的未实现错误。
- WAV 识别不依赖系统声，三个平台的 `transcribe` 路径一致（非 Linux 需本机 cgo 工具链和对应的 sherpa 库）。

## 代码布局

```
cmd/hearline              窗口、demo、transcribe、notes、download-model
internal/audio            WAV、能量切句（含自适应底噪）、Linux Pulse 采集、其它 OS 的缺口
internal/asr              Recognizer 接口
internal/asr/sensevoice   SenseVoice-Small Q8 下载与 sherpa-onnx CPU 推理
internal/asr/mock         无模型时的字幕接线
internal/asr/qwen3        占位，返回未接入
internal/pipeline         音源块 → 切句 → 有界解码队列 → 字幕行
internal/resident         识别器与本机模型不能同时留在内存里
internal/fault            稳定错误码和短中文
internal/notes            确定的句子、单线程整理、TXT / SRT / Markdown
internal/llm              本地 / 远程配置与 OpenAI 兼容客户端
internal/config           config.json
internal/ui               MyGo 窗口；无显示器时也能画 PNG
```

## 三期路线

### 一期 · 桌面本地听写与字幕（本分支在做的范围）

- MyGo 单窗口：字幕舞台、设置、纪要
- Linux 系统声 / 麦克风（Pulse / PipeWire）；macOS、Windows 采集仍是明确缺口
- SenseVoice-Small Q8，CPU，按句解码；半句只解尾窗；灵敏度可调；线程最多 4
- 无模型时可用 `--engine mock` 看字幕是否在动
- 确定的句子可导出 TXT / SRT。远程纪要配齐后可以整理；本地 llama.cpp 仍未接入，按钮保持停用
- LLM 不进字幕环。识别器在内存里时不加载本机模型

### 二期 · 本机模型与传译

- 把本地小模型真正跑起来
- 按住说话 → ASR + 翻译 + TTS → 虚拟麦
- 收起停靠留到这一期，不在 v0.1 里做假

### 三期 · 移动端遥控与看字幕

- 手机不做系统声采集主体
- 遥控桌面会话、同步看字幕 / 纪要
- 默认不上传音频

## 非目标（初期）

- 不往会议里塞机器人 Bot
- 不以云端 ASR 为默认路径
- 不打包 GPL 虚拟声卡驱动

## 已知限制

- SenseVoice 在 sherpa-onnx 里是离线模型，v0.1 用切句模拟实时，不是 token 级流式。
- 打开 ITN 时英文样例可能把 “gold” 记成 “code”。这是模型行为，不是后处理写死的。
- 本地 LLM 推理、Qwen3-ASR、macOS 系统声、Windows WASAPI 都还没接。正在听时关掉窗口会先确认，确认后停止采集。没有独立的停靠窗。
- 参考客户端 `client-pulse-vcam` 未能读取，因此没有照搬它的托盘或预览实现。

## 许可

待定（开源前补齐 LICENSE）。SenseVoice 的 sherpa-onnx 发行包本身是 Apache-2.0，模型文件不进 git。
