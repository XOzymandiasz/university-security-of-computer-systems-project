package httpapi

import (
	"encoding/json"
	"net/http"
)

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	s.healthCheck.HealthCheck()
}

func (s *Server) handleMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	text, err := s.readMessage.ReadMessage()
	if err != nil {
		http.Error(w, "server error: "+err.Error(), http.StatusBadGateway)
	}

	w.Header().Set("Content-Type", "application/json")

	_ = json.NewEncoder(w).Encode(MessageResponse{
		Body: text,
	})
}
