package notes

import (
	"context"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/miaoboai/hearline/internal/config"
	"github.com/miaoboai/hearline/internal/pipeline"
)

func TestSRTSkipsPartials(t *testing.T) {
	body := SRT([]pipeline.Line{
		{Text: "第一句", Start: 720 * time.Millisecond, End: 5200 * time.Millisecond},
		{Text: "半句", Partial: true, Start: time.Second, End: 2 * time.Second},
	})
	if strings.Contains(body, "半句") || !strings.Contains(body, "00:00:00,720 --> 00:00:05,200") {
		t.Fatalf("srt:\n%s", body)
	}
	plain := Plain([]pipeline.Line{{Text: "第一句"}, {Text: "半句", Partial: true}})
	if plain != "第一句\n" {
		t.Fatalf("plain %q", plain)
	}
}

func TestRunnerKeepsLatestOnly(t *testing.T) {
	var n atomic.Int32
	started := make(chan struct{})
	prev := generate
	generate = func(ctx context.Context, _ config.LLM, lines []pipeline.Line) (Doc, error) {
		if n.Add(1) == 1 {
			close(started)
			<-ctx.Done()
			return Doc{}, ctx.Err()
		}
		return Doc{Points: []string{lines[0].Text}}, nil
	}
	t.Cleanup(func() { generate = prev })

	cfg := config.LLM{Backend: "remote", Remote: config.RemoteLLM{BaseURL: "https://example.invalid", Model: "m"}}
	var run Runner
	got := make(chan Result, 1)
	run.Submit(cfg, []pipeline.Line{{Text: "先"}, {Text: "半句", Partial: true}}, func(res Result) {
		got <- res
	})
	select {
	case <-started:
	case <-time.After(2 * time.Second):
		t.Fatal("first request did not start")
	}
	run.Submit(cfg, []pipeline.Line{{Text: "后"}}, func(res Result) {
		got <- res
	})
	select {
	case res := <-got:
		if res.Err != nil {
			t.Fatal(res.Err)
		}
		if res.Gen != 2 || len(res.Doc.Points) != 1 || res.Doc.Points[0] != "后" {
			t.Fatalf("result %+v", res)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("runner did not finish the latest request")
	}
}
