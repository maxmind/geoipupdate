package internal

import (
	"context"
	"fmt"
	"sync"

	"golang.org/x/sync/errgroup"
)

// JobProcessor runs jobs with a set number of workers.
type JobProcessor struct {
	// sync.Mutex prevents adding new jobs while the processor is running.
	mu sync.Mutex
	// processor is used to parallelize and limit the number of workers
	// processing read requests.
	processor *errgroup.Group
	// jobs defines the jobs to be processed.
	jobs []func(context.Context) error
	// cancel cancels the context and stops the processing of the queue.
	cancel context.CancelFunc
}

// NewJobProcessor inits a new JobProcessor struct.
func NewJobProcessor(ctx context.Context, workers int) *JobProcessor {
	processor, _ := errgroup.WithContext(ctx)
	processor.SetLimit(workers)

	return &JobProcessor{
		processor: processor,
		jobs:      []func(context.Context) error{},
	}
}

// Add queues a job for processing.
func (j *JobProcessor) Add(job func(context.Context) error) {
	j.mu.Lock()
	defer j.mu.Unlock()

	j.jobs = append(j.jobs, job)
}

// Run processes the job queue and returns the first error encountered, if any.
func (j *JobProcessor) Run(ctx context.Context) error {
	j.mu.Lock()
	defer j.mu.Unlock()

	ctx, j.cancel = context.WithCancel(ctx)
	for _, job := range j.jobs {
		job := job
		j.processor.Go(func() error {
			if err := ctx.Err(); err != nil {
				return fmt.Errorf("processing canceled: %w", err)
			}
			return job(ctx)
		})
	}

	return j.Wait()
}

// Wait waits for all jobs to finish processing and returns the first
// error encountered, if any.
func (j *JobProcessor) Wait() error {
	if err := j.processor.Wait(); err != nil {
		return fmt.Errorf("running job: %w", err)
	}
	return nil
}

// Stop cancels all queued jobs.
func (j *JobProcessor) Stop() {
	if j.cancel != nil {
		j.cancel()
	}
}
