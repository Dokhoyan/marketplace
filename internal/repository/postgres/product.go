package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/Dokhoyan/marketplace/internal/model"
	"github.com/Dokhoyan/marketplace/internal/repository"
	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func (r *Repository) GetProducts(ctx context.Context) ([]model.Product, error) {
	builder := sq.Select("id", "name", "price", "stock").
		PlaceholderFormat(sq.Dollar).
		From("products").
		Where(sq.Eq{"deleted_at": nil}).
		OrderBy("id")

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build select query: %w", err)
	}

	conn := r.getConn(ctx)

	rows, err := conn.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get products: %w", err)
	}
	defer rows.Close()

	products := make([]model.Product, 0)
	for rows.Next() {
		var product model.Product
		if err := rows.Scan(&product.ID, &product.Name, &product.Price, &product.Stock); err != nil {
			return nil, fmt.Errorf("failed to scan product: %w", err)
		}
		products = append(products, product)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate products: %w", err)
	}

	return products, nil
}

func (r *Repository) GetProduct(ctx context.Context, id int64) (model.Product, error) {
	builder := sq.Select("id", "name", "price", "stock").
		PlaceholderFormat(sq.Dollar).
		From("products").
		Where(sq.Eq{"id": id, "deleted_at": nil})

	query, args, err := builder.ToSql()
	if err != nil {
		return model.Product{}, fmt.Errorf("failed to build select query: %w", err)
	}

	conn := r.getConn(ctx)

	var product model.Product
	err = conn.QueryRow(ctx, query, args...).Scan(&product.ID, &product.Name, &product.Price, &product.Stock)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Product{}, repository.ErrProductNotFound
		}
		return model.Product{}, fmt.Errorf("failed to get product: %w", err)
	}

	return product, nil
}

func (r *Repository) CreateProduct(ctx context.Context, name string, price, stock int64) (model.Product, error) {
	builder := sq.Insert("products").
		PlaceholderFormat(sq.Dollar).
		Columns("name", "price", "stock").
		Values(name, price, stock).
		Suffix("RETURNING id, name, price, stock")

	query, args, err := builder.ToSql()
	if err != nil {
		return model.Product{}, fmt.Errorf("failed to build insert query: %w", err)
	}

	conn := r.getConn(ctx)

	var product model.Product
	err = conn.QueryRow(ctx, query, args...).Scan(&product.ID, &product.Name, &product.Price, &product.Stock)
	if err != nil {
		return model.Product{}, fmt.Errorf("failed to create product: %w", err)
	}

	return product, nil
}

func (r *Repository) UpdateProduct(ctx context.Context, id int64, name *string, price *int64) (model.Product, error) {
	builder := sq.Update("products").
		PlaceholderFormat(sq.Dollar)

	if name != nil {
		builder = builder.Set("name", *name)
	}
	if price != nil {
		builder = builder.Set("price", *price)
	}
	builder = builder.
		Where(sq.Eq{"id": id, "deleted_at": nil}).
		Suffix("RETURNING id, name, price, stock")

	query, args, err := builder.ToSql()
	if err != nil {
		return model.Product{}, fmt.Errorf("failed to build update query: %w", err)
	}

	conn := r.getConn(ctx)

	var product model.Product
	err = conn.QueryRow(ctx, query, args...).Scan(&product.ID, &product.Name, &product.Price, &product.Stock)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Product{}, repository.ErrProductNotFound
		}
		return model.Product{}, fmt.Errorf("failed to update product: %w", err)
	}

	return product, nil
}

func (r *Repository) SetProductStock(ctx context.Context, id, stock int64) (model.Product, error) {
	builder := sq.Update("products").
		PlaceholderFormat(sq.Dollar).
		Set("stock", stock).
		Where(sq.Eq{"id": id, "deleted_at": nil}).
		Suffix("RETURNING id, name, price, stock")

	query, args, err := builder.ToSql()
	if err != nil {
		return model.Product{}, fmt.Errorf("failed to build update query: %w", err)
	}

	conn := r.getConn(ctx)

	var product model.Product
	err = conn.QueryRow(ctx, query, args...).Scan(&product.ID, &product.Name, &product.Price, &product.Stock)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Product{}, repository.ErrProductNotFound
		}
		return model.Product{}, fmt.Errorf("failed to set product stock: %w", err)
	}

	return product, nil
}

func (r *Repository) AddProductStock(ctx context.Context, id, quantity int64) (model.Product, error) {
	builder := sq.Update("products").
		PlaceholderFormat(sq.Dollar).
		Set("stock", sq.Expr("stock + ?", quantity)).
		Where(sq.Eq{"id": id, "deleted_at": nil}).
		Suffix("RETURNING id, name, price, stock")

	query, args, err := builder.ToSql()
	if err != nil {
		return model.Product{}, fmt.Errorf("failed to build update query: %w", err)
	}

	conn := r.getConn(ctx)

	var product model.Product
	err = conn.QueryRow(ctx, query, args...).Scan(&product.ID, &product.Name, &product.Price, &product.Stock)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Product{}, repository.ErrProductNotFound
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "22003" {
			return model.Product{}, repository.ErrStockOverflow
		}
		return model.Product{}, fmt.Errorf("failed to add product stock: %w", err)
	}

	return product, nil
}

func (r *Repository) DeleteProduct(ctx context.Context, id int64) error {
	builder := sq.Update("products").
		PlaceholderFormat(sq.Dollar).
		Set("deleted_at", sq.Expr("NOW()")).
		Where(sq.Eq{"id": id, "deleted_at": nil})

	query, args, err := builder.ToSql()
	if err != nil {
		return fmt.Errorf("failed to build update query: %w", err)
	}

	conn := r.getConn(ctx)

	result, err := conn.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}
	if result.RowsAffected() == 0 {
		return repository.ErrProductNotFound
	}

	return nil
}
