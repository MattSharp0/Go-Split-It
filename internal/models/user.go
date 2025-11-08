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

type CreateUserRequest struct {
	Name string `json:"name"`
}

type UpdateUserRequest struct {
	Name string `json:"name"`
}

type RegisterRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}
