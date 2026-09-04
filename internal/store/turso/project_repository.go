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

var _ domain.ProjectRepository = (*ProjectRepository)(nil)

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

func (r *ProjectRepository) ListByOwner(ctx context.Context, ownerID int64) ([]*domain.Project, error) {
	rows, err := r.db.QueryContext(
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
			WHERE owner_id = ?
			ORDER BY created_at DESC, id DESC
		`,
		ownerID,
	)

	if err != nil {
		return nil, fmt.Errorf(
			"list projects by owner: %w",
			err,
		)
	}

	defer rows.Close()
	var projects []*domain.Project

	for rows.Next() {
		project, err := scanProject(rows)

		if err != nil {
			return nil, fmt.Errorf("scan project: %w", err)
		}

		projects = append(projects, project)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate projects: %w", err)
	}

	return projects, nil
}

func (r *ProjectRepository) Update(ctx context.Context, project *domain.Project) error {
	result, err := r.db.ExecContext(
		ctx, // Review this utter buffoonery at a later time
		`
			UPDATE projects
			SET
				title = ?,
				slug = ?,
				description = ?,
				status = ?,
				published_at = ?,
				updated_at = CURRENT_TIMESTAMP
			WHERE id = ?
				AND NOT EXISTS (
					SELECT 1
					FROM projects
					WHERE SLUG = ?
						AND id <> ?
				)	
		`,
		project.Title,
		project.Slug,
		project.Description,
		project.Status,
		formatNullableTimestamp(
			project.PublishedAt,
		),
		project.ID,
		project.Slug,
		project.ID,
	)

	if err != nil {
		return fmt.Errorf("udpate project: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get affected rows: %w", err)
	}

	if rows == 0 {
		_, err := r.ByID(ctx, project.ID)

		if errors.Is(err, domain.ErrProjectNotFound) {
			return domain.ErrProjectNotFound
		}

		if err != nil {
			return err
		}

		return domain.ErrSlugTaken
	}

	updated, err := r.ByID(ctx, project.ID)
	if err != nil {
		return fmt.Errorf("reload updated project: %w", err)
	}

	*project = *updated
	return nil
}

func (r *ProjectRepository) Delete(ctx context.Context, id int64) error {
	result, err := r.db.ExecContext(
		ctx,
		`
			DELETE FROM projects
			WHERE id = ?	
		`,
		id,
	)

	if err != nil {
		return fmt.Errorf("delete project: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get affected rows: %w", err)
	}

	if rows == 0 {
		return domain.ErrProjectNotFound
	}

	return nil
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
