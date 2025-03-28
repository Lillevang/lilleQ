package api

import (
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/lillevang/lilleQ/internal/broker"
	"github.com/lillevang/lilleQ/internal/queue"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Handle WebSocket connections for queues
func HandleQueueWebSocket(queue *queue.Queue, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Failed to upgrade connection", http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	// Subscribe to queue
	subscriber := queue.AddSubscriber()
	defer close(subscriber)

	// Send messages to WebSocket client
	for msg := range subscriber {
		if err := conn.WriteJSON(map[string]string{"message": msg}); err != nil {
			return // Exit if the client disconnects or an error occurs
		}
	}
}

// Handle WebSocket connections for topics
func HandleTopicWebSocket(topic *broker.Topic, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, "Failed to upgrade connection", http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	// Subscribe to topic
	subscriber := topic.Subscribe()
	defer close(subscriber)

	// Send messages to WebSocket client
	for msg := range subscriber {
		if err := conn.WriteJSON(map[string]string{"message": msg}); err != nil {
			return // Exit if the client disconnects or an error occurs
		}
	}
}
