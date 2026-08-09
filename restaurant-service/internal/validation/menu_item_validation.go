package validation

import (
	"errors"

	"github.com/ErenKarakus1/Food-Delivery-System/restaurant-service/internal/model"
)

func ValidateCreateMenuItemRequest(r model.CreateMenuItemRequest) error {
	if len(r.Name) < 2 {
		return errors.New("menu item name must be at least 2 characters")
	}
	if len(r.Name) > 50 {
		return errors.New("menu item name must be at most 50 characters")
	}
	if len(r.Description) > 500 {
		return errors.New("menu item description must be at most 500 characters")
	}
	if r.PriceCents <= 0 {
		return errors.New("invalid price")
	}
	return nil
}
