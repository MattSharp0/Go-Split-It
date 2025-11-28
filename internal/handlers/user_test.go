package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	db "github.com/MattSharp0/transaction-split-go/db/sqlc"
	"github.com/MattSharp0/transaction-split-go/internal/logger"
	"github.com/MattSharp0/transaction-split-go/internal/mocks"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	// Initialize logger for tests - disable output
	cfg := logger.Config{
		Level:      logger.LevelDebug,
		Output:     logger.OutputDiscard,
		JSONFormat: false,
	}
	_, err := logger.InitLogger(cfg)
	if err != nil {
		os.Exit(1)
	}
	code := m.Run()
	os.Exit(code)
}

// storeAsInterface converts MockStore to db.Store interface
// This is now a simple type conversion since MockStore fully implements db.Store
func storeAsInterface(ms *mocks.MockStore) db.Store {
	return ms
}

func TestListUsers(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(*mocks.MockStore)
		requestURL     string
		expectedStatus int
		expectedCount  int
	}{
		{
			name: "success with default pagination",
			setupMock: func(ms *mocks.MockStore) {
				users := []db.User{
					{ID: 1, Name: "Alice", CreatedAt: time.Now(), ModifiedAt: time.Now()},
					{ID: 2, Name: "Bob", CreatedAt: time.Now(), ModifiedAt: time.Now()},
				}
				ms.On("ListUsers", mock.Anything, db.ListUsersParams{Limit: 100, Offset: 0}).Return(users, nil)
			},
			requestURL:     "/users",
			expectedStatus: http.StatusOK,
			expectedCount:  2,
		},
		{
			name: "success with custom pagination",
			setupMock: func(ms *mocks.MockStore) {
				users := []db.User{
					{ID: 1, Name: "Alice", CreatedAt: time.Now(), ModifiedAt: time.Now()},
				}
				ms.On("ListUsers", mock.Anything, db.ListUsersParams{Limit: 50, Offset: 10}).Return(users, nil)
			},
			requestURL:     "/users?limit=50&offset=10",
			expectedStatus: http.StatusOK,
			expectedCount:  1,
		},
		{
			name: "empty list",
			setupMock: func(ms *mocks.MockStore) {
				ms.On("ListUsers", mock.Anything, db.ListUsersParams{Limit: 100, Offset: 0}).Return([]db.User{}, nil)
			},
			requestURL:     "/users",
			expectedStatus: http.StatusOK,
			expectedCount:  0,
		},
		{
			name: "database error",
			setupMock: func(ms *mocks.MockStore) {
				ms.On("ListUsers", mock.Anything, db.ListUsersParams{Limit: 100, Offset: 0}).Return(nil, errors.New("database error"))
			},
			requestURL:     "/users",
			expectedStatus: http.StatusInternalServerError,
			expectedCount:  0,
		},
		{
			name:           "invalid limit parameter",
			setupMock:      func(ms *mocks.MockStore) {},
			requestURL:     "/users?limit=abc",
			expectedStatus: http.StatusBadRequest,
			expectedCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := mocks.NewMockStore(t)
			tt.setupMock(mockStore)

			req := httptest.NewRequest("GET", tt.requestURL, nil)
			rr := httptest.NewRecorder()

			handler := listUsers(storeAsInterface(mockStore))
			handler(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, float64(tt.expectedCount), response["count"])
			}
			mockStore.AssertExpectations(t)
		})
	}
}

func TestGetUserByID(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(*mocks.MockStore)
		pathValue      string
		expectedStatus int
		expectUser     bool
	}{
		{
			name: "success",
			setupMock: func(ms *mocks.MockStore) {
				user := db.User{ID: 1, Name: "Alice", CreatedAt: time.Now(), ModifiedAt: time.Now()}
				ms.On("GetUserByID", mock.Anything, int64(1)).Return(user, nil)
			},
			pathValue:      "1",
			expectedStatus: http.StatusOK,
			expectUser:     true,
		},
		{
			name:           "invalid ID format",
			setupMock:      func(ms *mocks.MockStore) {},
			pathValue:      "abc",
			expectedStatus: http.StatusBadRequest,
			expectUser:     false,
		},
		{
			name: "user not found",
			setupMock: func(ms *mocks.MockStore) {
				ms.On("GetUserByID", mock.Anything, int64(999)).Return(db.User{}, pgx.ErrNoRows)
			},
			pathValue:      "999",
			expectedStatus: http.StatusNotFound,
			expectUser:     false,
		},
		{
			name: "database error",
			setupMock: func(ms *mocks.MockStore) {
				ms.On("GetUserByID", mock.Anything, int64(1)).Return(db.User{}, errors.New("database error"))
			},
			pathValue:      "1",
			expectedStatus: http.StatusInternalServerError,
			expectUser:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := mocks.NewMockStore(t)
			tt.setupMock(mockStore)

			req := httptest.NewRequest("GET", "/users/"+tt.pathValue, nil)
			req.SetPathValue("id", tt.pathValue)
			rr := httptest.NewRecorder()

			handler := getUserByID(storeAsInterface(mockStore))
			handler(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.expectUser {
				var userResponse map[string]interface{}
				err := json.Unmarshal(rr.Body.Bytes(), &userResponse)
				require.NoError(t, err)
				assert.NotNil(t, userResponse["id"])
			}
			mockStore.AssertExpectations(t)
		})
	}
}

func TestCreateUser(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(*mocks.MockStore)
		requestBody    interface{}
		expectedStatus int
		expectUser     bool
	}{
		{
			name: "success",
			setupMock: func(ms *mocks.MockStore) {
				user := db.User{ID: 1, Name: "Alice", CreatedAt: time.Now(), ModifiedAt: time.Now()}
				ms.On("CreateUser", mock.Anything, "Alice").Return(user, nil)
			},
			requestBody:    map[string]string{"name": "Alice"},
			expectedStatus: http.StatusCreated,
			expectUser:     true,
		},
		{
			name:           "missing name",
			setupMock:      func(ms *mocks.MockStore) {},
			requestBody:    map[string]string{},
			expectedStatus: http.StatusBadRequest,
			expectUser:     false,
		},
		{
			name:           "empty name",
			setupMock:      func(ms *mocks.MockStore) {},
			requestBody:    map[string]string{"name": ""},
			expectedStatus: http.StatusBadRequest,
			expectUser:     false,
		},
		{
			name:           "invalid JSON",
			setupMock:      func(ms *mocks.MockStore) {},
			requestBody:    "invalid json",
			expectedStatus: http.StatusBadRequest,
			expectUser:     false,
		},
		{
			name: "database error",
			setupMock: func(ms *mocks.MockStore) {
				ms.On("CreateUser", mock.Anything, "Alice").Return(db.User{}, errors.New("database error"))
			},
			requestBody:    map[string]string{"name": "Alice"},
			expectedStatus: http.StatusInternalServerError,
			expectUser:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := mocks.NewMockStore(t)
			tt.setupMock(mockStore)

			var bodyBytes []byte
			if str, ok := tt.requestBody.(string); ok {
				bodyBytes = []byte(str)
			} else {
				var err error
				bodyBytes, err = json.Marshal(tt.requestBody)
				require.NoError(t, err)
			}

			req := httptest.NewRequest("POST", "/users", bytes.NewBuffer(bodyBytes))
			rr := httptest.NewRecorder()

			handler := createUser(storeAsInterface(mockStore))
			handler(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.expectUser {
				var userResponse map[string]interface{}
				err := json.Unmarshal(rr.Body.Bytes(), &userResponse)
				require.NoError(t, err)
				assert.NotNil(t, userResponse["id"])
			}
			mockStore.AssertExpectations(t)
		})
	}
}

func TestUpdateUser(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(*mocks.MockStore)
		pathValue      string
		requestBody    interface{}
		expectedStatus int
		expectUser     bool
	}{
		{
			name: "success",
			setupMock: func(ms *mocks.MockStore) {
				user := db.User{ID: 1, Name: "Alice Updated", CreatedAt: time.Now(), ModifiedAt: time.Now()}
				ms.On("UpdateUser", mock.Anything, db.UpdateUserParams{ID: 1, Name: "Alice Updated"}).Return(user, nil)
				// Note: CheckOwnUser validates authenticatedUserID == id, so we use userID 1 for both
			},
			pathValue:      "1",
			requestBody:    map[string]string{"name": "Alice Updated"},
			expectedStatus: http.StatusOK,
			expectUser:     true,
		},
		{
			name:           "invalid ID format",
			setupMock:      func(ms *mocks.MockStore) {},
			pathValue:      "abc",
			requestBody:    map[string]string{"name": "Alice"},
			expectedStatus: http.StatusBadRequest,
			expectUser:     false,
		},
		{
			name:           "missing name",
			setupMock:      func(ms *mocks.MockStore) {},
			pathValue:      "1",
			requestBody:    map[string]string{},
			expectedStatus: http.StatusBadRequest,
			expectUser:     false,
		},
		{
			name: "user not found",
			setupMock: func(ms *mocks.MockStore) {
				// CheckOwnUser validates authenticatedUserID == id, so use 1 for both
				ms.On("UpdateUser", mock.Anything, db.UpdateUserParams{ID: 1, Name: "Alice"}).Return(db.User{}, pgx.ErrNoRows)
			},
			pathValue:      "1",
			requestBody:    map[string]string{"name": "Alice"},
			expectedStatus: http.StatusNotFound,
			expectUser:     false,
		},
		{
			name:           "invalid JSON",
			setupMock:      func(ms *mocks.MockStore) {},
			pathValue:      "1",
			requestBody:    "invalid json",
			expectedStatus: http.StatusBadRequest,
			expectUser:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := mocks.NewMockStore(t)
			tt.setupMock(mockStore)

			var bodyBytes []byte
			if str, ok := tt.requestBody.(string); ok {
				bodyBytes = []byte(str)
			} else {
				var err error
				bodyBytes, err = json.Marshal(tt.requestBody)
				require.NoError(t, err)
			}

			req := createRequestWithUserID("PUT", "/users/"+tt.pathValue, bodyBytes, 1)
			req.SetPathValue("id", tt.pathValue)
			rr := httptest.NewRecorder()

			handler := updateUser(storeAsInterface(mockStore))
			handler(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.expectUser {
				var userResponse map[string]interface{}
				err := json.Unmarshal(rr.Body.Bytes(), &userResponse)
				require.NoError(t, err)
				assert.NotNil(t, userResponse["id"])
			}
			mockStore.AssertExpectations(t)
		})
	}
}

func TestDeleteUser(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(*mocks.MockStore)
		pathValue      string
		expectedStatus int
		expectUser     bool
	}{
		{
			name: "success",
			setupMock: func(ms *mocks.MockStore) {
				user := db.User{ID: 1, Name: "Alice", CreatedAt: time.Now(), ModifiedAt: time.Now()}
				ms.On("DeleteUser", mock.Anything, int64(1)).Return(user, nil)
			},
			pathValue:      "1",
			expectedStatus: http.StatusOK,
			expectUser:     true,
		},
		{
			name:           "invalid ID format",
			setupMock:      func(ms *mocks.MockStore) {},
			pathValue:      "abc",
			expectedStatus: http.StatusBadRequest,
			expectUser:     false,
		},
		{
			name: "user not found",
			setupMock: func(ms *mocks.MockStore) {
				// CheckOwnUser validates authenticatedUserID == id, so use 1 for both
				ms.On("DeleteUser", mock.Anything, int64(1)).Return(db.User{}, pgx.ErrNoRows)
			},
			pathValue:      "1",
			expectedStatus: http.StatusNotFound,
			expectUser:     false,
		},
		{
			name: "database error",
			setupMock: func(ms *mocks.MockStore) {
				ms.On("DeleteUser", mock.Anything, int64(1)).Return(db.User{}, errors.New("database error"))
			},
			pathValue:      "1",
			expectedStatus: http.StatusInternalServerError,
			expectUser:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := mocks.NewMockStore(t)
			tt.setupMock(mockStore)

			req := createRequestWithUserID("DELETE", "/users/"+tt.pathValue, nil, 1)
			req.SetPathValue("id", tt.pathValue)
			rr := httptest.NewRecorder()

			handler := deleteUser(storeAsInterface(mockStore))
			handler(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.expectUser {
				var userResponse map[string]interface{}
				err := json.Unmarshal(rr.Body.Bytes(), &userResponse)
				require.NoError(t, err)
				assert.NotNil(t, userResponse["id"])
			}
			mockStore.AssertExpectations(t)
		})
	}
}

func TestGetUserBalances(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(*mocks.MockStore)
		pathValue      string
		expectedStatus int
	}{
		{
			name: "success",
			setupMock: func(ms *mocks.MockStore) {
				summary := db.UserBalancesSummaryRow{
					NetBalance:      decimal.NewFromInt(100),
					TotalOwed:       decimal.NewFromInt(50),
					TotalOwedToUser: decimal.NewFromInt(150),
				}
				ms.On("UserBalancesSummary", mock.Anything, mock.MatchedBy(func(id *int64) bool {
					return *id == 1
				})).Return(summary, nil)

				balancesByGroup := []db.UserBalancesByGroupRow{
					{GroupID: 1, GroupName: "Group 1", NetBalance: decimal.NewFromInt(50)},
				}
				ms.On("UserBalancesByGroup", mock.Anything, mock.MatchedBy(func(id *int64) bool {
					return *id == 1
				})).Return(balancesByGroup, nil)

				balancesByMember := []db.UserBalancesByMemberRow{
					{MemberUserID: intPtr(2), MemberName: "Bob", NetBalance: decimal.NewFromInt(25)},
				}
				ms.On("UserBalancesByMember", mock.Anything, mock.MatchedBy(func(id *int64) bool {
					return *id == 1
				})).Return(balancesByMember, nil)
			},
			pathValue:      "1",
			expectedStatus: http.StatusOK,
		},
		{
			name: "user not found",
			setupMock: func(ms *mocks.MockStore) {
				// CheckOwnUser validates authenticatedUserID == pathUserID, so use 1 for both
				ms.On("UserBalancesSummary", mock.Anything, mock.MatchedBy(func(id *int64) bool {
					return id != nil && *id == 1
				})).Return(db.UserBalancesSummaryRow{}, pgx.ErrNoRows)
			},
			pathValue:      "1",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := mocks.NewMockStore(t)
			tt.setupMock(mockStore)

			req := createRequestWithUserID("GET", "/users/"+tt.pathValue+"/balances", nil, 1)
			req.SetPathValue("user_id", tt.pathValue)
			rr := httptest.NewRecorder()

			handler := getUserBalances(storeAsInterface(mockStore))
			handler(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.NotNil(t, response["summary"])
			}
			mockStore.AssertExpectations(t)
		})
	}
}

func TestGetTransactionsByUserNested(t *testing.T) {
	tests := []struct {
		name               string
		setupMock          func(*mocks.MockStore)
		pathValue          string
		queryParams        string
		expectedStatus     int
		expectTransactions bool
	}{
		{
			name: "success with transactions",
			setupMock: func(ms *mocks.MockStore) {
				transactions := []db.Transaction{
					{
						ID:              1,
						GroupID:         1,
						Name:            "Transaction 1",
						TransactionDate: time.Now(),
						Amount:          decimal.NewFromInt(100),
						ByUser:          1,
						CreatedAt:       time.Now(),
						ModifiedAt:      time.Now(),
					},
				}
				ms.On("GetTransactionsByUserInPeriod", mock.Anything, mock.AnythingOfType("db.GetTransactionsByUserInPeriodParams")).Return(transactions, nil)
			},
			pathValue:          "1",
			queryParams:        "",
			expectedStatus:     http.StatusOK,
			expectTransactions: true,
		},
		{
			name: "success with empty transactions",
			setupMock: func(ms *mocks.MockStore) {
				ms.On("GetTransactionsByUserInPeriod", mock.Anything, mock.AnythingOfType("db.GetTransactionsByUserInPeriodParams")).Return([]db.Transaction{}, nil)
			},
			pathValue:          "1",
			queryParams:        "",
			expectedStatus:     http.StatusOK,
			expectTransactions: false,
		},
		{
			name: "database error",
			setupMock: func(ms *mocks.MockStore) {
				ms.On("GetTransactionsByUserInPeriod", mock.Anything, mock.AnythingOfType("db.GetTransactionsByUserInPeriodParams")).Return(nil, errors.New("database error"))
			},
			pathValue:          "1",
			queryParams:        "",
			expectedStatus:     http.StatusInternalServerError,
			expectTransactions: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := mocks.NewMockStore(t)
			tt.setupMock(mockStore)

			url := "/users/" + tt.pathValue + "/transactions"
			if tt.queryParams != "" {
				url += "?" + tt.queryParams
			}
			req := createRequestWithUserID("GET", url, nil, 1)
			req.SetPathValue("user_id", tt.pathValue)
			rr := httptest.NewRecorder()

			handler := getTransactionsByUserNested(storeAsInterface(mockStore))
			handler(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.expectTransactions {
				var response map[string]interface{}
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.NotNil(t, response["transactions"])
			}
			mockStore.AssertExpectations(t)
		})
	}
}

func intPtr(i int64) *int64 {
	return &i
}
