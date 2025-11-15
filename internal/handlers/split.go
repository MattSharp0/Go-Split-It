package handlers

import (
	"fmt"
	"net/http"

	db "github.com/MattSharp0/transaction-split-go/db/sqlc"
	"github.com/MattSharp0/transaction-split-go/internal/auth"
	"github.com/MattSharp0/transaction-split-go/internal/logger"
	"github.com/MattSharp0/transaction-split-go/internal/models"
	"github.com/MattSharp0/transaction-split-go/internal/server"
)

func SplitRoutes(s *server.Server, q db.Store) *http.ServeMux {
	mux := http.NewServeMux()

	// Batch operations
	mux.HandleFunc("POST /batch", createSplitsForTransaction(q)) // Deprecated
	mux.HandleFunc("POST /transaction/{transaction_id}/batch", createTransactionSplits(q))
	mux.HandleFunc("PATCH /transaction/{transaction_id}/batch", updateTransactionSplits(q))
	mux.HandleFunc("PUT /transaction/{transaction_id}/batch", updateTransactionSplits(q))

	// Transaction-specific routes
	mux.HandleFunc("GET /transaction/{transaction_id}", getSplitsByTransactionID(q))

	// User-specific routes
	mux.HandleFunc("GET /user/{user_id}", getSplitsByUser(q))

	// ID path handlers
	mux.HandleFunc("GET /{id}", getSplitByID(q))
	mux.HandleFunc("PUT /{id}", updateSplit(q))
	mux.HandleFunc("PATCH /{id}", updateSplit(q))
	mux.HandleFunc("DELETE /{id}", deleteSplit(q))

	// Root path handlers
	mux.HandleFunc("POST /", createSplit(q))
	mux.HandleFunc("GET /", listSplits(q))

	return mux
}

func listSplits(store db.Store) http.HandlerFunc {
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

		var listSplitParams db.ListSplitsByUserGroupsParams
		listSplitParams.UserID = &userID
		listSplitParams.Limit = limit
		listSplitParams.Offset = offset

		logger.Debug("Listing splits", "limit", listSplitParams.Limit, "offset", listSplitParams.Offset)

		splits, err := store.ListSplitsByUserGroups(r.Context(), listSplitParams)
		if HandleDBListError(w, err, "An error has occurred", "Failed to list splits", "limit", listSplitParams.Limit, "offset", listSplitParams.Offset) {
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
			Limit:  listSplitParams.Limit,
			Offset: listSplitParams.Offset,
		}

		if err := WriteJSONResponseOK(w, listSplitResponse); err != nil {
			http.Error(w, "An error has occurred", http.StatusInternalServerError)
			return
		}
	}
}

func getSplitsByTransactionID(store db.Store) http.HandlerFunc {
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

func getSplitsByUser(store db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get authenticated user ID
		authenticatedUserID, ok := GetAuthenticatedUserID(w, r)
		if !ok {
			return
		}

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

		var listParams db.GetSplitsByUserFilteredParams
		listParams.SplitUser = &userID
		listParams.UserID = &authenticatedUserID
		listParams.Limit = limit
		listParams.Offset = offset

		logger.Debug("Getting splits for user", "user_id", userID, "limit", listParams.Limit, "offset", listParams.Offset)

		splits, err := store.GetSplitsByUserFiltered(r.Context(), listParams)
		if HandleDBListError(w, err, "An error has occurred", "Failed to get splits by user", "user_id", userID) {
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
			Limit:  listParams.Limit,
			Offset: listParams.Offset,
		}

		if err := WriteJSONResponseOK(w, listSplitResponse); err != nil {
			http.Error(w, "An error has occurred", http.StatusInternalServerError)
			return
		}
	}
}

func getSplitByID(store db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get authenticated user ID
		userID, ok := GetAuthenticatedUserID(w, r)
		if !ok {
			return
		}

		// Extract {id} from path parameter
		id, ok := ParsePathInt64(w, r, "id", "Split ID is required")
		if !ok {
			return
		}

		logger.Debug("Getting split by ID", "split_id", id)

		// Get split from database
		split, err := store.GetSplitByID(r.Context(), id)
		if HandleDBError(w, err, "Split not found", "An error has occurred", "Failed to get split by ID", "split_id", id) {
			return
		}

		// Get transaction to find its group
		transaction, err := store.GetTransactionByID(r.Context(), split.TransactionID)
		if HandleDBError(w, err, "Transaction not found", "An error has occurred", "Failed to get transaction", "transaction_id", split.TransactionID) {
			return
		}

		// Verify user is a member of the group
		if err := auth.CheckGroupMembership(r.Context(), store, transaction.GroupID, userID); err != nil {
			http.Error(w, "Forbidden: User is not a current group member", http.StatusForbidden)
			return
		}

		// Convert to response format
		splitResponse := models.SplitResponse{
			ID:            split.ID,
			TransactionID: split.TransactionID,
			TxAmount:      split.TxAmount,
			SplitPercent:  split.SplitPercent,
			SplitAmount:   split.SplitAmount,
			SplitUser:     split.SplitUser,
			CreatedAt:     split.CreatedAt,
			ModifiedAt:    split.ModifiedAt,
		}

		// Send response
		if err := WriteJSONResponseOK(w, splitResponse); err != nil {
			http.Error(w, "An error has occurred", http.StatusInternalServerError)
			return
		}
	}
}

func createSplit(store db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get authenticated user ID
		userID, ok := GetAuthenticatedUserID(w, r)
		if !ok {
			return
		}

		// Decode request body
		var createSplitReq models.CreateSplitRequest
		if err := DecodeJSONBody(r, &createSplitReq); err != nil {
			http.Error(w, "Bad request: invalid JSON", http.StatusBadRequest)
			return
		}

		// Validate input
		if createSplitReq.TransactionID == 0 {
			http.Error(w, "Transaction ID is required", http.StatusBadRequest)
			return
		}

		// Get transaction to find its group
		transaction, err := store.GetTransactionByID(r.Context(), createSplitReq.TransactionID)
		if HandleDBError(w, err, "Transaction not found", "An error has occurred", "Failed to get transaction by ID", "transaction_id", createSplitReq.TransactionID) {
			return
		}

		// Verify user is a member of the group
		if err := auth.CheckGroupMembership(r.Context(), store, transaction.GroupID, userID); err != nil {
			http.Error(w, "Forbidden: you must be a member of this group", http.StatusForbidden)
			return
		}

		// Verify split_user is a group member if provided
		if createSplitReq.SplitUser != nil {
			groupMember, err := store.GetGroupMemberByID(r.Context(), *createSplitReq.SplitUser)
			if err != nil || groupMember.GroupID != transaction.GroupID {
				logger.Warn("Split user is not a member of this group", "split_user", *createSplitReq.SplitUser, "group_id", transaction.GroupID)
				http.Error(w, "Split user must be a member of the transaction's group", http.StatusBadRequest)
				return
			}
		}

		logger.Debug("Creating split", "transaction_id", createSplitReq.TransactionID, "user_id", userID)

		// Create split in database
		split, err := store.CreateSplit(r.Context(), db.CreateSplitParams{
			TransactionID: createSplitReq.TransactionID,
			SplitPercent:  createSplitReq.SplitPercent,
			SplitAmount:   createSplitReq.SplitAmount,
			SplitUser:     createSplitReq.SplitUser,
		})
		if HandleDBListError(w, err, "An error has occurred", "Failed to create split", "transaction_id", createSplitReq.TransactionID) {
			return
		}
		logger.Debug("Split created successfully", "split_id", split.ID, "transaction_id", createSplitReq.TransactionID)

		// Convert to response format
		splitResponse := models.SplitResponse{
			ID:            split.ID,
			TransactionID: split.TransactionID,
			TxAmount:      split.TxAmount,
			SplitPercent:  split.SplitPercent,
			SplitAmount:   split.SplitAmount,
			SplitUser:     split.SplitUser,
			CreatedAt:     split.CreatedAt,
			ModifiedAt:    split.ModifiedAt,
		}

		// Send response with 201 Created status
		if err := WriteJSONResponseCreated(w, splitResponse); err != nil {
			http.Error(w, "An error has occurred", http.StatusInternalServerError)
			return
		}
	}
}

// updateSplit updates an individual split
// WARNING: This can leave the parent transaction in an invalid state where splits don't add up to 100%
// Consider using updateTransactionSplits (PUT /transaction/{id}/batch) instead for safer operations
func updateSplit(store db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get authenticated user ID
		userID, ok := GetAuthenticatedUserID(w, r)
		if !ok {
			return
		}

		// Extract {id} from path parameter
		id, ok := ParsePathInt64(w, r, "id", "Split ID is required")
		if !ok {
			return
		}

		// Get split to find transaction
		split, err := store.GetSplitByID(r.Context(), id)
		if HandleDBError(w, err, "Split not found", "An error has occurred", "Failed to get split by ID", "split_id", id) {
			return
		}

		// Get transaction to find its group
		transaction, err := store.GetTransactionByID(r.Context(), split.TransactionID)
		if HandleDBError(w, err, "Transaction not found", "An error has occurred", "Failed to get transaction", "transaction_id", split.TransactionID) {
			return
		}

		// Verify user is a member of the group
		if err := auth.CheckGroupMembership(r.Context(), store, transaction.GroupID, userID); err != nil {
			http.Error(w, "Forbidden: you must be a member of this group", http.StatusForbidden)
			return
		}

		// Decode request body
		var updateSplitReq models.UpdateSplitRequest
		if err := DecodeJSONBody(r, &updateSplitReq); err != nil {
			http.Error(w, "Bad request: invalid JSON", http.StatusBadRequest)
			return
		}

		// Verify split_user is a group member if provided
		if updateSplitReq.SplitUser != nil {
			groupMember, err := store.GetGroupMemberByID(r.Context(), *updateSplitReq.SplitUser)
			if err != nil || groupMember.GroupID != transaction.GroupID {
				logger.Warn("Split user is not a member of this group", "split_user", *updateSplitReq.SplitUser, "group_id", transaction.GroupID)
				http.Error(w, "Split user must be a member of the transaction's group", http.StatusBadRequest)
				return
			}
		}

		logger.Warn("Updating individual split - this may leave transaction in invalid state", "split_id", id, "user_id", userID)

		// Update split in database
		split, err = store.UpdateSplit(r.Context(), db.UpdateSplitParams{
			ID:           id,
			SplitPercent: updateSplitReq.SplitPercent,
			SplitAmount:  updateSplitReq.SplitAmount,
			SplitUser:    updateSplitReq.SplitUser,
		})
		if HandleDBError(w, err, "Split not found", "An error has occurred", "Failed to update split", "split_id", id) {
			return
		}

		// Convert to response format
		splitResponse := models.SplitResponse{
			ID:            split.ID,
			TransactionID: split.TransactionID,
			TxAmount:      split.TxAmount,
			SplitPercent:  split.SplitPercent,
			SplitAmount:   split.SplitAmount,
			SplitUser:     split.SplitUser,
			CreatedAt:     split.CreatedAt,
			ModifiedAt:    split.ModifiedAt,
		}

		// Send response
		if err := WriteJSONResponseOK(w, splitResponse); err != nil {
			http.Error(w, "An error has occurred", http.StatusInternalServerError)
			return
		}
	}
}

// deleteSplit deletes an individual split
// WARNING: This can leave the parent transaction in an invalid state where splits don't add up to 100%
// Use updateTransactionSplits (PUT /transaction/{id}/batch) instead for safer operations
func deleteSplit(store db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get authenticated user ID
		userID, ok := GetAuthenticatedUserID(w, r)
		if !ok {
			return
		}

		// Extract {id} from path parameter
		id, ok := ParsePathInt64(w, r, "id", "Split ID is required")
		if !ok {
			return
		}

		// Get split to find transaction
		split, err := store.GetSplitByID(r.Context(), id)
		if HandleDBError(w, err, "Split not found", "An error has occurred", "Failed to get split by ID", "split_id", id) {
			return
		}

		// Get transaction to find its group
		transaction, err := store.GetTransactionByID(r.Context(), split.TransactionID)
		if HandleDBError(w, err, "Transaction not found", "An error has occurred", "Failed to get transaction", "transaction_id", split.TransactionID) {
			return
		}

		// Verify user is a member of the group
		if err := auth.CheckGroupMembership(r.Context(), store, transaction.GroupID, userID); err != nil {
			http.Error(w, "Forbidden: you must be a member of this group", http.StatusForbidden)
			return
		}

		logger.Warn("Deleting individual split - this may leave transaction in invalid state", "split_id", id, "user_id", userID)

		// Delete split from database
		split, err = store.DeleteSplit(r.Context(), id)
		if HandleDBError(w, err, "Split not found", "An error has occurred", "Failed to delete split", "split_id", id) {
			return
		}

		// Convert to response format
		splitResponse := models.SplitResponse{
			ID:            split.ID,
			TransactionID: split.TransactionID,
			TxAmount:      split.TxAmount,
			SplitPercent:  split.SplitPercent,
			SplitAmount:   split.SplitAmount,
			SplitUser:     split.SplitUser,
			CreatedAt:     split.CreatedAt,
			ModifiedAt:    split.ModifiedAt,
		}

		// Send response with deleted split data
		if err := WriteJSONResponseOK(w, splitResponse); err != nil {
			http.Error(w, "An error has occurred", http.StatusInternalServerError)
			return
		}
	}
}

// Recommended to use createTransactionSplits instead
func createSplitsForTransaction(store db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get authenticated user ID
		userID, ok := GetAuthenticatedUserID(w, r)
		if !ok {
			return
		}

		var req struct {
			TransactionID int64                       `json:"transaction_id"`
			Splits        []models.CreateSplitRequest `json:"splits"`
		}

		if err := DecodeJSONBody(r, &req); err != nil {
			http.Error(w, "Bad request: invalid JSON", http.StatusBadRequest)
			return
		}

		if req.TransactionID == 0 {
			http.Error(w, "Transaction ID is required", http.StatusBadRequest)
			return
		}

		if len(req.Splits) == 0 {
			http.Error(w, "At least one split is required", http.StatusBadRequest)
			return
		}

		// Get transaction to find its group
		transaction, err := store.GetTransactionByID(r.Context(), req.TransactionID)
		if HandleDBError(w, err, "Transaction not found", "An error has occurred", "Failed to get transaction by ID", "transaction_id", req.TransactionID) {
			return
		}

		// Verify user is a member of the group
		if err := auth.CheckGroupMembership(r.Context(), store, transaction.GroupID, userID); err != nil {
			http.Error(w, "Forbidden: User is not a current group member", http.StatusForbidden)
			return
		}

		// Validate all SplitUser values are group members
		for _, split := range req.Splits {
			if split.SplitUser != nil {
				groupMember, err := store.GetGroupMemberByID(r.Context(), *split.SplitUser)
				if err != nil {
					logger.Warn("Group member not found for SplitUser", "split_user", *split.SplitUser)
					http.Error(w, "Group member not found", http.StatusBadRequest)
					return
				}
				if groupMember.GroupID != transaction.GroupID {
					logger.Warn("Split user is not a member of this group", "split_user", *split.SplitUser, "group_id", transaction.GroupID)
					http.Error(w, "Split user must be a member of the transaction's group", http.StatusBadRequest)
					return
				}
			}
		}

		logger.Debug("Creating multiple splits", "split_count", len(req.Splits), "transaction_id", req.TransactionID)

		// Convert to DB params
		dbSplits := make([]db.CreateSplitParams, len(req.Splits))
		for i, split := range req.Splits {
			dbSplits[i] = db.CreateSplitParams{
				TransactionID: req.TransactionID,
				SplitPercent:  split.SplitPercent,
				SplitAmount:   split.SplitAmount,
				SplitUser:     split.SplitUser,
			}
		}

		// Execute transaction
		result, err := store.CreateSplitsTx(r.Context(), db.CreateSplitsTxParams{
			TransactionID: req.TransactionID,
			Splits:        dbSplits,
		})
		if err != nil {
			logger.Error("Failed to create splits", "error", err) // TODO: check error type to determine if splits not found or unable to create splits
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
			Transaction models.TransactionResponse `json:"transaction"`
			Splits      []models.SplitResponse     `json:"splits"`
		}{
			Transaction: models.TransactionResponse{
				ID:              result.Transaction.ID,
				GroupID:         result.Transaction.GroupID,
				Name:            result.Transaction.Name,
				TransactionDate: result.Transaction.TransactionDate,
				Amount:          result.Transaction.Amount,
				Category:        result.Transaction.Category,
				Note:            result.Transaction.Note,
				ByUser:          result.Transaction.ByUser,
				CreatedAt:       result.Transaction.CreatedAt,
				ModifiedAt:      result.Transaction.ModifiedAt,
			},
			Splits: splitResponses,
		}

		if err := WriteJSONResponseCreated(w, response); err != nil {
			http.Error(w, "An error has occurred", http.StatusInternalServerError)
			return
		}
	}
}

func createTransactionSplits(store db.Store) http.HandlerFunc {
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

		// Validate all SplitUser values are group members
		for _, split := range req.Splits {
			if split.SplitUser != nil {
				groupMember, err := store.GetGroupMemberByID(r.Context(), *split.SplitUser)
				if err != nil {
					logger.Warn("Group member not found for SplitUser", "split_user", *split.SplitUser)
					http.Error(w, "Group member not found", http.StatusBadRequest)
					return
				}
				if groupMember.GroupID != transaction.GroupID {
					logger.Warn("Split user is not a member of this group", "split_user", *split.SplitUser, "group_id", transaction.GroupID)
					http.Error(w, "Split user must be a member of the transaction's group", http.StatusBadRequest)
					return
				}
			}
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
		result, err := store.CreateSplitsTx(r.Context(), db.CreateSplitsTxParams{
			TransactionID: transactionID,
			Splits:        dbSplits,
		})
		if err != nil {
			logger.Error("Failed to create transaction splits", "error", err, "transaction_id", transactionID) // TODO: check error type to determine if splits not found or unable to update splits
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

func updateTransactionSplits(store db.Store) http.HandlerFunc {
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

		// Validate all SplitUser values are group members
		for _, split := range req.Splits {
			if split.SplitUser != nil {
				groupMember, err := store.GetGroupMemberByID(r.Context(), *split.SplitUser)
				if err != nil {
					logger.Warn("Group member not found for SplitUser", "split_user", *split.SplitUser)
					http.Error(w, "Group member not found", http.StatusBadRequest)
					return
				}
				if groupMember.GroupID != transaction.GroupID {
					logger.Warn("Split user is not a member of this group", "split_user", *split.SplitUser, "group_id", transaction.GroupID)
					http.Error(w, "Split user must be a member of the transaction's group", http.StatusBadRequest)
					return
				}
			}
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
			logger.Error("Failed to update transaction splits", "error", err, "transaction_id", transactionID) // TODO: check error type to determine if splits not found or unable to update splits
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
