package storage

type MemoryStore struct {
	urlMap map[string]string
}

func InitMemoryStore() *MemoryStore {
	return &MemoryStore{
		urlMap: make(map[string]string),
	}
}

func (s *MemoryStore) AddURL(originalURL, shortURL string) error {
	s.urlMap[shortURL] = originalURL
	return nil
}

func (s *MemoryStore) GetURL(shortURL string) (string, bool) {
	originalURL, exists := s.urlMap[shortURL]
	return originalURL, exists
}

func (s *MemoryStore) Ping() error {
	return nil
}

func (s *MemoryStore) Close() error {
	return nil
}

func (s *MemoryStore) AddURLBatch(urls []URLRecord) error {
	for _, url := range urls {
		s.urlMap[url.ShortURL] = url.OriginalURL
	}
	return nil
}
