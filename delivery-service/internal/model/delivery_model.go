package model

import (
	"time"

	"github.com/google/uuid"
)

type StatusUpdateRequest struct {
	Status string `json:"status"`
}

type Delivery struct {
	ID        uuid.UUID `json:"id"`
	OrderID   uuid.UUID `json:"order_id"`
	CourierID uuid.UUID `json:"courier_id"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Courier struct {
	ID          uuid.UUID `json:"id"`
	IsAvailable bool      `json:"is_available"`
	CreatedAt   time.Time `json:"created_at"`
}

type DeliveryRejecetion struct {
	ID         uuid.UUID `json:"id"`
	DeliveryID uuid.UUID `json:"delivery_id"`
	CourierID  uuid.UUID `json:"courier_id"`
	CreatedAt  time.Time `json:"created_at"`
}

type Order struct {
	ID           uuid.UUID `json:"id"`
	CustomerID   uuid.UUID `json:"customer_id"`
	RestaurantID uuid.UUID `json:"restaurant_id"`
	TotalCents   int64     `json:"total_cents"`
	Status       string    `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type CreateCourierRequest struct {
	ID string `json:"id"`
}
