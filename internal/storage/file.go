package storage

import (
	"bufio"
	"encoding/json"
	"os"

	"github.com/ma-shulgin/go-link-shortener/internal/logger"
)

type URLRecord struct {
	UUID        int    `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

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

func (s *FileStore) AddURL(originalURL, shortURL string) error {
	if _, exists := s.urlMap[shortURL]; exists {
		logger.Log.Warnf("short URL already exists: %s", shortURL)
		return nil
	}

	s.urlMap[shortURL] = originalURL

	record := URLRecord{
		UUID:        s.nextID,
		ShortURL:    shortURL,
		OriginalURL: originalURL,
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

func (s *FileStore) GetURL(shortURL string) (string, bool) {
	url, ok := s.urlMap[shortURL]
	return url, ok
}

func (s *FileStore) Close() error {
	if s.file != nil {
		return s.file.Close()
	}
	return nil
}

func (s *FileStore) Ping() error {
	_, err := s.file.Stat()
	return err
}
