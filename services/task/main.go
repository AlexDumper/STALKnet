package main

import (
	"log"
	"os"

	"github.com/stalknet/services/task/handlers"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8084"
	}

	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	router := handlers.SetupRouter(dbHost, dbPort, dbUser, dbPassword, dbName)

	log.Printf("Task service starting on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start task service: %v", err)
	}
}
