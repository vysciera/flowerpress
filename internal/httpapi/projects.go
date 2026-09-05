package httpapi

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"time"

	"flowerpress/internal/domain"
	"flowerpress/internal/service"
)

type createProjectRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type updateProjectRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

type projectResponse struct {
	ID          int64                `json:"id"`
	Title       string               `json:"title"`
	Slug        string               `json:"slug"`
	Description string               `json:"description"`
	Status      domain.ProjectStatus `json:"status"`

	PublishedAt *time.Time `json:"published_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func projectToResponse(project *domain.Project) projectResponse {
	return projectResponse{
		ID:          project.ID,
		Title:       project.Title,
		Slug:        project.Slug,
		Description: project.Description,
		Status:      project.Status,
		PublishedAt: project.PublishedAt,
		CreatedAt:   project.CreatedAt,
		UpdatedAt:   project.UpdatedAt,
	}
}

func projectIDFromRequest(r *http.Request) (int64, error) {
	return strconv.ParseInt(r.PathValue("id"), 10, 64)
}

// !!Server methods
func (s *Server) handleCreateProject(w http.ResponseWriter, r *http.Request) {
	var request createProjectRequest

	if !decodeJSON(w, r, &request) {
		return
	}

	project, err := s.projects.Create(
		r.Context(),
		request.Title,
		request.Description,
	)

	switch {
	case errors.Is(err, service.ErrProjectTitleRequired):
		writeJSON(
			w,
			http.StatusBadRequest,
			map[string]string{
				"error": "project title is required",
			},
		)

		return

	case err != nil:
		writeJSON(
			w,
			http.StatusInternalServerError,
			map[string]string{
				"error": "internal server error",
			},
		)

		return
	}

	writeJSON(
		w,
		http.StatusCreated,
		projectToResponse(project),
	)
}

func (s *Server) handleListProjects(w http.ResponseWriter, r *http.Request) {
	projects, err := s.projects.List(
		r.Context(),
	)

	if err != nil {
		writeJSON(
			w,
			http.StatusInternalServerError,
			map[string]string{
				"error": "internal server error",
			},
		)

		return
	}

	response := make([]projectResponse, 0, len(projects))
	for _, project := range projects {
		response = append(response, projectToResponse(project))
	}

	writeJSON(w, http.StatusOK, response)
}

func (s *Server) handleGetProject(w http.ResponseWriter, r *http.Request) {
	projectID, err := projectIDFromRequest(r)
	if err != nil || projectID <= 0 {
		writeJSON(
			w,
			http.StatusBadRequest,
			map[string]string{
				"err": "invalid project id",
			},
		)
		return
	}

	project, err := s.projects.ByID(
		r.Context(),
		projectID,
	)

	switch {
	case errors.Is(err, domain.ErrProjectNotFound):
		writeJSON(
			w,
			http.StatusNotFound,
			map[string]string{
				"error": "project not found",
			},
		)
		return

	case err != nil:
		writeJSON(
			w,
			http.StatusInternalServerError,
			map[string]string{
				"error": "internal server error",
			},
		)
		return
	}

	writeJSON(
		w,
		http.StatusOK,
		projectToResponse(project),
	)
}

func (s *Server) handleUpdateProject(w http.ResponseWriter, r *http.Request) {
	projectID, err := projectIDFromRequest(r)
	if err != nil || projectID <= 0 {
		writeJSON(
			w,
			http.StatusBadRequest,
			map[string]string{
				"error": "invalid project id",
			},
		)
		return
	}

	var request updateProjectRequest

	if !decodeJSON(w, r, &request) {
		return
	}

	project, err := s.projects.Update(
		r.Context(),
		projectID,
		request.Title,
		request.Description,
	)

	switch {
	case errors.Is(err, service.ErrProjectTitleRequired):
		writeJSON(
			w,
			http.StatusBadRequest,
			map[string]string{
				"error": "project title is required",
			},
		)
		return

	case errors.Is(err, domain.ErrProjectNotFound):
		writeJSON(
			w,
			http.StatusNotFound,
			map[string]string{
				"error": "project not found",
			},
		)
		return

	case err != nil:
		writeJSON(
			w,
			http.StatusInternalServerError,
			map[string]string{
				"error": "internal server error",
			},
		)
		return
	}

	writeJSON(
		w,
		http.StatusOK,
		projectToResponse(project),
	)
}

func (s *Server) handleDeleteProject(w http.ResponseWriter, r *http.Request) {
	projectID, err := projectIDFromRequest(r)
	if err != nil || projectID <= 0 {
		writeJSON(
			w,
			http.StatusBadRequest,
			map[string]string{
				"error": "invalid project id",
			},
		)
		return
	}

	err = s.projects.Delete(
		r.Context(),
		projectID,
	)

	switch {
	case errors.Is(err, domain.ErrProjectNotFound):
		writeJSON(
			w,
			http.StatusNotFound,
			map[string]string{
				"error": "project not found",
			},
		)
		return

	case err != nil:
		writeJSON(
			w,
			http.StatusInternalServerError,
			map[string]string{
				"error": "internal server error",
			},
		)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handlePublicProject(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")

	if slug == "" {
		writeJSON(
			w,
			http.StatusBadRequest,
			map[string]string{
				"error": "invalid project slug",
			},
		)
		return
	}

	project, err := s.projects.ByPublicSlug(
		r.Context(),
		slug,
	)

	switch {
	case errors.Is(err, domain.ErrProjectNotFound):
		writeJSON(
			w,
			http.StatusNotFound,
			map[string]string{
				"error": "project not found",
			},
		)
		return

	case err != nil:
		writeJSON(
			w,
			http.StatusInternalServerError,
			map[string]string{
				"error": "internal server error",
			},
		)
		return
	}

	writeJSON(
		w,
		http.StatusOK,
		projectToResponse(project),
	)
}

func (s *Server) handlePublicProjects(w http.ResponseWriter, r *http.Request) {
	projects, err := s.projects.ListPublic(r.Context())
	if err != nil {
		writeJSON(
			w,
			http.StatusInternalServerError,
			map[string]string{
				"error": "internal server error",
			},
		)
		return
	}

	response := make(
		[]projectResponse,
		0,
		len(projects),
	)

	for _, project := range projects {
		response = append(response, projectToResponse(project))
	}

	writeJSON(
		w,
		http.StatusOK,
		response,
	)
}

// !Project lifecycle handling
type projectLifecycleAction func(ctx context.Context, projectID int64) (*domain.Project, error)

func (s *Server) handleProjectLifecycle(w http.ResponseWriter, r *http.Request, action projectLifecycleAction) {
	projectID, err := projectIDFromRequest(r)
	if err != nil || projectID <= 0 {
		writeJSON(
			w,
			http.StatusBadRequest,
			map[string]string{
				"error": "invalid project id",
			},
		)
		return
	}

	project, err := action(
		r.Context(),
		projectID,
	)

	switch {
	case errors.Is(err, domain.ErrProjectNotFound):
		writeJSON(
			w,
			http.StatusNotFound,
			map[string]string{
				"error": "project not found",
			},
		)
		return

	case errors.Is(err, service.ErrInvalidProjectTransition):
		writeJSON(
			w,
			http.StatusConflict,
			map[string]string{
				"error": "invalid project status transition",
			},
		)
		return

	case err != nil:
		writeJSON(
			w,
			http.StatusInternalServerError,
			map[string]string{
				"error": "internal server error",
			},
		)
		return
	}

	writeJSON(
		w,
		http.StatusOK,
		projectToResponse(project),
	)
}

func (s *Server) handlePublishProject(w http.ResponseWriter, r *http.Request) {
	s.handleProjectLifecycle(
		w,
		r,
		s.projects.Publish,
	)
}

func (s *Server) handleUnpublishProject(w http.ResponseWriter, r *http.Request) {
	s.handleProjectLifecycle(
		w,
		r,
		s.projects.Unpublish,
	)
}

func (s *Server) handleUnlistProject(w http.ResponseWriter, r *http.Request) {
	s.handleProjectLifecycle(
		w,
		r,
		s.projects.Unlist,
	)
}

func (s *Server) handleArchiveProject(w http.ResponseWriter, r *http.Request) {
	s.handleProjectLifecycle(
		w,
		r,
		s.projects.Archive,
	)
}
