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

func TestListSplits(t *testing.T) {
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
				splits := []db.Split{
					{
						ID:            1,
						TransactionID: 1,
						TxAmount:      decimal.NewFromInt(100),
						SplitPercent:  decimal.NewFromFloat(0.5),
						SplitAmount:   decimal.NewFromInt(50),
						CreatedAt:     time.Now(),
						ModifiedAt:    time.Now(),
					},
				}
				userID := int64Ptr(1)
				ms.On("ListSplitsByUserGroups", mock.Anything, db.ListSplitsByUserGroupsParams{UserID: userID, Limit: 100, Offset: 0}).Return(splits, nil)
			},
			requestURL:     "/splits",
			expectedStatus: http.StatusOK,
			expectedCount:  1,
		},
		{
			name: "empty list",
			setupMock: func(ms *mocks.MockStore) {
				userID := int64Ptr(1)
				ms.On("ListSplitsByUserGroups", mock.Anything, db.ListSplitsByUserGroupsParams{UserID: userID, Limit: 100, Offset: 0}).Return([]db.Split{}, nil)
			},
			requestURL:     "/splits",
			expectedStatus: http.StatusOK,
			expectedCount:  0,
		},
		{
			name: "database error",
			setupMock: func(ms *mocks.MockStore) {
				userID := int64Ptr(1)
				ms.On("ListSplitsByUserGroups", mock.Anything, db.ListSplitsByUserGroupsParams{UserID: userID, Limit: 100, Offset: 0}).Return(nil, errors.New("database error"))
			},
			requestURL:     "/splits",
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

			handler := listSplits(mockStore)
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

func TestGetSplitByID(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(*mocks.MockStore)
		pathValue      string
		expectedStatus int
		expectSplit    bool
	}{
		{
			name: "success",
			setupMock: func(ms *mocks.MockStore) {
				split := db.Split{
					ID:            1,
					TransactionID: 1,
					TxAmount:      decimal.NewFromInt(100),
					SplitPercent:  decimal.NewFromFloat(0.5),
					SplitAmount:   decimal.NewFromInt(50),
					SplitUser:     int64Ptr(1),
					CreatedAt:     time.Now(),
					ModifiedAt:    time.Now(),
				}
				ms.On("GetSplitByID", mock.Anything, int64(1)).Return(split, nil)
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
			},
			pathValue:      "1",
			expectedStatus: http.StatusOK,
			expectSplit:    true,
		},
		{
			name:           "invalid ID format",
			setupMock:      func(ms *mocks.MockStore) {},
			pathValue:      "abc",
			expectedStatus: http.StatusBadRequest,
			expectSplit:    false,
		},
		{
			name: "split not found",
			setupMock: func(ms *mocks.MockStore) {
				ms.On("GetSplitByID", mock.Anything, int64(999)).Return(db.Split{}, pgx.ErrNoRows)
			},
			pathValue:      "999",
			expectedStatus: http.StatusNotFound,
			expectSplit:    false,
		},
		{
			name: "database error",
			setupMock: func(ms *mocks.MockStore) {
				ms.On("GetSplitByID", mock.Anything, int64(1)).Return(db.Split{}, errors.New("database error"))
			},
			pathValue:      "1",
			expectedStatus: http.StatusInternalServerError,
			expectSplit:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := mocks.NewMockStore(t)
			tt.setupMock(mockStore)

			req := createRequestWithUserID("GET", "/splits/"+tt.pathValue, nil, 1)
			req.SetPathValue("id", tt.pathValue)
			rr := httptest.NewRecorder()

			handler := getSplitByID(mockStore)
			handler(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.expectSplit {
				var splitResponse map[string]interface{}
				err := json.Unmarshal(rr.Body.Bytes(), &splitResponse)
				require.NoError(t, err)
				assert.NotNil(t, splitResponse["id"])
			}
			mockStore.AssertExpectations(t)
		})
	}
}
