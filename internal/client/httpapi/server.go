// Package httpapi zawiera lokalne API HTTP klienta.
//
// Pakiet udostępnia endpointy używane przez interfejs użytkownika do uruchamiania
// uwierzytelniania oraz wysyłania wiadomości do serwera po zestawieniu sesji.
package httpapi

import "net/http"

// ReadMessageUseCase definiuje operację wysyłania wiadomości do serwera.
//
// Interfejs oddziela warstwę HTTP od logiki aplikacyjnej odpowiedzialnej
// za szyfrowanie wiadomości, komunikację z serwerem i odszyfrowanie odpowiedzi.
type ReadMessageUseCase interface {
	ReadMessage(msg string) (string, error)
}

// AuthenticateUseCase definiuje operację uwierzytelniania klienta.
//
// Implementacja tego interfejsu wykonuje proces uwierzytelnienia z udziałem
// klienta, serwera oraz TTP, a po sukcesie zapisuje lokalny klucz sesyjny.
type AuthenticateUseCase interface {
	Authenticate() error
}

// Server reprezentuje lokalny serwer HTTP klienta.
//
// Struktura przechowuje zależności do przypadków użycia oraz katalog bazowy,
// w którym zapisywany jest lokalny stan klienta, między innymi klucz sesyjny.
type Server struct {
	readMessage  ReadMessageUseCase
	authenticate AuthenticateUseCase
	baseDir      string
}

// New tworzy nową instancję lokalnego serwera HTTP klienta.
//
// Funkcja przyjmuje implementacje przypadków użycia oraz katalog bazowy
// lokalnej tożsamości, a następnie zwraca gotową strukturę Server.
//
// @param readMessage Przypadek użycia odpowiedzialny za wysyłanie wiadomości.
// @param authenticate Przypadek użycia odpowiedzialny za uwierzytelnianie klienta.
// @param baseDir Katalog bazowy lokalnej tożsamości klienta.
// @return Wskaźnik do nowej instancji Server.
func New(
	readMessage ReadMessageUseCase,
	authenticate AuthenticateUseCase,
	baseDir string,
) *Server {
	return &Server{
		readMessage:  readMessage,
		authenticate: authenticate,
		baseDir:      baseDir,
	}
}

// Handler buduje główny router HTTP lokalnego API klienta.
//
// Funkcja rejestruje endpoint sprawdzający stan aplikacji, endpoint wysyłania
// wiadomości oraz endpoint uruchamiający proces uwierzytelniania.
//
// @return Handler HTTP gotowy do przekazania do serwera net/http.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/api/message", s.handleMessage)
	mux.HandleFunc("/api/authenticate", s.handleAuthenticate)

	return mux
}
