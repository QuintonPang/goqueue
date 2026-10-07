package api // Use the api package so the tests can call CreateJob and GetJob directly.

import (
	"bytes"         // bytes lets us build an in-memory request body from JSON text.
	"encoding/json" // encoding/json decodes the handler's JSON response into a Go struct.
	"net/http"      // net/http provides method names and status-code constants.
	"net/http/httptest" // httptest creates fake HTTP requests and response recorders for tests.
	"testing"       // testing is Go's built-in test framework.

	"goqueue/job"   // job provides the Job struct and status constants.
	"goqueue/store" // store lets us prepare and verify job data used by the handlers.
)

func TestCreateJob(t *testing.T) { // Test that POST /jobs accepts valid JSON and creates a pending job.
	body := bytes.NewBufferString(`{"type":"email","payload":"Send test email"}`) // Build the JSON request body in memory.
	req := httptest.NewRequest(http.MethodPost, "/jobs", body)                    // Create a fake POST request without opening a real network port.
	rec := httptest.NewRecorder()                                                 // Create a fake ResponseWriter that records what the handler sends back.

	CreateJob(rec, req) // Call the real HTTP handler directly.

	if rec.Code != http.StatusAccepted { // POST /jobs should return HTTP 202 Accepted.
		t.Fatalf("expected status %d, got %d", http.StatusAccepted, rec.Code) // Stop because the response is already wrong.
	}

	var createdJob job.Job // Create a Job variable that will receive the JSON response.
	if err := json.NewDecoder(rec.Body).Decode(&createdJob); err != nil { // Decode the recorded JSON response.
		t.Fatalf("failed to decode response: %v", err) // Stop if the handler did not return valid Job JSON.
	}

	if createdJob.ID <= 0 { // A successfully created job should receive a positive numeric ID.
		t.Errorf("expected a positive job ID, got %d", createdJob.ID)
	}

	if createdJob.Type != "email" { // The response should preserve the type sent by the client.
		t.Errorf("expected type %q, got %q", "email", createdJob.Type)
	}

	if createdJob.Status != job.Pending { // A newly created asynchronous job should begin as pending.
		t.Errorf("expected status %q, got %q", job.Pending, createdJob.Status)
	}

	storedJob, ok := store.Get(createdJob.ID) // Verify that CreateJob also saved the job in the shared store.
	if !ok {
		t.Fatal("expected created job to exist in store")
	}

	if storedJob.Payload != "Send test email" { // Confirm the payload was stored correctly.
		t.Errorf("expected payload %q, got %q", "Send test email", storedJob.Payload)
	}
}

func TestGetJob(t *testing.T) { // Test that GET /jobs/{id} returns an existing job as JSON.
	testJob := job.Job{ // Prepare a known job directly in the store.
		ID:      700002,
		Type:    "email",
		Status:  job.Running,
		Payload: "Get-job test",
	}
	store.Add(testJob) // Store the job before calling the GET handler.

	req := httptest.NewRequest(http.MethodGet, "/jobs/700002", nil) // Create a fake GET request for that exact job ID.
	rec := httptest.NewRecorder()                                  // Record the response produced by GetJob.

	GetJob(rec, req) // Call the real GET handler directly.

	if rec.Code != http.StatusOK { // A successful GET should return HTTP 200.
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var returnedJob job.Job // Prepare a variable for decoding the response JSON.
	if err := json.NewDecoder(rec.Body).Decode(&returnedJob); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if returnedJob.ID != testJob.ID { // The handler should return the same job requested in the URL.
		t.Errorf("expected ID %d, got %d", testJob.ID, returnedJob.ID)
	}

	if returnedJob.Status != job.Running { // The GET response should show the job's current stored status.
		t.Errorf("expected status %q, got %q", job.Running, returnedJob.Status)
	}
}

func TestGetJobWithInvalidID(t *testing.T) { // Test the negative case where the URL does not contain a numeric ID.
	req := httptest.NewRequest(http.MethodGet, "/jobs/abc", nil) // "abc" cannot be converted to an integer.
	rec := httptest.NewRecorder()

	GetJob(rec, req)

	if rec.Code != http.StatusBadRequest { // Invalid ID syntax should return HTTP 400.
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestGetMissingJob(t *testing.T) { // Test the negative case where the ID is valid but no job exists.
	req := httptest.NewRequest(http.MethodGet, "/jobs/99999999", nil) // Use a valid integer that this test suite does not add.
	rec := httptest.NewRecorder()

	GetJob(rec, req)

	if rec.Code != http.StatusNotFound { // A valid but unknown ID should return HTTP 404.
		t.Errorf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}
}
