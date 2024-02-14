package storage

import (
	"context"
	"errors"
)

type URLStore interface {
	AddURL(ctx context.Context, originalURL, shortURL string) error
	AddURLBatch(ctx context.Context, urls []URLRecord) error
	DeleteURLs(ctx context.Context, shortURLs []string) error
	GetURL(ctx context.Context, shortURL string) (string, error)
	GetUserURLs(ctx context.Context) ([]URLRecord, error)
	Ping(ctx context.Context) error
	Close() error
}

var ErrConflict = errors.New("data conflict")
var ErrNoUserID = errors.New("userID not found in context")
var ErrDeleted = errors.New("url was deleted")
var ErrNotFound = errors.New("record not found")

type URLRecord struct {
	UUID        int    `json:"uuid,omitempty"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
	CreatorID   string `json:"creator_id,omitempty"`
	DeletedFlag bool   `json:"is_deleted,omitempty"`
}
