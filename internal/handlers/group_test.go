package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	db "github.com/MattSharp0/transaction-split-go/db/sqlc"
	"github.com/MattSharp0/transaction-split-go/internal/mocks"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestListGroups(t *testing.T) {
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
				groups := []db.Group{
					{ID: 1, Name: "Group 1"},
					{ID: 2, Name: "Group 2"},
				}
				ms.On("ListGroups", mock.Anything, db.ListGroupsParams{Limit: 100, Offset: 0}).Return(groups, nil)
			},
			requestURL:     "/groups",
			expectedStatus: http.StatusOK,
			expectedCount:  2,
		},
		{
			name: "success with custom pagination",
			setupMock: func(ms *mocks.MockStore) {
				groups := []db.Group{
					{ID: 1, Name: "Group 1"},
				}
				ms.On("ListGroups", mock.Anything, db.ListGroupsParams{Limit: 50, Offset: 10}).Return(groups, nil)
			},
			requestURL:     "/groups?limit=50&offset=10",
			expectedStatus: http.StatusOK,
			expectedCount:  1,
		},
		{
			name: "empty list",
			setupMock: func(ms *mocks.MockStore) {
				ms.On("ListGroups", mock.Anything, db.ListGroupsParams{Limit: 100, Offset: 0}).Return([]db.Group{}, nil)
			},
			requestURL:     "/groups",
			expectedStatus: http.StatusOK,
			expectedCount:  0,
		},
		{
			name: "database error",
			setupMock: func(ms *mocks.MockStore) {
				ms.On("ListGroups", mock.Anything, db.ListGroupsParams{Limit: 100, Offset: 0}).Return(nil, errors.New("database error"))
			},
			requestURL:     "/groups",
			expectedStatus: http.StatusInternalServerError,
			expectedCount:  0,
		},
		{
			name:           "invalid limit parameter",
			setupMock:      func(ms *mocks.MockStore) {},
			requestURL:     "/groups?limit=abc",
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

			handler := listGroups(mockStore)
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

func TestGetGroupByID(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(*mocks.MockStore)
		pathValue      string
		expectedStatus int
		expectGroup    bool
	}{
		{
			name: "success",
			setupMock: func(ms *mocks.MockStore) {
				group := db.Group{ID: 1, Name: "Group 1"}
				ms.On("GetGroupByID", mock.Anything, int64(1)).Return(group, nil)
			},
			pathValue:      "1",
			expectedStatus: http.StatusOK,
			expectGroup:    true,
		},
		{
			name:           "invalid ID format",
			setupMock:      func(ms *mocks.MockStore) {},
			pathValue:      "abc",
			expectedStatus: http.StatusBadRequest,
			expectGroup:    false,
		},
		{
			name: "group not found",
			setupMock: func(ms *mocks.MockStore) {
				ms.On("GetGroupByID", mock.Anything, int64(999)).Return(db.Group{}, pgx.ErrNoRows)
			},
			pathValue:      "999",
			expectedStatus: http.StatusNotFound,
			expectGroup:    false,
		},
		{
			name: "database error",
			setupMock: func(ms *mocks.MockStore) {
				ms.On("GetGroupByID", mock.Anything, int64(1)).Return(db.Group{}, errors.New("database error"))
			},
			pathValue:      "1",
			expectedStatus: http.StatusInternalServerError,
			expectGroup:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := mocks.NewMockStore(t)
			tt.setupMock(mockStore)

			req := httptest.NewRequest("GET", "/groups/"+tt.pathValue, nil)
			req.SetPathValue("id", tt.pathValue)
			rr := httptest.NewRecorder()

			handler := getGroupByID(mockStore)
			handler(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.expectGroup {
				var groupResponse map[string]interface{}
				err := json.Unmarshal(rr.Body.Bytes(), &groupResponse)
				require.NoError(t, err)
				assert.NotNil(t, groupResponse["id"])
			}
			mockStore.AssertExpectations(t)
		})
	}
}

func TestCreateGroup(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(*mocks.MockStore)
		requestBody    interface{}
		expectedStatus int
		expectGroup    bool
	}{
		{
			name: "success",
			setupMock: func(ms *mocks.MockStore) {
				group := db.Group{ID: 1, Name: "New Group"}
				ms.On("CreateGroup", mock.Anything, "New Group").Return(group, nil)
			},
			requestBody:    map[string]string{"name": "New Group"},
			expectedStatus: http.StatusCreated,
			expectGroup:    true,
		},
		{
			name:           "missing name",
			setupMock:      func(ms *mocks.MockStore) {},
			requestBody:    map[string]string{},
			expectedStatus: http.StatusBadRequest,
			expectGroup:    false,
		},
		{
			name:           "empty name",
			setupMock:      func(ms *mocks.MockStore) {},
			requestBody:    map[string]string{"name": ""},
			expectedStatus: http.StatusBadRequest,
			expectGroup:    false,
		},
		{
			name:           "invalid JSON",
			setupMock:      func(ms *mocks.MockStore) {},
			requestBody:    "invalid json",
			expectedStatus: http.StatusBadRequest,
			expectGroup:    false,
		},
		{
			name: "database error",
			setupMock: func(ms *mocks.MockStore) {
				ms.On("CreateGroup", mock.Anything, "New Group").Return(db.Group{}, errors.New("database error"))
			},
			requestBody:    map[string]string{"name": "New Group"},
			expectedStatus: http.StatusInternalServerError,
			expectGroup:    false,
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

			req := httptest.NewRequest("POST", "/groups", bytes.NewBuffer(bodyBytes))
			rr := httptest.NewRecorder()

			handler := createGroup(mockStore)
			handler(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.expectGroup {
				var groupResponse map[string]interface{}
				err := json.Unmarshal(rr.Body.Bytes(), &groupResponse)
				require.NoError(t, err)
				assert.NotNil(t, groupResponse["id"])
			}
			mockStore.AssertExpectations(t)
		})
	}
}

func TestUpdateGroup(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(*mocks.MockStore)
		pathValue      string
		requestBody    interface{}
		expectedStatus int
		expectGroup    bool
	}{
		{
			name: "success",
			setupMock: func(ms *mocks.MockStore) {
				group := db.Group{ID: 1, Name: "Updated Group"}
				ms.On("UpdateGroup", mock.Anything, db.UpdateGroupParams{ID: 1, Name: "Updated Group"}).Return(group, nil)
			},
			pathValue:      "1",
			requestBody:    map[string]string{"name": "Updated Group"},
			expectedStatus: http.StatusOK,
			expectGroup:    true,
		},
		{
			name:           "invalid ID format",
			setupMock:      func(ms *mocks.MockStore) {},
			pathValue:      "abc",
			requestBody:    map[string]string{"name": "Updated Group"},
			expectedStatus: http.StatusBadRequest,
			expectGroup:    false,
		},
		{
			name:           "missing name",
			setupMock:      func(ms *mocks.MockStore) {},
			pathValue:      "1",
			requestBody:    map[string]string{},
			expectedStatus: http.StatusBadRequest,
			expectGroup:    false,
		},
		{
			name: "group not found",
			setupMock: func(ms *mocks.MockStore) {
				ms.On("UpdateGroup", mock.Anything, db.UpdateGroupParams{ID: 999, Name: "Updated Group"}).Return(db.Group{}, pgx.ErrNoRows)
			},
			pathValue:      "999",
			requestBody:    map[string]string{"name": "Updated Group"},
			expectedStatus: http.StatusNotFound,
			expectGroup:    false,
		},
		{
			name:           "invalid JSON",
			setupMock:      func(ms *mocks.MockStore) {},
			pathValue:      "1",
			requestBody:    "invalid json",
			expectedStatus: http.StatusBadRequest,
			expectGroup:    false,
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

			req := httptest.NewRequest("PUT", "/groups/"+tt.pathValue, bytes.NewBuffer(bodyBytes))
			req.SetPathValue("id", tt.pathValue)
			rr := httptest.NewRecorder()

			handler := updateGroup(mockStore)
			handler(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.expectGroup {
				var groupResponse map[string]interface{}
				err := json.Unmarshal(rr.Body.Bytes(), &groupResponse)
				require.NoError(t, err)
				assert.NotNil(t, groupResponse["id"])
			}
			mockStore.AssertExpectations(t)
		})
	}
}

func TestDeleteGroup(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(*mocks.MockStore)
		pathValue      string
		expectedStatus int
		expectGroup    bool
	}{
		{
			name: "success",
			setupMock: func(ms *mocks.MockStore) {
				group := db.Group{ID: 1, Name: "Group 1"}
				ms.On("DeleteGroup", mock.Anything, int64(1)).Return(group, nil)
			},
			pathValue:      "1",
			expectedStatus: http.StatusOK,
			expectGroup:    true,
		},
		{
			name:           "invalid ID format",
			setupMock:      func(ms *mocks.MockStore) {},
			pathValue:      "abc",
			expectedStatus: http.StatusBadRequest,
			expectGroup:    false,
		},
		{
			name: "group not found",
			setupMock: func(ms *mocks.MockStore) {
				ms.On("DeleteGroup", mock.Anything, int64(999)).Return(db.Group{}, pgx.ErrNoRows)
			},
			pathValue:      "999",
			expectedStatus: http.StatusNotFound,
			expectGroup:    false,
		},
		{
			name: "database error",
			setupMock: func(ms *mocks.MockStore) {
				ms.On("DeleteGroup", mock.Anything, int64(1)).Return(db.Group{}, errors.New("database error"))
			},
			pathValue:      "1",
			expectedStatus: http.StatusInternalServerError,
			expectGroup:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := mocks.NewMockStore(t)
			tt.setupMock(mockStore)

			req := httptest.NewRequest("DELETE", "/groups/"+tt.pathValue, nil)
			req.SetPathValue("id", tt.pathValue)
			rr := httptest.NewRecorder()

			handler := deleteGroup(mockStore)
			handler(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.expectGroup {
				var groupResponse map[string]interface{}
				err := json.Unmarshal(rr.Body.Bytes(), &groupResponse)
				require.NoError(t, err)
				assert.NotNil(t, groupResponse["id"])
			}
			mockStore.AssertExpectations(t)
		})
	}
}
