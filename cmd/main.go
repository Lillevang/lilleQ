package main

import (
	"log"
	"net/http"

	"github.com/lillevang/lilleQ/internal/api"
	"github.com/lillevang/lilleQ/internal/queue"
)

func main() {
	queueManager := queue.NewQueueManager()
	server := api.NewServer(queueManager)

	http.HandleFunc("/queue", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			server.PublishHandler(w, r)
		case http.MethodGet:
			server.ConsumeHandler(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/queues", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			server.CreateQueueHandler(w, r)
		case http.MethodGet:
			server.ListQueuesHandler(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	log.Println("Server running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
