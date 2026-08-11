package repository

import (
	"context"
	"errors"

	"github.com/ErenKarakus1/Food-Delivery-System/order-service/internal/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrOrderNotFound = errors.New("order not found")

const getCustomerOrdersQuery = `
	SELECT
		id,
		customer_id,
		restaurant_id,
		total_cents,
		status,
		created_at,
		updated_at
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
		created_at,
		updated_at
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

const getCustomerOrderByIDQuery = `
	SELECT
		id,
		customer_id,
		restaurant_id,
		total_cents,
		status,
		created_at,
		updated_at
	FROM orders
	WHERE id=$1
	AND customer_id=$2
`

const getOrderItemsByIDQuery = `
	SELECT
		id,
		order_id,
		menu_item_id,
		quantity,
		name,
		price_cents
	FROM order_items
	WHERE order_id=$1
`

const getRestaurantOrderByIDQuery = `
	SELECT
		id,
		customer_id,
		restaurant_id,
		total_cents,
		status,
		created_at,
		updated_at
	FROM orders
	WHERE id=$1
`

const getOrdersByRestaurantIDQuery = `
	SELECT
		id,
		customer_id,
		restaurant_id,
		total_cents,
		status,
		created_at,
		updated_at
	FROM orders
	WHERE restaurant_id=$1
`

const updateOrderStatusByOrderIDQuery = `
	UPDATE orders
	SET 
		status = $1,
		updated_at = NOW()
	WHERE id = $2
	RETURNING
		id,
		customer_id,
		restaurant_id,
		total_cents,
		status,
		created_at,
		updated_at
`

const getOrdersReadyForPickupQuery = `
	SELECT
		id,
		customer_id,
		restaurant_id,
		total_cents,
		status,
		created_at,
		updated_at
	FROM orders
	WHERE status='ready_for_pickup'
`

func GetOrderItemsByOrderID(ctx context.Context, pool *pgxpool.Pool, orderID uuid.UUID) ([]model.OrderItem, error) {
	var order_items []model.OrderItem
	rows, err := pool.Query(
		ctx,
		getOrderItemsByIDQuery,
		orderID,
	)
	if err != nil {
		return []model.OrderItem{}, errors.New("internal server error")
	}
	defer rows.Close()

	for rows.Next() {
		var item model.OrderItem
		err := rows.Scan(
			&item.ID,
			&item.OrderID,
			&item.MenuItemID,
			&item.Quantity,
			&item.Name,
			&item.PriceCents,
		)
		if err != nil {
			return []model.OrderItem{}, errors.New("internal server error")
		}
		order_items = append(order_items, item)
	}
	if err := rows.Err(); err != nil {
		return []model.OrderItem{}, errors.New("internal server error")
	}
	if order_items == nil {
		order_items = []model.OrderItem{}
	}
	return order_items, nil
}

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
			&order.UpdatedAt,
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
		&createOrderResponse.UpdatedAt,
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

func GetCustomerOrderByID(ctx context.Context, pool *pgxpool.Pool, orderID uuid.UUID, customerID uuid.UUID) (model.Order, []model.OrderItem, error) {
	var order model.Order
	err := pool.QueryRow(
		ctx,
		getCustomerOrderByIDQuery,
		orderID,
		customerID,
	).Scan(
		&order.ID,
		&order.CustomerID,
		&order.RestaurantID,
		&order.TotalCents,
		&order.Status,
		&order.CreatedAt,
		&order.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Order{}, []model.OrderItem{}, ErrOrderNotFound
		}
		return model.Order{}, []model.OrderItem{}, errors.New("internal server error")
	}
	var order_items []model.OrderItem
	rows, err := pool.Query(
		ctx,
		getOrderItemsByIDQuery,
		orderID,
	)
	if err != nil {
		return model.Order{}, []model.OrderItem{}, errors.New("internal server error")
	}
	defer rows.Close()

	for rows.Next() {
		var item model.OrderItem
		err := rows.Scan(
			&item.ID,
			&item.OrderID,
			&item.MenuItemID,
			&item.Quantity,
			&item.Name,
			&item.PriceCents,
		)
		if err != nil {
			return model.Order{}, []model.OrderItem{}, errors.New("internal server error")
		}
		order_items = append(order_items, item)
	}
	if err := rows.Err(); err != nil {
		return model.Order{}, []model.OrderItem{}, errors.New("internal server error")
	}
	if order_items == nil {
		order_items = []model.OrderItem{}
	}
	return order, order_items, nil
}

func GetRestaurantOrderByID(ctx context.Context, pool *pgxpool.Pool, orderID uuid.UUID) (model.Order, error) {
	var order model.Order
	err := pool.QueryRow(
		ctx,
		getRestaurantOrderByIDQuery,
		orderID,
	).Scan(
		&order.ID,
		&order.CustomerID,
		&order.RestaurantID,
		&order.TotalCents,
		&order.Status,
		&order.CreatedAt,
		&order.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Order{}, ErrOrderNotFound
		}
		return model.Order{}, errors.New("internal server error")
	}

	return order, nil
}

func GetOrdersByRestaurantID(ctx context.Context, pool *pgxpool.Pool, restaurantID uuid.UUID) ([]model.Order, error) {
	rows, err := pool.Query(
		ctx,
		getOrdersByRestaurantIDQuery,
		restaurantID,
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
			&order.UpdatedAt,
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

func UpdateOrderStatusByOrderID(ctx context.Context, pool *pgxpool.Pool, newStatus string, orderID uuid.UUID) (model.Order, error) {
	var updatedOrder model.Order
	err := pool.QueryRow(
		ctx,
		updateOrderStatusByOrderIDQuery,
		newStatus,
		orderID,
	).Scan(
		&updatedOrder.ID,
		&updatedOrder.CustomerID,
		&updatedOrder.RestaurantID,
		&updatedOrder.TotalCents,
		&updatedOrder.Status,
		&updatedOrder.CreatedAt,
		&updatedOrder.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Order{}, ErrOrderNotFound
		}
		return model.Order{}, errors.New("internal server error")
	}
	return updatedOrder, nil
}

func GetOrdersReadyForPickup(ctx context.Context, pool *pgxpool.Pool) ([]model.Order, error) {
	var orders []model.Order
	rows, err := pool.Query(
		ctx,
		getOrdersReadyForPickupQuery,
	)
	if err != nil {
		return []model.Order{}, errors.New("internal server error")
	}
	defer rows.Close()
	for rows.Next() {
		var order model.Order
		err := rows.Scan(
			&order.ID,
			&order.CustomerID,
			&order.RestaurantID,
			&order.TotalCents,
			&order.Status,
			&order.CreatedAt,
			&order.UpdatedAt,
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
