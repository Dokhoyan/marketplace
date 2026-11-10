package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/Dokhoyan/marketplace/internal/model"
	passwordutil "github.com/Dokhoyan/marketplace/internal/password"
	"github.com/Dokhoyan/marketplace/internal/repository"
	"github.com/Dokhoyan/marketplace/internal/service"
)

// Проверка неизвестного логина тоже выполняет Argon2id.
// #nosec G101 -- Фиктивный хеш для проверки неизвестного логина, не учётные данные.
const dummyPasswordHash = "$argon2id$v=19$m=19456,t=2,p=1$AAAAAAAAAAAAAAAAAAAAAA$AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"

func (s *serv) Login(ctx context.Context, login, password string) (model.AccessToken, error) {
	if len(login) < 3 || len(login) > 32 || utf8.RuneCountInString(password) < 15 || utf8.RuneCountInString(password) > 128 {
		return model.AccessToken{}, service.ErrInvalidCredentials
	}
	credentials, err := s.userRepo.GetUserCredentials(ctx, strings.ToLower(login))
	if err != nil && !errors.Is(err, repository.ErrUserNotFound) {
		return model.AccessToken{}, fmt.Errorf("failed to get user credentials: %w", err)
	}
	passwordHash := credentials.PasswordHash
	if passwordHash == "" {
		passwordHash = dummyPasswordHash
	}
	passwordValid := passwordutil.Verify(password, passwordHash)
	if !passwordValid || err != nil || credentials.PasswordHash == "" {
		return model.AccessToken{}, service.ErrInvalidCredentials
	}
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return model.AccessToken{}, fmt.Errorf("failed to generate access token: %w", err)
	}
	accessToken := base64.RawURLEncoding.EncodeToString(tokenBytes)
	expiresAt, err := s.authRepo.CreateSession(ctx, credentials.User.ID, hashToken(accessToken))
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return model.AccessToken{}, service.ErrInvalidCredentials
		}
		return model.AccessToken{}, fmt.Errorf("failed to create session: %w", err)
	}

	return model.AccessToken{AccessToken: accessToken, TokenType: "Bearer", ExpiresAt: expiresAt}, nil
}
