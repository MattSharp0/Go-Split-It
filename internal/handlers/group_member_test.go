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

func TestListGroupMembersByGroupID(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(*mocks.MockStore)
		pathValue      string
		queryParams    string
		expectedStatus int
		expectedCount  int
	}{
		{
			name: "success with members",
			setupMock: func(ms *mocks.MockStore) {
				// Mock group membership check
				userID := int64Ptr(1)
				members := []db.ListGroupMembersByGroupIDRow{
					{
						ID:         1,
						GroupID:    1,
						GroupName:  "Group 1",
						MemberName: stringPtr("Member 1"),
						UserID:     userID,
						UserName:   "User 1",
						CreatedAt:  time.Now(),
					},
				}
				ms.On("ListGroupMembersByGroupID", mock.Anything, db.ListGroupMembersByGroupIDParams{GroupID: 1, Limit: 1000, Offset: 0}).Return(members, nil)
				ms.On("ListGroupMembersByGroupID", mock.Anything, db.ListGroupMembersByGroupIDParams{GroupID: 1, Limit: 100, Offset: 0}).Return(members, nil)
			},
			pathValue:      "1",
			queryParams:    "",
			expectedStatus: http.StatusOK,
			expectedCount:  1,
		},
		{
			name: "success with empty members",
			setupMock: func(ms *mocks.MockStore) {
				// Mock group membership check
				userID := int64Ptr(1)
				members := []db.ListGroupMembersByGroupIDRow{
					{ID: 1, GroupID: 1, UserID: userID},
				}
				ms.On("ListGroupMembersByGroupID", mock.Anything, db.ListGroupMembersByGroupIDParams{GroupID: 1, Limit: 1000, Offset: 0}).Return(members, nil)
				ms.On("ListGroupMembersByGroupID", mock.Anything, db.ListGroupMembersByGroupIDParams{GroupID: 1, Limit: 100, Offset: 0}).Return([]db.ListGroupMembersByGroupIDRow{}, nil)
			},
			pathValue:      "1",
			queryParams:    "",
			expectedStatus: http.StatusOK,
			expectedCount:  0,
		},
		{
			name:           "invalid group ID format",
			setupMock:      func(ms *mocks.MockStore) {},
			pathValue:      "abc",
			queryParams:    "",
			expectedStatus: http.StatusBadRequest,
			expectedCount:  0,
		},
		{
			name: "database error",
			setupMock: func(ms *mocks.MockStore) {
				// Mock group membership check
				userID := int64Ptr(1)
				members := []db.ListGroupMembersByGroupIDRow{
					{ID: 1, GroupID: 1, UserID: userID},
				}
				ms.On("ListGroupMembersByGroupID", mock.Anything, db.ListGroupMembersByGroupIDParams{GroupID: 1, Limit: 1000, Offset: 0}).Return(members, nil)
				ms.On("ListGroupMembersByGroupID", mock.Anything, db.ListGroupMembersByGroupIDParams{GroupID: 1, Limit: 100, Offset: 0}).Return(nil, errors.New("database error"))
			},
			pathValue:      "1",
			queryParams:    "",
			expectedStatus: http.StatusInternalServerError,
			expectedCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := mocks.NewMockStore(t)
			tt.setupMock(mockStore)

			url := "/group_members/group/" + tt.pathValue
			if tt.queryParams != "" {
				url += "?" + tt.queryParams
			}
			req := createRequestWithUserID("GET", url, nil, 1)
			req.SetPathValue("group_id", tt.pathValue)
			rr := httptest.NewRecorder()

			handler := listGroupMembersByGroupID(mockStore)
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

func TestGetGroupMemberByID(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(*mocks.MockStore)
		pathValue      string
		expectedStatus int
		expectMember   bool
	}{
		{
			name: "success",
			setupMock: func(ms *mocks.MockStore) {
				member := db.GetGroupMemberByIDRow{
					ID:         1,
					GroupID:    1,
					GroupName:  "Group 1",
					MemberName: stringPtr("Member 1"),
					UserID:     int64Ptr(1),
					UserName:   "User 1",
					CreatedAt:  time.Now(),
				}
				ms.On("GetGroupMemberByID", mock.Anything, int64(1)).Return(member, nil)
				// Mock group membership check
				userID := int64Ptr(1)
				members := []db.ListGroupMembersByGroupIDRow{
					{ID: 1, GroupID: 1, UserID: userID},
				}
				ms.On("ListGroupMembersByGroupID", mock.Anything, db.ListGroupMembersByGroupIDParams{GroupID: 1, Limit: 1000, Offset: 0}).Return(members, nil)
			},
			pathValue:      "1",
			expectedStatus: http.StatusOK,
			expectMember:   true,
		},
		{
			name:           "invalid ID format",
			setupMock:      func(ms *mocks.MockStore) {},
			pathValue:      "abc",
			expectedStatus: http.StatusBadRequest,
			expectMember:   false,
		},
		{
			name: "member not found",
			setupMock: func(ms *mocks.MockStore) {
				ms.On("GetGroupMemberByID", mock.Anything, int64(999)).Return(db.GetGroupMemberByIDRow{}, pgx.ErrNoRows)
			},
			pathValue:      "999",
			expectedStatus: http.StatusNotFound,
			expectMember:   false,
		},
		{
			name: "database error",
			setupMock: func(ms *mocks.MockStore) {
				ms.On("GetGroupMemberByID", mock.Anything, int64(1)).Return(db.GetGroupMemberByIDRow{}, errors.New("database error"))
			},
			pathValue:      "1",
			expectedStatus: http.StatusInternalServerError,
			expectMember:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := mocks.NewMockStore(t)
			tt.setupMock(mockStore)

			req := createRequestWithUserID("GET", "/group_members/"+tt.pathValue, nil, 1)
			req.SetPathValue("id", tt.pathValue)
			rr := httptest.NewRecorder()

			handler := getGroupMemberByID(mockStore)
			handler(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.expectMember {
				var memberResponse map[string]interface{}
				err := json.Unmarshal(rr.Body.Bytes(), &memberResponse)
				require.NoError(t, err)
				assert.NotNil(t, memberResponse["id"])
			}
			mockStore.AssertExpectations(t)
		})
	}
}

func TestCreateGroupMember(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(*mocks.MockStore)
		requestBody    interface{}
		expectedStatus int
		expectMember   bool
	}{
		{
			name: "success with user ID",
			setupMock: func(ms *mocks.MockStore) {
				// Mock group membership check
				userID := int64Ptr(1)
				members := []db.ListGroupMembersByGroupIDRow{
					{ID: 1, GroupID: 1, UserID: userID},
				}
				ms.On("ListGroupMembersByGroupID", mock.Anything, db.ListGroupMembersByGroupIDParams{GroupID: 1, Limit: 1000, Offset: 0}).Return(members, nil)
				userIDVal := int64(1)
				member := db.GroupMember{
					ID:        1,
					GroupID:   1,
					UserID:    &userIDVal,
					CreatedAt: time.Now(),
				}
				ms.On("CreateGroupMember", mock.Anything, db.CreateGroupMemberParams{GroupID: 1, UserID: &userIDVal}).Return(member, nil)
			},
			requestBody:    map[string]interface{}{"group_id": 1, "user_id": 1},
			expectedStatus: http.StatusCreated,
			expectMember:   true,
		},
		{
			name: "success without user ID",
			setupMock: func(ms *mocks.MockStore) {
				// Mock group membership check
				userID := int64Ptr(1)
				members := []db.ListGroupMembersByGroupIDRow{
					{ID: 1, GroupID: 1, UserID: userID},
				}
				ms.On("ListGroupMembersByGroupID", mock.Anything, db.ListGroupMembersByGroupIDParams{GroupID: 1, Limit: 1000, Offset: 0}).Return(members, nil)
				member := db.GroupMember{
					ID:        1,
					GroupID:   1,
					UserID:    nil,
					CreatedAt: time.Now(),
				}
				ms.On("CreateGroupMember", mock.Anything, db.CreateGroupMemberParams{GroupID: 1, UserID: nil}).Return(member, nil)
			},
			requestBody:    map[string]interface{}{"group_id": 1, "user_id": nil},
			expectedStatus: http.StatusCreated,
			expectMember:   true,
		},
		{
			name:           "missing group_id",
			setupMock:      func(ms *mocks.MockStore) {},
			requestBody:    map[string]interface{}{"user_id": 1},
			expectedStatus: http.StatusBadRequest,
			expectMember:   false,
		},
		{
			name:           "invalid JSON",
			setupMock:      func(ms *mocks.MockStore) {},
			requestBody:    "invalid json",
			expectedStatus: http.StatusBadRequest,
			expectMember:   false,
		},
		{
			name: "database error",
			setupMock: func(ms *mocks.MockStore) {
				// Mock group membership check
				userIDPtr := int64Ptr(1)
				members := []db.ListGroupMembersByGroupIDRow{
					{ID: 1, GroupID: 1, UserID: userIDPtr},
				}
				ms.On("ListGroupMembersByGroupID", mock.Anything, db.ListGroupMembersByGroupIDParams{GroupID: 1, Limit: 1000, Offset: 0}).Return(members, nil)
				userID := int64(1)
				ms.On("CreateGroupMember", mock.Anything, db.CreateGroupMemberParams{GroupID: 1, UserID: &userID}).Return(db.GroupMember{}, errors.New("database error"))
			},
			requestBody:    map[string]interface{}{"group_id": 1, "user_id": 1},
			expectedStatus: http.StatusInternalServerError,
			expectMember:   false,
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

			req := createRequestWithUserID("POST", "/group_members", bodyBytes, 1)
			rr := httptest.NewRecorder()

			handler := createGroupMember(mockStore)
			handler(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.expectMember {
				var memberResponse map[string]interface{}
				err := json.Unmarshal(rr.Body.Bytes(), &memberResponse)
				require.NoError(t, err)
				assert.NotNil(t, memberResponse["id"])
			}
			mockStore.AssertExpectations(t)
		})
	}
}

func TestUpdateGroupMember(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(*mocks.MockStore)
		pathValue      string
		requestBody    interface{}
		expectedStatus int
		expectMember   bool
	}{
		{
			name: "success",
			setupMock: func(ms *mocks.MockStore) {
				// Get group member first
				groupMemberRow := db.GetGroupMemberByIDRow{
					ID:      1,
					GroupID: 1,
				}
				ms.On("GetGroupMemberByID", mock.Anything, int64(1)).Return(groupMemberRow, nil)
				// Mock group membership check
				userID := int64Ptr(1)
				members := []db.ListGroupMembersByGroupIDRow{
					{ID: 1, GroupID: 1, UserID: userID},
				}
				ms.On("ListGroupMembersByGroupID", mock.Anything, db.ListGroupMembersByGroupIDParams{GroupID: 1, Limit: 1000, Offset: 0}).Return(members, nil)
				userIDVal := int64(2)
				member := db.GroupMember{
					ID:        1,
					GroupID:   1,
					UserID:    &userIDVal,
					CreatedAt: time.Now(),
				}
				ms.On("UpdateGroupMember", mock.Anything, db.UpdateGroupMemberParams{ID: 1, GroupID: 1, UserID: &userIDVal}).Return(member, nil)
			},
			pathValue:      "1",
			requestBody:    map[string]interface{}{"group_id": 1, "user_id": 2},
			expectedStatus: http.StatusOK,
			expectMember:   true,
		},
		{
			name:           "invalid ID format",
			setupMock:      func(ms *mocks.MockStore) {},
			pathValue:      "abc",
			requestBody:    map[string]interface{}{"group_id": 1, "user_id": 1},
			expectedStatus: http.StatusBadRequest,
			expectMember:   false,
		},
		{
			name: "missing group_id",
			setupMock: func(ms *mocks.MockStore) {
				// Get group member first
				groupMemberRow := db.GetGroupMemberByIDRow{
					ID:      1,
					GroupID: 1,
				}
				ms.On("GetGroupMemberByID", mock.Anything, int64(1)).Return(groupMemberRow, nil)
				// Mock group membership check
				userID := int64Ptr(1)
				members := []db.ListGroupMembersByGroupIDRow{
					{ID: 1, GroupID: 1, UserID: userID},
				}
				ms.On("ListGroupMembersByGroupID", mock.Anything, db.ListGroupMembersByGroupIDParams{GroupID: 1, Limit: 1000, Offset: 0}).Return(members, nil)
			},
			pathValue:      "1",
			requestBody:    map[string]interface{}{"user_id": 1},
			expectedStatus: http.StatusBadRequest,
			expectMember:   false,
		},
		{
			name: "member not found",
			setupMock: func(ms *mocks.MockStore) {
				ms.On("GetGroupMemberByID", mock.Anything, int64(999)).Return(db.GetGroupMemberByIDRow{}, pgx.ErrNoRows)
			},
			pathValue:      "999",
			requestBody:    map[string]interface{}{"group_id": 1, "user_id": 1},
			expectedStatus: http.StatusNotFound,
			expectMember:   false,
		},
		{
			name: "invalid JSON",
			setupMock: func(ms *mocks.MockStore) {
				// Get group member first
				groupMemberRow := db.GetGroupMemberByIDRow{
					ID:      1,
					GroupID: 1,
				}
				ms.On("GetGroupMemberByID", mock.Anything, int64(1)).Return(groupMemberRow, nil)
				// Mock group membership check
				userID := int64Ptr(1)
				members := []db.ListGroupMembersByGroupIDRow{
					{ID: 1, GroupID: 1, UserID: userID},
				}
				ms.On("ListGroupMembersByGroupID", mock.Anything, db.ListGroupMembersByGroupIDParams{GroupID: 1, Limit: 1000, Offset: 0}).Return(members, nil)
			},
			pathValue:      "1",
			requestBody:    "invalid json",
			expectedStatus: http.StatusBadRequest,
			expectMember:   false,
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

			req := createRequestWithUserID("PUT", "/group_members/"+tt.pathValue, bodyBytes, 1)
			req.SetPathValue("id", tt.pathValue)
			rr := httptest.NewRecorder()

			handler := updateGroupMember(mockStore)
			handler(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.expectMember {
				var memberResponse map[string]interface{}
				err := json.Unmarshal(rr.Body.Bytes(), &memberResponse)
				require.NoError(t, err)
				assert.NotNil(t, memberResponse["id"])
			}
			mockStore.AssertExpectations(t)
		})
	}
}

func TestDeleteGroupMember(t *testing.T) {
	tests := []struct {
		name           string
		setupMock      func(*mocks.MockStore)
		pathValue      string
		expectedStatus int
		expectMember   bool
	}{
		{
			name: "success",
			setupMock: func(ms *mocks.MockStore) {
				// Get group member first
				groupMemberRow := db.GetGroupMemberByIDRow{
					ID:      1,
					GroupID: 1,
				}
				ms.On("GetGroupMemberByID", mock.Anything, int64(1)).Return(groupMemberRow, nil)
				// Mock group membership check
				userID := int64Ptr(1)
				members := []db.ListGroupMembersByGroupIDRow{
					{ID: 1, GroupID: 1, UserID: userID},
				}
				ms.On("ListGroupMembersByGroupID", mock.Anything, db.ListGroupMembersByGroupIDParams{GroupID: 1, Limit: 1000, Offset: 0}).Return(members, nil)
				userIDVal := int64(1)
				member := db.GroupMember{
					ID:        1,
					GroupID:   1,
					UserID:    &userIDVal,
					CreatedAt: time.Now(),
				}
				ms.On("UnlinkGroupMember", mock.Anything, int64(1)).Return(member, nil)
			},
			pathValue:      "1",
			expectedStatus: http.StatusOK,
			expectMember:   true,
		},
		{
			name:           "invalid ID format",
			setupMock:      func(ms *mocks.MockStore) {},
			pathValue:      "abc",
			expectedStatus: http.StatusBadRequest,
			expectMember:   false,
		},
		{
			name: "member not found",
			setupMock: func(ms *mocks.MockStore) {
				ms.On("GetGroupMemberByID", mock.Anything, int64(999)).Return(db.GetGroupMemberByIDRow{}, pgx.ErrNoRows)
			},
			pathValue:      "999",
			expectedStatus: http.StatusNotFound,
			expectMember:   false,
		},
		{
			name: "database error",
			setupMock: func(ms *mocks.MockStore) {
				groupMemberRow := db.GetGroupMemberByIDRow{
					ID:      1,
					GroupID: 1,
				}
				ms.On("GetGroupMemberByID", mock.Anything, int64(1)).Return(groupMemberRow, nil)
				// Mock group membership check
				userID := int64Ptr(1)
				members := []db.ListGroupMembersByGroupIDRow{
					{ID: 1, GroupID: 1, UserID: userID},
				}
				ms.On("ListGroupMembersByGroupID", mock.Anything, db.ListGroupMembersByGroupIDParams{GroupID: 1, Limit: 1000, Offset: 0}).Return(members, nil)
				ms.On("UnlinkGroupMember", mock.Anything, int64(1)).Return(db.GroupMember{}, errors.New("database error"))
			},
			pathValue:      "1",
			expectedStatus: http.StatusInternalServerError,
			expectMember:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStore := mocks.NewMockStore(t)
			tt.setupMock(mockStore)

			req := createRequestWithUserID("DELETE", "/group_members/"+tt.pathValue, nil, 1)
			req.SetPathValue("id", tt.pathValue)
			rr := httptest.NewRecorder()

			handler := deleteGroupMember(mockStore)
			handler(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.expectMember {
				var memberResponse map[string]interface{}
				err := json.Unmarshal(rr.Body.Bytes(), &memberResponse)
				require.NoError(t, err)
				assert.NotNil(t, memberResponse["id"])
			}
			mockStore.AssertExpectations(t)
		})
	}
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}
