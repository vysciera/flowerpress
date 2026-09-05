package domain

import (
	"context"
	"errors"
)

var ErrMediaAssetNotFound = errors.New("media asset not found")

type MediaAssetRepository interface {
	Create(ctx context.Context, asset *MediaAsset) error
	Update(ctx context.Context, asset *MediaAsset) error
	Delete(ctx context.Context, id int64) error

	ByID(ctx context.Context, id int64) (*MediaAsset, error)
	ByStorageKey(ctx context.Context, storageKey string) (*MediaAsset, error)
	BySHA256(ctx context.Context, hash string) (*MediaAsset, error)

	List(ctx context.Context) ([]*MediaAsset, error)
}
