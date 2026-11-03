package pipeline

import (
	"context"
	"testing"

	"github.com/miaoboai/hearline/internal/audio"
)

func TestQueueKeepsOnePartialAndAllFinals(t *testing.T) {
	q := newDecodeQ()
	ctx := context.Background()
	for i := 0; i < 5; i++ {
		samples := make([]float32, partialWindow+audio.SampleRate)
		samples[0] = float32(i + 1)
		if err := q.push(ctx, audio.Segment{Partial: true, Samples: samples, Start: i}); err != nil {
			t.Fatal(err)
		}
	}
	q.mu.Lock()
	if len(q.items) != 1 || !q.items[0].Partial {
		t.Fatalf("partials = %#v", q.items)
	}
	if len(q.items[0].Samples) != partialWindow {
		t.Fatalf("trailing window %d", len(q.items[0].Samples))
	}
	if q.items[0].Samples[0] == 1 {
		t.Fatal("oldest partial was kept")
	}
	if q.items[0].Start != 4+audio.SampleRate {
		t.Fatalf("start %d", q.items[0].Start)
	}
	q.mu.Unlock()

	for i := 0; i < decodeQueueCap; i++ {
		if err := q.push(ctx, audio.Segment{Samples: []float32{float32(i)}, Start: 100 + i}); err != nil {
			t.Fatal(err)
		}
	}
	if err := q.push(ctx, audio.Segment{Partial: true, Samples: []float32{9}}); err != nil {
		t.Fatal(err)
	}
	q.mu.Lock()
	defer q.mu.Unlock()
	if len(q.items) != decodeQueueCap {
		t.Fatalf("len %d", len(q.items))
	}
	for _, item := range q.items {
		if item.Partial {
			t.Fatal("partial kept while finals filled the queue")
		}
	}
}

func TestQueueDropsPartialToAdmitFinal(t *testing.T) {
	q := newDecodeQ()
	ctx := context.Background()
	for i := 0; i < decodeQueueCap-1; i++ {
		if err := q.push(ctx, audio.Segment{Samples: []float32{1}, Start: i}); err != nil {
			t.Fatal(err)
		}
	}
	if err := q.push(ctx, audio.Segment{Partial: true, Samples: []float32{2}}); err != nil {
		t.Fatal(err)
	}
	if err := q.push(ctx, audio.Segment{Samples: []float32{3}, Start: 50}); err != nil {
		t.Fatal(err)
	}
	q.mu.Lock()
	defer q.mu.Unlock()
	var partials, finals int
	for _, item := range q.items {
		if item.Partial {
			partials++
			continue
		}
		finals++
	}
	if partials != 0 || finals != decodeQueueCap {
		t.Fatalf("partials %d finals %d items %d", partials, finals, len(q.items))
	}
}
