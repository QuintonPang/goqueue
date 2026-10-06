package main // main makes this package an executable program.

import (
	"context"   // context carries cancellation signals and is used for graceful shutdown.
	"errors"    // errors lets us compare returned errors safely.
	"fmt"       // fmt prints readable messages to the terminal.
	"log"       // log prints fatal errors and can terminate the program.
	"net/http"  // net/http provides Go's built-in HTTP server.
	"os"        // os gives us operating-system values such as interrupt signals.
	"os/signal" // os/signal lets the program listen for Ctrl+C and shutdown signals.
	"sync"      // sync provides WaitGroup for waiting for all workers to stop.
	"syscall"   // syscall provides SIGTERM, commonly sent when a process is being stopped.

	"goqueue/api"    // api contains our HTTP handlers such as POST /jobs and GET /jobs/{id}.
	"goqueue/queue"  // queue contains the shared job channel.
	"goqueue/worker" // worker contains the background worker logic.
)

func main() { // main is the first function Go runs when the program starts.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM) // Create a context cancelled by Ctrl+C or SIGTERM.
	defer stop() // Ensure signal resources are released when main eventually exits.

	var wg sync.WaitGroup // Create a WaitGroup to track running workers.
	wg.Add(3)             // Tell the WaitGroup that three workers must finish.

	go worker.Start(1, &wg) // Start Worker 1 in its own goroutine.
	go worker.Start(2, &wg) // Start Worker 2 in its own goroutine.
	go worker.Start(3, &wg) // Start Worker 3 in its own goroutine.

	mux := http.NewServeMux()                    // Create an HTTP router (ServeMux).
	mux.HandleFunc("/jobs", api.CreateJob)       // Route exactly /jobs to the job-creation handler.
	mux.HandleFunc("/jobs/", api.GetJob)         // Route paths such as /jobs/1 to the job-reading handler.
	server := &http.Server{Addr: ":8080", Handler: mux} // Configure the HTTP server to listen on port 8080.

	serverErrors := make(chan error, 1) // Create a buffered channel that can receive one server error.
	go func() {                         // Start the HTTP server in a separate goroutine so main can also wait for shutdown signals.
		serverErrors <- server.ListenAndServe() // Run the server and send its final error into serverErrors when it stops.
	}() // Finish and immediately start the anonymous goroutine.

	fmt.Println("GoQueue API listening on http://localhost:8080") // Tell the developer that the API is ready.

	var listenErr error // Keep any unexpected HTTP-server error so we can report it after cleanup.

	select { // Wait until either a shutdown signal arrives or the HTTP server stops unexpectedly.
	case <-ctx.Done(): // This case runs when Ctrl+C or SIGTERM cancels ctx.
		fmt.Println("Shutdown requested") // Inform the developer that graceful shutdown is starting.
	case err := <-serverErrors: // This case runs if ListenAndServe returns first.
		if !errors.Is(err, http.ErrServerClosed) { // Ignore the normal error produced when an HTTP server is intentionally closed.
			listenErr = err // Save any unexpected server error for later reporting.
		} // Finish checking the server error.
	} // Finish waiting for the first shutdown event.

	stop() // Stop signal notification and restore normal signal behavior; another Ctrl+C can force the process to exit.

	fmt.Println("Stopping HTTP server...") // Explain the first shutdown step to the terminal.
	// Shut down HTTP first so request handlers finish before we close the queue they send jobs into.
	if err := server.Shutdown(context.Background()); err != nil { // Gracefully stop accepting requests and wait for active handlers.
		log.Fatalf("HTTP server shutdown failed: %v", err) // Exit with an error if the HTTP server cannot shut down cleanly.
	} // HTTP handlers have now finished.

	fmt.Println("Closing job queue...") // Explain the second shutdown step.
	close(queue.JobQueue)               // Tell workers that no more jobs will ever be sent to the channel.

	fmt.Println("Waiting for workers to finish...") // Explain the final waiting step.
	wg.Wait() // Block here until all three workers call wg.Done() and the counter reaches zero.

	if listenErr != nil { // Check whether the server originally stopped because of an unexpected error.
		log.Fatal(listenErr) // Report that error and terminate with a failure status.
	} // No unexpected server error occurred if execution continues.

	fmt.Println("Shutdown complete") // All HTTP work and queued worker jobs finished successfully.
}
