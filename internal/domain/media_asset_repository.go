package domain

import (
	"context"
	"errors"
)

var ErrMediaAssetNotFound = errors.New("media asset not found")

type MediaAssetRepository interface {
	Create(ctx context.Context, asset *MediaAsset) error
	Update(ctx context.Context, asset *MediaAsset) error
	Delete(ctx context.Context, asset *MediaAsset) error

	ByID(ctx context.Context, id int64) (*MediaAsset, error)
	ByStorageKey(ctx context.Context, storageKey string) (*MediaAsset, error)
	BySHA256(ctx context.Context, ownerID int64, hash string) (*MediaAsset, error)

	ListByOwner(ctx context.Context, ownerID int64) ([]*MediaAsset, error)
}
