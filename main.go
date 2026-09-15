package main

import (
	"sync"

	"goqueue/job"
	"goqueue/queue"
	"goqueue/worker"
)

func main() {
	var wg sync.WaitGroup

	wg.Add(3)

	go worker.Start(1, &wg)
	go worker.Start(2, &wg)
	go worker.Start(3, &wg)

	for i := 1; i <= 20; i++ {
		queue.JobQueue <- job.Job{
			ID:     i,
			Type:   "email",
			Status: job.Pending,
		}
	}

	close(queue.JobQueue)
	wg.Wait()
}
