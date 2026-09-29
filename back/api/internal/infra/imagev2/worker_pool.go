package imagev2

import (
	"context"
	"ismelen/inkomi/internal/domain/manga"
)

type Job struct {
	ImgData []byte
	Opts    manga.ProcessOptions
	Result  chan JobResult
	Ctx     context.Context
}

type JobResult struct {
	Data [][]byte
	Err  error
}

type WorkerPool struct {
	workersCount int
	jobs         chan Job
}

func NewWorkerPool(workersCount int) *WorkerPool {
	return &WorkerPool{
		workersCount: workersCount,
		jobs:         make(chan Job, workersCount*2), // buffered channel
	}
}

func (wp *WorkerPool) Start(processor *processorImpl) {
	for i := 0; i < wp.workersCount; i++ {
		go func() {
			for job := range wp.jobs {
				if job.Ctx.Err() != nil {
					job.Result <- JobResult{Err: job.Ctx.Err()}
					continue
				}

				res, err := processor.processInternal(job.ImgData, job.Opts)
				job.Result <- JobResult{Data: res, Err: err}
			}
		}()
	}
}

func (wp *WorkerPool) Submit(ctx context.Context, imgData []byte, opts manga.ProcessOptions) ([][]byte, error) {
	resChan := make(chan JobResult, 1)
	job := Job{
		ImgData: imgData,
		Opts:    opts,
		Result:  resChan,
		Ctx:     ctx,
	}

	select {
	case wp.jobs <- job:
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	select {
	case res := <-resChan:
		return res.Data, res.Err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}
