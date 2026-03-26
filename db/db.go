package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

type Database struct {
	db *sql.DB
}

func InitDB() (*Database, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("could not determine user home dir: %w", err)
	}

	appDir := filepath.Join(homeDir, ".paprawatch")
	return InitDBAtPath(appDir)
}

func InitDBAtPath(dbDir string) (*Database, error) {
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return nil, fmt.Errorf("could not create directory %s: %w", dbDir, err)
	}

	dbPath := filepath.Join(dbDir, "paprawatch.db")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("could not open database: %w", err)
	}

	createTableQuery := `
	CREATE TABLE IF NOT EXISTS scanned_files (
		hash TEXT PRIMARY KEY,
		path TEXT,
		scanned_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	`
	if _, err := db.Exec(createTableQuery); err != nil {
		db.Close()
		return nil, fmt.Errorf("could not create table: %w", err)
	}

	return &Database{db: db}, nil
}

func (d *Database) Close() error {
	return d.db.Close()
}

func (d *Database) HasFile(hash string) (bool, error) {
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM scanned_files WHERE hash = ?)"
	err := d.db.QueryRow(query, hash).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("could not query database: %w", err)
	}
	return exists, nil
}

func (d *Database) RecordFile(hash, path string) error {
	query := "INSERT INTO scanned_files (hash, path) VALUES (?, ?)"
	_, err := d.db.Exec(query, hash, path)
	if err != nil {
		return fmt.Errorf("could not insert record: %w", err)
	}
	return nil
}
