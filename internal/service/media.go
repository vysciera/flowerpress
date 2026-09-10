package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
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

	ErrProjectThumbnailExists	 = errors.New("project already has a thumbnail")
)

type byteCounter struct {
	n int64
}

type MediaService struct {
	assets     domain.MediaAssetRepository
	placements domain.MediaPlacementRepository
	projects   domain.ProjectRepository
	storage    MediaStorage
}

func (c *byteCounter) Write(p []byte) (int, error) {
	c.n += int64(len(p))
	return len(p), nil
}

func newMediaStorageKey() (string, error) {
	var random [16]byte

	if _, err := rand.Read(random[:]); err != nil {
		return "", fmt.Errorf("generate media storage key: %w", err)
	}

	return "objects/" + hex.EncodeToString(random[:]), nil
}

func (s *MediaService) UploadAsset(ctx context.Context, originalName string, mimeType string, source io.Reader, width, height *int) (*domain.MediaAsset, error) {
	originalName = strings.TrimSpace(originalName)
	mimeType = strings.TrimSpace(mimeType)

	switch {
	case originalName == "":
		return nil, ErrMediaOriginalNameRequired

	case mimeType == "":
		return nil, ErrMediaMIMETypeRequired

	case source == nil:
		return nil, errors.New("media source is required")

	case width != nil && *width <= 0:
		return nil, ErrInvalidMediaDimensions

	case height != nil && *height <= 0:
		return nil, ErrInvalidMediaDimensions
	}

	storageKey, err := newMediaStorageKey()
	if err != nil {
		return nil, err
	}

	hasher := sha256.New()
	counter := &byteCounter{}

	reader := io.TeeReader(source, io.MultiWriter(hasher, counter))
	if err := s.storage.Put(ctx, storageKey, reader); err != nil {
		return nil, fmt.Errorf("store uploaded media: %w", err)
	}

	hash := hex.EncodeToString(hasher.Sum(nil))
	existing, err := s.assets.BySHA256(ctx, hash)

	switch {
	case err == nil:
		if err := s.storage.Delete(ctx, storageKey); err != nil { // Review later
			return nil, fmt.Errorf("remove duplicate media object: %w", err)
		}

		return existing, nil

	case !errors.Is(err, domain.ErrMediaAssetNotFound):
		_ = s.storage.Delete(ctx, storageKey)

		return nil, err
	}

	asset := &domain.MediaAsset{
		StorageKey:   storageKey,
		OriginalName: originalName,
		MIMEType:     mimeType,
		SizeBytes:    counter.n,
		SHA256:       hash,
		Width:        width,
		Height:       height,
	}

	if err := s.assets.Create(ctx, asset); err != nil {
		existing, findErr := s.assets.BySHA256(ctx, hash)

		if findErr == nil {
			if deleteErr := s.storage.Delete(ctx, storageKey); deleteErr != nil {
				return nil, fmt.Errorf("remove concurrent duplicate: %w", deleteErr)
			}

			return existing, nil
		}

		_ = s.storage.Delete(ctx, storageKey)

		return nil, err
	}

	return asset, nil
}

func (s *MediaService) OpenAssetContent(ctx context.Context, id int64) (*domain.MediaAsset, io.ReadCloser, error) {
	asset, err := s.assets.ByID(ctx, id)
	if err != nil {
		return nil, nil, err
	}

	content, err := s.storage.Open(ctx, asset.StorageKey)
	if err != nil {
		return nil, nil, fmt.Errorf("open media content: %w", err)
	}

	return asset, content, nil
}

func NewMediaService(
	assets domain.MediaAssetRepository,
	placements domain.MediaPlacementRepository,
	projects domain.ProjectRepository,
	storage MediaStorage,
) *MediaService {
	return &MediaService{
		assets:     assets,
		placements: placements,
		projects:   projects,
		storage:    storage,
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

// Please for the love of god reorganize all your file structures soon

func (s *MediaService) UpdatePlacement(
	ctx context.Context,
	placementID int64,
	role domain.MediaPlacementRole,
	position int,
	caption string,
	altText string,
) (*domain.MediaPlacement, error) {
	if !validPlacementRole(role) {
		return nil, ErrInvalidMediaPlacementRole
	}

	if err := validatePlacementPosition(role, position); err != nil {
		return nil, err
	}

	placement, err := s.placements.ByID(ctx, placementID)
	if err != nil {
		return nil, err
	}

	if role == domain.MediaPlacementThumbnail {
		if err := s.ensureThumbnailAvailable(ctx, placement.ProjectID, placement.ID); err != nil {
			return nil, err
		}
	}

	placement.Role = role
	placement.Position = position
	placement.Caption = caption
	placement.AltText = altText

	if err := s.placements.Update(ctx, placement); err != nil {
		return nil, err
	}

	return placement, nil
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

	if err := validatePlacementPosition(role, position); err != nil {
		return nil, err
	}

	if _, err := s.projects.ByID(ctx, projectID); err != nil {
		return nil, err
	}

	if _, err := s.assets.ByID(ctx, assetID); err != nil {
		return nil, err
	}

	if role == domain.MediaPlacementThumbnail {
		if err := s.ensureThumbnailAvailable(ctx, projectID, 0); err != nil {
			return nil, err
		}
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

func (s *MediaService) ensureThumbnailAvailable(ctx context.Context, projectID int64, exceptPlacementID int64) error {
	placements, err := s.placements.ListByProject(ctx, projectID)
	if err != nil {
		return err
	}

	for _, placement := range placements {
		if placement.Role != domain.MediaPlacementThumbnail {
			continue
		}

		if placement.ID == exceptPlacementID {
			continue
		}

		return ErrProjectThumbnailExists
	}

	return nil
}

// validatePlacementPosition ensures thumbnails always have position 0
func validatePlacementPosition(role domain.MediaPlacementRole, position int) error {
	if position < 0 {
		return ErrInvalidMediaPosition
	}

	if role == domain.MediaPlacementThumbnail && position != 0 {
		return ErrInvalidMediaPosition
	}

	return nil
}
