package job // job defines the data model shared by the API, store, queue, and workers.

import "time" // time provides time.Time for recording when a job was created.

const (
	Pending   = "pending"   // Pending means the job is waiting in the queue for a worker.
	Running   = "running"   // Running means a worker is currently executing the job.
	Retrying  = "retrying"  // Retrying means an attempt failed and another attempt will be made.
	Completed = "completed" // Completed means the worker finished the job successfully.
	Failed    = "failed"    // Failed means all allowed attempts were unsuccessful.
)

const DefaultMaxRetries = 3 // New jobs may retry up to three times after their first failed attempt.

type Job struct { // Job represents one unit of background work in GoQueue.
	ID         int       `json:"id"`                   // ID uniquely identifies the job and becomes the map key.
	Type       string    `json:"type"`                 // Type describes what kind of task the worker should perform.
	Status     string    `json:"status"`               // Status tracks pending, running, retrying, completed, or failed.
	Payload    string    `json:"payload"`              // Payload stores the input data needed by the task.
	RetryCount int       `json:"retry_count"`          // RetryCount records how many retry attempts have been started.
	MaxRetries int       `json:"max_retries"`          // MaxRetries limits how many retries are allowed after the first attempt.
	Error      string    `json:"error,omitempty"`      // Error stores the most recent failure message; omitted from JSON when empty.
	CreatedAt  time.Time `json:"created_at"`           // CreatedAt records when the API accepted the job.
}
