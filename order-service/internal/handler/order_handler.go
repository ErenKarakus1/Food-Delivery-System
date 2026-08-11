package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/ErenKarakus1/Food-Delivery-System/order-service/internal/model"
	"github.com/ErenKarakus1/Food-Delivery-System/order-service/internal/repository"
	"github.com/ErenKarakus1/Food-Delivery-System/order-service/internal/service"
	"github.com/ErenKarakus1/Food-Delivery-System/order-service/internal/validation"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func getRestaurantByID(restaurantID uuid.UUID) (model.Restaurant, error) {
	url := fmt.Sprintf(
		"http://localhost:8081/restaurants/%s",
		restaurantID,
	)

	resp, err := http.Get(url)
	if err != nil {
		return model.Restaurant{}, errors.New("internal server error")
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return model.Restaurant{}, errors.New("restaurant not found")
	}

	if resp.StatusCode != http.StatusOK {
		return model.Restaurant{}, errors.New("restaurant service error")
	}

	var restaurant model.Restaurant

	if err := json.NewDecoder(resp.Body).Decode(&restaurant); err != nil {
		return model.Restaurant{}, errors.New("internal server error")
	}

	return restaurant, nil
}

func getMyRestaurants(userID uuid.UUID, role string) ([]model.Restaurant, error) {
	url := "http://localhost:8081/restaurants/me"
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return []model.Restaurant{}, errors.New("internal server error")
	}
	req.Header.Set("X-User-ID", userID.String())
	req.Header.Set("X-User-Role", role)

	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return []model.Restaurant{}, errors.New("restaurant service error")
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return []model.Restaurant{}, errors.New("restaurant service error")
	}
	var restaurants []model.Restaurant
	if err := json.NewDecoder(resp.Body).Decode(&restaurants); err != nil {
		return []model.Restaurant{}, errors.New("restaurant service error")
	}
	if restaurants == nil {
		return []model.Restaurant{}, nil
	}
	return restaurants, nil
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

func CreateCustomerOrderHandler(pool *pgxpool.Pool) gin.HandlerFunc {
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

func GetCustomerOrderByIdHandler(pool *pgxpool.Pool) gin.HandlerFunc {
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
		orderID := ctx.Param("id")
		if orderID == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "order id required"})
			return
		}
		parsedOrderID, err := uuid.Parse(orderID)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
			return
		}
		order, order_items, err := repository.GetCustomerOrderByID(ctx.Request.Context(), pool, parsedOrderID, parsedCustomerID)
		if err != nil {
			if errors.Is(err, repository.ErrOrderNotFound) {
				ctx.JSON(http.StatusNotFound, gin.H{"error": repository.ErrOrderNotFound.Error()})
				return
			}
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
		ctx.JSON(http.StatusOK, gin.H{"order": order, "order_items": order_items})
	}
}

func GetRestaurantOrdersHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userID := ctx.GetHeader("X-User-ID")
		if userID == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "user id required"})
			return
		}
		parsedUserID, err := uuid.Parse(userID)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
			return
		}
		role := ctx.GetHeader("X-User-Role")
		if role == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "role required"})
			return
		}
		if role != "restaurant" {
			ctx.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		myRestaurants, err := getMyRestaurants(parsedUserID, role)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		var myOrders []model.Order
		for _, restaurant := range myRestaurants {
			orders, err := repository.GetOrdersByRestaurantID(ctx.Request.Context(), pool, restaurant.ID)
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
				return
			}
			for _, order := range orders {
				myOrders = append(myOrders, order)
			}
		}
		if myOrders == nil {
			myOrders = []model.Order{}
		}
		ctx.JSON(http.StatusOK, myOrders)
	}
}

func GetRestaurantOrderByOrderIDHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userID := ctx.GetHeader("X-User-ID")
		if userID == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "user id required"})
			return
		}
		parsedUserID, err := uuid.Parse(userID)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
			return
		}
		role := ctx.GetHeader("X-User-Role")
		if role == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "role required"})
			return
		}
		if role != "restaurant" {
			ctx.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		orderID := ctx.Param("id")
		if orderID == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "order id required"})
			return
		}
		parsedOrderID, err := uuid.Parse(orderID)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
			return
		}
		order, err := repository.GetOrderByID(ctx.Request.Context(), pool, parsedOrderID)
		if err != nil {
			if errors.Is(err, repository.ErrOrderNotFound) {
				ctx.JSON(http.StatusNotFound, gin.H{"error": repository.ErrOrderNotFound.Error()})
				return
			}
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
		restaurant, err := getRestaurantByID(order.RestaurantID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
		if restaurant.OwnerID != parsedUserID {
			ctx.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		orderItems, err := repository.GetOrderItemsByOrderID(ctx.Request.Context(), pool, order.ID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
		ctx.JSON(http.StatusOK, gin.H{"order": order, "order_items": orderItems})
	}
}

func ChangeRestaurantOrderStatusHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var updateStatus model.UpdateOrderStatusRequest
		if err := ctx.ShouldBindJSON(&updateStatus); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}
		updateStatus.Normalize()
		if !validation.ValidateStatusRequest(updateStatus.Status) {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid status"})
			return
		}
		userID := ctx.GetHeader("X-User-ID")
		if userID == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "user id required"})
			return
		}
		parsedUserID, err := uuid.Parse(userID)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
			return
		}
		role := ctx.GetHeader("X-User-Role")
		if role == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "role required"})
			return
		}
		if role != "restaurant" {
			ctx.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		orderID := ctx.Param("id")
		if orderID == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "order id required"})
			return
		}
		parsedOrderID, err := uuid.Parse(orderID)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
			return
		}
		order, err := repository.GetOrderByID(ctx.Request.Context(), pool, parsedOrderID)
		if err != nil {
			if errors.Is(err, repository.ErrOrderNotFound) {
				ctx.JSON(http.StatusNotFound, gin.H{"error": repository.ErrOrderNotFound.Error()})
				return
			}
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
		restaurant, err := getRestaurantByID(order.RestaurantID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
		if restaurant.OwnerID != parsedUserID {
			ctx.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		if !validation.ValidateStatusTransition(order.Status, updateStatus.Status) {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid status transition"})
			return
		}
		updatedOrder, err := repository.UpdateOrderStatusByOrderID(ctx.Request.Context(), pool, updateStatus.Status, parsedOrderID)
		if err != nil {
			if errors.Is(err, repository.ErrOrderNotFound) {
				ctx.JSON(http.StatusNotFound, gin.H{"error": repository.ErrOrderNotFound.Error()})
				return
			}
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
		ctx.JSON(http.StatusOK, updatedOrder)
	}
}

func GetOrdersReadyForPickupHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		role := ctx.GetHeader("X-User-Role")
		if role == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "role required"})
			return
		}
		if role != "courier" {
			ctx.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		orders, err := repository.GetOrdersReadyForPickup(ctx.Request.Context(), pool)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
		ctx.JSON(http.StatusOK, orders)
	}
}

func UpdateCourierOrderStatusHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var updateStatus model.CourierUpdateOrderStatusRequest
		if err := ctx.ShouldBindJSON(&updateStatus); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}
		updateStatus.Normalize()
		if !validation.ValidateCourierStatusRequest(updateStatus.Status) {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid status"})
			return
		}
		role := ctx.GetHeader("X-User-Role")
		if role == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "role required"})
			return
		}
		if role != "courier" {
			ctx.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		orderID := ctx.Param("id")
		if orderID == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "order id required"})
			return
		}
		parsedOrderID, err := uuid.Parse(orderID)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
			return
		}
		order, err := repository.GetOrderByID(ctx.Request.Context(), pool, parsedOrderID)
		if err != nil {
			if errors.Is(err, repository.ErrOrderNotFound) {
				ctx.JSON(http.StatusNotFound, gin.H{"error": repository.ErrOrderNotFound.Error()})
				return
			}
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
		if !validation.ValidateCourierStatusTransition(order.Status, updateStatus.Status) {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid status transition"})
			return
		}
		updatedOrder, err := repository.UpdateOrderStatusByOrderID(ctx.Request.Context(), pool, updateStatus.Status, parsedOrderID)
		if err != nil {
			if errors.Is(err, repository.ErrOrderNotFound) {
				ctx.JSON(http.StatusNotFound, gin.H{"error": repository.ErrOrderNotFound.Error()})
				return
			}
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
		ctx.JSON(http.StatusOK, updatedOrder)
	}
}

func GetCourierOrderByOrderIDHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		role := ctx.GetHeader("X-User-Role")
		if role == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "role required"})
			return
		}
		if role != "courier" {
			ctx.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		orderID := ctx.Param("id")
		if orderID == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "order id required"})
			return
		}
		parsedOrderID, err := uuid.Parse(orderID)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
			return
		}
		order, err := repository.GetOrderByID(ctx.Request.Context(), pool, parsedOrderID)
		if err != nil {
			if errors.Is(err, repository.ErrOrderNotFound) {
				ctx.JSON(http.StatusNotFound, gin.H{"error": repository.ErrOrderNotFound.Error()})
				return
			}
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
		if !validation.ValidateCourierStatusRequest(order.Status) {
			ctx.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		ctx.JSON(http.StatusOK, order)
	}
}
