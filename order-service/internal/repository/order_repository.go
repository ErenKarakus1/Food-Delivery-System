package repository

import (
	"context"
	"errors"

	"github.com/ErenKarakus1/Food-Delivery-System/order-service/internal/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

const getCustomerOrdersQuery = `
	SELECT
		id,
		customer_id,
		restaurant_id,
		total_cents,
		status,
		created_at
	FROM orders
	WHERE customer_id=$1
	ORDER BY created_at DESC
`

const createOrderQuery = `
	INSERT INTO orders (
		id,
		customer_id,
		restaurant_id,
		total_cents,
		status
	)
	VALUES ($1,$2,$3,$4,$5)
	RETURNING
		id,
		customer_id,
		restaurant_id,
		total_cents,
		status,
		created_at
`

const createOrderItemQuery = `
	INSERT INTO order_items (
		id,
		order_id,
		menu_item_id,
		quantity,
		name,
		price_cents
	)
	VALUES ($1,$2,$3,$4,$5,$6)
`

func GetCustomerOrders(ctx context.Context, pool *pgxpool.Pool, customerID uuid.UUID) ([]model.Order, error) {
	rows, err := pool.Query(
		ctx,
		getCustomerOrdersQuery,
		customerID,
	)
	if err != nil {
		return []model.Order{}, errors.New("internal server error")
	}
	defer rows.Close()

	var orders []model.Order
	for rows.Next() {
		var order model.Order
		err := rows.Scan(
			&order.ID,
			&order.CustomerID,
			&order.RestaurantID,
			&order.TotalCents,
			&order.Status,
			&order.CreatedAt,
		)
		if err != nil {
			return []model.Order{}, errors.New("internal server error")
		}
		orders = append(orders, order)
	}
	if err := rows.Err(); err != nil {
		return []model.Order{}, errors.New("internal server error")
	}
	if orders == nil {
		orders = []model.Order{}
	}
	return orders, nil
}

func CreateOrder(ctx context.Context, pool *pgxpool.Pool, order model.Order, items []model.OrderItem) (model.Order, error) {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return model.Order{}, errors.New("internal server error")
	}
	defer tx.Rollback(ctx)

	var createOrderResponse model.Order
	err = tx.QueryRow(
		ctx,
		createOrderQuery,
		order.ID,
		order.CustomerID,
		order.RestaurantID,
		order.TotalCents,
		order.Status,
	).Scan(
		&createOrderResponse.ID,
		&createOrderResponse.CustomerID,
		&createOrderResponse.RestaurantID,
		&createOrderResponse.TotalCents,
		&createOrderResponse.Status,
		&createOrderResponse.CreatedAt,
	)
	if err != nil {
		return model.Order{}, errors.New("internal server error")
	}
	for _, item := range items {
		_, err := tx.Exec(
			ctx,
			createOrderItemQuery,
			item.ID,
			item.OrderID,
			item.MenuItemID,
			item.Quantity,
			item.Name,
			item.PriceCents,
		)
		if err != nil {
			return model.Order{}, errors.New("internal server error")
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return model.Order{}, errors.New("internal server error")
	}
	return order, nil

}
