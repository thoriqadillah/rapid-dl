package rapid

import (
	"context"
	"fmt"
	"log"
	"sync"
)

type Pool interface {
	Start()
	Stop()
	Add(job Job)
}

type worker struct {
	poolsize int
	jobs     chan Job
	start    sync.Once
	stop     sync.Once
	quit     chan struct{}
	ctx      context.Context
}

var errPoolsize = fmt.Errorf("worker pool can't be less than 1")
var errJobsize = fmt.Errorf("job size can't be negative")

func New(ctx context.Context, poolsize int, amount ...int) (Pool, error) {
	if poolsize <= 0 {
		return nil, errPoolsize
	}

	jobsize := 0
	if len(amount) > 0 {
		jobsize = amount[0]
	}

	if jobsize < 0 {
		return nil, errJobsize
	}

	return &worker{
		poolsize: poolsize,
		jobs:     make(chan Job, jobsize),
		start:    sync.Once{},
		stop:     sync.Once{},
		quit:     make(chan struct{}),
		ctx:      ctx,
	}, nil
}

func (w *worker) Start() {
	w.start.Do(func() {
		log.Println("Starting worker...")

		for i := 0; i < w.poolsize; i++ {
			go func(id int) {
				log.Printf("Starting worker id: %d\n", id)

				for {
					select {
					case <-w.quit:
						log.Printf("Stopping worker id: %d with quit channel. Still waiting worker to finish...\n", id)
						return
					case <-w.ctx.Done():
						log.Printf("Cancelling worker id: %d process...\n", id)
						return
					case job, ok := <-w.jobs:
						if !ok {
							log.Printf("Stopping worker id: %d with closed channel\n", id)
							return
						}

						if err := job.Execute(w.ctx); err != nil {
							job.OnError(err)
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
		log.Println("Stopping worker")
		close(w.quit)
	})
}
