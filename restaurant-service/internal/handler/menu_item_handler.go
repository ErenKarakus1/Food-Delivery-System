package handler

import (
	"errors"
	"net/http"

	"github.com/ErenKarakus1/Food-Delivery-System/restaurant-service/internal/model"
	"github.com/ErenKarakus1/Food-Delivery-System/restaurant-service/internal/repository"
	"github.com/ErenKarakus1/Food-Delivery-System/restaurant-service/internal/service"
	"github.com/ErenKarakus1/Food-Delivery-System/restaurant-service/internal/validation"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func CreateMenuItemHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req model.CreateMenuItemRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}
		role := ctx.GetHeader("X-User-Role")
		if role == "" {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "role required"})
			return
		}
		if role != "restaurant" {
			ctx.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		req.Normalize()
		if err := validation.ValidateCreateMenuItemRequest(req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		userID := ctx.GetHeader("X-User-ID")
		if userID == "" {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "user id required"})
			return
		}
		parsedUserID, err := uuid.Parse(userID)
		if err != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id"})
			return
		}
		if _, err := repository.RestaurantExistsAndOwned(ctx.Request.Context(), pool, req.RestaurantID, parsedUserID); err != nil {
			if errors.Is(err, repository.ErrRestaurantNotFound) {
				ctx.JSON(http.StatusNotFound, gin.H{"error": repository.ErrRestaurantNotFound.Error()})
				return
			}
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
		createdMenuItem := service.CreateMenuItemFromRequest(req)
		menuItemResponse, err := repository.CreateMenuItem(ctx.Request.Context(), pool, createdMenuItem)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusCreated, menuItemResponse)
	}
}

func GetMenuItemByIDHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		menuItemID := ctx.Param("id")
		if menuItemID == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "menu item id required"})
			return
		}
		parsedMenuItemID, err := uuid.Parse(menuItemID)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid menu item id"})
			return
		}
		menuItem, err := repository.GetMenuItemByID(ctx.Request.Context(), pool, parsedMenuItemID)
		if err != nil {
			if errors.Is(err, repository.ErrMenuItemNotFound) {
				ctx.JSON(http.StatusNotFound, gin.H{"error": repository.ErrMenuItemNotFound.Error()})
				return
			}
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
		ctx.JSON(http.StatusOK, menuItem)
	}
}

func DeleteMenuItemHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		role := ctx.GetHeader("X-User-Role")
		if role == "" {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "role required"})
			return
		}
		if role != "restaurant" {
			ctx.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		userID := ctx.GetHeader("X-User-ID")
		if userID == "" {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "user id required"})
			return
		}
		parsedUserID, err := uuid.Parse(userID)
		if err != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id"})
			return
		}
		itemID := ctx.Param("id")
		if itemID == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "menu item id required"})
			return
		}
		parsedItemID, err := uuid.Parse(itemID)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid menu item id"})
			return
		}
		menuItem, err := repository.GetMenuItemByID(ctx.Request.Context(), pool, parsedItemID)
		if err != nil {
			if errors.Is(err, repository.ErrMenuItemNotFound) {
				ctx.JSON(http.StatusNotFound, gin.H{"error": repository.ErrMenuItemNotFound.Error()})
				return
			}
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
		restaurant, err := repository.GetRestaurantByID(ctx.Request.Context(), pool, menuItem.RestaurantID)
		if err != nil {
			if errors.Is(err, repository.ErrRestaurantNotFound) {
				ctx.JSON(http.StatusNotFound, gin.H{"error": repository.ErrRestaurantNotFound.Error()})
				return
			}
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
		if parsedUserID != restaurant.OwnerID {
			ctx.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		err = repository.DeleteMenuItem(ctx.Request.Context(), pool, parsedItemID)
		if err != nil {
			if errors.Is(err, repository.ErrMenuItemNotFound) {
				ctx.JSON(http.StatusNotFound, gin.H{"error": repository.ErrMenuItemNotFound.Error()})
				return
			}
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
		ctx.Status(http.StatusNoContent)
	}
}

func UpdateMenuItemHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req model.CreateMenuItemRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}

		req.Normalize()

		if err := validation.ValidateCreateMenuItemRequest(req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		itemID := ctx.Param("id")
		if itemID == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "menu item id required"})
			return
		}
		parsedItemID, err := uuid.Parse(itemID)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid menu item id"})
			return
		}
		menuItem, err := repository.GetMenuItemByID(ctx.Request.Context(), pool, parsedItemID)
		if err != nil {
			if errors.Is(err, repository.ErrMenuItemNotFound) {
				ctx.JSON(http.StatusNotFound, gin.H{"error": repository.ErrMenuItemNotFound.Error()})
				return
			}
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}

		role := ctx.GetHeader("X-User-Role")
		if role == "" {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "role required"})
			return
		}
		if role != "restaurant" {
			ctx.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		userID := ctx.GetHeader("X-User-ID")
		if userID == "" {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "user id required"})
			return
		}
		parsedUserID, err := uuid.Parse(userID)
		if err != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id"})
			return
		}

		restaurant, err := repository.GetRestaurantByID(ctx.Request.Context(), pool, menuItem.RestaurantID)
		if err != nil {
			if errors.Is(err, repository.ErrRestaurantNotFound) {
				ctx.JSON(http.StatusNotFound, gin.H{"error": repository.ErrRestaurantNotFound.Error()})
				return
			}
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
		if parsedUserID != restaurant.OwnerID {
			ctx.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		updatedMenuItem, err := repository.UpdateMenuItem(ctx.Request.Context(), pool, parsedItemID, req)
		if err != nil {
			if errors.Is(err, repository.ErrMenuItemNotFound) {
				ctx.JSON(http.StatusNotFound, gin.H{"error": repository.ErrMenuItemNotFound.Error()})
				return
			}
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
		ctx.JSON(http.StatusOK, updatedMenuItem)
	}
}
