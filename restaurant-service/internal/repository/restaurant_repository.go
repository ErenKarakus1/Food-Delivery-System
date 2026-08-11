package repository

import (
	"context"
	"errors"
	"log"

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

const restaurantExistsAndOwnedQuery = `
	SELECT
		id,
		owner_id,
		name,
		created_at
	FROM restaurants
	WHERE id=$1
	AND owner_id=$2
`

const getRestaurantByIDQuery = `
	SELECT
		id,
		owner_id,
		name,
		created_at
	FROM restaurants
	where id=$1
`

const getAllRestaurantQuery = `
	SELECT
		id,
		owner_id,
		name,
		created_at
	FROM restaurants
	ORDER BY created_at DESC
`

const getMenuQuery = `
	SELECT
		id,
		restaurant_id,
		name,
		description,
		price_cents,
		created_at
	FROM menu_items
	WHERE restaurant_id=$1
	ORDER BY created_at DESC
`

const checkRestaurantExistsQuery = `
	SELECT 1
	FROM restaurants
	WHERE id=$1
	LIMIT 1
`

const getRestaurantsByOwnerIDQuery = `
	SELECT
		id,
		owner_id,
		name,
		created_at
	FROM restaurants
	WHERE owner_id=$1
`

func checkRestaurantExists(ctx context.Context, pool *pgxpool.Pool, restaurantID uuid.UUID) error {
	var exists int
	err := pool.QueryRow(
		ctx,
		checkRestaurantExistsQuery,
		restaurantID,
	).Scan(&exists)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrRestaurantNotFound
		}
		return errors.New("internal server error")
	}
	return nil
}

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

func RestaurantExistsAndOwned(ctx context.Context, pool *pgxpool.Pool, restaurantID uuid.UUID, ownerID uuid.UUID) (model.Restaurant, error) {
	var restaurant model.Restaurant
	err := pool.QueryRow(
		ctx,
		restaurantExistsAndOwnedQuery,
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

func GetRestaurantByID(ctx context.Context, pool *pgxpool.Pool, restaurantID uuid.UUID) (model.Restaurant, error) {
	var restaurant model.Restaurant
	err := pool.QueryRow(
		ctx,
		getRestaurantByIDQuery,
		restaurantID,
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

func GetAllRestaurants(ctx context.Context, pool *pgxpool.Pool) ([]model.Restaurant, error) {
	rows, err := pool.Query(
		ctx,
		getAllRestaurantQuery,
	)
	if err != nil {
		return nil, errors.New("internal server error")
	}
	defer rows.Close()

	var restaurants []model.Restaurant
	for rows.Next() {
		var restaurant model.Restaurant
		err := rows.Scan(
			&restaurant.ID,
			&restaurant.OwnerID,
			&restaurant.Name,
			&restaurant.CreatedAt,
		)
		if err != nil {
			return nil, errors.New("internal server error")
		}
		restaurants = append(restaurants, restaurant)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.New("internal server error")
	}
	return restaurants, nil
}

func GetMenu(ctx context.Context, pool *pgxpool.Pool, restaurantID uuid.UUID) ([]model.MenuItem, error) {
	if err := checkRestaurantExists(ctx, pool, restaurantID); err != nil {
		if errors.Is(err, ErrRestaurantNotFound) {
			return []model.MenuItem{}, ErrRestaurantNotFound
		}
		return []model.MenuItem{}, errors.New("internal server error")
	}
	rows, err := pool.Query(
		ctx,
		getMenuQuery,
		restaurantID,
	)
	if err != nil {
		return []model.MenuItem{}, errors.New("internal server error")
	}
	defer rows.Close()

	var menu []model.MenuItem
	for rows.Next() {
		var menuItem model.MenuItem
		err := rows.Scan(
			&menuItem.ID,
			&menuItem.RestaurantID,
			&menuItem.Name,
			&menuItem.Description,
			&menuItem.PriceCents,
			&menuItem.CreatedAt,
		)
		if err != nil {
			return []model.MenuItem{}, errors.New("internal server error")
		}
		menu = append(menu, menuItem)
	}
	if err := rows.Err(); err != nil {
		return []model.MenuItem{}, errors.New("internal server error")
	}
	if menu == nil {
		menu = []model.MenuItem{}
	}
	return menu, nil
}

func GetRestaurantsByOwnerID(ctx context.Context, pool *pgxpool.Pool, ownerID uuid.UUID) ([]model.Restaurant, error) {
	rows, err := pool.Query(
		ctx,
		getRestaurantsByOwnerIDQuery,
		ownerID,
	)
	if err != nil {
		log.Println(err)
		return []model.Restaurant{}, errors.New("internal server error")
	}
	defer rows.Close()

	var restaurants []model.Restaurant
	for rows.Next() {
		var restaurant model.Restaurant
		err := rows.Scan(
			&restaurant.ID,
			&restaurant.OwnerID,
			&restaurant.Name,
			&restaurant.CreatedAt,
		)
		if err != nil {
			return []model.Restaurant{}, errors.New("internal server error")
		}
		restaurants = append(restaurants, restaurant)
	}
	if err := rows.Err(); err != nil {
		return []model.Restaurant{}, errors.New("internal server error")
	}
	if restaurants == nil {
		return []model.Restaurant{}, ErrRestaurantNotFound
	}
	return restaurants, nil
}
