package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/miaoboai/hearline/internal/config"
	"github.com/miaoboai/hearline/internal/fault"
	"github.com/miaoboai/hearline/internal/notes"
	"github.com/miaoboai/hearline/internal/pipeline"
)

// WriteReviewFrames draws the listen, empty, error, settings, about, and notes
// surfaces. main may be nil; a sample transcript is used then.
func WriteReviewFrames(dir string, main *Model) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	if main == nil {
		main = sampleListen()
	}
	frames := []struct {
		name string
		m    *Model
	}{
		{"hearline-subs.png", main},
		{"hearline-empty.png", sampleEmpty()},
		{"hearline-error.png", sampleError()},
		{"hearline-settings.png", sampleSettings()},
		{"hearline-about.png", sampleAbout()},
		{"hearline-notes.png", sampleNotes()},
	}
	for _, frame := range frames {
		if err := WritePNG(frame.m, filepath.Join(dir, frame.name)); err != nil {
			return fmt.Errorf("%s: %w", frame.name, err)
		}
	}
	return nil
}

func sampleLines() []pipeline.Line {
	return []pipeline.Line{
		{Text: "开饭时间早上9点至下午5点。", Language: "zh", Start: 720 * time.Millisecond, End: 5200 * time.Millisecond},
		{Text: "The tribal chieftain called for the boy.", Language: "en", Start: 1100 * time.Millisecond, End: 3460 * time.Millisecond},
		{Text: "And presented him with 50 pieces of gold.", Language: "en", Start: 3820 * time.Millisecond, End: 7100 * time.Millisecond},
	}
}

func sampleListen() *Model {
	m := Snapshot(sampleLines(), "官方测试 WAV（zh + en）", "已准备", "")
	m.WAV = "models/…/test_wavs/zh.wav"
	return m
}

func sampleEmpty() *Model {
	m := Snapshot(nil, "系统声", "已准备", "")
	m.WAV = ""
	return m
}

func sampleError() *Model {
	m := Snapshot(nil, "系统声", "已准备", "")
	m.Ready = false
	m.Fault = fault.ModelMissing
	return m
}

func sampleSettings() *Model {
	m := Snapshot(nil, "", "已准备", "")
	m.Screen = screenSettings
	m.loadDraft(config.Config{
		ASR: config.ASR{Language: "auto", VAD: "mid"},
		LLM: config.LLM{Backend: "local", Local: config.LocalLLM{Runtime: "llama.cpp"}},
	})
	m.ExportPath = "Hearline-纪要.md"
	return m
}

func sampleAbout() *Model {
	m := sampleSettings()
	m.Screen = screenAbout
	return m
}

func sampleNotes() *Model {
	m := Snapshot(sampleLines(), "官方测试 WAV（zh + en）", "已准备", "")
	m.Screen = screenNotes
	m.wantNotes = true
	m.llm = config.LLM{Backend: "remote", Remote: config.RemoteLLM{BaseURL: "https://example.invalid/v1", Model: "notes"}}
	m.Doc = notes.Doc{
		Points: []string{"开饭时间是早上 9 点到下午 5 点。", "族长召来了男孩。"},
		Todos:  []string{"把 50 块金子交给男孩。"},
		Terms:  []notes.Term{{Word: "tribal chieftain", Note: "族长"}},
	}
	return m
}
