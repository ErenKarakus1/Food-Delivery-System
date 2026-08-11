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
	router.GET("/orders/customer", handler.GetCustomerOrdersHandler(pool))
	router.POST("orders/customer", handler.CreateCustomerOrderHandler(pool))
	router.GET("orders/customer/:id", handler.GetCustomerOrderByIdHandler(pool))
	router.GET("/orders/restaurant", handler.GetRestaurantOrdersHandler(pool))
	router.GET("/orders/restaurant/:id", handler.GetRestaurantOrderByOrderIDHandler(pool))
	router.PATCH("/orders/restaurant/:id", handler.ChangeRestaurantOrderStatusHandler(pool))
	router.GET("/orders/courier/ready-for-pickup", handler.GetOrdersReadyForPickupHandler(pool))
	router.PATCH("/orders/courier/:id", handler.UpdateCourierOrderStatusHandler(pool))
	router.Run(":8082")
}
