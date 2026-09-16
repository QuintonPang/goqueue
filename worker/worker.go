package worker

import (
	"fmt"
	"sync"
	"time"

	"goqueue/job"
	"goqueue/queue"
	"goqueue/store"
)

func Start(id int, wg *sync.WaitGroup) {
	defer wg.Done()

	fmt.Println("Worker", id, "started")

	for queuedJob := range queue.JobQueue {
		store.UpdateStatus(queuedJob.ID, job.Running)
		fmt.Println("Worker", id, "processing", queuedJob.ID)

		time.Sleep(2 * time.Second)

		store.UpdateStatus(queuedJob.ID, job.Completed)
		fmt.Println("Worker", id, "completed", queuedJob.ID)
	}

	fmt.Println("Worker", id, "stopped")
}
