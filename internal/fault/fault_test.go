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
		{"wav: bad riff header", FileInvalid},
		{"permission denied opening monitor", Permission},
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
