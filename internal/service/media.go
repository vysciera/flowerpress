package service

import (
	"context"
	"errors"
	"strings"

	"flowerpress/internal/domain"
)

var (
	ErrMediaStorageKeyRequired   = errors.New("media storage key is required")
	ErrMediaOriginalNameRequired = errors.New("media original name is required")
	ErrMediaMIMETypeRequired     = errors.New("media MIME type is required")
	ErrMediaSHA256Required       = errors.New("media SHA256 is required")

	ErrInvalidMediaSize          = errors.New("invalid media size")
	ErrInvalidMediaDimensions    = errors.New("invalid media dimensions")
	ErrInvalidMediaPlacementRole = errors.New("invalid media placement role")
	ErrInvalidMediaPosition      = errors.New("invalid media position")
)

type MediaService struct {
	assets     domain.MediaAssetRepository
	placements domain.MediaPlacementRepository
	projects   domain.ProjectRepository
}

func NewMediaService(assets domain.MediaAssetRepository, placements domain.MediaPlacementRepository, projects domain.ProjectRepository) *MediaService {
	return &MediaService{
		assets:     assets,
		placements: placements,
		projects:   projects,
	}
}

func (s *MediaService) RegisterAsset(
	ctx context.Context,
	storageKey string,
	originalName string,
	mimeType string,
	sizeBytes int64,
	sha256 string,
	width *int,
	height *int,
) (*domain.MediaAsset, error) {
	storageKey = strings.TrimSpace(storageKey)
	originalName = strings.TrimSpace(originalName)
	mimeType = strings.TrimSpace(mimeType)
	sha256 = strings.TrimSpace(sha256)

	switch {
	case storageKey == "":
		return nil, ErrMediaStorageKeyRequired

	case originalName == "":
		return nil, ErrMediaOriginalNameRequired

	case mimeType == "":
		return nil, ErrMediaMIMETypeRequired

	case sha256 == "":
		return nil, ErrMediaSHA256Required

	case sizeBytes < 0:
		return nil, ErrInvalidMediaSize

	case width != nil && *width <= 0:
		return nil, ErrInvalidMediaDimensions

	case height != nil && *height <= 0:
		return nil, ErrInvalidMediaDimensions
	}

	existing, err := s.assets.BySHA256(ctx, sha256)

	switch {
	case err == nil:
		return existing, nil

	case !errors.Is(err, domain.ErrMediaAssetNotFound):
		return nil, err
	}

	asset := &domain.MediaAsset{
		StorageKey:   storageKey,
		OriginalName: originalName,
		MIMEType:     mimeType,
		SizeBytes:    sizeBytes,
		SHA256:       sha256,
		Width:        width,
		Height:       height,
	}

	if err := s.assets.Create(ctx, asset); err != nil {
		existing, findErr := s.assets.BySHA256(ctx, sha256)
		if findErr == nil {
			return existing, nil
		}

		return nil, err
	}

	return asset, nil
}

func (s *MediaService) AssetByID(ctx context.Context, id int64) (*domain.MediaAsset, error) {
	return s.assets.ByID(ctx, id)
}

func (s *MediaService) ListAssets(ctx context.Context) ([]*domain.MediaAsset, error) {
	return s.assets.List(ctx)
}

func validPlacementRole(role domain.MediaPlacementRole) bool {
	switch role {
	case domain.MediaPlacementThumbnail,
		domain.MediaPlacementContent,
		domain.MediaPlacementAttachment:
		return true

	default:
		return false
	}
}

func (s *MediaService) PlaceAsset(
	ctx context.Context,
	projectID int64,
	assetID int64,
	role domain.MediaPlacementRole,
	position int,
	caption string,
	altText string,
) (*domain.MediaPlacement, error) {
	if !validPlacementRole(role) {
		return nil, ErrInvalidMediaPlacementRole
	}

	if position < 0 {
		return nil, ErrInvalidMediaPosition
	}

	if _, err := s.projects.ByID(ctx, projectID); err != nil {
		return nil, err
	}

	if _, err := s.assets.ByID(ctx, assetID); err != nil {
		return nil, err
	}

	placement := &domain.MediaPlacement{
		AssetID:   assetID,
		ProjectID: projectID,
		Role:      role,
		Position:  position,
		Caption:   strings.TrimSpace(caption),
		AltText:   strings.TrimSpace(altText),
	}

	if err := s.placements.Create(ctx, placement); err != nil {
		return nil, err
	}

	return placement, nil
}

func (s *MediaService) ListProjectMedia(ctx context.Context, projectID int64) ([]*domain.MediaPlacement, error) {
	if _, err := s.projects.ByID(ctx, projectID); err != nil {
		return nil, err
	}

	return s.placements.ListByProject(ctx, projectID)
}

func (s *MediaService) RemovePlacement(ctx context.Context, placementID int64) error {
	return s.placements.Delete(ctx, placementID)
}
