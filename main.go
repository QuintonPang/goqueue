package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"goqueue/api"
	"goqueue/queue"
	"goqueue/worker"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var wg sync.WaitGroup
	wg.Add(3)
	go worker.Start(1, &wg)
	go worker.Start(2, &wg)
	go worker.Start(3, &wg)

	mux := http.NewServeMux()
	mux.HandleFunc("/jobs", api.CreateJob)
	mux.HandleFunc("/jobs/", api.GetJob)
	server := &http.Server{Addr: ":8080", Handler: mux}

	serverErrors := make(chan error, 1)
	go func() {
		serverErrors <- server.ListenAndServe()
	}()
	fmt.Println("GoQueue API listening on http://localhost:8080")

	var listenErr error
	select {
	case <-ctx.Done():
		fmt.Println("Shutdown requested")
	case err := <-serverErrors:
		if !errors.Is(err, http.ErrServerClosed) {
			listenErr = err
		}
	}
	stop() // Restore normal signal handling; another Ctrl+C can force exit.

	fmt.Println("Stopping HTTP server...")
	// Wait for handlers to finish before closing the channel they send jobs to.
	if err := server.Shutdown(context.Background()); err != nil {
		log.Fatalf("HTTP server shutdown failed: %v", err)
	}

	fmt.Println("Closing job queue...")
	close(queue.JobQueue)

	fmt.Println("Waiting for workers to finish...")
	wg.Wait()
	if listenErr != nil {
		log.Fatal(listenErr)
	}
	fmt.Println("Shutdown complete")
}
