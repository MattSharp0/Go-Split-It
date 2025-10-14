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

func UserRoutes(s *server.Server, q db.Store) *http.ServeMux {
	mux := http.NewServeMux()

	// Root path handlers - different methods handled within functions
	mux.HandleFunc("POST /", createUser(q)) // POST: Create user
	mux.HandleFunc("GET /", listUsers(q))   // GET: List users (with body params)

	// ID path handlers - different methods handled within functions
	mux.HandleFunc("GET /{id}", getUserByID(q))   // GET: Get user by ID
	mux.HandleFunc("PUT /{id}", updateUser(q))    // PUT: Update user
	mux.HandleFunc("PATCH /{id}", updateUser(q))  // PATCH: Update user
	mux.HandleFunc("DELETE /{id}", deleteUser(q)) // DELETE: Delete user

	return mux
}

func listUsers(store db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		decoder := json.NewDecoder(r.Body)
		defer r.Body.Close()

		var listuserparams db.ListUsersParams
		err := decoder.Decode(&listuserparams)
		if err != nil {
			http.Error(w, "Bad ListUsers request", http.StatusBadRequest)
			return
		}
		log.Printf("List User parameters: %v", listuserparams)

		users, err := store.ListUsers(context.Background(), listuserparams)
		if err != nil {
			log.Printf("ListUsers returned an error: %v", err)
			http.Error(w, "An error has occured", http.StatusInternalServerError)
			return
		}

		userResponses := make([]models.UserResponse, len(users))
		for i, user := range users {
			userResponses[i] = models.UserResponse{
				ID:         user.ID,
				Name:       user.Name,
				CreatedAt:  user.CreatedAt,
				ModifiedAt: user.ModifiedAt,
			}
		}

		count := len(userResponses)

		listuserresponse := models.ListUserResponse{
			Users:  userResponses,
			Count:  int32(count),
			Limit:  listuserparams.Limit,
			Offset: listuserparams.Offset,
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(listuserresponse); err != nil {
			log.Printf("Error encoding user responses: %v", err)
			http.Error(w, "An error has occured", http.StatusInternalServerError)
			return
		}

	}
}

func getUserByID(store db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract {id} from path parameter
		idStr := r.PathValue("id")
		if idStr == "" {
			http.Error(w, "User ID is required", http.StatusBadRequest)
			return
		}

		// Convert string ID to int64
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid user ID format", http.StatusBadRequest)
			return
		}

		log.Printf("Getting user with ID: %d", id)

		// Get user from database
		user, err := store.GetUserByID(context.Background(), id)
		if err != nil {
			log.Printf("GetUserByID (%v) returned an error: %v", id, err)
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}

		// Convert to response format
		userResponse := models.UserResponse{
			ID:         user.ID,
			Name:       user.Name,
			CreatedAt:  user.CreatedAt,
			ModifiedAt: user.ModifiedAt,
		}

		// Send response
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(userResponse); err != nil {
			log.Printf("Error encoding user response: %v", err)
			http.Error(w, "An error has occurred", http.StatusInternalServerError)
			return
		}
	}
}

func createUser(store db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Decode request body
		decoder := json.NewDecoder(r.Body)
		defer r.Body.Close()

		var createUserReq models.CreateUserRequest
		err := decoder.Decode(&createUserReq)
		if err != nil {
			http.Error(w, "Bad request: invalid JSON", http.StatusBadRequest)
			return
		}

		// Validate input
		if createUserReq.Name == "" {
			http.Error(w, "Name is required", http.StatusBadRequest)
			return
		}

		log.Printf("Creating user with name: %s", createUserReq.Name)

		// Create user in database
		user, err := store.CreateUser(context.Background(), createUserReq.Name)
		if err != nil {
			log.Printf("CreateUser returned an error: %v", err)
			http.Error(w, "An error has occurred", http.StatusInternalServerError)
			return
		}
		log.Printf("Created user '%s' with ID: %d", user.Name, user.ID)

		// Convert to response format
		userResponse := models.UserResponse{
			ID:         user.ID,
			Name:       user.Name,
			CreatedAt:  user.CreatedAt,
			ModifiedAt: user.ModifiedAt,
		}

		// Send response with 201 Created status
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(userResponse); err != nil {
			log.Printf("Error encoding user response: %v", err)
			http.Error(w, "An error has occurred", http.StatusInternalServerError)
			return
		}
	}
}

func updateUser(store db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract {id} from path parameter
		idStr := r.PathValue("id")
		if idStr == "" {
			http.Error(w, "User ID is required", http.StatusBadRequest)
			return
		}

		// Convert string ID to int64
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid user ID format", http.StatusBadRequest)
			return
		}

		// Decode request body
		decoder := json.NewDecoder(r.Body)
		defer r.Body.Close()

		var updateUserReq models.UpdateUserRequest
		err = decoder.Decode(&updateUserReq)
		if err != nil {
			http.Error(w, "Bad request: invalid JSON", http.StatusBadRequest)
			return
		}

		// Validate input
		if updateUserReq.Name == "" {
			http.Error(w, "Name is required", http.StatusBadRequest)
			return
		}

		log.Printf("Updating user with ID: %d, new name: %s", id, updateUserReq.Name)

		// Update user in database
		user, err := store.UpdateUser(context.Background(), db.UpdateUserParams{
			ID:   id,
			Name: updateUserReq.Name,
		})
		if err != nil {
			log.Printf("UpdateUser returned an error: %v", err)
			http.Error(w, "User not found or unable to update", http.StatusNotFound)
			return
		}

		// Convert to response format
		userResponse := models.UserResponse{
			ID:         user.ID,
			Name:       user.Name,
			CreatedAt:  user.CreatedAt,
			ModifiedAt: user.ModifiedAt,
		}

		// Send response
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(userResponse); err != nil {
			log.Printf("Error encoding user response: %v", err)
			http.Error(w, "An error has occurred", http.StatusInternalServerError)
			return
		}
	}
}

func deleteUser(store db.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract {id} from path parameter
		idStr := r.PathValue("id")
		if idStr == "" {
			http.Error(w, "User ID is required", http.StatusBadRequest)
			return
		}

		// Convert string ID to int64
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			http.Error(w, "Invalid user ID format", http.StatusBadRequest)
			return
		}

		log.Printf("Deleting user with ID: %d", id)

		// Delete user from database
		user, err := store.DeleteUser(context.Background(), id)
		if err != nil {
			log.Printf("DeleteUser returned an error: %v", err)
			http.Error(w, "User not found or unable to delete", http.StatusNotFound)
			return
		}

		// Convert to response format
		userResponse := models.UserResponse{
			ID:         user.ID,
			Name:       user.Name,
			CreatedAt:  user.CreatedAt,
			ModifiedAt: user.ModifiedAt,
		}

		// Send response with deleted user data
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(userResponse); err != nil {
			log.Printf("Error encoding user response: %v", err)
			http.Error(w, "An error has occurred", http.StatusInternalServerError)
			return
		}
	}
}
