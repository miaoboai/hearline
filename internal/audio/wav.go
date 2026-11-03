package audio

import (
	"context"
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"time"
)

const maxMonoSamples = SampleRate * 60 * 60 // one hour at 16 kHz

// ReadWAV reads a PCM16 or float32 WAV file. samples are interleaved and
// scaled to about [-1, 1].
func ReadWAV(path string) (samples []float32, sampleRate, channels int, err error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, 0, 0, err
	}
	defer f.Close()
	var riff [12]byte
	if _, err := f.Read(riff[:]); err != nil {
		return nil, 0, 0, fmt.Errorf("%s: %w", path, err)
	}
	if string(riff[0:4]) != "RIFF" || string(riff[8:12]) != "WAVE" {
		return nil, 0, 0, fmt.Errorf("%s: not a RIFF WAVE file", path)
	}
	var (
		format      uint16
		sampleRateU uint32
		channelsU   uint16
		bits        uint16
		data        []byte
	)
	for {
		var hdr [8]byte
		if _, err := f.Read(hdr[:]); err != nil {
			return nil, 0, 0, fmt.Errorf("%s: truncated wav: %w", path, err)
		}
		id := string(hdr[0:4])
		size := binary.LittleEndian.Uint32(hdr[4:8])
		payload := make([]byte, size)
		if _, err := f.Read(payload); err != nil {
			return nil, 0, 0, fmt.Errorf("%s: short chunk %s: %w", path, id, err)
		}
		if size%2 == 1 {
			if _, err := f.Read(make([]byte, 1)); err != nil {
				return nil, 0, 0, fmt.Errorf("%s: pad: %w", path, err)
			}
		}
		switch id {
		case "fmt ":
			if len(payload) < 16 {
				return nil, 0, 0, fmt.Errorf("%s: short fmt chunk", path)
			}
			format = binary.LittleEndian.Uint16(payload[0:2])
			channelsU = binary.LittleEndian.Uint16(payload[2:4])
			sampleRateU = binary.LittleEndian.Uint32(payload[4:8])
			bits = binary.LittleEndian.Uint16(payload[14:16])
		case "data":
			data = payload
		}
		if id == "data" {
			break
		}
	}
	if format == 0 || sampleRateU == 0 || channelsU == 0 || data == nil {
		return nil, 0, 0, fmt.Errorf("%s: missing fmt or data", path)
	}
	switch format {
	case 1, 3:
	default:
		return nil, 0, 0, fmt.Errorf("%s: unsupported wav format %d", path, format)
	}
	ch := int(channelsU)
	sr := int(sampleRateU)
	switch {
	case format == 1 && bits == 16:
		if len(data)%2 != 0 {
			return nil, 0, 0, fmt.Errorf("%s: odd pcm16 size", path)
		}
		n := len(data) / 2
		samples = make([]float32, n)
		for i := 0; i < n; i++ {
			v := int16(binary.LittleEndian.Uint16(data[i*2:]))
			samples[i] = float32(v) / 32768
		}
	case format == 3 && bits == 32:
		if len(data)%4 != 0 {
			return nil, 0, 0, fmt.Errorf("%s: odd float32 size", path)
		}
		n := len(data) / 4
		samples = make([]float32, n)
		for i := 0; i < n; i++ {
			samples[i] = math.Float32frombits(binary.LittleEndian.Uint32(data[i*4:]))
		}
	default:
		return nil, 0, 0, fmt.Errorf("%s: unsupported pcm format %d bits %d", path, format, bits)
	}
	return samples, sr, ch, nil
}

// ReadMono16k reads path and returns mono 16 kHz audio.
func ReadMono16k(path string) ([]float32, error) {
	samples, sr, ch, err := ReadWAV(path)
	if err != nil {
		return nil, err
	}
	mono := Downmix(samples, ch)
	out := Resample(mono, sr, SampleRate)
	if len(out) > maxMonoSamples {
		return nil, fmt.Errorf("%s: longer than one hour; use live capture or a shorter file", path)
	}
	return out, nil
}

// StreamWAV sends ~100 ms chunks. realtime paces them like a microphone.
func StreamWAV(ctx context.Context, path string, realtime bool) (<-chan []float32, error) {
	samples, err := ReadMono16k(path)
	if err != nil {
		return nil, err
	}
	ch := make(chan []float32, 8)
	go func() {
		defer close(ch)
		frame := SampleRate / 10
		if frame < 1 {
			frame = 1
		}
		for i := 0; i < len(samples); i += frame {
			j := i + frame
			if j > len(samples) {
				j = len(samples)
			}
			cp := append([]float32(nil), samples[i:j]...)
			select {
			case <-ctx.Done():
				return
			case ch <- cp:
			}
			if !realtime {
				continue
			}
			d := time.Duration(len(cp)) * time.Second / SampleRate
			timer := time.NewTimer(d)
			select {
			case <-ctx.Done():
				timer.Stop()
				return
			case <-timer.C:
			}
		}
	}()
	return ch, nil
}
