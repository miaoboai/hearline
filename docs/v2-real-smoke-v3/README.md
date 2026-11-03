# Hearline real-window E pack (v3)

Built from `cursor/product-notes-76b4` at **fe3b6956b3af107ff228bea34540dc0be375250f**
(“Seat the stage on the window edge”). Binary: `/workspace/hearline-verify/hearline`
(`vcs.revision` matches). Not committed and not pushed.

Capture: xfwm4 on `DISPLAY=:5`, `ffmpeg` x11grab of the frame window (title bar included). No GPU. PulseAudio is not installed, so idle 「开始听」 stays disabled. Listening is 「听这个文件」 + `/tmp/hearline-long-zh.wav`.

## What this tip shows

- **Stage flush.** Subtitle surface is square (`RadiusStage = 0`) and runs to the client left, right, and bottom. Sampled edge pixels are `#15171C`; there is no `#F3F4F6` margin ring (the old floating card).
- **Empty copy.** Idle line is 「开始听写后字幕会出现在这里」.
- **E9 same control, not 「保存」.** Four states of the dictation primary only: enabled 「停止」 default `#C2410C`, hover `#A83410`, pressed `#8F2C0D`, plus disabled 「开始听」 `#F0B3A3` (no mic, so an enabled idle primary was not available). Hover and pressed crops include the pointer. Settings 「保存」 is not in this strip.

## H2

Clicking 「下载模型」 disabled that button (`#F0B3A3`) and entered `Download` (target dir created). The process was killed during the HTTP request, before the archive body and before the “downloading …” log line.

## Files

| file | what |
|---|---|
| e1-idle-1024.png | idle, client 1120 wide, flush stage, empty hint, disabled 开始听 |
| e2-listening-1024.png | file listen, 停止 `#C2410C`, subtitles |
| e3-idle-720.png | idle, client 720 (frame 730) |
| e4-settings.png | 高级 collapsed, 保存 visible |
| e5-notes-or-absent.png | header crop: Hearline / 设置 / 导出, no 纪要 |
| e6-error-model.png | `--model-dir` empty → 找不到识别模型 + 下载模型 |
| h2-download-click.png | after click, 下载模型 disabled |
| e9-primary-states.png | 停止 default / hover / pressed + disabled 开始听 |
| pulse-ref.png | copied from real-smoke-v2 (Pulse Overview, no browser chrome) |
| e7-pulse-hearline-idle.png | Pulse \| idle |
| e8-pulse-hearline-listening.png | Pulse \| listening |
