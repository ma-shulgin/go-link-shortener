package storage
 
import (
	"errors"
	"context"
)

type URLStore interface {
	AddURL(ctx context.Context, originalURL, shortURL string) error
	AddURLBatch(ctx context.Context, urls []URLRecord) error
	GetURL(ctx context.Context, shortURL string) (string, bool)
	Ping(ctx context.Context) error
	Close() error
}

var ErrConflict = errors.New("data conflict")

type URLRecord struct {
	UUID        int    `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}
