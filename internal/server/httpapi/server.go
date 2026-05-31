// Package httpapi zawiera API HTTP serwera aplikacyjnego.
//
// Pakiet definiuje zależności serwera, konstruktor oraz główny router HTTP.
// Udostępnione endpointy obsługują sprawdzanie stanu aplikacji, uwierzytelnianie
// klienta z udziałem TTP oraz szyfrowaną wymianę wiadomości.
package httpapi

import (
	"net/http"

	"scs/internal/shared/protocol"
)

// TTPClient definiuje operację uwierzytelnienia wykonywaną przez TTP.
//
// Interfejs pozwala serwerowi aplikacyjnemu przekazać żądanie uwierzytelnienia
// klienta i serwera do zaufanej strony trzeciej bez zależności od konkretnej
// implementacji klienta HTTP.
//
// @param req Żądanie uwierzytelnienia zawierające dane klienta i serwera.
// @return Odpowiedź TTP z wynikiem walidacji oraz zaszyfrowanymi kluczami sesyjnymi.
type TTPClient interface {
	Authenticate(req protocol.AuthenticateRequest) (protocol.AuthenticateResponse, error)
}

// New tworzy nową instancję serwera HTTP aplikacji serwerowej.
//
// Funkcja przyjmuje ścieżkę wiadomości, katalog bazowy lokalnej tożsamości
// serwera oraz klienta TTP używanego podczas procesu uwierzytelniania.
//
// @param messagePath Ścieżka lub identyfikator zasobu wiadomości obsługiwanego przez serwer.
// @param baseDir Katalog bazowy lokalnej tożsamości serwera.
// @param ttpClient Klient umożliwiający komunikację z usługą TTP.
// @return Wskaźnik do nowej instancji Server.
func New(messagePath string, baseDir string, ttpClient TTPClient) *Server {
	return &Server{
		messagePath: messagePath,
		baseDir:     baseDir,
		ttpClient:   ttpClient,
	}
}

// Server reprezentuje serwer HTTP aplikacji serwerowej.
//
// Struktura przechowuje konfigurację potrzebną do obsługi wiadomości,
// lokalny katalog tożsamości serwera oraz klienta TTP wykorzystywanego
// podczas uwierzytelniania.
type Server struct {
	messagePath string
	baseDir     string
	ttpClient   TTPClient
}

// Handler buduje główny router HTTP serwera aplikacyjnego.
//
// Funkcja rejestruje endpoint sprawdzający stan działania, endpoint obsługi
// zaszyfrowanej wiadomości oraz endpoint uwierzytelniania klienta.
//
// @return Handler HTTP gotowy do uruchomienia przez net/http.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/api/message", s.handleMessage)
	mux.HandleFunc("/api/authenticate", s.handleAuthenticate)

	return mux
}
