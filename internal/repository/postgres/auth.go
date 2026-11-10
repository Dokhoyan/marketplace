package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Dokhoyan/marketplace/internal/model"
	"github.com/Dokhoyan/marketplace/internal/repository"
	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
)

func (r *Repository) CreateSession(ctx context.Context, userID int64, tokenHash string) (time.Time, error) {
	user := sq.Select("id").
		Column("?", tokenHash).
		From("users").
		Where(sq.Eq{"id": userID, "deleted_at": nil})

	builder := sq.Insert("sessions").
		PlaceholderFormat(sq.Dollar).
		Columns("user_id", "token_hash").
		Select(user).
		Suffix("RETURNING expires_at")

	query, args, err := builder.ToSql()
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to build insert query: %w", err)
	}

	conn := r.getConn(ctx)

	var expiresAt time.Time
	err = conn.QueryRow(ctx, query, args...).Scan(&expiresAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return time.Time{}, repository.ErrUserNotFound
		}
		return time.Time{}, fmt.Errorf("failed to create session: %w", err)
	}

	return expiresAt.UTC(), nil
}

func (r *Repository) GetSession(ctx context.Context, tokenHash string) (model.Session, error) {
	builder := sq.Select("s.user_id", "u.role", "s.expires_at").
		PlaceholderFormat(sq.Dollar).
		From("sessions s").
		Join("users u ON u.id = s.user_id").
		Where(sq.Eq{"s.token_hash": tokenHash, "u.deleted_at": nil}).
		Where("s.expires_at > NOW()")

	query, args, err := builder.ToSql()
	if err != nil {
		return model.Session{}, fmt.Errorf("failed to build select query: %w", err)
	}

	conn := r.getConn(ctx)

	session := model.Session{TokenHash: tokenHash}
	err = conn.QueryRow(ctx, query, args...).Scan(&session.UserID, &session.Role, &session.ExpiresAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Session{}, repository.ErrSessionNotFound
		}
		return model.Session{}, fmt.Errorf("failed to get session: %w", err)
	}

	session.ExpiresAt = session.ExpiresAt.UTC()
	return session, nil
}

func (r *Repository) DeleteSession(ctx context.Context, tokenHash string) error {
	builder := sq.Delete("sessions").
		PlaceholderFormat(sq.Dollar).
		Where(sq.Eq{"token_hash": tokenHash})

	query, args, err := builder.ToSql()
	if err != nil {
		return fmt.Errorf("failed to build delete query: %w", err)
	}

	conn := r.getConn(ctx)

	_, err = conn.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	return nil
}

func (r *Repository) DeleteUserSessions(ctx context.Context, userID int64) error {
	builder := sq.Delete("sessions").
		PlaceholderFormat(sq.Dollar).
		Where(sq.Eq{"user_id": userID})

	query, args, err := builder.ToSql()
	if err != nil {
		return fmt.Errorf("failed to build delete query: %w", err)
	}

	conn := r.getConn(ctx)

	_, err = conn.Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to delete user sessions: %w", err)
	}

	return nil
}
