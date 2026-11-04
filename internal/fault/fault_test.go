package fault

import "testing"

func TestClassifyBindsRealCodes(t *testing.T) {
	cases := []struct {
		msg  string
		code string
	}{
		{"open model.int8.onnx: no such file", ModelMissing},
		{"sherpa-onnx 未能加载 model.int8.onnx", ModelUnreadable},
		{"pulse: connection refused", CaptureUnavailable},
		{"Core Audio process tap 未实现", CaptureUnavailable},
		{"macOS 系统声采集尚未实现：v0.1 计划用 Core Audio process tap（系统输出环回）。当前请使用 WAV 音源（hearline transcribe --wav）", CaptureUnavailable},
		{"Core Audio 打开麦克风失败: miniaudio: Failed to open backend device", CaptureUnavailable},
		{"wav: bad riff header", FileInvalid},
		{"permission denied opening monitor", Permission},
		{"麦克风权限被拒绝（permission denied）。请在「系统设置 → 隐私与安全性 → 麦克风」中允许 Hearline，或允许启动它的终端", Permission},
		{"miniaudio: Permission denied", Permission},
		{"something else", ""},
	}
	for _, tc := range cases {
		if got := Classify(tc.msg); got != tc.code {
			t.Fatalf("%q → %q, want %q", tc.msg, got, tc.code)
		}
		if tc.code != "" && UserMessage(tc.code) == "" {
			t.Fatalf("%s has no user message", tc.code)
		}
	}
	if Action(DecodeBacklog) != "" || Action(Permission) != "" {
		t.Fatal("backlog and permission must not invent a button")
	}
	if Action(ModelMissing) != "下载模型" || Action(CaptureUnavailable) != "改音源" {
		t.Fatal("action labels drifted")
	}
}
