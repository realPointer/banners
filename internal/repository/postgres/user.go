package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/realPointer/banners/internal/entity"
	"github.com/realPointer/banners/pkg/postgres"
	"github.com/rs/zerolog"
)

type UserRepo struct {
	*postgres.Postgres
	l *zerolog.Logger
}

func NewUserRepo(pg *postgres.Postgres, l *zerolog.Logger) *UserRepo {
	return &UserRepo{
		Postgres: pg,
		l:        l,
	}
}

func (r *UserRepo) CreateUser(ctx context.Context, user *entity.User) error {
	sql, args, _ := r.Builder.
		Insert("users").
		Columns("username", "password", "role").
		Values(user.Username, user.Password, user.Role).
		ToSql()

	_, err := r.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return fmt.Errorf("UserRepo.Create - r.Pool.Exec: %w", err)
	}

	return nil
}

func (r *UserRepo) GetUserByUsername(ctx context.Context, username string) (*entity.User, error) {
	sql, args, _ := r.Builder.
		Select("username", "password", "role").
		From("users").
		Where(squirrel.Eq{"username": username}).
		ToSql()

	var user entity.User
	err := r.Pool.QueryRow(ctx, sql, args...).Scan(
		&user.Username,
		&user.Password,
		&user.Role,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, errors.New("user not found")
		}
		return nil, fmt.Errorf("UserRepo.GetByUsername - r.Pool.QueryRow.Scan: %w", err)
	}

	return &user, nil
}
