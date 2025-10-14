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
