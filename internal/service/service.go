package service

import (
	"context"
	"errors"

	"github.com/Dokhoyan/marketplace/internal/model"
)

var (
	ErrInvalidProductName = errors.New("название должно содержать от 1 до 200 символов")
	ErrInvalidPrice       = errors.New("цена должна быть положительным целым числом")
	ErrInvalidStock       = errors.New("остаток должен быть неотрицательным целым числом")
	ErrInvalidQuantity    = errors.New("количество должно быть положительным целым числом")
	ErrEmptyProductUpdate = errors.New("передайте name и/или price")
	ErrProductNotFound    = errors.New("товар не найден")
	ErrStockOverflow      = errors.New("остаток превышает допустимое значение")
	ErrInvalidLogin       = errors.New("логин должен содержать от 3 до 32 латинских букв, цифр или знаков подчеркивания")
	ErrInvalidPassword    = errors.New("пароль должен содержать от 15 до 128 символов")
	ErrLoginTaken         = errors.New("логин уже занят")
	ErrUserNotFound       = errors.New("пользователь не найден")
	ErrInvalidCredentials = errors.New("неверный логин или пароль")
	ErrUnauthorized       = errors.New("требуется действующий Bearer-токен")
	ErrInvalidName        = errors.New("имя должно содержать от 1 до 200 символов")
)

type ProductService interface {
	GetProducts(ctx context.Context) ([]model.Product, error)
	GetProduct(ctx context.Context, id int64) (model.Product, error)
	CreateProduct(ctx context.Context, name string, price, stock int64) (model.Product, error)
	UpdateProduct(ctx context.Context, id int64, name *string, price *int64) (model.Product, error)
	SetProductStock(ctx context.Context, id, stock int64) (model.Product, error)
	AddProductStock(ctx context.Context, id, quantity int64) (model.Product, error)
	DeleteProduct(ctx context.Context, id int64) error
}

type UserService interface {
	GetUser(ctx context.Context, id int64) (model.User, error)
	UpdateUser(ctx context.Context, id int64, name string) (model.User, error)
	DeleteUser(ctx context.Context, id int64) error
	GetUsers(ctx context.Context) ([]model.User, error)
	CreateUser(ctx context.Context, name, login, password string) (model.User, error)
}

type AuthService interface {
	Login(ctx context.Context, login, password string) (model.AccessToken, error)
	Authenticate(ctx context.Context, token string) (model.Session, error)
	Logout(ctx context.Context, tokenHash string) error
}
