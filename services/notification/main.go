package main

import (
	"log"
	"os"

	"github.com/stalknet/services/notification/handlers"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8085"
	}

	redisHost := os.Getenv("REDIS_HOST")
	redisPort := os.Getenv("REDIS_PORT")

	router := handlers.SetupRouter(redisHost, redisPort)

	log.Printf("Notification service starting on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start notification service: %v", err)
	}
}
