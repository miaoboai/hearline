// Package ui is the MyGo native window: pick a source, capture, and show
// subtitle lines. Rendering a frame does not need a display; WritePNG uses
// that for the headless demo.
package ui

import (
	"context"
	"fmt"
	"image/png"
	"os"
	"sync/atomic"

	"github.com/egoist/mygo"
	mygoui "github.com/egoist/mygo/ui"

	"github.com/miaoboai/hearline/internal/asr"
	"github.com/miaoboai/hearline/internal/audio"
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

	load    func() (asr.Recognizer, error)
	capture func(context.Context, string) (<-chan []float32, error)
	chunk   pipeline.Options

	win    *mygo.Window
	rec    asr.Recognizer
	cancel context.CancelFunc
	closed atomic.Bool
}

// Run opens the window and blocks until it closes.
func Run(opt Options) error {
	m := &Model{
		load:      opt.LoadRecognizer,
		capture:   opt.Capture,
		chunk:     opt.Chunk,
		LLMStatus: opt.LLMStatus,
		Status:    "正在加载识别器…",
	}
	if m.load == nil || m.capture == nil {
		return fmt.Errorf("ui: missing recognizer or capture")
	}
	m.refresh(opt.ListSources)
	mygo.App.WhenReady(func() {
		win := mygo.NewWindow(mygo.WindowOptions{
			Title:     "Hearline",
			Width:     880,
			Height:    640,
			MinWidth:  640,
			MinHeight: 480,
			Content:   mygoui.View(m.view),
		})
		m.win = win
		win.OnClosed(func() {
			m.closed.Store(true)
			m.stop()
		})
		go m.loadModel()
	})
	return mygo.App.Run()
}

// WritePNG renders one frame of the subtitle surface.
func WritePNG(m *Model, path string) error {
	img := mygoui.Render(m.view, 960, 640, 1)
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
	rec, err := m.load()
	if m.closed.Load() {
		if rec != nil {
			_ = rec.Close()
		}
		return
	}
	m.win.Update(func() {
		if err != nil {
			m.Status = "识别器未就绪: " + err.Error()
			return
		}
		m.rec = rec
		m.Ready = true
		m.Engine = rec.ID()
		m.Status = "已加载 " + rec.ID() + " · CPU"
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
	} else {
		m.DeviceNote = ""
	}
}

func prefer(sources []audio.Source, labels []string) string {
	for i, s := range sources {
		if s.Kind == "system" {
			return labels[i]
		}
	}
	if len(labels) > 0 {
		return labels[0]
	}
	return ""
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
		m.Status = "请选择音源，或填写 WAV 路径后播放"
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	m.cancel = cancel
	m.Listening = true
	m.Status = "正在听 " + id
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
			})
		}
	}()
	ch, err := m.capture(ctx, id)
	if err != nil {
		m.report(err.Error())
		return
	}
	err = pipeline.Run(ctx, m.rec, ch, m.chunk, func(line pipeline.Line) {
		if m.closed.Load() || m.win == nil {
			return
		}
		m.win.Update(func() { m.push(line) })
	})
	if err != nil && ctx.Err() == nil {
		m.report(err.Error())
	}
}

func (m *Model) report(msg string) {
	if m.closed.Load() || m.win == nil {
		return
	}
	m.win.Update(func() { m.Status = msg })
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
}

func (m *Model) stop() {
	if m.cancel != nil {
		m.cancel()
		m.cancel = nil
	}
	m.Listening = false
}

// Snapshot is a display-only model used by the headless screenshot.
func Snapshot(lines []pipeline.Line, source, status, llm string) *Model {
	m := &Model{
		Lines:     lines,
		Status:    status,
		LLMStatus: llm,
		Engine:    "sensevoice-small-q8",
		Ready:     true,
		Selected:  source,
	}
	if source != "" {
		m.Labels = []string{source}
		m.Sources = []audio.Source{{ID: "wav:", Name: source, Kind: "file"}}
	}
	return m
}
