package api // api contains the HTTP layer of GoQueue.

import (
	"encoding/json" // encoding/json converts between JSON request/response data and Go structs.
	"net/http"      // net/http provides HTTP request, response, status-code, and method types.
	"strconv"       // strconv converts the job ID in the URL from string to int.
	"strings"       // strings helps remove "/jobs/" from the URL path.
	"sync"          // sync provides Mutex for safely generating job IDs across concurrent requests.
	"time"          // time supplies the creation timestamp for each job.

	"goqueue/job"   // job defines the Job struct and status constants.
	"goqueue/queue" // queue provides the channel used to send jobs to workers.
	"goqueue/store" // store keeps job state so GET /jobs/{id} can retrieve it.
)

var (
	nextJobID int        // nextJobID stores the latest numeric ID assigned to a job.
	idMu      sync.Mutex // idMu protects nextJobID from concurrent HTTP requests.
)

type createJobRequest struct { // createJobRequest represents only the JSON fields a client is allowed to submit.
	Type    string `json:"type"`    // Type describes the kind of work, for example "email".
	Payload string `json:"payload"` // Payload contains the data the worker should process.
}

func CreateJob(w http.ResponseWriter, r *http.Request) { // CreateJob handles POST /jobs.
	if r.Method != http.MethodPost { // Reject any HTTP method other than POST.
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed) // Return HTTP 405 to the client.
		return // Stop this handler immediately after returning the error.
	}

	var req createJobRequest // Create an empty struct that will receive the client's JSON body.
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil { // Decode JSON from the request body into req using a pointer.
		http.Error(w, "invalid JSON body", http.StatusBadRequest) // Return HTTP 400 if the JSON cannot be decoded.
		return // Stop because we cannot create a valid job from malformed JSON.
	}

	if req.Type == "" { // Validate that the required type field was provided.
		http.Error(w, "type is required", http.StatusBadRequest) // Return HTTP 400 when type is empty.
		return // Stop before assigning an ID or adding anything to the queue.
	}

	idMu.Lock()   // Obtain exclusive access to nextJobID.
	nextJobID++   // Increment the shared ID counter.
	id := nextJobID // Copy the newly generated ID into a local variable for this request.
	idMu.Unlock() // Release the mutex so another request can generate its ID.

	newJob := job.Job{ // Build the complete Job that GoQueue will store and process.
		ID:        id,          // Use the unique ID generated above.
		Type:      req.Type,    // Copy the requested job type.
		Status:    job.Pending, // New jobs begin in the pending state.
		Payload:   req.Payload, // Copy the client's payload.
		CreatedAt: time.Now(),  // Record when the job was created.
	}

	store.Add(newJob)         // Save the pending job so clients can query its status immediately.
	queue.JobQueue <- newJob // Send the job into the buffered channel for one worker to receive.

	w.Header().Set("Content-Type", "application/json") // Tell the client that the response body is JSON.
	w.WriteHeader(http.StatusAccepted)                 // Return HTTP 202 because processing happens asynchronously.
	_ = json.NewEncoder(w).Encode(newJob)              // Encode the created job as JSON; "_" intentionally ignores Encode's error.
}

func GetJob(w http.ResponseWriter, r *http.Request) { // GetJob handles requests such as GET /jobs/12.
	if r.Method != http.MethodGet { // Reject methods other than GET.
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed) // Return HTTP 405.
		return // Stop handling this request.
	}

	idText := strings.TrimPrefix(r.URL.Path, "/jobs/") // Convert "/jobs/12" into the string "12".
	id, err := strconv.Atoi(idText)                    // Convert the string ID into an int because our map uses int keys.
	if err != nil || id <= 0 {                         // Reject text such as "abc", zero, and negative IDs.
		http.Error(w, "invalid job id", http.StatusBadRequest) // Return HTTP 400 because the URL contains an invalid ID.
		return // Stop before trying to read the store.
	}

	storedJob, ok := store.Get(id) // Ask the shared store for this job; ok tells us whether the key existed.
	if !ok {                       // Enter this block when no job exists with that ID.
		http.Error(w, "job not found", http.StatusNotFound) // Return HTTP 404 for a valid but unknown ID.
		return // Stop because there is no job to encode.
	}

	w.Header().Set("Content-Type", "application/json") // Tell the client that the response is JSON.
	_ = json.NewEncoder(w).Encode(storedJob)           // Send the current stored job state as JSON.
}
