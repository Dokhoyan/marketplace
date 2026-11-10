package auth

import (
	"crypto/sha256"
	"encoding/hex"

	"github.com/Dokhoyan/marketplace/internal/repository"
	"github.com/Dokhoyan/marketplace/internal/service"
)

type serv struct {
	authRepo repository.AuthRepository
	userRepo repository.UserRepository
}

func NewService(authRepo repository.AuthRepository, userRepo repository.UserRepository) service.AuthService {
	return &serv{authRepo: authRepo, userRepo: userRepo}
}

func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
