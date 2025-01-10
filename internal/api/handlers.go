package api

import (
	"encoding/json"
	"github.com/lillevang/lilleQ/internal/queue"
	"net/http"
)

type Server struct {
	queue *queue.Queue
}

func NewServer(q *queue.Queue) *Server {
	return &Server{queue: q}
}

func (s *Server) PublishHandler(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Message string `json:"message"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	s.queue.Publish(body.Message)
	w.WriteHeader(http.StatusCreated)
}

func (s *Server) ConsumeHandler(w http.ResponseWriter, r *http.Request) {
	msg, ok := s.queue.Consume()
	if !ok {
		http.Error(w, "No messages in queue", http.StatusNoContent)
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"message": msg})
}
