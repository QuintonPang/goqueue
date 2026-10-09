package api // Use the api package so the tests can call CreateJob and GetJob directly.

import (
	"bytes"             // bytes lets us build an in-memory request body from JSON text.
	"encoding/json"     // encoding/json decodes the handler's JSON response into a Go struct.
	"net/http"          // net/http provides method names and status-code constants.
	"net/http/httptest" // httptest creates fake HTTP requests and response recorders for tests.
	"testing"           // testing is Go's built-in test framework.

	"goqueue/job"   // job provides the Job struct and status constants.
	"goqueue/store" // store lets us prepare and verify job data used by the handlers.
)

func TestCreateJob(t *testing.T) { // Test that POST /jobs accepts valid JSON and creates a pending job.
	body := bytes.NewBufferString(`{"type":"email","payload":"Send test email"}`) // Build valid JSON in memory.
	req := httptest.NewRequest(http.MethodPost, "/jobs", body)                   // Create a fake POST request.
	rec := httptest.NewRecorder()                                                // Record the handler's response.

	CreateJob(rec, req) // Call the real handler directly.

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

	if createdJob.Type != "email" {
		t.Errorf("expected type %q, got %q", "email", createdJob.Type)
	}

	if createdJob.Status != job.Pending {
		t.Errorf("expected status %q, got %q", job.Pending, createdJob.Status)
	}

	if createdJob.MaxRetries != job.DefaultMaxRetries {
		t.Errorf("expected max retries %d, got %d", job.DefaultMaxRetries, createdJob.MaxRetries)
	}

	storedJob, ok := store.Get(createdJob.ID)
	if !ok {
		t.Fatal("expected created job to exist in store")
	}

	if storedJob.Payload != "Send test email" {
		t.Errorf("expected payload %q, got %q", "Send test email", storedJob.Payload)
	}
}

func TestGetJob(t *testing.T) {
	testJob := job.Job{
		ID:         700002,
		Type:       "email",
		Status:     job.Running,
		Payload:    "Get-job test",
		MaxRetries: job.DefaultMaxRetries,
	}
	store.Add(testJob)

	req := httptest.NewRequest(http.MethodGet, "/jobs/700002", nil)
	rec := httptest.NewRecorder()

	GetJob(rec, req)

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

	if returnedJob.Status != job.Running {
		t.Errorf("expected status %q, got %q", job.Running, returnedJob.Status)
	}
}

func TestGetJobWithInvalidID(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/jobs/abc", nil)
	rec := httptest.NewRecorder()

	GetJob(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestGetMissingJob(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/jobs/99999999", nil)
	rec := httptest.NewRecorder()

	GetJob(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}
}
