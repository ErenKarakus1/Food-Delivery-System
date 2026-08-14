package main

import (
	"log"

	"github.com/ErenKarakus1/Food-Delivery-System/auth-service/internal/config"
	"github.com/ErenKarakus1/Food-Delivery-System/auth-service/internal/db"
	"github.com/ErenKarakus1/Food-Delivery-System/auth-service/internal/handler"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.LoadConfig()

	pool, err := db.NewPool(cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()
	log.Println("Connected to Postgres!")

	router := gin.Default()
	router.POST("/auth/register", handler.RegisterHandler(pool))
	router.POST("/auth/login", handler.LoginHandler(pool, cfg.JWT_SECRET))
	router.Run(":8084")
}
