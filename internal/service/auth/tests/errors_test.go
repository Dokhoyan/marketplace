package auth_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/Dokhoyan/marketplace/internal/model"
	"github.com/Dokhoyan/marketplace/internal/repository"
	repomocks "github.com/Dokhoyan/marketplace/internal/repository/mocks"
	"github.com/Dokhoyan/marketplace/internal/service"
	"github.com/Dokhoyan/marketplace/internal/service/auth"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestLoginDatabaseError(t *testing.T) {
	ctx := context.Background()
	userRepo := repomocks.NewMockUserRepository(t)
	dbErr := errors.New("database unavailable")
	userRepo.EXPECT().GetUserCredentials(ctx, "anna").Return(model.UserCredentials{}, dbErr)
	_, err := auth.NewService(nil, userRepo).Login(ctx, "anna", "valid_password_123")
	require.ErrorIs(t, err, dbErr)
	require.ErrorContains(t, err, "failed to get user credentials")
}

func TestAuthenticateErrors(t *testing.T) {
	ctx := context.Background()
	dbErr := errors.New("database unavailable")
	for _, tc := range []struct {
		err, want error
		message   string
	}{
		{repository.ErrSessionNotFound, service.ErrUnauthorized, ""},
		{dbErr, dbErr, "failed to get session"},
	} {
		authRepo := repomocks.NewMockAuthRepository(t)
		authRepo.EXPECT().GetSession(ctx, mock.AnythingOfType("string")).Return(model.Session{}, tc.err)
		_, err := auth.NewService(authRepo, nil).Authenticate(ctx, strings.Repeat("a", 43))
		require.ErrorIs(t, err, tc.want)
		if tc.message != "" {
			require.ErrorContains(t, err, tc.message)
		}
	}
}

func TestLogoutDatabaseError(t *testing.T) {
	ctx := context.Background()
	authRepo := repomocks.NewMockAuthRepository(t)
	dbErr := errors.New("database unavailable")
	authRepo.EXPECT().DeleteSession(ctx, "hash").Return(dbErr)
	err := auth.NewService(authRepo, nil).Logout(ctx, "hash")
	require.ErrorIs(t, err, dbErr)
	require.ErrorContains(t, err, "failed to delete session")
}
