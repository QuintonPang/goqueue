package worker

import (
	"context"
	"testing"

	"goqueue/job"
	"goqueue/store"
)

func TestRecoverUnfinished(t *testing.T) {
	ctx := context.Background()
	s := store.NewMemoryStore()

	inputs := []job.Job{
		{Type: "pending", Status: job.Pending, MaxRetries: 3},
		{Type: "running", Status: job.Running, RetryCount: 1, MaxRetries: 3},
		{Type: "retrying", Status: job.Retrying, RetryCount: 2, MaxRetries: 3},
		{Type: "completed", Status: job.Completed, MaxRetries: 3},
		{Type: "failed", Status: job.Failed, MaxRetries: 3},
	}

	for i := range inputs {
		if err := s.Create(ctx, &inputs[i]); err != nil {
			t.Fatalf("Create returned error: %v", err)
		}
	}

	recoveryQueue := make(chan job.Job, 10)

	recovered, err := RecoverUnfinished(ctx, s, recoveryQueue)
	if err != nil {
		t.Fatalf("RecoverUnfinished returned error: %v", err)
	}
	if recovered != 3 {
		t.Fatalf("expected 3 recovered jobs, got %d", recovered)
	}
	if len(recoveryQueue) != 3 {
		t.Fatalf("expected 3 jobs in recovery queue, got %d", len(recoveryQueue))
	}

	for i := 0; i < 3; i++ {
		recoveredJob := <-recoveryQueue

		if recoveredJob.Status != job.Pending {
			t.Errorf("expected recovered job %d status %q, got %q", recoveredJob.ID, job.Pending, recoveredJob.Status)
		}

		storedJob, ok, err := s.Get(ctx, recoveredJob.ID)
		if err != nil {
			t.Fatalf("Get returned error: %v", err)
		}
		if !ok {
			t.Fatalf("expected recovered job %d to remain in store", recoveredJob.ID)
		}
		if storedJob.Status != job.Pending {
			t.Errorf("expected stored job %d status %q, got %q", storedJob.ID, job.Pending, storedJob.Status)
		}
	}

	// Recovery keeps the retry count so restarting GoQueue does not grant extra retries.
	secondRecovered := inputs[1]
	storedRunningJob, ok, err := s.Get(ctx, secondRecovered.ID)
	if err != nil || !ok {
		t.Fatalf("expected originally running job to exist after recovery")
	}
	if storedRunningJob.RetryCount != 1 {
		t.Errorf("expected retry count 1 to be preserved, got %d", storedRunningJob.RetryCount)
	}
}
