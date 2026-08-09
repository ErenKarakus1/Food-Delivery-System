package model

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

type MenuItem struct {
	ID           uuid.UUID `json:"id"`
	RestaurantID uuid.UUID `json:"restaurant_id"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	PriceCents   int64     `json:"price_cents"`
	CreatedAt    time.Time `json:"created_at"`
}

type CreateMenuItemRequest struct {
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	RestaurantID uuid.UUID `json:"restaurant_id"`
	PriceCents   int64     `json:"price_cents"`
}

func (r *CreateMenuItemRequest) Normalize() {
	r.Name = strings.TrimSpace(r.Name)
	r.Description = strings.TrimSpace(r.Description)
}
