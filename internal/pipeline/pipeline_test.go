package pipeline

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/miaoboai/hearline/internal/asr/mock"
	"github.com/miaoboai/hearline/internal/audio"
)

func TestMockClosedLoop(t *testing.T) {
	silence := make([]float32, audio.SampleRate/2)
	loud := make([]float32, audio.SampleRate)
	for i := range loud {
		loud[i] = 0.4
	}
	samples := append(append(append([]float32{}, silence...), loud...), silence...)
	lines, err := TranscribePCM(context.Background(), mock.New(), samples, audio.SampleRate, Options{
		Threshold: 0.05,
		MinSpeech: 300 * time.Millisecond,
		Silence:   300 * time.Millisecond,
		MaxChunk:  8 * time.Second,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(lines) != 1 {
		t.Fatalf("lines %#v", lines)
	}
	if !strings.HasPrefix(lines[0].Text, "mock ") {
		t.Fatalf("text %q", lines[0].Text)
	}
	if lines[0].Start < 400*time.Millisecond {
		t.Fatalf("start %s", lines[0].Start)
	}
}

func TestLiveStream(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	in := make(chan []float32, 1)
	loud := make([]float32, audio.SampleRate/2)
	for i := range loud {
		loud[i] = 0.3
	}
	var got []Line
	done := make(chan error, 1)
	go func() {
		done <- Run(ctx, mock.New(), in, Options{Threshold: 0.05, MinSpeech: 200 * time.Millisecond, Silence: 200 * time.Millisecond}, func(l Line) {
			got = append(got, l)
		})
	}()
	in <- loud
	close(in)
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Text == "" {
		t.Fatalf("got %#v", got)
	}
}
