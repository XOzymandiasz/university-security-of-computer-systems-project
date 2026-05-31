// Package httpapi zawiera API HTTP aplikacji TTP.
//
// Pakiet definiuje interfejs logiki TTP, strukturę serwera HTTP,
// konstruktor oraz routing endpointów wykorzystywanych podczas inicjalizacji,
// rejestracji i uwierzytelniania.
package httpapi

import (
	"net/http"

	"scs/internal/shared/protocol"
)

// TTPService definiuje operacje udostępniane przez usługę Trusted Third Party.
//
// Interfejs oddziela warstwę HTTP od logiki domenowej TTP. Implementacja
// odpowiada za udostępnienie publicznego klucza TTP, rejestrację stron,
// wydawanie certyfikatów oraz uwierzytelnianie klienta i serwera.
type TTPService interface {
	Init() (protocol.InitResponse, error)
	Register(req protocol.RegisterRequest) (protocol.RegisterResponse, error)
	Authenticate(req protocol.AuthenticateRequest) (protocol.AuthenticateResponse, error)
}

// Server reprezentuje serwer HTTP aplikacji TTP.
//
// Struktura przechowuje zależność do logiki TTP, która jest wywoływana
// przez handlery HTTP obsługujące poszczególne etapy protokołu.
type Server struct {
	ttp TTPService
}

// New tworzy nową instancję serwera HTTP TTP.
//
// Funkcja przyjmuje implementację usługi TTP i opakowuje ją w warstwę HTTP,
// dzięki czemu logika zaufanej strony trzeciej może być dostępna przez REST API.
//
// @param ttp Implementacja logiki Trusted Third Party.
// @return Wskaźnik do nowej instancji Server.
func New(ttp TTPService) *Server {
	return &Server{
		ttp: ttp,
	}
}

// Handler buduje główny router HTTP aplikacji TTP.
//
// Funkcja rejestruje endpointy sprawdzania stanu, inicjalizacji,
// rejestracji klienta lub serwera oraz uwierzytelniania stron.
//
// @return Handler HTTP gotowy do przekazania do serwera net/http.
func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/api/init", s.handleInit)
	mux.HandleFunc("/api/register", s.handleRegister)
	mux.HandleFunc("/api/authenticate", s.handleAuthentication)

	return mux
}
