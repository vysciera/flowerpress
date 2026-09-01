package turso

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"flowerpress/internal/domain"
)

type SessionRepository struct {
	db *sql.DB
}

var _ domain.SessionRepository = (*SessionRepository)(nil)

func NewSessionRepository(db *sql.DB) *SessionRepository {
	return &SessionRepository{
		db: db,
	}
}

// New Session
func (r *SessionRepository) Create(ctx context.Context, session *domain.Session) error {
	result, err := r.db.ExecContext(
		ctx,
		`
			INSERT INTO sessions (
				user_id,
				token_hash,
				session_version,
				expires_at
			)	
			VALUES (?, ?, ?, ?)
		`,
		session.UserID,
		session.TokenHash,
		session.SessionVersion,
		formatTimestamp(session.ExpiresAt),
	)

	if err != nil {
		return fmt.Errorf("create session: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("get session id: %w", err)
	}

	session.ID = id

	created, err := r.ByTokenHash(ctx, session.TokenHash)
	if err != nil {
		return fmt.Errorf("reload created session: %w", err)
	}

	*session = *created

	return nil
}

// Retrieve Session
func (r *SessionRepository) ByTokenHash(ctx context.Context, tokenHash string) (*domain.Session, error) {
	row := r.db.QueryRowContext(
		ctx,
		`
			SELECT
				id,
				user_id,
				token_hash,
				session_version,
				expires_at,
				created_at
			FROM sessions
			WHERE token_hash = ?	
		`,
		tokenHash,
	)

	session, err := scanSession(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrSessionNotFound
		}

		return nil, fmt.Errorf("get session by token hash: %w", err)
	}

	return session, nil
}

// Delete Session
func (r *SessionRepository) Delete(ctx context.Context, id int64) error {
	result, err := r.db.ExecContext(
		ctx,
		`
			DELETE FROM sessions
			WHERE id = ?	
		`,
		id,
	)

	if err != nil {
		return fmt.Errorf("delete session: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get affected rows: %w", err)
	}

	if rows == 0 {
		return domain.ErrSessionNotFound
	}

	return nil
}

// Delete all sessions
func (r *SessionRepository) DeleteByUserID(ctx context.Context, userID int64) error {
	_, err := r.db.ExecContext(
		ctx,
		`
			DELETE FROM sessions
			WHERE user_id = ?	
		`,
		userID,
	)

	if err != nil {
		return fmt.Errorf("delete sessions by user id: %w", err)
	}

	return nil
}

// Delete expired session
func (r *SessionRepository) DeleteExpired(ctx context.Context, before time.Time) error {
	_, err := r.db.ExecContext(
		ctx,
		`
			DELETE FROM sessions
			WHERE expires_at <= ?	
		`,
		formatTimestamp(before),
	)

	if err != nil {
		return fmt.Errorf("delete expired sessions: %w", err)
	}

	return nil
}

func scanSession(row scanner) (*domain.Session, error) {
	var (
		session   domain.Session
		expiresAt string
		createdAt string
	)

	err := row.Scan(
		&session.ID,
		&session.UserID,
		&session.TokenHash,
		&session.SessionVersion,
		&expiresAt,
		&createdAt,
	)

	if err != nil {
		return nil, err
	}

	session.ExpiresAt, err = parseTimestamp(expiresAt)
	if err != nil {
		return nil, fmt.Errorf("parse expires_at: %w", err)
	}

	session.CreatedAt, err = parseTimestamp(createdAt)
	if err != nil {
		return nil, fmt.Errorf("parse created_at: %w", err)
	}

	return &session, nil
}
