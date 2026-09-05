package turso

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"flowerpress/internal/domain"
)

type MediaAssetRepository struct {
	db *sql.DB
}

func NewMediaAssetRepository(db *sql.DB) *MediaAssetRepository {
	return &MediaAssetRepository{
		db: db,
	}
}

func (r *MediaAssetRepository) Create(ctx context.Context, asset *domain.MediaAsset) error {
	result, err := r.db.ExecContext(
		ctx,
		`
			INSERT INTO media_assets (
				owner_id,
				storage_key,
				original_name,
				mime_types,
				size_bytes,
				sha256,
				width,
				height
			)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?)	
		`,
		asset.OwnerID,
		asset.StorageKey,
		asset.OriginalName,
		asset.MIMEType,
		asset.SizeBytes,
		asset.SHA256,
		asset.Width,
		asset.Height,
	)
	if err != nil {
		return fmt.Errorf("create media asset: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("get media asset id: %w", err)
	}

	asset.ID = id

	created, err := r.ByID(ctx, id)
	if err != nil {
		return fmt.Errorf("reload media asset: %w", err)
	}

	*asset = *created

	return nil
}

func (r *MediaAssetRepository) ByID(ctx context.Context, id int64) (*domain.MediaAsset, error) {
	row := r.db.QueryRowContext(
		ctx,
		`
			SELECT
				id,
				owner_id,
				storage_key,
				original_name,
				mime_type,
				size_bytes,
				sha256,
				width,
				height,
				created_at,
				updated_at
			FROM media_assets
			WHERE id = ?	
		`,
		id,
	)

	asset, err := scanMediaAsset(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrMediaAssetNotFound
		}

		return nil, fmt.Errorf("get media asset by id: %w", err)
	}

	return asset, nil
}

func (r *MediaAssetRepository) ByStorageKey(ctx context.Context, storageKey string) (*domain.MediaAsset, error) {
	row := r.db.QueryRowContext(
		ctx,
		`
			SELECT
				id,
				owner_id,
				storage_key,
				original_name,
				mime_type,
				size_bytes,
				sha256,
				width,
				height,
				created_at,
				updated_at
			FROM media_assets
			WHERE storage_key = ?	
		`,
		storageKey,
	)

	asset, err := scanMediaAsset(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrMediaAssetNotFound
		}

		return nil, fmt.Errorf("get media by aseet storage key: %w", err)
	}

	return asset, nil
}

func (r *MediaAssetRepository) BySHA256(ctx context.Context, ownerID int64, hash string) (*domain.MediaAsset, error) {
	row := r.db.QueryRowContext(
		ctx, // LIMIT 1 (??? its a SHA)
		`
			SELECT
				id,
				owner_id,
				storage_key,
				original_name,
				mime_type,
				size_bytes,
				sha256,
				width,
				height,
				created_at,
				updated_at
			FROM media_assets
			WHERE owner_id = ?
				AND sha256 = ?
			ORDER BY id ASC
			LIMIT 1	
		`,
		ownerID,
		hash,
	)

	asset, err := scanMediaAsset(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrMediaAssetNotFound
		}

		return nil, fmt.Errorf("get media asset by sha256: %w", err)
	}

	return asset, nil
}

func scanMediaAsset(row scanner) (*domain.MediaAsset, error) {
	var (
		asset     domain.MediaAsset
		width     sql.NullInt64
		height    sql.NullInt64
		createdAt string
		updatedAt string
	)

	err := row.Scan(
		&asset.ID,
		&asset.OwnerID,
		&asset.StorageKey,
		&asset.OriginalName,
		&asset.MIMEType,
		&asset.SizeBytes,
		&asset.SHA256,
		&width,
		&height,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		return nil, err
	}

	if width.Valid {
		value := int(width.Int64)
		asset.Width = &value
	}

	if height.Valid {
		value := int(height.Int64)
		asset.Height = &value
	}

	asset.CreatedAt, err = parseTimestamp(createdAt)
	if err != nil {
		return nil, fmt.Errorf("parse created_at: %w", err)
	}

	asset.UpdatedAt, err = parseTimestamp(updatedAt)
	if err != nil {
		return nil, fmt.Errorf("parse updated_at: %w", err)
	}

	return &asset, nil
}
