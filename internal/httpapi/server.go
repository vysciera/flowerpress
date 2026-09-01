package httpapi

import (
	"net/http"

	"flowerpress/internal/service"
)

type Server struct {
	users    *service.UserService
	sessions *service.SessionService
	mux      *http.ServeMux
}

func NewServer(users *service.UserService, sessions *service.SessionService) *Server {
	s := &Server{
		users:    users,
		sessions: sessions,
		mux:      http.NewServeMux(),
	}

	s.routes()

	return s
}

func (s *Server) Handler() http.Handler {
	return s.mux
}

func (s *Server) routes() {
	s.mux.HandleFunc("GET /health", s.handleHealth)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	w.WriteHeader(http.StatusOK)

	_, _ = w.Write([]byte(`{"status":"ok"}`))
}
