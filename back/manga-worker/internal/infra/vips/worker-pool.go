package vips

import (
	"context"

	"github.com/ismelen/inkomi/back/manga-worker/internal/domain/models"
)

type pageFunc func(data []byte, cfg models.MangaConfig, rtl bool) ([]models.ProcessedPage, error)

type poolJob struct {
	ctx    context.Context
	data   []byte
	cfg    models.MangaConfig
	rtl    bool
	result chan<- poolResult
}

type poolResult struct {
	pages []models.ProcessedPage
	err   error
}

// WorkerPool bounds how many pages are processed at once across every running job, since each
// one holds several decoded copies of a page in memory.
type WorkerPool struct {
	jobs chan poolJob
}

func NewWorkerPool(ctx context.Context, workers int, process pageFunc) *WorkerPool {
	p := &WorkerPool{jobs: make(chan poolJob, workers*2)}
	for range workers {
		go p.run(ctx, process)
	}
	return p
}

func (p *WorkerPool) run(ctx context.Context, process pageFunc) {
	for {
		select {
		case <-ctx.Done():
			return
		case job := <-p.jobs:
			if err := job.ctx.Err(); err != nil {
				job.result <- poolResult{err: err}
				continue
			}
			pages, err := process(job.data, job.cfg, job.rtl)
			job.result <- poolResult{pages: pages, err: err}
		}
	}
}

func (p *WorkerPool) Submit(ctx context.Context, data []byte, cfg models.MangaConfig, rtl bool) ([]models.ProcessedPage, error) {
	result := make(chan poolResult, 1)
	select {
	case p.jobs <- poolJob{ctx: ctx, data: data, cfg: cfg, rtl: rtl, result: result}:
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	select {
	case res := <-result:
		return res.pages, res.err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
