// Package pipeline turns PCM into subtitle lines. It is the closed loop
// shared by the window and the headless demo: audio chunks, silence
// splitting, then whatever Recognizer the caller loaded.
package pipeline

import (
	"context"
	"fmt"
	"time"

	"github.com/miaoboai/hearline/internal/asr"
	"github.com/miaoboai/hearline/internal/audio"
)

// Line is one subtitle row.
type Line struct {
	Text     string
	Language string
	Partial  bool
	Start    time.Duration
	End      time.Duration
	// Slow means this sentence took long enough that the live caption lags.
	Slow bool
}

// String is a stable text form for logs and --out files.
func (l Line) String() string {
	lang := l.Language
	if lang == "" {
		lang = "--"
	}
	prefix := ""
	if l.Partial {
		prefix = "~ "
	}
	return fmt.Sprintf("%s%s-%s\t%s\t%s", prefix, fmtDur(l.Start), fmtDur(l.End), lang, l.Text)
}

func fmtDur(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	cs := d.Milliseconds()
	return fmt.Sprintf("%02d:%02d.%02d", cs/60000, (cs/1000)%60, (cs%1000)/10)
}

// Options controls utterance splitting. Zero values use the chunker defaults
// (about 8 s max, 0.45 s silence). PartialEvery enables in-progress lines.
type Options struct {
	MaxChunk     time.Duration
	Silence      time.Duration
	MinSpeech    time.Duration
	PartialEvery time.Duration
	Threshold    float32
	// Sensitivity is high, mid, or low. It applies when Threshold is 0
	// and the chunker tracks the quiet level itself.
	Sensitivity string
}

func (o Options) chunker() *audio.Chunker {
	c := audio.Chunker{SampleRate: audio.SampleRate, Threshold: o.Threshold, Sensitivity: o.Sensitivity}
	if o.MaxChunk > 0 {
		c.MaxSpeech = int(o.MaxChunk * audio.SampleRate / time.Second)
	}
	if o.Silence > 0 {
		c.Silence = int(o.Silence * audio.SampleRate / time.Second)
	}
	if o.MinSpeech > 0 {
		c.MinSpeech = int(o.MinSpeech * audio.SampleRate / time.Second)
	}
	if o.PartialEvery > 0 {
		c.PartialEvery = int(o.PartialEvery * audio.SampleRate / time.Second)
	}
	return audio.NewChunker(c)
}

// TranscribePCM decodes a whole mono buffer. Audio that never crosses the
// energy threshold and is at most 20 s is decoded as a single utterance, so
// a quiet fixture still reaches the model.
func TranscribePCM(ctx context.Context, rec asr.Recognizer, samples []float32, sampleRate int, opt Options) ([]Line, error) {
	if rec == nil {
		return nil, fmt.Errorf("nil recognizer")
	}
	if sampleRate <= 0 {
		return nil, fmt.Errorf("invalid sample rate")
	}
	if sampleRate != audio.SampleRate {
		samples = audio.Resample(samples, sampleRate, audio.SampleRate)
	}
	ch := opt.chunker()
	segs := ch.Push(samples)
	segs = append(segs, ch.Flush()...)
	if len(segs) == 0 && len(samples) > 0 && len(samples) <= audio.SampleRate*20 {
		segs = []audio.Segment{{Samples: samples, Start: 0, End: len(samples)}}
	}
	var lines []Line
	for _, seg := range segs {
		if err := ctx.Err(); err != nil {
			return lines, err
		}
		line, err := decode(rec, seg)
		if err != nil {
			return lines, err
		}
		if line.Text == "" {
			continue
		}
		lines = append(lines, line)
	}
	return lines, nil
}

// Run reads a live stream until ctx is cancelled or in closes.
// Decode errors are returned. Partial lines are emitted when configured.
func Run(ctx context.Context, rec asr.Recognizer, in <-chan []float32, opt Options, emit func(Line)) error {
	if rec == nil {
		return fmt.Errorf("nil recognizer")
	}
	ch := opt.chunker()
	emitSegs := func(segs []audio.Segment) error {
		for _, seg := range segs {
			if err := ctx.Err(); err != nil {
				return err
			}
			line, err := decode(rec, seg)
			if err != nil {
				return err
			}
			if line.Text == "" || emit == nil {
				continue
			}
			emit(line)
		}
		return nil
	}
	for {
		select {
		case <-ctx.Done():
			_ = emitSegs(ch.Flush())
			return ctx.Err()
		case buf, ok := <-in:
			if !ok {
				return emitSegs(ch.Flush())
			}
			if err := emitSegs(ch.Push(buf)); err != nil {
				return err
			}
		}
	}
}

func decode(rec asr.Recognizer, seg audio.Segment) (Line, error) {
	start := time.Now()
	res, err := rec.DecodePCM(seg.Samples, audio.SampleRate)
	if err != nil {
		return Line{}, err
	}
	elapsed := time.Since(start)
	audioDur := time.Duration(len(seg.Samples)) * time.Second / audio.SampleRate
	slow := !seg.Partial && audioDur > 0 && elapsed > 2*time.Second && elapsed > audioDur*3/2
	return Line{
		Text:     res.Text,
		Language: res.Language,
		Partial:  seg.Partial,
		Start:    sampleTime(seg.Start),
		End:      sampleTime(seg.End),
		Slow:     slow,
	}, nil
}

func sampleTime(n int) time.Duration {
	if n < 0 {
		n = 0
	}
	return time.Duration(n) * time.Second / audio.SampleRate
}
