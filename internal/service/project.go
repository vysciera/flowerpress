package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
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

func (s *ProjectService) Create(ctx context.Context, ownerID int64, title string, description string) (*domain.Project, error) {
	title = strings.TrimSpace(title)

	if title == "" {
		return nil, ErrProjectTitleRequired
	}

	baseSlug := slugify(title)

	project := &domain.Project{
		OwnerID:     ownerID,
		Title:       title,
		Slug:        baseSlug,
		Description: strings.TrimSpace(description),
		Status:      domain.ProjectStatusDraft,
	}

	// Slug sequencing for duplicate titles
	for suffix := 1; ; suffix++ {
		if suffix == 1 {
			project.Slug = baseSlug
		} else {
			project.Slug = fmt.Sprintf(
				"%s-%d",
				baseSlug,
				suffix,
			)
		}

		err := s.projects.Create(ctx, project)

		if err == nil {
			return project, nil
		}

		if !errors.Is(err, domain.ErrSlugTaken) {
			return nil, err
		}
	}
}

func (s *ProjectService) Update(ctx context.Context, ownerID int64, projectID int64, title string, description string) (*domain.Project, error) {
	title = strings.TrimSpace(title)

	if title == "" {
		return nil, ErrProjectTitleRequired
	}

	project, err := s.projectForOwner(ctx, ownerID, projectID)
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

func (s *ProjectService) projectForOwner(ctx context.Context, ownerID int64, projectID int64) (*domain.Project, error) {
	project, err := s.projects.ByID(ctx, projectID)

	if err != nil {
		return nil, err
	}

	if project.OwnerID != ownerID {
		return nil, domain.ErrProjectNotFound
	}

	return project, nil
}

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
