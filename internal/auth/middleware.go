package auth

import (
	"net/http"

	"github.com/MattSharp0/transaction-split-go/internal/logger"
)

// RequireAuth middleware validates JWT token and adds user ID to context
func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			logger.Warn("Missing authorization header")
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		tokenString, err := ExtractTokenFromHeader(authHeader)
		if err != nil {
			logger.Warn("Invalid authorization header format", "error", err)
			http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
			return
		}

		userID, err := ValidateToken(tokenString)
		if err != nil {
			if err == ErrExpiredToken {
				logger.Warn("Token expired")
				http.Error(w, "Token expired", http.StatusUnauthorized)
				return
			}
			logger.Warn("Invalid token", "error", err)
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Add user ID to context
		ctx := SetUserID(r.Context(), userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// OptionalAuth middleware extracts user from token if present but doesn't require it
func OptionalAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader != "" {
			tokenString, err := ExtractTokenFromHeader(authHeader)
			if err == nil {
				userID, err := ValidateToken(tokenString)
				if err == nil {
					ctx := SetUserID(r.Context(), userID)
					r = r.WithContext(ctx)
				}
			}
		}
		next.ServeHTTP(w, r)
	})
}
