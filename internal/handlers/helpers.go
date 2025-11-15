package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/MattSharp0/transaction-split-go/internal/auth"
	"github.com/MattSharp0/transaction-split-go/internal/logger"
)

// ParsePathInt64 extracts and parses an int64 path parameter from the request.
// Returns the parsed value and true if successful. On error, writes an HTTP error response
// and returns false (caller should return immediately).
func ParsePathInt64(w http.ResponseWriter, r *http.Request, paramName, errorMessage string) (int64, bool) {
	paramStr := r.PathValue(paramName)
	if paramStr == "" {
		http.Error(w, errorMessage, http.StatusBadRequest)
		return 0, false
	}

	param, err := strconv.ParseInt(paramStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid "+paramName+" format", http.StatusBadRequest)
		return 0, false
	}

	return param, true
}

// ParseQueryInt32 extracts and parses an int32 query parameter from the request.
// Returns the parsed value, or defaultValue if the parameter is not present.
// Returns an error if the parameter is present but invalid.
func ParseQueryInt32(r *http.Request, paramName string, defaultValue int32) (int32, error) {
	queryParams := r.URL.Query()
	paramStr := queryParams.Get(paramName)
	if paramStr == "" {
		return defaultValue, nil
	}

	param, err := strconv.ParseInt(paramStr, 10, 32)
	if err != nil {
		return defaultValue, err
	}

	return int32(param), nil
}

// ParseQueryDate extracts and parses a date query parameter from the request.
// Expected format is "2006-01-02" (YYYY-MM-DD).
// Returns the parsed value, or defaultValue if the parameter is not present.
// Returns an error if the parameter is present but invalid.
func ParseQueryDate(r *http.Request, paramName string, defaultValue time.Time) (time.Time, error) {
	queryParams := r.URL.Query()
	paramStr := queryParams.Get(paramName)
	if paramStr == "" {
		return defaultValue, nil
	}

	date, err := time.Parse("2006-01-02", paramStr)
	if err != nil {
		return defaultValue, err
	}

	return date, nil
}

// ParseLimitOffset parses limit and offset query parameters from the request.
// Returns limit and offset with default values of 100 and 0 respectively.
// Returns an error if either parameter is invalid. The error will indicate which parameter failed.
func ParseLimitOffset(r *http.Request) (limit int32, offset int32, err error) {
	limit, err = ParseQueryInt32(r, "limit", 100)
	if err != nil {
		return 100, 0, fmt.Errorf("limit: %w", err)
	}

	offset, err = ParseQueryInt32(r, "offset", 0)
	if err != nil {
		return limit, 0, fmt.Errorf("offset: %w", err)
	}

	return limit, offset, nil
}

// DecodeJSONBody decodes the request body as JSON into the provided destination.
// Automatically closes the request body. Returns an error if decoding fails.
func DecodeJSONBody(r *http.Request, dest interface{}) error {
	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()

	if err := decoder.Decode(dest); err != nil {
		return err
	}

	return nil
}

// WriteJSONResponse writes a JSON response with the given status code.
// Sets the Content-Type header and encodes the data as JSON.
// Returns an error if encoding fails.
func WriteJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		logger.Error("Failed to encode response", "error", err)
		return err
	}
	return nil
}

// WriteJSONResponseCreated is a convenience function that writes a JSON response with 201 Created status.
func WriteJSONResponseCreated(w http.ResponseWriter, data interface{}) error {
	return WriteJSONResponse(w, http.StatusCreated, data)
}

// WriteJSONResponseOK is a convenience function that writes a JSON response with 200 OK status.
func WriteJSONResponseOK(w http.ResponseWriter, data interface{}) error {
	return WriteJSONResponse(w, http.StatusOK, data)
}

// GetAuthenticatedUserID extracts the authenticated user ID from the request context.
// Returns the user ID and true if successful. On error, writes an HTTP error response
// and returns false (caller should return immediately).
func GetAuthenticatedUserID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	userID, ok := auth.GetUserID(r.Context())
	if !ok {
		logger.Warn("User ID not found in context")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return 0, false
	}
	return userID, true
}
