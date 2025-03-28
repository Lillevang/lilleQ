package api

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/lillevang/lilleQ/internal/broker"
	"github.com/lillevang/lilleQ/internal/queue"
)

type Server struct {
	queueManager *queue.QueueManager
	broker       *broker.Broker
}

func NewServer(q *queue.QueueManager, b *broker.Broker) *Server {
	return &Server{
		queueManager: q,
		broker:       b,
	}
}

func getQueueNameFromRequest(r *http.Request) string {
	queueName := r.URL.Query().Get("queue")
	if queueName == "" {
		queueName = "defeault"
	}
	return queueName
}

func sendJSONResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// Queue handlers

func (s *Server) CreateQueueHandler(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Name == "" {
		http.Error(w, "Invalid request", http.StatusBadRequest)
	}

	id, err := s.queueManager.CreateQueue(body.Name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}
	log.Printf("CreateQueueHandler: Created queue '%s'", body.Name)
	sendJSONResponse(w, map[string]int{"id": id})
}

func (s *Server) ListQueuesHandler(w http.ResponseWriter, r *http.Request) {
	queues := s.queueManager.ListQueues()
	log.Printf("ListQueuesHandler: Listed all queues")
	sendJSONResponse(w, queues)
}

func (s *Server) PublishHandler(w http.ResponseWriter, r *http.Request) {
	queueName := getQueueNameFromRequest(r)
	queue, exists := s.queueManager.GetQueue(queueName)
	if !exists {
		log.Printf("PublishHandler: Queue '%s' not found", queueName)
		http.Error(w, "Queue not found", http.StatusNotFound)
		return
	}

	var body struct {
		Message string `json:"message"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		log.Printf("PublishHandler: Invalid request body for queue '%s'", queueName)
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	queue.Publish(body.Message)
	log.Printf("PublishHandler: Published message to queue '%s'", queueName)
	w.WriteHeader(http.StatusCreated)
}

func (s *Server) ConsumeHandler(w http.ResponseWriter, r *http.Request) {
	queueName := getQueueNameFromRequest(r)
	queue, exists := s.queueManager.GetQueue(queueName)
	if !exists {
		http.Error(w, "Queue not found", http.StatusNotFound)
		return
	}

	msg, ok := queue.Consume()
	if !ok {
		http.Error(w, "No messages in queue", http.StatusNoContent)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"message": msg})
}

func (s *Server) WebSocketQueueHandler(w http.ResponseWriter, r *http.Request) {
	queueName := getQueueNameFromRequest(r)
	queue, exists := s.queueManager.GetQueue(queueName)
	if !exists {
		http.Error(w, "Queue not found", http.StatusNotFound)
		return
	}

	HandleQueueWebSocket(queue, w, r)
}

// Broker handlers

func (s *Server) CreateTopicHandler(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.Name == "" {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if err := s.broker.CreateTopic(body.Name); err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (s *Server) PublishToTopicHandler(w http.ResponseWriter, r *http.Request) {
	topicName := r.URL.Query().Get("topic")
	topic, exists := s.broker.GetTopic(topicName)
	if !exists {
		http.Error(w, "Topic not found", http.StatusNotFound)
		return
	}

	var body struct {
		Message string `json:"message"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	topic.Publish(body.Message)
	w.WriteHeader(http.StatusOK)
}

func (s *Server) WebSocketTopicHandler(w http.ResponseWriter, r *http.Request) {
	topicName := r.URL.Query().Get("topic")
	topic, exists := s.broker.GetTopic(topicName)
	if !exists {
		http.Error(w, "Topic not found", http.StatusNotFound)
		return
	}

	HandleTopicWebSocket(topic, w, r)
}

func (s *Server) SubscribeToTopicHandler(w http.ResponseWriter, r *http.Request) {
	topicName := r.URL.Query().Get("topic")
	topic, exists := s.broker.GetTopic(topicName)
	if !exists {
		http.Error(w, "Topic not found", http.StatusNotFound)
		return
	}

	subscriber := topic.Subscribe()
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	for {
		select {
		case msg := <-subscriber:
			if err := json.NewEncoder(w).Encode(map[string]string{"message": msg}); err != nil {
				http.Error(w, "Failed to send message", http.StatusInternalServerError)
				return
			}
			flusher.Flush()
		case <-r.Context().Done():
			// client disconnected
			return
		}
	}
}
