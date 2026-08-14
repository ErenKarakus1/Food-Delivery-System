package main

import (
	"github.com/ErenKarakus1/Food-Delivery-System/api-gateway/internal/config"
	"github.com/ErenKarakus1/Food-Delivery-System/api-gateway/internal/middleware"
	"github.com/ErenKarakus1/Food-Delivery-System/api-gateway/internal/proxy"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.LoadConfig()
	router := gin.Default()

	restaurantProxy := proxy.NewProxy("http://localhost:8081")
	orderProxy := proxy.NewProxy("http://localhost:8082")
	deliveryProxy := proxy.NewProxy("http://localhost:8083")
	authProxy := proxy.NewProxy("http://localhost:8084")

	router.POST("/auth/register", authProxy)
	router.POST("/auth/login", authProxy)

	router.GET("/menu-items/:id", restaurantProxy)
	router.GET("/restaurants", restaurantProxy)
	router.GET("/restaurants/:id/menu", restaurantProxy)
	router.GET("/restaurants/:id", restaurantProxy)

	auth := middleware.AuthMiddleware(cfg.JWT_SECRET)
	protected := router.Group("/")
	protected.Use(auth)
	{
		protected.GET("/couriers/me", deliveryProxy)
		protected.PATCH("/couriers/me/available", deliveryProxy)
		protected.PATCH("/couriers/me/unavailable", deliveryProxy)
		protected.GET("/deliveries/me", deliveryProxy)
		protected.PATCH("/deliveries/me/pickup", deliveryProxy)
		protected.PATCH("/deliveries/me/deliver", deliveryProxy)
		protected.PATCH("/deliveries/me/reject", deliveryProxy)

		protected.POST("/menu-items", restaurantProxy)
		protected.DELETE("/menu-items/:id", restaurantProxy)
		protected.PUT("/menu-items/:id", restaurantProxy)
		protected.POST("/restaurants", restaurantProxy)
		protected.GET("/restaurants/me", restaurantProxy)

		protected.GET("/orders/customer", orderProxy)
		protected.POST("/orders/customer", orderProxy)
		protected.GET("/orders/customer/:id", orderProxy)
		protected.GET("/orders/restaurant", orderProxy)
		protected.GET("/orders/restaurant/:id", orderProxy)
		protected.PATCH("/orders/restaurant/:id", orderProxy)
		protected.GET("/orders/courier/:id", orderProxy)
	}
	router.Run(":8080")
}
