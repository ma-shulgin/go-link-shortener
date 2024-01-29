package storage

import (
    "database/sql"
    "github.com/ma-shulgin/go-link-shortener/internal/logger"
    _ "github.com/jackc/pgx/v5/stdlib"
)

type PostgresStore struct {
    db *sql.DB
}


func InitPostgresStore(dsn string) (*PostgresStore, error) {
    db, err := sql.Open("pgx", dsn)
    if err != nil {
        logger.Log.Fatal("Failed to connect to database: ", err)
    }
    _, err = db.Exec(`CREATE TABLE IF NOT EXISTS urls (
        id SERIAL PRIMARY KEY,
        original_url TEXT NOT NULL,
        short_url TEXT NOT NULL UNIQUE
    );`)
    if err != nil {
        logger.Log.Fatalln("Failed to ping the database:", err)
    }
    logger.Log.Info("Database initalized successfully")
    s := &PostgresStore{db:db}
    return s, nil
}

func (s *PostgresStore) AddURL(originalURL, shortURL string) error {
    _, err := s.db.Exec("INSERT INTO urls (original_url, short_url) VALUES ($1, $2)", originalURL, shortURL)
    return err
}

func (s *PostgresStore) GetURL(shortURL string) (string, bool) {
    var originalURL string
    err := s.db.QueryRow("SELECT original_url FROM urls WHERE short_url = $1", shortURL).Scan(&originalURL)
    if err != nil {
        return "", false
    }
    return originalURL, true
}

func (s *PostgresStore) Ping() error {
    return s.db.Ping()
}

func (s *PostgresStore) Close() error {
    return s.db.Close()
}