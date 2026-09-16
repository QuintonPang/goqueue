package store

import (
	"sync"

	"goqueue/job"
)

var (
	jobs = make(map[int]job.Job)
	mu   sync.RWMutex
)

func Add(newJob job.Job) {
	mu.Lock()
	defer mu.Unlock()

	jobs[newJob.ID] = newJob
}

func Get(id int) (job.Job, bool) {
	mu.RLock()
	defer mu.RUnlock()

	storedJob, ok := jobs[id]
	return storedJob, ok
}

func UpdateStatus(id int, status string) bool {
	mu.Lock()
	defer mu.Unlock()

	storedJob, ok := jobs[id]
	if !ok {
		return false
	}

	storedJob.Status = status
	jobs[id] = storedJob
	return true
}
