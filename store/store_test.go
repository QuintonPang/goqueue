package store // Use the same package so these tests exercise the real store functions directly.

import (
	"testing" // testing provides Go's built-in test framework.

	"goqueue/job" // job gives us the Job struct and status constants used by the store.
)

func TestAddAndGet(t *testing.T) { // Test that Add stores a job and Get can retrieve it.
	testJob := job.Job{ // Create a sample job used only by this test.
		ID:      1001,        // Use a unique ID so this test does not conflict with other tests.
		Type:    "email",     // Give the job a simple example type.
		Status:  job.Pending, // A newly created job should begin as pending.
		Payload: "test",      // Give the job a small sample payload.
	}

	Add(testJob) // Store the test job in our in-memory map.

	storedJob, ok := Get(testJob.ID) // Retrieve the job; ok tells us whether that ID existed.
	if !ok {                         // If ok is false, Add did not store the job correctly.
		t.Fatal("expected job to exist") // Fail this test immediately because later checks would be meaningless.
	}

	if storedJob.ID != testJob.ID { // Check that the retrieved job has the same ID we stored.
		t.Errorf("expected ID %d, got %d", testJob.ID, storedJob.ID) // Report the mismatch but allow the test to continue.
	}

	if storedJob.Status != job.Pending { // Check that the original pending status was preserved.
		t.Errorf("expected status %q, got %q", job.Pending, storedJob.Status) // Report the wrong status.
	}
}

func TestUpdateStatus(t *testing.T) { // Test that UpdateStatus changes an existing job.
	testJob := job.Job{ // Create another independent sample job.
		ID:     1002,        // Use a different ID from the first test.
		Type:   "email",     // The exact job type is not important for this test.
		Status: job.Pending, // Start the job in the pending state.
	}

	Add(testJob) // Put the job into the store before trying to update it.

	updated := UpdateStatus(testJob.ID, job.Running) // Change the stored job from pending to running.
	if !updated {                                    // false would mean the store could not find the job.
		t.Fatal("expected status update to succeed") // Stop because the update itself failed.
	}

	storedJob, ok := Get(testJob.ID) // Read the job again to verify what is actually stored.
	if !ok {                         // The job should still exist after updating it.
		t.Fatal("expected job to exist after update") // Fail immediately if it disappeared.
	}

	if storedJob.Status != job.Running { // Compare the stored status with the status we requested.
		t.Errorf("expected status %q, got %q", job.Running, storedJob.Status) // Report an incorrect update.
	}
}

func TestUpdateStatusForMissingJob(t *testing.T) { // Test how UpdateStatus behaves when the ID does not exist.
	updated := UpdateStatus(999999, job.Completed) // Try updating an ID that this test suite never adds.
	if updated {                                   // A missing job should return false, not true.
		t.Error("expected update of missing job to return false") // Fail if the store incorrectly reports success.
	}
}
