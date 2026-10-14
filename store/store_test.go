package store

import (
	"context"
	"testing"

	"goqueue/job"
)

func TestMemoryStoreCreateAndGet(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()

	testJob := job.Job{
		Type:       "email",
		Status:     job.Pending,
		Payload:    "test",
		MaxRetries: job.DefaultMaxRetries,
	}

	if err := s.Create(ctx, &testJob); err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	if testJob.ID <= 0 {
		t.Fatalf("expected generated ID, got %d", testJob.ID)
	}

	storedJob, ok, err := s.Get(ctx, testJob.ID)
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if !ok {
		t.Fatal("expected job to exist")
	}
	if storedJob.Status != job.Pending {
		t.Errorf("expected status %q, got %q", job.Pending, storedJob.Status)
	}
}

func TestMemoryStoreUpdate(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()

	testJob := job.Job{
		Type:       "email",
		Status:     job.Pending,
		MaxRetries: job.DefaultMaxRetries,
	}

	if err := s.Create(ctx, &testJob); err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	testJob.Status = job.Running
	testJob.RetryCount = 1

	if err := s.Update(ctx, testJob); err != nil {
		t.Fatalf("Update returned error: %v", err)
	}

	storedJob, ok, err := s.Get(ctx, testJob.ID)
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if !ok {
		t.Fatal("expected job to exist")
	}
	if storedJob.Status != job.Running {
		t.Errorf("expected status %q, got %q", job.Running, storedJob.Status)
	}
	if storedJob.RetryCount != 1 {
		t.Errorf("expected retry count 1, got %d", storedJob.RetryCount)
	}
}

func TestMemoryStoreMissingJob(t *testing.T) {
	ctx := context.Background()
	s := NewMemoryStore()

	_, ok, err := s.Get(ctx, 999999)
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if ok {
		t.Error("expected missing job to return ok=false")
	}
}
