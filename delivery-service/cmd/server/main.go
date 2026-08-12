package main

import (
	"log"

	"github.com/ErenKarakus1/Food-Delivery-System/delivery-service/internal/config"
	"github.com/ErenKarakus1/Food-Delivery-System/delivery-service/internal/db"
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

	router.Run(":8083")
}
