package postgres

import (
	"context"

	"github.com/Dokhoyan/marketplace/internal/repository"
	txmpgx "github.com/avito-tech/go-transaction-manager/drivers/pgxv5/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db     *pgxpool.Pool
	getter *txmpgx.CtxGetter
}

var (
	_ repository.ProductRepository = (*Repository)(nil)
	_ repository.CartRepository    = (*Repository)(nil)
	_ repository.AuthRepository    = (*Repository)(nil)
	_ repository.UserRepository    = (*Repository)(nil)
)

func New(db *pgxpool.Pool, getter *txmpgx.CtxGetter) *Repository {
	return &Repository{db: db, getter: getter}
}

func (r *Repository) Close() {
	r.db.Close()
}

func (r *Repository) getConn(ctx context.Context) txmpgx.Tr {
	return r.getter.DefaultTrOrDB(ctx, r.db)
}
