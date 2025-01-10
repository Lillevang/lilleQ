package api

import (
	"encoding/json"
	"github.com/lillevang/lilleQ/internal/queue"
	"net/http"
)

type Server struct {
	queueManager *queue.QueueManager
}

func NewServer(q *queue.QueueManager) *Server {
	return &Server{queueManager: q}
}

func (s *Server) PublishHandler(w http.ResponseWriter, r *http.Request) {
	queueName := GetQueueName(r)
	queue, exists := s.queueManager.GetQueue(queueName)
	if !exists {
		http.Error(w, "Queue not found", http.StatusNotFound)
		return
	}

	var body struct {
		Message string `json:"message"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	queue.Publish(body.Message)
	w.WriteHeader(http.StatusCreated)
}

func (s *Server) ConsumeHandler(w http.ResponseWriter, r *http.Request) {
	queueName := GetQueueName(r)
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

func GetQueueName(r *http.Request) string {
	queueName := r.URL.Query().Get("queue")
	if queueName == "" {
		queueName = "defeault"
	}
	return queueName
}

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
	}

	json.NewEncoder(w).Encode(map[string]int{"id": id})

}

func (s *Server) ListQueuesHandler(w http.ResponseWriter, r *http.Request) {
	queues := s.queueManager.ListQueues()
	json.NewEncoder(w).Encode(queues)
}
