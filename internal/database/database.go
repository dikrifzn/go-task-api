package database

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

func Open(ctx context.Context) (*sql.DB, error) {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "data/tasks.db"
	}

	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return nil, fmt.Errorf("membuat folder database: %w", err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("membuka database: %w", err)
	}

	db.SetMaxOpenConns(1)

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("menghubungkan database: %w", err)
	}

	if err := migrate(ctx, db); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

func migrate(ctx context.Context, db *sql.DB) error {
	schema := `
	CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		email TEXT NOT NULL UNIQUE,
		password_hash TEXT NOT NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS tasks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		title TEXT NOT NULL,
		description TEXT NOT NULL DEFAULT '',
		completed BOOLEAN NOT NULL DEFAULT FALSE,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);
	
	CREATE TABLE IF NOT EXISTS sessions (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		user_id INTEGER NOT NULL,
		token_hash TEXT NOT NULL UNIQUE,
		expires_at DATETIME NOT NULL,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);
	`

	if _, err := db.ExecContext(ctx, "PRAGMA foreign_keys = ON"); err != nil {
		return fmt.Errorf("mengaktifkan foreign keys: %w", err)
	}

	if _, err := db.ExecContext(ctx, schema); err != nil {
		return fmt.Errorf("membuat tabel database: %w", err)
	}

	return nil
}
