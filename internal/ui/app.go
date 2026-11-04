// Package ui is the MyGo native window: pick a source, capture, and show
// subtitle lines. Rendering a frame does not need a display; WritePNG uses
// that for the headless demo.
package ui

import (
	"context"
	"fmt"
	"image/png"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"github.com/egoist/mygo"
	mygoui "github.com/egoist/mygo/ui"

	"github.com/miaoboai/hearline/internal/asr"
	"github.com/miaoboai/hearline/internal/audio"
	"github.com/miaoboai/hearline/internal/config"
	"github.com/miaoboai/hearline/internal/fault"
	"github.com/miaoboai/hearline/internal/notes"
	"github.com/miaoboai/hearline/internal/pipeline"
)

// Options wires the window to capture and recognition without this package
// importing sherpa-onnx.
type Options struct {
	LoadRecognizer func() (asr.Recognizer, error)
	ListSources    func() ([]audio.Source, error)
	Capture        func(context.Context, string) (<-chan []float32, error)
	LLMStatus      string
	Chunk          pipeline.Options
	Config         config.Config
	Version        string
	SaveConfig     func(config.Config) error
	DownloadModel  func(context.Context) error
}

// Model is the subtitle window state. The view reads it on the UI thread.
type Model struct {
	Sources    []audio.Source
	Labels     []string
	Selected   string
	WAV        string
	Status     string
	DeviceNote string
	LLMStatus  string
	Engine     string
	Lines      []pipeline.Line
	Listening  bool
	Ready      bool
	Screen     int
	Fault      string
	Focus      string
	Version    string
	ExportPath string
	ExportNote string
	SaveNote   string
	LangLabel  string
	VADLabel   string
	ModelDir   string
	Backend    int
	BaseURL    string
	ModelName  string
	APIKey     string
	Doc        notes.Doc
	NotesBusy  bool
	NotesErr   string
	// follow is the line count last scrolled into view.
	follow       int
	notesGen     int
	wantNotes    bool
	reloadASR    bool
	loadGen      atomic.Uint64
	Panel        int
	More         bool
	ConfirmQuit  bool
	allowClose   bool
	OnTop        bool
	Started      time.Time
	ShownFor     time.Duration
	FileName     string
	DockNote     string
	Downloading  bool
	Threads      int
	ShowAdvanced bool
	ExportOpen   bool
	NotesOpen    bool
	Gallery      string
	notesRun     notes.Runner

	llm  config.LLM
	cfg  config.Config
	save func(config.Config) error

	load     func() (asr.Recognizer, error)
	capture  func(context.Context, string) (<-chan []float32, error)
	download func(context.Context) error
	chunk    pipeline.Options

	win    *mygo.Window
	rec    asr.Recognizer
	cancel context.CancelFunc
	closed atomic.Bool
}

// Run opens the window and blocks until it closes.
func Run(opt Options) error {
	m := &Model{
		load:       opt.LoadRecognizer,
		capture:    opt.Capture,
		chunk:      opt.Chunk,
		LLMStatus:  opt.LLMStatus,
		Status:     "正在准备",
		Version:    opt.Version,
		ExportPath: defaultExportPath(),
		save:       opt.SaveConfig,
		download:   opt.DownloadModel,
		Threads:    opt.Config.ASR.Threads(),
	}
	m.loadDraft(opt.Config)
	if m.load == nil || m.capture == nil {
		return fmt.Errorf("ui: missing recognizer or capture")
	}
	m.refresh(opt.ListSources)
	mygo.App.WhenReady(func() {
		win := mygo.NewWindow(mygo.WindowOptions{
			Title:           "Hearline",
			Width:           frameWidth,
			Height:          frameHeight,
			MinWidth:        720,
			MinHeight:       640,
			BackgroundColor: "#F3F4F6",
			Content:         mygoui.View(m.view),
		})
		m.win = win
		win.OnClose(func(e *mygo.CloseEvent) {
			if m.Listening && !m.allowClose {
				e.PreventDefault()
				win.Update(func() { m.ConfirmQuit = true })
			}
		})
		win.OnClosed(func() {
			m.closed.Store(true)
			m.notesRun.Stop()
			m.stop()
		})
		go m.loadModel()
	})
	return mygo.App.Run()
}

// WritePNG renders one frame of the subtitle surface.
func WritePNG(m *Model, path string) error {
	return WritePNGSize(m, path, frameWidth, frameHeight)
}

func WritePNGSize(m *Model, path string, w, h int) error {
	img := mygoui.Render(m.view, w, h, 1)
	if img == nil {
		return fmt.Errorf("subtitle surface did not render")
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

func (m *Model) loadModel() {
	gen := m.loadGen.Add(1)
	rec, err := m.load()
	stale := func() bool {
		return m.closed.Load() || gen != m.loadGen.Load()
	}
	if stale() {
		if rec != nil {
			_ = rec.Close()
		}
		return
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "recognizer: %v\n", err)
		if rec != nil {
			_ = rec.Close()
		}
		code := fault.Classify(err.Error())
		if code == "" {
			code = fault.ModelMissing
		}
		if m.win == nil {
			return
		}
		m.win.Update(func() {
			if stale() {
				return
			}
			m.Ready = false
			m.setFault(code)
		})
		return
	}
	if m.win == nil {
		_ = rec.Close()
		return
	}
	m.win.Update(func() {
		if stale() {
			go rec.Close()
			return
		}
		if m.Listening {
			m.reloadASR = true
			go rec.Close()
			return
		}
		if m.rec != nil {
			old := m.rec
			go old.Close()
		}
		m.rec = rec
		m.Ready = true
		m.Engine = rec.ID()
		m.Status = "已准备"
		m.clearFault(fault.ModelMissing)
	})
}

func (m *Model) refresh(list func() ([]audio.Source, error)) {
	if list == nil {
		list = audio.List
	}
	sources, err := list()
	m.Sources = sources
	m.Labels = make([]string, len(sources))
	seen := map[string]int{}
	for i, s := range sources {
		label := s.Name
		if label == "" {
			label = s.ID
		}
		seen[label]++
		if seen[label] > 1 {
			label = fmt.Sprintf("%s (%d)", label, seen[label])
		}
		m.Labels[i] = label
	}
	if !contains(m.Labels, m.Selected) {
		m.Selected = prefer(sources, m.Labels)
	}
	if err != nil {
		m.DeviceNote = err.Error()
		code := fault.Classify(err.Error())
		if code == "" {
			code = fault.CaptureUnavailable
		}
		m.setFault(code)
		return
	}
	m.DeviceNote = ""
	m.clearFault(fault.CaptureUnavailable)
}

func prefer(sources []audio.Source, labels []string) string {
	for i, s := range sources {
		if s.Kind == "system" {
			return labels[i]
		}
	}
	// Prefer the built-in MacBook mic when several inputs are present.
	for i, s := range sources {
		if s.Kind == "mic" && (containsFold(s.Name, "MacBook") || containsFold(s.Name, "Built-in")) {
			return labels[i]
		}
	}
	for i, s := range sources {
		if s.Kind == "mic" {
			return labels[i]
		}
	}
	if len(labels) > 0 {
		return labels[0]
	}
	return ""
}

func containsFold(s, sub string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(sub))
}

func contains(list []string, s string) bool {
	for _, item := range list {
		if item == s {
			return true
		}
	}
	return false
}

func (m *Model) selectedID() string {
	for i, label := range m.Labels {
		if label == m.Selected && i < len(m.Sources) {
			return m.Sources[i].ID
		}
	}
	return ""
}

func (m *Model) start(id string) {
	if m.Listening || !m.Ready || m.rec == nil || m.win == nil {
		return
	}
	if id == "" {
		m.More = true
		m.DockNote = "先选择一个音频文件"
		return
	}
	m.clearFault(fault.DecodeBacklog)
	m.clearFault(fault.CaptureUnavailable)
	m.clearFault(fault.FileInvalid)
	m.DockNote = ""
	m.ConfirmQuit = false
	m.Started = time.Now()
	m.ShownFor = 0
	ctx, cancel := context.WithCancel(context.Background())
	m.cancel = cancel
	m.Listening = true
	m.Status = "正在听"
	go m.session(ctx, id)
}

func (m *Model) session(ctx context.Context, id string) {
	defer func() {
		if m.win != nil && !m.closed.Load() {
			m.win.Update(func() {
				m.Listening = false
				if m.cancel != nil {
					m.cancel = nil
				}
				if m.Status == "正在听 "+id || m.closed.Load() {
					m.Status = "已停止"
				}
				m.maybeReload()
			})
		}
	}()
	ch, err := m.capture(ctx, id)
	if err != nil {
		m.reportCapture(id, err.Error())
		return
	}
	err = pipeline.Run(ctx, m.rec, ch, m.chunk, func(line pipeline.Line) {
		if m.closed.Load() || m.win == nil {
			return
		}
		m.win.Update(func() { m.push(line) })
	})
	if err != nil && ctx.Err() == nil {
		m.reportDecode(err.Error())
	}
}

func (m *Model) reportCapture(id, msg string) {
	if m.closed.Load() || m.win == nil {
		return
	}
	fmt.Fprintf(os.Stderr, "capture: %s\n", msg)
	code := fault.Classify(msg)
	if len(id) > 4 && id[:4] == "wav:" && code != fault.Permission {
		if code == "" || code == fault.ModelMissing || code == fault.CaptureUnavailable {
			code = fault.FileInvalid
		}
	}
	if code == "" {
		code = fault.CaptureUnavailable
	}
	m.win.Update(func() { m.setFault(code) })
}

func (m *Model) reportDecode(msg string) {
	if m.closed.Load() || m.win == nil {
		return
	}
	fmt.Fprintf(os.Stderr, "decode: %s\n", msg)
	code := fault.Classify(msg)
	if code == "" {
		return
	}
	m.win.Update(func() { m.setFault(code) })
}

func (m *Model) push(line pipeline.Line) {
	if line.Text == "" {
		return
	}
	if n := len(m.Lines); n > 0 && m.Lines[n-1].Partial {
		m.Lines[n-1] = line
	} else {
		m.Lines = append(m.Lines, line)
	}
	if len(m.Lines) > 200 {
		m.Lines = m.Lines[len(m.Lines)-200:]
	}
	if line.Slow {
		m.setFault(fault.DecodeBacklog)
	}
	if !line.Partial {
		m.considerNotes()
	}
}

func (m *Model) stop() {
	if m.cancel != nil {
		m.cancel()
		m.cancel = nil
	}
	m.Listening = false
	m.maybeReload()
}

func (m *Model) maybeReload() {
	if m == nil || m.Listening || !m.reloadASR {
		return
	}
	m.reloadASR = false
	m.reload()
}

const (
	frameWidth  = 1120
	frameHeight = 760
)

// Snapshot is a display-only model used by the headless screenshot.
func Snapshot(lines []pipeline.Line, source, status, llm string) *Model {
	m := &Model{
		Lines:      lines,
		Status:     status,
		LLMStatus:  llm,
		Engine:     "sensevoice-small-q8",
		Ready:      true,
		Selected:   source,
		Version:    "0.1",
		ExportPath: "Hearline-字幕.txt",
		LangLabel:  "自动",
		VADLabel:   "标准",
		Threads:    2,
	}
	m.llm = config.LLM{Backend: "local", Local: config.LocalLLM{Runtime: "llama.cpp"}}
	if source != "" {
		m.Labels = []string{source}
		m.Sources = []audio.Source{{ID: "system", Name: source, Kind: "system"}}
	}
	return m
}
