package handler

import "context"

type authUserKey struct{}

// AuthUser holds authenticated user data attached to a request context.
type AuthUser struct {
	ID    int64
	Email string
}

// WithAuthUser stores authenticated user data in the context.
func WithAuthUser(ctx context.Context, user AuthUser) context.Context {
	return context.WithValue(ctx, authUserKey{}, user)
}

// AuthUserFromContext reads authenticated user data from the context.
func AuthUserFromContext(ctx context.Context) (AuthUser, bool) {
	user, ok := ctx.Value(authUserKey{}).(AuthUser)
	return user, ok
}
