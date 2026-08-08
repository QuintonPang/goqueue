package worker

import (
	"fmt"
	"time"

	"goqueue/queue"
)

func Start(id int) {

	fmt.Println("Worker", id, "started")

	for job := range queue.JobQueue {
		fmt.Println("Worker", id, "processing", job.ID)
		time.Sleep(2 * time.Second)
		fmt.Println("Worker", id, "completed", job.ID)
	}

	fmt.Println("Worker", id, "stopped")
}
