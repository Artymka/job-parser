package postgres

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

type Storage struct {
	db                    *sql.DB
	requestTimeoutSeconds time.Duration
}

func (s *Storage) Close() error {
	return s.db.Close()
}

func NewStorage(dbname, user, password string) (*Storage, error) {
	const op = "postgres_new_storage"

	connStr := fmt.Sprintf("dbname=%s user=%s password=%s sslmode=disable", dbname, user, password)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	storage := Storage{
		db:                    db,
		requestTimeoutSeconds: 5,
	}

	return &storage, nil
}
