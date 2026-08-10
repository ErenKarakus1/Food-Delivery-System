package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/ErenKarakus1/Food-Delivery-System/order-service/internal/model"
	"github.com/ErenKarakus1/Food-Delivery-System/order-service/internal/repository"
	"github.com/ErenKarakus1/Food-Delivery-System/order-service/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func GetCustomerOrdersHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		customerID := ctx.GetHeader("X-User-ID")
		if customerID == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "user id required"})
			return
		}
		parsedCustomerID, err := uuid.Parse(customerID)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
			return
		}
		role := ctx.GetHeader("X-User-Role")
		if role == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "role required"})
			return
		}
		if role != "customer" {
			ctx.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		orders, err := repository.GetCustomerOrders(ctx.Request.Context(), pool, parsedCustomerID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
		ctx.JSON(http.StatusOK, orders)
	}
}

func returnMenuItem(key uuid.UUID) (model.MenuItem, error) {
	url := fmt.Sprintf("http://localhost:8081/menu-items/%s", key)
	resp, err := http.Get(url)
	if err != nil {
		return model.MenuItem{}, errors.New("internal server error")
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return model.MenuItem{}, errors.New("menu item not found")
	}

	if resp.StatusCode != http.StatusOK {
		return model.MenuItem{}, errors.New("restaurant service error")
	}
	var menuItem model.MenuItem
	if err := json.NewDecoder(resp.Body).Decode(&menuItem); err != nil {
		return model.MenuItem{}, errors.New("internal server error")
	}
	return menuItem, nil
}

func createOrderItems(items []model.OrderItemRequest, orderID uuid.UUID) ([]model.OrderItem, error) {
	orderItemsMap := make(map[uuid.UUID]int)
	for _, item := range items {
		if item.Quantity <= 0 {
			return []model.OrderItem{}, errors.New("invalid menu item quantity")
		}
		orderItemsMap[item.MenuItemID] += item.Quantity
	}
	var orderItems []model.OrderItem
	var restaurantID uuid.UUID
	for orderItemID, quantity := range orderItemsMap {
		menuItem, err := returnMenuItem(orderItemID)
		if err != nil {
			return []model.OrderItem{}, err
		}
		if restaurantID == uuid.Nil {
			restaurantID = menuItem.RestaurantID
		} else if restaurantID != menuItem.RestaurantID {
			return []model.OrderItem{}, errors.New("all menu items must belong to the same restaurant")
		}
		item := service.CreateOrderItemFromMenuItem(menuItem, orderID, quantity)
		orderItems = append(orderItems, item)
	}
	if orderItems == nil {
		return []model.OrderItem{}, errors.New("no order items exist")
	}
	return orderItems, nil
}

func CreateOrderHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		customerID := ctx.GetHeader("X-User-ID")
		if customerID == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "user id required"})
			return
		}
		parsedCustomerID, err := uuid.Parse(customerID)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
			return
		}
		role := ctx.GetHeader("X-User-Role")
		if role == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "role required"})
			return
		}
		if role != "customer" {
			ctx.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		var req model.OrderRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}

		order := service.CreateOrderFromRequest(req, parsedCustomerID, -1)

		orderItems, err := createOrderItems(req.Items, order.ID)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var totalCents int64
		for _, item := range orderItems {
			totalCents += (item.PriceCents * int64(item.Quantity))
		}

		order.TotalCents = totalCents

		createdOrder, err := repository.CreateOrder(ctx.Request.Context(), pool, order, orderItems)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
		ctx.JSON(http.StatusCreated, createdOrder)
	}
}
