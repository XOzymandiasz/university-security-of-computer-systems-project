package httpapi

import (
	"net/http"
	"scs/internal/protocol"
)

type TTPService interface {
	Init() (protocol.InitResponse, error)
	Register(req protocol.RegisterRequest) (protocol.RegisterResponse, error)
	Authenticate(req protocol.AuthenticateRequest) (protocol.AuthenticateResponse, error)
}

type Server struct {
	ttp TTPService
}

func New(ttp TTPService) *Server {
	return &Server{
		ttp: ttp,
	}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/api/init", s.handleInit)
	mux.HandleFunc("/api/register", s.handleRegister)
	mux.HandleFunc("/api/authenticate", s.handleAuthentication)

	return mux
}
