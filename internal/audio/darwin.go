//go:build darwin

package audio

import (
	"bufio"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

// macOS live capture uses ffmpeg's AVFoundation input. System output
// loopback (Core Audio process tap) is still unimplemented; we only
// list and capture microphones so 「开始听」 works on Darwin.

var (
	avAudioDeviceLine = regexp.MustCompile(`\[(\d+)\]\s+(.+)$`)
	ffmpegCandidates  = []string{
		"ffmpeg",
		"/opt/homebrew/bin/ffmpeg",
		"/usr/local/bin/ffmpeg",
	}
)

func findFFmpeg() (string, error) {
	for _, c := range ffmpegCandidates {
		if c == "ffmpeg" {
			if p, err := exec.LookPath(c); err == nil {
				return p, nil
			}
			continue
		}
		if st, err := os.Stat(c); err == nil && !st.IsDir() {
			return c, nil
		}
	}
	return "", fmt.Errorf("找不到 ffmpeg（Homebrew: brew install ffmpeg）。macOS 麦克风采集需要它")
}

func systemSources() ([]Source, error) {
	ff, err := findFFmpeg()
	if err != nil {
		return nil, err
	}
	devs, err := listAVFoundationAudio(ff)
	if err != nil {
		return nil, err
	}
	if len(devs) == 0 {
		return nil, fmt.Errorf("没有找到麦克风。请在系统设置 → 隐私与安全性 → 麦克风中允许，并确认本机有输入设备")
	}
	out := make([]Source, 0, len(devs))
	for _, d := range devs {
		out = append(out, Source{
			ID:   fmt.Sprintf("source:avfoundation:%d", d.Index),
			Name: "麦克风 · " + d.Name,
			Kind: "mic",
		})
	}
	return out, nil
}

type avDevice struct {
	Index int
	Name  string
}

func listAVFoundationAudio(ffmpeg string) ([]avDevice, error) {
	cmd := exec.Command(ffmpeg, "-hide_banner", "-f", "avfoundation", "-list_devices", "true", "-i", "")
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("启动 ffmpeg 列出音源失败: %w", err)
	}
	var (
		inAudio bool
		devs    []avDevice
	)
	sc := bufio.NewScanner(stderr)
	for sc.Scan() {
		line := sc.Text()
		low := strings.ToLower(line)
		if strings.Contains(low, "avfoundation audio devices") {
			inAudio = true
			continue
		}
		if inAudio && strings.Contains(low, "video devices") {
			inAudio = false
			continue
		}
		if !inAudio {
			continue
		}
		m := avAudioDeviceLine.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		idx, _ := strconv.Atoi(m[1])
		name := strings.TrimSpace(m[2])
		name = strings.ReplaceAll(name, "\"", "")
		name = strings.ReplaceAll(name, "“", "")
		name = strings.ReplaceAll(name, "”", "")
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		devs = append(devs, avDevice{Index: idx, Name: name})
	}
	_ = cmd.Wait()
	if len(devs) == 0 {
		return nil, fmt.Errorf("ffmpeg 未列出可用麦克风（权限不足或无输入设备）")
	}
	return devs, nil
}

func captureSystem(ctx context.Context, sourceName string) (<-chan []float32, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	ff, err := findFFmpeg()
	if err != nil {
		return nil, err
	}
	idx, err := resolveAVIndex(ff, sourceName)
	if err != nil {
		return nil, err
	}
	input := fmt.Sprintf(":%d", idx)
	cmd := exec.CommandContext(ctx, ff,
		"-hide_banner", "-loglevel", "error",
		"-f", "avfoundation",
		"-i", input,
		"-ac", "1",
		"-ar", strconv.Itoa(SampleRate),
		"-f", "f32le",
		"pipe:1",
	)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("打开麦克风失败: %w", err)
	}

	errCh := make(chan string, 1)
	go func() {
		b, _ := io.ReadAll(stderr)
		msg := strings.TrimSpace(string(b))
		waitErr := cmd.Wait()
		if msg == "" && waitErr != nil && ctx.Err() == nil {
			msg = waitErr.Error()
		}
		if msg != "" && ctx.Err() == nil {
			errCh <- classifyFFmpegMicErr(msg)
			return
		}
		errCh <- ""
	}()

	out := make(chan []float32, 64)
	first := make(chan struct{})
	var firstOnce sync.Once
	var killOnce sync.Once
	kill := func() {
		killOnce.Do(func() {
			if cmd.Process != nil {
				_ = cmd.Process.Kill()
			}
		})
	}

	go func() {
		defer close(out)
		defer kill()
		frame := SampleRate / 10
		buf := make([]byte, frame*4)
		for {
			if ctx.Err() != nil {
				return
			}
			n, readErr := io.ReadFull(stdout, buf)
			if n >= 4 {
				samples := bytesToFloat32LE(buf[: n&^3])
				if len(samples) > 0 {
					firstOnce.Do(func() { close(first) })
					cp := append([]float32(nil), samples...)
					select {
					case <-ctx.Done():
						return
					case out <- cp:
					}
				}
			}
			if readErr == nil {
				continue
			}
			return
		}
	}()

	select {
	case <-first:
		return out, nil
	case msg := <-errCh:
		kill()
		// Drain so the reader goroutine can exit.
		go func() {
			for range out {
			}
		}()
		if msg == "" {
			msg = "麦克风采集意外退出"
		}
		return nil, fmt.Errorf("%s", msg)
	case <-time.After(3 * time.Second):
		kill()
		go func() {
			for range out {
			}
		}()
		return nil, fmt.Errorf("打开麦克风超时。请在系统设置 → 隐私与安全性 → 麦克风中检查权限")
	case <-ctx.Done():
		kill()
		return nil, ctx.Err()
	}
}

func resolveAVIndex(ffmpeg, sourceName string) (int, error) {
	sourceName = strings.TrimSpace(sourceName)
	if strings.HasPrefix(sourceName, "avfoundation:") {
		n := strings.TrimPrefix(sourceName, "avfoundation:")
		idx, err := strconv.Atoi(n)
		if err != nil {
			return 0, fmt.Errorf("无效音源 %s", sourceName)
		}
		return idx, nil
	}
	if idx, err := strconv.Atoi(sourceName); err == nil {
		return idx, nil
	}
	devs, err := listAVFoundationAudio(ffmpeg)
	if err != nil {
		return 0, err
	}
	for _, d := range devs {
		if d.Name == sourceName || strings.Contains(sourceName, d.Name) || strings.Contains(d.Name, sourceName) {
			return d.Index, nil
		}
	}
	return 0, fmt.Errorf("找不到音源 %s", sourceName)
}

func bytesToFloat32LE(b []byte) []float32 {
	n := len(b) / 4
	out := make([]float32, n)
	for i := 0; i < n; i++ {
		out[i] = math.Float32frombits(binary.LittleEndian.Uint32(b[i*4 : i*4+4]))
	}
	return out
}

func classifyFFmpegMicErr(msg string) string {
	low := strings.ToLower(msg)
	switch {
	case strings.Contains(low, "permission"),
		strings.Contains(low, "not authorized"),
		strings.Contains(low, "denied"),
		strings.Contains(msg, "权限"):
		return "没有麦克风权限：请在系统设置 → 隐私与安全性 → 麦克风中允许"
	case strings.Contains(low, "input/output error"),
		strings.Contains(low, "cannot open"),
		strings.Contains(low, "error opening"):
		return "打开麦克风失败（权限或设备占用）。请在系统设置 → 隐私与安全性 → 麦克风中检查权限"
	default:
		return msg
	}
}
