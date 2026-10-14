package worker // worker contains the goroutines that process queued jobs.

import (
	"context" // context is used for persistence operations.
	"errors"  // errors creates simulated failures for retry demonstrations.
	"fmt"     // fmt prints worker activity to the terminal.
	"sync"    // sync provides WaitGroup.
	"time"    // time provides processing and retry delays.

	"goqueue/job"   // job provides Job and status constants.
	"goqueue/queue" // queue provides the shared JobQueue channel.
	"goqueue/store" // store defines how workers persist job state.
)

const retryDelay = time.Second

// Start runs one worker until JobQueue is closed and drained.
func Start(id int, wg *sync.WaitGroup, jobStore store.Store) {
	defer wg.Done()

	fmt.Println("Worker", id, "started")

	for queuedJob := range queue.JobQueue {
		processWithRetries(id, queuedJob, jobStore, executeJob, retryDelay)
	}

	fmt.Println("Worker", id, "stopped")
}

// processWithRetries executes one job until it succeeds or exhausts retries.
func processWithRetries(
	workerID int,
	queuedJob job.Job,
	jobStore store.Store,
	execute func(job.Job) error,
	delay time.Duration,
) {
	ctx := context.Background() // Worker persistence continues independently of the HTTP request that created the job.

	queuedJob.Status = job.Running
	queuedJob.Error = ""
	if err := jobStore.Update(ctx, queuedJob); err != nil {
		fmt.Println("Worker", workerID, "could not persist running state:", err)
		return
	}

	for {
		fmt.Println("Worker", workerID, "processing", queuedJob.ID, "retry", queuedJob.RetryCount)

		err := execute(queuedJob)
		if err == nil {
			queuedJob.Status = job.Completed
			queuedJob.Error = ""
			if err := jobStore.Update(ctx, queuedJob); err != nil {
				fmt.Println("Worker", workerID, "could not persist completed state:", err)
				return
			}
			fmt.Println("Worker", workerID, "completed", queuedJob.ID)
			return
		}

		queuedJob.Error = err.Error()

		if queuedJob.RetryCount >= queuedJob.MaxRetries {
			queuedJob.Status = job.Failed
			if err := jobStore.Update(ctx, queuedJob); err != nil {
				fmt.Println("Worker", workerID, "could not persist failed state:", err)
				return
			}
			fmt.Println("Worker", workerID, "failed", queuedJob.ID, "error:", queuedJob.Error)
			return
		}

		queuedJob.RetryCount++
		queuedJob.Status = job.Retrying
		if err := jobStore.Update(ctx, queuedJob); err != nil {
			fmt.Println("Worker", workerID, "could not persist retry state:", err)
			return
		}
		fmt.Println("Worker", workerID, "retrying", queuedJob.ID, "attempt", queuedJob.RetryCount)

		if delay > 0 {
			time.Sleep(delay)
		}

		queuedJob.Status = job.Running
		if err := jobStore.Update(ctx, queuedJob); err != nil {
			fmt.Println("Worker", workerID, "could not persist running retry state:", err)
			return
		}
	}
}

// executeJob simulates real job execution for this learning project.
func executeJob(queuedJob job.Job) error {
	time.Sleep(2 * time.Second)

	switch queuedJob.Type {
	case "fail":
		return errors.New("simulated permanent failure")
	case "flaky":
		if queuedJob.RetryCount == 0 {
			return errors.New("simulated temporary failure")
		}
	}

	return nil
}
