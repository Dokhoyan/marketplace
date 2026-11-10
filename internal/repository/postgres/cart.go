package postgres

import (
	"context"
	"fmt"

	sq "github.com/Masterminds/squirrel"
)

func (r *Repository) CreateCart(ctx context.Context, userID int64) error {
	builder := sq.Insert("carts").
		PlaceholderFormat(sq.Dollar).
		Columns("user_id").
		Values(userID)

	query, args, err := builder.ToSql()
	if err != nil {
		return fmt.Errorf("failed to build insert query: %w", err)
	}

	conn := r.getConn(ctx)

	_, err = conn.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to create cart: %w", err)
	}

	return nil
}
