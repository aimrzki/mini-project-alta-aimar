package main

import (
	"github.com/joho/godotenv"
	"log"
	"myproject/config"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}
	// Setup router
	router := config.SetupRouter()

	// Mulai server Echo pada alamat dan port tertentu (misalnya, :8080)
	err := router.Start(":8080")
	if err != nil {
		log.Fatal(err)
	}
}
