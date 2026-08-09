package repository

import (
	"context"
	"errors"

	"github.com/ErenKarakus1/Food-Delivery-System/auth-service/internal/model"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrEmailExists = errors.New("this email is already registered")

const createUserQuery = `
	INSERT INTO users (
		id,
		name,
		email,
		password_hash,
		role
	)
	VALUES ($1,$2,$3,$4,$5)
	RETURNING
		id,
		name,
		email,
		role,
		created_at
`

const findUserByEmailQuery = `
	SELECT
		id,
		name,
		email,
		password_hash,
		role,
		created_at
	FROM users
	WHERE email=$1
`

func RegisterUser(ctx context.Context, pool *pgxpool.Pool, req model.User) (model.RegisterResponse, error) {
	var createdUser model.RegisterResponse
	err := pool.QueryRow(
		ctx,
		createUserQuery,
		req.ID,
		req.Name,
		req.Email,
		req.PasswordHash,
		req.Role,
	).Scan(
		&createdUser.ID,
		&createdUser.Name,
		&createdUser.Email,
		&createdUser.Role,
		&createdUser.CreatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return model.RegisterResponse{}, ErrEmailExists
		}
		return model.RegisterResponse{}, errors.New("couldnt register user")
	}
	return createdUser, nil
}

func FindUserByEmail(ctx context.Context, pool *pgxpool.Pool, email string) (model.User, error) {
	var user model.User
	err := pool.QueryRow(
		ctx,
		findUserByEmailQuery,
		email,
	).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.PasswordHash,
		&user.Role,
		&user.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.User{}, errors.New("invalid email or password")
		}
		return model.User{}, errors.New("internal server error")
	}
	return user, nil
}
