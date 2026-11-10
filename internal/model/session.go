package model

import "time"

type Session struct {
	UserID    int64
	Role      string
	TokenHash string
	ExpiresAt time.Time
}

type AccessToken struct {
	AccessToken string    `json:"access_token"`
	TokenType   string    `json:"token_type"`
	ExpiresAt   time.Time `json:"expires_at"`
}
