package domain

import (
	"context"
	"errors"
)

var ErrMediaPlacementNotFound = errors.New("media placement not found")

type MediaPlacementRepository interface {
	Create(ctx context.Context, placement *MediaPlacement) error
	Update(ctx context.Context, placement *MediaPlacement) error
	Delete(ctx context.Context, id int64) error

	ById(ctx context.Context, id int64) (*MediaPlacement, error)
	ListByProject(ctx context.Context, projectID int64) ([]*MediaPlacement, error)
}
