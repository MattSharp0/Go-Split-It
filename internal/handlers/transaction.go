package handlers

import (
	"fmt"
	"log/slog"
	"net/http"

	db "github.com/MattSharp0/transaction-split-go/db/sqlc"
	"github.com/MattSharp0/transaction-split-go/internal/auth"
	"github.com/MattSharp0/transaction-split-go/internal/logger"
	"github.com/MattSharp0/transaction-split-go/internal/models"
	"github.com/MattSharp0/transaction-split-go/internal/server"
)

func TransactionRoutes(s *server.Server, q db.Store) *http.ServeMux {
	mux := http.NewServeMux()

	// Root path handlers
	mux.HandleFunc("POST /", createTransaction(q)) // POST: Create transaction
	mux.HandleFunc("GET /", listTransactions(q))   // GET: List transactions

	// ID path handlers
	mux.HandleFunc("GET /{id}", getTransactionByID(q))   // GET: Get transaction by ID
	mux.HandleFunc("PUT /{id}", updateTransaction(q))    // PUT: Update transaction
	mux.HandleFunc("PATCH /{id}", updateTransaction(q))  // PATCH: Update transaction
	mux.HandleFunc("DELETE /{id}", deleteTransaction(q)) // DELETE: Delete transaction

	// Nested resource handlers - RESTful approach for splits (batch operations only)
	mux.HandleFunc("GET /{transaction_id}/splits", getSplitsByTransactionNested(q))   // GET: List splits for transaction
	mux.HandleFunc("POST /{transaction_id}/splits", createTransactionSplitsBatch(q))  // POST: Create/replace all splits (batch)
	mux.HandleFunc("PUT /{transaction_id}/splits", updateTransactionSplitsBatch(q))   // PUT: Replace all splits (batch)
	mux.HandleFunc("PATCH /{transaction_id}/splits", updateTransactionSplitsBatch(q)) // PATCH: Replace all splits (batch)

	return mux
}

func listTransactions(store db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get authenticated user ID
		userID, ok := GetAuthenticatedUserID(w, r)
		if !ok {
			return
		}

		// Parse query parameters
		limit, offset, err := ParseLimitOffset(r)
		if err != nil {
			http.Error(w, "Invalid parameter: "+err.Error(), http.StatusBadRequest)
			return
		}

		var listTransactionParams db.ListTransactionsByUserGroupsParams
		listTransactionParams.UserID = &userID
		listTransactionParams.Limit = limit
		listTransactionParams.Offset = offset

		logger.Debug("Listing transactions",
			"limit", listTransactionParams.Limit,
			"offset", listTransactionParams.Offset,
		)

		transactions, err := store.ListTransactionsByUserGroups(r.Context(), listTransactionParams)
		if HandleDBListError(w, err, "An error has occurred", "Failed to list transactions", "limit", listTransactionParams.Limit, "offset", listTransactionParams.Offset) {
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
			Limit:        listTransactionParams.Limit,
			Offset:       listTransactionParams.Offset,
		}

		if err := WriteJSONResponseOK(w, listTransactionResponse); err != nil {
			http.Error(w, "An error has occurred", http.StatusInternalServerError)
			return
		}
	}
}

func getTransactionByID(store db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get authenticated user ID
		userID, ok := GetAuthenticatedUserID(w, r)
		if !ok {
			return
		}

		// Extract {id} from path parameter
		id, ok := ParsePathInt64(w, r, "id", "Transaction ID is required")
		if !ok {
			return
		}

		logger.Debug("Getting transaction by ID", "transaction_id", id)

		// Get transaction from database
		transaction, err := store.GetTransactionByID(r.Context(), id)
		if HandleDBError(w, err, "Transaction not found", "An error has occurred", "Failed to get transaction by ID", "transaction_id", id) {
			return
		}

		// Verify user is a member of the group
		if err := auth.CheckGroupMembership(r.Context(), store, transaction.GroupID, userID); err != nil {
			http.Error(w, "Forbidden: User is not a current group member", http.StatusForbidden)
			return
		}

		// Convert to response format
		transactionResponse := models.TransactionResponse{
			ID:              transaction.ID,
			GroupID:         transaction.GroupID,
			Name:            transaction.Name,
			TransactionDate: transaction.TransactionDate,
			Amount:          transaction.Amount,
			Category:        transaction.Category,
			Note:            transaction.Note,
			ByUser:          transaction.ByUser,
			CreatedAt:       transaction.CreatedAt,
			ModifiedAt:      transaction.ModifiedAt,
		}

		// Send response
		if err := WriteJSONResponseOK(w, transactionResponse); err != nil {
			http.Error(w, "An error has occurred", http.StatusInternalServerError)
			return
		}
	}
}

func createTransaction(store db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get authenticated user ID
		userID, ok := GetAuthenticatedUserID(w, r)
		if !ok {
			return
		}

		// Decode request body
		var createTransactionReq models.CreateTransactionRequest
		if err := DecodeJSONBody(r, &createTransactionReq); err != nil {
			http.Error(w, "Bad request: invalid JSON", http.StatusBadRequest)
			return
		}

		// Validate input
		if createTransactionReq.Name == "" {
			http.Error(w, "Name is required", http.StatusBadRequest)
			return
		}
		if createTransactionReq.GroupID == 0 {
			http.Error(w, "Group ID is required", http.StatusBadRequest)
			return
		}
		if createTransactionReq.ByUser == 0 {
			http.Error(w, "ByUser is required", http.StatusBadRequest)
			return
		}

		// Verify user is a member of the group
		if err := auth.CheckGroupMembership(r.Context(), store, createTransactionReq.GroupID, userID); err != nil {
			http.Error(w, "Forbidden: you must be a member of this group", http.StatusForbidden)
			return
		}

		// Verify ByUser is a group member (it should be a group_member ID)
		groupMember, err := store.GetGroupMemberByID(r.Context(), createTransactionReq.ByUser)
		if err != nil {
			logger.Warn("Group member not found for ByUser", "by_user", createTransactionReq.ByUser)
			http.Error(w, "Group member not found", http.StatusBadRequest)
			return
		}
		if groupMember.GroupID != createTransactionReq.GroupID {
			logger.Warn("Group member does not belong to this group", "by_user", createTransactionReq.ByUser, "group_id", createTransactionReq.GroupID)
			http.Error(w, "Group member does not belong to this group", http.StatusBadRequest)
			return
		}

		logger.Debug("Creating transaction",
			slog.String("name", createTransactionReq.Name),
			slog.Int64("group_id", createTransactionReq.GroupID),
			slog.Int64("user_id", userID),
		)

		// Create transaction in database
		transaction, err := store.CreateTransaction(r.Context(), db.CreateTransactionParams{
			GroupID:         createTransactionReq.GroupID,
			Name:            createTransactionReq.Name,
			TransactionDate: createTransactionReq.TransactionDate,
			Amount:          createTransactionReq.Amount,
			Category:        createTransactionReq.Category,
			Note:            createTransactionReq.Note,
			ByUser:          createTransactionReq.ByUser,
		})
		if HandleDBListError(w, err, "An error has occurred", "Failed to create transaction", "group_id", createTransactionReq.GroupID) {
			return
		}
		logger.Debug("Transaction created successfully",
			slog.Int64("transaction_id", transaction.ID),
			slog.String("name", transaction.Name),
		)

		// Convert to response format
		transactionResponse := models.TransactionResponse{
			ID:              transaction.ID,
			GroupID:         transaction.GroupID,
			Name:            transaction.Name,
			TransactionDate: transaction.TransactionDate,
			Amount:          transaction.Amount,
			Category:        transaction.Category,
			Note:            transaction.Note,
			ByUser:          transaction.ByUser,
			CreatedAt:       transaction.CreatedAt,
			ModifiedAt:      transaction.ModifiedAt,
		}

		// Send response with 201 Created status
		if err := WriteJSONResponseCreated(w, transactionResponse); err != nil {
			http.Error(w, "An error has occurred", http.StatusInternalServerError)
			return
		}
	}
}

func updateTransaction(store db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get authenticated user ID
		userID, ok := GetAuthenticatedUserID(w, r)
		if !ok {
			return
		}

		// Extract {id} from path parameter
		id, ok := ParsePathInt64(w, r, "id", "Transaction ID is required")
		if !ok {
			return
		}

		// Get transaction to find its group
		transaction, err := store.GetTransactionByID(r.Context(), id)
		if HandleDBError(w, err, "Transaction not found", "An error has occurred", "Failed to get transaction by ID", "transaction_id", id) {
			return
		}

		// Verify user is a member of the group
		if err := auth.CheckGroupMembership(r.Context(), store, transaction.GroupID, userID); err != nil {
			http.Error(w, "Forbidden: you must be a member of this group", http.StatusForbidden)
			return
		}

		// Decode request body
		var updateTransactionReq models.UpdateTransactionRequest
		if err := DecodeJSONBody(r, &updateTransactionReq); err != nil {
			http.Error(w, "Bad request: invalid JSON", http.StatusBadRequest)
			return
		}

		// Validate input
		if updateTransactionReq.Name == "" {
			http.Error(w, "Name is required", http.StatusBadRequest)
			return
		}
		if updateTransactionReq.GroupID == 0 {
			http.Error(w, "Group ID is required", http.StatusBadRequest)
			return
		}
		if updateTransactionReq.ByUser == 0 {
			http.Error(w, "ByUser is required", http.StatusBadRequest)
			return
		}

		// Verify ByUser is a group member
		groupMember, err := store.GetGroupMemberByID(r.Context(), updateTransactionReq.ByUser)
		if err != nil {
			logger.Warn("Group member not found for ByUser", "by_user", updateTransactionReq.ByUser)
			http.Error(w, "Group member not found", http.StatusBadRequest)
			return
		}
		if groupMember.GroupID != updateTransactionReq.GroupID {
			logger.Warn("Group member does not belong to this group", "by_user", updateTransactionReq.ByUser, "group_id", updateTransactionReq.GroupID)
			http.Error(w, "Group member does not belong to this group", http.StatusBadRequest)
			return
		}

		logger.Debug("Updating transaction", "transaction_id", id, "user_id", userID)

		// Update transaction in database
		transaction, err = store.UpdateTransaction(r.Context(), db.UpdateTransactionParams{
			ID:              id,
			GroupID:         updateTransactionReq.GroupID,
			Name:            updateTransactionReq.Name,
			TransactionDate: updateTransactionReq.TransactionDate,
			Amount:          updateTransactionReq.Amount,
			Category:        updateTransactionReq.Category,
			Note:            updateTransactionReq.Note,
			ByUser:          updateTransactionReq.ByUser,
		})
		if HandleDBError(w, err, "Transaction not found", "An error has occurred", "Failed to update transaction", "transaction_id", id) {
			return
		}

		// Convert to response format
		transactionResponse := models.TransactionResponse{
			ID:              transaction.ID,
			GroupID:         transaction.GroupID,
			Name:            transaction.Name,
			TransactionDate: transaction.TransactionDate,
			Amount:          transaction.Amount,
			Category:        transaction.Category,
			Note:            transaction.Note,
			ByUser:          transaction.ByUser,
			CreatedAt:       transaction.CreatedAt,
			ModifiedAt:      transaction.ModifiedAt,
		}

		// Send response
		if err := WriteJSONResponseOK(w, transactionResponse); err != nil {
			http.Error(w, "An error has occurred", http.StatusInternalServerError)
			return
		}
	}
}

func deleteTransaction(store db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get authenticated user ID
		userID, ok := GetAuthenticatedUserID(w, r)
		if !ok {
			return
		}

		// Extract {id} from path parameter
		id, ok := ParsePathInt64(w, r, "id", "Transaction ID is required")
		if !ok {
			return
		}

		// Get transaction to find its group
		transaction, err := store.GetTransactionByID(r.Context(), id)
		if HandleDBError(w, err, "Transaction not found", "An error has occurred", "Failed to get transaction by ID", "transaction_id", id) {
			return
		}

		// Verify user is a member of the group
		if err := auth.CheckGroupMembership(r.Context(), store, transaction.GroupID, userID); err != nil {
			http.Error(w, "Forbidden: you must be a member of this group", http.StatusForbidden)
			return
		}

		logger.Debug("Deleting transaction", "transaction_id", id, "user_id", userID)

		// Delete transaction from database
		transaction, err = store.DeleteTransaction(r.Context(), id)
		if HandleDBError(w, err, "Transaction not found", "An error has occurred", "Failed to delete transaction", "transaction_id", id) {
			return
		}

		// Convert to response format
		transactionResponse := models.TransactionResponse{
			ID:              transaction.ID,
			GroupID:         transaction.GroupID,
			Name:            transaction.Name,
			TransactionDate: transaction.TransactionDate,
			Amount:          transaction.Amount,
			Category:        transaction.Category,
			Note:            transaction.Note,
			ByUser:          transaction.ByUser,
			CreatedAt:       transaction.CreatedAt,
			ModifiedAt:      transaction.ModifiedAt,
		}

		// Send response with deleted transaction data
		if err := WriteJSONResponseOK(w, transactionResponse); err != nil {
			http.Error(w, "An error has occurred", http.StatusInternalServerError)
			return
		}
	}
}

// Nested resource handlers

// List splits for transaction
// GET /transactions/{transaction_id}/splits
func getSplitsByTransactionNested(store db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get authenticated user ID
		userID, ok := GetAuthenticatedUserID(w, r)
		if !ok {
			return
		}

		// Extract {transaction_id} from path parameter
		transactionID, ok := ParsePathInt64(w, r, "transaction_id", "Transaction ID is required")
		if !ok {
			return
		}

		// Get transaction to find its group
		transaction, err := store.GetTransactionByID(r.Context(), transactionID)
		if HandleDBError(w, err, "Transaction not found", "An error has occurred", "Failed to get transaction by ID", "transaction_id", transactionID) {
			return
		}

		// Verify user is a member of the group
		if err := auth.CheckGroupMembership(r.Context(), store, transaction.GroupID, userID); err != nil {
			http.Error(w, "Forbidden: User is not a current group member", http.StatusForbidden)
			return
		}

		logger.Debug("Getting splits for transaction", "transaction_id", transactionID)

		splits, err := store.GetSplitsByTransactionID(r.Context(), transactionID)
		if HandleDBListError(w, err, "An error has occurred", "Failed to get splits by transaction ID", "transaction_id", transactionID) {
			return
		}

		splitResponses := make([]models.SplitResponse, len(splits))
		for i, split := range splits {
			splitResponses[i] = models.SplitResponse{
				ID:            split.ID,
				TransactionID: split.TransactionID,
				TxAmount:      split.TxAmount,
				SplitPercent:  split.SplitPercent,
				SplitAmount:   split.SplitAmount,
				SplitUser:     split.SplitUser,
				CreatedAt:     split.CreatedAt,
				ModifiedAt:    split.ModifiedAt,
			}
		}

		count := len(splitResponses)

		listSplitResponse := models.ListSplitResponse{
			Splits: splitResponses,
			Count:  int32(count),
			Limit:  int32(count),
			Offset: 0,
		}

		if err := WriteJSONResponseOK(w, listSplitResponse); err != nil {
			http.Error(w, "An error has occurred", http.StatusInternalServerError)
			return
		}
	}
}

// Create/replace all splits for transaction (batch)
// POST /transactions/{transaction_id}/splits
func createTransactionSplitsBatch(store db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get authenticated user ID
		userID, ok := GetAuthenticatedUserID(w, r)
		if !ok {
			return
		}

		// Extract {transaction_id} from path parameter
		transactionID, ok := ParsePathInt64(w, r, "transaction_id", "Transaction ID is required")
		if !ok {
			return
		}

		// Get transaction to find its group
		transaction, err := store.GetTransactionByID(r.Context(), transactionID)
		if HandleDBError(w, err, "Transaction not found", "An error has occurred", "Failed to get transaction by ID", "transaction_id", transactionID) {
			return
		}

		// Verify user is a member of the group
		if err := auth.CheckGroupMembership(r.Context(), store, transaction.GroupID, userID); err != nil {
			http.Error(w, "Forbidden: User is not a current group member", http.StatusForbidden)
			return
		}

		// Decode request body
		var req struct {
			Splits []models.CreateSplitRequest `json:"splits"`
		}

		if err := DecodeJSONBody(r, &req); err != nil {
			http.Error(w, "Bad request: invalid JSON", http.StatusBadRequest)
			return
		}

		if len(req.Splits) == 0 {
			http.Error(w, "At least one split is required", http.StatusBadRequest)
			return
		}

		// Get group member list to check all users
		groupMembers, err := store.ListGroupMembersByGroupID(r.Context(), db.ListGroupMembersByGroupIDParams{GroupID: transaction.GroupID, Limit: 1000, Offset: 0})
		if HandleDBError(w, err, "Group members not found", "An error has occurred", "Failed to get group members by group ID", "group id", transaction.GroupID) {
			return
		}

		// Validate split group members are in tx group
		if err := ValidateSplitMembersInGroup(req.Splits, groupMembers, transaction.GroupID); err != nil {

			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Validate splits total Tx amount & 100%
		if err := ValidateSplitsTotals(req.Splits, transaction.Amount); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}

		logger.Debug("Creating transaction splits", "transaction_id", transactionID, "split_count", len(req.Splits))

		// Convert to DB params
		dbSplits := make([]db.CreateSplitParams, len(req.Splits))
		for i, split := range req.Splits {
			dbSplits[i] = db.CreateSplitParams{
				TransactionID: transactionID,
				SplitPercent:  split.SplitPercent,
				SplitAmount:   split.SplitAmount,
				SplitUser:     split.SplitUser,
			}
		}

		// Execute transaction to replace all splits
		result, err := store.CreateSplitsTx(r.Context(), db.CreateSplitsTxParams{
			TransactionID: transactionID,
			Splits:        dbSplits,
		})
		if err != nil {
			logger.Error("Failed to create transaction splits", "error", err, "transaction_id", transactionID)
			http.Error(w, fmt.Sprintf("Failed to create splits: %v", err), http.StatusBadRequest)
			return
		}

		// Convert to response format
		splitResponses := make([]models.SplitResponse, len(result.Splits))
		for i, split := range result.Splits {
			splitResponses[i] = models.SplitResponse{
				ID:            split.ID,
				TransactionID: split.TransactionID,
				TxAmount:      split.TxAmount,
				SplitPercent:  split.SplitPercent,
				SplitAmount:   split.SplitAmount,
				SplitUser:     split.SplitUser,
				CreatedAt:     split.CreatedAt,
				ModifiedAt:    split.ModifiedAt,
			}
		}

		response := struct {
			Splits  []models.SplitResponse `json:"splits"`
			Message string                 `json:"message"`
		}{
			Splits:  splitResponses,
			Message: fmt.Sprintf("Successfully created %d splits", len(result.Splits)),
		}

		if err := WriteJSONResponseCreated(w, response); err != nil {
			http.Error(w, "An error has occurred", http.StatusInternalServerError)
			return
		}
	}
}

// Update/replace all splits for transaction (batch)
// PUT /transactions/{transaction_id}/splits or PATCH /transactions/{transaction_id}/splits
func updateTransactionSplitsBatch(store db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get authenticated user ID
		userID, ok := GetAuthenticatedUserID(w, r)
		if !ok {
			return
		}

		// Extract {transaction_id} from path parameter
		transactionID, ok := ParsePathInt64(w, r, "transaction_id", "Transaction ID is required")
		if !ok {
			return
		}

		// Get transaction to find its group
		transaction, err := store.GetTransactionByID(r.Context(), transactionID)
		if HandleDBError(w, err, "Transaction not found", "An error has occurred", "Failed to get transaction by ID", "transaction_id", transactionID) {
			return
		}

		// Verify user is a member of the group
		if err := auth.CheckGroupMembership(r.Context(), store, transaction.GroupID, userID); err != nil {
			http.Error(w, "Forbidden: User is not a current group member", http.StatusForbidden)
			return
		}

		// Decode request body
		var req struct {
			Splits []models.CreateSplitRequest `json:"splits"`
		}

		if err := DecodeJSONBody(r, &req); err != nil {
			http.Error(w, "Bad request: invalid JSON", http.StatusBadRequest)
			return
		}

		if len(req.Splits) == 0 {
			http.Error(w, "At least one split is required", http.StatusBadRequest)
			return
		}

		// Get group member list to check all users
		groupMembers, err := store.ListGroupMembersByGroupID(r.Context(), db.ListGroupMembersByGroupIDParams{GroupID: transaction.GroupID, Limit: 1000, Offset: 0})
		if HandleDBError(w, err, "Group members not found", "An error has occurred", "Failed to get group members by group ID", "group id", transaction.GroupID) {
			return
		}

		// Validate split group members are in tx group
		if err := ValidateSplitMembersInGroup(req.Splits, groupMembers, transaction.GroupID); err != nil {

			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// Validate splits total Tx amount & 100%
		if err := ValidateSplitsTotals(req.Splits, transaction.Amount); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}

		logger.Debug("Updating transaction splits", "transaction_id", transactionID, "new_split_count", len(req.Splits))

		// Convert to DB params
		dbSplits := make([]db.CreateSplitParams, len(req.Splits))
		for i, split := range req.Splits {
			dbSplits[i] = db.CreateSplitParams{
				TransactionID: transactionID,
				SplitPercent:  split.SplitPercent,
				SplitAmount:   split.SplitAmount,
				SplitUser:     split.SplitUser,
			}
		}

		// Execute transaction to replace all splits
		result, err := store.UpdateTransactionSplitsTx(r.Context(), db.UpdateTransactionSplitsTxParams{
			TransactionID: transactionID,
			Splits:        dbSplits,
		})
		if err != nil {
			logger.Error("Failed to update transaction splits", "error", err, "transaction_id", transactionID)
			http.Error(w, fmt.Sprintf("Failed to update splits: %v", err), http.StatusBadRequest)
			return
		}

		// Convert to response format
		splitResponses := make([]models.SplitResponse, len(result.NewSplits))
		for i, split := range result.NewSplits {
			splitResponses[i] = models.SplitResponse{
				ID:            split.ID,
				TransactionID: split.TransactionID,
				TxAmount:      split.TxAmount,
				SplitPercent:  split.SplitPercent,
				SplitAmount:   split.SplitAmount,
				SplitUser:     split.SplitUser,
				CreatedAt:     split.CreatedAt,
				ModifiedAt:    split.ModifiedAt,
			}
		}

		deletedSplitResponses := make([]models.SplitResponse, len(result.DeletedSplits))
		for i, split := range result.DeletedSplits {
			deletedSplitResponses[i] = models.SplitResponse{
				ID:            split.ID,
				TransactionID: split.TransactionID,
				TxAmount:      split.TxAmount,
				SplitPercent:  split.SplitPercent,
				SplitAmount:   split.SplitAmount,
				SplitUser:     split.SplitUser,
				CreatedAt:     split.CreatedAt,
				ModifiedAt:    split.ModifiedAt,
			}
		}

		response := struct {
			DeletedSplits []models.SplitResponse `json:"deleted_splits"`
			NewSplits     []models.SplitResponse `json:"new_splits"`
			Message       string                 `json:"message"`
		}{
			DeletedSplits: deletedSplitResponses,
			NewSplits:     splitResponses,
			Message:       fmt.Sprintf("Successfully replaced %d splits with %d new splits", len(result.DeletedSplits), len(result.NewSplits)),
		}

		if err := WriteJSONResponseOK(w, response); err != nil {
			http.Error(w, "An error has occurred", http.StatusInternalServerError)
			return
		}
	}
}
