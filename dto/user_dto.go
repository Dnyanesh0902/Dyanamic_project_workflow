package dto

import "time"

type UserCreateRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Rolename string `json:"role_name" binding:"required"`
}

type UserUpdateRequest struct {
	UUID     string `json:"uuid" binding:"required"`
	Name     string `json:"name"`
	Email    string `json:"email" binding:"omitempty,email"`
	Password string `json:"password" binding:"omitempty,min=6"`
	Rolename string `json:"role_name"`
}

type UserLoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type UserByUUIDRequest struct {
	UUID string `json:"uuid" binding:"required"`
}

type UserByEmailRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type UserResponse struct {
	ID        int        `json:"id"`
	UUID      string     `json:"uuid"`
	Name      string     `json:"name"`
	Email     string     `json:"email"`
	Rolename  string     `json:"role_name"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

type LoginResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}
