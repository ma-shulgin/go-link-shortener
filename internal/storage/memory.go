package storage

import (
	"context"
	"errors"

	"github.com/ma-shulgin/go-link-shortener/internal/appcontext"
)

type MemoryStore struct {
	urlMap map[string]URLRecord
	nextID int
}

func InitMemoryStore() *MemoryStore {
	return &MemoryStore{
		urlMap: make(map[string]URLRecord),
		nextID: 1,
	}
}

func (s *MemoryStore) addAndReturnURL(ctx context.Context, originalURL, shortURL string) (*URLRecord, error) {
	if _, exists := s.urlMap[shortURL]; exists {
		return nil,ErrConflict
	}
	
	userID, ok := ctx.Value(appcontext.KeyUserID).(string)
	if !ok {
		return nil, ErrNoUserID
	}

	record := URLRecord{
		ShortURL: shortURL,
		OriginalURL: originalURL,
		CreatorID:   userID,
		DeletedFlag: false,
		UUID: s.nextID,
	}

	s.urlMap[shortURL] = record
	s.nextID++
	return &record, nil
}
func (s *MemoryStore) AddURL(ctx context.Context, originalURL, shortURL string) error {
	_, err := s.addAndReturnURL(ctx,originalURL,shortURL)
	return err
}

func (s *MemoryStore) GetURL(ctx context.Context, shortURL string) (string, error) {
	record, ok := s.urlMap[shortURL]
	if !ok {
		return "", errors.New("not found")
	}
	if record.DeletedFlag {
		return "", ErrDeleted
	}
	return record.OriginalURL, nil
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

func (s *MemoryStore) DeleteURLs(ctx context.Context, shortURLs []string) error {
	userID, ok := ctx.Value(appcontext.KeyUserID).(string)
	if !ok {
		return ErrNoUserID
	}
	for _, shortURL := range shortURLs {
			if url, ok := s.urlMap[shortURL]; ok && url.CreatorID == userID {
					url.DeletedFlag = true
					s.urlMap[shortURL] = url 
			}
	}
	return nil
}