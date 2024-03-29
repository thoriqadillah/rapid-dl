package rapid

import (
	"context"
	"fmt"
	"sync"
)

type (
	Job interface {
		Execute(ctx context.Context) error
		OnError(ctx context.Context, err error)
	}

	Pool interface {
		Start()
		Stop()
		Add(job Job)
	}

	worker struct {
		poolsize int
		jobs     chan Job
		start    sync.Once
		stop     sync.Once
		quit     chan struct{}
		ctx      context.Context
		logger   Logger
	}
)

var errPoolsize = fmt.Errorf("worker pool can't be less than 1")
var errJobsize = fmt.Errorf("job size can't be negative")

func NewWorker(ctx context.Context, poolsize int, amount int, setting Setting) (Pool, error) {
	if poolsize <= 0 {
		return nil, errPoolsize
	}

	if amount < 0 {
		return nil, errJobsize
	}

	return &worker{
		poolsize: poolsize,
		jobs:     make(chan Job, amount),
		start:    sync.Once{},
		stop:     sync.Once{},
		quit:     make(chan struct{}),
		ctx:      ctx,
		logger:   NewLogger(setting),
	}, nil
}

func (w *worker) Start() {
	w.start.Do(func() {
		w.logger.Print("Starting worker...")

		for i := 0; i < w.poolsize; i++ {
			go func(id int) {
				for {
					select {
					case <-w.quit:
						return
					case job, ok := <-w.jobs:
						if !ok {
							return
						}

						if err := job.Execute(w.ctx); err != nil {
							job.OnError(w.ctx, err)
						}
					}
				}
			}(i)
		}
	})
}

func (w *worker) Add(job Job) {
	select {
	case w.jobs <- job:
	case <-w.quit:
	}
}

func (w *worker) Stop() {
	w.stop.Do(func() {
		w.logger.Print("Stopping worker")
		close(w.quit)
	})
}
