package handlers

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	db "github.com/MattSharp0/transaction-split-go/db/sqlc"
	"github.com/MattSharp0/transaction-split-go/internal/auth"
	"github.com/MattSharp0/transaction-split-go/internal/logger"
	"github.com/MattSharp0/transaction-split-go/internal/models"
	"github.com/MattSharp0/transaction-split-go/internal/server"
	"github.com/shopspring/decimal"
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
	mux.HandleFunc("GET /{user_id}/balances", getUserBalances(q))                 // GET: Get user balances

	return mux
}

func listUsers(store db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse query parameters
		limit, offset, err := ParseLimitOffset(r)
		if err != nil {
			logger.Warn("Invalid parameter", "error", err)
			http.Error(w, "Invalid parameter: "+err.Error(), http.StatusBadRequest)
			return
		}

		var listuserparams db.ListUsersParams
		listuserparams.Limit = limit
		listuserparams.Offset = offset

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

		if err := WriteJSONResponseOK(w, listuserresponse); err != nil {
			http.Error(w, "An error has occurred", http.StatusInternalServerError)
			return
		}

		logger.Debug("Successfully listed users", "count", count)
	}
}

func getUserByID(store db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract {id} from path parameter
		id, ok := ParsePathInt64(w, r, "id", "User ID is required")
		if !ok {
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
		if err := WriteJSONResponseOK(w, userResponse); err != nil {
			http.Error(w, "An error has occurred", http.StatusInternalServerError)
			return
		}
	}
}

func createUser(store db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Decode request body
		var createUserReq models.CreateUserRequest
		if err := DecodeJSONBody(r, &createUserReq); err != nil {
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
		if err := WriteJSONResponseCreated(w, userResponse); err != nil {
			http.Error(w, "An error has occurred", http.StatusInternalServerError)
			return
		}
	}
}

func updateUser(store db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get authenticated user ID
		authenticatedUserID, ok := auth.GetUserID(r.Context())
		if !ok {
			logger.Warn("User ID not found in context")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Extract {id} from path parameter
		id, ok := ParsePathInt64(w, r, "id", "User ID is required")
		if !ok {
			return
		}

		// Verify user can only update their own account
		if err := auth.CheckOwnUser(authenticatedUserID, id); err != nil {
			http.Error(w, "Forbidden: cannot update another user's account", http.StatusForbidden)
			return
		}

		// Decode request body
		var updateUserReq models.UpdateUserRequest
		if err := DecodeJSONBody(r, &updateUserReq); err != nil {
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
		if err := WriteJSONResponseOK(w, userResponse); err != nil {
			http.Error(w, "An error has occurred", http.StatusInternalServerError)
			return
		}
	}
}

func deleteUser(store db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get authenticated user ID
		authenticatedUserID, ok := auth.GetUserID(r.Context())
		if !ok {
			logger.Warn("User ID not found in context")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Extract {id} from path parameter
		id, ok := ParsePathInt64(w, r, "id", "User ID is required")
		if !ok {
			return
		}

		// Verify user can only delete their own account
		if err := auth.CheckOwnUser(authenticatedUserID, id); err != nil {
			http.Error(w, "Forbidden: cannot delete another user's account", http.StatusForbidden)
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
		if err := WriteJSONResponseOK(w, userResponse); err != nil {
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
		userID, ok := ParsePathInt64(w, r, "user_id", "User ID is required")
		if !ok {
			return
		}

		// Parse query parameters
		limit, offset, err := ParseLimitOffset(r)
		if err != nil {
			http.Error(w, "Invalid parameter: "+err.Error(), http.StatusBadRequest)
			return
		}

		// Default values
		var listParams db.GetTransactionsByUserInPeriodParams
		listParams.ByUser = userID

		// Default to past year, TODO: make this configurable
		defaultStartDate := time.Now().AddDate(-1, 0, 0)
		defaultEndDate := time.Now()

		// Parse dates
		startDate, err := ParseQueryDate(r, "start_date", defaultStartDate)
		if err != nil {
			http.Error(w, "Invalid start_date format, use YYYY-MM-DD", http.StatusBadRequest)
			return
		}

		endDate, err := ParseQueryDate(r, "end_date", defaultEndDate)
		if err != nil {
			http.Error(w, "Invalid end_date format, use YYYY-MM-DD", http.StatusBadRequest)
			return
		}

		listParams.StartDate = startDate
		listParams.EndDate = endDate
		listParams.Limit = limit
		listParams.Offset = offset

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

		if err := WriteJSONResponseOK(w, listTransactionResponse); err != nil {
			http.Error(w, "An error has occurred", http.StatusInternalServerError)
			return
		}

		logger.Debug("Successfully listed user transactions", "user_id", userID, "count", count)
	}
}

// Get user balances
// GET /users/{user_id}/balances
func getUserBalances(store db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get authenticated user ID
		authenticatedUserID, ok := auth.GetUserID(r.Context())
		if !ok {
			logger.Warn("User ID not found in context")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Extract {user_id} from path parameter
		pathUserID, ok := ParsePathInt64(w, r, "user_id", "User ID is required")
		if !ok {
			return
		}

		// Verify user can only access their own balances
		if err := auth.CheckOwnUser(authenticatedUserID, pathUserID); err != nil {
			http.Error(w, "Forbidden: cannot access another user's balances", http.StatusForbidden)
			return
		}

		logger.Debug("Getting balances for user", "user_id", pathUserID)

		// Get summary
		summaryRow, err := store.UserBalancesSummary(context.Background(), &pathUserID)
		if HandleDBError(w, err, "User not found", "An error has occurred", "Failed to get user balances summary", "user_id", pathUserID) {
			return
		}

		// logger.Debug("User balances summary",
		// 	"user_id", userID,
		// 	"net_balance", summaryRow.NetBalance.String(),
		// 	"total_owed", summaryRow.TotalOwed.String(),
		// 	"total_owed_to_user", summaryRow.TotalOwedToUser.String())

		summary := models.UserBalancesSummaryResponse{
			NetBalance:      summaryRow.NetBalance,
			TotalOwed:       summaryRow.TotalOwed,
			TotalOwedToUser: summaryRow.TotalOwedToUser,
		}

		// Get balances by group
		balancesByGroupRows, err := store.UserBalancesByGroup(context.Background(), &pathUserID)
		if HandleDBListError(w, err, "An error has occurred", "Failed to get user balances by group", "user_id", pathUserID) {
			return
		}

		balancesByGroup := make([]models.UserBalanceByGroupResponse, len(balancesByGroupRows))
		for i, bg := range balancesByGroupRows {
			totalOwed := decimal.Zero
			totalOwedToUser := decimal.Zero
			if bg.NetBalance.IsNegative() {
				totalOwed = bg.NetBalance.Neg()
			} else if bg.NetBalance.IsPositive() {
				totalOwedToUser = bg.NetBalance
			}
			balancesByGroup[i] = models.UserBalanceByGroupResponse{
				GroupID:         bg.GroupID,
				GroupName:       bg.GroupName,
				NetBalance:      bg.NetBalance,
				TotalOwed:       totalOwed,
				TotalOwedToUser: totalOwedToUser,
			}
		}

		// Get balances by member
		balancesByMemberRows, err := store.UserBalancesByMember(context.Background(), &pathUserID)
		if HandleDBListError(w, err, "An error has occurred", "Failed to get user balances by member", "user_id", pathUserID) {
			return
		}

		balancesByMember := make([]models.UserBalanceByMemberResponse, len(balancesByMemberRows))
		for i, bm := range balancesByMemberRows {
			memberID := int64(0)
			if bm.MemberUserID != nil {
				memberID = *bm.MemberUserID
			}
			totalOwed := decimal.Zero
			totalOwedToUser := decimal.Zero
			if bm.NetBalance.IsNegative() {
				totalOwed = bm.NetBalance.Neg()
			} else if bm.NetBalance.IsPositive() {
				totalOwedToUser = bm.NetBalance
			}
			balancesByMember[i] = models.UserBalanceByMemberResponse{
				MemberID:        memberID,
				MemberName:      bm.MemberName,
				NetBalance:      bm.NetBalance,
				TotalOwed:       totalOwed,
				TotalOwedToUser: totalOwedToUser,
			}
		}

		response := models.UserBalancesResponse{
			UserID:           pathUserID,
			Summary:          summary,
			BalancesByGroup:  balancesByGroup,
			BalancesByMember: balancesByMember,
			GroupCount:       int32(len(balancesByGroup)),
			MemberCount:      int32(len(balancesByMember)),
		}

		if err := WriteJSONResponseOK(w, response); err != nil {
			http.Error(w, "An error has occurred", http.StatusInternalServerError)
			return
		}
	}
}
