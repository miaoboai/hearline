// Package mock is a CPU-free recognizer used to wire the subtitle loop
// when the SenseVoice model is not loaded. It is not a speech model.
package mock

import (
	"fmt"

	"github.com/miaoboai/hearline/internal/asr"
	"github.com/miaoboai/hearline/internal/audio"
)

// Recognizer reports the duration of each non-silent buffer.
type Recognizer struct{}

// New returns a mock recognizer.
func New() *Recognizer { return &Recognizer{} }

// ID reports the engine name.
func (r *Recognizer) ID() string { return "mock" }

// DecodePCM returns a placeholder line when the buffer has audible energy.
func (r *Recognizer) DecodePCM(samples []float32, sampleRate int) (asr.Result, error) {
	if sampleRate <= 0 {
		sampleRate = audio.SampleRate
	}
	if len(samples) == 0 || audio.RMS(samples) < 0.01 {
		return asr.Result{}, nil
	}
	sec := float64(len(samples)) / float64(sampleRate)
	return asr.Result{
		Text:     fmt.Sprintf("mock %.2fs", sec),
		Language: "und",
		Event:    "mock",
	}, nil
}

// Close releases nothing.
func (r *Recognizer) Close() error { return nil }
