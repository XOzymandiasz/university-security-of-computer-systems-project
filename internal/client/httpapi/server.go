package httpapi

import (
	"net/http"
)

type ReadMessageUseCase interface {
	ReadMessage(msg string) (string, error)
}

type Server struct {
	readMessage ReadMessageUseCase
}

func New(readMessage ReadMessageUseCase) *Server {
	return &Server{
		readMessage: readMessage,
	}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/api/message", s.handleMessage)

	return mux
}
