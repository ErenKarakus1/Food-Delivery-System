package service

import (
	"github.com/ErenKarakus1/Food-Delivery-System/restaurant-service/internal/model"
	"github.com/google/uuid"
)

func CreateRestaurantFromRequest(r model.CreateRestaurantRequest, ownerID uuid.UUID) model.Restaurant {
	return model.Restaurant{
		ID:      uuid.New(),
		OwnerID: ownerID,
		Name:    r.Name,
	}
}
