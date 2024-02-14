package storage

import (
	"context"

	"github.com/ma-shulgin/go-link-shortener/internal/appcontext"
)

type MemoryStore struct {
	urlMap map[string]URLRecord
}

func InitMemoryStore() *MemoryStore {
	return &MemoryStore{
		urlMap: make(map[string]URLRecord),
	}
}

func (s *MemoryStore) AddURL(ctx context.Context, originalURL, shortURL string) error {
	if _, exists := s.urlMap[shortURL]; exists {
		return ErrConflict
	}

	userID, ok := ctx.Value(appcontext.KeyUserID).(string)
	if !ok {
		return ErrNoUserID
	}

	record := URLRecord{
		OriginalURL: originalURL,
		CreatorID:   userID,
	}

	s.urlMap[shortURL] = record
	return nil
}

func (s *MemoryStore) GetURL(ctx context.Context, shortURL string) (string, bool) {
	record, exists := s.urlMap[shortURL]
	return record.OriginalURL, exists
}

func (s *MemoryStore) Ping(ctx context.Context) error {
	return nil
}

func (s *MemoryStore) Close() error {
	return nil
}

func (s *MemoryStore) AddURLBatch(ctx context.Context, urls []URLRecord) error {
	for _, url := range urls {
		if err := s.AddURL(ctx, url.OriginalURL, url.ShortURL); err != nil {
			return err
		}
	}
	return nil
}

func (s *MemoryStore) GetUserURLs(ctx context.Context) ([]URLRecord, error) {
	userID, ok := ctx.Value(appcontext.KeyUserID).(string)
	if !ok {
		return nil, ErrNoUserID
	}

	var urls []URLRecord
	for _, url := range s.urlMap {
		if url.CreatorID == userID {
			urls = append(urls, URLRecord{ShortURL: url.ShortURL, OriginalURL: url.OriginalURL})
		}
	}
	return urls, nil
}
