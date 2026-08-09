package model

import (
	"strings"
	"time"

	"github.com/google/uuid"
)

type Restaurant struct {
	ID        uuid.UUID `json:"id"`
	OwnerID   uuid.UUID `json:"ownder_id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type CreateRestaurantRequest struct {
	Name string `json:"name"`
}

func (r *CreateRestaurantRequest) Normalize() {
	r.Name = strings.TrimSpace(r.Name)
}
