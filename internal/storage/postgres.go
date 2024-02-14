package storage

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/ma-shulgin/go-link-shortener/internal/appcontext"
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
        original_url TEXT NOT NULL UNIQUE,
        short_url TEXT NOT NULL,
				creator_id TEXT NOT NULL,
				is_deleted BOOLEAN DEFAULT FALSE
    );`)

	tx.Exec(`CREATE INDEX IF NOT EXISTS short_url_idx ON urls (short_url)`)
	tx.Exec(`CREATE INDEX IF NOT EXISTS creator_id_idx ON urls (creator_id)`)

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		logger.Log.Fatalln("Failed to create the database:", err)
	}

	logger.Log.Info("Database initalized successfully")
	s := &PostgresStore{db: db}
	return s, nil
}

func (s *PostgresStore) AddURL(ctx context.Context, originalURL, shortURL string) error {
	userID, ok := ctx.Value(appcontext.KeyUserID).(string)
	if !ok {
		return ErrNoUserID
	}

	_, err := s.db.ExecContext(ctx, "INSERT INTO urls (original_url, short_url, creator_id) VALUES ($1, $2, $3)", originalURL, shortURL, userID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgerrcode.IsIntegrityConstraintViolation(pgErr.Code) {
			err = ErrConflict
		}
	}
	return err
}

func (s *PostgresStore) GetURL(ctx context.Context, shortURL string) (string, error) {
	var originalURL string
	err := s.db.QueryRowContext(ctx, "SELECT original_url FROM urls WHERE short_url = $1", shortURL).Scan(&originalURL)
	if err != nil {
		if errors.Is(err,sql.ErrNoRows){
			return "", ErrDeleted
		}
		return "", err
	}
	return originalURL, nil
}

func (s *PostgresStore) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

func (s *PostgresStore) Close() error {
	return s.db.Close()
}

func (s *PostgresStore) AddURLBatch(ctx context.Context, urls []URLRecord) error {
	userID, ok := ctx.Value(appcontext.KeyUserID).(string)
	if !ok {
		return ErrNoUserID
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	for _, url := range urls {
		if _, err := tx.ExecContext(ctx, "INSERT INTO urls (original_url, short_url, creator_id) VALUES ($1, $2, $3)", url.OriginalURL, url.ShortURL, userID); err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

func (s *PostgresStore) GetUserURLs(ctx context.Context) ([]URLRecord, error) {
	userID, ok := ctx.Value(appcontext.KeyUserID).(string)
	if !ok {
		return nil, ErrNoUserID
	}
	rows, err := s.db.Query("SELECT short_url, original_url FROM urls WHERE creator_id = $1", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var urls []URLRecord
	for rows.Next() {
		var url URLRecord
		if err := rows.Scan(&url.ShortURL, &url.OriginalURL); err != nil {
			return nil, err
		}
		urls = append(urls, url)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return urls, nil
}

func (s *PostgresStore) DeleteURLs(ctx context.Context, shortURLIDs []string) error {
	userID, ok := ctx.Value(appcontext.KeyUserID).(string)
	if !ok {
		return ErrNoUserID
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
			return err
	}

	for _, shortURLID := range shortURLIDs {
			_, err := tx.ExecContext(ctx, "UPDATE urls SET is_deleted = TRUE WHERE short_url = $1 AND creator_id = $2", shortURLID, userID)
			if err != nil {
					tx.Rollback()
					return err
			}
	}

	return tx.Commit()
}