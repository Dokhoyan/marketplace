package user

import (
	"context"
	"errors"
	"fmt"

	"github.com/Dokhoyan/marketplace/internal/repository"
	"github.com/Dokhoyan/marketplace/internal/service"
)

func (s *serv) DeleteUser(ctx context.Context, id int64) error {
	err := s.trManager.Do(ctx, func(ctx context.Context) error {
		if err := s.userRepo.DeleteUser(ctx, id); err != nil {
			if errors.Is(err, repository.ErrUserNotFound) {
				return service.ErrUserNotFound
			}
			return fmt.Errorf("failed to delete user: %w", err)
		}
		if err := s.authRepo.DeleteUserSessions(ctx, id); err != nil {
			return fmt.Errorf("failed to delete user sessions: %w", err)
		}
		return nil
	})
	return err
}
