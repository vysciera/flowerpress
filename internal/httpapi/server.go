package httpapi

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"flowerpress/internal/service"
)

type Server struct {
	users         *service.UserService
	sessions      *service.SessionService
	projects      *service.ProjectService
	secureCookies bool

	router chi.Router
}

func NewServer(
	users *service.UserService,
	sessions *service.SessionService,
	projects *service.ProjectService,
	secureCookies bool,
) *Server {
	s := &Server{
		users:         users,
		sessions:      sessions,
		projects:      projects,
		secureCookies: secureCookies,
		router:        chi.NewRouter(),
	}

	s.routes()

	return s
}

func (s *Server) Handler() http.Handler {
	return s.router
}

func (s *Server) routes() {
	r := s.router

	r.Get("/health", s.handleHealth)

	r.Route("/api", func(r chi.Router) {
		r.Post("/auth/register", s.handleRegister)
		r.Post("/auth/login", s.handleLogin)
		r.Post("/auth/logout", s.handleLogout)

		r.Get("/public/projects", s.handlePublicProjects)
		r.Get("/public/projects/{slug}", s.handlePublicProject)

		r.Group(func(r chi.Router) {
			r.Use(s.requireAuth)

			r.Get("/auth/me", s.handleMe)

			r.Get("/projects", s.handleListProjects)
			r.Post("/projects", s.handleCreateProject)

			r.Get("/projects/{id}", s.handleGetProject)
			r.Put("/projects/{id}", s.handleUpdateProject)
			r.Delete("/projects/{id}", s.handleDeleteProject)

			r.Post("/projects/{id}/publish", s.handlePublishProject)
			r.Post("/projects/{id}/unpublish", s.handleUnpublishProject)
			r.Post("/projects/{id}/unlist", s.handleUnlistProject)
			r.Post("/projects/{id}/archive", s.handleArchiveProject)
		})
	})
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	w.WriteHeader(http.StatusOK)

	_, _ = w.Write([]byte(`{"status":"ok"}`))
}
