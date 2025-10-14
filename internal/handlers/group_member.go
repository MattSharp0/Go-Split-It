package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	db "github.com/MattSharp0/transaction-split-go/db/sqlc"
	"github.com/MattSharp0/transaction-split-go/internal/models"
	"github.com/MattSharp0/transaction-split-go/internal/server"
)

func GroupMemberRoutes(s *server.Server, q db.Store) *http.ServeMux {
	mux := http.NewServeMux()

	// Root path handlers
	mux.HandleFunc("POST /", createGroupMember(q)) // POST: Create group member

	// ID path handlers
	mux.HandleFunc("GET /{id}", getGroupMemberByID(q))   // GET: Get group member by ID
	mux.HandleFunc("PUT /{id}", updateGroupMember(q))    // PUT: Update group member
	mux.HandleFunc("PATCH /{id}", updateGroupMember(q))  // PATCH: Update group member
	mux.HandleFunc("DELETE /{id}", deleteGroupMember(q)) // DELETE: Delete group member

	// Group-specific routes
	mux.HandleFunc("GET /group/{group_id}", listGroupMembersByGroupID(q)) // GET: List members by group

	return mux
}

func listGroupMembersByGroupID(store db.Store) http.HandlerFunc {
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

func getGroupMemberByID(store db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract {id} from path parameter
		idStr := r.PathValue("id")
		if idStr == "" {
			http.Error(w, "Group Member ID is required", http.StatusBadRequest)
			return
		}

		// Convert string ID to int64
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid group member ID format", http.StatusBadRequest)
			return
		}

		log.Printf("Getting group member with ID: %d", id)

		// Get group member from database
		groupMember, err := store.GetGroupMemberByID(context.Background(), id)
		if err != nil {
			log.Printf("GetGroupMemberByID (%v) returned an error: %v", id, err)
			http.Error(w, "Group member not found", http.StatusNotFound)
			return
		}

		// Convert to response format
		groupMemberResponse := models.GroupMemberResponse{
			ID:         groupMember.ID,
			GroupID:    groupMember.GroupID,
			GroupName:  groupMember.GroupName,
			MemberName: groupMember.MemberName,
			UserID:     groupMember.UserID,
			UserName:   groupMember.UserName,
			CreatedAt:  groupMember.CreatedAt,
		}

		// Send response
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(groupMemberResponse); err != nil {
			log.Printf("Error encoding group member response: %v", err)
			http.Error(w, "An error has occurred", http.StatusInternalServerError)
			return
		}
	}
}

func createGroupMember(store db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Decode request body
		decoder := json.NewDecoder(r.Body)
		defer r.Body.Close()

		var createGroupMemberReq models.CreateGroupMemberRequest
		err := decoder.Decode(&createGroupMemberReq)
		if err != nil {
			http.Error(w, "Bad request: invalid JSON", http.StatusBadRequest)
			return
		}

		// Validate input
		if createGroupMemberReq.GroupID == 0 {
			http.Error(w, "Group ID is required", http.StatusBadRequest)
			return
		}

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

func updateGroupMember(store db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract {id} from path parameter
		idStr := r.PathValue("id")
		if idStr == "" {
			http.Error(w, "Group Member ID is required", http.StatusBadRequest)
			return
		}

		// Convert string ID to int64
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid group member ID format", http.StatusBadRequest)
			return
		}

		// Decode request body
		decoder := json.NewDecoder(r.Body)
		defer r.Body.Close()

		var updateGroupMemberReq models.UpdateGroupMemberRequest
		err = decoder.Decode(&updateGroupMemberReq)
		if err != nil {
			http.Error(w, "Bad request: invalid JSON", http.StatusBadRequest)
			return
		}

		// Validate input
		if updateGroupMemberReq.GroupID == 0 {
			http.Error(w, "Group ID is required", http.StatusBadRequest)
			return
		}

		log.Printf("Updating group member with ID: %d", id)

		// Update group member in database
		groupMember, err := store.UpdateGroupMember(context.Background(), db.UpdateGroupMemberParams{
			ID:      id,
			GroupID: updateGroupMemberReq.GroupID,
			UserID:  updateGroupMemberReq.UserID,
		})
		if err != nil {
			log.Printf("UpdateGroupMember returned an error: %v", err)
			http.Error(w, "Group member not found or unable to update", http.StatusNotFound)
			return
		}

		// Convert to response format
		groupMemberResponse := models.GroupMemberResponse{
			ID:         groupMember.ID,
			GroupID:    groupMember.GroupID,
			MemberName: groupMember.MemberName,
			UserID:     groupMember.UserID,
			CreatedAt:  groupMember.CreatedAt,
		}

		// Send response
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(groupMemberResponse); err != nil {
			log.Printf("Error encoding group member response: %v", err)
			http.Error(w, "An error has occurred", http.StatusInternalServerError)
			return
		}
	}
}

func deleteGroupMember(store db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract {id} from path parameter
		idStr := r.PathValue("id")
		if idStr == "" {
			http.Error(w, "Group Member ID is required", http.StatusBadRequest)
			return
		}

		// Convert string ID to int64
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid group member ID format", http.StatusBadRequest)
			return
		}

		log.Printf("Deleting group member with ID: %d", id)

		// Delete group member from database
		groupMember, err := store.DeleteGroupMember(context.Background(), id)
		if err != nil {
			log.Printf("DeleteGroupMember returned an error: %v", err)
			http.Error(w, "Group member not found or unable to delete", http.StatusNotFound)
			return
		}

		// Convert to response format
		groupMemberResponse := models.GroupMemberResponse{
			ID:         groupMember.ID,
			GroupID:    groupMember.GroupID,
			MemberName: groupMember.MemberName,
			UserID:     groupMember.UserID,
			CreatedAt:  groupMember.CreatedAt,
		}

		// Send response with deleted group member data
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(groupMemberResponse); err != nil {
			log.Printf("Error encoding group member response: %v", err)
			http.Error(w, "An error has occurred", http.StatusInternalServerError)
			return
		}
	}
}
