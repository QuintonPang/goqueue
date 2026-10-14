package api // api contains the HTTP layer of GoQueue.

import (
	"encoding/json" // encoding/json converts between JSON and Go structs.
	"net/http"      // net/http provides HTTP request and response types.
	"strconv"       // strconv converts the job ID in the URL from string to int.
	"strings"       // strings helps remove "/jobs/" from the URL path.
	"time"          // time supplies the creation timestamp for each job.

	"goqueue/job"   // job defines the Job struct and status constants.
	"goqueue/queue" // queue provides the channel used to send jobs to workers.
	"goqueue/store" // store defines the persistence interface used by the API.
)

// Handler owns the dependencies needed by our HTTP endpoints.
type Handler struct {
	Store store.Store // Store may be PostgreSQL in production or MemoryStore in tests.
}

// NewHandler creates an API handler that uses the supplied Store.
func NewHandler(s store.Store) *Handler {
	return &Handler{Store: s}
}

// createJobRequest contains only fields clients are allowed to send.
type createJobRequest struct {
	Type    string `json:"type"`
	Payload string `json:"payload"`
}

// CreateJob handles POST /jobs.
func (h *Handler) CreateJob(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req createJobRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	if req.Type == "" {
		http.Error(w, "type is required", http.StatusBadRequest)
		return
	}

	newJob := job.Job{
		Type:       req.Type,
		Status:     job.Pending,
		Payload:    req.Payload,
		RetryCount: 0,
		MaxRetries: job.DefaultMaxRetries,
		CreatedAt:  time.Now(),
	}

	// PostgreSQL assigns the ID and writes it back into newJob.
	if err := h.Store.Create(r.Context(), &newJob); err != nil {
		http.Error(w, "failed to save job", http.StatusInternalServerError)
		return
	}

	queue.JobQueue <- newJob // Once persisted, send the job to a worker.

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	_ = json.NewEncoder(w).Encode(newJob)
}

// GetJob handles GET /jobs/{id}.
func (h *Handler) GetJob(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	idText := strings.TrimPrefix(r.URL.Path, "/jobs/")
	id, err := strconv.Atoi(idText)
	if err != nil || id <= 0 {
		http.Error(w, "invalid job id", http.StatusBadRequest)
		return
	}

	storedJob, ok, err := h.Store.Get(r.Context(), id)
	if err != nil {
		http.Error(w, "failed to load job", http.StatusInternalServerError)
		return
	}
	if !ok {
		http.Error(w, "job not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(storedJob)
}
