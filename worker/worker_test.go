package worker

import (
	"context"
	"errors"
	"testing"

	"goqueue/job"
	"goqueue/store"
)

func TestProcessWithRetriesEventuallySucceeds(t *testing.T) {
	ctx := context.Background()
	s := store.NewMemoryStore()

	testJob := job.Job{
		Type:       "test",
		Status:     job.Pending,
		MaxRetries: 2,
	}
	if err := s.Create(ctx, &testJob); err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	attempts := 0
	execute := func(current job.Job) error {
		attempts++
		if attempts == 1 {
			return errors.New("temporary failure")
		}
		return nil
	}

	processWithRetries(1, testJob, s, execute, 0)

	storedJob, ok, err := s.Get(ctx, testJob.ID)
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if !ok {
		t.Fatal("expected job to exist")
	}
	if storedJob.Status != job.Completed {
		t.Errorf("expected status %q, got %q", job.Completed, storedJob.Status)
	}
	if storedJob.RetryCount != 1 {
		t.Errorf("expected retry count 1, got %d", storedJob.RetryCount)
	}
	if attempts != 2 {
		t.Errorf("expected 2 attempts, got %d", attempts)
	}
}

func TestProcessWithRetriesEventuallyFails(t *testing.T) {
	ctx := context.Background()
	s := store.NewMemoryStore()

	testJob := job.Job{
		Type:       "test",
		Status:     job.Pending,
		MaxRetries: 2,
	}
	if err := s.Create(ctx, &testJob); err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	attempts := 0
	execute := func(current job.Job) error {
		attempts++
		return errors.New("permanent failure")
	}

	processWithRetries(1, testJob, s, execute, 0)

	storedJob, ok, err := s.Get(ctx, testJob.ID)
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
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
	if attempts != 3 {
		t.Errorf("expected 3 attempts, got %d", attempts)
	}
}
