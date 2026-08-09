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

func CreateRestaurantHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req model.CreateRestaurantRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}
		req.Normalize()
		if err := validation.ValidateCreateRestaurantRequest(req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
		ownerID := ctx.GetHeader("X-User-ID")
		if ownerID == "" {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "user id required"})
			return
		}
		parsedOwnerID, err := uuid.Parse(ownerID)
		if err != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user id"})
			return
		}
		createdRestaurant := service.CreateRestaurantFromRequest(req, parsedOwnerID)
		createRestaurantResponse, err := repository.CreateRestaurant(ctx.Request.Context(), pool, createdRestaurant)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "couldnt create restaurant"})
			return
		}
		ctx.JSON(http.StatusCreated, createRestaurantResponse)
	}
}

func GetRestauranByIDHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		restaurantID := ctx.Param("id")
		if restaurantID == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "restaurant id required"})
			return
		}
		parsedRestaurantID, err := uuid.Parse(restaurantID)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid restaurant id"})
			return
		}
		restaurant, err := repository.GetRestaurantByID(ctx.Request.Context(), pool, parsedRestaurantID)
		if err != nil {
			if errors.Is(err, repository.ErrRestaurantNotFound) {
				ctx.JSON(http.StatusNotFound, gin.H{"error": repository.ErrRestaurantNotFound.Error()})
				return
			}
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
		ctx.JSON(http.StatusOK, restaurant)
	}
}

func GetAllRestaurants(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		restaurants, err := repository.GetAllRestaurants(ctx.Request.Context(), pool)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
		ctx.JSON(http.StatusOK, restaurants)
	}
}

func GetMenu(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		restaurantID := ctx.Param("id")
		if restaurantID == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "restaurant id required"})
			return
		}
		parsedRestaurantID, err := uuid.Parse(restaurantID)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid restaurant id"})
			return
		}
		menu, err := repository.GetMenu(ctx.Request.Context(), pool, parsedRestaurantID)
		if err != nil {
			if errors.Is(err, repository.ErrRestaurantNotFound) {
				ctx.JSON(http.StatusNotFound, gin.H{"error": repository.ErrRestaurantNotFound.Error()})
				return
			}
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
		ctx.JSON(http.StatusOK, menu)
	}
}
