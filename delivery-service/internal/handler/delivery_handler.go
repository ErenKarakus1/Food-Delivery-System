package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/ErenKarakus1/Food-Delivery-System/delivery-service/internal/model"
	"github.com/ErenKarakus1/Food-Delivery-System/delivery-service/internal/repository"
	"github.com/ErenKarakus1/Food-Delivery-System/delivery-service/internal/service"
	"github.com/ErenKarakus1/Food-Delivery-System/delivery-service/internal/validation"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func updateStatusOrderService(orderID uuid.UUID, status string) error {
	url := fmt.Sprintf("http://localhost:8082/orders/courier/%s", orderID)
	body := map[string]string{
		"status": status,
	}
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return errors.New("internal server error")
	}
	req, err := http.NewRequest(http.MethodPatch, url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return errors.New("internal server error")
	}
	req.Header.Set("X-User-Role", "courier")

	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return errors.New("order service error")
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return errors.New("order service error")
	}
	return nil

}

func GetCurrentDeliveryHandler(pool *pgxpool.Pool) gin.HandlerFunc {
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
		courierID := ctx.GetHeader("X-User-ID")
		if courierID == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "courier id required"})
			return
		}
		parsedCourierID, err := uuid.Parse(courierID)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid courier id"})
			return
		}
		currentDelivery, err := repository.GetCurrentDeliveryByCourierID(ctx.Request.Context(), pool, parsedCourierID)
		if err != nil {
			if errors.Is(err, repository.ErrDeliveryNotFound) {
				ctx.JSON(http.StatusNotFound, gin.H{"error": repository.ErrDeliveryNotFound.Error()})
				return
			}
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
		ctx.JSON(http.StatusOK, currentDelivery)
	}
}

func UnavailableHandler(pool *pgxpool.Pool) gin.HandlerFunc {
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
		courierID := ctx.GetHeader("X-User-ID")
		if courierID == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "courier id required"})
			return
		}
		parsedCourierID, err := uuid.Parse(courierID)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid courier id"})
			return
		}
		err = repository.SetUnavailable(ctx.Request.Context(), pool, parsedCourierID)
		if err != nil {
			if errors.Is(err, repository.ErrCourierNotFound) {
				ctx.JSON(http.StatusNotFound, gin.H{"error": repository.ErrCourierNotFound.Error()})
				return
			}
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
		ctx.Status(http.StatusOK)
	}
}

func AvailableHandler(pool *pgxpool.Pool) gin.HandlerFunc {
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
		courierID := ctx.GetHeader("X-User-ID")
		if courierID == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "courier id required"})
			return
		}
		parsedCourierID, err := uuid.Parse(courierID)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid courier id"})
			return
		}
		_, err = repository.GetCurrentDeliveryByCourierID(ctx.Request.Context(), pool, parsedCourierID)
		if err != nil {
			if errors.Is(err, repository.ErrDeliveryNotFound) {
				if err := repository.SetAvailable(ctx.Request.Context(), pool, parsedCourierID); err != nil {
					ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
					return
				}
				ctx.Status(http.StatusOK)
				return
			}
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
		ctx.JSON(http.StatusConflict, gin.H{"error": "delivery is active"})
	}
}

func PickUpDeliveryStatusHandler(pool *pgxpool.Pool) gin.HandlerFunc {
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
		courierID := ctx.GetHeader("X-User-ID")
		if courierID == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "courier id required"})
			return
		}
		parsedCourierID, err := uuid.Parse(courierID)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid courier id"})
			return
		}
		var statusRequest model.StatusUpdateRequest
		if err := ctx.ShouldBindJSON(&statusRequest); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}
		if !validation.ValidateStatus(statusRequest.Status) {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid status update"})
			return
		}
		currentDelivery, err := repository.GetCurrentDeliveryByCourierID(ctx.Request.Context(), pool, parsedCourierID)
		if err != nil {
			if errors.Is(err, repository.ErrDeliveryNotFound) {
				ctx.JSON(http.StatusNotFound, gin.H{"error": repository.ErrDeliveryNotFound.Error()})
				return
			}
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
		if !validation.ValidateStatusTransition(currentDelivery.Status, statusRequest.Status) {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid status transition"})
			return
		}
		updatedCurrentDelivery, err := repository.UpdateDeliveryStatus(ctx.Request.Context(), pool, currentDelivery.ID, statusRequest.Status)
		if err != nil {
			if errors.Is(err, repository.ErrDeliveryNotFound) {
				ctx.JSON(http.StatusNotFound, gin.H{"error": repository.ErrDeliveryNotFound.Error()})
				return
			}
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
		if err := updateStatusOrderService(currentDelivery.OrderID, statusRequest.Status); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, updatedCurrentDelivery)
	}
}

func DeliverDeliveryStatusHandler(pool *pgxpool.Pool) gin.HandlerFunc {
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
		courierID := ctx.GetHeader("X-User-ID")
		if courierID == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "courier id required"})
			return
		}
		parsedCourierID, err := uuid.Parse(courierID)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid courier id"})
			return
		}
		var statusRequest model.StatusUpdateRequest
		if err := ctx.ShouldBindJSON(&statusRequest); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}
		if !validation.ValidateStatus(statusRequest.Status) {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid status update"})
			return
		}
		currentDelivery, err := repository.GetCurrentDeliveryByCourierID(ctx.Request.Context(), pool, parsedCourierID)
		if err != nil {
			if errors.Is(err, repository.ErrDeliveryNotFound) {
				ctx.JSON(http.StatusNotFound, gin.H{"error": repository.ErrDeliveryNotFound.Error()})
				return
			}
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
		if !validation.ValidateStatusTransition(currentDelivery.Status, statusRequest.Status) {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid status transition"})
			return
		}
		updatedCurrentDelivery, err := repository.UpdateDeliveryStatus(ctx.Request.Context(), pool, currentDelivery.ID, statusRequest.Status)
		if err != nil {
			if errors.Is(err, repository.ErrDeliveryNotFound) {
				ctx.JSON(http.StatusNotFound, gin.H{"error": repository.ErrDeliveryNotFound.Error()})
				return
			}
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
		if err := updateStatusOrderService(currentDelivery.OrderID, statusRequest.Status); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if err := repository.SetAvailable(ctx.Request.Context(), pool, parsedCourierID); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
		ctx.JSON(http.StatusOK, updatedCurrentDelivery)
	}
}

func RejectDeliveryHandler(pool *pgxpool.Pool) gin.HandlerFunc {
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
		courierID := ctx.GetHeader("X-User-ID")
		if courierID == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "courier id required"})
			return
		}
		parsedCourierID, err := uuid.Parse(courierID)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid courier id"})
			return
		}
		currentDelivery, err := repository.GetCurrentDeliveryByCourierID(ctx.Request.Context(), pool, parsedCourierID)
		if err != nil {
			if errors.Is(err, repository.ErrDeliveryNotFound) {
				ctx.JSON(http.StatusNotFound, gin.H{"error": repository.ErrDeliveryNotFound.Error()})
				return
			}
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
		if err := repository.RejectDelivery(ctx.Request.Context(), pool, currentDelivery.ID, parsedCourierID); err != nil {
			if errors.Is(err, repository.ErrDeliveryNotFound) {
				ctx.JSON(http.StatusNotFound, gin.H{"error": repository.ErrDeliveryNotFound.Error()})
				return
			}
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
		deliveryRejection := service.CreateDeliveryRejection(currentDelivery.ID, parsedCourierID)
		if err := repository.CreateDeliveryRejection(ctx.Request.Context(), pool, deliveryRejection); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
		if err := repository.SetAvailable(ctx.Request.Context(), pool, parsedCourierID); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
		ctx.Status(http.StatusOK)
	}
}

func GetMeHandler(pool *pgxpool.Pool) gin.HandlerFunc {
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
		courierID := ctx.GetHeader("X-User-ID")
		if courierID == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "courier id required"})
			return
		}
		parsedCourierID, err := uuid.Parse(courierID)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid courier id"})
			return
		}
		courier, err := repository.GetCourierByID(ctx.Request.Context(), pool, parsedCourierID)
		if err != nil {
			if errors.Is(err, repository.ErrCourierNotFound) {
				ctx.JSON(http.StatusNotFound, gin.H{"error": repository.ErrCourierNotFound.Error()})
				return
			}
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
		ctx.JSON(http.StatusOK, courier)
	}
}

func CreateCourierHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req model.CreateCourierRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, errors.New("invalid create courier request body"))
			return
		}
		parsedRequestID, err := uuid.Parse(req.ID)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid create courier id"})
			return
		}
		err = repository.CreateCourier(ctx.Request.Context(), pool, parsedRequestID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
		ctx.Status(http.StatusOK)
	}
}

func GetDeliveryByOrderIDHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req model.GetDeliveryByOrderIDRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}
		if req.OrderID == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "order id required"})
			return
		}
		orderID, err := uuid.Parse(req.OrderID)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
			return
		}
		delivery, err := repository.GetDeliveryByOrderID(ctx.Request.Context(), pool, orderID)
		if err != nil {
			if errors.Is(err, repository.ErrDeliveryNotFound) {
				ctx.JSON(http.StatusNotFound, gin.H{"error": repository.ErrDeliveryNotFound.Error()})
				return
			}
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		}
		ctx.JSON(http.StatusOK, delivery)
	}
}
