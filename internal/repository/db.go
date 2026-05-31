package repository

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

// InitDB opens the SQLite database and runs migrations.
func InitDB(dbPath string) (*sql.DB, error) {
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("create db directory: %w", err)
	}

	db, err := sql.Open("sqlite3", dbPath+"?_journal_mode=WAL&_foreign_keys=on&_busy_timeout=5000")
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	db.SetMaxOpenConns(1) // SQLite serializes writes
	db.SetMaxIdleConns(1)

	if err := runMigrations(db); err != nil {
		db.Close()
		return nil, fmt.Errorf("run migrations: %w", err)
	}

	return db, nil
}

func runMigrations(db *sql.DB) error {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS assets (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			dir_path TEXT NOT NULL,
			match_status TEXT NOT NULL DEFAULT 'orphan',
			rating INTEGER NOT NULL DEFAULT 0,
			color_label TEXT NOT NULL DEFAULT '',
			ai_status TEXT NOT NULL DEFAULT '',
			captured_at DATETIME,
			grid_thumb TEXT NOT NULL DEFAULT '',
			full_thumb TEXT NOT NULL DEFAULT '',
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS media_files (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			asset_id INTEGER NOT NULL,
			file_path TEXT NOT NULL UNIQUE,
			file_name TEXT NOT NULL,
			file_size INTEGER NOT NULL DEFAULT 0,
			media_type TEXT NOT NULL,
			camera_model TEXT NOT NULL DEFAULT '',
			lens_model TEXT NOT NULL DEFAULT '',
			focal_length REAL NOT NULL DEFAULT 0,
			aperture REAL NOT NULL DEFAULT 0,
			shutter_speed TEXT NOT NULL DEFAULT '',
			iso INTEGER NOT NULL DEFAULT 0,
			captured_at DATETIME,
			width INTEGER NOT NULL DEFAULT 0,
			height INTEGER NOT NULL DEFAULT 0,
			orientation INTEGER NOT NULL DEFAULT 1,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (asset_id) REFERENCES assets(id) ON DELETE CASCADE
		)`,
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_assets_dir_name ON assets(dir_path, name)`,
		`CREATE INDEX IF NOT EXISTS idx_assets_rating ON assets(rating)`,
		`CREATE INDEX IF NOT EXISTS idx_assets_captured_at ON assets(captured_at)`,
		`CREATE INDEX IF NOT EXISTS idx_media_files_asset_id ON media_files(asset_id)`,
		`CREATE INDEX IF NOT EXISTS idx_media_files_captured_at ON media_files(captured_at)`,
	}

	for _, m := range migrations {
		if _, err := db.Exec(m); err != nil {
			return fmt.Errorf("migration failed: %w\nSQL: %s", err, m)
		}
	}

	// Soft-delete migration: add deleted_at column if not present
	var hasDeletedAt bool
	if err := db.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('assets') WHERE name='deleted_at'`).Scan(&hasDeletedAt); err == nil && !hasDeletedAt {
		if _, err := db.Exec(`ALTER TABLE assets ADD COLUMN deleted_at DATETIME`); err != nil {
			return fmt.Errorf("add deleted_at column: %w", err)
		}
	}
	if _, err := db.Exec(`CREATE INDEX IF NOT EXISTS idx_assets_deleted_at ON assets(deleted_at)`); err != nil {
		return fmt.Errorf("create deleted_at index: %w", err)
	}

	return nil
}
