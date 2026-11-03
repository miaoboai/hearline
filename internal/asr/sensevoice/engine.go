//go:build cgo

package sensevoice

import (
	"fmt"
	"os"
	"sync"

	sherpa "github.com/k2-fsa/sherpa-onnx-go/v2/sherpa_onnx"

	"github.com/miaoboai/hearline/internal/asr"
	"github.com/miaoboai/hearline/internal/config"
)

// Engine is a CPU SenseVoice-Small Q8 recognizer.
type Engine struct {
	mu  sync.Mutex
	rec *sherpa.OfflineRecognizer
	id  string
}

// Open loads the Q8 model described by cfg. Provider other than cpu is rejected
// by config.Validate; this runner always asks sherpa for the CPU provider.
func Open(cfg config.ASR) (asr.Recognizer, error) {
	paths, err := Resolve(cfg.ModelDir)
	if err != nil {
		return nil, err
	}
	lang := cfg.Language
	if lang == "" {
		lang = "auto"
	}
	threads := cfg.Threads()
	useITN := 0
	if cfg.UseITN {
		useITN = 1
	}
	conf := sherpa.OfflineRecognizerConfig{}
	conf.FeatConfig.SampleRate = 16000
	conf.FeatConfig.FeatureDim = 80
	conf.ModelConfig.SenseVoice.Model = paths.Model
	conf.ModelConfig.SenseVoice.Language = lang
	conf.ModelConfig.SenseVoice.UseInverseTextNormalization = useITN
	conf.ModelConfig.Tokens = paths.Tokens
	conf.ModelConfig.NumThreads = threads
	conf.ModelConfig.Provider = "cpu"
	conf.ModelConfig.Debug = 0
	conf.DecodingMethod = "greedy_search"

	fmt.Fprintf(os.Stderr, "loading SenseVoice-Small Q8 (%s) cpu threads=%d itn=%d\n", paths.Model, threads, useITN)
	rec := sherpa.NewOfflineRecognizer(&conf)
	if rec == nil {
		return nil, fmt.Errorf("sherpa-onnx 未能加载 %s（确认 tokens.txt 与 CPU 版 libsherpa-onnx 可用）", paths.Model)
	}
	return &Engine{rec: rec, id: "sensevoice-small-q8"}, nil
}

// ID is the stable engine name shown in the UI.
func (e *Engine) ID() string { return e.id }

// DecodePCM runs one offline pass. samples are mono floats in about [-1, 1].
// The sherpa stream is released before this returns.
func (e *Engine) DecodePCM(samples []float32, sampleRate int) (asr.Result, error) {
	if sampleRate <= 0 {
		sampleRate = 16000
	}
	if len(samples) < sampleRate/5 {
		return asr.Result{}, nil
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.rec == nil {
		return asr.Result{}, fmt.Errorf("sensevoice recognizer is closed")
	}
	stream := sherpa.NewOfflineStream(e.rec)
	if stream == nil {
		return asr.Result{}, fmt.Errorf("sherpa-onnx 未能创建 offline stream")
	}
	defer sherpa.DeleteOfflineStream(stream)
	stream.AcceptWaveform(sampleRate, samples)
	e.rec.Decode(stream)
	got := stream.GetResult()
	if got == nil {
		return asr.Result{}, nil
	}
	return asr.Result{
		Text:       asr.CleanText(got.Text),
		Language:   asr.CleanTag(got.Lang),
		Emotion:    asr.CleanTag(got.Emotion),
		Event:      asr.CleanTag(got.Event),
		Tokens:     append([]string(nil), got.Tokens...),
		Timestamps: append([]float32(nil), got.Timestamps...),
	}, nil
}

// Close frees the sherpa recognizer.
func (e *Engine) Close() error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.rec != nil {
		sherpa.DeleteOfflineRecognizer(e.rec)
		e.rec = nil
	}
	return nil
}
