package config

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestRoundTripRedactsKey(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "hearline", "config.json")
	cfg := Default()
	cfg.LLM.Backend = "remote"
	cfg.LLM.Remote.BaseURL = "https://example.test/v1"
	cfg.LLM.Remote.APIKey = "sk-test-secret"
	cfg.LLM.Remote.Model = "demo-model"
	if err := cfg.Save(path); err != nil {
		t.Fatal(err)
	}
	if runtime.GOOS != "windows" {
		fi, err := os.Stat(path)
		if err != nil {
			t.Fatal(err)
		}
		if fi.Mode().Perm() != 0o600 {
			t.Fatalf("config mode = %o, want 0600", fi.Mode().Perm())
		}
	}
	got, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if got.LLM.Remote.APIKey != "sk-test-secret" {
		t.Fatalf("key not restored: %q", got.LLM.Remote.APIKey)
	}
	if err := got.Validate(); err != nil {
		t.Fatal(err)
	}
	red := got.Redacted()
	if red.LLM.Remote.APIKey != "***" {
		t.Fatalf("redacted key = %q", red.LLM.Remote.APIKey)
	}
	if got.LLM.Remote.APIKey != "sk-test-secret" {
		t.Fatal("Redacted mutated the original key")
	}
}

func TestMissingFileIsDefault(t *testing.T) {
	cfg, err := Load(filepath.Join(t.TempDir(), "missing.json"))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.ASR.Engine != "sensevoice" || cfg.LLM.Backend != "local" {
		t.Fatalf("defaults: %+v", cfg)
	}
	if cfg.ASR.Threads() < 1 || cfg.ASR.Threads() > 4 {
		t.Fatalf("threads %d", cfg.ASR.Threads())
	}
}

func TestRejectsNonCPUProvider(t *testing.T) {
	cfg := Default()
	cfg.ASR.Provider = "cuda"
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected cuda to be rejected")
	}
}
