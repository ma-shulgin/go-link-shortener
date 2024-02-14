package storage

import (
	"bufio"
	"context"
	"encoding/json"
	"os"

	"github.com/ma-shulgin/go-link-shortener/internal/appcontext"
	"github.com/ma-shulgin/go-link-shortener/internal/logger"
)

type FileStore struct {
	file   *os.File
	urlMap map[string]string
	nextID int
}

func InitFileStore(filePath string) (*FileStore, error) {
	store := &FileStore{
		urlMap: make(map[string]string),
		nextID: 1,
	}

	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	maxID := 0
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var record URLRecord
		if err := json.Unmarshal(scanner.Bytes(), &record); err != nil {
			return nil, err
		}

		if record.UUID > maxID {
			maxID = record.UUID
		}
		store.urlMap[record.ShortURL] = record.OriginalURL
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	store.nextID = maxID + 1
	store.file = file

	return store, nil
}

func (s *FileStore) AddURL(ctx context.Context, originalURL, shortURL string) error {
	if _, exists := s.urlMap[shortURL]; exists {
		return ErrConflict
	}

	userID, ok := ctx.Value(appcontext.KeyUserID).(string)
	if !ok {
		return ErrNoUserID
	}

	s.urlMap[shortURL] = originalURL

	record := URLRecord{
		UUID:        s.nextID,
		ShortURL:    shortURL,
		OriginalURL: originalURL,
		CreatorID:   userID,
	}

	data, err := json.Marshal(record)
	if err != nil {
		logger.Log.Errorf("error marshaling JSON: %w", err)
		return err
	}

	if _, err := s.file.Write(append(data, '\n')); err != nil {
		logger.Log.Errorf("error writing to file: %w", err)
		return err
	}

	s.nextID++
	return nil
}

func (s *FileStore) GetURL(ctx context.Context, shortURL string) (string, bool) {
	url, ok := s.urlMap[shortURL]
	return url, ok
}

func (s *FileStore) Close() error {
	if s.file != nil {
		return s.file.Close()
	}
	return nil
}

func (s *FileStore) Ping(ctx context.Context) error {
	_, err := s.file.Stat()
	return err
}

func (s *FileStore) AddURLBatch(ctx context.Context, urls []URLRecord) error {
	for _, url := range urls {
		if err := s.AddURL(ctx, url.OriginalURL, url.ShortURL); err != nil {
			return err
		}
	}
	return nil
}

func (s *FileStore) GetUserURLs(ctx context.Context) ([]URLRecord, error) {
	userID, ok := ctx.Value(appcontext.KeyUserID).(string)
	if !ok {
		return nil, ErrNoUserID
	}

	var urls []URLRecord
	scanner := bufio.NewScanner(s.file)
	for scanner.Scan() {
		var url URLRecord
		if err := json.Unmarshal(scanner.Bytes(), &url); err == nil && url.CreatorID == userID {
			urls = append(urls, URLRecord{ShortURL: url.ShortURL, OriginalURL: url.OriginalURL})
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return urls, nil
}
