// Package config loads Hearline settings. ASR stays on-device. The LLM
// section is local-first, with an optional OpenAI-compatible remote that
// the v0.1 subtitle loop does not call.
package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// Config is the on-disk settings file.
type Config struct {
	Path string `json:"-"`
	ASR  ASR    `json:"asr"`
	LLM  LLM    `json:"llm"`
}

// ASR selects the recognizer. Engine "sensevoice" is the v0.1 default.
// "qwen3" is accepted by the config and rejected by the runner until that
// model is linked. "mock" feeds the UI without a model.
type ASR struct {
	Engine     string  `json:"engine"`
	ModelDir   string  `json:"model_dir,omitempty"`
	Language   string  `json:"language"`
	UseITN     bool    `json:"use_itn"`
	NumThreads int     `json:"num_threads"`
	Provider   string  `json:"provider"`
	ChunkSec   float64 `json:"chunk_seconds"`
}

// LLM chooses where translation and notes will run later.
// Backend is "local" or "remote".
type LLM struct {
	Backend         string    `json:"backend"`
	FallbackToLocal bool      `json:"fallback_to_local"`
	Local           LocalLLM  `json:"local"`
	Remote          RemoteLLM `json:"remote"`
}

// LocalLLM names the on-device model. The llama.cpp runner is not linked
// in v0.1; the path is stored so a later release can load it.
type LocalLLM struct {
	ModelPath string `json:"model_path,omitempty"`
	Runtime   string `json:"runtime"`
}

// RemoteLLM is an OpenAI-compatible chat endpoint.
// APIKey is stored only in the user's config file (mode 0600) and must
// not be logged.
type RemoteLLM struct {
	BaseURL string `json:"base_url,omitempty"`
	APIKey  string `json:"api_key,omitempty"`
	Model   string `json:"model,omitempty"`
}

// Default returns the v0.1 defaults: SenseVoice on CPU, local LLM.
func Default() Config {
	return Config{
		ASR: ASR{
			Engine:     "sensevoice",
			Language:   "auto",
			UseITN:     true,
			NumThreads: 0,
			Provider:   "cpu",
			ChunkSec:   8,
		},
		LLM: LLM{
			Backend:         "local",
			FallbackToLocal: true,
			Local: LocalLLM{
				Runtime: "llama.cpp",
			},
		},
	}
}

// DefaultPath is ~/.config/hearline/config.json (or the platform equivalent).
func DefaultPath() string {
	if p := os.Getenv("HEARLINE_CONFIG"); p != "" {
		return p
	}
	dir, err := os.UserConfigDir()
	if err != nil {
		return filepath.Join(".", "hearline-config.json")
	}
	return filepath.Join(dir, "hearline", "config.json")
}

// DefaultThreads is a CPU-only cap aimed at a 16 GiB machine.
// Eight cores use 4 threads; smaller machines use fewer so ORT does not
// multiply activation memory.
func DefaultThreads() int {
	n := runtime.NumCPU()
	switch {
	case n >= 8:
		return 4
	case n >= 4:
		return 2
	default:
		return 1
	}
}

// Threads resolves NumThreads and clamps it to 1..4.
func (a ASR) Threads() int {
	n := a.NumThreads
	if n <= 0 {
		n = DefaultThreads()
	}
	if n > 4 {
		n = 4
	}
	if n < 1 {
		n = 1
	}
	return n
}

// Load reads path. A missing file returns Default without creating it.
func Load(path string) (Config, error) {
	if path == "" {
		path = DefaultPath()
	}
	cfg := Default()
	cfg.Path = path
	b, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return cfg, nil
	}
	if err != nil {
		return cfg, err
	}
	if err := json.Unmarshal(b, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse %s: %w", path, err)
	}
	cfg.Path = path
	cfg.applyDefaults()
	return cfg, nil
}

func (c *Config) applyDefaults() {
	d := Default()
	if c.ASR.Engine == "" {
		c.ASR.Engine = d.ASR.Engine
	}
	if c.ASR.Language == "" {
		c.ASR.Language = d.ASR.Language
	}
	if c.ASR.Provider == "" {
		c.ASR.Provider = "cpu"
	}
	if c.ASR.ChunkSec <= 0 {
		c.ASR.ChunkSec = d.ASR.ChunkSec
	}
	if c.LLM.Backend == "" {
		c.LLM.Backend = d.LLM.Backend
	}
	if c.LLM.Local.Runtime == "" {
		c.LLM.Local.Runtime = d.LLM.Local.Runtime
	}
}

// Validate checks the LLM switch. Remote needs base_url and model.
// An empty API key is allowed for keyless local gateways.
func (c Config) Validate() error {
	switch c.ASR.Engine {
	case "", "sensevoice", "mock", "qwen3":
	default:
		return fmt.Errorf("unknown asr engine %q", c.ASR.Engine)
	}
	if c.ASR.Provider != "" && c.ASR.Provider != "cpu" {
		return fmt.Errorf("v0.1 ASR provider must be cpu, got %q", c.ASR.Provider)
	}
	switch c.LLM.Backend {
	case "", "local":
		return nil
	case "remote":
		if c.LLM.Remote.BaseURL == "" || c.LLM.Remote.Model == "" {
			return errors.New("remote LLM requires base_url and model")
		}
		return nil
	default:
		return fmt.Errorf("unknown llm backend %q", c.LLM.Backend)
	}
}

// Save writes the config with mode 0600. The directory is mode 0700.
func (c Config) Save(path string) error {
	if path == "" {
		path = c.Path
	}
	if path == "" {
		path = DefaultPath()
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	b, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	b = append(b, '\n')
	return os.WriteFile(path, b, 0o600)
}

// Redacted is a copy safe to print. A non-empty API key becomes "***".
func (c Config) Redacted() Config {
	out := c
	if out.LLM.Remote.APIKey != "" {
		out.LLM.Remote.APIKey = "***"
	}
	return out
}
