package user

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/Dokhoyan/marketplace/internal/model"
	"github.com/Dokhoyan/marketplace/internal/repository"
	"github.com/Dokhoyan/marketplace/internal/service"
)

func (s *serv) UpdateUser(ctx context.Context, id int64, name string) (model.User, error) {
	name = strings.TrimSpace(name)
	if length := utf8.RuneCountInString(name); length < 1 || length > 200 {
		return model.User{}, service.ErrInvalidName
	}
	user, err := s.userRepo.UpdateUser(ctx, id, name)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return model.User{}, service.ErrUserNotFound
		}
		return model.User{}, fmt.Errorf("failed to update user: %w", err)
	}

	return user, nil
}
