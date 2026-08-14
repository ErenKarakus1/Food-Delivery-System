package main

import (
	"log"

	"github.com/ErenKarakus1/Food-Delivery-System/order-service/internal/config"
	"github.com/ErenKarakus1/Food-Delivery-System/order-service/internal/db"
	"github.com/ErenKarakus1/Food-Delivery-System/order-service/internal/handler"
	"github.com/ErenKarakus1/Food-Delivery-System/order-service/rabbitmq"
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

	publisher, err := rabbitmq.NewPublisher(cfg.RabbitmqURL)
	if err != nil {
		log.Fatal(err)
	}
	defer publisher.Close()

	router := gin.Default()
	router.GET("/orders/customer", handler.GetCustomerOrdersHandler(pool))
	router.POST("/orders/customer", handler.CreateCustomerOrderHandler(pool))
	router.GET("/orders/customer/:id", handler.GetCustomerOrderByIdHandler(pool))
	router.GET("/orders/restaurant", handler.GetRestaurantOrdersHandler(pool))
	router.GET("/orders/restaurant/:id", handler.GetRestaurantOrderByOrderIDHandler(pool))
	router.PATCH("/orders/restaurant/:id", handler.ChangeRestaurantOrderStatusHandler(pool, publisher))
	router.GET("/orders/courier/ready-for-pickup", handler.GetOrdersReadyForPickupHandler(pool))
	router.GET("/orders/courier/:id", handler.GetCourierOrderByOrderIDHandler(pool))
	router.PATCH("/orders/courier/:id", handler.UpdateCourierOrderStatusHandler(pool))
	router.PATCH("/orders/courier/:id/delivery_created", handler.DeliveryCreatedHandler(pool))
	router.Run(":8082")
}
