package user

import (
	"github.com/Dokhoyan/marketplace/internal/repository"
	"github.com/Dokhoyan/marketplace/internal/service"
)

type serv struct {
	userRepo  repository.UserRepository
	cartRepo  repository.CartRepository
	authRepo  repository.AuthRepository
	trManager service.TrManager
}

func NewService(
	userRepo repository.UserRepository,
	cartRepo repository.CartRepository,
	authRepo repository.AuthRepository,
	trManager service.TrManager,
) service.UserService {
	return &serv{
		userRepo:  userRepo,
		cartRepo:  cartRepo,
		authRepo:  authRepo,
		trManager: trManager,
	}
}
