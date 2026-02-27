package main

import (
	"log"
	"os"

	"github.com/stalknet/gateway/handlers"
	"github.com/stalknet/gateway/middleware"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	authURL := os.Getenv("AUTH_SERVICE_URL")
	userURL := os.Getenv("USER_SERVICE_URL")
	chatURL := os.Getenv("CHAT_SERVICE_URL")
	taskURL := os.Getenv("TASK_SERVICE_URL")

	if authURL == "" {
		authURL = "http://localhost:8081"
	}
	if userURL == "" {
		userURL = "http://localhost:8082"
	}
	if chatURL == "" {
		chatURL = "http://localhost:8083"
	}
	if taskURL == "" {
		taskURL = "http://localhost:8084"
	}

	router := handlers.SetupRouter(
		authURL,
		userURL,
		chatURL,
		taskURL,
	)

	router.Use(middleware.CORS())
	router.Use(middleware.Logging())

	log.Printf("Gateway starting on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start gateway: %v", err)
	}
}
