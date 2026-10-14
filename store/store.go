package store // store defines how GoQueue saves and retrieves jobs.

import (
	"context" // context lets database operations be cancelled or timed out.
	"sync"    // sync protects the in-memory test store from concurrent access.

	"goqueue/job" // job provides the Job data model.
)

// Store describes the persistence operations the API and workers need.
// PostgreSQL and the in-memory test store both implement this interface.
type Store interface {
	Create(ctx context.Context, newJob *job.Job) error
	Get(ctx context.Context, id int) (job.Job, bool, error)
	Update(ctx context.Context, updatedJob job.Job) error
	Close()
}

// MemoryStore is used by unit tests so tests do not require a real PostgreSQL server.
type MemoryStore struct {
	mu   sync.RWMutex       // Protect jobs because tests may exercise concurrent code.
	jobs map[int]job.Job     // Map job ID -> Job.
	nextID int              // Generate IDs similarly to PostgreSQL SERIAL.
}

// NewMemoryStore creates an empty in-memory implementation of Store.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		jobs: make(map[int]job.Job),
	}
}

// Create assigns an ID and stores a new job.
func (s *MemoryStore) Create(ctx context.Context, newJob *job.Job) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.nextID++
	newJob.ID = s.nextID
	s.jobs[newJob.ID] = *newJob
	return nil
}

// Get returns a copy of a stored job.
func (s *MemoryStore) Get(ctx context.Context, id int) (job.Job, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	storedJob, ok := s.jobs[id]
	return storedJob, ok, nil
}

// Update replaces the stored copy of a job.
func (s *MemoryStore) Update(ctx context.Context, updatedJob job.Job) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.jobs[updatedJob.ID] = updatedJob
	return nil
}

// Close exists so MemoryStore satisfies Store. It has nothing to clean up.
func (s *MemoryStore) Close() {}
