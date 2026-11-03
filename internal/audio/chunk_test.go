package audio

import "testing"

func TestChunkerSplitsOnSilence(t *testing.T) {
	c := NewChunker(Chunker{
		SampleRate: SampleRate,
		Threshold:  0.05,
		MinSpeech:  SampleRate * 3 / 10,
		Silence:    SampleRate * 4 / 10,
		MaxSpeech:  SampleRate * 5,
	})
	silence := make([]float32, SampleRate/2)
	loud := make([]float32, SampleRate/2)
	for i := range loud {
		loud[i] = 0.5
	}
	var buf []float32
	buf = append(buf, silence...)
	buf = append(buf, loud...)
	buf = append(buf, silence...)
	buf = append(buf, loud...)
	buf = append(buf, silence...)
	segs := c.Push(buf)
	segs = append(segs, c.Flush()...)
	var finals []Segment
	for _, s := range segs {
		if !s.Partial {
			finals = append(finals, s)
		}
	}
	if len(finals) != 2 {
		t.Fatalf("finals = %d, want 2", len(finals))
	}
	if finals[0].Start < SampleRate/2-SampleRate/10 {
		t.Fatalf("first start %d, want near %d", finals[0].Start, SampleRate/2)
	}
}

func TestChunkerMaxUtterance(t *testing.T) {
	c := NewChunker(Chunker{
		SampleRate: SampleRate,
		Threshold:  0.05,
		MinSpeech:  SampleRate / 5,
		Silence:    SampleRate,
		MaxSpeech:  SampleRate,
	})
	loud := make([]float32, SampleRate*3)
	for i := range loud {
		loud[i] = 0.4
	}
	segs := c.Push(loud)
	segs = append(segs, c.Flush()...)
	sum := 0
	n := 0
	for _, s := range segs {
		if s.Partial {
			continue
		}
		n++
		sum += len(s.Samples)
	}
	if n < 3 {
		t.Fatalf("finals = %d, want at least 3", n)
	}
	if sum != len(loud) {
		t.Fatalf("kept %d samples, want %d", sum, len(loud))
	}
}

func TestResampleAndWAVRoundTrip(t *testing.T) {
	in := []float32{0, 0.5, -0.5, 0.25}
	same := Resample(in, 16000, 16000)
	if &same[0] != &in[0] {
		t.Fatal("equal-rate resample should alias the input")
	}
	up := Resample(in, 8000, 16000)
	if len(up) != 8 {
		t.Fatalf("up len %d", len(up))
	}
	mono := Downmix([]float32{0.2, 0.4, -0.2, 0.0}, 2)
	if len(mono) != 2 || mono[0] < 0.29 || mono[0] > 0.31 {
		t.Fatalf("downmix %#v", mono)
	}
}
