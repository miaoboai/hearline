//go:build !linux && !darwin

package audio

import (
	"context"
	"fmt"
	"runtime"
)

// systemSources is the gap for systems without a capture backend.
// macOS microphones are implemented in coreaudio_darwin.go; Linux uses Pulse.
func systemSources() ([]Source, error) {
	return nil, fmt.Errorf("%s", platformGap())
}

func captureSystem(ctx context.Context, sourceName string) (<-chan []float32, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return nil, fmt.Errorf("%s (%s)", platformGap(), sourceName)
}

func platformGap() string {
	switch runtime.GOOS {
	case "windows":
		return "Windows 系统声采集尚未实现：v0.1 计划用 WASAPI loopback。当前请使用 WAV 音源（hearline transcribe --wav）"
	default:
		return runtime.GOOS + " 没有系统声采集实现。当前请使用 WAV 音源"
	}
}
