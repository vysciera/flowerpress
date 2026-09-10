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

var _ domain.MediaAssetRepository = (*MediaAssetRepository)(nil)

func (r *MediaAssetRepository) Create(ctx context.Context, asset *domain.MediaAsset) error {
	result, err := r.db.ExecContext(
		ctx,
		`
			INSERT INTO media_assets (
				storage_key,
				original_name,
				mime_type,
				size_bytes,
				sha256,
				width,
				height
			)
			VALUES (?, ?, ?, ?, ?, ?, ?)	
		`,
		asset.StorageKey,
		asset.OriginalName,
		asset.MIMEType,
		asset.SizeBytes,
		asset.SHA256,
		nullableInt(asset.Width),
		nullableInt(asset.Height),
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

func (r *MediaAssetRepository) Update(ctx context.Context, asset *domain.MediaAsset) error {
	result, err := r.db.ExecContext(
		ctx,
		`
			UPDATE media_assets
			SET
				storage_key = ?,
				original_name = ?,
				mime_type = ?,
				size_bytes = ?,
				sha256 = ?,
				width = ?,
				height = ?,
				updated_at = CURRENT_TIMESTAMP
			WHERE id = ?
		`,
		asset.StorageKey,
		asset.OriginalName,
		asset.MIMEType,
		asset.SizeBytes,
		asset.SHA256,
		nullableInt(asset.Width),
		nullableInt(asset.Height),
		asset.ID,
	)
	if err != nil {
		return fmt.Errorf("update media asset: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get media asset rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrMediaAssetNotFound
	}

	updated, err := r.ByID(ctx, asset.ID)
	if err != nil {
		return fmt.Errorf("reload updated media asset: %w", err)
	}

	*asset = *updated

	return nil
}

func (r *MediaAssetRepository) Delete(ctx context.Context, id int64) error {
	result, err := r.db.ExecContext(
		ctx,
		`
			DELETE FROM media_assets
			WHERE id = ?
		`,
		id,
	)
	if err != nil {
		return fmt.Errorf("delete media asset: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get media asset rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrMediaAssetNotFound
	}

	return nil
}

func (r *MediaAssetRepository) ByID(ctx context.Context, id int64) (*domain.MediaAsset, error) {
	row := r.db.QueryRowContext(
		ctx,
		`
			SELECT
				id,
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

func (r *MediaAssetRepository) BySHA256(ctx context.Context, hash string) (*domain.MediaAsset, error) {
	row := r.db.QueryRowContext(
		ctx, // ok man.
		`
			SELECT
				id,
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
			WHERE sha256 = ?
		`,
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

func (r *MediaAssetRepository) List(ctx context.Context) ([]*domain.MediaAsset, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`
			SELECT
				id,
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
			ORDER BY created_at DESC, id DESC
		`,
	)
	if err != nil {
		return nil, fmt.Errorf("list media assets: %w", err)
	}
	defer rows.Close()

	assets := make([]*domain.MediaAsset, 0)
	for rows.Next() {
		asset, err := scanMediaAsset(rows)
		if err != nil {
			return nil, fmt.Errorf("scan media asset: %w", err)
		}

		assets = append(assets, asset)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate media assets: %w", err)
	}

	return assets, nil
}

func scanMediaAsset(row scanner) (*domain.MediaAsset, error) {
	var (
		asset     domain.MediaAsset
		width     sql.NullInt64
		height    sql.NullInt64
		createdAt string
		updatedAt string
	)

	if err := row.Scan(
		&asset.ID,
		&asset.StorageKey,
		&asset.OriginalName,
		&asset.MIMEType,
		&asset.SizeBytes,
		&asset.SHA256,
		&width,
		&height,
		&createdAt,
		&updatedAt,
	); err != nil {
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

	var err error

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

func nullableInt(value *int) any {
	if value == nil {
		return nil
	}

	return *value
}
