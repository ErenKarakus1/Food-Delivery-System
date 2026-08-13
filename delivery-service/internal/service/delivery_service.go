package service

import (
	"github.com/ErenKarakus1/Food-Delivery-System/delivery-service/internal/model"
	"github.com/google/uuid"
)

func CreateDeliveryRejection(deliveryID uuid.UUID, courierID uuid.UUID) model.DeliveryRejecetion {
	return model.DeliveryRejecetion{
		ID:         uuid.New(),
		DeliveryID: deliveryID,
		CourierID:  courierID,
	}
}
