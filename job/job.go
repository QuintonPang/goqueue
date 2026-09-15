package job

import "time"

const (
	Pending   = "pending"
	Running   = "running"
	Completed = "completed"
	Failed    = "failed"
)

type Job struct {
	ID        int       `json:"id"`
	Type      string    `json:"type"`
	Status    string    `json:"status"`
	Payload   string    `json:"payload"`
	CreatedAt time.Time `json:"created_at"`
}
