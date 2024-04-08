package internal

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// TestJobQueueRun tests the parallel job queue functionality
// and ensures that the maximum number of allowed goroutines does not exceed the value
// set in the config.
func TestJobQueueRun(t *testing.T) {
	simulatedJobDuration := 5 * time.Millisecond
	jobsNumber := 10

	tests := []struct {
		Description string
		Parallelism int
	}{{
		Description: "sequential jobs",
		Parallelism: 1,
	}, {
		Description: "parallel jobs",
		Parallelism: 3,
	}}

	for _, test := range tests {
		t.Run(test.Description, func(t *testing.T) {
			doneCh := make(chan struct{})
			var lock sync.Mutex
			runningGoroutines := 0
			maxConcurrentGoroutines := 0

			// A mock processor function that is used to gather data
			// about the number of goroutines called.
			processorFunc := func(_ context.Context) error {
				lock.Lock()
				runningGoroutines++
				if runningGoroutines > maxConcurrentGoroutines {
					maxConcurrentGoroutines = runningGoroutines
				}
				lock.Unlock()

				time.Sleep(simulatedJobDuration)

				lock.Lock()
				runningGoroutines--
				lock.Unlock()
				return nil
			}

			ctx := context.Background()
			jobProcessor := NewJobProcessor(ctx, test.Parallelism)
			for i := 0; i < jobsNumber; i++ {
				jobProcessor.Add(processorFunc)
			}

			// Execute run in a goroutine so that we can exit early if the test
			// hangs or takes too long to execute.
			go func() {
				if err := jobProcessor.Run(ctx); err != nil {
					t.Error(err)
				}
				close(doneCh)
			}()

			// Wait for run to complete or timeout after a certain duration
			select {
			case <-doneCh:
			case <-time.After(1000 * time.Millisecond):
				t.Errorf("Timeout waiting for function completion")
			}

			// The maximum number of parallel downloads executed should not exceed
			// the number defined in the configuration.
			require.Equal(t, maxConcurrentGoroutines, test.Parallelism)
		})
	}
}

// TestJobQueueStop cancels a job queue and makes sure queued
// jobs are not processed.
func TestJobQueueStop(t *testing.T) {
	doneCh := make(chan struct{})
	processedJobs := 0
	maxProcessedJobs := 5

	ctx := context.Background()
	jobProcessor := NewJobProcessor(ctx, 1)

	processorFunc := func(_ context.Context) error {
		processedJobs++
		if processedJobs == maxProcessedJobs {
			jobProcessor.Stop()
		}
		return nil
	}

	for i := 0; i < 10; i++ {
		jobProcessor.Add(processorFunc)
	}

	// Execute run in a goroutine so that we can exit early if the test
	// hangs or takes too long to execute.
	go func() {
		err := jobProcessor.Run(ctx)
		if err == nil {
			t.Error(`expected "processing canceled" error`)
		}
		if !strings.Contains(err.Error(), "processing canceled") {
			t.Errorf(`expected "processing canceled" error, got %q`, err)
		}
		close(doneCh)
	}()

	// Wait for run to complete or timeout after a certain duration
	select {
	case <-doneCh:
	case <-time.After(1000 * time.Millisecond):
		t.Errorf("Timeout waiting for function completion")
	}

	require.Equal(t, processedJobs, maxProcessedJobs)
}
