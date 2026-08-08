package job

import "time"

const (
    Pending   = "pending"
    Running   = "running"
    Completed = "completed"
    Failed    = "failed"
)

type Job struct {
    ID        int
    Type      string
    Status    string
    Payload   string
    CreatedAt time.Time
}