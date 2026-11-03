package ui

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	g "github.com/egoist/mygo/ui"

	"github.com/miaoboai/hearline/internal/config"
	"github.com/miaoboai/hearline/internal/fault"
	"github.com/miaoboai/hearline/internal/notes"
	"github.com/miaoboai/hearline/internal/pipeline"
)

func TestRenderSnapshot(t *testing.T) {
	m := Snapshot([]pipeline.Line{
		{Text: "开饭时间早上9点至下午5点。", Language: "zh", Start: 720 * time.Millisecond},
		{Text: "还在听", Partial: true},
		{Text: "And presented him with 50 pieces of gold.", Language: "en", Start: 3820 * time.Millisecond},
	}, "官方测试 WAV（zh + en）", "已准备", "")
	if len(m.finals()) != 2 {
		t.Fatalf("finals %d", len(m.finals()))
	}
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

func TestRenderStates(t *testing.T) {
	for _, m := range []*Model{sampleEmpty(), sampleListening(), sampleError(), sampleSettings(), sampleNotes()} {
		if g.Render(m.view, frameWidth, frameHeight, 1) == nil {
			t.Fatal("nil frame")
		}
	}
	if sampleError().Fault != fault.ModelMissing {
		t.Fatal("error frame lost its code")
	}
	ok, reason := notes.Available(sampleListen().llm)
	if ok || !strings.Contains(reason, "尚未接入") {
		t.Fatalf("default notes ok=%v reason=%q", ok, reason)
	}
	if sampleListening().clock() != "00:12" {
		t.Fatalf("clock %s", sampleListening().clock())
	}
}

func TestDumpFrames(t *testing.T) {
	dir := t.TempDir()
	if env := os.Getenv("HEARLINE_FRAMES"); env != "" {
		dir = env
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := WriteReviewFrames(dir, nil); err != nil {
		t.Fatal(err)
	}
	partial := sampleListen()
	partial.Lines = append(partial.Lines, pipeline.Line{Text: "还在听这一句", Partial: true})
	if err := WritePNG(partial, filepath.Join(dir, "hearline-partial.png")); err != nil {
		t.Fatal(err)
	}
	remote := sampleSettings()
	remote.Backend = 1
	remote.BaseURL = "https://example.invalid/v1"
	remote.ModelName = "notes"
	remote.llm.Backend = "remote"
	remote.llm.Remote = config.RemoteLLM{BaseURL: remote.BaseURL, Model: remote.ModelName}
	if err := WritePNG(remote, filepath.Join(dir, "hearline-settings-remote.png")); err != nil {
		t.Fatal(err)
	}
	about := sampleSettings()
	about.Panel = panelListen
	about.llm.Backend = "remote"
	about.llm.Remote = config.RemoteLLM{BaseURL: "https://example.invalid/v1", Model: "notes"}
	if err := WritePNG(about, filepath.Join(dir, "hearline-about-remote.png")); err != nil {
		t.Fatal(err)
	}
	miss := sampleEmpty()
	miss.Fault = fault.CaptureUnavailable
	if err := WritePNG(miss, filepath.Join(dir, "hearline-capture.png")); err != nil {
		t.Fatal(err)
	}
	t.Log(dir)
}

func TestPrivacyFollowsBackend(t *testing.T) {
	if privacyAudio != "听写在本机完成，音频不会上传。" || privacyRemote != "若纪要选用远程服务，只会发送已识别的文字，不会发送音频。" || privacyOff != "当前纪要未启用远程" {
		t.Fatal("privacy copy drifted")
	}
	settings := sampleSettings()
	if settings.remoteDisclosure() {
		t.Fatal("local settings should mute the remote line")
	}
	settings.Backend = 1
	if !settings.remoteDisclosure() {
		t.Fatal("settings toggle should show the remote sentence before save")
	}
	about := sampleSettings()
	about.Panel = panelListen
	if about.remoteDisclosure() {
		t.Fatal("listen surface follows the saved backend")
	}
	about.Backend = 1
	if about.remoteDisclosure() {
		t.Fatal("listen surface must not follow an unsaved settings toggle")
	}
	about.Panel = panelSettings
	if !about.remoteDisclosure() {
		t.Fatal("settings toggle should show the remote sentence before save")
	}
	about.Panel = panelListen
	about.llm.Backend = "remote"
	if !about.remoteDisclosure() {
		t.Fatal("saved remote backend should disclose text leaving the machine")
	}
	if asrNeedsReload(config.ASR{Language: "auto"}, config.ASR{Language: "zh"}) != true {
		t.Fatal("language change should reload")
	}
	if asrNeedsReload(config.ASR{Language: "zh", VAD: "mid"}, config.ASR{Language: "zh", VAD: "high"}) {
		t.Fatal("sensitivity is applied on the chunker, not by reloading the model")
	}
}

func TestDownloadModelIsWired(t *testing.T) {
	if fault.Action(fault.ModelMissing) != "下载模型" {
		t.Fatal("model missing action drifted")
	}
	m := sampleError()
	called := make(chan struct{}, 1)
	m.download = func(context.Context) error {
		called <- struct{}{}
		return context.Canceled
	}
	tt := g.NewTester(m.view, frameWidth, frameHeight)
	if err := tt.Click("下载模型"); err != nil {
		t.Fatal(err)
	}
	if !m.Downloading {
		t.Fatal("下载模型 did not start")
	}
	select {
	case <-called:
	case <-time.After(2 * time.Second):
		t.Fatal("下载模型 did not call DownloadModel")
	}
	plain := sampleError()
	tt = g.NewTester(plain.view, frameWidth, frameHeight)
	if err := tt.Click("下载模型"); err != nil {
		t.Fatal(err)
	}
	if plain.Panel != panelSettings || plain.SaveNote != modelCacheNote {
		t.Fatalf("panel %d note %q", plain.Panel, plain.SaveNote)
	}
	if !tt.HasText("模型路径") || !tt.HasText(modelCacheNote) {
		t.Fatal("下载模型 did not open the model field")
	}
	tt.Type("models/sensevoice")
	if plain.ModelDir != "models/sensevoice" {
		t.Fatalf("model field did not take focus, dir %q", plain.ModelDir)
	}
}

func TestFootnoteQuietOnFault(t *testing.T) {
	m := sampleError()
	if got := m.footnote(); got != "" {
		t.Fatalf("error footnote %q", got)
	}
	cold := sampleEmpty()
	cold.Ready = false
	if got := cold.footnote(); got != "正在加载识别器" {
		t.Fatalf("cold footnote %q", got)
	}
}

func TestLineMetaAndStamp(t *testing.T) {
	if got := stamp(3820 * time.Millisecond); got != "00:03.8" {
		t.Fatalf("stamp %q", got)
	}
	if got := lineMeta(pipeline.Line{Language: "zh", Start: time.Second}); got != "中文 · 00:01.0" {
		t.Fatalf("meta %q", got)
	}
}

func near8(got, want, tol int) bool {
	return got >= want-tol && got <= want+tol
}
