package httpapi

import (
	"encoding/json"
	"log"
	"net/http"

	"scs/internal/shared/protocol"
)

// handleHealth obsługuje endpoint sprawdzający stan działania usługi TTP.
//
// Handler zwraca kod HTTP 200 oraz komunikat tekstowy. Endpoint może być
// używany do sprawdzenia, czy aplikacja TTP jest uruchomiona.
//
// @param w Obiekt odpowiedzi HTTP.
// @param r Obiekt żądania HTTP.
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("healthy"))
}

// handleInit obsługuje początkowe żądanie inicjalizacji protokołu.
//
// Handler zwraca publiczny klucz szyfrujący TTP. Klient i serwer używają
// tego klucza do szyfrowania danych przesyłanych do TTP podczas rejestracji
// oraz uwierzytelniania.
//
// @param w Obiekt odpowiedzi HTTP.
// @param r Obiekt żądania HTTP.
func (s *Server) handleInit(w http.ResponseWriter, r *http.Request) {
	log.Printf("handleInit called by remoteAddr=%s userAgent=%q",
		r.RemoteAddr,
		r.UserAgent(),
	)

	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	response, err := s.ttp.Init()
	if err != nil {
		log.Println("ttp init failed:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")

	if err = json.NewEncoder(w).Encode(response); err != nil {
		log.Println("encode init response:", err)
	}
}

// handleRegister obsługuje rejestrację klienta lub serwera w TTP.
//
// Handler przyjmuje dane rejestracyjne aplikacji, przekazuje je do logiki TTP,
// a następnie zwraca certyfikat X.509 wygenerowany dla zarejestrowanej strony.
//
// @param w Obiekt odpowiedzi HTTP.
// @param r Obiekt żądania HTTP.
func (s *Server) handleRegister(w http.ResponseWriter, r *http.Request) {
	log.Printf("handleRegister called by remoteAddr=%s userAgent=%q",
		r.RemoteAddr,
		r.UserAgent(),
	)
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req protocol.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response, err := s.ttp.Register(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if err = json.NewEncoder(w).Encode(response); err != nil {
		log.Println(err)
	}
}

// handleAuthentication obsługuje uwierzytelnianie klienta i serwera przez TTP.
//
// Handler odbiera żądanie zawierające dane serwera oraz zaszyfrowany pakiet
// klienta. TTP weryfikuje certyfikaty, podpisy i tożsamości stron, a następnie
// w przypadku powodzenia zwraca klucz sesyjny zaszyfrowany osobno dla klienta
// oraz serwera.
//
// @param writer Obiekt odpowiedzi HTTP.
// @param request Obiekt żądania HTTP.
func (s *Server) handleAuthentication(writer http.ResponseWriter, request *http.Request) {
	log.Printf("handleAuthentication called by remoteAddr=%s userAgent=%q",
		request.RemoteAddr,
		request.UserAgent(),
	)
	if request.Method != http.MethodPost {
		http.Error(writer, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req protocol.AuthenticateRequest
	if err := json.NewDecoder(request.Body).Decode(&req); err != nil {
		http.Error(writer, "invalid request body", http.StatusBadRequest)
		return
	}

	msg, err := s.ttp.Authenticate(req)
	if err != nil {
		http.Error(writer, err.Error(), http.StatusUnauthorized)
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)

	if err = json.NewEncoder(writer).Encode(msg); err != nil {
		http.Error(writer, "encode response", http.StatusInternalServerError)
		return
	}
}
