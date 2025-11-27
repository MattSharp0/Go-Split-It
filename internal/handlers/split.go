package handlers

import (
	"net/http"

	db "github.com/MattSharp0/transaction-split-go/db/sqlc"
	"github.com/MattSharp0/transaction-split-go/internal/auth"
	"github.com/MattSharp0/transaction-split-go/internal/logger"
	"github.com/MattSharp0/transaction-split-go/internal/models"
	"github.com/MattSharp0/transaction-split-go/internal/server"
)

func SplitRoutes(s *server.Server, q db.Store) *http.ServeMux {
	mux := http.NewServeMux()

	// ID path handlers (read-only)
	mux.HandleFunc("GET /{id}", getSplitByID(q))

	// Root path handlers (read-only - filtered by authenticated user's groups)
	mux.HandleFunc("GET /", listSplits(q))

	// Split create & update operations handled via transactions/{tx_id}/splits batch endpoint

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
