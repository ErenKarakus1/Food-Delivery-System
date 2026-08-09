package main

import (
	"log"

	"github.com/ErenKarakus1/Food-Delivery-System/restaurant-service/internal/config"
	"github.com/ErenKarakus1/Food-Delivery-System/restaurant-service/internal/db"
	"github.com/ErenKarakus1/Food-Delivery-System/restaurant-service/internal/handler"
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

	router.POST("/menu-items", handler.CreateMenuItemHandler(pool))
	router.POST("/restaurants", handler.CreateRestaurantHandler(pool))

	router.Run(":8081")
}
