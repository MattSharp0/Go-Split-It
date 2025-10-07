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
	mux.HandleFunc("/", listUsers(q))       // User Root
	mux.HandleFunc("/{id}", getUserByID(q)) // User ID
	// mux.HandleFunc("/", createUser(q)) // User Create
	// mux.HandleFunc("/{id}", updateUser(q)) // User Update
	// mux.HandleFunc("/{id}", deleteUser(q)) // User Delete

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
		// Only handle GET requests
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

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
			log.Printf("GetUserByID returned an error: %v", err)
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
