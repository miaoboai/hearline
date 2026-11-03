package llm

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/miaoboai/hearline/internal/config"
)

func TestRemoteRequestShape(t *testing.T) {
	var gotAuth, gotPath, gotBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotPath = r.URL.Path
		b, _ := io.ReadAll(r.Body)
		gotBody = string(b)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":"你好"}}]}`))
	}))
	defer srv.Close()

	cfg := config.Default().LLM
	cfg.Backend = "remote"
	cfg.Remote.BaseURL = srv.URL + "/v1"
	cfg.Remote.APIKey = "sk-test-secret"
	cfg.Remote.Model = "unit-model"
	text, err := Complete(context.Background(), cfg, "hello")
	if err != nil {
		t.Fatal(err)
	}
	if text != "你好" {
		t.Fatalf("text %q", text)
	}
	if gotAuth != "Bearer sk-test-secret" {
		t.Fatalf("auth %q", gotAuth)
	}
	if gotPath != "/v1/chat/completions" {
		t.Fatalf("path %q", gotPath)
	}
	if !strings.Contains(gotBody, `"model":"unit-model"`) {
		t.Fatalf("body %s", gotBody)
	}
	desc := Describe(cfg)
	if strings.Contains(desc, "sk-test-secret") {
		t.Fatalf("description leaked the key: %s", desc)
	}
}

func TestLocalNotLinked(t *testing.T) {
	_, err := Complete(context.Background(), config.Default().LLM, "hello")
	if err == nil || !strings.Contains(err.Error(), "尚未接入") {
		t.Fatalf("err = %v", err)
	}
}

func TestRemoteFallsBackToLocalError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "nope", http.StatusBadGateway)
	}))
	defer srv.Close()
	cfg := config.Default().LLM
	cfg.Backend = "remote"
	cfg.FallbackToLocal = true
	cfg.Remote.BaseURL = srv.URL
	cfg.Remote.Model = "m"
	_, err := Complete(context.Background(), cfg, "hello")
	if err == nil {
		t.Fatal("expected remote error")
	}
	if strings.Contains(err.Error(), "本地回退") || strings.Contains(err.Error(), "尚未接入") {
		t.Fatalf("fell back to a local runner that is not linked: %v", err)
	}
}
