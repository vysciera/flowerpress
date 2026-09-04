package turso

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"flowerpress/internal/domain"
)

type ProjectRepository struct {
	db *sql.DB
}

func NewProjectRepository(db *sql.DB) *ProjectRepository {
	return &ProjectRepository{
		db: db,
	}
}

func (r *ProjectRepository) Create(ctx context.Context, project *domain.Project) error {
	result, err := r.db.ExecContext(
		ctx,
		`
			INSERT INTO projects (
				owner_id,
				title,
				slug,
				description,
				status,
				published_at
			)
			VALUES (?, ?, ?, ?, ?, ?)
			ON CONFLICT(slug) DO NOTHING	
		`,
		project.OwnerID,
		project.Title,
		project.Slug,
		project.Description,
		project.Status,
		formatNullableTimestamp(project.PublishedAt),
	)

	if err != nil {
		return fmt.Errorf(
			"create project: %w",
			err,
		)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf(
			"get affected rows: %w",
			err,
		)
	}

	if rows == 0 {
		return domain.ErrSlugTaken
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf(
			"get project id: %w",
			err,
		)
	}

	project.ID = id
	created, err := r.ByID(ctx, id)

	if err != nil {
		return fmt.Errorf(
			"reload created project: %w",
			err,
		)
	}

	*project = *created

	return nil
}

func (r *ProjectRepository) ByID(ctx context.Context, id int64) (*domain.Project, error) {
	row := r.db.QueryRowContext(
		ctx,
		`
			SELECT
				id,
				owner_id,
				title,
				slug,
				description,
				status,
				published_at,
				created_at,
				updated_at
			FROM projects
			WHERE id = ?	
		`,
		id,
	)

	project, err := scanProject(row)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrProjectNotFound
		}

		return nil, fmt.Errorf(
			"get project by id: %w",
			err,
		)
	}

	return project, nil
}

func (r *ProjectRepository) BySlug(ctx context.Context, slug string) (*domain.Project, error) {
	row := r.db.QueryRowContext(
		ctx,
		`
			SELECT
				id,
				owner_id,
				title,
				slug,
				description,
				status,
				published_at,
				created_at,
				updated_at
			FROM projects
			WHERE slug = ?	
		`,
		slug,
	)

	project, err := scanProject(row)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrProjectNotFound
		}

		return nil, fmt.Errorf(
			"get project by slug: %w",
			err,
		)
	}

	return project, nil
}

func scanProject(row scanner) (*domain.Project, error) {
	var (
		project     domain.Project
		publishedAt sql.NullString
		createdAt   string
		updatedAt   string
	)

	err := row.Scan(
		&project.ID,
		&project.OwnerID,
		&project.Title,
		&project.Slug,
		&project.Description,
		&project.Status,
		&publishedAt,
		&createdAt,
		&updatedAt,
	)

	if err != nil {
		return nil, err
	}

	if publishedAt.Valid {
		value, err := parseTimestamp(
			publishedAt.String,
		)

		if err != nil {
			return nil, fmt.Errorf(
				"parse published_at: %w",
				err,
			)
		}

		project.PublishedAt = &value
	}

	var errCreated error

	project.CreatedAt, errCreated = parseTimestamp(
		createdAt,
	)

	if errCreated != nil {
		return nil, fmt.Errorf(
			"parse created_at: %w",
			errCreated,
		)
	}

	var errUpdated error

	project.UpdatedAt, errUpdated = parseTimestamp(
		updatedAt,
	)

	if errUpdated != nil {
		return nil, fmt.Errorf(
			"parse updated_at: %w",
			errUpdated,
		)
	}

	return &project, nil
}
