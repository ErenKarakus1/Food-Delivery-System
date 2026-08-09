package repository

import (
	"context"
	"errors"

	"github.com/ErenKarakus1/Food-Delivery-System/restaurant-service/internal/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
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

const getMenuItemByIDQuery = `
	SELECT
		id,
		restaurant_id,
		name,
		description,
		price_cents,
		created_at
	FROM menu_items
	WHERE id=$1
`

const deleteMenuItemQuery = `
	DELETE FROM menu_items
	WHERE id=$1
`

var ErrMenuItemNotFound = errors.New("menu item not found")

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

func GetMenuItemByID(ctx context.Context, pool *pgxpool.Pool, menuItemID uuid.UUID) (model.MenuItem, error) {
	var menuItem model.MenuItem
	err := pool.QueryRow(
		ctx,
		getMenuItemByIDQuery,
		menuItemID,
	).Scan(
		&menuItem.ID,
		&menuItem.RestaurantID,
		&menuItem.Name,
		&menuItem.Description,
		&menuItem.PriceCents,
		&menuItem.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.MenuItem{}, ErrMenuItemNotFound
		}
		return model.MenuItem{}, errors.New("internal server error")
	}
	return menuItem, nil
}

func DeleteMenuItem(ctx context.Context, pool *pgxpool.Pool, menuItemID uuid.UUID) error {
	tag, err := pool.Exec(
		ctx,
		deleteMenuItemQuery,
		menuItemID,
	)
	if err != nil {
		return errors.New("internal server error")
	}
	if tag.RowsAffected() == 0 {
		return ErrMenuItemNotFound
	}
	return nil
}
