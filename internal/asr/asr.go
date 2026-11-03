// Package asr is the speech-recognizer boundary used by the live subtitle
// loop. SenseVoice-Small is the v0.1 implementation; Qwen3-ASR can be added
// behind the same Recognizer without changing capture or the UI.
package asr

import (
	"errors"
	"regexp"
	"strings"
)

// ErrUnsupported means this build knows the engine name but cannot run it.
var ErrUnsupported = errors.New("asr engine is not available in this build")

// Result is one decode of a PCM utterance.
type Result struct {
	Text       string
	Language   string
	Emotion    string
	Event      string
	Tokens     []string
	Timestamps []float32
}

// Recognizer decodes mono PCM. Implementations must be safe for one caller
// at a time; Hearline uses a single decode goroutine per session.
// DecodePCM may be called repeatedly. Close releases the model.
type Recognizer interface {
	ID() string
	DecodePCM(samples []float32, sampleRate int) (Result, error)
	Close() error
}

var specialToken = regexp.MustCompile(`<\|[^|]*\|>`)

// CleanText removes SenseVoice special tokens and surrounding space.
func CleanText(s string) string {
	s = specialToken.ReplaceAllString(s, "")
	return strings.TrimSpace(s)
}

// CleanTag turns "<|zh|>" into "zh". Other strings are trimmed.
func CleanTag(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "<|")
	s = strings.TrimSuffix(s, "|>")
	return strings.TrimSpace(s)
}
