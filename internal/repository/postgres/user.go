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

func (r *Repository) CreateUser(ctx context.Context, name, login, passwordHash string) (model.User, error) {
	builder := sq.Insert("users").
		PlaceholderFormat(sq.Dollar).
		Columns("name", "login", "password_hash").
		Values(name, login, passwordHash).
		Suffix("RETURNING id, name, login, role, balance, created_at")

	query, args, err := builder.ToSql()
	if err != nil {
		return model.User{}, fmt.Errorf("failed to build insert query: %w", err)
	}

	conn := r.getConn(ctx)

	var user model.User
	err = conn.QueryRow(ctx, query, args...).Scan(
		&user.ID,
		&user.Name,
		&user.Login,
		&user.Role,
		&user.Balance,
		&user.CreatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" && pgErr.ConstraintName == "users_login_lower_idx" {
			return model.User{}, repository.ErrLoginTaken
		}
		return model.User{}, fmt.Errorf("failed to create user: %w", err)
	}

	user.CreatedAt = user.CreatedAt.UTC()
	return user, nil
}

func (r *Repository) GetUserCredentials(ctx context.Context, login string) (model.UserCredentials, error) {
	builder := sq.Select("id", "name", "login", "role", "balance", "created_at", "COALESCE(password_hash, '')").
		PlaceholderFormat(sq.Dollar).
		From("users").
		Where(sq.Eq{"deleted_at": nil}).
		Where(sq.Expr("lower(login) = lower(?)", login))

	query, args, err := builder.ToSql()
	if err != nil {
		return model.UserCredentials{}, fmt.Errorf("failed to build select query: %w", err)
	}

	conn := r.getConn(ctx)

	var credentials model.UserCredentials
	user := &credentials.User
	err = conn.QueryRow(ctx, query, args...).Scan(
		&user.ID,
		&user.Name,
		&user.Login,
		&user.Role,
		&user.Balance,
		&user.CreatedAt,
		&credentials.PasswordHash,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.UserCredentials{}, repository.ErrUserNotFound
		}
		return model.UserCredentials{}, fmt.Errorf("failed to get user credentials: %w", err)
	}

	user.CreatedAt = user.CreatedAt.UTC()
	return credentials, nil
}

func (r *Repository) GetUser(ctx context.Context, id int64) (model.User, error) {
	builder := sq.Select("id", "name", "login", "role", "balance", "created_at").
		PlaceholderFormat(sq.Dollar).
		From("users").
		Where(sq.Eq{"id": id, "deleted_at": nil})

	query, args, err := builder.ToSql()
	if err != nil {
		return model.User{}, fmt.Errorf("failed to build select query: %w", err)
	}

	conn := r.getConn(ctx)

	var user model.User
	err = conn.QueryRow(ctx, query, args...).Scan(
		&user.ID,
		&user.Name,
		&user.Login,
		&user.Role,
		&user.Balance,
		&user.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.User{}, repository.ErrUserNotFound
		}
		return model.User{}, fmt.Errorf("failed to get user: %w", err)
	}

	user.CreatedAt = user.CreatedAt.UTC()
	return user, nil
}

func (r *Repository) UpdateUser(ctx context.Context, id int64, name string) (model.User, error) {
	builder := sq.Update("users").
		PlaceholderFormat(sq.Dollar).
		Set("name", name).
		Where(sq.Eq{"id": id, "deleted_at": nil}).
		Suffix("RETURNING id, name, login, role, balance, created_at")

	query, args, err := builder.ToSql()
	if err != nil {
		return model.User{}, fmt.Errorf("failed to build update query: %w", err)
	}

	conn := r.getConn(ctx)

	var user model.User
	err = conn.QueryRow(ctx, query, args...).Scan(
		&user.ID,
		&user.Name,
		&user.Login,
		&user.Role,
		&user.Balance,
		&user.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.User{}, repository.ErrUserNotFound
		}
		return model.User{}, fmt.Errorf("failed to update user: %w", err)
	}

	user.CreatedAt = user.CreatedAt.UTC()
	return user, nil
}

func (r *Repository) GetUsers(ctx context.Context) ([]model.User, error) {
	builder := sq.Select("id", "name", "login", "role", "balance", "created_at").
		PlaceholderFormat(sq.Dollar).
		From("users").
		Where(sq.Eq{"deleted_at": nil}).
		OrderBy("id")

	query, args, err := builder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build select query: %w", err)
	}

	conn := r.getConn(ctx)

	rows, err := conn.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %w", err)
	}
	defer rows.Close()

	users := make([]model.User, 0)
	for rows.Next() {
		var user model.User
		if err := rows.Scan(&user.ID, &user.Name, &user.Login, &user.Role, &user.Balance, &user.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		user.CreatedAt = user.CreatedAt.UTC()
		users = append(users, user)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate users: %w", err)
	}

	return users, nil
}

func (r *Repository) DeleteUser(ctx context.Context, id int64) error {
	builder := sq.Update("users").
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
		return fmt.Errorf("failed to delete user: %w", err)
	}
	if result.RowsAffected() == 0 {
		return repository.ErrUserNotFound
	}

	return nil
}
