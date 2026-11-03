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

const galleryStates = "states"

// WriteReviewFrames draws the listen, empty, listening, error, settings, and notes
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
		{"hearline-listening.png", sampleListening()},
		{"hearline-error.png", sampleError()},
		{"hearline-settings.png", sampleSettings()},
		{"hearline-notes.png", sampleNotes()},
		{"hearline-settings-advanced.png", sampleSettingsAdvanced()},
	}
	for _, frame := range frames {
		if err := WritePNG(frame.m, filepath.Join(dir, frame.name)); err != nil {
			return fmt.Errorf("%s: %w", frame.name, err)
		}
	}
	if err := WritePNGSize(sampleEmpty(), filepath.Join(dir, "hearline-narrow.png"), 720, frameHeight); err != nil {
		return fmt.Errorf("hearline-narrow.png: %w", err)
	}
	if err := WritePNGSize(samplePrimaryStates(), filepath.Join(dir, "hearline-primary-states.png"), 880, 220); err != nil {
		return fmt.Errorf("hearline-primary-states.png: %w", err)
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
	m := Snapshot(sampleLines(), "系统声", "已停止", "")
	m.DockNote = "已停止"
	return m
}

func sampleEmpty() *Model {
	return Snapshot(nil, "系统声", "已准备", "")
}

func sampleListening() *Model {
	m := Snapshot(sampleLines()[:1], "系统声", "正在听", "")
	m.Listening = true
	m.ShownFor = 12 * time.Second
	m.Lines = append(m.Lines, pipeline.Line{Text: "还在听这一句", Partial: true})
	return m
}

func sampleError() *Model {
	m := Snapshot(nil, "系统声", "已准备", "")
	m.Ready = false
	m.Fault = fault.ModelMissing
	return m
}

func sampleSettings() *Model {
	m := Snapshot(nil, "系统声", "已准备", "")
	m.Panel = panelSettings
	m.loadDraft(config.Config{
		ASR: config.ASR{Language: "auto", VAD: "mid", NumThreads: 2},
		LLM: config.LLM{Backend: "local", Local: config.LocalLLM{Runtime: "llama.cpp"}},
	})
	m.Threads = 2
	m.ExportPath = "Hearline-字幕.txt"
	return m
}

func sampleNotes() *Model {
	m := Snapshot(sampleLines(), "系统声", "已停止", "")
	m.Panel = panelNotes
	m.NotesOpen = true
	m.llm = config.LLM{Backend: "remote", Remote: config.RemoteLLM{BaseURL: "https://example.invalid/v1", Model: "notes"}}
	m.Backend = 1
	m.BaseURL = m.llm.Remote.BaseURL
	m.ModelName = m.llm.Remote.Model
	m.Doc = notesDocSample()
	return m
}

func notesDocSample() notes.Doc {
	return notes.Doc{Points: []string{"开饭时间早上9点至下午5点。"}}
}

func sampleSettingsAdvanced() *Model {
	m := sampleSettings()
	m.ShowAdvanced = true
	m.Focus = focusModel
	m.ModelDir = "~/.cache/hearline/sensevoice-small-q8"
	return m
}

func samplePrimaryStates() *Model {
	m := sampleEmpty()
	m.Gallery = galleryStates
	return m
}
