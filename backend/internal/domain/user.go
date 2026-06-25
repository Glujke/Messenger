package domain

import "time"

// User represents a registered messenger account.
type User struct {
	ID        int64
	Email     string
	Username  string
	Verified  bool
	CreatedAt time.Time
}
