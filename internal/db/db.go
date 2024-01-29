package db

import (
    "database/sql"
    "github.com/ma-shulgin/go-link-shortener/internal/logger"
    _ "github.com/jackc/pgx/v5/stdlib"
)

func ConnectToDB(dsn string) *sql.DB {
    db, err := sql.Open("pgx", dsn)
    if err != nil {
        logger.Log.Fatal("Failed to connect to database: ", err)
    }
    err = db.Ping()
    if err != nil {
        logger.Log.Fatal("Failed to ping the database: ", err)
    }
    logger.Log.Info("Connected to the database successfully")
    return db
}
