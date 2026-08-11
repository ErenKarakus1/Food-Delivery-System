package model

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

type OrderItem struct {
	ID         uuid.UUID `json:"id"`
	OrderID    uuid.UUID `json:"order_id"`
	MenuItemID uuid.UUID `json:"menu_item_id"`
	Quantity   int       `json:"quantity"`
	Name       string    `json:"name"`
	PriceCents int64     `json:"price_cents"`
}

type OrderItemRequest struct {
	MenuItemID uuid.UUID `json:"menu_item_id"`
	Quantity   int       `json:"quantity"`
}

type Order struct {
	ID           uuid.UUID `json:"id"`
	CustomerID   uuid.UUID `json:"customer_id"`
	RestaurantID uuid.UUID `json:"restaurant_id"`
	TotalCents   int64     `json:"total_cents"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
}

type OrderRequest struct {
	RestaurantID uuid.UUID          `json:"restaurant_id"`
	Items        []OrderItemRequest `json:"items"`
}

type MenuItem struct {
	ID           uuid.UUID `json:"id"`
	RestaurantID uuid.UUID `json:"restaurant_id"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	PriceCents   int64     `json:"price_cents"`
	CreatedAt    time.Time `json:"created_at"`
}

type Restaurant struct {
	ID        uuid.UUID `json:"id"`
	OwnerID   uuid.UUID `json:"ownder_id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type UpdateOrderStatusRequest struct {
	Status string `json:"status"`
}

func (r *UpdateOrderStatusRequest) Normalize() {
	r.Status = strings.ToLower(strings.TrimSpace(r.Status))
}
