package middleware

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

func EnsureValidToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := extractToken(r)
		if token == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		tokenClaims, err := decodeJWTPayload(token)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		//check tokenclaims is valid
		exp, ok := tokenClaims["exp"].(float64)
		if !ok || int64(exp) < time.Now().Unix() {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		id, ok := tokenClaims["user"].(map[string]interface{})["id"]
		if !ok || id == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), "tokenClaims", tokenClaims)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

// extractToken extracts the token from the Authorization header
func extractToken(r *http.Request) string {
	authHeader := r.Header.Get("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		return strings.TrimPrefix(authHeader, "Bearer ")
	}
	return ""
}

func decodeJWTPayload(tokenString string) (map[string]interface{}, error) {
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid token format")
	}

	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("error decoding payload: %v", err)
	}

	var claims map[string]interface{}
	err = json.Unmarshal(payload, &claims)
	if err != nil {
		return nil, fmt.Errorf("error unmarshaling JSON: %v", err)
	}

	return claims, nil
}
