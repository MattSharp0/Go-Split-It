package models

import "time"

type GroupMemberResponse struct {
	ID         int64     `json:"id"`
	GroupID    int64     `json:"group_id"`
	GroupName  string    `json:"group_name,omitempty"`
	MemberName *string   `json:"member_name"`
	UserID     *int64    `json:"user_id"`
	UserName   string    `json:"user_name,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

type ListGroupMemberResponse struct {
	GroupMembers []GroupMemberResponse `json:"group_members"`
	Count        int32                 `json:"count"`
	Limit        int32                 `json:"limit"`
	Offset       int32                 `json:"offset"`
}

type CreateGroupMemberRequest struct {
	GroupID int64  `json:"group_id"`
	UserID  *int64 `json:"user_id"`
}

type UpdateGroupMemberRequest struct {
	GroupID int64  `json:"group_id"`
	UserID  *int64 `json:"user_id"`
}

// Batch operation models
type BatchCreateGroupMemberRequest struct {
	Members []BatchGroupMemberItem `json:"members"`
}

type BatchGroupMemberItem struct {
	UserID *int64 `json:"user_id"`
}

type BatchUpdateGroupMemberRequest struct {
	Members []BatchGroupMemberItem `json:"members"`
}

type BatchCreateGroupMemberResponse struct {
	Group        GroupResponse         `json:"group"`
	GroupMembers []GroupMemberResponse `json:"group_members"`
	Count        int32                 `json:"count"`
}

type BatchUpdateGroupMemberResponse struct {
	Group          GroupResponse         `json:"group"`
	DeletedMembers []GroupMemberResponse `json:"deleted_members"`
	NewMembers     []GroupMemberResponse `json:"new_members"`
	DeletedCount   int32                 `json:"deleted_count"`
	NewCount       int32                 `json:"new_count"`
}

type BatchDeleteGroupMemberResponse struct {
	GroupID int64  `json:"group_id"`
	Message string `json:"message"`
}
