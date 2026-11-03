package pipeline

import (
	"context"
	"sync"

	"github.com/miaoboai/hearline/internal/audio"
)

const (
	decodeQueueCap = 4
	// partialWindow is the only audio a partial job may send to the model.
	// Finals still decode the whole utterance, once.
	partialWindow = 3 * audio.SampleRate
)

// decodeQ is a bounded queue with one consumer. Partials collapse to a
// single waiting job. Finals are never dropped; the producer waits.
type decodeQ struct {
	mu     sync.Mutex
	items  []audio.Segment
	closed bool
	wake   chan struct{}
}

func newDecodeQ() *decodeQ {
	return &decodeQ{wake: make(chan struct{}, 1)}
}

func (q *decodeQ) signal() {
	select {
	case q.wake <- struct{}{}:
	default:
	}
}

func (q *decodeQ) push(ctx context.Context, seg audio.Segment) error {
	if seg.Partial {
		seg = trailingPartial(seg, partialWindow)
	}
	for {
		q.mu.Lock()
		if q.closed {
			q.mu.Unlock()
			return context.Canceled
		}
		if seg.Partial {
			for i := range q.items {
				if q.items[i].Partial {
					q.items[i] = seg
					q.mu.Unlock()
					q.signal()
					return nil
				}
			}
			if len(q.items) >= decodeQueueCap {
				q.mu.Unlock()
				return nil
			}
			q.items = append(q.items, seg)
			q.mu.Unlock()
			q.signal()
			return nil
		}
		if len(q.items) >= decodeQueueCap && !dropOldestPartial(q) {
			q.mu.Unlock()
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-q.wake:
			}
			continue
		}
		q.items = append(q.items, seg)
		q.mu.Unlock()
		q.signal()
		return nil
	}
}

func dropOldestPartial(q *decodeQ) bool {
	for i := range q.items {
		if q.items[i].Partial {
			q.items = append(q.items[:i], q.items[i+1:]...)
			return true
		}
	}
	return false
}

func (q *decodeQ) pop() (audio.Segment, bool) {
	for {
		q.mu.Lock()
		if len(q.items) > 0 {
			seg := q.items[0]
			q.items = q.items[1:]
			q.mu.Unlock()
			q.signal()
			return seg, true
		}
		if q.closed {
			q.mu.Unlock()
			return audio.Segment{}, false
		}
		q.mu.Unlock()
		<-q.wake
	}
}

func (q *decodeQ) close() {
	q.mu.Lock()
	q.closed = true
	q.mu.Unlock()
	q.signal()
}

func trailingPartial(seg audio.Segment, n int) audio.Segment {
	if !seg.Partial || n <= 0 || len(seg.Samples) <= n {
		return seg
	}
	drop := len(seg.Samples) - n
	seg.Samples = append([]float32(nil), seg.Samples[drop:]...)
	seg.Start += drop
	return seg
}
