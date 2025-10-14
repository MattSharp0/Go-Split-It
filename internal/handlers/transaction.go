package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	db "github.com/MattSharp0/transaction-split-go/db/sqlc"
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

	// User-specific routes
	mux.HandleFunc("GET /user/{user_id}", getTransactionsByUser(q)) // GET: List transactions by user

	// Group-specific routes
	mux.HandleFunc("GET /group/{group_id}", getTransactionsByGroup(q)) // GET: List transactions by group

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

		log.Printf("List Transaction parameters: %v", listTransactionParams)

		transactions, err := store.ListTransactions(context.Background(), listTransactionParams)
		if err != nil {
			log.Printf("ListTransactions returned an error: %v", err)
			http.Error(w, "An error has occurred", http.StatusInternalServerError)
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
			log.Printf("Error encoding transaction responses: %v", err)
			http.Error(w, "An error has occurred", http.StatusInternalServerError)
			return
		}
	}
}

func getTransactionsByUser(store db.Store) http.HandlerFunc {
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
		var listParams db.GetTransactionsByUserParams
		listParams.ByUser = userID
		listParams.Limit = 100
		listParams.Offset = 0

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

		log.Printf("List Transactions for user %d, parameters: %v", userID, listParams)

		transactions, err := store.GetTransactionsByUser(context.Background(), listParams)
		if err != nil {
			log.Printf("GetTransactionsByUser returned an error: %v", err)
			http.Error(w, "An error has occurred", http.StatusInternalServerError)
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
			log.Printf("Error encoding transaction responses: %v", err)
			http.Error(w, "An error has occurred", http.StatusInternalServerError)
			return
		}
	}
}

func getTransactionsByGroup(store db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract {group_id} from path parameter
		groupIDStr := r.PathValue("group_id")
		if groupIDStr == "" {
			http.Error(w, "Group ID is required", http.StatusBadRequest)
			return
		}

		// Convert string ID to int64
		groupID, err := strconv.ParseInt(groupIDStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid group ID format", http.StatusBadRequest)
			return
		}

		// Parse query parameters
		queryParams := r.URL.Query()

		// Default values
		var listParams db.GetTransactionsByGroupInPeriodParams
		listParams.GroupID = groupID
		listParams.Limit = 100
		listParams.Offset = 0

		// Parse start_date (required)
		startDateStr := queryParams.Get("start_date")
		if startDateStr == "" {
			http.Error(w, "start_date parameter is required in ISO 8601 format", http.StatusBadRequest)
			return
		}
		startDate, err := time.Parse("2006-01-02", startDateStr)
		if err != nil {
			http.Error(w, "Invalid start_date format, use YYYY-MM-DD", http.StatusBadRequest)
			return
		}
		listParams.StartDate = startDate

		// Parse end_date (required)
		endDateStr := queryParams.Get("end_date")
		if endDateStr == "" {
			http.Error(w, "end_date parameter is required in ISO 8601 format", http.StatusBadRequest)
			return
		}
		endDate, err := time.Parse("2006-01-02", endDateStr)
		if err != nil {
			http.Error(w, "Invalid end_date format, use YYYY-MM-DD", http.StatusBadRequest)
			return
		}
		listParams.EndDate = endDate

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

		log.Printf("List Transactions for group %d, parameters: %v", groupID, listParams)

		transactions, err := store.GetTransactionsByGroupInPeriod(context.Background(), listParams)
		if err != nil {
			log.Printf("GetTransactionsByGroupInPeriod returned an error: %v", err)
			http.Error(w, "An error has occurred", http.StatusInternalServerError)
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
			log.Printf("Error encoding transaction responses: %v", err)
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

		log.Printf("Getting transaction with ID: %d", id)

		// Get transaction from database
		transaction, err := store.GetTransactionByID(context.Background(), id)
		if err != nil {
			log.Printf("GetTransactionByID (%v) returned an error: %v", id, err)
			http.Error(w, "Transaction not found", http.StatusNotFound)
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
			log.Printf("Error encoding transaction response: %v", err)
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

		log.Printf("Creating transaction: %s for group: %d", createTransactionReq.Name, createTransactionReq.GroupID)

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
		if err != nil {
			log.Printf("CreateTransaction returned an error: %v", err)
			http.Error(w, "An error has occurred", http.StatusInternalServerError)
			return
		}
		log.Printf("Created transaction '%s' with ID: %d", transaction.Name, transaction.ID)

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
			log.Printf("Error encoding transaction response: %v", err)
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

		log.Printf("Updating transaction with ID: %d", id)

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
		if err != nil {
			log.Printf("UpdateTransaction returned an error: %v", err)
			http.Error(w, "Transaction not found or unable to update", http.StatusNotFound)
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
			log.Printf("Error encoding transaction response: %v", err)
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

		log.Printf("Deleting transaction with ID: %d", id)

		// Delete transaction from database
		transaction, err := store.DeleteTransaction(context.Background(), id)
		if err != nil {
			log.Printf("DeleteTransaction returned an error: %v", err)
			http.Error(w, "Transaction not found or unable to delete", http.StatusNotFound)
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
			log.Printf("Error encoding transaction response: %v", err)
			http.Error(w, "An error has occurred", http.StatusInternalServerError)
			return
		}
	}
}
