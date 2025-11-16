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
				userID := int64Ptr(1)
				ms.On("ListGroupsByUser", mock.Anything, db.ListGroupsByUserParams{UserID: userID, Limit: 100, Offset: 0}).Return(groups, nil)
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
				userID := int64Ptr(1)
				ms.On("ListGroupsByUser", mock.Anything, db.ListGroupsByUserParams{UserID: userID, Limit: 50, Offset: 10}).Return(groups, nil)
			},
			requestURL:     "/groups?limit=50&offset=10",
			expectedStatus: http.StatusOK,
			expectedCount:  1,
		},
		{
			name: "empty list",
			setupMock: func(ms *mocks.MockStore) {
				userID := int64Ptr(1)
				ms.On("ListGroupsByUser", mock.Anything, db.ListGroupsByUserParams{UserID: userID, Limit: 100, Offset: 0}).Return([]db.Group{}, nil)
			},
			requestURL:     "/groups",
			expectedStatus: http.StatusOK,
			expectedCount:  0,
		},
		{
			name: "database error",
			setupMock: func(ms *mocks.MockStore) {
				userID := int64Ptr(1)
				ms.On("ListGroupsByUser", mock.Anything, db.ListGroupsByUserParams{UserID: userID, Limit: 100, Offset: 0}).Return(nil, errors.New("database error"))
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

			req := createRequestWithUserID("GET", tt.requestURL, nil, 1)
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
				// Mock group membership check
				userID := int64Ptr(1)
				members := []db.ListGroupMembersByGroupIDRow{
					{ID: 1, GroupID: 1, UserID: userID},
				}
				ms.On("ListGroupMembersByGroupID", mock.Anything, db.ListGroupMembersByGroupIDParams{GroupID: 1, Limit: 1000, Offset: 0}).Return(members, nil)
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

			req := createRequestWithUserID("GET", "/groups/"+tt.pathValue, nil, 1)
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
				// Automatically add creator as group member
				userID := int64Ptr(1)
				groupMember := db.GroupMember{
					ID:        1,
					GroupID:   1,
					UserID:    userID,
					CreatedAt: time.Now(),
				}
				ms.On("CreateGroupMember", mock.Anything, db.CreateGroupMemberParams{GroupID: 1, UserID: userID}).Return(groupMember, nil)
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
				// Note: CreateGroupMember won't be called if CreateGroup fails
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

			req := createRequestWithUserID("POST", "/groups", bodyBytes, 1)
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
				// Mock group membership check
				userID := int64Ptr(1)
				members := []db.ListGroupMembersByGroupIDRow{
					{ID: 1, GroupID: 1, UserID: userID},
				}
				ms.On("ListGroupMembersByGroupID", mock.Anything, db.ListGroupMembersByGroupIDParams{GroupID: 1, Limit: 1000, Offset: 0}).Return(members, nil)
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
			name: "missing name",
			setupMock: func(ms *mocks.MockStore) {
				// Mock group membership check (handler checks this before decoding body)
				userID := int64Ptr(1)
				members := []db.ListGroupMembersByGroupIDRow{
					{ID: 1, GroupID: 1, UserID: userID},
				}
				ms.On("ListGroupMembersByGroupID", mock.Anything, db.ListGroupMembersByGroupIDParams{GroupID: 1, Limit: 1000, Offset: 0}).Return(members, nil)
			},
			pathValue:      "1",
			requestBody:    map[string]string{},
			expectedStatus: http.StatusBadRequest,
			expectGroup:    false,
		},
		{
			name: "group not found",
			setupMock: func(ms *mocks.MockStore) {
				// Mock group membership check first
				userID := int64Ptr(1)
				members := []db.ListGroupMembersByGroupIDRow{
					{ID: 1, GroupID: 999, UserID: userID},
				}
				ms.On("ListGroupMembersByGroupID", mock.Anything, db.ListGroupMembersByGroupIDParams{GroupID: 999, Limit: 1000, Offset: 0}).Return(members, nil)
				ms.On("UpdateGroup", mock.Anything, db.UpdateGroupParams{ID: 999, Name: "Updated Group"}).Return(db.Group{}, pgx.ErrNoRows)
			},
			pathValue:      "999",
			requestBody:    map[string]string{"name": "Updated Group"},
			expectedStatus: http.StatusNotFound,
			expectGroup:    false,
		},
		{
			name: "invalid JSON",
			setupMock: func(ms *mocks.MockStore) {
				// Mock group membership check (handler checks this before decoding body)
				userID := int64Ptr(1)
				members := []db.ListGroupMembersByGroupIDRow{
					{ID: 1, GroupID: 1, UserID: userID},
				}
				ms.On("ListGroupMembersByGroupID", mock.Anything, db.ListGroupMembersByGroupIDParams{GroupID: 1, Limit: 1000, Offset: 0}).Return(members, nil)
			},
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

			req := createRequestWithUserID("PUT", "/groups/"+tt.pathValue, bodyBytes, 1)
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
				// Mock group membership check
				userID := int64Ptr(1)
				members := []db.ListGroupMembersByGroupIDRow{
					{ID: 1, GroupID: 1, UserID: userID},
				}
				ms.On("ListGroupMembersByGroupID", mock.Anything, db.ListGroupMembersByGroupIDParams{GroupID: 1, Limit: 1000, Offset: 0}).Return(members, nil)
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
				// Mock group membership check first
				userID := int64Ptr(1)
				members := []db.ListGroupMembersByGroupIDRow{
					{ID: 1, GroupID: 999, UserID: userID},
				}
				ms.On("ListGroupMembersByGroupID", mock.Anything, db.ListGroupMembersByGroupIDParams{GroupID: 999, Limit: 1000, Offset: 0}).Return(members, nil)
				ms.On("DeleteGroup", mock.Anything, int64(999)).Return(db.Group{}, pgx.ErrNoRows)
			},
			pathValue:      "999",
			expectedStatus: http.StatusNotFound,
			expectGroup:    false,
		},
		{
			name: "database error",
			setupMock: func(ms *mocks.MockStore) {
				// Mock group membership check first
				userID := int64Ptr(1)
				members := []db.ListGroupMembersByGroupIDRow{
					{ID: 1, GroupID: 1, UserID: userID},
				}
				ms.On("ListGroupMembersByGroupID", mock.Anything, db.ListGroupMembersByGroupIDParams{GroupID: 1, Limit: 1000, Offset: 0}).Return(members, nil)
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

			req := createRequestWithUserID("DELETE", "/groups/"+tt.pathValue, nil, 1)
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
