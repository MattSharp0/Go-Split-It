package handlers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	db "github.com/MattSharp0/transaction-split-go/db/sqlc"
	"github.com/MattSharp0/transaction-split-go/internal/logger"
	"github.com/MattSharp0/transaction-split-go/internal/models"
	"github.com/MattSharp0/transaction-split-go/internal/server"
)

func UserRoutes(s *server.Server, q db.Store) *http.ServeMux {
	mux := http.NewServeMux()

	// Root path handlers - different methods handled within functions
	mux.HandleFunc("POST /", createUser(q)) // POST: Create user
	mux.HandleFunc("GET /", listUsers(q))   // GET: List users (with body params)

	// ID path handlers - different methods handled within functions
	mux.HandleFunc("GET /{id}", getUserByID(q))   // GET: Get user by ID
	mux.HandleFunc("PUT /{id}", updateUser(q))    // PUT: Update user
	mux.HandleFunc("PATCH /{id}", updateUser(q))  // PATCH: Update user
	mux.HandleFunc("DELETE /{id}", deleteUser(q)) // DELETE: Delete user

	// Nested resource handlers - RESTful approach for user transactions
	mux.HandleFunc("GET /{user_id}/transactions", getTransactionsByUserNested(q)) // GET: List transactions by user

	return mux
}

func listUsers(store db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse query parameters
		queryParams := r.URL.Query()

		// Default values
		var listuserparams db.ListUsersParams
		listuserparams.Limit = 100
		listuserparams.Offset = 0

		// Parse limit
		if limitStr := queryParams.Get("limit"); limitStr != "" {
			limit, err := strconv.ParseInt(limitStr, 10, 32)
			if err != nil {
				logger.Warn("Invalid limit parameter", "error", err, "value", limitStr)
				http.Error(w, "Invalid limit parameter", http.StatusBadRequest)
				return
			}
			listuserparams.Limit = int32(limit)
		}

		// Parse offset
		if offsetStr := queryParams.Get("offset"); offsetStr != "" {
			offset, err := strconv.ParseInt(offsetStr, 10, 32)
			if err != nil {
				logger.Warn("Invalid offset parameter", "error", err, "value", offsetStr)
				http.Error(w, "Invalid offset parameter", http.StatusBadRequest)
				return
			}
			listuserparams.Offset = int32(offset)
		}

		logger.Debug("Listing users",
			slog.Int("limit", int(listuserparams.Limit)),
			slog.Int("offset", int(listuserparams.Offset)),
		)

		users, err := store.ListUsers(context.Background(), listuserparams)
		if HandleDBListError(w, err, "An error has occurred", "Failed to list users", "limit", listuserparams.Limit, "offset", listuserparams.Offset) {
			return
		}

		userResponses := make([]models.UserResponse, len(users))
		for i, user := range users {
			userResponses[i] = models.UserResponse{
				ID:         user.ID,
				Name:       user.Name,
				CreatedAt:  user.CreatedAt,
				ModifiedAt: user.ModifiedAt,
			}
		}

		count := len(userResponses)

		listuserresponse := models.ListUserResponse{
			Users:  userResponses,
			Count:  int32(count),
			Limit:  listuserparams.Limit,
			Offset: listuserparams.Offset,
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(listuserresponse); err != nil {
			logger.Error("Failed to encode user response", "error", err)
			http.Error(w, "An error has occured", http.StatusInternalServerError)
			return
		}

		logger.Debug("Successfully listed users", "count", count)
	}
}

func getUserByID(store db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract {id} from path parameter
		idStr := r.PathValue("id")
		if idStr == "" {
			http.Error(w, "User ID is required", http.StatusBadRequest)
			return
		}

		// Convert string ID to int64
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			logger.Warn("Invalid user ID format", "error", err, "value", idStr)
			http.Error(w, "Invalid user ID format", http.StatusBadRequest)
			return
		}

		logger.Debug("Getting user by ID", "user_id", id)

		// Get user from database
		user, err := store.GetUserByID(context.Background(), id)
		if HandleDBError(w, err, "User not found", "An error has occurred", "Failed to get user by ID", "user_id", id) {
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
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(userResponse); err != nil {
			logger.Error("Failed to encode user response", "error", err)
			http.Error(w, "An error has occurred", http.StatusInternalServerError)
			return
		}
	}
}

func createUser(store db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Decode request body
		decoder := json.NewDecoder(r.Body)
		defer r.Body.Close()

		var createUserReq models.CreateUserRequest
		err := decoder.Decode(&createUserReq)
		if err != nil {
			http.Error(w, "Bad request: invalid JSON", http.StatusBadRequest)
			return
		}

		// Validate input
		if createUserReq.Name == "" {
			logger.Warn("Create user request missing name")
			http.Error(w, "Name is required", http.StatusBadRequest)
			return
		}

		logger.Info("Creating user", slog.String("name", createUserReq.Name))

		// Create user in database
		user, err := store.CreateUser(context.Background(), createUserReq.Name)
		if HandleDBListError(w, err, "An error has occurred", "Failed to create user", "name", createUserReq.Name) {
			return
		}

		logger.Debug("User created successfully",
			slog.Int64("user_id", user.ID),
			slog.String("name", user.Name),
		)

		// Convert to response format
		userResponse := models.UserResponse{
			ID:         user.ID,
			Name:       user.Name,
			CreatedAt:  user.CreatedAt,
			ModifiedAt: user.ModifiedAt,
		}

		// Send response with 201 Created status
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(userResponse); err != nil {
			logger.Error("Failed to encode user response", "error", err)
			http.Error(w, "An error has occurred", http.StatusInternalServerError)
			return
		}
	}
}

func updateUser(store db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract {id} from path parameter
		idStr := r.PathValue("id")
		if idStr == "" {
			http.Error(w, "User ID is required", http.StatusBadRequest)
			return
		}

		// Convert string ID to int64
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid user ID format", http.StatusBadRequest)
			return
		}

		// Decode request body
		decoder := json.NewDecoder(r.Body)
		defer r.Body.Close()

		var updateUserReq models.UpdateUserRequest
		err = decoder.Decode(&updateUserReq)
		if err != nil {
			http.Error(w, "Bad request: invalid JSON", http.StatusBadRequest)
			return
		}

		// Validate input
		if updateUserReq.Name == "" {
			logger.Warn("Update user request missing name")
			http.Error(w, "Name is required", http.StatusBadRequest)
			return
		}

		logger.Debug("Updating user", "user_id", id, "new_name", updateUserReq.Name)

		// Update user in database
		user, err := store.UpdateUser(context.Background(), db.UpdateUserParams{
			ID:   id,
			Name: updateUserReq.Name,
		})
		if HandleDBError(w, err, "User not found", "An error has occurred", "Failed to update user", "user_id", id) {
			return
		}

		// Convert to response format
		userResponse := models.UserResponse{
			ID:         user.ID,
			Name:       user.Name,
			CreatedAt:  user.CreatedAt,
			ModifiedAt: user.ModifiedAt,
		}

		logger.Debug("User updated successfully", "user_id", user.ID)

		// Send response
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(userResponse); err != nil {
			logger.Error("Failed to encode user response", "error", err)
			http.Error(w, "An error has occurred", http.StatusInternalServerError)
			return
		}
	}
}

func deleteUser(store db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract {id} from path parameter
		idStr := r.PathValue("id")
		if idStr == "" {
			http.Error(w, "User ID is required", http.StatusBadRequest)
			return
		}

		// Convert string ID to int64
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid user ID format", http.StatusBadRequest)
			return
		}

		logger.Debug("Deleting user", "user_id", id)

		// Delete user from database
		user, err := store.DeleteUser(context.Background(), id)
		if HandleDBError(w, err, "User not found", "An error has occurred", "Failed to delete user", "user_id", id) {
			return
		}

		logger.Debug("User deleted successfully", "user_id", user.ID)

		// Convert to response format
		userResponse := models.UserResponse{
			ID:         user.ID,
			Name:       user.Name,
			CreatedAt:  user.CreatedAt,
			ModifiedAt: user.ModifiedAt,
		}

		// Send response with deleted user data
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(userResponse); err != nil {
			logger.Error("Failed to encode user response", "error", err)
			http.Error(w, "An error has occurred", http.StatusInternalServerError)
			return
		}
	}
}

// Nested resource handler

// List transactions for user
// GET /users/{user_id}/transactions
func getTransactionsByUserNested(store db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract {user_id} from path parameter
		userIDStr := r.PathValue("user_id")
		if userIDStr == "" {
			http.Error(w, "User ID is required", http.StatusBadRequest)
			return
		}

		// Convert string ID to int64
		userID, err := strconv.ParseInt(userIDStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid user ID format", http.StatusBadRequest)
			return
		}

		// Parse query parameters
		queryParams := r.URL.Query()

		// Default values
		var listParams db.GetTransactionsByUserInPeriodParams
		listParams.ByUser = userID
		listParams.StartDate = time.Now().AddDate(-1, 0, 0)
		listParams.EndDate = time.Now()
		listParams.Limit = 100
		listParams.Offset = 0

		// Parse start_date

		// Parse start_date
		if startDateStr := queryParams.Get("start_date"); startDateStr != "" {
			startDate, err := time.Parse("2006-01-02", startDateStr)
			if err != nil {
				http.Error(w, "Invalid start_date format, use YYYY-MM-DD", http.StatusBadRequest)
				return
			}
			listParams.StartDate = startDate
		}

		// Parse end_date
		if endDateStr := queryParams.Get("end_date"); endDateStr != "" {
			endDate, err := time.Parse("2006-01-02", endDateStr)
			if err != nil {
				http.Error(w, "Invalid end_date format, use YYYY-MM-DD", http.StatusBadRequest)
				return
			}
			listParams.EndDate = endDate
		}

		// Parse limit
		if limitStr := queryParams.Get("limit"); limitStr != "" {
			limit, err := strconv.ParseInt(limitStr, 10, 32)
			if err != nil {
				http.Error(w, "Invalid limit parameter", http.StatusBadRequest)
				return
			}
			listParams.Limit = int32(limit)
		}

		// Parse offset
		if offsetStr := queryParams.Get("offset"); offsetStr != "" {
			offset, err := strconv.ParseInt(offsetStr, 10, 32)
			if err != nil {
				http.Error(w, "Invalid offset parameter", http.StatusBadRequest)
				return
			}
			listParams.Offset = int32(offset)
		}

		logger.Debug("Listing transactions for user",
			"user_id", userID,
			"start_date", listParams.StartDate,
			"end_date", listParams.EndDate,
			"limit", listParams.Limit,
			"offset", listParams.Offset,
		)

		transactions, err := store.GetTransactionsByUserInPeriod(context.Background(), listParams)
		if HandleDBListError(w, err, "An error has occurred", "Failed to get transactions by user", "user_id", userID) {
			return
		}

		transactionResponses := make([]models.TransactionResponse, len(transactions))
		for i, tx := range transactions {
			transactionResponses[i] = models.TransactionResponse{
				ID:              tx.ID,
				GroupID:         tx.GroupID,
				Name:            tx.Name,
				TransactionDate: tx.TransactionDate,
				Amount:          tx.Amount,
				Category:        tx.Category,
				Note:            tx.Note,
				ByUser:          tx.ByUser,
				CreatedAt:       tx.CreatedAt,
				ModifiedAt:      tx.ModifiedAt,
			}
		}

		count := len(transactionResponses)

		listTransactionResponse := models.ListTransactionResponse{
			Transactions: transactionResponses,
			Count:        int32(count),
			Limit:        listParams.Limit,
			Offset:       listParams.Offset,
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(listTransactionResponse); err != nil {
			logger.Error("Failed to encode transaction response", "error", err)
			http.Error(w, "An error has occurred", http.StatusInternalServerError)
			return
		}

		logger.Debug("Successfully listed user transactions", "user_id", userID, "count", count)
	}
}
