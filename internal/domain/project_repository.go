package domain

import (
	"context"
	"errors"
)

var (
	ErrProjectNotFound = errors.New("project not found")
	ErrSlugTaken       = errors.New("project slug already taken")
)

type ProjectRepository interface {
	Create(ctx context.Context, project *Project) error

	ByID(ctx context.Context, id int64) (*Project, error)
	BySlug(ctx context.Context, slug string) (*Project, error)
	List(ctx context.Context) ([]*Project, error)

	Update(ctx context.Context, project *Project) error
	Delete(ctx context.Context, id int64) error

	ListPublished(ctx context.Context) ([]*Project, error)
}
