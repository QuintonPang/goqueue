package api

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"goqueue/job"
	"goqueue/queue"
)

var (
	nextJobID int
	idMu      sync.Mutex
)

type createJobRequest struct {
	Type    string `json:"type"`
	Payload string `json:"payload"`
}

func CreateJob(w http.ResponseWriter, r *http.Request) {
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

	idMu.Lock()
	nextJobID++
	id := nextJobID
	idMu.Unlock()

	newJob := job.Job{
		ID:        id,
		Type:      req.Type,
		Status:    job.Pending,
		Payload:   req.Payload,
		CreatedAt: time.Now(),
	}

	queue.JobQueue <- newJob

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	_ = json.NewEncoder(w).Encode(newJob)
}
