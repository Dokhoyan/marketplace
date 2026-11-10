package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/Dokhoyan/marketplace/internal/model"
	"github.com/Dokhoyan/marketplace/internal/repository"
	"github.com/Dokhoyan/marketplace/internal/service"
)

func (s *serv) Authenticate(ctx context.Context, token string) (model.Session, error) {
	if len(token) != 43 {
		return model.Session{}, service.ErrUnauthorized
	}
	session, err := s.authRepo.GetSession(ctx, hashToken(token))
	if err != nil {
		if errors.Is(err, repository.ErrSessionNotFound) {
			return model.Session{}, service.ErrUnauthorized
		}
		return model.Session{}, fmt.Errorf("failed to get session: %w", err)
	}
	return session, nil
}
