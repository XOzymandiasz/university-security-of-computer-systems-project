package httpapi

import (
	"encoding/json"
	"log"
	"net/http"
	"scs/internal/identity"
	"scs/internal/protocol"
)

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("healthy"))
}

func (s *Server) handleMessage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var request protocol.UIMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	defer func() {
		_ = r.Body.Close()
	}()

	if request.Body == "" {
		http.Error(w, "message body is required", http.StatusBadRequest)
		return
	}

	if _, err := identity.LoadSessionKey(s.baseDir); err != nil {
		http.Error(w, "authenticate first - missing client session key", http.StatusUnauthorized)
		return
	}

	text, err := s.readMessage.ReadMessage(request.Body)
	if err != nil {
		http.Error(w, "server error: "+err.Error(), http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	_ = json.NewEncoder(w).Encode(protocol.UIMessageResponse{
		Body: text,
	})
}

func (s *Server) handleAuthenticate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := s.authenticate.Authenticate(); err != nil {
		log.Println("client authenticate failed:", err)
		http.Error(w, "authentication failed: "+err.Error(), http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	_ = json.NewEncoder(w).Encode(map[string]any{
		"ok":      true,
		"message": "authenticated",
	})
}
