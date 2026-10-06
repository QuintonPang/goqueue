package store // store keeps the current in-memory state of every job.

import (
	"sync" // sync provides RWMutex for safe concurrent reads and writes.

	"goqueue/job" // job provides the Job struct stored in the map.
)

var (
	jobs = make(map[int]job.Job) // Create an empty map: integer job ID -> complete Job value.
	mu   sync.RWMutex            // Protect the map because HTTP handlers and workers access it concurrently.
)

func Add(newJob job.Job) { // Add inserts a newly created job into the shared map.
	mu.Lock()         // Take an exclusive write lock because the map will change.
	defer mu.Unlock() // Guarantee the write lock is released when Add returns.

	jobs[newJob.ID] = newJob // Store the Job value using its ID as the map key.
}

func Get(id int) (job.Job, bool) { // Get returns both the Job and a bool indicating whether the ID exists.
	mu.RLock()         // Take a read lock; multiple Get calls can hold read locks at the same time.
	defer mu.RUnlock() // Release this read lock when Get returns.

	storedJob, ok := jobs[id] // Copy the Job out of the map; ok is true only when that key exists.
	return storedJob, ok      // Return both the copied job and the existence flag to the caller.
}

func UpdateStatus(id int, status string) bool { // UpdateStatus changes one stored job's Status field.
	mu.Lock()         // Take an exclusive lock because we will modify the shared map.
	defer mu.Unlock() // Always release the lock when this function returns.

	storedJob, ok := jobs[id] // Read a copy of the Job and check whether the requested ID exists.
	if !ok {                  // Run this branch if no such job exists.
		return false // Tell the caller that nothing was updated.
	}

	storedJob.Status = status // Change the status on our local copy of the Job.
	jobs[id] = storedJob      // Write the changed copy back into the map.
	return true               // Tell the caller the update succeeded.
}
