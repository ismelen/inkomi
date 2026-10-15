package vips

import "context"

// WorkerPool bounds how many pages are processed at once across every running job and preview,
// since each one holds several decoded copies of a page in memory.
type WorkerPool struct {
	tasks chan func()
}

func NewWorkerPool(ctx context.Context, workers int) *WorkerPool {
	p := &WorkerPool{tasks: make(chan func(), workers*2)}
	for range workers {
		go p.run(ctx)
	}
	return p
}

func (p *WorkerPool) run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case task := <-p.tasks:
			task()
		}
	}
}

// Do runs task on a worker and waits for it. When ctx is done first it returns ctx.Err() and
// task, if already queued, still runs but its result is discarded.
func (p *WorkerPool) Do(ctx context.Context, task func()) error {
	done := make(chan struct{})
	ran := false
	wrapped := func() {
		defer close(done)
		if ctx.Err() == nil {
			task()
			ran = true
		}
	}

	select {
	case p.tasks <- wrapped:
	case <-ctx.Done():
		return ctx.Err()
	}

	select {
	case <-done:
		if ran {
			return nil
		}
		return ctx.Err()
	case <-ctx.Done():
		return ctx.Err()
	}
}
