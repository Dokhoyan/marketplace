package user

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/Dokhoyan/marketplace/internal/model"
	passwordutil "github.com/Dokhoyan/marketplace/internal/password"
	"github.com/Dokhoyan/marketplace/internal/repository"
	"github.com/Dokhoyan/marketplace/internal/service"
)

var loginPattern = regexp.MustCompile(`^[A-Za-z0-9_]{3,32}$`)

func (s *serv) CreateUser(ctx context.Context, name, login, password string) (model.User, error) {
	name = strings.TrimSpace(name)
	if length := utf8.RuneCountInString(name); length < 1 || length > 200 {
		return model.User{}, service.ErrInvalidName
	}
	if !loginPattern.MatchString(login) {
		return model.User{}, service.ErrInvalidLogin
	}
	if length := utf8.RuneCountInString(password); length < 15 || length > 128 {
		return model.User{}, service.ErrInvalidPassword
	}
	login = strings.ToLower(login)

	passwordHash, err := passwordutil.Hash(password)
	if err != nil {
		return model.User{}, fmt.Errorf("failed to hash password: %w", err)
	}
	var user model.User
	err = s.trManager.Do(ctx, func(ctx context.Context) error {
		var err error
		user, err = s.userRepo.CreateUser(ctx, name, login, passwordHash)
		if err != nil {
			if errors.Is(err, repository.ErrLoginTaken) {
				return service.ErrLoginTaken
			}
			return fmt.Errorf("failed to create user: %w", err)
		}
		if err := s.cartRepo.CreateCart(ctx, user.ID); err != nil {
			return fmt.Errorf("failed to create cart: %w", err)
		}
		return nil
	})
	if err != nil {
		return model.User{}, err
	}
	return user, nil
}
