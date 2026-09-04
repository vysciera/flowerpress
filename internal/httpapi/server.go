package httpapi

import (
	"net/http"

	"flowerpress/internal/service"
)

type Server struct {
	users         *service.UserService
	sessions      *service.SessionService
	secureCookies bool
	mux           *http.ServeMux
}

func NewServer(users *service.UserService, sessions *service.SessionService, secureCookies bool) *Server {
	s := &Server{
		users:         users,
		sessions:      sessions,
		secureCookies: secureCookies,
		mux:           http.NewServeMux(),
	}

	s.routes()

	return s
}

func (s *Server) Handler() http.Handler {
	return s.mux
}

func (s *Server) routes() {
	s.mux.HandleFunc("GET /health", s.handleHealth)

	s.mux.Handle("GET /api/auth/me", // wuh
		s.requireAuth(
			http.HandlerFunc(
				s.handleMe,
			),
		),
	)

	s.mux.HandleFunc("POST /api/auth/login", s.handleLogin)
	s.mux.HandleFunc("POST /api/auth/logout", s.handleLogout)
	s.mux.HandleFunc("POST /api/auth/register", s.handleRegister)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	w.WriteHeader(http.StatusOK)

	_, _ = w.Write([]byte(`{"status":"ok"}`))
}
