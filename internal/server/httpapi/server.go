package httpapi

import "net/http"

func New(messagePath string) *Server {
	return &Server{
		messagePath: messagePath,
	}
}

type Server struct {
	messagePath string
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/api/message", s.handleMessage)

	return mux
}
