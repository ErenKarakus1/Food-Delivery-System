package repository

import (
	"context"
	"errors"

	"github.com/ErenKarakus1/Food-Delivery-System/restaurant-service/internal/model"
	"github.com/jackc/pgx/v5/pgxpool"
)

const createMenuItemQuery = `
	INSERT INTO menu_items (
		id,
		restaurant_id,
		name,
		description,
		price_cents
	)
	VALUES ($1,$2,$3,$4,$5)
	RETURNING
		id,
		restaurant_id,
		name,
		description,
		price_cents,
		created_at
`

func CreateMenuItem(ctx context.Context, pool *pgxpool.Pool, req model.MenuItem) (model.MenuItem, error) {
	var createdMenuItem model.MenuItem
	err := pool.QueryRow(
		ctx,
		createMenuItemQuery,
		req.ID,
		req.RestaurantID,
		req.Name,
		req.Description,
		req.PriceCents,
	).Scan(
		&createdMenuItem.ID,
		&createdMenuItem.RestaurantID,
		&createdMenuItem.Name,
		&createdMenuItem.Description,
		&createdMenuItem.PriceCents,
		&createdMenuItem.CreatedAt,
	)
	if err != nil {
		return model.MenuItem{}, errors.New("couldnt create menu item")
	}
	return createdMenuItem, nil
}
