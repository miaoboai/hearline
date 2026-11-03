package ui

import (
	"testing"
	"time"

	g "github.com/egoist/mygo/ui"

	"github.com/miaoboai/hearline/internal/pipeline"
)

func TestRenderSnapshot(t *testing.T) {
	m := Snapshot([]pipeline.Line{
		{Text: "开饭时间早上9点至下午5点。", Language: "zh", Start: 720 * time.Millisecond, End: 5200 * time.Millisecond},
		{Text: "The tribal chieftain called for the boy.", Language: "en", Start: 1100 * time.Millisecond, End: 3460 * time.Millisecond},
		{Text: "And presented him with 50 pieces of gold.", Language: "en", Start: 3820 * time.Millisecond, End: 7100 * time.Millisecond},
	}, "官方测试 WAV（zh + en）", "SenseVoice-Small Q8 · CPU · 4 threads", "本地优先 · llama.cpp · 未设置模型路径（字幕环不调用）")
	img := g.Render(m.view, frameWidth, frameHeight, 1)
	if img == nil {
		t.Fatal("nil image")
	}
	if img.Bounds().Dx() != frameWidth || img.Bounds().Dy() != frameHeight {
		t.Fatalf("bounds %v", img.Bounds())
	}
	r, gg, b, _ := img.At(16, 16).RGBA()
	if !near8(int(r>>8), 0xF3, 12) || !near8(int(gg>>8), 0xF4, 12) || !near8(int(b>>8), 0xF6, 12) {
		t.Fatalf("ground pixel = #%02x%02x%02x, want about #F3F4F6", r>>8, gg>>8, b>>8)
	}
	var dark, orange int
	bounds := img.Bounds()
	for y := bounds.Min.Y; y < bounds.Max.Y; y += 2 {
		for x := bounds.Min.X; x < bounds.Max.X; x += 2 {
			rr, gg, bb, _ := img.At(x, y).RGBA()
			R, G, B := int(rr>>8), int(gg>>8), int(bb>>8)
			if R < 40 && G < 42 && B < 48 {
				dark++
			}
			if R > 190 && G > 40 && G < 130 && B < 90 && R > G+60 {
				orange++
			}
		}
	}
	if dark < 20000 {
		t.Fatalf("caption stage looks missing, dark samples %d", dark)
	}
	if orange < 30 {
		t.Fatalf("accent control looks missing, orange samples %d", orange)
	}
}

func TestRenderHistoryScroll(t *testing.T) {
	lines := make([]pipeline.Line, 8)
	for i := range lines {
		lines[i] = pipeline.Line{
			Text:     "字幕行",
			Language: "zh",
			Start:    time.Duration(i) * time.Second,
			End:      time.Duration(i+1) * time.Second,
		}
	}
	m := Snapshot(lines, "系统声", "已加载 sensevoice-small-q8 · CPU", "远程 OpenAI 兼容 · gpt-4.1-mini · https://example · key 已保存（字幕环不调用）")
	m.Listening = true
	img := g.Render(m.view, frameWidth, frameHeight, 1)
	if img == nil {
		t.Fatal("nil image")
	}
}

func TestStatusChips(t *testing.T) {
	if got := engineChip("sensevoice-small-q8", ""); got != "SenseVoice · Q8" {
		t.Fatalf("engine chip %q", got)
	}
	if got := cpuChip("SenseVoice-Small Q8 · CPU · 4 threads"); got != "CPU · 4 线程" {
		t.Fatalf("cpu chip %q", got)
	}
	if got := llmChip("本地优先 · llama.cpp · 未设置模型路径（字幕环不调用）"); got != "LLM · 本地" {
		t.Fatalf("local llm chip %q", got)
	}
	if got := llmChip("远程 OpenAI 兼容 · gpt · https://example · key 已保存（字幕环不调用）"); got != "LLM · 远程" {
		t.Fatalf("remote llm chip %q", got)
	}
	if got := footerStatus(&Model{Status: "SenseVoice-Small Q8 · CPU · 4 threads"}); got != "可以开始听" {
		t.Fatalf("footer %q", got)
	}
	if got := deviceNote("pulse: connection refused"); got != "没有连上系统声，WAV 仍然可以播放" {
		t.Fatalf("device note %q", got)
	}
	if got := langLabel("zh"); got != "中文" {
		t.Fatalf("lang %q", got)
	}
	if got := stamp(3820 * time.Millisecond); got != "00:03.8" {
		t.Fatalf("stamp %q", got)
	}
}

func near8(got, want, tol int) bool {
	return got >= want-tol && got <= want+tol
}
