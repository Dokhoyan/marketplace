package user

import (
	"context"
	"fmt"

	"github.com/Dokhoyan/marketplace/internal/model"
)

func (s *serv) GetUsers(ctx context.Context) ([]model.User, error) {
	users, err := s.userRepo.GetUsers(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %w", err)
	}
	return users, nil
}
