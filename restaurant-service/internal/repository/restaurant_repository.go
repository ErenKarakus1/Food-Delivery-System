package repository

import (
	"context"
	"errors"

	"github.com/ErenKarakus1/Food-Delivery-System/restaurant-service/internal/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrRestaurantNotFound = errors.New("restaurant not found")

const createRestaurantQuery = `
	INSERT INTO restaurants (
		id,
		owner_id,
		name
	)
	VALUES ($1,$2,$3)
	RETURNING
		id,
		owner_id,
		name,
		created_at
`

const findRestaurantQuery = `
	SELECT
		id,
		owner_id,
		name,
		created_at
	FROM restaurants
	WHERE id=$1
	AND owner_id=$2
`

func CreateRestaurant(ctx context.Context, pool *pgxpool.Pool, req model.Restaurant) (model.Restaurant, error) {
	var createdRestaurant model.Restaurant
	err := pool.QueryRow(
		ctx,
		createRestaurantQuery,
		req.ID,
		req.OwnerID,
		req.Name,
	).Scan(
		&createdRestaurant.ID,
		&createdRestaurant.OwnerID,
		&createdRestaurant.Name,
		&createdRestaurant.CreatedAt,
	)
	if err != nil {
		return model.Restaurant{}, errors.New("couldnt create restaurant")
	}
	return createdRestaurant, nil
}

func FindRestaurant(ctx context.Context, pool *pgxpool.Pool, restaurantID uuid.UUID, ownerID uuid.UUID) (model.Restaurant, error) {
	var restaurant model.Restaurant
	err := pool.QueryRow(
		ctx,
		findRestaurantQuery,
		restaurantID,
		ownerID,
	).Scan(
		&restaurant.ID,
		&restaurant.OwnerID,
		&restaurant.Name,
		&restaurant.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Restaurant{}, ErrRestaurantNotFound
		}
		return model.Restaurant{}, errors.New("internal server error")
	}
	return restaurant, nil
}
