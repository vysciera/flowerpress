package domain

import (
	"context"
	"errors"
)

var (
	ErrMediaPlacementNotFound      = errors.New("media placement not found")
	ErrMediaPlacementOrderMismatch = errors.New("media placement order does not match project media")
)

type MediaPlacementRepository interface {
	Create(ctx context.Context, placement *MediaPlacement) error
	Update(ctx context.Context, placement *MediaPlacement) error
	Delete(ctx context.Context, id int64) error

	Reorder(ctx context.Context, projectID int64, role MediaPlacementRole, placementsIDs []int64) error

	ByID(ctx context.Context, id int64) (*MediaPlacement, error)
	ListByProject(ctx context.Context, projectID int64) ([]*MediaPlacement, error)
}
