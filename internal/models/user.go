package models

import "time"

type UserResponse struct {
	Name       string    `json:"name"`
	ID         int64     `json:"id"`
	CreatedAt  time.Time `json:"created_at"`
	ModifiedAt time.Time `json:"modified_at"`
}

type ListUserResponse struct {
	Users  []UserResponse `json:"users"`
	Count  int32          `json:"count"`
	Limit  int32          `json:"limit"`
	Offset int32          `json:"offset"`
}
