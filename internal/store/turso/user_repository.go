package turso

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"flowerpress/internal/domain"
)

type UserRepository struct {
	db *sql.DB
}

// Compile-time interface check
var _ domain.UserRepository = (*UserRepository)(nil)

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	result, err := r.db.ExecContext(
		ctx,
		`
			INSERT INTO users (
				username,
				password_hash,
				recovery_hash,
				session_version
			)
				VALUES (?, ?, ?, ?)
		`,
		user.Username,
		user.PasswordHash,
		user.RecoveryHash,
		user.SessionVersion,
	)

	if err != nil {
		return fmt.Errorf("create user: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("get user id: %w", err)
	}

	user.ID = id

	created, err := r.ByID(ctx, id)
	if err != nil {
		return fmt.Errorf("reload created user: %w", err)
	}

	*user = *created

	return nil
}

// NOTE: Revisit this func later
func (r *UserRepository) Update(ctx context.Context, user *domain.User) error {
	result, err := r.db.ExecContext(
		ctx,
		`
			UPDATE users
			SET
				username = ?,
				password_hash = ?,
				recovery_hash = ?,
				session_version = ?,
				updated_at = CURRENT_TIMESTAMP
			WHERE id = ?
		`,
		user.Username,
		user.PasswordHash,
		user.RecoveryHash,
		user.SessionVersion,
		user.ID,
	)

	if err != nil {
		return fmt.Errorf("update user: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get affected rows: %w", err)
	}

	if rows == 0 {
		return domain.ErrUserNotFound
	}

	updated, err := r.ByID(ctx, user.ID)
	if err != nil {
		return fmt.Errorf("reload updated user: %w", err)
	}

	*user = *updated

	return nil
}

func (r *UserRepository) Delete(ctx context.Context, id int64) error {
	result, err := r.db.ExecContext(
		ctx,
		`
			DELETE FROM users
			WHERE id = ?	
		`,
		id,
	)

	if err != nil {
		return fmt.Errorf("delete user: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get affected rows: %w", err)
	}

	if rows == 0 {
		return domain.ErrUserNotFound
	}

	return nil
}

func (r *UserRepository) ByUsername(ctx context.Context, username string) (*domain.User, error) {
	row := r.db.QueryRowContext(
		ctx,
		`
			SELECT
				id,
				username,
				password_hash,
				recovery_hash,
				session_version,
				created_at,
				updated_at
			FROM users
			WHERE username = ?
		`,
		username,
	)

	user, err := scanUser(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}

		return nil, fmt.Errorf("get user by username: %w", err)
	}

	return user, nil
}

func (r *UserRepository) ByID(ctx context.Context, id int64) (*domain.User, error) {
	row := r.db.QueryRowContext(
		ctx,
		`
			SELECT
				id,
				username,
				password_hash,
				recovery_hash,
				session_version,
				created_at,
				updated_at
			FROM users
			WHERE id = ?	
		`,
		id,
	)

	user, err := scanUser(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}

		return nil, fmt.Errorf("get user by id: %w", err)
	}

	return user, nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanUser(row scanner) (*domain.User, error) {
	var (
		user      domain.User
		createdAt string
		updatedAt string
	)

	err := row.Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.RecoveryHash,
		&user.SessionVersion,
		&createdAt,
		&updatedAt,
	)

	if err != nil {
		return nil, err
	}

	user.CreatedAt, err = parseTimestamp(createdAt)
	if err != nil {
		return nil, fmt.Errorf("parse created_at: %w", err)
	}

	user.UpdatedAt, err = parseTimestamp(updatedAt)
	if err != nil {
		return nil, fmt.Errorf("parse updated_at: %w", err)
	}

	return &user, nil
}

func parseTimestamp(value string) (time.Time, error) {
	return time.Parse("2006-01-02 15:04:05", value)
}
