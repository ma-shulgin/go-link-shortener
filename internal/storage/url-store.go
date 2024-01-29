package storage
type URLStore interface {
    AddURL(originalURL, shortURL string) error
    GetURL(shortURL string) (string, bool)
		Ping() error
		Close() error
}