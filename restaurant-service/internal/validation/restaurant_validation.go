package validation

import (
	"errors"

	"github.com/ErenKarakus1/Food-Delivery-System/restaurant-service/internal/model"
)

func ValidateCreateRestaurantRequest(r model.CreateRestaurantRequest) error {
	if len(r.Name) < 5 {
		return errors.New("restaurant name must be at least 5 characters")
	}
	if len(r.Name) > 50 {
		return errors.New("restaurant name must be at most 50 characters")
	}
	return nil
}
