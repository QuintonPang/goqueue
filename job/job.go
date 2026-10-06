package job // job defines the data model shared by the API, store, queue, and workers.

import "time" // time provides time.Time for recording when a job was created.

const (
	Pending   = "pending"   // Pending means the job is waiting in the queue for a worker.
	Running   = "running"   // Running means a worker has received and is processing the job.
	Completed = "completed" // Completed means the worker finished the job successfully.
	Failed    = "failed"    // Failed is reserved for jobs whose processing does not succeed.
)

type Job struct { // Job represents one unit of background work in GoQueue.
	ID        int       `json:"id"`         // ID uniquely identifies the job and becomes the map key.
	Type      string    `json:"type"`       // Type describes what kind of task the worker should perform.
	Status    string    `json:"status"`     // Status tracks pending, running, completed, or failed.
	Payload   string    `json:"payload"`    // Payload stores the input data needed by the task.
	CreatedAt time.Time `json:"created_at"` // CreatedAt records when the API accepted the job.
}
