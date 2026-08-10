package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPool(databaseURL string) (*pgxpool.Pool, error) {
	return pgxpool.New(
		context.Background(),
		databaseURL,
	)
}
