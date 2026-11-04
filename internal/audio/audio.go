// Package audio captures mono PCM and splits it into utterances.
// Linux talks to PulseAudio, including PipeWire's Pulse compatibility.
// macOS captures microphone input through Core Audio. System-output
// loopback is not implemented there, and Windows still has no capture
// backend. WAV input works on every OS.
package audio

import (
	"context"
	"fmt"
	"math"
	"strings"
)

// SampleRate is the rate SenseVoice-Small expects.
const SampleRate = 16000

// Source is one input the user can pick.
type Source struct {
	ID   string // "source:<pulse name>" or "wav:<path>"
	Name string
	Kind string // "system", "mic", or "file"
}

// List returns live capture devices. On Linux a failure to reach
// PulseAudio/PipeWire is returned with an empty list. On macOS the list
// is Core Audio inputs; a missing system-output loopback is not an error
// and does not invent a system source. WAV playback does not need a device.
func List() ([]Source, error) {
	return systemSources()
}

// Capture starts a 16 kHz mono float32 stream. id is "source:…" from List
// or "wav:<path>". The channel closes when ctx is done or the file ends.
// Each chunk is owned by the caller.
func Capture(ctx context.Context, id string) (<-chan []float32, error) {
	switch {
	case strings.HasPrefix(id, "wav:"):
		return StreamWAV(ctx, strings.TrimPrefix(id, "wav:"), true)
	case strings.HasPrefix(id, "source:"):
		return captureSystem(ctx, strings.TrimPrefix(id, "source:"))
	default:
		return nil, fmt.Errorf("unknown audio source %q", id)
	}
}

// RMS is the root-mean-square energy of samples in about [-1, 1].
func RMS(samples []float32) float32 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, v := range samples {
		sum += float64(v) * float64(v)
	}
	return float32(math.Sqrt(sum / float64(len(samples))))
}

// Downmix averages interleaved channels. One channel is returned as-is.
func Downmix(interleaved []float32, channels int) []float32 {
	if channels <= 1 || len(interleaved) == 0 {
		return interleaved
	}
	n := len(interleaved) / channels
	out := make([]float32, n)
	for i := 0; i < n; i++ {
		var sum float32
		for c := 0; c < channels; c++ {
			sum += interleaved[i*channels+c]
		}
		out[i] = sum / float32(channels)
	}
	return out
}

// Resample linearly resamples mono audio. Equal rates return in unchanged.
func Resample(in []float32, from, to int) []float32 {
	if from <= 0 || to <= 0 || len(in) == 0 || from == to {
		return in
	}
	outLen := int(float64(len(in)) * float64(to) / float64(from))
	if outLen < 1 {
		outLen = 1
	}
	out := make([]float32, outLen)
	scale := float64(from) / float64(to)
	last := len(in) - 1
	for i := range out {
		pos := float64(i) * scale
		j := int(pos)
		if j >= last {
			out[i] = in[last]
			continue
		}
		frac := float32(pos - float64(j))
		out[i] = in[j]*(1-frac) + in[j+1]*frac
	}
	return out
}
