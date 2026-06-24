package repository

import "context"

// Store defines persistence operations implemented by infrastructure adapters.
// Business logic should depend on this interface, not on PostgreSQL details.
type Store interface {
	Ping(ctx context.Context) error
}
