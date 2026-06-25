package api

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
)

type jwtClaims struct {
	UserID int64 `json:"uid"`
}

// ParseUserIDFromJWT extracts uid claim from JWT without verifying signature.
// This is used only for client-side UI filtering.
func ParseUserIDFromJWT(token string) (int64, error) {
	parts := strings.Split(token, ".")
	if len(parts) < 2 {
		return 0, fmt.Errorf("invalid token")
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return 0, fmt.Errorf("invalid token payload")
	}

	var claims jwtClaims
	if err := json.Unmarshal(payload, &claims); err != nil {
		return 0, fmt.Errorf("invalid token claims")
	}
	if claims.UserID <= 0 {
		return 0, fmt.Errorf("uid missing in token")
	}
	return claims.UserID, nil
}

