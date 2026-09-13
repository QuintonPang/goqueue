package worker

import (
	"fmt"
	"sync"
	"time"

	"goqueue/queue"
)

func Start(id int, wg *sync.WaitGroup) {
	defer wg.Done()

	fmt.Println("Worker", id, "started")

	for job := range queue.JobQueue {
		fmt.Println("Worker", id, "processing", job.ID)
		time.Sleep(2 * time.Second)
		fmt.Println("Worker", id, "completed", job.ID)
	}

	fmt.Println("Worker", id, "stopped")
}
