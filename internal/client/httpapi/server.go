package httpapi

import "net/http"

type ReadMessageUseCase interface {
	ReadMessage() (string, error)
}

type HealthCheckUseCase interface {
	HealthCheck() string
}

type Server struct {
	readMessage ReadMessageUseCase
	healthCheck HealthCheckUseCase
}

func New(readMessage ReadMessageUseCase, healthCheck HealthCheckUseCase) *Server {
	return &Server{
		readMessage: readMessage,
		healthCheck: healthCheck,
	}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/api/message", s.handleMessage)

	return mux
}
