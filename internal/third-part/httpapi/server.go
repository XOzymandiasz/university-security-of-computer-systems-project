package httpapi

import (
	"net/http"
	protocol2 "scs/internal/shared/protocol"
)

type TTPService interface {
	Init() (protocol2.InitResponse, error)
	Register(req protocol2.RegisterRequest) (protocol2.RegisterResponse, error)
	Authenticate(req protocol2.AuthenticateRequest) (protocol2.AuthenticateResponse, error)
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
