package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode"

	"flowerpress/internal/domain"
)

var (
	ErrProjectTitleRequired     = errors.New("project title is required")
	ErrInvalidProjectTransition = errors.New("invalid project status transition")
)

type ProjectService struct {
	projects domain.ProjectRepository
}

func NewProjectService(projects domain.ProjectRepository) *ProjectService {
	return &ProjectService{
		projects: projects,
	}
}

func (s *ProjectService) Create(ctx context.Context, title string, description string) (*domain.Project, error) {
	title = strings.TrimSpace(title)

	if title == "" {
		return nil, ErrProjectTitleRequired
	}

	baseSlug := slugify(title)

	project := &domain.Project{
		Title:       title,
		Description: strings.TrimSpace(description),
		Status:      domain.ProjectStatusDraft,
	}

	// Slug sequencing for duplicate titles
	for suffix := 1; ; suffix++ {
		project.Slug = sequencedSlug(
			baseSlug,
			suffix,
		)

		err := s.projects.Create(ctx, project)

		if err == nil {
			return project, nil
		}

		if !errors.Is(err, domain.ErrSlugTaken) {
			return nil, err
		}
	}
}

func (s *ProjectService) Update(ctx context.Context, projectID int64, title string, description string) (*domain.Project, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return nil, ErrProjectTitleRequired
	}

	project, err := s.projects.ByID(ctx, projectID)
	if err != nil {
		return nil, err
	}

	project.Title = title
	project.Description = strings.TrimSpace(description)

	if err := s.projects.Update(ctx, project); err != nil {
		return nil, err
	}

	return project, nil
}

func (s *ProjectService) Delete(ctx context.Context, projectID int64) error {
	project, err := s.projects.ByID(ctx, projectID)
	if err != nil {
		return err
	}

	return s.projects.Delete(ctx, project.ID)
}

func (s *ProjectService) Publish(ctx context.Context, projectID int64) (*domain.Project, error) {
	project, err := s.projects.ByID(ctx, projectID)
	if err != nil {
		return nil, err
	}

	switch project.Status {
	case domain.ProjectStatusArchived:
		return nil, ErrInvalidProjectTransition

	case domain.ProjectStatusPublished:
		return project, nil
	}

	// PublishedAt records first time publication - not currently reset on unpublish/republish
	project.Status = domain.ProjectStatusPublished

	if project.PublishedAt == nil {
		now := time.Now().UTC()
		project.PublishedAt = &now
	}

	if err := s.projects.Update(ctx, project); err != nil {
		return nil, err
	}

	return project, nil
}

func (s *ProjectService) Unpublish(ctx context.Context, projectID int64) (*domain.Project, error) {
	project, err := s.projects.ByID(ctx, projectID)
	if err != nil {
		return nil, err
	}

	switch project.Status {
	case domain.ProjectStatusArchived:
		return nil, ErrInvalidProjectTransition

	case domain.ProjectStatusDraft:
		return project, nil
	}

	// Same here
	project.Status = domain.ProjectStatusDraft

	if err := s.projects.Update(ctx, project); err != nil {
		return nil, err
	}

	return project, nil
}

func (s *ProjectService) Unlist(ctx context.Context, projectID int64) (*domain.Project, error) {
	project, err := s.projects.ByID(ctx, projectID)
	if err != nil {
		return nil, err
	}

	switch project.Status {
	case domain.ProjectStatusArchived:
		return nil, ErrInvalidProjectTransition

	case domain.ProjectStatusUnlisted:
		return project, nil
	}

	project.Status = domain.ProjectStatusUnlisted

	if project.PublishedAt == nil {
		now := time.Now().UTC()
		project.PublishedAt = &now
	}

	if err := s.projects.Update(ctx, project); err != nil {
		return nil, err
	}

	return project, nil
}

func (s *ProjectService) Archive(ctx context.Context, projectID int64) (*domain.Project, error) {
	project, err := s.projects.ByID(ctx, projectID)
	if err != nil {
		return nil, err
	}

	if project.Status == domain.ProjectStatusArchived {
		return project, nil
	}

	project.Status = domain.ProjectStatusArchived

	if err := s.projects.Update(ctx, project); err != nil {
		return nil, err
	}

	return project, nil
}

func (s *ProjectService) ByID(ctx context.Context, projectID int64) (*domain.Project, error) {
	return s.projects.ByID(ctx, projectID)
}

func (s *ProjectService) List(ctx context.Context) ([]*domain.Project, error) {
	return s.projects.List(ctx)
}

func (s *ProjectService) ByPublicSlug(ctx context.Context, slug string) (*domain.Project, error) {
	project, err := s.projects.BySlug(ctx, slug)
	if err != nil {
		return nil, err
	}

	switch project.Status {
	case domain.ProjectStatusPublished, domain.ProjectStatusUnlisted:
		return project, nil

	default:
		return nil, domain.ErrProjectNotFound
	}
}

func (s *ProjectService) ListPublic(ctx context.Context) ([]*domain.Project, error) {
	return s.projects.ListPublished(ctx)
}

// Misc
func slugify(value string) string {
	value = strings.ToLower(
		strings.TrimSpace(value),
	)

	var builder strings.Builder
	previousHyphen := false

	for _, r := range value {
		switch {
		case unicode.IsLetter(r), unicode.IsDigit(r):
			builder.WriteRune(r)
			previousHyphen = false

		case !previousHyphen && builder.Len() > 0:
			builder.WriteByte('-')
			previousHyphen = true
		}
	}

	return strings.Trim(
		builder.String(),
		"-",
	)
}

func sequencedSlug(base string, number int) string {
	if number <= 1 {
		return base
	}

	return fmt.Sprintf(
		"%s-%d",
		base,
		number,
	)
}
