package auth

import (
	"net/http"

	"github.com/MattSharp0/transaction-split-go/internal/logger"
)

// RequireAuth middleware validates JWT token from header or cookie and adds user ID to context
func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var tokenString string
		var err error

		// Try to get token from cookie first, then fall back to header
		tokenString = GetTokenFromCookie(r, AccessTokenCookieName)
		if tokenString == "" {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				logger.Warn("Missing authorization header or cookie")
				http.Error(w, "Authorization required", http.StatusUnauthorized)
				return
			}

			tokenString, err = ExtractTokenFromHeader(authHeader)
			if err != nil {
				logger.Warn("Invalid authorization header format", "error", err)
				http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
				return
			}
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
		var tokenString string
		var err error

		// Try cookie first
		tokenString = GetTokenFromCookie(r, AccessTokenCookieName)
		if tokenString == "" {
			// Fall back to header
			authHeader := r.Header.Get("Authorization")
			if authHeader != "" {
				tokenString, err = ExtractTokenFromHeader(authHeader)
				if err == nil {
					userID, err := ValidateToken(tokenString)
					if err == nil {
						ctx := SetUserID(r.Context(), userID)
						r = r.WithContext(ctx)
					}
				}
			}
		} else {
			userID, err := ValidateToken(tokenString)
			if err == nil {
				ctx := SetUserID(r.Context(), userID)
				r = r.WithContext(ctx)
			}
		}
		next.ServeHTTP(w, r)
	})
}

// RequireCSRF middleware validates CSRF token on state-changing requests
func RequireCSRF(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only check CSRF on state-changing methods
		if r.Method == "GET" || r.Method == "HEAD" || r.Method == "OPTIONS" {
			next.ServeHTTP(w, r)
			return
		}

		// Get user ID from context (must be authenticated)
		userID, ok := GetUserID(r.Context())
		if !ok {
			logger.Warn("CSRF check failed: user not authenticated")
			http.Error(w, "Authentication required", http.StatusUnauthorized)
			return
		}

		// Get CSRF token from cookie
		csrfTokenCookie := GetTokenFromCookie(r, CSRFTokenCookieName)
		if csrfTokenCookie == "" {
			logger.Warn("CSRF token cookie missing")
			http.Error(w, "CSRF token required", http.StatusForbidden)
			return
		}

		// Get CSRF token from header
		csrfTokenHeader := r.Header.Get("X-CSRF-Token")
		if csrfTokenHeader == "" {
			logger.Warn("CSRF token header missing")
			http.Error(w, "CSRF token required", http.StatusForbidden)
			return
		}

		// Validate that cookie and header tokens match (double-submit pattern)
		if csrfTokenCookie != csrfTokenHeader {
			logger.Warn("CSRF token mismatch")
			http.Error(w, "Invalid CSRF token", http.StatusForbidden)
			return
		}

		// Validate CSRF token signature and user ID
		if err := ValidateCSRFToken(csrfTokenHeader, userID); err != nil {
			logger.Warn("CSRF token validation failed", "error", err)
			http.Error(w, "Invalid CSRF token", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}
