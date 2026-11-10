package auth

import (
	"context"
	"fmt"
)

func (s *serv) Logout(ctx context.Context, tokenHash string) error {
	if err := s.authRepo.DeleteSession(ctx, tokenHash); err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}
	return nil
}
