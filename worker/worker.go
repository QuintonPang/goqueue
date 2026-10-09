package worker // worker contains the goroutines that process queued jobs.

import (
	"errors" // errors creates simulated failures so we can learn and test retry behavior.
	"fmt"    // fmt prints worker activity to the terminal.
	"sync"   // sync provides the WaitGroup type shared with main.
	"time"   // time provides processing delays and retry delays.

	"goqueue/job"   // job provides the Job struct and status constants.
	"goqueue/queue" // queue provides the shared JobQueue channel.
	"goqueue/store" // store lets workers persist status, retry count, and errors.
)

const retryDelay = time.Second // Wait one second before trying a failed job again.

func Start(id int, wg *sync.WaitGroup) { // Start runs one worker; id identifies it in logs and wg tracks when it stops.
	defer wg.Done() // When this worker function exits, reduce the WaitGroup counter by one.

	fmt.Println("Worker", id, "started") // Log that this worker goroutine is ready.

	for queuedJob := range queue.JobQueue { // Receive jobs until the queue is closed and completely drained.
		processWithRetries(id, queuedJob, executeJob, retryDelay) // Process this job and retry it when execution returns an error.
	}

	fmt.Println("Worker", id, "stopped") // The queue is closed and empty, so this worker is about to exit.
}

func processWithRetries(workerID int, queuedJob job.Job, execute func(job.Job) error, delay time.Duration) { // Run one job until it succeeds or exhausts its retries.
	queuedJob.Status = job.Running // The worker has picked up the job, so it is no longer merely pending.
	queuedJob.Error = ""           // Clear any old error before the first attempt.
	store.Update(queuedJob)        // Persist the running state so GET /jobs/{id} can see it.

	for { // Continue attempting until we explicitly return on success or final failure.
		fmt.Println("Worker", workerID, "processing", queuedJob.ID, "retry", queuedJob.RetryCount) // Log the current attempt.

		err := execute(queuedJob) // Execute the actual job logic; nil means success and a non-nil error means failure.
		if err == nil {           // Enter this branch when the attempt succeeds.
			queuedJob.Status = job.Completed // Mark successful work as completed.
			queuedJob.Error = ""             // Clear an earlier temporary error after eventual success.
			store.Update(queuedJob)          // Persist the final completed state.
			fmt.Println("Worker", workerID, "completed", queuedJob.ID) // Log successful completion.
			return // Stop retrying because the job is done.
		}

		queuedJob.Error = err.Error() // Save the latest error message so clients can inspect the failure.

		if queuedJob.RetryCount >= queuedJob.MaxRetries { // Check whether every allowed retry has already been used.
			queuedJob.Status = job.Failed // No attempts remain, so this is now a permanent failure.
			store.Update(queuedJob)       // Persist the failed state and final error message.
			fmt.Println("Worker", workerID, "failed", queuedJob.ID, "error:", queuedJob.Error) // Log the permanent failure.
			return // Stop because retrying again would exceed MaxRetries.
		}

		queuedJob.RetryCount++       // Record that we are about to start one retry.
		queuedJob.Status = job.Retrying // Expose a temporary retrying state to API clients.
		store.Update(queuedJob)      // Persist the retry count and latest error before waiting.
		fmt.Println("Worker", workerID, "retrying", queuedJob.ID, "attempt", queuedJob.RetryCount) // Log the retry.

		if delay > 0 { // Tests can pass zero to avoid waiting; production passes the real retry delay.
			time.Sleep(delay) // Pause briefly before another attempt instead of retrying immediately.
		}

		queuedJob.Status = job.Running // The next retry attempt is now beginning.
		store.Update(queuedJob)        // Persist running again before executing the retry.
	}
}

func executeJob(queuedJob job.Job) error { // executeJob contains the simulated task logic used by the learning project.
	time.Sleep(2 * time.Second) // Simulate a real task taking time, such as an HTTP call or email send.

	switch queuedJob.Type { // Use special learning-only job types to demonstrate different failure patterns.
	case "fail": // A "fail" job always fails, which lets us observe the final Failed state.
		return errors.New("simulated permanent failure")
	case "flaky": // A "flaky" job fails once and then succeeds on its first retry.
		if queuedJob.RetryCount == 0 {
			return errors.New("simulated temporary failure")
		}
	}

	return nil // Normal jobs, and flaky jobs after their first retry, succeed.
}
