package validation

import (
	"errors"
	"net/mail"

	"github.com/ErenKarakus1/Food-Delivery-System/auth-service/internal/model"
)

const (
	roleCustomer   = "customer"
	roleRestaurant = "restaurant"
	roleCourier    = "courier"
)

func validateRole(role string) bool {
	switch role {
	case roleCustomer, roleRestaurant, roleCourier:
		return true
	default:
		return false
	}
}

func ValidateRegister(req model.RegisterRequest) error {
	if _, err := mail.ParseAddress(req.Email); err != nil {
		return errors.New("invalid email")
	}
	if len(req.Name) < 3 {
		return errors.New("name must be at least 3 characters")
	}
	if len(req.Name) > 50 {
		return errors.New("name must be at most 50 characters")
	}
	if len(req.Password) < 8 {
		return errors.New("password must be at least 8 characters")
	}
	if len(req.Password) > 80 {
		return errors.New("password must be at most 80 characters")
	}
	if req.Role == "" {
		return errors.New("role is required")
	}
	if ok := validateRole(req.Role); !ok {
		return errors.New("invalid role")
	}
	return nil
}

func ValidateLogin(req model.LoginRequest) error {
	if _, err := mail.ParseAddress(req.Email); err != nil {
		return errors.New("invalid email")
	}
	if req.Password == "" {
		return errors.New("password is required")
	}
	return nil
}
