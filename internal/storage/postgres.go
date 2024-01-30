package storage

import (
	"database/sql"
	"errors"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/ma-shulgin/go-link-shortener/internal/logger"
)

type PostgresStore struct {
	db *sql.DB
}

func InitPostgresStore(dsn string) (*PostgresStore, error) {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		logger.Log.Fatal("Failed to connect to database: ", err)
	}
	tx, err := db.Begin()

	if err != nil {
		return nil, err
	}

	tx.Exec(`CREATE TABLE IF NOT EXISTS urls (
        id SERIAL PRIMARY KEY,
        original_url TEXT NOT NULL,
        short_url TEXT NOT NULL UNIQUE
    );`)

	tx.Exec(`CREATE INDEX IF NOT EXISTS short_url_idx ON urls (short_url)`)

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		logger.Log.Fatalln("Failed to create the database:", err)
	}

	logger.Log.Info("Database initalized successfully")
	s := &PostgresStore{db: db}
	return s, nil
}

func (s *PostgresStore) AddURL(originalURL, shortURL string) error {
	_, err := s.db.Exec("INSERT INTO urls (original_url, short_url) VALUES ($1, $2)", originalURL, shortURL)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
			err = ErrConflict
		}
	}
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

func (s *PostgresStore) AddURLBatch(urls []URLRecord) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}

	for _, url := range urls {
		if _, err := tx.Exec("INSERT INTO urls (original_url, short_url) VALUES ($1, $2)", url.OriginalURL, url.ShortURL); err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}
