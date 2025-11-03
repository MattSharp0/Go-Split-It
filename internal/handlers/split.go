package handlers

import (
	"context"
	"fmt"
	"net/http"

	db "github.com/MattSharp0/transaction-split-go/db/sqlc"
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
		// Parse query parameters
		limit, offset, err := ParseLimitOffset(r)
		if err != nil {
			http.Error(w, "Invalid parameter: "+err.Error(), http.StatusBadRequest)
			return
		}

		var listSplitParams db.ListSplitsParams
		listSplitParams.Limit = limit
		listSplitParams.Offset = offset

		logger.Debug("Listing splits", "limit", listSplitParams.Limit, "offset", listSplitParams.Offset)

		splits, err := store.ListSplits(context.Background(), listSplitParams)
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
		// Extract {transaction_id} from path parameter
		transactionID, ok := ParsePathInt64(w, r, "transaction_id", "Transaction ID is required")
		if !ok {
			return
		}

		logger.Debug("Getting splits for transaction", "transaction_id", transactionID)

		splits, err := store.GetSplitsByTransactionID(context.Background(), transactionID)
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

		var listParams db.GetSplitsByUserParams
		listParams.SplitUser = &userID
		listParams.Limit = limit
		listParams.Offset = offset

		logger.Debug("Getting splits for user", "user_id", userID, "limit", listParams.Limit, "offset", listParams.Offset)

		splits, err := store.GetSplitsByUser(context.Background(), listParams)
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
		// Extract {id} from path parameter
		id, ok := ParsePathInt64(w, r, "id", "Split ID is required")
		if !ok {
			return
		}

		logger.Debug("Getting split by ID", "split_id", id)

		// Get split from database
		split, err := store.GetSplitByID(context.Background(), id)
		if HandleDBError(w, err, "Split not found", "An error has occurred", "Failed to get split by ID", "split_id", id) {
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

		logger.Debug("Creating split", "transaction_id", createSplitReq.TransactionID)

		// Create split in database
		split, err := store.CreateSplit(context.Background(), db.CreateSplitParams{
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
		// Extract {id} from path parameter
		id, ok := ParsePathInt64(w, r, "id", "Split ID is required")
		if !ok {
			return
		}

		// Decode request body
		var updateSplitReq models.UpdateSplitRequest
		if err := DecodeJSONBody(r, &updateSplitReq); err != nil {
			http.Error(w, "Bad request: invalid JSON", http.StatusBadRequest)
			return
		}

		logger.Warn("Updating individual split - this may leave transaction in invalid state", "split_id", id)

		// Update split in database
		split, err := store.UpdateSplit(context.Background(), db.UpdateSplitParams{
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
		// Extract {id} from path parameter
		id, ok := ParsePathInt64(w, r, "id", "Split ID is required")
		if !ok {
			return
		}

		logger.Warn("Deleting individual split - this may leave transaction in invalid state", "split_id", id)

		// Delete split from database
		split, err := store.DeleteSplit(context.Background(), id)
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
		result, err := store.CreateSplitsTx(context.Background(), db.CreateSplitsTxParams{
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
		// Extract {transaction_id} from path parameter
		transactionID, ok := ParsePathInt64(w, r, "transaction_id", "Transaction ID is required")
		if !ok {
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
		result, err := store.CreateSplitsTx(context.Background(), db.CreateSplitsTxParams{
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
		// Extract {transaction_id} from path parameter
		transactionID, ok := ParsePathInt64(w, r, "transaction_id", "Transaction ID is required")
		if !ok {
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
		result, err := store.UpdateTransactionSplitsTx(context.Background(), db.UpdateTransactionSplitsTxParams{
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
