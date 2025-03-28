package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/lillevang/lilleQ/config"
	"github.com/lillevang/lilleQ/internal/api"
	"github.com/lillevang/lilleQ/internal/broker"
	"github.com/lillevang/lilleQ/internal/logging"
	"github.com/lillevang/lilleQ/internal/queue"
	"go.uber.org/zap"
)

func main() {
	cfg, err := config.LoadConfig("./config.yaml")
	log.Printf("Loaded config: %+v", cfg)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	logger := logging.NewLogger(cfg.Logging)
	defer logger.Sync()

	queueManager := queue.NewQueueManager()
	broker := broker.NewBroker()
	server := api.NewServer(queueManager, broker)

	// Queue routes
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

	http.HandleFunc("/queue/ws", func(w http.ResponseWriter, r *http.Request) {
		server.WebSocketQueueHandler(w, r)
	})

	// Broker routes
	http.HandleFunc("/topics", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			server.CreateTopicHandler(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/topics/publish", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			server.PublishToTopicHandler(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/topics/subscribe", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			server.SubscribeToTopicHandler(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	logger.Info("Server starting", zap.String("port", fmt.Sprintf("%d", cfg.Server.Port)))
	log.Fatal(http.ListenAndServe(":8080", nil))
}
