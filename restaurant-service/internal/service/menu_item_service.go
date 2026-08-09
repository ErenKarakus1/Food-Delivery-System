package service

import (
	"strings"

	"github.com/ErenKarakus1/Food-Delivery-System/restaurant-service/internal/model"
	"github.com/google/uuid"
)

func CreateMenuItemFromRequest(req model.CreateMenuItemRequest) model.MenuItem {
	return model.MenuItem{
		ID:           uuid.New(),
		RestaurantID: req.RestaurantID,
		Name:         strings.TrimSpace(req.Name),
		Description:  strings.TrimSpace(req.Description),
		PriceCents:   req.PriceCents,
	}
}
