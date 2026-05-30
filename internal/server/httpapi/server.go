package httpapi

import (
	"net/http"
	"scs/internal/shared/protocol"
)

type TTPClient interface {
	Authenticate(req protocol.AuthenticateRequest) (protocol.AuthenticateResponse, error)
}

func New(messagePath string, baseDir string, ttpClient TTPClient) *Server {
	return &Server{
		messagePath: messagePath,
		baseDir:     baseDir,
		ttpClient:   ttpClient,
	}
}

type Server struct {
	messagePath string
	baseDir     string
	ttpClient   TTPClient
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/api/message", s.handleMessage)
	mux.HandleFunc("/api/authenticate", s.handleAuthenticate)

	return mux
}
