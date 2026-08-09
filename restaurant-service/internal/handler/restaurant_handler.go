package handler

import (
	"net/http"

	"github.com/ErenKarakus1/Food-Delivery-System/restaurant-service/internal/model"
	"github.com/ErenKarakus1/Food-Delivery-System/restaurant-service/internal/repository"
	"github.com/ErenKarakus1/Food-Delivery-System/restaurant-service/internal/service"
	"github.com/ErenKarakus1/Food-Delivery-System/restaurant-service/validation"
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
