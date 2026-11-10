package user_test

import (
	"context"
	"errors"
	"testing"

	"github.com/Dokhoyan/marketplace/internal/repository"
	repomocks "github.com/Dokhoyan/marketplace/internal/repository/mocks"
	"github.com/Dokhoyan/marketplace/internal/service"
	"github.com/Dokhoyan/marketplace/internal/service/user"
	"github.com/stretchr/testify/require"
)

func TestDeleteUser(t *testing.T) {
	ctx := context.Background()
	userRepo := repomocks.NewMockUserRepository(t)
	authRepo := repomocks.NewMockAuthRepository(t)
	userRepo.EXPECT().DeleteUser(ctx, int64(1)).Return(nil)
	authRepo.EXPECT().DeleteUserSessions(ctx, int64(1)).Return(nil)
	users := user.NewService(userRepo, nil, authRepo, transactionManager(t))
	require.NoError(t, users.DeleteUser(ctx, 1))
}

func TestDeleteUserNotFound(t *testing.T) {
	ctx := context.Background()
	userRepo := repomocks.NewMockUserRepository(t)
	authRepo := repomocks.NewMockAuthRepository(t)
	userRepo.EXPECT().DeleteUser(ctx, int64(1)).Return(repository.ErrUserNotFound)
	users := user.NewService(userRepo, nil, authRepo, transactionManager(t))
	require.ErrorIs(t, users.DeleteUser(ctx, 1), service.ErrUserNotFound)
}

func TestDeleteUserSessionFailure(t *testing.T) {
	ctx := context.Background()
	userRepo := repomocks.NewMockUserRepository(t)
	authRepo := repomocks.NewMockAuthRepository(t)
	dbErr := errors.New("session deletion failed")
	userRepo.EXPECT().DeleteUser(ctx, int64(1)).Return(nil)
	authRepo.EXPECT().DeleteUserSessions(ctx, int64(1)).Return(dbErr)
	users := user.NewService(userRepo, nil, authRepo, transactionManager(t))
	require.ErrorIs(t, users.DeleteUser(ctx, 1), dbErr)
}
