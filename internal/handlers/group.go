package handlers

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	db "github.com/MattSharp0/transaction-split-go/db/sqlc"
	"github.com/MattSharp0/transaction-split-go/internal/logger"
	"github.com/MattSharp0/transaction-split-go/internal/models"
	"github.com/MattSharp0/transaction-split-go/internal/server"
	"github.com/MattSharp0/transaction-split-go/internal/services"
)

func GroupRoutes(s *server.Server, q db.Store) *http.ServeMux {
	mux := http.NewServeMux()

	// Batch Path Handlers
	mux.HandleFunc("POST /{group_id}/members/batch", createGroupMembersForGroup(q))   // POST: Create group members for group
	mux.HandleFunc("PATCH /{group_id}/members/batch", updateGroupMembersForGroup(q))  // PATCH: Update group members for group
	mux.HandleFunc("PUT /{group_id}/members/batch", updateGroupMembersForGroup(q))    // PUT: Update group members for group
	mux.HandleFunc("DELETE /{group_id}/members/batch", deleteGroupMembersForGroup(q)) // DELETE: Delete group members for group

	// Root path handlers
	mux.HandleFunc("POST /", createGroup(q)) // POST: Create group
	mux.HandleFunc("GET /", listGroups(q))   // GET: List groups

	// ID path handlers
	mux.HandleFunc("GET /{id}", getGroupByID(q))   // GET: Get group by ID
	mux.HandleFunc("PUT /{id}", updateGroup(q))    // PUT: Update group
	mux.HandleFunc("PATCH /{id}", updateGroup(q))  // PATCH: Update group
	mux.HandleFunc("DELETE /{id}", deleteGroup(q)) // DELETE: Delete group

	// Nested resource handlers
	mux.HandleFunc("GET /{group_id}/members", listGroupMembers(q))         // GET: List group members
	mux.HandleFunc("POST /{group_id}/members", createGroupMemberNested(q)) // POST: Create group member

	mux.HandleFunc("GET /{group_id}/transactions", getTransactionsByGroupNested(q)) // GET: List group transactions
	mux.HandleFunc("POST /{group_id}/transactions", createTransactionNested(q))     // POST: Create transaction in group

	mux.HandleFunc("GET /{group_id}/balances", getGroupBalances(q)) // GET: Get group balances

	return mux
}

func listGroups(store db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse query parameters
		limit, offset, err := ParseLimitOffset(r)
		if err != nil {
			http.Error(w, "Invalid parameter: "+err.Error(), http.StatusBadRequest)
			return
		}

		var listGroupParams db.ListGroupsParams
		listGroupParams.Limit = limit
		listGroupParams.Offset = offset

		logger.Debug("Listing groups", "limit", listGroupParams.Limit, "offset", listGroupParams.Offset)

		groups, err := store.ListGroups(context.Background(), listGroupParams)
		if HandleDBListError(w, err, "An error has occurred", "Failed to list groups", "limit", listGroupParams.Limit, "offset", listGroupParams.Offset) {
			return
		}

		groupResponses := make([]models.GroupResponse, len(groups))
		for i, group := range groups {
			groupResponses[i] = models.GroupResponse{
				ID:   group.ID,
				Name: group.Name,
			}
		}

		count := len(groupResponses)

		listGroupResponse := models.ListGroupResponse{
			Groups: groupResponses,
			Count:  int32(count),
			Limit:  listGroupParams.Limit,
			Offset: listGroupParams.Offset,
		}

		if err := WriteJSONResponseOK(w, listGroupResponse); err != nil {
			http.Error(w, "An error has occurred", http.StatusInternalServerError)
			return
		}
	}
}

func getGroupByID(store db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract {id} from path parameter
		id, ok := ParsePathInt64(w, r, "id", "Group ID is required")
		if !ok {
			return
		}

		logger.Debug("Getting group by ID", "group_id", id)

		// Get group from database
		group, err := store.GetGroupByID(context.Background(), id)
		if HandleDBError(w, err, "Group not found", "An error has occurred", "Failed to get group by ID", "group_id", id) {
			return
		}

		// Convert to response format
		groupResponse := models.GroupResponse{
			ID:   group.ID,
			Name: group.Name,
		}

		// Send response
		if err := WriteJSONResponseOK(w, groupResponse); err != nil {
			http.Error(w, "An error has occurred", http.StatusInternalServerError)
			return
		}
	}
}

func createGroup(store db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Decode request body
		var createGroupReq models.CreateGroupRequest
		if err := DecodeJSONBody(r, &createGroupReq); err != nil {
			http.Error(w, "Bad request: invalid JSON", http.StatusBadRequest)
			return
		}

		// Validate input
		if createGroupReq.Name == "" {
			http.Error(w, "Name is required", http.StatusBadRequest)
			return
		}

		logger.Debug("Creating group", slog.String("name", createGroupReq.Name))

		// Create group in database
		group, err := store.CreateGroup(context.Background(), createGroupReq.Name)
		if HandleDBListError(w, err, "An error has occurred", "Failed to create group", "name", createGroupReq.Name) {
			return
		}
		logger.Debug("Group created successfully", slog.Int64("group_id", group.ID), slog.String("name", group.Name))

		// Convert to response format
		groupResponse := models.GroupResponse{
			ID:   group.ID,
			Name: group.Name,
		}

		// Send response with 201 Created status
		if err := WriteJSONResponseCreated(w, groupResponse); err != nil {
			http.Error(w, "An error has occurred", http.StatusInternalServerError)
			return
		}
	}
}

func updateGroup(store db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract {id} from path parameter
		id, ok := ParsePathInt64(w, r, "id", "Group ID is required")
		if !ok {
			return
		}

		// Decode request body
		var updateGroupReq models.UpdateGroupRequest
		if err := DecodeJSONBody(r, &updateGroupReq); err != nil {
			http.Error(w, "Bad request: invalid JSON", http.StatusBadRequest)
			return
		}

		// Validate input
		if updateGroupReq.Name == "" {
			http.Error(w, "Name is required", http.StatusBadRequest)
			return
		}

		logger.Debug("Updating group", "group_id", id, "new_name", updateGroupReq.Name)

		// Update group in database
		group, err := store.UpdateGroup(context.Background(), db.UpdateGroupParams{
			ID:   id,
			Name: updateGroupReq.Name,
		})
		if HandleDBError(w, err, "Group not found", "An error has occurred", "Failed to update group", "group_id", id) {
			return
		}

		// Convert to response format
		groupResponse := models.GroupResponse{
			ID:   group.ID,
			Name: group.Name,
		}

		// Send response
		if err := WriteJSONResponseOK(w, groupResponse); err != nil {
			http.Error(w, "An error has occurred", http.StatusInternalServerError)
			return
		}
	}
}

func deleteGroup(store db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract {id} from path parameter
		id, ok := ParsePathInt64(w, r, "id", "Group ID is required")
		if !ok {
			return
		}

		logger.Debug("Deleting group", "group_id", id)

		// Delete group from database
		group, err := store.DeleteGroup(context.Background(), id)
		if HandleDBError(w, err, "Group not found", "An error has occurred", "Failed to delete group", "group_id", id) {
			return
		}

		// Convert to response format
		groupResponse := models.GroupResponse{
			ID:   group.ID,
			Name: group.Name,
		}

		// Send response with deleted group data
		if err := WriteJSONResponseOK(w, groupResponse); err != nil {
			http.Error(w, "An error has occurred", http.StatusInternalServerError)
			return
		}
	}
}

// Nested resource handlers

// List group members in group
// GET /groups/{group_id}/members
func listGroupMembers(store db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract {group_id} from path parameter
		groupID, ok := ParsePathInt64(w, r, "group_id", "Group ID is required")
		if !ok {
			return
		}

		// Parse query parameters
		limit, offset, err := ParseLimitOffset(r)
		if err != nil {
			http.Error(w, "Invalid parameter: "+err.Error(), http.StatusBadRequest)
			return
		}

		var listParams db.ListGroupMembersByGroupIDParams
		listParams.GroupID = groupID
		listParams.Limit = limit
		listParams.Offset = offset

		logger.Debug("Listing group members", "group_id", groupID, "limit", listParams.Limit, "offset", listParams.Offset)

		groupMembers, err := store.ListGroupMembersByGroupID(context.Background(), listParams)
		if HandleDBListError(w, err, "An error has occurred", "Failed to list group members", "group_id", groupID) {
			return
		}

		groupMemberResponses := make([]models.GroupMemberResponse, len(groupMembers))
		for i, gm := range groupMembers {
			groupMemberResponses[i] = models.GroupMemberResponse{
				ID:         gm.ID,
				GroupID:    gm.GroupID,
				GroupName:  gm.GroupName,
				MemberName: gm.MemberName,
				UserID:     gm.UserID,
				UserName:   gm.UserName,
				CreatedAt:  gm.CreatedAt,
			}
		}

		count := len(groupMemberResponses)

		listGroupMemberResponse := models.ListGroupMemberResponse{
			GroupMembers: groupMemberResponses,
			Count:        int32(count),
			Limit:        listParams.Limit,
			Offset:       listParams.Offset,
		}

		if err := WriteJSONResponseOK(w, listGroupMemberResponse); err != nil {
			http.Error(w, "An error has occurred", http.StatusInternalServerError)
			return
		}
	}
}

// Create group member in group
// POST /groups/{group_id}/members
func createGroupMemberNested(store db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract {group_id} from path parameter
		groupID, ok := ParsePathInt64(w, r, "group_id", "Group ID is required")
		if !ok {
			return
		}

		// Decode request body
		var createGroupMemberReq models.CreateGroupMemberRequest
		if err := DecodeJSONBody(r, &createGroupMemberReq); err != nil {
			http.Error(w, "Bad request: invalid JSON", http.StatusBadRequest)
			return
		}

		// Override group_id from URL
		createGroupMemberReq.GroupID = groupID

		logger.Debug("Creating group member", "group_id", createGroupMemberReq.GroupID, "user_id", createGroupMemberReq.UserID)

		// Create group member in database
		groupMember, err := store.CreateGroupMember(context.Background(), db.CreateGroupMemberParams{
			GroupID: createGroupMemberReq.GroupID,
			UserID:  createGroupMemberReq.UserID,
		})
		if HandleDBListError(w, err, "An error has occurred", "Failed to create group member", "group_id", createGroupMemberReq.GroupID, "user_id", createGroupMemberReq.UserID) {
			return
		}
		logger.Debug("Group member created successfully", "group_member_id", groupMember.ID)

		// Convert to response format
		groupMemberResponse := models.GroupMemberResponse{
			ID:         groupMember.ID,
			GroupID:    groupMember.GroupID,
			MemberName: groupMember.MemberName,
			UserID:     groupMember.UserID,
			CreatedAt:  groupMember.CreatedAt,
		}

		// Send response with 201 Created status
		if err := WriteJSONResponseCreated(w, groupMemberResponse); err != nil {
			http.Error(w, "An error has occurred", http.StatusInternalServerError)
			return
		}
	}
}

// List transactions in group
// GET /groups/{group_id}/transactions
func getTransactionsByGroupNested(store db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract {group_id} from path parameter
		groupID, ok := ParsePathInt64(w, r, "group_id", "Group ID is required")
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
		var listParams db.GetTransactionsByGroupInPeriodParams
		listParams.GroupID = groupID

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

		logger.Debug("Listing transactions for group", "group_id", groupID, "start_date", listParams.StartDate, "end_date", listParams.EndDate, "limit", listParams.Limit, "offset", listParams.Offset)

		transactions, err := store.GetTransactionsByGroupInPeriod(context.Background(), listParams)
		if HandleDBListError(w, err, "An error has occurred", "Failed to get transactions by group", "group_id", groupID) {
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
	}
}

// Create transaction in group
// POST /groups/{group_id}/transactions
func createTransactionNested(store db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract {group_id} from path parameter
		groupID, ok := ParsePathInt64(w, r, "group_id", "Group ID is required")
		if !ok {
			return
		}

		// Decode request body
		var createTransactionReq models.CreateTransactionRequest
		if err := DecodeJSONBody(r, &createTransactionReq); err != nil {
			http.Error(w, "Bad request: invalid JSON", http.StatusBadRequest)
			return
		}

		// Override group_id from URL
		createTransactionReq.GroupID = groupID

		// Validate input
		if createTransactionReq.Name == "" {
			http.Error(w, "Name is required", http.StatusBadRequest)
			return
		}
		if createTransactionReq.ByUser == 0 {
			http.Error(w, "ByUser is required", http.StatusBadRequest)
			return
		}

		logger.Debug("Creating transaction", slog.String("name", createTransactionReq.Name), slog.Int64("group_id", createTransactionReq.GroupID))

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
		logger.Debug("Transaction created successfully", slog.Int64("transaction_id", transaction.ID), slog.String("name", transaction.Name))

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

// Get group balances
// GET /groups/{group_id}/balances
func getGroupBalances(store db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract {group_id} from path parameter
		groupID, ok := ParsePathInt64(w, r, "group_id", "Group ID is required")
		if !ok {
			return
		}

		logger.Debug("Getting balances for group", "group_id", groupID)

		balances, err := store.GroupBalances(context.Background(), groupID)
		if HandleDBListError(w, err, "An error has occurred", "Failed to get group balances", "group_id", groupID) {
			return
		}

		netBalances, err := store.GroupBalancesNet(context.Background(), groupID)
		if HandleDBListError(w, err, "An error has occurred", "Failed to get group net balances", "group_id", groupID) {
			return
		}

		netBalancesForSimplification := make([]*models.NetBalance, len(netBalances))
		for i, nb := range netBalances {
			userID := int64(0)
			if nb.UserID != nil {
				userID = *nb.UserID
			}
			netBalancesForSimplification[i] = &models.NetBalance{
				UserID:     userID,
				NetBalance: nb.NetBalance,
			}
		}

		simplifiedBalances, err := services.SimplifyDebts(netBalancesForSimplification)
		if err != nil {
			logger.Error("Failed to simplify debts", "error", err, "group_id", groupID)
			http.Error(w, "An error has occurred", http.StatusInternalServerError)
			return
		}

		balanceResponses := make([]models.BalanceResponse, len(balances))
		for i, b := range balances {
			creditor := ""
			if b.Creditor != nil {
				creditor = *b.Creditor
			}
			debtor := ""
			if b.Debtor != nil {
				debtor = *b.Debtor
			}
			creditorID := int64(0)
			if b.CreditorID != nil {
				creditorID = *b.CreditorID
			}
			debtorID := int64(0)
			if b.DebtorID != nil {
				debtorID = *b.DebtorID
			}
			balanceResponses[i] = models.BalanceResponse{
				CreditorID: creditorID,
				Creditor:   creditor,
				DebtorID:   debtorID,
				Debtor:     debtor,
				TotalOwed:  b.TotalOwed,
			}
		}

		netBalanceResponses := make([]models.NetBalanceResponse, len(netBalances))
		for i, nb := range netBalances {
			userID := int64(0)
			if nb.UserID != nil {
				userID = *nb.UserID
			}

			memberName := ""
			if nb.UserName != nil {
				memberName = *nb.UserName
			}
			netBalanceResponses[i] = models.NetBalanceResponse{
				UserID:     userID,
				MemberName: memberName,
				NetBalance: nb.NetBalance,
			}
		}

		simplifiedResponses := make([]models.SimplifiedPaymentsResponse, len(simplifiedBalances))
		for i, sb := range simplifiedBalances {
			simplifiedResponses[i] = models.SimplifiedPaymentsResponse{
				FromUserID: sb.FromUserID,
				ToUserID:   sb.ToUserID,
				Amount:     sb.Amount,
			}
		}

		response := models.GroupBalancesResponse{
			GroupID:                 groupID,
			Balances:                balanceResponses,
			NetBalances:             netBalanceResponses,
			SimplifiedPayments:      simplifiedResponses,
			Count:                   int32(len(balanceResponses)),
			NetCount:                int32(len(netBalanceResponses)),
			SimplifiedPaymentsCount: int32(len(simplifiedResponses)),
		}

		if err := WriteJSONResponseOK(w, response); err != nil {
			http.Error(w, "An error has occurred", http.StatusInternalServerError)
			return
		}
	}
}

// Batch operation handlers

// Create group members for group (batch)
// POST /groups/{group_id}/members/batch
func createGroupMembersForGroup(store db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract {group_id} from path parameter
		groupID, ok := ParsePathInt64(w, r, "group_id", "Group ID is required")
		if !ok {
			return
		}

		// Decode request body
		var batchReq models.BatchCreateGroupMemberRequest
		if err := DecodeJSONBody(r, &batchReq); err != nil {
			http.Error(w, "Bad request: invalid JSON", http.StatusBadRequest)
			return
		}

		// Validate input
		if len(batchReq.Members) == 0 {
			http.Error(w, "At least one member is required", http.StatusBadRequest)
			return
		}

		// Convert request to transaction params
		groupMembers := make([]db.CreateGroupMemberParams, len(batchReq.Members))
		for i, member := range batchReq.Members {
			groupMembers[i] = db.CreateGroupMemberParams{
				GroupID: groupID,
				UserID:  member.UserID,
			}
		}

		logger.Debug("Creating group members in batch", "group_id", groupID, "count", len(groupMembers))

		// Create group members using transaction
		result, err := store.CreateGroupMembersTx(context.Background(), db.CreateGroupMemberTxParams{
			GroupID:      groupID,
			GroupMembers: groupMembers,
		})
		if err != nil {
			logger.Error("Failed to create group members", "error", err, "group_id", groupID)
			http.Error(w, "An error has occurred", http.StatusInternalServerError)
			return
		}

		logger.Debug("Group members created successfully", "group_id", groupID, "count", len(result.GroupMembers))

		// Convert to response format
		groupMemberResponses := make([]models.GroupMemberResponse, len(result.GroupMembers))
		for i, gm := range result.GroupMembers {
			groupMemberResponses[i] = models.GroupMemberResponse{
				ID:         gm.ID,
				GroupID:    gm.GroupID,
				MemberName: gm.MemberName,
				UserID:     gm.UserID,
				CreatedAt:  gm.CreatedAt,
			}
		}

		response := models.BatchCreateGroupMemberResponse{
			Group: models.GroupResponse{
				ID:   result.Group.ID,
				Name: result.Group.Name,
			},
			GroupMembers: groupMemberResponses,
			Count:        int32(len(groupMemberResponses)),
		}

		// Send response with 201 Created status
		if err := WriteJSONResponseCreated(w, response); err != nil {
			http.Error(w, "An error has occurred", http.StatusInternalServerError)
			return
		}
	}
}

// Update group members for group (batch) - replaces all members
// PUT/PATCH /groups/{group_id}/members/batch
func updateGroupMembersForGroup(store db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract {group_id} from path parameter
		groupID, ok := ParsePathInt64(w, r, "group_id", "Group ID is required")
		if !ok {
			return
		}

		// Decode request body
		var batchReq models.BatchUpdateGroupMemberRequest
		if err := DecodeJSONBody(r, &batchReq); err != nil {
			http.Error(w, "Bad request: invalid JSON", http.StatusBadRequest)
			return
		}

		// Convert request to transaction params
		groupMembers := make([]db.CreateGroupMemberParams, len(batchReq.Members))
		for i, member := range batchReq.Members {
			groupMembers[i] = db.CreateGroupMemberParams{
				GroupID: groupID,
				UserID:  member.UserID,
			}
		}

		logger.Debug("Updating group members in batch", "group_id", groupID, "new_count", len(groupMembers))

		// Update group members using transaction (replaces all)
		result, err := store.UpdateGroupMembersTx(context.Background(), db.UpdateGroupMemberTxParams{
			GroupID:      groupID,
			GroupMembers: groupMembers,
		})
		if err != nil {
			logger.Error("Failed to update group members", "error", err, "group_id", groupID)
			http.Error(w, "An error has occurred", http.StatusInternalServerError)
			return
		}

		logger.Debug("Group members updated successfully", "group_id", groupID, "deleted_count", len(result.DeletedMembers), "new_count", len(result.NewMembers))

		// Convert deleted members to response format
		deletedResponses := make([]models.GroupMemberResponse, len(result.DeletedMembers))
		for i, gm := range result.DeletedMembers {
			deletedResponses[i] = models.GroupMemberResponse{
				ID:         gm.ID,
				GroupID:    gm.GroupID,
				MemberName: gm.MemberName,
				UserID:     gm.UserID,
				CreatedAt:  gm.CreatedAt,
			}
		}

		// Convert new members to response format
		newResponses := make([]models.GroupMemberResponse, len(result.NewMembers))
		for i, gm := range result.NewMembers {
			newResponses[i] = models.GroupMemberResponse{
				ID:         gm.ID,
				GroupID:    gm.GroupID,
				MemberName: gm.MemberName,
				UserID:     gm.UserID,
				CreatedAt:  gm.CreatedAt,
			}
		}

		response := models.BatchUpdateGroupMemberResponse{
			Group: models.GroupResponse{
				ID:   result.Group.ID,
				Name: result.Group.Name,
			},
			DeletedMembers: deletedResponses,
			NewMembers:     newResponses,
			DeletedCount:   int32(len(deletedResponses)),
			NewCount:       int32(len(newResponses)),
		}

		// Send response
		if err := WriteJSONResponseOK(w, response); err != nil {
			http.Error(w, "An error has occurred", http.StatusInternalServerError)
			return
		}
	}
}

// Delete group members for group (batch) - deletes all members
// DELETE /groups/{group_id}/members/batch
func deleteGroupMembersForGroup(store db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract {group_id} from path parameter
		groupID, ok := ParsePathInt64(w, r, "group_id", "Group ID is required")
		if !ok {
			return
		}

		logger.Debug("Deleting group members in batch", "group_id", groupID)

		// Delete all group members using transaction
		err := store.DeleteGroupMembersTx(context.Background(), groupID)
		if err != nil {
			logger.Error("Failed to delete group members", "error", err, "group_id", groupID)
			http.Error(w, "An error has occurred", http.StatusInternalServerError)
			return
		}

		logger.Debug("Group members deleted successfully", "group_id", groupID)

		response := models.BatchDeleteGroupMemberResponse{
			GroupID: groupID,
			Message: "All group members deleted successfully",
		}

		// Send response
		if err := WriteJSONResponseOK(w, response); err != nil {
			http.Error(w, "An error has occurred", http.StatusInternalServerError)
			return
		}
	}
}
