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

func GroupRoutes(s *server.Server, q db.Store) *http.ServeMux {
	mux := http.NewServeMux()

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
		queryParams := r.URL.Query()

		// Default values
		var listGroupParams db.ListGroupsParams
		listGroupParams.Limit = 100
		listGroupParams.Offset = 0

		// Parse limit
		if limitStr := queryParams.Get("limit"); limitStr != "" {
			limit, err := strconv.ParseInt(limitStr, 10, 32)
			if err != nil {
				http.Error(w, "Invalid limit parameter", http.StatusBadRequest)
				return
			}
			listGroupParams.Limit = int32(limit)
		}

		// Parse offset
		if offsetStr := queryParams.Get("offset"); offsetStr != "" {
			offset, err := strconv.ParseInt(offsetStr, 10, 32)
			if err != nil {
				http.Error(w, "Invalid offset parameter", http.StatusBadRequest)
				return
			}
			listGroupParams.Offset = int32(offset)
		}

		log.Printf("List Group parameters: %v", listGroupParams)

		groups, err := store.ListGroups(context.Background(), listGroupParams)
		if err != nil {
			log.Printf("ListGroups returned an error: %v", err)
			http.Error(w, "An error has occurred", http.StatusInternalServerError)
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

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(listGroupResponse); err != nil {
			log.Printf("Error encoding group responses: %v", err)
			http.Error(w, "An error has occurred", http.StatusInternalServerError)
			return
		}
	}
}

func getGroupByID(store db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract {id} from path parameter
		idStr := r.PathValue("id")
		if idStr == "" {
			http.Error(w, "Group ID is required", http.StatusBadRequest)
			return
		}

		// Convert string ID to int64
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid group ID format", http.StatusBadRequest)
			return
		}

		log.Printf("Getting group with ID: %d", id)

		// Get group from database
		group, err := store.GetGroupByID(context.Background(), id)
		if err != nil {
			log.Printf("GetGroupByID (%v) returned an error: %v", id, err)
			http.Error(w, "Group not found", http.StatusNotFound)
			return
		}

		// Convert to response format
		groupResponse := models.GroupResponse{
			ID:   group.ID,
			Name: group.Name,
		}

		// Send response
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(groupResponse); err != nil {
			log.Printf("Error encoding group response: %v", err)
			http.Error(w, "An error has occurred", http.StatusInternalServerError)
			return
		}
	}
}

func createGroup(store db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Decode request body
		decoder := json.NewDecoder(r.Body)
		defer r.Body.Close()

		var createGroupReq models.CreateGroupRequest
		err := decoder.Decode(&createGroupReq)
		if err != nil {
			http.Error(w, "Bad request: invalid JSON", http.StatusBadRequest)
			return
		}

		// Validate input
		if createGroupReq.Name == "" {
			http.Error(w, "Name is required", http.StatusBadRequest)
			return
		}

		log.Printf("Creating group with name: %s", createGroupReq.Name)

		// Create group in database
		group, err := store.CreateGroup(context.Background(), createGroupReq.Name)
		if err != nil {
			log.Printf("CreateGroup returned an error: %v", err)
			http.Error(w, "An error has occurred", http.StatusInternalServerError)
			return
		}
		log.Printf("Created group '%s' with ID: %d", group.Name, group.ID)

		// Convert to response format
		groupResponse := models.GroupResponse{
			ID:   group.ID,
			Name: group.Name,
		}

		// Send response with 201 Created status
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(groupResponse); err != nil {
			log.Printf("Error encoding group response: %v", err)
			http.Error(w, "An error has occurred", http.StatusInternalServerError)
			return
		}
	}
}

func updateGroup(store db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract {id} from path parameter
		idStr := r.PathValue("id")
		if idStr == "" {
			http.Error(w, "Group ID is required", http.StatusBadRequest)
			return
		}

		// Convert string ID to int64
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid group ID format", http.StatusBadRequest)
			return
		}

		// Decode request body
		decoder := json.NewDecoder(r.Body)
		defer r.Body.Close()

		var updateGroupReq models.UpdateGroupRequest
		err = decoder.Decode(&updateGroupReq)
		if err != nil {
			http.Error(w, "Bad request: invalid JSON", http.StatusBadRequest)
			return
		}

		// Validate input
		if updateGroupReq.Name == "" {
			http.Error(w, "Name is required", http.StatusBadRequest)
			return
		}

		log.Printf("Updating group with ID: %d, new name: %s", id, updateGroupReq.Name)

		// Update group in database
		group, err := store.UpdateGroup(context.Background(), db.UpdateGroupParams{
			ID:   id,
			Name: updateGroupReq.Name,
		})
		if err != nil {
			log.Printf("UpdateGroup returned an error: %v", err)
			http.Error(w, "Group not found or unable to update", http.StatusNotFound)
			return
		}

		// Convert to response format
		groupResponse := models.GroupResponse{
			ID:   group.ID,
			Name: group.Name,
		}

		// Send response
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(groupResponse); err != nil {
			log.Printf("Error encoding group response: %v", err)
			http.Error(w, "An error has occurred", http.StatusInternalServerError)
			return
		}
	}
}

func deleteGroup(store db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract {id} from path parameter
		idStr := r.PathValue("id")
		if idStr == "" {
			http.Error(w, "Group ID is required", http.StatusBadRequest)
			return
		}

		// Convert string ID to int64
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid group ID format", http.StatusBadRequest)
			return
		}

		log.Printf("Deleting group with ID: %d", id)

		// Delete group from database
		group, err := store.DeleteGroup(context.Background(), id)
		if err != nil {
			log.Printf("DeleteGroup returned an error: %v", err)
			http.Error(w, "Group not found or unable to delete", http.StatusNotFound)
			return
		}

		// Convert to response format
		groupResponse := models.GroupResponse{
			ID:   group.ID,
			Name: group.Name,
		}

		// Send response with deleted group data
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(groupResponse); err != nil {
			log.Printf("Error encoding group response: %v", err)
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
		var listParams db.ListGroupMembersByGroupIDParams
		listParams.GroupID = groupID
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

		log.Printf("List GroupMembers for group %d, parameters: %v", groupID, listParams)

		groupMembers, err := store.ListGroupMembersByGroupID(context.Background(), listParams)
		if err != nil {
			log.Printf("ListGroupMembersByGroupID returned an error: %v", err)
			http.Error(w, "An error has occurred", http.StatusInternalServerError)
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

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(listGroupMemberResponse); err != nil {
			log.Printf("Error encoding group member responses: %v", err)
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

		// Decode request body
		decoder := json.NewDecoder(r.Body)
		defer r.Body.Close()

		var createGroupMemberReq models.CreateGroupMemberRequest
		err = decoder.Decode(&createGroupMemberReq)
		if err != nil {
			http.Error(w, "Bad request: invalid JSON", http.StatusBadRequest)
			return
		}

		// Override group_id from URL
		createGroupMemberReq.GroupID = groupID

		log.Printf("Creating group member for group: %d, user: %v", createGroupMemberReq.GroupID, createGroupMemberReq.UserID)

		// Create group member in database
		groupMember, err := store.CreateGroupMember(context.Background(), db.CreateGroupMemberParams{
			GroupID: createGroupMemberReq.GroupID,
			UserID:  createGroupMemberReq.UserID,
		})
		if err != nil {
			log.Printf("CreateGroupMember returned an error: %v", err)
			http.Error(w, "An error has occurred", http.StatusInternalServerError)
			return
		}
		log.Printf("Created group member with ID: %d", groupMember.ID)

		// Convert to response format
		groupMemberResponse := models.GroupMemberResponse{
			ID:         groupMember.ID,
			GroupID:    groupMember.GroupID,
			MemberName: groupMember.MemberName,
			UserID:     groupMember.UserID,
			CreatedAt:  groupMember.CreatedAt,
		}

		// Send response with 201 Created status
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(groupMemberResponse); err != nil {
			log.Printf("Error encoding group member response: %v", err)
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

		// Default to past year, TODO: make this configurable
		listParams.StartDate = time.Now().AddDate(-1, 0, 0)
		listParams.EndDate = time.Now()

		listParams.Limit = 100
		listParams.Offset = 0

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

// Create transaction in group
// POST /groups/{group_id}/transactions
func createTransactionNested(store db.Store) http.HandlerFunc {
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

		// Decode request body
		decoder := json.NewDecoder(r.Body)
		defer r.Body.Close()

		var createTransactionReq models.CreateTransactionRequest
		err = decoder.Decode(&createTransactionReq)
		if err != nil {
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

// Get group balances
// GET /groups/{group_id}/balances
func getGroupBalances(store db.Store) http.HandlerFunc {
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

		log.Printf("Getting balances for group %d", groupID)

		// Get all three types of balances
		balances, err := store.GroupBalances(context.Background(), groupID)
		if err != nil {
			log.Printf("GroupBalances returned an error: %v", err)
			http.Error(w, "An error has occurred", http.StatusInternalServerError)
			return
		}

		netBalances, err := store.GroupBalancesNet(context.Background(), groupID)
		if err != nil {
			log.Printf("GroupBalancesNet returned an error: %v", err)
			http.Error(w, "An error has occurred", http.StatusInternalServerError)
			return
		}

		simplifiedBalances, err := store.GroupBalancesSimplified(context.Background(), groupID)
		if err != nil {
			log.Printf("GroupBalancesSimplified returned an error: %v", err)
			http.Error(w, "An error has occurred", http.StatusInternalServerError)
			return
		}

		// Convert to response format
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
			balanceResponses[i] = models.BalanceResponse{
				Creditor:  creditor,
				Debtor:    debtor,
				TotalOwed: b.TotalOwed,
			}
		}

		netBalanceResponses := make([]models.NetBalanceResponse, len(netBalances))
		for i, nb := range netBalances {
			memberName := ""
			if nb.MemberName != nil {
				memberName = *nb.MemberName
			}
			netBalanceResponses[i] = models.NetBalanceResponse{
				MemberName: memberName,
				NetBalance: nb.NetBalance,
			}
		}

		simplifiedResponses := make([]models.BalanceResponse, len(simplifiedBalances))
		for i, sb := range simplifiedBalances {
			creditor := ""
			if sb.Creditor != nil {
				creditor = *sb.Creditor
			}
			debtor := ""
			if sb.Debtor != nil {
				debtor = *sb.Debtor
			}
			simplifiedResponses[i] = models.BalanceResponse{
				Creditor:  creditor,
				Debtor:    debtor,
				TotalOwed: sb.TotalOwed,
			}
		}

		response := models.GroupBalancesResponse{
			GroupID:         groupID,
			Balances:        balanceResponses,
			NetBalances:     netBalanceResponses,
			SimplifiedOwes:  simplifiedResponses,
			Count:           int32(len(balanceResponses)),
			NetCount:        int32(len(netBalanceResponses)),
			SimplifiedCount: int32(len(simplifiedResponses)),
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			log.Printf("Error encoding balance response: %v", err)
			http.Error(w, "An error has occurred", http.StatusInternalServerError)
			return
		}
	}
}
