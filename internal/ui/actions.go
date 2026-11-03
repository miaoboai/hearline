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
	panelListen = iota
	panelSettings
	panelNotes

	focusSource = "source"
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
	if m.Panel != panelListen {
		return
	}
	if m.Listening {
		m.stop()
		return
	}
	if !m.Ready {
		return
	}
	id := m.selectedID()
	if id == "" {
		m.More = true
		m.DockNote = "没找到系统声，可从文件听写"
		return
	}
	m.start(id)
}

func (m *Model) listenFile() {
	if m.Listening || !m.Ready {
		return
	}
	path := strings.TrimSpace(m.WAV)
	if path == "" {
		m.More = true
		m.DockNote = "先选择一个音频文件"
		return
	}
	m.FileName = filepath.Base(path)
	m.start("wav:" + path)
}

func (m *Model) clearLines() {
	if m == nil || m.Listening {
		return
	}
	m.Lines = nil
	m.Doc = notes.Doc{}
	m.NotesErr = ""
	m.DockNote = ""
	m.clearFault(fault.DecodeBacklog)
}

func (m *Model) faultAction(code string) {
	switch code {
	case fault.ModelMissing:
		m.downloadModel()
	case fault.ModelUnreadable:
		m.Panel = panelListen
		m.reload()
	case fault.CaptureUnavailable:
		m.Panel = panelListen
		m.More = len(m.Labels) == 0
		m.Focus = focusSource
	case fault.FileInvalid:
		m.Panel = panelListen
		m.More = true
		m.browse()
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

func (m *Model) exportKind(kind string) {
	finals := m.finals()
	if len(finals) == 0 {
		return
	}
	var body, ext string
	switch kind {
	case "srt":
		body = notes.SRT(finals)
		ext = ".srt"
	case "md":
		if docEmpty(m.Doc) {
			return
		}
		body = notes.Markdown(docOrNil(m.Doc), finals)
		ext = ".md"
	default:
		body = notes.Plain(finals)
		ext = ".txt"
	}
	path := strings.TrimSpace(m.ExportPath)
	if path == "" {
		path = defaultExportPath()
	}
	path = replaceExt(path, ext)
	m.writeExport(path, body)
}

func replaceExt(path, ext string) string {
	base := strings.TrimSuffix(path, filepath.Ext(path))
	if base == "" {
		base = "Hearline-字幕"
	}
	return base + ext
}

func (m *Model) writeExport(path, body string) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil && filepath.Dir(path) != "." {
		fmt.Fprintf(os.Stderr, "export: %v\n", err)
		m.DockNote = "没能写入"
		return
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		fmt.Fprintf(os.Stderr, "export: %v\n", err)
		m.DockNote = "没能写入"
		return
	}
	m.ExportPath = path
	m.DockNote = "已导出到「" + filepath.Base(path) + "」"
}

func (m *Model) downloadModel() {
	if m.Downloading {
		return
	}
	if m.download == nil {
		m.Panel = panelSettings
		m.Focus = focusModel
		m.DockNote = "先填写模型路径"
		return
	}
	m.Downloading = true
	m.DockNote = "正在下载模型"
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()
		err := m.download(ctx)
		if m.closed.Load() || m.win == nil {
			return
		}
		m.win.Update(func() {
			m.Downloading = false
			if err != nil {
				fmt.Fprintf(os.Stderr, "download: %v\n", err)
				m.DockNote = "没能下载模型"
				m.setFault(fault.ModelMissing)
				return
			}
			m.DockNote = ""
			m.clearFault(fault.ModelMissing)
			m.reload()
		})
	}()
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
		return "Hearline-字幕.txt"
	}
	return filepath.Join(home, "Hearline-字幕.txt")
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
	m.Panel = panelNotes
	m.ConfirmQuit = false
	ok, _ := notes.Available(m.llm)
	if !ok {
		return
	}
	m.wantNotes = true
	if len(m.finals()) == 0 {
		return
	}
	m.generateNotes()
}

func (m *Model) generateNotes() {
	ok, _ := notes.Available(m.llm)
	if !ok {
		return
	}
	finals := append([]pipeline.Line(nil), m.finals()...)
	if len(finals) == 0 {
		return
	}
	cfg := m.llm
	m.NotesBusy = true
	m.NotesErr = ""
	m.notesRun.Submit(cfg, finals, func(res notes.Result) {
		if m.closed.Load() || m.win == nil {
			return
		}
		m.win.Update(func() {
			m.NotesBusy = false
			if res.Err != nil {
				if res.Err.Error() == notes.ReasonLocal || res.Err.Error() == notes.ReasonRemote {
					return
				}
				m.NotesErr = "这次没整理出来"
				return
			}
			m.Doc = res.Doc
			m.NotesErr = ""
		})
	})
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
