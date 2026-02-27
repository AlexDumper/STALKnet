package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// SetupRouter настраивает роутер auth service
func SetupRouter(
	dbHost, dbPort, dbUser, dbPassword, dbName string,
	redisHost, redisPort, jwtSecret string,
) *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())

	authHandler := NewAuthHandler(
		dbHost, dbPort, dbUser, dbPassword, dbName,
		redisHost, redisPort,
		jwtSecret,
	)

	auth := router.Group("/api/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
		auth.POST("/logout", authHandler.Logout)
		auth.POST("/refresh", authHandler.Refresh)
	}

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	return router
}
