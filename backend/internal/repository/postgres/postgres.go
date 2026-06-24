package postgres

import (
	"context"
	"database/sql"

	"messenger/backend/internal/repository"
)

// Store provides PostgreSQL-backed persistence.
type Store struct {
	db *sql.DB
}

// NewStore creates a PostgreSQL repository adapter.
func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

// Ping verifies that the database connection is alive.
func (s *Store) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

var _ repository.Store = (*Store)(nil)
