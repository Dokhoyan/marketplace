package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/Dokhoyan/marketplace/internal/model"
	"github.com/Dokhoyan/marketplace/internal/repository"
	"github.com/Dokhoyan/marketplace/internal/service"
)

func (s *serv) GetUser(ctx context.Context, id int64) (model.User, error) {
	user, err := s.userRepo.GetUser(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return model.User{}, service.ErrUserNotFound
		}
		return model.User{}, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}
