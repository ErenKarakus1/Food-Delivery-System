package main

import (
	"context"
	"log"

	"github.com/ErenKarakus1/Food-Delivery-System/delivery-service/internal/config"
	"github.com/ErenKarakus1/Food-Delivery-System/delivery-service/internal/db"
	"github.com/ErenKarakus1/Food-Delivery-System/delivery-service/internal/handler"
	"github.com/ErenKarakus1/Food-Delivery-System/delivery-service/internal/worker"
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

	worker := worker.NewWorker(pool)
	go worker.Start(context.Background())
	router := gin.Default()

	router.POST("/couriers/create", handler.CreateCourierHandler(pool))
	router.GET("/couriers/me", handler.GetMeHandler(pool))
	router.PATCH("/couriers/me/available", handler.AvailableHandler(pool))
	router.PATCH("/couriers/me/unavailable", handler.UnavailableHandler(pool))
	router.GET("/deliveries/me", handler.GetCurrentDeliveryHandler(pool))
	router.PATCH("/deliveries/me/pickup", handler.PickUpDeliveryStatusHandler(pool))
	router.PATCH("/deliveries/me/deliver", handler.DeliverDeliveryStatusHandler(pool))
	router.PATCH("/deliveries/me/reject", handler.RejectDeliveryHandler(pool))
	router.Run(":8083")
}
