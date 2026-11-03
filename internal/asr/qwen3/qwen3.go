// Package qwen3 reserves the Qwen3-ASR plug-in point.
//
// v0.1 does not download or run Qwen3. When a runner is added, implement
// Recognizer here — sherpa-onnx exposes it as OfflineQwen3ASRModelConfig on
// the same offline recognizer SenseVoice uses — and register the engine
// name "qwen3" without changing capture, the subtitle pipeline, or the UI.
package qwen3

import (
	"fmt"

	"github.com/miaoboai/hearline/internal/asr"
	"github.com/miaoboai/hearline/internal/config"
)

// New reports that Qwen3-ASR is not linked in this release.
func New(cfg config.ASR) (asr.Recognizer, error) {
	hint := cfg.ModelDir
	if hint == "" {
		hint = "(no model_dir)"
	}
	return nil, fmt.Errorf("qwen3-asr is not linked in v0.1 (model_dir %s): %w", hint, asr.ErrUnsupported)
}
