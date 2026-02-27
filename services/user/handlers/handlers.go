package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// SetupRouter настраивает роутер user service
func SetupRouter(
	dbHost, dbPort, dbUser, dbPassword, dbName string,
	redisHost, redisPort string,
) *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())

	userHandler := NewUserHandler(dbHost, dbPort, dbUser, dbPassword, dbName, redisHost, redisPort)

	user := router.Group("/api/user")
	{
		user.GET("/profile/:id", userHandler.GetProfile)
		user.GET("/profile/me", userHandler.GetMyProfile)
		user.PUT("/profile/me", userHandler.UpdateProfile)
		user.GET("/status", userHandler.GetStatus)
		user.PUT("/status", userHandler.SetStatus)
		user.GET("/online", userHandler.GetOnlineUsers)
	}

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	return router
}

type UserHandler struct {
	// repository будет добавлен
}

func NewUserHandler(dbHost, dbPort, dbUser, dbPassword, dbName, redisHost, redisPort string) *UserHandler {
	return &UserHandler{}
}

func (h *UserHandler) GetProfile(c *gin.Context) {
	id := c.Param("id")
	userID, err := strconv.Atoi(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	// Получение профиля из БД будет добавлено

	c.JSON(http.StatusOK, gin.H{"id": userID, "username": "user" + id})
}

func (h *UserHandler) GetMyProfile(c *gin.Context) {
	// Получение профиля текущего пользователя
	c.JSON(http.StatusOK, gin.H{"message": "My profile"})
}

func (h *UserHandler) UpdateProfile(c *gin.Context) {
	// Обновление профиля
	c.JSON(http.StatusOK, gin.H{"message": "Profile updated"})
}

func (h *UserHandler) GetStatus(c *gin.Context) {
	// Получение статуса пользователя
	c.JSON(http.StatusOK, gin.H{"status": "online"})
}

func (h *UserHandler) SetStatus(c *gin.Context) {
	// Установка статуса
	c.JSON(http.StatusOK, gin.H{"message": "Status updated"})
}

func (h *UserHandler) GetOnlineUsers(c *gin.Context) {
	// Список онлайн пользователей
	c.JSON(http.StatusOK, gin.H{"users": []string{}})
}
