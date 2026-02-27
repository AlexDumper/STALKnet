package main

import (
	"log"
	"os"

	"github.com/stalknet/services/chat/handlers"
	"github.com/stalknet/services/chat/hub"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8083"
	}

	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	redisHost := os.Getenv("REDIS_HOST")
	redisPort := os.Getenv("REDIS_PORT")

	// Создаём хаб для управления WebSocket соединениями
	wsHub := hub.NewHub()
	go wsHub.Run()

	router := handlers.SetupRouter(
		dbHost, dbPort, dbUser, dbPassword, dbName,
		redisHost, redisPort,
		wsHub,
	)

	log.Printf("Chat service starting on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start chat service: %v", err)
	}
}
