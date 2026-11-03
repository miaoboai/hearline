// Package engine constructs the recognizer named in config.
package engine

import (
	"fmt"

	"github.com/miaoboai/hearline/internal/asr"
	"github.com/miaoboai/hearline/internal/asr/mock"
	"github.com/miaoboai/hearline/internal/asr/qwen3"
	"github.com/miaoboai/hearline/internal/asr/sensevoice"
	"github.com/miaoboai/hearline/internal/config"
)

// New returns the configured recognizer.
func New(cfg config.ASR) (asr.Recognizer, error) {
	switch cfg.Engine {
	case "", "sensevoice":
		return sensevoice.Open(cfg)
	case "mock":
		return mock.New(), nil
	case "qwen3":
		return qwen3.New(cfg)
	default:
		return nil, fmt.Errorf("unknown asr engine %q", cfg.Engine)
	}
}
