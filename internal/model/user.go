package model

import "time"

const (
	RoleUser  = "user"
	RoleAdmin = "admin"
)

type UserCredentials struct {
	User         User   `json:"-"`
	PasswordHash string `json:"-"`
}

type User struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Login     string    `json:"login"`
	Role      string    `json:"role"`
	Balance   int64     `json:"balance"`
	CreatedAt time.Time `json:"created_at"`
}
