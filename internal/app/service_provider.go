package app

import (
	"context"
	"fmt"
	"time"

	"github.com/Dokhoyan/common/pkg/closer"
	"github.com/Dokhoyan/marketplace/internal/config"
	authImpl "github.com/Dokhoyan/marketplace/internal/http-server/handlers/auth"
	productImpl "github.com/Dokhoyan/marketplace/internal/http-server/handlers/product"
	userImpl "github.com/Dokhoyan/marketplace/internal/http-server/handlers/user"
	"github.com/Dokhoyan/marketplace/internal/http-server/middleware"
	postgresRepo "github.com/Dokhoyan/marketplace/internal/repository/postgres"
	"github.com/Dokhoyan/marketplace/internal/service"
	authserv "github.com/Dokhoyan/marketplace/internal/service/auth"
	productserv "github.com/Dokhoyan/marketplace/internal/service/product"
	userserv "github.com/Dokhoyan/marketplace/internal/service/user"
	txmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/avito-tech/go-transaction-manager/trm/v2/manager"
	"github.com/jackc/pgx/v5/pgxpool"
)

type serviceProvider struct {
	txManager      *manager.Manager
	authService    service.AuthService
	authImpl       *authImpl.Implementation
	authMiddleware *middleware.Auth
	pgConfig       config.PGConfig
	httpConfig     config.HTTPConfig
	dbPool         *pgxpool.Pool
	repo           *postgresRepo.Repository
	productService service.ProductService
	userService    service.UserService
	productImpl    *productImpl.Implementation
	userImpl       *userImpl.Implementation
}

func newServiceProvider(ctx context.Context) (*serviceProvider, error) {
	pgConfig, err := config.NewPGConfig()
	if err != nil {
		return nil, err
	}
	httpConfig, err := config.NewHTTPConfig()
	if err != nil {
		return nil, err
	}
	pool, err := pgxpool.New(ctx, pgConfig.DSN())
	if err != nil {
		return nil, fmt.Errorf("failed to create database pool: %w", err)
	}
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}
	closer.Add(func() error {
		pool.Close()
		return nil
	})

	return &serviceProvider{
		pgConfig:   pgConfig,
		httpConfig: httpConfig,
		dbPool:     pool,
	}, nil
}

func (s *serviceProvider) HTTPConfig() config.HTTPConfig {
	return s.httpConfig
}

func (s *serviceProvider) Repository() *postgresRepo.Repository {
	if s.repo == nil {
		s.repo = postgresRepo.New(s.dbPool, txmpgx.DefaultCtxGetter)
	}
	return s.repo
}

func (s *serviceProvider) ProductService() service.ProductService {
	if s.productService == nil {
		s.productService = productserv.NewService(s.Repository())
	}
	return s.productService
}

func (s *serviceProvider) UserService() service.UserService {
	if s.userService == nil {
		s.userService = userserv.NewService(s.Repository(), s.Repository(), s.Repository(), s.TxManager())
	}
	return s.userService
}

func (s *serviceProvider) ProductImpl() *productImpl.Implementation {
	if s.productImpl == nil {
		s.productImpl = productImpl.NewImplementation(s.ProductService())
	}
	return s.productImpl
}

func (s *serviceProvider) UserImpl() *userImpl.Implementation {
	if s.userImpl == nil {
		s.userImpl = userImpl.NewImplementation(s.UserService())
	}
	return s.userImpl
}

func (s *serviceProvider) AuthService() service.AuthService {
	if s.authService == nil {
		s.authService = authserv.NewService(s.Repository(), s.Repository())
	}
	return s.authService
}

func (s *serviceProvider) AuthImpl() *authImpl.Implementation {
	if s.authImpl == nil {
		s.authImpl = authImpl.NewImplementation(s.AuthService())
	}
	return s.authImpl
}

func (s *serviceProvider) AuthMiddleware() *middleware.Auth {
	if s.authMiddleware == nil {
		s.authMiddleware = middleware.NewAuth(s.AuthService())
	}
	return s.authMiddleware
}

func (s *serviceProvider) TxManager() *manager.Manager {
	if s.txManager == nil {
		s.txManager = manager.Must(txmpgx.NewDefaultFactory(s.dbPool))
	}
	return s.txManager
}
