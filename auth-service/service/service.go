package service

import (
	"github.com/ErenKarakus1/Food-Delivery-System/auth-service/internal/model"
	"github.com/ErenKarakus1/Food-Delivery-System/auth-service/password"
	"github.com/google/uuid"
)

func CreateUserFromRegisterRequest(req model.RegisterRequest) (model.User, error) {
	passwordHash, err := password.HashPassword(req.Password)
	if err != nil {
		return model.User{}, err
	}
	return model.User{
		ID:           uuid.New(),
		Name:         req.Name,
		Email:        req.Email,
		PasswordHash: passwordHash,
		Role:         req.Role,
	}, nil
}
