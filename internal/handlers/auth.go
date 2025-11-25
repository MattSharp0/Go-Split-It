package handlers

import (
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"

	db "github.com/MattSharp0/transaction-split-go/db/sqlc"
	"github.com/MattSharp0/transaction-split-go/internal/auth"
	"github.com/MattSharp0/transaction-split-go/internal/logger"
	"github.com/MattSharp0/transaction-split-go/internal/models"
	"github.com/MattSharp0/transaction-split-go/internal/server"
)

func AuthRoutes(s *server.Server, store db.Store) *http.ServeMux {
	mux := http.NewServeMux()

	// Public routes
	mux.HandleFunc("POST /register", register(store)) // POST auth/register: Register new user
	mux.HandleFunc("POST /login", login(store))       // POST auth/login: Login
	mux.HandleFunc("POST /refresh", refresh(store))   // POST auth/refresh: Refresh tokens

	// Protected routes
	mux.HandleFunc("GET /me", auth.RequireAuth(http.HandlerFunc(getMe(store))).ServeHTTP)                         // GET auth/me: Get current user
	mux.HandleFunc("POST /logout", auth.RequireAuth(auth.RequireCSRF(http.HandlerFunc(logout(store)))).ServeHTTP) // POST auth/logout: Logout
	mux.HandleFunc("GET /csrf-token", auth.RequireAuth(http.HandlerFunc(getCSRFToken())).ServeHTTP)               // GET auth/csrf-token: Get CSRF token

	return mux
}

func register(store db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.RegisterRequest
		if err := DecodeJSONBody(r, &req); err != nil {
			http.Error(w, "Bad request: invalid JSON", http.StatusBadRequest)
			return
		}

		// Validate input
		if req.Name == "" {
			http.Error(w, "Name is required", http.StatusBadRequest)
			return
		}
		if req.Email == "" {
			http.Error(w, "Email is required", http.StatusBadRequest)
			return
		}
		if req.Password == "" {
			http.Error(w, "Password is required", http.StatusBadRequest)
			return
		}
		if len(req.Password) < 8 {
			http.Error(w, "Password must be at least 8 characters", http.StatusBadRequest)
			return
		}

		logger.Info("Registering user", slog.String("email", req.Email))

		// Hash password
		passwordHash, err := auth.HashPassword(req.Password)
		if err != nil {
			logger.Error("Failed to hash password", "error", err)
			http.Error(w, "An error has occurred", http.StatusInternalServerError)
			return
		}

		// Create user
		user, err := store.CreateUserWithAuth(r.Context(), db.CreateUserWithAuthParams{
			Name:         req.Name,
			Email:        req.Email,
			PasswordHash: passwordHash,
		})
		if err != nil {
			// Check if email already exists
			if err.Error() == "duplicate key value violates unique constraint \"users_email_key\"" {
				logger.Warn("Email already exists", "email", req.Email)
				http.Error(w, "Email already registered", http.StatusConflict)
				return
			}
			logger.Error("Failed to create user", "error", err)
			http.Error(w, "An error has occurred", http.StatusInternalServerError)
			return
		}

		logger.Debug("User registered successfully", slog.Int64("user_id", user.ID), slog.String("email", user.Email))

		// Generate access and refresh tokens
		accessToken, err := auth.GenerateAccessToken(user.ID)
		if err != nil {
			logger.Error("Failed to generate access token", "error", err)
			http.Error(w, "An error has occurred", http.StatusInternalServerError)
			return
		}

		refreshToken, err := auth.GenerateRefreshToken()
		if err != nil {
			logger.Error("Failed to generate refresh token", "error", err)
			http.Error(w, "An error has occurred", http.StatusInternalServerError)
			return
		}

		// Store refresh token in database
		refreshTokenHash := auth.HashRefreshToken(refreshToken)
		expirationDays := 7 // default 7 days
		if expStr := os.Getenv("REFRESH_TOKEN_EXPIRATION_DAYS"); expStr != "" {
			if parsed, err := strconv.Atoi(expStr); err == nil {
				expirationDays = parsed
			}
		}
		expiresAt := time.Now().Add(time.Duration(expirationDays) * 24 * time.Hour)

		_, err = store.CreateRefreshToken(r.Context(), db.CreateRefreshTokenParams{
			TokenHash:  refreshTokenHash,
			UserID:     user.ID,
			ExpiresAt:  expiresAt,
			DeviceInfo: nil, // Could extract from User-Agent header
		})
		if err != nil {
			logger.Error("Failed to store refresh token", "error", err)
			http.Error(w, "An error has occurred", http.StatusInternalServerError)
			return
		}

		// Generate CSRF token
		csrfToken, err := auth.GenerateCSRFToken(user.ID)
		if err != nil {
			logger.Error("Failed to generate CSRF token", "error", err)
			http.Error(w, "An error has occurred", http.StatusInternalServerError)
			return
		}

		// Set cookies
		accessTokenMaxAge := 30 * 60 // 30 minutes in seconds
		refreshTokenMaxAge := expirationDays * 24 * 60 * 60
		csrfTokenMaxAge := 24 * 60 * 60 // 24 hours

		auth.SetAuthCookie(w, auth.AccessTokenCookieName, accessToken, accessTokenMaxAge)
		auth.SetAuthCookie(w, auth.RefreshTokenCookieName, refreshToken, refreshTokenMaxAge)
		auth.SetCSRFCookie(w, csrfToken, csrfTokenMaxAge)

		// Convert to response format
		userResponse := models.UserResponse{
			ID:         user.ID,
			Name:       user.Name,
			CreatedAt:  user.CreatedAt,
			ModifiedAt: user.ModifiedAt,
		}

		loginResponse := models.LoginResponse{
			Token:     accessToken, // Still return access token in JSON for header-based clients
			User:      userResponse,
			CSRFToken: csrfToken, // Include CSRF token in response
		}

		// Send response with 201 Created status
		if err := WriteJSONResponseCreated(w, loginResponse); err != nil {
			http.Error(w, "An error has occurred", http.StatusInternalServerError)
			return
		}
	}
}

func login(store db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req models.LoginRequest
		if err := DecodeJSONBody(r, &req); err != nil {
			http.Error(w, "Bad request: invalid JSON", http.StatusBadRequest)
			return
		}

		// Validate input
		if req.Email == "" {
			http.Error(w, "Email is required", http.StatusBadRequest)
			return
		}
		if req.Password == "" {
			http.Error(w, "Password is required", http.StatusBadRequest)
			return
		}

		logger.Debug("Login attempt", slog.String("email", req.Email))

		// Get user by email
		user, err := store.GetUserByEmail(r.Context(), req.Email)
		if err != nil {
			logger.Warn("Login failed: user not found", "email", req.Email)
			http.Error(w, "Invalid email or password", http.StatusUnauthorized)
			return
		}

		// Verify password
		if err := auth.VerifyPassword(user.PasswordHash, req.Password); err != nil {
			logger.Warn("Login failed: invalid password", "email", req.Email)
			http.Error(w, "Invalid email or password", http.StatusUnauthorized)
			return
		}

		logger.Debug("Password verified successfully", slog.Int64("user_id", user.ID))

		// Generate access and refresh tokens
		accessToken, err := auth.GenerateAccessToken(user.ID)
		if err != nil {
			logger.Error("Failed to generate access token", "error", err)
			http.Error(w, "An error has occurred", http.StatusInternalServerError)
			return
		}

		refreshToken, err := auth.GenerateRefreshToken()
		if err != nil {
			logger.Error("Failed to generate refresh token", "error", err)
			http.Error(w, "An error has occurred", http.StatusInternalServerError)
			return
		}

		// Store refresh token in database
		refreshTokenHash := auth.HashRefreshToken(refreshToken)
		expirationDays := 7 // default 7 days
		if expStr := os.Getenv("REFRESH_TOKEN_EXPIRATION_DAYS"); expStr != "" {
			if parsed, err := strconv.Atoi(expStr); err == nil {
				expirationDays = parsed
			}
		}
		expiresAt := time.Now().Add(time.Duration(expirationDays) * 24 * time.Hour)

		_, err = store.CreateRefreshToken(r.Context(), db.CreateRefreshTokenParams{
			TokenHash:  refreshTokenHash,
			UserID:     user.ID,
			ExpiresAt:  expiresAt,
			DeviceInfo: nil, // TODO: extract from User-Agent header
		})
		if err != nil {
			logger.Error("Failed to store refresh token", "error", err)
			http.Error(w, "An error has occurred", http.StatusInternalServerError)
			return
		}

		// Generate CSRF token
		csrfToken, err := auth.GenerateCSRFToken(user.ID)
		if err != nil {
			logger.Error("Failed to generate CSRF token", "error", err)
			http.Error(w, "An error has occurred", http.StatusInternalServerError)
			return
		}

		// Set cookies
		accessTokenMaxAge := 30 * 60 // 30 minutes in seconds
		refreshTokenMaxAge := expirationDays * 24 * 60 * 60
		csrfTokenMaxAge := 24 * 60 * 60 // 24 hours

		auth.SetAuthCookie(w, auth.AccessTokenCookieName, accessToken, accessTokenMaxAge)
		auth.SetAuthCookie(w, auth.RefreshTokenCookieName, refreshToken, refreshTokenMaxAge)
		auth.SetCSRFCookie(w, csrfToken, csrfTokenMaxAge)

		// Convert to response format
		userResponse := models.UserResponse{
			ID:         user.ID,
			Name:       user.Name,
			CreatedAt:  user.CreatedAt,
			ModifiedAt: user.ModifiedAt,
		}

		loginResponse := models.LoginResponse{
			Token:     accessToken, // Still return access token in JSON for header-based clients
			User:      userResponse,
			CSRFToken: csrfToken, // Include CSRF token in response
		}

		logger.Debug("Login successful", slog.Int64("user_id", user.ID))

		// Send response
		if err := WriteJSONResponseOK(w, loginResponse); err != nil {
			http.Error(w, "An error has occurred", http.StatusInternalServerError)
			return
		}
	}
}

func getMe(store db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := GetAuthenticatedUserID(w, r)
		if !ok {
			return
		}

		logger.Debug("Getting current user", "user_id", userID)

		// Get user from database
		user, err := store.GetUserByID(r.Context(), userID)
		if HandleDBError(w, err, "User not found", "An error has occurred", "Failed to get user by ID", "user_id", userID) {
			return
		}

		// Convert to response format
		userResponse := models.UserResponse{
			ID:         user.ID,
			Name:       user.Name,
			CreatedAt:  user.CreatedAt,
			ModifiedAt: user.ModifiedAt,
		}

		// Send response
		if err := WriteJSONResponseOK(w, userResponse); err != nil {
			http.Error(w, "An error has occurred", http.StatusInternalServerError)
			return
		}
	}
}

func refresh(store db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get refresh token from cookie or header
		refreshToken := auth.GetTokenFromCookie(r, auth.RefreshTokenCookieName)
		if refreshToken == "" {
			// Try header as fallback
			authHeader := r.Header.Get("Authorization")
			if authHeader != "" {
				if token, err := auth.ExtractTokenFromHeader(authHeader); err == nil {
					refreshToken = token
				}
			}
		}

		if refreshToken == "" {
			logger.Warn("Refresh token missing")
			http.Error(w, "Refresh token required", http.StatusUnauthorized)
			return
		}

		// Validate refresh token
		userID, err := auth.ValidateRefreshToken(r.Context(), store, refreshToken)
		if err != nil {
			if err == auth.ErrRevokedToken {
				logger.Warn("Refresh token revoked")
				http.Error(w, "Token revoked", http.StatusUnauthorized)
				return
			}
			if err == auth.ErrExpiredToken {
				logger.Warn("Refresh token expired")
				http.Error(w, "Token expired", http.StatusUnauthorized)
				return
			}
			logger.Warn("Invalid refresh token", "error", err)
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Revoke old refresh token (token rotation)
		if err := auth.RevokeRefreshToken(r.Context(), store, refreshToken); err != nil {
			logger.Warn("Failed to revoke old refresh token", "error", err)
			// Continue anyway - token rotation is best practice but not critical
		}

		// Generate new access and refresh tokens
		accessToken, err := auth.GenerateAccessToken(userID)
		if err != nil {
			logger.Error("Failed to generate access token", "error", err)
			http.Error(w, "An error has occurred", http.StatusInternalServerError)
			return
		}

		newRefreshToken, err := auth.GenerateRefreshToken()
		if err != nil {
			logger.Error("Failed to generate refresh token", "error", err)
			http.Error(w, "An error has occurred", http.StatusInternalServerError)
			return
		}

		// Store new refresh token in database
		refreshTokenHash := auth.HashRefreshToken(newRefreshToken)
		expirationDays := 7 // default 7 days
		if expStr := os.Getenv("REFRESH_TOKEN_EXPIRATION_DAYS"); expStr != "" {
			if parsed, err := strconv.Atoi(expStr); err == nil {
				expirationDays = parsed
			}
		}
		expiresAt := time.Now().Add(time.Duration(expirationDays) * 24 * time.Hour)

		_, err = store.CreateRefreshToken(r.Context(), db.CreateRefreshTokenParams{
			TokenHash:  refreshTokenHash,
			UserID:     userID,
			ExpiresAt:  expiresAt,
			DeviceInfo: nil,
		})
		if err != nil {
			logger.Error("Failed to store refresh token", "error", err)
			http.Error(w, "An error has occurred", http.StatusInternalServerError)
			return
		}

		// Generate new CSRF token
		csrfToken, err := auth.GenerateCSRFToken(userID)
		if err != nil {
			logger.Error("Failed to generate CSRF token", "error", err)
			http.Error(w, "An error has occurred", http.StatusInternalServerError)
			return
		}

		// Set new cookies
		accessTokenMaxAge := 30 * 60 // 30 minutes in seconds
		refreshTokenMaxAge := expirationDays * 24 * 60 * 60
		csrfTokenMaxAge := 24 * 60 * 60 // 24 hours

		auth.SetAuthCookie(w, auth.AccessTokenCookieName, accessToken, accessTokenMaxAge)
		auth.SetAuthCookie(w, auth.RefreshTokenCookieName, newRefreshToken, refreshTokenMaxAge)
		auth.SetCSRFCookie(w, csrfToken, csrfTokenMaxAge)

		refreshResponse := models.RefreshResponse{
			Token:     accessToken,
			CSRFToken: csrfToken,
		}

		logger.Debug("Token refresh successful", slog.Int64("user_id", userID))

		// Send response
		if err := WriteJSONResponseOK(w, refreshResponse); err != nil {
			http.Error(w, "An error has occurred", http.StatusInternalServerError)
			return
		}
	}
}

func logout(store db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := GetAuthenticatedUserID(w, r)
		if !ok {
			http.Error(w, "Authentication required", http.StatusUnauthorized)
			return
		}

		// Get refresh token from cookie or header
		refreshToken := auth.GetTokenFromCookie(r, auth.RefreshTokenCookieName)
		if refreshToken == "" {
			// Try header as fallback
			authHeader := r.Header.Get("Authorization")
			if authHeader != "" {
				if token, err := auth.ExtractTokenFromHeader(authHeader); err == nil {
					refreshToken = token
				}
			}
		}

		// Revoke the refresh token if provided
		if refreshToken != "" {
			if err := auth.RevokeRefreshToken(r.Context(), store, refreshToken); err != nil {
				logger.Warn("Failed to revoke refresh token", "error", err)
				// Continue anyway
			}
		}

		// Optionally revoke all user tokens
		// if err := auth.RevokeAllUserTokens(r.Context(), store, userID); err != nil {
		// 	logger.Warn("Failed to revoke all user tokens", "error", err)
		// }

		// Clear all auth cookies
		auth.ClearAuthCookies(w)

		logger.Debug("Logout successful", slog.Int64("user_id", userID))

		// Send success response
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message":"Logged out successfully"}`))
	}
}

func getCSRFToken() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, ok := GetAuthenticatedUserID(w, r)
		if !ok {
			http.Error(w, "Authentication required", http.StatusUnauthorized)
			return
		}

		// Generate CSRF token
		csrfToken, err := auth.GenerateCSRFToken(userID)
		if err != nil {
			logger.Error("Failed to generate CSRF token", "error", err)
			http.Error(w, "An error has occurred", http.StatusInternalServerError)
			return
		}

		// Set CSRF cookie
		csrfTokenMaxAge := 24 * 60 * 60 // 24 hours
		auth.SetCSRFCookie(w, csrfToken, csrfTokenMaxAge)

		// Return CSRF token in response
		response := map[string]string{
			"csrf_token": csrfToken,
		}

		if err := WriteJSONResponseOK(w, response); err != nil {
			http.Error(w, "An error has occurred", http.StatusInternalServerError)
			return
		}
	}
}
