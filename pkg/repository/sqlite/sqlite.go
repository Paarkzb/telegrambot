package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"telegrambot/internal/e"
	"telegrambot/pkg/repository"

	_ "modernc.org/sqlite"
)

type RepositorySQLite struct {
	db *sql.DB
}

// New creates new SQLite repository.
func New(path string) (*RepositorySQLite, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, e.Wrap("can't open sqlite db", err)
	}

	if err := db.Ping(); err != nil {
		return nil, e.Wrap("can't ping sqlite db", err)
	}

	return &RepositorySQLite{db: db}, nil
}

// Save saves page to repository.
func (r *RepositorySQLite) Save(ctx context.Context, p *repository.Page) error {
	q := `INSERT INTO pages (url, username) VALUES (?, ?)`

	if _, err := r.db.ExecContext(ctx, q, p.URL, p.Username); err != nil {
		return e.Wrap("can't save page", err)
	}

	return nil
}

// PickRandom pick random page from repository.
func (r *RepositorySQLite) PickRandom(ctx context.Context, username string) (*repository.Page, error) {
	q := `SELECT url FROM pages WHERE username = ? ORDER BY RANDOM() LIMIT 1`

	var url string

	err := r.db.QueryRowContext(ctx, q, username).Scan(&url)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, repository.ErrNoSavedPages
	}
	if err != nil {
		return nil, e.Wrap("can't pick random page", err)
	}

	return &repository.Page{URL: url, Username: username}, nil
}

// Remove removes page from repository.
func (r *RepositorySQLite) Remove(ctx context.Context, p *repository.Page) error {
	q := `DELETE FROM pages WHERE url = ? and username = ?`

	if _, err := r.db.ExecContext(ctx, q, p.URL, p.Username); err != nil {
		return e.Wrap("can't remove page", err)
	}

	return nil
}

// IsExists checks if page exists in repository.
func (r *RepositorySQLite) IsExists(ctx context.Context, p *repository.Page) (bool, error) {
	q := `SELECT COUNT(*) FROM pages WHERE url = ? and username = ?`

	var count int

	if err := r.db.QueryRowContext(ctx, q, p.URL, p.Username).Scan(&count); err != nil {
		return false, e.Wrap("can't check if page exists", err)
	}

	return count > 0, nil
}

func (r *RepositorySQLite) Init(ctx context.Context) error {
	q := `CREATE TABLE IF NOT EXISTS pages (url TEXT, username TEXT); 
CREATE INDEX IF NOT EXISTS pages_username_idx ON pages (username);
CREATE INDEX IF NOT EXISTS pages_url_idx ON pages(url);`

	_, err := r.db.ExecContext(ctx, q)
	if err != nil {
		return e.Wrap("can't create table", err)
	}

	return nil
}
