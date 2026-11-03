# Hearline real-window pack (v4) — idle bottom-align

Built from `origin/cursor/product-notes-76b4` tip **9f48f3450ea630a2d7080b3e8736b1aad966257f**
(“Seat the idle hint and primary on the listening bottom band.”). That commit is the branch tip; nothing newer was on the remote. Binary: `/workspace/hearline-verify/hearline` (`vcs.revision=9f48f3450ea630a2d7080b3e8736b1aad966257f`, `vcs.modified=false`). Not committed and not pushed.

Capture: xfwm4 on `DISPLAY=:5`, `ffmpeg` x11grab of the frame window (title bar included). Client sized to 1120×760 (frame 1130×794) and 720×760 (frame 730×794). No GPU. No PulseAudio, so idle 「开始听」 stays disabled (`#F0B3A3`). Listening is 「听这个文件」 + `/tmp/hearline-long-zh.wav`. Model dir: `/workspace/hearline/models/sherpa-onnx-sense-voice-zh-en-ja-ko-yue-int8-2024-07-17`.

## Idle bottom-align

The empty stage is no longer a centered island. Hint 「开始听写后字幕会出现在这里」 and disabled 「开始听」 sit on the same lower band as the listening current line and 「停止」.

Measured on the 1130×794 frame (stage `#15171C`, flush to the client left, right, and bottom; no `#F3F4F6` margin ring):

- Idle primary fill: y 738–776
- Listening 「停止」 fill `#C2410C`: lower edge y 776 (same edge; the shorter label makes the top of the fill a few pixels lower)
- Narrow 720 idle primary: y 738–776 as well

So the idle bottom matches the listening band.

## Pulse reference

https://pulse.egoist.dev/ has no camera, preview, or device surface (those paths 404; the site is one marketing page). Left side is **not** the Overview CPU/Mem card dashboard. It is the Processes product window (list + Safari inspector), the densest non-Overview app face on the page: `https://pulse.egoist.dev/shots/processes-light.jpg`, shown in the “Every process, with an icon you know it by.” figure. Saved with the Mac window chrome and no browser chrome.

## E9

Dictation primary only: 「停止」 default `#C2410C`, hover `#A83410`, pressed `#8F2C0D` (pointer on the control), plus disabled 「开始听」 `#F0B3A3`. No 「保存」.

## Files

| file | what |
|---|---|
| e1-idle-1024.png | idle, client 1120, hint + disabled 开始听 on the bottom band, stage flush |
| e2-listening-1024.png | file listen, 停止 `#C2410C`, subtitles on the same band |
| e3-idle-720.png | idle, client 720 (frame 730), same bottom band |
| e4-settings.png | copied from real-smoke-v3 (settings chrome unchanged by 9f48f34; not recaptured) |
| e5-notes-or-absent.png | copied from v3 (header crop, no 纪要) |
| e6-error-model.png | copied from v3 |
| h2-download-click.png | copied from v3 |
| e9-primary-states.png | 停止 default / hover / pressed + disabled 开始听 |
| pulse-ref.png | Pulse Processes window, 1200×735, no browser chrome |
| e7-pulse-hearline-idle.png | Processes \| idle (Hearline 1:1, 12px gap) |
| e8-pulse-hearline-listening.png | Processes \| listening |
