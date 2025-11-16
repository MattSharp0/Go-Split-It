package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	db "github.com/MattSharp0/transaction-split-go/db/sqlc"
	"github.com/MattSharp0/transaction-split-go/internal/mocks"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestListTransactions(t *testing.T) {
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
				userID := int64Ptr(1)
				ms.On("ListTransactionsByUserGroups", mock.Anything, db.ListTransactionsByUserGroupsParams{UserID: userID, Limit: 100, Offset: 0}).Return(transactions, nil)
			},
			requestURL:     "/transactions",
			expectedStatus: http.StatusOK,
			expectedCount:  1,
		},
		{
			name: "empty list",
			setupMock: func(ms *mocks.MockStore) {
				userID := int64Ptr(1)
				ms.On("ListTransactionsByUserGroups", mock.Anything, db.ListTransactionsByUserGroupsParams{UserID: userID, Limit: 100, Offset: 0}).Return([]db.Transaction{}, nil)
			},
			requestURL:     "/transactions",
			expectedStatus: http.StatusOK,
			expectedCount:  0,
		},
		{
			name: "database error",
			setupMock: func(ms *mocks.MockStore) {
				userID := int64Ptr(1)
				ms.On("ListTransactionsByUserGroups", mock.Anything, db.ListTransactionsByUserGroupsParams{UserID: userID, Limit: 100, Offset: 0}).Return(nil, errors.New("database error"))
			},
			requestURL:     "/transactions",
			expectedStatus: http.StatusInternalServerError,
			expectedCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := mocks.NewMockStore(t)
			tt.setupMock(mockStore)

			req := createRequestWithUserID("GET", tt.requestURL, nil, 1)
			rr := httptest.NewRecorder()

			handler := listTransactions(mockStore)
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

func TestGetTransactionByID(t *testing.T) {
	tests := []struct {
		name              string
		setupMock         func(*mocks.MockStore)
		pathValue         string
		expectedStatus    int
		expectTransaction bool
	}{
		{
			name: "success",
			setupMock: func(ms *mocks.MockStore) {
				transaction := db.Transaction{
					ID:              1,
					GroupID:         1,
					Name:            "Transaction 1",
					TransactionDate: time.Now(),
					Amount:          decimal.NewFromInt(100),
					ByUser:          1,
					CreatedAt:       time.Now(),
					ModifiedAt:      time.Now(),
				}
				ms.On("GetTransactionByID", mock.Anything, int64(1)).Return(transaction, nil)
				// Mock group membership check
				userID := int64Ptr(1)
				members := []db.ListGroupMembersByGroupIDRow{
					{ID: 1, GroupID: 1, UserID: userID},
				}
				ms.On("ListGroupMembersByGroupID", mock.Anything, db.ListGroupMembersByGroupIDParams{GroupID: 1, Limit: 1000, Offset: 0}).Return(members, nil)
			},
			pathValue:         "1",
			expectedStatus:    http.StatusOK,
			expectTransaction: true,
		},
		{
			name:              "invalid ID format",
			setupMock:         func(ms *mocks.MockStore) {},
			pathValue:         "abc",
			expectedStatus:    http.StatusBadRequest,
			expectTransaction: false,
		},
		{
			name: "transaction not found",
			setupMock: func(ms *mocks.MockStore) {
				ms.On("GetTransactionByID", mock.Anything, int64(999)).Return(db.Transaction{}, pgx.ErrNoRows)
			},
			pathValue:         "999",
			expectedStatus:    http.StatusNotFound,
			expectTransaction: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := mocks.NewMockStore(t)
			tt.setupMock(mockStore)

			req := createRequestWithUserID("GET", "/transactions/"+tt.pathValue, nil, 1)
			req.SetPathValue("id", tt.pathValue)
			rr := httptest.NewRecorder()

			handler := getTransactionByID(mockStore)
			handler(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.expectTransaction {
				var transactionResponse map[string]interface{}
				err := json.Unmarshal(rr.Body.Bytes(), &transactionResponse)
				require.NoError(t, err)
				assert.NotNil(t, transactionResponse["id"])
			}
			mockStore.AssertExpectations(t)
		})
	}
}

func TestCreateTransaction(t *testing.T) {
	tests := []struct {
		name              string
		setupMock         func(*mocks.MockStore)
		requestBody       interface{}
		expectedStatus    int
		expectTransaction bool
	}{
		{
			name: "success",
			setupMock: func(ms *mocks.MockStore) {
				// Mock group membership check
				userID := int64Ptr(1)
				members := []db.ListGroupMembersByGroupIDRow{
					{ID: 1, GroupID: 1, UserID: userID},
				}
				ms.On("ListGroupMembersByGroupID", mock.Anything, db.ListGroupMembersByGroupIDParams{GroupID: 1, Limit: 1000, Offset: 0}).Return(members, nil)
				// Mock GetGroupMemberByID for ByUser validation
				groupMember := db.GetGroupMemberByIDRow{
					ID:      1,
					GroupID: 1,
					UserID:  userID,
				}
				ms.On("GetGroupMemberByID", mock.Anything, int64(1)).Return(groupMember, nil)
				transaction := db.Transaction{
					ID:              1,
					GroupID:         1,
					Name:            "New Transaction",
					TransactionDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					Amount:          decimal.NewFromInt(100),
					ByUser:          1,
					CreatedAt:       time.Now(),
					ModifiedAt:      time.Now(),
				}
				ms.On("CreateTransaction", mock.Anything, mock.AnythingOfType("db.CreateTransactionParams")).Return(transaction, nil)
			},
			requestBody: map[string]interface{}{
				"group_id":         1,
				"name":             "New Transaction",
				"transaction_date": "2024-01-01T00:00:00Z",
				"amount":           "100.00",
				"by_user":          1,
			},
			expectedStatus:    http.StatusCreated,
			expectTransaction: true,
		},
		{
			name:              "missing name",
			setupMock:         func(ms *mocks.MockStore) {},
			requestBody:       map[string]interface{}{"group_id": 1, "by_user": 1},
			expectedStatus:    http.StatusBadRequest,
			expectTransaction: false,
		},
		{
			name:              "missing group_id",
			setupMock:         func(ms *mocks.MockStore) {},
			requestBody:       map[string]interface{}{"name": "Transaction", "by_user": 1},
			expectedStatus:    http.StatusBadRequest,
			expectTransaction: false,
		},
		{
			name:              "missing by_user",
			setupMock:         func(ms *mocks.MockStore) {},
			requestBody:       map[string]interface{}{"name": "Transaction", "group_id": 1},
			expectedStatus:    http.StatusBadRequest,
			expectTransaction: false,
		},
		{
			name:              "invalid JSON",
			setupMock:         func(ms *mocks.MockStore) {},
			requestBody:       "invalid json",
			expectedStatus:    http.StatusBadRequest,
			expectTransaction: false,
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

			req := createRequestWithUserID("POST", "/transactions", bodyBytes, 1)
			rr := httptest.NewRecorder()

			handler := createTransaction(mockStore)
			handler(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.expectTransaction {
				var transactionResponse map[string]interface{}
				err := json.Unmarshal(rr.Body.Bytes(), &transactionResponse)
				require.NoError(t, err)
				assert.NotNil(t, transactionResponse["id"])
			}
			mockStore.AssertExpectations(t)
		})
	}
}

func TestUpdateTransaction(t *testing.T) {
	tests := []struct {
		name              string
		setupMock         func(*mocks.MockStore)
		pathValue         string
		requestBody       interface{}
		expectedStatus    int
		expectTransaction bool
	}{
		{
			name: "success",
			setupMock: func(ms *mocks.MockStore) {
				// Get transaction first
				existingTransaction := db.Transaction{
					ID:      1,
					GroupID: 1,
				}
				ms.On("GetTransactionByID", mock.Anything, int64(1)).Return(existingTransaction, nil)
				// Mock group membership check
				userID := int64Ptr(1)
				members := []db.ListGroupMembersByGroupIDRow{
					{ID: 1, GroupID: 1, UserID: userID},
				}
				ms.On("ListGroupMembersByGroupID", mock.Anything, db.ListGroupMembersByGroupIDParams{GroupID: 1, Limit: 1000, Offset: 0}).Return(members, nil)
				// Mock GetGroupMemberByID for ByUser validation
				groupMember := db.GetGroupMemberByIDRow{
					ID:      1,
					GroupID: 1,
					UserID:  userID,
				}
				ms.On("GetGroupMemberByID", mock.Anything, int64(1)).Return(groupMember, nil)
				transaction := db.Transaction{
					ID:              1,
					GroupID:         1,
					Name:            "Updated Transaction",
					TransactionDate: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					Amount:          decimal.NewFromInt(200),
					ByUser:          1,
					CreatedAt:       time.Now(),
					ModifiedAt:      time.Now(),
				}
				ms.On("UpdateTransaction", mock.Anything, mock.AnythingOfType("db.UpdateTransactionParams")).Return(transaction, nil)
			},
			pathValue: "1",
			requestBody: map[string]interface{}{
				"group_id":         1,
				"name":             "Updated Transaction",
				"transaction_date": "2024-01-01T00:00:00Z",
				"amount":           "200.00",
				"by_user":          1,
			},
			expectedStatus:    http.StatusOK,
			expectTransaction: true,
		},
		{
			name:              "invalid ID format",
			setupMock:         func(ms *mocks.MockStore) {},
			pathValue:         "abc",
			requestBody:       map[string]interface{}{"name": "Transaction", "group_id": 1, "by_user": 1},
			expectedStatus:    http.StatusBadRequest,
			expectTransaction: false,
		},
		{
			name: "transaction not found",
			setupMock: func(ms *mocks.MockStore) {
				// Get transaction first
				ms.On("GetTransactionByID", mock.Anything, int64(999)).Return(db.Transaction{}, pgx.ErrNoRows)
			},
			pathValue: "999",
			requestBody: map[string]interface{}{
				"group_id": 1,
				"name":     "Transaction",
				"by_user":  1,
			},
			expectedStatus:    http.StatusNotFound,
			expectTransaction: false,
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

			req := createRequestWithUserID("PUT", "/transactions/"+tt.pathValue, bodyBytes, 1)
			req.SetPathValue("id", tt.pathValue)
			rr := httptest.NewRecorder()

			handler := updateTransaction(mockStore)
			handler(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.expectTransaction {
				var transactionResponse map[string]interface{}
				err := json.Unmarshal(rr.Body.Bytes(), &transactionResponse)
				require.NoError(t, err)
				assert.NotNil(t, transactionResponse["id"])
			}
			mockStore.AssertExpectations(t)
		})
	}
}

func TestDeleteTransaction(t *testing.T) {
	tests := []struct {
		name              string
		setupMock         func(*mocks.MockStore)
		pathValue         string
		expectedStatus    int
		expectTransaction bool
	}{
		{
			name: "success",
			setupMock: func(ms *mocks.MockStore) {
				transaction := db.Transaction{
					ID:              1,
					GroupID:         1,
					Name:            "Transaction 1",
					TransactionDate: time.Now(),
					Amount:          decimal.NewFromInt(100),
					ByUser:          1,
					CreatedAt:       time.Now(),
					ModifiedAt:      time.Now(),
				}
				ms.On("GetTransactionByID", mock.Anything, int64(1)).Return(transaction, nil)
				// Mock group membership check
				userID := int64Ptr(1)
				members := []db.ListGroupMembersByGroupIDRow{
					{ID: 1, GroupID: 1, UserID: userID},
				}
				ms.On("ListGroupMembersByGroupID", mock.Anything, db.ListGroupMembersByGroupIDParams{GroupID: 1, Limit: 1000, Offset: 0}).Return(members, nil)
				ms.On("DeleteTransaction", mock.Anything, int64(1)).Return(transaction, nil)
			},
			pathValue:         "1",
			expectedStatus:    http.StatusOK,
			expectTransaction: true,
		},
		{
			name:              "invalid ID format",
			setupMock:         func(ms *mocks.MockStore) {},
			pathValue:         "abc",
			expectedStatus:    http.StatusBadRequest,
			expectTransaction: false,
		},
		{
			name: "transaction not found",
			setupMock: func(ms *mocks.MockStore) {
				ms.On("GetTransactionByID", mock.Anything, int64(999)).Return(db.Transaction{}, pgx.ErrNoRows)
			},
			pathValue:         "999",
			expectedStatus:    http.StatusNotFound,
			expectTransaction: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := mocks.NewMockStore(t)
			tt.setupMock(mockStore)

			req := createRequestWithUserID("DELETE", "/transactions/"+tt.pathValue, nil, 1)
			req.SetPathValue("id", tt.pathValue)
			rr := httptest.NewRecorder()

			handler := deleteTransaction(mockStore)
			handler(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.expectTransaction {
				var transactionResponse map[string]interface{}
				err := json.Unmarshal(rr.Body.Bytes(), &transactionResponse)
				require.NoError(t, err)
				assert.NotNil(t, transactionResponse["id"])
			}
			mockStore.AssertExpectations(t)
		})
	}
}

func TestGetSplitsByTransactionNested(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(*mocks.MockStore)
		pathValue      string
		expectedStatus int
		expectSplits   bool
	}{
		{
			name: "success with splits",
			setupMock: func(ms *mocks.MockStore) {
				transaction := db.Transaction{
					ID:      1,
					GroupID: 1,
				}
				ms.On("GetTransactionByID", mock.Anything, int64(1)).Return(transaction, nil)
				// Mock group membership check
				userID := int64Ptr(1)
				members := []db.ListGroupMembersByGroupIDRow{
					{ID: 1, GroupID: 1, UserID: userID},
				}
				ms.On("ListGroupMembersByGroupID", mock.Anything, db.ListGroupMembersByGroupIDParams{GroupID: 1, Limit: 1000, Offset: 0}).Return(members, nil)
				splits := []db.Split{
					{
						ID:            1,
						TransactionID: 1,
						TxAmount:      decimal.NewFromInt(100),
						SplitPercent:  decimal.NewFromFloat(0.5),
						SplitAmount:   decimal.NewFromInt(50),
						SplitUser:     int64Ptr(1),
						CreatedAt:     time.Now(),
						ModifiedAt:    time.Now(),
					},
				}
				ms.On("GetSplitsByTransactionID", mock.Anything, int64(1)).Return(splits, nil)
			},
			pathValue:      "1",
			expectedStatus: http.StatusOK,
			expectSplits:   true,
		},
		{
			name:           "invalid transaction ID format",
			setupMock:      func(ms *mocks.MockStore) {},
			pathValue:      "abc",
			expectedStatus: http.StatusBadRequest,
			expectSplits:   false,
		},
		{
			name: "database error",
			setupMock: func(ms *mocks.MockStore) {
				transaction := db.Transaction{
					ID:      1,
					GroupID: 1,
				}
				ms.On("GetTransactionByID", mock.Anything, int64(1)).Return(transaction, nil)
				// Mock group membership check
				userID := int64Ptr(1)
				members := []db.ListGroupMembersByGroupIDRow{
					{ID: 1, GroupID: 1, UserID: userID},
				}
				ms.On("ListGroupMembersByGroupID", mock.Anything, db.ListGroupMembersByGroupIDParams{GroupID: 1, Limit: 1000, Offset: 0}).Return(members, nil)
				ms.On("GetSplitsByTransactionID", mock.Anything, int64(1)).Return(nil, errors.New("database error"))
			},
			pathValue:      "1",
			expectedStatus: http.StatusInternalServerError,
			expectSplits:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := mocks.NewMockStore(t)
			tt.setupMock(mockStore)

			req := createRequestWithUserID("GET", "/transactions/"+tt.pathValue+"/splits", nil, 1)
			req.SetPathValue("transaction_id", tt.pathValue)
			rr := httptest.NewRecorder()

			handler := getSplitsByTransactionNested(mockStore)
			handler(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.expectSplits {
				var response map[string]interface{}
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.NotNil(t, response["splits"])
			}
			mockStore.AssertExpectations(t)
		})
	}
}
