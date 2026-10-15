package main // main makes this package an executable program.

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
	"goqueue/store"
	"goqueue/worker"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	databaseURL := os.Getenv("DATABASE_URL") // Keep credentials/configuration outside source code.
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	jobStore, err := store.NewPostgresStore(ctx, databaseURL) // Connect to PostgreSQL once at startup.
	if err != nil {
		log.Fatalf("failed to connect to PostgreSQL: %v", err)
	}
	defer jobStore.Close()

	handler := api.NewHandler(jobStore) // Inject PostgreSQL into the HTTP layer.

	var wg sync.WaitGroup
	wg.Add(3)

	// Start workers before recovery so they can drain the queue even if many jobs are restored.
	go worker.Start(1, &wg, jobStore)
	go worker.Start(2, &wg, jobStore)
	go worker.Start(3, &wg, jobStore)

	recovered, err := worker.RecoverUnfinished(ctx, jobStore, queue.JobQueue)
	if err != nil {
		log.Fatalf("failed to recover unfinished jobs: %v", err)
	}
	fmt.Println("Recovered jobs:", recovered)

	mux := http.NewServeMux()
	mux.HandleFunc("/jobs", handler.CreateJob)
	mux.HandleFunc("/jobs/", handler.GetJob)
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

	stop()

	fmt.Println("Stopping HTTP server...")
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
