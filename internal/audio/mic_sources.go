package audio

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"strings"
)

// inputDevice is one capture endpoint before it becomes a Source.
// ID is the stable device UID, not a display name.
type inputDevice struct {
	ID      string
	Name    string
	Default bool
}

// microphoneDeniedMessage is the capture error the UI maps to
// fault.Permission. It must not contain "wav" or "无法读取", or the
// window shows a file error instead of the microphone permission line.
const microphoneDeniedMessage = "麦克风权限被拒绝（permission denied）。请在「系统设置 → 隐私与安全性 → 麦克风」中允许 Hearline，或允许启动它的终端"

func microphoneDenied() error {
	return errors.New(microphoneDeniedMessage)
}

// micSources builds live microphone sources. The default input is first
// so the window selects it. System loopback is never invented here.
func micSources(devs []inputDevice) []Source {
	ordered := append([]inputDevice(nil), devs...)
	// Default first; keep the enumeration order otherwise.
	stableDefaultFirst(ordered)
	out := make([]Source, 0, len(ordered))
	for _, d := range ordered {
		id := strings.TrimSpace(d.ID)
		if id == "" {
			continue
		}
		name := strings.TrimSpace(d.Name)
		if name == "" {
			name = id
		}
		out = append(out, Source{
			ID:   "source:" + id,
			Name: "麦克风 · " + name,
			Kind: "mic",
		})
	}
	return out
}

func stableDefaultFirst(devs []inputDevice) {
	// Insertion move keeps relative order of the non-default devices.
	n := len(devs)
	for i := 0; i < n; i++ {
		if !devs[i].Default {
			continue
		}
		d := devs[i]
		copy(devs[1:i+1], devs[0:i])
		devs[0] = d
		return
	}
}

// deviceUID reads a NUL-terminated UID from a Core Audio device id.
func deviceUID(b []byte) string {
	if i := bytes.IndexByte(b, 0); i >= 0 {
		return string(b[:i])
	}
	return string(b)
}

// putDeviceUID writes uid into dst, leaving a trailing NUL.
func putDeviceUID(dst []byte, uid string) error {
	if uid == "" {
		return errors.New("empty device uid")
	}
	if len(uid) >= len(dst) {
		return errors.New("device uid too long")
	}
	for i := range dst {
		dst[i] = 0
	}
	copy(dst, uid)
	return nil
}

// wrapCapture turns a backend failure into a string the fault package
// can bind. Permission stays on the permission path; anything else is a
// Core Audio capture failure, never a WAV read.
func wrapCapture(err error) error {
	if err == nil {
		return nil
	}
	msg := err.Error()
	low := strings.ToLower(msg)
	if strings.Contains(low, "permission") || strings.Contains(low, "denied") || strings.Contains(msg, "权限") {
		return fmt.Errorf("%s: %w", microphoneDeniedMessage, err)
	}
	return fmt.Errorf("Core Audio 打开麦克风失败: %w", err)
}

// float32FromPCM converts interleaved little-endian PCM to mono float32
// in about [-1, 1]. sampleBytes is 2 (s16) or 4 (f32).
func float32FromPCM(b []byte, sampleBytes, channels int) []float32 {
	if sampleBytes != 2 && sampleBytes != 4 {
		return nil
	}
	if channels < 1 {
		channels = 1
	}
	n := len(b) / sampleBytes
	if n == 0 {
		return nil
	}
	interleaved := make([]float32, n)
	switch sampleBytes {
	case 4:
		for i := 0; i < n; i++ {
			interleaved[i] = math.Float32frombits(binary.LittleEndian.Uint32(b[i*4 : i*4+4]))
		}
	case 2:
		for i := 0; i < n; i++ {
			interleaved[i] = float32(int16(binary.LittleEndian.Uint16(b[i*2:i*2+2]))) / 32768
		}
	}
	return Downmix(interleaved, channels)
}
