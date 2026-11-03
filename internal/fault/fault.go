// Package fault is the small set of failures the window is allowed to act on.
// Anything else stays out of the subtitle stage.
package fault

import "strings"

const (
	ModelMissing       = "model_missing"
	CaptureUnavailable = "capture_unavailable"
	DecodeBacklog      = "decode_backlog"
)

// Classify maps an error string onto a code the UI can bind to a real action.
// An empty code means the window should not invent a button for it.
func Classify(msg string) string {
	low := strings.ToLower(msg)
	switch {
	case strings.Contains(low, "model"),
		strings.Contains(low, "onnx"),
		strings.Contains(low, "no such"),
		strings.Contains(low, "not found"),
		strings.Contains(msg, "未找到"),
		strings.Contains(msg, "模型"):
		return ModelMissing
	case strings.Contains(low, "pulse"),
		strings.Contains(low, "pipewire"),
		strings.Contains(low, "connection refused"),
		strings.Contains(low, "wasapi"),
		strings.Contains(low, "core audio"),
		strings.Contains(msg, "未实现"):
		return CaptureUnavailable
	default:
		return ""
	}
}
