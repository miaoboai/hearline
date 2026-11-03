package ui

import (
	"testing"

	g "github.com/egoist/mygo/ui"

	"github.com/miaoboai/hearline/internal/pipeline"
)

func TestRenderSnapshot(t *testing.T) {
	m := Snapshot([]pipeline.Line{{
		Text:     "开放时间早上9点至下午5点。",
		Language: "zh",
	}}, "官方测试 WAV", "SenseVoice-Small Q8 · CPU", "本地优先 · llama.cpp · 未设置模型路径（字幕环不调用）")
	img := g.Render(m.view, 960, 640, 1)
	if img == nil {
		t.Fatal("nil image")
	}
	if img.Bounds().Dx() != 960 || img.Bounds().Dy() != 640 {
		t.Fatalf("bounds %v", img.Bounds())
	}
	var ink int
	b := img.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y += 4 {
		for x := b.Min.X; x < b.Max.X; x += 4 {
			r, gg, bl, _ := img.At(x, y).RGBA()
			if r|gg|bl != 0 {
				ink++
			}
		}
	}
	if ink < 100 {
		t.Fatalf("frame looks empty, ink samples %d", ink)
	}
}
