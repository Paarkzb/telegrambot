package repository

import (
	"context"
	"crypto/sha1"
	"errors"
	"fmt"
	"telegrambot/internal/e"
)

var ErrNoSavedPages = errors.New("no saved pages")

type Repository interface {
	Save(ctx context.Context, p *Page) error
	PickRandom(ctx context.Context, username string) (*Page, error)
	Remove(ctx context.Context, p *Page) error
	IsExists(ctx context.Context, p *Page) (bool, error)
}

type Page struct {
	URL      string
	Username string
}

func (p *Page) Hash() (string, error) {
	h := sha1.New()

	if _, err := h.Write([]byte(p.URL)); err != nil {
		return "", e.Wrap("can't calculate hash", err)
	}

	if _, err := h.Write([]byte(p.Username)); err != nil {
		return "", e.Wrap("can't calculate hash", err)
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
