//go:build linux

package audio

import (
	"context"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/jfreymuth/pulse"
)

// systemSources lists PulseAudio sources. PipeWire exposes the same
// protocol when pipewire-pulse is running. Monitor sources (system
// loopback) are labeled 系统声.
func systemSources() ([]Source, error) {
	c, err := pulse.NewClient()
	if err != nil {
		return nil, fmt.Errorf("PulseAudio/PipeWire 不可用（需要本机 pulse 套接字，PipeWire 请启用 pipewire-pulse）: %w", err)
	}
	defer c.Close()
	list, err := c.ListSources()
	if err != nil {
		return nil, fmt.Errorf("列出 Pulse 音源失败: %w", err)
	}
	out := make([]Source, 0, len(list))
	for _, s := range list {
		id := s.ID()
		kind := "mic"
		name := "麦克风 · " + s.Name()
		if strings.Contains(id, ".monitor") {
			kind = "system"
			name = "系统声 · " + s.Name()
		}
		out = append(out, Source{ID: "source:" + id, Name: name, Kind: kind})
	}
	return out, nil
}

// captureSystem records one Pulse source as 16 kHz mono float32.
func captureSystem(ctx context.Context, sourceName string) (<-chan []float32, error) {
	c, err := pulse.NewClient()
	if err != nil {
		return nil, fmt.Errorf("PulseAudio/PipeWire 不可用: %w", err)
	}
	src, err := c.SourceByID(sourceName)
	if err != nil {
		c.Close()
		return nil, fmt.Errorf("找不到音源 %s: %w", sourceName, err)
	}
	out := make(chan []float32, 64)
	var mu sync.Mutex
	var inflight sync.WaitGroup
	stopped := false
	rate := SampleRate
	writer := pulse.Float32Writer(func(p []float32) (int, error) {
		mu.Lock()
		if stopped || ctx.Err() != nil {
			mu.Unlock()
			return 0, io.EOF
		}
		cp := append([]float32(nil), p...)
		r := rate
		inflight.Add(1)
		mu.Unlock()
		if r != 0 && r != SampleRate {
			cp = Resample(cp, r, SampleRate)
		}
		var sendErr error
		select {
		case <-ctx.Done():
			sendErr = ctx.Err()
		case out <- cp:
		}
		inflight.Done()
		if sendErr != nil {
			return 0, sendErr
		}
		return len(p), nil
	})
	// Sample rate must be set before RecordLatency, which reads it.
	stream, err := c.NewRecord(writer,
		pulse.RecordMono,
		pulse.RecordSampleRate(SampleRate),
		pulse.RecordLatency(0.08),
		pulse.RecordSource(src),
		pulse.RecordMediaName("Hearline"),
	)
	if err != nil {
		c.Close()
		return nil, fmt.Errorf("打开录音流失败: %w", err)
	}
	mu.Lock()
	if got := stream.SampleRate(); got > 0 {
		rate = got
	}
	mu.Unlock()
	stream.Start()
	go func() {
		<-ctx.Done()
		mu.Lock()
		stopped = true
		mu.Unlock()
		stream.Stop()
		inflight.Wait()
		close(out)
		stream.Close()
		c.Close()
	}()
	return out, nil
}
