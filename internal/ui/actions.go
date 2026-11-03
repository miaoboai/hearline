package ui

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/miaoboai/hearline/internal/config"
	"github.com/miaoboai/hearline/internal/fault"
	"github.com/miaoboai/hearline/internal/notes"
	"github.com/miaoboai/hearline/internal/pipeline"
)

const (
	screenListen = iota
	screenSettings
	screenAbout
	screenNotes

	focusSource = "source"
	focusWAV    = "wav"
	focusModel  = "model"
)

func (m *Model) finals() []pipeline.Line {
	if m == nil {
		return nil
	}
	return notes.Finals(m.Lines)
}

func (m *Model) livePartial() (pipeline.Line, bool) {
	if m == nil || len(m.Lines) == 0 {
		return pipeline.Line{}, false
	}
	last := m.Lines[len(m.Lines)-1]
	if !last.Partial || strings.TrimSpace(last.Text) == "" {
		return pipeline.Line{}, false
	}
	return last, true
}

func (m *Model) toggleListen() {
	if m.Listening {
		m.stop()
		return
	}
	if !m.Ready {
		return
	}
	id := m.selectedID()
	if id == "" {
		if path := strings.TrimSpace(m.WAV); path != "" {
			id = "wav:" + path
		}
	}
	m.start(id)
}

func (m *Model) faultAction(code string) {
	switch code {
	case fault.ModelMissing:
		m.Screen = screenSettings
		m.Focus = focusModel
	case fault.CaptureUnavailable:
		m.Screen = screenListen
		if len(m.Labels) == 0 {
			m.Focus = focusWAV
			return
		}
		m.Focus = focusSource
	}
}

func (m *Model) setFault(code string) {
	switch {
	case code == "":
		return
	case m.Fault == fault.ModelMissing && code != fault.ModelMissing:
		return
	case m.Fault == fault.CaptureUnavailable && code == fault.DecodeBacklog:
		return
	}
	m.Fault = code
}

func (m *Model) clearFault(code string) {
	if m.Fault == code {
		m.Fault = ""
	}
}

func (m *Model) export() {
	finals := m.finals()
	if len(finals) == 0 {
		return
	}
	path := strings.TrimSpace(m.ExportPath)
	if path == "" {
		path = defaultExportPath()
		m.ExportPath = path
	}
	body := notes.Markdown(docOrNil(m.Doc), finals)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil && filepath.Dir(path) != "." {
		fmt.Fprintf(os.Stderr, "export: %v\n", err)
		m.ExportNote = "没能写入"
		return
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "export: %v\n", err)
		m.ExportNote = "没能写入"
		return
	}
	m.ExportNote = "已导出"
}

func docOrNil(doc notes.Doc) *notes.Doc {
	if docEmpty(doc) {
		return nil
	}
	return &doc
}

func defaultExportPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "Hearline-纪要.md"
	}
	return filepath.Join(home, "Hearline-纪要.md")
}

func (m *Model) loadDraft(cfg config.Config) {
	m.cfg = cfg
	m.llm = cfg.LLM
	m.LangLabel = langChoice(cfg.ASR.Language)
	m.VADLabel = vadChoice(cfg.ASR.VAD)
	m.ModelDir = cfg.ASR.ModelDir
	m.BaseURL = cfg.LLM.Remote.BaseURL
	m.ModelName = cfg.LLM.Remote.Model
	m.APIKey = cfg.LLM.Remote.APIKey
	if cfg.LLM.Backend == "remote" {
		m.Backend = 1
	}
}

func (m *Model) openNotes() {
	ok, _ := notes.Available(m.llm)
	if !ok {
		return
	}
	m.Screen = screenNotes
	m.wantNotes = true
	if len(m.finals()) == 0 {
		return
	}
	m.generateNotes()
}

func (m *Model) generateNotes() {
	ok, _ := notes.Available(m.llm)
	if !ok || m.win == nil {
		return
	}
	finals := append([]pipeline.Line(nil), m.finals()...)
	if len(finals) == 0 {
		return
	}
	m.notesGen++
	gen := m.notesGen
	cfg := m.llm
	m.NotesBusy = true
	m.NotesErr = ""
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
		defer cancel()
		doc, err := notes.Generate(ctx, cfg, finals)
		if m.closed.Load() || m.win == nil {
			return
		}
		m.win.Update(func() {
			if gen != m.notesGen {
				return
			}
			m.NotesBusy = false
			if err != nil {
				if err.Error() == notes.ReasonLocal || err.Error() == notes.ReasonRemote {
					return
				}
				m.NotesErr = "这次没整理出来"
				return
			}
			m.Doc = doc
			m.NotesErr = ""
		})
	}()
}

func (m *Model) considerNotes() {
	if !m.wantNotes {
		return
	}
	m.generateNotes()
}

func (m *Model) reload() {
	if m.Listening || m.load == nil {
		return
	}
	go m.loadModel()
}
