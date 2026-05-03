package sqlite

import (
	"database/sql"
	"errors"
	"fmt"
	"url-shortener/internal/storage"

	"github.com/mattn/go-sqlite3"
)

type Storage struct {
	db *sql.DB
}

func New(storagePath string) (*Storage, error) {
	const op = "storage.sqlite.NewStorage"

	db, err := sql.Open("sqlite3", storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// Create table if not exists
	stmt, err := db.Prepare(`
	CREATE TABLE IF NOT EXISTS url (
	    id INTEGER PRIMARY KEY,
	    alias TEXT NOT NULL UNIQUE,
	    url TEXT NOT NULL);
	CREATE INDEX IF NOT EXISTS idx_alias ON url(alias);
	`)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	_, err = stmt.Exec()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) SaveURL(urlToSave string, alias string) (int64, error) {
	const op = "storage.sqlite.SaveUrl"

	// Prepare stmt
	stmt, err := s.db.Prepare("INSERT INTO url (alias, url) VALUES (?, ?)")
	if err != nil {
		return 0, fmt.Errorf("%s prepare statement: %w", op, err)
	}

	// Executing
	res, err := stmt.Exec(alias, urlToSave)
	if err != nil {
		if sqliteErr, ok := err.(sqlite3.Error); ok && sqliteErr.ExtendedCode == sqlite3.ErrConstraintUnique {
			return 0, fmt.Errorf("%s: %w", op, storage.ErrURLExists)
		}
	}

	// getting id
	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s failed to get last insert id: %w", op, err)
	}

	// returnin ID
	return id, nil
}

func (s *Storage) GetURL(alias string) (string, error) {
	const op = "storage.sqlite.GetURL"

	stmt, err := s.db.Prepare("SELECT url FROM url WHERE alias = ?")
	if err != nil {
		return "", fmt.Errorf("%s prepare statement: %w", op, err)
	}

	var resURL string

	err = stmt.QueryRow(alias).Scan(&resURL)
	if errors.Is(err, sql.ErrNoRows) {
		return "", storage.ErrURLNotFound
	}
	if err != nil {
		return "", fmt.Errorf("%s execute statement: %w", op, err)
	}
	return resURL, nil
}

func (s *Storage) DeleteURL(alias string) (int64, error) {
	const op = "storage.sqlite.DeleteUrl"

	// prepare statement
	stmt, err := s.db.Prepare("DELETE FROM url WHERE alias = ?")
	if err != nil {
		return 0, fmt.Errorf("%s prepare statement: %w", op, err)
	}
	res, err := stmt.Exec(alias)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, storage.ErrURLNotFound
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s failed to get last insert id: %w", op, err)
	}
	// returnin ID
	return id, nil
}
