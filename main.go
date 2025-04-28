package main

import (
	"context"
	"net/http"
	"os"
	"riverqueue/worker"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
)

func main() {
	workers := river.NewWorkers()
	if err := river.AddWorkerSafely(workers, &worker.SortWorker{}); err != nil {
		panic("error adding worker: " + err.Error())
	}
	context := context.Background()

	dbPool, err := pgxpool.New(context, os.Getenv("DATABASE_URL"))
	if err != nil {
		// handle error
	}

	riverClient, err := river.NewClient(riverpgxv5.New(dbPool), &river.Config{
		Queues: map[string]river.QueueConfig{
			river.QueueDefault: {MaxWorkers: 100},
		},
		Workers: workers,
	})
	defer func() {
		// Stop fetching new work and wait for active jobs to finish.
		if err := riverClient.Stop(context); err != nil {
			// handle error
		}
	}()

	// Run the client inline. All executed jobs will inherit from ctx:
	if err := riverClient.Start(context); err != nil {
		// handle error
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		_, err = riverClient.Insert(context, worker.SortArgs{
			Strings: []string{
				"whale", "tiger", "bear",
			},
		}, nil)
		if err != nil {
			// handle error
		}
	})
	http.ListenAndServe("localhost:3000", r)
}
