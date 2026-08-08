package queue

import "goqueue/job"

var JobQueue = make(chan job.Job)