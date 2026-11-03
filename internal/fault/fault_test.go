package fault

import "testing"

func TestClassifyBindsRealCodes(t *testing.T) {
	cases := []struct {
		msg  string
		code string
	}{
		{"open model.int8.onnx: no such file", ModelMissing},
		{"pulse: connection refused", CaptureUnavailable},
		{"Core Audio process tap 未实现", CaptureUnavailable},
		{"something else", ""},
	}
	for _, tc := range cases {
		if got := Classify(tc.msg); got != tc.code {
			t.Fatalf("%q → %q, want %q", tc.msg, got, tc.code)
		}
	}
}
