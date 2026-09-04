package service

import (
	"context"
	"errors"
	"strings"
	"unicode"

	"flowerpress/internal/domain"
)

var ErrProjectTitleRequired = errors.New("project title is required")

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

	project := &domain.Project{
		OwnerID:     ownerID,
		Title:       title,
		Slug:        slugify(title),
		Description: strings.TrimSpace(description),
		Status:      domain.ProjectStatusDraft,
	}

	if err := s.projects.Create(
		ctx,
		project,
	); err != nil {
		return nil, err
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
