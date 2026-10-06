package queue // queue owns the channel that connects HTTP job creation to background workers.

import "goqueue/job" // job supplies the Job type that travels through the channel.

// JobQueue is a buffered channel shared by API handlers and workers.
var JobQueue = make(chan job.Job, 100) // The buffer can hold up to 100 waiting jobs before a sender must block.
