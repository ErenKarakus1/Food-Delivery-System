package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/ErenKarakus1/Food-Delivery-System/auth-service/internal/jwt"
	"github.com/ErenKarakus1/Food-Delivery-System/auth-service/internal/model"
	"github.com/ErenKarakus1/Food-Delivery-System/auth-service/internal/password"
	"github.com/ErenKarakus1/Food-Delivery-System/auth-service/internal/repository"
	"github.com/ErenKarakus1/Food-Delivery-System/auth-service/internal/service"
	"github.com/ErenKarakus1/Food-Delivery-System/auth-service/internal/validation"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func createCourier(courierID uuid.UUID) error {
	url := "http://localhost:8083/couriers/create"
	payload := map[string]string{
		"id": courierID.String(),
	}
	jsonBody, err := json.Marshal(payload)
	if err != nil {
		return errors.New("internal server error")
	}
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonBody))
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return errors.New("delivery service error")
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return errors.New("delivery service error")
	}
	return nil
}

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
		if createdUser.Role == "courier" {
			err := createCourier(createdUser.ID)
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
				return
			}
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
