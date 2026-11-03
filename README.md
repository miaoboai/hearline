# Hearline

本地实时字幕与同声传译（桌面优先）。灵感来自 [Cicada / 知了](https://github.com/wquguru/cicada) 的产品形态，但是**全新项目**：壳层用 [MyGo](https://github.com/egoist/mygo)，识别走本机模型。翻译 / 纪要的架构是本地优先，并可选手动切到 OpenAI 兼容远程，不把设计锁死在「只能本地」。

> 不是 Cicada 的 fork，也不是 NSGo / MyGo 本身。参考仓库 `miaoboai/client-pulse-vcam` 当前不可读，v0.1 按 MyGo 原生窗口（`cmd/` + `internal/`）自己搭，不拷贝别人的业务代码。

## 目标

- 任意会议、直播：本机听系统声 → 实时字幕（可双语）
- 可选：按住说话做同声传译（虚拟麦回灌）
- 可选：边听边写结构化纪要
- 隐私默认：音频与 ASR 不出机

## 技术选型（已定）

| 层 | 选择 |
|---|---|
| 桌面 UI 壳 | **MyGo** 原生 UI（Go + GPU 绘制，不用 Electron / Flutter / Tauri / Wails） |
| ASR | 默认 **SenseVoice-Small Q8**（sherpa-onnx `model.int8.onnx`，约 229MB，纯 CPU）。接口上预留 **Qwen3-ASR**，v0.1 不加载它 |
| 翻译 / 纪要（LLM） | **本地优先**：配置默认本机小模型（如 Qwen2.5-0.5B / 1.5B Q4，llama.cpp）。**可选手动切到远程** OpenAI 兼容 API。v0.1 字幕环不调用 LLM，但配置、校验和远程客户端已经接上 |
| 采集 | **Linux：PulseAudio，含 PipeWire 的 pulse 兼容层**（`.monitor` 为系统声）。macOS / Windows 在 v0.1 明确返回未实现，WAV 三个平台都能播 |
| 传译输出 | 各端虚拟麦（BlackHole / VB-Cable 等，驱动不随包分发）。不属于 v0.1 |

## LLM 配置（本地优先）

翻译与纪要走同一套 LLM 后端，配置在 `~/.config/hearline/config.json`（权限 `0600`）：

1. **默认：本地** — `llm.backend` 为 `local`。不要求云端 key。v0.1 只保存 `llm.local.model_path` 和 `runtime`（默认 `llama.cpp`），推理器下一期再接。
2. **可选：远程 OpenAI 兼容** — `llm.backend` 设为 `remote`，填写：
   - `base_url`（例如 `https://api.openai.com/v1` 或任意兼容网关）
   - `api_key`
   - `model`（远端模型名）
3. **切换粒度** — v0.1 是全局一个后端。
4. **安全** — key 只在这份用户文件里，`hearline` 打印配置时会打成 `***`，不进字幕日志。
5. **失败退回** — `llm.fallback_to_local` 为 true 时，远程失败会再走本地路径。v0.1 的本地路径会返回「尚未接入」，不会假装译出来。

ASR 与 LLM 独立。v0.1 的 ASR 始终是本机 CPU。远程选项只覆盖将来的翻译和纪要。

示例：

```json
{
  "asr": {
    "engine": "sensevoice",
    "language": "auto",
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

`hearline config init` 会按 `0600` 写出这份默认文件。`num_threads` 为 0 时自动选择，并且**上限是 4**，避免 16GiB 机器上 ONNX Runtime 把激活内存乘太大。`provider` 只接受 `cpu`。

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
3. 把字幕行写到 `/tmp/hearline-subs.txt`，并把字幕界面画成 `/tmp/hearline-subs.png`（不需要显示器）。仓库里的参考画面是 `docs/v0.1-subtitles.png`

模型地址：

`https://github.com/k2-fsa/sherpa-onnx/releases/download/asr-models/sherpa-onnx-sense-voice-zh-en-ja-ko-yue-int8-2024-07-17.tar.bz2`

也可以手动：

```bash
./bin/hearline download-model --dir "$PWD/models/sherpa-onnx-sense-voice-zh-en-ja-ko-yue-int8-2024-07-17"
./bin/hearline transcribe --wav models/sherpa-onnx-sense-voice-zh-en-ja-ko-yue-int8-2024-07-17/test_wavs/zh.wav
./bin/hearline sources
./bin/hearline          # 有显示器时：选音源或播放 WAV，字幕往下长
```

查找模型的顺序：`--model-dir`、`HEARLINE_MODEL_DIR`、仓库里的 `models/...`、用户缓存 `~/Library` 或 `~/.cache/hearline/sensevoice-small-q8`。

窗口里的闭环是：刷新音源 → 选一个 Pulse 源（名字以 `.monitor` 结尾的是系统声）或填 WAV → 开始听 / 播放 WAV → 字幕区更新。SenseVoice 是离线模型，所以字幕按句出现，连续说话最长约 8 秒出一行；直播时每 2 秒会尝试刷新半句。

`go test -short ./...` 不加载模型。`make test-asr` 会真正加载 Q8 并检查中文样例里有「点」和「下午」。

第一次 `make build` 会通过 Go 模块下载 `sherpa-onnx-go-linux` 里的 CPU 版 `libsherpa-onnx-c-api.so` 和 `libonnxruntime.so`，并用 rpath 指到模块缓存。在**将要运行的那台 Linux** 上编译，不要把别的机器编出来的二进制直接拷走。

### macOS / Windows

- 识别代码与 Linux 共用 sherpa-onnx 的 Go 绑定（macOS / Windows 用各自的预编译库）。本仓库的验证在 Linux 上完成。
- 系统声：macOS 计划用 Core Audio process tap，Windows 计划用 WASAPI loopback。v0.1 调用采集会得到明确的未实现错误。
- WAV 识别不依赖系统声，三个平台的 `transcribe` 路径一致（非 Linux 需本机 cgo 工具链和对应的 sherpa 库）。

## 代码布局

```
cmd/hearline              窗口、demo、transcribe、download-model
internal/audio            WAV、能量切句、Linux Pulse 采集、其它 OS 的缺口
internal/asr              Recognizer 接口
internal/asr/sensevoice   SenseVoice-Small Q8 下载与 sherpa-onnx CPU 推理
internal/asr/mock         无模型时的字幕接线
internal/asr/qwen3        占位，返回未接入
internal/pipeline         音源块 → 切句 → 识别 → 字幕行
internal/llm              本地 / 远程配置与 OpenAI 兼容客户端
internal/config           config.json
internal/ui               MyGo 字幕窗口；无窗口时也能画一帧 PNG
```

## 三期路线

### 一期 · 桌面本地听写与字幕（本分支在做的范围）

- MyGo 字幕窗口和音源选择
- Linux 系统声 / 麦克风（Pulse / PipeWire）；macOS、Windows 采集仍是明确缺口
- SenseVoice-Small Q8，CPU，按句解码
- 无模型时可用 `--engine mock` 看字幕是否在动
- LLM 配置与远程客户端就位，**不进字幕环**

### 二期 · 纪要与传译

- 把本地小模型真正跑起来，并允许切远程 OpenAI 兼容 API
- 按住说话 → ASR + 翻译 + TTS → 虚拟麦
- 会话落盘（Markdown / SRT）

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
- 本地 LLM 推理、Qwen3-ASR、macOS 系统声、Windows WASAPI 都还没接。
- 参考客户端 `client-pulse-vcam` 未能读取，因此没有照搬它的托盘或预览实现。

## 许可

待定（开源前补齐 LICENSE）。SenseVoice 的 sherpa-onnx 发行包本身是 Apache-2.0，模型文件不进 git。
