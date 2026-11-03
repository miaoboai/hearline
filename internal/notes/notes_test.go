package notes

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/miaoboai/hearline/internal/config"
	"github.com/miaoboai/hearline/internal/pipeline"
)

func TestUnavailableWithoutRemote(t *testing.T) {
	ok, reason := Available(config.Default().LLM)
	if ok || reason != ReasonLocal {
		t.Fatalf("ok=%v reason=%q", ok, reason)
	}
	_, err := Generate(context.Background(), config.Default().LLM, []pipeline.Line{{Text: "你好"}})
	if err == nil || !strings.Contains(err.Error(), ReasonLocal) {
		t.Fatalf("err = %v", err)
	}
}

func TestRemoteNotesParse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer sk-test-secret" {
			t.Errorf("auth %q", r.Header.Get("Authorization"))
		}
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"{\"points\":[\"周五发布\"],\"todos\":[\"确认设计稿\"],\"terms\":[{\"word\":\"SenseVoice\",\"note\":\"本机识别\"}]}"}}]}`))
	}))
	defer srv.Close()
	cfg := config.Default().LLM
	cfg.Backend = "remote"
	cfg.Remote.BaseURL = srv.URL + "/v1"
	cfg.Remote.Model = "unit"
	cfg.Remote.APIKey = "sk-test-secret"
	lines := []pipeline.Line{
		{Text: "半句", Partial: true},
		{Text: "周五发布。设计稿请确认。"},
	}
	doc, err := Generate(context.Background(), cfg, lines)
	if err != nil {
		t.Fatal(err)
	}
	if len(doc.Points) != 1 || doc.Points[0] != "周五发布" {
		t.Fatalf("points %#v", doc.Points)
	}
	if len(doc.Todos) != 1 || len(doc.Terms) != 1 {
		t.Fatalf("doc %#v", doc)
	}
	md := Markdown(&doc, lines)
	if strings.Contains(md, "半句") || strings.Contains(md, "sk-test-secret") {
		t.Fatalf("markdown leaked partial or key:\n%s", md)
	}
	if !strings.Contains(md, "周五发布") || !strings.Contains(md, "SenseVoice") {
		t.Fatalf("markdown:\n%s", md)
	}
}

func TestMarkdownSkipsPartials(t *testing.T) {
	md := Markdown(nil, []pipeline.Line{{Text: "确定的一句"}, {Text: "还在听", Partial: true}})
	if strings.Contains(md, "还在听") || !strings.Contains(md, "确定的一句") {
		t.Fatalf("markdown:\n%s", md)
	}
}
