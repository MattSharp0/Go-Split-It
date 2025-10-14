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
