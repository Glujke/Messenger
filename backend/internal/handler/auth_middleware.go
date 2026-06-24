package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"messenger/backend/internal/service"
)

// RequireAuth validates Bearer JWT and attaches user data to the request context.
func RequireAuth(jwtSecret string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, ok := bearerToken(r.Header.Get("Authorization"))
		if !ok {
			writeUnauthorized(w)
			return
		}

		claims, err := service.ParseToken(token, jwtSecret)
		if err != nil {
			writeUnauthorized(w)
			return
		}

		user := AuthUser{
			ID:    claims.UserID,
			Email: claims.Email,
		}
		next.ServeHTTP(w, r.WithContext(WithAuthUser(r.Context(), user)))
	})
}

func bearerToken(header string) (string, bool) {
	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return "", false
	}

	token := strings.TrimSpace(header[len(prefix):])
	if token == "" {
		return "", false
	}
	return token, true
}

func writeUnauthorized(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_ = json.NewEncoder(w).Encode(errorResponse{Error: "unauthorized"})
}
