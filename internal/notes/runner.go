package notes

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/miaoboai/hearline/internal/config"
	"github.com/miaoboai/hearline/internal/pipeline"
)

// Result is one finished notes generation. Gen matches the Submit that
// produced it. A superseded request is not delivered.
type Result struct {
	Gen int
	Doc Doc
	Err error
}

// generate is replaced in tests.
var generate = Generate

type request struct {
	gen   int
	cfg   config.LLM
	lines []pipeline.Line
}

// Runner is a single worker. A new Submit cancels the one in flight and
// keeps only the latest finals.
type Runner struct {
	mu      sync.Mutex
	seq     int
	running bool
	cancel  context.CancelFunc
	pending *request
}

// Submit merges with anything already running. done runs on the worker
// goroutine, and only for the request that was not replaced.
func (r *Runner) Submit(cfg config.LLM, lines []pipeline.Line, done func(Result)) {
	if r == nil {
		return
	}
	r.mu.Lock()
	r.seq++
	job := request{gen: r.seq, cfg: cfg, lines: append([]pipeline.Line(nil), Finals(lines)...)}
	if r.running {
		r.pending = &job
		if r.cancel != nil {
			r.cancel()
		}
		r.mu.Unlock()
		return
	}
	r.running = true
	r.mu.Unlock()
	go r.loop(job, done)
}

// Stop cancels the current request and drops anything waiting.
func (r *Runner) Stop() {
	if r == nil {
		return
	}
	r.mu.Lock()
	r.pending = nil
	if r.cancel != nil {
		r.cancel()
	}
	r.mu.Unlock()
}

func (r *Runner) loop(job request, done func(Result)) {
	for {
		ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
		r.mu.Lock()
		r.cancel = cancel
		r.mu.Unlock()
		doc, err := generate(ctx, job.cfg, job.lines)
		cancel()
		r.mu.Lock()
		if r.pending != nil {
			job = *r.pending
			r.pending = nil
			r.mu.Unlock()
			continue
		}
		r.running = false
		r.cancel = nil
		r.mu.Unlock()
		if done == nil || errors.Is(err, context.Canceled) {
			return
		}
		done(Result{Gen: job.gen, Doc: doc, Err: err})
		return
	}
}
