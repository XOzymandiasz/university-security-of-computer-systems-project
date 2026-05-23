package httpapi

import "net/http"

type HealthCheckUseCase interface {
	HealthCheck() string
}

func New(healthCheck HealthCheckUseCase) *Server {
	return &Server{
		healthCheck: healthCheck,
	}
}

type Server struct {
	healthCheck HealthCheckUseCase
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", s.handleHealth)

	return mux
}
