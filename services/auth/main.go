package main

import (
	"log"
	"os"

	"github.com/stalknet/services/auth/handlers"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}

	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	redisHost := os.Getenv("REDIS_HOST")
	redisPort := os.Getenv("REDIS_PORT")
	jwtSecret := os.Getenv("JWT_SECRET")

	router := handlers.SetupRouter(
		dbHost, dbPort, dbUser, dbPassword, dbName,
		redisHost, redisPort,
		jwtSecret,
	)

	log.Printf("Auth service starting on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start auth service: %v", err)
	}
}
