package main

import (
	"log"

	"github.com/ErenKarakus1/Food-Delivery-System/order-service/internal/config"
	"github.com/ErenKarakus1/Food-Delivery-System/order-service/internal/db"
	"github.com/ErenKarakus1/Food-Delivery-System/order-service/internal/handler"
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
	router.GET("/orders", handler.GetCustomerOrdersHandler(pool))
	router.POST("/orders", handler.CreateOrderHandler(pool))
	router.Run(":8082")
}
