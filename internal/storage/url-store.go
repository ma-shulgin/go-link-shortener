package storage
type URLStore interface {
    AddURL(originalURL, shortURL string) error
		AddURLBatch(urls []URLRecord) error
    GetURL(shortURL string) (string, bool)
		Ping() error
		Close() error
}

type URLRecord struct {
	UUID        int    `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}