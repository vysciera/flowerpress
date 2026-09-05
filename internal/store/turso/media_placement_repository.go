package turso

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"flowerpress/internal/domain"
)

type MediaPlacementRepository struct {
	db *sql.DB
}

func NewMediaPlacementRepository(db *sql.DB) *MediaPlacementRepository {
	return &MediaPlacementRepository{
		db: db,
	}
}

var _ domain.MediaPlacementRepository = (*MediaPlacementRepository)(nil)

func (r *MediaPlacementRepository) Create(ctx context.Context, placement *domain.MediaPlacement) error {
	result, err := r.db.ExecContext(
		ctx,
		`
			INSERT INTO media_placements (
				asset_id,
				project_id,
				role,
				position,
				caption,
				alt_text
			)
			VALUES (?, ?, ?, ?, ?, ?)
		`,
		placement.AssetID,
		placement.ProjectID,
		placement.Role,
		placement.Position,
		placement.Caption,
		placement.AltText,
	)
	if err != nil {
		return fmt.Errorf(
			"create media placement: %w",
			err,
		)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf(
			"get media placement id: %w",
			err,
		)
	}

	placement.ID = id

	created, err := r.ByID(ctx, id)
	if err != nil {
		return fmt.Errorf(
			"reload created media placement: %w",
			err,
		)
	}

	*placement = *created

	return nil
}

func (r *MediaPlacementRepository) ByID(ctx context.Context, id int64) (*domain.MediaPlacement, error) {
	row := r.db.QueryRowContext(
		ctx,
		`
			SELECT
				id,
				asset_id,
				project_id,
				role,
				position,
				caption,
				alt_text,
				created_at,
				updated_at
			FROM media_placements
			WHERE id = ?
		`,
		id,
	)

	placement, err := scanMediaPlacement(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrMediaPlacementNotFound
		}

		return nil, fmt.Errorf("get media placement by id: %w", err)
	}

	return placement, nil
}

func (r *MediaPlacementRepository) ListByProject(ctx context.Context, projectID int64) ([]*domain.MediaPlacement, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`
			SELECT
				id,
				asset_id,
				project_id,
				role,
				position,
				caption,
				alt_text,
				created_at,
				updated_at
			FROM media_placements
			WHERE project_id = ?
			ORDER BY position ASC, id ASC
		`,
		projectID,
	)
	if err != nil {
		return nil, fmt.Errorf("list media placements: %w", err)
	}
	defer rows.Close()

	placements := make([]*domain.MediaPlacement, 0)

	for rows.Next() {
		placement, err := scanMediaPlacement(rows)
		if err != nil {
			return nil, fmt.Errorf("scan media placement: %w", err)
		}

		placements = append(placements, placement)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate media placements: %w", err)
	}

	return placements, nil
}

func (r *MediaPlacementRepository) Update(ctx context.Context, placement *domain.MediaPlacement) error {
	result, err := r.db.ExecContext(
		ctx,
		`
			UPDATE media_placements
			SET
				asset_id = ?,
				project_id = ?,
				role = ?,
				position = ?,
				caption = ?,
				alt_text = ?,
				updated_at = CURRENT_TIMESTAMP
			WHERE id = ?
		`,
		placement.AssetID,
		placement.ProjectID,
		placement.Role,
		placement.Position,
		placement.Caption,
		placement.AltText,
		placement.ID,
	)
	if err != nil {
		return fmt.Errorf("update media placement: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get media placement rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrMediaPlacementNotFound
	}

	updated, err := r.ByID(ctx, placement.ID)
	if err != nil {
		return fmt.Errorf("reload updated media placement: %w", err)
	}

	*placement = *updated

	return nil
}

func (r *MediaPlacementRepository) Delete(ctx context.Context, id int64) error {
	result, err := r.db.ExecContext(
		ctx,
		`
			DELETE FROM media_placements
			WHERE id = ?
		`,
		id,
	)
	if err != nil {
		return fmt.Errorf("delete media placement: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get media placement rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return domain.ErrMediaPlacementNotFound
	}

	return nil
}

func scanMediaPlacement(
	row interface {
		Scan(dest ...any) error
	},
) (*domain.MediaPlacement, error) {
	var (
		placement domain.MediaPlacement
		createdAt string
		updatedAt string
	)

	if err := row.Scan(
		&placement.ID,
		&placement.AssetID,
		&placement.ProjectID,
		&placement.Role,
		&placement.Position,
		&placement.Caption,
		&placement.AltText,
		&createdAt,
		&updatedAt,
	); err != nil {
		return nil, err
	}

	var err error

	placement.CreatedAt, err = parseTimestamp(createdAt)
	if err != nil {
		return nil, fmt.Errorf("parse created_at: %w", err)
	}

	placement.UpdatedAt, err = parseTimestamp(updatedAt)
	if err != nil {
		return nil, fmt.Errorf("parse updated_at: %w", err)
	}

	return &placement, nil
}
