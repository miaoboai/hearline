//go:build !cgo

package sensevoice

import (
	"fmt"

	"github.com/miaoboai/hearline/internal/asr"
	"github.com/miaoboai/hearline/internal/config"
)

// Open reports that this binary was built without cgo, so sherpa-onnx is absent.
func Open(cfg config.ASR) (asr.Recognizer, error) {
	return nil, fmt.Errorf("SenseVoice-Small Q8 需要 CGO_ENABLED=1 才能链接 sherpa-onnx（model_dir %s）: %w", cfg.ModelDir, asr.ErrUnsupported)
}
