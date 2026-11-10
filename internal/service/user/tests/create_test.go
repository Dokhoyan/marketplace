package user_test

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"
	"testing"

	"github.com/Dokhoyan/marketplace/internal/model"
	"github.com/Dokhoyan/marketplace/internal/repository"
	repomocks "github.com/Dokhoyan/marketplace/internal/repository/mocks"
	"github.com/Dokhoyan/marketplace/internal/service"
	servicemocks "github.com/Dokhoyan/marketplace/internal/service/mocks"
	"github.com/Dokhoyan/marketplace/internal/service/user"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/argon2"
)

func TestCreateUser(t *testing.T) {
	ctx := context.Background()
	repo := repomocks.NewMockUserRepository(t)
	expected := model.User{ID: 1, Name: "Анна", Login: "anna_1", Role: "user"}
	password := strings.Repeat("я", 128)
	var hashes []string
	repo.EXPECT().CreateUser(ctx, "Анна", "anna_1", mock.MatchedBy(func(encoded string) bool {
		parts := strings.Split(encoded, "$")
		if len(parts) != 6 || parts[1] != "argon2id" || parts[2] != "v=19" {
			return false
		}
		var memory, iterations uint32
		var threads uint8
		if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &memory, &iterations, &threads); err != nil {
			return false
		}
		salt, saltErr := base64.RawStdEncoding.DecodeString(parts[4])
		hash, hashErr := base64.RawStdEncoding.DecodeString(parts[5])
		if saltErr != nil || hashErr != nil || len(salt) < 16 || len(hash) != 32 {
			return false
		}
		actual := argon2.IDKey([]byte(password), salt, iterations, memory, threads, uint32(len(hash)))
		return string(actual) == string(hash)
	})).Run(func(_ context.Context, _, _, hash string) {
		hashes = append(hashes, hash)
	}).Return(expected, nil).Twice()

	cartRepo := repomocks.NewMockCartRepository(t)
	cartRepo.EXPECT().CreateCart(ctx, expected.ID).Return(nil).Twice()
	s := user.NewService(repo, cartRepo, nil, transactionManager(t))
	for range 2 {
		result, err := s.CreateUser(ctx, " Анна ", "Anna_1", password)
		require.NoError(t, err)
		require.Equal(t, expected, result)
	}
	require.Len(t, hashes, 2)
	require.NotEqual(t, hashes[0], hashes[1], "each password hash must use a fresh salt")
}

func TestCreateUserInvalidData(t *testing.T) {
	for _, tc := range []struct {
		name, userName, login, password string
		want                            error
	}{
		{"empty name", " ", "anna", "valid_password_123", service.ErrInvalidName},
		{"long name", strings.Repeat("я", 201), "anna", "valid_password_123", service.ErrInvalidName},
		{"short login", "Анна", "ab", "valid_password_123", service.ErrInvalidLogin},
		{"long login", "Анна", strings.Repeat("a", 33), "valid_password_123", service.ErrInvalidLogin},
		{"unicode login", "Анна", "анна", "valid_password_123", service.ErrInvalidLogin},
		{"login punctuation", "Анна", "anna-1", "valid_password_123", service.ErrInvalidLogin},
		{"login whitespace", "Анна", " anna ", "valid_password_123", service.ErrInvalidLogin},
		{"short password", "Анна", "anna", strings.Repeat("я", 14), service.ErrInvalidPassword},
		{"long password", "Анна", "anna", strings.Repeat("a", 129), service.ErrInvalidPassword},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo := repomocks.NewMockUserRepository(t)
			_, err := user.NewService(repo, nil, nil, nil).CreateUser(context.Background(), tc.userName, tc.login, tc.password)
			require.ErrorIs(t, err, tc.want)
		})
	}
}

func TestCreateUserLoginTaken(t *testing.T) {
	ctx := context.Background()
	repo := repomocks.NewMockUserRepository(t)
	repo.EXPECT().CreateUser(ctx, "Анна", "anna", mock.AnythingOfType("string")).Return(model.User{}, repository.ErrLoginTaken)
	_, err := user.NewService(repo, nil, nil, transactionManager(t)).CreateUser(ctx, "Анна", "ANNA", "valid_password_123")
	require.ErrorIs(t, err, service.ErrLoginTaken)
}

func transactionManager(t *testing.T) service.TrManager {
	trManager := servicemocks.NewMockTrManager(t)
	trManager.EXPECT().Do(mock.Anything, mock.Anything).RunAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
		return fn(ctx)
	}).Maybe()
	return trManager
}
