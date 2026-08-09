package handler

import (
	"errors"
	"net/http"

	"github.com/ErenKarakus1/Food-Delivery-System/auth-service/internal/jwt"
	"github.com/ErenKarakus1/Food-Delivery-System/auth-service/internal/model"
	"github.com/ErenKarakus1/Food-Delivery-System/auth-service/internal/repository"
	"github.com/ErenKarakus1/Food-Delivery-System/auth-service/internal/validation"
	"github.com/ErenKarakus1/Food-Delivery-System/auth-service/password"
	"github.com/ErenKarakus1/Food-Delivery-System/auth-service/service"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func RegisterHandler(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req model.RegisterRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}
		req.Normalize()
		if err := validation.ValidateRegister(req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		createdUser, err := service.CreateUserFromRegisterRequest(req)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		registerResponse, err := repository.RegisterUser(ctx.Request.Context(), pool, createdUser)
		if err != nil {
			if errors.Is(err, repository.ErrEmailExists) {
				ctx.JSON(http.StatusConflict, gin.H{"error": err.Error()})
				return
			}
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusCreated, registerResponse)
	}
}

func LoginHandler(pool *pgxpool.Pool, jwt_secret string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req model.LoginRequest
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
			return
		}
		req.Normalize()
		if err := validation.ValidateLogin(req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		user, err := repository.FindUserByEmail(ctx.Request.Context(), pool, req.Email)
		if err != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		if ok := password.CompareHashAndPassword(user.PasswordHash, req.Password); !ok {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "invalid email or password"})
			return
		}
		token, err := jwt.GenerateToken(user.ID, user.Role, jwt_secret)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
			return
		}
		ctx.JSON(http.StatusOK, model.LoginResponse{
			Token: token,
		})
	}
}
