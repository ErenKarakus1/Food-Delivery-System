package service

import (
	"github.com/ErenKarakus1/Food-Delivery-System/order-service/internal/model"
	"github.com/google/uuid"
)

func CreateOrderItemFromMenuItem(r model.MenuItem, orderID uuid.UUID, quantity int) model.OrderItem {
	return model.OrderItem{
		ID:         uuid.New(),
		OrderID:    orderID,
		MenuItemID: r.ID,
		Quantity:   quantity,
		Name:       r.Name,
		PriceCents: r.PriceCents,
	}
}

func CreateOrderFromRequest(r model.OrderRequest, customerID uuid.UUID, totalCents int64) model.Order {
	return model.Order{
		ID:           uuid.New(),
		CustomerID:   customerID,
		RestaurantID: r.RestaurantID,
		TotalCents:   totalCents,
		Status:       "pending",
	}
}
