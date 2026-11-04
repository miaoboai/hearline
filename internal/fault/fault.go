// Package fault is the small set of failures the window is allowed to act on.
// Anything else stays out of the subtitle stage.
package fault

import "strings"

const (
	ModelMissing       = "model_missing"
	ModelUnreadable    = "model_unreadable"
	CaptureUnavailable = "capture_unavailable"
	FileInvalid        = "file_invalid"
	Permission         = "permission"
	DecodeBacklog      = "decode_backlog"
)

// UserMessage is the short Chinese line for a code. Unknown codes are blank.
func UserMessage(code string) string {
	switch code {
	case ModelMissing:
		return "找不到识别模型"
	case ModelUnreadable:
		return "识别器无法启动"
	case CaptureUnavailable:
		return "音源不可用"
	case FileInvalid:
		return "无法读取这个文件"
	case Permission:
		return "没有麦克风/系统声权限"
	case DecodeBacklog:
		return "识别有点跟不上，句子会稍晚出现"
	default:
		return ""
	}
}

// Action is the button label for a code, or empty when there is nothing honest to do.
func Action(code string) string {
	switch code {
	case ModelMissing:
		return "下载模型"
	case ModelUnreadable:
		return "重试"
	case CaptureUnavailable:
		return "改音源"
	case FileInvalid:
		return "重选"
	default:
		return ""
	}
}

// Classify maps an error string onto a code the UI can bind to a real action.
// An empty code means the window should not invent a button for it.
func Classify(msg string) string {
	low := strings.ToLower(msg)
	switch {
	case strings.Contains(low, "permission"),
		strings.Contains(low, "denied"),
		strings.Contains(low, "access"),
		strings.Contains(msg, "权限"):
		return Permission
	case strings.Contains(low, "no such"),
		strings.Contains(low, "not found"),
		strings.Contains(msg, "未找到"),
		strings.Contains(msg, "找不到"):
		return ModelMissing
	case strings.Contains(low, "未能加载"),
		strings.Contains(low, "invalid model"),
		strings.Contains(msg, "损坏"):
		return ModelUnreadable
	case strings.Contains(low, "pulse"),
		strings.Contains(low, "pipewire"),
		strings.Contains(low, "connection refused"),
		strings.Contains(low, "wasapi"),
		strings.Contains(low, "core audio"),
		strings.Contains(low, "coreaudio"),
		strings.Contains(msg, "未实现"):
		return CaptureUnavailable
	case strings.Contains(low, "wav"),
		strings.Contains(low, "riff"),
		strings.Contains(low, "audio format"),
		strings.Contains(msg, "无法读取"):
		return FileInvalid
	case strings.Contains(low, "model"),
		strings.Contains(low, "onnx"),
		strings.Contains(msg, "模型"):
		return ModelMissing
	default:
		return ""
	}
}
