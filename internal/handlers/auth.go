package handlers

import (
	"context"
	"log/slog"
	"net/http"

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

	// Protected routes
	mux.HandleFunc("GET /me", auth.RequireAuth(http.HandlerFunc(getMe(store))).ServeHTTP) // GET auth/me: Get current user

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
		user, err := store.CreateUserWithAuth(context.Background(), db.CreateUserWithAuthParams{
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

		// Generate token
		token, err := auth.GenerateToken(user.ID)
		if err != nil {
			logger.Error("Failed to generate token", "error", err)
			http.Error(w, "An error has occurred", http.StatusInternalServerError)
			return
		}

		// Convert to response format
		userResponse := models.UserResponse{
			ID:         user.ID,
			Name:       user.Name,
			CreatedAt:  user.CreatedAt,
			ModifiedAt: user.ModifiedAt,
		}

		loginResponse := models.LoginResponse{
			Token: token,
			User:  userResponse,
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
		user, err := store.GetUserByEmail(context.Background(), req.Email)
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

		// Generate token
		token, err := auth.GenerateToken(user.ID)
		if err != nil {
			logger.Error("Failed to generate token", "error", err)
			http.Error(w, "An error has occurred", http.StatusInternalServerError)
			return
		}

		// Convert to response format
		userResponse := models.UserResponse{
			ID:         user.ID,
			Name:       user.Name,
			CreatedAt:  user.CreatedAt,
			ModifiedAt: user.ModifiedAt,
		}

		loginResponse := models.LoginResponse{
			Token: token,
			User:  userResponse,
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
		userID, ok := auth.GetUserID(r.Context())
		if !ok {
			logger.Warn("User ID not found in context")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		logger.Debug("Getting current user", "user_id", userID)

		// Get user from database
		user, err := store.GetUserByID(context.Background(), userID)
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
