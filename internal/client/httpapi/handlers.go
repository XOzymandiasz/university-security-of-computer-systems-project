package httpapi

import (
	"encoding/json"
	"net/http"
	"scs/internal/shared/identity"
	"scs/internal/shared/protocol"
)

// handleHealth obsługuje endpoint sprawdzający stan działania lokalnego API.
//
// Handler zwraca kod HTTP 200 oraz prosty komunikat tekstowy. Endpoint może być
// używany do sprawdzenia, czy aplikacja klienta jest uruchomiona.
// @param w Obiekt odpowiedzi HTTP.
// @param r Obiekt żądania HTTP.
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("healthy"))
}

// handleMessage obsługuje żądanie wysłania wiadomości z interfejsu użytkownika.
//
// Handler odbiera jawną treść wiadomości z UI, sprawdza czy klient posiada
// lokalnie zapisany klucz sesyjny, a następnie przekazuje wiadomość do warstwy
// odpowiedzialnej za szyfrowaną komunikację z serwerem.
//
// @param w Obiekt odpowiedzi HTTP.
// @param r Obiekt żądania HTTP.
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

// handleAuthenticate obsługuje żądanie rozpoczęcia uwierzytelniania klienta.
//
// Handler uruchamia logikę uwierzytelnienia klienta z udziałem serwera i TTP.
// Po poprawnym zakończeniu procesu aplikacja zapisuje lokalnie klucz sesyjny,
// który jest później używany do szyfrowania wiadomości.
//
// @param w Obiekt odpowiedzi HTTP.
// @param r Obiekt żądania HTTP.
func (s *Server) handleAuthenticate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := s.authenticate.Authenticate(); err != nil {
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
