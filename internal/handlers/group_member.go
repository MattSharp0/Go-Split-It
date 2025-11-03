package handlers

import (
	"context"
	"net/http"

	db "github.com/MattSharp0/transaction-split-go/db/sqlc"
	"github.com/MattSharp0/transaction-split-go/internal/logger"
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

func getGroupMemberByID(store db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract {id} from path parameter
		id, ok := ParsePathInt64(w, r, "id", "Group Member ID is required")
		if !ok {
			return
		}

		logger.Debug("Getting group member by ID", "group_member_id", id)

		// Get group member from database
		groupMember, err := store.GetGroupMemberByID(context.Background(), id)
		if HandleDBError(w, err, "Group member not found", "An error has occurred", "Failed to get group member by ID", "group_member_id", id) {
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
		if err := WriteJSONResponseOK(w, groupMemberResponse); err != nil {
			http.Error(w, "An error has occurred", http.StatusInternalServerError)
			return
		}
	}
}

func createGroupMember(store db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Decode request body
		var createGroupMemberReq models.CreateGroupMemberRequest
		if err := DecodeJSONBody(r, &createGroupMemberReq); err != nil {
			http.Error(w, "Bad request: invalid JSON", http.StatusBadRequest)
			return
		}

		// Validate input
		if createGroupMemberReq.GroupID == 0 {
			http.Error(w, "Group ID is required", http.StatusBadRequest)
			return
		}

		logger.Debug("Creating group member", "group_id", createGroupMemberReq.GroupID, "user_id", createGroupMemberReq.UserID)

		// Create group member in database
		groupMember, err := store.CreateGroupMember(context.Background(), db.CreateGroupMemberParams{
			GroupID: createGroupMemberReq.GroupID,
			UserID:  createGroupMemberReq.UserID,
		})
		if HandleDBListError(w, err, "An error has occurred", "Failed to create group member", "group_id", createGroupMemberReq.GroupID, "user_id", createGroupMemberReq.UserID) {
			return
		}
		logger.Debug("Group member created successfully", "group_member_id", groupMember.ID, "group_id", createGroupMemberReq.GroupID, "user_id", createGroupMemberReq.UserID)

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

func updateGroupMember(store db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract {id} from path parameter
		id, ok := ParsePathInt64(w, r, "id", "Group Member ID is required")
		if !ok {
			return
		}

		// Decode request body
		var updateGroupMemberReq models.UpdateGroupMemberRequest
		if err := DecodeJSONBody(r, &updateGroupMemberReq); err != nil {
			http.Error(w, "Bad request: invalid JSON", http.StatusBadRequest)
			return
		}

		// Validate input
		if updateGroupMemberReq.GroupID == 0 {
			http.Error(w, "Group ID is required", http.StatusBadRequest)
			return
		}

		logger.Debug("Updating group member", "group_member_id", id, "group_id", updateGroupMemberReq.GroupID, "user_id", updateGroupMemberReq.UserID)

		// Update group member in database
		groupMember, err := store.UpdateGroupMember(context.Background(), db.UpdateGroupMemberParams{
			ID:      id,
			GroupID: updateGroupMemberReq.GroupID,
			UserID:  updateGroupMemberReq.UserID,
		})
		if HandleDBError(w, err, "Group member not found", "An error has occurred", "Failed to update group member", "group_member_id", id) {
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
		if err := WriteJSONResponseOK(w, groupMemberResponse); err != nil {
			http.Error(w, "An error has occurred", http.StatusInternalServerError)
			return
		}
	}
}

func deleteGroupMember(store db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract {id} from path parameter
		id, ok := ParsePathInt64(w, r, "id", "Group Member ID is required")
		if !ok {
			return
		}

		logger.Debug("Deleting group member", "group_member_id", id)

		// Delete group member from database
		groupMember, err := store.DeleteGroupMember(context.Background(), id)
		if HandleDBError(w, err, "Group member not found", "An error has occurred", "Failed to delete group member", "group_member_id", id) {
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
		if err := WriteJSONResponseOK(w, groupMemberResponse); err != nil {
			http.Error(w, "An error has occurred", http.StatusInternalServerError)
			return
		}
	}
}
