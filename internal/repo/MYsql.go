package repo

import (
	"database/sql"
	"errors"
	"fmt"
	"friend-help/internal/errs"
	"os"

	_ "github.com/go-sql-driver/mysql"
)

func ConnectToBase() (*sql.DB, error) {
	connStr := os.Getenv("DB_DSN")
	if connStr == "" {
		return nil, errors.New("var DB_DSN not found")
	}
	db, err := sql.Open("mysql", connStr)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errs.ErrFailedToOpenDB, err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("%w: %w", errs.ErrFailedToPingDB, err)
	}
	return db, nil
}

func RunMigrations(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id INT PRIMARY KEY AUTO_INCREMENT,
			login VARCHAR(32) NOT NULL UNIQUE,
			username VARCHAR(32) NOT NULL,
			email VARCHAR(100) NULL UNIQUE,
			password_hash TEXT NOT NULL,
			is_activated BOOLEAN NOT NULL DEFAULT true,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
	`)
	if err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}
	return nil
}
