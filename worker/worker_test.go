package worker // Use the worker package so tests can call the unexported retry helper directly.

import (
	"errors"  // errors creates controlled failures for our fake executors.
	"testing" // testing is Go's built-in test framework.

	"goqueue/job"   // job provides Job and status constants.
	"goqueue/store" // store lets us inspect the final state written by the worker.
)

func TestProcessWithRetriesEventuallySucceeds(t *testing.T) { // Fail once, then succeed.
	testJob := job.Job{
		ID:         800001,
		Type:       "test",
		Status:     job.Pending,
		MaxRetries: 2,
	}
	store.Add(testJob)

	attempts := 0
	execute := func(current job.Job) error { // Fake job logic used only by this test.
		attempts++
		if attempts == 1 {
			return errors.New("temporary failure")
		}
		return nil
	}

	processWithRetries(1, testJob, execute, 0) // Zero delay keeps the test fast.

	storedJob, ok := store.Get(testJob.ID)
	if !ok {
		t.Fatal("expected job to exist")
	}

	if storedJob.Status != job.Completed {
		t.Errorf("expected status %q, got %q", job.Completed, storedJob.Status)
	}

	if storedJob.RetryCount != 1 {
		t.Errorf("expected retry count 1, got %d", storedJob.RetryCount)
	}

	if storedJob.Error != "" {
		t.Errorf("expected final error to be empty, got %q", storedJob.Error)
	}

	if attempts != 2 {
		t.Errorf("expected 2 attempts, got %d", attempts)
	}
}

func TestProcessWithRetriesEventuallyFails(t *testing.T) { // Keep failing until MaxRetries is exhausted.
	testJob := job.Job{
		ID:         800002,
		Type:       "test",
		Status:     job.Pending,
		MaxRetries: 2,
	}
	store.Add(testJob)

	attempts := 0
	execute := func(current job.Job) error {
		attempts++
		return errors.New("permanent failure")
	}

	processWithRetries(1, testJob, execute, 0)

	storedJob, ok := store.Get(testJob.ID)
	if !ok {
		t.Fatal("expected job to exist")
	}

	if storedJob.Status != job.Failed {
		t.Errorf("expected status %q, got %q", job.Failed, storedJob.Status)
	}

	if storedJob.RetryCount != 2 {
		t.Errorf("expected retry count 2, got %d", storedJob.RetryCount)
	}

	if storedJob.Error != "permanent failure" {
		t.Errorf("expected final error %q, got %q", "permanent failure", storedJob.Error)
	}

	if attempts != 3 { // Initial attempt + 2 retries.
		t.Errorf("expected 3 attempts, got %d", attempts)
	}
}
