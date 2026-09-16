package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"goqueue/api"
	"goqueue/worker"
)

func main() {
	var wg sync.WaitGroup

	wg.Add(3)
	go worker.Start(1, &wg)
	go worker.Start(2, &wg)
	go worker.Start(3, &wg)

	http.HandleFunc("/jobs", api.CreateJob)
	http.HandleFunc("/jobs/", api.GetJob)

	fmt.Println("GoQueue API listening on http://localhost:8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
