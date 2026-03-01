package main

import (
	"log"
	"os"

	"github.com/stalknet/services/compliance/handlers"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8086"
	}

	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	// Значения по умолчанию
	if dbHost == "" {
		dbHost = "localhost"
	}
	if dbPort == "" {
		dbPort = "5432"
	}
	if dbUser == "" {
		dbUser = "stalknet"
	}
	if dbPassword == "" {
		dbPassword = "stalknet_secret"
	}
	if dbName == "" {
		dbName = "stalknet"
	}

	router := handlers.SetupRouter(dbHost, dbPort, dbUser, dbPassword, dbName)

	log.Printf("Compliance service starting on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start compliance service: %v", err)
	}
}
