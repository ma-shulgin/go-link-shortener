package storage

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"os"

	"github.com/ma-shulgin/go-link-shortener/internal/appcontext"
	"github.com/ma-shulgin/go-link-shortener/internal/logger"
)

type FileStore struct {
	file        *os.File
	memoryStore *MemoryStore
}

func InitFileStore(filePath string) (*FileStore, error) {
	store := &FileStore{
		memoryStore: InitMemoryStore(),
	}

	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var record URLRecord
		if err := json.Unmarshal(scanner.Bytes(), &record); err != nil {
			return nil, err
		}
		store.memoryStore.AddURL(context.WithValue(context.Background(), appcontext.KeyUserID, record.CreatorID), record.OriginalURL, record.ShortURL)

	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	store.file = file

	return store, nil
}

func (s *FileStore) AddURL(ctx context.Context, originalURL, shortURL string) error {
	record, err := s.memoryStore.addAndReturnURL(ctx, originalURL, shortURL)
	if err != nil {
		return err
	}
	return s.appendRecord(record)
}

func (s *FileStore) DeleteURLs(ctx context.Context, shortURLs []string) error {
	if err := s.memoryStore.DeleteURLs(ctx, shortURLs); err != nil {
		return err
	}
	if err := s.file.Truncate(0); err != nil {
		return err
	}
	if _, err := s.file.Seek(0, io.SeekStart); err != nil {
		return err
	}

	for _, record := range s.memoryStore.urlMap {
		if err := s.appendRecord(&record); err != nil {
			return err
		}
	}
	return nil
}

func (s *FileStore) appendRecord(record *URLRecord) error {
	data, err := json.Marshal(*record)
	if err != nil {
		logger.Log.Errorf("error marshaling JSON: %w", err)
		return err
	}

	if _, err := s.file.Write(append(data, '\n')); err != nil {
		logger.Log.Errorf("error writing to file: %w", err)
		return err
	}

	return nil
}

func (s *FileStore) GetURL(ctx context.Context, shortURL string) (string, error) {
	return s.memoryStore.GetURL(ctx, shortURL)
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
	return s.memoryStore.GetUserURLs(ctx)
}
