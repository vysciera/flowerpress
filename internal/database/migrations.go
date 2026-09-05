package database

import (
	"database/sql"
	"fmt"
)

type Migration struct {
	Version int
	Name    string
	SQL     string
}

// God I love SQL.
var migrations = []Migration{
	{
		Version: 1,
		Name:    "create users",
		SQL: `
			CREATE TABLE users (
				id INTEGER PRIMARY KEY,
				username TEXT NOT NULL COLLATE NOCASE UNIQUE,
				password_hash TEXT NOT NULL,
				recovery_hash TEXT,
				session_version INTEGER NOT NULL DEFAULT 0,
				created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
				updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
			);
		`,
	},
	{
		Version: 2,
		Name:    "create sessions",
		SQL: `
			CREATE TABLE sessions (
				id INTEGER PRIMARY KEY,
				user_id INTEGER NOT NULL,
				token_hash TEXT NOT NULL UNIQUE,
				session_version INTEGER NOT NULL,
				expires_at TEXT NOT NULL,
				created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,

				FOREIGN KEY (user_id)
					REFERENCES users(id)
					ON DELETE CASCADE
			);

			CREATE INDEX idx_sessions_user_id
				ON sessions(user_id);

			CREATE INDEX idx_sessions_expires_at
				ON sessions(expires_at);
		`,
	},
	{
		Version: 3,
		Name:    "create projects",
		SQL: `
			CREATE TABLE projects (
				id INTEGER PRIMARY KEY,

				title TEXT NOT NULL,
				slug TEXT NOT NULL UNIQUE,
				description TEXT NOT NULL DEFAULT '',
				status TEXT NOT NULL DEFAULT 'draft',

				published_at TEXT,
				created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
				updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,

				CHECK (
					status IN (
						'draft',
						'published',
						'unlisted',
						'archived'
					)
				)
			);

			CREATE INDEX idx_projects_status
				ON projects(status);

			CREATE INDEX idx_projects_created_at
				ON projects(created_at);
		`,
	},
	{
		Version: 4,
		Name:    "create media",
		SQL: `
		CREATE TABLE media_assets (
			id INTEGER PRIMARY KEY,

			storage_key TEXT NOT NULL UNIQUE,
			original_name TEXT NOT NULL,
			mime_type TEXT NOT NULL,
			size_bytes INTEGER NOT NULL,
			sha256 TEXT NOT NULL UNIQUE,

			width INTEGER,
			height INTEGER,

			created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,

			CHECK (size_bytes >= 0),
			CHECK (width IS NULL OR width > 0),
			CHECK (height IS NULL OR height > 0)
		);

		CREATE TABLE media_placements (
			id INTEGER PRIMARY KEY,

			asset_id INTEGER NOT NULL,
			project_id INTEGER NOT NULL,

			role TEXT NOT NULL,
			position INTEGER NOT NULL DEFAULT 0,

			caption TEXT NOT NULL DEFAULT '',
			alt_text TEXT NOT NULL DEFAULT '',

			created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,

			FOREIGN KEY (asset_id)
				REFERENCES media_assets(id)
				ON DELETE CASCADE,

			FOREIGN KEY (project_id)
				REFERENCES projects(id)
				ON DELETE CASCADE,

			CHECK (
				role IN (
					'thumbnail',
					'content',
					'attachment'
				)
			),

			CHECK (position >= 0)
		);

		CREATE INDEX idx_media_placements_asset_id
			ON media_placements(asset_id);

		CREATE INDEX idx_media_placements_project_id
			ON media_placements(project_id);

		CREATE INDEX idx_media_placements_project_position
			ON media_placements(project_id, position);

		CREATE UNIQUE INDEX idx_media_placements_project_thumbnail
			ON media_placements(project_id)
			WHERE role = 'thumbnail';
	`,
	},
}

func Migrate(db *sql.DB) error {
	if err := ensureMigrationsTable(db); err != nil {
		return fmt.Errorf("initialize migrations table: %w", err)
	}

	for _, migration := range migrations {
		applied, err := migrationApplied(db, migration.Version)
		if err != nil {
			return fmt.Errorf("check migration %d: %w", migration.Version, err)
		}

		if applied {
			continue
		}

		if err := applyMigration(db, migration); err != nil {
			return fmt.Errorf("apply migration %d (%s): %w", migration.Version, migration.Name, err)
		}
	}

	return nil
}

func ensureMigrationsTable(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			applied_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
			)
	`)

	return err
}

func migrationApplied(db *sql.DB, version int) (bool, error) {
	var exists int

	err := db.QueryRow(`
		SELECT COUNT(*)
		FROM schema_migrations
		WHERE version = ?
	`, version).Scan(&exists)

	if err != nil {
		return false, err
	}

	return exists > 0, nil
}

func applyMigration(db *sql.DB, migration Migration) error {
	tx, err := db.Begin()

	if err != nil {
		return err
	}

	defer func() {
		_ = tx.Rollback()
	}()

	if _, err := tx.Exec(migration.SQL); err != nil {
		return err
	}

	if _, err := tx.Exec(`
		INSERT INTO schema_migrations (version, name)
		VALUES (?, ?)
	`, migration.Version, migration.Name); err != nil {
		return err
	}

	return tx.Commit()
}
