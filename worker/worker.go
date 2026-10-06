package worker // worker contains the goroutines that process queued jobs.

import (
	"fmt"  // fmt prints worker activity to the terminal.
	"sync" // sync provides the WaitGroup type shared with main.
	"time" // time is currently used to simulate two seconds of job processing.

	"goqueue/job"   // job provides status constants such as Running and Completed.
	"goqueue/queue" // queue provides the shared JobQueue channel.
	"goqueue/store" // store lets workers update the job status visible to the API.
)

func Start(id int, wg *sync.WaitGroup) { // Start runs one worker; id identifies it in logs and wg tracks when it stops.
	defer wg.Done() // When this worker function exits, reduce the WaitGroup counter by one.

	fmt.Println("Worker", id, "started") // Log that this worker goroutine is ready.

	for queuedJob := range queue.JobQueue { // Continuously receive jobs; the loop ends after JobQueue is closed and drained.
		store.UpdateStatus(queuedJob.ID, job.Running)           // Mark the job as running as soon as this worker takes it.
		fmt.Println("Worker", id, "processing", queuedJob.ID)   // Log which worker received which job.

		time.Sleep(2 * time.Second) // Simulate doing the actual job for two seconds.

		store.UpdateStatus(queuedJob.ID, job.Completed)         // Mark the job completed after the simulated work finishes.
		fmt.Println("Worker", id, "completed", queuedJob.ID)    // Log successful completion.
	}

	fmt.Println("Worker", id, "stopped") // The queue is closed and empty, so this worker is about to exit.
}
