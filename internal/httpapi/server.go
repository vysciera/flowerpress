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
		router:           chi.NewRouter(),
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
		// Auth endpoints
		// Logout unprotected - idempotent
		// absent/invalid sessions can still have coocie cleared

		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", s.handleRegister)
			r.Post("/login", s.handleLogin)
			r.Post("/logout", s.handleLogout)

			r.With(s.requireAuth).Get("/me", s.handleMe) 
		})

		// Public Projects API

		r.Route("/public/projects", func(r chi.Router) {
			r.Get("/", s.handlePublicProject)
			r.Get("/{slug}", s.handlePublicProject)
		})

		// Flowerpress Owner

		r.Group(func(r chi.Router) {
			r.Use(s.requireAuth)

			r.Route("/projects", func(r chi.Router) {
				r.Get("/", s.handleListProjects)
				r.Post("/", s.handleCreateProject)

				r.Route("/{id}", func(r chi.Router) {
					r.Get("/", s.handleGetProject)
					r.Put("/", s.handleUpdateProject)
					r.Delete("/", s.handleDeleteProject)

					r.Post("/publish", s.handlePublicProject)
					r.Post("/unpublish", s.handleUnpublishProject)
					r.Post("/unlist", s.handleUnlistProject)
					r.Post("/archive", s.handleArchiveProject)
				})
			})
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
