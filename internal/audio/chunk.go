package audio

// Segment is one utterance, or a partial look at the current utterance.
type Segment struct {
	Samples []float32
	Start   int // sample index in the stream
	End     int
	Partial bool
}

// Chunker splits a 16 kHz mono stream on silence, and cuts a long utterance
// so a non-streaming model such as SenseVoice can emit subtitles without
// holding the whole session.
type Chunker struct {
	SampleRate   int
	Threshold    float32
	MinSpeech    int
	MaxSpeech    int
	Silence      int
	PartialEvery int

	pending     []float32
	buf         []float32
	bufStart    int
	abs         int
	inSpeech    bool
	silenceRun  int
	lastPartial int
	frame       int
}

// NewChunker fills zero fields with defaults for 16 kHz speech.
func NewChunker(c Chunker) *Chunker {
	if c.SampleRate <= 0 {
		c.SampleRate = SampleRate
	}
	if c.Threshold <= 0 {
		c.Threshold = 0.01
	}
	if c.MinSpeech <= 0 {
		c.MinSpeech = c.SampleRate * 3 / 10
	}
	if c.MaxSpeech <= 0 {
		c.MaxSpeech = c.SampleRate * 8
	}
	if c.Silence <= 0 {
		c.Silence = c.SampleRate * 45 / 100
	}
	c.frame = c.SampleRate / 50
	if c.frame < 1 {
		c.frame = 1
	}
	return &c
}

// Push consumes samples and returns any utterances that closed, plus an
// optional partial copy of the open utterance.
func (c *Chunker) Push(samples []float32) []Segment {
	if len(samples) == 0 {
		return nil
	}
	c.pending = append(c.pending, samples...)
	var out []Segment
	for len(c.pending) >= c.frame {
		f := append([]float32(nil), c.pending[:c.frame]...)
		c.pending = c.pending[c.frame:]
		out = append(out, c.consume(f)...)
	}
	return out
}

// Flush ends the stream and returns a trailing utterance, if any.
func (c *Chunker) Flush() []Segment {
	var out []Segment
	if len(c.pending) > 0 {
		f := c.pending
		c.pending = nil
		out = append(out, c.consume(f)...)
	}
	if c.inSpeech && len(c.buf) >= c.MinSpeech {
		out = append(out, c.finish(false))
	}
	c.inSpeech = false
	c.buf = nil
	return out
}

func (c *Chunker) consume(f []float32) []Segment {
	c.abs += len(f)
	loud := RMS(f) >= c.Threshold
	if !c.inSpeech {
		if !loud {
			return nil
		}
		c.inSpeech = true
		c.buf = append(c.buf[:0], f...)
		c.bufStart = c.abs - len(f)
		c.silenceRun = 0
		c.lastPartial = c.abs
		return nil
	}
	c.buf = append(c.buf, f...)
	if loud {
		c.silenceRun = 0
	} else {
		c.silenceRun += len(f)
	}
	if c.silenceRun >= c.Silence || len(c.buf) >= c.MaxSpeech {
		seg := c.finish(false)
		if len(seg.Samples) < c.MinSpeech {
			return nil
		}
		return []Segment{seg}
	}
	if c.PartialEvery > 0 && c.abs-c.lastPartial >= c.PartialEvery && len(c.buf) >= c.MinSpeech {
		c.lastPartial = c.abs
		return []Segment{c.peek()}
	}
	return nil
}

func (c *Chunker) peek() Segment {
	cp := append([]float32(nil), c.buf...)
	return Segment{Samples: cp, Start: c.bufStart, End: c.bufStart + len(cp), Partial: true}
}

func (c *Chunker) finish(partial bool) Segment {
	samples := c.buf
	if !partial && c.silenceRun > 0 && c.silenceRun < len(samples) {
		keep := len(samples) - c.silenceRun
		pad := c.SampleRate / 10
		end := keep + pad
		if end > len(samples) {
			end = len(samples)
		}
		samples = samples[:end]
	}
	cp := append([]float32(nil), samples...)
	seg := Segment{Samples: cp, Start: c.bufStart, End: c.bufStart + len(cp), Partial: partial}
	c.buf = nil
	c.inSpeech = false
	c.silenceRun = 0
	return seg
}
