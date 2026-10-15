package worker // Recovery helpers for unfinished jobs after a process restart.

import (
	"context"
	"fmt"

	"goqueue/job"
	"goqueue/store"
)

// RecoverUnfinished reloads jobs that were not completed before the previous process stopped.
// A recovered job is reset to pending, persisted, then sent back to the worker queue.
func RecoverUnfinished(
	ctx context.Context,
	jobStore store.Store,
	jobQueue chan<- job.Job,
) (int, error) {
	unfinished, err := jobStore.ListUnfinished(ctx)
	if err != nil {
		return 0, err
	}

	recovered := 0
	for _, recoveredJob := range unfinished {
		recoveredJob.Status = job.Pending // A restarted process has no job that is truly still running.

		if err := jobStore.Update(ctx, recoveredJob); err != nil {
			return recovered, fmt.Errorf("reset job %d before recovery: %w", recoveredJob.ID, err)
		}

		select {
		case jobQueue <- recoveredJob: // Requeue the durable job so a worker can process it again.
			recovered++
		case <-ctx.Done(): // Stop cleanly if startup itself is cancelled.
			return recovered, ctx.Err()
		}
	}

	return recovered, nil
}
