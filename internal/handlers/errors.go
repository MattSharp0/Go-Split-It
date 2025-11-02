package handlers

import (
	"errors"
	"net/http"

	"github.com/MattSharp0/transaction-split-go/internal/logger"
	"github.com/jackc/pgx/v5"
)

// HandleDBError handles database errors and writes appropriate HTTP responses.
// It differentiates between 404 (not found) and 500 (server error) based on pgx.ErrNoRows.
//
// Parameters:
//   - w: HTTP response writer
//   - err: The database error to handle
//   - notFoundMessage: Message to send for 404 errors (e.g., "Group not found")
//   - errorMessage: Message to send for 500 errors (e.g., "An error has occurred")
//   - logMessage: Message for logging the error (e.g., "Failed to get group by ID")
//   - logFields: Additional fields for structured logging (can be nil)
//
// Returns true if an error response was written (caller should return), false otherwise.
func HandleDBError(w http.ResponseWriter, err error, notFoundMessage, errorMessage, logMessage string, logFields ...interface{}) bool {
	if err == nil {
		return false
	}

	// Check if the error is "not found" (pgx.ErrNoRows)
	if errors.Is(err, pgx.ErrNoRows) {
		logger.Debug(logMessage+": not found", append([]interface{}{"error", err}, logFields...)...)
		http.Error(w, notFoundMessage, http.StatusNotFound)
		return true
	}

	// All other errors are treated as server errors (500)
	logArgs := []interface{}{"error", err}
	logArgs = append(logArgs, logFields...)
	logger.Error(logMessage, logArgs...)
	http.Error(w, errorMessage, http.StatusInternalServerError)
	return true
}

// HandleDBListError handles database errors for list operations.
// List operations typically don't return ErrNoRows (empty lists aren't errors),
// so this always treats errors as 500 Internal Server Error.
//
// Parameters:
//   - w: HTTP response writer
//   - err: The database error to handle
//   - errorMessage: Message to send for 500 errors (e.g., "An error has occurred")
//   - logMessage: Message for logging the error (e.g., "Failed to list groups")
//   - logFields: Additional fields for structured logging (can be nil)
//
// Returns true if an error response was written (caller should return), false otherwise.
func HandleDBListError(w http.ResponseWriter, err error, errorMessage, logMessage string, logFields ...interface{}) bool {
	if err == nil {
		return false
	}

	// List operations don't return ErrNoRows, so all errors are server errors
	logArgs := []interface{}{"error", err}
	logArgs = append(logArgs, logFields...)
	logger.Error(logMessage, logArgs...)
	http.Error(w, errorMessage, http.StatusInternalServerError)
	return true
}
