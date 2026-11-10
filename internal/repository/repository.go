package repository

import (
	"context"
	"errors"
	"time"

	"github.com/Dokhoyan/marketplace/internal/model"
)

var (
	ErrProductNotFound = errors.New("товар не найден")
	ErrUserNotFound    = errors.New("пользователь не найден")
	ErrSessionNotFound = errors.New("сессия не найдена")
	ErrLoginTaken      = errors.New("логин уже занят")
	ErrStockOverflow   = errors.New("остаток превышает допустимое значение")
)

type ProductRepository interface {
	GetProducts(ctx context.Context) ([]model.Product, error)
	GetProduct(ctx context.Context, id int64) (model.Product, error)
	CreateProduct(ctx context.Context, name string, price, stock int64) (model.Product, error)
	UpdateProduct(ctx context.Context, id int64, name *string, price *int64) (model.Product, error)
	SetProductStock(ctx context.Context, id, stock int64) (model.Product, error)
	AddProductStock(ctx context.Context, id, quantity int64) (model.Product, error)
	DeleteProduct(ctx context.Context, id int64) error
}

type UserRepository interface {
	GetUser(ctx context.Context, id int64) (model.User, error)
	UpdateUser(ctx context.Context, id int64, name string) (model.User, error)
	DeleteUser(ctx context.Context, id int64) error
	GetUsers(ctx context.Context) ([]model.User, error)
	GetUserCredentials(ctx context.Context, login string) (model.UserCredentials, error)
	CreateUser(ctx context.Context, name, login, passwordHash string) (model.User, error)
}

type CartRepository interface {
	CreateCart(ctx context.Context, userID int64) error
}

type AuthRepository interface {
	DeleteUserSessions(ctx context.Context, userID int64) error
	CreateSession(ctx context.Context, userID int64, tokenHash string) (time.Time, error)
	GetSession(ctx context.Context, tokenHash string) (model.Session, error)
	DeleteSession(ctx context.Context, tokenHash string) error
}
