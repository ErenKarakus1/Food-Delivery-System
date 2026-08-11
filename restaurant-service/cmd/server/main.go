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
	router.GET("/menu-items/:id", handler.GetMenuItemByIDHandler(pool))
	router.DELETE("/menu-items/:id", handler.DeleteMenuItemHandler(pool))
	router.PUT("/menu-items/:id", handler.UpdateMenuItemHandler(pool))
	router.POST("/restaurants", handler.CreateRestaurantHandler(pool))
	router.GET("/restaurants", handler.GetAllRestaurants(pool))
	router.GET("/restaurants/:id/menu", handler.GetMenu(pool))
	router.GET("/restaurants/:id", handler.GetRestauranByIDHandler(pool))
	router.GET("/restaurants/me", handler.GetMyRestaurants(pool))

	router.Run(":8081")
}
