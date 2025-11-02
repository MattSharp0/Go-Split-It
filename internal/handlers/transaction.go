package handlers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	db "github.com/MattSharp0/transaction-split-go/db/sqlc"
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

	// Nested resource handlers - RESTful approach for splits
	mux.HandleFunc("GET /{transaction_id}/splits", getSplitsByTransactionNested(q)) // GET: List splits for transaction
	mux.HandleFunc("POST /{transaction_id}/splits", createSplitNested(q))           // POST: Create split for transaction

	return mux
}

func listTransactions(store db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse query parameters
		queryParams := r.URL.Query()

		// Default values
		var listTransactionParams db.ListTransactionsParams
		listTransactionParams.Limit = 100
		listTransactionParams.Offset = 0

		// Parse limit
		if limitStr := queryParams.Get("limit"); limitStr != "" {
			limit, err := strconv.ParseInt(limitStr, 10, 32)
			if err != nil {
				http.Error(w, "Invalid limit parameter", http.StatusBadRequest)
				return
			}
			listTransactionParams.Limit = int32(limit)
		}

		// Parse offset
		if offsetStr := queryParams.Get("offset"); offsetStr != "" {
			offset, err := strconv.ParseInt(offsetStr, 10, 32)
			if err != nil {
				http.Error(w, "Invalid offset parameter", http.StatusBadRequest)
				return
			}
			listTransactionParams.Offset = int32(offset)
		}

		logger.Debug("Listing transactions",
			"limit", listTransactionParams.Limit,
			"offset", listTransactionParams.Offset,
		)

		transactions, err := store.ListTransactions(context.Background(), listTransactionParams)
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

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(listTransactionResponse); err != nil {
			logger.Error("Failed to encode transaction responses", "error", err)
			http.Error(w, "An error has occurred", http.StatusInternalServerError)
			return
		}
	}
}

func getTransactionByID(store db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract {id} from path parameter
		idStr := r.PathValue("id")
		if idStr == "" {
			http.Error(w, "Transaction ID is required", http.StatusBadRequest)
			return
		}

		// Convert string ID to int64
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid transaction ID format", http.StatusBadRequest)
			return
		}

		logger.Debug("Getting transaction by ID", "transaction_id", id)

		// Get transaction from database
		transaction, err := store.GetTransactionByID(context.Background(), id)
		if HandleDBError(w, err, "Transaction not found", "An error has occurred", "Failed to get transaction by ID", "transaction_id", id) {
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
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(transactionResponse); err != nil {
			logger.Error("Failed to encode transaction response", "error", err)
			http.Error(w, "An error has occurred", http.StatusInternalServerError)
			return
		}
	}
}

func createTransaction(store db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Decode request body
		decoder := json.NewDecoder(r.Body)
		defer r.Body.Close()

		var createTransactionReq models.CreateTransactionRequest
		err := decoder.Decode(&createTransactionReq)
		if err != nil {
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

		logger.Debug("Creating transaction",
			slog.String("name", createTransactionReq.Name),
			slog.Int64("group_id", createTransactionReq.GroupID),
		)

		// Create transaction in database
		transaction, err := store.CreateTransaction(context.Background(), db.CreateTransactionParams{
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
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(transactionResponse); err != nil {
			logger.Error("Failed to encode transaction response", "error", err)
			http.Error(w, "An error has occurred", http.StatusInternalServerError)
			return
		}
	}
}

func updateTransaction(store db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract {id} from path parameter
		idStr := r.PathValue("id")
		if idStr == "" {
			http.Error(w, "Transaction ID is required", http.StatusBadRequest)
			return
		}

		// Convert string ID to int64
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid transaction ID format", http.StatusBadRequest)
			return
		}

		// Decode request body
		decoder := json.NewDecoder(r.Body)
		defer r.Body.Close()

		var updateTransactionReq models.UpdateTransactionRequest
		err = decoder.Decode(&updateTransactionReq)
		if err != nil {
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

		logger.Debug("Updating transaction", "transaction_id", id)

		// Update transaction in database
		transaction, err := store.UpdateTransaction(context.Background(), db.UpdateTransactionParams{
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
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(transactionResponse); err != nil {
			logger.Error("Failed to encode transaction response", "error", err)
			http.Error(w, "An error has occurred", http.StatusInternalServerError)
			return
		}
	}
}

func deleteTransaction(store db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract {id} from path parameter
		idStr := r.PathValue("id")
		if idStr == "" {
			http.Error(w, "Transaction ID is required", http.StatusBadRequest)
			return
		}

		// Convert string ID to int64
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid transaction ID format", http.StatusBadRequest)
			return
		}

		logger.Debug("Deleting transaction", "transaction_id", id)

		// Delete transaction from database
		transaction, err := store.DeleteTransaction(context.Background(), id)
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
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(transactionResponse); err != nil {
			logger.Error("Failed to encode transaction response", "error", err)
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
		// Extract {transaction_id} from path parameter
		transactionIDStr := r.PathValue("transaction_id")
		if transactionIDStr == "" {
			http.Error(w, "Transaction ID is required", http.StatusBadRequest)
			return
		}

		// Convert string ID to int64
		transactionID, err := strconv.ParseInt(transactionIDStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid transaction ID format", http.StatusBadRequest)
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

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(listSplitResponse); err != nil {
			logger.Error("Failed to encode split responses", "error", err)
			http.Error(w, "An error has occurred", http.StatusInternalServerError)
			return
		}
	}
}

// Create split for transaction
// POST /transactions/{transaction_id}/splits
func createSplitNested(store db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract {transaction_id} from path parameter
		transactionIDStr := r.PathValue("transaction_id")
		if transactionIDStr == "" {
			http.Error(w, "Transaction ID is required", http.StatusBadRequest)
			return
		}

		// Convert string ID to int64
		transactionID, err := strconv.ParseInt(transactionIDStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid transaction ID format", http.StatusBadRequest)
			return
		}

		// Decode request body
		decoder := json.NewDecoder(r.Body)
		defer r.Body.Close()

		var createSplitReq models.CreateSplitRequest
		err = decoder.Decode(&createSplitReq)
		if err != nil {
			http.Error(w, "Bad request: invalid JSON", http.StatusBadRequest)
			return
		}

		// Override transaction_id from URL
		createSplitReq.TransactionID = transactionID

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
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(splitResponse); err != nil {
			logger.Error("Failed to encode split response", "error", err)
			http.Error(w, "An error has occurred", http.StatusInternalServerError)
			return
		}
	}
}
