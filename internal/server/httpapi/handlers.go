package httpapi

import "net/http"

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	s.healthCheck.HealthCheck()
}
