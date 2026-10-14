package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"goqueue/job"
	"goqueue/queue"
	"goqueue/store"
)

func TestCreateJob(t *testing.T) {
	s := store.NewMemoryStore()
	handler := NewHandler(s)

	body := bytes.NewBufferString("{\"type\":\"email\",\"payload\":\"Send test email\"}")
	req := httptest.NewRequest(http.MethodPost, "/jobs", body)
	rec := httptest.NewRecorder()

	handler.CreateJob(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected status %d, got %d", http.StatusAccepted, rec.Code)
	}

	var createdJob job.Job
	if err := json.NewDecoder(rec.Body).Decode(&createdJob); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if createdJob.ID <= 0 {
		t.Errorf("expected a positive job ID, got %d", createdJob.ID)
	}
	if createdJob.Status != job.Pending {
		t.Errorf("expected status %q, got %q", job.Pending, createdJob.Status)
	}
	if createdJob.MaxRetries != job.DefaultMaxRetries {
		t.Errorf("expected max retries %d, got %d", job.DefaultMaxRetries, createdJob.MaxRetries)
	}

	storedJob, ok, err := s.Get(context.Background(), createdJob.ID)
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if !ok {
		t.Fatal("expected created job to exist")
	}
	if storedJob.Payload != "Send test email" {
		t.Errorf("expected payload %q, got %q", "Send test email", storedJob.Payload)
	}

	// CreateJob also puts the job onto the global worker queue. Drain it so tests remain isolated.
	<-queue.JobQueue
}

func TestGetJob(t *testing.T) {
	ctx := context.Background()
	s := store.NewMemoryStore()
	handler := NewHandler(s)

	testJob := job.Job{
		Type:       "email",
		Status:     job.Running,
		Payload:    "Get-job test",
		MaxRetries: job.DefaultMaxRetries,
	}
	if err := s.Create(ctx, &testJob); err != nil {
		t.Fatalf("Create returned error: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/jobs/%d", testJob.ID), nil)
	rec := httptest.NewRecorder()

	handler.GetJob(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var returnedJob job.Job
	if err := json.NewDecoder(rec.Body).Decode(&returnedJob); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if returnedJob.ID != testJob.ID {
		t.Errorf("expected ID %d, got %d", testJob.ID, returnedJob.ID)
	}
}

func TestGetJobWithInvalidID(t *testing.T) {
	handler := NewHandler(store.NewMemoryStore())
	req := httptest.NewRequest(http.MethodGet, "/jobs/abc", nil)
	rec := httptest.NewRecorder()

	handler.GetJob(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestGetMissingJob(t *testing.T) {
	handler := NewHandler(store.NewMemoryStore())
	req := httptest.NewRequest(http.MethodGet, "/jobs/99999999", nil)
	rec := httptest.NewRecorder()

	handler.GetJob(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}
}
