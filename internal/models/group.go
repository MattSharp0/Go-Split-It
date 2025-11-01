package models

type GroupResponse struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type ListGroupResponse struct {
	Groups []GroupResponse `json:"groups"`
	Count  int32           `json:"count"`
	Limit  int32           `json:"limit"`
	Offset int32           `json:"offset"`
}

type CreateGroupRequest struct {
	Name string `json:"name"`
}

type UpdateGroupRequest struct {
	Name string `json:"name"`
}
