package httpapi

import (
	"net/http"

	"flowerpress/internal/service"
)

type Server struct {
	users         *service.UserService
	sessions      *service.SessionService
	projects      *service.ProjectService
	secureCookies bool
	mux           *http.ServeMux
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

	// !!User Routes
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

	//  !!Project Routes
	s.mux.Handle(
		"GET /api/projects",
		s.requireAuth(
			http.HandlerFunc(
				s.handleListProjects,
			),
		),
	)

	s.mux.Handle(
		"POST /api/projects",
		s.requireAuth(
			http.HandlerFunc(
				s.handleCreateProject,
			),
		),
	)

	s.mux.Handle(
		"GET /api/projects/{id}",
		s.requireAuth(
			http.HandlerFunc(
				s.handleGetProject,
			),
		),
	)

	s.mux.Handle(
		"PUT /api/projects/{id}",
		s.requireAuth(
			http.HandlerFunc(
				s.handleUpdateProject,
			),
		),
	)

	s.mux.HandleFunc(
		"GET /api/public/projects/{slug}",
		s.handlePublicProject,
	)

	s.mux.HandleFunc(
		"GET /api/public/projects",
		s.handlePublicProjects,
	)

	// Project Actions

	s.mux.Handle(
		"POST /api/projects/{id}/publish",
		s.requireAuth(
			http.HandlerFunc(
				s.handlePublishProject,
			),
		),
	)

	s.mux.Handle(
		"POST /api/projects/{id}/unpublish",
		s.requireAuth(
			http.HandlerFunc(
				s.handleUnpublishProject,
			),
		),
	)

	s.mux.Handle(
		"POST /api/projects/{id}/unlist",
		s.requireAuth(
			http.HandlerFunc(
				s.handleUnlistProject,
			),
		),
	)

	s.mux.Handle(
		"POST /api/projects/{id}/archive",
		s.requireAuth(
			http.HandlerFunc(
				s.handleArchiveProject,
			),
		),
	)

	s.mux.Handle(
		"DELETE /api/projects/{id}",
		s.requireAuth(
			http.HandlerFunc(
				s.handleDeleteProject,
			),
		),
	)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	w.WriteHeader(http.StatusOK)

	_, _ = w.Write([]byte(`{"status":"ok"}`))
}
