package main

import (
	"time"

	"goqueue/job"
	"goqueue/queue"
	"goqueue/worker"
)

func main() {

	go worker.Start(1)
	go worker.Start(2)
	go worker.Start(3)

	for i := 1; i <= 20; i++ {
		queue.JobQueue <- job.Job{
			ID:     i,
			Type:   "email",
			Status: job.Pending,
		}
	}
	close(queue.JobQueue)

	time.Sleep(10 * time.Second)
}
